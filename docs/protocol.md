# Protocol v0

Wire protocol between the `cmd_and_ctrl` Go game server and its TypeScript
clients. Version 0 started as the minimum needed to prove the seam in S01
(`ping`/`pong` only) and was extended in S03 with `action` and `snapshot`
frame kinds plus the supporting action-type catalog and view schema. All
additions within v0 are backwards-compatible — an old ping-only client
still works against a newer v0 server.

- **Transport:** WebSocket (`ws://host:8080/ws`, `wss://` in production)
- **Framing:** one JSON object per WebSocket text frame
- **Authoritative direction:** the server. Clients propose actions; the
  server accepts, validates, and emits state deltas.
- **Version negotiation:** every frame carries a `v` field (integer). Server
  refuses any frame whose `v` does not match its supported version.

---

## Connection lifecycle (S04)

`GET /ws` requires authentication and game binding:

- **Credential** transported as a session cookie (`cmdctrl_session`, browser),
  `Authorization: Bearer <token>` header (CLI), or `?token=<token>` query
  parameter (browser WebSocket — the only transport that works cross-
  origin, since browsers cannot set Authorization headers on WS upgrade).
- **Game / player binding** via query parameters:
  - `?game=<uuid>` — the game the connection is for. A RolePlayer session
    implicitly fixes this; supplying a mismatched value returns 403.
  - `?player=<uuid>` — the viewer's seat. A RolePlayer session also fixes
    this. Omitting it for an admin connection yields a spectator view
    (every opponent hand and library rendered as hidden counts).

Upgrade errors surface as HTTP status codes before the WebSocket handshake
completes:

| Status | Meaning |
|---|---|
| 401 | missing or invalid credential |
| 403 | session not valid for the requested game / player |
| 404 | unknown game |
| 503 | server shutting down |

### Per-connection visibility (S04)

Every snapshot frame is filtered per recipient before it goes on the wire:

- The recipient's own seat carries full `hand.cards` and `library.cards`.
- Every other seat's `hand.cards` and `library.cards` are replaced with
  an empty array `[]`. The `count` field is preserved so the UI can still
  render a placeholder stack.
- Shared zones (`battlefield`, `stack`, `exile`), plus every seat's
  `graveyard` and `command` zones, are unchanged.
