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
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type OwError struct {
	Status  int
	Message string
}

func (e OwError) Error() string { return e.Message }
func (e OwError) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: e.Status, Code: 12001, Message: e.Message}
}
func owInvalid(message string) error { return OwError{Status: http.StatusBadRequest, Message: message} }
func owDenied() error {
	return OwError{Status: http.StatusForbidden, Message: "You do not have access to this project or operation"}
}

func OwDefaultSettings(project int64, key string) *OwProjectSettings {
	states := []OwState{}
	for _, item := range []struct{ id, name, category string }{{"backlog", "Backlog", "open"}, {"selected", "Selected", "open"}, {"in-progress", "In Progress", "open"}, {"in-review", "In Review", "open"}, {"testing", "Testing", "open"}, {"done", "Done", "done"}, {"blocked", "Blocked", "open"}, {"cancelled", "Cancelled", "cancelled"}} {
		states = append(states, OwState{ID: item.id, Name: item.name, Category: item.category, Working: slices.Contains([]string{"in-progress", "in-review", "testing"}, item.id)})
	}
	transitions := []OwTransition{}
	for i := 0; i < 5; i++ {
		transitions = append(transitions, OwTransition{From: states[i].ID, To: states[i+1].ID})
	}
	transitions = append(transitions, OwTransition{From: "*", To: "blocked"}, OwTransition{From: "*", To: "cancelled"}, OwTransition{From: "blocked", To: "in-progress"}, OwTransition{From: "done", To: "selected"}, OwTransition{From: "cancelled", To: "backlog"}, OwTransition{From: "in-review", To: "in-progress"}, OwTransition{From: "testing", To: "in-progress"}, OwTransition{From: "selected", To: "backlog"})
	return &OwProjectSettings{ProjectID: project, Key: key, Types: []string{"Epic", "Story", "Task", "Bug", "Improvement", "Subtask"}, States: states, Transitions: transitions, Fields: []OwField{}}
}

func owSettings(s *xorm.Session, project int64) (*OwProjectSettings, error) {
	settings := &OwProjectSettings{}
	found, err := s.ID(project).Get(settings)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, owInvalid("Set up My Tickets for this project first")
	}
	return settings, nil
}

func owAdmin(s *xorm.Session, project int64, a web.Auth) (bool, error) {
	p := &Project{ID: project}
	ok, permission, err := p.CanRead(s, a)
	return ok && permission == int(PermissionAdmin), err
}

func (p *OwProjectSettings) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	return (&Project{ID: p.ProjectID}).CanRead(s, a)
}
func (p *OwProjectSettings) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return owAdmin(s, p.ProjectID, a)
}
func (p *OwProjectSettings) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return p.CanCreate(s, a)
}

