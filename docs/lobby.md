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

## Authenticated routes

### `POST /games` *(admin only)*

Create a new game.

**Request**

```json
{ "name": "Friday Night Magic", "host_discord_id": "123456789012345678" }
```

`host_discord_id` is optional. It names the table host by Discord user ID
([ADR 0075 §2.1](decisions/0075-table-settings-and-host-controls.md)). The
Discord bot's `/cc-invite` sends the user who ran it. The ID is held on the
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

### `POST /games/{id}/host`

Transfer hosting to another seat. Host or admin only.

**Request**

```json
{ "player_id": "<uuid of the new host>" }
```

**Response 200**: the updated `GameMeta`, with `host_player_id` and the
`is_host` flags moved. Connected clients get a fresh snapshot with the new
`is_host`.

An explicit transfer also clears a pending named host, so a late `/cc-invite`
claimant does not take the table back.

**Errors**

| Status | Reason |
|---|---|
| 400 | missing or malformed `player_id` |
| 401 | unauthenticated |
| 403 | caller is neither the host of this table nor the admin |
| 404 | game not found |
| 422 | `player_id` is not a seat at this table, is a bot, or has left the game |

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
      "host_player_id": "<uuid>"
    }
  ]
}
```

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
seated in the game may call this.

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

### `POST /games/{id}/invites/rotate` *(admin only)*

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

**Admin-only for now.** The issue that will let a game's own creator
rotate their invites (not just an admin) is
[#1044](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1044),
which populates `games.created_by`; there is a `TODO(#1044)` on the
route registration in `server/internal/lobby/http.go`.

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
those routes brute-force, even though the admin gate already keeps a
stranger from calling it at all.

**Errors**

| Status | Reason |
|---|---|
| 400 | `kind` is neither `"player"` nor `"spectator"` |
| 401 | unauthenticated |
| 403 | caller is not an admin |
| 404 | game not found |
| 429 | rate-limited |

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
session with a non-nil `user_id`. Every other caller gets **401**:
no session, a guest seat's session, an admin, a spectator, and
everyone on a deployment with no database (there are no users).

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
| 401 | not a signed-in person (see `GET /me/games`) |
| 403 | the caller holds no seat at this table |
| 404 | the table is not live in this process |
| 409 | the table has been archived |

---

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
