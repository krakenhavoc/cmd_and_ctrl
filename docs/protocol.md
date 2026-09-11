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
| `payload.code` | string | yes | Short machine-readable code. v0 codes: `bad_version`, `bad_json`, `bad_request`, `internal`. S15 adds `insufficient_mana` — see below. |
| `payload.message` | string | yes | Human-readable message safe to display to the user. |
| `payload.missing` | string[] | no | Present only when `code == "insufficient_mana"` (S15). List of mana symbols the caller's pool could not cover, in the order they appear in the printed cost (e.g. `["{R}", "{1}"]`). Client renders these verbatim into the override toast. |
| `payload.card_id` | string (UUID) | no | Present only when `code == "insufficient_mana"` (S15). Instance ID of the card whose cast was rejected. Lets the client's "Cast anyway" / "Auto-tap & cast" buttons re-fire the same cast without needing to round-trip through the user's last click. |

`id` matches the originating client frame's `id` when possible; otherwise
empty string.

#### `insufficient_mana` (S15)

Emitted only when a `cast_spell` arrives with `strict: true` and the
caller's mana pool cannot cover the effective cost (printed cost +
commander tax for command-zone casts). The `missing` list carries the
remaining unpaid symbols; the client's error-toast override surfaces
them and offers two buttons:

1. **Cast anyway** — re-fire the same `cast_spell` with
   `force_cast: true`, which bypasses the gate without deducting
   any mana (sandbox posture — the player accepts paper tracking
   for this one cast).
2. **Auto-tap & cast** — fetch
   `GET /games/:id/auto-tap-preview?card=<id>`, render the plan in
   `AutoTapPreviewModal`, and on confirm re-fire the cast with
   `auto_tap: true` (atomic server-side tap-and-cast, see below).

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

The client mirrors the accepted verbs as a string-literal `ActionType`
union in `client/src/lib/protocol.ts`, so a mistyped action name is a
compile error rather than a runtime `bad_request`. The server registry
(`server/internal/actions/actions.go` plus the hub-level `undo` verb)
remains the source of truth; the union is a client-side convenience and
does **not** change the wire schema — add literals as the UI grows call
sites for verbs the server already accepts.

