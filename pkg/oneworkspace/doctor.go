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
	"encoding/json"
	"errors"
	"flag"
	"net"
	"os"
	"path/filepath"
	"time"
)

func doctorCommand(args []string) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	data := flags.String("data", "", "Dedicated private-data root")
	listen := flags.String("listen", "127.0.0.1:9026", "Loopback listener to inspect")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *data == "" || !filepath.IsAbs(*data) {
		return errors.New("doctor requires --data with an absolute directory")
	}
	report := map[string]any{"diagnostics": Diagnostics(), "dataExists": false, "embeddedDatabaseExists": false, "portInUse": false, "checked": time.Now().UTC()}
	if info, err := os.Stat(*data); err == nil {
		report["dataExists"] = info.IsDir()
	}
	if info, err := os.Stat(filepath.Join(*data, "core", "mytickets.db")); err == nil {
		report["embeddedDatabaseExists"] = info.Size() > 0
	}
	dialer := net.Dialer{Timeout: time.Second}
	conn, err := dialer.DialContext(context.Background(), "tcp", *listen)
	if err == nil {
		report["portInUse"] = true
		conn.Close()
	}
	return json.NewEncoder(os.Stdout).Encode(report)
}
