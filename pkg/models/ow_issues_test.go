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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func owTestData(t *testing.T) (*xorm.Session, *user.User, *OwProjectSettings) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	t.Cleanup(func() { _ = s.Close() })
	actor := &user.User{ID: 1}
	// Every test owns a unique project/key, so neither shared upstream fixtures nor another test's records are reused.
	p := &Project{Title: t.Name()}
	require.NoError(t, p.Create(s, actor))
	settings := OwDefaultSettings(p.ID, "T"+strconv.FormatInt(p.ID, 10))
	require.NoError(t, settings.Create(s, actor))
	return s, actor, settings
}

func TestOwIssueLifecycleAndPermissions(t *testing.T) {
	s, actor, settings := owTestData(t)
	issue := &OwIssue{ProjectID: settings.ProjectID, Summary: "Synthetic bug", Type: "Bug", AssigneeIDs: []int64{actor.ID}}
	ok, err := issue.CanCreate(s, actor)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, issue.Create(s, actor))
	require.Equal(t, settings.Key+"-1", issue.Key)
	other := &user.User{ID: 13}
	ok, err = issue.CanCreate(s, other)
	require.NoError(t, err)
	require.False(t, ok)
	ok, _, err = issue.CanRead(s, actor)
	require.NoError(t, err)
	require.True(t, ok)
	ok, _, err = issue.CanRead(s, other)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = issue.CanUpdate(s, other)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = issue.CanDelete(s, other)
	require.NoError(t, err)
	require.False(t, ok)
	for _, state := range []string{"selected", "in-progress", "in-review", "testing", "done"} {
		issue.State = state
		require.NoError(t, issue.Update(s, actor))
	}
	stored := &OwIssue{Key: issue.Key}
	require.NoError(t, stored.ReadOne(s, actor))
	require.True(t, stored.Task.Done)
	require.False(t, stored.CompletedAt.IsZero())
	require.False(t, stored.StartedAt.IsZero())
	metric, err := OwMetric(s, actor, "completed", OwFilter{ProjectID: settings.ProjectID})
	require.NoError(t, err)
	require.InDelta(t, 1.0, metric.Value, 0.000001)
	_, err = OwMetric(s, other, "completed", OwFilter{ProjectID: settings.ProjectID})
	require.Error(t, err)
	metrics, err := OwMetric(s, other, "completed", OwFilter{})
	require.NoError(t, err)
	require.Zero(t, metrics.Value)
	activities := []*OwActivity{}
	require.NoError(t, s.Where("task_id = ?", issue.TaskID).Find(&activities))
	require.Len(t, activities, 6)
	for _, entry := range activities {
		require.Equal(t, actor.ID, entry.ActorID)
	}
	ok, err = issue.CanDelete(s, actor)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, issue.Delete(s, actor))
	items, err := OwQuery(s, actor, OwFilter{ProjectID: settings.ProjectID})
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestOwFailedTransitionAndCustomValidation(t *testing.T) {
	s, actor, settings := owTestData(t)
	settings.Fields = []OwField{{ID: "approved", Name: "Approved", Type: "boolean"}, {ID: "impact", Name: "Impact", Type: "select", Options: []string{"low", "high"}}}
	settings.Transitions = append(settings.Transitions, OwTransition{From: "backlog", To: "done", Required: []string{"approved"}, AdminOnly: true})
	require.NoError(t, settings.Update(s, actor))
	i := &OwIssue{ProjectID: settings.ProjectID, Summary: "Validation test", Type: "Bug", Custom: map[string]any{"impact": "unknown"}}
	require.Error(t, i.Create(s, actor))
	i.Custom = map[string]any{"impact": "high"}
	require.NoError(t, i.Create(s, actor))
	i.State = "done"
	require.Error(t, i.Update(s, actor))
	stored := &OwIssue{Key: i.Key}
	require.NoError(t, stored.ReadOne(s, actor))
	require.Equal(t, "backlog", stored.State)
	require.False(t, stored.Task.Done)
	i.Custom["approved"] = true
	require.NoError(t, i.Update(s, actor))
	require.NoError(t, stored.ReadOne(s, actor))
	require.Equal(t, "done", stored.State)
}

func TestOwSprintAndDashboardIsolation(t *testing.T) {
	s, actor, settings := owTestData(t)
	sprint := &OwSprint{ProjectID: settings.ProjectID, Name: "Iteration 1"}
	require.NoError(t, sprint.Create(s, actor))
	issue := &OwIssue{ProjectID: settings.ProjectID, Summary: "Sprint work", Type: "Task", SprintID: sprint.ID}
	require.NoError(t, issue.Create(s, actor))
	started, err := OwSprintLifecycle(s, actor, sprint.ID, "start", 0)
	require.NoError(t, err)
	require.Equal(t, []int64{issue.TaskID}, started.Committed)
	closed, err := OwSprintLifecycle(s, actor, sprint.ID, "close", 0)
	require.NoError(t, err)
	require.Equal(t, "closed", closed.State)
	require.Empty(t, closed.Completed)
	require.NoError(t, issue.ReadOne(s, actor))
	require.Zero(t, issue.SprintID)
	dash := OwDashboardTemplate("Project Overview", settings.ProjectID)
	require.NoError(t, dash.Create(s, actor))
	other := &user.User{ID: 13}
	ok, _, err := dash.CanRead(s, other)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = dash.CanUpdate(s, other)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = dash.CanDelete(s, other)
	require.NoError(t, err)
	require.False(t, ok)
	data, err := OwDashboardMetrics(s, actor, dash.ID)
	require.NoError(t, err)
	require.Len(t, data.Widgets, 8)
	require.InDelta(t, 1.0, data.Widgets[0].Value, 0.000001)
	dash.Shares = []int64{other.ID}
	dash.Filter = OwFilter{}
	require.NoError(t, dash.Update(s, actor))
	data, err = OwDashboardMetrics(s, other, dash.ID)
	require.NoError(t, err)
	require.Zero(t, data.Widgets[0].Value)
	_, err = OwDashboardMetrics(s, other, dash.ID+999)
	require.Error(t, err)
}
