# My Git / Gitea

A project administrator opens **Project settings → My Git / Gitea** and saves
its trusted HTTPS Git server and exact `owner/repository`. The service token needs
read access to that repository. Provide a random webhook secret of at least 32
characters. Tokens/secrets are encrypted with AES-GCM using a key derived from the
protected app key; they are never returned by GET or stored in ticket text.
Protect and back up the private-data directory and `core.key` together.

Create a Gitea JSON webhook:

- Target: `https://YOUR-WORKSPACE/tickets/api/v2/tickets/projects/PROJECT_ID/git/webhook`
- Secret: the same configured secret.
- Authorization header: `Bearer YOUR_SCOPED_MY_TICKETS_API_TOKEN`.
- Events: push and pull request.

Use a dedicated service identity with access to this project. Its API token must
include the webhook operation from `/api/v2/routes`. Token routing through
Workspace is limited to the Tickets API; the native app still validates the key,
route scope, project write permission and HMAC-SHA256 of the exact body.
Never place a credential in a webhook URL.

Mention `WORK-123` in a commit or pull request title/body. Links are accepted only
for the bound repository and trusted server; issue read permission is checked.
Push branch references and commit metadata, plus pull request state, are retained
with each link. A previously linked PR can still run an authorized merge rule.
Delivery IDs are idempotent. A failed transaction is not marked delivered, so
Gitea can retry it safely. Existing links remain available during Git outages.
**Resync recent commits** imports up to 50 commits without triggering historical
workflow transitions. Network requests do not hold the core SQLite transaction.

Workflow automation is off until an administrator adds an explicit event/state
rule. A merged/closed pull request rule applies only when `merged` is true.
Normal workflow transitions, required fields and project permissions still apply;
commit/PR text cannot invent actions or execute commands.

The acceptance report separates signed synthetic webhook tests from any live
repository check. Repository-wide historical import and CI/build synchronization
are not implied by a recent-commit resync.
