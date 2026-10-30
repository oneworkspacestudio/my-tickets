# API extensions

All new resource endpoints are under `/api/v2/tickets`:

- Catalog, project creation and workflow/field configuration.
- Issue CRUD/archive, query, compound filters, activity, duplicate, move and bulk.
- Sprint CRUD/start/close with immutable scope snapshots.
- Dashboard CRUD, templates, shared filters, computed data and personal layouts.
- Saved views and real reports.
- Git repository settings, signed/idempotent webhooks, links and recent resync.
- Attachment upload/list/download and recoverable deletion/restoration.

OpenAPI is generated from the current Huma registrations. The TypeScript client
is regenerated with `mage generate:frontend-client`; never edit generated files.
GET/PUT issue operations retain model-level access checks. Bulk operations commit
only if every selected issue authorizes and validates. A moved issue's old key
resolves to its current identity, followed by current-project authorization.

Native cookies, CSRF and account fencing protect browser requests. API `tk_`
credentials retain route scopes and record-level permissions. Workspace forwards
only the narrow Tickets API path when an agent-shaped credential is supplied;
credential presence alone is never authentication.