func (p *OwProjectSettings) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return owAdmin(s, p.ProjectID, a)
}
func (p *OwProjectSettings) Delete(_ *xorm.Session, _ web.Auth) error {
	return owInvalid("Project keys and workflows are retained with their issues; archive the project instead")
}
func (p *OwProjectSettings) ReadAll(s *xorm.Session, a web.Auth, _ string, page, perPage int) (any, int, int64, error) {
	catalog, err := OwGetCatalog(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	items := []*OwProjectSettings{}
	for _, project := range catalog.Projects {
		if project.Settings != nil {
			items = append(items, project.Settings)
		}
	}
	start, end := owPage(len(items), page, perPage)
	return items[start:end], end - start, int64(len(items)), nil
}
func (p *OwProjectSettings) ReadOne(s *xorm.Session, _ web.Auth) error {
	v, err := owSettings(s, p.ProjectID)
	if err == nil {
		*p = *v
	}
	return err
}

func (p *OwProjectSettings) validate() error {
	if !regexp.MustCompile(`^[A-Z][A-Z0-9]{0,9}$`).MatchString(p.Key) {
		return owInvalid("Use an uppercase project key of one to ten letters/numbers")
	}
	if len(p.States) == 0 || len(p.States) > 32 || len(p.Types) == 0 || len(p.Types) > 32 || len(p.Fields) > 64 || len(p.Transitions) > 512 {
		return owInvalid("Check the workflow, issue types and field limits")
	}
	states := map[string]bool{}
	for _, state := range p.States {
		if state.ID == "" || len(state.ID) > 80 || state.Name == "" || states[state.ID] || !slices.Contains([]string{"open", "done", "cancelled"}, state.Category) || state.Limit < 0 {
			return owInvalid("Workflow states need unique IDs, names and valid categories")
		}
		states[state.ID] = true
	}
	for _, transition := range p.Transitions {
		if (!states[transition.From] && transition.From != "*") || !states[transition.To] {
			return owInvalid("A transition references a missing state")
		}
	}
	seen := map[string]bool{}
	for _, kind := range p.Types {
		if kind == "" || len(kind) > 80 || seen[kind] {
			return owInvalid("Issue types must be unique and nonempty")
		}
		seen[kind] = true
	}
	seen = map[string]bool{}
	for _, field := range p.Fields {
		if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,39}$`).MatchString(field.ID) || field.Name == "" || seen[field.ID] || !slices.Contains([]string{"text", "number", "date", "boolean", "select", "multi-select"}, field.Type) {
			return owInvalid("Check custom field IDs and types")
		}
		seen[field.ID] = true
		if len(field.Options) > 100 {
			return owInvalid("A field supports at most 100 options")
		}
	}
	return nil
}

func (p *OwProjectSettings) Create(s *xorm.Session, _ web.Auth) error {
	if err := p.validate(); err != nil {
		return err
	}
	project, err := GetProjectSimpleByID(s, p.ProjectID)
	if err != nil {
		return err
	}
	if project.Identifier != "" && project.Identifier != p.Key {
		return owInvalid("The project already has a different key")
	}
	duplicate, err := s.Where("identifier = ? AND id != ?", p.Key, p.ProjectID).Exist(&Project{})
	if err != nil {
		return err
	}
	if duplicate {
		return owInvalid("That project key is already used")
	}
	if _, err = s.ID(p.ProjectID).Cols("identifier").Update(&Project{Identifier: p.Key}); err != nil {
		return err
	}
	_, err = s.Insert(p)
	return err
}

func (p *OwProjectSettings) Update(s *xorm.Session, a web.Auth) error {
	old, err := owSettings(s, p.ProjectID)
	if err != nil {
		return err
	}
	if p.Key != old.Key {
		return owInvalid("Project keys are permanent")
	}
	if err = p.validate(); err != nil {
		return err
	}
	issues := []*OwIssue{}
	if err = s.Where("project_id = ?", p.ProjectID).Find(&issues); err != nil {
		return err
	}
	for _, issue := range issues {
		if !slices.Contains(p.Types, issue.Type) {
			return owInvalid("Move issues to another type before removing an issue type")
		}
		oldState := owState(old, issue.State)
		newState := owState(p, issue.State)
		if newState == nil || oldState == nil || oldState.Category != newState.Category {
			return owInvalid("A used state cannot be removed or reclassified; transition its issues first")
		}
		for _, field := range old.Fields {
			if _, present := issue.Custom[field.ID]; present && !slices.ContainsFunc(p.Fields, func(f OwField) bool { return f.ID == field.ID && f.Type == field.Type }) {
				return owInvalid("A populated field cannot be removed or change type")
			}
		}
	}
	_, err = s.ID(p.ProjectID).AllCols().Update(p)
	if err != nil {
		return err
	}
	return owActivity(s, a, p.ProjectID, 0, "workflow.updated", "Project workflow and field configuration updated")
}

func OwCreateProject(s *xorm.Session, a web.Auth, title, key string) (*OwProjectSettings, error) {
	settings := OwDefaultSettings(0, strings.ToUpper(strings.TrimSpace(key)))
	if err := settings.validate(); err != nil {
		return nil, err
	}
	p := &Project{Title: strings.TrimSpace(title), Identifier: settings.Key}
	allowed, err := p.CanCreate(s, a)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, owDenied()
	}
	if err = p.Create(s, a); err != nil {
		return nil, err
	}
	settings.ProjectID = p.ID
	if err = settings.Create(s, a); err != nil {
		return nil, err
	}
	return settings, nil
}

func owState(settings *OwProjectSettings, id string) *OwState {
	for _, state := range settings.States {
		if state.ID == id {
			v := state
			return &v
		}
	}
	return nil
}
func owActivity(s *xorm.Session, a web.Auth, project, task int64, action, detail string) error {
	_, err := s.Insert(&OwActivity{ProjectID: project, TaskID: task, ActorID: a.GetID(), Action: action, Detail: detail})
	return err
}

func (i *OwIssue) stored(s *xorm.Session) (*OwIssue, error) {
	v := &OwIssue{}
	query := s
	if i.Key != "" {
		query = s.Where("`key` = ?", i.Key)
	} else {
		query = s.Where("task_id = ?", i.TaskID)
	}
	ok, err := query.Get(v)
	if err != nil {
		return nil, err
	}
	if !ok && i.Key != "" {
		cut := strings.LastIndex(i.Key, "-")
		if cut > 0 {
			number, parseErr := strconv.ParseInt(i.Key[cut+1:], 10, 64)
			project := &OwProjectSettings{}
			found, lookupErr := s.Where("`key` = ?", i.Key[:cut]).Get(project)
			if lookupErr != nil {
				return nil, lookupErr
			}
			if found && parseErr == nil {
				alias := &TaskIndexAlias{}
				found, lookupErr = s.Where("project_id = ? AND `index` = ?", project.ProjectID, number).Get(alias)
				if lookupErr != nil {
					return nil, lookupErr
				}
				if found {
					ok, err = s.ID(alias.TaskID).Get(v)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if !ok {
		return nil, OwError{Status: 404, Message: "This issue does not exist"}
	}
	return v, nil
}
func (i *OwIssue) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	v, err := i.stored(s)
	if err != nil {
		return false, 0, err
	}
	return (&Task{ID: v.TaskID}).CanRead(s, a)
}
func (i *OwIssue) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return (&Project{ID: i.ProjectID}).CanWrite(s, a)
}
func (i *OwIssue) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	v, err := i.stored(s)
	if err != nil {
		return false, err
	}
	return (&Task{ID: v.TaskID}).CanUpdate(s, a)
}
func (i *OwIssue) CanDelete(s *xorm.Session, a web.Auth) (bool, error) { return i.CanUpdate(s, a) }
func (i *OwIssue) hydrate(s *xorm.Session, a web.Auth) error {
	t := &Task{ID: i.TaskID}
	if err := t.ReadOne(s, a); err != nil {
		return err
	}
	i.Task = t
	i.Summary = t.Title
	i.Description = t.Description
	i.Priority = t.Priority
	i.DueDate = t.DueDate
	i.AssigneeIDs = []int64{}
	for _, u := range t.Assignees {
		i.AssigneeIDs = append(i.AssigneeIDs, u.ID)
	}
	return nil
}
func (i *OwIssue) ReadOne(s *xorm.Session, a web.Auth) error {
	v, err := i.stored(s)
	if err != nil {
		return err
	}
	*i = *v
	return i.hydrate(s, a)
}

func owEmpty(v any) bool {
	if v == nil {
		return true
	}
	if text, ok := v.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	if values, ok := v.([]any); ok {
		return len(values) == 0
	}
	return false
}
func owValidateCustom(settings *OwProjectSettings, values map[string]any) error {
	if len(values) > 64 {
		return owInvalid("Too many custom field values")
	}
	defs := map[string]OwField{}
	for _, field := range settings.Fields {
		defs[field.ID] = field
		if field.Required && owEmpty(values[field.ID]) {
			return owInvalid("Complete the required field: " + field.Name)
		}
	}
	for id, value := range values {
		field, ok := defs[id]
		if !ok {
			return owInvalid("Unknown custom field: " + id)
		}
		if value == nil {
			continue
		}
		valid := false
		switch field.Type {
		case "text":
			text, ok := value.(string)
			valid = ok && len(text) <= 16000
		case "number":
			_, valid = value.(float64)
		case "boolean":
			_, valid = value.(bool)
		case "date":
			text, ok := value.(string)
			_, err := time.Parse("2006-01-02", text)
			valid = ok && err == nil
		case "select":
			text, ok := value.(string)
			valid = ok && slices.Contains(field.Options, text)
		case "multi-select":
			values, ok := value.([]any)
			valid = ok
			for _, entry := range values {
				text, ok := entry.(string)
				if !ok || !slices.Contains(field.Options, text) {
					valid = false
				}
			}
		}
		if !valid {
			return owInvalid("Invalid value for " + field.Name)
		}
	}
	return nil
}

func (i *OwIssue) validate(s *xorm.Session, a web.Auth, old *OwIssue, settings *OwProjectSettings) error {
	if strings.TrimSpace(i.Summary) == "" || len(i.Summary) > 1000 || len(i.Description) > 1<<20 || i.Priority < 0 || i.Priority > 5 || i.Points < 0 || i.Estimate < 0 {
		return owInvalid("Check summary, priority and estimates")
	}
	if !slices.Contains(settings.Types, i.Type) || owState(settings, i.State) == nil {
		return owInvalid("Choose a configured issue type and workflow state")
	}
	if err := owValidateCustom(settings, i.Custom); err != nil {
		return err
	}
	if err := i.validatePlanning(s, old); err != nil {
		return err
	}
	if old != nil && old.State != i.State {
		if err := i.validateTransition(s, a, old, settings); err != nil {
			return err
		}
	}
	state := owState(settings, i.State)
	if state.Limit > 0 && (old == nil || old.State != i.State) {
		count, err := s.Where("project_id = ? AND state = ? AND archived = ?", i.ProjectID, i.State, false).Count(&OwIssue{})
		if err != nil {
			return err
		}
		if count >= int64(state.Limit) {
			return owInvalid("This workflow column has reached its work-in-progress limit")
		}
	}
	return nil
}

func (i *OwIssue) validatePlanning(s *xorm.Session, old *OwIssue) error {
	if i.SprintID != 0 {
		sp := &OwSprint{}
		ok, err := s.ID(i.SprintID).Get(sp)
		if err != nil {
			return err
		}
		if !ok || sp.ProjectID != i.ProjectID || sp.State == "closed" && (old == nil || old.SprintID != i.SprintID) {
			return owInvalid("Choose an open sprint in this project")
		}
	}
	if i.EpicID != 0 {
		epic := &OwIssue{}
		ok, err := s.ID(i.EpicID).Get(epic)
		if err != nil {
			return err
		}
		if !ok || epic.ProjectID != i.ProjectID || epic.Type != "Epic" || epic.TaskID == i.TaskID {
			return owInvalid("Choose another Epic in this project")
		}
	}
	return nil
}

func (i *OwIssue) Create(s *xorm.Session, a web.Auth) error {
	settings, err := owSettings(s, i.ProjectID)
	if err != nil {
		return err
	}
	if i.Type == "" {
		i.Type = "Task"
	}
	if i.State == "" {
		i.State = settings.States[0].ID
	}
	if i.Custom == nil {
		i.Custom = map[string]any{}
	}
	i.TaskID = 0
	i.Revision = 1
	i.Key = ""
	i.ReporterID = a.GetID()
	i.StartedAt = time.Time{}
	i.CompletedAt = time.Time{}
	i.Created = time.Time{}
	i.Updated = time.Time{}
	if err = i.validate(s, a, nil, settings); err != nil {
		return err
	}
	t := &Task{owManagedWrite: true, ProjectID: i.ProjectID, Title: strings.TrimSpace(i.Summary), Description: i.Description, Priority: i.Priority, DueDate: i.DueDate, Done: owState(settings, i.State).Category != "open"}
	for _, id := range i.AssigneeIDs {
		t.Assignees = append(t.Assignees, &user.User{ID: id})
	}
	if err = t.Create(s, a); err != nil {
		return err
	}
	i.TaskID = t.ID
	if owState(settings, i.State).Working {
		i.StartedAt = time.Now()
	}
	i.Key = settings.Key + "-" + strconv.FormatInt(t.Index, 10)
	if i.Position == 0 {
		i.Position = float64(t.Index) * 1024
	}
	if owState(settings, i.State).Category == "done" {
		i.CompletedAt = time.Now()
	}
	if _, err = s.Insert(i); err != nil {
		return err
	}
	return owRecordChange(s, a, nil, i)
}

func (i *OwIssue) Update(s *xorm.Session, a web.Auth) error {
	old, err := i.stored(s)
	if err != nil {
		return err
	}
	if i.ExpectedRevision > 0 && i.ExpectedRevision != old.Revision {
		return OwError{Status: 409, Message: "This issue changed while you were editing. Your draft is retained; review the latest issue before saving."}
	}
	if i.ExpectedRevision == 0 && i.ExpectedUpdated != "" {
		expected, parseErr := time.Parse(time.RFC3339Nano, i.ExpectedUpdated)
		if parseErr != nil {
			return owInvalid("The issue revision is invalid; reload the issue")
		}
		if !expected.Equal(old.Updated) {
			return OwError{Status: 409, Message: "This issue changed while you were editing. Your draft is retained; review the latest issue before saving."}
		}
	}

	if i.ProjectID != old.ProjectID {
		return owInvalid("Use the move action to change projects")
	}
	settings, err := owSettings(s, i.ProjectID)
	if err != nil {
		return err
	}
	i.TaskID = old.TaskID
	i.Revision = old.Revision + 1
	i.Key = old.Key
	i.ReporterID = old.ReporterID
	i.Created = old.Created
	i.StartedAt = old.StartedAt
	i.CompletedAt = old.CompletedAt
	if err = i.validate(s, a, old, settings); err != nil {
		return err
	}
	t := &Task{ID: i.TaskID}
	if err = t.ReadOne(s, a); err != nil {
		return err
	}
	t.Title = strings.TrimSpace(i.Summary)
	t.Description = i.Description
	t.Priority = i.Priority
	t.DueDate = i.DueDate
	t.Done = owState(settings, i.State).Category != "open"
	t.owManagedWrite = true
	if err = t.Update(s, a); err != nil {
		return err
	}
	assignees := []*user.User{}
	for _, id := range i.AssigneeIDs {
		assignees = append(assignees, &user.User{ID: id})
	}
	if err = t.updateTaskAssignees(s, assignees, a); err != nil {
		return err
	}
	if old.State != i.State {
		if (owState(settings, i.State).Working || i.State == "in-progress") && i.StartedAt.IsZero() {
			i.StartedAt = time.Now()
		}
		if owState(settings, i.State).Category == "done" {
			i.CompletedAt = time.Now()
		} else {
			i.CompletedAt = time.Time{}
		}
	}
	if _, err = s.ID(i.TaskID).AllCols().Update(i); err != nil {
		return err
	}
	i.Task = t
	return owRecordChange(s, a, old, i)
}
func (i *OwIssue) Delete(s *xorm.Session, a web.Auth) error {
	v, err := i.stored(s)
	if err != nil {
		return err
	}
	_, err = s.ID(v.TaskID).Cols("archived", "revision").Update(&OwIssue{Archived: true, Revision: v.Revision + 1})
	if err != nil {
		return err
	}
	return owActivity(s, a, v.ProjectID, v.TaskID, "issue.archived", v.Key+" archived; identifier preserved")
}

func owAccessibleProjects(s *xorm.Session, a web.Auth) ([]int64, error) {
	if _, ok := a.(*user.User); !ok {
		return nil, owDenied()
	}
	access, err := getProjectAccessForUser(s, a.GetID())
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(access.permissions))
	for id := range access.permissions {
		ids = append(ids, id)
	}
	return ids, nil
}

// OwQuery filters before pagination, so list totals and dashboard aggregates share exactly the same authorized records.
//
//nolint:gocyclo // Filtering, hydration and deterministic sorting share one authorized cohort.
func OwQuery(s *xorm.Session, a web.Auth, filter OwFilter) ([]*OwIssue, error) {
	if err := owValidateFilter(filter); err != nil {
		return nil, err
	}
	projects, err := owAccessibleProjects(s, a)
	if err != nil {
		return nil, err
	}
	if filter.ProjectID != 0 && !slices.Contains(projects, filter.ProjectID) {
		return nil, owDenied()
	}
	cond := builder.In("project_id", projects)
	conditions := []builder.Cond{cond}
	if filter.ProjectID != 0 {
		conditions = append(conditions, builder.Eq{"project_id": filter.ProjectID})
	}
	if filter.SprintID != 0 {
		conditions = append(conditions, builder.Eq{"sprint_id": filter.SprintID})
	}
	if filter.EpicID != 0 {
		conditions = append(conditions, builder.Eq{"epic_id": filter.EpicID})
	}
	if !filter.Archived {
		conditions = append(conditions, builder.Eq{"archived": false})
	}
	issues := []*OwIssue{}
	if err = s.Where(builder.And(conditions...)).Find(&issues); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(issues))
	for _, i := range issues {
		ids = append(ids, i.TaskID)
	}
	tasks := map[int64]*Task{}
	if err = s.In("id", ids).Find(&tasks); err != nil {
		return nil, err
	}
	links := []*TaskAssginee{}
	if err = s.In("task_id", ids).Find(&links); err != nil {
		return nil, err
	}
	assignees := map[int64][]int64{}
	for _, link := range links {
		assignees[link.TaskID] = append(assignees[link.TaskID], link.UserID)
	}
	labels := []*LabelTask{}
	if err = s.In("task_id", ids).Find(&labels); err != nil {
		return nil, err
	}
	labelIDs := map[int64][]int64{}
	for _, link := range labels {
		labelIDs[link.TaskID] = append(labelIDs[link.TaskID], link.LabelID)
	}
	result := []*OwIssue{}
	q := strings.ToLower(strings.TrimSpace(filter.Q))
	for _, i := range issues {
		t := tasks[i.TaskID]
		if t == nil {
			continue
		}
		i.Task = t
		i.Summary = t.Title
		i.Description = t.Description
		i.Priority = t.Priority
		i.DueDate = t.DueDate
		i.AssigneeIDs = assignees[i.TaskID]
		if i.AssigneeIDs == nil {
			i.AssigneeIDs = []int64{}
		}
		if q != "" && !strings.Contains(strings.ToLower(i.Key+" "+i.Summary+" "+i.Description), q) {
			continue
		}
		matched := 0
		for _, rule := range filter.Rules {
			if owMatch(i, rule, labelIDs[i.TaskID]) {
				matched++
			}
		}
		if len(filter.Rules) > 0 && ((filter.Match == "or" && matched == 0) || (filter.Match != "or" && matched != len(filter.Rules))) {
			continue
		}
		result = append(result, i)
	}
	sort.SliceStable(result, func(l, r int) bool {
		left, right := result[l], result[r]
		comparison := 0
		switch filter.Sort {
		case "summary":
			comparison = strings.Compare(strings.ToLower(left.Summary), strings.ToLower(right.Summary))
		case "priority":
			comparison = compareFloat(float64(left.Priority), float64(right.Priority))
		case "created":
			comparison = left.Created.Compare(right.Created)
		case "updated":
			comparison = left.Updated.Compare(right.Updated)
		case "due_date":
			comparison = left.DueDate.Compare(right.DueDate)
		case "position":
			comparison = compareFloat(left.Position, right.Position)
		default:
			comparison = strings.Compare(left.Key, right.Key)
		}
		if comparison == 0 {
			comparison = compareFloat(float64(left.TaskID), float64(right.TaskID))
		}
		if filter.Descending {
			return comparison > 0
		}
		return comparison < 0
	})
	return result, nil
}
func compareFloat(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

//nolint:gocyclo // Explicit typed predicates keep empty/date/number comparisons together.
func owMatch(i *OwIssue, rule OwRule, labels []int64) bool {
	var value any
	switch rule.Field {
	case "type":
		value = i.Type
	case "state":
		value = i.State
	case "priority":
		value = i.Priority
	case "reporter":
		value = i.ReporterID
	case "severity":
		value = i.Severity
	case "component":
		value = i.Component
	case "fix_version":
		value = i.FixVersion
	case "assignee":
		id, _ := strconv.ParseInt(rule.Value, 10, 64)
		yes := slices.Contains(i.AssigneeIDs, id)
		if rule.Operator == "empty" {
			return len(i.AssigneeIDs) == 0
		}
		if rule.Operator == "ne" {
			return !yes
		}
		return yes
	case "label":
		id, _ := strconv.ParseInt(rule.Value, 10, 64)
		yes := slices.Contains(labels, id)
		if rule.Operator == "ne" {
			return !yes
		}
		return yes
	case "due_date":
		if !i.DueDate.IsZero() {
			value = i.DueDate.Format(time.RFC3339)
		}
	case "created":
		value = i.Created.Format(time.RFC3339)
	case "updated":
		value = i.Updated.Format(time.RFC3339)
	default:
		if strings.HasPrefix(rule.Field, "custom.") {
			value = i.Custom[strings.TrimPrefix(rule.Field, "custom.")]
		}
	}
	if rule.Operator == "empty" {
		return owEmpty(value)
	}
	if value != nil && (rule.Field == "due_date" || rule.Field == "created" || rule.Field == "updated") && len(rule.Value) == 10 {
		instant, err := time.Parse(time.RFC3339, fmt.Sprint(value))
		if err != nil {
			return false
		}
		day := instant.In(config.GetTimeZone()).Format("2006-01-02")
		switch rule.Operator {
		case "eq":
			return day == rule.Value
		case "ne":
			return day != rule.Value
		case "before":
			return day < rule.Value
		case "after":
			return day > rule.Value
		}
	}
	text := ""
	if value != nil {
		text = fmt.Sprint(value)
	}
	switch rule.Operator {
	case "ne":
		return text != rule.Value
	case "contains":
		return strings.Contains(strings.ToLower(text), strings.ToLower(rule.Value))
	case "before", "after":
		if text == "" {
			return false
		}
		comparison := strings.Compare(text, rule.Value)
		if rule.Field == "due_date" || rule.Field == "created" || rule.Field == "updated" {
			left, leftErr := time.Parse(time.RFC3339, text)
			right, rightErr := time.Parse(time.RFC3339, rule.Value)
			if rightErr != nil {
				right, rightErr = time.ParseInLocation("2006-01-02", rule.Value, config.GetTimeZone())
			}
			if leftErr != nil || rightErr != nil {
				return false
			}
			comparison = left.Compare(right)
		} else if left, leftErr := strconv.ParseFloat(text, 64); leftErr == nil {
			right, rightErr := strconv.ParseFloat(rule.Value, 64)
			if rightErr != nil {
				return false
			}
			comparison = compareFloat(left, right)
		}
		if rule.Operator == "before" {
			return comparison < 0
		}
		return comparison > 0
	default:
		return text == rule.Value
	}
}

func (i *OwIssue) ReadAll(s *xorm.Session, a web.Auth, search string, page, perPage int) (any, int, int64, error) {
	issues, err := OwQuery(s, a, OwFilter{ProjectID: i.ProjectID, Q: search})
	if err != nil {
		return nil, 0, 0, err
	}
	total := int64(len(issues))
	if perPage <= 0 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}
	if page < 1 {
		page = 1
	}
	start := min((page-1)*perPage, len(issues))
	end := min(start+perPage, len(issues))
	return issues[start:end], end - start, total, nil
}

// owGuardCoreTask prevents older generic task routes from bypassing ticket invariants.
func owGuardCoreTask(s *xorm.Session, ids []int64) error {
	found, err := s.In("task_id", ids).Exist(&OwIssue{})
	if err != nil {
		return err
	}
	if found {
		return owInvalid("Use the My Tickets API to preserve issue history and workflow")
	}
	return nil
}
func owValidateFilter(f OwFilter) error {
	if len(f.Q) > 1000 || len(f.Rules) > 32 || !slices.Contains([]string{"", "and", "or"}, f.Match) {
		return owInvalid("Check search and filter limits")
	}
	for _, r := range f.Rules {
		if len(r.Value) > 16000 || !slices.Contains([]string{"eq", "ne", "contains", "before", "after", "empty"}, r.Operator) {
			return owInvalid("Unsupported filter comparison")
		}
		if !strings.HasPrefix(r.Field, "custom.") && !slices.Contains([]string{"type", "state", "priority", "assignee", "reporter", "label", "severity", "component", "fix_version", "due_date", "created", "updated"}, r.Field) {
			return owInvalid("Unsupported filter field")
		}
	}
	return nil
}

func owGuardCoreCreate(s *xorm.Session, project int64, tasks []*Task) error {
	configured, err := s.ID(project).Exist(&OwProjectSettings{})
	if err != nil {
		return err
	}
	if !configured {
		return nil
	}
	for _, task := range tasks {
		if !task.owManagedWrite {
			return owInvalid("Create tickets through the My Tickets API")
		}
	}
	return nil
}

func (i *OwIssue) validateTransition(s *xorm.Session, a web.Auth, old *OwIssue, settings *OwProjectSettings) error {
	allowed := false
	for _, transition := range settings.Transitions {
		if (transition.From == old.State || transition.From == "*") && transition.To == i.State {
			if transition.AdminOnly {
				ok, err := owAdmin(s, i.ProjectID, a)
				if err != nil {
					return err
				}
				if !ok {
					return owDenied()
				}
			}
			for _, required := range transition.Required {
				var value any
				switch required {
				case "assignee":
					if len(i.AssigneeIDs) > 0 {
						value = "assigned"
					}
				case "description":
					value = i.Description
				case "severity":
					value = i.Severity
				case "estimate":
					if i.Estimate > 0 {
						value = "estimated"
					}
				default:
					value = i.Custom[required]
				}
				if owEmpty(value) {
					return owInvalid("Complete " + required + " before this transition")
				}
			}
			allowed = true
			break
		}
	}
	if !allowed {
		return owInvalid("That workflow transition is not allowed")
	}

	return nil
}

func owBumpIssueRevision(s *xorm.Session, task int64) error {
	_, err := s.Where("task_id = ?", task).Incr("revision", 1).Update(&OwIssue{})
	return err
}
