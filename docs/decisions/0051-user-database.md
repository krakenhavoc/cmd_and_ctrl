# ADR 0051 — A persistent user database: people, their games, and their decks

**Status:** Accepted · 2026-09-16 · Sprint S34 · Tracking issue [#607](https://github.com/krakenhavoc/cmd_and_ctrl/issues/607)
**Builds on:** [ADR 0003](0003-auth-and-lobby.md) (roles and the
`Authenticator` seam), [ADR 0004](0004-discord-identity.md) (Discord
as the identity source), [ADR 0041](0041-game-persistence.md) and
[ADR 0044](0044-surviving-a-deploy.md) (what survives a restart, and
stateless sessions), [ADR 0050](0050-discord-login-identity.md)
(identity-only sessions).
**Supersedes in part:** ADR 0041's `<dumpDir>/lobby/<id>.json` artifact,
and ADR 0044 decision 4's need for a per-seat reclaim secret.

## Context

The server has never had a notion of a *person*. It has seats.

A session principal is bound to one game and one player ID. The
Discord fields on it are cached at sign-in so a seat can render a
name and an avatar, and nothing else reads them. `GameMeta` is one
JSON file per game, and the only record of who sat where is the
`SeatInfo` inside it. Two games played by the same four friends share
no row, no key and no file. ADR 0050 added the first person-shaped
thing — a Discord-authenticated session with no seat — but it is an
identity waiting for an invite code, and it evaporates on restart like
every other session.

So none of the following can be answered today, and the owner has
asked for all of them:

- **"Show me my games."** Every game I have sat in, live or ended,
  from any device, without an invite link.
- **"Invite the people I usually play with."** Today an invite is a
  link pasted into a channel. There is no list of people to pick
  from, because there is no list of people.
- **"Stay signed in, stay connected to Discord."** ADR 0050's identity
  session dies with the process, and the OAuth flow discards the
  refresh token, so the server cannot refresh a name or an avatar
  without another consent screen.
- **"Keep my decks."** A deck is uploaded per seat, per game, and
  parsed into the engine. The text the player pasted is not kept.

Every one of these is identity that outlives a game. That is the
trigger, named before this ADR was written, at which the file-per-game
layout stops being the right store: a query that spans games, and
rows that belong to a person rather than to a table.

Two ground rules from [AGENTS.md](../../AGENTS.md) bound the answer.
*One VPS, one database, one deployable per service.* And the scale is
eight people. Nothing here is built for more.

## Decision 1 — SQLite, in the data directory, owned by the server process

The store is a single SQLite file at `<dataDir>/db/cmdctrl.sqlite`,
opened in WAL mode by the game server and by nothing else. The driver
is `modernc.org/sqlite`: pure Go, no cgo, so the binary stays static
and the rsync-and-restart deploy in [`ci-cd.yml`](../../.github/workflows/ci-cd.yml)
is unchanged.

**Not a database server.** Postgres earns its place when a second host
exists or a second writer needs transactions against the same rows.
Neither is true and neither is planned. ADR 0044 decision 1 already
records that the topology is one unit on one VM; this ADR inherits
that and adds nothing that would need a network-reachable store.

**One writer, and it is the server.** The Discord bot is a separate
process (ADR 0004 §2) and it keeps talking to the server over HTTP.
It never opens the file. This is the rule that makes the current
design safe — every piece of state has exactly one owner — carried
into the database rather than abandoned at it. A second process with
a write handle on the same SQLite file is how you get `SQLITE_BUSY`
at the worst moment and a bot that half-wrote an invite.

**The file is mode `0600`**, for the same reason ADR 0041 gave
`lobby/<id>.json` that mode: it holds invite tokens and, with
decision 5, encrypted OAuth tokens.

**Schema migrations are numbered SQL files embedded in the binary**,
applied at boot inside a transaction before `RestoreFromDisk` runs,
recorded in a `schema_migrations` table, forward-only. The same
posture as `SnapshotSchemaVersion`: a file written by a newer binary
than the one booting is refused, loudly, and the process exits — a
rolled-back binary must not run against a schema it does not
understand. No ORM, no query builder; `database/sql` and hand-written
SQL, which matches the stdlib-first convention and keeps the surface
auditable.

**Backups become a prerequisite, not a nice-to-have.** Nothing backs
up `/var/lib/cmd_and_ctrl/data` today. Losing a restore point costs
one game; losing this file costs every account, every deck and every
invite. A nightly off-node copy of the data disk lands in
[HomeLab](https://github.com/krakenhavoc/HomeLab) before sub-PR 2
below merges, and one restore is rehearsed. SQLite's online backup API
(`VACUUM INTO`) gives a consistent copy without stopping the service;
the sweep that produces it belongs to the server, on a timer, writing
beside the live file so the disk-level backup picks up a clean copy.

**Amended (2026-09-19): how the off-node copy is done** ([#1031](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1031)).
The off-node copy is not a HomeLab job. HomeLab has a single Proxmox
node and no target off it, so each VM backs itself up nightly with
restic to its own Cloudflare R2 bucket. It backs up the `VACUUM INTO`
copy, never the live file, along with `restore/`, `replays/`,
`lobby/`, `bugreports/` and `games/`. The job, its credentials and the
restore runbook ship from this repo's CD, as the Caddyfile and the bot
unit do. HomeLab owns only the buckets
([krakenhavoc/HomeLab#58](https://github.com/krakenhavoc/HomeLab/issues/58)).
It costs $0 within R2's free tier. See
[docs/environments.md](../environments.md#backups).

## Decision 2 — A user is our ID; a Discord account is one identity attached to it

```sql
users (
  id                       TEXT PRIMARY KEY,   -- uuid, minted by us
  display_name             TEXT NOT NULL,
  avatar_url               TEXT,
  created_at               INTEGER NOT NULL,
  last_seen_at             INTEGER NOT NULL,
  sessions_invalid_before  INTEGER NOT NULL DEFAULT 0   -- decision 6
)

identities (
  provider        TEXT NOT NULL,             -- 'discord' today
  subject         TEXT NOT NULL,             -- the snowflake
  user_id         TEXT NOT NULL REFERENCES users(id),
  display_name    TEXT NOT NULL,             -- as the provider last reported it
  avatar_hash     TEXT,
  refresh_token   BLOB,                      -- encrypted; decision 5
  scopes          TEXT NOT NULL,
  linked_at       INTEGER NOT NULL,
  refreshed_at    INTEGER,
  PRIMARY KEY (provider, subject),
  UNIQUE (user_id, provider)
)
```

The Discord snowflake is deliberately **not** the primary key of
`users`. The owner has said an external identity provider (Entra
External ID was the example) may follow. That is a new `provider`
value and nothing else — no schema change, no re-keying of every seat
and deck that points at a user. The cost today is one join, on a
table of eight rows.

**Guests remain first-class at the table.** A friend who clicks an
invite and types a name still gets a seat. That seat has a `NULL`
`user_id` and a `guest_name`, and it is not retrievable later — no
"my games", no decks, no tablemates. ADR 0050's posture ("manual name
entry stays first-class on every path") is kept exactly. What changes
is what *ownership* requires:

| Action | Needs a signed-in user |
|---|---|
| Sit at a table via an invite | no |
| Create a game | **yes** |
| Save a deck to the library | **yes** |
| See your game history | **yes** |
| Be suggested as a tablemate, or invited by DM | **yes** |

Admin sessions are unchanged: the admin token is a server credential,
not a person, and it gets no user row. A game created by an admin
session has a `NULL` `created_by`.

## Decision 3 — Sessions are still stateless; they now carry a user

S33 sub-PR 1 ([#517](https://github.com/krakenhavoc/cmd_and_ctrl/issues/517))
is replacing `MemoryAuthenticator` with an HMAC-signed `Principal`.
This ADR asks that the `Principal` gain **one field, `UserID`**, in
that PR rather than in a second migration of the credential shape:

```go
type Principal struct {
    Role     Role
    UserID   uuid.UUID   // zero for admin and guest sessions
    GameID   uuid.UUID
    PlayerID uuid.UUID
    // Discord* fields: retained for now, see below
}
```

The four roles from ADR 0003 and ADR 0050 stay. The relationship
between them becomes:

- **`identified`** is the durable, per-person session: minted at
  Discord sign-in, long TTL (30 days, `CMDCTRL_IDENTITY_TTL`),
  carries `UserID` and nothing table-specific. This is the credential
  a browser holds across restarts and across games.
- **`player`** is minted *from* an identified session when a seat is
  claimed, exactly as ADR 0050 already does, and now also carries
  `UserID`. It stays bound to a `(GameID, PlayerID)` because the WS
  authorizer, the room, and the bot-seat runner all reason about one
  socket as one seat. A guest's `player` session has a zero `UserID`.
- **`spectator`** and **`admin`** are unchanged.

**Collapsing `player` into `identified`** — one user session, and
"which seat is yours" answered by the `seats` table at upgrade time —
was considered and deferred. It is the cleaner end state and the
`seats` table makes it possible, but it rewires `WSAuthorizer`, the
bot seat path and reclaim in one move, on top of S33 rewiring the
credential itself. Two credential migrations in one quarter is one
too many. It is recorded under *Deferred* with the shape it should
take.

**The cached `Discord*` fields on `Principal` become redundant** once
`UserID` resolves to a row that holds the same data. They are kept
through S34 so the seat label and avatar path does not change in the
same PR as the store, and removed in the collapse above.

**Seat reclaim** (ADR 0044 decision 4, [#520](https://github.com/krakenhavoc/cmd_and_ctrl/issues/520))
no longer needs a per-seat secret. `seats.user_id` is the proof. A
signed-in user whose `player` token has expired asks for the seat
whose `user_id` is theirs and gets a fresh `player` session. Guests
keep the invite-link path #520 builds, since they have no identity to
prove; it is the backstop for them, not the mechanism for everyone.

## Decision 4 — Games, seats and invites become rows; the engine's files do not

```sql
games (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  created_by    TEXT REFERENCES users(id),   -- NULL: admin-created
  state         TEXT NOT NULL,               -- lobby | active | ended
  created_at    INTEGER NOT NULL,
  started_at    INTEGER,
  ended_at      INTEGER,
  archived_at   INTEGER,
  winner_seat   INTEGER                      -- NULL until the engine reports one
)

seats (
  game_id       TEXT NOT NULL REFERENCES games(id),
  seat          INTEGER NOT NULL,
  player_id     TEXT NOT NULL,               -- the engine's Player.ID
  user_id       TEXT REFERENCES users(id),   -- NULL: guest or bot
  guest_name    TEXT,
  bot_tier      TEXT,                        -- NULL: human
  deck_id       TEXT REFERENCES decks(id),   -- NULL: placeholder or ad-hoc upload
  deck_name     TEXT,
  PRIMARY KEY (game_id, seat)
)

invites (
  token_hash    BLOB PRIMARY KEY,            -- sha256 of the 16 random bytes
  game_id       TEXT NOT NULL REFERENCES games(id),
  kind          TEXT NOT NULL,               -- player | spectator
  created_by    TEXT REFERENCES users(id),
  created_at    INTEGER NOT NULL,
  expires_at    INTEGER,
  revoked_at    INTEGER
)
```

This **replaces** `<dumpDir>/lobby/<id>.json` from ADR 0041. Every
field that file carried is a column above, and the one reason it
existed — "a restored table nobody can open is indistinguishable from
a lost game" — is served by `invites` surviving in the database
instead.

**What stays on disk, untouched:**

| Path | Why it stays |
|---|---|
| `restore/<id>.json` | the engine snapshot is a blob of one game; rows gain nothing |
| `replays/<id>.jsonl` | append-only, large, read by the scrubber and bug reports |
| `games/<id>.json` | forensic view dump |
| `scryfall/`, `images/`, `avatars/` | reference data and caches |
| `bugreports/` | screenshots are public artifacts by design |

A `games` row is the index; the files are the payload. `Delete` reaps
both, as `RoomManager.Delete` does today; `Archive` flips
`archived_at` and touches no file, as today.

**Invite tokens are stored as a hash.** ADR 0050 made a bare token
resolvable to its game with `FindByInvite`, and worried, correctly,
about a timing oracle from short-circuiting the scan. A SHA-256
lookup on a primary key has no such oracle and is O(1), so the
"no short-circuit" scan and its comment go away. Tokens gain an
optional expiry and an explicit revocation, which the file never had
room for; the default remains "no expiry", matching current behaviour.

**"My games"** is `SELECT … FROM seats JOIN games WHERE user_id = ?`,
ordered by `created_at`, which is the query that justifies the whole
change. It is exposed as `GET /me/games` and replaces nothing: the
admin listing under `/games` stays as it is.

## Decision 5 — The Discord link is durable, and the server can act on it

The OAuth callback stores the **refresh token**, encrypted, on the
`identities` row. What that buys, stated precisely:

- **Silent profile refresh.** Names and avatars are re-fetched on a
  timer or on next sign-in without a consent screen, so a renamed
  friend does not stay stale on every seat forever.
- **Future scopes** without redesigning the link. `guilds` for
  "which servers do we share" is the likely first.

What it does **not** buy, and is not needed for: **direct-message
invites.** A bot can open a DM with any user who shares a guild with
it using only the bot token and the user's snowflake. The user's own
tokens are irrelevant to that path. This is spelled out because the
question that produced this decision listed DMs as depending on the
refresh token, and they do not.

**Encryption at rest:** AES-256-GCM with a key from
`CMDCTRL_IDENTITY_KEY` in `/etc/cmd_and_ctrl/env`, beside the admin
token and the HMAC key from #517. The same "missing key must not
fail open" rule as ADR 0044 decision 3: with the key absent, sign-in
still works, the refresh token is **discarded** at the callback, the
row records `refresh_token = NULL`, and the boot log warns naming the
variable. Nothing is ever stored in the clear.

**Direct-message invites** therefore need one more secret on the
*server*, `CMDCTRL_DISCORD_BOT_TOKEN`, and a plain REST call — open a
DM channel, post a message — from the same stdlib HTTP client ADR
0004 already uses for OAuth. The gateway bot binary does not send
DMs. `POST /games/{id}/invites/dm` takes a target
`user_id` and requires the caller to be seated in, or the creator of,
the game. The message carries the ordinary invite link; nothing new
is minted. Rate-limited per caller like every invite-adjacent route.

**Amended at acceptance (2026-09-16):** the bot *does* gain a
`/cc-invite-dm @user` slash command (#613), carried over from ADR 0004
and S12.5 (#59). It is a thin client of this route: it creates the game
with its existing admin credentials and calls
`POST /games/{id}/invites/dm`. It never opens a DM from its gateway
session. There is still exactly one place that builds and sends an
invite DM, and it is this route.

## Decision 6 — Revocation comes back, per user

ADR 0044 decision 3 accepted that a stateless credential cannot be
withdrawn by forgetting it and left `Revoke` advisory. The `users`
table makes a bounded, honest version cheap:

`users.sessions_invalid_before` is a timestamp. `Validate` on any
principal with a `UserID` compares the token's `IssuedAt` against it;
older tokens are rejected. Logout-everywhere, and an admin removing a
person, set it to now. The read is one indexed lookup per request on
a table of eight rows, cached in memory and invalidated on write, so
the hot path — a WS frame — never touches the database.

Admin and guest tokens have no `UserID` and stay exactly as ADR 0044
left them: short TTL, advisory revoke. That is recorded on the type.

## Decision 7 — Decks belong to people, and are stored as text

```sql
decks (
  id            TEXT PRIMARY KEY,
  owner_id      TEXT NOT NULL REFERENCES users(id),
  name          TEXT NOT NULL,
  source_format TEXT NOT NULL,               -- moxfield | text
  source_text   TEXT NOT NULL,               -- what the player pasted
  commanders    TEXT NOT NULL,               -- json array of names, for listing
  card_count    INTEGER NOT NULL,
  created_at    INTEGER NOT NULL,
  updated_at    INTEGER NOT NULL
)
```

**Source text, not the parsed card list.** The catalog changes weekly
and ADR 0044 decision 7 already worries about oracle IDs moving
between dumps. A deck stored as resolved cards would drift silently;
a deck stored as the pasted text is re-parsed through `deck.Parse` and
validated against the current catalog at the moment it is seated, and
a card that no longer resolves surfaces as a validation error the
player can see and fix. Parsing a hundred lines is microseconds.

`POST /games/{id}/decks` is unchanged for guests. For a signed-in
user it additionally creates or updates a `decks` row, and a new
`POST /games/{id}/decks/{deck_id}` seats a library deck without
re-pasting. `GET /me/decks` lists them. No deck builder ships here;
Moxfield and a text box are the editors, as they are today.

## Decision 8 — Tablemates are a query, not a table

"Invite friends" means: people you have shared a table with, offered
as suggestions when you invite. That is

```sql
SELECT DISTINCT s2.user_id
FROM seats s1 JOIN seats s2 ON s1.game_id = s2.game_id
WHERE s1.user_id = ? AND s2.user_id IS NOT NULL AND s2.user_id <> ?
```

with a recency order. No friend requests, no acceptance, no
`friendships` table. An explicit list, and the UI a social graph
drags in, is deferred until the group is not eight people who already
know each other. Should it ever be wanted, it is one table on top of
this and changes nothing below.

## Migration

One import, at boot, when the database has no `games` rows and
`<dumpDir>/lobby/` has files:

1. Each `lobby/<id>.json` becomes a `games` row, its `SeatInfo`
   entries become `seats` rows, and its two tokens become `invites`
   rows (hashed). A seat whose `SeatInfo` carries a `DiscordID` gets
   `user_id = NULL` **and** a `pending_discord_id` side column, because
   no `users` row exists yet.
2. The files are renamed to `lobby/<id>.json.imported`, never deleted
   by the importer. A later sweep removes them once the owner is
   satisfied; ADR 0041's habit of leaving the file so a rollback can
   find it applies.
3. On a person's first Discord sign-in after the migration, every seat
   with a matching `pending_discord_id` is linked to the new user and
   the side column cleared. History appears retroactively, which is
   the point of storing the snowflake on `SeatInfo` in the first place.

The engine restore points are not touched. A restored game's `Room`
finds its metadata in the database instead of in a file, through the
same `Lobby` methods it calls today.

**Rollback:** a binary from before this ADR ignores the database and
reads `lobby/*.json`, which the importer left in place under a
different name. A short shell one-liner to rename them back is
documented in `docs/environments.md` alongside the existing
"resetting preview state" recipe.

## Sub-PRs

One PR each, in this order:

| # | Scope | Depends on |
|---|---|---|
| 0 | This ADR, `docs/sprints.md`, the tracking issue | — |
| 1 | `internal/db`: open, WAL, migrations runner, `VACUUM INTO` timer; backup lands in HomeLab and is rehearsed | — |
| 2 | `users` + `identities`; OAuth callback writes them; `Principal.UserID` (coordinated with #517); refresh token encryption | 1, #517 |
| 3 | `games` / `seats` / `invites` replace `lobby/*.json`; importer; hashed invite lookup | 1 |
| 4 | `GET /me/games` and a "My games" view; seat linking on first sign-in; linking Discord to an already-held seat mid-game (`GET /auth/discord/link`, carried over from S12.5 #59) | 2, 3 |
| 5 | `decks`: library rows from upload, seat-from-library route, `GET /me/decks` | 2, 3 |
| 6 | Tablemates query, invite picker, `POST /games/{id}/invites/dm` via bot REST; the `/cc-invite-dm` slash command that calls it (#613) can follow separately | 4 |
| 7 | `sessions_invalid_before`, logout-everywhere, admin remove-user | 2 |

Sub-PRs 2 and 3 are independent of each other and can run in
parallel. Nothing here blocks S33's remaining sub-PRs, and sub-PR 2
should land *after* #517 rather than race it.

### Implementation notes — sub-PR 2 (users and identities)

Recorded where the code settled a detail the decisions above left
open, or did something differently from how they read.

- **The person columns are rebuilt into real foreign keys.** Migration
  0002 (sub-PR 3) left `games.created_by`, `invites.created_by` and
  `seats.user_id` as plain TEXT because `users` did not exist yet.
  Migration 0003 creates `users` and `identities` and rebuilds those
  three tables with `REFERENCES users(id)`, keeping every row and
  index. `seats.deck_id` stays unenforced until `decks` exists (sub-PR
  5).
- **Migrations run with `foreign_keys` off, then checked.** SQLite's
  documented table-rebuild procedure needs `PRAGMA foreign_keys = OFF`,
  and that pragma is a silent no-op inside a transaction. With it on,
  `DROP TABLE games` fires `ON DELETE CASCADE` and deletes every seat
  and invite. `PRAGMA defer_foreign_keys` defers the check, not the
  cascade, so it does not help. The runner now pins one connection,
  switches enforcement off before `BEGIN`, runs the migration and
  `PRAGMA foreign_key_check` in the transaction (a violation rolls the
  migration back), and switches enforcement back on before the
  connection returns to the pool. This applies to every migration, not
  only 0003. A dangling reference is still refused. It is caught once
  at the end of the migration instead of row by row.
- **Key format.** `CMDCTRL_IDENTITY_KEY` has the same format as
  `CMDCTRL_SESSION_KEY`: a random string of at least 32 bytes, used as
  text rather than decoded. The AES-256 key is SHA-256 of a fixed label
  followed by that string. A sealed value is a version byte (`1`), then
  a 12-byte random nonce, then the ciphertext and GCM tag. The
  identity's `provider` and `subject` are bound in as associated data,
  so a blob copied to another row does not open. The key must differ
  from the admin token and the session key, and the boot fails
  otherwise.
- **Key absent.** The refresh token is discarded and `refresh_token`
  is written NULL, including over a value sealed while a key was
  configured. With a key, a sign-in that returns no refresh token
  leaves the stored one in place.
- **The OAuth exchange did not keep a refresh token before this
  sub-PR.** Discord returns one on every authorization-code grant.
  `TokenResponse` now reads it. The requested scope is unchanged
  (`identify`). `identities.scopes` records the scope Discord reports
  granting, falling back to the requested one if Discord omits it.
- **`users.avatar_url`** is the same-origin proxy path
  `/avatars/<snowflake>/<hash>.png` that the client already renders,
  not a Discord CDN URL. It is NULL for an account with no custom
  avatar. Both `users.display_name` and `identities.display_name` hold
  Discord's global name, falling back to the username.
- **A failed user-store write fails the sign-in** with a 500. It does
  not degrade to a session with a zero `UserID`. With no database at
  all (`CMDCTRL_DATA_DIR` empty) the store is a no-op, sign-in works,
  and sessions carry a zero `UserID`, as before this ADR.
- **"Create a game needs a signed-in user" (decision 2) is already
  true.** `POST /games` is admin-only. The only other ways a game comes
  into being are `CMDCTRL_SEED_DEMO`, which is boot-time and bypasses
  the lobby, and the one-time `lobby/*.json` import. No non-admin path
  exists, so nothing new is enforced. `games.created_by` (and both
  invites' `created_by`) is taken from the creating principal's
  `UserID`, which is NULL for an admin. Opening the route to signed-in
  users later is a change to its middleware and nothing else.
- **Left for later sub-PRs.** `seats.user_id` is still never written:
  linking a seat to its person is sub-PR 4, alongside
  `pending_discord_id`. `sessions_invalid_before` exists but nothing
  reads it until sub-PR 7. Decision 3's 30-day `identified` TTL
  (`CMDCTRL_IDENTITY_TTL`) is not in any sub-PR's scope yet. Identity
  sessions still use `CMDCTRL_SESSION_TTL`.

### Implementation notes — sub-PR 7 (per-user revocation, identity TTL)

Decision 6 in full, and decision 3's identity TTL, which sub-PR 2 left
unclaimed.

- **Where the check lives.** `auth.WithRevocation(inner,
  list)` wraps the HMAC authenticator (or the in-memory one). Its
  `Validate` runs the inner check first, so a bad signature or an
  expiry is reported as such. It then refuses a principal with a
  non-zero `UserID` whose `IssuedAt` is **at or before** that user's
  watermark, with the new sentinel `auth.ErrRevokedCredential`. The
  middleware answers `401 "session revoked"`. `auth` defines only the
  one-method `RevocationList` interface and has no database import.
  `users.Revocations` implements it. `main` wraps the authenticator
  before anything else takes a reference to it, so the lobby routes,
  `POST /join`'s optional session and the WS upgrade authorizer all
  validate through the same wrapper. Issue and Revoke pass through, so
  a single token's `Revoke` stays advisory.
- **"At or before", not "before".** The request that revokes carries a
  token issued earlier, and a token minted in the same millisecond as
  the revocation is refused rather than let through on a tie. The only
  cost is that a sign-in in that same millisecond has to sign in again.
- **The cache.** `users.NewRevocations` reads every non-zero
  `sessions_invalid_before` once at boot. A failed read stops the
  boot, because booting without the watermarks would accept revoked
  sessions. After that, `Revoked` is a map lookup under a read lock,
  and a user not in the map has watermark 0. That is also true of any
  user created after boot, since the column defaults to 0. The request
  path never touches SQLite: not a WS frame, not a WS upgrade, not an
  HTTP request. `RevokeAll` writes the row
  (`MAX(sessions_invalid_before, now)`, so it never moves back) and
  then replaces the cache entry with the value written. The write is
  the only event that can change the answer, so the cache is
  invalidated exactly when it goes stale. A failed write leaves the
  cache alone and the route returns 500. This is sound because the
  server is the database's only writer (decision 1) and `RevokeAll` is
  the only code that writes the column. A hand-run `UPDATE` on a live
  server is not seen until the next restart.
- **Open WebSockets are closed.** A token is validated once, at the
  upgrade, so revocation alone would leave an open socket playing on.
  `ws.Binding` now carries the session's `UserID` and `IssuedAt`, which
  the lobby's `WSAuthorizer` fills in from the principal. The new
  `Hub.EvictUserSessions(user, before)` closes every socket of that
  user opened with a session issued at or before the watermark, the
  same rule as `Validate`. It sends close `1000 "session revoked"`,
  which is terminal on the client (ADR 0044 decision 2), so the client
  does not redial with a dead token. It is the same shape as
  `EvictGame`: one scan of the client map under the read lock, which
  holds a playgroup's sockets. Both revocation routes call it, and the
  admin route reports how many sockets it closed.
- **Routes.** `POST /logout/everywhere` sits beside `POST /logout`.
  Unlike `/logout` it needs a valid session: it acts on the caller's
  own `UserID` and has no way to name another user. A session with no
  user is refused with 403, and `/logout` is the sign-out for those.
  It returns 204 and clears the cookie. The admin route is
  `POST /admin/users/{id}/revoke-sessions`. It is named for what it
  does, because "remove" is revocation only here. No row is deleted,
  and the user can sign in with Discord again straight away. Keeping
  someone out for good would need a flag checked at sign-in, and that
  is not part of this decision. `/logout/*` was added to
  `deploy/Caddyfile`'s `@api` matcher, since Caddy's `/logout` matches
  that exact path only. `/logout` was also missing from the Vite dev
  proxy and has been added. The service worker's prefix match already
  covered both.
- **With no database** there is no revocation list. The authenticator
  is not wrapped, sessions carry no `UserID`, logout-everywhere answers
  403 because the session has no user, and the admin route answers 503.
  Everything else behaves as it did before this ADR.
- **Identity TTL.** `CMDCTRL_IDENTITY_TTL` is a Go duration, default
  `720h`, and invalid or `<= 0` fails the boot. It applies only to the
  `RoleIdentified` session the Discord callback mints on the login
  page, through `lobby.Config.IdentityTTL`. Seat sessions keep
  `CMDCTRL_SESSION_TTL`, including those minted from an identity
  session by `POST /join` or by the callback's invite flow, and so do
  spectator and admin sessions. The session cookie's `Expires` follows
  the token.
- **The client did not drop long sessions at 12 hours, but it did at
  about 23 days.** It stores whatever `expires_at` the server hands it.
  Its expiry timer, though, clamped the `setTimeout` delay to 2·10⁹ ms
  (the API overflows past 2³¹−1 ms) and then cleared the session when
  the clamped timer fired. A 30-day session would have been discarded
  on day 23. The timer now re-arms for the remainder when the session
  is still good. The callback's `#/oauth-complete` fragment now
  carries `user_id`, and the client's principal keeps it. Its presence
  is what shows "log out everywhere" in the lobby header and "sign out
  everywhere" on the login page's signed-in card. That card also gains
  a plain "sign out", since a 30-day identity session needs a way out
  before it joins a table.

## Consequences

- The server gains its first stateful dependency beyond the
  filesystem, and a real backup obligation with it.
- A person exists across games. "My games", a deck library, and
  tablemate suggestions become one query each.
- The `lobby/*.json` artifact from ADR 0041 is retired; the engine
  artifacts are untouched and ADR 0041/0044's guarantees are
  unchanged.
- Invite tokens are hashed at rest and gain expiry and revocation.
- Revocation is real again for anyone with a user row (decision 6).
- Guests lose nothing they have today, and gain nothing.
- The bot binary is unchanged; the *server* gains a bot token and a
  REST path to Discord for DMs.
- The principal shape changes once, inside #517, not twice.

## Deferred

- **Collapse `player` sessions into `identified`.** One durable
  credential per person; the WS upgrade names a game and the `seats`
  table says whether you may sit. Removes the `Discord*` fields from
  `Principal` and most of the reclaim machinery. Do this after S33 and
  S34 have both settled.
- **A second identity provider — specifically Microsoft Entra External
  ID.** Deferred on purpose, not for lack of time. The `identities`
  table is shaped for it and that is the whole of the investment made
  here. It is not adopted now because the playgroup arrives through
  Discord: Discord supplies the names, the avatars, the channel the
  invite is pasted into and the bot, and Entra cannot federate Discord
  cleanly (Discord's OAuth2 is not OpenID Connect and is not a built-in
  social provider there), so in practice Entra would become a Microsoft
  or email login with Discord demoted to a linked account — the reverse
  of how people actually show up. It would add a tenant to administer,
  a second consent screen for every player and a third way in on the
  login page, in exchange for MFA, account recovery and conditional
  access, which are public-launch problems this project has ruled out
  building for. **Triggers that would reverse this:** (a) a wish for
  single sign-on across several self-hosted apps, where one Entra
  tenant beats one Discord application per project; (b) the group
  stops being people who know each other, so recovery and abuse
  controls start to matter; (c) wanting to exercise Entra for reasons
  outside this repo. Only (a) is likely. When it fires: register the
  OIDC client, add `provider = 'entra'` rows, and put a second button
  on the login page — nothing below the `identities` table changes.
  Note the product: Azure AD B2C stopped accepting new tenants in 2025;
  Entra External ID is the current name.
- **Explicit friends list.** One table over decision 8 if wanted.
- **Stats and ratings** — win rates per commander, ELO. The rows to
  compute them from now exist; the queries are not written.
- **Spectator accounts** — spectator sessions stay token-only.
- **Guest-to-user upgrade** — claiming a past guest seat after signing
  in. Needs a proof the guest seat can offer, which today it cannot.

## Alternatives considered

**Keep files; add a `users/` directory.** One JSON per user with a
list of game IDs would answer "my games" until the first cross-cutting
query (tablemates) or the first invariant (a seat links to a user
that exists). It reinvents foreign keys as a sweep script, which is
the failure mode named at the top.

**Postgres.** Rejected in decision 1. It would be the right call the
day there are two hosts, and it is strictly worse before then: a unit
to run, credentials to manage, a network hop on every request, and a
second thing to back up. Moving from SQLite later is a data export
and a driver swap behind the same `database/sql` interface.

**Persist sessions to the database instead of HMAC.** ADR 0044
decision 3 already chose stateless for the reasons given there.
Decision 6 gets revocation back without a session table.

**Discord ID as the user primary key.** Simpler today, and it forecloses
the second provider the owner has already asked about. One join is the
price of not re-keying every table later.

**Store parsed decks.** Rejected in decision 7; source text is what
the player owns and what survives a catalog change.