- Spectator connections (admin without `?player=`, or a `RoleSpectator`
  session minted via `POST /games/{id}/spectate` — S11) see all opponent
  hand cards hidden. `RoleSpectator` connections additionally have every
  inbound `action` frame rejected with `bad_request` ("spectator
  connections are read-only"). Admin-spectator connections keep their
  ability to mutate state — admins are the moderator escape hatch.

The filter runs inside the hub after `Room.Apply`'s state capture, so
all recipients see the same `seq` for a given state even though the
payloads differ.

---

## Frame envelope

Every frame is a JSON object with at least:

```json
{
  "v": 0,
  "kind": "<kind>",
  "id": "<uuid-v4>"
}
```

- `v` — protocol version. Currently `0`.
- `kind` — discriminator. Each kind has its own additional fields below.
- `id` — client-generated UUIDv4 (client → server frames) or
  server-generated (server → client frames). Used to correlate
  responses/errors to the originating action.

Unknown fields MUST be ignored by readers to allow forward-compatible
additions within a version.

---

## v0 frame kinds

### `ping` (client → server)

A no-op action used to prove the WebSocket round-trip. The server replies
with a `pong` delta carrying the same `id`.

```json
{
  "v": 0,
  "kind": "ping",
  "id": "3c2f8d7e-6b5a-4f1e-9a2c-0d8e7f6a5b4c",
  "payload": {
    "msg": "hello"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.msg` | string | no | Echoed back in the `pong`. For S01 debugging only. |

### `pong` (server → client)

Server response to a `ping`.

```json
{
  "v": 0,
  "kind": "pong",
  "id": "3c2f8d7e-6b5a-4f1e-9a2c-0d8e7f6a5b4c",
  "payload": {
    "msg": "hello",
    "server_time": "2026-04-10T17:42:31Z"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.msg` | string | no | Echoed from the originating `ping`. Omitted if the ping carried no msg. |
| `payload.server_time` | string (RFC3339) | yes | Server wall clock at the moment of reply. |

### `error` (server → client)

Any server-side failure in processing a client frame.

```json
{
  "v": 0,
  "kind": "error",
  "id": "3c2f8d7e-6b5a-4f1e-9a2c-0d8e7f6a5b4c",
  "payload": {
    "code": "bad_request",
    "message": "unknown kind 'explode'"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.code` | string | yes | Short machine-readable code. v0 codes: `bad_version`, `bad_json`, `bad_request`, `internal`. |
| `payload.message` | string | yes | Human-readable message safe to display to the user. |

`id` matches the originating client frame's `id` when possible; otherwise
empty string.

### `action` (client → server) — added in S03

A request to mutate the authoritative game state. The server validates,
applies the mutation, and broadcasts a `snapshot` frame to every connected
client on success. On failure it replies with an `error` frame addressed
to the originating client only (no broadcast).

```json
{
  "v": 0,
  "kind": "action",
  "id": "a4f7b8e1-2c3d-4e5f-9a2c-0d8e7f6a5b4c",
  "payload": {
    "type": "draw_card",
    "player": "8a2c0d8e-7f6a-5b4c-9a2c-0d8e7f6a5b4c"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.type` | string | yes | One of the action types in the table below. |
| `payload.player` | string (UUID) | conditional | See the per-action `player required` column in the catalog below. Required for most "a specific player does X" actions; omitted for actions that operate on a specific card by instance ID (`tap`, `untap`, `move_card`, `add_counter`, `set_commander_damage`) and for game-wide actions (`pass_priority`, `pass_turn`). |
| `payload.params` | object | conditional | Action-specific parameters, shape depends on `type`. See the table. |

#### Action types (v0)

All are accepted at face value — there is no rules enforcement at S03.
Rules enforcement grows incrementally in the S13+ B→C graft track (see
PLAN.md §2.1).

| `type` | `player` required | `params` shape | Effect |
|---|---|---|---|
| `draw_card` | yes | — | Moves the top of `player`'s library to their hand. S13: server-side no-op when called during the active player's `draw` step (the step-entry hook has already drawn automatically, except on the starting player's turn 1 per CR 103.7c). Outside that window the action behaves as before. |
| `play_card` | yes | `{ "instance_id": "<uuid>" }` | Moves the card from `player`'s hand to the battlefield; stamps controller to `player`. |
| `move_card` | no | `{ "src": <ZoneRef>, "dst": <ZoneRef>, "instance_id": "<uuid>" }` | General-purpose zone-to-zone move. Clears tapped state + counters if leaving the battlefield (CR 400.7). Caller must control the card. |
| `tap` | no | `{ "instance_id": "<uuid>" }` | Sets a battlefield card's tapped state to true. Caller must control the card (admin sessions bypass) — see "controller-only card actions" below. |
| `untap` | no | `{ "instance_id": "<uuid>" }` | Sets a battlefield card's tapped state to false. Same controller gate as `tap`. |
| `untap_all` | yes | — | Untaps all of `player`'s cards on the battlefield. S13: server-side no-op when called during the active player's `untap` step (the step-entry hook has already untapped automatically). Outside that window the action behaves as before — sandbox / replay support. |
| `pass_priority` | no | — | Rotates priority to the next non-eliminated seat. When priority would wrap back to the active seat, the step auto-advances and `priority_holder` resets per the new step's rules. S13: skips eliminated seats during the rotation; rejects with `bad_request` ("no player holds priority this step") on `untap` / `cleanup` (those don't grant priority per CR 502.4 / 514.3). Stack-aware semantics (priority resets on spell resolution) arrive with S13.1. |
| `pass_turn` | no | — | Jumps to the next seat's turn. Lands on Untap with `priority_holder = -1`; the step-entry hook auto-untaps and walks the cursor on to Upkeep before the next snapshot is broadcast. |
| `mulligan` | yes | `{ "hand_size": <int> }` | Shuffles `player`'s hand back into their library and draws `hand_size` cards. Simplified London mulligan — no card-to-bottom penalty. |
| `shuffle_library` | yes | — | Reshuffles `player`'s library. |
| `change_life` | yes | `{ "delta": <int> }` | Adjusts `player`'s life by `delta` (positive for gain, negative for loss). |
| `add_counter` | no | `{ "instance_id": "<uuid>", "name": "<string>", "delta": <int> }` | Modifies a named counter on a card. Delta ≤ 0 that drives the counter to zero removes the entry. Caller must control the card. |
| `set_commander_damage` | no | `{ "from": "<uuid>", "to": "<uuid>", "amount": <int> }` | Sets total commander damage dealt from `from`'s commander(s) to `to`. Set semantics, not additive. |
| `set_battlefield_position` | no | `{ "instance_id": "<uuid>", "x": <float>, "y": <float> }` | Stamps a normalised (x, y) position in `[0, 1]` on a battlefield card. Server clamps out-of-range inputs rather than erroring. Target must be on the battlefield — other zones return `card_not_found`. Caller must control the card. Added in S06; controller gate added in S08.5. |
| `concede` | yes | — | Marks the calling player as eliminated and advances the turn cursor past them if they were the active seat. When exactly one non-eliminated seat remains, the game's `state` transitions to `ended` (the survivor is the implicit winner — no separate field). Idempotent calls return `bad_request` with "player is already eliminated". Added in S08. |
| `keep_hand` | yes | — | Commits the calling player to their current opening hand during the mulligan window. Once every seated, non-eliminated player has called `keep_hand`, `mulligans_open` flips to false. Idempotent (no-op if already kept). Added in S08. |
| `declare_attacker` | no | `{ "attacker": "<uuid>", "target": "<uuid>" }` | Marks a battlefield card as attacking the target player. Step-gated to `declare_attackers`; rejects with `bad_request` outside that step. Attacker must be a creature (`type_line` contains "Creature"); rejects with `bad_request` otherwise. Re-declaring the same attacker against a different target overwrites. Caller must control the attacker (admin sessions bypass) — see "controller-only card actions" below. Added in S08; controller gate added in S08.5. |
| `declare_blocker` | no | `{ "blocker": "<uuid>", "attacker": "<uuid>" }` | Marks a battlefield card as blocking the named attacker. Step-gated to `declare_blockers`. Blocker must be a creature; attacker must exist on the battlefield. Caller must control the blocker. Added in S08; controller gate added in S08.5. |
| `clear_combat` | no | — | Resets `attacking_target` and `blocking_target` on every battlefield card. Not step-gated (escape hatch). Added in S08. |
| `advance_step` | no | — | Advances the turn cursor by one step. Sandbox affordance — does NOT require the caller to hold priority (cf. `pass_priority` which does). Used by the "next step" / "done" buttons so the active player can drive combat progression without waiting for opponent priority passes. Auto-resolves combat damage on entry to `combat_damage` (same hook as `pass_priority`). Added in S08. |
| `set_monarch` | optional | — | Designates `player` as the monarch (Conspiracy mechanic). Empty / omitted `player` clears the marker. Any seated player may flip the marker — sandbox does not enforce the must-attack-when-able or combat-damage transfer rules. Added in S10. |
| `set_initiative` | optional | — | Designates `player` as holding the initiative (BG3 mechanic). Empty / omitted `player` clears. Same posture as `set_monarch`. Added in S10. |
| `set_goaded` | no | `{ "instance_id": "<uuid>", "by": "<uuid>" }` | Marks a battlefield creature as goaded by the named player. Empty `by` clears the goad. Sandbox marker; the must-attack-not-the-goader rule is not enforced. Added in S10. |
| `set_poison` | yes | `{ "amount": <int> }` | Sets `player`'s poison counter total. Set semantics (not increment) so two stale tabs don't double-count. Clamped at 0 from below. Added in S10. |
| `set_energy` | yes | `{ "amount": <int> }` | Sets `player`'s energy counter total. Same shape and clamp posture as `set_poison`. Added in S10. |
| `set_promise` | no | `{ "from": "<uuid>", "to": "<uuid>", "count": <int> }` | Sets the per-pair "I owe you" promise-token count from `from` → `to`. Set semantics; clamped at 0. Politics scaffold — visual reminder only. Added in S10. |
| `start_vote` | yes | `{ "topic": "<string>", "options": ["<string>", …] }` | Opens a vote with `player` as initiator. At least 2 options required. Rejects with `bad_request` if a vote is already open (clients must `end_vote` first). Added in S10. |
| `cast_vote` | yes | `{ "option": <int> }` | Records / overwrites `player`'s ballot at the given option index. Re-voting overwrites. Added in S10. |
| `end_vote` | no | — | Closes the currently open vote. Any seated caller may end any vote (sandbox). Added in S10. |
| `undo` | no | — | Pops the room's most recent pre-mutation snapshot off its undo stack and restores the game state. Bumps `seq` like a normal action so clients see a regular `snapshot` frame. Authorization: a seated caller may only pop an entry whose stored caller matches their own seat (rewinding an opponent's move requires their cooperation — they undo first). Each successful seated undo debits the caller's `Player.UndosRemaining` (refreshed to `Game.UndoLimit` on entering their untap step). Admin (`?player=` omitted) bypasses both the caller and budget gates. Stack capped at 32 per room. No redo at v1. Errors: `bad_request` for empty stack ("nothing to undo"), cross-player request ("you can only undo your own most recent action"), or exhausted budget ("no undos remaining this turn"). Added in S11. |
| `set_undo_limit` | no | `{ "limit": <int> }` | Sets `Game.UndoLimit` (per-player per-turn undo budget) and immediately refreshes every seat's `UndosRemaining` to the new value. Default at game start is 1. Sandbox — any seated player or admin may change it. Negative values clamp to 0 (effectively disables undo). Added in S11. |

