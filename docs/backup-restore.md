# Backup and restore

Workspace Backups covers Tickets data and attachment files. Update recovery copies
cover the program/settings/database needed for rollback, while keeping immutable
attachment bytes in their separate data folder. Updates do not copy all uploaded
files again or put data in the program directory.

The EXE also provides encrypted, password-protected standalone backups. Stop a
standalone server before the CLI command, because the private folder has an
exclusive process lock:

```powershell
$env:WORKSPACE_BACKUP_PASSWORD = 'choose-your-own-long-backup-passphrase'
./MyTickets.exe --data D:/Private/MyTickets --backup E:/Backups/tickets.owbackup --no-browser
./MyTickets.exe --data D:/Private/MyTickets-restored --restore E:/Backups/tickets.owbackup --no-browser
Remove-Item Env:WORKSPACE_BACKUP_PASSWORD
```

The backup file must be new. The restore destination must not already exist.
Never put a real password in a shared script or shell history. The owner-only
native live-backup endpoint is `POST /api/backups`, using the native session and
CSRF protection; Workspace's usual backup UI is the simpler managed path.

The snapshot includes both SQLite databases, protected keys and attachment bytes.
The core database uses SQLite's online backup API; the live main/WAL files are
not copied directly. The archive is encrypted and authenticated. Restore rejects
wrong passwords, unexpected paths, traversal, symlinks, missing required files,
invalid integrity and existing destinations. Browser sessions are revoked after
restore. Keep the backup password separately; it cannot be recovered from the EXE.
