# Build My Tickets for Windows x64

Use Go 1.27.3, Node 24 or later, pnpm 11.27.0 and a Windows x64 C compiler.
The recorded build uses Zig 0.15.2 targeting x86_64-windows-gnu with static linking.
Frontend dependencies and Go modules are pinned in the repository. Keep pnpm's
project/store and Go linker scratch on NTFS; large dependency caches can use another
drive. Never build in an installed application's data or program directories.

```powershell
$env:CC = 'C:/Tools/zig/zig.exe cc -target x86_64-windows-gnu -static'
$env:CXX = 'C:/Tools/zig/zig.exe c++ -target x86_64-windows-gnu -static'
./scripts/build-windows.ps1 -Out D:/Build/MyTickets
```

The script builds the full upstream frontend and the My Tickets Vue entry, embeds
both with Go embed, adds the Windows icon, and builds one `MyTickets.exe`. Build
tools download pinned dependencies during compilation, never during app startup.
`-FrontendDirectory` allows an isolated NTFS frontend copy. `-PreparedFrontend`
is intended for a separately completed and verified frontend build. Set
`SOURCE_DATE_EPOCH` to the source commit timestamp for repeatable build metadata.
Use the same toolchain, source and embedded assets when comparing binary hashes.

The public Workspace SDK commit is pinned in go.mod. Release builds explicitly use
`GOWORK=off`; no local checkout replacement is needed. `doctor --data <absolute>`
reports app/source/upstream versions and data/listener status. See
[acceptance](ACCEPTANCE.md) for the actual tested environment and limits.

Run Go checks through Mage: `go tool mage test:tickets`, `go tool mage test:packaging`
and `go tool mage lint`. For isolated browser acceptance use
`go tool mage test:e2e tickets` with the documented `TICKETS_TEST_*` fixture paths.
The browser harness creates only synthetic accounts and data, stops its child
processes, checks API access and restores an encrypted snapshot to a new directory.