`<ZoneRef>` is `{ "kind": "<zone_kind>", "owner": "<uuid>" }`. Owner is
omitted for shared zones (`battlefield`, `stack`, `exile`). Zone kinds are
`library`, `hand`, `battlefield`, `graveyard`, `exile`, `command`, `stack`.

#### Controller-only card actions (added in S08.5)

`tap`, `untap`, `move_card`, `add_counter`, `set_battlefield_position`,
`declare_attacker`, and `declare_blocker` are gated on the caller being
the current controller of the named card. A seated player issuing one
of these actions against a card controlled by a different seat receives
an `error` frame with `payload.code = "bad_request"` and
`payload.message = "you do not control that card"`. Admin sessions
(`role: "admin"`) bypass the gate so a moderator can fix wedged board
state. `clear_combat` intentionally stays loose — it's the "combat is
wedged" escape hatch and any seated player may invoke it.

This was deliberately permissive at S08 to support casual cross-seat
play; S08.5 wave 1 tightened it after a 2-player playtest surfaced
players accidentally tapping each other's permanents.

### `chat` (both directions) — added in S07

A chat message addressed to every client bound to the same game.
Bypasses the action / snapshot pipeline entirely: chat does not
mutate game state, does not bump the snapshot `seq`, and is not
included in crash-recovery dumps. Reconnecting clients receive an
empty chat history; persistent chat arrives with the S11 replay log.