| `type` | `player` required | `params` shape | Effect |
|---|---|---|---|
| `draw_card` | yes | — | Moves the top of `player`'s library to their hand. S13: server-side no-op when called during the active player's `draw` step (the step-entry hook has already drawn automatically, except on the starting player's turn 1 per CR 103.7c). Outside that window the action behaves as before. |
| `play_card` | yes | `{ "instance_id": "<uuid>" }` | Moves the card from `player`'s hand to the battlefield; stamps controller to `player`. |
| `move_card` | no | `{ "src": <ZoneRef>, "dst": <ZoneRef>, "instance_id": "<uuid>", "as_commander"?: bool, "to_bottom"?: bool }` | General-purpose zone-to-zone move. Clears tapped state + counters if leaving the battlefield (CR 400.7). Caller must control the card. S13.1: when `as_commander` is true AND the card is a commander AND `dst` is graveyard / exile / hand / library, the destination is rewritten to the owner's command zone (CR 903.9 commander zone replacement, exercised as an explicit player choice rather than an automatic engine transform). #170: when `to_bottom` is true the card is seated at the BOTTOM of the destination zone instead of the top — only meaningful for a library destination, and a no-op if a replacement rewrote the destination. |
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
| `declare_attackers` | no | `{ "attackers": [{ "attacker": "<uuid>", "target": "<uuid>" }, ...] }` | Declares a whole attacking set in ONE action — the wire verb behind the client's "attack with all" cluster (#318). Step-gated to `declare_attackers`. Each entry is checked independently and **skipped silently** when the creature is unknown, is not a creature, is tapped, is summoning-sick (CR 302.1), has `defender` (CR 702.3), is already declared this combat, or names a seat that is the creature's own controller / eliminated / nonexistent — one ineligible creature must never sink a wide declaration. Rejects with `bad_request` when the batch is empty, exceeds 256 entries, contains a malformed UUID, or when EVERY entry was skipped. Caller must control every listed creature; a foreign creature rejects the whole batch (authorization is not part of the skip contract). Declared creatures tap unless they have `vigilance` (CR 508.1f / 702.20), emit one `EventAttack` each, and drain through a single state-check pass so simultaneous attack triggers reach the stack together (CR 508.1 / 508.2). Prefer this over looping `declare_attacker`: each action is one undo entry and one broadcast snapshot, so N creatures would otherwise cost N of both against a per-turn undo budget of 1. Added in S31.
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
| `cast_spell` | yes | `{ "instance_id": "<uuid>", "from_zone"?: "hand" \| "command", "targets"?: [...], "modes"?: [int...], "x_value"?: int, "distribution"?: { "<uuid>": int }, "hold_priority"?: bool, "split_second"?: bool, "strict"?: bool, "force_cast"?: bool, "auto_tap"?: bool, "locked_sources"?: ["<uuid>", ...] }` | S13.1 canonical "play a card from hand" verb (CR 601). Lands route to battlefield (CR 305 special action); other types go to the stack with a fresh metadata entry, the caster retains priority (CR 117.3c). Sorcery-speed gate (sorceries + lands) requires main phase + stack empty + caller is active. Casting from `command` increments the per-commander tax counter (CR 903.8). Caller must hold priority. Targets are `{ "kind": "player" \| "card" \| "self" \| "none", "id"?: "<uuid>" }`. S14: if the card is in the effect catalog with a declared `CardView.target_mode`, the client enters a targeting prompt before sending `cast_spell`; the server re-validates at resolve per CR 608.2b ("countered by game rules" fizzle if every targeted slot is illegal at resolution time). **S15:** `strict` engages the mana-cost gate — server parses `ManaCost` (plus commander tax), rejects with `insufficient_mana` error frame when the pool falls short, deducts on success. Sourced from the client's `gameplay.strictMana` setting. `force_cast` overrides the gate for one cast (the "Cast anyway" toast button) — proceeds permissively without touching the pool. `auto_tap` runs the auto-tapper before the gate, taps the planned permanents, and drops produced mana into the pool — all atomic under one write lock. `locked_sources` excludes specific permanents from the auto-tapper's plan (the modal's lock-tap UI). Plan failure under `auto_tap: true` returns `insufficient_mana` before any cards tap (all-or-nothing). |
| `counter_spell` | no | `{ "instance_id": "<uuid>", "to_zone"?: <ZoneRef> }` | Removes a spell item from the stack and routes its card to `to_zone` (default = owner's graveyard). Covers Counterspell (default), Hinder (`to_zone = library`), Remand (`to_zone = hand`), exile-bound counters. Battlefield + stack rejected as destinations (`bad_request: invalid stack-counter destination`). Caller must hold priority. Added in S13.1. |
| `counter_ability` | no | `{ "instance_id": "<uuid>" }` | Removes an activated / triggered ability item from the stack. Abilities cease to exist on removal (CR 608.2m); no destination needed. Caller must hold priority. Added in S13.1. |
| `activate_ability` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"? }` | Creates an activated-ability stack item linked to the source card. Same announce-time params as `cast_spell` (minus the source-zone routing). The source card stays in its origin zone; the stack item carries a synthetic ID. Mana abilities are NOT modeled this way (CR 605 — they don't use the stack). Caller must hold priority. Added in S13.1. |
| `activate_loyalty` | yes | `{ "planeswalker_id": "<uuid>", "label"?: string, "delta": int }` | Applies a planeswalker loyalty ability. Sandbox shape: the engine doesn't model the activation as a proper stack item (deferred to S14+) — `delta` is applied immediately to the planeswalker's `loyalty` counter. Sorcery-speed gated; once-per-turn-per-planeswalker (CR 606.5). Caller must hold priority. Added in S13.1. |
| `announce_trigger` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"? }` | Queues a triggered ability for APNAP-ordered drain onto the stack (CR 603.3b). The drain happens at every priority-grant boundary inside the SBA + trigger loop. Sandbox: the player whose card has the trigger announces it manually — auto-fire from card events is the bright line between S13.1 and S14+. Added in S13.1. |
| `mark_damage` | no | `{ "instance_id": "<uuid>", "delta": int }` | Adjusts the damage noted on a creature on the battlefield by `delta` (positive to add, negative to remove). Drives the lethal-damage SBA (CR 704.5g). Positive deltas require controller (admins / spectators bypass); negative deltas are caller-loose so opponents can undo. SBA loop fires after the mutation. Cleanup-step turn-based action zeros every creature's marked damage (CR 514.2). Added in S13.1. |
| `add_player_counter` | yes | `{ "name": string, "delta": int }` | Modifies a named player-level counter (poison, energy, experience, rad, plus homebrew names) by `delta`. Negative deltas clamp at zero. Drives the SBA loop after the mutation so 10 poison or future counter-driven losses fire immediately. The legacy `set_poison` / `set_energy` actions stay synchronised with the underlying `Player.Counters` map. Player-scoped: a seated caller may only adjust their own counters; admins bypass. Added in S13.2. |
| `discard_selection` | yes | `{ "card_ids": ["<uuid>", ...] }` | Resolves the cleanup-step interactive discard prompt (S13.4, CR 402.2). Caller must be in `discard_pending`; the count must exactly match the over-max amount; every supplied card ID must live in the caller's hand. Selected cards move to the caller's graveyard, the caller's pending entry clears, and the cleanup hook re-fires so the cursor resumes its auto-advance. Idempotent no-op for callers not in the pending map. Player-scoped. Added in S13.4. |
| `set_max_hand_size` | yes | `{ "value": int }` | Sets the named player's per-player cleanup-step hand-size cap (CR 402.2). Default 7; `-1` disables the cap (Reliquary Tower / Thought Vessel). Sandbox helper until the S14+ effect catalog wires this to real cards via the S16 layer pipeline. Player-scoped. Added in S13.4. |
| `resolve_choice` | yes | `{ "choice_id": "<uuid>", "card_ids"?: ["<uuid>", ...], "color"?: "W" \| "U" \| "B" \| "R" \| "G" \| "C", "apply"?: bool, "order"?: ["<id>", ...], "assignments"?: [{ "blocker_id": "<uuid>", "amount": int }, ...], "trample_to_player"?: int }` | Resolves a pending async effect-driven choice (S14, CR 701.8a — "player chooses" discard). Caller must be the `chooser` of the named entry in `GameView.pending_choices`; `card_ids.length` must equal the entry's `count`; every supplied card ID must currently live in the entry's `from_player`'s hand. Picked cards move to `from_player`'s graveyard (for `discard_from_hand`), the entry drains from the queue, and an `EventDiscardCard` event fires per card. Distinct from `discard_selection` (S13.4 cleanup, chooser == discarder); `resolve_choice` covers effects where chooser != discarder (Thoughtseize — caster picks from target's revealed hand). **S15:** the entry's `kind` may be `mana_pick` (Birds of Paradise / Arcane Signet); `card_ids` is omitted and `color` carries the chosen single-letter color, validated against the entry's `color_options` slice. The picked color drops into the chooser's mana pool as one `ManaToken`. **S17:** `kind` may be `replacement_order` (CR 616 multi-replacement order prompt — `order` carries the chosen permutation of replacement-effect IDs) or `optional_replacement` (CR 614.10 yes/no — `apply` is the answer). **S18:** `kind` may be `damage_assignment` (CR 510.1c multi-blocker assignment — `assignments` is the per-blocker amounts in chosen order, `trample_to_player` is the overflow when the attacker has trample). **S19:** `kind` may be `trigger_prompt` (CR 603.4 "you may" optional triggered ability — `apply` is the answer; the server routes by inspecting the choice's kind, since `optional_replacement` and `trigger_prompt` share the `apply` payload). Added in S14. |
| `activate_mana_ability` | yes | `{ "instance_id": "<uuid>", "ability_idx"?: int }` | S15. Activates the `ability_idx`-th mana ability on a battlefield permanent the caller controls. `ability_idx` defaults to 0 (the only ability for basic lands and most rocks). Single code path for catalog-declared abilities (Sol Ring, Arcane Signet, Birds of Paradise — looked up via the `CatalogManaAbilities` hook) and synthetic basic-land abilities (Forest → `{G}`) the engine derives from `TypeLine`. Pays the tap cost (rejects with `bad_request: "card already tapped"` if the source is tapped), then materialises the produced mana: single-color slots drop straight into the controller's pool as one `ManaToken`; multi-option slots (pipe syntax) queue a `mana_pick` PendingChoice for the controller. Mana abilities don't use the stack (CR 605.3) — synchronous under the write lock. Caller need NOT hold priority (mana abilities are special, CR 605.1a). Player-scoped: caller must control the source. |

`<ZoneRef>` is `{ "kind": "<zone_kind>", "owner": "<uuid>" }`. Owner is
omitted for shared zones (`battlefield`, `stack`, `exile`). Zone kinds are
`library`, `hand`, `battlefield`, `graveyard`, `exile`, `command`, `stack`.

#### Controller-only card actions (added in S08.5)

`tap`, `untap`, `move_card`, `add_counter`, `set_battlefield_position`,
`declare_attacker`, `declare_attackers`, and `declare_blocker` are gated
on the caller being the current controller of the named card.
(`declare_attackers` applies the gate to every entry in the batch, and
one foreign creature rejects the whole batch.) A seated player issuing one
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

- **GameView**: `{ id, state, seats[], battlefield, stack, exile, turn, mulligans_open, starting_seat, stack_items?, pending_triggers?, delayed_triggers?, split_second_active?, events?, pending_choices? }` — `mulligans_open` is true between `Start` and the moment all seated, non-eliminated players have called `keep_hand` (S08). `starting_seat` is the seat index that took the first turn; used by the server to enforce the CR 103.7c turn-1 skip-draw rule, surfaced on the wire so spectators / reconnects render the same first-player decision. Pre-S13 snapshots decode as `0` (S13). `stack_items` is the announce-time metadata for every item on the stack (S13.1) — see StackItemView below. `pending_triggers` is the APNAP queue of triggered abilities waiting to drain (CR 603.3b). `delayed_triggers` (S22, omitempty) is the queue of CR 603.7 delayed triggered abilities still owed — "at the beginning of the next end step, return that card to the battlefield" — as `{ id, controller, source?, label?, at, created_turn?, cards? }`, where `at` is the step whose beginning fires it and `cards` are the instance IDs the effect acts on. Public information, so it is not redacted per viewer. `split_second_active` mirrors the engine's "no responses allowed" gate (CR 702.79). `events` (S14, omitempty) is a rolling window of the authoritative per-game event log the client uses for auto-resolve toasts — pre-S14 snapshots decode as empty. `pending_choices` (S14, omitempty) is the async effect-driven decision queue — see PendingChoiceView below.
- **PlayerView**: `{ id, name, seat, life, poison?, energy?, library, hand, graveyard, command, commander_damage, life_history, eliminated?, hand_kept?, mulligans_taken?, mana_pool? }` — `mana_pool` (S15, omitempty) is the player's typed mana pool as an ordered list of single-character color strings (`"W"`, `"U"`, `"B"`, `"R"`, `"G"`, `"C"`) in tap order. Drives the `ManaPoolPips` row in the player header. Server-side empties at every step boundary (CR 106.4) so the wire surface stays small. Public to all viewers (matches paper Magic — floating mana sits visibly next to the player). `life_history` is a per-player rolling log of `LifeChangeView` entries (`{ delta, new_total, at }`, RFC3339 timestamp), bounded server-side at 50 entries and never filtered (life is public). `eliminated` is omitted unless true; set when the player has conceded (S08) or, in future S13+ work, has lost to a state-based action. `hand_kept` is true once the player has committed via `keep_hand`; `mulligans_taken` counts mulligans in the current opening-hand window. Both omitted when false/zero. All five fields added in S08. `is_bot` / `bot_tier` (S31, both omitempty) mark a seat driven by a server-side bot runner and name its policy tier (`"random"` today); the client renders a BOT chip in place of the avatar. A bot's moves arrive as ordinary snapshot frames — there is no bot-specific frame kind.
- **ZoneView**: `{ kind, owner?, count, cards[] }` — `owner` omitted for shared zones
- **CardView**: `{ instance_id, name, owner, controller, scryfall_id?, oracle_id?, type_line?, power?, toughness?, tapped?, counters?, is_commander?, battle_x?, battle_y?, attacking_target?, blocking_target?, auto?, target_mode?, mana_cost?, produced_mana?, mana_abilities?, abilities? }` — **S16:** `power`, `toughness`, `type_line`, and the new `abilities` field carry the *effective* (post-CR-613-layer-resolution) characteristics, not printed. The wire keys are unchanged; only the resolution semantics are. Cards with no static ability affecting them produce identical wire output to pre-S16 (printed == effective). When an anthem is in play, the controlled creature's `power` / `toughness` arrive +1; when Mycosynth Lattice is in play, every permanent's `type_line` includes `"Artifact"`; when Lord of Atlantis is in play, other Merfolk creatures' `abilities` includes `"islandwalk"`. `abilities` (S16, omitempty) is a `[]string` of keyword names ("flying", "first strike", "trample", "islandwalk", etc.) — drives S18's keyword renderer; behavior of the keywords lands with S18 (combat sub-step rewrite). Non-protocol-bumping change: pre-S16 clients ignore the unknown `abilities` field harmlessly. `mana_cost` (S15, omitempty) is the card's printed cost string in Scryfall format (`"{2}{R}{R}"`, `"{X}{G}"`, `"{W/U}{W/U}"`); the parser accepts generic, colored (`{W|U|B|R|G|C}`), variable (`{X}`), hybrid two-color (`{W/U}`), hybrid generic (`{N/W}`), phyrexian (`{W/P}`), and snow (`{S}`) symbols. Drives the cost chip on the cast modal and the strict-mode gate. Empty / omitted ⇒ treated as costless (for cards Scryfall ingestion couldn't parse, which the cost gate emits as an `EventCostWarning` rather than rejecting). Redacted to empty for cards the viewer is not a `known_by` of. `produced_mana` (S15, omitempty) is the parsed Scryfall `produced_mana` field (e.g. `"{C}{C}"` for Sol Ring, `"{W|U|B|R|G}"` for Birds of Paradise) — informational; the canonical activation surface is `mana_abilities`. `mana_abilities` (S15, omitempty) is the per-ability projection the client renders into the right-click "tap for mana" menu: each entry is `{ "label": string, "produced": string, "cost_tap"?: bool, "cost_sacrifice"?: bool }`. Synthetic basic-land abilities are rendered into this list when the catalog has nothing registered (Forest → `[{"label": "Add {G}", "produced": "{G}", "cost_tap": true}]`). `scryfall_id` is stamped at deck-import time and lets the client resolve images via `GET /cards/{id}/image`. Omitted for placeholder cards (demo game seeded via `CMDCTRL_SEED_DEMO`). `oracle_id` (S14) is the Scryfall oracle-level identifier (stable across printings) and is the catalog lookup key — every printing of Lightning Bolt shares one `oracle_id`. `type_line` (e.g. "Legendary Creature — Human Wizard"), `power`, and `toughness` carry Scryfall data the client uses to filter creature-only UIs and label combat rows; all three omitted for placeholder cards. `battle_x` / `battle_y` are normalised positions in `[0, 1]` for cards on the battlefield (S06+); both are cleared on zone exit and omitted for cards that have never been positioned. `attacking_target` (player UUID) and `blocking_target` (attacker instance UUID) are set by `declare_attacker` / `declare_blocker` and cleared on zone exit and by `clear_combat`. Both omitted when not set. `type_line`, `power`, `toughness`, `attacking_target`, `blocking_target` added in S08. `attached_to` (S24, omitempty) is the CR 301.5c / CR 303.4 attachment relation for an Equipment or an Aura — a `TargetRefView` (`{ kind, id }`) naming the permanent (`kind: "card"`) or player (`kind: "player"`) this card is attached to. Omitted for every card attached to nothing. Shipped in ONE direction only: "what is attached to this creature" is derived client-side by partitioning the battlefield on this field, so the two directions cannot disagree. Not redacted — attachment is public battlefield state exactly like `attacking_target`. See [ADR 0036](decisions/0036-attachments.md). `auto` (S14, omitempty) is true when the card's `oracle_id` is registered in the catalog (drives the gold-leaf AUTO badge). `target_mode` (S14, omitempty) is the announce-time target-prompt shape the client should show when casting: `"any"`, `"player"`, `"creature"`, `"stack_spell"`, `"card_in_graveyard"`. Empty ⇒ no prompt (cast fires immediately). Both fields redacted to zero for cards the viewer is not a `known_by` of.
- **TurnView**: `{ number, active_seat, priority_holder, phase, step }` — `priority_holder` is the seat index (0-based) that currently holds priority within the step, OR `-1` (the `NoPriority` sentinel) during steps that don't grant priority. S13 lands the cursor with `priority_holder = -1` for `untap` and `cleanup` (CR 502.4 / 514.3); auto turn-based actions (auto-untap, auto-draw, auto-advance through cleanup) fire from the step-entry hook so the cursor never sits idle on a no-priority step. Added in S07; sentinel added in S13.
- **StackItemView** (S13.1): `{ id, kind, controller, owner, source_card_id, label?, targets?, modes?, x_value?, distribution?, hold_priority?, split_second? }` — announce-time metadata for one item on the stack. `kind` is `"spell"` (card lives in `stack` zone), `"activated"` or `"triggered"` (no card; source stays in its origin zone). `targets` is a list of `TargetRefView` slots: `{ kind: "player" | "card" | "self" | "none", id?: "<uuid>" }`. Resolution re-checks targets per CR 608.2b — if every targeted slot is illegal, the spell is "countered by game rules" and routes to its owner's graveyard. `PlayerView.commander_casts` (S13.1, omitempty) is the per-commander cast tax counter map keyed by commander UUID string. `CardView.damage_marked` (omitempty) drives the lethal-damage SBA.
- **CardView.known_by_you / face_down** (S13.5): per-viewer card visibility. `known_by_you` is true when the viewer is in the server-side KnownBy set for that card; the wire surfaces full printed characteristics in that case. When false, the wire zeroes name / type_line / scryfall_id / power / toughness / counters / is_commander but keeps instance_id / owner / controller / tapped / position / damage_marked / face_down — the engine state stays observable, but identity doesn't leak. `face_down` is the visual flip flag (CR 708 — morph / manifest); a face-down creature renders as a back to everyone, but knowers can still hover-reveal the printed characteristics. Library cards have no knowers post-shuffle; opening-hand cards are known to their owner only; battlefield / stack / exile / graveyard / command-zone cards are public (all seated players are knowers). `ShuffleLibrary` clears every library card's KnownBy; `Mulligan` clears hand + library and re-grants the owner on the new opening hand. S14's `keepKnownInHandZone` view-filter path preserves revealed opponent-hand cards end-to-end: a seated viewer receives opponent hand zones containing only their `KnownByYou == true` entries (the `Count` stays accurate so the client can render the hidden remainder as face-down placeholders); spectator / admin connections still see a fully-hidden opponent hand via `hideZoneContents`.
- **EventView** (S14, omitempty): `{ kind, actor?, source?, target?, amount?, card_id?, old_zone?, new_zone?, error?, seq }` — one entry in the authoritative per-game event log. `kind` is one of `cast`, `resolve`, `fizzle`, `deal_damage`, `change_life`, `draw_card`, `discard_card`, `mill`, `zone_move`, `tap_card`, `untap_card`, `counter_placed`, `token_created`, `search_library`, `counter_spell`, `concede`, `etb`, `ltb`, `effect_error`. **S15 adds:** `mana_added` (one per single-color slot dropped into a pool), `mana_spent` (one per successful strict-mode cast), `mana_ability_activated` (one per `activate_mana_ability` call), `cost_warning` (permissive cast / unparseable cost / `force_cast` override). **S16:** `etb` and `ltb` are now consumed by the layer-engine listener for invalidation (battlefield zone changes bump `Game.layerVersion`); `counter_placed` likewise bumps. No new event kinds in S16 — the engine listens to existing ones. `seq` is monotonic within a game; `actor` / `source` / `target` / `card_id` are UUID strings keyed to the seated player / source card / target card. Rolling window on the wire (server-side cap) — pre-S14 replay snapshots decode as empty.
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.4 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder).

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

### Auto-tap preview (S15)

`GET /games/{id}/auto-tap-preview?card=<uuid>&x=<int>&exclude=<uuid>,<uuid>...`
returns a read-only projection of what the auto-tapper would do for
a given cast — the client uses it to render the `AutoTapPreviewModal`
before committing a `cast_spell` with `auto_tap: true`. No game state
mutates. Authorization: any seated player at the named game; spectators
and non-seated callers receive 403.

| Query param | Required | Notes |
|---|---|---|
| `card` | yes | Instance UUID of the card to plan for. The server reads its `ManaCost` (plus commander tax for command-zone casts) to derive the effective cost. |
| `x` | no | Caller-supplied X value for spells with `{X}` in their cost. Defaults to 0. |
| `exclude` | no | Comma-separated permanent UUIDs the auto-tapper must NOT consider — the lock-tap UI's reservation list. |

Response shape:

```json
{
  "ok": true,
  "plan": ["<uuid>", "<uuid>"],
  "missing": null,
  "cost": "{2}{R}{R}"
}
```

`ok` is true when the auto-tapper found a satisfying plan; `plan` is
the ordered list of permanent IDs to tap. When `ok` is false, `plan`
is omitted and `missing` carries the unpaid mana symbols (same shape
as the `insufficient_mana` error frame's `missing` list). `cost` is
always the parsed printed cost — the modal renders it next to the
plan for context. Errors: 400 on missing / malformed query params;
404 when the game or card doesn't exist.

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
- [ADR 0011 — Mana pool, cost model, and auto-tapper](decisions/0011-mana-pool-and-auto-tapper.md)
- [ADR 0012 — Continuous effects + layer system (CR 613)](decisions/0012-layer-system.md)
- Sprint S03 — [action protocol + state deltas](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3)
- `server/internal/protocol/` — Go source of truth for wire types
- `server/internal/actions/` — action type catalog and dispatch
