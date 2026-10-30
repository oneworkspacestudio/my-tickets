# My Tickets Phase 0 status

This is an unfinished product baseline. No My Tickets repository has been
published, no release has been created, and no live Workspace installation or
tunnel has been changed.

## Source and build

- Upstream: Vikunja v2.7.1, AGPL-3.0-or-later.
- Commit: `5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0`.
- The source-built baseline used the unchanged upstream application source and
  complete 361-file frontend. The fork separately repairs the Windows Mage
  build-driver compilation by keeping Unix process controls out of Windows, and
  fixes an upstream embedded-test-fixture path separator error. Those compatibility
  patches are not included in the earlier unchanged-source baseline EXE.
- Binary: `Vikunja-2.7.1-baseline.exe`, 86,813,696 bytes.
- SHA256: `537e158c9b555878eeca190cbbda8f954b5886d33d26b0f5171cd26b96937772`.
- Build command and verification steps: [Windows proof](../windows-single-exe.md).
- Source build details: [machine-readable evidence](test-evidence/source-build.json).

## Evidence and limits

The developer-host smoke began with only an EXE. With PATH restricted to Windows
system directories it served embedded HTML/JS, wrote a project and task using
SQLite, rejected anonymous task access, and retained the task after forced
termination/restart. It stopped its own processes afterward. The disposable
fixtures were removed after saving sanitized evidence.

PE imports contain Windows KERNEL32/UCRT API sets and no delay imports. The
runtime also had a validly signed Citrix application-protection module injected
by this developer PC. It is also present in the test driver. No security software
was disabled. This environment is not a clean Windows machine; runtime-module
acceptance remains UNVERIFIED. See [functional result](test-evidence/source-runtime.json),
[PE imports](test-evidence/source-imports.json), and
[host module evidence](test-evidence/developer-host-module.json).

The initial broader filtered test run found the embedded-fixture bug and was
stopped after that failure. After the fix the scoped packaging target passed.
The full application suite remains UNVERIFIED; see [test evidence](test-evidence/packaging-tests.json).

The automated test driver does not assert browser interaction, a production
security audit, application feature completeness, or live integrations.

## Requirement status

| Requirement | Status | Evidence or outstanding work |
| --- | --- | --- |
| Full upstream frontend build | PASS | 361 files, pinned pnpm and frozen lockfile |
| Windows x64 source build and SQLite writes | PASS | Source-build and developer-host functional proof |
| Clean Windows VM single-EXE acceptance | UNVERIFIED | No clean VM available; explicit brief gate pending |
| Windows Mage compilation | PASS | `go tool mage -l` after platform helper separation |
| Focused packaging tests | PASS | `mage test:packaging`: DB/config/doctor/files; named-fixture regression failed before and passed after fix |
| Backend lint | UNVERIFIED | `mage lint:fix` attempted; golangci-lint v2.14.0 unavailable |
| Interactive browser and mobile acceptance | UNVERIFIED | Not exercised by the HTTP smoke |
| My Tickets branding and Jira-style extensions | FAIL | Not implemented; upstream baseline only |
| Custom workflows/fields, sprints, dashboards/reports | FAIL | Not implemented |
| OneAccount, My Git, My Drive adapters and scoped MCP extensions | FAIL | Not implemented |
| Native Windows service and required operations UX | FAIL | Not implemented |
| Org publication, release, installer/navigation/tunnel integration | FAIL | Not performed; product not ready |

FAIL here means an unmet product requirement, not a claim that upstream Vikunja
is broken. The implementation checklist remains authoritative for remaining work.

## Gate and resume

The brief explicitly requires: “Verify on clean Windows environment. Gate: pass
single-EXE proof.” The operator was asked for a clean Windows VM or an explicit
change to use an isolated developer-host test first with VM acceptance pending.
No answer has been received. Do not silently waive that gate, label this a
finished Jira-style product, or install the baseline as the completed app.
