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
	"mime"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"xorm.io/xorm"
)

func init() { AddRouteRegistrar(RegisterOwAttachments) }
func RegisterOwAttachments(api huma.API) {
	Register(api, huma.Operation{OperationID: "tickets-attachments-list", Summary: "List issue attachments", Description: "Requires issue read permission. Includes recoverable deletion state.", Method: http.MethodGet, Path: "/tickets/issues/{key}/attachments", Tags: []string{"tickets"}}, owAttachments)
	Register(api, huma.Operation{OperationID: "tickets-attachments-upload", Summary: "Upload an attachment", Description: "Streams a file into the protected Drive source, with issue write authorization, SHA256 and a 64 MB cap.", Method: http.MethodPost, Path: "/tickets/issues/{key}/attachments", Tags: []string{"tickets"}, MaxBodyBytes: 66 << 20}, owAttachmentUpload)
	Register(api, huma.Operation{OperationID: "tickets-attachments-download", Summary: "Download an attachment", Description: "Authorizes current issue membership before streaming bytes; filenames cannot select paths.", Method: http.MethodGet, Path: "/tickets/attachments/{id}", Tags: []string{"tickets"}}, owAttachmentDownload)
	Register(api, huma.Operation{OperationID: "tickets-attachments-state", Summary: "Delete or restore an attachment", Description: "Requires write permission. Recoverable deletion retains bytes and history.", Method: http.MethodPut, Path: "/tickets/attachments/{id}/state", Tags: []string{"tickets"}}, owAttachmentState)
}

type owAttachmentList struct{ Body []*models.OwAttachment }

func owAttachments(ctx context.Context, in *struct {
	Key string `path:"key"`
}) (*owAttachmentList, error) {
	var out []*models.OwAttachment
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		out, err = models.OwAttachments(s, a, in.Key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &owAttachmentList{Body: out}, nil
}
func owAttachmentUpload(ctx context.Context, in *struct {
	Key     string `path:"key"`
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"file" required:"true" doc:"One file, up to 64 MB."`
	}]
}) (*singleBody[models.OwAttachment], error) {
	var out *models.OwAttachment
	file := in.RawBody.Data().File
	defer file.Close()
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error {
		var err error
		out, err = models.OwUploadAttachment(s, a, in.Key, file.Filename, file)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &singleBody[models.OwAttachment]{Body: out}, nil
}
func owAttachmentState(ctx context.Context, in *struct {
	ID   string `path:"id"`
	Body struct {
		Deleted bool `json:"deleted" doc:"False restores the attachment."`
	}
}) (*emptyBody, error) {
	err := owTransaction(ctx, func(s *xorm.Session, a web.Auth) error { return models.OwAttachmentState(s, a, in.ID, in.Body.Deleted) })
	if err != nil {
		return nil, err
	}
	return &emptyBody{}, nil
}
func owAttachmentDownload(ctx context.Context, in *struct {
	ID string `path:"id"`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	v, f, err := models.OwOpenAttachment(s, a, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err = s.Commit(); err != nil {
		f.Close()
		return nil, err
	}
	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		defer f.Close()
		w := (*humaecho.Unwrap(hctx)).Response()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": v.Name}))
		w.Header().Set("X-Content-SHA256", v.SHA256)
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Content-Length", fmt.Sprint(v.Size))
		http.ServeContent(w, (*humaecho.Unwrap(hctx)).Request(), v.Name, v.Created, f)
	}}, nil
}
