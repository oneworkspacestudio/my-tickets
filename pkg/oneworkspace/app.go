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

package oneworkspace

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"code.vikunja.io/api/frontend"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/plugins"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/version"
	"github.com/mattn/go-sqlite3"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

const ID = "tickets"

var Version = "0.1.0-preview.1"
var Commit = "development"
var BuildTime = "unknown"
var errInactiveTicketAccount = errors.New("ticket account is inactive")

type nativeSession struct {
	ID      string
	Expires time.Time
}
type App struct {
	Host     *apphost.Host
	Core     http.Handler
	SQL      *sql.DB
	mu       sync.Mutex
	sessions map[string]nativeSession
}

var runContext = context.Background()

func Main() error {
	if handled, err := nativeService(); handled {
		return err
	}
	if len(os.Args) > 1 && os.Args[1] == "doctor" {
		return doctorCommand(os.Args[2:])
	}
	return serve()
}
func serve() error {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	}
	for index, arg := range os.Args {
		if arg == "--data-dir" {
			os.Args[index] = "--data"
		}
	}
	hasData := false
	for _, arg := range os.Args[1:] {
		if arg == "--data" || strings.HasPrefix(arg, "--data=") {
			hasData = true
		}
	}
	if !hasData {
		root, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		os.Args = append(os.Args, "--data", filepath.Join(root, "OneWorkspace", "MyTickets"))
	}
	return apphost.Main(apphost.Config{ParentContext: runContext, ID: ID, Name: "My Tickets", Version: Version, Listen: "127.0.0.1:9026", RestoreDatabases: []string{"core/mytickets.db"}, RestoreFiles: []string{"core/mytickets.db", "core.key", "core/attachments/"}}, nil, New)
}

func New(h *apphost.Host) (apphost.Application, error) {
	webOrigin, err := nativeWebOrigin(h.Config.Origin)
	if err != nil {
		return nil, err
	}
	temp := filepath.Join(h.Config.Data, "tmp")
	if err := os.MkdirAll(temp, 0700); err != nil {
		return nil, err
	}
	if err := os.Setenv("TEMP", temp); err != nil {
		return nil, err
	}
	if err := os.Setenv("TMP", temp); err != nil {
		return nil, err
	}
	core := filepath.Join(h.Config.Data, "core")
	if err := os.MkdirAll(filepath.Join(core, "attachments"), 0700); err != nil {
		return nil, err
	}
	secretPath := filepath.Join(h.Config.Data, "core.key")
	secret, err := os.ReadFile(secretPath)
	if os.IsNotExist(err) {
		bytes := make([]byte, 64)
		if _, err = rand.Read(bytes); err != nil {
			return nil, err
		}
		secret = []byte(hex.EncodeToString(bytes))
		file, err := os.OpenFile(secretPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		_, writeErr := file.Write(secret)
		closeErr := file.Close()
		if writeErr != nil {
			return nil, writeErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
	} else if err != nil {
		return nil, err
	}
	if len(secret) != 128 {
		return nil, errors.New("ticket signing key is invalid; restore a verified backup")
	}
	config.ServiceRootpath.Set(core)
	config.DatabaseType.Set("sqlite")
	config.DatabaseMaxOpenConnections.Set(1)
	config.DatabaseMaxIdleConnections.Set(1)
	config.DatabasePath.Set(filepath.Join(core, "mytickets.db"))
	config.FilesBasePath.Set(filepath.Join(core, "attachments"))
	config.ServiceSecret.Set(string(secret))
	config.ServicePublicURL.Set(h.Config.Origin + "/")
	config.CorsOrigins.Set([]string{webOrigin})
	config.ServiceEnableRegistration.Set(false)
	config.AuthLocalEnabled.Set(false)
	config.AuthOpenIDEnabled.Set(false)
	config.MailerEnabled.Set(false)
	config.RedisEnabled.Set(false)
	config.ServiceEnableCaldav.Set(false)
	config.ServiceEnableLinkSharing.Set(false)
	config.ServiceJWTTTL.Set(15 * 60)
	initialize.FullInit()
	s := db.NewSession()
	engine := s.Engine().DB().DB
	_ = s.Close()
	// Upstream's SQLite constructor does not apply its configurable pool limits.
	// One writer connection serializes transaction upgrades and atomic issue counters.
	engine.SetMaxOpenConns(1)
	engine.SetMaxIdleConns(1)
	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	a := &App{Host: h, Core: e, SQL: engine, sessions: map[string]nativeSession{}}
	gateway := a.gateway()
	for _, method := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
		h.Mux.Handle(method+" /api/v2/", gateway)
		h.Mux.HandleFunc(method+" /api/v1/", func(w http.ResponseWriter, _ *http.Request) { apphost.Problem(w, 404, "Use the My Tickets API v2") })
	}
	h.Mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if h.DB.PingContext(ctx) != nil || engine.PingContext(ctx) != nil {
			apphost.Problem(w, 503, "My Tickets is still preparing its databases")
			return
		}
		apphost.JSON(w, 200, map[string]string{"status": "ready", "version": Version})
	})
	h.UIHandler = http.HandlerFunc(a.ui)
	h.SnapshotFiles = map[string]func(string) error{"core/mytickets.db": func(dest string) error { return snapshotSQLite(engine, dest) }}
	h.BackupPaths = []string{"core.key", "core/attachments"}
	h.CloseHooks = append(h.CloseHooks, engine.Close)
	return a, nil
}

