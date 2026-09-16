# ADR 0050 — Discord sign-in before the invite (identity-only sessions)

**Status:** Accepted · 2026-09-16 · Sprint S12.5 follow-up (landed during S33)

## Context

[ADR 0004](0004-discord-identity.md) wired Discord sign-in to an
invite: `/auth/discord/start` requires `?game=&t=`, parks both in the
state store, and the callback claims that seat in a single shot. The
button therefore only exists on the Join page, which you can only
reach by clicking a link that already names a table.

That is a fine flow for "click the link in the channel", and a bad one
for everything else. A player who has the game open in one tab, or who
was sent a code rather than a URL, or who simply lands on the site,
sees a login page offering an admin token and a box demanding a full
`…/#/games/<uuid>/join?t=<token>` URL. Signing in with Discord — the
thing the playgroup already has — is not on offer at all.

The ask was to put the Discord button on the login page and collect
the invite afterwards. That inverts the order the whole flow was built
around, and surfaces two problems.

**There was no session for "signed in, no seat."** ADR 0003's three
roles are each bound to something: `player` to a (game, seat),
`spectator` to a game, `admin` to the server. A Discord identity with
no table is none of those, and reusing `player` with a nil `GameID`
would hand the WS authorizer a principal that binds to the zero game.

**An invite code did not identify a table.** Invite tokens are
per-game: 16 bytes of `crypto/rand`, compared in constant time against
one specific game's token. Nothing could answer "which table does this
code belong to?", because nothing had ever needed to ask.

## Decision

### A fourth role: `auth.RoleIdentified`

A Discord-authenticated principal with the `Discord*` fields, no
`GameID` and no `PlayerID`. It can do exactly one thing — `POST /join`
with an invite code, which trades it for a `RolePlayer` session.

`lobby.WSAuthorizer` rejects it **by name** rather than letting it
reach the `default` arm. The default's "unrecognised principal role"
would be a lie about a role the server mints itself, and the explicit
arm can say what to do next instead.

The identity session is deliberately **not** revoked when it is traded
in. It is how the same person joins a second table later without
another consent screen, and it is inert on its own.

### An unbound OAuth round-trip

`/auth/discord/start` now takes either `?game=&t=` (invite flow,
unchanged) or no parameters at all (login flow). `StateEntry.Unbound()`
reports which, and the callback branches on it: claim the seat, or mint
an identity session and bounce to `#/oauth-complete` with no `game` in
the fragment. The fragment's shape is how the client tells the two
apart.

One of the pair without the other is a 400. Treating a half-filled link
as unbound would silently discard a seat claim the user asked for, and
the missing half cannot be invented.

### `Lobby.FindByInvite` and `POST /join`

`FindByInvite` resolves a bare player-invite token to its game. Two
properties are deliberate:

- **No short-circuit on match.** Returning early would make response
  time a function of where the table sits in the map — a weak oracle,
  but free to close at a scale of a handful of games.
- **Player invites only, live tables only.** A spectator code
  resolving here would let a read-only link open a seat-claiming flow,
  which is exactly the distinction the two tokens draw. Archived
  tables are skipped so a stale code reads as expired rather than
  quietly reopening a retired table.

`POST /join` is the path-free sibling of `POST /games/{id}/join`. Its
session is **optional**, which is the unusual part and the point: an
identity session supplies the seat's name and avatar (and the body's
`name` is ignored, so a caller cannot relabel a seat Discord vouched
for), no session at all behaves exactly like the classic manual join,
and any other role is a 409. A credential that fails to validate falls
through to the manual path rather than 401-ing — an identity session
that expired while the user hunted for the code should feel like "type
your name", not like being thrown out.

It shares the `/games/{id}/join` rate-limit bucket: same credential
being guessed, and a code-only endpoint is the easier one to script.

## Consequences

- Manual name entry stays first-class on every path. A deployment with
  the three `CMDCTRL_DISCORD_*` values unset hides the button and
  everything else still works, which is the same posture ADR 0004 set.
- Invite tokens are now globally searchable by anyone who can reach
  `POST /join`. The entropy (16 random bytes) and the shared limiter
  make guessing impractical, but this is a real widening of what a
  token is: previously you needed the game id **and** the token, and
  now the token alone is sufficient. Recorded here because it is a
  deliberate softening of ADR 0003's model, not an oversight.
- Sessions remain in memory, so a restart still logs everyone out —
  including identity sessions, which simply means signing in again.
- The login page now has three ways in (Discord, invite code or link,
  admin token). The first two share one input, because a link that
  already names its table needs no lookup.

## Alternatives considered

**Keep OAuth per-game; teach the login box to take a code.** Smallest
change — no new role, no lookup — but it puts the invite before the
sign-in, which is the order the request was explicitly trying to
reverse.

**A short-lived identity cookie instead of a session.** Avoids a
fourth role, but invents a second credential type sitting outside
`auth.Authenticator`, which is the seam ADR 0003 built specifically so
identity questions have one answer.
