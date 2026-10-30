// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// One Workspace extensions, copyright 2026 One Workspace contributors.

package models

import (
	"sort"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type OwPoint struct {
	Label     string  `json:"label" doc:"Accessible label for this observation."`
	Value     float64 `json:"value" doc:"Computed value."`
	Secondary float64 `json:"secondary" doc:"Second series value, where applicable."`
}
type OwWidgetData struct {
	ID         string        `json:"id" doc:"Widget ID."`
	Kind       string        `json:"kind" doc:"Calculation kind."`
	Title      string        `json:"title" doc:"Widget title."`
	Value      float64       `json:"value" doc:"Primary computed metric."`
	Total      float64       `json:"total" doc:"Denominator for progress metrics."`
	Points     []OwPoint     `json:"points" doc:"Computed chart observations, including zero-value days."`
	Issues     []*OwIssue    `json:"issues" doc:"Authorized recent issue rows."`
	Activities []*OwActivity `json:"activities" doc:"Authorized activity feed."`
}
type OwDashboardData struct {
	Dashboard *OwDashboard   `json:"dashboard" doc:"Dashboard definition."`
	Widgets   []OwWidgetData `json:"widgets" doc:"Metrics computed with the current viewer's permissions."`
	Generated time.Time      `json:"generated" doc:"Server calculation time."`
	Timezone  string         `json:"timezone" doc:"Server reporting timezone."`
}

func userForOw(s *xorm.Session, a web.Auth) (*user.User, error) {
	if _, ok := a.(*user.User); !ok {
		return nil, owDenied()
	}
	return user.GetUserByID(s, a.GetID())
}

//nolint:gocyclo // Metric variants use one permission-filtered cohort and one completion definition.
func OwMetric(s *xorm.Session, a web.Auth, kind string, filter OwFilter) (OwWidgetData, error) {
	result := OwWidgetData{Kind: kind, Points: []OwPoint{}, Issues: []*OwIssue{}, Activities: []*OwActivity{}}
	issues, err := OwQuery(s, a, filter)
	if err != nil {
		return result, err
	}
	settings := map[int64]*OwProjectSettings{}
	for _, i := range issues {
		if settings[i.ProjectID] == nil {
			v, err := owSettings(s, i.ProjectID)
			if err != nil {
				return result, err
			}
			settings[i.ProjectID] = v
		}
	}
	now := time.Now()
	groups := map[string]float64{}
	created := map[string]float64{}
	resolved := map[string]float64{}
	cycleCount := 0.0
	cycleHours := 0.0
	ids := make([]int64, 0, len(issues))
	for _, i := range issues {
		ids = append(ids, i.TaskID)
		state := owState(settings[i.ProjectID], i.State)
		if state == nil {
			continue
		}
		open := state.Category == "open"
		done := state.Category == "done"
		created[i.Created.In(config.GetTimeZone()).Format("2006-01-02")]++
		if done && !i.CompletedAt.IsZero() {
			resolved[i.CompletedAt.In(config.GetTimeZone()).Format("2006-01-02")]++
			if !i.StartedAt.IsZero() && !i.CompletedAt.Before(i.StartedAt) {
				cycleHours += i.CompletedAt.Sub(i.StartedAt).Hours()
				cycleCount++
			}
		}
		switch kind {
		case "open":
			if open {
				result.Value++
			}
		case "completed":
			if done {
				result.Value++
			}
		case "overdue":
			if open && !i.DueDate.IsZero() && i.DueDate.Before(now) {
				result.Value++
			}
		case "unassigned":
			if open && len(i.AssigneeIDs) == 0 {
				result.Value++
			}
		case "bugs-by-priority":
			if i.Type == "Bug" && open {
				groups[strconv.FormatInt(i.Priority, 10)]++
			}
		case "by-status":
			groups[state.Name]++
		case "by-assignee", "workload":
			if len(i.AssigneeIDs) == 0 {
				groups["Unassigned"]++
			} else {
				for _, id := range i.AssigneeIDs {
					label := strconv.FormatInt(id, 10)
					groups[label]++
					if kind == "workload" {
						groups[label] += i.Estimate - 1
					}
				}
			}
		case "sprint-progress":
			result.Total++
			if done {
				result.Value++
			}
		case "aging":
			if open {
				age := int(now.Sub(i.Created).Hours() / 24)
				bucket := "0–7 days"
				switch {
				case age > 30:
					bucket = "31+ days"
				case age > 14:
					bucket = "15–30 days"
				case age > 7:
					bucket = "8–14 days"
				}
				groups[bucket]++
			}
		case "epic-progress":
			if i.Type == "Epic" {
				if _, exists := groups[i.Key]; !exists {
					groups[i.Key] = 0
				}
			}
			if i.EpicID != 0 && done {
				epic := &OwIssue{TaskID: i.EpicID}
				ok, _, lookupErr := epic.CanRead(s, a)
				if lookupErr != nil {
					return result, lookupErr
				}
				if ok {
					if lookupErr = epic.ReadOne(s, a); lookupErr != nil {
						return result, lookupErr
					}
					groups[epic.Key]++
				}
			}
		}
	}
	if kind == "cycle-time" && cycleCount > 0 {
		result.Value = cycleHours / cycleCount
	}
	if kind == "created-resolved" || kind == "throughput" {
		date := now.In(config.GetTimeZone())
		for d := 29; d >= 0; d-- {
			day := date.AddDate(0, 0, -d).Format("2006-01-02")
			point := OwPoint{Label: day, Value: created[day], Secondary: resolved[day]}
			if kind == "throughput" {
				point.Value = resolved[day]
				point.Secondary = 0
			}
			result.Points = append(result.Points, point)
		}
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		label := key
		if kind == "by-assignee" || kind == "workload" {
			if id, err := strconv.ParseInt(key, 10, 64); err == nil {
				person, err := user.GetUserByID(s, id)
				if err != nil {
					return result, err
				}
				label = person.GetName()
			}
		}
		result.Points = append(result.Points, OwPoint{Label: label, Value: groups[key]})
	}
	if kind == "recent" {
		sort.Slice(issues, func(i, j int) bool { return issues[i].Updated.After(issues[j].Updated) })
		result.Issues = issues[:min(len(issues), 10)]
	}
	if kind == "activity" {
		if err = s.In("task_id", ids).Desc("id").Limit(20).Find(&result.Activities); err != nil {
			return result, err
		}
	}
	if kind == "velocity" {
		projects, err := owAccessibleProjects(s, a)
		if err != nil {
			return result, err
		}
		sprints := []*OwSprint{}
		query := s.In("project_id", projects).And("state = ?", "closed")
		if filter.ProjectID != 0 {
			query = query.And("project_id = ?", filter.ProjectID)
		}
		if err = query.Desc("closed").Limit(12).Find(&sprints); err != nil {
			return result, err
		}
		for i := len(sprints) - 1; i >= 0; i-- {
			sp := sprints[i]
			result.Points = append(result.Points, OwPoint{Label: sp.Name, Value: float64(len(sp.Completed)), Secondary: float64(len(sp.Committed))})
		}
	}
	if kind == "burndown" || kind == "cumulative-flow" {
		return OwHistoricalMetric(s, a, kind, filter, issues)
	}
	return result, nil
}

func owMergeFilter(base, override OwFilter) OwFilter {
	if override.ProjectID != 0 {
		base.ProjectID = override.ProjectID
	}
	if override.SprintID != 0 {
		base.SprintID = override.SprintID
	}
	if override.EpicID != 0 {
		base.EpicID = override.EpicID
	}
	if override.Q != "" {
		base.Q = override.Q
	}
	if len(override.Rules) > 0 {
		base.Rules = append(base.Rules, override.Rules...)
	}
	return base
}

func OwDashboardMetrics(s *xorm.Session, a web.Auth, id int64) (*OwDashboardData, error) {
	d := &OwDashboard{ID: id}
	allowed, _, err := d.CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, owDenied()
	}
	if err = d.ReadOne(s, a); err != nil {
		return nil, err
	}
	layout := &OwDashboardLayout{}
	found, err := s.Where("dashboard_id = ? AND user_id = ?", id, a.GetID()).Get(layout)
	if err != nil {
		return nil, err
	}
	if found {
		for index, widget := range d.Widgets {
			for _, personal := range layout.Widgets {
				if widget.ID == personal.ID {
					d.Widgets[index].X = personal.X
					d.Widgets[index].Y = personal.Y
					d.Widgets[index].W = personal.W
					d.Widgets[index].H = personal.H
				}
			}
		}
	}
	result := &OwDashboardData{Dashboard: d, Widgets: []OwWidgetData{}, Generated: time.Now(), Timezone: config.ServiceTimeZone.GetString()}
	filter := d.Filter
	if d.Scope == "project" {
		filter.ProjectID = d.ProjectID
	}
	for _, widget := range d.Widgets {
		data, err := OwMetric(s, a, widget.Kind, owMergeFilter(filter, widget.Filter))
		if err != nil {
			return nil, err
		}
		data.ID = widget.ID
		data.Title = widget.Title
		result.Widgets = append(result.Widgets, data)
	}
	return result, nil
}

