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
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/user"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type OwProjectInfo struct {
	Project    *Project           `json:"project" doc:"Accessible upstream project."`
	Settings   *OwProjectSettings `json:"settings" doc:"My Tickets configuration; null for an unconfigured legacy project."`
	Permission int                `json:"permission" doc:"Current project permission: 0 viewer, 1 contributor, 2 administrator."`
}
type OwCatalog struct {
	User        *user.User      `json:"user" doc:"The current upstream account; authorization is derived from this server-side identity."`
	Projects    []OwProjectInfo `json:"projects" doc:"Only the requesting user's accessible projects."`
	WidgetKinds []string        `json:"widget_kinds" doc:"Supported dashboard calculations."`
	Templates   []string        `json:"templates" doc:"Available dashboard templates."`
}

func OwGetCatalog(s *xorm.Session, a web.Auth) (*OwCatalog, error) {
	ids, err := owAccessibleProjects(s, a)
	if err != nil {
		return nil, err
	}
	projects := []*Project{}
	if err = s.In("id", ids).Asc("title").Find(&projects); err != nil {
		return nil, err
	}
	current, err := userForOw(s, a)
	if err != nil {
		return nil, err
	}
	result := &OwCatalog{User: current, Projects: []OwProjectInfo{}, WidgetKinds: OwWidgetKinds, Templates: []string{"Personal", "Project Overview", "Engineering", "Sprint Overview"}}
	for _, p := range projects {
		ok, permission, err := p.CanRead(s, a)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		settings := &OwProjectSettings{}
		found, err := s.ID(p.ID).Get(settings)
		if err != nil {
			return nil, err
		}
		if !found {
			settings = nil
		}
		result.Projects = append(result.Projects, OwProjectInfo{Project: p, Settings: settings, Permission: permission})
	}
	return result, nil
}

func OwIssueActivity(s *xorm.Session, a web.Auth, key string) ([]*OwActivity, error) {
	i := &OwIssue{Key: key}
	ok, _, err := i.CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	if err = i.ReadOne(s, a); err != nil {
		return nil, err
	}
	rows := []*OwActivity{}
	err = s.Where("task_id = ?", i.TaskID).Desc("id").Limit(200).Find(&rows)
	return rows, err
}

func OwDuplicateIssue(s *xorm.Session, a web.Auth, key string) (*OwIssue, error) {
	i := &OwIssue{Key: key}
	ok, _, err := i.CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	if err = i.ReadOne(s, a); err != nil {
		return nil, err
	}
	ok, err = i.CanCreate(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	i.Summary = "Copy of " + i.Summary
	i.State = ""
	i.Archived = false
	i.SprintID = 0
	i.Position = 0
	if err = i.Create(s, a); err != nil {
		return nil, err
	}
	return i, nil
}

type OwBulkChange struct {
	Keys   []string `json:"keys" doc:"At most 100 issue keys."`
	Action string   `json:"action" enum:"state,priority,sprint,archive,duplicate" doc:"Atomic bulk operation; all issues must authorize successfully."`
	Value  string   `json:"value" doc:"New state, priority or sprint ID."`
}

func OwBulk(s *xorm.Session, a web.Auth, change OwBulkChange) ([]string, error) {
	if len(change.Keys) == 0 || len(change.Keys) > 100 {
		return nil, owInvalid("Select one to 100 issues")
	}
	keys := []string{}
	for _, key := range change.Keys {
		if change.Action == "duplicate" {
			duplicate, err := OwDuplicateIssue(s, a, key)
			if err != nil {
				return nil, err
			}
			keys = append(keys, duplicate.Key)
			continue
		}
		i := &OwIssue{Key: key}
		ok, err := i.CanUpdate(s, a)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, owDenied()
		}
		if err = i.ReadOne(s, a); err != nil {
			return nil, err
		}
		switch change.Action {
		case "archive":
			if err = i.Delete(s, a); err != nil {
				return nil, err
			}
			keys = append(keys, key)
			continue
		case "state":
			i.State = change.Value
		case "priority":
			v, err := strconv.ParseInt(change.Value, 10, 64)
			if err != nil {
				return nil, owInvalid("Choose a numeric priority")
			}
			i.Priority = v
		case "sprint":
			v, err := strconv.ParseInt(change.Value, 10, 64)
			if err != nil {
				return nil, owInvalid("Choose a sprint ID")
			}
			i.SprintID = v
		default:
			return nil, owInvalid("Unsupported bulk action")
		}
		if err = i.Update(s, a); err != nil {
			return nil, err
		}
		keys = append(keys, i.Key)
	}
	return keys, nil
}

// OwMoveIssue preserves the upstream task and old identifier alias while resetting project-scoped metadata.
func OwMoveIssue(s *xorm.Session, a web.Auth, key string, project int64) (*OwIssue, error) {
	i := &OwIssue{Key: key}
	ok, err := i.CanUpdate(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	if err = i.ReadOne(s, a); err != nil {
		return nil, err
	}
	if i.ProjectID == project {
		return i, nil
	}
	ok, err = (&Project{ID: project}).CanWrite(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	settings, err := owSettings(s, project)
	if err != nil {
		return nil, err
	}
	previous := *i
	i.ProjectID = project
	i.State = settings.States[0].ID
	i.SprintID = 0
	i.EpicID = 0
	i.Custom = map[string]any{}
	if !slices.Contains(settings.Types, i.Type) {
		i.Type = settings.Types[0]
	}
	if err = i.validate(s, a, nil, settings); err != nil {
		return nil, err
	}
	t := i.Task
	t.ProjectID = project
	t.owManagedWrite = true
	t.Done = false
	// Assignees from the old project are removed; the destination's members can be assigned afterwards.
	t.Assignees = []*user.User{}
	if err = t.Update(s, a); err != nil {
		return nil, err
	}
	i.Key = settings.Key + "-" + strconv.FormatInt(t.Index, 10)
	i.Revision++
	i.AssigneeIDs = []int64{}
	i.CompletedAt = time.Time{}
	i.StartedAt = time.Time{}
	if _, err = s.ID(i.TaskID).AllCols().Update(i); err != nil {
		return nil, err
	}
	if err = owActivity(s, a, project, i.TaskID, "issue.moved", previous.Key+" moved to "+i.Key); err != nil {
		return nil, err
	}
	return i, nil
}
