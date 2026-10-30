# Upstream

- Project: Vikunja, official GitHub repository https://github.com/go-vikunja/vikunja
- Canonical module URL: https://code.vikunja.io/vikunja
- Stable tag: v2.7.1 (annotated tag object 91107279377ecf45f652a12589a687819f8d8e90)
- Source commit: 5619c4a9f66ddfc13f7bbba24ada07c3e82ae7a0
- Release notes: https://vikunja.io/changelog/vikunja-2.7.1-was-released/
- License: AGPL-3.0-or-later; preserve LICENSE, copyright headers and network source availability.
- Reason: stable 20 October 2026 patch with security fixes, API v2 and built-in MCP.
- Build baseline: unmodified upstream first. Frontend embedding is in frontend/embed.go.
  SQLite currently imports github.com/mattn/go-sqlite3 and needs a proven CGo build.
- Keep the upstream remote. Rebase/merge pinned stable security releases into a
  review branch, run baseline and namespaced-extension acceptance, then promote.
- Do not remove upstream license checks or claim Pro features as free upstream features.

The user explicitly chose this machine and waived clean Windows VM verification.
Tests use isolated synthetic NTFS data. Developer-host checks do not establish
clean-host, Windows service or reboot acceptance; see docs/ACCEPTANCE.md.
