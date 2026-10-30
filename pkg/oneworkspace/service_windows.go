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

//go:build windows

// One Workspace extensions, copyright 2026 One Workspace contributors.

package oneworkspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "OneWorkspaceMyTickets"

type serviceHandler struct{}

func (serviceHandler) Execute(_ []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runContext = ctx
	status <- svc.Status{State: svc.StartPending}
	done := make(chan error, 1)
	go func() { done <- serve() }()
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case err := <-done:
			if err != nil {
				return true, 1
			}
			return false, 0
		case r := <-requests:
			switch r.Cmd {
			case svc.Pause, svc.Continue, svc.ParamChange, svc.NetBindAdd, svc.NetBindRemove, svc.NetBindEnable, svc.NetBindDisable, svc.DeviceEvent, svc.HardwareProfileChange, svc.PowerEvent, svc.SessionChange, svc.PreShutdown:
				status <- r.CurrentStatus
			case svc.Interrogate:
				status <- r.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending}
				cancel()
				select {
				case err := <-done:
					if err != nil {
						return true, 1
					}
					return false, 0
				case <-time.After(25 * time.Second):
					return true, 2
				}
			}
		}
	}
}
func nativeService() (bool, error) {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return true, err
	}
	if isService {
		return true, svc.Run(serviceName, serviceHandler{})
	}
	if len(os.Args) < 2 || os.Args[1] != "service" {
		return false, nil
	}
	if len(os.Args) < 3 {
		return true, errors.New("use service install, start, stop, status or uninstall")
	}
	manager, err := mgr.Connect()
	if err != nil {
		return true, fmt.Errorf("open Windows service manager: %w", err)
	}
	defer func() { _ = manager.Disconnect() }()
	action := os.Args[2]
	if action == "install" {
		data := ""
		for i := 3; i+1 < len(os.Args); i++ {
			if os.Args[i] == "--data" || os.Args[i] == "--data-dir" {
				data = os.Args[i+1]
			}
		}
		if !filepath.IsAbs(data) {
			return true, errors.New("service install requires --data with a dedicated absolute NTFS path")
		}
		exe, err := os.Executable()
		if err != nil {
			return true, err
		}
		args := append([]string{"--no-browser"}, os.Args[3:]...)
		service, err := manager.CreateService(serviceName, exe, mgr.Config{DisplayName: "One Workspace My Tickets", Description: "Native issue tracking, API, dashboards and embedded SQLite.", StartType: mgr.StartAutomatic}, args...)
		if err != nil {
			return true, err
		}
		defer service.Close()
		return true, service.SetRecoveryActions([]mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: 5 * time.Second}, {Type: mgr.ServiceRestart, Delay: 30 * time.Second}}, 300)
	}
	service, err := manager.OpenService(serviceName)
	if err != nil {
		return true, err
	}
	defer service.Close()
	switch action {
	case "start":
		return true, service.Start()
	case "stop":
		_, err = service.Control(svc.Stop)
		return true, err
	case "uninstall":
		status, err := service.Query()
		if err != nil {
			return true, err
		}
		if status.State != svc.Stopped {
			return true, errors.New("stop My Tickets before uninstalling the service; private data is preserved")
		}
		return true, service.Delete()
	case "status":
		status, err := service.Query()
		if err == nil {
			fmt.Println(status.State)
		}
		return true, err
	default:
		return true, errors.New("unknown service operation")
	}
}
