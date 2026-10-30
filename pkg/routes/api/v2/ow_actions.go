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

package apiv2

import (
	"context"
	"net/http"
	"slices"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

func init() { AddRouteRegistrar(RegisterOwActions) }
func RegisterOwActions(api huma.API) {
	Register(api, huma.Operation{OperationID: "tickets-issue-move", Summary: "Move an issue to another project", Description: "Authorizes both projects and retains a resolvable old-key alias.", Method: http.MethodPost, Path: "/tickets/issues/{key}/move", Tags: []string{"tickets"}}, owMove)
	Register(api, huma.Operation{OperationID: "tickets-catalog", Summary: "List accessible ticket projects", Description: "Includes role and workflow configuration only for projects the caller can read.", Method: http.MethodGet, Path: "/tickets/catalog", Tags: []string{"tickets"}}, owCatalog)
	Register(api, huma.Operation{OperationID: "tickets-query", Summary: "Search and filter issues", Description: "Combines AND/OR predicates before pagination. Totals never include hidden projects.", Method: http.MethodPost, Path: "/tickets/query", Tags: []string{"tickets"}}, owQuery)
	Register(api, huma.Operation{OperationID: "tickets-project-create", Summary: "Create a ticket project", Description: "Atomically creates the upstream project and its immutable key/workflow settings.", Method: http.MethodPost, Path: "/tickets/projects", Tags: []string{"tickets"}}, owProjectCreate)
	Register(api, huma.Operation{OperationID: "tickets-project-settings-read", Summary: "Read project workflow", Description: "Requires project read permission.", Method: http.MethodGet, Path: "/tickets/projects/{id}", Tags: []string{"tickets"}}, owProjectSettingsRead)
	Register(api, huma.Operation{OperationID: "tickets-project-settings-update", Summary: "Configure project workflow", Description: "Requires project administration; rejects key changes and orphaning existing field values or states.", Method: http.MethodPut, Path: "/tickets/projects/{id}", Tags: []string{"tickets"}}, owProjectSettingsUpdate)
	Register(api, huma.Operation{OperationID: "tickets-issue-activity", Summary: "Read issue audit history", Description: "Authorizes against the current issue project before returning server-recorded actors and events.", Method: http.MethodGet, Path: "/tickets/issues/{key}/activity", Tags: []string{"tickets"}}, owActivity)
	Register(api, huma.Operation{OperationID: "tickets-issue-duplicate", Summary: "Duplicate an issue", Description: "Requires source read and project create permission; creates a new permanent identifier and audit record.", Method: http.MethodPost, Path: "/tickets/issues/{key}/duplicate", Tags: []string{"tickets"}}, owDuplicate)
	Register(api, huma.Operation{OperationID: "tickets-bulk", Summary: "Apply an atomic bulk change", Description: "All selected issues must authorize and validate before commit. A failure rolls back every change.", Method: http.MethodPost, Path: "/tickets/bulk", Tags: []string{"tickets"}}, owBulk)
	Register(api, huma.Operation{OperationID: "tickets-sprint-start", Summary: "Start a sprint", Description: "Snapshots starting scope; only one sprint may be active in a project.", Method: http.MethodPost, Path: "/tickets/sprints/{id}/start", Tags: []string{"tickets"}}, owSprintStart)
	Register(api, huma.Operation{OperationID: "tickets-sprint-close", Summary: "Close a sprint", Description: "Records completed scope and sends unfinished work to backlog or a chosen next sprint.", Method: http.MethodPost, Path: "/tickets/sprints/{id}/close", Tags: []string{"tickets"}}, owSprintClose)
	Register(api, huma.Operation{OperationID: "tickets-dashboard-data", Summary: "Compute dashboard widgets", Description: "Every widget is calculated with the current viewer's project permissions. Definitions and personal layouts persist independently.", Method: http.MethodGet, Path: "/tickets/dashboards/{id}/data", Tags: []string{"tickets"}}, owDashboardData)
	Register(api, huma.Operation{OperationID: "tickets-dashboard-layout", Summary: "Save a personal widget layout", Description: "Viewers can personalize positions and dimensions without changing shared definitions or filters.", Method: http.MethodPost, Path: "/tickets/dashboards/{id}/layout", Tags: []string{"tickets"}}, owDashboardLayout)
	Register(api, huma.Operation{OperationID: "tickets-dashboard-template", Summary: "Create a dashboard from a template", Description: "Creates a private dashboard from actual authorized issue data; Personal defaults to assigned work.", Method: http.MethodPost, Path: "/tickets/dashboard-templates/{name}", Tags: []string{"tickets"}}, owDashboardTemplate)
	Register(api, huma.Operation{OperationID: "tickets-report", Summary: "Calculate a report", Description: "Uses the same permission-scoped filter engine as issues and dashboards; historical charts replay audit events.", Method: http.MethodPost, Path: "/tickets/reports/{kind}", Tags: []string{"tickets"}}, owReport)
}

func owTransaction(ctx context.Context, fn func(*xorm.Session, web.Auth) error) error {
	a, err := authFromCtx(ctx)
	if err != nil {
		return err
	}
	s := db.NewSession()
	defer s.Close()
	if err = fn(s, a); err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		events.CleanupPending(s)
		return translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return nil
}
func owCatalog(ctx context.Context, _ *struct{}) (*singleBody[models.OwCatalog], error) {
	var value *models.OwCatalog
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		value, err = models.OwGetCatalog(s, a)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwCatalog]{Body: value}, nil
}

