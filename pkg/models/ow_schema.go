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
	"time"
)

type OwState struct {
	Working  bool   `json:"working" doc:"An open state that starts cycle-time tracking when entered."`
	ID       string `json:"id" doc:"Stable workflow state identifier."`
	Name     string `json:"name" doc:"Display name."`
	Category string `json:"category" enum:"open,done,cancelled" doc:"Defines completion calculations independently of the display name."`
	Limit    int    `json:"limit" doc:"Optional work-in-progress limit; zero is unlimited."`
}

type OwTransition struct {
	From      string   `json:"from" doc:"Source state, or * for any source."`
	To        string   `json:"to" doc:"Destination state identifier."`
	Required  []string `json:"required" doc:"Field identifiers required before transition."`
	AdminOnly bool     `json:"admin_only" doc:"Requires project administrator permission."`
}

type OwField struct {
	ID       string   `json:"id" doc:"Stable custom field identifier."`
	Name     string   `json:"name" doc:"Display label."`
	Type     string   `json:"type" enum:"text,number,date,boolean,select,multi-select" doc:"Validation type."`
	Options  []string `json:"options" doc:"Allowed values for selection fields."`
	Required bool     `json:"required" doc:"Whether every issue must supply the field."`
}

type OwProjectSettings struct {
	ProjectID   int64          `xorm:"bigint pk" json:"project_id" doc:"Upstream project ID."`
	Key         string         `xorm:"varchar(10) unique not null" json:"key" doc:"Immutable uppercase project key."`
	Types       []string       `xorm:"json" json:"types" doc:"Allowed issue types."`
	States      []OwState      `xorm:"json" json:"states" doc:"Ordered workflow columns."`
	Transitions []OwTransition `xorm:"json" json:"transitions" doc:"Authorized workflow transitions."`
	Fields      []OwField      `xorm:"json" json:"fields" doc:"Project-scoped typed custom fields."`
	Updated     time.Time      `xorm:"updated" json:"updated" readOnly:"true" doc:"Last settings change."`
}

func (*OwProjectSettings) TableName() string { return "ow_project_settings" }

type OwIssue struct {
	Revision         int64          `xorm:"bigint not null default 1" json:"revision" readOnly:"true" doc:"Monotonic issue revision for conditional edits."`
	ExpectedRevision int64          `xorm:"-" json:"expected_revision,omitempty" doc:"Optional last-seen revision; stale edits return 409 without changing the issue."`
	ExpectedUpdated  string         `xorm:"-" json:"expected_updated,omitempty" doc:"Optional last-seen updated timestamp; a mismatch rejects stale edits with 409."`
	TaskID           int64          `xorm:"bigint pk" json:"task_id" readOnly:"true" doc:"Unique upstream task mapping; never client-assigned."`
	ProjectID        int64          `xorm:"bigint index not null" json:"project_id" doc:"Owning project; use the authorized move action to change it."`
	Key              string         `xorm:"varchar(32) unique not null" json:"key" readOnly:"true" doc:"Permanent human-readable issue identifier."`
	Type             string         `xorm:"varchar(80) index" json:"type" doc:"Project-configured issue type."`
	State            string         `xorm:"varchar(80) index" json:"state" doc:"Workflow state identifier."`
	Severity         string         `xorm:"varchar(80)" json:"severity" doc:"Impact or severity classification."`
	Environment      string         `xorm:"text" json:"environment" doc:"Environment and reproduction context."`
	Component        string         `xorm:"varchar(250) index" json:"component" doc:"Component name."`
	AffectedVersion  string         `xorm:"varchar(120)" json:"affected_version" doc:"Version where the problem occurs."`
	FixVersion       string         `xorm:"varchar(120) index" json:"fix_version" doc:"Target release or milestone."`
	Estimate         float64        `json:"estimate" doc:"Estimated hours."`
	Points           float64        `json:"points" doc:"Story points."`
	SprintID         int64          `xorm:"bigint index" json:"sprint_id" doc:"Sprint in the same project, or zero for backlog."`
	EpicID           int64          `xorm:"bigint index" json:"epic_id" doc:"Epic task ID in the same project, or zero."`
	Position         float64        `xorm:"index" json:"position" doc:"Backlog ordering; editable without drag-and-drop."`
	Custom           map[string]any `xorm:"json" json:"custom" doc:"Values validated against the project's field definitions."`
	Archived         bool           `xorm:"index" json:"archived" doc:"Archived issues are excluded from ordinary lists."`
	ReporterID       int64          `xorm:"bigint index" json:"reporter_id" readOnly:"true" doc:"Authenticated creator."`
	StartedAt        time.Time      `json:"started_at" readOnly:"true" doc:"First entry into active work."`
	CompletedAt      time.Time      `xorm:"index" json:"completed_at" readOnly:"true" doc:"Entry into a completed workflow category; cleared when reopened."`
	Created          time.Time      `xorm:"created" json:"created" readOnly:"true" doc:"Issue creation time."`
	Updated          time.Time      `xorm:"updated index" json:"updated" readOnly:"true" doc:"Last issue change."`
	Summary          string         `xorm:"-" json:"summary" doc:"Upstream task title."`
	Description      string         `xorm:"-" json:"description" doc:"Rich-text upstream description."`
	Priority         int64          `xorm:"-" json:"priority" doc:"Priority, zero to five."`
	DueDate          time.Time      `xorm:"-" json:"due_date" doc:"Optional due date."`
	AssigneeIDs      []int64        `xorm:"-" json:"assignee_ids" doc:"Assignees with access to the project."`
	Task             *Task          `xorm:"-" json:"task" readOnly:"true" doc:"Expanded upstream task, including labels, relations and attachments."`
}

