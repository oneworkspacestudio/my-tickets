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
	"strings"
	"time"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func (v *OwSprint) stored(s *xorm.Session) (*OwSprint, error) {
	old := &OwSprint{}
	ok, err := s.ID(v.ID).Get(old)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, OwError{Status: 404, Message: "Sprint not found"}
	}
	return old, nil
}
func (v *OwSprint) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	old, err := v.stored(s)
	if err != nil {
		return false, 0, err
	}
	return (&Project{ID: old.ProjectID}).CanRead(s, a)
}
func (v *OwSprint) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return (&Project{ID: v.ProjectID}).CanWrite(s, a)
}
func (v *OwSprint) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	old, err := v.stored(s)
	if err != nil {
		return false, err
	}
	return (&Project{ID: old.ProjectID}).CanWrite(s, a)
}
func (v *OwSprint) CanDelete(s *xorm.Session, a web.Auth) (bool, error) { return v.CanUpdate(s, a) }
func (v *OwSprint) ReadOne(s *xorm.Session, _ web.Auth) error {
	old, err := v.stored(s)
	if err == nil {
		*v = *old
	}
	return err
}
func (v *OwSprint) validate() error {
	if strings.TrimSpace(v.Name) == "" || len(v.Name) > 250 || (!v.Start.IsZero() && !v.End.IsZero() && !v.End.After(v.Start)) {
		return owInvalid("A sprint needs a name and an end after its start")
	}
	return nil
}
func (v *OwSprint) Create(s *xorm.Session, a web.Auth) error {
	if err := v.validate(); err != nil {
		return err
	}
	if _, err := owSettings(s, v.ProjectID); err != nil {
		return err
	}
	v.ID = 0
	v.State = "planned"
	v.Started = time.Time{}
	v.Closed = time.Time{}
	v.Committed = []int64{}
	v.Completed = []int64{}
	_, err := s.Insert(v)
	if err != nil {
		return err
	}
	return owActivity(s, a, v.ProjectID, 0, "sprint.created", v.Name)
}
func (v *OwSprint) Update(s *xorm.Session, a web.Auth) error {
	old, err := v.stored(s)
	if err != nil {
		return err
	}
	if v.ProjectID != old.ProjectID {
		return owInvalid("A sprint cannot change projects")
	}
	if err = v.validate(); err != nil {
		return err
	}
	if old.State == "closed" {
		return owInvalid("Closed sprints are historical records")
	}
	v.State = old.State
	v.Started = old.Started
	v.Closed = old.Closed
	v.Committed = old.Committed
	v.Completed = old.Completed
	_, err = s.ID(v.ID).AllCols().Update(v)
	if err != nil {
		return err
	}
	return owActivity(s, a, v.ProjectID, 0, "sprint.updated", v.Name)
}
func (v *OwSprint) Delete(s *xorm.Session, a web.Auth) error {
	old, err := v.stored(s)
	if err != nil {
		return err
	}
	if old.State != "planned" {
		return owInvalid("Only planned sprints can be deleted")
	}
	if _, err = s.Where("sprint_id = ?", old.ID).Cols("sprint_id").Update(&OwIssue{SprintID: 0}); err != nil {
		return err
	}
	_, err = s.ID(old.ID).Delete(&OwSprint{})
	if err != nil {
		return err
	}
	return owActivity(s, a, old.ProjectID, 0, "sprint.deleted", old.Name)
}
func (v *OwSprint) ReadAll(s *xorm.Session, a web.Auth, _ string, page, perPage int) (any, int, int64, error) {
	ids, err := owAccessibleProjects(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	query := s.In("project_id", ids)
	if v.ProjectID != 0 {
		query = query.And("project_id = ?", v.ProjectID)
	}
	items := []*OwSprint{}
	if err = query.Desc("id").Find(&items); err != nil {
		return nil, 0, 0, err
	}
	total := len(items)
	start, end := owPage(total, page, perPage)
	return items[start:end], end - start, int64(total), nil
}

func OwSprintLifecycle(s *xorm.Session, a web.Auth, id int64, action string, nextID int64) (*OwSprint, error) {
	v := &OwSprint{ID: id}
	allowed, err := v.CanUpdate(s, a)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, owDenied()
	}
	if err = v.ReadOne(s, a); err != nil {
		return nil, err
	}
	issues, err := OwQuery(s, a, OwFilter{ProjectID: v.ProjectID, SprintID: v.ID})
	if err != nil {
		return nil, err
	}
	settings, err := owSettings(s, v.ProjectID)
	if err != nil {
		return nil, err
	}
	switch action {
	case "start":
		if v.State != "planned" {
			return nil, owInvalid("Only a planned sprint can start")
		}
		active, err := s.Where("project_id = ? AND state = ?", v.ProjectID, "active").Exist(&OwSprint{})
		if err != nil {
			return nil, err
		}
		if active {
			return nil, owInvalid("Close the active sprint before starting another")
		}
		v.State = "active"
		v.Started = time.Now()
		v.Committed = []int64{}
		for _, issue := range issues {
			v.Committed = append(v.Committed, issue.TaskID)
		}
	case "close":
		if v.State != "active" {
			return nil, owInvalid("Only an active sprint can close")
		}
		if nextID != 0 {
			next := &OwSprint{ID: nextID}
			if err = next.ReadOne(s, a); err != nil {
				return nil, err
			}
			if next.ID == id || next.ProjectID != v.ProjectID || next.State == "closed" {
				return nil, owInvalid("Choose another open sprint in this project")
			}
		}
		v.State = "closed"
		v.Closed = time.Now()
		v.Completed = []int64{}
		for _, issue := range issues {
			state := owState(settings, issue.State)
			if state != nil && state.Category == "done" {
				v.Completed = append(v.Completed, issue.TaskID)
				continue
			}
			if _, err = s.ID(issue.TaskID).Cols("sprint_id").Update(&OwIssue{SprintID: nextID}); err != nil {
				return nil, err
			}
			if err = owActivity(s, a, issue.ProjectID, issue.TaskID, "sprint.carryover", issue.Key+" returned to backlog or next sprint"); err != nil {
				return nil, err
			}
		}
	default:
		return nil, owInvalid("Choose start or close")
	}
	if _, err = s.ID(v.ID).AllCols().Update(v); err != nil {
		return nil, err
	}
	if err = owActivity(s, a, v.ProjectID, 0, "sprint."+action, v.Name); err != nil {
		return nil, err
	}
	return v, nil
}