func nativeWebOrigin(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", errors.New("configured HTTP or HTTPS origin is required")
	}
	return u.Scheme + "://" + u.Host, nil
}

func (a *App) Run(ctx context.Context) error {
	<-ctx.Done()
	cron.Stop()
	license.Shutdown()
	plugins.Shutdown()
	return nil
}

func (a *App) gateway() http.Handler {
	native := a.Host.Require(func(w http.ResponseWriter, r *http.Request) {
		session := apphost.Current(r)
		cookie, err := r.Cookie(ID + "_session")
		if err != nil {
			apphost.Problem(w, 401, "Sign in again")
			return
		}
		token, err := a.coreToken(r.Context(), session, cookie.Value)
		if err != nil {
			a.Host.Log.Error("ticket identity could not be resolved", "kind", fmt.Sprintf("%T", err))
			if errors.Is(err, errInactiveTicketAccount) {
				apphost.Problem(w, 403, "Your ticket account is inactive. Contact your Workspace administrator.")
				return
			}
			apphost.Problem(w, 503, "Your ticket account could not be prepared. Retry in a moment; your Workspace sign-in is retained.")
			return
		}
		r = r.Clone(r.Context())
		r.Header = r.Header.Clone()
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("X-Real-IP")
		r.Header.Del("Forwarded")
		a.Core.ServeHTTP(w, r)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			apphost.Problem(w, 404, "This client uses revocable HTTP sessions")
			return
		}
		// Agent credentials retain upstream per-route scopes and model permissions. Browser bearer JWTs never bypass the native cookie gate.
		if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer tk_") {
			a.Host.Freeze.RLock()
			defer a.Host.Freeze.RUnlock()
			r = r.Clone(r.Context())
			r.Header = r.Header.Clone()
			r.Header.Del("X-Forwarded-For")
			r.Header.Del("X-Real-IP")
			r.Header.Del("Forwarded")
			a.Core.ServeHTTP(w, r)
			return
		}
		native(w, r)
	})
}

