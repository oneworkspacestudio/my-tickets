# 2026-10-30 My Tickets

User authorizes implementation of the supplied My Tickets brief, new oneworkspacestudio
repository, releases and installation. Single Windows executable gate precedes heavy
UI work; no companion server/database/runtime. Source pinned to Vikunja v2.7.1,
commit5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0 (tag object91107279377ecf45f652a12589a687819f8d8e90).
Direct code.vikunja.io clone failed403; official site links go-vikunja GitHub org.
I:/Projects/my-tickets is source branch codex/my-tickets, upstream remote retained.
C:/Projects/.cache/my-tickets-20261030/upstream is dedicated NTFS stock build copy;
I:/Projects/tooling/my-tickets-20261030 holds public artifacts/dependency caches.
No owner installations used as fixtures. No app/repository/release/install yet.
Stock upstream imports mattn/go-sqlite3 (CGo needed); frontend requires Node>=24,
pnpm11.27.0; existing Node26.8.2 and Go1.27.3 available. Existing Zig0.15.2 located.
Windows Sandbox, Hyper-V management and VMware/VirtualBox commands not found;
hypervisor is present. Async user question asks for a clean VM or permission to
keep clean-VM acceptance pending while using isolated local tests; no answer yet.
Do not silently waive the brief's clean-machine gate. Track checklist in
 docs/oneworkspace/IMPLEMENTATION.md. Preserve upstream AGPL and license checks;
no GitHub Actions. Use repository's .agents skills for models/migrations/API routes.

Stock frontend completed (361 files, archived on I, SHA f2d8a5cd6412df11f24467c6cea11be969c8e4619873413f8a3946985950c194). Pnpm11 needs NTFS project/store links; npm cache on exFAT caused ECOMPROMISED and store links caused EISDIR. Active pnpm/npm caches moved C; after frontend success archived dist, verified link destinations inside task root and removed only owned node_modules/pnpm-store. Separate official Windows reference ZIP verified sha69f07c3b6515b621c347cc23791c6b4bbf71cca0a576e74487d74d22317c6718. Reference46MB exe imports KERNEL32.DLL/msvcrt.dll only. First readiness harness with localhost/2s timeout falsely timed out although manual health200; its orphan test PID27996 was identified by exact path and stopped. Fresh fixture2 uses127.0.0.1/10s, succeeds EXE-only UI/API/SQLite project+task and forced-kill restart persistence. All reference processes stopped; verify port19456 closed. Evidence explicitly reference/not self-built/not cleanVM. Own build and broad module download still active; Zig compiles cold MinGW CRT cache on I. Own C task cache small; other concurrent work also consumes C, never delete others' files. No heavy feature implementation until proof gate/explicit user waiver.


## Completed Phase 0 developer-host work (still gated)

- Stock C checkout at pinned commit built source EXE with complete frontend;
  direct Go build was required because stock Mage used Unix syscall fields on Windows.
  EXE 86,813,696 bytes, SHA537e158c9b555878eeca190cbbda8f954b5886d33d26b0f5171cd26b96937772.
  Artifact remains under I:/Projects/tooling/my-tickets-20261030.
- Source EXE-only smoke PASS: embedded HTML/JS, SQLite project/task, anonymous
  denial, force-kill/restart persistence. PATH only Windows system directories.
  It is an unchanged upstream baseline, NOT the completed custom app.
- PE imports Windows KERNEL32/UCRT API sets only, no delay imports. Developer host
  injects signed Citrix ctxapclient64.dll into both the app and test driver.
  First strict module test failed; developer-host functional test records REVIEW_REQUIRED
  and singleExecutableAcceptance UNVERIFIED. No security setting disabled. Clean VM
  still absent and unanswered; never silently accept developer host as a clean VM.
- Fork source fixes Mage process helpers with mage/platform build tags; Windows
  build-tool taskkill is excluded from normal runtime. Mage -l compiled successfully.
  Added mage test:packaging to constrain relevant package checks.
- Broad original filter passed three desired tests but failed files TestMain:
  embed.FS paths used filepath.Join backslashes. It was stopped after the known
  failure. Added TestInitFixturesNamedTables, observed same error before fix,
  changed pkg/db/test_fixtures.go to path.Join, then packaging target passed
  pkg/db/config/doctor/files. Full suite NOT run/accepted. No models/migrations/API/UI
  extensions yet. Baseline EXE predates these build/test compatibility patches.
- mage lint:fix attempted but golangci-lint2.14.0 unavailable; lint unverified.
  No commit/push/repository/release/owner installation/tunnel change performed.
