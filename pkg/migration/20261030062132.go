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

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type owState20261030062132 struct {
	Working  bool   `json:"working"`
	ID       string `json:"id" doc:"Stable workflow state identifier."`
	Name     string `json:"name" doc:"Display name."`
	Category string `json:"category" enum:"open,done,cancelled" doc:"Defines completion calculations independently of the display name."`
	Limit    int    `json:"limit" doc:"Optional work-in-progress limit; zero is unlimited."`
}

type owTransition20261030062132 struct {
	From      string   `json:"from" doc:"Source state, or * for any source."`
	To        string   `json:"to" doc:"Destination state identifier."`
	Required  []string `json:"required" doc:"Field identifiers required before transition."`
	AdminOnly bool     `json:"admin_only" doc:"Requires project administrator permission."`
}

type owField20261030062132 struct {
	ID       string   `json:"id" doc:"Stable custom field identifier."`
	Name     string   `json:"name" doc:"Display label."`
	Type     string   `json:"type" enum:"text,number,date,boolean,select,multi-select" doc:"Validation type."`
	Options  []string `json:"options" doc:"Allowed values for selection fields."`
	Required bool     `json:"required" doc:"Whether every issue must supply the field."`
}

type owProjectSettings20261030062132 struct {
	ProjectID   int64                        `xorm:"bigint pk" json:"project_id" doc:"Upstream project ID."`
	Key         string                       `xorm:"varchar(10) unique not null" json:"key" doc:"Immutable uppercase project key."`
	Types       []string                     `xorm:"json" json:"types" doc:"Allowed issue types."`
	States      []owState20261030062132      `xorm:"json" json:"states" doc:"Ordered workflow columns."`
	Transitions []owTransition20261030062132 `xorm:"json" json:"transitions" doc:"Authorized workflow transitions."`
	Fields      []owField20261030062132      `xorm:"json" json:"fields" doc:"Project-scoped typed custom fields."`
	Updated     time.Time                    `xorm:"updated" json:"updated" readOnly:"true" doc:"Last settings change."`
}

func (*owProjectSettings20261030062132) TableName() string { return "ow_project_settings" }

type owIssue20261030062132 struct {
	Revision        int64          `xorm:"bigint not null default 1" json:"revision"`
	TaskID          int64          `xorm:"bigint pk" json:"task_id" readOnly:"true" doc:"Unique upstream task mapping; never client-assigned."`
	ProjectID       int64          `xorm:"bigint index not null" json:"project_id" doc:"Owning project; use the authorized move action to change it."`
	Key             string         `xorm:"varchar(32) unique not null" json:"key" readOnly:"true" doc:"Permanent human-readable issue identifier."`
	Type            string         `xorm:"varchar(80) index" json:"type" doc:"Project-configured issue type."`
	State           string         `xorm:"varchar(80) index" json:"state" doc:"Workflow state identifier."`
	Severity        string         `xorm:"varchar(80)" json:"severity" doc:"Impact or severity classification."`
	Environment     string         `xorm:"text" json:"environment" doc:"Environment and reproduction context."`
	Component       string         `xorm:"varchar(250) index" json:"component" doc:"Component name."`
	AffectedVersion string         `xorm:"varchar(120)" json:"affected_version" doc:"Version where the problem occurs."`
	FixVersion      string         `xorm:"varchar(120) index" json:"fix_version" doc:"Target release or milestone."`
	Estimate        float64        `json:"estimate" doc:"Estimated hours."`
	Points          float64        `json:"points" doc:"Story points."`
	SprintID        int64          `xorm:"bigint index" json:"sprint_id" doc:"Sprint in the same project, or zero for backlog."`
	EpicID          int64          `xorm:"bigint index" json:"epic_id" doc:"Epic task ID in the same project, or zero."`
	Position        float64        `xorm:"index" json:"position" doc:"Backlog ordering; editable without drag-and-drop."`
	Custom          map[string]any `xorm:"json" json:"custom" doc:"Values validated against the project's field definitions."`
	Archived        bool           `xorm:"index" json:"archived" doc:"Archived issues are excluded from ordinary lists."`
	ReporterID      int64          `xorm:"bigint index" json:"reporter_id" readOnly:"true" doc:"Authenticated creator."`
	StartedAt       time.Time      `json:"started_at" readOnly:"true" doc:"First entry into active work."`
	CompletedAt     time.Time      `xorm:"index" json:"completed_at" readOnly:"true" doc:"Entry into a completed workflow category; cleared when reopened."`
	Created         time.Time      `xorm:"created" json:"created" readOnly:"true" doc:"Issue creation time."`
	Updated         time.Time      `xorm:"updated index" json:"updated" readOnly:"true" doc:"Last issue change."`
}

