# Architecture

My Tickets keeps Vikunja's API, authorization, project/task identity, comments,
relations and notification models. Extensions live in `pkg/models/ow_*`,
`pkg/routes/api/v2/ow_*`, `pkg/oneworkspace` and `frontend/src/oneworkspace`.

One native process owns one loopback HTTP listener, two embedded SQLite engines
and the embedded browser interface. The Workspace apphost SDK owns private-data
locking/ACLs, OIDC with PKCE, secure browser cookies, CSRF and encrypted backups.
Its authenticated identity maps by **issuer and subject** to a Vikunja user.
The verified Workspace owner is the core administrator. A short server-side JWT
bridges the existing core handler; it is never returned to the browser.

Agent `tk_` credentials use the existing API scope and model permission checks.
There are no native-client bearer credentials in localStorage. Browser requests
carry an account fence and CSRF header. Switching accounts cancels and clears
queries. WebSocket upgrades are disabled; this client refreshes lists/dashboard
calculations over revocable HTTP sessions.

`ow_issues.task_id` is a unique mapping to the core task. Task title, description,
priority, due date, assignees and atomic project-local counters stay upstream-owned.
Namespaced metadata owns workflow, type, sprint, epic and typed custom values.
Generic task writes cannot bypass ticket workflow validation. Moves use the core
counter/alias mechanism and reauthorize both projects; old keys resolve to the
current issue and are checked against its current permissions.

Metrics and reports share the authorized query cohort. Historical charts replay
server-recorded state/scope events. Sprint start and close snapshots are retained.
Personal widget positions are separate from shared definitions; only a dashboard
owner changes shared definitions. Sharing never grants access to hidden issues.

Default standalone data is in the user's application-data folder, never beside
the executable. Managed data is `${DATA}/tickets`: SDK state in `state.sqlite`,
core state in `core/mytickets.db`, signing/encryption key in `core.key`, temporary
work in `tmp`, and attachment bytes in `core/attachments`. No mutable data belongs
in the replaceable program directory.

Backups freeze native requests and snapshot each SQLite engine through its
supported online snapshot mechanism. The core uses the SQLite backup API, not
copies of live WAL/database files. Restore uses an application-owned allowlist,
requires all declared files, verifies archive/object integrity and rejects unsafe
paths and existing destinations. Workspace updates retain attachment bytes while
backing up the database/settings needed for schema recovery.