func (*OwIssue) TableName() string { return "ow_issues" }

type OwSprint struct {
	ID        int64     `xorm:"bigint autoincr pk" json:"id" readOnly:"true" doc:"Sprint ID."`
	ProjectID int64     `xorm:"bigint index not null" json:"project_id" doc:"Owning project."`
	Name      string    `xorm:"varchar(250)" json:"name" doc:"Sprint name."`
	Goal      string    `xorm:"text" json:"goal" doc:"Sprint goal."`
	Start     time.Time `json:"start" doc:"Planned start."`
	End       time.Time `json:"end" doc:"Planned end."`
	State     string    `xorm:"varchar(16)" json:"state" readOnly:"true" doc:"planned, active or closed; changed through lifecycle actions."`
	Committed []int64   `xorm:"json" json:"committed" readOnly:"true" doc:"Immutable starting scope, as task IDs."`
	Completed []int64   `xorm:"json" json:"completed" readOnly:"true" doc:"Completed scope at close."`
	Started   time.Time `json:"started" readOnly:"true" doc:"Actual start."`
	Closed    time.Time `json:"closed" readOnly:"true" doc:"Actual completion."`
}

func (*OwSprint) TableName() string { return "ow_sprints" }

type OwWidget struct {
	Chart  string   `json:"chart" enum:",metric,line,bar,donut,table" doc:"Presentation choice; empty uses the calculation's default."`
	ID     string   `json:"id" doc:"Stable widget ID."`
	Kind   string   `json:"kind" doc:"Metric, chart, list or activity widget kind."`
	Title  string   `json:"title" doc:"Widget heading."`
	X      int      `json:"x" doc:"Grid column, zero-based."`
	Y      int      `json:"y" doc:"Grid row, zero-based."`
	W      int      `json:"w" doc:"Width in grid columns."`
	H      int      `json:"h" doc:"Height in grid rows."`
	Filter OwFilter `json:"filter" doc:"Optional widget-level filter overrides."`
}

type OwDashboard struct {
	ID        int64      `xorm:"bigint autoincr pk" json:"id" readOnly:"true" doc:"Dashboard ID."`
	OwnerID   int64      `xorm:"bigint index" json:"owner_id" readOnly:"true" doc:"Authenticated owner."`
	ProjectID int64      `xorm:"bigint index" json:"project_id" doc:"Project dashboard scope, or zero."`
	Name      string     `xorm:"varchar(250)" json:"name" doc:"Dashboard title."`
	Scope     string     `xorm:"varchar(16)" json:"scope" enum:"private,project,global" doc:"Sharing scope; data remains permission-filtered for every viewer."`
	Shares    []int64    `xorm:"json" json:"shares" doc:"Additional user IDs permitted to view a private dashboard."`
	Filter    OwFilter   `xorm:"json" json:"filter" doc:"Dashboard filters."`
	Widgets   []OwWidget `xorm:"json" json:"widgets" doc:"Persistent shared widget definitions and layout."`
	Updated   time.Time  `xorm:"updated" json:"updated" readOnly:"true" doc:"Last modification."`
}