**Client → server:** the client supplies only `payload.text` (trimmed,
non-empty, ≤ 1000 characters). The server **discards** any client-
supplied `author_id`, `author_name`, or `timestamp` and re-stamps all
three from the connection's authenticated principal before broadcast.

**Server → client:** broadcast to every client bound to the same game
(including the originator, so the local UI surfaces the canonical
server-stamped message rather than echoing the unstamped local copy).
Spectator / admin connections without a bound `playerID` chat under
the synthetic name `"spectator"` and an empty `author_id` — the
client distinguishes spectator messages by the missing `author_id`.

```json
{
  "v": 0,
  "kind": "chat",
  "id": "c1d2e3f4-...",
  "payload": {
    "author_id": "8a2c0d8e-7f6a-5b4c-9a2c-0d8e7f6a5b4c",
    "author_name": "Alice",
    "text": "any responses?",
    "timestamp": "2026-04-18T14:32:11Z"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.text` | string | yes | 1 – 1000 characters after trimming. Empty or oversized payloads are rejected with `bad_request`. |
| `payload.author_id` | string (UUID) | no | Server-stamped from the connection's player principal. Empty for spectator connections. Ignored on inbound frames. |
| `payload.author_name` | string | yes (server) | Server-stamped from the seated player's name (or `"spectator"`). Ignored on inbound frames. |
| `payload.timestamp` | string (RFC3339) | yes (server) | Server wall clock at broadcast time. Ignored on inbound frames. |