func owQuery(ctx context.Context, in *struct {
	Body struct {
		Filter  models.OwFilter `json:"filter" doc:"Compound authorized query."`
		Page    int             `json:"page" doc:"One-based page."`
		PerPage int             `json:"per_page" doc:"Page size, at most 200."`
	}
}) (*owIssueListBody, error) {
	var rows []*models.OwIssue
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		rows, err = models.OwQuery(s, a, in.Body.Filter)
		return err
	})
	if err != nil {
		return nil, err
	}
	page := max(1, in.Body.Page)
	perPage := in.Body.PerPage
	if perPage <= 0 {
		perPage = 50
	}
	perPage = min(perPage, 200)
	start := min(len(rows), (page-1)*perPage)
	end := min(len(rows), start+perPage)
	return &owIssueListBody{Body: NewPaginated(rows[start:end], int64(len(rows)), page, perPage)}, nil
}

func owProjectCreate(ctx context.Context, in *struct {
	Body struct {
		Title string `json:"title" doc:"Project title."`
		Key   string `json:"key" doc:"Unique immutable uppercase key."`
	}
}) (*singleBody[models.OwProjectSettings], error) {
	var value *models.OwProjectSettings
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		value, err = models.OwCreateProject(s, a, in.Body.Title, in.Body.Key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwProjectSettings]{Body: value}, nil
}

type owProjectSettingsBody struct {
	models.OwProjectSettings
	MaxPermission int `json:"max_permission" readOnly:"true" doc:"Current project permission."`
}

func owProjectSettingsRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*singleBody[owProjectSettingsBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwProjectSettings{ProjectID: in.ID}
	permission, err := handler.DoReadOne(ctx, m, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[owProjectSettingsBody]{Body: &owProjectSettingsBody{OwProjectSettings: *m, MaxPermission: permission}}, nil
}
func owProjectSettingsUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body owProjectSettingsBody
}) (*singleBody[models.OwProjectSettings], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ID
	if err = handler.DoUpdate(ctx, &in.Body.OwProjectSettings, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwProjectSettings]{Body: &in.Body.OwProjectSettings}, nil
}

type owActivityBody struct{ Body []*models.OwActivity }

func owActivity(ctx context.Context, in *struct {
	Key string `path:"key"`
}) (*owActivityBody, error) {
	rows := []*models.OwActivity{}
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		rows, err = models.OwIssueActivity(s, a, in.Key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &owActivityBody{Body: rows}, nil
}
func owDuplicate(ctx context.Context, in *struct {
	Key string `path:"key"`
}) (*singleBody[models.OwIssue], error) {
	var value *models.OwIssue
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		value, err = models.OwDuplicateIssue(s, a, in.Key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwIssue]{Body: value}, nil
}
func owBulk(ctx context.Context, in *struct{ Body models.OwBulkChange }) (*struct{ Body []string }, error) {
	var keys []string
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		keys, err = models.OwBulk(s, a, in.Body)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &struct{ Body []string }{Body: keys}, nil
}

type owSprintLifecycleInput struct {
	ID   int64 `path:"id"`
	Body struct {
		NextSprintID int64 `json:"next_sprint_id" doc:"Zero returns unfinished issues to backlog."`
	}
}

func owSprintStart(ctx context.Context, in *owSprintLifecycleInput) (*singleBody[models.OwSprint], error) {
	return owSprintAction(ctx, in, "start")
}
func owSprintClose(ctx context.Context, in *owSprintLifecycleInput) (*singleBody[models.OwSprint], error) {
	return owSprintAction(ctx, in, "close")
}
func owSprintAction(ctx context.Context, in *owSprintLifecycleInput, action string) (*singleBody[models.OwSprint], error) {
	var value *models.OwSprint
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		value, err = models.OwSprintLifecycle(s, a, in.ID, action, in.Body.NextSprintID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwSprint]{Body: value}, nil
}
func owDashboardData(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*singleBody[models.OwDashboardData], error) {
	var value *models.OwDashboardData
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		value, err = models.OwDashboardMetrics(s, a, in.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwDashboardData]{Body: value}, nil
}
func owDashboardLayout(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body struct {
		Widgets []models.OwWidget `json:"widgets" doc:"Only layout dimensions and positions are changed."`
	}
}) (*emptyBody, error) {
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error { return models.OwSaveLayout(s, a, in.ID, in.Body.Widgets) })
	if err != nil {
		return nil, err
	}
	return &emptyBody{}, nil
}
func owDashboardTemplate(ctx context.Context, in *struct {
	Name string `path:"name" enum:"Personal,Project Overview,Engineering,Sprint Overview"`
	Body struct {
		ProjectID int64 `json:"project_id" doc:"Optional project scope."`
	}
}) (*singleBody[models.OwDashboard], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	d := models.OwDashboardTemplate(in.Name, in.Body.ProjectID)
	if in.Name == "Personal" {
		d.Filter.Rules = []models.OwRule{{Field: "assignee", Operator: "eq", Value: strconv.FormatInt(a.GetID(), 10)}}
	}
	if err = handler.DoCreate(ctx, d, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwDashboard]{Body: d}, nil
}
func owReport(ctx context.Context, in *struct {
	Kind string `path:"kind"`
	Body models.OwFilter
}) (*singleBody[models.OwWidgetData], error) {
	if !slices.Contains(models.OwWidgetKinds, in.Kind) {
		return nil, huma.Error400BadRequest("Choose an available report")
	}
	var value models.OwWidgetData
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		value, err = models.OwMetric(s, a, in.Kind, in.Body)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwWidgetData]{Body: &value}, nil
}

func owMove(ctx context.Context, in *struct {
	Key  string `path:"key"`
	Body struct {
		ProjectID int64 `json:"project_id"`
	}
}) (*singleBody[models.OwIssue], error) {
	var value *models.OwIssue
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		value, err = models.OwMoveIssue(s, a, in.Key, in.Body.ProjectID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwIssue]{Body: value}, nil
}
