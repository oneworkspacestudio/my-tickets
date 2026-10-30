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

import "testing"

func TestNativeWebOriginExcludesMountedPath(t *testing.T) {
	for _, input := range []struct{ value, want string }{
		{value: "https://workspace.example.test/tickets", want: "https://workspace.example.test"},
		{value: "http://localhost:19580/tickets/", want: "http://localhost:19580"},
	} {
		got, err := nativeWebOrigin(input.value)
		if err != nil || got != input.want {
			t.Fatal(got, err)
		}
	}
	for _, invalid := range []string{"/tickets", "javascript:alert(1)", "https://user:password@example.test/tickets"} {
		if _, err := nativeWebOrigin(invalid); err == nil {
			t.Fatal("invalid origin accepted")
		}
	}
}
