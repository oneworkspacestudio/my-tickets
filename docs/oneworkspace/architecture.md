# My Tickets architecture and proof gate

## Baseline and runtime
Vikunja v2.7.1 is pinned in UPSTREAM.md. Its Go server embeds the entire frontend
through frontend/embed.go. The official Windows reference has only KERNEL32.DLL
and msvcrt.dll in its PE import table. An isolated NTFS, EXE-only starting folder
served UI/API, wrote SQLite records and retained them after forced termination.
That reference test is not a self-build proof or a clean Windows VM test.
The stock self-build uses Go1.27.3, upstream pnpm11.27.0/lockfile, Node26.8.2, and
Zig0.15.2 for the existing mattn/go-sqlite3 CGo binding. Static C linking must be
validated by inspecting and running the resulting EXE, not assumed from flags.

## Extension boundaries (planned; not implemented)
- Reuse Vikunja task/project/user/permission/attachment/relations primitives and
  API v2 Huma/generated client. No new API v1 endpoints.
- Add migration-managed ow_* tables for project keys, issue sequences, fields,
  workflows, sprints, dashboard layouts, Git links and attachment storage index.
  An upstream task maps to exactly one issue, with a unique task reference.
- Allocate project-local issue numbers transactionally; issue IDs and project
  keys remain stable. Authorization is applied before row/aggregate queries.
- Reuse audited upstream OIDC issuer/sub mapping where it meets the requirement;
  verify browser token lifetime/storage and refresh cookies before adapting it.
  Workspace production mode disables public registration/local password login.
- Preserve all existing upstream license checks. Implement the requested missing
  ticket/dashboard functionality in independent namespaced AGPL extensions;
  do not claim or bypass optional upstream Pro features.
- Configure Gitea and Drive through inspected ecosystem contracts and isolated
  test services. External integration outages must not prevent ticket/search use.
- One process: core HTTP server, SQLite, frontend and native SCM/CLI adapters all
  remain in the same executable. Mutable DB/files/logs/config remain outside it.

## Current acceptance boundary
The source-built baseline passed developer-host HTML/JS/API/SQLite and crash
persistence checks. Clean-VM proof, My Tickets extensions and ecosystem integration
are not accepted yet. The VM question is pending. No owner service has been used
as a development fixture, no tunnel is created, and no My Tickets release/install
has been performed.


## Accepted continuation architecture

The user explicitly waived clean-VM verification in favor of tests on this machine.
Use synthetic NTFS roots. Implement custom Vue UI beside the complete embedded
upstream UI, not a second program. Reuse upstream task/project/index/permission,
comment/relation/notification primitives. New missing concepts use ow_* migrations.

Use the current public Workspace apphost SDK in the same executable for secure
cookie OIDC, owner binding, CSRF, private folder ownership and lifecycle. Mount
the core Echo handler directly inside the SDK HTTP server: no child process or
second listener. A server-side gateway maps verified issuer/sub to an upstream
user and injects a short-lived upstream JWT. Browser resource calls use SDK session
cookies and CSRF, no long-lived bearer in browser storage. Scoped tk_ agent tokens
retain upstream token authorization and project permission checks. UI request data
uses generated API v2 clients/query modules. Extend the SDK with an optional custom
UI handler, preserving existing consumers. Native runtime keeps all mutable data
under its managed private root; backup must snapshot BOTH SQLite databases safely.

Managed definition: tickets, /tickets, foundation=true, general default, app update/
recovery, one MyTickets.exe. Fresh integration clone I:/Projects/workspace-my-tickets
starts at current main05ded8d. No generated source/editing API client by hand.
