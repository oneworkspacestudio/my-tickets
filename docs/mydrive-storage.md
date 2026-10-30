# Drive and attachments

Attachments are stored through a namespaced local storage adapter in
`core/attachments/PROJECT_KEY/ISSUE_KEY/ATTACHMENT_ID/safe-basename` under the
protected Tickets data root. The database records original name, byte length,
uploader, SHA256 and recoverable deletion state. Uploads are capped at 64 MB,
written to a partial file, synchronized, then atomically renamed. Failed streams
are removed. Downloads reauthorize current issue membership and check integrity.
A supplied filename never selects a filesystem path.

The managed installer registers this folder as the **My Tickets** Drive source.
Its default permissions are disabled: broad Drive access must never bypass ticket
project membership. Members use the authorized attachment interface in Tickets;
a Drive administrator can inspect the source. Other Drive sources are preserved.
The source contains attachment bytes, never the ticket database, keys or sessions.

Delete moves an attachment into recoverable deletion; **Restore** restores the
same ID and bytes. No background job permanently purges these files. Renaming an
issue's project does not expose another path; a moved issue retains authorized
references to its existing bytes. A missing/offline storage path produces a
retryable error while issues and dashboards remain available.

The standalone EXE uses the same local adapter without requiring Drive. This
preview's managed integration is a protected shared filesystem source, not a
separate remote Drive-server upload adapter. Do not enable the raw source for
ordinary users without considering project permissions. Use Workspace Backups
for the complete data set; update recovery deliberately preserves blob bytes.
