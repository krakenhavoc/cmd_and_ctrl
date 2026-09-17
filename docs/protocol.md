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
| `draw_card` | yes | — | Moves the top of `player`'s library to their hand. S13: server-side no-op when called during the active player's `draw` step (the step-entry hook has already drawn automatically, except on the starting player's turn 1 of a **two-player** game per CR 103.8a — CR 103.8c has nobody skip in a larger multiplayer game). Outside that window the action behaves as before. |
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
| `declare_attackers` | no | `{ "attackers": [{ "attacker": "<uuid>", "target": "<uuid>" }, ...] }` | Declares a whole attacking set in ONE action — the wire verb behind the client's "attack with all" cluster (#318). Step-gated to `declare_attackers`. Each entry is checked independently and **skipped silently** when the creature is unknown, is not a creature, is tapped, is summoning-sick (CR 302.6), has `defender` (CR 702.3), is already declared this combat, or names a seat that is the creature's own controller / eliminated / nonexistent — one ineligible creature must never sink a wide declaration. Rejects with `bad_request` when the batch is empty, exceeds 256 entries, contains a malformed UUID, or when EVERY entry was skipped. Caller must control every listed creature; a foreign creature rejects the whole batch (authorization is not part of the skip contract). Declared creatures tap unless they have `vigilance` (CR 508.1f / 702.20), emit one `EventAttack` each, and drain through a single state-check pass so simultaneous attack triggers reach the stack together (CR 508.1 / 508.2). Prefer this over looping `declare_attacker`: each action is one undo entry and one broadcast snapshot, so N creatures would otherwise cost N of both against a per-turn undo budget of 1. Added in S31.
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
| `counter_ability` | no | `{ "instance_id": "<uuid>" }` | Removes an activated / triggered ability item from the stack. Abilities cease to exist on removal (CR 608.2n); no destination needed. Caller must hold priority. Added in S13.1. |
| `activate_ability` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"?, "ability_index"?: int, "sacrifice_ids"?, "crew_ids"?, "counter_source_ids"?, "counter_kind"?, "strict"?, "auto_tap"? }` | Creates an activated-ability stack item linked to the source card. Same announce-time params as `cast_spell` (minus the source-zone routing). The source card stays in its origin zone; the stack item carries a synthetic ID. Mana abilities are NOT modeled this way (CR 605 — they don't use the stack). Caller must hold priority. Added in S13.1. **With `ability_index`** the payload names a catalog ability (S21 sub-PR 2) and the engine validates and pays the cost: `sacrifice_ids` pays a "Sacrifice a creature" component, `crew_ids` pays a crew number (CR 702.122a), and `x_value` is the X announced for an `{X}` in the ability's mana component (CR 602.2b). X is validated at announce — non-negative, zero unless the cost has an `{X}`, and at least the cost's printed floor — then locked onto the stack item, where the ability's effect reads it back. The card's `activated_abilities[i]` carries `demands_x`, `min_x` and `x_slots` so the client knows to prompt, what floor to enforce, and what a given X actually costs. **#625:** `counter_source_ids` (one `<uuid>`) and `counter_kind` pay a "remove N counters" component. The ability's view carries `counter_cost_n` (present means the ability has the component), `counter_cost_kind` (absent means "a counter" of any kind), `counter_cost_self` (the counters come off the source), `counter_cost_label` (the "from" clause, e.g. "a planeswalker you control") and `counter_cost_options: [{ "card_id": "<uuid>", "kinds": [{ "kind": string, "count": int }] }]`. The options list the permanents the viewer controls that could pay right now, most counters first. They come from the non-targeting candidate walk, so hexproof permanents are included, and the field is absent when nothing can pay. Send `counter_source_ids` for the other-permanent form. Omit it for the self form, or send the source's own ID. Send `counter_kind` for the any-kind form. It is optional for a printed kind, and if sent it must match that kind. The removal is paid at announce with the rest of the cost and is never modified by counter replacement effects. It is not a loyalty activation and has no timing restriction of its own. Errors: `not enough counters to pay that cost` when the named permanent holds too few, `you do not control that card` for another player's permanent, and `bad_request` for a payload that names a counter choice the ability has no component for. **#743:** an ability with an activation condition (CR 602.1b: "Activate only if an opponent controls four or more lands", "Activate only during your turn") is refused with `ability's activation condition is not met` while the condition is false, before X, targets or any cost are checked, so nothing is paid. The condition is checked only at activation, never at resolution. When sorcery speed and the condition both fail, the sorcery-speed error is the one returned. `activated_abilities[i].condition_unmet: true` marks an ability whose condition is false right now. It is absent when the ability has no condition or the condition holds, is evaluated with the permanent's controller as "you", and is sent to every viewer (a condition reads only public information). The client greys the row, as it does for `sorcery_speed`. |
| `activate_loyalty` | yes | `{ "planeswalker_id": "<uuid>", "label"?: string, "delta": int }` | Applies a planeswalker loyalty ability. Sandbox shape: the engine doesn't model the activation as a proper stack item (deferred to S14+) — `delta` is applied immediately to the planeswalker's `loyalty` counter. Sorcery-speed gated; once-per-turn-per-planeswalker (CR 606.3). Caller must hold priority. Added in S13.1. |
| `announce_trigger` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"? }` | Queues a triggered ability for APNAP-ordered drain onto the stack (CR 603.3b). The drain happens at every priority-grant boundary inside the SBA + trigger loop. Sandbox: the player whose card has the trigger announces it manually — auto-fire from card events is the bright line between S13.1 and S14+. Added in S13.1. |
| `mark_damage` | no | `{ "instance_id": "<uuid>", "delta": int }` | Adjusts the damage noted on a creature on the battlefield by `delta` (positive to add, negative to remove). Drives the lethal-damage SBA (CR 704.5g). Positive deltas require controller (admins / spectators bypass); negative deltas are caller-loose so opponents can undo. SBA loop fires after the mutation. Cleanup-step turn-based action zeros every creature's marked damage (CR 514.2). Added in S13.1. |
| `add_player_counter` | yes | `{ "name": string, "delta": int }` | Modifies a named player-level counter (poison, energy, experience, rad, plus homebrew names) by `delta`. Negative deltas clamp at zero. Drives the SBA loop after the mutation so 10 poison or future counter-driven losses fire immediately. The legacy `set_poison` / `set_energy` actions stay synchronised with the underlying `Player.Counters` map. Player-scoped: a seated caller may only adjust their own counters; admins bypass. Added in S13.2. |
| `discard_selection` | yes | `{ "card_ids": ["<uuid>", ...] }` | Resolves the cleanup-step interactive discard prompt (S13.4, CR 402.2). Caller must be in `discard_pending`; the count must exactly match the over-max amount; every supplied card ID must live in the caller's hand. Selected cards move to the caller's graveyard, the caller's pending entry clears, and the cleanup hook re-fires so the cursor resumes its auto-advance. Idempotent no-op for callers not in the pending map. Player-scoped. Added in S13.4. |
| `set_max_hand_size` | yes | `{ "value": int }` | Sets the named player's per-player cleanup-step hand-size cap (CR 402.2). Default 7; `-1` disables the cap (Reliquary Tower / Thought Vessel). Sandbox helper until the S14+ effect catalog wires this to real cards via the S16 layer pipeline. Player-scoped. Added in S13.4. |
| `resolve_choice` | yes | `{ "choice_id": "<uuid>", "card_ids"?: ["<uuid>", ...], "color"?: "W" \| "U" \| "B" \| "R" \| "G" \| "C", "apply"?: bool, "order"?: ["<id>", ...], "assignments"?: [{ "blocker_id": "<uuid>", "amount": int }, ...], "trample_to_player"?: int }` | Resolves a pending async effect-driven choice (S14, CR 701.9b — "player chooses" discard). Caller must be the `chooser` of the named entry in `GameView.pending_choices`; `card_ids.length` must equal the entry's `count`; every supplied card ID must currently live in the entry's `from_player`'s hand. Picked cards move to `from_player`'s graveyard (for `discard_from_hand`), the entry drains from the queue, and an `EventDiscardCard` event fires per card. Distinct from `discard_selection` (S13.4 cleanup, chooser == discarder); `resolve_choice` covers effects where chooser != discarder (Thoughtseize — caster picks from target's revealed hand). **S15:** the entry's `kind` may be `mana_pick` (Birds of Paradise / Arcane Signet); `card_ids` is omitted and `color` carries the chosen single-letter color, validated against the entry's `color_options` slice. The picked color drops into the chooser's mana pool as one `ManaToken`. **S17:** `kind` may be `replacement_order` (CR 616 multi-replacement order prompt — `order` carries the chosen permutation of replacement-effect IDs) or `optional_replacement` (CR 614.10 yes/no — `apply` is the answer). **S18:** `kind` may be `damage_assignment` (CR 510.1c multi-blocker assignment — `assignments` is the per-blocker amounts in chosen order, `trample_to_player` is the overflow when the attacker has trample). **S19:** `kind` may be `trigger_prompt` (CR 603.5 "you may" optional triggered ability — `apply` is the answer; the server routes by inspecting the choice's kind, since `optional_replacement` and `trigger_prompt` share the `apply` payload). **S26:** `kind` may be `choose_creature_type` (CR 614.12 "as this permanent enters, choose a creature type" — Cavern of Souls, Door of Destinies, Vanquisher's Banner, Adaptive Automaton); `creature_type` carries one name from the CR 205.3m vocabulary, validated and normalised server-side and stamped onto the source permanent's named tribe. Added in S14. |
| `resolve_choice` | yes | `{ "choice_id": "<uuid>", "card_ids"?: ["<uuid>", ...], "color"?: "W" \| "U" \| "B" \| "R" \| "G" \| "C", "apply"?: bool, "order"?: ["<id>", ...], "assignments"?: [{ "blocker_id": "<uuid>", "amount": int }, ...], "trample_to_player"?: int }` | Resolves a pending async effect-driven choice (S14, CR 701.9b — "player chooses" discard). Caller must be the `chooser` of the named entry in `GameView.pending_choices`; `card_ids.length` must equal the entry's `count`; every supplied card ID must currently live in the entry's `from_player`'s hand. Picked cards move to `from_player`'s graveyard (for `discard_from_hand`), the entry drains from the queue, and an `EventDiscardCard` event fires per card. Distinct from `discard_selection` (S13.4 cleanup, chooser == discarder); `resolve_choice` covers effects where chooser != discarder (Thoughtseize — caster picks from target's revealed hand). **S15:** the entry's `kind` may be `mana_pick` (Birds of Paradise / Arcane Signet); `card_ids` is omitted and `color` carries the chosen single-letter color, validated against the entry's `color_options` slice. The picked color drops into the chooser's mana pool as one `ManaToken`. **S17:** `kind` may be `replacement_order` (CR 616 multi-replacement order prompt — `order` carries the chosen permutation of replacement-effect IDs) or `optional_replacement` (CR 614.10 yes/no — `apply` is the answer). **S18:** `kind` may be `damage_assignment` (CR 510.1c multi-blocker assignment — `assignments` is the per-blocker amounts in chosen order, `trample_to_player` is the overflow when the attacker has trample). **S19:** `kind` may be `trigger_prompt` (CR 603.5 "you may" optional triggered ability — `apply` is the answer; the server routes by inspecting the choice's kind, since `optional_replacement` and `trigger_prompt` share the `apply` payload). **S21:** `kind` may be `scry` (CR 701.22 — `bottom` is the looked-at cards going under the library and `top_order` the ones staying on top, listed TOP-FIRST; every looked-at card must appear in exactly one list, because a scry moves all of them). **S22:** `kind` may be `surveil` (CR 701.25 — the same shape with `graveyard` in place of `bottom`) or `look_at_top` ("look at the top N cards of your library, then put them back in any order" — Ponder, Sensei's Divining Top; answered with `top_order` alone, since there is no away lane). The three are the **scry family** and the dispatcher routes them **by the pending choice's `kind`**, not by the payload shape every other `resolve_choice` branch keys off: all three carry `top_order` and `look_at_top` carries nothing else, so no key's presence identifies them. Send the destination key that matches the kind — `bottom` for scry, `graveyard` for surveil, neither for `look_at_top`. Added in S14. |
| `activate_mana_ability` | yes | `{ "instance_id": "<uuid>", "ability_idx"?: int }` | S15. Activates the `ability_idx`-th mana ability on a battlefield permanent the caller controls. `ability_idx` defaults to 0 (the only ability for basic lands and most rocks). Single code path for catalog-declared abilities (Sol Ring, Arcane Signet, Birds of Paradise — looked up via the `CatalogManaAbilities` hook) and synthetic basic-land abilities (Forest → `{G}`) the engine derives from `TypeLine`. Pays the tap cost (rejects with `bad_request: "card already tapped"` if the source is tapped), then materialises the produced mana: single-color slots drop straight into the controller's pool as one `ManaToken`; multi-option slots (pipe syntax) queue a `mana_pick` PendingChoice for the controller. Mana abilities don't use the stack (CR 605.3) — synchronous under the write lock. Caller need NOT hold priority (mana abilities are special, CR 605.1a). Player-scoped: caller must control the source. **#743:** `mana_abilities[i].condition_unmet: true` is the same flag for a mana ability whose activation condition is false right now (Temple of the False God with four lands, Mox Opal without metalcraft). Same evaluation and visibility as `activated_abilities[i].condition_unmet`; the activation is refused with `ability's activation condition is not met` and nothing is tapped or paid. |

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

### Set-level rules on a card-set prompt (#624)

A `search_library` or `choose_cards` prompt can carry a rule about the
chosen cards **as a set**: "two basic lands that share a land type",
"discard two cards unless you discard a creature card". The rule stays
on the server. Nothing new is added to `PendingChoiceView`, because the
prompt's `reason` is already the card's own sentence and says the rule.
A `resolve_choice` that has the right number of cards, all of them
candidates, but breaks the rule gets an `error` frame with `payload.code
= "bad_request"`. The prompt **stays in `pending_choices`**, so the
client leaves its modal open and the player can choose again, and the
modal shows the error's message itself: the board's error toast sits
under the modal's backdrop, so a refusal shown only there is unreadable
while the prompt is open. For
`choose_cards` the message is `"that selection doesn't meet the card's
condition — check its text and choose again"`. (`search_library` still
answers with the generic `"invalid parameter"`.) Every set listed in
`legal_moves` for such a prompt already passes the rule.

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
| `payload.kind` | string | no | `"say"` (or absent) for a human message; `"bot_improvisation"` / `"bot_reasoning"` for the server-originated bot lines below. Ignored on inbound frames — a player who sends one gets `"say"`. Added in S31 sub-PR 8. |
| `payload.reason` | string | no | A bot policy's `Decision.Reason`. Present only on bot lines. Added in S31 sub-PR 8. |

The originating client receives the server-stamped frame with the
same `id` it sent. Other recipients see the same `id` — clients can
use this for deduplication if they ever render a local optimistic
copy before the broadcast lands.

**Server-originated bot lines — added in S31 sub-PR 8.** A bot seat
has no socket, so its lines come from `Hub.BroadcastChat` rather than
from `handleChat`, and each carries a freshly minted frame `id`.
There are two kinds and they are not interchangeable:

- `"bot_improvisation"` — the bot applied an effect by hand with the
  sandbox verbs because the catalog cannot execute that card
  ([ADR 0033 §8](decisions/0033-ai-bot-seat.md)). **Mandatory
  disclosure: a client must render these.** An unannounced
  improvisation is a bot cheating, and a client that hides the
  announcement is what makes it one. The same commit writes a tagged
  line to the replay log (see `snapshot` below).
- `"bot_reasoning"` — the bot narrating why it took an ordinary move.
  Debug output; the reference client hides it unless the viewer turns
  on **Settings → Gameplay → Show bot reasoning**, and the server
  only emits it when the runner's `Narrate` config is on. It can name
  cards in the bot's *own* hand, which is a disadvantage the bot
  accepts rather than a leak — a policy is handed the same filtered
  view a human at that seat gets and never sees another seat's hidden
  state.

`payload.reason` rides along on both, split out of `text` so a client
can show the announcement while keeping the reasoning behind the
setting.

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

- **GameView**: `{ id, state, seats[], battlefield, stack, exile, turn, mulligans_open, starting_seat, stack_items?, pending_triggers?, delayed_triggers?, split_second_active?, pending_choices?, log?, reveals? }` — `mulligans_open` is true between `Start` and the moment all seated, non-eliminated players have called `keep_hand` (S08). `starting_seat` is the seat index that took the first turn; used by the server to enforce the CR 103.8a turn-1 skip-draw rule (two-player games only — CR 103.8c has nobody skip in a larger multiplayer game), surfaced on the wire so spectators / reconnects render the same first-player decision. Pre-S13 snapshots decode as `0` (S13). `stack_items` is the announce-time metadata for every item on the stack (S13.1) — see StackItemView below. `pending_triggers` is the APNAP queue of triggered abilities waiting to drain (CR 603.3b). `delayed_triggers` (S22, omitempty) is the queue of CR 603.7 delayed triggered abilities still owed — "at the beginning of the next end step, return that card to the battlefield" — as `{ id, controller, source?, label?, at, created_turn?, cards? }`, where `at` is the step whose beginning fires it and `cards` are the instance IDs the effect acts on. Public information, so it is not redacted per viewer. `split_second_active` mirrors the engine's "no responses allowed" gate (CR 702.61). `pending_choices` (S14, omitempty) is the async effect-driven decision queue — see PendingChoiceView below. `log` (S31, omitempty) is the public game log — see LogEvent below. `reveals` (S22, omitempty) is the broadcast reveal window — see RevealView below.
- **PlayerView**: `{ id, name, seat, life, poison?, energy?, library, hand, graveyard, command, commander_damage, life_history, eliminated?, hand_kept?, mulligans_taken?, mana_pool?, is_bot?, bot_tier?, bot_deck? }` — `mana_pool` (S15, omitempty) is the player's typed mana pool as an ordered list of single-character color strings (`"W"`, `"U"`, `"B"`, `"R"`, `"G"`, `"C"`) in tap order. Drives the `ManaPoolPips` row in the player header. Server-side empties at every step boundary (CR 106.4) so the wire surface stays small. Public to all viewers (matches paper Magic — floating mana sits visibly next to the player). `life_history` is a per-player rolling log of `LifeChangeView` entries (`{ delta, new_total, at }`, RFC3339 timestamp), bounded server-side at 50 entries and never filtered (life is public). `eliminated` is omitted unless true; set when the player has conceded (S08) or, in future S13+ work, has lost to a state-based action. `hand_kept` is true once the player has committed via `keep_hand`; `mulligans_taken` counts mulligans in the current opening-hand window. Both omitted when false/zero. All five fields added in S08. `is_bot`, `bot_tier` and `bot_deck` (S31, all omitempty) mark a bot seat, one driven by an in-process `aiseat` runner instead of a browser ([ADR 0033 §9](decisions/0033-ai-bot-seat.md), [docs/bot.md](bot.md)). `is_bot` is true only on such a seat. `bot_tier` is its policy tier: `"random"`, `"heuristic"`, `"assisted"` or `"strong"`. `bot_deck` is the curated deck ID it was seated with, such as `"izzet-aggro"`. It is empty when the bot was seated with a pasted decklist (`{tier, format, source}`). All three are public, unredacted, and absent on a human seat. The client renders the BOT chip, the robot avatar mark and the thinking pulse from them. They also ride the engine snapshot and the persisted lobby metadata, so a bot seat survives a restart ([docs/lobby.md](lobby.md)).
- **ZoneView**: `{ kind, owner?, count, cards[] }` — `owner` omitted for shared zones
- **CardView**: `{ instance_id, name, owner, controller, scryfall_id?, oracle_id?, type_line?, power?, toughness?, tapped?, counters?, is_commander?, battle_x?, battle_y?, attacking_target?, blocking_target?, auto?, target_mode?, mana_cost?, produced_mana?, mana_abilities?, abilities? }` — **S16:** `power`, `toughness`, `type_line`, and the new `abilities` field carry the *effective* (post-CR-613-layer-resolution) characteristics, not printed. The wire keys are unchanged; only the resolution semantics are. Cards with no static ability affecting them produce identical wire output to pre-S16 (printed == effective). When an anthem is in play, the controlled creature's `power` / `toughness` arrive +1; when Mycosynth Lattice is in play, every permanent's `type_line` includes `"Artifact"`; when Lord of Atlantis is in play, other Merfolk creatures' `abilities` includes `"islandwalk"`. `abilities` (S16, omitempty) is a `[]string` of keyword names ("flying", "first strike", "trample", "islandwalk", etc.) — drives S18's keyword renderer; behavior of the keywords lands with S18 (combat sub-step rewrite). Non-protocol-bumping change: pre-S16 clients ignore the unknown `abilities` field harmlessly. `mana_cost` (S15, omitempty) is the card's printed cost string in Scryfall format (`"{2}{R}{R}"`, `"{X}{G}"`, `"{W/U}{W/U}"`); the parser accepts generic, colored (`{W|U|B|R|G|C}`), variable (`{X}`), hybrid two-color (`{W/U}`), hybrid generic (`{N/W}`), phyrexian (`{W/P}`), and snow (`{S}`) symbols. Drives the cost chip on the cast modal and the strict-mode gate. Empty / omitted ⇒ treated as costless (for cards Scryfall ingestion couldn't parse, which the cost gate emits as an `EventCostWarning` rather than rejecting). Redacted to empty for cards the viewer is not a `known_by` of. `produced_mana` (S15, omitempty) is the parsed Scryfall `produced_mana` field (e.g. `"{C}{C}"` for Sol Ring, `"{W|U|B|R|G}"` for Birds of Paradise) — informational; the canonical activation surface is `mana_abilities`. `mana_abilities` (S15, omitempty) is the per-ability projection the client renders into the right-click "tap for mana" menu: each entry is `{ "label": string, "produced": string, "cost_tap"?: bool, "cost_sacrifice"?: bool }`. Synthetic basic-land abilities are rendered into this list when the catalog has nothing registered (Forest → `[{"label": "Add {G}", "produced": "{G}", "cost_tap": true}]`). `scryfall_id` is stamped at deck-import time and lets the client resolve images via `GET /cards/{id}/image`. Omitted for placeholder cards (demo game seeded via `CMDCTRL_SEED_DEMO`). `oracle_id` (S14) is the Scryfall oracle-level identifier (stable across printings) and is the catalog lookup key — every printing of Lightning Bolt shares one `oracle_id`. `type_line` (e.g. "Legendary Creature — Human Wizard"), `power`, and `toughness` carry Scryfall data the client uses to filter creature-only UIs and label combat rows; all three omitted for placeholder cards. `battle_x` / `battle_y` are normalised positions in `[0, 1]` stamped by `set_battlefield_position` (S06+). **Sent for battlefield cards only, and for every one of them (#29)** — so an absent pair means "this card is not on the battlefield", never "this card is at the origin". They previously carried `omitempty` on a plain float and so vanished whenever the position was `(0, 0)`, which conflated a deliberate origin stamp with a permanent nobody had positioned. The origin is a routine value, not an exotic one: `set_battlefield_position` *clamps* rather than rejects, so every negative and NaN coordinate lands on exactly 0, and `x = 0` is "first in the row" for the client's within-row sort. Outside the battlefield the pair is omitted because it is meaningless there — zone exit clears it — and those zones are the bulk of every frame. Note this makes the wire honest rather than more expressive: `game.Card` carries no "positioned" flag, so "on the battlefield but never positioned" is not a state the server can distinguish from `(0, 0)`; such a card reports `(0, 0)`. Clients should keep an `?? 0` fallback — replays captured before this change omit the pair on battlefield cards too. `attacking_target` (player UUID) and `blocking_target` (attacker instance UUID) are set by `declare_attacker` / `declare_blocker` and cleared on zone exit and by `clear_combat`. Both omitted when not set. `type_line`, `power`, `toughness`, `attacking_target`, `blocking_target` added in S08. `attached_to` (S24, omitempty) is the CR 301.5c / CR 303.4 attachment relation for an Equipment or an Aura — a `TargetRefView` (`{ kind, id }`) naming the permanent (`kind: "card"`) or player (`kind: "player"`) this card is attached to. Omitted for every card attached to nothing. Shipped in ONE direction only: "what is attached to this creature" is derived client-side by partitioning the battlefield on this field, so the two directions cannot disagree. Not redacted — attachment is public battlefield state exactly like `attacking_target`. See [ADR 0036](decisions/0036-attachments.md). `restrictions` (S24, omitempty) is the CR 508.1c / 509.1b / 602.5 restriction set the engine computed for this permanent, as stable snake_case tokens: `"cant_attack"`, `"cant_block"`, `"cant_be_blocked"`, `"cant_activate"`, `"cant_activate_mana"`. Omitted for the permanent nothing is restricting, which is nearly all of them. Deliberately NOT folded into `abilities`: a restriction is not a keyword the permanent has, it is an effect something else has (Pacifism, Arrest, a Whispersilk Cloak), so it renders as a disabled control with a reason rather than as a keyword badge. The client READS it and derives nothing — who may attack is the server's decision and this is how it says so. Public battlefield state, redacted only for a card the viewer is not a `known_by` of (#95). See [ADR 0045](decisions/0045-combat-restrictions.md). `auto` (S14, omitempty) is true when the card's `oracle_id` is registered in the catalog (drives the gold-leaf AUTO badge). `target_mode` (S14, omitempty) is the announce-time target-prompt shape the client should show when casting: `"any"`, `"player"`, `"creature"`, `"stack_spell"`, `"card_in_graveyard"`. Empty ⇒ no prompt (cast fires immediately). Both fields redacted to zero for cards the viewer is not a `known_by` of.
- **TurnView**: `{ number, active_seat, priority_holder, phase, step }` — `priority_holder` is the seat index (0-based) that currently holds priority within the step, OR `-1` (the `NoPriority` sentinel) during steps that don't grant priority. S13 lands the cursor with `priority_holder = -1` for `untap` and `cleanup` (CR 502.4 / 514.3); auto turn-based actions (auto-untap, auto-draw, auto-advance through cleanup) fire from the step-entry hook so the cursor never sits idle on a no-priority step. Added in S07; sentinel added in S13.
- **StackItemView** (S13.1): `{ id, kind, controller, owner, source_card_id, label?, targets?, modes?, x_value?, distribution?, hold_priority?, split_second? }` — announce-time metadata for one item on the stack. `kind` is `"spell"` (card lives in `stack` zone), `"activated"` or `"triggered"` (no card; source stays in its origin zone). `targets` is a list of `TargetRefView` slots: `{ kind: "player" | "card" | "self" | "none", id?: "<uuid>" }`. Resolution re-checks targets per CR 608.2b — if every targeted slot is illegal, the spell is "countered by game rules" and routes to its owner's graveyard. `PlayerView.commander_casts` (S13.1, omitempty) is the per-commander cast tax counter map keyed by commander UUID string. `CardView.damage_marked` (omitempty) drives the lethal-damage SBA.
- **CardView.known_by_you / face_down** (S13.5): per-viewer card visibility. `known_by_you` is true when the viewer is in the server-side KnownBy set for that card; the wire surfaces full printed characteristics in that case. When false, the wire keeps only instance_id / owner / controller / tapped / damage_marked / face_down / battlefield position / combat declarations (`attacking_target`, `attacking_target_kind`, `blocking_target`) / `goaded_by` / `attached_to`, and zeroes everything read off the card itself — name, type line, Scryfall ID, P/T, counters, costs, faces, `auto`, `target_mode`, `mana_abilities`, `activated_abilities`, `restrictions`, `exile_play`, the hand-zone cast stamps, and the type-derived `summoning_sick` / `loyalty_activated` / `defense` / `protector_player` (#95; `face_down_view_test.go` pins this as an allowlist). The engine state stays observable, but identity doesn't leak. `face_down` is the visual flip flag (CR 406.3 / 708 — Necropotence's exile today; morph / manifest / foretell later); the client draws a card back for a face-down card the viewer does not know, and the face for one it does. Library cards have no knowers post-shuffle; opening-hand cards are known to their owner only; battlefield / stack / exile / graveyard / command-zone cards are public (all seated players are knowers). `ShuffleLibrary` clears every library card's KnownBy; `Mulligan` clears hand + library and re-grants the owner on the new opening hand. S14's `keepKnownInHandZone` view-filter path preserves revealed opponent-hand cards end-to-end: a seated viewer receives opponent hand zones containing only their `KnownByYou == true` entries (the `Count` stays accurate so the client can render the hidden remainder as face-down placeholders); spectator / admin connections still see a fully-hidden opponent hand via `hideZoneContents`.
- **The engine event log is not on the wire.** S14 documented an `events` field on `GameView`, a rolling window of `EventView` entries (#139), but it was never added: no `GameView` in `server/internal/protocol/view.go` or `client/src/lib/protocol.ts` has ever carried one. The engine's log (`game.Game.Events`, kinds in `server/internal/game/events.go`) stays on the server. Game-state persistence saves it (`server/internal/game/snapshot.go`), and it reaches clients only through two table-visible projections built from it: `log` (LogEvent, below) and `reveals` (RevealView, below).
- **LogEvent** (S31, omitempty): `{ seq, kind, turn?, step?, seat, target_seat?, card_id?, target?, amount?, old_zone?, new_zone?, combat?, combat_step?, text }` — one line of the **public game log** ([ADR 0033](decisions/0033-ai-bot-seat.md) §4). `GameView.log` is the last 200 table-visible events, **oldest first**, and it is a projection of the engine's own event log rather than a stored buffer — nothing on `game.Game` holds it, so it survives an undo, a snapshot restore and a deploy by riding `Game.Events`, which already does.

  `kind` is one of `step`, `cast`, `resolve`, `fizzle`, `counter`, `zone`, `draw`, `life`, `damage`, `attack`, `block`, `token`, `sacrifice`, `eliminated`, `reveal` — deliberately coarser than the engine's event kinds, because several engine events are one line to a reader and most engine events are no line at all.

  **Players are seat indices, not UUIDs.** `seat` is the responsible player (`-1` when there is none — an SBA life loss, a spell resolving with nobody to credit) and `target_seat` is the player acted on. `card_id` and `target` are card instance IDs; exactly one of `target_seat` / `target` is set on an entry that has a target at all. `step` rides only on `step` entries: every entry after one belongs to that step until the next, and a consumer that wants the step on every line carries it forward. All of this is wire-cost discipline — the struct repeats 200 times on every snapshot frame.

  `text` is the rendered line ("Aang cast Lightning Bolt", "Turn 7 — Katara · precombat main") and is what a panel prints.

  **Combat damage steps.** `combat` (S19) marks a `damage` entry as combat damage (CR 510). `combat_step` (#187, omitempty) says which combat damage step dealt it: `"first_strike"` or `"regular"` (CR 510.4 — a combat with a first-strike or double-strike creature has two combat damage steps). It is set **only when the combat had a first-strike step**, on the damage of both steps; a combat with no first strike or double strike anywhere has untagged damage, including damage that lands later from a CR 510.1c damage-assignment prompt or a CR 616 replacement-ordering prompt. So the tag's presence alone says there are two beats to show, and a client never works out from keywords which creature dealt damage in which step. A tagged entry's `text` says so: "Fencing Ace dealt 1 combat damage to Grizzly Bears (first strike)", "Grizzly Bears dealt 2 combat damage to Fencing Ace (regular damage)"; untagged lines are unchanged. The tag rides the engine's `game.Event.CombatStep` (`combat_step` in the snapshot file), so it survives an undo, a restore and a deploy like the rest of the log; an event written before the field existed decodes untagged. A paused prompt keeps the step it was queued in: `DamageAssignmentFrame.CombatStep` is **server-side only** and is not on `PendingChoiceView.damage_assignment`. Both fields are additive and their zero value means untagged, so there is no snapshot schema bump, and `snapshot_drift_test.go` needs no new entry because `Event` and `DamageAssignmentFrame` are embedded in the snapshot by value. Engine bugs that change *which* damage is dealt, not how it is tagged: [#702](https://github.com/krakenhavoc/cmd_and_ctrl/issues/702) (a first-strike multi-blocker prompt's damage lands after the regular step's), [#715](https://github.com/krakenhavoc/cmd_and_ctrl/issues/715), [#716](https://github.com/krakenhavoc/cmd_and_ctrl/issues/716). See [ADR 0053](decisions/0053-combat-damage-beats.md) Decision 1.

  **Visibility.** The log goes through the same per-viewer filter as every other field, in two places. At build time, an event whose card sits in a hidden zone at *both* ends (library → hand, hand → library) is emitted with no card reference at all — not even the instance ID, which would otherwise let a client follow a tutored card onto the battlefield — and a draw never names the card for anyone, including the drawer. At filter time, every surviving `card_id` / `target` goes through the same `known_by` predicate that redacts `CardView`, so a face-down permanent's entry reads "a card entered the battlefield" for everyone but its controller, and `text` is re-rendered to match. A log entry can never say more than the zones alongside it already do.

  **Reveals.** A `reveal` entry (#170 follow-up) is one line per reveal: the engine's per-card `reveal_cards` events that share a `reveal_seq` collapse into one entry, with `amount` the card count and `old_zone` the zone they were revealed from. It **never carries `card_id`**, for any viewer, whatever zone the cards came from; the instance ID of a card revealed out of a hand or library is the same correlation handle the build-time rule above withholds, and the knower filter cannot stand in for it because a reveal makes every seat a knower. A reveal to the whole table names the cards by printed name in `text` ("Aang revealed 3 cards from their library: Island, Swamp, Opt"), up to five names and then "and N more", and the names are not redacted per viewer: the table saw them, and a later shuffle clearing `known_by` does not un-see them (the same argument as `reveals` below). A reveal to one player sets `target_seat` and names nothing for anyone ("Aang revealed a card from their hand to Katara"); no engine path emits that shape yet.

- **RevealView** (S22, omitempty): `{ seq, turn?, seat, source?, reason?, from?, cards[], count? }` — one **broadcast reveal** (CR 701.20). `GameView.reveals` is the reveals that happened **this turn**, oldest first, at most **4** of them. Like `log` it is a projection of `Game.Events` and nothing on `game.Game` stores it, so it survives an undo, a snapshot restore and a deploy for free.

  This is the counterpart to the controller-only look-at-cards prompt that `scry` / `surveil` / `look_at_top` ride, and the two are shaped nothing alike on purpose. A look-at names one chooser and ships `CardView`s complete with instance IDs, and `filterPendingChoices` drops every option a non-chooser does not know. A reveal names no chooser, goes to every seat **identically**, is never redacted, and carries **no instance IDs at all**.

  **What a non-controller learns, and for how long.** They learn the printed identity of each revealed card — `name`, `mana_cost`, `type_line`, and `scryfall_id` (a *printing* id, so the client can draw the face). That is exactly what a player sitting at the table sees and nothing more: a reveal off the top of a library does not make the rest of the library's order public. They learn it for two different durations, and the split is deliberate. The **frame entry** lives for the rest of the turn it happened in, or until four newer reveals push it off, whichever comes first — long enough that a client which dropped a frame still catches it. The **knowledge** is permanent and is not carried by this field at all: it is the ordinary `known_by` set, which `RevealForEffect` writes for every seat, and which a shuffle clears. So a card revealed on its way to hand stays readable to the whole table in that hand; a card revealed on top of a library about to be shuffled does not.

  **No correlation handle.** `source` is the *name* of the card that revealed, never its instance ID, and it is omitted when that card is not in a zone the table can see. `RevealedCardView` has no ID field to fill in — a type with nowhere to put one cannot forget to blank it. This is the same "the instance ID alone is the leak" argument the public log settled at build time and the pending-choice options had to settle again at filter time, arriving here a third time: a reveal makes a card public *for that moment and to that extent*, and the cards usually go straight back into a hidden zone, so a stable UUID would let any client — or any bot policy, which reads the same bytes under [ADR 0033](decisions/0033-ai-bot-seat.md) §3 — recognise the card turns later in a zone it was never entitled to read.

  **Bounded.** `cards[]` is truncated to 8 entries and `count` reports the true total, so a client renders "+N more". The cap exists for Hermit Druid, which reveals a whole library when it finds no basic land; every card in the catalog that reveals a *fixed* number fits under it (Fact or Fiction's five is the largest). Truncating what is drawn never truncates what is known — the identities are public through `known_by` either way. A saturated window costs ~7 KB on a four-player frame, against the log's 32 KiB and `legal_moves`' 24 KiB budgets.

  `seq` is the engine sequence number of the reveal's first event: monotonic and stable across frames, which makes it the dedupe key a client must use, since there are deliberately no card IDs to key on. `seat` is a seat index (`-1` when there is none). `from` is the zone the cards were revealed *out of* — the cards did not move, because a reveal is not a zone change.

- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder).
- **LegalMoveView** (S31, omitempty): `{ type, player, params?, kind, label, source?, always_legal?, cost? }` — one fully-specified thing the viewer's seat may do right now, produced by `server/internal/legal` ([ADR 0033](decisions/0033-ai-bot-seat.md) §1). `type` is the action type the move performs (`pass_priority`, `cast_spell`, `activate_ability`, `activate_mana_ability`, `declare_attacker`, `declare_blocker`, `resolve_choice`, `keep_hand`, `mulligan`, `discard_selection`) and `params` is **exactly** the `ActionPayload.params` object that performs it — a client fires a move by sending `{ type, player, params }` unaltered, and the server is contractually obliged to accept it. `kind` coarsely groups the move (`pass` / `land` / `cast` / `activate` / `mana` / `attack` / `block` / `choice` / `mulligan`) so a consumer can ask "is there anything here but pass?" without parsing labels. `label` is human-readable and menu-ready (`"Cast Lightning Bolt targeting Kess"`). `source` is the instance ID of the card the move is about, when there is one — this is the join key the client greys hand cards on.
  **Own seat only, always.** `legal_moves` names cards in a hand; another seat's list would leak exactly the hidden information every other redaction in this document protects. The server enumerates per seat into an unexported map and `FilterViewFor` hands back only the viewer's own entry; spectators, admins and replay readers (empty viewer ID) get nothing, which is the one place the "empty viewer sees everything" convention deliberately does not apply. The unfiltered `GameView` that reaches the crash dump and the replay log carries no move list at all.
  **Empty is the normal case.** The list is populated only when the seat actually owes a decision — priority, a pending choice, a combat declaration, the mulligan window or a cleanup discard. A seat with nothing to do gets the field omitted, and a client must read an absent `legal_moves` as "no information", not as "nothing is legal": the safe reading when the field is missing is to stay permissive and let the server reject.
  **`always_legal`.** Present and `true` on a move the server cannot refuse whatever else happens between this frame and the click: `pass_priority`, and a pending choice's one unconditional answer (a `search_library` prompt's "fail to find", CR 701.23b; a `choose_cards` prompt's "choose nothing" when its `choose_min` is 0). Absent means "legal right now", which is what every entry in this list already promises — the flag is the stronger claim that it is still legal after the board has moved or after some other answer to the same prompt was rejected. Automated seats read it as the way out of a prompt they cannot otherwise answer: a seat owing a choice is enumerated that choice's answers and nothing else, so there is no pass to fall back on, and a bot with nowhere to go stops playing and holds the table (#544). Choice kinds that have no unconditional answer mark nothing.
  **`cost`** (omitempty): `{ life?, loyalty?, counters?: [{ card_id, counter, n }] }`. It is the part of a move's price that `params` cannot name, because the engine reads it off the ability rather than off the payload (#74, #547). It is advice about the move, never sent back. `life` is life paid at announce, and `loyalty` is a loyalty ability's signed counter delta on the source. `counters` (#625) lists the counters a "remove N counters" cost takes and which permanent they come off. That permanent is often not `source`: Heart of Kiran's alternative crew is paid with a planeswalker's loyalty counter. A policy should price the removal against that permanent, including the whole permanent when the counter is its last loyalty.
  **Capped, twice.** Target and mode expansion is capped at 12 concrete moves per source card (`legal.Options.MaxExpansionPerSource`), so a spell with thirty legal targets ships twelve of them. For a "search your library for up to N" prompt, and for a `choose_cards` prompt, that budget is spread ACROSS pick sizes rather than spent smallest-first, so a card that fetches a pair still offers pairs when the library holds more matching cards than the cap (#544), and a "discard two unless you discard a creature card" prompt still offers pairs when most single cards break its rule (#624). On top of that the wire projection caps the whole list at 48 moves: past that it *degrades* rather than truncates, keeping the first move of every `(source, kind)` pair and dropping only the alternatives. The invariant a client may rely on is therefore **"every card that has a legal move is represented by at least one entry"** — never "this is the complete set of targets". Targeting UI reads `CardView.legal_targets`, not this field.
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder). **S16.5 adds `copy_target`** — "you may have this permanent enter as a copy of ..." (Clone, Phyrexian Metamorph, Spark Double, Sakashima the Impostor). `options[]` carries the permanents that may be copied (public battlefield cards, unredacted); answered with the shared `{choice_id, card_ids}` payload, where an EMPTY list declines and the permanent enters as its own printed self. The permanent is still on the stack while the prompt is open — the answer decides what it enters AS, so its own ETB trigger has not fired yet either.
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`; `type_options` is the full creature-type vocabulary for `choose_creature_type`, materialised from the engine rather than stored on the choice. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder).
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder). **S21/S22 look-at kinds:** `scry`, `surveil` and `look_at_top` put the looked-at library cards in `options[]`, top-first. Both keywords are "look at", not "reveal" — only the chooser is a knower. The **count** stays visible to the table, which is correct: "scry 2" is a printed number, and `search_library` goes further still by dropping `options[]` outright for non-choosers, because the number of MATCHES is itself hidden information about a hidden zone. **S31 correction:** every kind now drops an option the viewer is not a knower of, rather than projecting it as a back. Redacting kept the `instance_id`, which is the right trade for a face-down permanent (the ID is already on that viewer's wire) and the wrong one for a card whose zone the seat projection has already stripped for that same viewer — a stable UUID for a specific card in a hidden zone is a correlation handle, since the card is drawn in private and cast in public under the same ID. The chooser always keeps their whole list, known or not; picking a back is a legal answer.

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

**Annotations (S31 sub-PR 8).** A replay line may carry an optional
`annotation` object alongside `seq` and `game`. It exists because the
log is a stream of full snapshots: "what happened here" normally has
to be diffed back out of two consecutive lines, and some events are
worth saying outright. The field is written only on the replay /
crash-dump path — no WebSocket `snapshot` frame ever carries one.

```json
{
  "seq": 42,
  "game": { "...": "..." },
  "annotation": {
    "tag": "bot_improvisation",
    "seat": "8a2c0d8e-...",
    "seat_name": "Kess",
    "card": "Grim Tutor",
    "effect": "search my library, then lose 3 life",
    "text": "Grim Tutor: search my library, then lose 3 life (improvised — …)",
    "reason": "no catalog spec for this card",
    "steps": ["move_card", "change_life"]
  }
}
```

| Field | Type | Notes |
|---|---|---|
| `annotation.tag` | string | Grep handle. `"bot_improvisation"` is the only tag today — `grep bot_improvisation replays/*.jsonl` finds every bot improvisation in a game that went wrong. |
| `annotation.seat` / `seat_name` | string | The acting seat. |
| `annotation.card` / `effect` | string | The card played by hand and the intended effect, in the words the table was given. |
| `annotation.text` | string | The exact chat line broadcast, verbatim — the record is what the table was *told*, not a reconstruction. |
| `annotation.reason` | string | The policy's `Decision.Reason`. Always recorded here even when no client is showing reasoning. |
| `annotation.steps` | string[] | The wire action types the bundle applied, in order. |

Everything in an annotation is public information (it repeats the
chat line), so it is safe inside a replay artifact a player can
download.

---

## Schema evolution rules

- **Breaking changes** bump `v` and require updating both server and client
  in lockstep. Keep breaking changes rare.
- **Additive changes within a version** (new optional fields, new error
  codes) are allowed and do not bump `v`. Readers ignore unknown fields.
- **Removing a field** is always breaking. Don't do it without a `v` bump.
- **`omitempty` on a numeric or boolean field is a trap.** Go's zero value
  is indistinguishable from "unset" on the wire, so `omitempty` silently
  deletes a meaningful `0` / `false`. Only use it where the zero genuinely
  carries no information. Emitting such a field *more* often — dropping
  `omitempty`, or widening it from a Go zero-check to a real condition — is
  **not** breaking and does not bump `v`, because readers already fall back
  to the zero value for an absent field. Narrowing it is breaking.
  `battle_x` / `battle_y` (#29) are the worked example: they now go out for
  every battlefield card and no others, so "absent" states a fact about the
  zone instead of accidentally encoding two different things.
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
