# Lobby HTTP API (S04)

REST surface for game creation, seat claiming, and admin bootstrap.
All responses are JSON. All errors use a uniform shape:

```json
{ "error": "<human-readable message>" }
```

Authentication transport and the full credential-precedence rules
live in [ADR 0003 — Auth and lobby architecture](decisions/0003-auth-and-lobby.md).

---

## Public routes (no credential required)

### `POST /admin/login`

Exchange the shared admin token for an admin session.

**Request**

```json
{ "token": "<CMDCTRL_ADMIN_TOKEN>" }
```

**Response 200**

```json
{
  "token": "<opaque base64url>",
  "expires_at": "2026-04-14T08:00:00Z",
  "principal": {
    "role": "admin",
    "admin_id": "<uuid>",
    "issued_at": "2026-04-13T20:00:00Z",
    "expires_at": "2026-04-14T08:00:00Z"
  }
}
```

Also sets the `cmdctrl_session` HttpOnly cookie.

**Errors**

| Status | Reason |
|---|---|
| 401 | invalid admin token |
| 503 | admin login disabled (no `CMDCTRL_ADMIN_TOKEN` set) |

### `POST /games/{id}/join`

The invite-link flow. An invite token minted at game creation is
enough to claim a seat and mint a RolePlayer session.

**Request**

```json
{
  "invite_token": "<token from invite URL ?t= param>",
  "name": "Alice"
}
```

**Response 200**

```json
{
  "token": "<session token>",
  "expires_at": "2026-04-14T08:00:00Z",
  "principal": {
    "role": "player",
    "game_id": "<game uuid>",
    "player_id": "<new player uuid>",
    "name": "Alice",
    "issued_at": "2026-04-13T20:00:00Z",
    "expires_at": "2026-04-14T08:00:00Z"
  },
  "game": { "...GameMeta without invite_token..." },
  "player_id": "<new player uuid>"
}
```

Sets the `cmdctrl_session` cookie. Subsequent calls to `/ws?game=<id>`
automatically bind to this seat.

The session is **optional** here too (S34 sub-PR 4,
[ADR 0051](decisions/0051-user-database.md)). When the request carries
a signed-in person's session, that person takes the seat as
themselves: the seat gets the session's Discord name and avatar,
`seats.user_id` is their users row, `name` in the body is ignored, and
the minted principal carries `user_id` and the `discord_*` fields. A
signed-in session is an `identified` one, or a `player` one with a
non-nil `user_id` (the same person at their next table). Any other
session, and a credential that no longer validates, joins by `name`
exactly as before.

**Errors**

| Status | Reason |
|---|---|
| 400 | empty name |
| 401 | invite token did not match |
| 404 | game not found |
| 409 | game already started, game full, or the signed-in person already holds a seat at this table |

---

### `POST /join`

The login-page flow: the same seat claim as above, but the caller has
only an invite **code** and no game id — the shape you get when
somebody pastes a code out of a Discord message rather than clicking a
link. The server resolves the table from the code itself
(`Lobby.FindByInvite`).

Player invites only; a spectator code does not resolve here. Archived
tables are skipped, so a stale code from an old chat message reads as
expired rather than quietly reopening one.

**Request**

```json
{
  "invite_token": "<invite code>",
  "name": "Alice"
}
```

The session is **optional**, and decides where the seat's identity
comes from:

| Session attached | Behaviour |
|---|---|
| `identified` (Discord sign-in, no seat yet) | Name and avatar come from the Discord identity; `name` in the body is ignored |
| none, or a credential that no longer validates | Classic manual join — `name` is required |
| `player` / `admin` / `spectator` | 409 — that session already belongs somewhere |

**Response 200** — identical to `POST /games/{id}/join`, cookie
included. On the Discord path the principal also carries `discord_id`,
`discord_username`, `discord_global_name` and `discord_avatar_hash`,
and `user_id` is the signed-in person's users row, copied from the
identity session (S34, [ADR 0051](decisions/0051-user-database.md)
decision 3). `user_id` is the nil uuid for a guest seat, for admin and
spectator sessions, and for everyone on a deployment with no database.

The identity session stays valid afterwards: it is how the same person
joins a second table later without signing in to Discord again. On its
own it can do nothing else — the WS authorizer refuses it outright.

**Errors**

| Status | Reason |
|---|---|
| 400 | empty name on the anonymous path |
| 401 | no live table has that invite code (an archived one reads the same way) |
| 409 | game already started, game full, a session that already belongs to a table, or the signed-in person already holds a seat at this table |