func (*owIssue20261030062132) TableName() string { return "ow_issues" }

type owSprint20261030062132 struct {
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

func (*owSprint20261030062132) TableName() string { return "ow_sprints" }

type owWidget20261030062132 struct {
	Chart  string                 `json:"chart"`
	ID     string                 `json:"id" doc:"Stable widget ID."`
	Kind   string                 `json:"kind" doc:"Metric, chart, list or activity widget kind."`
	Title  string                 `json:"title" doc:"Widget heading."`
	X      int                    `json:"x" doc:"Grid column, zero-based."`
	Y      int                    `json:"y" doc:"Grid row, zero-based."`
	W      int                    `json:"w" doc:"Width in grid columns."`
	H      int                    `json:"h" doc:"Height in grid rows."`
	Filter owFilter20261030062132 `json:"filter" doc:"Optional widget-level filter overrides."`
}

type owDashboard20261030062132 struct {
	ID        int64                    `xorm:"bigint autoincr pk" json:"id" readOnly:"true" doc:"Dashboard ID."`
	OwnerID   int64                    `xorm:"bigint index" json:"owner_id" readOnly:"true" doc:"Authenticated owner."`
	ProjectID int64                    `xorm:"bigint index" json:"project_id" doc:"Project dashboard scope, or zero."`
	Name      string                   `xorm:"varchar(250)" json:"name" doc:"Dashboard title."`
	Scope     string                   `xorm:"varchar(16)" json:"scope" enum:"private,project,global" doc:"Sharing scope; data remains permission-filtered for every viewer."`
	Shares    []int64                  `xorm:"json" json:"shares" doc:"Additional user IDs permitted to view a private dashboard."`
	Filter    owFilter20261030062132   `xorm:"json" json:"filter" doc:"Dashboard filters."`
	Widgets   []owWidget20261030062132 `xorm:"json" json:"widgets" doc:"Persistent shared widget definitions and layout."`
	Updated   time.Time                `xorm:"updated" json:"updated" readOnly:"true" doc:"Last modification."`
}

func (*owDashboard20261030062132) TableName() string { return "ow_dashboards" }

type owDashboardLayout20261030062132 struct {
	ID          int64                    `xorm:"bigint autoincr pk"`
	DashboardID int64                    `xorm:"bigint unique(ow_layout)"`
	UserID      int64                    `xorm:"bigint unique(ow_layout)"`
	Widgets     []owWidget20261030062132 `xorm:"json"`
}

func (*owDashboardLayout20261030062132) TableName() string { return "ow_dashboard_layouts" }

type owActivity20261030062132 struct {
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

func (*owActivity20261030062132) TableName() string { return "ow_issue_activity" }

type owSavedView20261030062132 struct {
	ID      int64                  `xorm:"bigint autoincr pk" json:"id" readOnly:"true" doc:"Saved view ID."`
	OwnerID int64                  `xorm:"bigint index" json:"owner_id" readOnly:"true" doc:"Authenticated owner."`
	Name    string                 `xorm:"varchar(250)" json:"name" doc:"View name."`
	Shared  bool                   `json:"shared" doc:"Shared definition; results are always scoped to each viewer."`
	Filter  owFilter20261030062132 `xorm:"json" json:"filter" doc:"Saved compound filter."`
}

func (*owSavedView20261030062132) TableName() string { return "ow_saved_views" }

type owFilter20261030062132 struct {
	ProjectID  int64                  `json:"project_id" doc:"Zero means all accessible projects."`
	SprintID   int64                  `json:"sprint_id" doc:"Zero means any sprint."`
	EpicID     int64                  `json:"epic_id" doc:"Zero means any epic."`
	Q          string                 `json:"q" doc:"Case-insensitive key, summary and description search."`
	Rules      []owRule20261030062132 `json:"rules" doc:"Compound filter conditions."`
	Match      string                 `json:"match" enum:",and,or" doc:"Rule combination, default and."`
	Sort       string                 `json:"sort" doc:"key, summary, priority, due_date, created, updated or position."`
	Descending bool                   `json:"descending" doc:"Descending sort order."`
	Archived   bool                   `json:"archived" doc:"Include archived issues."`
}

type owRule20261030062132 struct {
	Field    string `json:"field" doc:"type, state, priority, assignee, reporter, label, due_date, created, updated or custom.fieldID."`
	Operator string `json:"operator" enum:"eq,ne,contains,before,after,empty" doc:"Typed comparison."`
	Value    string `json:"value" doc:"Comparison value; dates use YYYY-MM-DD or RFC3339."`
}

type owGitSettings20261030062132 struct {
	ProjectID    int64                     `xorm:"bigint pk" json:"project_id" doc:"Bound ticket project."`
	URL          string                    `xorm:"varchar(500)" json:"url" doc:"Explicit trusted Gitea server base URL."`
	Repository   string                    `xorm:"varchar(250)" json:"repository" doc:"Exact repository owner/name."`
	Token        string                    `xorm:"-" json:"token,omitempty" writeOnly:"true" doc:"Service token; empty preserves the current encrypted token."`
	Secret       string                    `xorm:"-" json:"secret,omitempty" writeOnly:"true" doc:"Webhook HMAC secret; empty preserves the current secret."`
	TokenCipher  string                    `xorm:"text" json:"-"`
	SecretCipher string                    `xorm:"text" json:"-"`
	Rules        []owGitRule20261030062132 `xorm:"json" json:"rules" doc:"Explicit opt-in workflow automation, disabled when empty."`
	Configured   bool                      `xorm:"-" json:"configured" readOnly:"true" doc:"Whether credentials and webhook secret exist."`
}

type owGitRule20261030062132 struct {
	Event string `json:"event" doc:"push or pull_request.closed."`
	State string `json:"state" doc:"Validated destination workflow state."`
}

type owGitLink20261030062132 struct {
	Repository string `xorm:"varchar(250)"`
	Revision   string `xorm:"varchar(100)"`
	Author     string `xorm:"varchar(250)"`
	State      string `xorm:"varchar(30)"`
	Occurred   time.Time
	URLHash    string    `xorm:"varchar(64) unique(ow_git_link)" json:"-"`
	ID         int64     `xorm:"bigint pk autoincr" json:"id" readOnly:"true" doc:"Link ID."`
	TaskID     int64     `xorm:"bigint index unique(ow_git_link)" json:"task_id" doc:"Related issue task."`
	URL        string    `xorm:"varchar(1000)" json:"url" doc:"Validated link on the configured Git server."`
	Kind       string    `xorm:"varchar(80)" json:"kind" doc:"Commit or pull request."`
	Title      string    `xorm:"text" json:"title" doc:"External title; always treated as untrusted text."`
	At         time.Time `xorm:"created" json:"at" readOnly:"true" doc:"First association time."`
}

type owGitDelivery20261030062132 struct {
	ID        string    `xorm:"varchar(64) pk"`
	ProjectID int64     `xorm:"bigint index"`
	At        time.Time `xorm:"created"`
}

func (*owGitSettings20261030062132) TableName() string { return "ow_git_settings" }

func (*owGitLink20261030062132) TableName() string { return "ow_git_links" }

func (*owGitDelivery20261030062132) TableName() string { return "ow_git_deliveries" }

type owAttachment20261030062132 struct {
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

func (*owAttachment20261030062132) TableName() string { return "ow_attachments" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261030062132",
		Description: "Add namespaced One Workspace issue, sprint, workflow and dashboard tables",
		Migrate: func(tx *xorm.Engine) error {
			// These are new tables only; no upstream table/index is synchronized.
			//nolint:forbidigo // Creates only new ow_* tables; no upstream schema is synchronized.
			return tx.Sync2(&owProjectSettings20261030062132{}, &owIssue20261030062132{}, &owSprint20261030062132{}, &owDashboard20261030062132{}, &owDashboardLayout20261030062132{}, &owActivity20261030062132{}, &owSavedView20261030062132{}, &owGitSettings20261030062132{}, &owGitLink20261030062132{}, &owGitDelivery20261030062132{}, &owAttachment20261030062132{})
		},
		Rollback: func(_ *xorm.Engine) error { return nil },
	})
}