func (a *App) coreToken(ctx context.Context, session apphost.Session, rawCookie string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := db.NewSession()
	defer s.Close()
	u := &user.User{}
	found, err := s.Where("issuer = ? AND subject = ?", session.User.Issuer, session.User.Subject).Get(u)
	if err != nil {
		return "", err
	}
	if !found {
		u, err = auth.CreateUserWithRandomUsername(s, &user.User{Issuer: session.User.Issuer, Subject: session.User.Subject, Name: session.User.Name, Email: session.User.Email}, user.CreateUserOptions{SkipEmailConfirm: true, AllowMissingExternalEmail: true})
		if err != nil {
			return "", err
		}
	}
	if u.Status != user.StatusActive {
		return "", errInactiveTicketAccount
	}
	if u.Name != session.User.Name || u.Email != session.User.Email || u.IsAdmin != (session.User.Role == "owner") {
		u.Name = session.User.Name
		u.Email = session.User.Email
		u.IsAdmin = session.User.Role == "owner"
		if err = user.InvalidateSessionAuthCacheOnCommit(s, u.ID); err != nil {
			return "", err
		}
		if _, err = s.ID(u.ID).Cols("name", "email", "is_admin").Update(u); err != nil {
			return "", err
		}
	}
	key := apphost.Hash(rawCookie)
	native, exists := a.sessions[key]
	if exists {
		if !time.Now().Before(native.Expires) {
			delete(a.sessions, key)
			exists = false
		} else if _, err = models.GetSessionByID(s, native.ID); err != nil {
			return "", err
		}
	}
	if !exists {
		for id, item := range a.sessions {
			if !time.Now().Before(item.Expires) {
				delete(a.sessions, id)
			}
		}
		if len(a.sessions) >= 10000 {
			return "", errors.New("too many native sessions")
		}
		created, err := models.CreateSession(s, u.ID, "One Workspace browser", "", false, nil)
		if err != nil {
			return "", err
		}
		native = nativeSession{ID: created.ID, Expires: session.Expires}
	}
	token, err := auth.NewUserJWTAuthtoken(u, native.ID)
	if err != nil {
		return "", err
	}
	if err = s.Commit(); err != nil {
		return "", err
	}
	events.DispatchPending(ctx, s)
	a.sessions[key] = native
	return token, nil
}

func (a *App) ui(w http.ResponseWriter, r *http.Request) {
	assets, err := fs.Sub(frontend.Files, "dist")
	if err != nil {
		http.Error(w, "Embedded interface unavailable", 500)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/")
	if strings.HasPrefix(name, "assets/") || strings.HasSuffix(name, ".svg") || strings.HasSuffix(name, ".woff2") || name == "favicon.ico" {
		http.FileServer(http.FS(assets)).ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(name, "api/") || strings.HasPrefix(name, "auth/") {
		http.NotFound(w, r)
		return
	}
	page, err := fs.ReadFile(assets, "mytickets.html")
	if err != nil {
		http.Error(w, "Build the complete My Tickets interface before running this app", http.StatusServiceUnavailable)
		return
	}
	base := a.Host.Base + "/"
	content := strings.Replace(string(page), "<!--workspace-base-->", `<base href="`+html.EscapeString(base)+`"><meta name="workspace-base" content="`+html.EscapeString(a.Host.Base)+`">`, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(content))
}

func snapshotSQLite(source *sql.DB, dest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := sql.Open("sqlite3", dest)
	if err != nil {
		return err
	}
	defer out.Close()
	src, err := source.Conn(ctx)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := out.Conn(ctx)
	if err != nil {
		return err
	}
	defer dst.Close()
	return src.Raw(func(driver any) error {
		from, ok := driver.(*sqlite3.SQLiteConn)
		if !ok {
			return errors.New("unsupported ticket database driver")
		}
		return dst.Raw(func(driver any) (result error) {
			to, ok := driver.(*sqlite3.SQLiteConn)
			if !ok {
				return errors.New("unsupported snapshot database driver")
			}
			backup, err := to.Backup("main", from, "main")
			if err != nil {
				return err
			}
			defer func() {
				if finishErr := backup.Finish(); result == nil {
					result = finishErr
				}
			}()
			for {
				done, err := backup.Step(256)
				if err != nil {
					return err
				}
				if done {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(5 * time.Millisecond):
				}
			}
		})
	})
}

func Diagnostics() map[string]string {
	return map[string]string{"app": "My Tickets", "version": Version, "commit": Commit, "built": BuildTime, "upstream": version.Version, "upstream_commit": "5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0"}
}