The originating client receives the server-stamped frame with the
same `id` it sent. Other recipients see the same `id` — clients can
use this for deduplication if they ever render a local optimistic
copy before the broadcast lands.

### `snapshot` (server → client) — added in S03

A full authoritative view of the game state. The server emits a `snapshot`
on initial connect (to the joining client only) and after every successful
action (broadcast to all connected clients). There is no incremental delta
format at v0 — bandwidth is trivial at ≤4 players and the full-snapshot
design is much simpler to reason about. A future version may introduce
diffs; new frame kinds can be added inside v0 additively.

```json
{
  "v": 0,
  "kind": "snapshot",
  "id": "",
  "payload": {
    "seq": 42,
    "game": { "id": "...", "state": "active", "seats": [ ... ], "turn": { ... }, "battlefield": { ... }, "stack": { ... }, "exile": { ... } }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.seq` | uint64 | yes | Per-room monotonically **non-decreasing** sequence number. Clients use it to detect dropped or out-of-order frames. `seq` is bumped only on successful `action` dispatches (see `Room.Apply` in the server); the initial snapshot a joining client receives reuses the current (not-yet-bumped) value, so a new joiner during an action race may briefly see two consecutive snapshots with identical `seq` — both carrying the same state. Clients must treat snapshots idempotently: duplicate `seq` always means "same state, re-apply is a no-op". |
| `payload.game` | object | yes | Complete `GameView`. See the schema below. |

The `id` field on a snapshot is always empty — snapshots are not correlated
to any particular client request.

#### GameView schema (summary)

The server's Go source at `server/internal/protocol/view.go` is the
canonical type definition. High-level shape:

- **GameView**: `{ id, state, seats[], battlefield, stack, exile, turn, mulligans_open, starting_seat }` — `mulligans_open` is true between `Start` and the moment all seated, non-eliminated players have called `keep_hand` (S08). `starting_seat` is the seat index that took the first turn; used by the server to enforce the CR 103.7c turn-1 skip-draw rule, surfaced on the wire so spectators / reconnects render the same first-player decision. Pre-S13 snapshots decode as `0`, which matches the only seat games started on before this field existed (S13).
- **PlayerView**: `{ id, name, seat, life, poison?, energy?, library, hand, graveyard, command, commander_damage, life_history, eliminated?, hand_kept?, mulligans_taken? }` — `life_history` is a per-player rolling log of `LifeChangeView` entries (`{ delta, new_total, at }`, RFC3339 timestamp), bounded server-side at 50 entries and never filtered (life is public). `eliminated` is omitted unless true; set when the player has conceded (S08) or, in future S13+ work, has lost to a state-based action. `hand_kept` is true once the player has committed via `keep_hand`; `mulligans_taken` counts mulligans in the current opening-hand window. Both omitted when false/zero. All five fields added in S08.
- **ZoneView**: `{ kind, owner?, count, cards[] }` — `owner` omitted for shared zones
- **CardView**: `{ instance_id, name, owner, controller, scryfall_id?, type_line?, power?, toughness?, tapped?, counters?, is_commander?, battle_x?, battle_y?, attacking_target?, blocking_target? }` — `scryfall_id` is stamped at deck-import time and lets the client resolve images via `GET /cards/{id}/image`. Omitted for placeholder cards (demo game seeded via `CMDCTRL_SEED_DEMO`). `type_line` (e.g. "Legendary Creature — Human Wizard"), `power`, and `toughness` carry Scryfall data the client uses to filter creature-only UIs and label combat rows; all three omitted for placeholder cards. `battle_x` / `battle_y` are normalised positions in `[0, 1]` for cards on the battlefield (S06+); both are cleared on zone exit and omitted for cards that have never been positioned. `attacking_target` (player UUID) and `blocking_target` (attacker instance UUID) are set by `declare_attacker` / `declare_blocker` and cleared on zone exit and by `clear_combat`. Both omitted when not set. `type_line`, `power`, `toughness`, `attacking_target`, `blocking_target` added in S08.
- **TurnView**: `{ number, active_seat, priority_holder, phase, step }` — `priority_holder` is the seat index (0-based) that currently holds priority within the step, OR `-1` (the `NoPriority` sentinel) during steps that don't grant priority. S13 lands the cursor with `priority_holder = -1` for `untap` and `cleanup` (CR 502.4 / 514.3); auto turn-based actions (auto-untap, auto-draw, auto-advance through cleanup) fire from the step-entry hook so the cursor never sits idle on a no-priority step. Added in S07; sentinel added in S13.