- I:/ temp failed Go linker chtimes on exFAT; use dedicated NTFS scratch for Go
  native linking as well as npm/pnpm. Large caches/artifacts remain I. All synthetic
  servers stopped. Runtime fixtures and obsolete C Go cache were removed after
  copying sanitized JSON. Remaining dedicated C task cache/stock clone cleanup
  started after all own Go/Node processes ended; verify its completion.
- Evidence/report under docs/oneworkspace/, docs/windows-single-exe.md.
  scripts/verify-single-exe.ps1 records environment declaration, does not attest VM;
  inspect-pe-imports.py uses only Python stdlib. Do not commit raw logs/databases.

Resume: resolve explicit clean-Windows gate (VM or human change to acceptance),
then implement the actual brief phases. I source branch codex/my-tickets remains
uncommitted atop upstream stable. Fetch/unshallow as needed for full fork publication,
create oneworkspacestudio/my-tickets and disable Actions before first push. Future
Workspace integration must fresh-fetch its concurrently modified repo and use its
normal signed-APK/channel publisher, not an assumed old release number. No heavy
new UI/modules until gate is resolved. Source C clone can be rehydrated from pinned
upstream if needed; all stock frontend files preserved on I in archive and ignored dist.

Cleanup completion confirmed: dedicated C:/Projects/.cache/my-tickets-20261030 no longer exists; artifacts/source/evidence remain on I. All known task test/build sessions ended and fixture port19456 is closed.

## User gate change
User explicitly waived VM testing and instructed testing on this machine, managed Workspace availability, push and release. Continue all remaining features and managed delivery with isolated synthetic NTFS test data. Clean VM remains unperformed by user decision, no longer a blocker. Use current APP-LIFECYCLE; older Workspace checkouts lack the new managed registry, so a fresh integration clone is being obtained on I.


## Active continuation implementation checkpoint

User explicitly removed VM as gate. Continue until full app, managed integration,
push/release/install; do not end at baseline again. Current work is still uncommitted.

Implemented in I:/Projects/my-tickets:
- pkg/models/ow_schema.go: namespaced project config, issue metadata, sprint,
  dashboard/layout, saved view, activity, filter/rules. Existing task/project rows
  remain authoritative for ordinary fields; existing atomic Task.Index reused.
- ow_issues.go: configuration validation, immutable keys, typed fields, workflow
  validation/required fields/admin transitions/WIP limits, issue CRUD/archive,
  authorized filtering/sort/pagination/hydration. Initial key-only lookup bug fixed.
- ow_planning.go, ow_analytics.go, ow_history.go, ow_actions.go: sprint start/close
  scope/carryover; dashboard CRUD/shares/personal layouts/templates; real metrics
  and replayed history charts; catalog, activity, duplicate and atomic bulk actions.
- Registered tables in models.GetTables. Generated migration skeleton through Mage,
  then added versioned schema snapshot pkg/migration/20261030062132.go. New tables
  only Sync2. Later schema changes must be reflected in this unreleased migration.
- API v2 resources/actions under pkg/routes/api/v2/ow_*.go. Follow existing pointer
  singleBody and conditionalReadResponse conventions. All model logic authorizes.
- mage test:tickets runs TestOw*. Three model tests PASS in tickets-phase1-tests-4.log:
  issue lifecycle/permissions/hidden metrics, failed-transition atomic validation,
  sprint carryover and shared dashboard isolation. No API/browser acceptance yet.
- pkg/oneworkspace/app.go implements SDK/native single HTTP server. SDK secure-cookie
  Require wraps core Echo handler directly (no companion process/private listener).
  Stable SDK issuer/sub maps to upstream user and verified SDK owner maps IsAdmin;
  gateway mints short JWT per request, never returns it to browser. Core session IDs
  cached by SDK cookie hash, API tk_ scopes preserved. WebSocket upgrades denied
  (custom UI will poll); standard SDK logout revokes cookie sessions immediately.
  Core SQLite/attachments under private/core; core.key is protected private data.
  Snapshot uses mattn online backup API, SDK extra snapshot hooks, no raw WAL copy.
  Needs audit/tests and more lifecycle/doctor/service/control implementation.
- cmd.Execute delegates default/serve/flags to native Main; original CLI remains
  available under explicit `upstream`. Native MyTickets-dev.exe not built yet.
  No My Tickets frontend implemented yet; native UI intentionally503 until built.
