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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOwTenThousandIssueSearch(t *testing.T) {
	s, a, p := owTestData(t)
	tasks := make([]*Task, 10000)
	issues := make([]*OwIssue, 10000)
	for n := range tasks {
		id := int64(100000 + n)
		tasks[n] = &Task{ID: id, ProjectID: p.ProjectID, Index: int64(n + 1), Title: fmt.Sprintf("launch-%05d", n)}
		issues[n] = &OwIssue{TaskID: id, ProjectID: p.ProjectID, Key: fmt.Sprintf("%s-%d", p.Key, n+1), Type: "Task", State: "backlog", Custom: map[string]any{}}
	}
	for start := 0; start < len(tasks); start += 100 {
		_, err := s.Insert(tasks[start : start+100])
		require.NoError(t, err)
		_, err = s.Insert(issues[start : start+100])
		require.NoError(t, err)
	}
	start := time.Now()
	rows, err := OwQuery(s, a, OwFilter{ProjectID: p.ProjectID, Q: "launch-09999"})
	elapsed := time.Since(start)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	t.Logf("10,000 issue search on this developer host: %s", elapsed)
}
