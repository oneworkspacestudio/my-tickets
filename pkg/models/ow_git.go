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
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type OwGitSettings struct {
	ProjectID    int64       `xorm:"bigint pk" json:"project_id" doc:"Bound ticket project."`
	URL          string      `xorm:"varchar(500)" json:"url" doc:"Explicit trusted Gitea server base URL."`
	Repository   string      `xorm:"varchar(250)" json:"repository" doc:"Exact repository owner/name."`
	Token        string      `xorm:"-" json:"token,omitempty" writeOnly:"true" doc:"Service token; empty preserves the current encrypted token."`
	Secret       string      `xorm:"-" json:"secret,omitempty" writeOnly:"true" doc:"Webhook HMAC secret; empty preserves the current secret."`
	TokenCipher  string      `xorm:"text" json:"-"`
	SecretCipher string      `xorm:"text" json:"-"`
	Rules        []OwGitRule `xorm:"json" json:"rules" doc:"Explicit opt-in workflow automation, disabled when empty."`
	Configured   bool        `xorm:"-" json:"configured" readOnly:"true" doc:"Whether credentials and webhook secret exist."`
}

func (*OwGitSettings) TableName() string { return "ow_git_settings" }

type OwGitRule struct {
	Event string `json:"event" doc:"push or pull_request.closed."`
	State string `json:"state" doc:"Validated destination workflow state."`
}
type OwGitLink struct {
	Repository string    `xorm:"varchar(250)" json:"repository" doc:"Bound repository name."`
	Revision   string    `xorm:"varchar(100)" json:"revision" doc:"Commit hash or branch reference."`
	Author     string    `xorm:"varchar(250)" json:"author" doc:"External author name."`
	State      string    `xorm:"varchar(30)" json:"state" doc:"External pull request state."`
	Occurred   time.Time `json:"occurred" doc:"External event timestamp when supplied."`
	URLHash    string    `xorm:"varchar(64) unique(ow_git_link)" json:"-"`
	ID         int64     `xorm:"bigint pk autoincr" json:"id" readOnly:"true" doc:"Link ID."`
	TaskID     int64     `xorm:"bigint index unique(ow_git_link)" json:"task_id" doc:"Related issue task."`
	URL        string    `xorm:"varchar(1000)" json:"url" doc:"Validated link on the configured Git server."`
	Kind       string    `xorm:"varchar(80)" json:"kind" doc:"Commit or pull request."`
	Title      string    `xorm:"text" json:"title" doc:"External title; always treated as untrusted text."`
	At         time.Time `xorm:"created" json:"at" readOnly:"true" doc:"First association time."`
}

func (*OwGitLink) TableName() string { return "ow_git_links" }

type OwGitDelivery struct {
	ID        string    `xorm:"varchar(64) pk"`
	ProjectID int64     `xorm:"bigint index"`
	At        time.Time `xorm:"created"`
}

