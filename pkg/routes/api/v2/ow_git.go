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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

func init() { AddRouteRegistrar(RegisterOwGit) }
func RegisterOwGit(api huma.API) {
	Register(api, huma.Operation{OperationID: "tickets-git-settings-read", Summary: "Read project Git settings", Description: "Only project administrators; secrets are never returned.", Method: http.MethodGet, Path: "/tickets/projects/{id}/git", Tags: []string{"tickets"}}, owGitRead)
	Register(api, huma.Operation{OperationID: "tickets-git-settings-update", Summary: "Configure a Git repository", Description: "Encrypts credentials and validates exact repository and opt-in transition rules. Requires project administration.", Method: http.MethodPut, Path: "/tickets/projects/{id}/git", Tags: []string{"tickets"}}, owGitUpdate)
	Register(api, huma.Operation{OperationID: "tickets-git-resync", Summary: "Resync recent Git history", Description: "Imports up to 50 commits from the configured trusted origin; never executes transitions for historical records.", Method: http.MethodPost, Path: "/tickets/projects/{id}/git/resync", Tags: []string{"tickets"}}, owGitResync)
	Register(api, huma.Operation{OperationID: "tickets-git-webhook", Summary: "Receive a signed Gitea webhook", Description: "Requires a project-scoped API credential plus HMAC-SHA256 of the exact body and a delivery ID. Duplicate deliveries are idempotent.", Method: http.MethodPost, Path: "/tickets/projects/{id}/git/webhook", Tags: []string{"tickets"}, MaxBodyBytes: 8 << 20}, owGitWebhook)
	Register(api, huma.Operation{OperationID: "tickets-git-links", Summary: "Read issue Git links", Description: "Authorizes the issue before returning its commit and pull request links.", Method: http.MethodGet, Path: "/tickets/issues/{key}/git", Tags: []string{"tickets"}}, owGitLinks)
}
func owGitRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*singleBody[models.OwGitSettings], error) {
	var out *models.OwGitSettings
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		out, err = models.OwGitConfig(s, a, in.ID, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwGitSettings]{Body: out}, nil
}
func owGitUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body models.OwGitSettings
}) (*singleBody[models.OwGitSettings], error) {
	var out *models.OwGitSettings
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		out, err = models.OwGitConfig(s, a, in.ID, &in.Body)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwGitSettings]{Body: out}, nil
}

type owGitLinksBody struct{ Body []*models.OwGitLink }

func owGitLinks(ctx context.Context, in *struct {
	Key string `path:"key"`
}) (*owGitLinksBody, error) {
	var out []*models.OwGitLink
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		out, err = models.OwGitLinks(s, a, in.Key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &owGitLinksBody{Body: out}, nil
}
func owGitResync(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*owGitLinksBody, error) {
	var cfg *models.OwGitSettings
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		cfg, err = models.OwGitConfig(s, a, in.ID, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	batch, err := models.OwFetchGit(ctx, cfg)
	if err != nil {
		return nil, translateDomainError(err)
	}
	var out []*models.OwGitLink
	err = owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		out, err = models.OwApplyGitBatch(s, a, in.ID, batch)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &owGitLinksBody{Body: out}, nil
}
func owGitWebhook(ctx context.Context, in *struct {
	ID        int64  `path:"id"`
	Event     string `header:"X-Gitea-Event"`
	Delivery  string `header:"X-Gitea-Delivery"`
	Signature string `header:"X-Gitea-Signature"`
	RawBody   []byte
}) (*owGitLinksBody, error) {
	var out []*models.OwGitLink
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		out, err = models.OwGitWebhook(s, a, in.ID, in.Event, in.Delivery, in.Signature, in.RawBody)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &owGitLinksBody{Body: out}, nil
}
