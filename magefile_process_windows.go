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

//go:build mage && windows

// One Workspace extensions, copyright 2026 One Workspace contributors.

package main

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

// setProcessGroup isolates build/test children from the invoking console.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

// taskkill is a Windows build-tool dependency, never an application runtime dependency.
func killProcessGroup(ctx context.Context, cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	systemDirectory, err := windows.GetSystemDirectory()
	if err != nil {
		return err
	}
	systemTool := filepath.Join(systemDirectory, "taskkill.exe")
	killErr := exec.CommandContext(context.WithoutCancel(ctx), systemTool, "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F").Run()
	waitErr := cmd.Wait()
	if killErr != nil {
		return killErr
	}
	var exitError *exec.ExitError
	if errors.As(waitErr, &exitError) {
		return nil // Forced termination is expected for a test server.
	}
	return waitErr
}