Since S34 sub-PR 4 a seat claimed on the Discord path records its
person in `seats.user_id`. One person holds at most one seat per
table, because reclaiming a seat by user
([`POST /me/games/{id}/session`](#post-megamesidsession)) looks up
"the seat whose `user_id` is yours".

---

### `POST /games/{id}/reclaim`

Redeem a **seat-reclaim ticket**: a one-shot, short-lived credential
minted by an admin (see
[`POST /games/{id}/seats/{player_id}/reclaim`](#post-gamesidseatsplayer_idreclaim))
that returns one specific player to their own seat in a game that has
already started.

Public by necessity — the caller has lost their session, which is the
whole problem — so the ticket **is** the credential. Rate-limited in
the same bucket as `/join` and `/spectate`.

**Request**

```json
{ "ticket": "<the ?t= value from the reclaim link>" }
```

There is deliberately no `name` field: the seat already has one, and
accepting a label here would make reclaim a way to rename another
player.

**Response 200** — the same `sessionResponse` shape as `/join`: a
RolePlayer session bound to the seat's `(game_id, player_id)`, with
the seat's Discord identity copied onto the principal so the avatar
renders exactly as it did before the disconnect. Both invite tokens
are stripped from the embedded `game`.

**Errors**

| Status | Reason |
|---|---|
| 401 | unknown, expired, already-redeemed, or wrong-table ticket — all four answer identically, on purpose |
| 403 | the seat left the table between minting and redemption |
| 404 | game not found |
| 409 | the table has been archived |

### `GET /roadmap`

The public engine roadmap ([ADR 0092](decisions/0092-public-roadmap-and-site-portal.md)):
every keyword, mechanic and engine gap in `server/internal/roadmap`'s
registry, with a status a player can read. The client's `#/roadmap`
page renders it. No session is needed, and none is read.

It is public, while `GET /catalog` is not, because of what the body
holds (ADR 0092 Decision 1): card **names** and caveat sentences this
repository wrote. It never holds card art, an image URL, a Scryfall
printing ID or oracle text. `roadmap`'s handler tests fail if an
`image`, `oracle_text`, `scryfall_id` or `engine_notes` key appears
anywhere in the body.

The body is fixed for the life of the binary, so the server encodes it
once. `Cache-Control: public, max-age=300`, the same as `/catalog`.
`GET` and `HEAD` only; any other method gets 405.

**Response 200**

```json
{
  "counts": {
    "implemented": 90,
    "partial": 13,
    "missing": 9,
    "cards": { "total": 2378, "full": 2039, "caveats": 282, "unreviewed": 57 }
  },
  "next": ["abilities-granted-to-other-permanents", "extra-combats"],
  "items": [
    {
      "slug": "extra-combats",
      "name": "Extra combat and main phases",
      "kind": "seam",
      "status": "missing",
      "summary": "Spells and abilities that give you an additional combat phase, ...",
      "missing": "The turn's phases are fixed, so nothing can add another combat or main phase yet.",
      "issue": 753,
      "unblocks": 28,
      "waiting": ["Relentless Assault", "Aggravated Assault", "Seize the Day"]
    },
    {
      "slug": "cascade",
      "name": "Cascade",
      "kind": "mechanic",
      "status": "implemented",
      "summary": "When you cast a spell with cascade, reveal cards ...",
      "rules": ["702.85"],
      "examples": [
        {
          "name": "Bloodbraid Elf",
          "oracle_id": "3f0c9466-5ab9-4205-a84f-b4b27b5a678e"
        }
      ]
    }
  ]
}
```

- `counts.implemented` / `partial` / `missing` count registry items.
  `counts.cards` is the card catalogue's own tally by declared
  completeness: the numbers the signed-in `#/catalog` page shows,
  published as bare numbers.
- `next` is the "up next" list, as slugs in order: pinned items first,
  then unfinished items ranked by `unblocks` plus the number of
  `waiting` cards, at most six.
- `kind` is `keyword`, `mechanic` or `seam` (an engine gap). `status`
  is `implemented`, `partial` or `missing`.
- `missing` says what does not work yet; absent on an implemented item.
- `examples` (at most three) are fully automated cards that use the
  item. `partial_examples` (at most three) are cards that work with a
  gap this item explains, each with its first caveat. `waiting` are
  cards known to be blocked on it.
- `issue` is the tracking issue on the public repository and `adr` the
  decision record's file name. Either may be absent, as may `rules`,
  `pin` and `unblocks`.

### `POST /deck-coverage`

The deck coverage checker ([ADR 0095](decisions/0095-deck-coverage-and-deck-requests.md)
§1–§2): how much of a decklist the engine automates. The site's
`#/deck-check` page and the bot's `/c2-deck-check` both render this
response. No session is needed.

**Request** — exactly one of:

```json
{ "url": "https://moxfield.com/decks/AbC123" }
{ "text": "1 Atraxa, Praetors' Voice *CMDR*\n1 Sol Ring\n..." }
```

`url` must be a Moxfield or Archidekt deck link. The server fetches it
itself, with `deck.FetchFromURL`'s 10-second timeout and 2 MiB cap, and
always fetches the deck's canonical link, never the caller's query
string. `text` is the same plain-text list `POST /games/{id}/decks`
accepts. Nothing else is fetched: an unsupported host is refused before
any outbound request.

**Limits.** The route fetches a third-party URL for anonymous callers,
so it is limited per client IP to about one check every 10 seconds,
with a burst of 3 (429 past that; the key follows
`CMDCTRL_TRUST_FORWARDED`). An **admin** session has its own bucket of
1/s, burst 10, because the bot calls from loopback on behalf of every
guild member at once. Any other session is ignored. A link report is
cached for 10 minutes under its deck key, so a second check of the same
deck, however the link is spelled, fetches nothing. A pasted list is
never cached.

**Response 200**

```json
{
  "deck_name": "Needy Deck",
  "source": "moxfield",
  "source_url": "https://moxfield.com/decks/AbC123",
  "deck_key": "moxfield:AbC123",
  "commanders": ["Atraxa, Praetors' Voice"],
  "counts": { "manual": 12, "unreviewed": 1, "caveats": 6, "automated": 40, "no_effect": 29 },
  "cards": [
    { "name": "Doubling Season", "oracle_id": "…", "count": 1, "bucket": "manual" },
    { "name": "Abzan Charm", "oracle_id": "…", "count": 1, "bucket": "caveats",
      "caveats": ["…"] },
    { "name": "Forest", "oracle_id": "…", "count": 7, "bucket": "no_effect" }
  ],
  "unknown": ["Some Misspelled Card"],
  "violations": [
    { "code": "wrong_card_count", "message": "deck has 99 cards; expected 100" }
  ]
}
```

- `bucket` is one of five, decided in one place
  (`server/internal/deckcoverage`):
  - `manual`: prints rules the engine will not run. Exactly
    `game.Unimplemented`, the bit behind the stack's `manual` chip.
  - `unreviewed`: catalogued, but not audited against its text.
  - `caveats`: catalogued with declared simplifications. `caveats` is
    the catalogue's player-facing sentences.
  - `automated`: catalogued and complete.
  - `no_effect`: nothing to automate. A vanilla creature, a card whose
    text is only keywords the engine enforces, or a basic land.
- `counts` is by distinct card, and always carries all five keys.
  `count` on a card is its copies, summed across the command zone and
  the main deck. Sideboard rows are not bucketed.
- `cards` is sorted by bucket in the order above, then by name.
- `unknown` is the names the card index could not resolve; they are in
  no bucket.
- `violations` is `deck.Validate`'s answer. It is informational and
  never blocks a report: a 60-card list still gets its buckets.
- `source` is `moxfield`, `archidekt` or `text`. `source_url` and
  `deck_key` are absent for `text`.
- `commanders`, `cards`, `unknown` and `violations` are always arrays,
  never `null`.

The body carries names, oracle IDs, buckets and caveat sentences, and
never card art, an image URL or oracle text: the line ADR 0092 drew for
`GET /roadmap`.

**Errors**

A link that cannot be read answers with a sentence a player can act on,
the typed code, and the same `violations[]` shape a deck upload uses:

```json
{
  "error": "That deck is private. Make it public or unlisted on the deck site, or paste the list as text.",
  "code": "deck_private",
  "violations": [{ "code": "deck_private", "card": "<the url>", "message": "…" }]
}
```

| Status | `code` | When |
|---|---|---|
| 400 | `unknown_source` | Not a Moxfield or Archidekt deck link |
| 404 | `deck_not_found` | The deck site answered 404 |
| 422 | `deck_private` | The deck is private |
| 422 | `unreadable_deck` | The deck was fetched but is not a list the importer can read |
| 502 | `upstream_blocked` | The deck site's CDN is blocking this server |
| 502 | `external_api_unavailable` | The deck site did not answer |

The upstream failures are 502 rather than 4xx: the caller's link was
fine. Other errors use the uniform `{error}` shape: 400 for a body with
neither or both of `url` and `text`, or an unreadable pasted list; 503
when the card index is not loaded.

## Authenticated routes

### `POST /games` *(admin only)*

Create a new game.

**Request**

```json
{ "name": "Friday Night Magic", "host_discord_id": "123456789012345678" }
```

`host_discord_id` is optional. It names the table host by Discord user ID
([ADR 0075 §2.1](decisions/0075-table-settings-and-host-controls.md)). The
Discord bot's `/c2-invite` sends the user who ran it. The ID is held on the
table, unserved, until that Discord identity claims a seat through the OAuth
join. That seat then becomes host. See [The table host](#the-table-host).

**Response 201**

```json
{
  "id": "<game uuid>",
  "name": "Friday Night Magic",
  "created_at": "2026-04-13T20:00:00Z",
  "invite_token": "<16-byte base64url>",
  "spectator_invite": "<16-byte base64url>",
  "players": [],
  "state": "lobby"
}
```

This response is the one reliable place to get the invite links. The
server stores only their hashes, so after a restart it cannot show them
again (see `GET /games/{id}`).

**Errors**

| Status | Reason |
|---|---|
| 401 | unauthenticated |
| 403 | session is not RoleAdmin |
| 400 | empty name |

### The table host

Every table has at most one **host**: the seat that may manage the table
(settings, and later spawning) alongside the server admin
([ADR 0075 §2.1](decisions/0075-table-settings-and-host-controls.md)).

- **A named host** (`host_discord_id` on `POST /games`) hosts as soon as that
  Discord identity claims a seat.
- **Otherwise the first human seat to join hosts.** When the named host has
  not arrived yet, the first human hosts in the meantime and hands over when
  they sit down.
- **A bot seat never hosts.** A table with only bots has no host.
- **Transfer**: `POST /games/{id}/host` (below).
- **What hosting lets you do**: change the table's settings, through
  `PATCH /games/{id}/settings` (below) or the `set_table_settings` WebSocket
  action. The deprecated `set_undo_limit` action is gated the same way.
- **The host leaving passes it on.** When the host concedes or loses
  ([ADR 0060](decisions/0060-leaving-the-game.md)), hosting passes to the
  next human seat in turn order, skipping bots and departed seats and
  wrapping around the table. If no human is left, nobody hosts and only the
  admin can manage the table. The pass is permanent: an undo that brings the
  old host back does not hand the table back.

`GameMeta.host_player_id` is the host's player ID. When there is no host it
is the zero UUID (`00000000-0000-0000-0000-000000000000`). Each seat in
`players` carries `is_host: true` on the host's seat, and the game view's
`PlayerView.is_host` matches it ([protocol.md](protocol.md)). Both the host
and a pending named host are stored on the game's `games` row, in the
`host_player_id` and `host_discord_id` columns added by migration 0004. They
survive a restart, and a pass that happened while the server was up is
written to the row as soon as the lobby sees it.

"Host or admin" is one predicate, `lobby.CanManageTable(principal, meta)`.
It is true for `RoleAdmin`, or for a `RolePlayer` session bound to this game
whose `player_id` is `host_player_id`. It is false for spectators, unseated
Discord sign-ins, other seats, and the host of a different table.

**Not the same thing as the game's creator.** `games.created_by` (ADR
0051 decision 2) is whoever called `POST /games` while signed in, and
is a *different* predicate, `lobby.CanRotateInvites` — used only by
`POST /games/{id}/invites/rotate` and the Discord bot's `/c2-end` host
check, below. A table's creator need not ever sit down (no seat, no
`is_host`), and a seated host need not be the creator — the first
human to join hosts by default regardless of who created the table.

### `POST /games/{id}/host`

Transfer hosting to another seat. Host or admin only.

**Request**

```json
{ "player_id": "<uuid of the new host>" }
```

**Response 200**: the updated `GameMeta`, with `host_player_id` and the
`is_host` flags moved. Connected clients get a fresh snapshot with the new
`is_host`.

An explicit transfer also clears a pending named host, so a late `/c2-invite`
claimant does not take the table back.

**Errors**

| Status | Reason |
|---|---|
| 400 | missing or malformed `player_id` |
| 401 | unauthenticated |
| 403 | caller is neither the host of this table nor the admin |
| 404 | game not found |
| 422 | `player_id` is not a seat at this table, is a bot, or has left the game |

### `PATCH /games/{id}/settings`

Change the table's settings
([ADR 0075 §2.3](decisions/0075-table-settings-and-host-controls.md)). **Host
or admin only** — the same `CanManageTable` predicate the transfer route
uses. Valid in the lobby **and** on a live game: a host sitting on the lobby
page of a running table should not have to open the board to turn undos back
on. The WebSocket twin is the `set_table_settings` action
([protocol.md](protocol.md)), which takes the identical body.

**Request** — a *partial*. Only the fields present are applied; a field left
out is untouched, which is why this is a `PATCH` and not a `PUT`.

```json
{
  "undo_limit": 3,
  "undo_scope": "own",
  "starting_life": 40,
  "commander_damage": 21,
  "bot_pace": "normal",
  "allow_spawn": false
}
```

| Field | Range | Effect |
|---|---|---|
| `undo_limit` | `-1` and up | Per-player per-turn undo budget. `-1` is unlimited (never debited, shown as ∞), `0` turns undo off. Takes effect immediately and refreshes every seat's `undos_remaining`. |
| `undo_scope` | `"own"` \| `"host_any"` | Whose entries a seat may take back. |
| `starting_life` | 1..999 | Each seat's life at `start`. In the lobby it also rewrites the already-seated players' totals. **Rejected once the game is active.** |
| `commander_damage` | 1..99 | Damage from one commander that loses the game. Read at the next state-based action check, so lowering it can lose somebody the game at that check. |
| `bot_pace` | `"fast"` \| `"normal"` \| `"slow"` | AI seat pacing preset. |
| `allow_spawn` | bool | Whether the host and admin may spawn cards and tokens on a live table. |

The **whole patch is validated first**, so a patch with one bad field changes
nothing — including the good fields beside it.

**Response 200** — the table's settings *after* the patch, whole, so the
client can render its panel from the answer instead of waiting for the
WebSocket broadcast:

```json
{
  "settings": {
    "undo_limit": 3,
    "undo_scope": "own",
    "starting_life": 40,
    "commander_damage": 21,
    "bot_pace": "normal",
    "allow_spawn": false
  }
}
```

Every connected client also gets a fresh snapshot carrying the new
`settings`, and the game log gains one `settings` line per field that
actually changed ("Luke (host) set undos to 3 per turn"). A settings change
is **not undoable** — it pushes no undo entry, and a later `undo` carries the
settings forward rather than rolling them back.

**Errors**

| Status | Reason |
|---|---|
| 400 | malformed body, or a value outside its range (`undo_limit` below `-1`, `starting_life` outside 1..999, `commander_damage` outside 1..99, an unknown `undo_scope` or `bot_pace`) |
| 401 | unauthenticated |
| 403 | caller is neither the host of this table nor the admin |
| 404 | game not found |
| 422 | `starting_life` changed after the game started — the value has already been applied; use the life controls to adjust totals |

### `POST /games/{id}/spawn`

Put cards or tokens onto a **live** table from nowhere
([ADR 0075 §2.4](decisions/0075-table-settings-and-host-controls.md), which amends
[ADR 0023](decisions/0023-develop-environment.md)). This is the production
spawner and is **not** behind the dev-feature gate — the dev route
`POST /games/{id}/dev/spawn` still exists, unchanged, with its own dev-only,
anyone-at-the-table semantics.

Two gates, both required:

1. **`CanManageTable`** — the caller is the table host or the server admin.
2. **`Settings.allow_spawn`** — the table has switched spawning on. It is
   **off by default** and changed through the table-settings surface.

A refusal says **which** gate fired, because they are different problems with
different fixes: "you are not the host" sends the caller to ask the host, and
"this table has spawning off" sends the host to the settings panel.

Every spawn is **announced in the public game log** ("Luke (host) spawned
2 × Treasure onto Ana's battlefield" — see `spawn` in
[protocol.md](protocol.md)) and is **undoable** with the ordinary undo,
attributed to the spawner and flagged free, so repairing a typo does not cost
the host their per-turn take-back. A battlefield spawn emits `EventETB`, so
enters-the-battlefield triggers fire and the spawn can put abilities on the
stack; that is the point of the feature.

**Request**

```json
{
  "name": "Sol Ring",
  "player_id": "<uuid of the seat that will own and control the cards>",
  "zone": "battlefield",
  "count": 2,
  "commander": false
}
```

Exactly one of three identifies what to make, checked in this order:

| Field | What it names |
|---|---|
| `token` | a token template key, exactly as `GET /games/{id}/spawn/tokens` lists it (`"Treasure"`, `"1/1 white Soldier"`) |
| `scryfall_id` | an indexed Scryfall printing — what the client pins after a `GET /games/{id}/spawn/cards` search |
| `name` | a card name, resolved against the same index |

`zone` is one of `battlefield`, `hand`, `graveyard`, `exile`, `library`,
`command` (never `stack`). `count` defaults to 1 and is capped at 20.
`commander` stamps the card as a commander and is ignored for a token.

**Tokens may only be spawned onto the battlefield.** CR 704.5d removes a token
from every other zone at the next state-based action check, so spawning one
into a hand would appear to work and then silently undo itself. Tokens go
through the same `CreateToken` primitive the card catalog uses, so a spawned
Treasure taps and sacrifices for mana like a real one.

**Response 200**

```json
{
  "spawned": ["<instance uuid>", "..."],
  "name": "Treasure",
  "zone": "battlefield",
  "count": 2,
  "token": true
}
```

`scryfall_id` echoes the printing that was resolved, and is absent for a token
(a token has no Scryfall printing — that is why the dev spawner could never
make one).

**Errors**

| Status | Reason |
|---|---|
| 400 | malformed `player_id`; unknown or missing `zone`; `count` outside 1..20; a token into a zone other than the battlefield; no card or token identifier |
| 401 | unauthenticated |
| 403 | caller is neither the host of this table nor the admin |
| 403 | the table has `allow_spawn` off (message: "spawning is switched off for this table") |
| 403 | `player_id` is not a seat at this table |
| 404 | game not found; no such card in the index; no such token template |
| 409 | the game has not started (spawning into a lobby-state game would be erased by `Start` dealing opening hands) |
| 503 | card index not loaded, or the server was built without token templates |

### `GET /games/{id}/spawn/cards`

Search the Scryfall index for the spawn picker: `?q=<text>&limit=<n>` (capped
at 40), same response shape as the dev spawner's `GET /dev/cards`, so one
client component serves both.

It exists as its own route because `/dev/cards` is behind `requireDevFeature`
and 404s in production, which would leave the production spawner with a name
field and no way to find a name. Gated like the spawn itself, and scoped to a
game for the same proxy-prefix reason as the token list below.

**Response 200**

```json
{ "cards": [{ "id": "<scryfall uuid>", "name": "Sol Ring", "type_line": "Artifact", "mana_cost": "{1}", "set": "lea" }] }
```

**Errors**: 401 unauthenticated · 403 not the host or admin · 404 game not
found · 503 card index not loaded.

### `GET /games/{id}/spawn/tokens`

The token template keys the `token` field of a spawn request accepts, sorted.
Gated exactly like the spawn itself, so the picker is not offered to a seat
that cannot use it.

The answer is the same for every table; it is scoped to a game only so that it
rides the `/games` prefix, which `deploy/Caddyfile`'s `@api` matcher, the Vite
proxy and the service worker's `API_PATH` already carry. A new top-level prefix
would have to be added to all three or it would 404 in production only.

**Response 200**

```json
{ "tokens": ["0/1 colorless Eldrazi Spawn", "1/1 white Soldier", "Treasure", "..."] }
```

Two families share one list: the plain templates from
`server/internal/cards/effects/tokens_table.go`, keyed the way a card prints
them, and the behaviour tokens from `tokens.go` (Treasure, Gold, Food, Clue,
Blood, Powerstone), keyed by name because that is how a card prints those.

### `GET /games`

List the **active** games known to the lobby. Invite tokens are
STRIPPED — listing is not proof of ownership.

Archived tables are excluded. `GET /games?archived=1` returns those
instead (same shape, each carrying `archived_at`), so the default
answer to "what tables are there" does not grow without bound as old
games pile up.

**Response 200**

```json
{
  "games": [
    {
      "id": "<uuid>",
      "name": "FNM",
      "created_at": "2026-04-13T20:00:00Z",
      "players": [{ "player_id": "<uuid>", "name": "Alice", "seat": 0, "is_host": true }],
      "state": "lobby",
      "host_player_id": "<uuid>",
      "is_creator": true
    }
  ]
}
```

`is_creator` ([#1098](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1098))
is computed per **viewer**, not stored: `true` only when the
requesting principal's `UserID` equals this game's `created_by`, and
omitted (`false`) otherwise — including for every other caller
looking at the same game. The raw `created_by` never reaches the
wire. It is what the lobby UI's rotate-invite buttons check (see
`POST /games/{id}/invites/rotate` above); a table's creator need not
even be seated at it, so this can be `true` on a game where `mySeat`
finds nothing.

### `GET /games/{id}`

Full metadata for a single game. The `invite_token` and
`spectator_invite` fields are included only for admins and for players
seated in this specific game.

**They are also absent after a server restart.** Invites are stored
only as SHA-256 hashes (ADR 0051 decision 4, S34 sub-PR 3). The
plaintext exists only in the process that minted the game. A game that
came back from a restart still accepts every link handed out before
it. The server just cannot show those tokens again. The `POST /games`
response is where a creator gets the link. The lobby UI and the
Discord bot both take it from there, and neither ever read it from
this route.

### `POST /games/{id}/start`

Transition the game from `lobby` → `active`. Only admins and players
seated in the game may call this. The server has every seated player roll a
d20, rerolling only tied leaders until one winner remains; that seat takes the
first turn. The rolls and winner are public in the opening-hand view and game
log, and use the game's persisted RNG so reconnects and replay agree.

**Errors**

| Status | Reason |
|---|---|
| 403 | not a seat in this game |
| 404 | game not found |
| 409 | not enough players (min 2), or one or more seats haven't uploaded a deck |

### `POST /games/{id}/decks`

Install a deck for a seat — either an uploaded decklist or one of the
pre-built decks from `GET /decks`. RolePlayer may only set their own
seat's deck (`player_id` in the body must match the session's bound
player); RoleAdmin may set any seat.

**Request** — exactly one of `deck` and `source`:

```json
{
  "format": "text",   // or "moxfield"; empty → auto-detect
  "source": "Commander:\n1 Atraxa, Praetors' Voice\n\nMainboard:\n...\n",
  "player_id": "<uuid>"
}
```

```json
{
  "deck": "izzet-aggro",   // a deck id from GET /decks
  "player_id": "<uuid>"
}
```

`deck` is a field on this route rather than a route of its own so that
there is exactly **one** legality path in the server: the deck id is
expanded to that deck's plain-text decklist and run through the same
parse → resolve → validate → install pipeline a pasted list takes. A
pre-built deck cannot be legal by a rule an uploaded one is not held
to. `POST /games/{id}/seats/bot` has taken the same shape since S31,
for the same reason.

Sending both `deck` and `source` is a 400 rather than a precedence
rule; an unknown deck id is a 422 that names the ids this build has.

When `format` is empty, the server auto-detects from the first
non-whitespace bytes of `source`:

| Prefix | Detected format |
|---|---|
| `http://` / `https://` | `url` (S06.5+) |
| `{` | `moxfield` (JSON paste) |
| anything else | `text` |

### `format: "url"` (S06.5+)

When `format` is `"url"`, `source` is a deck URL the server fetches
and parses upstream. No JSON shape to paste; the server handles the
outbound request and reuses the same validation pipeline as the paste-
based importers. Supported hosts at S06.5:

- `moxfield.com` — `/decks/<id>` or `/decks/<id>/<slug>`
- `archidekt.com` — `/decks/<id>` or `/decks/<id>/<slug>`

Private decks (upstream 401/403) and unsupported hosts surface as
`deck_private` / `unknown_source` violations in the 422 response.

**Response 200**

```json
{
  "game": { "...GameMeta with updated seat..." },
  "deck_name": "Atraxa Superfriends",
  "card_count": 100,
  "deck_id": "izzet-aggro",
  "commanders": ["Atraxa, Praetors' Voice"],
  "warnings": [
    { "code": "sideboard_not_supported_in_commander", "message": "..." }
  ]
}
```

`deck_id` echoes the pre-built deck that was installed and is absent
for an uploaded list.

**Errors**

| Status | Reason |
|---|---|
| 400 | malformed request (neither `deck` nor `source`, both of them, missing player_id, unknown format) |
| 403 | not your seat (RolePlayer with mismatched player_id) |
| 413 | body exceeds the 2 MiB deck-source cap |
| 422 | validation failed — body carries `{"error", "violations": [...]}`; may also carry `warnings` |
| 422 | unknown pre-built `deck` id — body names the ids this build offers |
| 429 | too many requests — 2/s refill with 10-burst per IP |
| 503 | server card index not loaded — run `scripts/scryfall-refresh.sh` |

All 422 failure modes (validation, unknown cards, unsupported
mechanics) share the same response shape. A response with fatals in
`violations` may also include non-fatal advisories under `warnings`:

```json
{
  "error": "deck has validation errors",
  "violations": [
    { "code": "color_identity_violation", "card": "Lightning Bolt", "message": "..." }
  ],
  "warnings": [
    { "code": "sideboard_not_supported_in_commander", "message": "..." }
  ]
}
```

Violation codes (stable strings, keyable by the client):

- `wrong_card_count` — deck is not 100 cards
- `missing_commander` / `too_many_commanders` / `not_a_legal_commander`
- `color_identity_violation` — mainboard card outside commander's
  color identity (carries the offending `card` field)
- `singleton_violation` — non-basic-land card appears more than once
- `not_legal_in_format` — card is banned or not legal in Commander
- `unknown_card` — decklist name that did not resolve against the
  Scryfall index (carries `card`)
- `unsupported_mechanic` — resolved commander uses partner/companion,
  deferred to a later sprint (carries `card`)
- `unknown_source` — URL-based import (S06.5+) whose host is not one
  of the supported deck-builders; carries the URL in `card`
- `deck_not_found` — URL-based import where the upstream returned
  404; the deck was deleted or the ID is wrong
- `deck_private` — URL-based import where the upstream returned
  401/403; only publicly-readable decks are supported at S06.5
- `external_api_unavailable` — URL-based import where the upstream
  timed out or returned 5xx; treat as retry-after-a-bit
- `sideboard_not_supported_in_commander` — **warning only**, delivered
  under `warnings` on both success and 422 responses (the server
  ignores the sideboard either way)

### The deck library (ADR 0051 decision 7, S34 sub-PR 5)

A signed-in caller — a `player` or `identified` session whose
principal carries a non-zero `user_id` (ADR 0051 decision 3) — gets a
saved copy of every deck they paste, so it can be seated again without
re-pasting.

**`POST /games/{id}/decks` saves or updates a library row.** Unchanged
for a guest (no `user_id`): still installs the deck on the seat and
nothing else. For a signed-in caller it *additionally*:

- Creates or updates a `decks` row for **`format: "text"` or
  `"moxfield"` requests only** — not a `deck` (pre-built catalog) pick,
  which has its own id system and is never a library row, and not
  `format: "url"`, whose `source` is a link rather than the decklist
  text the library re-parses later. A URL-based import still installs
  the deck on the seat; it is just never saved to the library.
- Sets the seat's `deck_id` to the saved deck. Any other outcome
  (a guest, a catalog pick, a URL import, or a failed save) leaves the
  seat's `deck_id` empty, clearing a previous one if there was one.

**The update rule:** a caller's existing deck with the **same name**
is updated in place — `source_text`, `source_format`, `commanders`,
`card_count` and `updated_at` all overwrite the stored row, the
`decks.id` does not change, and no duplicate row is created. Anything
else (a new name, or no existing deck by that name) inserts a new row.
`name` is the parsed deck's name (Moxfield exports carry one); a
plain-text paste has none (`ParseText` has nowhere to put one), so it
falls back to the deck's first commander's name. Two different
plain-text pastes with the same commander and no other name therefore
collide under this rule — rename one to keep both, the same as two
Moxfield exports named identically would.

A save (or setting the seat's `deck_id`) that fails for its own
reasons (a transient store error) is logged and does **not** fail the
upload — the deck is already installed on the seat by that point.

**`GET /games/{id}/decks`'s response is unchanged** by any of this —
`deck_id` in the response still means the pre-built catalog id, per
`uploadDeckResponse` above. The library deck's own id isn't echoed
there; find it via `GET /me/decks`.

### `POST /games/{id}/decks/{deck_id}`

Seat a deck already in the caller's library, without re-pasting it.
`{deck_id}` is one of `GET /me/decks`' ids. No request body.

Only a `player` session already seated in this game may call it
(`session.game_id` must match the path), for their own seat — there is
no `player_id` field, unlike `POST /games/{id}/decks`, because a
`player` session already names exactly one seat. Admin sessions are
refused outright (403): the check that matters is deck ownership, and
an admin session never has a `user_id` to own a deck with (ADR 0051
decision 2), so admin access here could never do anything but fail
ownership one step later anyway.

The stored `source_text` is re-parsed through the same
parse → resolve → validate pipeline an upload takes, against the
catalog **this server has loaded right now** — not the catalog at save
time — so a card that stopped resolving (a rename, a ban, a dump that
dropped it) surfaces as the same 422 violation list an upload would
give, rather than silently installing something that changed.

**Response 200** — the same shape as `POST /games/{id}/decks`
(`uploadDeckResponse`), with `deck_id` echoing the library deck's id
(not a pre-built catalog id, though the field is the same one).

**Errors**

| Status | Reason |
|---|---|
| 400 | invalid `deck_id` in the path |
| 403 | caller is not a `player` session seated in this game, or the deck belongs to someone else |
| 404 | game not found, or no such deck id |
| 422 | the stored deck no longer validates against the current catalog — same violation shape as an upload |
| 429 | too many requests — shares `/decks`' rate bucket |
| 503 | card index not loaded, or no deck library configured on this deployment |

### `GET /decks`

The pre-built deck catalog the lobby's deck picker renders from, with
each deck's engine-coverage profile. Game-independent — the same
answer for every table — so the client fetches it once on mount.
Session-gated; there is no seat to install a deck into without one.

Served whether or not a Scryfall dump is loaded: the coverage grades
come out of the effect registry and are true regardless. The honest
failure for a deployment with no dump is the 503 on installing a
deck, not an empty picker implying there are none.

**Response 200**

```json
{
  "decks": [
    {
      "id": "izzet-aggro",
      "name": "Raid and Ransack",
      "archetype": "aggro",
      "summary": "...",
      "commander": "Mary Read and Anne Bonny",
      "colors": ["U", "R"],
      "card_count": 100,
      "coverage": {
        "cards": 88,
        "full": 55,
        "caveats": 22,
        "unreviewed": 11,
        "basics": 2,
        "unregistered": 0,
        "imperfect": [
          { "name": "Farseek", "caveats": ["Only basic lands are found — ..."] },
          { "name": "Thought Vessel", "unreviewed": true }
        ]
      }
    }
  ]
}
```

**Reading the coverage block.** Counts are per distinct card, not per
copy, which is why they do not add up to `card_count`: a deck's ten
basic lands are one or two rows in `basics`, and basics are exempt
because the engine synthesises their mana ability from the type line.

- `full` — plays exactly as printed.
- `caveats` — implemented with a declared simplification; the clauses
  the engine skips are in `imperfect[].caveats`, in player-facing
  language. Same strings `GET /catalog` serves, from the same
  computation (`internal/catalog`), so the two surfaces cannot drift.
- `unreviewed` — registered but never graded. Not a known gap and not
  a clean bill of health; listed in `imperfect` with
  `"unreviewed": true` and no caveat text.
- `unregistered` — no implementation at all. Always `0`: a build test
  in `internal/decks` fails the moment a card in a pre-built deck
  stops resolving to a registered effect spec. It is published anyway
  so that a deck which somehow shipped with one is visible in the
  picker rather than silently counted as fine.

"Every card in this deck is registered" and "every card in this deck
works fully" are different claims, and the client is expected to make
only the one that is true — see `client/src/lib/prebuiltDecks.ts`.

The same four decks back this route and `GET /bot/options`; they come
from one registry (`server/internal/decks`), which is the part that
must not fork.

### `GET /bot/options` (S31)

What the lobby's Add-bot picker renders from: the declared policy
tiers and the curated deck catalog. Game-independent — the same
answer for every table — so the client fetches it once.

**Response 200**

```json
{
  "enabled": true,
  "tiers": [
    { "tier": "random", "label": "Random", "description": "…", "available": true },
    { "tier": "heuristic", "label": "Heuristic", "description": "…", "available": true },
    { "tier": "assisted", "label": "Assisted", "description": "…", "available": false, "reason": "needs a model: set CMDCTRL_OPENAI_ENDPOINT … or CMDCTRL_ANTHROPIC_API_KEY …" },
    { "tier": "strong", "label": "Strong", "description": "…", "available": false, "reason": "needs a model: …" }
  ],
  "decks": [
    { "id": "izzet-aggro", "name": "Raid and Ransack", "description": "aggro — …", "colors": ["U", "R"], "commander": "Mary Read and Anne Bonny" },
    { "id": "simic-ramp", "name": "Deep Roots", "description": "ramp-stompy — …", "colors": ["U", "G"], "commander": "Tatyova, Benthic Druid" },
    { "id": "esper-control", "name": "The Long Answer", "description": "control — …", "colors": ["W", "U", "B"], "commander": "Hashaton, Scarab's Fist" },
    { "id": "mono-black-aristocrats", "name": "Body Count", "description": "aristocrats — …", "colors": ["B"], "commander": "Syr Konrad, the Grim" }
  ]
}
```

Every declared tier is listed, including the ones THIS server cannot
play. `available: false` is what the picker greys out — asking for one
is a 422, never a silent downgrade to a weaker tier, because a bot
labelled "strong" that plays at random is worse than no bot.

Availability is a property of the deployment rather than of the build.
`random` and `heuristic` need nothing; `assisted` and `strong` call a
model, and a server with no model endpoint configured reports them
unavailable with a `reason` naming what to set (`CMDCTRL_OPENAI_ENDPOINT`
for a local LLM, `CMDCTRL_ANTHROPIC_API_KEY` for the hosted one — see
the env list in `server/cmd/server/main.go`). `reason` is present only
on an unavailable tier and is safe to render verbatim.

Reporting them unavailable rather than offering them is deliberate: the
model tiers are complete policies without a model — they fall back to
the heuristic on every window — and a seat playing the heuristic under
the "assisted" label tells the player something false about the game
they are in.

`enabled` is `false` on a server with no bot host configured; the
client hides the Add-bot control rather than offering a button that
503s. `decks` is empty when no deck catalog is wired, in which case
only the raw-decklist form of the add request works.

A deck's `description` leads with its archetype — `aggro`,
`ramp-stompy`, `control`, `aristocrats` — because the names are
flavour and the archetype is what the player is actually choosing
between. The catalog is the four pre-built decks in
`server/internal/decks` — the same four `GET /decks` offers a human,
from one registry — every non-basic card of which is build-tested to
resolve to a registered effect spec. `GET /decks` adds the coverage
profile on top; `aiseat.DeckInfo` has no room for one and a bot does
not read disclosures.

### `POST /games/{id}/seats/bot` (S31)

Seat a bot at an unstarted table. Allowed for admins and for any
player already seated at this table — the request was "add to any
unstarted table", not an admin chore. The seat counts toward the
four-player limit exactly like a human, arrives with its deck
installed, and is driven by a server-side runner from the moment the
game starts (ADR 0033).

The deck travels exactly as it does for `POST /games/{id}/decks` —
`format` + `source`, same auto-detection, same parse / resolve /
validate pipeline, same 422 violation shape — so a bot cannot be
seated with a deck a human couldn't upload.

**Request** — exactly one of `deck` and `source`.

The player-facing form names a curated deck from `GET /bot/options`:

```json
{
  "tier": "random",
  "deck": "izzet-aggro",
  "name": "Bot 1"            // optional; defaults to "Bot N"
}
```

The escape hatch — the test harness, and trying a list that is not in
the catalog — pastes a decklist instead, in exactly the shape
`POST /games/{id}/decks` takes:

```json
{
  "tier": "random",
  "name": "Bot 1",
  "format": "text",          // as for /decks; empty → auto-detect
  "source": "Commander:\n1 Krenko, Mob Boss\n\nMainboard:\n..."
}
```

A named deck resolves to its decklist text server-side and then runs
the identical parse / resolve / validate pipeline, so there is one
code path and a curated deck that rots (a renamed card, a new ban)
fails exactly where a human's upload would.

**Response** `201 Created`

```json
{
  "game": { ...GameMeta, "players": [ ..., { "player_id": "<uuid>", "name": "Bot 1", "seat": 1, "deck_name": "...", "deck_uploaded": true, "is_bot": true, "bot_tier": "random", "bot_deck": "izzet-aggro" } ] },
  "player_id": "<uuid>",
  "deck_name": "...",
  "warnings": [ ... ],       // as for /decks, omitted when empty
  "unimplemented": [ ... ]   // as for /decks, omitted when empty
}
```

`unimplemented` is the same disclosure `POST /games/{id}/decks`
carries: the distinct names of the installed deck's cards whose
printed rules the engine will not carry out. Always empty for a
curated deck — every non-basic card in one is build-tested to resolve
to a registered effect spec — so it only ever appears on the `source`
escape hatch. It is not an error and does not block the seat: a bot is
held to the same catalog the table is, and the honest answer to a gap
is to name it, not to refuse the deck.

**Errors**

| Status | Reason |
|---|---|
| 400 | neither `deck` nor `source`, both of them, bad JSON, unknown `format` |
| 403 | caller is neither admin nor a **seated player** at this game (a spectator's session carries the game ID and is refused here) |
| 404 | game not found |
| 409 | game already started, or the table is full |
| 422 | unknown or unavailable `tier`, unknown `deck` ID, or the deck failed validation (violation list, as for `/decks`) |
| 503 | card index not loaded, no deck catalog configured (for a `deck` request), or bot seats are not enabled on this server |

Seat arithmetic worth knowing: bots take real seats and the table
holds four, so a seated human can add at most three. A four-bot table
is reachable only by an admin creating the game and adding four — it
is the engine's fuzz harness, not a player flow. One human plus one
bot is a legal game (`MinPlayers` is 2) and is the solo-practice case.

**Lifecycle.** Bot seats carry real decks, so `Start`'s
"every seat has uploaded a deck" gate needs no special case. `Start`
launches one runner goroutine per bot seat; a runner exits on its own
the moment the game leaves the active state, so the end of a game
needs no explicit stop. Deleting a game and shutting the server down
both cancel the runners and wait for them. A restart restores them:
`is_bot` / `bot_tier` / `bot_deck` ride the engine snapshot, `bot_tier`
is also on the game's `seats` row, and the lobby's restore path relaunches
a runner for every bot seat in a game that came back active.

### `DELETE /games/{id}/seats/bot/{player_id}` (S31)

Unseat a bot from an unstarted table and close the gap in seat
numbers. Same authorisation as adding one. Returns the updated
`GameMeta`.

| Status | Reason |
|---|---|
| 403 | caller is neither admin nor seated at this game; or `player_id` is not seated here |
| 404 | game not found |
| 409 | game already started |
| 422 | the seat is a human, not a bot |

### `DELETE /games/{id}` *(admin only)*

**Destroys** a table: the lobby entry, the engine snapshot, the
restore point and the replay JSONL, plus any WebSocket clients bound
to it. There is no undo.

Prefer `POST /games/{id}/archive` unless you actually want the
artifacts gone. A bug report filed against this game keeps working —
`bugstore.PinReplay` copies the replay into the report directory
precisely so a later delete cannot orphan the issue — but the live
replay of a game nobody reported is gone for good.

**Response 204** — no body.

### `POST /games/{id}/archive` *(admin only)*

Retire a table. It leaves `GET /games` (appearing under
`?archived=1`), its bot runners are stopped, and connected clients
are evicted — but **nothing on disk is removed**. The engine
snapshot, the replay log, the restore point and both invite tokens
all survive, the room stays registered, and the archived state
persists across a restart (a restored archived game is not
re-listed and its bots are not relaunched).

This is the action for clearing old games off the lobby screen.
`DELETE /games/{id}` is the irreversible one.

**Response 200** — the updated `GameMeta`, carrying `archived_at`.
Invite tokens are stripped.

**Errors**

| Status | Reason |
|---|---|
| 404 | game not found |

### `DELETE /games/{id}/archive` *(admin only)*

The undo: returns an archived table to the active listing, and
relaunches its bot runners if it was mid-game.

**Response 200** — the updated `GameMeta`, with `archived_at` absent.

### `POST /games/{id}/seats/{player_id}/reclaim` *(admin only)*

Mint a **seat-reclaim ticket** for one seat: a link that puts a
disconnected player back in their own seat at a table that has
already started. `Lobby.JoinWithIdentity` refuses any game past the
lobby state, and `/spectate` is the only state-agnostic path, so
without this a player who lost their session can watch their own game
but not play it ([ADR 0044](decisions/0044-surviving-a-deploy.md)
decision 4).

**This is a bearer credential for one player's seat**, hidden
information included. The shape reflects that:

| Property | What it means |
|---|---|
| Admin-only to mint | same `auth.RoleAdmin` gate as `POST /games`. A player seated at the table cannot mint one, not even for their own seat. |
| Minted, never derived | 32 bytes from `crypto/rand`. Not a function of the game ID, seat index, player ID or invite token. |
| Short TTL | 15 minutes (`lobby.ReclaimTTL`), reported in the response as `expires_at` + `ttl_seconds` so the host can see what they handed out. |
| Single use | redemption consumes it under the lobby mutex; a second presentation is indistinguishable from a forgery. |
| Returns a seat, never creates one | the seat must already exist in `GameMeta.Players` at mint time **and** again at redemption. `JoinWithIdentity` is not widened. |
| Not stored, not logged | held in memory keyed by the SHA-256 of the token, never written to `GameMeta` or to disk, never logged. |

Because the store is in memory, outstanding tickets die with the
process — a link does not survive a deploy. That needs the durable
HMAC sessions of #517; a 15-minute credential expiring early is the
right failure for now.

Archiving or deleting a table drops its outstanding tickets.

**Response 201**

```json
{
  "ticket": "<32 random bytes, base64url>",
  "game_id": "<uuid>",
  "player_id": "<uuid>",
  "seat": 0,
  "player_name": "Alice",
  "expires_at": "2026-09-14T12:15:00Z",
  "ttl_seconds": 900,
  "single_use": true
}
```

The client turns `ticket` into `#/games/{id}/reclaim?t=<ticket>`.

**Errors**

| Status | Reason |
|---|---|
| 400 | malformed player id |
| 403 | caller is not an admin |
| 404 | game not found |
| 409 | the table has been archived |
| 422 | that seat is a bot, not a disconnected player |
| 429 | too many outstanding tickets |

### `POST /games/{id}/invites/rotate` *(creator or admin)*

Revoke a game's current invite of one **kind** and mint its
replacement, atomically. This is the fix for the gap `GET
/games/{id}` documents above: only the process that minted a game can
show its invite plaintext, so a link lost after a restart could not
be recovered before this route existed — the table just sat there
with an invite nobody could read or reissue (#1038, [ADR
0051](decisions/0051-user-database.md) decision 4).

Rotating needs nothing from the OLD token. `Store.RotateInvite`
revokes every still-live invite of the requested kind for the game by
`(game_id, kind)`, not by hash, and inserts the new one in the same
transaction — which is exactly what makes this work when the old
plaintext is gone from every process's memory, restart or not.

**Creator or admin** ([#1098](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1098),
closing out the `TODO(#1044)` this used to carry, now that
`games.created_by` exists). The route itself is session-gated for any
authenticated role; `lobby.CanRotateInvites(principal, meta)` decides
inside the handler — true for `RoleAdmin`, or for a principal whose
`UserID` equals `meta.CreatedBy`. A guest (no `UserID`), a different
signed-in user, or the creator of a *different* game is refused. This
"creator" is `games.created_by` — whoever called `POST /games` while
signed in — which is a different concept from [the table
host](#the-table-host): the creator need not ever sit down, and a
seated host need not be the creator. The lobby UI's rotate buttons
follow the server's `is_creator` field on `GameMeta` (below), the
same "computed per-viewer, never the raw identity" pattern
`redactMetaFor` already uses for invite tokens.

**Request**

```json
{ "kind": "player" }
```

`kind` is `"player"` or `"spectator"`. Only the invite of that kind is
touched — rotating the player invite leaves the spectator invite (and
vice versa) working exactly as it did before.

**The old link of that kind stops working the instant this returns.**
Anyone still holding it gets the same 401 an unknown or expired token
gets from `/join` or `/spectate`. Send the new one to whoever needs
it; there is no way to see the old one again.

**Response 200**

```json
{ "kind": "player", "token": "<16-byte base64url>" }
```

`token` is the new plaintext, returned **once** — the same rule
`POST /games` and the seat-reclaim ticket follow. The lobby UI turns
it into a link the same way it does for `POST /games`'s
`invite_token` / `spectator_invite`. Also updates `GameMeta` in this
process's memory, so a subsequent `GET /games/{id}` keeps showing a
usable link — until the next restart, same as any other invite.

Rate-limited in the same bucket as `/join`, `/spectate` and
`/games/{id}/preview`: it mints and revokes the exact credential
those routes brute-force, even though the auth gate already keeps a
stranger from calling it at all.

**Errors**

| Status | Reason |
|---|---|
| 400 | `kind` is neither `"player"` nor `"spectator"` |
| 401 | unauthenticated |
| 403 | caller is neither the game's creator nor the admin |
| 404 | game not found |
| 429 | rate-limited |

### `POST /games/{id}/invites/dm`

Send one person this table's invite link as a Discord direct message
([ADR 0051](decisions/0051-user-database.md) decision 5, S34 sub-PR
6). The server opens the DM itself, with a bot token and two plain
REST calls — `POST /users/@me/channels` then `POST
/channels/{id}/messages`. The gateway bot binary is not involved. The
`/c2-invite-dm` slash command ([#613](https://github.com/krakenhavoc/cmd_and_ctrl/issues/613))
is a thin client of this route, so there is exactly one place that
builds and sends an invite DM.

**Nothing is minted.** The DM carries the game's *current* player
invite, the same link `GET /games/{id}` shows. A link already pasted
into a channel keeps working, and sending a DM does not rotate
anything.

**Who may call it:** someone seated at this table, the person who
created it (`games.created_by`), or the admin. Everyone else gets
**403** — a spectator, a player at a different table, and a signed-in
stranger alike. "Seated" is satisfied by a `player` session bound to
this game, or by any session whose `user_id` holds a seat here (the
same person signed in in another tab).

**Request**

```json
{ "user_id": "<uuid from GET /me/tablemates>" }
```

`user_id` is **our** user id, never a Discord snowflake. The server
resolves the Discord account from `identities` itself, so a snowflake
appears on the wire in neither direction.

`discord_id` is accepted as an alternative — a raw Discord snowflake,
exactly one of the two fields — but **only for an admin session**. It
exists for #613's `/c2-invite-dm @user`, which holds a mention and
nothing else: its target may never have signed in here, so there is no
user id to send. Letting every caller pass one would turn this route
into "DM any Discord user who shares a server with the bot", which is
a spam primitive; restricted to the admin credential the bot already
holds, it is the bot's own path and nothing more. A non-admin who
sends `discord_id` gets **403**.

**Response 200**

```json
{ "sent": true, "user_id": "<uuid>", "display_name": "Bob" }
```

The response carries no snowflake and no invite token: the caller
learns that the DM was sent, not what was in it.

**Rate-limited twice**, and a request has to satisfy both: the client-IP
bucket every invite-adjacent route rides (`/join`, `/spectate`,
`/games/{id}/preview`, `/invites/rotate`), and a **per-caller** bucket
of ~1 DM / 10 s with a burst of 3, keyed on the caller's user (or
their seat, or the admin credential). The per-caller bucket is the one
decision 5 asks for: every accepted call costs two outbound Discord
writes and lands an unsolicited DM in somebody's inbox, and behind a
reverse proxy the IP bucket is shared by the whole table.

**Configuration.** The route needs `CMDCTRL_DISCORD_BOT_TOKEN` on the
**server's** env file (`/etc/cmd_and_ctrl/env`), not only the bot's,
and an origin to build the link against (`CMDCTRL_PUBLIC_BASE_URL`,
falling back to `CMDCTRL_CLIENT_BASE_URL`). Either one missing is a
**503 naming the variable**; it never falls open, every other route is
unaffected, and the server says which state it is in once at boot. CD
writes the token on **production only** — one Discord application, and
a preview box DMing real people from the same identity is the
double-send the bot's prod-only rule already exists to prevent.

**The invite plaintext can be missing.** Only the process that minted
an invite holds its plaintext (see `GET /games/{id}`), so a table
recovered from the database after a restart has a hash and no link.
This route then answers **409** and points at
[`POST /games/{id}/invites/rotate`](#post-gamesidinvitesrotate)
rather than rotating by itself: rotating would silently revoke the
link the table has already shared, which is a startling side effect of
"DM this to Alice". Mint the replacement deliberately, then send it.

**Errors**

| Status | Reason |
|---|---|
| 400 | neither or both of `user_id` / `discord_id`, or `user_id` is not a uuid |
| 401 | unauthenticated |
| 403 | not seated here, not the creator, not the admin — or `discord_id` from a non-admin |
| 404 | no such game, no such user, or Discord has no user with that id |
| 409 | this process no longer holds the table's invite plaintext (rotate first) |
| 422 | the target has no Discord identity here, or Discord refused the DM (no shared server / DMs closed) |
| 429 | rate-limited, by us or by Discord |
| 502 | Discord rejected this server's bot credentials, or failed for another reason |
| 503 | `CMDCTRL_DISCORD_BOT_TOKEN` or the invite origin is not configured |

### `GET /games/{id}/creator` *(admin only)*

The Discord bot's `/c2-end` host check ([#1098](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1098)):
does the Discord user named by `?discord_id=<snowflake>` match this
game's creator. The bot calls the server with its own admin
credentials (same as every other bot call — see
[AGENTS.md](../AGENTS.md)), so the server has to tell it *whether*
the Discord user who ran the slash command created the table; this
route answers exactly that and nothing more.

It deliberately never says who the creator actually is — only
whether one named snowflake matches — so an admin token that guesses
wrong (or is compromised) never learns a game's creator identity from
this route, and nobody who isn't already an admin can call it at all.
`GET /games/{id}` does not carry the creator's raw identity either
(see `is_creator` below); this endpoint exists so the bot's one
actual question can be answered without that identity ever leaving
the server.

**Request**: `?discord_id=<snowflake>`, required.

**Response 200**

```json
{ "is_creator": true }
```

A game with no creator (`games.created_by` is `NULL` — admin-created,
or restored from a pre-ADR-0051 file import) always answers
`is_creator: false`, for every `discord_id` — there is nothing to
match, so the bot's `CMDCTRL_DISCORD_ADMIN_USER_IDS` /
`CMDCTRL_DISCORD_ADMIN_ROLE_IDS` allowlists stay the only route for
those games. An unresolvable `discord_id` (no linked `identities` row
— including every deployment with no user database) answers `false`
the same way rather than erroring, for the same "never fail open, and
never fail closed with something other than a clean refusal" reason
the allowlists already follow.

**Errors**

| Status | Reason |
|---|---|
| 400 | missing `discord_id` |
| 401 | unauthenticated |
| 403 | caller is not an admin |
| 404 | game not found |

### `GET /me`

Echo the principal attached to the request. Used by the client for
bootstrap — "am I still logged in, and as what?"

### `GET /me/games`

"My games" ([ADR 0051](decisions/0051-user-database.md) decision 4,
S34 sub-PR 4): every seat the caller's user holds, joined with its
game, **newest game first**. Ended and archived games are included,
as is a finished game whose table did not survive a restart; the
`games` and `seats` rows are the record.

Needs a signed-in person: an `identified` session, or a `player`
session with a non-nil `user_id`. A caller with no credential at all
gets **401**; an authenticated caller who is not a person gets **403**
— a guest seat's session, an admin, a spectator, and everyone on a
deployment with no database (there are no users).

The split matters to the client, not to the server (#1154): the SPA's
`authFetch` clears the session on **any** 401, so answering 401 to a
valid admin session logged the admin out of any page that happened to
fetch one of these routes. 403 says the true thing — you are
authenticated, you are just not a person.

**Response 200**

```json
{
  "games": [
    {
      "id": "<game uuid>",
      "name": "Friday Night Commander",
      "state": "ended",
      "seat": 1,
      "winner_seat": 1,
      "created_at": 1789820000000,
      "started_at": 1789820300000,
      "ended_at": 1789825000000,
      "archived_at": null,
      "others": [
        { "seat": 0, "name": "Bob" },
        { "seat": 2, "name": "Ghoul", "bot": true }
      ],
      "rejoin": "/me/games/<game uuid>/session"
    }
  ]
}
```

- Every time is **Unix milliseconds** (the unit the tables store), and
  a time that has not happened is `null`. `winner_seat` is `null` until
  the engine reports a winner.
- `seat` is the caller's seat. `others` is every other seat in seat
  order. A seat's `name` is its user's current display name when it
  has one, so a friend who renamed themselves on Discord reads
  correctly on old games, and otherwise the name stored on the seat.
- For a table live in this process, `state`, the times and
  `winner_seat` come from the engine, as `GET /games` reads them.
- `rejoin` is present only while the table is still open: live in this
  process and not archived. It is the path to POST for a seat session.
- No invite token appears anywhere in the response.

### `POST /me/games/{id}/session`

Seat reclaim by user ([ADR 0051](decisions/0051-user-database.md)
decision 3). A signed-in person gets a fresh `player` session for
**their own** seat at a live table, without the invite link and
without an admin-minted ticket. `seats.user_id` is the proof, and only
a Discord sign-in writes it. The ticket route
([`POST /games/{id}/reclaim`](#post-gamesidreclaim)) stays the backstop
for guests, who have no identity to prove.

Same caller rule as `GET /me/games`. No body.

**Response 200**: the `sessionResponse` shape `/join` returns, cookie
included. The principal is bound to the seat's `(game_id, player_id)`
and carries the caller's `user_id` and the seat's Discord identity.
Both invite tokens are stripped from the embedded `game`.

**Errors**

| Status | Reason |
|---|---|
| 401 | no session at all |
| 403 | authenticated but not a signed-in person (see `GET /me/games`) |
| 403 | the caller holds no seat at this table |
| 404 | the table is not live in this process |
| 409 | the table has been archived |

---

### `GET /me/tablemates` (ADR 0051 decision 8, S34 sub-PR 6)

The people the caller has shared a table with, **most recently shared
table first** — the list the invite picker offers. Decision 8's whole
design: a self-join of `seats`, not a `friendships` table. There are
no friend requests and no acceptance; if an explicit list is ever
wanted it is one table on top of this and changes nothing here.

Same caller rule as `GET /me/games` and `GET /me/decks`: a signed-in
person — an `identified` session, or a `player` session with a
non-nil `user_id`. No credential is **401**; every other caller is
**403**, including a guest's seat session, an admin (a credential, not
a person) and everyone on a deployment with no database.

Excluded, always: the caller themselves, and any seat with no user — a
guest, a bot, or a Discord seat still waiting on its `users` row.

**Response 200**

```json
{
  "tablemates": [
    {
      "user_id": "<uuid>",
      "display_name": "Bob",
      "avatar_url": "/avatars/<snowflake>/<hash>.png",
      "last_played_at": 1789820000000
    }
  ]
}
```

- `user_id` is **our** id. A Discord snowflake is never an identifier
  here: `POST /games/{id}/invites/dm` takes this id and resolves the
  Discord account server-side.
- `display_name` is the tablemate's *current* name, so a friend who
  renamed themselves on Discord reads correctly on old games too.
- `avatar_url` is the same-origin path the client already loads every
  seat's avatar from, and is absent for an account with no avatar.
- `last_played_at` is Unix milliseconds (the unit the tables store, as
  `GET /me/games` uses): when the most recent shared table was
  **created**. Deliberately `created_at` rather than `started_at` or
  `ended_at` — those are null on a table that never started, and a
  lobby you both sat in yesterday is a better suggestion than a game
  you finished a year ago.

### `GET /me/decks` (ADR 0051 decision 7, S34 sub-PR 5)

The caller's deck library — see "The deck library" under
`POST /games/{id}/decks` above for how a row gets there and the update
rule. Newest updated first.

**401** for any principal with no `user_id` — a guest's `player`
session, an admin session, or an `identified` session on a deployment
with no database — not only for a missing credential. There is
nothing partial to show: a `user_id`-less principal owns no decks by
construction.

**Response 200**

```json
{
  "decks": [
    {
      "id": "<uuid>",
      "name": "Atraxa Superfriends",
      "commanders": ["Atraxa, Praetors' Voice"],
      "card_count": 100,
      "updated_at": "2026-09-19T08:00:00Z"
    }
  ]
}
```

`source_text` and `source_format` are not included here — this is the
picker's list, not the re-seat payload; seating reads them server-side
via `POST /games/{id}/decks/{deck_id}`.

### `POST /deck-requests`

Ask for a deck's missing cards to be added to the engine
([ADR 0095](decisions/0095-deck-coverage-and-deck-requests.md) §3).
The server files a GitHub issue that is a checklist of the deck's
`manual` and `unreviewed` cards, or joins the open issue already filed
for that deck. The site's "Request these cards" button and the bot's
`/c2-deck-req` both call it.

**Who may call it:**

- **The bot**, with the admin session, naming the Discord member it is
  asking for: `requester: {discord_id, display_name}`. Required for an
  admin session (400 without it, or with a `discord_id` that is not a
  Discord user id).
- **A signed-in user whose account has a Discord identity.** The
  requester comes from the session and the users table; sending
  `requester` is 403.

No session is 401. A guest or spectator seat, or a signed-in account
with no Discord identity, is 403.

**Request**

```json
{ "url": "https://moxfield.com/decks/AbC123" }
{ "url": "https://moxfield.com/decks/AbC123",
  "requester": { "discord_id": "123456789012345678", "display_name": "Alice" } }
```

A link only. `text` is refused with 400: a pasted list has no stable
identity to deduplicate on.

**What it does**, in order:

1. **Rate limit.** Three asks per requester per rolling 24 hours, keyed
   `discord:<snowflake>` whether the ask came from the site or the bot,
   and counted in the database (`deck_request_asks`) so a deploy does
   not reset it. Over the limit is **429** `rate_limited`, decided
   before the deck is fetched. Only an ask that reaches GitHub (an
   issue filed, or a comment added) counts.
2. **The report.** The deck is fetched and bucketed exactly as
   `POST /deck-coverage` does it, from the same 10-minute cache. A
   fetch failure answers with that route's error table.
3. **Nothing to add.** No `manual` and no `unreviewed` card: no issue,
   **200** `nothing_to_add`, with the report (its caveated cards
   included).
4. **An open issue for the deck.** A comment, "Also requested by
   *name*", with the current counts: **200** `joined`. If this
   requester already asked on that issue, no comment is added, and the
   answer is `joined` with `already_requested: true`.
5. **A closed or deleted issue, or none.** A new issue, and the
   deck's row repointed at it: **201** `filed`.

The issue's title is `[deck-request] <deck name>` (the commander's name
when the deck has none), with labels `enhancement` and `deck-request`.
If GitHub refuses the labels, the issue is filed without them, as
`POST /bugreport` does. The body holds the deck link, the requester's
display name (never the snowflake), the counts, `## Cards to add` and
`## Cards to review` checklists (`- [ ] Name (oracle id)`),
`## Automated with caveats`, any names the index could not resolve, and
a footer naming the surface that filed it. The deck name, the display
name and any unresolved name went through ADR 0017's redaction before
they are published, and an `@` in them cannot ping anyone.

**Response**

```json
{
  "status": "filed",
  "issue_url": "https://github.com/krakenhavoc/cmd_and_ctrl/issues/1700",
  "issue_number": 1700,
  "report": { "deck_name": "…", "counts": { "manual": 12, "…": 0 }, "cards": [] }
}
```

| `status` | HTTP | Fields |
|---|---|---|
| `filed` | 201 | `issue_url`, `issue_number`, `report` |
| `joined` | 200 | `issue_url`, `issue_number`, `report`; `already_requested: true` when no comment was added |
| `nothing_to_add` | 200 | `report` |
| `rate_limited` | 429 | `retry_after`: whole seconds until the next ask is allowed (also sent as `Retry-After`) |

`report` is the `POST /deck-coverage` body. A 429 from the IP limiter
(`{"error": "too many requests"}`) has no `status` field; tell the two
apart by it.

**Other errors**

- **503** when the feature is off: `CMDCTRL_GITHUB_TOKEN` is not set
  (the message names it), or the server has no database
  (`CMDCTRL_DATA_DIR`). Never a silent success.
- **502** when GitHub fails while reading the deck's issue, commenting
  or filing.

## Signing out

Session lifetimes: the identity session a Discord sign-in mints from
the login page lasts `CMDCTRL_IDENTITY_TTL` (30 days by default). Every
other session (seat, spectator, admin) lasts `CMDCTRL_SESSION_TTL`
(12 hours by default). See
[ADR 0051](decisions/0051-user-database.md) decision 3.

### `POST /logout`

No credential required. Clears the session cookie and asks the
authenticator to revoke the presented token. Under HMAC sessions that
revoke is advisory ([ADR 0044](decisions/0044-surviving-a-deploy.md)
decision 3): a copy of the token held elsewhere stays valid until it
expires. Always **204**.

### `POST /logout/everywhere`

Requires a session **with a user**: a Discord sign-in, or a seat
claimed from one. Withdraws every session that user holds, in every
browser, including the caller's. It sets `users.sessions_invalid_before`
to now, closes the user's open game WebSockets, and clears the cookie
([ADR 0051](decisions/0051-user-database.md) decision 6). From then on,
any token of theirs issued at or before that instant fails with
`401 {"error":"session revoked"}`, on every route and on the WS upgrade.
A new Discord sign-in works straight away.

The route only acts on the caller's own user. There is no way to name
another one.

**Response 204**, with a `Set-Cookie` that evicts the session cookie.

**Errors**

| Status | Reason |
|---|---|
| 401 | no session, or one that is already expired or revoked |
| 403 | the session has no user: admin, guest or spectator, or any session on a server with no database. `POST /logout` is their sign-out |
| 404 | the user row no longer exists |
| 503 | the server has a user session but no revocation list (not a production configuration) |

### `POST /admin/users/{id}/revoke-sessions` *(admin only)*

ADR 0051's "admin remove-user". It does what the name says and nothing
more: the same revocation as `/logout/everywhere`, for the user `{id}`.
Every row stays: the user, identities, seats, games and decks. The
person can sign in with Discord again. Admin sessions have no user and
are unaffected, including the caller's.

**Response 200**

```json
{
  "user_id": "<uuid>",
  "sessions_invalid_before": "2026-09-19T12:00:00.123Z",
  "sockets_closed": 2
}
```

**Errors**

| Status | Reason |
|---|---|
| 400 | `{id}` is not a uuid, or is the zero uuid |
| 401 | no session |
| 403 | caller is not an admin |
| 404 | no such user |
| 503 | no user database (`CMDCTRL_DATA_DIR` empty) |

---

## Discord sign-in (S12.5, ADR 0004 / 0050 / 0051)

`GET /auth/discord/start` and `GET /auth/discord/callback` are the
OAuth round-trip ([ADR 0004](decisions/0004-discord-identity.md),
[ADR 0050](decisions/0050-discord-login-identity.md)). Since S34 the
callback also:

- records the person in `users` / `identities`
  ([ADR 0051](decisions/0051-user-database.md) decision 2), and puts
  `user_id` in the `#/oauth-complete?…` fragment when it did (absent
  on a deployment with no database);
- links every seat waiting on this Discord account's snowflake
  (`seats.pending_discord_id`, set on seats imported from
  `lobby/*.json` and on seats claimed while there was no database) to
  that user, and clears the pending id. One transaction, idempotent,
  on every sign-in. A failure is logged and the sign-in goes ahead;
  the next sign-in tries again.

### `GET /auth/discord/link`

Link Discord to a seat you already hold (S34 sub-PR 4, carried over
from S12.5 [#59](https://github.com/krakenhavoc/cmd_and_ctrl/issues/59)).
A navigation from the in-game menu: the server answers **302** to
Discord's consent screen, and the callback comes back to the same
seat. Works in any game state. A guest who signs in mid-game becomes
that seat's user, and a seat already linked to one Discord account can
be moved to another.

Needs a `player` session (401 without a session, 403 for any other
role). Optional `?game=<uuid>`: when present it must be the session's
own game (409 otherwise), so a browser whose cookie was replaced by a
join in another tab cannot link the wrong seat.

What the callback then does:

1. **Checks the browser.** The state parks `(game, player)`, and the
   callback requires the request's session **cookie** to be the player
   session for that same seat before it exchanges the code. Anything
   else is a **403**. This is the CSRF binding: without it, a link
   started by one person could be finished by another who is sent the
   consent URL, and their Discord account would end up on your seat.
2. Records the person and links pending seats, as for every sign-in.
3. Puts the identity on the seat: `display_name`, `discord_id` and
   `discord_avatar_hash` on the seat and on the engine's player, and
   `seats.user_id`. The seat's typed `name` is kept. The change goes
   through the room, so every client at the table gets a state
   broadcast carrying the new name and avatar straight away.
4. Mints a new `player` session for the same seat, now carrying
   `user_id` and the `discord_*` fields, sets the cookie, and
   redirects to `/#/oauth-complete?token=…&game=…&player_id=…&user_id=…`.

**Errors** (as JSON, like the other callback errors)

| Status | Reason |
|---|---|
| 401 | no session on `/link` |
| 403 | not a player session on `/link`; on the callback, the cookie is not the seat's own session |
| 404 | the table is not live in this process |
| 409 | `?game=` names another table; the table is archived; or the Discord account already holds a different seat at this table |
| 422 | the seat is a bot |
| 503 | Discord is not configured on this server |

---

## Bug reports

In-app "report a bug" button → GitHub issue, server-proxied so the
GitHub token never reaches the browser. See
[ADR 0017](decisions/0017-bug-report-button.md).

### `GET /bugreport/config`

Unauthenticated probe (mirrors `GET /auth/discord/config`): reports
whether the server can file issues, so the client knows whether to
render the report button.

**Response 200**

```json
{ "enabled": true, "attachments": true, "max_images": 4, "max_bytes": 4194304 }
```

`enabled` is `false` when `CMDCTRL_GITHUB_TOKEN` is unset.

`attachments` is reported separately because the two halves fail
independently: a server with a GitHub token but no `CMDCTRL_DATA_DIR`
or no `CMDCTRL_PUBLIC_BASE_URL` files text reports perfectly well and
cannot host screenshots. The client hides its file picker in that case
rather than offering an upload that would 503. `max_images` and
`max_bytes` are the server's caps, echoed so the modal can refuse an
oversized file before uploading it.

### `POST /bugreport`

File a bug report. Requires any authenticated session (player,
spectator, or admin — spectators hit bugs too). Rate-limited to a
burst of 3, then ~1 report / 30 s per client IP.

**Request**

```json
{
  "title": "cast dialog eats X value",
  "kind": "bug",
  "description": "set X=4, dialog sent X=0",
  "context": {
    "game_id": "<uuid>",
    "turn": 5,
    "phase": "main1",
    "step": "main",
    "seq": 731,
    "connection": "connected"
  }
}
```

`kind`, `description`, `context`, and `log` (and every field inside
them) are optional. `title` is capped at 200 characters,
`description` at 5000. Context strings are clipped server-side and
never trusted; the server adds the reporter identity, server time,
and `User-Agent` itself.

**Report kinds.** `kind` is what the reporter is telling us. It picks
the GitHub label and which artifacts the report collects:

| `kind` | label | client log | pinned replay | pinned game log |
|---|---|---|---|---|
| `bug` (default) | `bug` | yes | yes | yes |
| `idea` | `enhancement` | no | no | no |
| `question` | `question` | yes | no | yes |

Screenshots are attached for every kind. The client names a **kind**,
never a label: the mapping lives in the server so a session with a
report button can't attach (or create) arbitrary labels on the
tracker. An omitted or empty `kind` means `bug` — that is what every
client built before the field existed sends, and refusing those would
break the button mid-game. Any other value is a 400 listing the
accepted kinds. A `log` sent with a kind that doesn't take one is
dropped, not refused.

The title prefix is `[in-app] ` for every kind. The kind lives in the
label (and in a `Kind:` row in the issue body), not in the title, so
the convention every existing in-app issue follows keeps working and
retagging in triage doesn't leave a title that disagrees.

**A label never costs the report.** If GitHub refuses the create
because of the label — the repo doesn't have it, or the token may not
apply it (403/422) — the server re-files the issue **unlabelled** and
logs the failure at error level. Delivery failures (a timeout, a 5xx)
are *not* retried: GitHub may have created the issue already, and a
duplicate is worse than the 502 the reporter can act on.

`log` is the reporter's client-side activity ring buffer — up to 200
entries, kept from the tail:

```json
{
  "log": [
    { "at": 1789223415250, "kind": "sent", "text": "action cast_spell id=8f2a1b3c" },
    { "at": 1789223415310, "kind": "received", "text": "snapshot seq=731 turn=5" },
    { "at": 1789223415340, "kind": "console", "text": "TypeError: x is undefined" }
  ]
}
```

`at` is epoch milliseconds **from the reporter's clock** — the server
formats it and labels it as such rather than passing off a client
timestamp as server time. `kind` is one of `sent`, `received`,
`error`, `info`, `console`; anything else renders as `info`. Text is
clipped to 240 bytes and stripped of backticks and control characters
(it lands inside a fenced code block, and log lines can contain chat
other players typed).

The log holds only what the reporter's own browser already saw. Game
state — hands, libraries, decklists, any `GameView` content — is still
never embedded in an issue; see the replay handling below and
[ADR 0017 §7](decisions/0017-bug-report-button.md).

**Request — `multipart/form-data`**

To attach screenshots, send the same JSON as a `report` field plus up
to 4 `image` file parts:

```
POST /bugreport
Content-Type: multipart/form-data; boundary=…

--…
Content-Disposition: form-data; name="report"

{"title":"board rendered wrong","description":"see shots","context":{…}}
--…
Content-Disposition: form-data; name="image"; filename="shot.png"
Content-Type: image/png

<PNG bytes>
--…--
```

Images must be PNG, JPEG, GIF, or WebP — decided by **sniffing the
bytes**, not by the declared `Content-Type`. SVG is refused (it is a
script carrier, and the route that serves these is unauthenticated).
Limits: 4 images, 4 MiB each, 10 MiB per report, 12 MiB for the whole
request.

**Replay pinning.** When the report carries a `context.game_id` whose
game the caller is entitled to (admins anywhere; everyone else only
their own game), the server snapshots that game's replay JSONL
alongside the report and puts the report ID in the issue. The snapshot
is why this exists: the live `/games/{id}/replay` file keeps growing
after the report is filed and disappears when the game is evicted.

**Response 201**

```json
{
  "url": "https://github.com/krakenhavoc/cmd_and_ctrl/issues/123",
  "number": 123,
  "label": "bug",
  "report_id": "6f1c0b7e-9a2d-4c31-8f55-1b2d3e4f5a60"
}
```

`report_id` is present only when the report stored artifacts (images,
a pinned replay, or both). It is the key for both routes below.

`label` is the label the issue actually landed with, and is absent
when none did — so a client can say "filed as enhancement" only when
that is true.

**Errors**

| Status | Reason |
|---|---|
| 400 | missing/oversized title or description, unknown `kind`, unknown field, non-image attachment, missing `report` part |
| 401 | no valid session |
| 413 | an image over 4 MiB, attachments over 10 MiB, or a request over 12 MiB |
| 429 | rate-limited |
| 502 | GitHub rejected or timed out; retry later |
| 503 | bug reporting not configured (`CMDCTRL_GITHUB_TOKEN` unset), or attachments sent to a server with no artifact storage |

A rejected attachment fails the whole report — no issue is filed — and
a report whose GitHub call fails has its stored artifacts deleted, so
an outage never leaves orphaned images at live URLs.

### `GET /bugreport/att/{report_id}/{name}`

Serves one stored screenshot. **Unauthenticated**, by necessity:
GitHub renders an issue image by fetching it through its Camo proxy,
which presents no session and follows no login.

The `report_id` is the capability — a v4 uuid that appears only inside
a private-repo issue body. There is no listing endpoint, and an
unknown id is indistinguishable from a pruned one. `name` must match
`att-N.{png,jpg,gif,webp}`; the pinned replay and the manifest sitting
in the same directory are not reachable here.

Responses carry `X-Content-Type-Options: nosniff`, an allowlisted
`Content-Type` re-sniffed from the bytes on the way out,
`Content-Disposition: inline`, and
`Content-Security-Policy: default-src 'none'; sandbox`.

Artifacts are pruned after 90 days.

| Status | Reason |
|---|---|
| 404 | unknown report, malformed id or name, pruned, or not an allowlisted image |
| 429 | rate-limited (loose: 5/s, burst 60 — Camo plus readers) |
| 503 | artifact storage not configured |

### `GET /bugreport/{report_id}/replay`

Streams the replay JSONL pinned when the report was filed, as
`application/x-ndjson`.

**Admin only** — and unlike `GET /games/{id}/replay` there is no
"once the game has ended" relaxation for players. The pinned copy is
the UNFILTERED view (opponents' hands, full library order) and has no
live game whose state could justify loosening the rule. This is what
keeps a filed issue from being a hidden-information side channel for a
reporter who is also a repo collaborator.

| Status | Reason |
|---|---|
| 401 | no valid session |
| 403 | not an admin |
| 404 | unknown report, or the report pinned no replay |
| 503 | artifact storage not configured |

---

### `GET /bugreport/{report_id}/gamelog`

Streams the public game log pinned when the report was filed, as
`text/plain`. One line per entry: sequence number, turn, kind, and the
rendered sentence.

**Admin only**, the same gate as the pinned replay. The contents are
public information by construction — the log is built from what a
player at the table could observe — but the pinned copy is the
UNFILTERED projection, and this route has no seat to filter it for.

Where the replay says what the state *was* at every frame, the log says
what *happened*, which is usually the question a bug report is asking.
Added in S31 ([ADR 0033](decisions/0033-ai-bot-seat.md) §4).

| Status | Reason |
|---|---|
| 401 | no valid session |
| 403 | not an admin |
| 404 | unknown report, or the report pinned no game log |
| 503 | artifact storage not configured |

---

## Card routes (authenticated)

Served by the `cards` package. Both routes require authentication.

### `GET /cards/{id}`

Card metadata from the Scryfall bulk index. 404 if the card is
unknown or the index has not been loaded yet.

### `GET /cards/{id}/image?size=<name>`

Serves the card image bytes. On cache miss, downloads from Scryfall's
CDN and writes to `$CMDCTRL_DATA_DIR/images/<aa>/<id>.<size>.jpg`.

`size` defaults to `normal`; valid values: `small`, `normal`, `large`,
`png`, `art_crop`, `border_crop`. Response carries `Cache-Control:
public, max-age=604800, immutable` — Scryfall card UUIDs are immutable.

A failed CDN download is not retried server-side: it answers `502`
with a JSON error body and **no** cache headers, and writes nothing to
the disk cache, so the same URL re-attempts the CDN on the next
request. The client relies on that — every card-art `<img>` retries
the plain URL once after ~2 s, then shows a click-to-retry marker
(`client/src/lib/cardArt.ts`, #33). A cache-busting query parameter
would be wrong here: the service worker keys card art on the full
query string.
