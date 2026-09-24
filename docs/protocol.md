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
| `payload.code` | string | yes | Short machine-readable code. v0 codes: `bad_version`, `bad_json`, `bad_request`, `internal`. S15 adds `insufficient_mana`, #705 adds `illegal_block`, and #1063 adds `attack_tax_unpaid` — see below. |
| `payload.message` | string | yes | Human-readable message safe to display to the user. |
| `payload.missing` | string[] | no | Present only when `code == "insufficient_mana"` (S15). List of mana symbols the caller's pool could not cover, in the order they appear in the printed cost (e.g. `["{R}", "{1}"]`). Client renders these verbatim into the override toast. |
| `payload.card_id` | string (UUID) | no | Present when `code == "insufficient_mana"` (S15): instance ID of the card whose cast was rejected. Lets the client's "Cast anyway" / "Auto-tap & cast" buttons re-fire the same cast without needing to round-trip through the user's last click. Also present when `code == "illegal_block"` (#705): instance ID of the refused blocker. |
| `payload.reason` | string | no | Present only when `code == "illegal_block"` (#705). The stable snake_case token naming why the block was refused — see below. |

`id` matches the originating client frame's `id` when possible; otherwise
empty string.

#### `insufficient_mana` (S15)

Emitted only when a `cast_spell` arrives with `strict: true` and the
caller's mana pool cannot cover the effective cost (printed cost +
commander tax for command-zone casts) — or, since #1296, when a
catalog `activate_ability` arrives with `strict` / `auto_tap` and
neither the pool nor the auto-tapper can pay; that frame says
"insufficient mana to activate that ability" and carries no
`card_id`, so the two cast buttons below are never offered for it. The `missing` list carries the
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

#### `illegal_block` (#705)

Emitted when a `declare_blocker` or `declare_blockers` names a pair — or,
since #750, a block COUNT — the engine's block-legality check refuses
(CR 509.1b; [ADR 0045](decisions/0045-combat-restrictions.md) addendum,
Decisions 8, 12 and 13). Nothing is stored, logged or announced, and for a
set the WHOLE set is refused, not the offending entry. `message` is a
player-facing sentence the server builds and addresses to the caller, so the
client shows it verbatim and never re-derives the rule — for example
`"Cold-Eyed Selkie has islandwalk, and you control an Island (Island)."` (a
player other than the defending player reads the defender's name instead of
"you"). `card_id` is the blocker; `reason` is one of:

| `reason` | Refused because |
|---|---|
| `cant_block` | a "can't block" restriction is on the blocker (Pacifism, Carrion Feeder) |
| `cant_be_blocked` | a "can't be blocked" restriction is on the attacker (Whispersilk Cloak, Rogue's Passage) |
| `flying` | the attacker has flying and the blocker has neither flying nor reach (CR 702.9b) |
| `landwalk` | the attacker has a landwalk ability (`islandwalk`, …, `nonbasic landwalk`) and the defending player controls a land it names (CR 702.14c) |
| `fear` | the attacker has fear and the blocker is neither an artifact nor black |
| `intimidate` | the attacker has intimidate and the blocker is not an artifact and does not share a color with it |
| `shadow` | exactly one creature has shadow, so they cannot block each other |
| `horsemanship` | the attacker has horsemanship and the blocker does not |
| `skulk` | the attacker has skulk and the blocker has greater current power |
| `protection` | the attacker has protection from a quality the blocker has (CR 702.16f, #662) |
| `cant_be_blocked_by` | a block rule on the attacker names creatures that may not block it ("can't be blocked by creatures with power 2 or less", #750) |
| `cant_be_blocked_except_by` | a block rule on the attacker allows only the creatures it names, and the blocker is not one ("can't be blocked except by Walls", #750) |
| `cant_block_attacker` | a block rule on the blocker's side refuses this attacker (Champion of Lambholt, #750) |
| `too_few_blockers` | the declaration would leave fewer creatures blocking the attacker than its minimum — menace's 2 (CR 702.111b), Pathrazer of Ulamog's 3. `n` is that minimum |
| `too_many_blockers` | the declaration would put more creatures on the attacker than its maximum (Hungering Hydra's "can't be blocked by more than one creature"). `n` is that maximum |
| `not_defending` | the blocker's controller is not defending against that attacker: it is attacking another player, a planeswalker another player controls or a battle another player protects — or nothing at all (CR 802.4a / 509.1a, #1339). A creature whose planeswalker or battle has left is still defended by the player who was defending it (#1364, CR 506.4c): "Grizzly Bears is still attacking, though what it attacked is gone, so only P3 can block it." Like the count reasons it comes only from the declaration verbs. The sentence names where the attacker is pointed ("Grizzly Bears is attacking P3, so only P3 can block it."). A pairing the declaration already holds is not re-judged, so an attacker reselected after it was blocked (CR 508.7a) keeps its blockers |

The tokens are stable once shipped. The addendum reserves more
(`declaration_limit`, `tapped`, …);
each is added to this table in the change that first sends it.

The two COUNT reasons are about the whole declaration rather than about one
pair, which is why they can only come from the declaration verbs and never
from a per-pair legality query. `card_id` is a blocker named in the refused
declaration when there is one — a count broken by a blocker being re-pointed
AWAY names none, because no creature in that declaration is the problem.

#### `attack_tax_unpaid` (#1063)

Emitted when a `declare_attacker` or `declare_attackers` is refused
because the attacking player cannot pay the CR 508.1a attack tax the
declaration owes — Propaganda, Ghostly Prison, Sphere of Safety
([ADR 0080](decisions/0080-attack-taxes.md)).

Nothing was declared: no creature is tapped, no `EventAttack` fired,
nothing was spent, and the board is exactly as it was. `reason` is the
declaration's whole price as a cost string (`"{2}"`, `"{2}{2}"`) and
`missing` is the symbols the mana pool and the auto-tapper together
could not cover. `card_id` is absent — the refusal is about the
DECLARATION, not about one creature, and a bulk swing has no single
card to point at.

**There is no override.** `insufficient_mana` offers `force_cast: true`
because a sandbox table may track its mana on paper; an attack tax
waived is the enchantment played as a blank, which the engine will not
do. The player's move is to attack with fewer creatures — a smaller
batch is priced smaller — or to make more mana.

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
| `move_card` | no | `{ "src": <ZoneRef>, "dst": <ZoneRef>, "instance_id": "<uuid>", "as_commander"?: bool, "to_bottom"?: bool }` | General-purpose zone-to-zone move. Clears tapped state, counters, marked damage and the deathtouch mark if leaving the battlefield (CR 400.7). Caller must control the card. S13.1: when the card is a commander AND `dst` is graveyard / exile / hand / library, its owner is asked whether to put it in the command zone instead (CR 903.9 commander zone replacement, exercised as an explicit player choice rather than an automatic engine transform) — nothing moves until that prompt is answered, and #707 made the move complete afterwards from whatever zone the card was in; `as_commander` is a routing flavour on the request, not a gate on the offer. #170: when `to_bottom` is true the card is seated at the BOTTOM of the destination library instead of the top — a no-op if a replacement rewrote the destination, or if the destination is not a library. |
| `tap` | no | `{ "instance_id": "<uuid>" }` | Sets a battlefield card's tapped state to true. Caller must control the card (admin sessions bypass) — see "controller-only card actions" below. |
| `untap` | no | `{ "instance_id": "<uuid>" }` | Sets a battlefield card's tapped state to false. Same controller gate as `tap`. |
| `untap_all` | yes | — | Untaps all of `player`'s cards on the battlefield. S13: server-side no-op when called during the active player's `untap` step (the step-entry hook has already untapped automatically). Outside that window the action behaves as before — sandbox / replay support. |
| `pass_priority` | no | — | Rotates priority to the next non-eliminated seat. When priority would wrap back to the active seat, the step auto-advances and `priority_holder` resets per the new step's rules. S13: skips eliminated seats during the rotation; rejects with `bad_request` ("no player holds priority this step") on `untap` / `cleanup` (those don't grant priority per CR 502.4 / 514.3). Stack-aware semantics (priority resets on spell resolution) arrive with S13.1. |
| `pass_turn` | no | — | Ends the current turn at once and jumps to the next seat's turn. Attackers and blockers leave combat and the cleanup sweep runs (marked damage removed, "until end of turn" effects end); the steps in between are skipped, so there is no end step and no discard to hand size. Lands on Untap with `priority_holder = -1`; the step-entry hook auto-untaps and walks the cursor on to Upkeep before the next snapshot is broadcast. |
| `mulligan` | yes | `{ "hand_size": <int> }` | Shuffles `player`'s hand back into their library and draws `hand_size` cards. Simplified London mulligan — no card-to-bottom penalty. |
| `shuffle_library` | yes | — | Reshuffles `player`'s library. |
| `change_life` | yes | `{ "delta": <int> }` | Adjusts `player`'s life by `delta` (positive for gain, negative for loss). |
| `add_counter` | no | `{ "instance_id": "<uuid>", "name": "<string>", "delta": <int> }` | Modifies a named counter on a card. Delta ≤ 0 that drives the counter to zero removes the entry. Caller must control the card. |
| `set_commander_damage` | no | `{ "from": "<uuid>", "to": "<uuid>", "amount": <int> }` | Sets total commander damage dealt from `from`'s commander(s) to `to`. Set semantics, not additive. |
| `set_battlefield_position` | no | `{ "instance_id": "<uuid>", "x": <float>, "y": <float> }` | Stamps a normalised (x, y) position in `[0, 1]` on a battlefield card. Server clamps out-of-range inputs rather than erroring. Target must be on the battlefield — other zones return `card_not_found`. Caller must control the card. Added in S06; controller gate added in S08.5. |
| `concede` | yes | — | Marks the calling player as eliminated and advances the turn cursor past them if they were the active seat. When exactly one non-eliminated seat remains, the game's `state` transitions to `ended` (the survivor is the implicit winner — no separate field). Idempotent calls return `bad_request` with "player is already eliminated". Added in S08. |
| `keep_hand` | yes | — | Commits the calling player to their current opening hand during the mulligan window. Once every seated, non-eliminated player has called `keep_hand`, `mulligans_open` flips to false. Idempotent (no-op if already kept). Added in S08. |
| `declare_attacker` | no | `{ "attacker": "<uuid>", "target": "<uuid>" }` | Marks a battlefield card as attacking the target player. Step-gated to `declare_attackers`; rejects with `bad_request` outside that step. Attacker must be a creature (`type_line` contains "Creature"); rejects with `bad_request` otherwise. Re-declaring the same attacker against a different target overwrites. Caller must control the attacker (admin sessions bypass) — see "controller-only card actions" below. Added in S08; controller gate added in S08.5. Since [ADR 0080](decisions/0080-attack-taxes.md) it also carries the CR 508.1a payment trio — see the attack-tax paragraph below the table. |
| `declare_attackers` | no | `{ "attackers": [{ "attacker": "<uuid>", "target": "<uuid>" }, ...] }` | Declares a whole attacking set in ONE action — the wire verb behind the client's "attack with all" cluster (#318). Step-gated to `declare_attackers`. Each entry is checked independently and **skipped silently** when the creature is unknown, is not a creature, is tapped, is summoning-sick (CR 302.6), has `defender` (CR 702.3), is already declared this combat, or names a seat that is the creature's own controller / eliminated / nonexistent — one ineligible creature must never sink a wide declaration. Rejects with `bad_request` when the batch is empty, exceeds 256 entries, contains a malformed UUID, or when EVERY entry was skipped. Caller must control every listed creature; a foreign creature rejects the whole batch (authorization is not part of the skip contract). Declared creatures tap unless they have `vigilance` (CR 508.1f / 702.20), emit one `EventAttack` each, and drain through a single state-check pass so simultaneous attack triggers reach the stack together (CR 508.1 / 508.2). Prefer this over looping `declare_attacker`: each action is one undo entry and one broadcast snapshot, so N creatures would otherwise cost N of both against a per-turn undo budget of 1. Added in S31. Since [ADR 0080](decisions/0080-attack-taxes.md) there is ONE exception to the silent-skip contract — the CR 508.1a attack tax is all or nothing; see below.
| `declare_blocker` | no | `{ "blocker": "<uuid>", "attacker": "<uuid>" }` | Marks a battlefield card as blocking the named attacker. Step-gated to `declare_blockers`. Blocker must be a creature; attacker must exist on the battlefield. Caller must control the blocker. Exactly a one-entry `declare_blockers`, so since #750 it can also be refused for a block COUNT: one creature is not a legal block on a menace attacker and comes back as `illegal_block` / `too_few_blockers` with nothing stored. Send `declare_blockers` for a block that needs several creatures. Added in S08; controller gate added in S08.5; count refusal in S37. |
| `declare_blockers` | no | `{ "blocks": [{ "blocker": "<uuid>", "attacker": "<uuid>" }, ...] }` | Declares a whole block in ONE action (#750, [ADR 0045](decisions/0045-combat-restrictions.md) addendum Decision 13). Step-gated to `declare_blockers`. Exists for a RULES reason, not as a batching convenience: a block COUNT (menace's minimum of two, Hungering Hydra's maximum of one) is a property of the whole declaration, so a two-creature menace block is legal only as a pair and **cannot** be sent as two `declare_blocker` actions — the first would be refused with `too_few_blockers`. **All or nothing**, unlike `declare_attackers`: the set is validated as it will be after the action (each entry's pair legality, then the count bounds for every attacker whose blocker set the action changes, including one that LOSES a re-pointed blocker), and if anything is refused nothing is stored, nothing is announced and the first refusal comes back as `illegal_block`. Rejects with `bad_request` when the batch is empty, exceeds 256 entries, or contains a malformed UUID. Caller must control every listed blocker; a foreign creature rejects the whole batch. An entry naming a creature that is already blocking RE-POINTS it. Like `declare_blocker` the verb only STAGES the pairings — nothing is announced until the declaration is locked in at the first priority boundary in the step (#830). Added in S37. |
| `clear_combat` | no | — | Resets `attacking_target` and `blocking_target` on every battlefield card. Not step-gated (escape hatch). Added in S08. |
| `advance_step` | no | — | **Passes priority until the current step ends** (#914, CR 117.4). Sandbox affordance — does NOT require the caller to hold priority (cf. `pass_priority` which does). Used by the "next step" / "done" buttons so the active player can drive progression without waiting for opponent priority passes. With an EMPTY stack this is one move of the turn cursor and no pass, as it has always been. With anything on the stack it is the passes the rules require first: a step cannot end with objects on the stack, so the server resolves them (each resolution followed by state-based actions, the trigger drain and another priority round) and only then moves the cursor. The drive stops with the cursor unmoved and **no error** — the caller reads the broadcast state — when one of those resolutions raises a blocking prompt, when the CR 726 loop notice goes up, or when the game ends; a prompt that was already open when the action arrived is still refused with `ErrChoicePending` (see the choice gate below). Auto-resolves combat damage on entry to a combat damage step (same hook as `pass_priority`): first-strike damage on entry to `first_strike_damage`, regular damage on entry to `combat_damage`. A step this turn does not have is walked through, not entered — so an advance out of `declare_blockers` lands on `combat_damage` in a combat with no first or double strike in it, and on `first_strike_damage` in one that has (#717). Added in S08; the priority drive in S37. |
| `set_monarch` | optional | — | Designates `player` as the monarch (Conspiracy mechanic). Empty / omitted `player` clears the marker. Any seated player may flip the marker — sandbox does not enforce the must-attack-when-able or combat-damage transfer rules. Added in S10. |
| `set_initiative` | optional | — | Designates `player` as holding the initiative (BG3 mechanic). Empty / omitted `player` clears. Same posture as `set_monarch`. Added in S10. |
| `set_goaded` | no | `{ "instance_id": "<uuid>", "by": "<uuid>" }` | Marks a battlefield creature as goaded by the named player. Empty `by` clears the goad. Sandbox marker; the must-attack-not-the-goader rule is not enforced. Added in S10. |
| `set_poison` | yes | `{ "amount": <int> }` | Sets `player`'s poison counter total. Set semantics (not increment) so two stale tabs don't double-count. Clamped at 0 from below. Added in S10. |
| `set_energy` | yes | `{ "amount": <int> }` | Sets `player`'s energy counter total. Same shape and clamp posture as `set_poison`. Added in S10. |
| `set_promise` | no | `{ "from": "<uuid>", "to": "<uuid>", "count": <int> }` | Sets the per-pair "I owe you" promise-token count from `from` → `to`. Set semantics; clamped at 0. Politics scaffold — visual reminder only. Added in S10. |
| `start_vote` | yes | `{ "topic": "<string>", "options": ["<string>", …] }` | Opens a vote with `player` as initiator. At least 2 options required. Rejects with `bad_request` if a vote is already open (clients must `end_vote` first). Added in S10. |
| `cast_vote` | yes | `{ "option": <int> }` | Records / overwrites `player`'s ballot at the given option index. Re-voting overwrites. Added in S10. |
| `end_vote` | no | — | Closes the currently open vote. Any seated caller may end any vote (sandbox). Added in S10. |
| `undo` | no | — | Pops the room's most recent pre-mutation snapshot off its undo stack and restores the game state. Bumps `seq` like a normal action so clients see a regular `snapshot` frame. Authorization: a seated caller may only pop an entry whose stored caller matches their own seat (rewinding an opponent's move requires their cooperation — they undo first). Each successful seated undo debits the caller's `Player.UndosRemaining` (refreshed to `Game.Settings.UndoLimit` on entering their untap step; never debited when the limit is `-1`, unlimited — [ADR 0075](decisions/0075-table-settings-and-host-controls.md)). Admin (`?player=` omitted) bypasses both the caller and budget gates. Stack capped at 32 per room. No redo at v1. Errors: `bad_request` for empty stack ("nothing to undo"), cross-player request ("you can only undo your own most recent action"), or exhausted budget ("no undos remaining this turn"). Added in S11. |
| `set_table_settings` | no | a settings patch: `{ "undo_limit"?: <int>, "undo_scope"?: "own" \| "host_any", "starting_life"?: <int>, "commander_damage"?: <int>, "bot_pace"?: "fast" \| "normal" \| "slow", "allow_spawn"?: bool }` | Changes the table's settings ([ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.3). The params object **is** the patch: only the fields present are applied, and a field left out is untouched. **Host or admin only** — the server compares the connection against the table's effective host (`is_host` on the seat, `ws.Room.CanManageTable`) and refuses anyone else with `bad_request` ("only the table host or the server admin may change table settings"). Valid in the lobby and mid-game alike. The whole patch is validated before any of it is applied, so one bad field changes nothing: ranges are `undo_limit` ≥ `-1` (`-1` = unlimited, `0` = no undos), `starting_life` 1..999, `commander_damage` 1..99, and the two enums as listed. **`starting_life` after `start` is rejected** (`bad_request`, "starting life cannot change once the game has started") — the value has already been applied to every seat; use `change_life` to adjust totals. An `undo_limit` change refreshes every seat's `undos_remaining` to the new value immediately; a `commander_damage` change is read at the next state-based action check, so lowering it can lose somebody the game at that check. **Not undoable**: the change pushes no entry onto the room's undo stack, and a later `undo` carries the settings forward rather than rolling them back — otherwise lowering the undo limit could be taken back with the undo it was meant to stop. Emits one `settings_changed` event per field that actually changed, and one `settings` log line per event. Deliberately absent from `legal_moves`: a bot never hosts and never changes a table's rules. The HTTP twin is `PATCH /games/{id}/settings` ([docs/lobby.md](lobby.md)). Added in S35. |
| `set_undo_limit` | no | `{ "limit": <int> }` | **DEPRECATED since S35** — a one-field alias for `set_table_settings` (`{"undo_limit": n}`), kept for old clients and `gamecli` scripts. Sets the table's undo limit (`Game.Settings.UndoLimit`, the per-player per-turn undo budget) and immediately refreshes every seat's `UndosRemaining` to the new value. Default is 1. Active games only. **Host or admin only** since [ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.3 — it was any seated player until S35, which is the behaviour change that ADR exists for. Shares the alias's non-undoable apply path. Negative values clamp to 0 (disables undo): the limit `-1` means **unlimited**, and this legacy verb deliberately cannot select it — unlimited is set through `set_table_settings`. A limit of 0 is kept as written; the game no longer rewrites 0 to the default at start. Emits one `settings_changed` event when the value changes. Added in S11. |
| `cast_spell` | yes | `{ "instance_id": "<uuid>", "from_zone"?: "hand" \| "command", "targets"?: [...], "modes"?: [int...], "x_value"?: int, "distribution"?: { "<uuid>": int }, "hold_priority"?: bool, "split_second"?: bool, "strict"?: bool, "force_cast"?: bool, "auto_tap"?: bool, "locked_sources"?: ["<uuid>", ...] }` | S13.1 canonical "play a card from hand" verb (CR 601). Lands route to battlefield (CR 305 special action); other types go to the stack with a fresh metadata entry, the caster retains priority (CR 117.3c). Sorcery-speed gate (sorceries + lands) requires main phase + stack empty + caller is active. Casting from `command` increments the per-commander tax counter (CR 903.8). Caller must hold priority. Targets are `{ "kind": "player" \| "card" \| "self" \| "none", "id"?: "<uuid>" }`. S14: if the card is in the effect catalog with a declared `CardView.target_mode`, the client enters a targeting prompt before sending `cast_spell`; the server re-validates at resolve per CR 608.2b ("countered by game rules" fizzle if every targeted slot is illegal at resolution time). **S15:** `strict` engages the mana-cost gate — server parses `ManaCost` (plus commander tax), rejects with `insufficient_mana` error frame when the pool falls short, deducts on success. Sourced from the client's `gameplay.strictMana` setting. `force_cast` overrides the gate for one cast (the "Cast anyway" toast button) — proceeds permissively without touching the pool. `auto_tap` runs the auto-tapper before the gate, taps the planned permanents, and drops produced mana into the pool — all atomic under one write lock. `locked_sources` excludes specific permanents from the auto-tapper's plan (the modal's lock-tap UI). Plan failure under `auto_tap: true` returns `insufficient_mana` before any cards tap (all-or-nothing). **S21 / #747:** `discard_ids`? and `sacrifice_ids`? pay a catalog card's additional cost (CR 601.2f): `sacrifice_ids` names exactly `additional_cost.sacrifice_options.max` distinct permanents; see **Sacrifice costs of N permanents** below. **#500:** a land play past the seat's per-turn allowance (CR 305.2) is REJECTED with `bad_request` ("you've already played all the lands you can this turn"). The allowance is not always one — read `PlayerView.land_drops_per_turn` and `PlayerView.lands_played_this_turn` and disable the play before the click rather than only explaining the refusal. **#764:** a target entry may carry `slot` (which target CLAUSE it answers) and `mode` (which entry of `modes` — the mode OCCURRENCE — whose clause list that is). Both default to 0, which is what every single-clause non-modal cast has always meant, and a client that omits them keeps working: the server fills them in by walking the clauses in printed order. Send them when the card has more than one clause (`CardView.clauses`) or when more than one chosen mode targets, because that is the only way to say which pick answers which question. Modes may now REPEAT when `CardView.modes.repeatable` (CR 700.2d, Mystic Confluence) and every chosen bullet may target (CR 700.2c); the order of `modes` is the order they resolve in and the order their targets are asked for, so it is **not** sorted. **S45 / #1330 (CR 702.172a, Spree):** a mode option may carry its own additional cost — `CardView.modes.options[i].cost`, brace notation, present only on a bullet that prints one ("+ {1}{U} — Counter target spell."). Choosing that bullet adds its cost to the total the cast owes (CR 601.2f), on top of the printed cost and every OTHER chosen bullet's — there is no separate announcement for it, the choice is already `modes`. `activated_abilities[i].modes.options[i].cost` mirrors it for the shape's sake, but no printed activated ability charges more for choosing a mode, so it is always absent there today. **#787 / #916 (CR 107.4c/f, CR 601.2b):** `phyrexian_life`?: int is how many of the cost's Phyrexian mana symbols — `{U/P}`, and CR 107.4's ten hybrid Phyrexian symbols `{W/U/P}` … `{G/U/P}` — are paid with **2 life each** instead of mana. It is announced with the cast because CR 601.2b makes "how do you intend to pay each hybrid and Phyrexian symbol" part of announcing the spell; absent (0) pays every symbol with its coloured half. `CardView.phyrexian_symbols` is the ceiling (how many the printed cost prints) and `alternative_costs[i].phyrexian_symbols` the ceiling when that offer is claimed instead, so a client sizes its stepper without parsing a mana string. The engine strikes the symbols a life payment can actually save first, refuses a count larger than the cost prints and a payment CR 119.4 forbids (only down to 0) with `bad_request` before anything is paid, pays the life through the one cost-shaped life path before spending the pool, and plans the auto-tap for the mana half only. `activate_ability` carries the identical field for an activated ability's mana component (#917). |
| `counter_spell` | no | `{ "instance_id": "<uuid>", "to_zone"?: <ZoneRef> }` | Removes a spell item from the stack and routes its card to `to_zone` (default = owner's graveyard). Covers Counterspell (default), Hinder (`to_zone = library`), Remand (`to_zone = hand`), exile-bound counters. Battlefield + stack rejected as destinations (`bad_request: invalid stack-counter destination`). Caller must hold priority. Added in S13.1. |
| `counter_ability` | no | `{ "instance_id": "<uuid>" }` | Removes an activated / triggered ability item from the stack. Abilities cease to exist on removal (CR 608.2n); no destination needed. Caller must hold priority. Added in S13.1. |
| `activate_ability` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"?, "ability_index"?: int, "sacrifice_ids"?, "crew_ids"?, "counter_source_ids"?, "counter_counts"?, "counter_kind"?, "counter_kinds"?, "discard_ids"?, "exile_ids"?, "return_ids"?, "waterbend_ids"?, "tap_ids"?, "strict"?, "auto_tap"? }` | Creates an activated-ability stack item linked to the source card. Same announce-time params as `cast_spell` (minus the source-zone routing). The source card stays in its origin zone; the stack item carries a synthetic ID. Mana abilities are NOT modeled this way (CR 605 — they don't use the stack). Caller must hold priority. Added in S13.1. **With `ability_index`** the payload names a catalog ability (S21 sub-PR 2) and the engine validates and pays the cost: `sacrifice_ids` pays a "Sacrifice a creature" or "Sacrifice two artifacts" component (exactly `activated_abilities[i].sacrifice_options.max` distinct permanents; see **Sacrifice costs of N permanents** below), `crew_ids` pays a crew number (CR 702.122a), and `x_value` is the X announced for an `{X}` in the ability's mana component (CR 602.2b). X is validated at announce — non-negative, zero unless the cost has an `{X}`, and at least the cost's printed floor — then locked onto the stack item, where the ability's effect reads it back. The card's `activated_abilities[i]` carries `demands_x`, `min_x` and `x_slots` so the client knows to prompt, what floor to enforce, and what a given X actually costs. **#625:** `counter_source_ids` (one `<uuid>`) and `counter_kind` pay a "remove N counters" component. The ability's view carries `counter_cost_n` (present means the ability has the component), `counter_cost_kind` (absent means "a counter" of any kind), `counter_cost_self` (the counters come off the source), `counter_cost_label` (the "from" clause, e.g. "a planeswalker you control") and `counter_cost_options: [{ "card_id": "<uuid>", "kinds": [{ "kind": string, "count": int }] }]`. The options list the permanents the viewer controls that could pay right now, most counters first. They come from the non-targeting candidate walk, so hexproof permanents are included, and the field is absent when nothing can pay. Send `counter_source_ids` for the other-permanent form. Omit it for the self form, or send the source's own ID. Send `counter_kind` for the any-kind form. It is optional for a printed kind, and if sent it must match that kind. The removal is paid at announce with the rest of the cost and is never modified by counter replacement effects. It is not a loyalty activation and has no timing restriction of its own. Errors: `not enough counters to pay that cost` when the named permanent holds too few, `you do not control that card` for another player's permanent, and `bad_request` for a payload that names a counter choice the ability has no component for. **#743:** an ability with an activation condition (CR 602.1b: "Activate only if an opponent controls four or more lands", "Activate only during your turn") is refused with `ability's activation condition is not met` while the condition is false, before X, targets or any cost are checked, so nothing is paid. The condition is checked only at activation, never at resolution. When sorcery speed and the condition both fail, the sorcery-speed error is the one returned. **#789:** the counter component grew three more printed shapes, and one payment shape covers all of them. `counter_counts`?: `[int, ...]` is the per-permanent split, parallel to `counter_source_ids`: omit it for a fixed one-permanent cost (the #625 shape, unchanged), send counts totalling exactly `counter_cost_n` for a removal spread across permanents (`counter_cost_among: true` — "Remove two +1/+1 counters from among artifacts you control"), and send the announced count for a removal whose size the activator chooses (`counter_cost_variable: true` — "Remove any number of storage counters"; `counter_cost_n` is then the FLOOR and `counter_cost_max` the most the viewer could name right now). A split payment is validated as a set, like a crew payment: every permanent distinct, controlled by the activator, matching the clause, holding at least the count named against it, and the counts totalling exactly N — any failure refuses the whole activation with no counter removed. **#943:** the last printed shape is an among removal of ANY kind (`counter_cost_among: true` with no `counter_cost_kind` — Tekuthal, Inquiry Dominus' "Remove three counters from among other artifacts, creatures, and planeswalkers you control"), where the parts may differ in kind as well as in count. `counter_kinds`?: `[string, ...]` is the per-permanent kind, parallel to `counter_source_ids`: send it only when the payment actually mixes kinds, and `counter_kind` (one kind for the whole payment) whenever one kind will do — so a payment of one kind is byte-for-byte the shape earlier clients send. If both are sent they must agree, an empty kind in the array is refused, and the set's identity is `(permanent, kind)`: one permanent may appear twice with two different kinds (it pays in both) and never twice with the same one. Nothing new is projected for the shape — a `counter_cost_options` row has always been a (permanent, kind) pair, so the existing option list already answers "which kinds, off which permanent". The other direction is `counter_cost_add` / `counter_cost_add_kind`, a cost that PUTS counters on the source (Devoted Druid's "put a -1/-1 counter on this creature"): nothing is chosen, so there is nothing to send, and `counter_add_blocked: true` means CR 118.3 refuses the activation because the permanent cannot have those counters. Such a cost is never doubled by a counter-doubling replacement — such a replacement applies only to a counter placed by an effect (CR 614.16). `activated_abilities[i].condition_unmet: true` marks an ability whose condition is false right now. It is absent when the ability has no condition or the condition holds, is evaluated with the permanent's controller as "you", and is sent to every viewer (a condition reads only public information). The client greys the row, as it does for `sorcery_speed`. **#917 (CR 107.4f / CR 602.2b):** `phyrexian_life`?: int is how many of the ability's Phyrexian mana symbols are paid with **2 life each** instead of mana — Birthing Pod's `{1}{G/P}`, Solphim's `{1}{R/P}{R/P}`. It is the same field name, meaning and validation `cast_spell` carries, because CR 602.2b asks the activator exactly what CR 601.2b asks the caster and the engine runs one strike-and-pay helper for both. Absent (0) pays every symbol with its coloured half. `activated_abilities[i].phyrexian_symbols` is the ceiling — how many symbols the cost prints — so the client never parses a mana string to size its stepper. The engine strikes the symbols a life payment can actually save first, refuses a count larger than the cost prints, refuses a payment CR 119.4 forbids (a player may pay life only down to 0), and refuses a claim against an ability with no mana component at all; every refusal is `bad_request` with nothing paid. The life is paid through the one cost-shaped life path before the pool is spent, the auto-tapper plans only the mana the activation still owes, and `PaidCost.LifePaid` on the stack item records the printed `Life` component and this together. **#660, widened by #1221:** an ability may function from a zone other than the battlefield (CR 113.6). Cycling and typecycling (CR 702.29a/e) function from HAND, unearth (CR 702.82a), scavenge (CR 702.96a), embalm (CR 702.128a) and eternalize (CR 702.129a) from a GRAVEYARD, and any of them may declare exile or the command zone; all of them reach the client on the card's `zone_abilities` rather than on `activated_abilities` (the field was `hand_abilities` through #660, and was renamed rather than duplicated when the other zones opened); `ability_index` is the ability's index in the card's full list either way, so the payload is unchanged. Activating an ability from a zone it does not function in is refused with `this ability cannot be activated from that zone`, before anything is validated and before anything is paid — a cycling fired at a permanent and a sacrifice outlet fired at a card in hand get the same error. Off the battlefield the card's OWNER is the only player who may activate it (CR 108.4). `discard_ids` pays a discard component: the general "Discard a creature card" clause (`discard_cost_n` / `discard_cost_label` / `discard_cost_options` on the ability's view) names exactly `discard_cost_n` distinct cards from the viewer's own hand, each matching the clause, and none of them the source of a hand activation. Cycling's "Discard this card" (`discard_self: true`) sends NOTHING — the source is the payment — and a payload that names cards for an ability with no discard component is refused. The discard is paid at announce, with the rest of the cost, through the engine's one discard path, so `EventDiscardCard` fires per card and every discard payoff sees it; being a cost it cannot pause, so a commander pitched to one goes to the graveyard rather than opening a CR 903.9 prompt (CR 601.2h / 602.2b, [ADR 0062](decisions/0062-abilities-and-special-actions-from-the-hand.md)). **#1221:** scavenge, embalm and eternalize pay a second self-cost, `exile_self: true` on the ability's view — "Exile this card from your graveyard". Like `discard_self` it is ADVISORY and sends nothing: the source IS the payment, so there is no pick and no field on the payload, and the keyword's own label spells the clause out. The card is exiled at announce, through the same exit every other move takes and with the same cost-cannot-pause bit, and the ability's effect reads the card back out of exile afterwards. An ability that declares the component without declaring the graveyard is refused at server boot, not at activation. **#1297:** `exile_ids`?: `["<uuid>", ...]` pays an "Exile N cards from your graveyard" / "Exile a card from your hand" component — Grim Lavamancer's "Exile two cards from your graveyard", Moorland Haunt's "a creature card", Holistic Wisdom's "a card from your hand" — exactly `activated_abilities[i].exile_cost_n` distinct cards from `exile_cost_options`, none of them the source and none also named in `discard_ids`; see **Exile-N-cards costs** below. It is the mana ability's `exile_ids` (#1283) with a second owner. **#1208:** `activated_abilities[i].timing_closed: true` says the engine will refuse this activation RIGHT NOW for timing (CR 602.5d's "activate only as a sorcery", CR 606.3's loyalty window), as narrowed or OPENED by any per-player statement on the battlefield — The Wandering Emperor's "as long as she entered this turn, you may activate her loyalty abilities any time you could cast an instant", Leonin Shikari's "you may activate equip abilities any time you could cast an instant". It is the stamp of the one read `activate_ability` and the bot's move enumerator also ask, and it is the field a client greys the row on: `sorcery_speed` beside it is only the ability's PRINTED clause and does not change when a statement opens the window. Absent means the engine has no timing objection, which is every instant-speed ability at every moment; it is never stamped on a mana ability, because CR 605.3a gives one its own window and no timing statement reaches it (a card that stops mana abilities is a BAN and arrives as `cant_activate`). Evaluated once with the permanent's controller as "you" and sent to every viewer, exactly as `condition_unmet` is. |
| `activate_loyalty` | yes | `{ "planeswalker_id": "<uuid>", "label"?: string, "delta": int }` | Applies a planeswalker loyalty ability. Sandbox shape: the engine doesn't model the activation as a proper stack item (deferred to S14+) — `delta` is applied immediately to the planeswalker's `loyalty` counter. Sorcery-speed gated; once-per-turn-per-planeswalker (CR 606.3). Caller must hold priority. Added in S13.1. |
| `announce_trigger` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"? }` | Queues a triggered ability for APNAP-ordered drain onto the stack (CR 603.3b). The drain happens at every priority-grant boundary inside the SBA + trigger loop. Sandbox: the player whose card has the trigger announces it manually — auto-fire from card events is the bright line between S13.1 and S14+. Added in S13.1. |
| `mark_damage` | no | `{ "instance_id": "<uuid>", "delta": int }` | Adjusts the damage noted on a creature on the battlefield by `delta` (positive to add, negative to remove). Drives the lethal-damage SBA (CR 704.5g). Positive deltas require controller (admins / spectators bypass); negative deltas are caller-loose so opponents can undo. SBA loop fires after the mutation. Cleanup-step turn-based action zeros every creature's marked damage (CR 514.2). Added in S13.1. |
| `add_player_counter` | yes | `{ "name": string, "delta": int }` | Modifies a named player-level counter (poison, energy, experience, rad, plus homebrew names) by `delta`. Negative deltas clamp at zero. Drives the SBA loop after the mutation so 10 poison or future counter-driven losses fire immediately. The legacy `set_poison` / `set_energy` actions stay synchronised with the underlying `Player.Counters` map. Player-scoped: a seated caller may only adjust their own counters; admins bypass. Added in S13.2. |
| `discard_selection` | yes | `{ "card_ids": ["<uuid>", ...] }` | Resolves the cleanup-step interactive discard prompt (S13.4, CR 402.2). Caller must be in `discard_pending`; the count must exactly match the over-max amount; every supplied card ID must live in the caller's hand. Selected cards move to the caller's graveyard, the caller's pending entry clears, and the cleanup hook re-fires so the cursor resumes its auto-advance. Idempotent no-op for callers not in the pending map. Player-scoped. Added in S13.4. |
| `set_max_hand_size` | yes | `{ "value": int }` | Sets the named player's per-player cleanup-step hand-size cap (CR 402.2). Default 7; `-1` disables the cap (Reliquary Tower / Thought Vessel). Sandbox helper until the S14+ effect catalog wires this to real cards via the S16 layer pipeline. Player-scoped. Added in S13.4. |
| `resolve_choice` | yes | `{ "choice_id": "<uuid>", "card_ids"?: ["<uuid>", ...], "color"?: "W" \| "U" \| "B" \| "R" \| "G" \| "C", "apply"?: bool, "tap_ids"?: ["<uuid>", ...], "order"?: ["<id>", ...], "assignments"?: [{ "blocker_id": "<uuid>", "amount": int }, ...], "trample_to_player"?: int }` | Resolves a pending async effect-driven choice (S14, CR 701.9b — "player chooses" discard). Caller must be the `chooser` of the named entry in `GameView.pending_choices`; `card_ids.length` must equal the entry's `count`; every supplied card ID must currently live in the entry's `from_player`'s hand. Picked cards move to `from_player`'s graveyard (for `discard_from_hand`), the entry drains from the queue, and an `EventDiscardCard` event fires per card. Distinct from `discard_selection` (S13.4 cleanup, chooser == discarder); `resolve_choice` covers effects where chooser != discarder (Thoughtseize — caster picks from target's revealed hand). **S15:** the entry's `kind` may be `mana_pick` (Birds of Paradise / Arcane Signet); `card_ids` is omitted and `color` carries the chosen single-letter color, validated against the entry's `color_options` slice. The picked color drops into the chooser's mana pool as one `ManaToken`. **S17:** `kind` may be `replacement_order` (CR 616 multi-replacement order prompt — `order` carries the chosen permutation of replacement-effect IDs) or `optional_replacement` (a "may" yes/no — `apply` is the answer). **#1397:** an `optional_replacement` may also be the CR 903.9 question asked BEFORE a cost that moves a commander is paid ([ADR 0013 §5af](decisions/0013-replacement-effects.md#5af-amendment-2026-09-24-a-commander-paid-as-a-cost-is-asked-before-the-payment-not-during-it)). The cast or activation that asked is parked with nothing paid; answering it makes that cast or activation, so the answer can be followed by the spell or ability arriving on the stack. Same payload and same modal; the `chooser` is the commander's OWNER, who may not be the player paying, and `source` is the commander. **S18:** `kind` may be `damage_assignment` (CR 510.1c multi-blocker assignment — `assignments` is the per-blocker amounts in chosen order, `trample_to_player` is the overflow when the attacker has trample). **S19:** `kind` may be `trigger_prompt` (CR 603.5 "you may" optional triggered ability — `apply` is the answer; the server routes by inspecting the choice's kind, since `optional_replacement` and `trigger_prompt` share the `apply` payload). **S26:** `kind` may be `choose_creature_type` (CR 614.12 "as this permanent enters, choose a creature type" — Cavern of Souls, Door of Destinies, Vanquisher's Banner, Adaptive Automaton); `creature_type` carries one name from the CR 205.3m vocabulary, validated and normalised server-side and stamped onto the source permanent's named tribe. **#742:** `kind` may be `choose_color` (CR 105.4 "choose a color" — as a permanent enters, Coldsteel Heart / the Thriving lands, where the answer is stored on the permanent; or while a spell or ability resolves, Wash Out / Oona). It is answered with the same `color` field a `mana_pick` uses, validated against the entry's `color_options` (a subset of W/U/B/R/G, never C), and the dispatcher routes `color` **by the pending choice's `kind`**: the presence of `color` alone no longer says which resolver an answer belongs to. A `mana_pick` entry that carries `color_amounts` mints that many tokens of the picked colour instead of one. Added in S14. **#764:** `kind` may be `mode_pick` (CR 603.3c — a modal TRIGGERED ability's bullet, chosen as the ability is put on the stack, after the "you may" prompt and before the CR 603.3d target pick). It is answered with `modes`: the chosen ModeSpec indexes **in the order chosen**, repeats allowed only when `mode_repeatable`. The prompt carries `mode_options` (the oracle bullets), `mode_indexes` (the ModeSpec index each label belongs to — a bullet with no legal target is dropped server-side, so the row index is not the answer), `mode_min`, `mode_max` and `mode_repeatable`. The dispatcher routes this kind **by the choice's `kind`**, like `coin_call` and `loop_shortcut`: an index list of zeroes (`[0, 0, 0]` is Mystic Confluence drawing three cards) and the empty list are both ordinary answers, so there is no presence to route on. |
| `resolve_choice` | yes | `{ "choice_id": "<uuid>", "card_ids"?: ["<uuid>", ...], "color"?: "W" \| "U" \| "B" \| "R" \| "G" \| "C", "apply"?: bool, "tap_ids"?: ["<uuid>", ...], "order"?: ["<id>", ...], "assignments"?: [{ "blocker_id": "<uuid>", "amount": int }, ...], "trample_to_player"?: int }` | Resolves a pending async effect-driven choice (S14, CR 701.9b — "player chooses" discard). Caller must be the `chooser` of the named entry in `GameView.pending_choices`; `card_ids.length` must equal the entry's `count`; every supplied card ID must currently live in the entry's `from_player`'s hand. Picked cards move to `from_player`'s graveyard (for `discard_from_hand`), the entry drains from the queue, and an `EventDiscardCard` event fires per card. Distinct from `discard_selection` (S13.4 cleanup, chooser == discarder); `resolve_choice` covers effects where chooser != discarder (Thoughtseize — caster picks from target's revealed hand). **S15:** the entry's `kind` may be `mana_pick` (Birds of Paradise / Arcane Signet); `card_ids` is omitted and `color` carries the chosen single-letter color, validated against the entry's `color_options` slice. The picked color drops into the chooser's mana pool as one `ManaToken`. **S17:** `kind` may be `replacement_order` (CR 616 multi-replacement order prompt — `order` carries the chosen permutation of replacement-effect IDs) or `optional_replacement` (a "may" yes/no — `apply` is the answer). **#1397:** an `optional_replacement` may also be the CR 903.9 question asked BEFORE a cost that moves a commander is paid ([ADR 0013 §5af](decisions/0013-replacement-effects.md#5af-amendment-2026-09-24-a-commander-paid-as-a-cost-is-asked-before-the-payment-not-during-it)). The cast or activation that asked is parked with nothing paid; answering it makes that cast or activation, so the answer can be followed by the spell or ability arriving on the stack. Same payload and same modal; the `chooser` is the commander's OWNER, who may not be the player paying, and `source` is the commander. **S18:** `kind` may be `damage_assignment` (CR 510.1c multi-blocker assignment — `assignments` is the per-blocker amounts in chosen order, `trample_to_player` is the overflow when the attacker has trample). **S19:** `kind` may be `trigger_prompt` (CR 603.5 "you may" optional triggered ability — `apply` is the answer; the server routes by inspecting the choice's kind, since `optional_replacement` and `trigger_prompt` share the `apply` payload). **S21:** `kind` may be `scry` (CR 701.22 — `bottom` is the looked-at cards going under the library and `top_order` the ones staying on top, listed TOP-FIRST; every looked-at card must appear in exactly one list, because a scry moves all of them). **S22:** `kind` may be `surveil` (CR 701.25 — the same shape with `graveyard` in place of `bottom`) or `look_at_top` ("look at the top N cards of your library, then put them back in any order" — Ponder, Sensei's Divining Top; answered with `top_order` alone, since there is no away lane). The three are the **scry family** and the dispatcher routes them **by the pending choice's `kind`**, not by the payload shape every other `resolve_choice` branch keys off: all three carry `top_order` and `look_at_top` carries nothing else, so no key's presence identifies them. Send the destination key that matches the kind — `bottom` for scry, `graveyard` for surveil, neither for `look_at_top`. **#996 ([ADR 0088](decisions/0088-ordered-library-placement.md)):** `kind` may be `put_in_library` — "put these cards on top of / on the bottom of the library in any order" (Brainstorm's put-back, "the rest on the bottom of your library in any order", Aetherspouts' "top or bottom"), the family's fourth member and routed the same way. It is answered with scry's two keys, `top_order` and `bottom`, **both TOP-FIRST** (the first entry of `bottom` is the card nearest the top of the pile that goes under the library; the last is the bottom card). The entry's `placement` says which lanes are open — `top` (`top_order` only), `bottom` (`bottom` only) or `top_or_bottom` (both) — and an answer that puts a card in a closed lane is refused. Every card appears in exactly one list. **#1298:** the entry may also carry `top_count` — EXACTLY how many cards `top_order` must hold (Cream of the Crop's "put one of those cards on top"; any other count is refused) — and `top_depth` — where the top lane lands, counted from the top (2 is Temporal Cleansing's "second from the top"; absent is the top). The chooser need not own the cards: Jace, the Mind Sculptor's +2 and Portent order another player's library, and Hinder's counterer picks the end of the countered spell's owner's library. A scry's `bottom` has always been applied in the order sent, and is top-first by the same rule. **#742:** `kind` may be `choose_color` (CR 105.4 "choose a color" — as a permanent enters, Coldsteel Heart / the Thriving lands, where the answer is stored on the permanent; or while a spell or ability resolves, Wash Out / Oona). It is answered with the same `color` field a `mana_pick` uses, validated against the entry's `color_options` (a subset of W/U/B/R/G, never C), and the dispatcher routes `color` **by the pending choice's `kind`**: the presence of `color` alone no longer says which resolver an answer belongs to. A `mana_pick` entry that carries `color_amounts` mints that many tokens of the picked colour instead of one. Added in S14. |
| `activate_mana_ability` | yes | `{ "instance_id": "<uuid>", "ability_idx"?: int, "sacrifice_ids"?, "counter_source_ids"?, "counter_counts"?, "counter_kind"?, "counter_kinds"?, "discard_ids"?, "exile_ids"? }` | S15. Activates the `ability_idx`-th mana ability on a battlefield permanent the caller controls. `ability_idx` defaults to 0 (the only ability for basic lands and most rocks). Single code path for catalog-declared abilities (Sol Ring, Arcane Signet, Birds of Paradise — looked up via the `CatalogManaAbilities` hook) and synthetic basic-land abilities (Forest → `{G}`) the engine derives from `TypeLine`. Pays the tap cost (rejects with `bad_request: "card already tapped"` if the source is tapped), then materialises the produced mana: single-color slots drop straight into the controller's pool as one `ManaToken`; multi-option slots (pipe syntax) queue a `mana_pick` PendingChoice for the controller. **#742:** a pipe option may carry a count (`{W3|U3|B3|R3|G3}`, "three mana of any one color"), which queues ONE `mana_pick` with `color_amounts` rather than one pick per mana. Mana abilities don't use the stack (CR 605.3) — synchronous under the write lock. Caller need NOT hold priority (mana abilities are special, CR 605.1a). Player-scoped: caller must control the source. **S21 / #747:** `sacrifice_ids`?: `["<uuid>", ...]` pays a mana ability's "Sacrifice a creature" component (Ashnod's Altar): exactly `mana_abilities[i].sacrifice_options.max` distinct permanents; see **Sacrifice costs of N permanents** below. **#743:** `mana_abilities[i].condition_unmet: true` is the same flag for a mana ability whose activation condition is false right now (Temple of the False God with four lands, Mox Opal without metalcraft). Same evaluation and visibility as `activated_abilities[i].condition_unmet`; the activation is refused with `ability's activation condition is not met` and nothing is tapped or paid. **#844 (CR 903.4f):** `mana_abilities[i].adds_no_mana: true` marks an ability whose printed text says "any color in your commander's color identity" (Command Tower, Arcane Signet, Commander's Sphere, Path of Ancestry) activated by a controller with no commander, or one whose commander's colour identity is colourless (Kozilek, Karn). That quality is undefined or empty, so the ability adds no mana: the activation is NOT refused (the ability exists and may be activated; it simply does nothing, which is what the rulings on those four cards say), but no mana token is minted and no `mana_pick` is queued. Nothing offers the activation either — the client greys the row on this flag, and the bot's legal-move enumerator and the auto-tapper both skip the source. Absent for every other ability. **#764:** `modes` announces a modal activated ability's bullets (CR 602.2b) in the SAME message as the targets — activating is one indivisible step, so there is no prompt. `activated_abilities[i].modes` is the ModeSpec (same shape as `CardView.modes`) and `activated_abilities[i].clauses` its clause list when it has more than one; the target entries carry `slot` / `mode` exactly as a cast's do. **#789:** a mana ability may carry a COUNTER component, with exactly the field names, meanings and validation `activate_ability` uses — one component with two owners. Vivid Creek's "{T}, Remove a charge counter from this land" sends nothing at all (the self form with a printed kind and count); Mage-Ring Network's "Remove any number of storage counters" sends `counter_counts`. The ability's view carries the same `counter_cost_*` fields an activated ability's does. The auto-tapper only plans such a source when it can both decide and afford the cost — the self form, a printed kind, a fixed count, and enough counters right now — so a Vivid land with no charge counters left is not a five-colour source and is not planned as one. **#943:** `counter_kinds` is accepted here too, with the same meaning — one component, one payment shape, whichever ability kind carries it — though no printed mana ability needs it today. **#1283:** `exile_ids`?: `["<uuid>", ...]` pays an "Exile a card from your hand" component (Cadaverous Bloom) — exactly `mana_abilities[i].exile_cost_n` distinct cards from `exile_cost_options`; see **Exile-a-card costs on a MANA ability** below. |
| `special_action` | yes | `{ "card_id": "<uuid>", "kind": "foretell", "suspend", "plot" or "turn_face_up", "strict"?: bool, "auto_tap"?: bool }` | A CR 116.2 **special action**: a game action taken without using the stack and without passing priority, so there is no announce, no response window and nothing to respond to. Caller must hold priority; the caller keeps it. `kind` selects the action — `foretell` (CR 702.143a: pay {2} and exile the card from your hand face down, castable on a later turn for its foretell cost), `suspend` (CR 702.62a: pay the suspend cost and exile the card with N time counters), since #1342 `plot` (CR 702.170a: pay the plot cost and exile the card from your hand face up; it becomes plotted, castable without paying its mana cost in its owner's main phase with the stack empty on any later turn — the cast is an ordinary `cast_spell` with `from_zone: "exile"` under the card's granted permission) and, since #1194, `turn_face_up` (CR 708.6: pay a face-down permanent's morph cost and turn it face up). **Where the card is, is per kind:** foretell, suspend and plot act on a card in the CALLER'S OWN HAND, which must print the keyword; `turn_face_up` acts on a face-down permanent on the BATTLEFIELD that the caller CONTROLS (CR 708.5), and a permanent someone else controls is reported as `card not found` rather than as a refusal, because a face-down permanent's identity is not theirs to probe. A card that offers no such action is refused with `this card offers no such special action`, before anything is paid — which covers a face-up permanent, a manifested noncreature card (CR 701.34d: it stays face down) and a permanent an effect turned face down whose card prints no morph or disguise either (CR 708.7; since #1209 one that DOES print morph keeps its way back up, because CR 702.37e keys on the card having the ability and not on how the permanent came to be face down). **Timing is per kind, and the engine is the only place it is written down:** foretell any time you have priority during YOUR turn and **legal under split second** (CR 702.61b — a special action is neither a cast nor an activation); suspend any time you could begin to CAST the card, which means sorcery timing for a sorcery, instant timing for an instant or a card with flash, and **not** under split second (CR 702.62c); plot only in your own main phase with the stack empty (CR 702.170a), whatever the card's type or flash says; `turn_face_up` any time you have priority, on any turn, with anything on the stack, and **legal under split second** for the same reason foretell is (CR 702.37c). Outside the window the refusal is `that special action cannot be taken right now`. The cost is mana and nothing else — no targets, no modes, no picker — and `strict` / `auto_tap` behave exactly as they do on `cast_spell`. Mana restricted to casting spells or activating abilities cannot pay for a special action, because it is neither. ADR 0062 Decision 4; #658 / #659 / #1342. |

`<ZoneRef>` is `{ "kind": "<zone_kind>", "owner": "<uuid>" }`. Owner is
omitted for shared zones (`battlefield`, `stack`, `exile`). Zone kinds are
`library`, `hand`, `battlefield`, `graveyard`, `exile`, `command`, `stack`.

#### Controller-only card actions (added in S08.5)

`tap`, `untap`, `move_card`, `add_counter`, `set_battlefield_position`,
`declare_attacker`, `declare_attackers`, `declare_blocker`, and
`declare_blockers` are gated on the caller being the current controller of
the named card. (`declare_attackers` and `declare_blockers` apply the gate to
every entry in the batch, and one foreign creature rejects the whole batch.) A seated player issuing one
of these actions against a card controlled by a different seat receives
an `error` frame with `payload.code = "bad_request"` and
`payload.message = "you do not control that card"`. Admin sessions
(`role: "admin"`) bypass the gate so a moderator can fix wedged board
state. `clear_combat` intentionally stays loose — it's the "combat is
wedged" escape hatch and any seated player may invoke it.

This was deliberately permissive at S08 to support casual cross-seat
play; S08.5 wave 1 tightened it after a 2-player playtest surfaced
players accidentally tapping each other's permanents.

### The attack tax on a declaration (CR 508.1a, [ADR 0080](decisions/0080-attack-taxes.md))

Propaganda, Ghostly Prison and Sphere of Safety make attacking their
controller cost mana. The price is charged by the declaration action
itself, at CR 508.1a, before any creature is staged or tapped — never
as a prompt afterwards.

**Both `declare_attacker` and `declare_attackers` carry the same three
optional payment fields**, the trio `cast_spell`, `activate_ability`
and `special_action` already send:

| field | type | meaning |
|---|---|---|
| `auto_tap` | bool | tap lands for the tax when the mana pool alone cannot cover it, through the same planner and executor a cast uses |
| `locked_sources` | `["<uuid>", …]` | permanents the auto-tapper may not reach for |
| `phyrexian_life` | int | Phyrexian symbols in the tax paid with life instead of mana (CR 107.4f) |

There is deliberately **no `strict`**. An attack tax waived on paper is
the enchantment played as a blank, so the declaration always charges:
omitting all three fields means "pay out of the mana pool, refuse the
declaration if the pool is short".

**The payment is all or nothing** (CR 508.1a). `declare_attackers`
prices the ELIGIBLE set — what it would actually declare after the
silent skips above — and if that price cannot be paid in full it
declares nothing, taps nothing, announces nothing and returns
`attack_tax_unpaid`. Which attacks to drop when only some are
affordable is the player's choice, made by submitting a smaller batch;
the engine must not make it for them. This is the one exception to
`declare_attackers`' "skip what doesn't fit" contract.

The error frame carries the whole price in `payload.reason` (a cost
string such as `"{2}{2}"`) and the symbols the pool and the tapper
together could not cover in `payload.missing`. Unlike
`insufficient_mana` there is **no override**.

The price is public and the client never re-derives it: it rides
`turn.attack_targets[].tax` (see `TurnView` below) as the per-creature
cost of attacking that target, and the bot enumerator stamps the same
figure per move in `legal_moves[].cost.mana`.

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

### `untap_choice` — the untap step's own determination (#826, CR 502.3)

`pending_choices` may carry `kind: "untap_choice"`. It is CR 502.3's
first sentence — "the active player determines which permanents they
control will untap" — asked when a card makes it a real decision:
a cap ("players can't untap more than one land during their untap
steps", Winter Orb) or an opt-out ("you may choose not to untap this
during your untap step", Rust Tick).

It carries and is answered exactly like `choose_cards`: `options[]` are
the permanents in question, `choose_min` / `choose_max` bound the pick,
and `resolve_choice { choice_id, card_ids }` names the ones that untap.
The set-level rule above applies — several caps compose, and a set that
breaks one, or that leaves a mandatory untap on the table, comes back
as a `bad_request` with the prompt still open.

Two things differ from `choose_cards`:

- **The options and the bounds reach every seat**, not just the
  chooser. The candidates are tapped permanents on the battlefield,
  which everyone can already see; `choose_cards` hides both because its
  candidates are usually a hand.
- **It opens in a step that grants nobody priority.** `priority_holder`
  is `-1` and the turn cursor sits on `untap` until the prompt is
  answered; `advance_step`, `pass_priority` and `pass_turn` are refused
  while it is open. Answering it untaps the whole determined set at
  once and carries the cursor on to the upkeep in the same frame.

Permanents that *cannot* untap — held by a "doesn't untap" static or a
"next untap step" marker (`CardView.no_untap`) — are never among the
options. A chosen permanent with a stun counter still spends the
counter instead of untapping (CR 122.1d).

### `entry_reveal_from_hand` — a card choice inside an entry replacement (#1198, CR 614.1c)

`pending_choices` may carry `kind: "entry_reveal_from_hand"`. It is the
reveal-land cycle's own sentence, asked as the permanent enters:

> "As this land enters, you may reveal an Island or Swamp card from
> your hand. If you don't, this land enters tapped."   — Choked Estuary

It carries and is answered exactly like `choose_cards`: `options[]` are
the cards in the revealer's hand that the land's clause admits,
`choose_min` / `choose_max` bound the pick (always `0` / `1` for every
printed card today), and `resolve_choice { choice_id, card_ids }` names
what is revealed. **An empty `card_ids` is the decline**, and it is
always accepted — the floor is zero, so this prompt can never be left
without a legal answer.

Three things are worth saying out loud:

- **The permanent has not entered.** The CR 614 pipeline is suspended
  on the answer, so while the prompt is open the land is still in its
  owner's hand, nothing is on the battlefield, and no ETB trigger has
  fired. The answer decides HOW it enters — untapped if something was
  revealed, tapped if not — rather than fixing it up afterwards. Like
  every prompt that pauses an entry, it blocks the table:
  `advance_step`, `pass_priority` and `pass_turn` are refused while it
  is open.
- **The options and the bounds reach the CHOOSER ONLY**, exactly as
  `choose_cards`' do and unlike `untap_choice`'s. The candidates are
  cards in a hand, and HOW MANY of them match the land's clause is
  itself information about that hand. Other seats see that a choice is
  open and whose it is, and nothing else.
- **What was revealed arrives afterwards, as an ordinary reveal.** The
  answer produces the same grouped reveal frame any other reveal does
  (CR 701.20): every seat becomes entitled to read that card for as
  long as it stays where it is. Nothing moves — a reveal is not a zone
  change (CR 701.20b) — so the card is still in hand when the land has
  finished entering.

### `retarget` — changing a spell's or ability's target (#1196, CR 115.7)

`pending_choices` may carry `kind: "retarget"`. It is the prompt behind
Deflecting Swat's "you may choose new targets for target spell or
ability" and Bolt Bend's / Misdirection's / Imp's Mischief's "change
the target of target spell with a single target": one target SLOT of an
item already on the stack, offered to the retargeting effect's
controller.

It carries and is answered exactly like `pick_target` — `pick_target:
{players, cards, min, max}` with the legal alternatives, answered with
`resolve_choice { choice_id, targets }` — so the client's board picker
answers it with no second surface. The `kind` is what tells the server
the answer rewrites an item on the stack rather than building a new
one.

Three things differ from `pick_target`:

- **The target already in the slot is never among the alternatives.**
  CR 115.7a: a target can be changed only to *another* legal target.
  A slot with no alternative opens no prompt at all and stays where it
  is.
- **`min` is 0 for a printed "you may"**, and the empty list
  (`targets: []`) is the decline. `min` is 1 for a mandatory "change
  the target" that has somewhere to go, and a decline comes back
  `bad_request`.
- **Legality is judged for the ITEM, not for the chooser.** The
  alternatives are the targets the redirected spell could legally have
  chosen — its controller's point of view, its own colour and type for
  protection — while the prompt is addressed to whoever cast the
  redirect. That includes the PLAYER half of the keyword gate
  (CR 702.11d / CR 702.16i, #1197): a seat with hexproof is not among
  the alternatives offered to an opponent's spell, and a seat with
  protection from red is not among a red spell's.

A "choose new targets" prompt over a multi-target spell is a WALK: one
prompt per slot, in announce order, each answerable or declinable on
its own. Everything else about the announcement survives the walk — the
modes, X, the cost that was paid, and the division of damage, which
moves with the slot rather than staying on the old target's id.

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
  Since [#686](https://github.com/krakenhavoc/cmd_and_ctrl/issues/686)
  the kind also carries the *negative*: a bundle the server refused,
  naming the card and saying nothing was changed. It is the same
  mandatory disclosure and renders the same way — a client must not
  try to tell the two apart, and there is no replay annotation for a
  refusal, because nothing was committed.
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

- **GameView**: `{ id, state, seats[], battlefield, stack, exile, phased_out, turn, mulligans_open, starting_seat, undo_limit, settings, monarch?, initiative?, promises?, vote?, discard_pending?, legal_moves?, stack_items?, pending_triggers?, delayed_triggers?, split_second_active?, pending_choices?, log?, reveals?, loop_notice? }` — `mulligans_open` is true between `Start` and the moment all seated, non-eliminated players have called `keep_hand` (S08). `starting_seat` is the seat index that took the first turn; used by the server to enforce the CR 103.8a turn-1 skip-draw rule (two-player games only — CR 103.8c has nobody skip in a larger multiplayer game), surfaced on the wire so spectators / reconnects render the same first-player decision. Pre-S13 snapshots decode as `0` (S13). `stack_items` is the announce-time metadata for every item on the stack (S13.1) — see StackItemView below. `pending_triggers` is the APNAP queue of triggered abilities waiting to drain (CR 603.3b). `delayed_triggers` (S22, omitempty) is the queue of CR 603.7 delayed triggered abilities still owed — "at the beginning of the next end step, return that card to the battlefield" — as `{ id, controller, source?, label?, at, created_turn?, cards?, on? }`, where `at` is the step whose beginning fires it and `cards` are the instance IDs the effect acts on. `on` (#663, omitempty) is the EVENT condition of a "when you next cast an instant or sorcery spell this turn" trigger (CR 603.7b): such a trigger is owed on the next matching event rather than at a step, so it carries `on` and an empty `at`. Public information, so it is not redacted per viewer. `split_second_active` mirrors the engine's "no responses allowed" gate (CR 702.61). `pending_choices` (S14, omitempty) is the async effect-driven decision queue — see PendingChoiceView below. `log` (S31, omitempty) is the public game log — see LogEvent below. `reveals` (S22, omitempty) is the broadcast reveal window — see RevealView below. `loop_notice` (#628, omitempty) is the CR 726 loop breaker: `{ source?, label, controller?, count }`, present once the engine has watched one triggered ability resolve 25 times in a turn with no player decision in between. It is an instruction to the CLIENT, not a change to the rules — priority still rotates and `pass_priority` is still accepted; what stops is AUTOMATIC passing, so the autopass toggle holds and a person has to ask for the next iteration. Table-wide and identical for every seat. Cleared by the next cast, activation, answered prompt or combat declaration, and at the turn boundary. `undo_limit` (S11) is the per-player per-turn undo budget; since ADR 0075 it is always present and mirrors `settings.undo_limit` (`-1` unlimited, `0` no undos). `settings` (S35, [ADR 0075](decisions/0075-table-settings-and-host-controls.md)) is the table's configuration, `{ undo_limit, undo_scope, starting_life, commander_damage, bot_pace, allow_spawn }`: `undo_limit` as above; `undo_scope` `"own"` (a seat undoes only its own entries) or `"host_any"` (the host may undo anyone's); `starting_life` (1–999, default 40; fixed once the game is active); `commander_damage` (1–99, default 21 — the damage from one commander that loses the game, read at every state-based action check, so lowering it mid-game can eliminate a player at the next check); `bot_pace` `"fast"` / `"normal"` / `"slow"`; `allow_spawn` (default false). Public and identical for every viewer, spectators included. Settings are not rolled back by `undo`. `monarch` / `initiative` (S10) are the player ID holding that designation, empty when unassigned; `promises` (S10) is the "I owe you" tally keyed `"{from}->{to}"` with zero entries dropped; `vote` (S10) is the open council's-dilemma / politics vote, absent when none is open. `discard_pending` (S13.4) maps player IDs to the number of cards each still has to discard and drives the discard prompt. `legal_moves` (S31) is the viewer's own enumerated legal moves, filled per seat by `FilterViewFor` and absent from the unfiltered view.
- **PlayerView**: `{ id, name, seat, life, poison?, energy?, library, hand, graveyard, command, commander_damage, life_history, eliminated?, hand_kept?, mulligans_taken?, mana_pool?, is_bot?, bot_tier?, bot_deck?, is_host?, emblems?, keywords?, life_total_locked? }` — `mana_pool` (S15, omitempty) is the player's typed mana pool as an ordered list of single-character color strings (`"W"`, `"U"`, `"B"`, `"R"`, `"G"`, `"C"`) in tap order. Drives the `ManaPoolPips` row in the player header. Server-side empties at every step boundary (CR 106.4) so the wire surface stays small. Public to all viewers (matches paper Magic — floating mana sits visibly next to the player). `life_history` is a per-player rolling log of `LifeChangeView` entries (`{ delta, new_total, at, seq }`, RFC3339 timestamp), bounded server-side at 50 entries and never filtered (life is public). `seq` (#703) is a per-player counter starting at 1, stamped on every recorded change (a zero delta records nothing and consumes no `seq`) and monotonic across frames. It is the only field that identifies an entry, and it is what a client must key a transient life-change cue on: the log is trimmed from the FRONT, so the array index is not stable and the entry count stops growing at the cap, while `at` reaches the wire as RFC3339 **seconds**, so two changes in the same second are indistinguishable. `seq` keeps climbing after the log stops growing; it is derived from the newest surviving entry, so it rewinds with an undo and continues correctly across a snapshot restore. Entries from a snapshot taken before #703 decode as `0`. `eliminated` is omitted unless true; set when the player has conceded (S08) or, in future S13+ work, has lost to a state-based action. `hand_kept` is true once the player has committed via `keep_hand`; `mulligans_taken` counts mulligans in the current opening-hand window. Both omitted when false/zero. All five fields added in S08. `is_bot`, `bot_tier` and `bot_deck` (S31, all omitempty) mark a bot seat, one driven by an in-process `aiseat` runner instead of a browser ([ADR 0033 §9](decisions/0033-ai-bot-seat.md), [docs/bot.md](bot.md)). `is_bot` is true only on such a seat. `bot_tier` is its policy tier: `"random"`, `"heuristic"`, `"assisted"` or `"strong"`. `bot_deck` is the curated deck ID it was seated with, such as `"izzet-aggro"`. It is empty when the bot was seated with a pasted decklist (`{tier, format, source}`). All three are public, unredacted, and absent on a human seat. The client renders the BOT chip, the robot avatar mark and the thinking pulse from them. They also ride the engine snapshot, and `bot_tier` is on the game's `seats` row as well, so a bot seat survives a restart ([docs/lobby.md](lobby.md)). `is_host` (ADR 0075 §2.1, omitempty) is true on exactly one seat, the table host, and absent everywhere else, including on every seat of a table with no host. The host may manage the table alongside the server admin. It is public and unredacted. It is not engine state: the room stamps it on each capture from the host the lobby designated, and it moves to the next human seat in turn order when the host concedes or loses. It is never true on a bot seat. See [docs/lobby.md](lobby.md#the-table-host). `land_drops_per_turn` and `lands_played_this_turn` (#500, both always present) are the two halves of the CR 305.2 land-drop badge: how many lands this seat may play this turn and how many it already has. `land_drops_per_turn` is the EFFECTIVE allowance — a controlled permanent granting extra land plays, or a one-turn grant, is already summed in — so it is normally `1`. Since #500 the engine REFUSES a land play past the allowance, so a client should disable the hand's lands once `lands_played_this_turn >= land_drops_per_turn` rather than wait for the `bad_request`. Both are public. `emblems` (#623, omitempty) is the seat's emblems (CR 114), in creation order, each `{ instance_id, label, text }` — `label` is the board name ("Elspeth, Sun's Champion emblem") and `text` the emblem's printed ability, for the hover. It is deliberately NOT a `ZoneView`: an emblem has no characteristics at all (CR 114.1), so there is no `CardView` to build and nothing for the card renderer, the image route or the targeting layer to do with one. Emblems are **public and unredacted** — an emblem sits face up in the command zone and any player may read it — so every seat's list reaches every viewer intact. Absent for a seat with none, which is nearly every seat. An emblem never appears in the `command` `ZoneView`: that zone is the cast surface and the commander pile, and an emblem is neither castable nor a card. Label and text are read from the catalog on every projection rather than stored, so a wording fix in a card file reaches a game already in progress. See [ADR 0064](decisions/0064-emblems.md). `keywords` (#1197, #1201, CR 702.11d / CR 702.16i, omitempty) is the bare engine tokens a SEAT has right now — `"hexproof"`, or a general `"protection from <quality>"` (`"protection from everything"` is the only quality a catalogued card grants a player today: Teferi's Protection, The One Ring, Leyline of Sanctity, Aegis of the Gods). Derived grants from a controlled permanent come first, then ones granted for a duration. EFFECTIVE, like `max_hand_size` and `land_drops_per_turn` — computed on every projection, not read off stored state. PUBLIC and unredacted, like `emblems`: protection and hexproof on a seat are facts about the board, and `legal_targets` already excludes a protected seat for a viewer who could otherwise see the refusal but not its reason. Absent for a seat with none, which is nearly every seat. Unlike `CardView.protection`, the wire does NOT parse the quality into a structured `ProtectionView` here — there is exactly one production quality today, so a second structured field for one value wasn't worth shipping. The client (`client/src/lib/playerKeywordBadges.ts`) title-cases the parsed quality itself rather than switching on today's two spellings, renders a badge on the seat tile beside the life total in the same visual language as a card's protection badge (#979, `KeywordBadgeRow.svelte`), and reuses the hexproof icon from `keywordIcons.ts`. `life_total_locked` (#1200, CR 119.7 / CR 119.8, [ADR 0085](decisions/0085-life-total-cant-change.md), omitempty) is true while *"your life total can't change"* applies to this seat — Platinum Emperion's printed static, or a grant that lasts until this player's next turn (Teferi's Protection, Teferi's Reproach). EFFECTIVE and PUBLIC on the same terms as `keywords`: the derived half is not written anywhere on the engine's `Player`, so it is computed on every projection, and the lock changes what every player at the table may do — a Bolt that moves no life, a drain that drains nobody, a Phyrexian symbol this seat may not claim — so a viewer who can see the refusal but not its reason is strictly worse off. It is deliberately **not** another token inside `keywords`: that field is the bare engine ABILITY tokens a seat has and the client parses `protection from <quality>` out of it, while a life-total lock is not an ability the player has (ADR 0085 Decision 7). What it does NOT mean: damage is still dealt to the seat (it simply moves no life, so "whenever ~ is dealt damage" triggers and the CR 903.10a commander tally are unaffected), poison counters still land, and the player can still lose the game by every CR 104.3 route but their life total. The client renders it as a `LIFE` badge in the same `KeywordBadgeRow` language, appended after the keyword badges by `client/src/lib/playerKeywordBadges.ts`. `discord_id`, `discord_avatar_hash` and `display_name` (S12.5, all omitempty) are the Discord identity of a seat claimed through Discord sign-in. The client builds the avatar URL from the first two and prefers `display_name` over `name` as the seat label. **Since S34 sub-PR 4 they can change mid-game**: a seated player can link a Discord account to their seat, or move it to another one, through `GET /auth/discord/link` ([docs/lobby.md](lobby.md#get-authdiscordlink)). The change is applied through the room, so it arrives as an ordinary state broadcast to every viewer, and a client should read these fields from each frame rather than cache them per seat. No new message type.
- **ZoneView**: `{ kind, owner?, count, cards[] }` — `owner` omitted for shared zones
- **CardView**: `{ instance_id, name, owner, controller, scryfall_id?, oracle_id?, type_line?, colors?, power?, toughness?, tapped?, counters?, is_commander?, battle_x?, battle_y?, attacking_target?, blocking_target?, auto?, target_mode?, mana_cost?, produced_mana?, mana_abilities?, abilities?, face_down?, face_down_kind?, face_visible?, phased_out? }` — **S16:** `power`, `toughness`, `type_line`, `colors`, and `abilities` carry the *effective* (post-CR-613-layer-resolution) characteristics, not printed. `power` retains its zero clamp for combat damage; `negative_power` (#825, omitted unless below zero) carries the signed effective power including counters for comparisons such as skulk. Use `negative_power` when present and `power` otherwise. Both fields are redacted for unknown cards. The wire keys are unchanged; only the resolution semantics are. Cards with no static ability affecting them produce identical wire output to pre-S16 (printed == effective). When an anthem is in play, the controlled creature's `power` / `toughness` arrive +1; when Mycosynth Lattice is in play, every permanent's `type_line` includes `"Artifact"`; when Lord of Atlantis is in play, other Merfolk creatures' `abilities` includes `"islandwalk"`. `colors` is the current `[]string` of `"W"`, `"U"`, `"B"`, `"R"`, and `"G"`, including layer-5 changes; absent means colorless, and a consumer must never infer color from `mana_cost`. It is redacted with other card characteristics. `abilities` (S16, omitempty) is a `[]string` of canonical keyword names ("flying", "first strike", "trample", "islandwalk", "fear", "intimidate", "shadow", "horsemanship", "skulk", etc.) — drives S18's keyword renderer; behavior of the keywords lands with S18 (combat sub-step rewrite). Non-protocol-bumping change: pre-S16 clients ignore the unknown `abilities` field harmlessly. `mana_cost` (S15, omitempty) is the card's printed cost string in Scryfall format (`"{2}{R}{R}"`, `"{X}{G}"`, `"{W/U}{W/U}"`); the parser accepts generic, colored (`{W|U|B|R|G|C}`), variable (`{X}`), hybrid two-color (`{W/U}`), hybrid generic (`{N/W}`), phyrexian (`{W/P}`), and snow (`{S}`) symbols. Drives the cost chip on the cast modal and the strict-mode gate. Empty / omitted ⇒ treated as costless (for cards Scryfall ingestion couldn't parse, which the cost gate emits as an `EventCostWarning` rather than rejecting). Redacted to empty for cards the viewer is not a `known_by` of. `produced_mana` (S15, omitempty) is the parsed Scryfall `produced_mana` field (e.g. `"{C}{C}"` for Sol Ring, `"{W|U|B|R|G}"` for Birds of Paradise) — informational; the canonical activation surface is `mana_abilities`. `mana_abilities` (S15, omitempty) is the per-ability projection the client renders into the right-click "tap for mana" menu: each entry is `{ "label": string, "produced": string, "cost_tap"?: bool, "cost_sacrifice"?: bool }`. Synthetic basic-land abilities are rendered into this list when the catalog has nothing registered (Forest → `[{"label": "Add {G}", "produced": "{G}", "cost_tap": true}]`). `scryfall_id` is stamped at deck-import time and lets the client resolve images via `GET /cards/{id}/image`. Omitted for placeholder cards (demo game seeded via `CMDCTRL_SEED_DEMO`). `oracle_id` (S14) is the Scryfall oracle-level identifier (stable across printings) and is the catalog lookup key — every printing of Lightning Bolt shares one `oracle_id`. `type_line` (e.g. "Legendary Creature — Human Wizard"), `power`, and `toughness` carry Scryfall data the client uses to filter creature-only UIs and label combat rows; all three omitted for placeholder cards. `battle_x` / `battle_y` are normalised positions in `[0, 1]` stamped by `set_battlefield_position` (S06+). **Sent for battlefield cards only, and for every one of them (#29)** — so an absent pair means "this card is not on the battlefield", never "this card is at the origin". They previously carried `omitempty` on a plain float and so vanished whenever the position was `(0, 0)`, which conflated a deliberate origin stamp with a permanent nobody had positioned. The origin is a routine value, not an exotic one: `set_battlefield_position` *clamps* rather than rejects, so every negative and NaN coordinate lands on exactly 0, and `x = 0` is "first in the row" for the client's within-row sort. Outside the battlefield the pair is omitted because it is meaningless there — zone exit clears it — and those zones are the bulk of every frame. Note this makes the wire honest rather than more expressive: `game.Card` carries no "positioned" flag, so "on the battlefield but never positioned" is not a state the server can distinguish from `(0, 0)`; such a card reports `(0, 0)`. Clients should keep an `?? 0` fallback — replays captured before this change omit the pair on battlefield cards too. `attacking_target` (player UUID) and `blocking_target` (attacker instance UUID) are set by `declare_attacker` / `declare_blocker` and cleared on zone exit and by `clear_combat`. Both omitted when not set. `type_line`, `power`, `toughness`, `attacking_target`, `blocking_target` added in S08. `attached_to` (S24, omitempty) is the CR 301.5c / CR 303.4 attachment relation for an Equipment or an Aura — a `TargetRefView` (`{ kind, id }`) naming the permanent (`kind: "card"`) or player (`kind: "player"`) this card is attached to. Omitted for every card attached to nothing. Shipped in ONE direction only: "what is attached to this creature" is derived client-side by partitioning the battlefield on this field, so the two directions cannot disagree. Not redacted — attachment is public battlefield state exactly like `attacking_target`. See [ADR 0036](decisions/0036-attachments.md). `restrictions` (S24, omitempty) is the CR 508.1c / 509.1b / 602.5 restriction set the engine computed for this permanent, as stable snake_case tokens: `"cant_attack"`, `"cant_block"`, `"cant_be_blocked"`, `"cant_activate"`, `"cant_activate_mana"`. Omitted for the permanent nothing is restricting, which is nearly all of them. Deliberately NOT folded into `abilities`: a restriction is not a keyword the permanent has, it is an effect something else has (Pacifism, Arrest, a Whispersilk Cloak), so it renders as a disabled control with a reason rather than as a keyword badge. The client READS it and derives nothing — who may attack is the server's decision and this is how it says so. Public battlefield state, redacted only for a card the viewer is not a `known_by` of (#95). See [ADR 0045](decisions/0045-combat-restrictions.md). `auto` (S14, omitempty) is true when the card's `oracle_id` is registered in the catalog (drives the gold-leaf AUTO badge). `target_mode` (S14, omitempty) is the announce-time target-prompt shape the client should show when casting: `"any"`, `"player"`, `"creature"`, `"permanent"`, `"stack_spell"`, `"stack_ability"`, `"stack_item"`, `"card_in_graveyard"`. The three stack values (#1211, CR 115.4) differ only in the sentence the picker writes — `"stack_spell"` a spell, `"stack_ability"` an activated or triggered ability, `"stack_item"` either — and all three are answered by an ordinary `{kind: "card", id}` ref: a picked ABILITY carries its STACK ITEM id (the id in `stack_items[i].id`), which for a spell item is already the card instance id, so no new ref kind arrives on the wire. Legality is `legal_targets` and never the mode. Empty ⇒ no prompt (cast fires immediately). Both fields redacted to zero for cards the viewer is not a `known_by` of. `target_cost_notes` (#746, omitempty) is a `[]string` of the printed clauses of the card's own cost modifiers whose price depends on its targets — Fireball's "This spell costs {1} more to cast for each target beyond the first", strive's "{1}{W} more" — stamped with the other cast clauses on cards in the viewer's own hand, command zone, castable graveyard, castable library top and (since #978) an exiled card the viewer's own grant opens. The X picker opens before targeting, and the auto-tap preview it reads is priced with no targets (so at the one-target price); the picker shows these clauses under its readout rather than a surcharge it cannot know yet ([ADR 0048](decisions/0048-cost-modification.md) addendum, open question 2). The cast itself is priced with its real targets. Absent for nearly every card; redacted with the other card-identity fields for a card the viewer is not a `known_by` of, and stripped with `legal_targets` and the other cast clauses from an opponent's revealed hand card. It ships on the snapshot rather than on the auto-tap preview response the ADR first named ([ADR 0048](decisions/0048-cost-modification.md) §15, amendment). `phyrexian_symbols` (#916, omitempty) is how many symbols in the card's cost carry CR 107.4's "or 2 life" option — 1 for Gitaxian Probe's `{U/P}`, 2 for Dismember's `{1}{B/P}{B/P}` — and is the ceiling on the `phyrexian_life` a `cast_spell` may claim. Shipped as a count rather than left to the client for the reason `activated_abilities[i].demands_x` is: a client re-deriving it would be a second parser of the mana-cost syntax. Stamped with the other cast clauses on cards in the viewer's own hand, command zone, castable graveyard, castable library top and (since #978) an exiled card the viewer's own grant opens, stripped from an opponent's revealed hand card, and redacted with `mana_cost` for a card the viewer is not a `known_by` of. `alternative_costs[i].phyrexian_symbols` answers the same question for an offer that replaces the printed cost, and `activated_abilities[i].phyrexian_symbols` for an activated ability's mana component.
- **TurnView**: `{ number, active_seat, priority_holder, phase, step }` — `priority_holder` is the seat index (0-based) that currently holds priority within the step, OR `-1` (the `NoPriority` sentinel) during steps that don't grant priority. S13 lands the cursor with `priority_holder = -1` for `untap` and `cleanup` (CR 502.4 / 514.3); auto turn-based actions (auto-untap, auto-draw, auto-advance through cleanup) fire from the step-entry hook so the cursor never sits idle on a no-priority step. Added in S07; sentinel added in S13.
  Two omitempty fields join it during combat. **`block_decision_seats`** is the seat indices that still owe a block decision. **`attack_targets`** (S27) is what the ACTIVE player's creatures may attack right now, as `[{ kind: "player" | "planeswalker" | "battle", id, tax? }]`: `id` is a seat id for a player and an instance id for a permanent, and the client must not re-derive "who protects which battle". **`tax`** (omitempty, [ADR 0080](decisions/0080-attack-taxes.md)) is the CR 508.1a price ONE creature pays to attack that target — `"{2}"` against a seat with Propaganda out, `"{2}{2}"` against one with Propaganda and Ghostly Prison, absent when attacking it is free. Server-priced, per creature; a declaration's real price is this once per attacking creature.
- **StackItemView** (S13.1): `{ id, kind, controller, owner, source_card_id, label?, targets?, modes?, x_value?, distribution?, hold_priority?, split_second?, doubled_by?, doubled_by_name?, mana_spent?, colors_spent?, mana_spent_unknown? }` — announce-time metadata for one item on the stack. `kind` is `"spell"` (card lives in `stack` zone), `"activated"` or `"triggered"` (no card; source stays in its origin zone). `targets` is a list of `TargetRefView` slots: `{ kind: "player" | "card" | "self" | "none", id?: "<uuid>" }`. Resolution re-checks targets per CR 608.2b — if every targeted slot is illegal, the spell is "countered by game rules" and routes to its owner's graveyard. **#761:** `mana_spent` is how many mana paid for the spell and `colors_spent` the distinct COLOURS among them, in WUBRG order (colourless is not a colour, so it never appears there though it counts in `mana_spent`). Mana is spent face up, so both are public. `mana_spent_unknown: true` means the cast went through permissive mode or a strict-mode override: the engine never took the mana and has no record of what it was — render that as unknown, never as zero, because "nothing was spent" is a stronger claim and the one Vexing Bauble punishes. All three are absent on an ability item and on a copy of a spell (CR 707.10 — mana is not an object, so nothing was spent to cast the copy, and that zero is real). `PlayerView.commander_casts` (S13.1, omitempty) is the per-commander cast tax counter map keyed by commander UUID string. `CardView.damage_marked` (omitempty) drives the lethal-damage SBA. **#764:** a `TargetRefView` may additionally carry `slot` (the target CLAUSE it answered) and `mode` (the index into `modes` — the mode OCCURRENCE — whose clause list that is), both omitted at zero; and the item carries `mode_labels?: [string]`, the oracle bullet of each chosen mode in announce order and with repeats. The labels travel with the item because the caster's hand card is gone by the time the spell is on the stack, and `modes: [0, 2]` is not something a table can read.
- **Trigger doubling attribution** (#752, CR 603.2d): `StackItemView` may additionally carry `doubled_by?` (the public doubler permanent's UUID) and `doubled_by_name?` (its public card name). The fields are additive and omitted for ordinary spells, abilities, and undoubled triggers. A `trigger_prompt` or `pick_target` `PendingChoiceView` carries the same two optional fields, so the chooser can tell that the pending trigger is the additional trigger. The server projects these fields only from the engine's explicit trigger-doubler metadata; card identity in unrelated hidden zones remains subject to the normal per-viewer filter.
- **CardView.known_by_you / face_down** (S13.5): per-viewer card visibility. `known_by_you` is true when the viewer is in the server-side KnownBy set for that card; the wire surfaces full printed characteristics in that case. When false, the wire keeps only instance_id / owner / controller / tapped / damage_marked / face_down / battlefield position / combat declarations (`attacking_target`, `attacking_target_kind`, `defending_player`, `blocking_target`) / `goaded_by` / `attached_to`, and zeroes everything read off the card itself — name, type line, Scryfall ID, P/T, counters, costs, faces, `auto`, `target_mode`, `mana_abilities`, `activated_abilities`, `restrictions`, `exile_play`, the hand-zone cast stamps, and the type-derived `summoning_sick` / `loyalty_activated` / `defense` / `protector_player` (#95; `face_down_view_test.go` pins this as an allowlist). The engine state stays observable, but identity doesn't leak. `face_down` is the visual flip flag (CR 406.3 / 708); the client draws a card back for a face-down card the viewer does not know, and the face for one it does. **ADR 0069** adds `face_down_kind` and `face_visible`. `face_down_kind` is WHY it is face down — `exiled` (CR 406.3, Necropotence), `foretold` (CR 702.143b), `hideaway` (CR 702.75a, [ADR 0091](decisions/0091-hideaway.md)), or one of the CR 708.2 permanent states `manifested` / `morphed` / `disguised` / `cloaked` — and it is PUBLIC, so it survives the redaction and labels the card back. `face_visible` is whether THIS viewer may look at the face: its controller for a CR 708.5 permanent, its owner for a foretold card (CR 702.143d), the controller of the permanent that hid it for a `hideaway` card (and, by CR 406.3, anyone who has already looked — a hideaway land that changes hands hands the look over and the previous controller keeps theirs), nobody for a plain face-down exile (CR 406.3). A hidden SPELL that its linked ability has opened arrives with the ordinary `exile_play` stamp for the holder (`cost_override: "{0}"`, not `cast_only`, since hideaway says *play*); a hidden LAND is offered through the existing `may_cast` prompt and, on yes, played during the resolution as a land play (ADR 0091 decision 5). Since #1194 a spell CAST FACE DOWN (CR 708.4 — morph, megamorph, disguise) is one of these too while it is on the stack: it arrives as `face_down` with `face_down_kind: "morphed"` or `"disguised"`, the public 2/2 body, and `face_visible` true for its caster alone. It is stamped per-viewer beside `known_by_you`, equals `face_down && known_by_you`, and is what the client keys on to draw the real face plus a face-down badge. **The one exception to the redaction list above** is a face-down PERMANENT: CR 708.2 makes it a 2/2 colourless creature with no name, that body is public (an opponent has to see the 2/2 to block it), and so `type_line` (`"Creature"`), `power`, `toughness`, `colors`, `abilities`, `counters` and `summoning_sick` reach every viewer for it. When the effect that put it face down LISTED its characteristics (CR 708.2, #1270), that listed body is the public one: `type_line` is `"Artifact Creature — Cyberman"` for Cyber Conversion's victim and `"Land — Forest"` (no `power` / `toughness`) for Yedora's return, and `face_down_kind` is `"turned"` for both. No new field: the listing is read at layer 0, so the existing fields carry it. Everything that names the card underneath — `scryfall_id`, `mana_cost`, `faces`, `layout`, `auto`, `target_mode`, the ability lists — is still stripped from a non-knower, and most of it is never stamped at all because `CatalogKey` returns the empty key for a face-down permanent. Library cards have no knowers post-shuffle; opening-hand cards are known to their owner only; battlefield / stack / exile / graveyard / command-zone cards are public (all seated players are knowers). `ShuffleLibrary` clears every library card's KnownBy; `Mulligan` clears hand + library and re-grants the owner on the new opening hand. S14's `keepKnownInHandZone` view-filter path preserves revealed opponent-hand cards end-to-end: a seated viewer receives opponent hand zones containing only their `KnownByYou == true` entries (the `Count` stays accurate so the client can render the hidden remainder as face-down placeholders); spectator / admin connections still see a fully-hidden opponent hand via `hideZoneContents`.
- **CardView.defending_player** (#1339, omitempty): on an attacking creature, the seat defending against its attack — the only seat whose creatures may block it (CR 802.4a): the player attacked, the controller of the planeswalker attacked, or the PROTECTOR of the battle attacked (CR 310.9d). Server-computed from the same function the block option generator and `declare_blocker(s)` read, so the client's block pickers offer exactly the blocks the server accepts; a block on any other attacker is refused with `illegal_block` / `not_defending`. Absent when the card is not attacking. A creature attacking a planeswalker or battle that has since left the battlefield keeps the seat that was defending it (#1364, CR 506.4c "it may be blocked", CR 802.2a), and its `attacking_target_kind` is absent because it attacks nothing — it may be blocked by that seat, and if unblocked it deals no damage. The same holds since #1376 when the planeswalker or battle is removed from combat WITHOUT leaving the battlefield (CR 506.4 — a control change, or phasing out): `attacking_target` is then the reserved id `00000000-0000-0000-0000-000000000506` (`game.AttackingNothing`), which names no seat or card, so the creature still reads as attacking but no arrow can resolve and no damage is dealt, and `defending_player` is the seat that was defending it before the removal. Public, like `attacking_target` it is derived from. A client reading a frame that predates the field falls back to `attacking_target` for a player attack (`defendingPlayerOf`, `client/src/lib/attackTargets.ts`).
- **CardView.class_level / solved** (S46, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md), both omitempty): the two DESIGNATIONS a permanent can have. `class_level` is a Class's CR 716.2 level — 1 for a Class nobody has levelled, up from there — and is present only for a Class on the battlefield, so the client renders its badge on presence rather than by parsing the type line. An **uncatalogued** Class carries it too: the level is engine state, not catalog state, exactly as an uncatalogued Saga still shows its lore counters. `solved` is a Case's CR 719.3 solved marker, absent rather than `false` for everything else. Both are PUBLIC — a level and a solved marker are visible across the table in paper — and both are cleared on the non-knower redaction with the other type-derived bits (`summoning_sick`, `loyalty_activated`, `defense`, `protector_player`), because a level says "Class" and a solved flag says "Case" as loudly as loyalty says "planeswalker"; CR 708.2 gives a face-down permanent no subtypes, so it is neither.

  **There is no field for "which printed abilities are active", and there does not need to be.** The designation gate is evaluated where an object becomes abilities, so an inactive ACTIVATED ability is already missing from `activated_abilities` — the same accessor the engine, the legal-move enumerator and the lobby read — and an inactive static or trigger has no per-ability representation on the wire to grey out. A station threshold needs nothing at all: charge counters ride the existing `counters` map, exactly as lore counters do.

- **CardView.token_text** (S46, [ADR 0083](decisions/0083-token-abilities.md), omitempty): a TOKEN's printed ability text, verbatim — `"When this token dies, you gain 1 life."`, with `\n` between printed lines. Absent for every printed card and for a vanilla token, which is what makes it additive: a pre-ADR-0083 client ignores the unknown key harmlessly. **It exists because a token has no printing behind it.** `scryfall_id` is empty, no art loads (token art is [ADR 0078](decisions/0078-token-art.md), still Proposed), and the client renders the card through `Card.svelte`'s `.name-fallback` branch — the name on a grey rectangle. A mana or activated ability escapes that, because it reaches the player as a row in the right-click menu; a token's own TRIGGER ("when this token dies, create a 2/2 red Dragon") or STATIC has no control to hang text off, so without this field the words are nowhere on the client at all. `EmblemView.text` is the same field for the same reason, one object over. **Derived on every read** from the token template's catalog entry (`game.TokenTextForCard`), never stored, so a wording fix reaches a game already in progress. Keyed on `CatalogKey` and not `CatalogAbilityKey`: a token silenced by Dress Down still shows what it PRINTS, exactly as the client keeps rendering a silenced card's oracle text. CR 707.2 rides along — a token that is a copy of a printed card has that card's oracle ID, answers empty here, and renders that card's own printing instead. Cleared on the non-knower redaction with the other catalog reads, though a token is known to every seat so the cell never fires in a real game.

- **CardView.prepared** (#1328, [ADR 0090](decisions/0090-preparation-cards.md), omitempty): a permanent's CR 722.3a PREPARED designation. While it is set, the permanent's controller may cast the copy of its prepare spell that sits in exile — and that copy needs no new field: it is an ordinary card in `exile` whose `name` / `type_line` / `mana_cost` are the prepare spell's (face `1` active), carrying the `exile_play` stamp the client's exile button already reads, with `faces: [1]` and `player` naming the prepared permanent's controller. The stamp is DERIVED from the live permanent on every frame, so it disappears the moment the permanent leaves, is unprepared or its copy is cast; the copy itself leaves the `exile` zone view at the next state check. Public (the prepared creature and its copy are both visible across the table) and absent rather than `false` for every permanent that is not prepared; cleared on the non-knower redaction with `class_level` and `solved`. The client renders it through the same designation badge ("PREPARED").
- **CardView.harnessed** (#1321, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md) amendment 2026-09-23, omitempty): a permanent's CR 701.64 HARNESSED designation — the marker that switches its printed "∞ — [ability]" lines on (CR 702.186b). Unlike `class_level` / `solved` it carries no subtype gate: any permanent can print "Harness [this permanent]" (The Mind Stone is the first), so the field is read straight off the card once it is on the battlefield. Public (harnessed is visible across the table in paper) and absent rather than `false` for every permanent that is not harnessed; cleared on the non-knower redaction with `class_level`, `solved` and `prepared`. There is still no field for "which printed abilities are active" — an inactive `∞` trigger has no per-ability wire representation to grey out, exactly as an inactive Class or Case line has none.
- **CardView.chosen_color / named_tribe** (#781, both omitempty): the answers a player gave to this permanent's "as this enters, choose a color" (CR 105.4) and "as this enters, choose a creature type" (CR 614.12) instructions — one uppercase colour letter (`"W"` / `"U"` / `"B"` / `"R"` / `"G"`) and one canonical creature type (`"Elf"`). Written by the engine onto `game.Card.ChosenColor` / `game.Card.NamedTribe` when the entry prompt is answered, and **cleared on every battlefield exit** (`game/zone.go`, `game/entry_tail.go`) — so a non-empty value means exactly "a permanent on the battlefield whose controller has answered", and `viewOfCard` copies both with no zone gate of its own. Absent for the overwhelming majority of cards, and absent in the window between the permanent entering and the prompt being answered; a consumer must read an absent value as the weaker outcome (no anthem, no mana), **never** as "any colour". **PUBLIC.** The choice is announced at the table and hidden from nobody, including the player who made it, and CR 607.2d is why it has to be on the card rather than only in the log: the linked ability says "creatures you control of the chosen color", and that set cannot be computed without the answer — an opponent could not tell which of their creatures a Heraldic Banner was pumping, and the controller had to remember what they named. Cleared on the non-knower redaction with the other type-derived bits (`class_level`, `solved`, `summoning_sick`, `loyalty_activated`), because `"Elf"` names Cavern of Souls and `"G"` names Coldsteel Heart as loudly as loyalty says "planeswalker"; CR 708.2 also leaves a face-down permanent with no such ability to have asked the question. The client renders both through one module (`client/src/lib/chosenValues.ts`) — a chip on the card tile's keyword row and a line in the hover panel's footer, with the linked clause in the tooltip. The engine event kinds (`EventColorChosen`, `EventCreatureTypeChosen`) are still absent from the public log; see #781.

- **GameView.phased_out / CardView.phased_out** (S46, [ADR 0084](decisions/0084-phasing.md), #1199): the CR 702.26 phased-out permanents. A shared, owner-less, public `ZoneView` beside `battlefield`, `stack` and `exile`, with `kind: "phased_out"`, and every card in it carrying `phased_out: true`.

  **A phased-out permanent is absent from `battlefield`, not flagged inside it**, and that is the whole point rather than a detail. CR 702.26b says a phased-out permanent "is treated as though it does not exist"; ADR 0084 makes that true of the engine by keeping it out of `Game.Battlefield.Cards`, and the wire follows so that the two surfaces that read the view — the bot's sixteen battlefield walks in `internal/aiseat`, and the positional stamps in `view.go` that index `view.Battlefield.Cards[i]` against `g.Battlefield.Cards[i]` — are correct without learning the rule.

  Phasing is **not** a zone change (CR 702.26d), so there is no `zone` log entry beside it and the permanent's `instance_id`, counters, damage, tapped state and `attached_to` are all unchanged across a phase cycle. `phased_out` on the card is redundant with the zone it arrived in and is carried anyway, so a client that one day renders phased-out permanents in place on the board has the bit without another wire change. It is **public** like `face_down_kind` and survives the non-knower redaction: everyone can see the board stop showing a permanent, and a viewer who cannot identify the card still has to be able to tell "phased out" from "died".

- **The engine event log is not on the wire.** S14 documented an `events` field on `GameView`, a rolling window of `EventView` entries (#139), but it was never added: no `GameView` in `server/internal/protocol/view.go` or `client/src/lib/protocol.ts` has ever carried one. The engine's log (`game.Game.Events`, kinds in `server/internal/game/events.go`) stays on the server. Game-state persistence saves it (`server/internal/game/snapshot.go`), and it reaches clients only through two table-visible projections built from it: `log` (LogEvent, below) and `reveals` (RevealView, below).

  Two engine event kinds about **prompts that changed hands when a player left the game** are deliberately server-side only and are **not** projected into `log`: `pending_choice_dropped` (#864) and `pending_choice_reassigned` (#902, CR 800.4g/h — `actor` the departed chooser, `target` the player who inherits, `source` the object, `label` the `PendingChoiceKind`). Both are diagnostics for a stalled table. What a player needs to see is already on the wire without them: the prompt itself, which the very next `snapshot` carries under its new `PendingChoiceView.chooser` — no new field, and no client change, because the picker already renders a prompt exactly when `chooser` is the viewer.
- **LogEvent** (S31, omitempty): `{ seq, kind, turn?, step?, seat, target_seat?, card_id?, target?, amount?, old_zone?, new_zone?, combat?, combat_step?, choice?, label?, actor_is_host?, text }` — one line of the **public game log** ([ADR 0033](decisions/0033-ai-bot-seat.md) §4). `GameView.log` is the last 200 table-visible events, **oldest first**, and it is a projection of the engine's own event log rather than a stored buffer — nothing on `game.Game` holds it, so it survives an undo, a snapshot restore and a deploy by riding `Game.Events`, which already does.

  `kind` is one of `step`, `cast`, `resolve`, `fizzle`, `counter`, `zone`, `draw`, `life`, `damage`, `attack`, `block`, `token`, `sacrifice`, `eliminated`, `reveal`, `roll`, `flip`, `choose_color`, `choose_type`, `choose_player`, `control`, `special_action`, `cycle`, `counters`, `scry`, `surveil`, `saga_chapter`, `class_level`, `settings`, `spawn`, `transform`, `phase_out`, `phase_in`, `storm`, `turn_face_down` — deliberately coarser than the engine's event kinds, because several engine events are one line to a reader and most engine events are no line at all. **Which** engine kinds get no line is no longer a matter of taste: `server/internal/protocol/log_event_kind_gate_test.go` (#984) reads every declared `game.EventKind` and fails unless the projection has an arm for it or the file lists it as a deliberate silence with a written reason.

  **An ability's `resolve` / `fizzle` (#1257).** For a SPELL, `card_id` is the spell and there is no `label`. For a triggered or activated ABILITY, which has no card on the stack, `card_id` is the ability's **source** and `label` the stack item's label — the same string the stack overlay printed for it ("Grapeshot — storm", "Sacrifice a creature: scry 1"). `text` names the ability by the label, prefixed with the source's name when the label does not already start with it: "Grapeshot — storm resolved", "Viscera Seer — Sacrifice a creature: scry 1 resolved", "Prodigal Pyromancer — {T}: deal 1 damage to any target was countered by game rules (no legal targets)". A copy of an ability (CR 707.10) carries its original's source and label, so it reads the same. A `label` on a `resolve` entry is therefore how a client tells an ability from a spell. **Redaction keys on the source's knowers, not on a dropped name**: a viewer who may not identify the source gets `card_id` and `label` both cleared and reads "an ability resolved". The source's id goes with the label because a source in a hidden zone (a hand, a library) has an instance id the zone filter never hands a non-knower, and the knower test has to be direct because a face-down permanent's name is empty on every view (CR 708.2), so the ordinary name-drop test sees nothing to drop while the label names the card outright. A source the view cannot account for at all — a token or copy that has ceased to exist, both public — keeps its line. Before #1257 these entries carried neither field and every ability rendered as "a card resolved".

  **Chosen values (#984).** `choose_color` (CR 105.4), `choose_type` and `choose_player` (CR 614.12) are the answers a player gives out loud to a "choose a ..." prompt — Coldsteel Heart's colour, Cavern of Souls' tribe, True-Name Nemesis' player. All three carry `card_id` (the card the answer was given for). `choice` carries the VALUE on the first two — the colour **letter** (`"G"`), the canonical creature type (`"Elf"`) — and is absent on `choose_player`, whose answer is a seat and therefore rides `target_seat` like every other player in the log. `text` renders it as "P1 chose green for Coldsteel Heart". `choice` is **redacted with the card's name**: the answer identifies the card as loudly as the name does (which is why `CardView.chosen_color` / `named_tribe` are stripped for a non-knower, #781), so a viewer who may not identify the card gets "P1 chose a color for a card" and no `choice` at all.

  **The six narrated silences (#1021).** Writing the log's deliberate silences down (#984) showed six of them to be gaps rather than decisions, and each is now a line. They are eight `kind`s for six decisions, because scry is not surveil and a Saga chapter is not a Class level — a client tones and filters by `kind`.

  - `control` — a permanent changed controller (CR 613.1b). `seat` is the player who GAINED control and `target_seat` the one who lost it, which is one sentence for a gain, an exchange (CR 701.12) and a duration expiring: "P1 gained control of Grizzly Bears from P2".
  - `special_action` — a CR 116.2 special action: foretell, suspend, plot, turning face up. `label` is the action as the card prints it ("Foretell {2}"). The card's zone move says only that a card left a hand for exile; this says which action it was.
  - `cycle` — a cycling (CR 702.29b). It **replaces** the `zone` entry for the discard that paid the cost, the way a `sacrifice` entry replaces the zone move it causes, and keeps that entry's `seq`.
  - `counters` — the count of one counter kind on one card changed (CR 122). `label` is the kind (`"+1/+1"`), `amount` the count **after** the change — the engine's event carries no delta, so a placement and a removal are the same line with a different number, and `amount` ≤ 0 means the last one came off. **Loyalty and lore counters get no line**: a planeswalker's loyalty moves on every activation and every point of damage, both of which are already entries, and a lore counter's advance is the `saga_chapter` line below. The rule is `counterKindIsNarrated` in `log.go`.

    Counters on a **player** (poison, energy, experience, rad) are a different engine event — `player_counter_placed`, added by [ADR 0056](decisions/0056-infect-wither-toxic.md) Decision 5 — and get **no log entry yet**. It is a separate kind from `counter_placed` because the card-counter trigger helpers compare the event's target against card IDs, and it carries the **signed delta that landed** rather than the new total (the total is on `PlayerView.counters`). It is also the event the layer listener bumps on, which is what keeps a "corrupted" static reading an opponent's poison count from going stale. Its line — a `poison` entry reading "Alice got 3 poison counters (7/10)" — is [ADR 0056](decisions/0056-infect-wither-toxic.md) Decision 6 and lands with the client PR; until then `PlayerView.poison` carries the current total to every viewer on every frame, and the silence is recorded in `silentEventKinds` (`log_event_kind_gate_test.go`) with that reason.

    Counters on players also go through the **CR 614 replacement window** now, on the same `RepEventCounter` kind as counters on permanents, with `CounterPlayer` set instead of `CounterTarget`. That is what makes Vorinclex halve an opponent's poison and Lae'zel's "or on yourself" clause work. No action or view shape changed.
  - `scry` / `surveil` — a finished scry (CR 701.22) or surveil (CR 701.25). These carry **no `card_id` for anyone**, not even the spell that scried: the cards are hidden at both ends, and an entry with no card reference cannot leak one. `amount` is the public half — how many cards the table watched go to the bottom of the library, or into the graveyard. It is **not** the size of the scry ("scry 2"), which the engine's event does not carry.
  - `saga_chapter` / `class_level` — a Saga reached a chapter (CR 714.2b) or a Class became a level (CR 716.2); `amount` is the chapter or the level.
  - `settings` — the host or the admin changed a table setting ([ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.3, S35). `label` is the setting's key (`undo_limit`, `undo_scope`, `starting_life`, `commander_damage`, `bot_pace`, `allow_spawn`) and `choice` its **new** value as text — an int in decimal (`"-1"` is unlimited), a bool as `"true"`/`"false"`, an enum as its string. One line per field that actually moved; a no-op patch says nothing. `seat` is the host who changed it, or `-1` for the server admin, which the rendered text reads as "The admin". The only kind that is about the rules the game is being played under rather than about the game, which is exactly why it is written down: a budget that quietly halved mid-game is what this line prevents. It carries **no card**, so none of the card-identity redaction below applies to it.
  - `transform` — a permanent was turned over to its other face (CR 701.27a, S46, [ADR 0079](decisions/0079-transforming-a-permanent.md)). The entry's card name is the face it turned **into**, because `viewOfCard` reads the active face; `label` is the name of the face it turned **from**, which is the only place that name survives. A transform is **not** a zone change (CR 712.18 — the permanent doesn't become a new object), so there is no `zone` entry beside it and this is the only line the table gets. It is narrated rather than silent, unlike the other board-state changes, because a card physically turning over is announced out loud and a reader scrolling back wants to know when it happened.
  - `turn_face_down` — a permanent that was face up on the battlefield was turned face down (CR 708.2a, S46, [ADR 0082's 2026-09-23 amendment](decisions/0082-casting-face-down-and-turning-face-up.md), #1209). `card_id` is the permanent and `target` the object that did it (Ixidron, Cyber Conversion). Narrated for `phase_out`'s reason: turning face down is neither a zone change nor a transform (CR 701.27b says so in as many words), so no other line says it, and the board silently stops showing a card the table could read a second ago. The entry **names the source and not the permanent**, and that is not a redaction: a CR 708.2 object has no name for ANY viewer, its controller included (`CardView.Name` is the effective characteristic's, and the controller gets the art through `scryfall_id` / `face_visible` instead), so the line reads "Ixidron turned a card face down" for the whole table. The reverse direction has no line — `turned_face_up` is a deliberate silence whose `special_action` entry already carries it.
  - `phase_out` / `phase_in` — a permanent phased out or in (CR 702.26, S46, [ADR 0084](decisions/0084-phasing.md), #1199). Narrated for `transform`'s reason and one that is stronger here: phasing is **not** a zone change (CR 702.26d), so there is no `zone` entry beside it, and the board simply stops showing the permanent — which is indistinguishable from a permanent that died unless the log says which. Teferi's Protection removes a whole board this way, and a reader scrolling back has no other way to learn that it came back rather than never left.

  **`storm` (S46, [ADR 0086](decisions/0086-storm-and-the-turns-cast-order.md), #1238).** A storm trigger resolved and settled on its count (CR 702.40a — in the pinned August 2026 edition storm is **702.40**; 702.39 is provoke). `card_id` is the storm spell, `amount` is how many *other* spells were cast before it this turn, by any player, and `text` reads "Grapeshot — storm count 3". The count is how many copies are about to be created, and an entry is emitted at zero too ("storm count 0" is the turn's first spell), so a reader can tell a trigger that found nothing to copy from one that never fired. `amount` is **not** redacted with the card name, unlike `counters` / `saga_chapter` / `class_level`: it is a fact about the turn's casts, which every seat watched, not a value read off the permanent — a viewer who may not identify the card reads "a card — storm count 3". This entry is the only line the table gets for the copies: a copy of a spell is *created*, not cast (CR 707.10), so it emits no engine event and produces no `cast` line, and the trigger's own `resolve` entry names the ability ("Grapeshot — storm resolved", #1257) but not the count. Since #1255 a storm spell COUNTERED in response to its own trigger is still copied, from last-known information (CR 608.2h), so a `counter` line for the spell can be followed by this entry and its copies.

  **`spawn` (S35, [ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.4).** The host or the admin put cards or tokens onto a live table from nowhere. `label` is the card or token name, `amount` the count, `new_zone` where they went and `target_seat` the seat whose zone it was; there is no `card_id`, because one entry covers every copy the request made. `text` reads "Luke (host) spawned 2 × Treasure onto Ana's battlefield". This line is the condition the feature ships under: a spawned Treasure is indistinguishable from an earned one on the board, so if the log does not say where it came from, nothing does.

    Two things about it are not like the other kinds. **A spawn into a hand or a library names the zone and NOT the card** ("Luke (host) spawned a card into Ana's hand"), and that redaction happens at BUILD time rather than per viewer — the entry carries no card reference for the knower predicate to key on, so the name is dropped for everyone, the spawner included. And **`actor_is_host` is stamped by the room, not by the projection**: the host is a property of `ws.Room`, not of the engine, so `protocol.StampHostOnLog` runs on the same pass that sets `PlayerView.is_host`. It is absent on a spawn by somebody who is not the host — the dev route (`POST /games/{id}/dev/spawn`) still lets anyone at a preview table spawn — and absent on an admin spawn, which has no seat at all and renders as "The admin spawned …".

  `label` is **redacted with the card's name**, exactly as `choice` is and for the same reason: "Foretell {2}" prices the card as loudly as an `alternative_costs` entry does, which `redactCardForViewer` already strips from a face-down card. So does the `amount` of a `counters`, `saga_chapter` or `class_level` entry, because `redactCardForViewer` strips `counters` from a non-knower too — a viewer who may not identify the card reads "a card's counters changed" and gets neither the kind nor the number.

  **Players are seat indices, not UUIDs.** `seat` is the responsible player (`-1` when there is none — an SBA life loss, a spell resolving with nobody to credit) and `target_seat` is the player acted on. `card_id` and `target` are card instance IDs; exactly one of `target_seat` / `target` is set on an entry that has a target at all. `step` rides only on `step` entries: every entry after one belongs to that step until the next, and a consumer that wants the step on every line carries it forward. All of this is wire-cost discipline — the struct repeats 200 times on every snapshot frame.

  `text` is the rendered line ("Aang cast Lightning Bolt", "Turn 7 — Katara · precombat main") and is what a panel prints.

  **Combat damage steps.** `combat` (S19) marks a `damage` entry as combat damage (CR 510). `combat_step` (#187, omitempty) says which combat damage step dealt it: `"first_strike"` or `"regular"` (CR 510.4 — a combat with a first-strike or double-strike creature has two combat damage steps). It is set **only when the combat had a first-strike step**, on the damage of both steps; a combat with no first strike or double strike anywhere has untagged damage, including damage that lands later from a CR 510.1c damage-assignment prompt or a CR 616 replacement-ordering prompt. So the tag's presence alone says there are two beats to show, and a client never works out from keywords which creature dealt damage in which step. A tagged entry's `text` says so: "Fencing Ace dealt 1 combat damage to Grizzly Bears (first strike)", "Grizzly Bears dealt 2 combat damage to Fencing Ace (regular damage)"; untagged lines are unchanged. The tag rides the engine's `game.Event.CombatStep` (`combat_step` in the snapshot file), so it survives an undo, a restore and a deploy like the rest of the log; an event written before the field existed decodes untagged. A paused prompt keeps the step it was queued in: `DamageAssignmentFrame.CombatStep` is **server-side only** and is not on `PendingChoiceView.damage_assignment`. Both fields are additive and their zero value means untagged, so there is no snapshot schema bump, and `snapshot_drift_test.go` needs no new entry because `Event` and `DamageAssignmentFrame` are embedded in the snapshot by value. Engine bugs that change *which* damage is dealt, not how it is tagged: [#702](https://github.com/krakenhavoc/cmd_and_ctrl/issues/702) (a first-strike multi-blocker prompt's damage lands after the regular step's), [#715](https://github.com/krakenhavoc/cmd_and_ctrl/issues/715), [#716](https://github.com/krakenhavoc/cmd_and_ctrl/issues/716). See [ADR 0053](decisions/0053-combat-damage-beats.md) Decision 1.

  **`no_untap` (#751, omitempty).** This optional battlefield field is `{ static?, next? }`. `static` means the permanent does not untap during its controller's untap step right now. That is either a static restriction, or (#1313) a live "for as long as" hold from a resolved ability ("it doesn't untap during its controller's untap step for as long as you control Ty Lee"). A hold never appears in `next`. `next` is a deduplicated list of live player UUIDs whose next untap step the permanent skips; a controller-keyed marker is written as the current controller's ID. Face-down permanents retain `next` and keep `static` from a hold, but omit `static` from a restriction of their own (a face-down permanent has no abilities); the client shows a badge only when a tapped permanent's restriction applies to its controller and shows player names in the hover footer for any permanent carrying the field.

  **Visibility.** The log goes through the same per-viewer filter as every other field, in two places. At build time, an event whose card sits in a hidden zone at *both* ends (library → hand, hand → library) is emitted with no card reference at all — not even the instance ID, which would otherwise let a client follow a tutored card onto the battlefield — and a draw never names the card for anyone, including the drawer. At filter time, every surviving `card_id` / `target` goes through the same `known_by` predicate that redacts `CardView`, so a face-down permanent's entry reads "a card entered the battlefield" for everyone but its controller, and `text` is re-rendered to match. A log entry can never say more than the zones alongside it already do.

  **Reveals.** A `reveal` entry (#170 follow-up) is one line per reveal: the engine's per-card `reveal_cards` events that share a `reveal_seq` collapse into one entry, with `amount` the card count and `old_zone` the zone they were revealed from. It **never carries `card_id`**, for any viewer, whatever zone the cards came from; the instance ID of a card revealed out of a hand or library is the same correlation handle the build-time rule above withholds, and the knower filter cannot stand in for it because a reveal makes every seat a knower. A reveal to the whole table names the cards by printed name in `text` ("Aang revealed 3 cards from their library: Island, Swamp, Opt"), up to five names and then "and N more", and the names are not redacted per viewer: the table saw them, and a later shuffle clearing `known_by` does not un-see them (the same argument as `reveals` below). A reveal to one player sets `target_seat` and names nothing for anyone ("Aang revealed a card from their hand to Katara"); no engine path emits that shape yet.

- **RevealView** (S22, omitempty): `{ seq, turn?, seat, source?, reason?, from?, cards[], count? }` — one **broadcast reveal** (CR 701.20). `GameView.reveals` is the reveals that happened **this turn**, oldest first, at most **4** of them. Like `log` it is a projection of `Game.Events` and nothing on `game.Game` stores it, so it survives an undo, a snapshot restore and a deploy for free.

  This is the counterpart to the controller-only look-at-cards prompt that `scry` / `surveil` / `look_at_top` ride, and the two are shaped nothing alike on purpose. A look-at names one chooser and ships `CardView`s complete with instance IDs, and `filterPendingChoices` drops every option a non-chooser does not know. A reveal names no chooser, goes to every seat **identically**, is never redacted, and carries **no instance IDs at all**.

  **What a non-controller learns, and for how long.** They learn the printed identity of each revealed card — `name`, `mana_cost`, `type_line`, and `scryfall_id` (a *printing* id, so the client can draw the face). That is exactly what a player sitting at the table sees and nothing more: a reveal off the top of a library does not make the rest of the library's order public. They learn it for two different durations, and the split is deliberate. The **frame entry** lives for the rest of the turn it happened in, or until four newer reveals push it off, whichever comes first — long enough that a client which dropped a frame still catches it. The **knowledge** is permanent and is not carried by this field at all: it is the ordinary `known_by` set, which `RevealForEffect` writes for every seat, and which a shuffle clears. So a card revealed on its way to hand stays readable to the whole table in that hand; a card revealed on top of a library about to be shuffled does not.

  **No correlation handle.** `source` is the *name* of the card that revealed, never its instance ID, and it is omitted when that card is not in a zone the table can see. `RevealedCardView` has no ID field to fill in — a type with nowhere to put one cannot forget to blank it. This is the same "the instance ID alone is the leak" argument the public log settled at build time and the pending-choice options had to settle again at filter time, arriving here a third time: a reveal makes a card public *for that moment and to that extent*, and the cards usually go straight back into a hidden zone, so a stable UUID would let any client — or any bot policy, which reads the same bytes under [ADR 0033](decisions/0033-ai-bot-seat.md) §3 — recognise the card turns later in a zone it was never entitled to read.

  **Bounded.** `cards[]` is truncated to 8 entries and `count` reports the true total, so a client renders "+N more". The cap exists for Hermit Druid, which reveals a whole library when it finds no basic land; every card in the catalog that reveals a *fixed* number fits under it (Fact or Fiction's five is the largest). Truncating what is drawn never truncates what is known — the identities are public through `known_by` either way. A saturated window costs ~7 KB on a four-player frame, against the log's 32 KiB and `legal_moves`' 24 KiB budgets.

  `seq` is the engine sequence number of the reveal's first event: monotonic and stable across frames, which makes it the dedupe key a client must use, since there are deliberately no card IDs to key on. `seat` is a seat index (`-1` when there is none). `from` is the zone the cards were revealed *out of* — the cards did not move, because a reveal is not a zone change.

- **Zone abilities** (#660, widened by #1221; [ADR 0062](decisions/0062-abilities-and-special-actions-from-the-hand.md), [ADR 0020](decisions/0020-activated-abilities.md)'s 2026-09-22 addendum): `zone_abilities` is an `ActivatedAbilityView[]` on a card in a NON-BATTLEFIELD zone the viewer may activate it from — their own hand, their own graveyard, their own command zone, or their own cards in exile — listing the CR 602 activated abilities that function from THAT zone (CR 113.6): cycling and typecycling (CR 702.29) out of a hand, unearth (CR 702.82), scavenge (CR 702.96), embalm (CR 702.128) and eternalize (CR 702.129) out of a graveyard. Same shape as `activated_abilities`, same `index`, same `activate_ability` payload; a separate field because the two are read by different UI. **It is scoped to one seat, and since #1221 that is enforced rather than inherited.** Through #660 the field existed only for a hand and rode the hand's own secrecy — the per-viewer redaction strips it with the rest of the cost surface, because a row reading "Cycling {3}" names the card as loudly as its mana cost does. A GRAVEYARD is public, so the rows now ride the same per-seat carrier `castable_here` and `legal_targets` do (#1055): the seat whose card it is gets the list, every other seat gets the card and no list, and a spectator gets no list either. The two lists are disjoint by construction — the server filters each card's abilities by the zone it is in — so a permanent never carries `zone_abilities` and a card off the battlefield never carries `activated_abilities`. The LIBRARY is never stamped: CR 401.2 makes it hidden and no printed ability functions from one. **Discard cost components** ride `ActivatedAbilityView` for both: `discard_self` marks cycling's "Discard this card" (advisory; nothing to pick), and `discard_cost_n` / `discard_cost_label` / `discard_cost_options` describe a general "Discard a creature card" clause — the count, the clause as printed, and the cards in hand that could pay it right now, source excluded. Exactly `discard_cost_n` options means there is nothing to ask and the client skips its picker. The answer rides back as `discard_ids`. **#1369:** on a PERMANENT's `activated_abilities[i]` the list is read out of the controller's hand, so it reaches the controller's frame alone — every other seat, and a spectator, gets the row with its `discard_cost_n` and `discard_cost_label` and no `discard_cost_options`. For a filtered clause the list's length is how many matching cards that hand holds, and its IDs are a handle on specific hidden cards ([ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)'s 2026-09-23 amendment).

- **Zone mana abilities** (#1228; [ADR 0011](decisions/0011-auto-tapper.md)'s and [ADR 0020](decisions/0020-activated-abilities.md)'s 2026-09-23 addenda): `zone_mana_abilities` is a `ManaAbilityView[]` on a card in a NON-BATTLEFIELD zone the viewer may activate it from, listing the CR 605 MANA abilities that function from THAT zone (CR 113.6). Today that is one zone and one printed family: "Exile this card from your hand: Add {R}" on Simian Spirit Guide and Elvish Spirit Guide. `zone_abilities`' twin one ability kind over, and a separate field for the reason `mana_abilities` is separate from `activated_abilities` — the wire verb differs, so a client sends `activate_mana_ability` for a row it read here and `activate_ability` for one it read there. Same shape as `mana_abilities`, and `index` is the ability's index in the card's FULL mana-ability list, so the payload is byte-identical to a permanent's. **Scoped to one seat**, riding the same per-seat carrier `zone_abilities` does: a hand is already hidden wholesale, so today that is belt-and-braces, and it is written that way because the server's supported-zone list is one entry away from a public pile. The two lists are disjoint by construction — the server filters each card's abilities by the zone it is in — so a permanent never carries `zone_mana_abilities`, and **`mana_abilities` now means the battlefield and only the battlefield**: a Spirit Guide that reached the battlefield publishes neither list, because CR 113.6 makes "from your hand" the ability rather than a permission added to it. **The exile-this cost component** rides `ManaAbilityView.exile_self`, the same wire name `ActivatedAbilityView` carries it under (#1221): advisory, with nothing to pick — the payment IS the source — and the server validates the zone and makes the move.

- **Special actions** (CR 116.2 — foretell #658, suspend #659, turn face up #1194, [ADR 0062](decisions/0062-abilities-and-special-actions-from-the-hand.md) Decision 4 and [ADR 0082](decisions/0082-casting-face-down-and-turning-face-up.md) decision 9): a card in the viewer's OWN hand — and, since #1194, a FACE-DOWN PERMANENT the viewer controls — carries `special_actions`, an array of `{ kind, label, cost?, charged_cost?, available? }`. One row per CR 116.2 special action the card offers — `"Foretell {2}"`, `"Suspend 1—{R}"`, `"Plot {3}{U}"` — and the client fires the `special_action` verb straight from the row: there are no targets, no modes and no cost picker, because neither kind has a choice to make. `available` is the SERVER's per-kind timing answer for this frame, so a client greys the row rather than re-deriving a rule it would get backwards (foretell stays legal under split second, CR 702.61b; suspend does not, CR 702.62c). An unavailable row is greyed, never hidden: a player has to be able to see that the card has the keyword. Like `zone_abilities`, this is not public, and unlike it this one stays HAND-ONLY — foretell and suspend are announced out of a hand and nowhere else — `"Suspend 4—{U}"` names Ancestral Vision as loudly as its mana cost would — so the per-viewer redaction strips it with the rest of the cost surface. A kind the engine cannot carry out is never projected. The battlefield row is `turn_face_up` and nothing else: it is DERIVED from the face-down kind and the card underneath rather than declared by any card (CR 702.37b's morph cost for a morphed or disguised permanent, CR 701.34d's mana cost for a manifested or cloaked CREATURE card, and no row at all otherwise), so every face-up permanent on the board carries an empty list.

- **Sacrifice costs of N permanents** (#747, [ADR 0020](decisions/0020-activated-abilities.md) and [ADR 0021](decisions/0021-additional-costs.md) addenda): a sacrifice cost clause is shipped as `sacrifice_options` — a `LegalTargetsView` `{ cards[], min, max }` — in three places: `activated_abilities[i].sacrifice_options` (with `sacrifice_label`), `mana_abilities[i].sacrifice_options` (with `sacrifice_label`) and a hand card's `additional_cost.sacrifice_options` (with `additional_cost.label`). There is no separate count field. **`min` and `max` are the count and are always equal**: "Sacrifice a creature" is `1 / 1`, "Sacrifice two artifacts" is `2 / 2`, "Sacrifice five Treasures" is `5 / 5`. The answer rides back as `sacrifice_ids` on `activate_ability`, `activate_mana_ability` or `cast_spell`, and must name exactly `max` distinct permanents from `cards` (order does not matter). Anything else — too few, too many, one ID twice, a permanent the viewer does not control, one that does not match the clause, or the source itself when the cost also sacrifices the source — is refused with nothing paid. The N permanents leave the battlefield as one simultaneous event, so a "whenever a creature dies" watcher paid in with them sees every death (CR 603.10a). `cards` lists only permanents the viewer controls, from the non-targeting candidate walk (hexproof and shroud do not narrow it), without the source when the cost also sacrifices the source, and **in payment order**: tokens first, then lower mana value, then the ability's own source, then board order. That is the order the legal-move enumerator takes its one payment from (a sacrifice-N move names the first `max` of it), and the client's "Choose for me" button fills the picker with the first `max` too. A list shorter than `min` means the cost cannot be paid right now.

  **#1213: a count the ACTIVATOR announces.** `min` and `max` are no longer always equal. Two printed clauses vary:

  - **"Sacrifice one or more artifacts"** (Radiant Lotus) ships `min: 1` with `max: 0` — `LegalTargetsView`'s "no ceiling". Name any number at or above `min`; the board is the only bound.
  - **"Sacrifice X Treasures"** (Grim Hireling) ships `count_from_x: true`, and the count is the `x_value` the same message announces. Naming a different number of permanents is refused, not made cheap. Such an ability sets `demands_x: true` even when its mana component has no `{X}` at all — read the flag, never the cost string — and the announced X buys PERMANENTS rather than generic mana, so the mana owed is the same at every X.

  A fixed clause is byte-for-byte what it was: `min == max == N`. The bot's enumerator offers at most three counts for a variable clause ([docs/bot.md](bot.md)), which is policy rather than a rule; a client may name any number the bounds admit. Set-level restrictions ("with different names") still have no wire shape.

- **Return-a-permanent-to-hand costs** (#1213, [ADR 0073 amendment](decisions/0073-optional-additional-costs-and-the-cast-gate.md)): "Return a Forest you control to its owner's hand" (Quirion Ranger), "Return an artifact you control to its owner's hand" (Master Transmuter), "{1}, Return a land you control to its owner's hand" (Meloku the Clouded Mirror). `activated_abilities[i].return_label` is the clause as printed and `return_options` is a `LegalTargetsView` `{ cards[], min, max }` whose `min` and `max` are both the clause's count (1 for every printed card) and whose `cards` are the permanents that could pay right now, in the same payment order `sacrifice_options` uses — so the client reuses one picker and one "Choose for me". The answer rides back as `return_ids` on `activate_ability`: exactly `max` distinct permanents from `cards`. A TAPPED permanent is a legal pick (returning is not tapping) and so is the ability's own source when the clause admits it (Master Transmuter may return herself). An absent or empty list means the cost cannot be paid, and the engine refuses the activation (CR 118.3). The permanent goes to its OWNER's hand, not the activator's, and — being a cost — the move cannot pause: a commander returned this way takes its owner's hand rather than opening a CR 903.9 prompt (CR 601.2h / 602.2b). **#1227:** the clause may be a COMBAT one — ninjutsu's "Return an unblocked attacker you control to hand" (CR 702.49a) — in which case `cards` is empty outside the declare-blockers step and everything after it, and changes as blockers are declared. Nothing about the wire changes; a client that greys the row on an option list shorter than `min` (as both of this client's ability menus now do) is showing the player the same refusal the engine would give.
- **Waterbend costs that are not a spell's** (#1310 / #1311, CR 701.67a, [ADR 0020 amendment 2026-09-23](decisions/0020-activated-abilities.md), [ADR 0073 amendment 2026-09-23](decisions/0073-optional-additional-costs-and-the-cast-gate.md)): "Waterbend [cost]" means pay the cost, and for each generic mana in it you may tap an untapped artifact or creature you control instead. Two surfaces carry it, both in the `TapCostView` shape a hand card's convoke / waterbend already ships as `tap_cost` (`{ key: "waterbend", label, options: { cards[] }, max, demands_x? }`), so one client picker serves all three. **An activated ability** — "Waterbend {8}: Transform Aang", Katara's "Waterbend {X}" — carries `activated_abilities[i].waterbend`: `options.cards` are the viewer's untapped artifacts and creatures that could pay right now (from the walk the engine validates against; the ability's own source is among them unless its cost also prints `{T}`, and a hexproof permanent is too, because a cost does not target), and `max` is how many — the waterbend's generic, capped by what the priced cost still charges; `max: 0` with `demands_x` means "as many as the X you announce". The ability's `mana_cost` is the WHOLE cost (`{8}`, `{X}`), because paying mana for all of it is always legal. The picks ride `activate_ability` as `waterbend_ids`: zero up to `max` distinct permanents, each paying `{1}`, the rest paid from the pool and the auto-tapper, which will not also tap a permanent named here (CR 118.3). Tapping is not the `{T}` symbol, so a creature that arrived this turn may pay (CR 302.6). The taps land with the ability already on the stack; a list that names a permanent that cannot pay, more than `max`, a permanent twice, or a permanent another component of the same cost spends is refused as `bad_request` with nothing tapped and nothing paid, and so is `waterbend_ids` on an ability without the clause. **A pay-or-counter prompt** — "Ward—Waterbend {4}" (The Unagi of Kyoshi Island) — is a `pay_unless` whose `PendingChoiceView` carries `tap_cost` beside `pay_cost`, listing the CHOOSER's waterbenders. The "Pay" answer is `resolve_choice { choice_id, apply: true, tap_ids: [...] }`; the taps pay part of the generic and the pool and auto-tap pay the rest. If the rest cannot be paid, the answer is a decline with NOTHING tapped. A malformed `tap_ids` — on a prompt with no `tap_cost`, on `apply: false`, or naming a permanent that cannot pay — is refused and the prompt stays open, so a client bug never costs the payer their spell.

- **Tap-another costs** (#759, [ADR 0071 addendum 2026-09-23](decisions/0071-designations-that-switch-abilities-on.md); the component is #758's): "Tap another untapped creature you control" — the station ability's cost (CR 702.184a). `activated_abilities[i].tap_others_label` is the clause without the verb and `tap_others_options` is a `LegalTargetsView` `{ cards[], min, max }` whose `min` and `max` are both the clause's count (1 for station) and whose `cards` are the UNTAPPED permanents the viewer controls that could pay right now, in the payment order `sacrifice_options` uses — so the client reuses the same picker with the verb "Tap". The answer rides back as `tap_ids` on `activate_ability`: exactly `max` distinct permanents from `cards`. It is not the `{T}` symbol, so a creature that arrived this turn is a legal pick (CR 302.6), and paying it does not target, so a hexproof creature is too (CR 601.2h). The source is left off the list when the clause says "another", and when the same ability also prints `{T}` (CR 118.3 — it will already be tapped). An absent or short list means the cost cannot be paid and the engine refuses the activation. The ability's stack item records what was tapped (`PaidCost.TappedOthers`); station reads the tapped creature's power from it AS THE ABILITY RESOLVES, or its last-known power if it has left the battlefield (CR 608.2h). The MANA-ability half of the same component (Springleaf Drum) has no view options yet.

- **Discard costs on a MANA ability** (#1213): `mana_abilities[i]` carries `discard_cost_n` / `discard_cost_label` / `discard_cost_options` with exactly the meanings `activated_abilities[i]` gives them (#660), and the answer rides back as `discard_ids` on `activate_mana_ability`. Skirge Familiar's "Discard a card: Add {B}" is the one printed card. It is the same component with a second owner, so one client picker serves both ability kinds. The discard goes through the engine's one discard path, so `EventDiscardCard` fires per card and madness sees it; the AUTO-TAPPER never plans such a source, because which card to pitch is a decision the planner does not make. **#1369:** `discard_cost_options` reaches the controller's frame alone; every other viewer, spectators included, gets the count and the label.
- **Exile-a-card costs on a MANA ability** (#1283, [ADR 0011](decisions/0011-mana-pool-and-auto-tapper.md) amendment 2026-09-23): `mana_abilities[i]` carries `exile_cost_n` / `exile_cost_label` / `exile_cost_options` — the discard triple's shape under its own names — and the answer rides back as `exile_ids` on `activate_mana_ability`: exactly `exile_cost_n` distinct cards from the options, none of them also named in `discard_ids`. Cadaverous Bloom's "Exile a card from your hand: Add {B}{B} or {G}{G}" is the printed card. It is NOT a discard: the card goes to exile through the ordinary zone change, no discard event fires and madness does not see it, which is why it has its own fields. Like the discard, the AUTO-TAPPER never plans such a source. Not to be confused with `exile_self` (#1228), which exiles the SOURCE and sends nothing. **#1297:** the options come with `exile_cost_zone` — `"hand"` or `"graveyard"`, the pile they are in — so the client resolves them out of the right zone; a clause that names no pile is the hand, and every mana ability that prints the component today is one. **#1369:** a HAND list is read out of the controller's hand, so it reaches the controller's frame alone; every other viewer, spectators included, gets `exile_cost_n`, `exile_cost_label` and `exile_cost_zone` and no list. A graveyard list (`exile_cost_zone: "graveyard"`) is read off a public pile and goes to every viewer.
- **Exile-N-cards costs on an ACTIVATED ability** (#1297, [ADR 0020](decisions/0020-activated-abilities.md) and [ADR 0073](decisions/0073-optional-additional-costs-and-the-cast-gate.md) amendments 2026-09-23): `activated_abilities[i]` (and `zone_abilities[i]`) carry the same four fields a mana ability's exile cost does — `exile_cost_n` (the count; present marks the component), `exile_cost_label` (the clause as printed, without the verb: `"two cards"`, `"a creature card"`), `exile_cost_options` (the matching cards in the viewer's OWN hand or graveyard that could pay right now, in pile order, the source excluded; absent when nothing can pay) and `exile_cost_zone` (`"hand"` or `"graveyard"`). The answer rides back as `exile_ids` on `activate_ability`. The cards are exiled at ANNOUNCE (CR 602.2b), with the rest of the cost and before the ability is on the stack, through the ordinary zone change with the cost-cannot-pause bit — so a commander exiled this way still gets its CR 903.9 answer, no `EventDiscardCard` fires and madness does not see it. A graveyard card is not a hand payment, a hand card is not a graveyard payment, and a card that fails the clause's predicate (Tome Shredder's "an instant or sorcery card") is refused with nothing paid. The server records the exiled cards on the stack item's paid-cost record (`PaidCost.Exiled`, not on the wire), which is how an effect reads "the card exiled this way" (Holistic Wisdom). The client reuses the discard picker with the verb changed and the pile named, and skips it when the options number exactly `exile_cost_n`. **#1369:** when `exile_cost_zone` is `"hand"` the options reach the controller's frame alone — every other viewer, spectators included, gets the count, the label and the zone and no list, because the list's length and IDs are facts about a hidden hand. A graveyard list is read off a public pile and goes to every viewer ([ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)'s 2026-09-23 amendment).
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder).
- **LegalMoveView** (S31, omitempty): `{ type, player, params?, kind, label, source?, always_legal?, cost?, targets_stack? }` — one fully-specified thing the viewer's seat may do right now, produced by `server/internal/legal` ([ADR 0033](decisions/0033-ai-bot-seat.md) §1). `type` is the action type the move performs (`pass_priority`, `cast_spell`, `activate_ability`, `activate_mana_ability`, `declare_attacker`, `declare_blocker`, `declare_blockers`, `resolve_choice`, `keep_hand`, `mulligan`, `discard_selection`, `special_action`) and `params` is **exactly** the `ActionPayload.params` object that performs it — a client fires a move by sending `{ type, player, params }` unaltered, and the server is contractually obliged to accept it. `kind` coarsely groups the move (`pass` / `land` / `cast` / `activate` / `mana` / `attack` / `block` / `choice` / `mulligan` / `special_action`) so a consumer can ask "is there anything here but pass?" without parsing labels. `label` is human-readable and menu-ready (`"Cast Lightning Bolt targeting Kess"`). `source` is the instance ID of the card the move is about, when there is one — this is the join key the client greys hand cards on.
  **Own seat only, always.** `legal_moves` names cards in a hand; another seat's list would leak exactly the hidden information every other redaction in this document protects. The server enumerates per seat into an unexported map and `FilterViewFor` hands back only the viewer's own entry; spectators, admins and replay readers (empty viewer ID) get nothing, which is the one place the "empty viewer sees everything" convention deliberately does not apply. The unfiltered `GameView` that reaches the crash dump and the replay log carries no move list at all.
  **Empty is the normal case.** The list is populated only when the seat actually owes a decision — priority, a pending choice, a combat declaration, the mulligan window or a cleanup discard. A seat with nothing to do gets the field omitted, and a client must read an absent `legal_moves` as "no information", not as "nothing is legal": the safe reading when the field is missing is to stay permissive and let the server reject.
  **`always_legal`.** Present and `true` on a move the server cannot refuse whatever else happens between this frame and the click: `pass_priority`, and a pending choice's one unconditional answer (a `search_library` prompt's "fail to find", CR 701.23b; a `choose_cards` prompt's "choose nothing" when its `choose_min` is 0). Absent means "legal right now", which is what every entry in this list already promises — the flag is the stronger claim that it is still legal after the board has moved or after some other answer to the same prompt was rejected. Automated seats read it as the way out of a prompt they cannot otherwise answer: a seat owing a choice is enumerated that choice's answers and nothing else, so there is no pass to fall back on, and a bot with nowhere to go stops playing and holds the table (#544). Choice kinds that have no unconditional answer mark nothing.
  **`targets_stack`** (S35, #1307). Present and `true` on a `cast_spell` or `activate_ability` move when at least one of its chosen targets is an object already on the stack — a spell (a card in `stack`) or an activated / triggered ability (an entry in `stack_items` with no card behind it, CR 115.4). This is what lets a client tell a counterspell-shaped response (Counterspell, Stifle) apart from any other instant or ability it happens to be able to cast or activate right now, without re-deriving targeting rules client-side. Never set on `mana` or `land` moves — neither ever targets. On a modal move (Cryptic Command) the flag is per ANNOUNCEMENT: only the modes whose chosen targets are on the stack come back flagged, so a card with one counter mode and one burn mode ships both, correctly split.
  **`cost`** (omitempty): `{ life?, loyalty?, counters?: [{ card_id, counter, n }] }`. It is the part of a move's price that `params` cannot name, because the engine reads it off the ability rather than off the payload (#74, #547). It is advice about the move, never sent back. `life` is life paid at announce, and `loyalty` is a loyalty ability's signed counter delta on the source. `counters` (#625) lists the counters a "remove N counters" cost takes and which permanent they come off. That permanent is often not `source`: Heart of Kiran's alternative crew is paid with a planeswalker's loyalty counter. A policy should price the removal against that permanent, including the whole permanent when the counter is its last loyalty. **`mana`** ([ADR 0080](decisions/0080-attack-taxes.md)) is a cost string the move charges that `params` cannot name, and today that is exactly one thing: the CR 508.1a attack tax on a `declare_attacker` move, `"{2}"` for an attack into Propaganda. A cast's mana is its printed cost and a policy reads that off the CardView; an attack has no printed cost, so without this a bot prices an attack under Ghostly Prison exactly like a free one.
  **Capped, twice.** Target and mode expansion is capped at 12 concrete moves per source card (`legal.Options.MaxExpansionPerSource`), so a spell with thirty legal targets ships twelve of them. For a "search your library for up to N" prompt, and for a `choose_cards` prompt, that budget is spread ACROSS pick sizes rather than spent smallest-first, so a card that fetches a pair still offers pairs when the library holds more matching cards than the cap (#544), and a "discard two unless you discard a creature card" prompt still offers pairs when most single cards break its rule (#624). On top of that the wire projection caps the whole list at 48 moves: past that it *degrades* rather than truncates, keeping the first move of every `(source, kind, targets_stack)` tuple and dropping only the alternatives. `targets_stack` rides in that key, not just `(source, kind)`, so a capped board still keeps a modal card's counter mode alongside its non-counter mode rather than silently keeping only one. The invariant a client may rely on is therefore **"every card that has a legal move is represented by at least one entry"** — never "this is the complete set of targets". Targeting UI reads `CardView.legal_targets`, not this field.
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder). **S16.5 adds `copy_target`** — "you may have this permanent enter as a copy of ..." (Clone, Phyrexian Metamorph, Spark Double, Sakashima the Impostor). `options[]` carries the permanents that may be copied (public battlefield cards, unredacted); answered with the shared `{choice_id, card_ids}` payload, where an EMPTY list declines and the permanent enters as its own printed self. The permanent is still on the stack while the prompt is open — the answer decides what it enters AS, so its own ETB trigger has not fired yet either.
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`; `type_options` is the full creature-type vocabulary for `choose_creature_type`, materialised from the engine rather than stored on the choice. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder).
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); `color_options` is the legal-color button list for `mana_pick`. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder). **S21/S22 look-at kinds:** `scry`, `surveil` and `look_at_top` put the looked-at library cards in `options[]`, top-first. `put_in_library` (#996) puts the cards being placed there too — from a hand, a library, the battlefield or the stack — with the same chooser-only redaction, and adds `placement` (`"top"` / `"bottom"` / `"top_or_bottom"`), which stays public along with `count` — as do #1298's `top_count` and `top_depth`. A look at ANOTHER player's library (Jace's +2, Portent) projects the cards to the looker alone; the library's owner gets the prompt and its shape, not the cards. Both keywords are "look at", not "reveal" — only the chooser is a knower. The **count** stays visible to the table, which is correct: "scry 2" is a printed number, and `search_library` goes further still by dropping `options[]` outright for non-choosers, because the number of MATCHES is itself hidden information about a hidden zone. **S31 correction:** every kind now drops an option the viewer is not a knower of, rather than projecting it as a back. Redacting kept the `instance_id`, which is the right trade for a face-down permanent (the ID is already on that viewer's wire) and the wrong one for a card whose zone the seat projection has already stripped for that same viewer — a stable UUID for a specific card in a hidden zone is a correlation handle, since the card is drawn in private and cast in public under the same ID. The chooser always keeps their whole list, known or not; picking a back is a legal answer.
- **PendingChoiceView — #742 additions:** `kind` may be `choose_color` ("choose a color", CR 105.4). `color_options` carries the legal colours (a subset of `W`/`U`/`B`/`R`/`G` — "a color other than blue" is four entries, and colorless is never offered); the answer is `resolve_choice { choice_id, color }`. **#780:** such a prompt also carries `color_purpose` — `"mana"`, `"benefit"`, `"harm"`, `"filter"` or `"protect"` — the card's own declaration of what it will do with the colour it is handed. Public (it is a reading of the card's printed text) and present on `choose_color` only. It exists because every one of the five colours is a legal answer, so nothing else on the prompt distinguishes a good one. **#986:** it now drives three things rather than one — the bot seat's policy scores by it (`aiseat/heuristic`), the ORDER of `color_options` is ranked by it server-side (see the addendum below), and the client's picker WORDS itself from it (`colorPromptCopy` in `client/src/lib/manaPick.ts`): "Everything of the color you choose is hit, including your own permanents" for `harm`, "This produces mana of the color you choose" for `mana`. A prompt with no declared purpose keeps the neutral wording, and so does a purpose a client does not recognise — a card the catalog has not annotated must read as vague, never as the wrong question. A `mana_pick` may now carry `color_amounts?: { [color]: int }` — "N mana of any one color" (Gilded Lotus) is one pick that adds that many tokens of the picked colour, and the amount may differ per colour (Nyx Lotus adds your devotion to the colour you pick). A colour missing from the map adds one; the field is absent on ordinary picks.
- **PendingChoiceView — #752 additions:** `trigger_prompt` and `pick_target` may carry `doubled_by?` and `doubled_by_name?`, identifying the public permanent whose effect caused this additional trigger under CR 603.2d. The fields are omitted when no trigger doubler applies; they do not change the choice response payload.
- **PendingChoiceView — #744 additions:** `kind: "coin_call"` carries `coins` (one by default), optional `allow_stop`, `wins` (wins so far), and `max_useful_wins` (a bot hint, zero means unspecified). The chooser answers `resolve_choice { choice_id, call: "heads" | "tails" | "stop" }`; `stop` is legal only when offered. One call covers every coin in that instruction. Illegal calls leave the prompt and random stream untouched. Undo rewinds the stream: changing the call changes the displayed faces, not whether those flips win.
- **PendingChoiceView — #804 additions (CR 726):** `kind: "loop_shortcut"` is the loop breaker's shortcut proposal, queued to the controller of the repeating ability at the moment `GameView.loop_notice` goes up. `reason` carries the ability's label (`"<card> — <ability>"`), `loop_count` is how many times it has already resolved this turn, and `loop_max_iterations` is the largest answer the engine accepts (1000). The chooser answers `resolve_choice { choice_id, iterations }` — the table then resolves that ability exactly `iterations` more times with the notice down and automatic passing live, and the breaker raises the notice and re-offers the prompt on the last of them. `iterations: 0` is "stop here": the prompt clears and the table stays exactly where the breaker left it, notice standing and automatic passing suspended. The dispatcher routes this kind **by the pending choice's `kind`**, since its whole payload is an integer whose most meaningful value is the zero value. Unlike the `loop_notice` it belongs to, this prompt **blocks the table** — `advance_step`, `pass_priority` and `pass_turn` are refused while it is open (ADR 0018 §6), because the shortcut is proposed with the loop's trigger still on the stack. It is never queued to a seat that has left or been eliminated.
- **PendingChoiceView — #568 additions (CR 608.2):** `kind: "option_pick"` is "choose one of the following", asked while a spell or ability RESOLVES and frequently addressed to a seat that is not the effect's controller — Torment of Hailfire's "lose 3 life unless that player sacrifices a nonland permanent of their choice or discards a card", and the pile a Fact or Fiction chooser takes. `pick_options` carries one entry per branch in the card's printed order: `{ label, cards?: CardView[], life_cost?: int, player?: uuid }`.

  **`player` (#994, CR 800.4a)** is the SEAT a branch is about, and is present only on the prompts whose branches ARE players — "choose a player" / "choose an opponent" (#929) and True-Name Nemesis's as-enters sibling (#980). Absent on every other option. It exists because an option has to be able to name its subject: the engine re-checks an open prompt whenever the board moves under it, and a seat that carried its identity only as rendered text in `label` could not be matched back to a player, so a prompt went on offering somebody who had left the game. **An option list can therefore SHRINK while it is open** — a client holding an index across a broadcast must re-read the list rather than reuse the offset, and the server answers on the option, not the number. Public by CR 400.2, so unlike `cards` it rides no redaction and every viewer sees the same seats. The chooser answers `resolve_choice { choice_id, option_index }` — the INDEX, because an option is a consequence and not always a set of cards. The dispatcher routes this kind **by the pending choice's `kind`**, since zero (the first option) is both the field's zero value and the commonest answer. Every option listed is one the engine will accept: legality is enforced when the prompt is queued, so an option the chooser cannot take is never offered, and the FIRST option is the one `legal_moves` marks `always_legal`. `life_cost` is the same declaration `confirm` makes (#547) — a client holding only this payload must be able to tell "lose 3 life" from "discard a card". This kind **blocks the table**: the effect that asked it is paused mid-resolution (ADR 0018 §6, 2026-09-18 amendment).

  **A pile split is two prompts, not a kind.** Fact or Fiction sends a `choose_cards` prompt to the SPLITTER (an opponent) with `choose_min: 0` — an empty pile is a legal split — and, from its answer, an `option_pick` back to the controller whose two options each carry a pile in `cards`.

  **Redaction of a prompt's cards (#568, #549, PR #513).** One rule, applied to `options[]` and to every `pick_options[i].cards`. A viewer who is not a knower of a card does not receive it at all — it is dropped, not sent as a back, because a stable `instance_id` for a card in a hidden zone is a correlation handle. The CHOOSER keeps unknown cards as answerable backs only when the pool is their own material, or when it is another player's HAND (`discard_from_hand`), whose size is already public (CR 400.2). A prompt addressed to one seat over ANOTHER seat's library shows that seat only the cards a reveal made public — so a card of this family must reveal before it asks, and every printed one does. A client must therefore expect an option with a label and no `cards`, and keep its button live: that is the redaction working, not a missing render.

- **Public random outcomes — #744:** `log` kinds `roll` and `flip` aggregate one instruction into one entry, even when listeners emit other events between individual results. Rolls carry `sides` and `results: number[]`; flips carry `faces: ("heads" | "tails")[]`, and called flips also carry `call` and `wins` (omitted when zero). `text` is the complete rendered line. Results are public; source names retain normal knowledge filtering. No RNG keys, stream counters or source ordinals reach the gameplay view. The client shows fresh outcomes in the reveal strip, silently primes reconnect history, and removes rewound cues on undo.

- **PendingChoiceView.color_options order (2026-09-17, [ADR 0040](decisions/0040-mana-pipeline.md) addendum):** for a `mana_pick`, `color_options` is the source's printed colour set with the chooser's commander colour identity **listed first** (read from the chooser's commander wherever it is, so casting the commander does not change it), then the remaining colours, each group in printed (WUBRG) order. An "any color" source (Birds of Paradise, Treasure) under a mono-green commander sends `["G","W","U","B","R"]`. Nothing is narrowed away, except for a source whose printed text says "any color in your commander's color identity" (Command Tower, Arcane Signet, Commander's Sphere, Path of Ancestry), which sends only the identity's colours. Clients render the buttons in the order sent. Every listed colour is a legal answer. **#844 (CR 903.4f, 2026-09-17):** those four sources send NO `mana_pick` at all when the chooser has no commander, or a commander whose colour identity is colourless — the ability adds no mana, so there is nothing to pick. `color_options` is therefore never empty on a `mana_pick`; a client that receives an empty one should render no prompt. **#986 (2026-09-19, [ADR 0033](decisions/0033-ai-bot-seat.md) amendment):** a `choose_color` prompt's `color_options` is ordered too, by the prompt's own `color_purpose` — `harm` by what the opposition loses NET of what the chooser loses, `filter` by what the opponents have on the battlefield, `protect` by the greatest POWER among the creatures other seats control of that colour, and `mana` / `benefit` / undeclared by what the chooser already controls. Ties keep the list's own order (WUBRG, or the card's narrowed list). Every term is a count of permanents on the battlefield, so the order is public information and identical for every viewer, and it never narrows the list — all five (or all four of "a color other than blue") are still offered. One function does it, `legal.OrderColorOptionsLocked`, and the bot seat's move list is built from the same call, so the first button and the first enumerated answer are always the same colour. Clients render in the order sent, here as everywhere. A `choose_color`'s `color_options` is never empty either, and for a sharper reason than the `mana_pick` rule above: an unanswerable prompt BLOCKS, so a card that narrowed its list to nothing (`"C"` alone, a typo'd letter) would stall the table rather than ask a bad question. `normaliseColorOptions` now falls back to all five when the narrowing drops every entry — CR 105.4 says the answer is one of the five — and logs one warning per process naming the card bug.

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

`GET /games/{id}/auto-tap-preview?card=<uuid>&from_zone=<zone>&alternative_cost=<key>&optional_costs=<i>,<i>&tap_ids=<uuid>,<uuid>&face=<int>&x=<int>&phyrexian=<int>&ability=<int>&targets=<kind>:<uuid>,...&exclude=<uuid>,<uuid>...`
returns a read-only projection of what the auto-tapper would do for
a given cast — the client uses it to render the `AutoTapPreviewModal`
before committing a `cast_spell` with `auto_tap: true`. No game state
mutates. Authorization: any seated player at the named game; spectators
and non-seated callers receive 403.

**The endpoint prices the ANNOUNCEMENT, not the card (#696).** Every
cast-shaped param below is passed straight into `game.CastSpellParams`
and handed to `game.Game.PriceCast`, the engine's one cast pricer, so
the preview and `CastSpell` charge the same thing. The client builds
them from the cast payload it is about to send (`castPreviewParams` in
`client/src/lib/castPreview.ts`); omitting one asks about a different
cast. The endpoint used to parse the card's printed cost and re-apply
the commander tax and the cost modifiers itself, which knew nothing
about the alternative cost claimed at announce, a granted permission's
flat override, the "spend mana as though any colour" fold, the optional
additional costs or the face — so it disabled "Auto-tap & cast" on
casts that would have gone through, and planned taps for casts that
would fail.

| Query param | Required | Notes |
|---|---|---|
| `card` | yes | Instance UUID of the card to plan for. |
| `from_zone` | no | The zone the cast comes out of — `hand` (the default), `command`, `graveyard`, `exile`, `library`. Decides the commander tax, which cost modifiers see the cast, and which granted permission prices it. Must match the `cast_spell` payload the confirm button will send. |
| `alternative_cost` | no | The CR 118.9 cost the cast claims (`flashback`, `overload`, `evoke`). Empty means the printed cost. A key the card does not offer from that zone is a 400 — the same refusal the cast itself gets. |
| `optional_costs` | no | Comma-separated positions in the card's `optional_costs` (ADR 0073), repeated once per payment for a multikicker, so a kicked cast is previewed at the kicked price. |
| `tap_ids` | no | Comma-separated permanent UUIDs being tapped for convoke or waterbend. They pay part of the cost, and the plan must not tap them again for mana. |
| `face` | no | The printed face being cast (ADR 0034). A modal DFC's back face has its own mana cost. Defaults to 0, the front. |
| `x` | no | Caller-supplied X value for spells with `{X}` in their cost. Defaults to 0. |
| `ability` | no | Price the card's CR 602 activated ability at this index instead of its cast cost. Every cast-shaped param above is ignored on this branch — an ability is not a cast. #1405: priced by `Game.PriceActivation`, the function `ActivateCatalogAbility` charges, so board activation modifiers (Boom Scholar) and the ability's own cost clause (the channel lands, Dragonfire Blade) are in the plan and the `missing` list. No commander tax. |
| `targets` | no | #1405, with `ability` only — the targets the activation announces, as comma-separated `card:<uuid>` / `player:<uuid>` entries. A price that reads the target (Dragonfire Blade's "{1} less for each color of the creature it targets") is previewed at that target's price. Omitted, the no-target price, which is what the activation charges when it names no target and what the X and Phyrexian pickers want, since they open before targeting. Any other kind, or a malformed UUID, is a 400. Slot and mode are not carried; no cost reads them. |
| `sacrifice_ids` / `discard_ids` | no | #1242 — the permanents and cards the cast names to its additional cost's sacrifice and discard. They do not change the price; the plan must not ALSO spend them on mana (an Eldrazi Spawn offered to Village Rites, a Spirit Guide offered to Thrill of Possibility), exactly as `CastSpell`'s auto-tap will not. The same holds for `tap_ids`, which the plan now excludes too. |
| `exclude` | no | Comma-separated permanent UUIDs the auto-tapper must NOT consider — the lock-tap UI's reservation list. |
| `phyrexian` | no | #916 — how many of the cost's Phyrexian symbols the announcement will pay with 2 life each (CR 107.4f). Those symbols are struck before planning, exactly as the engine strikes them, so the plan and the `missing` breakdown describe the MANA the announcement still owes. Defaults to 0. Clamped to the number the cost prints rather than rejected: refusing a malformed announce is the announce gate's job, not a read-only preview's. |

Response shape:

```json
{
  "ok": true,
  "plan": ["<uuid>", "<uuid>"],
  "sources": [
    { "card_id": "<uuid>", "name": "Mountain", "zone": "battlefield", "tap": true },
    { "card_id": "<uuid>", "name": "Simian Spirit Guide", "zone": "hand", "exile": true }
  ],
  "missing": null,
  "cost": "{2}{R}{R}"
}
```

`ok` is true when the auto-tapper found a satisfying plan; `plan` is
the ordered list of source IDs the payment spends. **#1285:**
`sources` is the same list in the same order, DESCRIBED — a planned
source is no longer always an untapped permanent. `zone` is where it
is (`battlefield`, or `hand` for a Spirit Guide, #1228); `tap`,
`sacrifice` and `exile` say which components paying with it takes: a
land is `tap`, a Treasure `tap` + `sacrifice`, a Gold or an Eldrazi
Spawn `sacrifice` alone (#1242 — it is never tapped), a Spirit Guide
`exile`. `name` is the card's name; the preview is only answered for
the asking seat's own sources, so a hand card's name is the asker's
own information. A client should show every entry — the modal names
each card that leaves the hand — rather than looking the IDs up on
the battlefield, where a hand source is not. When `ok` is false, `plan`
is omitted and `missing` carries the unpaid mana symbols (same shape
as the `insufficient_mana` error frame's `missing` list). `cost` is
the cost string this cast PAYS — the claimed alternative cost, a
granted permission's override, or the card's printed cost when neither
applies — and the modal renders it next to the plan for context. The
commander tax and the cost modifiers are not folded into that string
(they are generic, and the string is Scryfall brace notation); `plan`
and `missing` are the authority on the total and are computed with
both. Errors: 400 on missing / malformed query params, and on a cast
the engine cannot price (an unclaimable `alternative_cost`, an unknown
`from_zone`, a face the card does not offer); 404 when the game or
card doesn't exist.

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

### Granted cast permissions and the top of your library (S42)

Four additive changes, none of them breaking (`v` unchanged):

- **`cast_spell.from_zone` accepts `"library"`** — CR 401.5's "you may
  play lands and cast spells from the top of your library". Only the
  TOP card is ever legal, and only while something grants both the
  permission and the visibility; the server refuses anything else. The
  library is the caller's own for every card in the catalog, and since
  #1035 may be another seat's under a permission that names it — see
  the `from_zone: "library"` note below. `"hand"`, `"command"`,
  `"graveyard"` and `"exile"` are unchanged.
- **`castable_here` and `alternative_costs` on a graveyard or library
  card now include GRANTED permissions**, not only the ones a card's
  own text prints (ADR 0066). A card Snapcaster Mage gave flashback to
  renders the same cast affordance a Faithless Looting does, with the
  synthesised offer in its picker. The offer list is computed for the
  zone's OWNER and is public; `castable_here` is the VIEWER's own
  answer (#1055), which covers both the owner's printed cast and the
  case where somebody else holds the permission (#1022 below).
- **An opponent's `library.cards` may now carry exactly one card** —
  the top one, when "play with the top card of your library revealed"
  (Oracle of Mul Daya, Courser of Kruphix) is in force and the viewer
  is a knower of it, or — since #1035 — when the viewer holds a cast
  permission over that library top whose grant carries CR 401.5's look
  (Xanathar, Guild Kingpin). Every other library card stays hidden,
  `count` stays public as before, and a one-shot "reveal the top two
  cards" does NOT open the zone: it makes those cards known without
  making them visible where they sit.

- **An exiled card a permission opens now carries the same announce
  surface a hand card does** (#978): `legal_targets`, `clauses`,
  `modes`, `additional_cost`, `tap_cost`, `target_cost_notes`,
  `phyrexian_symbols` and `alternative_costs`. It did not before,
  because those stamps walked hand, the command zone, the graveyard and
  the library top — all per-seat zones — and exile is a shared
  top-level one, so the client's cast chain had to fall back on
  heuristics for an impulse cast. The gate is the engine's own
  castability predicate (`CastPermissionForLocked`), the same one
  `cast_spell` validates with and the bot enumerator reads.

  **These fields are the GRANT HOLDER's alone.** A legal target set is
  narrowed by hexproof, shroud and "target opponent", so one seat's
  answer is not another's: every other viewer, spectators and admins
  included, gets the card and its public `exile_play` and none of the
  stamps. That is a per-viewer projection of a shared zone, which is
  new — everywhere else, "this stamp is yours" was implied by the zone
  being yours. A grant whose window has not OPENED yet (warp's
  `not_before_turn`) carries the grant and no stamps, because there is
  no cast to target for yet.

  A permission that names a `face` is read against THAT face's catalog
  entry, so a defeated Siege's back-face cast ships the back face's
  modes and target clause rather than the battle's — and so are
  `optional_costs` (ADR 0073), which ride the same key.

  `cant_cast` (ADR 0073 §7) reaches exile and the library top with
  them, because they reach the same stamping function: a permission
  opens a ZONE and a restriction shuts the cast anyway (CR 101.2).
  `castable_here` now respects it — a granted graveyard or library
  card the gate refuses is no longer marked a cast surface.

**A graveyard card somebody ELSE may cast is stamped for THEM (#1022).**
A `ScopeCards` permission names an object, not a pile — Wrexial's "you
may cast target instant or sorcery card from that player's graveyard" —
and until #1022 such a card reached the wire with nothing on it at all:
the per-seat walk asked "may the zone's owner cast this" and no other
question. The graveyard stamp is now **per holder**:

- the zone's owner gets the PUBLIC answer when the owner may cast the
  card at all (their own permission, or the card's printed flashback /
  escape), which is every case that existed before #1022;
- every OTHER seat that holds a permission over it gets its own copy of
  the stamps, computed for them and delivered on their frame alone —
  `FilterViewFor` drops them for every other viewer, the owner of the
  graveyard and spectators included. Until #1037 that was one seat (the
  first in seat order) and one set of stamps; see below.

**Every holder gets their own stamps (#1037), and the wire shape is
unchanged.** Until #1037 a `CardView` carried ONE seat's answer — the
graveyard took the first seat in seat order and exile the first live
permission — so a card two seats may both cast (a flashback on a card
in your own graveyard under somebody else's Wrexial-style grant, two
impulse grants over one exiled card) left the second holder with the
public zone, the public `exile_play` and no picker. The frame is built
per viewer already, so nothing on the wire had to grow: the server
computes each holder's answer, files it under their seat, and
`FilterViewFor` promotes exactly one.

- **The exported fields carry the answer that is PUBLIC.** Since #1055
  that is the half which is a fact about the CARD IN THIS ZONE rather
  than about a player: `alternative_costs`, `alternative_cost_required`,
  `modes`, `additional_cost`, `optional_costs`, `tap_cost`,
  `target_cost_notes`, `phyrexian_symbols` and `cant_cast`, computed for
  the pile's owner over public state, because a card in a graveyard is
  a card every player may pick up and read. For exile there is no owner,
  so nothing is public there but the grant.
- **A holder's answer replaces it, on their frame only.** Every field
  in the list above travels together, `castable_here` included.
- **`castable_here`, `legal_targets` and `clauses` are never in the
  public half** (#1055). All three answer "what may YOU announce" — the
  first one says so in its name, and the other two are narrowed by
  hexproof, shroud, protection and "target opponent", so one seat's set
  is not another's to read. A viewer with no answer of their own gets
  the card, its public price list, the public `exile_play`, and no cast
  surface.
- **`exile_play` is resolved for the viewer.** It stays PUBLIC and
  still names a seat, but a viewer who holds a permission over the card
  gets THEIR OWN grant rather than whichever live one came first — its
  `cost_override`, its `faces`, its `any_color`. A client that read
  somebody else's would render the wrong button for a cast it is
  allowed to make.
- **A non-knower gets neither.** The stamps name a card as loudly as
  its mana cost does, so the redaction runs first and a holder who
  cannot read the card keeps nothing.

**Reading `castable_here` on a card in somebody else's pile.** The bit
is YOURS (#1055). "The server marked it" IS "I may cast it", in every
pile and for every viewer, so both client readers (`castableFromZone`
for the zone browser, `libraryTopPlayable` for the library top) are
`card.castable_here === true` and nothing else. Until #1055 it was
public and carried the pile OWNER's answer, and each reader had to
pair it with "or `exile_play` names me" to find out which of the two
it was holding.

The engine reaches the same card through the same permission. CastSpell's
`from_zone: "graveyard"` resolves to the caster's own pile and, when the
card is not in it, to the pile it IS in — **only** under a permission
the caster holds (`castSourceZoneLocked`); the bot enumerator walks
every seat's graveyard once anything has granted anything, under the
same rule. A card's own text opens only its OWNER's graveyard, because
flashback, escape and Gravecrawler all print "your graveyard", and a
STANDING permission (Underworld Breach) is scoped the same way for the
same reason — `CastPermissionForLocked` says so now that anything can
ask it about another seat's pile.

**`from_zone: "library"` is the same sentence since #1035.** CR 401.5's
"the top card of your library" used to be checked against the HOLDER's
own library, which scoped every printed library clause by accident and
made a cross-seat permission open nothing at all. The position rule now
reads the library the CARD is in, so Xanathar, Guild Kingpin's "you may
play the top card of their library" is expressible: the cast path finds
the pile the card is in under a permission the caller holds, the
enumerator walks every seat's library TOP under a grant, and the view
stamps that seat's offers on the top card of the library it names. Two
rules ride with it:

- a STANDING library permission with no seat named is "YOUR library",
  exactly as a Breach is "your graveyard" — two Coursers of Kruphix do
  not play lands off each other's revealed top card;
- CR 401.5's other half, the LOOK, comes with the grant. "Play with the
  top card of your library revealed" and "you may look at the top card
  of your library" are per library and derived from the permanents its
  owner controls; "you may look at the top card of THEIR library" is
  per pair and granted by a resolution, so it rides the permission. A
  cross-seat grant without it opens nothing, and one with it puts that
  one card on the holder's wire — the second way an opponent's
  `library.cards` can carry exactly one entry.

`exile_play` is unchanged in name. It is now projected from the
per-player permission store rather than from a field on the card,
which is invisible on the wire. Its `face` key became `faces` in #719
— see below.

**`alternative_costs` lists only offers the caster could pay (#695).**
An offer is stamped when its condition holds, when the caster's life
total is at least its `life` (CR 119.4 — exactly N still appears,
because paying down to zero is legal) and when the board holds enough
cards for its card component (CR 601.2b). `pay_options` is therefore
never present-and-empty any more: an offer with nothing to pay it is
not offered. Mana is deliberately NOT part of the filter — CR 601.2g
lets the caster tap for it after the cost is chosen, so an offer the
seat cannot currently afford is still shown and the auto-tapper is
what answers it. The same predicate decides what the bot enumerator
offers and what `cast_spell` accepts, so a rendered offer is one the
server will take.

## Optional additional costs and the cast gate (S42, ADR 0073)

Two additive fields on `CardView` and one on `cast_spell`, both
readable by every client that ignores them.

- **`cast_spell.optional_costs`** (`[]int`, omitted when empty) is the
  CR 601.2b announcement of which optional additional costs the caster
  is paying — kicker, multikicker, buyback (#664). They are POSITIONS
  in the card's `optional_costs`, not keys, because a position is what
  the server's paid record holds. **Paying one N times is naming its
  index N times**: a Wolfbriar Elemental kicked three times sends
  `[0, 0, 0]`, which is how multikicker (CR 702.33d) announces its
  count without a second field. The server normalises the list to
  ascending order, so two clients that picked the same costs in
  different orders produce the same announcement.

  The MANA half of a claimed cost joins the total at CR 601.2f, after
  the alternative-cost swap and the commander tax and before the cost
  modifiers. The CARD half rides the existing flat `discard_ids` /
  `sacrifice_ids` lists, walked against one ordered payment plan — the
  mandatory additional cost first, then each claimed optional cost in
  index order. A claim the card does not offer, or one naming a
  once-only cost twice, is a rejected cast rather than an ignored
  field.

- **`CardView.optional_costs`** is the offer side: `index`, `key`
  (`"kicker"`, `"multikicker"`, `"buyback"`), the printed `label`, the
  `mana_cost`, `max_times` (1 for kicker and buyback, the multikicker
  cap above that — which is what turns the client's checkbox into a
  stepper), and the card-shaped halves `discard_cards` and
  `sacrifice_options` in the same shape `additional_cost` gives them.
  A present-and-empty `sacrifice_options` means the offer cannot be
  taken right now (a Constant Mists with no land), exactly as it does
  for a mandatory cost.

  Unlike `alternative_costs` these COMPOSE with the radio list: an
  alternative cost replaces the mana cost and an optional one adds to
  whichever cost is being paid, so the client renders them as toggles
  inside the same picker rather than as a prompt of their own.

- **`CardView.cant_cast`** is the printed clause that stops the card
  being cast from the zone it is in right now (#760) — "Each player
  can't cast more than one spell each turn", "Cast this spell only if
  you control a legendary creature or planeswalker". Absent, which is
  nearly always, means nothing refuses the cast. Since #1316 the same
  bit and the same clause also cover a per-player ban a resolved spell
  GRANTED with a duration ("Until your next turn, your opponents can't
  cast spells from anywhere other than their hands", Avatar's Wrath) —
  the source card is often gone (exiled) by the time the ban is read,
  which is exactly why it is stored rather than derived; the clause
  itself is unaffected and still travels as plain text.

  It is the STAMP of the one announce-time cast gate that `CastSpell`
  and the bot enumerator both call, so a card carrying it is one the
  server WILL refuse: grey it and show the clause rather than
  dispatching `cast_spell` and surfacing a toast. `castable_here` is
  cleared alongside it. PUBLIC — a Rule of Law on the battlefield is
  visible to everyone, and the clause is printed on a card in a public
  zone — which since #1055 `castable_here` is NOT: the clause is a fact
  about the card, the bit is an answer about you. Cleared with the rest
  of the cost surface on the non-knower redaction, because a
  legendary-sorcery clause says more about a face-down card than its
  mana cost does.

  A cast that races the stamp (the board changed between the snapshot
  and the click) comes back as `bad_request` with the clause in the
  message.

**`exile_play.face` became `exile_play.faces` (#719, CR 715.4).** The
field says which printed faces a grant opens, and it had to stop being
a single `omitempty` integer for the reason the trap below names: zero
is both "this grant does not speak about faces" and a real face index,
and an adventure card's creature half — the one half CR 715.4 opens
from exile — IS face 0. A list says "none" by being absent and "face
0" by being `[0]`. Absent still means the card's own layout decides,
which is what `faces` and `layout` already tell the client. Present
means those faces and no other; every grant anything declares today
names exactly one. Breaking within `v` only in the sense that a reader
of the old key sees nothing rather than something wrong — the field
was advisory, because the server settles the face from the grant
rather than from the request.

### Gift (#1267, ADR 0089)

Gift (CR 702.174) is one more optional cost, keyed `"gift"`, whose
payment is choosing an opponent rather than paying anything. Three
additive fields; a client that ignores them never promises a gift and
casts every gift card as its printed ungifted spell.

- **`cast_spell.gift_opponent`** (player ID string, omitted when no
  gift is promised) is the opponent the gift is promised to. It rides
  WITH the gift offer's index in `optional_costs`: required exactly
  when that index is announced, and a rejected cast — never an ignored
  field — when sent without it, when it names the caster, or when it
  names a player who has left the game.

- **`OptionalCostView`** gains `chooses_opponent` (true on a gift
  offer) and `opponent_options` (the player IDs the gift may be
  promised to — every other player still in the game; absent on a
  gift offer means nobody is left and the offer cannot be taken). It
  also gains the target-clause trio `target_mode` / `legal_targets` /
  `clauses`, in exactly the shape `alternative_costs` gives cleave: the
  clause the spell has WHEN THIS COST IS PAID (CR 702.174m — Long
  River's Pull's promised "target spell", Wear Down's two targets,
  Valley Rally's target that only a promised cast has). Absent when
  paying the cost leaves the card's own clause alone, which is every
  kicker and buyback. Per viewer, like every legal set.

- **`StackItemView.gift_to`** (player ID string) is who a spell on the
  stack promised its gift to; absent when it promised none. Public: a
  responder needs it, because a promised Long River's Pull counters
  any spell and an unpromised one only a creature spell.

## One list of cast prices, and the printed cost's place in it (#1012, #1015)

One additive field on `CardView`, and a sharper meaning for two that
were already there. A client that ignores the new one behaves exactly
as it did.

- **`CardView.alternative_cost_required`** (bool, omitted when false)
  says the PRINTED mana cost is not one of the prices this cast may
  claim out of the zone the card is sitting in, so the caster must name
  one of `alternative_costs`. A Faithless Looting in the graveyard is
  castable at its flashback cost and at nothing else (CR 702.34a), and
  a card whose zone a PERMISSION prices — a Snapcaster'd instant, a
  library top under Bolas's Citadel — is the same shape.

  Absent, which is every hand cast, every command-zone cast and a
  Gravecrawler whose graveyard permission carries no price, means the
  printed cost is on the menu as usual.

  Read it in the cost picker: the "its mana cost" row is not an option
  the player declined, it is one the card does not offer from here, so
  it is DROPPED rather than greyed, and the picker's default becomes
  the first offer. Before this field the client inferred the answer
  from the shape of the offer list, which is right for the two common
  cards and wrong for a Gravecrawler under an Underworld Breach, where
  the printed cost and the granted escape cost are both live.

  Stamped and stripped with `alternative_costs`: it is only meaningful
  beside the list it qualifies, so a bystander who gets no offers for
  an exiled card gets no flag either, and a non-knower's redaction
  clears it.

- **`castable_here` is now derived from the price list and the cast
  gate, and from nothing else** (#1015). It used to be set the moment
  a card's own text or a permission opened the zone, and never revisited
  when the offers that followed came back EMPTY — so an escape card in
  a graveyard too small to pay for it rendered a cast button with no
  offer behind it and the announce path refused the click with
  "an alternative cost must be claimed to cast this card from here".
  The bit is now `no cant_cast && at least one claimable price`, which
  covers the #978 gate case as a special case of the same sentence.

  Unchanged: it is never set on a hand or command-zone card (both are
  cast surfaces for everything in them), and exile keys its button off
  `exile_play`. The bit was still PUBLIC after #1015; #1055 below is
  where it stopped being.

## `castable_here` is the viewer's own answer (#1055, 2026-09-21)

One narrowing on `CardView`, additive in the `omitempty` sense — the
field goes out on FEWER frames, never on more, so a reader that already
falls back to `false` for an absent bit needs no change and `v` does
not move.

**`castable_here` now means "YOU may cast this from here", on every
frame.** It is stamped only for a seat that may actually make the cast
— the pile's owner for a printed flashback or escape, the holder of a
`CastPermission` over the card for a granted one, both of them on their
own frames when both are true — and is absent for everybody else,
spectators and admins included.

It used to be public and to carry the PILE OWNER's answer, which is two
different sentences depending on the card: "the owner may cast this",
which is not a statement about the viewer at all, and "YOU may cast
this" on a card the viewer held a grant over. Both client readers
therefore had to read a PAIR — the bit, and whether the (also public)
`exile_play` named this viewer — to answer the only question they
actually had. That is the S48 [#891](https://github.com/krakenhavoc/cmd_and_ctrl/issues/891)
shape: a field whose NAME is a statement about the viewer and whose
VALUE was a statement about somebody else.

**What stays public, and why.** The line is "is this a fact about the
card in this zone, or about a player". A card in a graveyard, or on top
of a revealed library, is a card every player may pick up and read, so
what it prints stays on every viewer's copy: `alternative_costs` and
`alternative_cost_required`, `modes`, `additional_cost`,
`optional_costs`, `tap_cost`, `target_cost_notes`, `phyrexian_symbols`
and `cant_cast`. They are computed for the pile's owner, over public
state — an escape offer is priced by the size of a graveyard everybody
can count.

`legal_targets` and `clauses` go with `castable_here` instead: hexproof,
shroud, protection and "target opponent" all narrow a target set by WHO
IS ASKING, so one seat's list is not another's to read. The server has
said exactly that about a spectator since #978; this is the same
sentence applied to the bystander a public stamp used to reach.

**For a client.** Both readers collapse to the bit:
`castableFromZone(card, zoneKind)` and `libraryTopPlayable(zone)` no
longer take the viewer's or the owner's seat. `exile_play` keeps its
own job — it is public, it names the seat the permission was granted
to, and the impulse button reads it for the grant's face, cost and
label — but it is no longer part of ANSWERING whether this viewer may
cast. A face swap (`cardAsFace`) replaces `castable_here` with the
chosen face's own, along with the rest of the announce surface — see
the per-face section below.

**The hand and the command zone answer per viewer too (#1166,
2026-09-21).** The same narrowing, on the two zones #1055 left alone.
The command zone is PUBLIC — every seated player is a knower of every
card in it — and nothing stripped it, so a commander with a target
clause put its owner's `legal_targets` on all four frames for a cast
CR 903.4 gives one seat. A REVEALED hand card (Thoughtseize, Telepathy)
is the same shape with a narrower audience. Both are routed through the
same per-viewer promotion every other cast surface uses now, so
`legal_targets` and `clauses` reach the hand or command zone's OWNER
and nobody else. `castable_here` was never involved: the server only
ever sets it for a graveyard or a library top, because every card in a
hand or a command zone is a cast candidate and an always-true flag
would be noise. A hand is still not a public zone, so a revealed card
also drops the cost-shaped announce fields on a non-owner's frame —
see the section below, which is where that rule now lives.

## Announce data per castable face (#992, 2026-09-21)

**`faces[i]` carries the same announce block `CardView` does.** A card
can be TWO castable objects (ADR 0034): a modal DFC's faces are
independently playable (CR 712.11b), and CR 715.3 lets an adventure
card's caster choose the creature or the Adventure. The wire carried
exactly one announce block — for the face that is UP, which for a card
in hand is always face 0 — so a client that let the player pick the
other half had to CLEAR it, because the front's answers are actively
wrong attached to the back. A targeted Adventure half (Stomp, Petty
Theft, Swift End) therefore reached `cast_spell` with no target picker
ever opening.

Each entry of `faces` now carries, with the same keys and the same
meanings they have on the card: `target_mode`, `legal_targets`,
`clauses`, `modes`, `additional_cost`, `optional_costs`, `tap_cost`,
`alternative_costs`, `alternative_cost_required`, `target_cost_notes`,
`phyrexian_symbols`, `cant_cast` and `castable_here`.

- **Only on a face a cast may actually choose.** That is
  `game.CastableFaces` — both halves of a modal DFC and of an adventure
  card, the front alone of a transform card — narrowed by a grant that
  NAMES faces, because CR 715.4's Adventure permission opens the
  creature and no other. On every other face the fields are simply
  absent, which reads correctly as "this half announces nothing". A
  single-faced card ships no `faces` at all, so this costs the ~33,000
  ordinary oracle IDs nothing.
- **The per-viewer split travels down with it.** `castable_here`,
  `legal_targets` and `clauses` answer "what may YOU announce" for a
  face exactly as for a card, so they reach one seat's frame; the rest
  of the block is public on a public zone under the same rule as above.
- **`exile_play` stays on the CARD.** A permission is granted over an
  object, not over a face of one; which face it opens is expressed by
  `exile_play.faces` and by which faces carry a block at all.
- **For a client.** `cardAsFace(card, i)` swaps `faces[i]`'s block in
  rather than clearing the card's, and the rest of the cast chain is
  unchanged — it reads `target_mode` and `legal_targets` off the card
  it was handed, which after a face pick is the face. `zone_abilities`
  and `zone_mana_abilities` are still cleared rather than swapped:
  cycling is an ability of the CARD IN HAND (CR 702.29a), not of a
  face being cast, and so is a Spirit Guide's "Exile this card from
  your hand: Add {R}" (#1228).

Additive: a reader that ignores the new keys behaves exactly as it did.

- **`alternative_costs` is `game.CastOffersForLocked`'s answer**, the
  same list the bot enumerator walks and `cast_spell` validates against
  (#673). The view used to build its own, and the two disagreed twice:
  the offer a permission synthesised was stamped first and then
  OVERWRITTEN wholesale by the card's printed set, so a card that both
  prints and is granted a price showed only the printed one; and the
  precedence between them was stated in two places. Both are now one
  read, so the picker and the move list are the same set by
  construction. A granted key the card also prints is dropped rather
  than listed twice, which is the precedence
  `resolveAlternativeCostLocked` applies at announce.

## A hand's public announce half, and the face that opens a zone (#1169, #1171, 2026-09-21)

Two narrowings on `CardView` and `faces[i]`, both in the `omitempty`
direction that is safe — a field goes out on fewer frames, never on
more — so `v` does not move.

**A revealed hand card carries four announce fields and no more
(#1169).** A hand is not a public zone the way a graveyard is: a
revealed card (Thoughtseize, Telepathy) is ONE card the viewer has
been shown, not a pile they may read. What a non-owner gets on it is
`target_mode`, `additional_cost`, `optional_costs` and `cant_cast` —
the fields that are not cost-shaped: a printed prompt shape, two
printed clauses whose pickers read the public battlefield, and a Rule
of Law everybody can see on the battlefield anyway. Everything else in
the announce block — `modes`, `alternative_costs`,
`alternative_cost_required`, `tap_cost`, `phyrexian_symbols`,
`target_cost_notes`, plus the three that were never public anywhere —
reaches the hand's OWNER and nobody else.

That is what the wire already did for the card, with two gaps this
closes:

- **`faces[i]` is narrowed the same way.** #992 publishes the announce
  block per castable face; the hand's narrowing was a list of field
  names in the per-viewer filter and never reached a face at all, so a
  knower of a revealed adventure card or modal DFC read its owner's
  whole per-face price list. An offer with a pitch cost (Force of
  Will's "exile a blue card from your hand") carries `pay_options`,
  which is a list of instance IDs out of the hand the viewer was shown
  exactly one card of.
- **`optional_costs` stays, deliberately.** It was never in that list —
  it was added after it (ADR 0073) and nobody updated it — and it is
  printed text with a public picker, so it is allowlisted rather than
  quietly dropped.

The rule is an ALLOWLIST now, in the one function that knows the field
list (`castStamps.publicIn`), so an announce field added to the wire
tomorrow is private in a hand until somebody decides otherwise. A
client reading these fields off somebody else's hand card should not:
they describe a cast only that seat may announce.

**`castable_here` and the price list answer for every castable FACE
(#1171).** "Which zones does this card's own text open" is asked of
every face a cast may choose — `game.CardCastableFromAnyFace`, the
predicate the bot's legal-move enumerator reads — and no longer of the
card's bare `oracle_id`, which is face 0's catalog entry (ADR 0034
keys a back face `<oracle_id>#N`).

Nothing on the wire changes for a single-faced card, or for any card
in the catalog today. It changes for the first card whose BACK face
prints flashback, escape or a Gravecrawler-shaped permission: such a
card was a legal move the server would have accepted and a card with
no announce stamps at all on the wire — no `castable_here`, no price
list, no target clause. **For a client:** the answer for a multi-face
card is the union over the card's block and its `faces[i]` blocks —
the front half of a card whose BACK opens the graveyard is not a cast
surface and the back half is, so a zone browser that reads the card's
`castable_here` alone will not offer a cast the server would accept.
This client's does (#1173, via the shared `castableFaces` walk in
`client/src/lib/faces.ts`), and hands the face picker the face the
union answered yes on rather than defaulting it to the front.

## A public announce field's nested legal sets are still per viewer (#1172, 2026-09-22)

One more narrowing, one level in, in the same safe direction — fields
go out on fewer frames, never on more — so `v` does not move.

`modes` and `alternative_costs` stay PUBLIC on a public pile, for the
reason the #1055 section gives: a card in a graveyard or on a revealed
library top is a card every player may pick up and read, and "choose
two of three" and "Overload {4}{R}" are what it prints. Each of them
carried board-derived lists of its own, computed for the seat the stamp
was built for and shipped inside the public field to everybody:

| field | what it is | who it now reaches |
| --- | --- | --- |
| `modes[i].legal_targets` | the bullet's legal set right now | the asking seat |
| `modes[i].clauses` | the bullet's per-clause legal sets (#764) | the asking seat |
| `alternative_costs[i].legal_targets` | the clause the spell has when this price is paid | the asking seat |
| `alternative_costs[i].pay_options` | the cards that can pay the cost's card-shaped half | the asking seat |
| everything else in both blocks | printed text: prompt, min / max, labels, `target_mode`, keys, costs, `x_locked_at_zero`, `phyrexian_symbols` | every viewer of the pile |

The rule that decides a row is the #1055 rule one level down: **does
this field come off the printed CARD, or off the board for one
PLAYER.** Hexproof, shroud, protection and "target opponent" narrow a
target set by who is asking, and "a blue card in your hand" / "the
cards in your graveyard" name one seat's cards — so a bystander reading
either was reading the pile owner's answer out of a field whose name
says it is theirs, which is the S48 [#891](https://github.com/krakenhavoc/cmd_and_ctrl/issues/891)
shape.

**The nested fields ride the same per-seat split as the top-level
ones.** There is no new mechanism and nothing new on the wire: the
public projection (`castStamps.publicIn`) copies the two blocks and
drops the four lists, and `FilterViewFor` promotes the asking seat's
whole `CastSurfaceView` — nested lists included — over the top. The
copy is load-bearing: the public half and the seat's own half come out
of ONE `castStampsFor` call and share the pointer and the slice header,
so blanking in place would take the sets off the owner's frame too.
`faces[i]` is narrowed identically, because a face's block is the same
type. In a HAND the parents are dropped wholesale already (#1169), so
nothing nested survives there either.

**For a client.** Nothing to change, and that is the point: every
reader takes a `CardView` out of the viewer's own snapshot, so
`modes[i].legal_targets` was already the viewer's own answer whenever
the viewer was the caster. A reader that finds a targeted-looking
bullet with no `legal_targets` must treat it as a bullet it cannot
open a picker for — never as a free-form prompt over the whole board —
which is what `beginForModes` already does.

## An ability row's charged cost, alongside its printed one (#1190, #1191, 2026-09-22)

One additive field, on each of the two ability rows that carry a mana
component. A client that ignores it behaves exactly as it did.

Since #1184 (S601.2f) an activated ability's mana cost can be modified
by a permanent someone else controls — Boom Scholar's "Exhaust
abilities of other permanents you control cost {2} less to activate"
— and `Game.AbilityManaCostForEffect` is what the engine actually
charges and what `internal/legal` checks affordability against. The
menu row kept stamping only the PRINTED string
(`activated_abilities[i].mana_cost` / `mana_abilities[i].mana_cost`),
so a discounted ability's row and its real price could disagree — the
gap Boom Scholar shipped as two declared caveats while #1184 closed.

- **`activated_abilities[i].charged_mana_cost`** and
  **`mana_abilities[i].charged_mana_cost`** (both `omitempty`) are what
  the engine will actually charge for the ability's mana component
  right now, rendered from the same `ParsedCost` the payment path
  charges (`ParsedCost.String()`, the CR 601.2f pass's own renderer)
  rather than a second copy of the printed string. CR 605.1a makes a
  mana ability an activated ability, so #1191 gave the CR 605 half the
  same pass through `Game.ManaAbilityManaCostForEffect` and the same
  field name — one shape, two owners, exactly as the counter-cost and
  sacrifice-cost fields already are.

  Present whenever the ability's `mana_cost` is, and **equal** to it
  when no cost modifier on the battlefield reaches this ability, which
  is nearly every activation in the game — a pre-#1190 client that
  only ever read `mana_cost` sees no discounted card differently than
  it always has. Absent when the printed cost could not be priced (an
  unparseable string, or a modifier that itself errors), in which case
  a client falls back to `mana_cost` exactly as it did before this
  field existed.

  **A client shows `charged_mana_cost` in the row where it showed
  `mana_cost`, and `mana_cost` as a tooltip only when the two strings
  differ.** There is no cost arithmetic on the client side — the
  server is still the one pricer, in one direction only, the same
  discipline `target_cost_notes` and `phyrexian_symbols` already hold
  to.

  Since #1319 a CR 116.2 special action's cost can be modified the
  same way — Ranar the Ever-Watchful's "The first card you foretell
  each turn costs {0} to foretell" — and `special_actions[i].charged_cost`
  (`omitempty`, present whenever `cost` is) is the same field, one
  more owner: `Game.SpecialActionManaCostForEffect` runs the printed
  `cost` through the identical CR 601.2f pass, keyed off
  `CostModifier.SpecialActions` rather than `.Activations`, so a
  modifier written for one family can never discount the other. A
  client reads `charged_cost` in the row exactly as it reads
  `charged_mana_cost`, `cost` as the tooltip only when they differ.

- **`activated_abilities[i].target_charged_mana_costs`** (#1296,
  `omitempty`, [ADR 0020](decisions/0020-activated-abilities.md)
  amendment 2026-09-24) is `{ "<target id>": "<charged cost>" }` — what
  the mana component costs if the ability targets each of its
  `legal_targets` (a card instance ID or a player ID), valued as
  `charged_mana_cost` is (`""` is free). It exists for an ability whose
  PRICE reads its target: Dragonfire Blade's "Equip {4}. This ability
  costs {1} less to activate for each color of the creature it
  targets", Ghostfire Blade's "{2} less if it targets a colorless
  creature". CR 602.2b runs 601.2c (targets) before 601.2f (the total
  cost), so such an ability has no single price, and
  `charged_mana_cost` on its row is the price with no target chosen —
  the printed cost for a reduction, an honest upper bound for a client
  that ignores the new field. Present only when the price reads the
  target AND the ability is one target clause with one pick and no
  modes (every printed card of the family); each entry is priced by the
  same `Game.AbilityManaCostForTargetsForEffect` the payment charges. A
  target whose price could not be computed is left out. The client
  shows the range in the menu row's hint ("{2}–{4} depending on the
  target") and each price with its targets in the targeting banner,
  because an equip's one click is also its confirm.

- **`activate_ability`'s `strict` / `auto_tap`** (#1296). The server
  has always read both on the catalog path (`ability_index`) and, with
  neither set, taken the sandbox posture: spend what the pool covers
  and, if it does not cover the cost, waive the charge and mark it paid
  on paper. The client stamped its `gameplay.strictMana` setting only
  on `cast_spell`, so every activated ability a player clicked was
  paper-paid whatever the setting said — an equip onto Vivi Ornitier
  with an empty pool was free (the report). A client with strict mana
  on now sends `strict: true, auto_tap: true` on every catalog
  `activate_ability`: the engine taps lands for the shortfall, as it
  does for every bot activation, special action and attack tax, and
  refuses with `insufficient_mana` ("insufficient mana to activate
  that ability", no `card_id` — the cast-only override toast stays
  away) when the board cannot pay. Strict mana off is unchanged.

- **The `{S}` (snow) round trip.** `ParsedCost.String()` is a full
  renderer — `{X}` × `XSlots`, then the generic component, then each
  colored requirement (a hybrid's options joined by `/`, `/P` for
  Phyrexian, the numeric alternative for `{2/W}`) — and needed one
  parser fix to be honest about snow: `{S}` and `{C}` used to parse to
  the identical `ColorRequirement{Options: {"C"}}`, so a cost with both
  ("`{1}{S}{C}`") had two indistinguishable colorless requirements and
  no way to print one back as `{S}`. `ColorRequirement` now carries a
  `Snow` bit per symbol (`ParsedCost.HasSnow` is unchanged — still the
  cost-level "at least one" flag its existing readers use), so
  `ParseCost(s).String() == s` for every shape `ParseCost` accepts,
  snow included.

## `castable_here` answers CR 307.1 too (#1195, 2026-09-22)

No new field, no field removed, `v` unmoved. What changed is the third
input to one bit.

Since #1015 `castable_here` has been "no `cant_cast` and at least one
claimable price", derived in one place. It said nothing about TIMING,
and the announce path always has: a sorcery in a graveyard under a
flashback grant is refused with `ErrSorcerySpeedRequired` in an
opponent's end step, and a bot is offered no move for it. The bit
therefore went out `true` on frames where the button behind it did not
work, and `false` on a board where a Vedalken Orrery had opened the
cast.

It is now the same bit plus `game.CastTimingOpenLocked` — the one
CR 307.1 read `CastSpell` and the bot enumerator already share (#1195,
[ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)'s
2026-09-22 amendment). That predicate folds the card's own timing
(instant, or flash), the granted permission's `Timing` override, the
per-player "you may cast spells as though they had flash" grants
(Vedalken Orrery, Leyline of Anticipation, Emergence Zone) and, last,
the per-player restrictions (Teferi, Time Raveler), because CR 101.2
says "can't" beats "can".

**For a client.** Two things follow, and both are things a client
should have been doing anyway:

- **`castable_here` is a right-NOW bit, not a property of the card.**
  It already moved with the graveyard's size (#1015) and with the
  viewer (#1055); it now also moves with the step and with the board.
  A client that caches it across frames will render a stale button.
  Every snapshot carries the current answer.
- **It is still not a reason.** `cant_cast` is the printed clause that
  BANS a cast ("Each player can't cast more than one spell each
  turn") and is what a toast should quote. "Not at this timing" is an
  ordinary state of nearly every sorcery on nearly every frame, so it
  is a `false` here and never a clause there — the two are separate
  reads for that reason, and folding the second into `cant_cast` would
  have put a refusal clause on half the cards in play.

Unchanged: the bit is never set on a hand or command-zone card, exile
keys its button off `exile_play`, and the per-viewer narrowing of
#1055 stands — each holder's stamp is computed for that holder, so two
seats that disagree about timing are told different things about the
same card, which is what the rules say.


## Casting from exile: `castable_here` and `cast_prices` (#1389, 2026-09-24)

Additive, `v` unmoved. Two answers reach the frame of a seat that holds
a LIVE cast permission over a card in `exile`, for the client's
castable-from-exile strip beside the hand.

- **`castable_here` is now stamped in exile too.** Same bit, same
  meaning as in a graveyard: "YOU may cast this from here right now" —
  no `cant_cast`, a claimable price, and `game.CastTimingOpenLocked`
  (so a plotted card is `false` outside its owner's main phase, and a
  sorcery is `false` in an end step). A LAND under a "you may play it"
  grant (Breeches) is answered by the land rule instead — the owner's
  main phase, an empty stack, a land drop left — and a land under a
  cast-only grant (Ragavan) is never `true` (CR 305.1). The note in
  #1195 that "exile keys its button off `exile_play`" is retired:
  `exile_play` still names the grant's holder, its face and its
  `not_before_turn`, and `castable_here` is the "now" half.
- **`cast_prices`** is a new `CastPriceView[]` on the cast surface
  (card and castable faces), exile only:

  ```
  { alternative_cost?: string, label?: string, cost: string, life?: number, printed?: boolean }
  ```

  One entry per price the cast may claim, **cheapest first**. `cost` is
  the total AFTER every CR 601.2f cost modifier, computed by
  `game.PriceCastForEffect` — the pricer `CastSpell`, the auto-tap
  preview and the bot enumerator share — so it is the number the
  auto-tapper will tap for. It is never empty: a free cast reads
  `"{0}"`. `alternative_cost` is the key the cast sends to claim that
  price (`"foretell"`, a granted `"flashback"`), empty for the path
  that claims none — the printed cost, or a permission's own flat
  price (airbend's `{2}`, a plotted card's `{0}`). `printed` is true
  when the price IS the card's printed mana cost, untouched; the client
  draws no badge for it. Priced with no targets and X = 0, as the
  auto-tap preview's first readout is.

  | Mechanic | `cast_prices[0].cost` | `printed` | when `castable_here` |
  |---|---|---|---|
  | impulse exile | the printed cost | yes | the card's own timing |
  | airbend | `{2}` | only on a card that prints `{2}` | the card's own timing |
  | warp | the printed cost | yes | from `not_before_turn` |
  | plot | `{0}` | no | a later turn, owner's main phase, empty stack |
  | foretell | the foretell cost, `alternative_cost: "foretell"` | no | a later turn |
  | adventure (CR 715.4) | the creature half's printed cost | yes | the creature's timing |
  | prepare (ADR 0090) | the prepare spell's printed cost | yes | the spell's timing |
  | ADR 0066 grants | per the grant | per the grant | per the grant |

  Cascade, discover and hideaway cast during resolution rather than
  under a lingering permission, so they never reach the strip.

**Whose answer it is (#1055).** Both are per viewer and ride the same
per-seat carrier as `legal_targets`: the holder's frame gets them and
nobody else's does, spectators and admin frames included. A cost
modifier can be scoped to one player, so one seat's price is not
another's. Both are cleared on the non-knower redaction, so a
face-down FORETOLD card still reads as a card back with no price to
anybody but its owner — the foretell cost would name the card.

**Before the window opens.** A warp, plot or foretell grant on the turn
it was made is not live, so the holder gets the public `exile_play`
(with `not_before_turn`) and neither `castable_here` nor
`cast_prices`: the engine would not accept the cast at any price. The
client shows such a card dimmed with a "next turn" hint and no badge.

## A land's `castable_here` is the land-play rule in every zone (#1407, 2026-09-24)

A narrowing, no new field, `v` unmoved. #1389 answered a LAND in exile
by the land-play rule; the graveyard and the library top still asked
the spell rules, so a land a play permission opens there (Courser of
Kruphix, Oracle of Mul Daya, a Crucible of Worlds) stayed `true` after
the turn's land drop was spent, and the click came back
`ErrLandDropUnavailable`.

For a land in a graveyard, on a library top or in exile, `castable_here`
is now `true` exactly when `cast_spell`'s land branch would accept the
play: the viewer's own main phase with an empty stack (CR 305.1), a
land drop left (CR 305.2, `game.LandPlayOpenForEffect`), something
opening the zone (a permission, or the card's own text), and no
cast-only grant (Realmwalker, Ragavan — CR 305.1 strands the
land). No timing statement reaches it: a Vedalken Orrery does not open
a land drop and Dosan does not shut one. A spell in the same zone is
unchanged. The client reads the bit and nothing else, so the zone
browser's button and the library-top affordance follow with no client
change.


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
