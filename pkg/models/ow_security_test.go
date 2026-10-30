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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/require"
)

func TestOwGenericAPICannotBypassTicketWorkflow(t *testing.T) {
	s, a, p := owTestData(t)
	i := &OwIssue{ProjectID: p.ProjectID, Summary: "Protected"}
	require.NoError(t, i.Create(s, a))
	require.NoError(t, i.ReadOne(s, a))
	generic := i.Task
	generic.Done = true
	require.Error(t, generic.Update(s, a))
	require.Error(t, generic.Delete(s, a))
	require.Error(t, (&Task{ProjectID: p.ProjectID, Title: "Bypass"}).Create(s, a))
	require.NoError(t, i.ReadOne(s, a))
	require.False(t, i.Task.Done)
}
func TestOwMovePreservesAliasAndPermission(t *testing.T) {
	s, a, p := owTestData(t)
	i := &OwIssue{ProjectID: p.ProjectID, Summary: "Move me"}
	require.NoError(t, i.Create(s, a))
	old := i.Key
	target, err := OwCreateProject(s, a, "Destination", "DEST")
	require.NoError(t, err)
	_, err = OwMoveIssue(s, &user.User{ID: 13}, old, target.ProjectID)
	require.Error(t, err)
	moved, err := OwMoveIssue(s, a, old, target.ProjectID)
	require.NoError(t, err)
	require.Equal(t, "DEST-1", moved.Key)
	alias := &OwIssue{Key: old}
	require.NoError(t, alias.ReadOne(s, a))
	require.Equal(t, moved.TaskID, alias.TaskID)
	require.Equal(t, moved.Key, alias.Key)
}
func TestOwGitWebhookSignatureDedupAndPermission(t *testing.T) {
	s, a, p := owTestData(t)
	i := &OwIssue{ProjectID: p.ProjectID, Summary: "Git fixture", State: "testing"}
	require.NoError(t, i.Create(s, a))
	secret := strings.Repeat("synthetic-", 5)
	cfg := &OwGitSettings{URL: "https://git.example.test", Repository: "team/test", Token: "synthetic-token", Secret: secret}
	_, err := OwGitConfig(s, a, p.ProjectID, cfg)
	require.NoError(t, err)
	require.NotContains(t, cfg.TokenCipher, "synthetic")
	_, err = OwGitConfig(s, &user.User{ID: 13}, p.ProjectID, nil)
	require.Error(t, err)
	raw := []byte(fmt.Sprintf(`{"repository":{"full_name":"team/test"},"commits":[{"id":"1","message":"Fix %s","url":"https://git.example.test/team/test/commit/1"}]}`, i.Key))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	sig := hex.EncodeToString(mac.Sum(nil))
	_, err = OwGitWebhook(s, a, p.ProjectID, "push", "test1", "bad", raw)
	require.Error(t, err)
	links, err := OwGitWebhook(s, a, p.ProjectID, "push", "test1", sig, raw)
	require.NoError(t, err)
	require.Len(t, links, 1)
	links, err = OwGitWebhook(s, a, p.ProjectID, "push", "test1", sig, raw)
	require.NoError(t, err)
	require.Empty(t, links)
	_, err = OwGitLinks(s, &user.User{ID: 13}, i.Key)
	require.Error(t, err)
	cfg.Rules = []OwGitRule{{Event: "pull_request.closed", State: "done"}}
	_, err = OwGitConfig(s, a, p.ProjectID, cfg)
	require.NoError(t, err)
	for _, merged := range []bool{false, true} {
		raw = []byte(fmt.Sprintf(`{"repository":{"full_name":"team/test"},"action":"closed","pull_request":{"title":"%s review","html_url":"https://git.example.test/team/test/pulls/1","merged":%t}}`, i.Key, merged))
		mac = hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write(raw)
		_, err = OwGitWebhook(s, a, p.ProjectID, "pull_request", fmt.Sprint("pr-", merged), hex.EncodeToString(mac.Sum(nil)), raw)
		require.NoError(t, err)
	}
	require.NoError(t, i.ReadOne(s, a))
	require.Equal(t, "done", i.State, "an existing PR link must still trigger an authorized merge rule")
}
func TestOwAttachmentLifecycleAndTraversal(t *testing.T) {
	s, a, p := owTestData(t)
	i := &OwIssue{ProjectID: p.ProjectID, Summary: "Attachment fixture"}
	require.NoError(t, i.Create(s, a))
	previous := config.FilesBasePath.GetString()
	config.FilesBasePath.Set(t.TempDir())
	t.Cleanup(func() { config.FilesBasePath.Set(previous) })
	_, err := OwUploadAttachment(s, a, i.Key, "../escape.txt", strings.NewReader("no"))
	require.Error(t, err)
	_, err = OwUploadAttachment(s, &user.User{ID: 13}, i.Key, "no.txt", strings.NewReader("no"))
	require.Error(t, err)
	attachment, err := OwUploadAttachment(s, a, i.Key, "evidence.txt", strings.NewReader("synthetic attachment"))
	require.NoError(t, err)
	require.Contains(t, attachment.Path, p.Key+"/"+i.Key+"/")
	_, f, err := OwOpenAttachment(s, a, attachment.ID)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.NoError(t, OwAttachmentState(s, a, attachment.ID, true))
	_, _, err = OwOpenAttachment(s, a, attachment.ID)
	require.Error(t, err)
	require.NoError(t, OwAttachmentState(s, a, attachment.ID, false))
	_, f, err = OwOpenAttachment(s, a, attachment.ID)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	_, _, err = OwOpenAttachment(s, &user.User{ID: 13}, attachment.ID)
	require.Error(t, err)
}
func TestOwTypedNumericFilters(t *testing.T) {
	require.True(t, owMatch(&OwIssue{Priority: 2}, OwRule{Field: "priority", Operator: "before", Value: "10"}, nil))
	require.Error(t, owValidateFilter(OwFilter{Rules: []OwRule{{Field: "invented", Operator: "eq"}}}))
}