func owPage(total, page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}
	start := min(total, (page-1)*perPage)
	return start, min(total, start+perPage)
}

var OwWidgetKinds = []string{"open", "completed", "overdue", "unassigned", "bugs-by-priority", "by-status", "by-assignee", "workload", "created-resolved", "throughput", "sprint-progress", "recent", "activity", "cycle-time", "aging", "cumulative-flow", "burndown", "velocity", "epic-progress"}

func (v *OwDashboard) stored(s *xorm.Session) (*OwDashboard, error) {
	old := &OwDashboard{}
	ok, err := s.ID(v.ID).Get(old)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, OwError{Status: 404, Message: "Dashboard not found"}
	}
	return old, nil
}
func (v *OwDashboard) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	old, err := v.stored(s)
	if err != nil {
		return false, 0, err
	}
	if old.OwnerID == a.GetID() {
		return true, 1, nil
	}
	if old.Scope == "global" || slices.Contains(old.Shares, a.GetID()) {
		return true, 0, nil
	}
	if old.Scope == "project" {
		ok, _, err := (&Project{ID: old.ProjectID}).CanRead(s, a)
		return ok, 0, err
	}
	return false, 0, nil
}
func (v *OwDashboard) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if v.Scope == "project" {
		return owAdmin(s, v.ProjectID, a)
	}
	if v.Scope == "global" {
		return owSystemAdmin(s, a)
	}
	_, err := owAccessibleProjects(s, a)
	return err == nil, err
}
func (v *OwDashboard) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	old, err := v.stored(s)
	return err == nil && old.OwnerID == a.GetID(), err
}
func (v *OwDashboard) CanDelete(s *xorm.Session, a web.Auth) (bool, error) { return v.CanUpdate(s, a) }
func owSystemAdmin(s *xorm.Session, a web.Auth) (bool, error) {
	u, err := userForOw(s, a)
	return err == nil && u.IsAdmin, err
}
func (v *OwDashboard) ReadOne(s *xorm.Session, _ web.Auth) error {
	old, err := v.stored(s)
	if err == nil {
		*v = *old
	}
	return err
}
func (v *OwDashboard) validate() error {
	if strings.TrimSpace(v.Name) == "" || len(v.Name) > 250 || !slices.Contains([]string{"private", "project", "global"}, v.Scope) || len(v.Shares) > 100 || len(v.Widgets) > 40 {
		return owInvalid("Check dashboard name, scope and widget limits")
	}
	if v.Scope == "project" && v.ProjectID == 0 {
		return owInvalid("Choose a project for a project dashboard")
	}
	seen := map[string]bool{}
	for _, widget := range v.Widgets {
		if widget.ID == "" || len(widget.ID) > 80 || seen[widget.ID] || !slices.Contains(OwWidgetKinds, widget.Kind) || !slices.Contains([]string{"", "metric", "line", "bar", "donut", "table"}, widget.Chart) || widget.W < 1 || widget.W > 12 || widget.H < 1 || widget.H > 12 || widget.X < 0 || widget.X+widget.W > 12 || widget.Y < 0 || widget.Y > 1000 {
			return owInvalid("Check widget IDs, kinds and grid dimensions")
		}
		for _, previous := range v.Widgets {
			if seen[previous.ID] && widget.X < previous.X+previous.W && widget.X+widget.W > previous.X && widget.Y < previous.Y+previous.H && widget.Y+widget.H > previous.Y {
				return owInvalid("Dashboard widgets cannot overlap")
			}
		}
		seen[widget.ID] = true
	}
	return nil
}
func (v *OwDashboard) Create(s *xorm.Session, a web.Auth) error {
	if v.Scope == "" {
		v.Scope = "private"
	}
	if err := v.validate(); err != nil {
		return err
	}
	v.ID = 0
	v.OwnerID = a.GetID()
	_, err := s.Insert(v)
	return err
}
func (v *OwDashboard) Update(s *xorm.Session, a web.Auth) error {
	old, err := v.stored(s)
	if err != nil {
		return err
	}
	v.OwnerID = old.OwnerID
	if err = v.validate(); err != nil {
		return err
	}
	allowed, err := v.CanCreate(s, a)
	if err != nil {
		return err
	}
	if !allowed {
		return owDenied()
	}
	_, err = s.ID(v.ID).AllCols().Update(v)
	return err
}
func (v *OwDashboard) Delete(s *xorm.Session, _ web.Auth) error {
	if _, err := s.Where("dashboard_id = ?", v.ID).Delete(&OwDashboardLayout{}); err != nil {
		return err
	}
	_, err := s.ID(v.ID).Delete(&OwDashboard{})
	return err
}
func (v *OwDashboard) ReadAll(s *xorm.Session, a web.Auth, _ string, page, perPage int) (any, int, int64, error) {
	all := []*OwDashboard{}
	if err := s.Desc("updated").Find(&all); err != nil {
		return nil, 0, 0, err
	}
	items := []*OwDashboard{}
	for _, item := range all {
		ok, _, err := item.CanRead(s, a)
		if err != nil {
			return nil, 0, 0, err
		}
		if ok {
			items = append(items, item)
		}
	}
	start, end := owPage(len(items), page, perPage)
	return items[start:end], end - start, int64(len(items)), nil
}

