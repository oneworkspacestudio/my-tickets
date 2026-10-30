# Operations

## Standalone

Run on Windows 11 x64 with a dedicated NTFS data directory:

```powershell
./MyTickets.exe serve --data D:/Private/MyTickets --listen 127.0.0.1:9026 --origin http://localhost:9026
```

Open the one-time setup link printed on first launch and choose an owner password.
Do not share that link. The browser interface and API are embedded; the EXE does
not launch another application server or database process. No Internet CDN is
required for the essential interface. External identity/Git integration naturally
requires its configured service to be reachable.

`--version` prints the app version. `doctor --data D:/Private/MyTickets` prints a
small diagnostic summary without credentials or issue contents. Its port check
reports whether the chosen listener is occupied; it does not prove another
service is My Tickets.

## Optional Windows service

From an elevated terminal, with an absolute private-data path:

```powershell
./MyTickets.exe service install --data D:/Private/MyTickets --listen 127.0.0.1:9026 --origin http://localhost:9026
./MyTickets.exe service start
./MyTickets.exe service status
./MyTickets.exe service stop
./MyTickets.exe service uninstall
```

The service is the same EXE, with native SCM stop handling and bounded recovery
restarts. Uninstall refuses while it is running and preserves private data.
MyAccounts client credentials for a manually managed service must be supplied
securely through its deployment environment. Prefer One Workspace supervision
for normal ecosystem installation; do not run both managers against one data root.

The acceptance report identifies whether elevated service/reboot testing was
actually performed. Developer-host testing is not a clean Windows VM test.

## Upgrades

Use Workspace's normal verified updater. It preserves accounts, settings, files
and recovery copies. Tickets needs Workspace preview105 or later; older
controllers request the newer installer before applying the catalog. Independent
Tickets updates carry an explicit package namespace and compatibility contract.

For standalone use: export a verified backup, stop the process/service, replace
only the EXE, restart and check readiness and your issues. Migrations move forward;
restore the matched executable and full database backup to recover a failed schema
upgrade. Never copy an old database over a live WAL database.
