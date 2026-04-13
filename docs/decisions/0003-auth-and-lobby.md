# ADR 0003 — Auth and lobby architecture

**Status:** Accepted · 2026-04-13 · Sprint S04

## Context

S04 adds the non-play surface of cmd_and_ctrl: game creation, seat
claiming, and identity. Three axes of decision were live:

1. **Who is allowed in?** The original issue called for a shared-
   password login. We evaluated alternatives: per-user accounts, HMAC-
   signed stateless tokens, magic-link flows. For four friends we
   landed on an **invite-link-as-credential** model — invites are the
   primary onboarding surface, a small admin token gates game
   creation.
2. **How is identity carried?** Browser WebSocket upgrades cannot set
   `Authorization` headers; we need a token transport that works for
   both HTTP and WS, ideally without breaking CLI clients.
3. **How portable is this to public deployment?** We want to ship the
   friends-only version quickly but avoid hard-coding assumptions
   that a future HMAC / JWT / OIDC layer cannot absorb.

## Decision

### Pluggable `auth.Authenticator` interface

All identity operations live behind a narrow interface:

```go
type Authenticator interface {
    Issue(ctx, Principal, ttl) (credential string, issued Principal, err error)
    Validate(ctx, credential) (Principal, error)
    Revoke(ctx, credential) error
}
```

The rest of the server speaks only in `Principal` — role (admin or
player), optional `GameID` / `PlayerID`, display name, IssuedAt /
ExpiresAt. Credential shape is entirely an implementation concern of
the `Authenticator`.

### Default S04 implementation: in-memory invite store

`auth.MemoryAuthenticator` is a server-side `map[token]Principal`
with lazy expiry and constant-time validation. Tokens are 32 bytes
of `crypto/rand`, base64url-encoded.

- Invites are per-game 16-byte tokens generated at game-create time
  and embedded in the share URL: `#/games/<id>/join?t=<token>`.
- The invite endpoint (`POST /games/{id}/join`) is the only public
  route; a valid invite mints a RolePlayer session bound to
  `(gameID, playerID)`.
- Admin sessions come from `POST /admin/login` with the shared
  `CMDCTRL_ADMIN_TOKEN` env var.

Sessions are in-memory: a server restart logs everyone out. This is
acceptable for a four-friends LAN deployment; the single restart
forces re-join via the still-valid invite URLs.

### Credential transport: cookie + bearer + query, in that order

- Browser HTTP: `cmdctrl_session` HttpOnly cookie set on login/join.
- Browser WS upgrade: same-origin cookie works; query `?token=` is
  the fallback for cross-origin deployments.
- CLI clients: `Authorization: Bearer <token>` header.

All three are honoured by `auth.CredentialFromRequest` with cookie >
header > query precedence.

### Cross-game session binding

The WS authorizer (`lobby.WSAuthorizer`) enforces that a RolePlayer
session cannot be reused to bind into a different game: if the
session's `GameID` doesn't match the requested `?game=` value, the
upgrade is rejected with 403. Admins are exempt; they may bind to
any game and optionally supply `?player=` to render a specific
seat's view.

### Deferred to later sprints

- **Persistent sessions** (SQLite-backed map) — not worth the
  complexity at ≤8 users.
- **Rate limiting on login / invite attempts** — trust-the-LAN for
  now; revisit when exposing publicly.
- **Per-invite revocation** — invite tokens are per-game, not per-
  person; leak mitigation is "rotate the game" (i.e. recreate).
- **TLS termination** — assumed handled by a reverse proxy in S12's
  deploy work. Until then, do not deploy publicly.
- **Public-deployment auth backend** — the `Authenticator` interface
  is sized for a stateless HMAC implementation to slot in without
  touching handlers. When we go public, write
  `auth.HMACAuthenticator`, swap the one construction site in
  `cmd/server/main.go`, done.

## Consequences

- Lobby / game-state handlers never see raw tokens. The seam is
  narrow (three methods) so swapping backends is a contained change.
- The S04 server requires exactly one secret (`CMDCTRL_ADMIN_TOKEN`);
  invites flow from there. No user database, no password file.
- Browser refresh preserves sessions via localStorage. A 401 from
  any API call clears the local session and bounces to `/login`.
- The invite link is credential-equivalent for joining a specific
  game; share it over a private channel (Discord DM, SMS), not a
  public one.