- go.mod now requires Workspace v0.3.0-preview.104 (Go version bumped1.27.3 by tidy).
  This is a TEMP dependency version: new SDK fields exist only in local integration
  checkout. Pin a published new SDK source commit before final app source/release.

Workspace integration clone I:/Projects/workspace-my-tickets at fresh05ded8d main:
- Read AGENTS, central docs/APP-LIFECYCLE and managed-distribution104 handoff.
- Added managed registry tickets: /tickets, org/my-tickets, OIDC foundation true,
  app recovery/update, general default. Reserved public9026 (no existing source use).
- SDK Host optional UIHandler; SnapshotFiles/BackupPaths; Config.RestoreFiles,
  RestoreWithFiles allowlist; CloseHooks before private lock close. Added extension
  tests. First SDK extension run PASS (sdk-extension-tests.log); CloseHooks added
  afterward, rerun once all SDK changes final. Existing APK publication mandatory.
- foundationApp now derives registry, seamless profile includes tickets/api/me and
  native auth/login/logout/cookie. No catalog/build/runtime receipt/update/pins,
  branding or installation implemented yet. No commits/pushes or org app repo yet.
- Need advance installerIntegration4->5/minimum manager105 because old104 cannot
  launch/provision native Tickets. Implement/test supported old-controller refusal,
  fresh setup, normal upgrade, independent update/recovery, preserving all33 old apps.
  Managed build source currently seeds103 and adds Remote5; extend pinned-component
  builder to add Tickets (not a local private seed). Current latest release104.

Tooling/cache notes:
- I:/Projects/tooling/my-tickets-20261030 still holds baseline exe + source/evidence,
  gocache/gomodcache/zig-cache/Mage cache/task scripts/logs. Use NTFS C scratch for
  native linker temp, npm cache and pnpm project/store; I exFAT chtimes/link issues.
- Task go.work is at tooling root with RELATIVE ../../my-tickets and
  ../../workspace-my-tickets. Absolute I:/ paths caused `current directory not in
  workspace` for Go build on Windows. Explicit GOWORK needed in task scripts.
  Temporary accidental blank repo go.work was MOVED to tooling and corrected.
- SDK module tidy completed exit0; first native build failed due absolute go.work.
  Retried explicit go.exe -C I:/Projects/my-tickets with relative workfile:
  ACTIVE exec session43132, log native-dev-build-2.log. Check it.
- C:/Projects/.cache/my-tickets-20261030 recreated for isolated scratch. Frontend
  source copy now on C:/.../frontend (robocopy, excludes node_modules/dist/.git).
  ACTIVE npm/pnpm install session39292, log frontend-install-continuation.log.
  Pinned npm exec pnpm11.27.0, frozen lock, C pnpm-store hardlinks, no browser download.
- Mage Generate.FrontendClient being patched to accept VIKUNJA_FRONTEND_DIR and
  VIKUNJA_PNPM_SCRIPT (node pnpm.cjs), with Windows cmd fallback. That command was
  ACTIVE session84556; inspect completion/code before using. Generate via Mage into
  C frontend then copy GENERATED results to I; never hand edit generated client.
- golangci-lint2.14.0 official Windows archive downloaded and checksummed; executable
  under tooling/golangci-lint/golangci-lint-2.14.0-windows-amd64/golangci-lint.exe.
  Put on PATH and run mage lint:fix before commits. New files currently use SPDX
  headers; may need upstream goheader template alignment and lint/style cleanup.

Next: finish native build/API smoke in fresh NTFS SDK fixture, generate client,
implement custom Vue My Tickets entry beside full embedded stock frontend, hook UI
and native credentials/CSRF to generated SDK queries (no long browser bearer).
Implement missing move/aliases, upstream-key/workflow bypass guards, full permission
coverage/migration/security/concurrency tests, integrations (Gitea/Drive adapters,
signed webhooks/retry, attachments), scoped MCP exposure, operations/service/doctor,
all required release/docs/branding and managed lifecycle delivery. Fix any flaws in
new code rather than claiming completeness. Source APIs/metrics not yet production
accepted. No owner data/test/live configuration mutation has occurred.

## Continuation checkpoint: native app and browser flow now functional

- VM gate remains explicitly waived by user. Continue to push/release/install;
  none of those delivery stages is complete yet. Do not stop at a baseline.
- Source repo `I:/Projects/my-tickets`, branch `codex/my-tickets`, full upstream
  history now fetched. Newly created public org repo is
  `https://github.com/oneworkspacestudio/my-tickets`; origin configured, Actions
  disabled before any push. Verified active org administrator fahimc. No source
  commit/push/release yet. Workspace clone remains `I:/Projects/workspace-my-tickets`.
