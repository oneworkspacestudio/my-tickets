# Security

Report sensitive vulnerabilities privately through this repository's Security tab.
Do not publish credentials, session cookies, database files or private attachments
in an issue. Describe the affected version and a synthetic reproduction.

Managed My Tickets uses MyAccounts OIDC, PKCE, exact redirect URLs and the native
Workspace application-host secure session. Browser resources use an HttpOnly
cookie with CSRF and account-change checks. Long-lived browser bearer tokens and
public self-registration are not enabled. Scoped API credentials remain subject
to both per-route permissions and project access checks.

Private SQLite databases, keys, temporary files and attachments belong in the
protected data directory. Never expose that directory through a generic web file
server. The optional Drive source is disabled by default to preserve project
membership boundaries. Ticket downloads recheck permission and content integrity.

Git integrations require a project administrator to choose a trusted service and
repository. Tokens and HMAC secrets are encrypted in SQLite with the protected
application secret; responses omit them. Webhooks require a scoped API credential,
signature and delivery ID. Workflow rules remain subject to normal validation.

Backups contain sensitive data even when encrypted. Keep their password separately,
restrict access to the files and retain a verified recovery copy before updating.
Restore checks archive authentication, paths, required files and SQLite integrity.

This evaluation release has not received an independent security assessment.
The [acceptance report](docs/ACCEPTANCE.md) distinguishes actual tests from pending
live-service, Windows service and clean-host checks. Upstream licence verification
remains intact; report upstream issues to the original project as appropriate.
