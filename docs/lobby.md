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

List games known to the lobby. Invite tokens are STRIPPED — listing is
not proof of ownership.

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
{ "enabled": true }
```

`enabled` is `false` when `CMDCTRL_GITHUB_TOKEN` is unset.

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

`description` and `context` (and every field inside `context`) are
optional. `title` is capped at 200 characters, `description` at
5000. Context strings are clipped server-side and never trusted; the
server adds the reporter identity, server time, and `User-Agent`
itself. Game state (hands, libraries, decklists) is deliberately
never embedded — the issue references `GET /games/{id}/replay` for
the operator instead.

**Response 201**

```json
{ "url": "https://github.com/krakenhavoc/cmd_and_ctrl/issues/123", "number": 123 }
```

**Errors**

| Status | Reason |
|---|---|
| 400 | missing/oversized title or description, unknown field |
| 401 | no valid session |
| 429 | rate-limited |
| 502 | GitHub rejected or timed out; retry later |
| 503 | bug reporting not configured (`CMDCTRL_GITHUB_TOKEN` unset) |

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
