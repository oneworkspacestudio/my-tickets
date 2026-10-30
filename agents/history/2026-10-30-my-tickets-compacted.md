# My Tickets implementation and delivery handoff — 2026-10-30

User authorizes brief implementation, managed availability, source push, release
and install on this PC. VM explicitly waived. No Actions, no owner test fixtures.
Fork pinned Vikunja2.7.1 upstream5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0;
AGPL/licence enforcement retained; original history and upstream remote retained.
Inherited workflows archived inactive because OAuth lacks workflow scope.

## Published app
Public v0.1.0-preview.1 targets8650cb8bacbef64c7a6a6c4b86bcc12f58dc5b20.
One Windows x64 executable, embedded full upstream/custom frontend/API/SQLite.
Native in-process apphost SDK438b9cfe, no helper executable or private worker port.
EXEf112e9635c179ac923458e75de1c6e0a5f5c2b03ca20a7c81fb39d04446e3b54,
93,282,816 bytes; component ZIP770b37f590caa23c6d1a732156658556fd9f080b25e7819d6fe6a7fc3a11e536,
30,522,844 bytes.23 release assets digest/size verified; anonymous ZIP verified.
Includes unchanged signed Android10 companion from Workspace104 with source,
original provenance/notices. Workspace105 rebuilds source-bound signed companion.

## Acceptance
Native Chrome desktop1440x1000/mobile375x812 including issue detail PASS; API
scope/project isolation/MCP initialize,16 concurrent unique issue keys, stale edit409,
attachment lifecycle/logout and encrypted two-database recovery PASS bound to exact
EXE hash.10K search159-170ms. Model/MCP tests, frontend type/lint/style/build,
backend lint0 issues PASS. PE imports OS UCRT/KERNEL32 only.
Real managed fresh105 and normal104 upgrade PASS on this PC with synthetic NTFS
private data. MyAccounts once then Tickets/no second password; issue create,
desktop/mobile static hashes verified. Independent synthetic Tickets2 update and
normal rollback retained SUITE-1, user files and manager/accounts binary hashes.
Synthetic2 never published. Evidence sanitized in release archives/receipts.

Regression fixes: SQLite constructor ignored pool config; explicitly bound actual
native sql.DB connections1. SDK strips unverified email correctly: opt in to absent
email only for authenticated stable external issuer/sub, no fabricated email/local
validation weakening. Unexpected core failures503 instead of401 reauth loop.
CORS uses origin not mounted path; phone rich-text toolbar wraps; already-linked
merged PR still applies explicit transition. All covered by tests.

## Preview limits
Protected shared-filesystem Drive adapter, remote Drive API absent. Webhook sender
retry only, no durable queue. Scheduled due-date automation absent. Git resync
latest50 commits. Destination required custom fields may block a move until policy
resolved. Live owner Gitea, elevated SCM, physical phone/reboot unverified. See
ACCEPTANCE.md. This is evaluation preview, no complete production certification.

## Remaining managed delivery
Workspace source integration+public app channel pushed. Normal published-pin
suite105 build, required signed Android suite publication, channel promotion and
protected owner104-to-105 install still need receipt. Exact accepted app binary
is already final; do not overwrite immutable release assets for doc corrections.
Task caches/evidence I:/Projects/tooling/my-tickets-20261030, small NTFS C fixtures;
frontend task dependencies and finished synthetic C fixtures removed after checks.
No owner data used as development fixtures. Avoid global lint --fix on CRLF source;
use read-only Mage lint. GOWORKoff release, I large Go temp, C only native CGO temp.