- Custom Vue UI now exists: TicketsApp, IssueDetail, IssueFields, DashboardView,
  DashboardSettings, ProjectSettings/Access, GitSettings, IssueAttachments,
  RichText, filters/charts, native cookie transport and generated query client.
  Strict typecheck and frontend ESLint/stylelint pass (four known warnings,
  sanitized v-html and main entrypoint). Source uses shared blue/purple/mint icon.
- Real isolated native Chrome flow PASS in tooling `browser-evidence-4`:
  sign-in, project/issue create, workflow edit, comment, attachment, list, board,
  actual dashboard totals, saved personal position, desktop and375px no overflow,
  no browser bearer credentials or console errors. Screenshots reviewed. Browser
  profile/data are synthetic; no owner install or physical-phone claim.
  Entrypoint `mage test:e2e tickets`, scripts/test-native-ui.ps1/.mjs. Latest log
  native-ui-test-7.log. Known test-driver fixes: ignore legacy/unconfigured Inbox
  when choosing ticket project; scope edit controls to visible issue region.
- Nine `TestOw*` model tests passed through Mage (tickets-phase4-tests.log), incl
  lifecycle/permissions, failed validation, sprints/dashboard isolation, generic
  API bypass denial, cross-project move/old-key alias, signed/idempotent Git webhook,
  attachments delete/restore/IDOR/traversal, numeric filter. Broader coverage needed.
- Git adapters added ow_git.go/v2 ow_git.go: encrypted settings, exact repository,
  trusted origin links, HMAC+API credential, replay protection, explicit transition
  rules, recent commit resync. Just refactored resync to fetch outside DB transaction
  (`OwFetchGit` then `OwApplyGitBatch`); NOT retested after this edit. Added hidden
  URLHash for portable unique index; migration snapshot updated. Need model/mock
  tests and live status accurate. Branch/tag/release and PR history resync incomplete.
- Attachments ow_attachments.go/v2:64MB stream/hash/atomic file, original metadata,
  authorization, safe namespace, soft delete/restore, checksum on download. Native
  temp now private data/tmp. Manager registers disabled-by-default My Tickets Drive
  source at tickets/core/attachments so raw Drive access cannot bypass project roles.
  This is a shared filesystem adapter, not remote Drive API upload; docs disclose.
- Native Windows SCM service control, serve/default app, doctor, core online SQLite
  backup, UI direct serving, native identity/status checks and forwarded-header
  scrubbing implemented. CLI/default data is user application-data, never program
  folder. SCM/elevated/reboot checks unverified. Core max SQLite connections1.
- SDK adds UIHandler, SnapshotFiles/BackupPaths/CloseHooks, Config.ParentContext,
  RestoreFiles(required exact files). Most recent edit adds RestoreDatabases and
  RestoreWithDatabaseFiles integrity checks before atomic restore; NOT retested.
  App declares core/mytickets.db as checked snapshot. Native backup/recovery full
  API acceptance still required. SDK dependency still temporary104 with local go.work;
  commit/push SDK source then pin public commit before app release.
- Workspace registration now includes Tickets, port9026/rebased+36, foundation SSO,
  update group/package owner, general/default, private/recovery profile, icon,
  README/table/overview. InstallerIntegration5 (minimum planned105). Added narrow
  tk_-shaped Tickets API forwarding for agents/webhooks; native app validates real
  scopes/row perms. These latest routing tests need rerun.
- Workspace pkg/apphost/catalog/branding/distribution/suite all passed before latest
  routing/restore edits (workspace-integration-tests-3.log). Mandatory signed-Android
  publisher retained. Need ordinary full catalog build, new component pin, required
  capability, app metadata, newapp omission/selection, fresh setup+104 upgrade,
  independent update/recovery, public channel and protected primary install.
- update_metadata preserved list now skips tickets/core/attachments and tickets/tmp
  while retaining core database/key metadata for rollback. Check suite tests again.
- Critical lint incident resolved: golangci-lint --fix on CRLF source corrupted
  headers in1007 unrelated upstream files. Stopped task-owned fixer, restored those
  exact HEAD blobs while preserving eight intended tracked Go implementations and
  all new files. Added *.go text eol=lf. Do NOT rerun global --fix until proven safe;
  use mage lint read-only. Last read-only log backend-lint-3.log listed30 issues.
  Most are now manually fixed; remaining check needed. Two migration OwnerID fields
  were wrongly renamed by initial snapshot regex; now exported/restored correctly.
  Explicit justified cyclo annotations cover metric/history/predicate dispatch.
  goimports run through Format-Tickets.ps1 just completed; inspect format-source.log.