func (*OwDashboard) TableName() string { return "ow_dashboards" }

type OwDashboardLayout struct {
	ID          int64      `xorm:"bigint autoincr pk"`
	DashboardID int64      `xorm:"bigint unique(ow_layout)"`
	UserID      int64      `xorm:"bigint unique(ow_layout)"`
	Widgets     []OwWidget `xorm:"json"`
}

func (*OwDashboardLayout) TableName() string { return "ow_dashboard_layouts" }

type OwActivity struct {
	ID               int64     `xorm:"bigint autoincr pk" json:"id" readOnly:"true" doc:"Activity sequence."`
	ProjectID        int64     `xorm:"bigint index" json:"project_id" doc:"Permission scope."`
	TaskID           int64     `xorm:"bigint index" json:"task_id" doc:"Related task, or zero for project activity."`
	ActorID          int64     `xorm:"bigint index" json:"actor_id" readOnly:"true" doc:"Authenticated actor; never read from ticket text."`
	Action           string    `xorm:"varchar(80)" json:"action" doc:"Recorded action."`
	Detail           string    `xorm:"text" json:"detail" doc:"Human-readable change summary."`
	PreviousState    string    `xorm:"varchar(80)" json:"previous_state" doc:"Previous issue state, when recording an issue change."`
	State            string    `xorm:"varchar(80)" json:"state" doc:"State after the change."`
	PreviousSprintID int64     `xorm:"bigint" json:"previous_sprint_id" doc:"Sprint before the change."`
	SprintID         int64     `xorm:"bigint index" json:"sprint_id" doc:"Sprint after the change."`
	Points           float64   `json:"points" doc:"Estimate at the time of this event."`
	At               time.Time `xorm:"created index" json:"at" readOnly:"true" doc:"Server time."`
}

func (*OwActivity) TableName() string { return "ow_issue_activity" }

type OwSavedView struct {
	ID      int64    `xorm:"bigint autoincr pk" json:"id" readOnly:"true" doc:"Saved view ID."`
	OwnerID int64    `xorm:"bigint index" json:"owner_id" readOnly:"true" doc:"Owner."`
	Name    string   `xorm:"varchar(250)" json:"name" doc:"View name."`
	Shared  bool     `json:"shared" doc:"Shared definition; results are always scoped to each viewer."`
	Filter  OwFilter `xorm:"json" json:"filter" doc:"Saved compound filter."`
}

func (*OwSavedView) TableName() string { return "ow_saved_views" }

type OwFilter struct {
	ProjectID  int64    `json:"project_id" doc:"Zero means all accessible projects."`
	SprintID   int64    `json:"sprint_id" doc:"Zero means any sprint."`
	EpicID     int64    `json:"epic_id" doc:"Zero means any epic."`
	Q          string   `json:"q" doc:"Case-insensitive key, summary and description search."`
	Rules      []OwRule `json:"rules" doc:"Compound filter conditions."`
	Match      string   `json:"match" enum:",and,or" doc:"Rule combination, default and."`
	Sort       string   `json:"sort" doc:"key, summary, priority, due_date, created, updated or position."`
	Descending bool     `json:"descending" doc:"Descending sort order."`
	Archived   bool     `json:"archived" doc:"Include archived issues."`
}

type OwRule struct {
	Field    string `json:"field" doc:"type, state, priority, assignee, reporter, label, due_date, created, updated or custom.fieldID."`
	Operator string `json:"operator" enum:"eq,ne,contains,before,after,empty" doc:"Typed comparison."`
	Value    string `json:"value" doc:"Comparison value; dates use YYYY-MM-DD or RFC3339."`
}
