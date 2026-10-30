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
	"fmt"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
)

func init() { AddRouteRegistrar(RegisterOwResources) }

func RegisterOwResources(api huma.API) {
	Register(api, huma.Operation{OperationID: "tickets-issues-list", Summary: "List ticket issues", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/issues", Tags: []string{"tickets"}}, owIssueList)
	Register(api, huma.Operation{OperationID: "tickets-issues-create", Summary: "Create ticket issues", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPost, Path: "/tickets/issues", Tags: []string{"tickets"}}, owIssueCreate)
	Register(api, huma.Operation{OperationID: "tickets-issues-read", Summary: "Read ticket issues", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/issues/{key}", Tags: []string{"tickets"}}, owIssueRead)
	Register(api, huma.Operation{OperationID: "tickets-issues-update", Summary: "Update ticket issues", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPut, Path: "/tickets/issues/{key}", Tags: []string{"tickets"}}, owIssueUpdate)
	Register(api, huma.Operation{OperationID: "tickets-issues-delete", Summary: "Delete ticket issues", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodDelete, Path: "/tickets/issues/{key}", Tags: []string{"tickets"}}, owIssueDelete)
	Register(api, huma.Operation{OperationID: "tickets-sprints-list", Summary: "List ticket sprints", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/sprints", Tags: []string{"tickets"}}, owSprintList)
	Register(api, huma.Operation{OperationID: "tickets-sprints-create", Summary: "Create ticket sprints", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPost, Path: "/tickets/sprints", Tags: []string{"tickets"}}, owSprintCreate)
	Register(api, huma.Operation{OperationID: "tickets-sprints-read", Summary: "Read ticket sprints", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/sprints/{id}", Tags: []string{"tickets"}}, owSprintRead)
	Register(api, huma.Operation{OperationID: "tickets-sprints-update", Summary: "Update ticket sprints", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPut, Path: "/tickets/sprints/{id}", Tags: []string{"tickets"}}, owSprintUpdate)
	Register(api, huma.Operation{OperationID: "tickets-sprints-delete", Summary: "Delete ticket sprints", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodDelete, Path: "/tickets/sprints/{id}", Tags: []string{"tickets"}}, owSprintDelete)
	Register(api, huma.Operation{OperationID: "tickets-dashboards-list", Summary: "List ticket dashboards", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/dashboards", Tags: []string{"tickets"}}, owDashboardList)
	Register(api, huma.Operation{OperationID: "tickets-dashboards-create", Summary: "Create ticket dashboards", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPost, Path: "/tickets/dashboards", Tags: []string{"tickets"}}, owDashboardCreate)
	Register(api, huma.Operation{OperationID: "tickets-dashboards-read", Summary: "Read ticket dashboards", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/dashboards/{id}", Tags: []string{"tickets"}}, owDashboardRead)
	Register(api, huma.Operation{OperationID: "tickets-dashboards-update", Summary: "Update ticket dashboards", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPut, Path: "/tickets/dashboards/{id}", Tags: []string{"tickets"}}, owDashboardUpdate)
	Register(api, huma.Operation{OperationID: "tickets-dashboards-delete", Summary: "Delete ticket dashboards", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodDelete, Path: "/tickets/dashboards/{id}", Tags: []string{"tickets"}}, owDashboardDelete)
	Register(api, huma.Operation{OperationID: "tickets-views-list", Summary: "List ticket views", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/views", Tags: []string{"tickets"}}, owSavedViewList)
	Register(api, huma.Operation{OperationID: "tickets-views-create", Summary: "Create ticket views", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPost, Path: "/tickets/views", Tags: []string{"tickets"}}, owSavedViewCreate)
	Register(api, huma.Operation{OperationID: "tickets-views-read", Summary: "Read ticket views", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodGet, Path: "/tickets/views/{id}", Tags: []string{"tickets"}}, owSavedViewRead)
	Register(api, huma.Operation{OperationID: "tickets-views-update", Summary: "Update ticket views", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodPut, Path: "/tickets/views/{id}", Tags: []string{"tickets"}}, owSavedViewUpdate)
	Register(api, huma.Operation{OperationID: "tickets-views-delete", Summary: "Delete ticket views", Description: "Project permissions are enforced; dashboard and saved-view ownership controls writes. Deletes archive issues and preserve their identifiers.", Method: http.MethodDelete, Path: "/tickets/views/{id}", Tags: []string{"tickets"}}, owSavedViewDelete)
}

type owIssueReadBody struct {
	models.OwIssue
	MaxPermission int `json:"max_permission" readOnly:"true" doc:"Current caller permission."`
}
type owIssueListBody struct{ Body Paginated[*models.OwIssue] }

func owIssueList(ctx context.Context, in *struct {
	ListParams
	ProjectID int64 `query:"project_id" doc:"Optional project scope."`
}) (*owIssueListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwIssue{}
	m.ProjectID = in.ProjectID
	rows, _, total, err := handler.DoReadAll(ctx, m, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := rows.([]*models.OwIssue)
	if !ok {
		return nil, fmt.Errorf("unexpected Issue list type %T", rows)
	}
	return &owIssueListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}
func owIssueCreate(ctx context.Context, in *struct{ Body models.OwIssue }) (*singleBody[models.OwIssue], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwIssue]{Body: &in.Body}, nil
}
func owIssueRead(ctx context.Context, in *struct {
	Key string `path:"key"`
	conditional.Params
}) (*singleReadBody[owIssueReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwIssue{Key: in.Key}
	permission, err := handler.DoReadOne(ctx, m, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return conditionalReadResponse(&in.Params, &owIssueReadBody{OwIssue: *m, MaxPermission: permission}, m.Updated, permission)
}
func owIssueUpdate(ctx context.Context, in *struct {
	Key  string `path:"key"`
	Body owIssueReadBody
}) (*singleBody[models.OwIssue], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.Key = in.Key
	if err = handler.DoUpdate(ctx, &in.Body.OwIssue, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwIssue]{Body: &in.Body.OwIssue}, nil
}
func owIssueDelete(ctx context.Context, in *struct {
	Key string `path:"key"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoDelete(ctx, &models.OwIssue{Key: in.Key}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

type owSprintReadBody struct {
	models.OwSprint
	MaxPermission int `json:"max_permission" readOnly:"true" doc:"Current caller permission."`
}
type owSprintListBody struct{ Body Paginated[*models.OwSprint] }

func owSprintList(ctx context.Context, in *struct {
	ListParams
	ProjectID int64 `query:"project_id" doc:"Optional project scope."`
}) (*owSprintListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwSprint{}
	m.ProjectID = in.ProjectID
	rows, _, total, err := handler.DoReadAll(ctx, m, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := rows.([]*models.OwSprint)
	if !ok {
		return nil, fmt.Errorf("unexpected Sprint list type %T", rows)
	}
	return &owSprintListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}
func owSprintCreate(ctx context.Context, in *struct{ Body models.OwSprint }) (*singleBody[models.OwSprint], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwSprint]{Body: &in.Body}, nil
}
func owSprintRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
	conditional.Params
}) (*singleReadBody[owSprintReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwSprint{ID: in.ID}
	permission, err := handler.DoReadOne(ctx, m, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return conditionalReadResponse(&in.Params, &owSprintReadBody{OwSprint: *m, MaxPermission: permission}, time.Time{}, permission)
}
func owSprintUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body owSprintReadBody
}) (*singleBody[models.OwSprint], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID
	if err = handler.DoUpdate(ctx, &in.Body.OwSprint, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwSprint]{Body: &in.Body.OwSprint}, nil
}
func owSprintDelete(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoDelete(ctx, &models.OwSprint{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

type owDashboardReadBody struct {
	models.OwDashboard
	MaxPermission int `json:"max_permission" readOnly:"true" doc:"Current caller permission."`
}
type owDashboardListBody struct {
	Body Paginated[*models.OwDashboard]
}

func owDashboardList(ctx context.Context, in *struct {
	ListParams
	ProjectID int64 `query:"project_id" doc:"Optional project scope."`
}) (*owDashboardListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwDashboard{}
	rows, _, total, err := handler.DoReadAll(ctx, m, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := rows.([]*models.OwDashboard)
	if !ok {
		return nil, fmt.Errorf("unexpected Dashboard list type %T", rows)
	}
	return &owDashboardListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}
func owDashboardCreate(ctx context.Context, in *struct{ Body models.OwDashboard }) (*singleBody[models.OwDashboard], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwDashboard]{Body: &in.Body}, nil
}
func owDashboardRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
	conditional.Params
}) (*singleReadBody[owDashboardReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwDashboard{ID: in.ID}
	permission, err := handler.DoReadOne(ctx, m, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return conditionalReadResponse(&in.Params, &owDashboardReadBody{OwDashboard: *m, MaxPermission: permission}, m.Updated, permission)
}
func owDashboardUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body owDashboardReadBody
}) (*singleBody[models.OwDashboard], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID
	if err = handler.DoUpdate(ctx, &in.Body.OwDashboard, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwDashboard]{Body: &in.Body.OwDashboard}, nil
}
func owDashboardDelete(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoDelete(ctx, &models.OwDashboard{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

type owSavedViewReadBody struct {
	models.OwSavedView
	MaxPermission int `json:"max_permission" readOnly:"true" doc:"Current caller permission."`
}
type owSavedViewListBody struct {
	Body Paginated[*models.OwSavedView]
}

func owSavedViewList(ctx context.Context, in *struct {
	ListParams
	ProjectID int64 `query:"project_id" doc:"Optional project scope."`
}) (*owSavedViewListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwSavedView{}
	rows, _, total, err := handler.DoReadAll(ctx, m, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := rows.([]*models.OwSavedView)
	if !ok {
		return nil, fmt.Errorf("unexpected SavedView list type %T", rows)
	}
	return &owSavedViewListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}
func owSavedViewCreate(ctx context.Context, in *struct{ Body models.OwSavedView }) (*singleBody[models.OwSavedView], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwSavedView]{Body: &in.Body}, nil
}
func owSavedViewRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
	conditional.Params
}) (*singleReadBody[owSavedViewReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.OwSavedView{ID: in.ID}
	permission, err := handler.DoReadOne(ctx, m, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return conditionalReadResponse(&in.Params, &owSavedViewReadBody{OwSavedView: *m, MaxPermission: permission}, time.Time{}, permission)
}
func owSavedViewUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body owSavedViewReadBody
}) (*singleBody[models.OwSavedView], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID
	if err = handler.DoUpdate(ctx, &in.Body.OwSavedView, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.OwSavedView]{Body: &in.Body.OwSavedView}, nil
}
func owSavedViewDelete(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err = handler.DoDelete(ctx, &models.OwSavedView{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
