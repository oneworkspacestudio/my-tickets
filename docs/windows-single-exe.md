# Windows single executable verification

## Baseline status

This branch is pinned to Vikunja v2.7.1, commit
`5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0`. This document preserves the initial
upstream packaging proof. Current My Tickets testing is recorded in
[ACCEPTANCE.md](ACCEPTANCE.md); native checks use `mage test:e2e tickets`.
See the implementation checklist in
[oneworkspace/IMPLEMENTATION.md](oneworkspace/IMPLEMENTATION.md).

The complete upstream frontend is embedded by `frontend/embed.go`. SQLite uses
`mattn/go-sqlite3`, compiled into the executable with CGo; a database service is
not required. Runtime self-containment must be demonstrated, not inferred from
an EXE filename. Windows operating-system DLLs are permitted; separately
installed database/Node/Java/Go runtimes are not.

The official upstream release passed a synthetic local test with only the EXE
in its initial directory and PATH limited to Windows system directories. That
reference result does **not** verify a source build or a clean Windows VM.

The separate source-built baseline also passed the developer-host functional
smoke. Its result records a signed Citrix module injected by this PC, so runtime
module acceptance and clean Windows VM verification remain UNVERIFIED. Only
Windows 11 build 22631 has been exercised. See the [Phase 0 report](oneworkspace/phase-0-report.md).
Recorded evidence explicitly distinguishes them.

## Reproduce the smoke test

Use a new Windows VM without Node, Go, PostgreSQL, MariaDB, Redis, or a companion
app server. Transfer the source-built EXE and the verification script. Use
PowerShell 7 as the test driver (it is not required to run the product).

```powershell
./scripts/verify-single-exe.ps1 `
  -Executable ./Vikunja-2.7.1-baseline.exe `
  -EvidenceDirectory C:/TicketsProof/run-1 `
  -EnvironmentKind CleanWindowsVM
```

`EnvironmentKind` records the operator's declaration; it does not magically
attest that the host is clean. Preserve the VM setup record with the result.
For a developer-host trial, omit that option. Use a new evidence directory for
every run. The script refuses existing fixture ports and inherited Vikunja
configuration, creates a disposable account, tests embedded HTML/JS, API writes,
anonymous denial, and persistence after forced process termination. It stops
only its own fixture process and restores its environment. It reports loaded
modules outside Windows. It never connects to live Workspace services.

Inspect `result.json`, then manually open the browser against a fresh fixture
and exercise login, project/task creation, editing and reload. The automated
script does not claim browser interaction, accessibility, application features,
service operation, backup/restore, or integrations were tested. Fixture data is
synthetic and can be removed after reviewing its evidence.

## Source build on Windows

Build the complete frontend with the package manager pinned in
`frontend/package.json` and its frozen lockfile. The development checkout and
pnpm store should be on NTFS. An exFAT pnpm store failed to register package
links in this environment. Build caches and published artifacts can be kept on
another drive.

The pinned upstream Mage file currently references Unix-only `Setpgid`,
`Getpgid`, and `Kill` symbols even on Windows. The stock application can be
built with the Go invocation used by the Mage build target, without compiling
that build-driver file:

```powershell
$env:CGO_ENABLED = '1'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
# CC must name an available static-capable Windows x64 C compiler.
# This baseline used Zig 0.15.2 with: cc -target x86_64-windows-gnu -static
# Select dedicated GOCACHE, GOMODCACHE, GOTMPDIR and ZIG_GLOBAL_CACHE_DIR.
go build -p 2 -trimpath -tags osusergo `
  -ldflags '-s -w -extldflags=-static -X code.vikunja.io/api/pkg/version.Version=v2.7.1 -X main.Tags=osusergo' `
  -o Vikunja-2.7.1-baseline.exe .
```

This is a build command, not permission to bypass the repository's requirement
that tests run through Mage. The fork separates the process helpers by platform so Mage compiles on
Windows; `go tool mage -l` passed. Focused application tests remain separately
recorded. Do not report the application tests as passing
because the source build succeeds.

## Acceptance gate

The supplied brief requires: “Verify on clean Windows environment. Gate: pass
single-EXE proof.” Clean Windows VM acceptance is pending. Do not label this
baseline a finished My Tickets release or install it into the owner's live
Workspace as if the requested Jira-style features had been delivered.
