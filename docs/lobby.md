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

**Errors**

| Status | Reason |
|---|---|
| 401 | invite token did not match |
| 404 | game not found |
| 409 | game already started, or game full, or empty name |

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
{ "name": "Friday Night Magic" }
```

**Response 201**

```json
{
  "id": "<game uuid>",
  "name": "Friday Night Magic",
  "created_at": "2026-04-13T20:00:00Z",
  "invite_token": "<16-byte base64url>",
  "players": [],
  "state": "lobby"
}
```

**Errors**

| Status | Reason |
|---|---|
| 401 | unauthenticated |
| 403 | session is not RoleAdmin |
| 400 | empty name |

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
      "players": [{ "player_id": "<uuid>", "name": "Alice", "seat": 0 }],
      "state": "lobby"
    }
  ]
}
```

### `GET /games/{id}`

Full metadata for a single game. The `invite_token` field is
included only for admins and for players seated in this specific
game.

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

Upload a deck for a seat. RolePlayer may only set their own seat's
deck (`player_id` in the body must match the session's bound player);
RoleAdmin may set any seat.

**Request**

```json
{
  "format": "text",   // or "moxfield"; empty → auto-detect
  "source": "Commander:\n1 Atraxa, Praetors' Voice\n\nMainboard:\n...\n",
  "player_id": "<uuid>"
}
```

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
  "commanders": ["Atraxa, Praetors' Voice"],
  "warnings": [
    { "code": "sideboard_not_supported_in_commander", "message": "..." }
  ]
}
```

**Errors**

| Status | Reason |
|---|---|
| 400 | malformed request (missing source / player_id, unknown format) |
| 403 | not your seat (RolePlayer with mismatched player_id) |
| 413 | body exceeds the 2 MiB deck-source cap |
| 422 | validation failed — body carries `{"error", "violations": [...]}`; may also carry `warnings` |
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
between. The catalog is the four curated decks in
`server/internal/aiseat/decks`, every non-basic card of which is
build-tested to resolve to a registered effect spec.

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
`is_bot` / `bot_tier` / `bot_deck` ride both the engine snapshot and
the persisted lobby metadata, and the lobby's restore path relaunches
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

### `GET /me`

Echo the principal attached to the request. Used by the client for
bootstrap — "am I still logged in, and as what?"

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

`description`, `context`, and `log` (and every field inside them) are
optional. `title` is capped at 200 characters, `description` at 5000.
Context strings are clipped server-side and never trusted; the server
adds the reporter identity, server time, and `User-Agent` itself.

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
  "report_id": "6f1c0b7e-9a2d-4c31-8f55-1b2d3e4f5a60"
}
```

`report_id` is present only when the report stored artifacts (images,
a pinned replay, or both). It is the key for both routes below.

**Errors**

| Status | Reason |
|---|---|
| 400 | missing/oversized title or description, unknown field, non-image attachment, missing `report` part |
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
