# Lobby HTTP API (S04)

REST surface for game creation, seat claiming, and admin bootstrap.
All responses are JSON. All errors use a uniform shape:

```json
{ "error": "<human-readable message>" }
```

Authentication transport and the full credential-precedence rules
live in [ADR 0003 — Auth and lobby architecture](decisions/0003-auth-and-lobby.md).

---

## Who is an admin (ADR 0110 §3, ADR 0112 §2)

Two kinds of session are admins, and every "admin only" or "admin" below
means either:

- the **shared token**'s session, from `POST /admin/login`
  (`role: "admin"`); and
- a **signed-in person on the allowlist, in admin mode**: any session
  with a `user_id` (identified, player or spectator) whose Discord ID is
  on `CMDCTRL_DISCORD_ADMIN_USER_IDS`, while that person has admin mode
  on. It is the same list the Discord bot uses for `/c2-end`.

Admin is decided on every request (`Config.isAdmin`) and is never in the
token, so taking an ID off the list takes effect on the next request.
The list is read at boot, so in practice that means at the next deploy.
A session with no `user_id` is never an allowlisted admin. That includes
a reclaim ticket's session for an admin's seat, which carries the seat's
Discord ID but no user. `GET /me` reports the answer as `admin`.

**Player mode** ([ADR 0112](decisions/0112-signed-in-home-player-mode-and-one-decks-page.md)
§2). Being on the list makes a person *able* to be an admin. Each
allowlisted person is in **player mode** until they switch admin mode on
with [`PUT /me/admin-mode`](#put-meadmin-mode-adr-0112-2), and it starts
off: every person on the list was in player mode the first time this
deployed. In player mode every route and gate below answers exactly as it
does for a signed-in person who is not on the list. The only extra things
they have are the switch and `admin_allowed` on `GET /me`.

- **Admin mode lasts 12 hours**, like `sudo`. A request after that sees
  player mode, and a sweep once a minute clears the mode and closes the
  person's admin WebSockets. Switching on again restarts the 12 hours.
- It belongs to the **person**, not the session: a switch reaches every
  tab and device at its next request, and nothing in any token changes.
- A switch, or a lapse, closes each of the person's WebSockets whose
  admin bit is now wrong with close code **4001**, reason `admin mode
  changed`. The client asks `GET /me` and reconnects, and the upgrade
  binds it again with the new answer. In player mode that is their own
  seat or spectator session and nothing else. When the connection it
  held needs admin mode (the seatless view, another seat, a table that
  isn't theirs), the client does not redial; it goes to the Lobby with
  "Player mode is on; this table isn't yours."
- **In the client** the switch is the Admin chip in the header's account
  menu: "Admin · until 23:40" on the accent colour while admin mode is
  on, "Player" in a muted outline while it is off. At a table, which has
  no header, the same switch is in the ⋯ menu. The shared token shows a
  static "Admin token" badge instead.
- [`POST /logout/everywhere`](#post-logouteverywhere) and
  [`POST /admin/users/{id}/revoke-sessions`](#post-adminusersidrevoke-sessions-admin-only)
  also end admin mode.
- A failure fails closed. If the server cannot load the modes at boot, it
  logs an ERROR and starts with everyone in player mode.
- With no database there are no users, so no allowlisted admins and no
  mode.

Player mode is not a defence against a stolen session: whoever holds the
session can switch admin mode back on. Revocation is that defence.

An allowlisted person joins tables the ordinary way, as their own Discord
identity. In admin mode their seat session is an admin at every table: it
can use the admin routes, pass every host-or-admin gate, and open any
table's WebSocket like the token (see `GET /ws` in
[protocol.md](protocol.md)).

Three paths exist for the Discord bot calling on other people's behalf,
and they stay on the **shared token alone** (`isServerCredential`). An
allowlisted person is a person on these paths:

- the bot's own rate buckets on `POST /deck-coverage` and the per-caller
  limits;
- `requester` on `POST /deck-requests`, which files a request in a named
  member's name. The token's own admin status does not lift that
  member's rate limits: only the member's own account, allowlisted and
  in admin mode, does ([#2052](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2052));
- `discord_id` on `POST /games/{id}/invites/dm`, which names a raw
  snowflake.

Every request an admin-only route lets through is logged at Info as
`admin action`, with the method and path and either `admin_user_id` (an
allowlisted person) or `admin_id` (the token). No token is ever logged.

## Public routes (no credential required)

### `POST /admin/login`

Exchange the shared admin token for an admin session.

**In the client, the token's form is on `#/admin` only**
([ADR 0112](decisions/0112-signed-in-home-player-mode-and-one-decks-page.md)
§2 item 8, S57). `#/login` is for signed-out visitors and has no token
form or link to one; operators open `#/admin` directly. What `#/admin`
shows:

| Browser holds | `#/admin` |
|---|---|
| no session | the token form |
| the token's own session | goes to `#/lobby` |
| an allowlisted person, in either mode (once `GET /me` has said `admin_allowed`) | goes to `#/lobby`, where the Admin chip is in the header's account menu |
| any other session (a signed-in person not on the list, a guest seat, a spectator) | the token form, with a note that it replaces this browser's session. A signed-in person's session is set aside and comes back when the token's session ends ([ADR 0110](decisions/0110-remember-me.md) §1 item 6). |

Every session that opens `#/login` goes to `#/lobby`, the signed-in
home, where the "Join a table" card and the header's account menu are
(ADR 0112 §1).

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
signed-in session is an `identified` one, or a `player` or `spectator`
one with a non-nil `user_id` (the same person at their next table).
Any other session, and a credential that no longer validates, joins by
`name` exactly as before.

**Lifetime** ([ADR 0110](decisions/0110-remember-me.md) §1, S55): a
signed-in joiner's seat session expires at the same instant as the
session they joined from, so `expires_at` is that session's, not
12 hours from now. A guest's seat session lasts `CMDCTRL_SESSION_TTL`
(12 hours by default). The same rule holds for every route that mints
a session; see [Session lifetimes](#session-lifetimes-adr-0110-1).

**An agent seat** ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md)
§7, S62). An AI agent's MCP client declares itself with an extra body
field:

```json
{
  "invite_token": "<token>",
  "name": "Claude",
  "agent": { "client": "claude-code" }
}
```

`client` is the MCP client's `clientInfo.name`. The server lower-cases
it, replaces every character outside `[a-z0-9._-]` with `-`, trims
leading and trailing `-` and cuts it to 32 characters. An empty result,
or `"agent": {}`, is `"unknown"`. Any `agent` object makes the seat an
agent seat:

- It is a **guest** seat. A request that carries a signed-in person's
  session (as above: `identified`, or `player` / `spectator` with a
  `user_id`) is refused with **400** `an agent seat joins as a guest`,
  rather than seating the person. So an agent seat never has a
  `user_id` or a Discord identity.
- The seat carries `is_agent: true` and `agent_client` in every seat
  list (`GET /games`, `GET /games/{id}`, the invite preview, the join
  response) and in the game view's `PlayerView`
  ([protocol.md](protocol.md)), for every viewer.
- **The badge never comes off.** It is set in the same commit that
  adds the seat, and no route, action, setting, reclaim, restart or
  admin path clears it. It rides the engine snapshot, so it needs no
  database column.
- It **never hosts** ([The table host](#the-table-host)) and **cannot
  link Discord** ([`GET /auth/discord/link`](#get-authdiscordlink)).
- It is a declaration by a cooperating client. A client that does not
  send `agent` is not marked, and the server cannot tell it from a
  person.

**Errors**

| Status | Reason |
|---|---|
| 400 | empty name; or `agent` with a signed-in session (`an agent seat joins as a guest`) |
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
| `player` or `spectator` with a non-nil `user_id` (a signed-in person already at a table) | The same: they take a seat at the code's table as themselves ([ADR 0110](decisions/0110-remember-me.md) §1 item 3, S55) |
| none, or a credential that no longer validates | Classic manual join — `name` is required |
| `admin`, or a guest's `player` / `spectator` (no `user_id`) | 409 — a guest has no identity to carry to a second table, and the admin token is not a person |

The body takes the same optional `agent` field as
[`POST /games/{id}/join`](#post-gamesidjoin), with the same rules: the
seat is a badged guest seat, and an `agent` join with a signed-in
session (the first two rows) is a **400** `an agent seat joins as a
guest` instead of a seat for that person. With no session it joins as a
guest agent. A guest's or the admin's session is the 409 above, as for
any join here.

**Response 200** — identical to `POST /games/{id}/join`, cookie
included. On the Discord path the principal also carries `discord_id`,
`discord_username`, `discord_global_name` and `discord_avatar_hash`,
and `user_id` is the signed-in person's users row, copied from the
session they joined with (S34, [ADR 0051](decisions/0051-user-database.md)
decision 3). `user_id` is the nil uuid for a guest seat and for
everyone on a deployment with no database. A signed-in joiner's seat
session expires with the session they joined from; a guest's lasts
`CMDCTRL_SESSION_TTL`.

The identity session stays valid afterwards: it is how the same person
joins a second table later without signing in to Discord again. On its
own it can do nothing else — the WS authorizer refuses it outright.

**Errors**

| Status | Reason |
|---|---|
| 400 | empty name on the anonymous path; or `agent` with a signed-in session |
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

The session it mints has **no `user_id`** and lasts
`CMDCTRL_SESSION_TTL`: the ticket is the whole credential, and copying
a seat's Discord fields never makes the holder that person. One
exception ([ADR 0110](decisions/0110-remember-me.md) §1 item 4, S55):
when the request **also** carries a valid signed-in session (cookie or
bearer) whose `user_id` owns this seat, the redemption is the same
thing as [`POST /me/games/{id}/session`](#post-megamesidsession), and
the session carries that `user_id` and expires with the session it
came from. A signed-in session for anyone else changes nothing.

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
decks page (`#/decks`, [ADR 0112](decisions/0112-signed-in-home-player-mode-and-one-decks-page.md)
§3, which `#/deck-check` and `#/deck-check?url=` still open, because
the bot links there) and the bot's `/c2-deck-check` both render this
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
guild member at once. Any other session is ignored. A report is cached
for 10 minutes under its deck key, so a second check of the same deck,
however the link is spelled, fetches nothing. A pasted list is cached
under its list key the same way, so the same deck pasted from two
exports is built once.

A pasted list may name at most 2,000 cards in total (400 past that).

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
  `copies` (#2220, additive) is the same five keys summed over each
  card's `count`, and `unknown_copies` the copies of the `unknown`
  names, sideboard excluded. `copies` plus `unknown_copies` totals the
  deck's size (100 for a Commander deck); the site and the bot's
  headline "N of M cards play as printed" read it.
  `count` on a card is its copies, summed across the command zone and
  the main deck. Sideboard rows are not bucketed.
- `cards` is sorted by bucket in the order above, then by name.
- `unknown` is the names the card index could not resolve; they are in
  no bucket.
- `violations` is `deck.Validate`'s answer. It is informational and
  never blocks a report: a 60-card list still gets its buckets.
- `source` is `moxfield`, `archidekt` or `text`. `source_url` is absent
  for `text`.
- `deck_key` is the deck's identity: `moxfield:<id>`, `archidekt:<id>`,
  or, for a pasted list, `list:<16 hex digits>` ([ADR 0095, amendment
  2026-09-25](decisions/0095-deck-coverage-and-deck-requests.md#amendment-2026-09-25-requests-from-a-pasted-list)).
  The list key is the first 16 hex digits of the SHA-256 of the list's
  canonical form: one line per copy, each a card's oracle ID (for a name
  the index cannot resolve, `name:` and the lowercased name), a
  commander's line prefixed `commander:`, sideboard rows left out,
  sorted and joined with `\n`. Set codes, collector numbers, row order,
  split rows and the way the commander is marked do not change it; a
  different commander does.
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
| 424 | `upstream_blocked` | The deck site's CDN is blocking this server |
| 424 | `external_api_unavailable` | The deck site did not answer |

The upstream failures are 424 (Failed Dependency): the caller's link was
fine, and a dependency failed. Not 502: Cloudflare, in front of both
hosts, replaces an origin 502's body with its own "error code: 502"
page, so the player-readable `error` and the `hint` never arrived.

**Moxfield.** Moxfield blocks this server: every Moxfield endpoint
answers it with a Cloudflare 403. So **any** fetch error for a Moxfield
link keeps its status and `code` but carries this sentence as `error`,
and `"hint": "paste_list"`:

```json
{
  "error": "Moxfield blocks our server. On Moxfield, open the deck → Export → Copy plain text, then paste the list here instead.",
  "code": "upstream_blocked",
  "hint": "paste_list",
  "violations": [{ "code": "upstream_blocked", "card": "<the url>", "message": "…" }]
}
```

A client keys its "paste the list instead" affordance on `hint`, not on
`code`. `POST /deck-requests` answers a Moxfield link the same way. Other errors use the uniform `{error}` shape: 400 for a body with
neither or both of `url` and `text`, or an unreadable pasted list; 503
when the card index is not loaded.

## Authenticated routes

### `POST /games` *(signed in, or the admin token)*

Create a new game. Since [ADR 0110](decisions/0110-remember-me.md) §5 item 4
(owner answer 2, Delivery PR 7), any signed-in person may create a table, not
only an admin:

- **A signed-in person** (an identified, player or spectator session that
  carries a user; allowlisted admins included) becomes the table's
  **creator** (`games.created_by`). The creator may rotate its invites, DM
  tablemates its link, and end it with `/c2-end`, and `GET /games` and
  `GET /games/{id}` mark it `is_creator: true` for them. The table also names
  its creator as host, so they host it once they sit down, whoever sat first.
  A person may have at most **3 open tables** (created, unstarted, unarchived)
  and may create **one table per 30 seconds**, with a burst of 3.
- **The shared admin token** is the Discord bot's `/c2-invite` and an
  operator. It is held to neither limit and records no creator, because a
  server credential is not a person.
- **Anyone else** (a guest seat or spectator, with no user) is refused with
  403.

**Request**

```json
{ "name": "Friday Night Magic", "host_discord_id": "123456789012345678", "setup": "last" }
```

`host_discord_id` is optional, and only the admin token may send it; a signed-in
person's table always names its creator. It names the table host by Discord
user ID ([ADR 0075 §2.1](decisions/0075-table-settings-and-host-controls.md)).
The Discord bot's `/c2-invite` sends the user who ran it. The ID is held on the
table, unserved, until that Discord identity claims a seat through the OAuth
join. That seat then becomes host. See [The table host](#the-table-host).

`setup` is optional. `"last"` applies the caller's last table setup to the new
table, exactly as [`POST /games/{id}/setup`](#post-gamesidsetup-adr-0110-5)
would, and the response carries what that did under `setup`. A caller with no
saved setup still gets the table, with a `setup` that says so. Any other value
is a 400 and creates nothing.

**Response 201**

```json
{
  "id": "<game uuid>",
  "name": "Friday Night Magic",
  "created_at": "2026-04-13T20:00:00Z",
  "invite_token": "<16-byte base64url>",
  "spectator_invite": "<16-byte base64url>",
  "players": [],
  "state": "lobby",
  "is_creator": true,
  "setup": { "settings": true, "bots_added": 2, "skipped": [] }
}
```

`setup` is present only when the request asked for one. The meta's fields stay
at the top level, so a client that reads this response as a game is unchanged.

This response is the one reliable place to get the invite links. The
server stores only their hashes, so after a restart it cannot show them
again (see `GET /games/{id}`).

**Errors**

| Status | Reason |
|---|---|
| 401 | unauthenticated |
| 403 | the caller is not a signed-in person and not the admin token |
| 400 | empty name, or an unknown `setup` |
| 409 | the caller already has 3 open tables; the message names them. Start one, or end one with `/c2-end` |
| 429 | the caller created tables too quickly (1 per 30 s, burst 3); `Retry-After: 30` |

### The table host

Every table has at most one **host**: the seat that may manage the table
(settings, and later spawning) alongside the server admin
([ADR 0075 §2.1](decisions/0075-table-settings-and-host-controls.md)).

- **A named host** (`host_discord_id` on `POST /games`) hosts as soon as that
  Discord identity claims a seat.
- **Otherwise the first human seat to join hosts.** When the named host has
  not arrived yet, the first human hosts in the meantime and hands over when
  they sit down.
- **A bot seat never hosts, and neither does an agent seat**
  ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md)
  §7). A table with only bots and agents has no host. An agent that
  joins before any person does not take the table; the first person
  to sit down does.
- **Transfer**: `POST /games/{id}/host` (below).
- **What hosting lets you do**: change the table's settings, through
  `PATCH /games/{id}/settings` (below) or the `set_table_settings` WebSocket
  action. The deprecated `set_undo_limit` action is gated the same way.
- **The host leaving passes it on.** When the host concedes or loses
  ([ADR 0060](decisions/0060-leaving-the-game.md)), hosting passes to the
  next human seat in turn order, skipping bots, agents and departed seats and
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
| 422 | `player_id` is not a seat at this table, is a bot or an agent seat, or has left the game |

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
`spectator_invite` fields are included only for admins, for players
seated in this specific game, and for the game's creator wherever they are
(ADR 0110 §5 item 4).

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
seated in the game may call this. The game opens with the **opening roll**
([ADR 0121](decisions/0121-animated-dice.md) §1, since S61): `GameView.opening_roll`
is present, nothing is shuffled or dealt, and each seat rolls its own d20
(`roll_opening`; a bot rolls on its own clock, and the host may `host_roll_remaining`).
Tied leaders roll again until one remains, and that seat chooses who takes the
first turn (`choose_starting_player`, CR 103.1). Only then are libraries
shuffled and hands dealt, and the mulligan proceeds. The rolls and the choice
are public in the view and the game log, and use the game's persisted RNG so
reconnects and replay agree. See [protocol.md](protocol.md) for the three
actions. The practice table rolls too, and its bot hands the first turn to the
player when it wins ([`POST /games/practice`](#post-gamespractice-adr-0076-s54)).

**The start captures a setup** (ADR 0110 §5 item 1). On the transition, the
server writes one person's last setup: the game's creator's, or, for a table
with no creator, the person who pressed start, if they are signed in. The
setup is the table settings as a complete `PATCH /games/{id}/settings` body,
every bot seat as `{tier, deck_id, name}` in seat order, and the user IDs of
the other signed-in people who sat there. A layout is final when the table
starts, so a table that never starts leaves no setup. A failed write is logged
and never fails the start. See [`GET /me/setup`](#get-mesetup-adr-0110-5).

**Errors**

| Status | Reason |
|---|---|
| 403 | not a seat in this game |
| 404 | game not found |
| 409 | not enough players (min 2), or one or more seats haven't uploaded a deck |

### `POST /games/{id}/setup` (ADR 0110 §5)

Apply the caller's last setup to an unstarted table: its settings through the
same path as `PATCH /games/{id}/settings`, then each bot through the same deck
pipeline as `POST /games/{id}/seats/bot`, up to the free seats. Rides the deck
upload rate bucket, since it seats decks.

**Who:** a signed-in person (the setup is theirs) who is the table's host, its
creator, or an admin.

**Request**

```json
{ "from": "last" }
```

**Response 200**

```json
{
  "game": { "id": "<game uuid>", "players": [ ... ], "state": "lobby" },
  "settings": true,
  "bots_added": 1,
  "skipped": [
    { "name": "Smart", "reason": "the \"strong\" tier is not available on this server" },
    { "name": "Old", "reason": "its deck \"retired-deck\" no longer exists" }
  ]
}
```

What cannot be applied is **skipped and named, never silently downgraded**:
a tier this server does not offer, a curated deck that no longer exists, a bot
that played a pasted list (a setup keeps only curated deck IDs), a deck the
validator now refuses, or no free seat left.

**Errors**

| Status | Reason |
|---|---|
| 401 | unauthenticated |
| 403 | not a signed-in person (a guest, the admin token), or not the table's host, creator or an admin |
| 400 | `from` is not `"last"` |
| 404 | game not found, or the caller has no saved setup |
| 409 | the table has started |

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
- `singleton_violation` — a non-basic-land card appears more than once,
  unless its own text allows it (CR 113.6n: "A deck can have any number
  of cards named ~" or "up to N cards named ~", read from the oracle
  text, so Nazgûl runs nine and Relentless Rats any number). Past a
  printed limit the message names it, e.g. `"Nazgûl" appears 10 times;
  a deck can have up to 9 cards named Nazgûl`.
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

  A failed URL import stays a 422 with this violations shape (it is not a
  502, which Cloudflare would replace with its own page, #1644). The body
  also carries `code`, and `error` is the same player sentence
  `POST /deck-coverage` uses. Moxfield blocks this server, so any failed
  Moxfield link says to export the list and paste it, and adds
  `"hint": "paste_list"`.
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

- Creates or updates a `decks` row for **`format: "text"`,
  `"moxfield"` or `"url"` requests** — not a `deck` (pre-built catalog)
  pick, which has its own id system and is never a library row. A
  **link import** (ADR 0110, owner decision 7) is saved as the list the
  fetcher returned, rendered as plain text (`source_format: "text"`),
  with the canonical link beside it in `decks.source_url`. Re-seating it
  re-parses the stored list and never calls the network; fetching it
  again from the link is a later, explicit action. Saving the same name
  from pasted text clears the link.
- **Saves only a deck the caller seats at their own seat** ([ADR
  0112](decisions/0112-signed-in-home-player-mode-and-one-decks-page.md)
  §3 item 8): the session's game and seat must be the path's game and
  the body's `player_id`, the rule `GET /me/last-deck` already has. An
  admin setting someone else's deck seats it and saves nothing to the
  admin's library. The person whose seat it is can save it from the
  decks page (`POST /me/decks`).
- Sets the seat's `deck_id` to the saved deck. Any other outcome
  (a guest, a catalog pick, someone else's seat, or a failed save)
  leaves the seat's `deck_id` empty, clearing a previous one if there
  was one.
- **The 200-deck cap (ADR 0110 section 6).** A signed-in caller keeps at
  most 200 decks. The 201st *new* deck is not saved (updating a deck by
  name is never refused), the upload still succeeds and seats the deck,
  and the response carries `library_note`, a player sentence saying the
  library is full and to delete a deck on `#/decks`.

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
    { "id": "esper-control", "name": "The Long Answer", "description": "control — …", "colors": ["W", "U", "B"], "commander": "Y'shtola, Night's Blessed" },
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

### `POST /games/practice` (ADR 0076, S54)

Open the tutorial's **practice table** for the caller, seated: one
human seat and one `random`-tier bot, each on its fixed tutorial deck
(`server/internal/decks/tutorial.go` — *First Steps* for the player,
*Practice Partner* for the bot). Any session may call it: a guest
seated at another table, a Discord sign-in, the admin. No body.

The table comes back **already started**, with the **opening roll**
open ([ADR 0125](decisions/0125-a-walkthrough-that-keeps-up.md) §5.2,
S65), exactly as at any table started with
[`POST /games/{id}/start`](#post-gamesidstart): nothing is dealt yet,
both seats roll a d20, ties roll again, and the winner chooses who
takes the first turn (CR 103.1). The player is the host, so they may
also `host_roll_remaining`. The tutorial's steps follow the player's
own first turn, so **a practice bot that wins the roll hands the first
turn to the player**: it sends the ordinary `choose_starting_player`
naming the player's seat, which CR 103.1 allows (the winner chooses any
player). A player who wins chooses for themselves, and may hand it to
the bot. A bot at any other table chooses as
[ADR 0121](decisions/0121-animated-dice.md) §4 has it. The response is a join's: a
fresh player session for the seat (`token`, `expires_at`,
`principal`), the `game` with `practice: true`, and `player_id`. The
session cookie is set to it. The seat carries the caller's Discord
identity and user, when the session had them.

A practice table is not like other tables:

- It is **never listed** by `GET /games`, to the admin included, and
  has no invite, so nobody else can join or watch it. `GET /games/{id}`
  still answers for it.
- It is **never persisted**: no `games` / `seats` rows (so it is not in
  `GET /me/games`), and its room writes no restore point, replay or
  crash dump. A restart ends it. There is no resume.
- **One per person.** Opening a second abandons the caller's first
  (the person is their user, else their Discord identity, else the
  admin credential, else their seat; a session that is itself a
  practice seat counts as whoever opened that table).
- **At most 8 open at once**, across everyone.
- **Reaped** after 20 minutes without a single commit, or 2 hours in
  all — the bot runner is stopped and any sockets closed.

| Status | Reason |
|---|---|
| 201 | opened; body as above |
| 401 | no session |
| 429 | this caller opened practice tables too quickly (or the IP deck bucket is empty) |
| 503 | no bot host; no card index; the `random` tier is not offered; or all 8 practice tables are in use |

### `POST /games/{id}/practice/leave` (ADR 0076, S54)

Abandon a practice table and put the caller's own session cookie
back. Called from every exit the tutorial has — Leave, a navigation, a
closed tab (as a `keepalive` request), and the next page load if that
one never landed — so it is idempotent and needs no session: the
credentials are in the body.

```json
{"practice_token": "<the practice seat's session>", "restore_token": "<the session to put back>"}
```

- `practice_token` authorises the leave. It must be a player session
  for this game's human seat. Empty, invalid or expired: nothing is
  left (the reaper will have the table) and the call still succeeds.
- `restore_token`, when it validates, becomes the session cookie
  again; otherwise the cookie is cleared. The server reads the cookie
  before the `Authorization` header, so this is the only way a client
  can stop talking as the practice seat.

`Content-Type` must be `application/json` (a cross-site form cannot
send it, so no other site can use this to plant a cookie). Rides the
same per-IP bucket as `POST /games/{id}/join`.

| Status | Reason |
|---|---|
| 204 | done, or there was nothing left to leave |
| 400 | malformed body or game id |
| 403 | `practice_token` is for this game but not its human seat |
| 409 | `{id}` is a real table, not a practice table |
| 415 | not `application/json` |
| 429 | rate-limited |

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
| Admin-only to mint | same admin gate as `POST /games` (`requireAdmin`). A player seated at the table cannot mint one, not even for their own seat, unless they are an allowlisted admin. |
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
`/c2-invite-dm <user> [game] [name]` slash command ([#613](https://github.com/krakenhavoc/cmd_and_ctrl/issues/613))
is a thin client of this route (it sends `discord_id` with the bot's admin session), so there is exactly one place that
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
| 424 | Discord rejected this server's bot credentials, or failed for another reason (not 502: Cloudflare, in front of both hosts, replaces an origin 502's body with its own "error code: 502" page (#1644)) |
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

The principal's fields are at the top level, as they always were, plus
computed fields that are not part of the token:

| Field | Meaning |
|---|---|
| `admin` | The effective answer (ADR 0110 §3 item 4): `true` for the shared token's session and for an allowlisted person **in admin mode**, `false` otherwise, player mode included. Every admin control in the client reads this. |
| `admin_allowed` | `true` when this person is on the allowlist and so may switch admin mode on, in either mode ([ADR 0112](decisions/0112-signed-in-home-player-mode-and-one-decks-page.md) §2 item 9). Persons only: `false` for the shared token. |
| `admin_mode` | `true` while this person has admin mode on. |
| `admin_mode_ends_at` | When admin mode lapses, in Unix milliseconds. Present only while `admin_mode` is `true`. |

The client asks once per installed session that has a `user_id`, and
shows admin UI to an admin seated at a table exactly as it does to the
token. It asks again after a 4001 close, when a hidden tab becomes
visible (for an allowlisted person), and when `admin_mode_ends_at`
passes; a switch takes its answer from `PUT /me/admin-mode`. It reads a
mode whose end time has passed as off without waiting for the answer. `/me` tells a person only about themselves: the allowlist itself
is never served.

```json
{ "role": "player", "user_id": "<uuid>", "game_id": "<uuid>", "player_id": "<uuid>",
  "discord_id": "<snowflake>", "issued_at": "…", "expires_at": "…",
  "admin": true, "admin_allowed": true, "admin_mode": true, "admin_mode_ends_at": 1759480800000 }
```

### `PUT /me/admin-mode` (ADR 0112 §2)

Switch the caller's admin mode on or off. **Allowlisted persons only**: a
session with a `user_id` whose Discord ID is on
`CMDCTRL_DISCORD_ADMIN_USER_IDS`, in either mode. See
[Who is an admin](#who-is-an-admin-adr-0110-3-adr-0112-2).

**Request**

```json
{ "on": true }
```

- `{"on": true}` switches admin mode on for 12 hours. Sent while it is
  already on, it restarts the 12 hours, which is how an admin extends it.
- `{"on": false}` switches it off. Sent while it is already off, it is a
  no-op that answers 200.

The switch is written to the database first and then takes effect on the
person's next request, from any session. The person's WebSockets whose
admin bit is now wrong are closed with **4001** `admin mode changed`, so
they reconnect with the new answer.

**Response 200**

```json
{ "admin": true, "admin_mode": true, "admin_mode_ends_at": 1759480800000 }
```

`admin` is the effective answer, as on `GET /me`. `admin_mode_ends_at`
is in Unix milliseconds and present only while admin mode is on.

**Errors**

| Status | Reason |
|---|---|
| 400 | the body is not `{"on": true}` or `{"on": false}` |
| 401 | no session, or an expired or revoked one |
| 403 | `not an admin`: anyone not on the allowlist, a guest, a reclaim ticket's session, and the shared token, which has no mode to switch |
| 404 | the user row no longer exists |
| 429 | more than one switch every 2 seconds per person, after a burst of 5 |
| 500 | the switch could not be saved. A failed switch on leaves player mode; a failed switch off is player mode in this process anyway |
| 503 | the server has no user database |

Every switch is logged at Info as `admin mode on` or `admin mode off`,
with `admin_user_id`; a lapse is logged as `admin mode lapsed`.

### `GET /me/games`

"My games" ([ADR 0051](decisions/0051-user-database.md) decision 4,
S34 sub-PR 4): every seat the caller's user holds, joined with its
game, **newest game first**. Ended and archived games are included,
as is a finished game whose table did not survive a restart; the
`games` and `seats` rows are the record.

Needs a signed-in person: an `identified` session, or a `player` or
`spectator` session with a non-nil `user_id` (a signed-in spectator
keeps their user since S55, [ADR 0110](decisions/0110-remember-me.md)
§1 item 2). A caller with no credential at all gets **401**; an
authenticated caller who is not a person gets **403** — a guest seat's
or guest spectator's session, an admin, and everyone on a deployment
with no database (there are no users).

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
  the engine reports a winner. `outcome` (`"win"` or `"draw"`) says how
  an ended table finished, which `winner_seat` alone cannot: a draw and
  a table an admin closed are both `winner_seat: null`. It is omitted
  when unknown, for a table that has not ended, one an admin closed,
  and any game that ended before `games.outcome` existed (migration
  0007, [ADR 0057](decisions/0057-win-and-lose-by-effect.md) Decision 7).
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
It expires with the caller's session, not 12 hours from now
([ADR 0110](decisions/0110-remember-me.md) §1). Both invite tokens are
stripped from the embedded `game`.

**Errors**

| Status | Reason |
|---|---|
| 401 | no session at all |
| 403 | authenticated but not a signed-in person (see `GET /me/games`) |
| 403 | the caller holds no seat at this table |
| 404 | the table is not live in this process |
| 409 | the table has been archived |

---

### `POST /me/session` (ADR 0110 §1 items 5 and 6, S55)

Renewal on use, and the reinstall of a saved identity
([ADR 0110](decisions/0110-remember-me.md), owner answer 1). It sets the
session cookie to the caller's own credential and returns that session.
When the session is **more than half spent** (more than half of the time
from its `issued_at` to its `expires_at` has passed), the server first
re-issues it for a fresh `CMDCTRL_IDENTITY_TTL`, and the cookie and the
response carry the new token.

**Caller:** a signed-in person: an `identified`, `player` or `spectator`
session with a `user_id`, the same rule as `GET /me/games`. Only those
sessions can be revoked, so nothing else is renewed or reinstalled.

**Credential:** the session cookie, else `Authorization: Bearer`. A
`?token=` query parameter is **never** read here, unlike every other
route. This route sets the cookie from the credential it reads, and a
cross-site form can put a token in a URL but cannot set a header (and
`SameSite=Lax` keeps the cookie off a cross-site POST), so the query
fallback would let another site plant its own session in this one. The
client's reinstall sends the saved token as a bearer once the session
that replaced it has expired, when the browser has already dropped that
cookie.

No body.

**Response 200**: the `sessionResponse` shape (`token`, `expires_at`,
`principal`, `player_id`), with no embedded `game`, and a `Set-Cookie` for
`token`.

- **Not yet half spent:** the same token and expiry the caller sent. This
  is the reinstall.
- **Renewed:** a new token for the **same principal**. Its role, `user_id`,
  `game_id`, `player_id`, `name` and `discord_*` fields are unchanged, so a
  seat session is still that seat and a spectator session still watches
  that table. Only `issued_at` and `expires_at` are new, and the new
  session lasts the full `CMDCTRL_IDENTITY_TTL`. The old token is **not**
  revoked: another tab may still hold it, and it runs out at its own
  expiry.

Renewal is how a signed-in session outlives its first 30 days. Revocation
is still the way to end one: a token issued before
`users.sessions_invalid_before` fails before the handler runs, and a
revocation that lands while a renewal is being minted refuses the
renewal rather than handing out a token the revocation missed.

**Rate limit:** a per-person bucket, one call every 6 seconds with a
burst of 10. The client calls the route at page load, at a session's
half-life, an hour after a renewal that renewed nothing or failed, and
when it reinstalls a saved identity.

**Errors**

| Status | Reason |
|---|---|
| 401 | no credential (a `?token=` alone counts as none), or one that is invalid, expired or revoked, including a revocation during the renewal |
| 403 | authenticated but not a signed-in person: the admin token, a guest seat or spectator, a reclaim ticket's session, or any session on a server with no database |
| 429 | the per-person bucket is spent |

**The client** (`client/src/lib/session.ts`):

- **Renewal.** A session with a `user_id` that is past half its life is
  renewed at page load and when a timer armed for its half-life fires.
  A failed renewal is quiet. There is no notice and no sign-out, the
  session lasts until it expires, and the client asks again an hour
  later.
- **The saved identity.** Installing a session with no `user_id` (the
  admin token, a reclaim ticket for someone else's seat) over a live
  signed-in session first copies that session to
  `localStorage["cmdctrl.identity"]`, which only this origin can read.
  When the session on top ends (its expiry timer, or a 401 for it), the
  client sends the saved token to this route as its bearer and installs
  the answer, so the person is still signed in rather than sent to
  `#/login`. A refused reinstall drops the saved copy and goes to
  `#/login` as before. Leaving a practice table whose replaced session
  has run out restores the saved identity through
  [`POST /games/{id}/practice/leave`](#post-gamesidpracticeleave-adr-0076-s54)'s
  `restore_token`. Signing out, signing out everywhere, and installing
  any session with a user clear the saved copy.
- **The cookie.** Both calls set the cookie when their answer arrives,
  so a request that mints a session (a join, a reclaim, a sign-out)
  waits for one in flight rather than race it. An answer for a
  different principal than the one sent (the cookie, read first, held
  another tab's seat) is not adopted.

---

### `GET /me/tablemates` (ADR 0051 decision 8, S34 sub-PR 6)

The people the caller has shared a table with, **most recently shared
table first** — the list the invite picker offers. Decision 8's whole
design: a self-join of `seats`, not a `friendships` table. There are
no friend requests and no acceptance; if an explicit list is ever
wanted it is one table on top of this and changes nothing here.

Same caller rule as `GET /me/games` and `GET /me/decks`: a signed-in
person — an `identified` session, or a `player` or `spectator` session
with a non-nil `user_id`. No credential is **401**; every other caller is
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

### `GET /me/setup` (ADR 0110 §5)

The caller's last table setup, for the create form's "Use my last setup".
Signed-in only: a guest and the admin token get 403 (never 401, #1154).

```json
{
  "setup": {
    "settings": { "undo_limit": 3, "undo_scope": "own", "starting_life": 40,
                  "commander_damage": 21, "bot_pace": "normal", "allow_spawn": false },
    "bots": [ { "tier": "heuristic", "deck_id": "raid-and-ransack", "name": "Bot 1" } ],
    "tablemates": [ "<user uuid>" ]
  },
  "game_id": "<the table it came from>",
  "updated_at": 1790000000000
}
```

`setup` is `null` (and the other fields absent) when the caller has none yet.
The body is written by the server when a table starts (see
[`POST /games/{id}/start`](#post-gamesidstart)), never by a client. A setup
lists the user IDs of the caller's own tablemates, the same class of data
`GET /me/tablemates` already returns to the same caller; the create flow
offers those people first, marked "at your last table".

### `GET /me/last-deck` (ADR 0110 §5)

The deck the caller last seated, for the deck panel to **preselect**. It is
never seated automatically: a seated deck is visible to the table.

```json
{ "last_deck": { "kind": "library", "id": "<deck uuid>" } }
```

`kind` is `library` (a `GET /me/decks` row) or `prebuilt` (a `GET /decks` ID).
`last_deck` is `null` when none is recorded. It is written whenever a signed-in
person seats a deck at their own seat: a library deck, an upload saved to the
library, or a pre-built deck. An admin installing somebody else's deck records
nothing. A deck that has since gone is simply not preselected. Signed-in only
(403 otherwise). A guest's last pre-built deck is kept in the browser instead
(`localStorage["cmdctrl.lastDeck"]`), and a guest's last typed name in
`localStorage["cmdctrl.guestName"]` (owner answer 8: a signed-in seat keeps its
Discord name and is never asked).

### `GET /me/decks` (ADR 0051 decision 7, S34 sub-PR 5)

The caller's deck library — see "The deck library" under
`POST /games/{id}/decks` above for how a row gets there and the update
rule. Newest updated first.

Same caller rule as `GET /me/games`. No credential is **401**; every
caller that is not a signed-in person is **403** — a guest's `player`
session, an admin session, or an `identified` session on a deployment
with no database. Before S55 this was a 401, which the client reads as
an expired session and signs the browser out
([ADR 0110](decisions/0110-remember-me.md) §1 item 7, #1154). There is
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
      "updated_at": "2026-09-19T08:00:00Z",
      "source_url": "https://moxfield.com/decks/abc",
      "coverage": {
        "counts": { "manual": 3, "unreviewed": 5, "caveats": 4, "automated": 40, "no_effect": 38 },
        "unknown": 0,
        "as_printed": 78,
        "resolved": 90
      }
    }
  ]
}
```

`source_text` and `source_format` are not included here — this is the
picker's list, not the re-seat payload; seating reads them server-side
via `POST /games/{id}/decks/{deck_id}`. `source_url` is absent for a
pasted deck.

**`coverage` is computed on read, never stored** (ADR 0110 section 6).
`counts` are [ADR 0095](decisions/0095-deck-coverage-and-deck-requests.md)'s
five buckets by distinct card. `as_printed` is `automated` plus
`no_effect`, `resolved` is every distinct card bucketed. Since #2220
the same block carries the copies form: `copies` (the buckets summed
over each card's copies), `unknown_copies`, `as_printed_copies`
(`automated` plus `no_effect` copies) and `deck_size` (every copy,
unresolved ones included). The page says "N of M cards play as
printed" from `as_printed_copies` and `deck_size`. The list builds the catalogue verdicts
once per request and reuses them for every deck. `coverage` is absent
when the server has no card index or a stored list no longer parses.
These reads, and the four routes below, share a per-client bucket of 1
request per second with a burst of 5 (429 past it).

### `POST /me/decks` (ADR 0112 §3 item 4)

Save a checked deck to the caller's library, from the decks page. A
check writes nothing; saving is always this explicit call, and it never
files a request (`POST /deck-requests` is its own button with its own
limit).

**Request**

```json
{ "url": "https://archidekt.com/decks/424242", "name": "Weekend deck" }
{ "text": "1 Atraxa, Praetors' Voice *CMDR*\n1 Sol Ring\n..." }
```

Exactly one of `url` and `text` (400 for neither or both); `name` is
optional.

- **A pasted list** is read the way `POST /deck-coverage` reads one and
  saved as pasted (`source_format: "text"`, no `source_url`).
- **A link** is saved as the list the server fetched, rendered as plain
  text, with the canonical link in `source_url`: the same row an import
  at a table writes. Seating it or requesting its cards later re-reads
  the stored list and never calls the network. The deck check's
  ten-minute cache keeps the fetched list beside the report, so a save
  after a check fetches nothing. A link the cache does not hold first
  spends one token from `POST /deck-coverage`'s per-IP bucket (1 per
  10 s, burst 3), then fetches, so saving cannot get around the fetch
  limit. With the bucket empty the answer is **429**
  `{"error": "…"}` with `Retry-After: 10`, and nothing is fetched. A
  fetch failure answers with `POST /deck-coverage`'s error table,
  including the Moxfield `paste_list` hint.
- **The name** is `name`, trimmed, at most 100 characters (400 past
  it). Without one: the deck site's name for a link, then the first
  commander, then "Untitled deck".
- **Nothing is validated against Commander rules here**, and names the
  card index cannot resolve are kept: the check showed both, the
  library's `coverage.unknown` counts the names, and seating the deck
  runs the full parse, resolve and validate pipeline as it always does
  (`POST /games/{id}/decks/{deck_id}`, 422 with the violations).
- **The update rule is the library's own:** a deck the caller already
  has by that name is replaced in place (same `id`), and the answer
  says so. A pasted list saved over a link deck clears the link.

**Response:** **201** for a new deck, **200** for a replaced one:

```json
{
  "deck": { "id": "<uuid>", "name": "Weekend deck", "commanders": ["…"], "card_count": 100,
            "updated_at": "…", "source_url": "https://archidekt.com/decks/424242",
            "coverage": { "counts": { "manual": 1, "…": 0 }, "unknown": 0, "as_printed": 78, "resolved": 90 } },
  "replaced": false
}
```

`deck` is the `GET /me/decks` entry, coverage included.

**Errors**

- **401** with no credential. **403** for any caller that is not a
  signed-in person: a guest seat or spectator, the admin token, or an
  `identified` session with no database (never 401, which would sign
  the browser out).
- **409** `{"error": "Your deck library is full (200 decks). Delete one
  below to save this one.", "code": "library_full"}` for a new name
  past the 200-deck cap. Replacing a deck is never refused.
- **400** for a body that is not one of the shapes above, a list over
  the length cap, or a link that is not a Moxfield or Archidekt deck.
- **503** with no card index.

The route rides the `/me/decks*` bucket (1 request per second, burst 5,
per client).

### `GET /me/decks/{id}/coverage` (ADR 0110 section 6)

The full ADR 0095 coverage report for one saved deck (the same body as
`POST /deck-coverage`), so the library can show the cards behind each
bucket and offer "Request these cards" through `POST /deck-requests`.
`403` for a session that is not a signed-in person (never 401, which would sign the browser out); `404` for a deck that is not the caller's or
does not exist (an id reveals nothing); `422` if the stored list no
longer parses; `503` with no card index.

### `PATCH /me/decks/{id}` (ADR 0110 section 6)

Body `{"name": "..."}`. Renames the caller's deck without touching its
`updated_at`. Returns the deck (`myDeckInfo` without coverage). `400`
for a blank name or one over 100 characters, `403` for a caller with no
`user_id`, `404` for a deck that is not theirs, `409` when they already
have another deck by that name (the update rule keys on the name).

### `DELETE /me/decks/{id}` (ADR 0110 section 6)

Removes the caller's deck. In the same transaction every seat that
pointed at it has `seats.deck_id` set to NULL; the seat keeps its
`deck_name`. `204` on success; `403` for a caller with no `user_id`;
`404` for a deck that is not theirs or already gone.

### `GET /me/settings` and `PUT /me/settings` (ADR 0110 §4)

The caller's **account copy of their settings**: the per-person half of
the client's `Settings`, so they follow a signed-in person to every
browser they sign in on
([ADR 0110](decisions/0110-remember-me.md) §4, owner answers 5 and 6).
Which fields that is is decided in one place, `SYNCED_FIELDS` in
`client/src/lib/settings.ts`. The per-device fields (volumes and mute,
card size, hand and table layout, opponent detail, the two expansion
settings, text scale, and the two that follow the OS reduced-motion
signal) never leave the browser.

The server checks the body's shape only — a JSON object of at most
32 KiB and depth 4. It never reads the fields. The client's `migrate`
is the one schema validator, so a server copy of the schema cannot
drift from it.

Same caller rule as the rest of `/me/*`: a signed-in person. No
credential is **401**. Every other caller is **403**, never 401 (#1154):
a guest's seat or spectator session, the admin token, and everyone on a
deployment with no database. The client reads 403 as "keep settings in
this browser". Each caller reads and writes only their own row.

**`GET /me/settings` → 200**

```json
{
  "version": 15,
  "revision": 4,
  "settings": { "display": { "theme": "dark" }, "gameplay": { "strictMana": false } },
  "updated_at": 1790000000000
}
```

- `version` is the client `SETTINGS_VERSION` that wrote the copy.
- `revision` goes up by one on every write. It is what `If-Match` names.
- `updated_at` is Unix milliseconds.
- A person with no copy yet gets `{"revision": 0}` and nothing else.

Responses carry `Cache-Control: no-store`.

**`PUT /me/settings`**

```
If-Match: 4
```

```json
{ "version": 15, "settings": { "display": { "theme": "light" } } }
```

`If-Match` is the revision the client last read, or `0` for the first
copy. A bare integer and a quoted entity tag (`"4"`) are both accepted.
The body replaces the whole copy. Unknown top-level fields are refused.

| Status | When | Body |
|---|---|---|
| 200 | Saved | The new copy, as `GET` returns it |
| 400 | `If-Match` is not a revision. Or the body is malformed: `settings` missing or not an object, nested deeper than 4, or `version` outside 1–1000 | `{error}` |
| 409 | `version` is **below** the stored copy's: a stale tab must not stamp an older schema over a newer client's copy | `{error, …}` with the current copy. The error says to reload |
| 412 | The revision has moved (another tab or device wrote first), or `If-Match: 0` when a copy exists | `{error, …}` with the current copy |
| 413 | `settings` is over 32 KiB | `{error}` |
| 428 | No `If-Match` | `{error}` |
| 429 | Over the per-person bucket: 1 write a second, a burst of 5, shared by all of that person's sessions. Only a write that passed the checks above spends a token. `Retry-After` is set | `{error}` |
| 503 | The server has no settings store | `{error}` |

The revision is checked before the version, so a stale revision is
always a 412.

**How the client uses it** (`client/src/lib/settingsSync.ts`):

- It downloads at sign-in and at page load while signed in.
- It uploads the synced fields one second after the last change.
- At sign-in the **account's copy wins**. If this browser's values
  differ, a toast offers "Keep this browser's instead" for the rest of
  the visit, which uploads them.
- On a 412 it merges field by field against the copy both sides last
  agreed on, and then uploads over the new revision. A field only this
  browser changed keeps this browser's value. Every other field takes
  the account's.
- A copy written by a newer client, or a 409, is applied and then this
  tab stops writing until it reloads.
- While the tutorial's practice table is open, the four settings it
  forces are uploaded as the player's own values
  (`withSettings(current, record.saved)`), never as the forced ones.
- A failed write never blocks the UI. It is retried with a backoff, and
  again when the browser comes back online.

### Playmats (ADR 0128)

A signed-in person can keep up to **three playmats** and show one of
them: an image every player at the table sees behind *their*
battlefield, as with a mat on a paper table. Bots, guests and agent
seats have none. The reasons for every choice here are in
[ADR 0128](decisions/0128-playmats.md); §11 is the three-slot design
and the best-size fit.

A person's mats live in **slots 1 to 3**, and **at most one is
active**: the one the table shows. "None" is allowed, so a person can
keep saved mats and show none. Switching the active mat never deletes a
file; removing or replacing a slot deletes its file.

| Route | Who | What |
|---|---|---|
| `GET /me/playmats` | a signed-in person | the saved slots, which one is active, the wash |
| `PUT /me/playmats/{slot}` | a signed-in person | upload into a slot: `multipart/form-data`, part `file` |
| `POST /me/playmats/{slot}/link` | a signed-in person | `{"url": "https://…"}`: the server fetches it once and stores it in the slot |
| `POST /me/playmats/{slot}/fit` | a signed-in person | `{"x": n, "y": n}`: crop the slot's image to the best size, the crop's top-left corner in the stored image's pixels |
| `DELETE /me/playmats/{slot}` | a signed-in person | remove the slot's playmat |
| `PUT /me/playmats/active` | a signed-in person | `{"slot": n \| null}`: which slot the table shows; `null` shows none |
| `PATCH /me/playmats` | a signed-in person | `{"wash": 30..90}`: how dark the active playmat is under the cards (ADR 0128 §10) |
| `GET /playmats/{id}` | any session | the image |
| `DELETE /admin/users/{id}/playmats/{slot}` | admin | remove one of anyone's playmats (moderation) |
| `DELETE /admin/users/{id}/playmat` | admin | remove all of anyone's playmats |

The v1 singular routes (`/me/playmat`, `/me/playmat/link`) are gone:
nothing but the client called them, and the client moved with the server.

Same caller rule as the rest of `/me/*`: no credential is **401**; a
guest's seat or spectator session, the admin token and every session on
a deployment with no database are **403**, never 401 (#1154). A server
with a database but no `CMDCTRL_DATA_DIR` has nowhere to store an
image: `GET /me/playmats` answers `{"enabled": false}` and the writes
are **503**. The client hides the Settings section in both cases.

A `{slot}` outside 1 to 3, or not a whole number, is **400** on every
route that takes one (the table also has a `CHECK`). Uploading into an
**empty slot while no mat is active** makes it active; any other write
leaves the active mat alone, except that replacing or fitting the
**active** slot moves the active pointer to the new image.

**The body of every `/me/playmats` route**, with `Cache-Control:
no-store`:

```json
{
  "enabled": true,
  "max_slots": 3,
  "ideal_width": 2400,
  "ideal_height": 1400,
  "slots": [
    { "slot": 1, "url": "/playmats/6f1c2a9e-1b2c-4d3e-8f40-0123456789ab", "width": 2400, "height": 1400, "fits": true },
    {
      "slot": 3, "url": "/playmats/0d1e2f30-4a5b-4c6d-8e7f-8091a2b3c4d5", "width": 2560, "height": 1800, "fits": false,
      "suggestion": {
        "target_width": 2400, "target_height": 1400,
        "crop": { "x": 0, "y": 153, "width": 2560, "height": 1493 },
        "smaller": false
      }
    }
  ],
  "active": 1,
  "slot": 3,
  "wash": 58
}
```

`slots` holds the occupied slots in slot order and is absent for none;
`active` is absent when none is on show; `slot` is present only on the
answer to an upload, link or fit, and names the slot it wrote, so a
client can open the best-size prompt on that slot's `suggestion`;
`wash` is sent while the feature is enabled. Every `url` is this
server's own route, never a link that was pasted.

**What is stored.** The bytes sent are never stored. They are decoded
(PNG, JPEG or WebP, by content, never by `Content-Type` or file name;
anything else, including GIF and SVG, is **415**), refused over 10 MB
(**413**) or over 40 megapixels (**413**, judged from the header before
any pixel buffer exists), downscaled so the long edge is at most 2560 px,
and re-encoded as a JPEG at quality 85. That strips EXIF (a phone photo
carries the location), XMP and anything appended to the original; the
EXIF orientation is applied first, so a portrait photo is not stored
sideways. Transparency is flattened onto white. The file is
`<data dir>/playmats/<uuid>.jpg`, a `user_playmats` row (migration
0013) names it for its slot, and `users.playmat_id` points at the
active one. A replacement is a new uuid, the old file is deleted, and
the old URL is a 404 from then on. Removing deletes the file.

**`POST /me/playmats/{slot}/link`.** The server makes the request, once,
and stores the result exactly like an upload; the link is never kept or
shown to anyone, and no other player's browser ever contacts that host.
The fetch is guarded against SSRF (ADR 0128 §3): `https` only (an
`http` link is refused); the address is checked **where it is
dialled**, on the resolved IP, so a name that resolves to a private
address, or rebinds to one, is refused; loopback, private, link-local
(including `169.254.169.254`), carrier-grade NAT, multicast,
unspecified, unique-local and reserved ranges are all refused, and an
IPv4 address written as IPv6 (`::ffff:`, NAT64, 6to4) is judged by the
IPv4 address inside it; every redirect is checked the same way, at most
3, and a redirect to `http` is refused; 10 seconds for the whole
request; the body is capped at 10 MB; no proxy, cookies, credentials or
`Referer`. A refusal is **422** `{error}` that says what to do and never
echoes what the guard saw.

**The best size, `fits` and `suggestion`** (ADR 0128 §11). Playmats look
best at `ideal_width` x `ideal_height`, **2400 x 1400**, the shape of a
paper playmat (24 x 14 in, 12:7). A stored image **fits** when its
aspect ratio is within 5% of 12:7 and its long edge is at least 1600 px.
An image that fits has `"fits": true` and no `suggestion`. When the
shape is off, `suggestion` carries the target size, the **centred crop**
rectangle in the stored image's own pixels (the largest 12:7 rectangle
that fits), and `smaller`: true when the result will be smaller than the
ideal, because **the server never upscales**; a small image is cropped
to the shape at its own resolution and the client warns it may look soft
at the table. A right-shaped image that is merely small has `"fits":
false` and no `suggestion`: a crop would change nothing, and the fix is
a bigger image.

**`POST /me/playmats/{slot}/fit`** takes `{"x": n, "y": n}`, the crop's
top-left corner. The crop's size is the server's (`suggestion.crop`'s
width and height); the client only chooses where the window sits along
the axis being cropped. The server validates that the rectangle lies
inside the stored image (**400** otherwise), crops the stored image,
scales the crop down to the target (never up), re-encodes it through the
same JPEG path, writes a **new file under a new uuid** (so a cache never
serves the old image), moves the slot's pointer, and the active pointer
if that slot was active, and deletes the old file. If the slot has no
`suggestion` (the image is already the best shape) it is a **409**; an
empty slot is a **404**; a slot that changed while the fit was being
made is a **409** to ask again. The answer is the same body as above.
A fit works from the stored image, which is already normalised to a
long edge of at most 2560 px: a portrait source has lost resolution
before the fit sees it, because the server keeps no original (disk use,
and EXIF never reaches disk).

| Status | When |
|---|---|
| 200 | Stored, activated, fitted or removed (removing an empty slot is a 200) |
| 400 | A slot outside 1 to 3; not multipart, or no `file` part; a link or fit body that is malformed; a crop outside the image; an activation body that is not `{"slot": 1..3 \| null}` |
| 401 / 403 | See the caller rule above |
| 404 | `PUT /me/playmats/active` or a fit names an empty slot |
| 409 | A fit of an image that is already the best shape, or one that lost a race with a replacement |
| 413 | Over 10 MB, or over 40 megapixels |
| 415 | Not a PNG, JPEG or WebP |
| 422 | The link was refused or could not be fetched |
| 429 | Upload, link, fit and remove: over the per-person bucket (a burst of 5, then one every 6 seconds, shared by all four) or, for the three that decode, the per-IP one (1 a second, a burst of 10). Activate and the wash share a cheaper bucket (one a second, a burst of 5). `Retry-After` is set |
| 503 | No data directory; or too many images are being decoded at once (`Retry-After: 5`: at most 2 at a time) |

**`GET /playmats/{id}`** needs a session, like `/avatars` and `/cards`,
and any session may fetch any id: every player must see every other
player's mat. There is no listing route. The `<img>` carries the session
as `?token=`, as an avatar's does. Served as `image/jpeg` with
`X-Content-Type-Options: nosniff`, `Content-Security-Policy:
default-src 'none'; sandbox`, `Referrer-Policy: no-referrer` and
`Cache-Control: private, max-age=31536000, immutable` (the bytes at a
URL never change). A malformed or unknown id is **404**, the same
answer for both.

**On the wire.** The seat's owner's **active** playmat is
`PlayerView.playmat_url` ([docs/protocol.md](protocol.md)), stamped by
the room on every capture, so a change reaches everyone on the next
state broadcast. Activating, removing or fitting the active mat, or
replacing the active slot, is pushed to every live seat the person holds
the way an upload was in v1; saving into an inactive slot costs the
table nothing, because nothing it shows changed.

**The wash** (ADR 0128 §10) is the owner's, one per account and not per
mat: how strongly their active playmat is darkened under the cards, a
whole percentage from 30 to 90, 58 until they set one. Every
`/me/playmats` answer carries it as `wash` while the feature is
enabled. `PATCH /me/playmats` with `{"wash": n}` sets it, image or not;
out of range, a non-number or any other field is **400**. The table sees
it as `PlayerView.playmat_wash` beside the URL. It is stored in
`users.playmat_wash` (migration 0012).

**`DELETE /admin/users/{id}/playmats/{slot}`** and **`DELETE
/admin/users/{id}/playmat`** are `requireAdmin`, so admin mode off gets
a non-admin's 403. The first removes one saved playmat, the second all
of them. **204**, also for a person with none; **404** for an unknown
user; **400** for a malformed id or a slot outside 1 to 3. The person
can upload again; this removes images and does not ban the feature. The
admin account view calls the per-slot route, one Remove per thumbnail.

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
{ "text": "1 Atraxa, Praetors' Voice *CMDR*\n1 Sol Ring\n..." }
{ "deck_id": "<uuid>" }
{ "url": "https://moxfield.com/decks/AbC123",
  "requester": { "discord_id": "123456789012345678", "display_name": "Alice" } }
```

Exactly one of `url`, `text` and `deck_id` (400 for none or more than
one). A pasted list
is deduplicated by its list key (`deck_key`, see `POST /deck-coverage`),
so the same deck pasted from two exports joins one issue, and the same
cards under another commander file another. Since the [2026-09-25
amendment](decisions/0095-deck-coverage-and-deck-requests.md#amendment-2026-09-25-requests-from-a-pasted-list):
Moxfield blocks this server, so pasting is a Moxfield player's only
route.

**`deck_id`** ([ADR
0112](decisions/0112-signed-in-home-player-mode-and-one-decks-page.md)
§3 item 5) is one of the caller's own saved decks (`GET /me/decks`),
so every library deck can ask for its missing cards. A deck that is not
theirs, or does not exist, is **404**, as on the other `/me/decks`
routes. The stored list is re-read, never fetched, and keyed the way
the original would have been: a deck with a `source_url` keys as that
link (`moxfield:<id>`, `archidekt:<id>`), so it joins the issue a
request from the link filed, and any other deck keys as its list
(`list:<hash>`). The admin token's `requester` path does not take
`deck_id` (**400**): the bot has no library. Everything below (the
limit, the outcomes, the issue) is the same whichever input named the
deck.

**What it does**, in order:

1. **Rate limit.** Three asks per requester per rolling 24 hours, keyed
   `discord:<snowflake>` whether the ask came from the site or the bot,
   and counted in the database (`deck_request_asks`) so a deploy does
   not reset it. Over the limit is **429** `rate_limited`, decided
   before the deck is fetched. Only an ask that reaches GitHub (an
   issue filed, or a comment added) counts. The route also rides the
   ordinary per-IP bucket (1/s, burst 5); the bot's calls spend it in
   the handler, for the member they name.

   **An admin is not rate-limited** ([#2052](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2052)):
   when the requester is an admin by the rule in [Who is an admin](#who-is-an-admin-adr-0110-3-adr-0112-2)
   (allowlisted and in admin mode), neither limit applies. A person in
   player mode is limited like everyone else. Through the bot, the
   **named member** decides: they are exempt only if their own account
   (looked up by Discord ID) is allowlisted and in admin mode. The
   bot's admin session exempts nobody it speaks for. An admin's asks
   are still recorded, and deduplication is unchanged: the same deck
   still joins the same issue.
2. **The report.** The deck is fetched (or the list parsed, or the
   saved deck's stored list read) and
   bucketed exactly as `POST /deck-coverage` does it, from the same
   10-minute cache (a saved deck with a link is built from its stored
   list instead, so the cache never holds a stale copy under the
   link's key). A fetch failure answers with that route's error
   table, the Moxfield hint included.
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
when the deck has none, which a pasted list never does), with labels
`enhancement` and `deck-request`.
If GitHub refuses the labels, the issue is filed without them, as
`POST /bugreport` does. The body holds the deck link, the requester's
display name (never the snowflake), the counts, `## Cards to add` and
`## Cards to review` checklists (`- [ ] Name (oracle id)`),
`## Automated with caveats`, any names the index could not resolve, and
a footer naming the surface that filed it. For a pasted list the deck
line reads "Pasted list" instead of a link, and the list itself follows
in a collapsed `<details>` block: `N Name` lines, commanders first and
marked `*CMDR*`, rendered from the card index's names and never from
the raw paste. The deck name, the display name and any unresolved name
went through ADR 0017's redaction before they are published, and an `@`
in them cannot ping anyone.

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
- **424** when GitHub fails while reading the deck's issue, commenting
  or filing (not 502, which Cloudflare replaces with its own page).

## Session lifetimes (ADR 0110 §1)

One rule for every route that mints a session
([ADR 0110](decisions/0110-remember-me.md) §1, S55, reversing
[ADR 0051](decisions/0051-user-database.md) sub-PR 7's "seat sessions
keep `CMDCTRL_SESSION_TTL`"). It lives in one function, `issueFor`
(`server/internal/lobby/session_ttl.go`):

| The new session | Lifetime | Routes |
|---|---|---|
| has a `user_id`, and came from a Discord sign-in | `CMDCTRL_IDENTITY_TTL` (30 days by default) | `GET /auth/discord/callback`, all three branches: the login page's identity session, the invite link's seat, and linking Discord to a seat |
| has a `user_id`, and renews one more than half spent | `CMDCTRL_IDENTITY_TTL` from now, same principal | [`POST /me/session`](#post-mesession-adr-0110-1-items-5-and-6-s55) |
| has a `user_id`, and came from a signed-in session | **the source session's own `expires_at`**, to the millisecond | `POST /games/{id}/join`, `POST /join`, `POST /games/{id}/spectate`, `POST /me/games/{id}/session`, `POST /games/practice`, and `POST /games/{id}/reclaim` redeemed by the seat's own user |
| has no `user_id` | `CMDCTRL_SESSION_TTL` (12 hours by default) | guest seats and spectators, `POST /admin/login`, a reclaim ticket on its own |

So joining, watching, practising or reclaiming a seat never shortens a
sign-in and never extends one. Only a new Discord sign-in, or the
renewal on use behind `POST /me/session`, extends it. Every session with a `user_id` can be
revoked (`POST /logout/everywhere`, below), which is what makes the
long lifetime safe. The `identified` session on a deployment with no
database has no `user_id` and still lasts `CMDCTRL_IDENTITY_TTL`, as it
did before.

**A signed-in spectator keeps their user** (§1 item 2):
`POST /games/{id}/spectate` with a signed-in session (cookie or bearer)
mints a `spectator` session that carries the caller's `user_id` and
`discord_*` fields, labelled with their Discord display name (a typed
`name` is for guests). `/me/*` keeps answering for it, the WebSocket
binding carries the user, and signing out everywhere closes it. A
guest spectator is unchanged.

## Signing out

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
to now, ends their admin mode if it was on (in the same statement, ADR
0112 §2 item 7), closes the user's open game WebSockets, and clears the
cookie ([ADR 0051](decisions/0051-user-database.md) decision 6). From then on,
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
| 403 | the session has no user: admin, guest seat or guest spectator, or any session on a server with no database. `POST /logout` is their sign-out |
| 404 | the user row no longer exists |
| 503 | the server has a user session but no revocation list (not a production configuration) |

### `POST /admin/users/{id}/revoke-sessions` *(admin only)*

ADR 0051's "admin remove-user". It does what the name says and nothing
more: the same revocation as `/logout/everywhere`, for the user `{id}`,
including the end of their admin mode. Every row stays: the user, identities, seats, games and decks. The
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

## Admin views (ADR 0124)

Read-only views of accounts and tables for an admin: the shared token,
or an allowlisted person in admin mode. They are what the Grafana
Overview's tiles link to, through the client's `#/admin/…` pages
([ADR 0124](decisions/0124-admin-views-accounts-games-and-who-is-on-now.md)).

- **Who may call them:** `requireAdmin`, like every admin route. No
  session is a 401. Any other session is a 403 `{"error":"admin only"}`,
  and an allowlisted person in player mode gets that same 403.
- **No database:** each of the four answers 503 `admin views need the
  user database, and this server has none`.
- **Times** are Unix milliseconds, absent when unknown. Every answer has
  `generated_at`.
- **Filters:** a missing or empty parameter means "any". An unknown value
  is a 400 whose error names the parameter.
- **What is never served:** Discord snowflakes as a field (an
  `avatar_url` path contains one, as it does everywhere the site shows an
  avatar), refresh tokens, scopes, invite tokens and hashes, reclaim
  tickets, deck lists, deck-request requesters, synced settings, table
  setups, admin mode and allowlist membership.
  `TestAdminViewsServeOnlyTheirFields` pins each route's fields.
- **The database is the record and memory the overlay.** History comes
  from SQL. A table's room in memory (`loaded`) gives its fresher state,
  its seats' kinds, agent clients and host, and practice tables, which
  have no row.
- **Logging:** each request is one `admin action` line with the method
  and path, never the query string, so a filter is never logged.
- **No rate limit** of their own: only admins reach them, and every query
  is bounded.

### `GET /admin/users`

Every account, one row each, at most 1,000; past that `truncated` is
`true`. Sorted playing now first, then by the latest play, then by the
latest sign-in, newest first.

| Parameter | Values |
|---|---|
| `played` | `1d`, `7d` or `30d`: an account seated at a started table whose end (`ended_at`, else `archived_at`) is in the window, or seated at a running table now. Exactly the accounts `cmdctrl_users_played{window}` counts. |

```json
{
  "generated_at": 1791206364278,
  "truncated": false,
  "accounts": [
    {
      "id": "<user uuid>",
      "name": "Ann",
      "avatar_url": "/avatars/<snowflake>/<hash>.png",
      "first_seen_at": 1791000000000,
      "last_sign_in_at": 1791200000000,
      "games_played": 12,
      "last_played_at": 1791100000000,
      "playing_now": false
    }
  ]
}
```

`first_seen_at` is the first Discord sign-in and `last_sign_in_at` the
latest. `games_played` counts distinct tables the account holds a seat at
that have started, in any state, and `last_played_at` is the latest end
of those. `playing_now` is a seat at a running, non-archived, non-practice
table in memory.

### `GET /admin/users/{id}`

One account: its row as above, its sign-in state, its tables (at most
500, newest first, then `games_truncated`), its saved decks, and its deck
requests (at most 100, newest first, then `deck_requests_truncated`).

```json
{
  "generated_at": 1791206364278,
  "account": { "id": "<uuid>", "name": "Ann", "games_played": 12, "playing_now": false },
  "playmats": [
    { "slot": 1, "url": "/playmats/<uuid>", "active": true },
    { "slot": 3, "url": "/playmats/<uuid>" }
  ],
  "sign_in": {
    "last_sign_in_at": 1791200000000,
    "discord_linked_at": 1790000000000,
    "sessions_invalid_before": 1791150000000,
    "revoke_path": "/admin/users/<uuid>/revoke-sessions"
  },
  "games": [ { "id": "<game uuid>", "their_seat": 0, "…": "a table row, below" } ],
  "games_truncated": false,
  "decks": [
    {
      "id": "<deck uuid>",
      "name": "Atraxa superfriends",
      "format": "moxfield",
      "source_url": "https://moxfield.com/decks/…",
      "commanders": ["Atraxa, Praetors' Voice"],
      "card_count": 100,
      "created_at": 1790500000000,
      "updated_at": 1790600000000
    }
  ],
  "deck_requests": [
    { "deck_key": "moxfield:abc", "asked_at": 1791000000000, "issue_number": 1234, "issue_url": "https://github.com/…/issues/1234" }
  ],
  "deck_requests_truncated": false
}
```

Sessions are HMAC tokens with no row ([ADR 0044](decisions/0044-surviving-a-deploy.md)
decision 3), so there is no list of them. `sessions_invalid_before`,
absent while unset, is the only per-person session state: a session
issued at or before it is refused. `revoke_path` is
[`POST /admin/users/{id}/revoke-sessions`](#post-adminusersidrevoke-sessions-admin-only).
A deck has no list and no coverage report. A deck request is matched to
the account through its Discord identity in SQL. `playmats` are the
account's saved playmats in slot order, absent for none, each with its
slot, its `/playmats/<uuid>` path and `active` on the one on show; the
account view's Remove button on each thumbnail calls `DELETE
/admin/users/{id}/playmats/{slot}` ([ADR 0124](decisions/0124-admin-views-accounts-games-and-who-is-on-now.md) amendment, [ADR 0128](decisions/0128-playmats.md) §11).

| Status | Reason |
|---|---|
| 400 | `{id}` is not a uuid |
| 404 | no such account |

### `GET /admin/games`

Tables, newest first, keyset-paginated.

| Parameter | Values |
|---|---|
| `state` | `lobby`, `active` or `ended` |
| `archived` | `true` or `false` |
| `practice` | `exclude` (the default), `include` or `only`. Practice tables have no row and come from memory: `include` puts them first on the first page, `only` lists just them, unpaginated. |
| `user` | an account id: the tables it holds a seat at |
| `limit` | 1 to 200, default 50; larger is capped at 200 |
| `cursor` | a `next_cursor` from the page before |

```json
{
  "generated_at": 1791206364278,
  "games": [
    {
      "id": "<game uuid>",
      "name": "Friday Night Commander",
      "state": "ended",
      "created_at": 1791200000000,
      "started_at": 1791200100000,
      "ended_at": 1791203000000,
      "archived_at": 1791204000000,
      "outcome": "win",
      "winner_seat": 2,
      "creator": { "id": "<user uuid>", "name": "Ann" },
      "practice": false,
      "loaded": false,
      "seats": [
        { "seat": 0, "kind": "human", "account": { "id": "<uuid>", "name": "Ann", "avatar_url": "/avatars/…" }, "deck_name": "Atraxa", "host": true },
        { "seat": 1, "kind": "human", "guest_name": "Gus", "host": false },
        { "seat": 2, "kind": "human", "guest_name": "Dee", "discord_pending": true, "host": false },
        { "seat": 3, "kind": "bot", "guest_name": "Bot 1", "bot_tier": "heuristic", "host": false },
        { "seat": 4, "kind": "agent", "guest_name": "Claude", "agent_client": "claude-code", "host": false }
      ]
    }
  ],
  "next_cursor": "<opaque>"
}
```

A table row:

- `state` is the lobby's cached state for a loaded table, which is
  fresher than the row.
- `outcome` is `win` or `draw` as recorded; `closed` for a table that
  ended, or was archived after it started, with none recorded (as
  `cmdctrl_games_ended_total{outcome="closed"}` counts it); absent for a
  table that has not ended.
- `creator` is absent for a table the token created.
- `loaded` is whether the table's room is in memory now. A row in state
  `lobby` or `active` that is not loaded did not come back after a
  restart.
- Each seat's `kind` is `human`, `bot` or `agent`, as `cmdctrl_seats`
  counts them. `account` is set for a signed-in seat, `guest_name`
  otherwise. `discord_pending` marks a Discord seat whose person has not
  signed in again; the snowflake is not served. `agent_client` comes from
  memory for a loaded table, else from `seats.agent_client` (migration
  0010), and a row written before that migration has none.
- `spectators_connected` and each seat's `connected` are the live sockets
  at a loaded table, counted as [`GET /admin/live`](#get-adminlive-admin-only)
  counts them: every socket bound to a seat is that seat's, an admin's
  included, and a spectator is a read-only socket that is not an admin's.
  They are absent for a table that is not loaded, and on a server with no
  WebSocket hub; absent never means zero.

### `GET /admin/games/{id}`

One table: the row above, flattened beside `generated_at`, plus its live
`connections`, earliest first. A practice table is served from memory.

```json
{
  "generated_at": 1791206364278,
  "id": "<game uuid>",
  "…": "the table row",
  "connections": [
    { "kind": "seat", "seat": 0, "account": { "id": "<uuid>", "name": "Ann", "avatar_url": "/avatars/…" }, "since": 1791206300000 },
    { "kind": "spectator", "since": 1791206310000 },
    { "kind": "admin", "since": 1791206320000 }
  ]
}
```

Each connection's `kind` is `seat`, `spectator` or `admin`; an admin
bound to a seat is an `admin` with that `seat`. `account` is absent for a
guest and for the shared token. `since` is when the hub admitted the
socket. A table that is not loaded has `connections: []`; the field is
absent only on a server with no WebSocket hub.

| Status | Reason |
|---|---|
| 400 | `{id}` is not a uuid |
| 404 | no such table in the database or in memory |

---

## Live now (ADR 0124 §3.4)

### `GET /admin/live` *(admin only)*

Who is connected now, to which table, as a seat, a spectator or an
admin, plus the bot and agent seats of every running table: the
drill-down behind the Overview's Players connected, Spectators and Bot
seats tiles ([ADR 0124](decisions/0124-admin-views-accounts-games-and-who-is-on-now.md)).
Read-only. It reads memory: the hub's sockets, then the lobby's tables,
one lock at a time and never a room's, and then, for names, one query of
the `users` table. With no database it still answers, with the names the
seats carry.

**Response 200**

```json
{
  "generated_at": 1759665600000,
  "tables": [
    {
      "id": "<uuid>",
      "name": "Friday night",
      "state": "active",
      "practice": false,
      "archived": false,
      "seats": [
        { "seat": 0, "kind": "human", "guest_name": "Alice", "deck_name": "Mono Red",
          "host": true, "connected": 2, "since": 1759665000000 },
        { "seat": 1, "kind": "human",
          "account": { "id": "<uuid>", "name": "Bobby", "avatar_url": "/avatars/<snowflake>/<hash>.png" },
          "host": false, "connected": 1, "since": 1759665100000 },
        { "seat": 2, "kind": "human", "guest_name": "Dave", "discord_pending": true,
          "host": false, "connected": 0 },
        { "seat": 3, "kind": "bot", "guest_name": "Bot 1", "bot_tier": "heuristic", "host": false, "connected": 0 }
      ],
      "spectators": [
        { "account": { "id": "<uuid>", "name": "Carol" }, "since": 1759665200000 },
        { "since": 1759665300000 }
      ],
      "admins": [
        { "as_seat": 0, "since": 1759665400000 }
      ]
    }
  ],
  "unbound_sockets": 0,
  "totals": {
    "players_connected": 2,
    "spectators": 2,
    "bot_seats": 1,
    "practice_tables": 0,
    "admin_views": 1
  }
}
```

- **`tables`** lists every table with a live socket, and every running
  table that is not archived (a bots-only table has none). Running tables
  come first, then waiting ones, then ended ones, each by name. Practice
  tables are listed, with `practice: true`.
- **A seat** is as in the games view (ADR 0124 §3.3): `kind` is `human`,
  `bot` or `agent`; `account` when a signed-in person holds it, else
  `guest_name`; `discord_pending` for a Discord seat whose person has not
  signed in again (the snowflake is never served); `bot_tier`,
  `agent_client` and `deck_name` when set; `host`. `connected` is the
  number of live sockets bound to the seat, an admin's included, and
  `since` the earliest of their connection times, absent when there is
  none. A seat's account takes its name from the seat and its avatar from
  the database.
- **`spectators`**: one entry per read-only socket. `account` is absent
  for a guest spectator; a signed-in one is named from the database.
- **`admins`**: one entry per admin socket. `account` is absent for the
  shared token; `as_seat` is the seat number an admin is bound to.
- **`unbound_sockets`** counts the sockets whose game the lobby does not
  hold. It should be 0.
- **`totals`**: `players_connected`, `spectators`, `bot_seats` and
  `practice_tables` are computed by the function the Overview's tiles are
  reported from (`metrics.Tally`), over the same copies, so they match
  `cmdctrl_seats_connected` (summed), `cmdctrl_spectators_connected`,
  `cmdctrl_seats{kind="bot"}` and `cmdctrl_practice_games` up to the time
  between a scrape and a load. Players connected counts running,
  unarchived, non-practice tables only, as the tile does; the people
  waiting at a lobby table are listed but not counted. `admin_views` is
  the number of `admins` entries listed.
- Times are Unix milliseconds. Nothing else is served: no snowflake
  except inside an `avatar_url`, no player ID, no token, no address.

**Errors**

| Status | Reason |
|---|---|
| 401 | no session |
| 403 | caller is not an admin, including an allowlisted person in player mode |
| 503 | the server wired no WebSocket hub (not a production configuration) |

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

### `GET /auth/discord/config`

Unauthenticated probe. Always registered; answers `200 {"enabled": bool}`.
`enabled` is false when `CMDCTRL_DISCORD_CLIENT_ID` / `_CLIENT_SECRET` /
`_REDIRECT_URI` are not all set, and the client hides its "Sign in with
Discord" button.

### `GET /auth/discord/start` and `GET /auth/discord/callback`

Both are registered on every deployment and both answer **503** when
Discord sign-in is not configured, rather than 404, so a misconfigured
deploy is visible. They are per-IP rate-limited.

`start` takes either `?game=<uuid>&t=<invite>` (the invite-link flow) or
no query at all (the login-page flow, which mints an identity-only
session). One of the pair without the other is a 400. It builds the
Discord authorize URL with PKCE (S256) and a `state` value, and parks the
`state`, the PKCE verifier and, in the invite flow, the game and invite in
an in-memory state store with a **5-minute TTL**. `callback` consumes the
state once, exchanges the code, reads `/users/@me` and either claims the
seat bound to the invite or mints the identity session, then redirects the
browser to the SPA's `#/oauth-complete?…` fragment.

**`prompt`** ([ADR 0110](decisions/0110-remember-me.md) §2, S55). Both
flows send Discord `prompt=none`, so a repeat sign-in with the same
`identify` scope skips Discord's screen. `?prompt=consent` on `start`
asks for the screen instead: it is the header account menu's "Sign in with a
different Discord account", because `prompt=none` silently uses
whichever account the browser is signed in to, and Discord's consent
screen has an account switcher. Any other `prompt` value is a **400**.

**The one retry.** What Discord does with `prompt=none` for someone who
has never authorized the app is undocumented. If the callback comes
back with `error=…` (for example `consent_required`,
`interaction_required` or `access_denied`) on a `prompt=none` round, it
is not shown: the callback consumes that round's state, parks a new
one for the same flow (the same game and invite, or neither) marked
`consent`, and answers **302** to Discord with `prompt=consent`. An
error on a `consent` round, the retry included, is a real refusal and
answers **400** with Discord's reason, as before. So a sign-in is
retried at most once and cannot loop. An error with a missing or
expired state is shown, since there is no flow to repeat.

### `GET /avatars/{discord_id}/{hash}`

A session is required (any role). The image is served from the server-side
avatar cache (`$CMDCTRL_DATA_DIR/avatars/<discord_id>/<hash>`), fetched
from Discord's CDN on the first miss and keyed on id plus hash, so a
changed avatar is a new key. The client appends `.png`; the server strips
it. 400 for a malformed id or hash, 503 when the cache is not configured
or `CMDCTRL_DATA_DIR` is empty. There is no `avatar_url` field anywhere in
the API: clients build this URL from `discord_id` and `discord_avatar_hash`.

### `SeatInfo` Discord fields

Each entry in a game's `players` (and the lobby's `GET /games`) may carry:

| Field | Meaning |
|---|---|
| `display_name` | The name to show. Discord global name, then Discord username, then the name the player typed. |
| `discord_id` | Discord snowflake, present only for a seat claimed through Discord. |
| `discord_avatar_hash` | Avatar hash, present only when the account has one. |

### `SeatInfo` agent fields

| Field | Meaning |
|---|---|
| `is_agent` | `true` on a seat an AI agent's MCP client claimed with `agent` in its join body ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §7). Absent on every other seat. Never cleared. |
| `agent_client` | The client's declared name, normalised (`"claude-code"`, `"codex"`, `"unknown"`). Present only with `is_agent`. |

### `GET /auth/discord/link`

Link Discord to a seat you already hold (S34 sub-PR 4, carried over
from S12.5 [#59](https://github.com/krakenhavoc/cmd_and_ctrl/issues/59)).
A navigation from the in-game menu: the server answers **302** to
Discord's consent screen, always with `prompt=consent` (ADR 0110 §2
item 4: it attaches an account to a seat, so the person sees which
account it is), and the callback comes back to the same seat. Works
in any game state. A guest who signs in mid-game becomes that seat's
user, and a seat already linked to one Discord account can be moved
to another.

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
   `user_id` and the `discord_*` fields and lasting
   `CMDCTRL_IDENTITY_TTL` (it is a sign-in), sets the cookie, and
   redirects to `/#/oauth-complete?token=…&game=…&player_id=…&user_id=…`.

**Errors** (as JSON, like the other callback errors)

| Status | Reason |
|---|---|
| 401 | no session on `/link` |
| 403 | not a player session on `/link`; on the callback, the cookie is not the seat's own session |
| 404 | the table is not live in this process |
| 409 | `?game=` names another table; the table is archived; or the Discord account already holds a different seat at this table |
| 422 | the seat is a bot, or an agent seat ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §7: it stays a guest) |
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
duplicate is worse than the 424 the reporter can act on.

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
| 424 | GitHub rejected or timed out; retry later (not 502: Cloudflare, in front of both hosts, replaces an origin 502's body with its own "error code: 502" page (#1644)) |
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

A failed CDN download is not retried server-side: it answers `424`
(Failed Dependency; not 502: Cloudflare, in front of both hosts, replaces an origin 502's body with its own "error code: 502" page (#1644)) with a JSON error body and **no** cache headers, and writes nothing to
the disk cache, so the same URL re-attempts the CDN on the next
request. The client relies on that — every card-art `<img>` retries
the plain URL once after ~2 s, then shows a click-to-retry marker
(`client/src/lib/cardArt.ts`, #33). A cache-busting query parameter
would be wrong here: the service worker keys card art on the full
query string.
