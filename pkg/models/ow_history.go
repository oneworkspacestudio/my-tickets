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
	"slices"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func owRecordChange(s *xorm.Session, a web.Auth, previous, current *OwIssue) error {
	event := &OwActivity{ProjectID: current.ProjectID, TaskID: current.TaskID, ActorID: a.GetID(), Action: "issue.created", Detail: current.Key + ": " + current.Summary, State: current.State, SprintID: current.SprintID, Points: current.Points}
	if previous != nil {
		event.Action = "issue.updated"
		event.PreviousState = previous.State
		event.PreviousSprintID = previous.SprintID
		event.Detail = current.Key + " updated"
		if previous.State != current.State {
			event.Detail = previous.State + " → " + current.State
		}
	}
	_, err := s.Insert(event)
	return err
}

// Historical charts replay structured server events, rather than estimating past states from today's totals.
//
//nolint:gocyclo // Bounded history replay handles sprint scope and each chart series consistently.
func OwHistoricalMetric(s *xorm.Session, a web.Auth, kind string, filter OwFilter, issues []*OwIssue) (OwWidgetData, error) {
	result := OwWidgetData{Kind: kind, Points: []OwPoint{}, Issues: []*OwIssue{}, Activities: []*OwActivity{}}
	start := time.Now().In(config.GetTimeZone()).AddDate(0, 0, -29)
	end := time.Now().In(config.GetTimeZone())
	scope := []int64{}
	var sprint *OwSprint
	if kind == "burndown" {
		if filter.SprintID == 0 {
			return result, nil
		}
		sprint = &OwSprint{ID: filter.SprintID}
		allowed, _, err := sprint.CanRead(s, a)
		if err != nil {
			return result, err
		}
		if !allowed {
			return result, owDenied()
		}
		if err = sprint.ReadOne(s, a); err != nil {
			return result, err
		}
		if sprint.Started.IsZero() {
			return result, nil
		}
		start = sprint.Started.In(config.GetTimeZone())
		if !sprint.Closed.IsZero() {
			end = sprint.Closed.In(config.GetTimeZone())
		}
		scope = append(scope, sprint.Committed...)
		all, err := OwQuery(s, a, OwFilter{ProjectID: sprint.ProjectID, Archived: true})
		if err != nil {
			return result, err
		}
		issues = all
	}
	if end.Sub(start) > 366*24*time.Hour {
		start = end.AddDate(0, 0, -366)
	}
	ids := make([]int64, 0, len(issues))
	settings := map[int64]*OwProjectSettings{}
	for _, issue := range issues {
		ids = append(ids, issue.TaskID)
		if settings[issue.ProjectID] == nil {
			value, err := owSettings(s, issue.ProjectID)
			if err != nil {
				return result, err
			}
			settings[issue.ProjectID] = value
		}
	}
	events := []*OwActivity{}
	if err := s.In("task_id", ids).Asc("at", "id").Find(&events); err != nil {
		return result, err
	}
	history := map[int64][]*OwActivity{}
	for _, event := range events {
		history[event.TaskID] = append(history[event.TaskID], event)
	}
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		cutoff := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, config.GetTimeZone()).AddDate(0, 0, 1)
		open, done := 0.0, 0.0
		for _, issue := range issues {
			if issue.Created.After(cutoff) {
				continue
			}
			state := ""
			membership := false
			if sprint != nil {
				membership = slices.Contains(scope, issue.TaskID)
			}
			for _, event := range history[issue.TaskID] {
				if !event.At.Before(cutoff) {
					break
				}
				if event.State != "" {
					state = event.State
				}
				if sprint != nil && event.At.After(sprint.Started) {
					if event.SprintID == sprint.ID {
						membership = true
					} else if event.PreviousSprintID == sprint.ID && event.SprintID != sprint.ID {
						membership = false
					}
				}
			}
			if state == "" {
				continue
			}
			if sprint != nil && !membership {
				continue
			}
			definition := owState(settings[issue.ProjectID], state)
			if definition == nil {
				continue
			}
			switch definition.Category {
			case "done":
				done++
			case "open":
				open++
			}
		}
		point := OwPoint{Label: date.Format("2006-01-02"), Value: open, Secondary: done}
		if kind == "burndown" {
			point.Secondary = open + done
		}
		result.Points = append(result.Points, point)
	}
	return result, nil
}
