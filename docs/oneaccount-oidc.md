# MyAccounts / OneAccount

Managed setup creates the `tickets` OIDC client and exact callback
`https://YOUR-WORKSPACE/tickets/auth/callback`. The manager supplies
`WORKSPACE_OIDC_ISSUER`, `WORKSPACE_OIDC_CLIENT_ID`,
`WORKSPACE_OIDC_CLIENT_SECRET`, `WORKSPACE_OWNER_SUBJECT` and
`WORKSPACE_HOME_URL` directly to the supervised process.

Launch through the Workspace apps picker. A valid MyAccounts browser session
finishes the standard authorization-code/PKCE flow without another password.
The application verifies issuer, audience, state and nonce, then issues its own
revocable HttpOnly/SameSite cookie. HTTPS deployments use Secure cookies.
Subject identity, rather than display name or a claimed administrator field,
controls account binding. Only the configured verified owner gains system admin.

Managed installations disable local account creation and password sign-in.
Standalone mode has a one-time owner setup link and owner-created local members;
there is no public registration endpoint. Passwords are hashed, not reversible.

Project creators administer their own projects. Project administrators can grant
viewer/contributor/administrator access in Project settings. A teammate must
first open My Tickets so their stable account exists; their username appears in
the profile menu. Project and team authorization remains server-side.

Workspace sign-out revokes remembered native sessions. An already-expired or
changed native session cannot be reused by an old browser resource request.
See the acceptance report for actual local OIDC and managed sign-out evidence.