- Newly added10,000-row performance test has NOT run; batch its inserts before run
  to avoid SQLite variable limits. New OwState.Working flag and closed-sprint edit
  validation change need regenerated TS and tests. Template titles now human-friendly.
- Docs README/architecture/SSO/Git/Drive/operations/backup/API/MCP drafted with limits;
  acceptance/checklist/build/security/provenance/dependency notices still pending.
- Build/tooling: all big caches/artifacts I:/Projects/tooling/my-tickets-20261030;
  Go1.27.3 and Zig paths/scripts same as above. C NTFS task root contains frontend,
  pnpm store/npm _npx, Go temp and synthetic UI fixtures. C free ~2.3GB. Preserve
  sources/evidence before final safe cleanup. All test browsers/servers stopped by
  harness finally; verify ports19457/19458 before cleanup.
- Mage compilation is NOT concurrent-safe in one root (mage_output_file.go race).
  Run one go-tool-mage compile at a time. A compiled Mage cache binary may invoke
  test:e2e while a separate lint process is running; no source mutation during lint.
  Current cache binary088bd9c166957d1e9c7ec93edb7283188a2d7dab194dd053acfe10cab70d639c.exe.

Next: finish permission/concurrency/snapshot/MCP/schema/performance acceptance,
regenerate client, verify final UI/native binary; finish ordinary managed builder,
public source commits/pin, full install+upgrade tests, app/suite publication through
signed APK guard and channel, then backed-up owner install and served-byte proof.

## Release preparation checkpoint
Public SDK commit438b9cfe76b364de531aaba1f224a788950fd099 pushed to Workspace main and pinned in go.mod. Full stock+custom frontend rebuilt. Expanded TestTickets model+MCP checks PASS; 10K search159-170ms. Native Chrome/API/restore harness PASS in browser-evidence-8: real labels/UI, permissions, scoped token, MCP initialize,16 concurrent unique creates, stale-edit409, file delete/restore, logout and encrypted two-DB restore. Found real SQLite lock failures: upstream SQLite constructor ignores configurable pool bounds; set native sql.DB maximum/idle connections explicitly1. Service/reboot/cleanVM remain unverified. Source app is evaluation preview with documented gaps (remote Drive API, durable webhook retry queue, due-date automation). Final release/managed setup/owner install still pending. Native artifact final must be rebuilt after source commit with GOWORKoff and exact provenance. Lint8 uses fresh task cache because template CRLF normalization was not reflected in cached goheader diagnostics. Entire tracked Go source and code-header-template now LF; Git semantic diff remains bounded. No Actions used.

## Managed acceptance uncovered missing-email identity bug
Source pushed to org main992225e8615baa0372b7f6dca2a523b20693f723, Actions disabled; upstream workflows archived inactive because OAuth cannot push active workflows without workflow scope. App/suite releases still NOT published. Candidate managed catalog validates34 apps and Setup links on I using CGO0 linker scratch there; C scratch cannot hold the1.3GB embedded bundle plus linker output. Removed own C frontend node_modules/store and15 completed native fixtures; Playwright runtime copied/checked on I. Managed fresh setup selected accounts+Tickets succeeds but first real OIDC browser test found core user factory requires email, while SDK correctly omits unverified email. Native SDK cookie/session was valid; core mapping returned ErrNoUsernamePassword, leading UI401 loops. Fixed via trusted internal AllowMissingExternalEmail CreateUserOptions for authenticated issuer/subject adapter only, no fabricated email/no local validation weakening. Update verified metadata separately, return503 for unexpected core failures instead of reauth loop,403 for disabled account. Also set core CORS to actual scheme/host rather than mounted app URL. TestTickets9 models/user/MCP/native origin tests PASS; lint9 pending. NoEmail fix, test EXE hash binding and docs currently uncommitted; must commit/push, rebuild, regenerate app ZIP/pin/update manifest and candidate before more managed checks. App release1 was never published; old ZIP553d.../EXE6934... are obsolete candidates. Tool prepare-release.py asserts binary source commit and three acceptance reports; enhance its hash binding/use newest evidence. Owner main installation remains104 untouched.


## Final native and managed acceptance checkpoint (supersedes pending rows above)

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
