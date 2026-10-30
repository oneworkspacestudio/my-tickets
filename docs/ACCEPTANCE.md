# My Tickets acceptance

This is an evaluation release. Tests use synthetic isolated NTFS data on the
maintainer's Windows machine. The user explicitly waived the clean-VM requirement;
this report does not certify a clean VM, physical phone, reboot or service account.

| Area | Status | Evidence / scope |
| --- | --- | --- |
| Pinned upstream and original licence enforcement | PASS | Vikunja v2.7.1 at 5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0; AGPL preserved |
| Embedded UI/API/SQLite, one Windows EXE | PASS | Source build, EXE-only developer-host smoke, HTML/JS/API, persistence and PE import inspection |
| Project keys, issue numbering and validation | PASS | Model and native browser tests; server-side workflow, custom field and priority validation |
| Roles and private project isolation | PASS | Model negatives and native cookie/scoped-token checks; data filtered before analytics |
| Issues, comments, files, boards and responsive UI | PASS | Real Chrome headless at 1440x1000 and 375x812, isolated profile; screenshots in release evidence |
| Sprints, epics, milestones and workload | PASS | Persisted planning model and lifecycle/carryover tests; UI uses actual issue data |
| Dashboards, reports and personal layouts | PASS | Persisted definitions and authorized calculations; charts have accessible data tables, CSV and print |
| 10,000 issue search | PASS | 159–170 ms on this developer host; test fixture, not a production benchmark |
| MyAccounts OIDC | PARTIAL | SDK identity mapping/cookie/session protection tested; final managed deployment receipt records live sign-in separately |
| Git links, signature, replay and opt-in transitions | PASS | Signed synthetic webhook model tests, including an already-linked PR later merged |
| Live Gitea repository and retry operations | UNVERIFIED | Explicit repository settings and resync supported; no owner repository used as a fixture |
| My Drive attachment adapter | PARTIAL | Protected shared filesystem source; Tickets authorizes every file operation. A remote Drive API backend is not implemented |
| Scoped API and MCP | PASS | Native API scope, hidden-project denial and MCP initialize tests; complete MCP package tests pass |
| Backup / restore | PASS | Live encrypted snapshot restored to a fresh NTFS directory; both databases, accounts, dashboard and exact file bytes verified |
| Windows SCM service | UNVERIFIED | Native commands implemented; elevated install/start/reboot not claimed |
| Clean Windows machine | UNVERIFIED | Explicitly waived in favour of this machine |
| Installer, upgrade, release and owner installation | UNVERIFIED | Delivery receipt is required; source integration alone is not delivery |

The final release evidence supersedes pending execution rows where it records a
successful check. No synthetic Git test establishes successful live Gitea delivery.

## Limitations

This release does not claim all Jira capabilities. Git resync imports the latest
50 commits; full historical repository, CI/build and Gitea-issue mirroring are not
included. Failed webhook transactions return an error for the sender to retry;
there is no durable inbound dead-letter queue yet. Scheduled due-date automations
are not implemented. Moving an issue into a project with required custom fields
requires resolving the destination's field policy first. Database-backed charts
are available, but advanced interactive chart gestures are not provided.

Standalone local passwords are for independent deployments; managed deployments
use MyAccounts. No public registration is enabled. Attachments remain private by
default when registered as a Drive source, because broad Drive membership cannot
replace ticket project permissions. Public service exposure and Git/Drive access
are configured by the owner, not silently enabled by the app.