func OwDashboardTemplate(name string, project int64) *OwDashboard {
	kinds := []string{"open", "completed", "overdue", "unassigned", "by-status", "created-resolved", "recent", "activity"}
	switch name {
	case "Engineering":
		kinds = []string{"open", "bugs-by-priority", "workload", "cycle-time", "throughput", "aging", "recent", "activity"}
	case "Sprint Overview":
		kinds = []string{"sprint-progress", "burndown", "velocity", "by-assignee", "recent", "activity"}
	case "Project Overview":
		kinds = []string{"open", "completed", "overdue", "unassigned", "by-status", "by-assignee", "created-resolved", "activity"}
	default:
		name = "Personal"
	}
	d := &OwDashboard{Name: name, ProjectID: project, Scope: "private", Filter: OwFilter{ProjectID: project}, Widgets: []OwWidget{}}
	for index, kind := range kinds {
		d.Widgets = append(d.Widgets, OwWidget{ID: kind + "-" + strconv.Itoa(index), Kind: kind, Title: owWidgetTitle(kind), X: (index % 2) * 6, Y: (index / 2) * 3, W: 6, H: 3})
	}
	return d
}

func owWidgetTitle(kind string) string {
	names := map[string]string{"open": "Open issues", "completed": "Completed issues", "overdue": "Overdue", "unassigned": "Unassigned", "by-status": "Status distribution", "by-assignee": "Issues by assignee", "created-resolved": "Created vs resolved", "recent": "Recent issues", "activity": "Activity", "bugs-by-priority": "Bugs by priority", "workload": "Workload", "cycle-time": "Average cycle time", "throughput": "Throughput", "aging": "Issue age", "cumulative-flow": "Cumulative flow", "burndown": "Sprint burndown", "velocity": "Sprint velocity", "epic-progress": "Epic progress", "sprint-progress": "Sprint progress"}
	if title, ok := names[kind]; ok {
		return title
	}
	return kind
}