func (*OwGitDelivery) TableName() string { return "ow_git_deliveries" }
func owCrypt(value string, decode bool) (string, error) {
	key := sha256.Sum256([]byte("mytickets-integration-v1\x00" + config.ServiceSecret.GetString()))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if decode {
		b, err := hex.DecodeString(value)
		if err != nil || len(b) < gcm.NonceSize() {
			return "", errors.New("integration credential cannot be decrypted")
		}
		p, err := gcm.Open(nil, b[:gcm.NonceSize()], b[gcm.NonceSize():], nil)
		return string(p), err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(value), nil)), nil
}
func OwGitConfig(s *xorm.Session, a web.Auth, id int64, in *OwGitSettings) (*OwGitSettings, error) {
	ok, err := owAdmin(s, id, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	old := &OwGitSettings{ProjectID: id}
	exists, err := s.ID(id).Get(old)
	if err != nil {
		return nil, err
	}
	if in == nil {
		old.Configured = old.TokenCipher != "" && old.SecretCipher != ""
		return old, nil
	}
	target, err := url.Parse(in.URL)
	if err != nil || target.Host == "" || target.User != nil || target.RawQuery != "" || target.Fragment != "" || !owAllowedGitOrigin(target) {
		return nil, owInvalid("Use HTTPS for the trusted Git server")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`).MatchString(in.Repository) || strings.Contains(in.Repository, "..") || len(in.Rules) > 16 {
		return nil, owInvalid("Choose one exact owner/repository")
	}
	settings, err := owSettings(s, id)
	if err != nil {
		return nil, err
	}
	for _, rule := range in.Rules {
		if rule.Event != "push" && rule.Event != "pull_request.closed" || owState(settings, rule.State) == nil {
			return nil, owInvalid("Check Git transition rules")
		}
	}
	in.ProjectID = id
	in.URL = strings.TrimRight(in.URL, "/")
	in.TokenCipher = old.TokenCipher
	in.SecretCipher = old.SecretCipher
	if in.Token != "" {
		in.TokenCipher, err = owCrypt(in.Token, false)
		if err != nil {
			return nil, err
		}
	}
	if in.Secret != "" {
		if len(in.Secret) < 32 {
			return nil, owInvalid("Use at least 32 characters for the webhook secret")
		}
		in.SecretCipher, err = owCrypt(in.Secret, false)
		if err != nil {
			return nil, err
		}
	}
	if exists {
		_, err = s.ID(id).AllCols().Update(in)
	} else {
		_, err = s.Insert(in)
	}
	if err != nil {
		return nil, err
	}
	in.Token = ""
	in.Secret = ""
	in.Configured = in.TokenCipher != "" && in.SecretCipher != ""
	return in, nil
}
func OwGitLinks(s *xorm.Session, a web.Auth, key string) ([]*OwGitLink, error) {
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
	out := []*OwGitLink{}
	err = s.Where("task_id = ?", i.TaskID).Desc("id").Find(&out)
	return out, err
}

type owGitCommit struct {
	Kind      string    `json:"-"`
	Timestamp time.Time `json:"timestamp"`
	Author    struct {
		Name string `json:"name"`
	} `json:"author"`
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
	HTMLURL string `json:"html_url"`
}
type owGitPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Commits     []owGitCommit `json:"commits"`
	Action      string        `json:"action"`
	PullRequest *struct {
		Title   string    `json:"title"`
		Body    string    `json:"body"`
		URL     string    `json:"html_url"`
		Merged  bool      `json:"merged"`
		State   string    `json:"state"`
		Updated time.Time `json:"updated_at"`
		User    struct {
			Name  string `json:"full_name"`
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
}

func OwGitWebhook(s *xorm.Session, a web.Auth, project int64, event, delivery, signature string, raw []byte) ([]*OwGitLink, error) {
	ok, err := (&Project{ID: project}).CanWrite(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	cfg := &OwGitSettings{}
	found, err := s.ID(project).Get(cfg)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, owInvalid("Configure this project's Git connection first")
	}
	secret, err := owCrypt(cfg.SecretCipher, true)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil || !hmac.Equal(mac.Sum(nil), sig) {
		return nil, owDenied()
	}
	if len(raw) > 8<<20 || len(delivery) > 250 || delivery == "" {
		return nil, owInvalid("Check Git delivery ID and size")
	}
	hash := sha256.Sum256([]byte(fmt.Sprint(project) + "\x00" + cfg.Repository + "\x00" + delivery))
	id := hex.EncodeToString(hash[:])
	found, err = s.ID(id).Exist(&OwGitDelivery{})
	if err != nil {
		return nil, err
	}
	if found {
		return []*OwGitLink{}, nil
	}
	var body owGitPayload
	if json.Unmarshal(raw, &body) != nil || body.Repository.FullName != cfg.Repository {
		return nil, owInvalid("Webhook repository does not match its binding")
	}
	effective := event
	if event == "pull_request" {
		effective += "." + body.Action
	}
	result, err := owGitApply(s, a, cfg, effective, body)
	if err != nil {
		return nil, err
	}
	_, err = s.Insert(&OwGitDelivery{ID: id, ProjectID: project})
	return result, err
}
func owGitEntries(cfg *OwGitSettings, body owGitPayload) ([]owGitCommit, error) {
	entries := body.Commits
	if len(entries) > 200 {
		return nil, owInvalid("Split large Git deliveries")
	}
	if body.PullRequest != nil {
		entry := owGitCommit{Kind: "pull_request", Message: body.PullRequest.Title + "\n" + body.PullRequest.Body, URL: body.PullRequest.URL, Timestamp: body.PullRequest.Updated}
		entry.Author.Name = body.PullRequest.User.Name
		if entry.Author.Name == "" {
			entry.Author.Name = body.PullRequest.User.Login
		}
		entries = append(entries, entry)
	}
	if strings.HasPrefix(body.Ref, "refs/heads/") {
		branch := strings.TrimPrefix(body.Ref, "refs/heads/")
		entries = append(entries, owGitCommit{Kind: "branch", ID: branch, Message: branch, URL: cfg.URL + "/" + cfg.Repository + "/src/branch/" + url.PathEscape(branch)})
	}
	return entries, nil
}

func owGitApply(s *xorm.Session, a web.Auth, cfg *OwGitSettings, event string, body owGitPayload) ([]*OwGitLink, error) {
	entries, err := owGitEntries(cfg, body)
	if err != nil {
		return nil, err
	}
	result := []*OwGitLink{}
	base, _ := url.Parse(cfg.URL)
	keys := regexp.MustCompile(`\b[A-Z][A-Z0-9]{0,9}-[0-9]+\b`)
	for _, entry := range entries {
		address := entry.HTMLURL
		if address == "" {
			address = entry.URL
		}
		target, err := url.Parse(address)
		if err != nil || target.User != nil || target.Scheme != base.Scheme || target.Host != base.Host || !strings.HasPrefix(target.Path, base.Path+"/"+cfg.Repository+"/") {
			continue
		}
		matches := keys.FindAllString(entry.Message, 101)
		if len(matches) > 100 {
			return nil, owInvalid("A Git record can link at most 100 issues")
		}
		for _, key := range matches {
			issue := &OwIssue{Key: key}
			ok, _, err := issue.CanRead(s, a)
			if err != nil {
				var missing OwError
				if errors.As(err, &missing) && missing.Status == 404 {
					continue
				}
				return nil, err
			}
			if !ok {
				continue
			}
			if err = issue.ReadOne(s, a); err != nil {
				return nil, err
			}
			if issue.ProjectID != cfg.ProjectID {
				continue
			}
			exists, err := s.Where("task_id = ? AND url = ?", issue.TaskID, address).Exist(&OwGitLink{})
			if err != nil {
				return nil, err
			}
			link := owGitLinkFor(cfg, issue, entry, body, address)
			if exists {
				_, err = s.Where("task_id = ? AND url_hash = ?", issue.TaskID, link.URLHash).Cols("title", "state", "author", "occurred").Update(link)
			} else {
				_, err = s.Insert(link)
			}
			if err != nil {
				return nil, err
			}
			result = append(result, link)
			if !exists {
				if err = owActivity(s, a, issue.ProjectID, issue.TaskID, "git.linked", link.Kind+" linked"); err != nil {
					return nil, err
				}
			}
			for _, rule := range cfg.Rules {
				if rule.Event == event {
					if event == "pull_request.closed" && (body.PullRequest == nil || !body.PullRequest.Merged) {
						continue
					}
					issue.State = rule.State
					if err = issue.Update(s, a); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return result, nil
}

func owGitLinkFor(cfg *OwGitSettings, issue *OwIssue, entry owGitCommit, body owGitPayload, address string) *OwGitLink {
	kind := entry.Kind
	if kind == "" {
		kind = "commit"
	}
	title := entry.Message
	if len(title) > 4000 {
		title = title[:4000]
	}
	state := ""
	if kind == "pull_request" && body.PullRequest != nil {
		state = body.PullRequest.State
		if body.PullRequest.Merged {
			state = "merged"
		}
	}
	urlHash := sha256.Sum256([]byte(address))
	return &OwGitLink{
		Repository: cfg.Repository,
		Revision:   entry.ID,
		Author:     entry.Author.Name,
		State:      state,
		Occurred:   entry.Timestamp,
		URLHash:    hex.EncodeToString(urlHash[:]),
		TaskID:     issue.TaskID,
		URL:        address,
		Kind:       kind,
		Title:      title,
	}
}

type OwGitBatch struct {
	server, repository string
	body               owGitPayload
}

// OwFetchGit makes network calls without holding a SQLite transaction.
func OwFetchGit(ctx context.Context, cfg *OwGitSettings) (*OwGitBatch, error) {
	token, err := owCrypt(cfg.TokenCipher, true)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", cfg.URL+"/api/v1/repos/"+cfg.Repository+"/commits?limit=50", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	client := http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return nil, owInvalid("Git is unavailable; existing links and issues are preserved")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, owInvalid("Git refused the connection; check the repository and service token")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20+1))
	if err != nil || len(raw) > 8<<20 {
		return nil, owInvalid("Git response exceeds the permitted size")
	}
	var commits []struct {
		SHA    string `json:"sha"`
		URL    string `json:"html_url"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string    `json:"name"`
				Date time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if json.Unmarshal(raw, &commits) != nil {
		return nil, owInvalid("Git returned an unsupported response")
	}
	body := owGitPayload{}
	for _, entry := range commits {
		item := owGitCommit{ID: entry.SHA, Message: entry.Commit.Message, URL: entry.URL, Timestamp: entry.Commit.Author.Date}
		item.Author.Name = entry.Commit.Author.Name
		body.Commits = append(body.Commits, item)
	}
	// Resync only links history; explicit workflow transitions run only on authenticated live events.
	return &OwGitBatch{server: cfg.URL, repository: cfg.Repository, body: body}, nil
}

func owAllowedGitOrigin(target *url.URL) bool {
	return target.Scheme == "https" || target.Scheme == "http" && (target.Hostname() == "127.0.0.1" || target.Hostname() == "localhost")
}

func OwApplyGitBatch(s *xorm.Session, a web.Auth, project int64, batch *OwGitBatch) ([]*OwGitLink, error) {
	cfg, err := OwGitConfig(s, a, project, nil)
	if err != nil {
		return nil, err
	}
	if cfg.URL != batch.server || cfg.Repository != batch.repository {
		return nil, OwError{Status: 409, Message: "Git settings changed while checking; retry the resync"}
	}
	return owGitApply(s, a, cfg, "resync", batch.body)
}