func OwSaveLayout(s *xorm.Session, a web.Auth, id int64, widgets []OwWidget) error {
	d := &OwDashboard{ID: id}
	ok, _, err := d.CanRead(s, a)
	if err != nil {
		return err
	}
	if !ok {
		return owDenied()
	}
	if err = d.ReadOne(s, a); err != nil {
		return err
	}
	if len(widgets) != len(d.Widgets) {
		return owInvalid("Personal layouts must retain every shared widget")
	}
	for _, widget := range widgets {
		if !slices.ContainsFunc(d.Widgets, func(original OwWidget) bool { return original.ID == widget.ID && original.Kind == widget.Kind }) {
			return owInvalid("Only widget positions and dimensions can be personalized")
		}
	}
	candidate := *d
	candidate.Widgets = widgets
	if err = candidate.validate(); err != nil {
		return err
	}
	if _, err = s.Where("dashboard_id = ? AND user_id = ?", id, a.GetID()).Delete(&OwDashboardLayout{}); err != nil {
		return err
	}
	_, err = s.Insert(&OwDashboardLayout{DashboardID: id, UserID: a.GetID(), Widgets: widgets})
	return err
}

func (v *OwSavedView) stored(s *xorm.Session) (*OwSavedView, error) {
	old := &OwSavedView{}
	ok, err := s.ID(v.ID).Get(old)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, OwError{Status: 404, Message: "Saved view not found"}
	}
	return old, nil
}
func (v *OwSavedView) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	old, err := v.stored(s)
	if err != nil {
		return false, 0, err
	}
	if old.OwnerID == a.GetID() {
		return true, 1, nil
	}
	return old.Shared, 0, nil
}
func (v *OwSavedView) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	_, err := owAccessibleProjects(s, a)
	return err == nil, err
}
func (v *OwSavedView) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	old, err := v.stored(s)
	return err == nil && old.OwnerID == a.GetID(), err
}
func (v *OwSavedView) CanDelete(s *xorm.Session, a web.Auth) (bool, error) { return v.CanUpdate(s, a) }
func (v *OwSavedView) ReadOne(s *xorm.Session, _ web.Auth) error {
	old, err := v.stored(s)
	if err == nil {
		*v = *old
	}
	return err
}
func (v *OwSavedView) Create(s *xorm.Session, a web.Auth) error {
	if v.Name == "" || len(v.Name) > 250 {
		return owInvalid("Give the saved view a name")
	}
	v.ID = 0
	v.OwnerID = a.GetID()
	_, err := s.Insert(v)
	return err
}
func (v *OwSavedView) Update(s *xorm.Session, _ web.Auth) error {
	old, err := v.stored(s)
	if err != nil {
		return err
	}
	v.OwnerID = old.OwnerID
	if v.Name == "" || len(v.Name) > 250 {
		return owInvalid("Give the saved view a name")
	}
	_, err = s.ID(v.ID).AllCols().Update(v)
	return err
}
func (v *OwSavedView) Delete(s *xorm.Session, _ web.Auth) error {
	_, err := s.ID(v.ID).Delete(&OwSavedView{})
	return err
}
func (v *OwSavedView) ReadAll(s *xorm.Session, a web.Auth, _ string, page, perPage int) (any, int, int64, error) {
	items := []*OwSavedView{}
	if err := s.Where("owner_id = ? OR shared = ?", a.GetID(), true).Desc("id").Find(&items); err != nil {
		return nil, 0, 0, err
	}
	start, end := owPage(len(items), page, perPage)
	return items[start:end], end - start, int64(len(items)), nil
}
