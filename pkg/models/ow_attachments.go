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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type OwAttachment struct {
	ID         string    `xorm:"varchar(48) pk" json:"id" readOnly:"true" doc:"Opaque attachment ID."`
	TaskID     int64     `xorm:"bigint index" json:"task_id" readOnly:"true" doc:"Authorized owning issue."`
	Name       string    `xorm:"varchar(240)" json:"name" doc:"Original basename, never used as an untrusted path."`
	Path       string    `xorm:"text" json:"-"`
	SHA256     string    `xorm:"varchar(64)" json:"sha256" readOnly:"true" doc:"Verified content digest."`
	Size       int64     `xorm:"bigint" json:"size" readOnly:"true" doc:"Byte length."`
	UploaderID int64     `xorm:"bigint" json:"uploader_id" readOnly:"true" doc:"Authenticated uploader."`
	Deleted    bool      `xorm:"index" json:"deleted" doc:"Recoverable deletion; bytes are retained for restore."`
	Created    time.Time `xorm:"created" json:"created" readOnly:"true" doc:"Upload time."`
}

func (*OwAttachment) TableName() string { return "ow_attachments" }
func (v *OwAttachment) stored(s *xorm.Session) (*OwAttachment, error) {
	old := &OwAttachment{}
	ok, err := s.ID(v.ID).Get(old)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, OwError{Status: 404, Message: "Attachment not found"}
	}
	return old, nil
}
func (v *OwAttachment) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	old, err := v.stored(s)
	if err != nil {
		return false, 0, err
	}
	return (&OwIssue{TaskID: old.TaskID}).CanRead(s, a)
}
func (v *OwAttachment) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return (&Task{ID: v.TaskID}).CanWrite(s, a)
}
func (v *OwAttachment) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	old, err := v.stored(s)
	if err != nil {
		return false, err
	}
	return (&OwIssue{TaskID: old.TaskID}).CanUpdate(s, a)
}
func (v *OwAttachment) CanDelete(s *xorm.Session, a web.Auth) (bool, error) { return v.CanUpdate(s, a) }
func (v *OwAttachment) ReadOne(s *xorm.Session, _ web.Auth) error {
	old, err := v.stored(s)
	if err == nil {
		*v = *old
	}
	return err
}
func (v *OwAttachment) Delete(s *xorm.Session, a web.Auth) error {
	return OwAttachmentState(s, a, v.ID, true)
}
func OwAttachmentState(s *xorm.Session, a web.Auth, id string, deleted bool) error {
	v := &OwAttachment{ID: id}
	ok, err := v.CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !ok {
		return owDenied()
	}
	if err = v.ReadOne(s, a); err != nil {
		return err
	}
	v.Deleted = deleted
	if _, err = s.ID(id).Cols("deleted").Update(v); err != nil {
		return err
	}
	issue, err := (&OwIssue{TaskID: v.TaskID}).stored(s)
	if err != nil {
		return err
	}
	action := "attachment.restored"
	if deleted {
		action = "attachment.deleted"
	}
	return owActivity(s, a, issue.ProjectID, issue.TaskID, action, v.Name)
}
func OwAttachments(s *xorm.Session, a web.Auth, key string) ([]*OwAttachment, error) {
	issue := &OwIssue{Key: key}
	ok, _, err := issue.CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	if err = issue.ReadOne(s, a); err != nil {
		return nil, err
	}
	rows := []*OwAttachment{}
	err = s.Where("task_id = ?", issue.TaskID).Desc("created").Find(&rows)
	return rows, err
}
func owAttachmentPath(rel string) (string, error) {
	root := filepath.Clean(config.FilesBasePath.GetString())
	if !filepath.IsAbs(root) || !filepath.IsLocal(filepath.FromSlash(rel)) || strings.Contains(rel, "\\") || strings.Contains(rel, ":") {
		return "", owInvalid("Invalid attachment location")
	}
	result := filepath.Join(root, filepath.FromSlash(rel))
	for current := result; current != root; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", owInvalid("Attachment links are not allowed")
		}
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if filepath.Dir(current) == current {
			return "", owInvalid("Invalid attachment root")
		}
	}
	return result, nil
}
func OwUploadAttachment(s *xorm.Session, a web.Auth, key, name string, reader io.Reader) (*OwAttachment, error) {
	issue := &OwIssue{Key: key}
	ok, err := issue.CanUpdate(s, a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, owDenied()
	}
	if err = issue.ReadOne(s, a); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 || strings.ContainsAny(name, "/\\:\x00\r\n") || name == "." || name == ".." {
		return nil, owInvalid("Choose a file with a simple basename")
	}
	random := make([]byte, 24)
	if _, err = rand.Read(random); err != nil {
		return nil, err
	}
	id := hex.EncodeToString(random)
	safe := regexp.MustCompile(`[^A-Za-z0-9._-]+`).ReplaceAllString(name, "_")
	if safe == "" {
		safe = "attachment"
	}
	project, err := owSettings(s, issue.ProjectID)
	if err != nil {
		return nil, err
	}
	rel := project.Key + "/" + issue.Key + "/" + id + "/" + safe
	final, err := owAttachmentPath(rel)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(final), 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(final+".partial", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(final + ".partial")
			_ = os.Remove(final)
		}
	}()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(reader, 64<<20+1))
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if size > 64<<20 {
		return nil, owInvalid("Attachments can be up to 64 MB")
	}
	if err = os.Rename(final+".partial", final); err != nil {
		return nil, err
	}
	attachment := &OwAttachment{ID: id, TaskID: issue.TaskID, Name: name, Path: rel, SHA256: hex.EncodeToString(hash.Sum(nil)), Size: size, UploaderID: a.GetID()}
	if _, err = s.Insert(attachment); err != nil {
		return nil, err
	}
	if err = owActivity(s, a, issue.ProjectID, issue.TaskID, "attachment.uploaded", name); err != nil {
		return nil, err
	}
	cleanup = false
	return attachment, nil
}
func OwOpenAttachment(s *xorm.Session, a web.Auth, id string) (*OwAttachment, *os.File, error) {
	v := &OwAttachment{ID: id}
	ok, _, err := v.CanRead(s, a)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, owDenied()
	}
	if err = v.ReadOne(s, a); err != nil {
		return nil, nil, err
	}
	if v.Deleted {
		return nil, nil, OwError{Status: 404, Message: "This attachment is in the recycle bin"}
	}
	path, err := owAttachmentPath(v.Path)
	if err != nil {
		return nil, nil, err
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, OwError{Status: 503, Message: "Attachment storage is unavailable; retry after reconnecting the drive"}
	}
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil || info.Size() != v.Size {
		_ = f.Close()
		return nil, nil, errors.New("attachment integrity check failed")
	}
	sum := sha256.New()
	if _, err = io.Copy(sum, f); err != nil {
		f.Close()
		return nil, nil, err
	}
	if hex.EncodeToString(sum.Sum(nil)) != v.SHA256 {
		f.Close()
		return nil, nil, errors.New("attachment checksum verification failed")
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, nil, err
	}
	return v, f, nil
}