S03 does not yet apply visibility filtering — every client receives every
card's contents, including opponents' hands. Per-connection visibility is
an S04 concern (ships alongside authentication).

Clients should treat every received snapshot as the new authoritative
state; no partial merge is needed.

---

## Connection lifecycle (v0)

1. Client opens WebSocket to `/ws`.
2. Server accepts, optionally logs client IP.
3. Client and server exchange frames. No authentication at v0 (S04 adds it).
4. Either side may close at any time. Server closes with a standard close
   code; client handles reconnection itself.

---

## Out of scope for v0

Deferred to later sprints, typically requiring a new frame kind but no `v`
bump unless they turn out to be wire-breaking:

- **Lobby and game creation** (S04) — per-connection auth, create/join
  rooms, list active games.
- **Visibility filtering** (S04) — opponent hands should show counts, not
  contents. At S03 every client sees every card face-up.
- **Incremental deltas** — S03's choice is full-snapshot broadcast. Deltas
  can arrive as a new frame kind without breaking v0.
- **Reconnection and session resume** — crash-recovery dumps let a
  restarted server reload state, but live client sessions drop on
  disconnect and must be re-established.
- **Spectator mode** (S11) — read-only connections.

### Replay log (S11)

Every successful `Apply` on a Room appends one JSONL line to
`$CMDCTRL_DATA_DIR/replays/<game-id>.jsonl`. Each line is a full
`SnapshotPayload` (the same shape broadcast over WS after an action),
in the order the server produced them. Consumers can stream the file
and re-derive the game timeline.

`GET /games/{id}/replay` returns the log as `application/x-ndjson`.
Authorization: admin or any seat (player / spectator) in this game.
Returns 204 when no Apply has fired yet (empty replay). The
crash-recovery dump (single `<game-id>.json` file holding the
latest snapshot) is orthogonal — it exists for server restarts;
the replay log is the additive history.

---

## Schema evolution rules

- **Breaking changes** bump `v` and require updating both server and client
  in lockstep. Keep breaking changes rare.
- **Additive changes within a version** (new optional fields, new error
  codes) are allowed and do not bump `v`. Readers ignore unknown fields.
- **Removing a field** is always breaking. Don't do it without a `v` bump.
- Each version of this document lives at `docs/protocol.md`; previous
  versions are preserved in git history and can be retrieved by commit hash
  when diagnosing issues with old deployments.

---

## References

- RFC 6455 — The WebSocket Protocol
- [ADR 0001 — WebSocket library](decisions/0001-ws-library.md)
- Sprint S03 — [action protocol + state deltas](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3)
- `server/internal/protocol/` — Go source of truth for wire types
- `server/internal/actions/` — action type catalog and dispatch
