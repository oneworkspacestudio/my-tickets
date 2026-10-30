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
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func TestOwTokenScopesDistinguishListAndIssue(t *testing.T) {
	previous := apiTokenRoutesV2
	apiTokenRoutesV2 = map[string]APITokenRoute{}
	t.Cleanup(func() { apiTokenRoutesV2 = previous })
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: http.MethodGet, Path: "/api/v2/tickets/issues"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: http.MethodGet, Path: "/api/v2/tickets/issues/:key"}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{Method: http.MethodPost, Path: "/api/v2/tickets/issues"}, true)
	routes := apiTokenRoutesV2["tickets_issues"]
	require.Equal(t, "/api/v2/tickets/issues", routes["read_all"].Path)
	require.Equal(t, "/api/v2/tickets/issues/:key", routes["read_one"].Path)
	require.Equal(t, http.MethodPost, routes["create"].Method)
}
