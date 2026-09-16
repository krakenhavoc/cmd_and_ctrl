# ADR 0051 — A persistent user database: people, their games, and their decks

**Status:** Proposed · 2026-09-16 · Sprint S34 (proposed) · Tracking issue [#607](https://github.com/krakenhavoc/cmd_and_ctrl/issues/607)
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
0004 already uses for OAuth. The gateway bot binary is not involved
and is not changed. `POST /games/{id}/invites/dm` takes a target
`user_id` and requires the caller to be seated in, or the creator of,
the game. The message carries the ordinary invite link; nothing new
is minted. Rate-limited per caller like every invite-adjacent route.

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

Proposed shape, one PR each, in this order:

| # | Scope | Depends on |
|---|---|---|
| 0 | This ADR, `docs/sprints.md`, the tracking issue | — |
| 1 | `internal/db`: open, WAL, migrations runner, `VACUUM INTO` timer; backup lands in HomeLab and is rehearsed | — |
| 2 | `users` + `identities`; OAuth callback writes them; `Principal.UserID` (coordinated with #517); refresh token encryption | 1, #517 |
| 3 | `games` / `seats` / `invites` replace `lobby/*.json`; importer; hashed invite lookup | 1 |
| 4 | `GET /me/games` and a "My games" view; seat linking on first sign-in | 2, 3 |
| 5 | `decks`: library rows from upload, seat-from-library route, `GET /me/decks` | 2, 3 |
| 6 | Tablemates query, invite picker, `POST /games/{id}/invites/dm` via bot REST | 4 |
| 7 | `sessions_invalid_before`, logout-everywhere, admin remove-user | 2 |

Sub-PRs 2 and 3 are independent of each other and can run in
parallel. Nothing here blocks S33's remaining sub-PRs, and sub-PR 2
should land *after* #517 rather than race it.

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
