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
    this. Omitting it for an admin connection yields the admin's
    omniscient debug view; any other seatless connection is a spectator,
    who sees public information only (see below).
- **Admins** ([ADR 0110](decisions/0110-remember-me.md) §3, owner
  answer 4) are the shared token's session and any signed-in session
  whose Discord ID is on `CMDCTRL_DISCORD_ADMIN_USER_IDS` (see
  [lobby.md](lobby.md#who-is-an-admin-adr-0110-3)). An admin may bind
  to any `?game=`, optionally as any `?player=`, and is not read-only.
  A signed-in admin's own seat or spectator session at its own game,
  asking for no other seat, binds exactly as anyone's would, and the
  connection is still marked admin for the table-host gates and the card
  overrides below. Every admin binding is logged at Info with
  `admin_user_id` (or `admin_id` for the token), the game and the seat.

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
- **Spectators see public information only** (#1588, owner decision
  2026-09-30, [ADR 0069's amendment](decisions/0069-face-down-objects.md)).
  A spectator is any seatless connection that is not the admin — a
  `RoleSpectator` session minted via `POST /games/{id}/spectate` (S11).
  A card is known to a spectator only when EVERY seat knows it, so a
  face-down card (foretold, hideaway, Necropotence's exile, Gonti, a
  morph's face), a card revealed to one seat, and every hand card read
  exactly as they do to a seat that is not a knower: `face_down` and
  `face_down_kind` stay public, `face_visible` is false, hand counts are
  kept and hand contents are withheld. Move lists, cast stamps, pending-
  choice options and log entries follow the same rule.
  `RoleSpectator` connections additionally have every inbound `action`
  frame rejected with `bad_request` ("spectator connections are
  read-only").
- **The admin keeps the omniscient debug view** (admin without
  `?player=`): every card is known, hands and libraries are still
  withheld, and the connection keeps its ability to mutate state —
  admins are the moderator escape hatch. Server-side the two are
  different `FilterViewFor` viewers: the empty string is the admin,
  `protocol.SpectatorViewerID` is a spectator.

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
| `payload.code` | string | yes | Short machine-readable code. v0 codes: `bad_version`, `bad_json`, `bad_request`, `internal`. S15 adds `insufficient_mana`, #705 adds `illegal_block`, #1063 adds `attack_tax_unpaid`, and #1507 adds `illegal_attack` (#1571 gives it a second reason) — see below. [ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §6.1 adds `no_decision` and `rate_limited`, answers to a [`legal_moves_request`](#legal_moves_request-client--server--adr-0122-61). |
| `payload.message` | string | yes | Human-readable message safe to display to the user. |
| `payload.missing` | string[] | no | Present only when `code == "insufficient_mana"` (S15). List of mana symbols the caller's pool could not cover, in the order they appear in the printed cost (e.g. `["{R}", "{1}"]`). Client renders these verbatim into the override toast. |
| `payload.card_id` | string (UUID) | no | Present when `code == "insufficient_mana"` (S15): instance ID of the card whose cast was rejected. Lets the client's "Cast anyway" / "Auto-tap & cast" buttons re-fire the same cast without needing to round-trip through the user's last click. Also present when `code == "illegal_block"` (#705): instance ID of the refused blocker; and when `code == "illegal_attack"` (#1507): a creature from the refused declaration. |
| `payload.reason` | string | no | Present when `code == "illegal_block"` (#705): the stable snake_case token naming why the block was refused — see below. Also carries the attack tax's price on `attack_tax_unpaid` (#1063) and `attack_limit` (#1507) or `attack_requirement` (#1571) on `illegal_attack`. |

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
   for this one cast). The cast's log entry carries `unpaid: true`
   and reads "<player> cast <card> without paying its mana cost"
   (ADR 0118 §2; see **Unpaid casts** under LogEvent).
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
| `declaration_limit` | the declaration would put more creatures into blocks in this COMBAT than a whole-combat limit allows — Silent Arbiter's "no more than one creature can block each combat" (CR 509.1b, #1507). Every stored block counts, whichever attacker it is on and whichever defender made it, so in a multiplayer combat the first defender to block can use the limit up. The sentence names the limit and its card: "No more than one creature can block each combat (Silent Arbiter)." A per-player limit counts each defending player's blocks on their own (Mirri, Weatherlight Duelist's "each opponent can't block with more than one creature this combat", #1534) and reads "Each opponent can't block with more than one creature this combat (Mirri, Weatherlight Duelist)." `card_id` is a blocker from the refused declaration |
| `not_defending` | the blocker's controller is not defending against that attacker: it is attacking another player, a planeswalker another player controls or a battle another player protects — or nothing at all (CR 802.4a / 509.1a, #1339). A creature whose planeswalker or battle has left is still defended by the player who was defending it (#1364, CR 506.4c): "Grizzly Bears is still attacking, though what it attacked is gone, so only P3 can block it." Like the count reasons it comes only from the declaration verbs. The sentence names where the attacker is pointed ("Grizzly Bears is attacking P3, so only P3 can block it."). A pairing the declaration already holds is not re-judged, so an attacker reselected after it was blocked (CR 508.7a) keeps its blockers |
| `blocker_capacity` | the declaration would have one creature block more attackers than it can (CR 509.1a/b, #1706): a creature that "can block an additional creature each combat" (High Ground, Two-Headed Giant of Foriys) already blocking two. `n` is its capacity. Only a creature that can block more than one is ever refused this — an ordinary blocker's second entry re-points it. "Two-Headed Giant of Foriys can't block more than two creatures each combat." |
| `block_requirement` | the declaration — or the defending player's `pass_priority`, `finish_blocks` or `advance_step` that ends it — would leave a CR 509.1c blocking requirement unobeyed that some legal declaration could obey (#1597, [ADR 0045](decisions/0045-combat-restrictions.md) Decisions 55-58): "blocks each combat if able" (Grand Melee, Razorgrass Screen), Lure's "all creatures able to block this creature do so", "must be blocked if able", "must be blocked by exactly one creature if able". A requirement never beats a restriction, so a Lure'd menace attacker with one potential blocker asks nothing. `card_id` is the creature that could obey it, and the sentence names the requirement and the card that prints it: "Wall One must block Bear if able (Lure).", "Their Wall must block this combat if able (Grand Melee).", "Gaea's Protector must be blocked if able." Only a defender whose declaration is still pending is judged. The creatures owed are on the snapshot as `CardView.must_block` |

| `blocks_declared` | the blocker's controller has already FINISHED declaring blockers this combat (CR 509.1, #1501, [ADR 0045](decisions/0045-combat-restrictions.md) amendment of 2026-10-03): by `finish_blocks`, by passing, or by having no legal block as the step began. The declaration is one turn-based action, so a block after it — on a ninja that entered attacking after the declaration, say — is not part of it. Returned only once every other check has passed, so a block that is illegal anyway reports what is wrong with it (`flying`, `too_few_blockers`, …) rather than its lateness. A pairing the declaration already holds is not re-judged, so repeating it is still a no-op. `card_id` is the blocker. "You have already finished declaring blockers this combat." to the defender, "P2 has already finished declaring blockers this combat." to anyone else |

The tokens are stable once shipped. The addendum reserves more
(`tapped`, …); each is added to this table in the change that first sends
it. `declaration_limit` joined with #1507, `block_requirement` with #1597,
`blocker_capacity` with #1706, `blocks_declared` with #1501.

The two COUNT reasons are about the whole declaration rather than about one
pair, which is why they can only come from the declaration verbs and never
from a per-pair legality query. `card_id` is a blocker named in the refused
declaration when there is one — a count broken by a blocker being re-pointed
AWAY names none, because no creature in that declaration is the problem.

#### `illegal_attack` (#1507)

Emitted when a `declare_attacker` or `declare_attackers` breaks a
CR 508.1c count limit on the whole declaration — Silent Arbiter's and
Dueling Grounds' "no more than one creature can attack each combat",
Crawlspace's "no more than two creatures can attack you each combat"
([ADR 0045](decisions/0045-combat-restrictions.md) Decisions 44-45).
`reason` is `attack_limit` for a limit, or `attack_requirement` for a
CR 508.1d requirement (#1571, below). For `attack_limit`, `card_id` is a creature
from the refused declaration that the limit counts. `message` is a sentence
the server builds and addresses to the caller: the seat a Crawlspace
protects reads "No more than two creatures can attack you each combat
(Crawlspace).", and anyone else reads that seat's name. A limit on one
planeswalker reads the same for everyone: "No more than one creature can
attack The Eternal Wanderer each combat." (#1534). A limit with a condition
(Mirri, Weatherlight Duelist's "as long as Mirri is tapped") is refused the
same way while the condition holds, and not at all otherwise.

Nothing was declared. No creature is tapped, no attack is announced and no
attack tax is charged, because the limit is judged first. A `declare_attackers`
batch is refused WHOLE. Every creature in an over-full swing is eligible on
its own, so there is no entry to skip, and which creatures stay home is the
attacking player's choice: they send a smaller set. "Attack you" means the
player, so an attack on a planeswalker they control does not count. A
declaration that does not raise a limit's count is never refused by it, so a
limit that arrives after the attackers were declared unmakes nothing.

How many creatures the smaller set may hold is on the snapshot before the
click: `turn.attack_targets[].attack_limit` (#1533, see `TurnView` below) is
the room each target has left. The client's "attack with all" answers this
refusal by offering its attackers picker capped at that number.

**`attack_requirement`** (#1571, [ADR 0045](decisions/0045-combat-restrictions.md)
Decisions 48-52). An attack declaration must obey the most CR 508.1d
requirements it can without breaking a restriction or paying a cost:
"attacks each combat if able", "creatures your opponents control attack this
turn if able" (Bident of Thassa), and goad's "attacks each combat if able and
attacks a player other than the goader if able" (CR 701.15b). Two things are
refused with this reason:

- a `declare_attacker` / `declare_attackers` that makes a requirement that
  could still be obeyed unobeyable — a goaded creature at its goader while
  another opponent is open, a creature with no requirement in the one slot
  Silent Arbiter leaves while a creature that must attack waits. A batch is
  refused whole; nothing is declared, tapped or paid;
- the ACTIVE player's `pass_priority` (and `advance_step`) in
  `declare_attackers` while a free attack would obey a requirement the
  declaration does not. That pass is where the declaration ends, and it is
  judged once per combat. Declaring nothing is refused only when a
  requirement can actually be met: a creature that is tapped, summoning sick,
  under "can't attack", or could attack only by paying an attack tax owes
  nothing.

`card_id` is the creature carrying the requirement, and `message` names it:
"Zurgo Helmsmasher must attack this combat if able.", "Late Bear must attack
this combat if able (Bident of Thassa).", "Grizzly Bears is goaded by you and
must attack a player other than you if able." The creatures owed are on the
snapshot before the click as `CardView.must_attack`.

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
to the originating client only (no broadcast). Since [ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md)
§6.4, on success the originating client also gets an [`ack`](#ack-server--client--adr-0122-64)
carrying the action's `id`, after the snapshot of the state the action
produced.

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
| `move_card` | no | `{ "src": <ZoneRef>, "dst": <ZoneRef>, "instance_id": "<uuid>", "as_commander"?: bool, "to_bottom"?: bool }` | General-purpose zone-to-zone move. Clears tapped state, counters, marked damage and the deathtouch mark if leaving the battlefield (CR 400.7). Caller must control the card. S13.1, amended by [ADR 0115](decisions/0115-commanders-die.md): when the card is a commander AND `dst` is a hand or a library, its owner is asked whether to put it in the command zone instead (the CR 903.9b replacement) — nothing moves until that prompt is answered, and #707 made the move complete afterwards from whatever zone the card was in. When `dst` is a graveyard or exile the card moves at once (a commander moved off the battlefield to its graveyard dies), and the CR 903.9a state-based action asks its owner afterwards with a `commander_return` prompt. `as_commander` pre-answers that question "yes" for a graveyard or exile destination, so the commander lands and then goes to the command zone in one action (the context menu's "Graveyard → command zone"); for a hand or library destination it is a routing flavour on the request, not a gate on the offer. #170: when `to_bottom` is true the card is seated at the BOTTOM of the destination library instead of the top — a no-op if a replacement rewrote the destination, or if the destination is not a library. |
| `tap` | no | `{ "instance_id": "<uuid>" }` | Sets a battlefield card's tapped state to true. Caller must control the card (admin sessions bypass) — see "controller-only card actions" below. |
| `untap` | no | `{ "instance_id": "<uuid>" }` | Sets a battlefield card's tapped state to false. Same controller gate as `tap`. |
| `untap_all` | yes | — | Untaps all of `player`'s cards on the battlefield. S13: server-side no-op when called during the active player's `untap` step (the step-entry hook has already untapped automatically). Outside that window the action behaves as before — sandbox / replay support. |
| `pass_priority` | no | — | Rotates priority to the next non-eliminated seat in turn order (CR 117.3d). Since #2275 the server counts the seats that have passed **in succession** (CR 117.4): once every seat still in the game has passed with nobody acting in between, the top of the stack resolves and `priority_holder` returns to the active seat (CR 117.3b), or, with an empty stack, the step ends and `priority_holder` resets per the new step's rules. Coming back round to the active seat is not that test — it was, before #2275, which let a non-active player's spell resolve on its caster's pass without the active player ever holding priority over it. A cast, an activation (a mana ability too, when its activator holds priority), a land play or other special action, anything put on the stack, and a player leaving the game all restart the succession, so the seats that passed before it pass again. The succession is server state only; nothing on the wire changed. S13: skips eliminated seats during the rotation and never waits for them; rejects with `bad_request` ("no player holds priority this step") on `untap` / `cleanup` (those don't grant priority per CR 502.4 / 514.3). Since #1571 the ACTIVE player's pass in `declare_attackers` is the attack declaration's CR 508.1d checkpoint and is refused with `illegal_attack` / `attack_requirement` while a creature marked `must_attack` stays home; the same holds for `advance_step` out of that step. Since #1597 a DEFENDING player's pass in `declare_blockers` is likewise their block declaration's CR 509.1c checkpoint, refused with `illegal_block` / `block_requirement` while a creature marked `must_block` stays home; `advance_step` out of that step is refused the same way for every defender still declaring. Since #1501 nobody holds priority in `declare_blockers` while any defender is still declaring (`priority_holder = -1`, CR 509.1), so every seated player's `pass_priority` there is refused with `bad_request` ("you do not hold priority") — and so are `cast_spell`, `activate_ability`, `activate_loyalty`, `counter_spell`, `counter_ability` and `special_action`, the other priority-gated verbs — until the last defender sends `finish_blocks`. |
| `pass_turn` | no | — | Ends the current turn at once and jumps to the next seat's turn. Attackers and blockers leave combat and the cleanup sweep runs (marked damage removed, "until end of turn" effects end); the steps in between are skipped, so there is no end step and no discard to hand size. Lands on Untap with `priority_holder = -1`; the step-entry hook auto-untaps and walks the cursor on to Upkeep before the next snapshot is broadcast. |
| `mulligan` | yes | `{ "hand_size": <int> }` | Shuffles `player`'s hand back into their library and draws `hand_size` cards. Simplified London mulligan — no card-to-bottom penalty. Refused out of turn like `keep_hand` (CR 103.5, #2237): the starting seat decides first and the others follow in turn order; a seat that mulligans decides again only in the next round, which is the seats that mulliganed, in turn order again. |
| `shuffle_library` | yes | — | Reshuffles `player`'s library. |
| `change_life` | yes | `{ "delta": <int> }` | Adjusts `player`'s life by `delta` (positive for gain, negative for loss). |
| `add_counter` | no | `{ "instance_id": "<uuid>", "name": "<string>", "delta": <int> }` | Modifies a named counter on a card. Delta ≤ 0 that drives the counter to zero removes the entry. Caller must control the card. |
| `set_commander_damage` | no | `{ "from": "<uuid>", "to": "<uuid>", "amount": <int> }` | Sets total commander damage dealt from `from`'s commander(s) to `to`. Set semantics, not additive. |
| `set_battlefield_position` | no | `{ "instance_id": "<uuid>", "x": <float>, "y": <float> }` | Stamps a normalised (x, y) position in `[0, 1]` on a battlefield card. Server clamps out-of-range inputs rather than erroring. Target must be on the battlefield — other zones return `card_not_found`. Caller must control the card. Added in S06; controller gate added in S08.5. |
| `concede` | yes | — | Marks the calling player as eliminated and advances the turn cursor past them if they were the active seat. When exactly one non-eliminated seat remains, the game's `state` transitions to `ended` and `GameView.outcome` names the survivor the winner (`cause: "last_standing"`); when none remains it is a draw. Never stopped by a "can't lose the game" effect (CR 104.3a). One concession writes one `eliminated` log line (`cause: "concede"`). Idempotent calls return `bad_request` with "player is already eliminated". Added in S08; outcome since ADR 0057 (#749). **During the opening roll** ([ADR 0121](decisions/0121-animated-dice.md) §1) a concession strikes the seat from every round and the standings are read again: a single leader among the seats left chooses, several roll again; it mints no undo entry. |
| `keep_hand` | yes | — | Commits the calling player to their current opening hand during the mulligan window. Once every seated, non-eliminated player has called `keep_hand`, `mulligans_open` flips to false. Idempotent (no-op if already kept). Added in S08. Since [#2237](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2237) (CR 103.5) it is refused (`bad_request`, "it is not your turn to decide on your opening hand") from any seat but the one whose turn it is: the one with `mulligan_turn` set. |
| `roll_opening` | yes | — | Rolls the calling player's d20 in the current round of the opening roll ([ADR 0121](decisions/0121-animated-dice.md) §3). Refused with `bad_request` ("you have no die to roll in this round of the opening roll") unless the seat is in the round's `seats` and has no die in its `rolls`. The die is drawn when the action is applied, on the seat's own keyed stream, so a seat's k-th opening die does not depend on when it rolls or who rolled first. Each die is a `roll` log line; a round every seat has rolled in ends with an `opening_roll` line: one leader becomes the chooser, several roll again in a new round. Mints no undo entry. Added in S61. |
| `host_roll_remaining` | no | — | Rolls, in seat order, for every seat in the **current** round of the opening roll that has not rolled, recording the presser as each die's `by` ([ADR 0121](decisions/0121-animated-dice.md) §3, owner decision 3: "Roll for everyone left"). **Host or admin only**, gated at the WebSocket edge like `set_table_settings` (`ws.Room.CanManageTable`); anyone else gets `bad_request` ("only the table host or the server admin may roll for everyone left"). A tie it causes opens a new round with its own dice to roll. Refused when every seat in the round has rolled. Writes an `opening_roll` line (`label: "rolled_for"`) naming the seats rolled for, the presser's own excluded. Mints no undo entry. Added in S61. |
| `choose_starting_player` | yes | `{ "seat": <int> }` | The winner of the opening roll chooses who takes the first turn (CR 103.1, [ADR 0121](decisions/0121-animated-dice.md) §2): `player` must be `opening_roll.chooser` (or the admin acting for that seat); `seat` is any seat still in the game, the chooser's own included. Then every library is shuffled and every opening hand dealt (CR 103.3, 103.5, 903.7), the starting seat takes turn 1, `opening_roll` disappears and the mulligan proceeds as before. Writes a `starting_player` line. **Final**: it mints no undo entry. Added in S61. |
| `roll_table_die` | yes | `{ "die": "d6" \| "d20" \| "coin" }` | "Roll a die" ([ADR 0121](decisions/0121-animated-dice.md) §5): rolls a d6 or a d20, or flips a coin, at the table, for the calling player. **Not a game roll**: no effect instructed it (CR 706.1), so it is never a `roll` or `flip` and no "whenever you roll" or "whenever you win a coin flip" ability sees it; it changes nothing but the log, where it writes one `table_roll` line. Legal at any time while the game is active, the opening roll and the mulligan included; refused for an eliminated seat, and with `bad_request` ("a table roll is a d6, a d20 or a coin") for any other `die`. Drawn from its own keyed stream outside the turn counters, so it moves no card's draw, and never replayed by an undo. **One per seat per 2 s**: the next inside that is refused with `bad_request` ("wait for your last roll to land"), so a held-down button cannot push the game's history out of the 200-entry log. Mints no undo entry, and an undo of another action does not erase its line (see `table_roll` below). Bots are never offered it. Added in S61. |
| `set_trigger_order_preference` | yes | `{ "always_ask": boolean }` (required) | "Always ask me to order my triggers" (#1530, [ADR 0018](decisions/0018-triggers-on-the-stack.md) 2026-10-05 amendment). Sets the calling seat's own preference (default off); the admin may set any seat's. When on, a batch of two or more of that seat's triggers raises the CR 603.3b `trigger_order` prompt even when every item commutes (an all-prowess batch, #1511) or is identical. Only that seat is affected. A setting, not a play: mints no undo entry, an undo of another action does not revert it, it is legal during the opening roll, and it is refused with `bad_request` once the game is not active. A missing `always_ask` is `bad_request`. The value rides the engine snapshot, so it survives a restart. The client re-sends it whenever its own seat's `trigger_order_always_ask` disagrees with the local setting. Bots are never offered it. |
| `set_auto_answers` | yes | `{ "rules": [{ "key": string, "answer": "always" \| "never" }] }` (required) | A seat's standing answers to repeated prompts ([ADR 0127](decisions/0127-answering-repeated-prompts-for-you.md) §3, #1961). The list **replaces** the calling seat's rules (an empty list clears them); the admin may set any seat's. Each `key` is a prompt's `auto_answer_key`, and a key with no rule is Ask. Refused with `bad_request` above 100 rules, for a key over 256 bytes, an empty key, a key named twice, an answer that is not `always` or `never`, a bot seat, a missing `rules`, or once the game is not active. A setting, not a play: mints no undo entry, an undo of another action does not revert it, and it is legal during the opening roll. It rides the engine snapshot. The client sends it whenever its own seat's `auto_answers` disagrees with the synced setting `gameplay.autoAnswers`. With a rule in place the server answers that prompt for the seat as its own commit, stamped with the chooser and free to undo, and writes an `auto_answer` log line; see **Automatic answers** below. Bots are never offered it, and the MCP seat does not expose it. |
| `declare_attacker` | no | `{ "attacker": "<uuid>", "target": "<uuid>" }` | Marks a battlefield card as attacking the target player. Step-gated to `declare_attackers`; rejects with `bad_request` outside that step. Attacker must be a creature (`type_line` contains "Creature"); rejects with `bad_request` otherwise. Re-declaring the same attacker against a different target overwrites. Caller must control the attacker (admin sessions bypass) — see "controller-only card actions" below. Added in S08; controller gate added in S08.5. Since [ADR 0080](decisions/0080-attack-taxes.md) it also carries the CR 508.1a payment trio — see the attack-tax paragraph below the table. Since #1507 it is refused with `illegal_attack` / `attack_limit` when one more attacker would break a CR 508.1c count limit (Silent Arbiter, Crawlspace). Since #1571 it is refused with `illegal_attack` / `attack_requirement` when it would make a CR 508.1d requirement unobeyable (a goaded creature at its goader while another opponent is open). Since #2462 `legal_moves` offers it only while the active player holds priority: their `pass_priority` ends the declaration, so after it no attack is offered (the verb itself still checks only the step). Since [ADR 0130](decisions/0130-exert.md) it takes an optional `"exert": true`, the choice to exert the attacker as it attacks (CR 701.43d, 508.1g): it is staged with the attack and paid as the declaration locks in, in the same event batch as the attacks, so the creature won't untap during the attacking player's next untap step. On a creature that can't be exerted as it attacks right now (no such ability, its condition fails, or it was already declared this combat) the action is refused with `bad_request` ("this creature can't be exerted as it attacks"). A creature re-pointed before the lock-in keeps a choice it already staged. Absent or `false` means no. |
| `declare_attackers` | no | `{ "attackers": [{ "attacker": "<uuid>", "target": "<uuid>" }, ...] }` | Declares a whole attacking set in ONE action — the wire verb behind the client's "attack with all" cluster (#318). Step-gated to `declare_attackers`. Each entry is checked independently and **skipped silently** when the creature is unknown, is not a creature, is tapped, is summoning-sick (CR 302.6), has `defender` (CR 702.3), is already declared this combat, or names a seat that is the creature's own controller / eliminated / nonexistent — one ineligible creature must never sink a wide declaration. Rejects with `bad_request` when the batch is empty, exceeds 256 entries, contains a malformed UUID, or when EVERY entry was skipped. Caller must control every listed creature; a foreign creature rejects the whole batch (authorization is not part of the skip contract). Declared creatures tap unless they have `vigilance` (CR 508.1f / 702.20), emit one `EventAttack` each, and drain through a single state-check pass so simultaneous attack triggers reach the stack together (CR 508.1 / 508.2). Prefer this over looping `declare_attacker`: each action is one undo entry and one broadcast snapshot, so N creatures would otherwise cost N of both against a per-turn undo budget of 1. Added in S31. Since [ADR 0080](decisions/0080-attack-taxes.md) there is ONE exception to the silent-skip contract — the CR 508.1a attack tax is all or nothing; see below. #1507 adds the second: a CR 508.1c count limit (Silent Arbiter, Crawlspace) refuses the whole eligible set with `illegal_attack` rather than choosing which creatures to drop. #1571 adds the third, for the same reason: a set that makes a CR 508.1d requirement unobeyable is refused whole with `illegal_attack` / `attack_requirement`. [ADR 0130](decisions/0130-exert.md) adds `"exert": true` on any `attackers[]` entry, as on `declare_attacker`; an entry the verb would declare whose creature can't be exerted as it attacks refuses the whole set with `bad_request` (an entry skipped as ineligible drops its choice with it).
| `declare_blocker` | no | `{ "blocker": "<uuid>", "attacker": "<uuid>" }` | Marks a battlefield card as blocking the named attacker. Step-gated to `declare_blockers`. Blocker must be a creature; attacker must exist on the battlefield. Caller must control the blocker. Exactly a one-entry `declare_blockers`, so since #750 it can also be refused for a block COUNT: one creature is not a legal block on a menace attacker and comes back as `illegal_block` / `too_few_blockers` with nothing stored. Send `declare_blockers` for a block that needs several creatures. Added in S08; controller gate added in S08.5; count refusal in S37. |
| `declare_blockers` | no | `{ "blocks": [{ "blocker": "<uuid>", "attacker": "<uuid>" }, ...] }` | Declares a whole block in ONE action (#750, [ADR 0045](decisions/0045-combat-restrictions.md) addendum Decision 13). Step-gated to `declare_blockers`. Exists for a RULES reason, not as a batching convenience: a block COUNT (menace's minimum of two, Hungering Hydra's maximum of one) is a property of the whole declaration, so a two-creature menace block is legal only as a pair and **cannot** be sent as two `declare_blocker` actions — the first would be refused with `too_few_blockers`. **All or nothing**, unlike `declare_attackers`: the set is validated as it will be after the action (each entry's pair legality, then the count bounds for every attacker whose blocker set the action changes, including one that LOSES a re-pointed blocker), and if anything is refused nothing is stored, nothing is announced and the first refusal comes back as `illegal_block`. Rejects with `bad_request` when the batch is empty, exceeds 256 entries, or contains a malformed UUID. Caller must control every listed blocker; a foreign creature rejects the whole batch. An entry naming a creature that is already blocking RE-POINTS it — unless the creature can block more than one attacker (#1706; `block_capacity` / `blocks_any_number` on its card), when it ADDS the attacker while there is room and is refused with `blocker_capacity` once there is none. Like `declare_blocker` the verb only STAGES the pairings — nothing is announced until the declaration is locked in (#830), which since #1279 is when the blocking seat's declaration COMPLETES (see `finish_blocks`). Added in S37. |
| `finish_blocks` | yes | — | Completes `player`'s CR 509.1 block declaration (#1279, [ADR 0045](decisions/0045-combat-restrictions.md) Decision 38) — the "Done blocking" / "No blocks" button. Whatever the seat has staged with `declare_blocker(s)` is its declaration; nothing staged is "declared, none". Step-gated to `declare_blockers`; **not** priority-gated, because nobody holds priority while a defender declares: since #1501 the step parks it (`priority_holder = -1`) from the moment it begins until the last defender still declaring finishes (CR 509.1 — the declaration comes before priority). This verb is therefore how a defender says "done", and it is offered in `legal_moves` as kind `finish_blocks` (AlwaysLegal) to every seat still declaring. Completing announces the seat's staged blocks (the `block` / `becomes_blocked` events and their triggers) and then a `blockers_declared` event, and when it was the last defender still declaring the ACTIVE player receives priority (CR 509.2 / 117.3a) — the post-block window. A defender with no legal block completes "declared, none" as the step begins and is never waited on; a defender who leaves the game while declaring is not waited on either. `pass_priority` by a defender who is still declaring completes their declaration the same way, for the one case in which a declaring defender holds priority (a table restored from before #1501 mid-step, or a player who became a defending player after the step began). Idempotent once declared. Rejects with `bad_request` outside the step, for a seat nothing is attacking ("you are not a defending player this combat"), and — player-scoped — for another seat's `player`. Since #1501 a `declare_blocker(s)` after completion is refused with `illegal_block` / `blocks_declared` (it was a sandbox allowance under #1279). Since #1597 `finish_blocks` is the declaration's CR 509.1c checkpoint, like the defender's `pass_priority` and `advance_step` out of the step: refused with `illegal_block` / `block_requirement` while a creature marked `must_block` stays home. Added in S37. |
| `clear_combat` | no | — | Resets `attacking_target` and `blocking_target` on every battlefield card. Not step-gated (escape hatch). Added in S08. |
| `advance_step` | no | — | **Passes priority until the current step ends** (#914, CR 117.4). Sandbox affordance — does NOT require the caller to hold priority (cf. `pass_priority` which does). Used by the "next step" / "done" buttons so the active player can drive progression without waiting for opponent priority passes. With an EMPTY stack this is one move of the turn cursor and no pass, as it has always been. With anything on the stack it is the passes the rules require first: a step cannot end with objects on the stack, so the server resolves them (each resolution followed by state-based actions, the trigger drain and another priority round) and only then moves the cursor. The drive stops with the cursor unmoved and **no error** — the caller reads the broadcast state — when one of those resolutions raises a blocking prompt, when the CR 732 loop notice goes up, or when the game ends; a prompt that was already open when the action arrived is still refused with `ErrChoicePending` (see the choice gate below). Auto-resolves combat damage on entry to a combat damage step (same hook as `pass_priority`): first-strike damage on entry to `first_strike_damage`, regular damage on entry to `combat_damage`. A step this turn does not have is walked through, not entered — so an advance out of `declare_blockers` lands on `combat_damage` in a combat with no first or double strike in it, and on `first_strike_damage` in one that has (#717). Added in S08; the priority drive in S37. |
| `set_monarch` | optional | — | Designates `player` as the monarch (Conspiracy mechanic). Empty / omitted `player` clears the marker. Any seated player may flip the marker. Added in S10. Since #375 the engine runs the CR 725.2 end-step draw and combat-damage transfer itself; since #1722 ([ADR 0096](decisions/0096-the-monarch-from-a-card-effect.md)) a manual set is the same write a card's "you become the monarch" makes, so it fires "whenever you become the monarch" triggers and re-reads "as long as you're the monarch" statics, does nothing when `player` already holds the crown, and is refused (`player not found`) for a seat that has left the game. |
| `set_initiative` | optional | — | Designates `player` as holding the initiative (BG3 mechanic). Empty / omitted `player` clears. Same posture as `set_monarch`. Added in S10. |
| `set_goaded` | no | `{ "instance_id": "<uuid>", "by": "<uuid>" }` | Marks a battlefield creature as goaded by the named player. Empty `by` clears the goad. Added in S10. Since #1571 the goad is enforced: the creature attacks each combat if able and attacks a player other than the goader if able (CR 701.15b), refused as `illegal_attack` / `attack_requirement`. Since #1598 goads stack, one per player (CR 701.15c): a second player's goad is added beside the first, a player who has already goaded the creature refreshes their goad, and each ends as its goader's next turn begins (CR 701.15a). Empty `by` clears every goad on the creature. |
| `set_poison` | yes | `{ "amount": <int> }` | Sets `player`'s poison counter total. Set semantics (not increment) so two stale tabs don't double-count. Clamped at 0 from below. Added in S10. |
| `set_energy` | yes | `{ "amount": <int> }` | Sets `player`'s energy counter total. Same shape and clamp posture as `set_poison`. Added in S10. |
| `set_promise` | no | `{ "from": "<uuid>", "to": "<uuid>", "count": <int> }` | Sets the per-pair "I owe you" promise-token count from `from` → `to`. Set semantics; clamped at 0. Politics scaffold — visual reminder only. Added in S10. |
| `start_vote` | yes | `{ "topic": "<string>", "options": ["<string>", …] }` | Opens a vote with `player` as initiator. At least 2 options required. Rejects with `bad_request` if a vote is already open (clients must `end_vote` first). Added in S10. |
| `cast_vote` | yes | `{ "option": <int> }` | Records / overwrites `player`'s ballot at the given option index. Re-voting overwrites. Added in S10. |
| `end_vote` | no | — | Closes the currently open vote. Any seated caller may end any vote (sandbox). Added in S10. |
| `undo` | no | — | Pops the room's most recent pre-mutation snapshot off its undo stack and restores the game state. Bumps `seq` like a normal action so clients see a regular `snapshot` frame. Authorization: a seated caller may only pop an entry whose stored caller matches their own seat (rewinding an opponent's move requires their cooperation — they undo first). Each successful seated undo debits the caller's `Player.UndosRemaining` (refreshed to `Game.Settings.UndoLimit` on entering their untap step; never debited when the limit is `-1`, unlimited — [ADR 0075](decisions/0075-table-settings-and-host-controls.md)). Admin (`?player=` omitted) bypasses both the caller and budget gates. Stack capped at 32 per room. No redo at v1. Errors: `bad_request` for empty stack ("nothing to undo"), cross-player request ("you can only undo your own most recent action"), or exhausted budget ("no undos remaining this turn"). Added in S11. |
| `set_table_settings` | no | a settings patch: `{ "undo_limit"?: <int>, "undo_scope"?: "own" \| "host_any", "starting_life"?: <int>, "commander_damage"?: <int>, "bot_pace"?: "fast" \| "normal" \| "slow", "allow_spawn"?: bool }` | Changes the table's settings ([ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.3). The params object **is** the patch: only the fields present are applied, and a field left out is untouched. **Host or admin only** — the server compares the connection against the table's effective host (`is_host` on the seat, `ws.Room.CanManageTable`) and refuses anyone else with `bad_request` ("only the table host or the server admin may change table settings"). Valid in the lobby and mid-game alike. The whole patch is validated before any of it is applied, so one bad field changes nothing: ranges are `undo_limit` ≥ `-1` (`-1` = unlimited, `0` = no undos), `starting_life` 1..999, `commander_damage` 1..99, and the two enums as listed. **`starting_life` after `start` is rejected** (`bad_request`, "starting life cannot change once the game has started") — the value has already been applied to every seat; use `change_life` to adjust totals. An `undo_limit` change refreshes every seat's `undos_remaining` to the new value immediately; a `commander_damage` change is read at the next state-based action check, so lowering it can lose somebody the game at that check. **Not undoable**: the change pushes no entry onto the room's undo stack, and a later `undo` carries the settings forward rather than rolling them back — otherwise lowering the undo limit could be taken back with the undo it was meant to stop. Emits one `settings_changed` event per field that actually changed, and one `settings` log line per event. Deliberately absent from `legal_moves`: a bot never hosts and never changes a table's rules. The HTTP twin is `PATCH /games/{id}/settings` ([docs/lobby.md](lobby.md)). Added in S35. |
| `set_undo_limit` | no | `{ "limit": <int> }` | **DEPRECATED since S35** — a one-field alias for `set_table_settings` (`{"undo_limit": n}`), kept for old clients and `gamecli` scripts. Sets the table's undo limit (`Game.Settings.UndoLimit`, the per-player per-turn undo budget) and immediately refreshes every seat's `UndosRemaining` to the new value. Default is 1. Active games only. **Host or admin only** since [ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.3 — it was any seated player until S35, which is the behaviour change that ADR exists for. Shares the alias's non-undoable apply path. Negative values clamp to 0 (disables undo): the limit `-1` means **unlimited**, and this legacy verb deliberately cannot select it — unlimited is set through `set_table_settings`. A limit of 0 is kept as written; the game no longer rewrites 0 to the default at start. Emits one `settings_changed` event when the value changes. Added in S11. |
| `cast_spell` | yes | `{ "instance_id": "<uuid>", "from_zone"?: "hand" \| "command", "targets"?: [...], "modes"?: [int...], "x_value"?: int, "distribution"?: { "<uuid>": int }, "hold_priority"?: bool, "split_second"?: bool, "strict"?: bool, "force_cast"?: bool, "auto_tap"?: bool, "locked_sources"?: ["<uuid>", ...] }` | S13.1 canonical "play a card from hand" verb (CR 601). Lands route to battlefield (CR 305 special action); other types go to the stack with a fresh metadata entry, the caster retains priority (CR 117.3c). Sorcery-speed gate (sorceries + lands) requires main phase + stack empty + caller is active. Casting from `command` increments the per-commander tax counter (CR 903.8). Caller must hold priority. Targets are `{ "kind": "player" \| "card" \| "self" \| "none", "id"?: "<uuid>" }`. S14: if the card is in the effect catalog with a declared `CardView.target_mode`, the client enters a targeting prompt before sending `cast_spell`; the server re-validates at resolve per CR 608.2b ("countered by game rules" fizzle if every targeted slot is illegal at resolution time). **S15:** `strict` engages the mana-cost gate — server parses `ManaCost` (plus commander tax), rejects with `insufficient_mana` error frame when the pool falls short, deducts on success. Sourced from the client's `gameplay.strictMana` setting, which is on by default since ADR 0118 (settings v19); with it on, a clicked cast is sent with `strict` and `auto_tap` both set, as a dragged one already was. The server's default for a payload with neither flag stays permissive. `force_cast` overrides the gate for one cast (the dock's "Cast anyway" and the "Cast anyway (don't pay)" row) — proceeds permissively without touching the pool. `auto_tap` runs the auto-tapper before the gate, taps the planned permanents, and drops produced mana into the pool — all atomic under one write lock. `locked_sources` excludes specific permanents from the auto-tapper's plan (the modal's lock-tap UI). Plan failure under `auto_tap: true` returns `insufficient_mana` before any cards tap (all-or-nothing). **S21 / #747:** `discard_ids`? and `sacrifice_ids`? pay a catalog card's additional cost (CR 601.2f): `sacrifice_ids` names exactly `additional_cost.sacrifice_options.max` distinct permanents; see **Sacrifice costs of N permanents** below. ADR 0100 §3: a VARIABLE clause on a cast takes the whole list — "sacrifice any number of …" (`min` 0 / `max` 0) any count including none, "sacrifice X …" (`count_from_x`) exactly `x_value` permanents; see **Variable sacrifice costs on a cast** below. **#1727:** a sacrifice that pays an ALTERNATIVE cost (Dread Return's flashback, Fireblast, Demon of Death's Gate) is named in `alt_cost_ids`, never `sacrifice_ids`, and the offer carries `sacrifice_options` instead of `pay_options`; see **A sacrifice as the price** below. **#500:** a land play past the seat's per-turn allowance (CR 305.2) is REJECTED with `bad_request` ("you've already played all the lands you can this turn"). The allowance is not always one — read `PlayerView.land_drops_per_turn` and `PlayerView.lands_played_this_turn` and disable the play before the click rather than only explaining the refusal. **#764:** a target entry may carry `slot` (which target CLAUSE it answers) and `mode` (which entry of `modes` — the mode OCCURRENCE — whose clause list that is). Both default to 0, which is what every single-clause non-modal cast has always meant, and a client that omits them keeps working: the server fills them in by walking the clauses in printed order. Send them when the card has more than one clause (`CardView.clauses`) or when more than one chosen mode targets, because that is the only way to say which pick answers which question. Modes may now REPEAT when `CardView.modes.repeatable` (CR 700.2d, Mystic Confluence) and every chosen bullet may target (CR 700.2c); the order of `modes` is the order they resolve in and the order their targets are asked for, so it is **not** sorted. **S45 / #1330 (CR 702.172a, Spree):** a mode option may carry its own additional cost — `CardView.modes.options[i].cost`, brace notation, present only on a bullet that prints one ("+ {1}{U} — Counter target spell."). Choosing that bullet adds its cost to the total the cast owes (CR 601.2f), on top of the printed cost and every OTHER chosen bullet's — there is no separate announcement for it, the choice is already `modes`. `activated_abilities[i].modes.options[i].cost` mirrors it for the shape's sake, but no printed activated ability charges more for choosing a mode, so it is always absent there today. **S45 / #1590 (CR 601.2b, a conditional mode count):** `CardView.modes.max` (and `activated_abilities[i].modes.max`) is the bound THIS caster is held to right now, not the printed one: Jeska's Will's "Choose one. If you control a commander as you cast this spell, you may choose both instead" stamps `max: 2` for a seat that controls a commander permanent and `max: 1` for everyone else, and the announce gate enforces the same number, read once at announce — a commander that leaves after `cast_spell` does not take a chosen bullet back. `min` does not move. A bystander's public copy of the card shows the printed `max`. No new field. **S45 / #1655 (CR 601.2b / 603.3c):** `min` may now move too. A forced count with no "may" ("If it was kicked, choose both instead", Prophetic Titan's delirium) raises `min` with `max`. A count that reads the caster's own optional costs (Inscription of Ruin's "If this spell was kicked, choose any number instead") stamps `modes.min` / `modes.max` for the cast with no optional cost announced, plus `modes.if_optional_paid: { "min": int, "max": int }` for the cast with the card's optional costs announced. It is absent when paying them changes nothing. The gate reads the `optional_costs` of the same `cast_spell`, because CR 601.2b announces both together. A bystander's copy shows the printed bounds and no `if_optional_paid`. A `mode_pick` prompt's `mode_min` / `mode_max` carry a trigger's forced count the same way; for a "when you cast this spell" trigger they read the spell's own announced kicker. **#787 / #916 (CR 107.4f, CR 601.2b):** `phyrexian_life`?: int is how many of the cost's Phyrexian mana symbols — `{U/P}`, and CR 107.4's ten hybrid Phyrexian symbols `{W/U/P}` … `{G/U/P}` — are paid with **2 life each** instead of mana. It is announced with the cast because CR 601.2b makes "how do you intend to pay each hybrid and Phyrexian symbol" part of announcing the spell; absent (0) pays every symbol with its coloured half. `CardView.phyrexian_symbols` is the ceiling (how many the printed cost prints) and `alternative_costs[i].phyrexian_symbols` the ceiling when that offer is claimed instead, so a client sizes its stepper without parsing a mana string. The engine strikes the symbols a life payment can actually save first, refuses a count larger than the cost prints and a payment CR 119.4 forbids (only down to 0) with `bad_request` before anything is paid, pays the life through the one cost-shaped life path before spending the pool, and plans the auto-tap for the mana half only. `activate_ability` carries the identical field for an activated ability's mana component (#917). **ADR 0103 (CR 709.3, 702.102, 702.127):** every split card casts either half — `face: 1` is the right half — and a Room is a split card too. An aftermath half is refused from any zone but a graveyard, where aftermath itself opens it (`from_zone: "graveyard"`, `face: 1`), and it is exiled as it leaves the stack. `fuse`?: bool casts BOTH halves of a split card with fuse, from hand, with `face` 0 and no `alternative_cost`: the spell pays both halves' costs, its `targets` answer the left half's clauses then the right half's (`slot` numbering runs on across both), and it resolves left then right. `CardView.fused` is the announce surface of that cast. Refused for a card without fuse, from any other zone, and for a card whose halves declare modes or additional costs. |
| `counter_spell` | no | `{ "instance_id": "<uuid>", "to_zone"?: <ZoneRef> }` | Removes a spell item from the stack and routes its card to `to_zone` (default = owner's graveyard). Covers Counterspell (default), Hinder (`to_zone = library`), Remand (`to_zone = hand`), exile-bound counters. Battlefield + stack rejected as destinations (`bad_request: invalid stack-counter destination`). Caller must hold priority. Added in S13.1. |
| `counter_ability` | no | `{ "instance_id": "<uuid>" }` | Removes an activated / triggered ability item from the stack. Abilities cease to exist on removal (CR 608.2n); no destination needed. Caller must hold priority. Added in S13.1. |
| `activate_ability` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"?, "ability_index"?: int, "ref"?: string, "sacrifice_ids"?, "crew_ids"?, "counter_source_ids"?, "counter_counts"?, "counter_kind"?, "counter_kinds"?, "discard_ids"?, "exile_ids"?, "top_ids"?, "return_ids"?, "exile_permanent_ids"?, "reveal_ids"?, "waterbend_ids"?, "tap_ids"?, "strict"?, "auto_tap"? }` | Creates an activated-ability stack item linked to the source card. Same announce-time params as `cast_spell` (minus the source-zone routing). The source card stays in its origin zone; the stack item carries a synthetic ID. Mana abilities are NOT modeled this way (CR 605 — they don't use the stack). Caller must hold priority. Added in S13.1. **With `ability_index`** the payload names a catalog ability (S21 sub-PR 2) and the engine validates and pays the cost: `sacrifice_ids` pays a "Sacrifice a creature" or "Sacrifice two artifacts" component (exactly `activated_abilities[i].sacrifice_options.max` distinct permanents; see **Sacrifice costs of N permanents** below), `crew_ids` pays a crew number (CR 702.122a), and `x_value` is the X announced for an `{X}` in the ability's mana component (CR 602.2b). X is validated at announce — non-negative, zero unless the cost has an `{X}`, and at least the cost's printed floor — then locked onto the stack item, where the ability's effect reads it back. The card's `activated_abilities[i]` carries `demands_x`, `min_x` and `x_slots` so the client knows to prompt, what floor to enforce, and what a given X actually costs. **#625:** `counter_source_ids` (one `<uuid>`) and `counter_kind` pay a "remove N counters" component. The ability's view carries `counter_cost_n` (present means the ability has the component), `counter_cost_kind` (absent means "a counter" of any kind), `counter_cost_self` (the counters come off the source), `counter_cost_label` (the "from" clause, e.g. "a planeswalker you control") and `counter_cost_options: [{ "card_id": "<uuid>", "kinds": [{ "kind": string, "count": int }] }]`. The options list the permanents the viewer controls that could pay right now, most counters first. They come from the non-targeting candidate walk, so hexproof permanents are included, and the field is absent when nothing can pay. Send `counter_source_ids` for the other-permanent form. Omit it for the self form, or send the source's own ID. Send `counter_kind` for the any-kind form. It is optional for a printed kind, and if sent it must match that kind. The removal is paid at announce with the rest of the cost and is never modified by counter replacement effects. It is not a loyalty activation and has no timing restriction of its own. Errors: `not enough counters to pay that cost` when the named permanent holds too few, `you do not control that card` for another player's permanent, and `bad_request` for a payload that names a counter choice the ability has no component for. **#743:** an ability with an activation condition (CR 602.1b: "Activate only if an opponent controls four or more lands", "Activate only during your turn") is refused with `ability's activation condition is not met` while the condition is false, before X, targets or any cost are checked, so nothing is paid. The condition is checked only at activation, never at resolution. When sorcery speed and the condition both fail, the sorcery-speed error is the one returned. **#789:** the counter component grew three more printed shapes, and one payment shape covers all of them. `counter_counts`?: `[int, ...]` is the per-permanent split, parallel to `counter_source_ids`: omit it for a fixed one-permanent cost (the #625 shape, unchanged), send counts totalling exactly `counter_cost_n` for a removal spread across permanents (`counter_cost_among: true` — "Remove two +1/+1 counters from among artifacts you control"), and send the announced count for a removal whose size the activator chooses (`counter_cost_variable: true` — "Remove any number of storage counters"; `counter_cost_n` is then the FLOOR and `counter_cost_max` the most the viewer could name right now). A split payment is validated as a set, like a crew payment: every permanent distinct, controlled by the activator, matching the clause, holding at least the count named against it, and the counts totalling exactly N — any failure refuses the whole activation with no counter removed. **#943:** the last printed shape is an among removal of ANY kind (`counter_cost_among: true` with no `counter_cost_kind` — Tekuthal, Inquiry Dominus' "Remove three counters from among other artifacts, creatures, and planeswalkers you control"), where the parts may differ in kind as well as in count. `counter_kinds`?: `[string, ...]` is the per-permanent kind, parallel to `counter_source_ids`: send it only when the payment actually mixes kinds, and `counter_kind` (one kind for the whole payment) whenever one kind will do — so a payment of one kind is byte-for-byte the shape earlier clients send. If both are sent they must agree, an empty kind in the array is refused, and the set's identity is `(permanent, kind)`: one permanent may appear twice with two different kinds (it pays in both) and never twice with the same one. Nothing new is projected for the shape — a `counter_cost_options` row has always been a (permanent, kind) pair, so the existing option list already answers "which kinds, off which permanent". The other direction is `counter_cost_add` / `counter_cost_add_kind`, a cost that PUTS counters on the source (Devoted Druid's "put a -1/-1 counter on this creature"): nothing is chosen, so there is nothing to send, and `counter_add_blocked: true` means CR 118.3 refuses the activation because the permanent cannot have those counters. Such a cost is never doubled by a counter-doubling replacement — such a replacement applies only to a counter placed by an effect (CR 614.16). `activated_abilities[i].condition_unmet: true` marks an ability whose condition is false right now. It is absent when the ability has no condition or the condition holds, is evaluated with the permanent's controller as "you", and is sent to every viewer (a condition reads only public information). The client greys the row, as it does for `sorcery_speed`. **#917 (CR 107.4f / CR 602.2b):** `phyrexian_life`?: int is how many of the ability's Phyrexian mana symbols are paid with **2 life each** instead of mana — Birthing Pod's `{1}{G/P}`, Solphim's `{1}{R/P}{R/P}`. It is the same field name, meaning and validation `cast_spell` carries, because CR 602.2b asks the activator exactly what CR 601.2b asks the caster and the engine runs one strike-and-pay helper for both. Absent (0) pays every symbol with its coloured half. `activated_abilities[i].phyrexian_symbols` is the ceiling — how many symbols the cost prints — so the client never parses a mana string to size its stepper. The engine strikes the symbols a life payment can actually save first, refuses a count larger than the cost prints, refuses a payment CR 119.4 forbids (a player may pay life only down to 0), and refuses a claim against an ability with no mana component at all; every refusal is `bad_request` with nothing paid. The life is paid through the one cost-shaped life path before the pool is spent, the auto-tapper plans only the mana the activation still owes, and `PaidCost.LifePaid` on the stack item records the printed `Life` component and this together. **#660, widened by #1221:** an ability may function from a zone other than the battlefield (CR 113.6). Cycling and typecycling (CR 702.29a/e) function from HAND, unearth (CR 702.82a), scavenge (CR 702.96a), embalm (CR 702.128a) and eternalize (CR 702.129a) from a GRAVEYARD, and any of them may declare exile or the command zone; all of them reach the client on the card's `zone_abilities` rather than on `activated_abilities` (the field was `hand_abilities` through #660, and was renamed rather than duplicated when the other zones opened); `ability_index` is the ability's index in the card's full list either way, so the payload is unchanged. Activating an ability from a zone it does not function in is refused with `this ability cannot be activated from that zone`, before anything is validated and before anything is paid — a cycling fired at a permanent and a sacrifice outlet fired at a card in hand get the same error. Off the battlefield the card's OWNER is the only player who may activate it (CR 108.4). `discard_ids` pays a discard component: the general "Discard a creature card" clause (`discard_cost_n` / `discard_cost_label` / `discard_cost_options` on the ability's view) names exactly `discard_cost_n` distinct cards from the viewer's own hand, each matching the clause, and none of them the source of a hand activation. Cycling's "Discard this card" (`discard_self: true`) sends NOTHING — the source is the payment — and a payload that names cards for an ability with no discard component is refused. The discard is paid at announce, with the rest of the cost, through the engine's one discard path, so `EventDiscardCard` fires per card and every discard payoff sees it; being a cost it cannot pause, so a commander pitched to one goes to the graveyard, and the CR 903.9a state-based action offers it the command zone afterwards (CR 601.2h / 602.2b, ADR 0115, [ADR 0062](decisions/0062-abilities-and-special-actions-from-the-hand.md)). **#1221:** scavenge, embalm and eternalize pay a second self-cost, `exile_self: true` on the ability's view — "Exile this card from your graveyard". Like `discard_self` it is ADVISORY and sends nothing: the source IS the payment, so there is no pick and no field on the payload, and the keyword's own label spells the clause out. The card is exiled at announce, through the same exit every other move takes and with the same cost-cannot-pause bit, and the ability's effect reads the card back out of exile afterwards. An ability that declares the component on a zone it cannot be paid from is refused at server boot, not at activation. **#1404:** the same `exile_self: true` now also appears on a BATTLEFIELD ability, in `activated_abilities`: "Exile this artifact" (Perpetual Timepiece) or "Exile this creature" (Hanged Executioner). The payload does not change, and nothing is sent for it. The permanent is exiled at announce, so it has left the battlefield before anyone can respond. Leaves-the-battlefield triggers see it; dies and sacrifice triggers do not. A commander exiled this way is asked about the command zone BEFORE anything is paid, as every commander paid as a cost is (#1397): the owner's `optional_replacement` answer makes the activation, which then pays without pausing. **#2028:** `return_self: true` on a BATTLEFIELD ability is "Return this enchantment to its owner's hand" (Gossamer Chains) or "Return Shigeki to its owner's hand". It is advisory like `exile_self`, and nothing is sent for it. The permanent goes to its OWNER's hand at announce, so it has left the battlefield before anyone can respond; leaves-the-battlefield triggers see it and the ability's effect reads it as it last existed there. A commander that returns itself is asked about the command zone before anything is paid, as every commander returned as a cost is. The source can't also be named in `sacrifice_ids`, `return_ids` or `exile_permanent_ids` (`bad_request`). **#1297:** `exile_ids`?: `["<uuid>", ...]` pays an "Exile N cards from your graveyard" / "Exile a card from your hand" component — Grim Lavamancer's "Exile two cards from your graveyard", Moorland Haunt's "a creature card", Holistic Wisdom's "a card from your hand" — exactly `activated_abilities[i].exile_cost_n` distinct cards from `exile_cost_options`, none of them the source and none also named in `discard_ids`; see **Exile-N-cards costs** below. It is the mana ability's `exile_ids` (#1283) with a second owner. **ADR 0109 §7 (#1902):** `top_ids`?: `["<uuid>", ...]` pays a "Put a card from your hand on top of your library" component (Penance, Leashling) — exactly `activated_abilities[i].top_cost_n` distinct cards from `top_cost_options`, none of them the source and none also named in `discard_ids` or `exile_ids`; an "Exile the top N cards of your library" component and a "Discard a card at random" component send nothing. See **Library costs and random discards** below. **#1208:** `activated_abilities[i].timing_closed: true` says the engine will refuse this activation RIGHT NOW for timing (CR 602.5d's "activate only as a sorcery", CR 606.3's loyalty window), as narrowed or OPENED by any per-player statement on the battlefield — The Wandering Emperor's "as long as she entered this turn, you may activate her loyalty abilities any time you could cast an instant", Leonin Shikari's "you may activate equip abilities any time you could cast an instant". It is the stamp of the one read `activate_ability` and the bot's move enumerator also ask, and it is the field a client greys the row on: `sorcery_speed` beside it is only the ability's PRINTED clause and does not change when a statement opens the window. Absent means the engine has no timing objection, which is every instant-speed ability at every moment; it is never stamped on a mana ability, because CR 605.3a gives one its own window and no timing statement reaches it (a card that stops mana abilities is a BAN and arrives as `cant_activate`). Evaluated once with the permanent's controller as "you" and sent to every viewer, exactly as `condition_unmet` is. **#2697 (CR 702.142a):** `activated_abilities[i].boast_blocked` is `"not_attacked"` or `"used"` on a boast ability ("Activate only if this creature attacked this turn and only once each turn") the engine will refuse right now, and absent otherwise, which is every other ability and every boast ability nothing objects to. `"not_attacked"` means the creature has not attacked this turn; `"used"` means the ability has been activated as many times this turn as the creature may (once, or twice under Birgi, God of Storytelling). It is the stamp of the one gate `activate_ability` and the bot's move enumerator also ask, and the activation is refused with `this creature hasn't attacked this turn, so its boast ability can't be activated` or `this boast ability has already been activated as many times as it can be this turn` before anything is paid. Two values rather than a flag because the halves recover differently, and the client owns the sentence. Public, like `condition_unmet`. **ADR 0093 (#754):** `ref`? is the row's stable ref (`activated_abilities[i].ref`); a ref that no longer names the row at `ability_index` is refused with `that ability is no longer at that position — the board changed` before anything is paid, and an absent ref is accepted. See **Granted abilities and ability refs** below. |
| `activate_loyalty` | yes | `{ "planeswalker_id": "<uuid>", "label"?: string, "delta": int }` | Applies a planeswalker loyalty ability. Sandbox shape: the engine doesn't model the activation as a proper stack item (deferred to S14+) — `delta` is applied immediately to the planeswalker's `loyalty` counter. Sorcery-speed gated; once-per-turn-per-planeswalker (CR 606.3). Caller must hold priority. Added in S13.1. |
| `announce_trigger` | yes | `{ "source_card_id": "<uuid>", "label"?: string, "targets"?, "modes"?, "x_value"?, "distribution"? }` | Queues a triggered ability for APNAP-ordered drain onto the stack (CR 603.3b). The drain happens at every priority-grant boundary inside the SBA + trigger loop. Sandbox: the player whose card has the trigger announces it manually — auto-fire from card events is the bright line between S13.1 and S14+. Added in S13.1. |
| `mark_damage` | no | `{ "instance_id": "<uuid>", "delta": int }` | Adjusts the damage noted on a creature on the battlefield by `delta` (positive to add, negative to remove). Drives the lethal-damage SBA (CR 704.5g). Positive deltas require controller (admins / spectators bypass); negative deltas are caller-loose so opponents can undo. SBA loop fires after the mutation. Cleanup-step turn-based action zeros every creature's marked damage (CR 514.2). Added in S13.1. |
| `add_player_counter` | yes | `{ "name": string, "delta": int }` | Modifies a named player-level counter (poison, energy, experience, rad, plus homebrew names) by `delta`. Negative deltas clamp at zero. Drives the SBA loop after the mutation so 10 poison or future counter-driven losses fire immediately. The legacy `set_poison` / `set_energy` actions stay synchronised with the underlying `Player.Counters` map. Player-scoped: a seated caller may only adjust their own counters; admins bypass. Added in S13.2. |
| `discard_selection` | yes | `{ "card_ids": ["<uuid>", ...] }` | Resolves the cleanup-step interactive discard prompt (S13.4, CR 402.2). Caller must be in `discard_pending`; the count must exactly match the over-max amount; every supplied card ID must live in the caller's hand. Selected cards move to the caller's graveyard, the caller's pending entry clears, and the cleanup hook re-fires so the cursor resumes its auto-advance. Idempotent no-op for callers not in the pending map. Player-scoped. Added in S13.4. |
| `set_max_hand_size` | yes | `{ "value": int }` | Sets the named player's own maximum-hand-size grant (CR 402.2). Default 7; `-1` is no maximum. The grant is stamped when it is set and folded with the battlefield's maximum-hand-size statics in CR 613.11 timestamp order ([ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md) §3), so a permanent that arrives later still applies after it. Sandbox helper. Player-scoped. Added in S13.4. |
| `resolve_choice` | yes | `{ "choice_id": "<uuid>", "card_ids"?: ["<uuid>", ...], "color"?: "W" \| "U" \| "B" \| "R" \| "G" \| "C", "apply"?: bool, "tap_ids"?: ["<uuid>", ...], "order"?: ["<id>", ...], "assignments"?: [{ "blocker_id": "<uuid>", "amount": int }, ...], "trample_to_player"?: int, "creature_type"?: string, "card_name"?: string, "distribution"?: { "<uuid>": int } }` | Resolves a pending async effect-driven choice (S14, CR 701.9b — "player chooses" discard). Caller must be the `chooser` of the named entry in `GameView.pending_choices`; `card_ids.length` must equal the entry's `count`; every supplied card ID must currently live in the entry's `from_player`'s hand. **ADR 0116:** for `discard_from_hand`, every card ID must also be in the entry's `eligible` list when it has one, and no ID may appear twice; anything else is refused and the entry stays open. **#2115:** `kind` may be `revealed_hand_pick`, the same pick with a variant. `card_ids` may be EMPTY when the entry's `choose_min` is 0 ("you may choose"); otherwise it holds exactly `count` IDs from `eligible`. With `pick_from_graveyard` set, a card may be in `from_player`'s graveyard as well as the hand. With `pick_destination: "exile"` the card is exiled and no `EventDiscardCard` fires, because exiling a card is not discarding it (CR 701.9a). Picked cards move to `from_player`'s graveyard (for `discard_from_hand`), the entry drains from the queue, and an `EventDiscardCard` event fires per card. Distinct from `discard_selection` (S13.4 cleanup, chooser == discarder); `resolve_choice` covers effects where chooser != discarder (Thoughtseize — caster picks from target's revealed hand). **S15:** the entry's `kind` may be `mana_pick` (Birds of Paradise / Arcane Signet); `card_ids` is omitted and `color` carries the chosen single-letter color, validated against the entry's `color_options` slice. The picked color drops into the chooser's mana pool as one `ManaToken`. **S17:** `kind` may be `replacement_order` (CR 616 multi-replacement order prompt — `order` carries the chosen permutation of replacement-effect IDs) or `optional_replacement` (a "may" yes/no — `apply` is the answer). **#1397:** an `optional_replacement` may also be the CR 903.9 question asked BEFORE a cost that moves a commander is paid (since [ADR 0115](decisions/0115-commanders-die.md), only a cost that puts it into a hand or a library: a sacrificed, discarded or exiled commander is paid and then offered by `commander_return`) ([ADR 0013 §5af](decisions/0013-replacement-effects.md#5af-amendment-2026-09-24-a-commander-paid-as-a-cost-is-asked-before-the-payment-not-during-it)). The cast or activation that asked is parked with nothing paid; answering it makes that cast or activation, so the answer can be followed by the spell or ability arriving on the stack. Same payload and same modal; the `chooser` is the commander's OWNER, who may not be the player paying, and `source` is the commander. **S18:** `kind` may be `damage_assignment` (CR 510.1c multi-blocker assignment — `assignments` is the per-blocker amounts, `trample_to_player` is the overflow when the attacker has trample). **#2692:** any split that adds up to `attacker_power` is accepted, in any order, because CR 510.1c divides the damage "as its controller chooses" with no damage assignment order; `trample_to_player` above 0 needs every blocker assigned lethal damage (CR 702.19b). **#1706:** the same kind carries CR 510.1d's division — `damage_assignment.blocker_divides: true`, `attacker_card_id` a BLOCKER that blocks every creature in `blocker_card_ids`, and the chooser its controller. The answer's shape is the same (`assignments` keyed by the attackers' ids); any split that adds up to `attacker_power` is accepted — no lethal-first order — and `trample_to_player` must be 0. **S19:** `kind` may be `trigger_prompt` (CR 603.5 "you may" optional triggered ability — `apply` is the answer; the server routes by inspecting the choice's kind, since `optional_replacement` and `trigger_prompt` share the `apply` payload). **S21:** `kind` may be `scry` (CR 701.22 — `bottom` is the looked-at cards going under the library and `top_order` the ones staying on top, listed TOP-FIRST; every looked-at card must appear in exactly one list, because a scry moves all of them). **S22:** `kind` may be `surveil` (CR 701.25 — the same shape with `graveyard` in place of `bottom`) or `look_at_top` ("look at the top N cards of your library, then put them back in any order" — Ponder, Sensei's Divining Top; answered with `top_order` alone, since there is no away lane). The three are the **scry family** and the dispatcher routes them **by the pending choice's `kind`**, not by the payload shape every other `resolve_choice` branch keys off: all three carry `top_order` and `look_at_top` carries nothing else, so no key's presence identifies them. Send the destination key that matches the kind — `bottom` for scry, `graveyard` for surveil, neither for `look_at_top`. **S26:** `kind` may be `choose_creature_type` (CR 614.12 "as this permanent enters, choose a creature type" — Cavern of Souls, Door of Destinies, Vanquisher's Banner, Adaptive Automaton); `creature_type` carries one name from the CR 205.3m vocabulary, validated and normalised server-side and stamped onto the source permanent's named tribe. Routed by presence. **#742:** `kind` may be `choose_color` (CR 105.4 "choose a color" — as a permanent enters, Coldsteel Heart / the Thriving lands, where the answer is stored on the permanent; or while a spell or ability resolves, Wash Out / Oona). It is answered with the same `color` field a `mana_pick` uses, validated against the entry's `color_options` (a subset of W/U/B/R/G, never C), and the dispatcher routes `color` **by the pending choice's `kind`**: the presence of `color` alone no longer says which resolver an answer belongs to. A `mana_pick` entry that carries `color_amounts` mints that many tokens of the picked colour instead of one. Added in S14. **#764:** `kind` may be `mode_pick` (CR 603.3c — a modal TRIGGERED ability's bullet, chosen as the ability is put on the stack, after the "you may" prompt and before the CR 603.3d target pick). It is answered with `modes`: the chosen ModeSpec indexes **in the order chosen**, repeats allowed only when `mode_repeatable`. The prompt carries `mode_options` (the oracle bullets), `mode_indexes` (the ModeSpec index each label belongs to — a bullet with no legal target is dropped server-side, so the row index is not the answer), `mode_min`, `mode_max` and `mode_repeatable`. The dispatcher routes this kind **by the choice's `kind`**, like `coin_call` and `loop_shortcut`: an index list of zeroes (`[0, 0, 0]` is Mystic Confluence drawing three cards) and the empty list are both ordinary answers, so there is no presence to route on. **#996 ([ADR 0088](decisions/0088-ordered-library-placement.md)):** `kind` may be `put_in_library` — "put these cards on top of / on the bottom of the library in any order" (Brainstorm's put-back, "the rest on the bottom of your library in any order", Aetherspouts' "top or bottom"), the family's fourth member and routed the same way. It is answered with scry's two keys, `top_order` and `bottom`, **both TOP-FIRST** (the first entry of `bottom` is the card nearest the top of the pile that goes under the library; the last is the bottom card). The entry's `placement` says which lanes are open — `top` (`top_order` only), `bottom` (`bottom` only) or `top_or_bottom` (both) — and an answer that puts a card in a closed lane is refused. Every card appears in exactly one list. **#1298:** the entry may also carry `top_count` — EXACTLY how many cards `top_order` must hold (Cream of the Crop's "put one of those cards on top"; any other count is refused) — and `top_depth` — where the top lane lands, counted from the top (2 is Temporal Cleansing's "second from the top"; absent is the top). The chooser need not own the cards: Jace, the Mind Sculptor's +2 and Portent order another player's library, and Hinder's counterer picks the end of the countered spell's owner's library. A scry's `bottom` has always been applied in the order sent, and is top-first by the same rule. **#1210 (CR 614.12, [ADR 0073 amendment 2026-09-22](decisions/0073-optional-additional-costs-and-the-cast-gate.md#amendment-2026-09-22-1210-the-activation-gate-and-the-chosen-card-name)):** `kind` may be `choose_card_name` — "as this permanent enters, choose a card name" (Pithing Needle, Phyrexian Revoker, Sorcerous Spyglass, Meddling Mage, Nevermore), the fourth as-enters choice after creature type, colour and player. `card_name` carries FREE TEXT, not a vocabulary pick like `creature_type`: CR 201.2 lets a player name any card at all, including one this server has never imported, so the engine checks only that it's trimmed, non-empty and under `MaxChosenNameLen` (200 characters) — it never rejects an unrecognised name. The prompt's `name_options` is a convenience suggestion list built from cards in PUBLIC zones only (so it can never leak a hidden card), never a closed set the answer is checked against. Matching a chosen name against a card is case-insensitive and checks every face of a split or transforming card (CR 201.2b), so naming "Fire" also matches "Fire // Ice". Routed by presence, like `creature_type`. **ADR 0131 §2 (#2531):** `phyrexian_life`? answers a `pay_unless` prompt for a mana cost (`pay_cost`): with `apply: true`, it is how many of the cost's symbols the chooser pays 2 life each for instead of the mana (a printed `{B/P}`, or a `{B}` under K'rrik, Son of Yawgmoth). The prompt's `phyrexian_symbols` is the ceiling and `phyrexian_granted` how many of those are the grant's. A claim past the ceiling, below 2 life, under a life lock, with `apply: false` or on a non-mana payment is refused with `bad_request` and the prompt stays open; one whose remaining mana cannot be paid is a decline and costs no life. Auto-tap never pays it. |
| `activate_mana_ability` | yes | `{ "card_id": "<uuid>", "ability_index"?: int, "ref"?: string, "sacrifice_ids"?, "counter_source_ids"?, "counter_counts"?, "counter_kind"?, "counter_kinds"?, "discard_ids"?, "exile_ids"?, "exile_permanent_ids"?, "color"?, "colors"?, "auto_tap"? }` | S15. Activates the `ability_index`-th mana ability on a battlefield permanent the caller controls. `ability_index` defaults to 0 (the only ability for basic lands and most rocks). Single code path for catalog-declared abilities (Sol Ring, Arcane Signet, Birds of Paradise — looked up via the `CatalogManaAbilities` hook) and synthetic basic-land abilities (Forest → `{G}`) the engine derives from `TypeLine`. Pays the tap cost (rejects with `bad_request: "card already tapped"` if the source is tapped), then materialises the produced mana: single-color slots drop straight into the controller's pool as one `ManaToken`; multi-option slots (pipe syntax) queue a `mana_pick` PendingChoice for the controller. **#742:** a pipe option may carry a count (`{W3|U3|B3|R3|G3}`, "three mana of any one color"), which queues ONE `mana_pick` with `color_amounts` rather than one pick per mana. Mana abilities don't use the stack (CR 605.3) — synchronous under the write lock. Caller need NOT hold priority (mana abilities are special, CR 605.1a). Player-scoped: caller must control the source. **S21 / #747:** `sacrifice_ids`?: `["<uuid>", ...]` pays a mana ability's "Sacrifice a creature" component (Ashnod's Altar): exactly `mana_abilities[i].sacrifice_options.max` distinct permanents; see **Sacrifice costs of N permanents** below. **#743:** `mana_abilities[i].condition_unmet: true` is the same flag for a mana ability whose activation condition is false right now (Temple of the False God with four lands, Mox Opal without metalcraft). Same evaluation and visibility as `activated_abilities[i].condition_unmet`; the activation is refused with `ability's activation condition is not met` and nothing is tapped or paid. **#844 (CR 903.4f):** `mana_abilities[i].adds_no_mana: true` marks an ability whose printed text says "any color in your commander's color identity" (Command Tower, Arcane Signet, Commander's Sphere, Path of Ancestry) activated by a controller with no commander, or one whose commander's colour identity is colourless (Kozilek, Karn). That quality is undefined or empty, so the ability adds no mana. **ADR 0117 §5:** the flag also marks an ability whose output is computed and computes to no mana right now (a Vivi Ornitier whose power is 0, a Selvala, Heart of the Wilds whose greatest power is 0, an Exotic Orchard with nothing to copy, a Mage-Ring Network with no counters); a static empty declaration is not flagged. In every case the activation is NOT refused (the ability exists and may be activated; it simply does nothing, which is what the rulings on those four cards say), but no mana token is minted and no `mana_pick` is queued. Nothing offers the activation either (the engine still accepts it, CR 605.1a) — the client greys the row on this flag, and the bot's legal-move enumerator and the auto-tapper both skip the source. Absent for every other ability. **#764:** `modes` announces a modal activated ability's bullets (CR 602.2b) in the SAME message as the targets — activating is one indivisible step, so there is no prompt. `activated_abilities[i].modes` is the ModeSpec (same shape as `CardView.modes`) and `activated_abilities[i].clauses` its clause list when it has more than one; the target entries carry `slot` / `mode` exactly as a cast's do. **#789:** a mana ability may carry a COUNTER component, with exactly the field names, meanings and validation `activate_ability` uses — one component with two owners. Vivid Creek's "{T}, Remove a charge counter from this land" sends nothing at all (the self form with a printed kind and count); Mage-Ring Network's "Remove any number of storage counters" sends `counter_counts`. The ability's view carries the same `counter_cost_*` fields an activated ability's does. The auto-tapper only plans such a source when it can both decide and afford the cost — the self form, a printed kind, a fixed count, and enough counters right now — so a Vivid land with no charge counters left is not a five-colour source and is not planned as one. **#943:** `counter_kinds` is accepted here too, with the same meaning — one component, one payment shape, whichever ability kind carries it — though no printed mana ability needs it today. **#1283:** `exile_ids`?: `["<uuid>", ...]` pays an "Exile a card from your hand" component (Cadaverous Bloom) — exactly `mana_abilities[i].exile_cost_n` distinct cards from `exile_cost_options`; see **Exile-a-card costs on a MANA ability** below. **#1443:** `color`?: `"W" \| "U" \| "B" \| "R" \| "G" \| "C"` names the colour a pipe slot adds BEFORE the source is tapped, so no `mana_pick` is queued: the colour is produced straight into the pool, through the same production the answered pick would have used (restrictions, the CR 106.12b window, the slot's amount). `colors`?: `["W", ...]` is the same for an output with several picking slots — one entry per list in `mana_abilities[i].color_options`, in output order (Mystic Gate's `{W|U}{W|U}` takes two). Send one or the other, never both (`bad_request`). The count must match `color_options` and each colour must be in its slot's list, checked before anything is validated or paid; otherwise the activation is refused with `that colour is not one this mana ability can add` and nothing is tapped, paid or produced. Absent is the ordinary activation, unchanged — every picking slot queues its `mana_pick` — and is what the auto-tapper and the bot seats use. See **Colour options on a MANA ability** below. **ADR 0093 (#754):** `ref`? is the row's stable ref (`mana_abilities[i].ref`), checked exactly as `activate_ability`'s is. See **Granted abilities and ability refs** below. **#2215:** `auto_tap`?: `true` lets a MANA component of the cost (Crystal Quarry's `{5}, {T}`, a Signet's `{1}, {T}`) be paid by tapping the caller's other sources for whatever the floating pool is missing, through the same auto-tapper and pool top-up a cast uses (CR 605.3a allows activating mana abilities while paying for one). The source itself and every card another component names are never tapped for it. It is planned before anything is paid, so an unpayable cost is still refused with `insufficient_mana` and nothing tapped; without the flag the mana must already be floating. The web client sends it on every mana activation, and the legal-move enumerator on every move whose ability has a mana component. **ADR 0131 §2 (#2531):** `phyrexian_life`?: how many symbols of the ability's mana component (`mana_abilities[i].mana_cost`) are paid with 2 life each instead of mana, the field `cast_spell` and `activate_ability` use. The row's `phyrexian_symbols` is the ceiling and `phyrexian_granted` how many of those are K'rrik's. Refused with `bad_request` before anything is paid when it exceeds the symbols, when the ability has no mana component, or when the caller cannot pay the life (CR 119.4, CR 119.8, and the ability's own `life_cost` counts toward the total). Auto-tap never claims it. |
| `special_action` | yes | `{ "card_id": "<uuid>", "kind": "foretell", "suspend", "plot", "turn_face_up" or "unlock", "strict"?: bool, "auto_tap"?: bool, "cost"?: "<printed cost>", "door"?: "left" \| "right" }` | A CR 116.2 **special action**: a game action taken without using the stack and without passing priority, so there is no announce, no response window and nothing to respond to. Caller must hold priority; the caller keeps it. `kind` selects the action — `foretell` (CR 702.143a: pay {2} and exile the card from your hand face down, castable on a later turn for its foretell cost), `suspend` (CR 702.62a: pay the suspend cost and exile the card with N time counters), since #1342 `plot` (CR 702.170a: pay the plot cost and exile the card from your hand face up; it becomes plotted, castable without paying its mana cost in its owner's main phase with the stack empty on any later turn — the cast is an ordinary `cast_spell` with `from_zone: "exile"` under the card's granted permission) and, since #1194, `turn_face_up` (CR 708.6: pay a face-down permanent's morph cost and turn it face up). **Where the card is, is per kind:** foretell, suspend and plot act on a card in the CALLER'S OWN HAND, which must print the keyword — or, since #1391, the top card of the caller's OWN LIBRARY when a permanent the caller controls grants the kind there (Fblthp, Lost on the Range: plot a nonland card from the top of your library, for its mana cost or for its own plot cost; a card on top of a library with no such grant is `card not found`). `cost` picks between two offers of the same kind on one card by the offer's printed cost — a Djinn of Fool's Fall on top under Fblthp offers plot for {3}{U} and for {4}{U} — and is omitted to take the first; `turn_face_up` acts on a face-down permanent on the BATTLEFIELD that the caller CONTROLS (CR 708.5), and a permanent someone else controls is reported as `card not found` rather than as a refusal, because a face-down permanent's identity is not theirs to probe. A card that offers no such action is refused with `this card offers no such special action`, before anything is paid — which covers a face-up permanent, a manifested noncreature card (CR 701.40b: it stays face down) and a permanent an effect turned face down whose card prints no morph or disguise either (CR 708.7; since #1209 one that DOES print morph keeps its way back up, because CR 702.37e keys on the card having the ability and not on how the permanent came to be face down). **Timing is per kind, and the engine is the only place it is written down:** foretell any time you have priority during YOUR turn and **legal under split second** (CR 702.61b — a special action is neither a cast nor an activation); suspend any time you could begin to CAST the card, which means sorcery timing for a sorcery, instant timing for an instant or a card with flash, and **not** under split second (CR 702.62c); plot only in your own main phase with the stack empty (CR 702.170a), whatever the card's type or flash says; `turn_face_up` any time you have priority, on any turn, with anything on the stack, and **legal under split second** for the same reason foretell is (CR 702.37c). Outside the window the refusal is `that special action cannot be taken right now`. The cost is mana and nothing else — no targets, no modes, no picker — and `strict` / `auto_tap` behave exactly as they do on `cast_spell`. Mana restricted to casting spells or activating abilities cannot pay for a special action, because it is neither. ADR 0062 Decision 4; #658 / #659 / #1342. **ADR 0103:** `unlock` (CR 709.5e, 116.2m) pays the mana cost of a LOCKED door of a Room the caller CONTROLS on the battlefield and gives the Room that door's unlocked designation. `door` is required for it (two doors can print the same cost) and refused on every other kind. Its window is plot's: your own main phase, stack empty. Mana restricted to casting or activating cannot pay it; mana restricted to unlocking doors can. A door that is already unlocked is `this card offers no such special action`. |
| `sacrifice_permanent` | yes | `{ "instance_id": "<uuid>" }` | CR 701.21 sacrifice: moves a battlefield permanent to its owner's graveyard and emits `EventSacrifice`, so "whenever you sacrifice a permanent" payoffs fire — distinct from a manual `move_card` to the graveyard, which does neither. Caller must control the card (admin sessions bypass) — see "controller-only card actions" below. `player` must be the permanent's controller: `game.SacrificePermanent` checks that directly (CR 701.21a — only a permanent's controller may sacrifice it), so even an admin session names the real controller in `player` rather than leaving it empty. Sandbox affordance rather than a stack action — deliberately absent from the bot's legal-move enumerator (`internal/legal`), alongside `advance_step` and `activate_loyalty`. Added in S21 sub-PR 1. |

`activate_mana_ability` also accepts `tap_ids?: ["<uuid>", ...]` for a
mana ability's fixed-count `tap_others_options` cost (#758). This field is
listed with the complete cost shape in **Tap-other-permanent costs** below.

`<ZoneRef>` is `{ "kind": "<zone_kind>", "owner": "<uuid>" }`. Owner is
omitted for shared zones (`battlefield`, `stack`, `exile`). Zone kinds are
`library`, `hand`, `battlefield`, `graveyard`, `exile`, `command`, `stack`.

#### Controller-only card actions (added in S08.5)

`tap`, `untap`, `move_card`, `add_counter`, `set_battlefield_position`,
`declare_attacker`, `declare_attackers`, `declare_blocker`,
`declare_blockers`, and `sacrifice_permanent` are gated on the caller being
the current controller of the named card. (`declare_attackers` and
`declare_blockers` apply the gate to every entry in the batch, and one
foreign creature rejects the whole batch.) A seated player issuing one
of these actions against a card controlled by a different seat receives
an `error` frame with `payload.code = "bad_request"` and
`payload.message = "you do not control that card"`. Admin connections
bypass the gate so a moderator can fix wedged board state. An admin
bound to a seat, their own included, keeps the bypass for `tap`,
`untap`, `move_card`, `add_counter`, `sacrifice_permanent`,
`mark_damage` and `set_battlefield_position` (the admin context menu's
overrides), and plays every other gate, the combat declarations
included, as that seat (ADR 0110 §3). `clear_combat` intentionally stays loose — it's the "combat is
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

### `choose_source` — "a source of your choice" (ADR 0107 §6, #1860, CR 609.7a)

`pending_choices` may carry `kind: "choose_source"`. It is asked as a
damage-prevention shield is made — "The next time a red source of your
choice would deal damage to you this turn, prevent that damage"
(Circle of Protection: Red, Deflecting Palm) — and the rest of the card
waits on it, so it blocks the table.

It carries and is answered exactly like `choose_cards` with a floor and
ceiling of one: `options[]` are the candidates, `choose_min` and
`choose_max` are both `1`, and `resolve_choice { choice_id, card_ids:
[id] }` names the source. The candidates are everything CR 609.7a
allows that has the card's property ("a red source"): permanents,
spells on the stack, face-up cards in a command zone, and cards in a
graveyard or exile that an item on the stack, a waiting prevention or
replacement effect or a waiting delayed trigger refers to. A card in a
hand or a library is never offered. All of them are public, so the
options and the bounds reach every seat. The client captions each one
with its controller and its zone.

The property is checked again when the source would deal the damage
(CR 615.9): if it no longer matches, the damage is dealt and the shield
is not used up.

### `ring_bearer` — choose your Ring-bearer (ADR 0114 §4, #2076, CR 701.54a)

`pending_choices` may carry `kind: "ring_bearer"`. It is asked when the
Ring tempts a player who controls two or more creatures: "choose a
creature you control" becomes their Ring-bearer. With exactly one
creature the server chooses it and says so in the log (owner decision
2); with none, nothing is asked. The rest of the tempting spell or
ability waits on the answer, so it blocks the table. Its `reason` is
`"choose your Ring-bearer"`.

It carries and is answered exactly like `choose_cards` with a floor and
ceiling of one: `options[]` are the chooser's creatures, `choose_min`
and `choose_max` are both `1`, and `resolve_choice { choice_id,
card_ids: [id] }` names the creature. It is not targeting: hexproof,
shroud, protection and ward do not apply. The candidates are battlefield
creatures, so the options and the bounds reach every seat. A candidate
that leaves the battlefield while the prompt is open is taken off it; if
none is left, the prompt is withdrawn and the Ring has still tempted the
player, with no Ring-bearer chosen (CR 701.54d).

### `proliferate` — choose what to proliferate (ADR 0013 amendment 2026-10-07, #2525, CR 701.34a)

`pending_choices` may carry `kind: "proliferate"`. It is asked of the
proliferating player once the proliferate has settled (so "proliferate
twice" asks twice, the second over the board the first answer left), and
the rest of the card waits on it, so it blocks the table. It is only
asked when some permanent or player has a counter.

It carries and is answered like `choose_cards` with a floor of zero:
`options[]` are the battlefield permanents with at least one counter,
`choose_min` is `0` and `choose_max` is the number of candidates,
permanents and seats together. Two fields are specific to the kind:

- `choose_players[]` — the seats that have a counter (player IDs). A seat
  is picked by sending its player ID in the answer's `card_ids`, beside
  the permanents.
- `choose_suggested[]` — the engine's beneficial pick (instance IDs and
  player IDs): everything of the chooser's a counter helps and everything
  of an opponent's one hurts. A default for the client to pre-select, not
  a rule; any subset is a legal answer.

`resolve_choice { choice_id, card_ids: [id…] }` names the picks, and an
empty list is a legal answer ("choose none"). Each pick gets one more
counter of every kind it already has. Counters are on the table, so the
options, the seats and the suggestion reach every seat unredacted.

### `divide_shield` — which damage a charged shield prevents (ADR 0108 §7, #1904, CR 615.7)

`pending_choices` may carry `kind: "divide_shield"`. It is asked when a
charged prevention shield — "prevent the next 3 damage that would be
dealt to any target this turn" (Mending Hands), "the next 3 damage that
a source of your choice would deal to you and/or permanents you control"
— would meet two or more damage events dealt at the same time (one
damage instance) for more than the charge it has left. The protected
player (or the controller of the protected permanent) chooses which of
that damage the shield prevents, before any of it is dealt; the damage
waits on the answer, so it blocks the table. When the charge covers all
of it, nothing is asked.

It carries `divide_shield: { label?, charge, entries: [{ id, source_id,
source_name?, target_id, target_name?, target_is_player?, amount,
combat? }] }`: one entry per source and recipient, `amount` being the
damage as it would be dealt before any other replacement. It is answered
with `resolve_choice { choice_id, distribution: { <entry id>: share } }`,
each share between 0 and its entry's `amount` and the shares adding up
to exactly `charge` (every 1 damage the shield can prevent, it does). An
entry given nothing is not prevented at all; an entry given N has at
most N prevented. Damage that can't be prevented is never an entry. If
the chooser leaves the game, the shield is divided in the order the
damage was opened and the damage is dealt.

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

### `entry_controller` — choosing who a permanent enters under (ADR 0102, CR 614.12a)

`pending_choices` may carry `kind: "entry_controller"`. It is the
question behind "This enchantment enters under the control of an
opponent of your choice" (Captive Audience, Pendant of Prosperity),
asked as the permanent would enter.

- **The options are seats.** They ride `pick_options`, exactly as a
  player-choosing `option_pick`'s do: one entry per opponent of the
  chooser still in the game, in turn order starting after the chooser,
  each with `player` set to the seat and `label` to its name.
- **The answer is `resolve_choice { choice_id, option_index }`**, the
  same payload `option_pick` takes. Every offered seat is a legal
  answer.
- **`control_purpose`** is `"harm"` (the permanent hurts whoever
  controls it) or `"benefit"` (it helps them). It is a reading of the
  card's own text, and the client uses it only for wording.

The permanent has not entered while the prompt is open: the CR 614
pipeline is suspended on the answer, nothing is on the battlefield and
no ETB trigger has fired. Like every prompt that pauses an entry, it
blocks the table. The permanent then lands under the chosen seat — it
was never under the chooser, so no control-change line is written — and
the choice is narrated as `choose_controller`.

There is no prompt when the choice is forced: with one eligible opponent
(every two-player game) that opponent is used, and an entry that cannot
pause uses the first opponent in turn order after the chooser. An
offered seat that leaves the game comes off the open prompt.

### `entry_riot` — riot's counter or haste (ADR 0109 §10, CR 702.136a)

`pending_choices` may carry `kind: "entry_riot"`. It is riot's "You may
have this permanent enter with an additional +1/+1 counter on it. If you
don't, it gains haste", asked of the would-be controller before the
permanent enters (CR 614.12a). `source` is the entering card,
`entry_keyword` is `"riot"`, and `accept_label` / `decline_label` name
the two answers ("+1/+1 counter", "Haste").

- **The answer is `resolve_choice { choice_id, apply }`**, the payload
  of the yes/no kinds: `apply: true` takes the counter, `apply: false`
  takes haste. Riot is mandatory, so both answers are always accepted
  and there is no third.
- **One prompt per instance.** Which riots apply is read off the
  permanent as it would exist on the battlefield (CR 614.12), so a
  creature with printed riot under Rhythm of the Wild is asked twice
  (CR 702.136b), and one entering under an effect that removes its
  abilities is not asked at all.
- **The haste is shown.** A permanent that took haste carries
  `CardView.riot_haste: true` (omitempty, public, battlefield only,
  cleared by the face-down redaction like the ability list), and the
  client labels its haste chip "Riot".

An entry that cannot pause takes the counter. Like every prompt that
pauses an entry, it blocks the table.

**Unleash** (CR 702.98a) is asked through the existing
`optional_replacement` prompt — "you may have this permanent enter with
an additional +1/+1 counter on it" — with `entry_keyword: "unleash"` and
`source` the entering card. Its other half, "can't block as long as it
has a +1/+1 counter on it", is an ordinary can't-block restriction the
block gate reads.

### `entry_read_ahead` — the chapter a Saga starts on (#2123, CR 702.155b)

`pending_choices` may carry `kind: "entry_read_ahead"`. It is read
ahead's "As this Saga enters, choose a number between one and this
Saga's final chapter number", asked of the would-be controller before
the Saga enters (CR 614.12a). `source` is the entering card and
`entry_keyword` is `"read ahead"`.

- **The options are the chapters.** They ride `pick_options`, one per
  chapter in order, labelled "Chapter I", "Chapter II" and so on.
- **The answer is `resolve_choice { choice_id, option_index }`**, the
  payload `option_pick` and `entry_controller` take: option N is
  chapter N+1, and the Saga enters with that many lore counters. Every
  offered chapter is a legal answer.
- **Skipped chapters never trigger.** The turn a Saga with read ahead
  entered, a chapter triggers only if the Saga has exactly that
  chapter's number of lore counters (CR 702.155a).

The answer is narrated as `choose_option` ("chose chapter III for The
Cruelty of Gix"). A Saga with one chapter is not asked, and an entry
that cannot pause starts on chapter I. Like every prompt that pauses an
entry, it blocks the table.

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

**A retarget TO a fixed object** (#1743 — Spellskite, Mizzium Meddler,
Hydroelectric Specimen: "change a target … to this creature") uses the
same `kind: "retarget"` and the same payload, but its options mean
something else: they are the spell's or ability's CURRENT targets that
could legally become that object, and clicking one changes that target
to it. The destination is never among the options and is never sent.
It opens only when there is a choice — two or more such targets, or a
printed "you may" (`min` 0, and `targets: []` declines); a single
forced change is made with no prompt, and a spell nothing could move
onto the object asks nothing at all. The `reason` text says which
object the target is changing to.

### `ack` (server → client) — ADR 0122 §6.4

Every action the server applies is acknowledged to the connection that
sent it, and to no other, so a client knows exactly which of its actions
landed and in which state. It covers all three apply paths: an ordinary
action, a table-settings change (`set_table_settings`, `set_undo_limit`)
and `undo`.

```json
{
  "v": 0,
  "kind": "ack",
  "id": "a4f7b8e1-2c3d-4e5f-9a2c-0d8e7f6a5b4c",
  "payload": { "seq": 43, "generation": 0 }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | string | yes | The `id` of the `action` frame being acknowledged. |
| `payload.seq` | uint64 | yes | The `seq` of the state the action produced: the snapshot that state was broadcast in. |
| `payload.generation` | uint64 | yes | That state's restore generation. |

- **Ordering.** On the originating connection the ack always arrives
  **after** the snapshot of the state it names. A client that sees the
  ack already holds that state, or a newer one: snapshots of other
  seats' actions may arrive between the two, and the ack's `seq` says
  which state was this action's.
- **Failure is unchanged.** A refused action gets its `error` frame
  with the action's `id` and never an ack.
- **The browser** handles the kind and records it in the protocol log,
  and does nothing else with it. A client that does not know the kind
  must ignore it (see [Frame envelope](#frame-envelope)); a browser
  holding a client from before this change logs it as an unknown kind
  until it reloads, which is harmless.

### `legal_moves_request` (client → server) — ADR 0122 §6.1

Asks for the bound seat's whole legal-move list. A snapshot's
`legal_moves` is capped at 48 moves and degrades past that
([`legal_moves_truncated`](#gameview-schema-summary) says when); this is
the list before the cap. It is also how a seat asks for one card's or
one prompt's moves past the enumerator's own caps (see
[The cut report](#the-cut-report-adr-0122-62)).

```json
{
  "v": 0,
  "kind": "legal_moves_request",
  "id": "5d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a",
  "payload": { "source": "9f3c…" }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload` | object | no | Empty, absent or `null` asks for the seat's whole list under the enumerator's ordinary caps. |
| `payload.source` | string (UUID) | no | One card's moves only: every move whose `source` is this instance, and every block on or by it. Lifts every count cap to 512. |
| `payload.choice` | string | no | One prompt's moves only: a pending choice's `id`, or `"cleanup_discard"` for the cleanup-step discard (CR 514.1). Lifts the caps like `source`. Wins over `source` when both are sent. |

- **The bound seat only.** The answer is always for the seat the
  connection is bound to — the seat `FilterViewFor` filters its frames
  for. No field names a seat, and a field that tries (`player`, `seat`)
  is ignored like any unknown field, so a connection cannot ask about
  another seat's moves.
- **Refusals**, each an `error` frame with the request's `id`:
  - `bad_request` from a connection with no seat: a spectator, or an
    admin with no `?player=`. An admin bound to a seat asks as that seat.
  - `no_decision` from a seat that owes no decision right now. A
    request for one card that has no moves is an empty answer, not a
    refusal.
  - `rate_limited` past **4 requests per second** per connection.
    Requests are answered one at a time, on the connection's own read
    loop, under the room's lock; a refused one is not counted.
  - `bad_request` for a `source` that is not a UUID or a `choice` that
    is neither a UUID nor `"cleanup_discard"`, and `bad_json` for a
    payload that is not an object.

### `legal_moves` (server → client) — ADR 0122 §6.1

The reply to a `legal_moves_request`, carrying its `id`.

```json
{
  "v": 0,
  "kind": "legal_moves",
  "id": "5d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a",
  "payload": {
    "seq": 42,
    "generation": 0,
    "moves": [ { "type": "cast_spell", "player": "…", "params": { … }, "kind": "cast", "label": "Cast Lightning Bolt targeting Bear", "source": "9f3c…" } ],
    "truncated": [ { "source": "9f3c…", "cap": "per_source", "omitted": 10 } ]
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.seq` | uint64 | yes | The seq of the state the moves describe, read under the same room lock as the moves. |
| `payload.generation` | uint64 | yes | That state's restore generation. |
| `payload.moves` | LegalMoveView[] | yes | The seat's enumeration with **no wire cap**, in the order `legal_moves` is projected from. With `source` or `choice`, that card's or prompt's moves with the caps lifted. Never `null`; `[]` when the named card has none. |
| `payload.truncated` | LegalCutView[] | no | The enumerator's cut report for what is listed: each count cap that still left moves out. Absent when nothing was cut. |

A client discards a reply whose `seq` or `generation` is not the window
it is deciding in, and asks again after the next snapshot: a reply can
name a state whose snapshot has not reached this connection yet. The
moves are the same enumeration the seat's own frame is projected from,
so they carry nothing that frame does not already give the seat.

#### The cut report (ADR 0122 §6.2)

The 48-move wire cap is not the only place a list is cut. The
enumerator (`server/internal/legal`) has count caps of its own, and each
reports what it left out as a **LegalCutView**:
`{ source?, choice?, cap, omitted, at_least? }`.

- `source` is the card being expanded; `choice` the prompt being
  answered (a pending choice's `id`, or `"cleanup_discard"`). A cut can
  carry both: a prompt a card raised.
- `cap` names the cap:

  | `cap` | What it bounds | Default |
  |---|---|---|
  | `per_source` | moves per card or prompt across targets, modes and cost choices; for a block, the groups per attacker | 12 |
  | `max_x` | the largest X tried | 20 |
  | `variable_counts` | how many counts a variable-count cost is offered ("sacrifice any number") | 3 |
  | `subset_scan` | candidate subsets a filtered card-set prompt tests | 64 per offered move |
  | `creature_types` | creature types offered for a choose-a-creature-type prompt | 5 |
  | `card_names` | card names suggested for a choose-a-card-name prompt | 5 |
  | `cost_payments` | ways one alternative cost's card component is offered | 3 |
  | `repeats` | payments of a repeatable optional cost (multikicker) | 3 |
  | `ceiling` | the 512 a `source` or `choice` request stops at | 512 |

- `omitted` is how many candidate moves the cap kept the enumerator
  from building — each legal as far as the cap site could tell, though
  an affordability check it never reached might still have refused
  some. Exact unless `at_least` is `true`, when it is a lower bound: the
  site stopped before it could count the rest without doing the work
  the cap saves. `omitted: 0` with `at_least` means the walk stopped
  with candidates untested and cannot say whether any were legal (a
  `subset_scan` cut).
- A `creature_types` cut counts the whole vocabulary the prompt did not
  offer, since the server accepts any creature type. A `source` or
  `choice` request lists the board's types first and then the rest of
  the vocabulary, about 300 entries, under the ceiling.

The report reaches the client in two places: the reply's `truncated`,
and per card in the digest as `legal_actions.sources[id].truncated`
(below), which carries every cut filed against a card the digest lists
— without `source`, which is the key. A prompt's cuts are in the reply
only.

**Two answers are open sets** and are marked rather than listed: see
`value` on LegalMoveView below.

The in-process bot seat does not use any of this. It enumerates in
process, never reads the wire, and keeps its own caps.

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
    "generation": 0,
    "game": { "id": "...", "state": "active", "seats": [ ... ], "turn": { ... }, "battlefield": { ... }, "stack": { ... }, "exile": { ... } }
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `payload.seq` | uint64 | yes | Per-room sequence number, monotonically **non-decreasing only within one `generation`** (#523, [ADR 0044](decisions/0044-surviving-a-deploy.md) decision 5 — see [Connection lifecycle (v0)](#connection-lifecycle-v0) below for the full rewind story). Clients use it, together with `generation`, to detect dropped or out-of-order frames. `seq` is bumped only on successful `action` dispatches (see `Room.Apply` in the server); the initial snapshot a joining client receives reuses the current (not-yet-bumped) value, so a new joiner during an action race may briefly see two consecutive snapshots with identical `seq` — both carrying the same state. Clients must treat snapshots idempotently: duplicate `seq` always means "same state, re-apply is a no-op". |
| `payload.generation` | uint64 | yes | This room's restore generation: 0 for a room that has never been rebuilt from disk, incremented by one every time a server restart restores it from its last written restore point. A change from the generation of the last snapshot a client rendered means the server rebuilt this room from an earlier point — accept and render the frame regardless of its `seq`, and discard whatever ordering state was being tracked for the old generation. |
| `payload.game` | object | yes | Complete `GameView`. See the schema below. |

The `id` field on a snapshot is always empty — snapshots are not correlated
to any particular client request.

#### GameView schema (summary)

The server's Go source at `server/internal/protocol/view.go` is the
canonical type definition. High-level shape:

- **GameView**: `{ id, state, seats[], battlefield, stack, exile, phased_out, turn, mulligans_open, starting_seat, opening_roll?, undo_limit, settings, monarch?, initiative?, day_night?, promises?, vote?, discard_pending?, legal_moves?, legal_moves_truncated?, legal_actions?, stack_items?, pending_triggers?, delayed_triggers?, split_second_active?, damage_cant_be_prevented?, exile_if_creatures_die?, damage_shields?, damage_multipliers?, damage_redirections?, graveyard_target_bans?, pending_choices?, log?, reveals?, loop_notice?, outcome? }` — `mulligans_open` is true between `Start` and the moment all seated, non-eliminated players have called `keep_hand` (S08). `starting_seat` is the seat index that took the first turn: the seat the winner of the opening roll chose (#1480, [ADR 0121](decisions/0121-animated-dice.md) §2). A client ignores it while `opening_roll` is present (it reads `0` then). `opening_roll` ([ADR 0121](decisions/0121-animated-dice.md) §3, omitempty) is the open opening roll, present only between a start that opens it and the winner's `choose_starting_player`: `{ rounds: [{ seats: number[], rolls: [{ seat, result, by? }] }], chooser? }`. Round 1 is every seat; each later round is the tied leaders of the one before. `rolls` are in the order rolled; `by` is present only when somebody other than the seat rolled the die (the host's seat for "Roll for everyone left", `-1` for the server admin from no seat). `chooser` is present once one leader remains. No hand is dealt while it is present, and only `roll_opening`, `host_roll_remaining`, `choose_starting_player`, `roll_table_die`, `concede`, `set_table_settings` and `set_undo_limit` are accepted; every other action is refused with `bad_request` ("the opening roll is not finished"). None of them mints an undo entry. Public and identical for every viewer. The lobby still starts tables with the automatic roll (the same window, played out at once, the winner going first), so a live table never shows it yet. It is used by the server to enforce the CR 103.8a turn-1 skip-draw rule (two-player games only — CR 103.8c has nobody skip in a larger multiplayer game), and surfaced so spectators / reconnects render the same first-player decision. The source-less, turn-0 `roll` entries in `log` are the durable public roll record. Pre-S13 snapshots decode as `0` (S13). `stack_items` is the announce-time metadata for every item on the stack (S13.1) — see StackItemView below. `pending_triggers` is the APNAP queue of triggered abilities waiting to drain (CR 603.3b). Since #1529 the whole queue waits while any trigger of the batch still has its `trigger_prompt`, `mode_pick` or `pick_target` open, so a targeted trigger joins its batch and is offered in the same `trigger_order` prompt; the wire shape is unchanged. `delayed_triggers` (S22, omitempty) is the queue of CR 603.7 delayed triggered abilities still owed — "at the beginning of the next end step, return that card to the battlefield" — as `{ id, controller, source?, label?, at, created_seq?, cards?, on? }`, where `at` is the step whose beginning fires it and `cards` are the instance IDs the effect acts on. `on` (#663, omitempty) is the EVENT condition of a "when you next cast an instant or sorcery spell this turn" trigger (CR 603.7b): such a trigger is owed on the next matching event rather than at a step, so it carries `on` and an empty `at`. Public information, so it is not redacted per viewer. `split_second_active` mirrors the engine's "no responses allowed" gate (CR 702.61). Since #1519 it is set by any spell that PRINTS split second (the keyword is stamped by the deck importer or declared by the catalog entry), not only by the sandbox `split_second` cast flag; while it is true, `castable_here` is false on every graveyard, exile and library-top card and `timing_closed` is present on every non-mana ability row, instant-speed ones included, because both stamps read the engine's timing functions, which refuse under split second. Mana ability rows are unaffected, and so are the foretell and turn-face-up special-action rows (CR 702.61b); suspend's row goes unavailable because its window is a cast window (CR 702.62c). `damage_cant_be_prevented` ([ADR 0107](decisions/0107-state-triggers-rebound-disturb-and-damage-prevention.md) §5, omitempty) lists the live "damage can't be prevented this turn" grants (CR 615.12) by their source's name, oldest first — Skullcrack, Stomp. While one is live, every prevention effect (a Fog, a shield, protection) prevents nothing. Battlefield statics that say the same (Leyline of Punishment) are not listed. Public and identical for every viewer. `exile_if_creatures_die` ([ADR 0108](decisions/0108-turn-scoped-effects-object-history-and-damage-shields.md) §1, omitempty) lists the live "if a creature would die this turn, exile it instead" effects over a live set by their source's name, oldest first — Flaying Tendrils, Malicious Eclipse ("a creature an opponent controls"). One pinned to a single creature is that permanent's chip instead: `CardView.exiled_if_it_dies?: [string]` names the effects that would exile it if it died now (Lava Coil, Disintegrate), and `CardView.cant_be_regenerated?: true` marks a permanent that can't be regenerated this turn (CR 701.19c; Incinerate, Whippoorwill). All three are public, survive the face-down redaction, and are identical for every viewer. `damage_multipliers` ([ADR 0108](decisions/0108-turn-scoped-effects-object-history-and-damage-shields.md) §3, omitempty) is one line per live "it deals double (triple) that damage instead" effect a resolved spell or ability made, oldest first, already worded for the banner: "Alice's sources deal double damage this turn — Insult", "Sources deal double damage to Bob and their permanents until Alice's next turn — Lightning, Army of One". Two Insults are two lines (×4). Battlefield statics that say the same (Furnace of Rath, Angrath's Marauders) are not listed. Public and identical for every viewer. `damage_redirections` ([ADR 0108](decisions/0108-turn-scoped-effects-object-history-and-damage-shields.md) §9, omitempty) is one line per live "that damage is dealt to <something> instead" effect a resolved spell or ability made, oldest first, already worded for the banner: "Damage to Alice from Goblin Guide is dealt to Beacon of Destiny instead, the next time — Beacon of Destiny", "Damage to Bob and their permanents from Lightning Bolt is dealt to Carol instead, the next 2 — Harm's Way". Battlefield statics that say the same (Pariah, Palisade Giant) are not listed. Public: the source and the destination are announced as the effect is made. Identical for every viewer. `graveyard_target_bans` ([ADR 0109](decisions/0109-rule-gates-land-types-mana-and-cost-components.md) §6, omitempty) is one line per live static that stops cards in graveyards being the targets of spells or abilities (CR 601.2c), the printed clause and the card that prints it: "Cards in graveyards can't be the targets of spells or abilities. — Ground Seal". It names the clause, not a verdict: Tomik's line covers only land cards and only his opponents' spells and abilities. Every legal target set already leaves the refused cards out; the graveyard viewer shows the lines as a banner. Public and identical for every viewer. `CardView.land_type_effects?: [{ types, in_addition?, loses_all?, loses_abilities?, gains?, until?, source? }]` ([ADR 0109](decisions/0109-rule-gates-land-types-mana-and-cost-components.md) §1 and §2, omitempty) lists the resolved effects changing a battlefield permanent's land types, oldest first: `types` are the land types the effect gives (always an array, empty for a loss), `in_addition` is "in addition to its other types" (CR 205.1b; absent is CR 305.7's replacement, which takes the old land types and the rules-text abilities away), `loses_all` is "loses all land types" (Ultima, Origin of Oblivion's blight), and with it `loses_abilities` is the same effect's "and abilities" and `gains` the printed texts of the abilities it gives (`["{T}: Add {C}."]`), `until` is the duration in the card's words ("until end of turn", "until Bob's next turn", absent for none) and `source` names the card — Tidal Warrior's "becomes an Island until end of turn", Navigator's Compass's added type. The effective `type_line` already shows the result; a static type change (Spreading Seas, Blood Moon, Lithoform Blight) is not listed. Public, survives the face-down redaction, identical for every viewer. `damage_shields` ([ADR 0108](decisions/0108-turn-scoped-effects-object-history-and-damage-shields.md) §7, omitempty) lists the live shields against a source ("prevent all damage a source of your choice would deal this turn" and the charged "prevent the next N damage … by a source of your choice"), oldest first, as `"<card> (<source>)"`, with `" — N left"` on a charged one — Pay No Heed (Goblin Guide), Healing Grace (Lightning Bolt) — 3 left. Public: the source is announced as the shield is made. `pending_choices` (S14, omitempty) is the async effect-driven decision queue — see PendingChoiceView below. `log` (S31, omitempty) is the public game log — see LogEvent below. `reveals` (S22, omitempty) is the broadcast reveal window — see RevealView below. `loop_notice` (#628, omitempty) is the CR 732 loop breaker: `{ source?, label, controller?, count }`, present once the engine has watched one triggered ability resolve 25 times in a turn with no player decision in between. It is an instruction to the CLIENT, not a change to the rules — priority still rotates and `pass_priority` is still accepted; what stops is AUTOMATIC passing, so the autopass toggle holds and a person has to ask for the next iteration. Table-wide and identical for every seat. Cleared by the next cast, activation, answered prompt or combat declaration, and at the turn boundary. `outcome` ([ADR 0057](decisions/0057-win-and-lose-by-effect.md), #749, omitempty) is the result of an ended game: `{ kind, winner?, winner_seat?, cause, source?, source_name? }`. `kind` is `"win"` or `"draw"`; `winner` / `winner_seat` name the winner of a win; `cause` is `"last_standing"` (every opponent has left, CR 104.2a), `"effect"` (an effect said this player wins, CR 104.2b — Felidar Sovereign, Laboratory Maniac, Thassa's Oracle) or `"all_lost"` (everyone left lost at once, CR 104.4a — a draw). `source` / `source_name` name the object whose effect won. An effect win ends the game with the other seats **still seated**, so a client must read the winner from here, not from "the one non-eliminated seat". Absent while the game is active and for a table an admin ended with no result; a view with `state: "ended"` and no `outcome` (older replays) falls back to the one seat left standing. Public and identical for every viewer. Once the game has ended, `undo` is refused for every caller except the server admin. `undo_limit` (S11) is the per-player per-turn undo budget; since ADR 0075 it is always present and mirrors `settings.undo_limit` (`-1` unlimited, `0` no undos). `settings` (S35, [ADR 0075](decisions/0075-table-settings-and-host-controls.md)) is the table's configuration, `{ undo_limit, undo_scope, starting_life, commander_damage, bot_pace, allow_spawn }`: `undo_limit` as above; `undo_scope` `"own"` (a seat undoes only its own entries) or `"host_any"` (the host may undo anyone's); `starting_life` (1–999, default 40; fixed once the game is active); `commander_damage` (1–99, default 21 — the damage from one commander that loses the game, read at every state-based action check, so lowering it mid-game can eliminate a player at the next check); `bot_pace` `"fast"` / `"normal"` / `"slow"`; `allow_spawn` (default false). Public and identical for every viewer, spectators included. Settings are not rolled back by `undo`. `day_night` ([ADR 0132](decisions/0132-day-and-night.md), omitempty) is the game's day/night designation (CR 731): `"day"` or `"night"`, absent while the game has neither. It belongs to the game, so it is public and identical for every viewer. `monarch` / `initiative` (S10) are the player ID holding that designation, empty when unassigned; `promises` (S10) is the "I owe you" tally keyed `"{from}->{to}"` with zero entries dropped; `vote` (S10) is the open council's-dilemma / politics vote, absent when none is open. `discard_pending` (S13.4) maps player IDs to the number of cards each still has to discard and drives the discard prompt. `legal_moves` (S31) is the viewer's own enumerated legal moves, filled per seat by `FilterViewFor` and absent from the unfiltered view. `legal_moves_truncated` ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §6.1, omitempty) is `true` when the 48-move wire cap dropped anything from the viewer's own `legal_moves`, and absent otherwise; own seat only, by the same construction. A seat that needs the rest sends a [`legal_moves_request`](#legal_moves_request-client--server--adr-0122-61). `legal_actions` ([ADR 0105](decisions/0105-legal-action-highlights.md), omitempty) is a per-card digest of the same enumeration, under the same own-seat rule — see LegalActionsView below.
- **PlayerView**: `{ id, name, seat, life, poison?, energy?, speed?, library, hand, graveyard, command, commander_damage, life_history, eliminated?, hand_kept?, mulligan_turn?, trigger_order_always_ask?, auto_answers?, undo_auto_answer?, mulligans_taken?, mana_pool?, is_bot?, bot_tier?, bot_deck?, is_agent?, agent_client?, citys_blessing?, playmat_url?, playmat_wash?, is_host?, emblems?, keywords?, life_total_locked?, cant_gain_life?, cant_play_lands?, cant_lose?, cant_win?, end_gates?, counter_shields? }` — `speed` (ADR 0138, #2122, omitempty) is the player's speed (CR 702.179): absent while they have none, then 1 to 4, and 4 is max speed (CR 702.179e). Public and identical for every viewer; the engine is its only writer (the start-your-engines state-based action and the inherent once-per-turn trigger), so there is no action that sets it. `mana_pool` (S15, omitempty) is the player's typed mana pool as an ordered list of single-character color strings (`"W"`, `"U"`, `"B"`, `"R"`, `"G"`, `"C"`) in tap order. Drives the `ManaPoolPips` row in the player header. Server-side empties at every step boundary (CR 106.4) so the wire surface stays small. Public to all viewers (matches paper Magic — floating mana sits visibly next to the player). `life_history` is a per-player rolling log of `LifeChangeView` entries (`{ delta, new_total, at, seq }`, RFC3339 timestamp), bounded server-side at 50 entries and never filtered (life is public). `seq` (#703) is a per-player counter starting at 1, stamped on every recorded change (a zero delta records nothing and consumes no `seq`) and monotonic across frames. It is the only field that identifies an entry, and it is what a client must key a transient life-change cue on: the log is trimmed from the FRONT, so the array index is not stable and the entry count stops growing at the cap, while `at` reaches the wire as RFC3339 **seconds**, so two changes in the same second are indistinguishable. `seq` keeps climbing after the log stops growing; it is derived from the newest surviving entry, so it rewinds with an undo and continues correctly across a snapshot restore. Entries from a snapshot taken before #703 decode as `0`. `eliminated` is omitted unless true; set when the player has left the game — a concession (S08), a state-based loss, or an effect that says they lose (ADR 0057). An effect WIN eliminates nobody: see `GameView.outcome`. `hand_kept` is true once the player has committed via `keep_hand`; `mulligan_turn` (#2237, CR 103.5, omitempty) is true on the one seat whose turn it is to keep or mulligan, and on no seat outside the mulligan window, while the opening roll is open, or once every seat has kept (a seat that conceded is skipped, so a departure moves it on) — public, identical for every viewer; `mulligans_taken` counts mulligans in the current opening-hand window. Both omitted when false/zero. All five fields added in S08. `trigger_order_always_ask` (#1530, omitempty) is true when the seat asked to be prompted for every trigger order; it is **private to its seat** (`FilterViewFor` clears it for every other viewer and spectators), and it is what the client compares with its local setting before re-sending `set_trigger_order_preference`. `auto_answers` ([ADR 0127](decisions/0127-answering-repeated-prompts-for-you.md) §3, omitempty) is the seat's standing answers, `[{ key, answer }]` sorted by `key`, with `answer` `"always"` or `"never"`; **private to its seat** on the same terms, and what the client compares with `gameplay.autoAnswers` before sending `set_auto_answers`. `undo_auto_answer` (ADR 0127 amendment 2026-10-07, omitempty) is the seq of this seat's `auto_answer` log line while that automatic answer is the **top undo entry**, so an `undo` from this seat takes it back; it is absent as soon as any other commit sits on top. It is room state, stamped on each capture like `is_host`, and **private to its seat**. The client greys its notice's Undo from exactly this. `is_bot`, `bot_tier` and `bot_deck` (S31, all omitempty) mark a bot seat, one driven by an in-process `aiseat` runner instead of a browser ([ADR 0033 §9](decisions/0033-ai-bot-seat.md), [docs/bot.md](bot.md)). `is_bot` is true only on such a seat. `bot_tier` is its policy tier: `"random"`, `"heuristic"`, `"assisted"` or `"strong"`. `bot_deck` is the curated deck ID it was seated with, such as `"izzet-aggro"`. It is empty when the bot was seated with a pasted decklist (`{tier, format, source}`). All three are public, unredacted, and absent on a human seat. The client renders the BOT chip, the robot avatar mark and the thinking pulse from them. They also ride the engine snapshot, and `bot_tier` is on the game's `seats` row as well, so a bot seat survives a restart ([docs/lobby.md](lobby.md)). `is_agent` and `agent_client` ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §7, S62, both omitempty) mark an **agent seat**: a guest seat played by an AI agent through an MCP client on someone's own machine, which declared itself in its join body (`agent`, [docs/lobby.md](lobby.md#post-gamesidjoin)). `is_agent` is true only on such a seat. `agent_client` is the client's declared name, normalised to at most 32 characters of `[a-z0-9._-]`, such as `"claude-code"` or `"codex"`, and `"unknown"` when the client named none. `citys_blessing` ([ADR 0096](decisions/0096-the-monarch-from-a-card-effect.md) amendment 2026-10-08, #2696, omitempty) is `true` on a seat that has the city's blessing (CR 702.131c): a player designation ascend gives and nothing takes away, so once present it stays for the rest of the game. Public and identical for every viewer, like the monarch. The agent fields: both are public, unredacted and identical for every viewer, seats, spectators and the admin alike: being an agent is a fact about the seat, like `is_bot`. A seat is a bot, an agent or neither, never two. **The badge never comes off**: the server sets it when the seat is claimed, in the same commit, and nothing clears it, not a reclaim, an undo, a restart or any admin route. An agent seat never hosts and cannot link Discord. The badge is a declaration by a cooperating client: a client that does not send `agent` is not marked, and the server cannot tell it from a person. Both keys ride the engine snapshot (`isAgent`, `agentClient`), so an agent seat survives a restart with its badge. `is_host` (ADR 0075 §2.1, omitempty) is true on exactly one seat, the table host, and absent everywhere else, including on every seat of a table with no host. The host may manage the table alongside the server admin. It is public and unredacted. It is not engine state: the room stamps it on each capture from the host the lobby designated, and it moves to the next human seat in turn order when the host concedes or loses. It is never true on a bot or agent seat. See [docs/lobby.md](lobby.md#the-table-host). `playmat_url` ([ADR 0128](decisions/0128-playmats.md), omitempty) is the same-origin path of the **active** playmat of the seat's signed-in owner (they keep up to three, ADR 0128 §11, and show one), `/playmats/<uuid>`, which the client draws behind that seat's battlefield. It is public and identical for every viewer, like the seat's name, and it is never a third-party URL: the server fetched and stored the image, so a browser contacts only this server (`GET /playmats/{id}` needs a session, [docs/lobby.md](lobby.md#playmats-adr-0128)). It is absent for a guest, a bot, an agent seat and a person with none, and it is not engine state: like `is_host`, the room stamps it on each capture from what the lobby set, so it is not in the snapshot or the undo stack, and a change made mid-game (a new image in the active slot, switching or removing the active mat, a fit, an admin removing it) reaches every viewer on the next state broadcast with no new message type. A client must load it only when it is exactly `/playmats/` and a lowercase uuid (`isPlaymatPath` in `client/src/lib/playmat.ts`); the `<img>` carries `?token=` as an avatar does. `playmat_wash` (ADR 0128 amendment, omitempty) is how strongly the owner darkens that playmat under the cards, a whole percentage from 30 to 90, chosen by the owner (`PATCH /me/playmats`) and the same for every viewer; it is present only beside `playmat_url`, stamped by the room the same way, and a client draws 58 when it is absent. `land_drops_per_turn` and `lands_played_this_turn` (#500, both always present) are the two halves of the CR 305.2 land-drop badge: how many lands this seat may play this turn and how many it already has. `land_drops_per_turn` is the EFFECTIVE allowance — a controlled permanent granting extra land plays, or a one-turn grant, is already summed in — so it is normally `1`. Since #500 the engine REFUSES a land play past the allowance, so a client should disable the hand's lands once `lands_played_this_turn >= land_drops_per_turn` rather than wait for the `bad_request`. Both are public. `max_hand_size` (S13.4, always present) is the seat's EFFECTIVE maximum hand size (CR 402.2): `-1` for no maximum, otherwise a number from `0` up, normally `7`. Since [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md) §3 (#2074) it is the CR 613.11 timestamp-order fold of every static that sets or changes it (Null Profusion's "is two", Jin-Gitaxias's "each opponent's … reduced by seven", Price of Knowledge's "players have no maximum") and the player's own rest-of-the-game grant; it never goes below `0`, so `-1` always means no maximum. The cleanup discard reads the same value for the active player only (CR 514.1). Public. The client shows a `HAND MAX n` or `NO HAND MAX` badge on a seat whose value is not `7`. `emblems` (#623, omitempty) is the seat's emblems (CR 114), in creation order, each `{ instance_id, label, text, level?, lines? }` — `label` is the board name ("Elspeth, Sun's Champion emblem") and `text` the emblem's printed ability, for the hover. The Ring ([ADR 0114](decisions/0114-the-ring-tempts-you.md) §2, §9, CR 701.54c) is an emblem too, labelled `"The Ring"`: its `level` is how many times the Ring has tempted the seat, its `text` is the lines it has gained so far (one per line), and `lines` is every line it can have, each `{ text, at }` with `at` the number of temptations that gains it, so a client can show the lines still to come. `level` and `lines` are absent on every other emblem. It is deliberately NOT a `ZoneView`: an emblem has no characteristics at all (CR 114.1), so there is no `CardView` to build and nothing for the card renderer, the image route or the targeting layer to do with one. Emblems are **public and unredacted** — an emblem sits face up in the command zone and any player may read it — so every seat's list reaches every viewer intact. Absent for a seat with none, which is nearly every seat. An emblem never appears in the `command` `ZoneView`: that zone is the cast surface and the commander pile, and an emblem is neither castable nor a card. Label and text are read from the catalog on every projection rather than stored, so a wording fix in a card file reaches a game already in progress. See [ADR 0064](decisions/0064-emblems.md). `keywords` (#1197, #1201, CR 702.11d / CR 702.16i, omitempty) is the bare engine tokens a SEAT has right now — `"hexproof"`, or a general `"protection from <quality>"` (`"protection from everything"` is the only quality a catalogued card grants a player today: Teferi's Protection, The One Ring, Leyline of Sanctity, Aegis of the Gods). Derived grants from a controlled permanent come first, then ones granted for a duration. EFFECTIVE, like `max_hand_size` and `land_drops_per_turn` — computed on every projection, not read off stored state. PUBLIC and unredacted, like `emblems`: protection and hexproof on a seat are facts about the board, and `legal_targets` already excludes a protected seat for a viewer who could otherwise see the refusal but not its reason. Absent for a seat with none, which is nearly every seat. Unlike `CardView.protection`, the wire does NOT parse the quality into a structured `ProtectionView` here — there is exactly one production quality today, so a second structured field for one value wasn't worth shipping. The client (`client/src/lib/playerKeywordBadges.ts`) title-cases the parsed quality itself rather than switching on today's two spellings, renders a badge on the seat tile beside the life total in the same visual language as a card's protection badge (#979, `KeywordBadgeRow.svelte`), and reuses the hexproof icon from `keywordIcons.ts`. `cant_play_lands` ([ADR 0109](decisions/0109-rule-gates-land-types-mana-and-cost-components.md) §4, CR 101.2, omitempty) is the clause that stops this seat playing any land from its hand right now ("Players can't play lands. — Territorial Dispute", "You can't play lands this turn — Turf Wound"), read from the same gate the play, the legal-move list and the land's own `cant_cast` read; a ban that names particular lands (City in a Bottle) or a zone other than the hand is on the land's `cant_cast` instead. Public, rendered as a NO LANDS badge. `cant_gain_life` ([ADR 0107](decisions/0107-state-triggers-rebound-disturb-and-damage-prevention.md) §5, CR 119.7, omitempty) is true while this seat can't gain life — a battlefield static (Leyline of Punishment, Erebos), a turn grant (Skullcrack) or the rest of the game (Screaming Nemesis). Effective and public, like `life_total_locked`. `life_total_locked` (#1200, CR 119.7 / CR 119.8, [ADR 0085](decisions/0085-life-total-cant-change.md), omitempty) is true while *"your life total can't change"* applies to this seat — Platinum Emperion's printed static, or a grant that lasts until this player's next turn (Teferi's Protection, Teferi's Reproach). EFFECTIVE and PUBLIC on the same terms as `keywords`: the derived half is not written anywhere on the engine's `Player`, so it is computed on every projection, and the lock changes what every player at the table may do — a Bolt that moves no life, a drain that drains nobody, a Phyrexian symbol this seat may not claim — so a viewer who can see the refusal but not its reason is strictly worse off. It is deliberately **not** another token inside `keywords`: that field is the bare engine ABILITY tokens a seat has and the client parses `protection from <quality>` out of it, while a life-total lock is not an ability the player has (ADR 0085 Decision 7). What it does NOT mean: damage is still dealt to the seat (it simply moves no life, so "whenever ~ is dealt damage" triggers and the CR 903.10a commander tally are unaffected), poison counters still land, and the player can still lose the game by every CR 104.3 route but their life total. The client renders it as a `LIFE` badge in the same `KeywordBadgeRow` language, appended after the keyword badges by `client/src/lib/playerKeywordBadges.ts`. `cant_lose`, `cant_win` and `end_gates` ([ADR 0057](decisions/0057-win-and-lose-by-effect.md) Decision 7, #749, all omitempty) are the "can't lose the game" / "can't win the game" gates on this seat (CR 104.3). `cant_lose` lists the causes that can't make the player lose right now — any of `"life"`, `"empty_draw"`, `"poison"`, `"commander_damage"`, `"effect"`; all five under a Platinum Angel. Concession is never in it: a player can always concede. `cant_win` is true when an effect can't make this player win (an opponent's Platinum Angel or Herald of Eternal Dawn, or their own Abyssal Persecutor); it does not stop them winning as the last player standing (CR 104.2a overrides it). `end_gates` names the sources, each `{ source?, source_name, cant_lose?, cant_win?, this_turn? }`, with `this_turn` on a gate a resolved spell granted until end of turn (Angel's Grace). EFFECTIVE and PUBLIC on the same terms as `keywords`: derived on every projection from the battlefield and resolved spells. A player behind a gate can sit at 0 or less life, 10 poison or 21 commander damage and stay in the game; the badge is how the table sees why. The client renders `CAN'T LOSE` / `CAN'T WIN` badges from these, with the sources in the tooltip. `counter_shields` ([ADR 0106](decisions/0106-five-small-seams-from-the-s50-rechecks.md) §4 decision 6, #1806, omitempty) is this seat's live "can't be countered" grants and unspent one-use promises, in the order they were made, each `{ source?, source_name, text, next_only? }`: `text` is the clause as printed ("Spells you control can't be countered this turn." from Veil of Summer, "The next creature spell you cast this turn can't be countered." from Insist), `source_name` the card it came from, and `next_only` marks a promise the seat's next matching spell will spend. A promise disappears from the list the moment a spell spends it (that spell's stack chip says it can't be countered from then on), and every entry disappears at cleanup. EFFECTIVE and PUBLIC on the same terms as `end_gates`: each came from a spell or ability the whole table watched resolve. A printed static on a permanent (Chimil, the Inner Sun) is not listed here; it is on the battlefield for everyone to read. The client renders a `NO COUNTER` badge from it, with the lines in the tooltip. `discord_id`, `discord_avatar_hash` and `display_name` (S12.5, all omitempty) are the Discord identity of a seat claimed through Discord sign-in. The client builds the avatar URL from the first two and prefers `display_name` over `name` as the seat label. **Since S34 sub-PR 4 they can change mid-game**: a seated player can link a Discord account to their seat, or move it to another one, through `GET /auth/discord/link` ([docs/lobby.md](lobby.md#get-authdiscordlink)). The change is applied through the room, so it arrives as an ordinary state broadcast to every viewer, and a client should read these fields from each frame rather than cache them per seat. No new message type.
- **ZoneView**: `{ kind, owner?, count, cards[] }` — `owner` omitted for shared zones
- **CardView**: `{ instance_id, name, owner, controller, scryfall_id?, is_token?, oracle_id?, type_line?, colors?, power?, toughness?, tapped?, counters?, is_commander?, battle_x?, battle_y?, attacking_target?, blocking_target?, auto?, target_mode?, mana_cost?, produced_mana?, mana_abilities?, abilities?, face_down?, face_down_kind?, face_visible?, phased_out? }` — **S16:** `power`, `toughness`, `type_line`, `colors`, and `abilities` carry the *effective* (post-CR-613-layer-resolution) characteristics, not printed. `power` retains its zero clamp for combat damage; `negative_power` (#825, omitted unless below zero) carries the signed effective power including counters for comparisons such as skulk. Use `negative_power` when present and `power` otherwise. Both fields are redacted for unknown cards. The wire keys are unchanged; only the resolution semantics are. Cards with no static ability affecting them produce identical wire output to pre-S16 (printed == effective). When an anthem is in play, the controlled creature's `power` / `toughness` arrive +1; when Mycosynth Lattice is in play, every permanent's `type_line` includes `"Artifact"`; when Lord of Atlantis is in play, other Merfolk creatures' `abilities` includes `"islandwalk"`. `colors` is the current `[]string` of `"W"`, `"U"`, `"B"`, `"R"`, and `"G"`, including layer-5 changes; absent means colorless, and a consumer must never infer color from `mana_cost` — a devoid card (CR 702.114a, #2152) prints coloured pips and has no `colors` in any zone, and carries `"devoid"` in `abilities`. It is redacted with other card characteristics. `abilities` (S16, omitempty) is a `[]string` of canonical keyword names ("flying", "first strike", "trample", "islandwalk", "fear", "intimidate", "shadow", "horsemanship", "skulk", "infect", "wither", etc.). Toxic (#748, [ADR 0056](decisions/0056-infect-wither-toxic.md), CR 702.164) is the one token that carries a number, `"toxic N"`, and it is CUMULATIVE: a creature may carry several entries (`["toxic 1", "toxic 1"]` for a Rat under Karumonix) and its toxic total is the sum of every N (CR 702.164b), so a client must not dedupe the list before summing. `"annihilator N"` (#2073, [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md) §2, CR 702.86) is numbered the same way and is cumulative too: each entry is one instance and one attack trigger, so an Eldrazi with a printed annihilator 4 under Eldrazi Conscription carries `["annihilator 4", "annihilator 2"]`. `"modular N"` (#2012, CR 702.43) is numbered and cumulative the same way. `"prowess"` (#706, CR 702.108b) is the other cumulative token: each entry is one instance and one trigger, so Ty Lee under Sokka carries `["flash", "prowess", "prowess"]`; `"evolve"`, `"undying"`, `"persist"` (#2075) and `"exalted"` (#2538, one entry per exalted counter too) and `"decayed"` (#2650, likewise) are cumulative the same way; the badge row may dedupe it for display — drives S18's keyword renderer; behavior of the keywords lands with S18 (combat sub-step rewrite). Non-protocol-bumping change: pre-S16 clients ignore the unknown `abilities` field harmlessly. `mana_cost` (S15, omitempty) is the card's printed cost string in Scryfall format (`"{2}{R}{R}"`, `"{X}{G}"`, `"{W/U}{W/U}"`); the parser accepts generic, colored (`{W|U|B|R|G|C}`), variable (`{X}`), hybrid two-color (`{W/U}`), hybrid generic (`{N/W}`), phyrexian (`{W/P}`), and snow (`{S}`) symbols. Drives the cost chip on the cast modal and the strict-mode gate. Empty / omitted ⇒ treated as costless (for cards Scryfall ingestion couldn't parse, which the cost gate emits as an `EventCostWarning` rather than rejecting). Redacted to empty for cards the viewer is not a `known_by` of. `produced_mana` (S15, omitempty) is the parsed Scryfall `produced_mana` field (e.g. `"{C}{C}"` for Sol Ring, `"{W|U|B|R|G}"` for Birds of Paradise) — informational; the canonical activation surface is `mana_abilities`. `mana_abilities` (S15, omitempty) is the per-ability projection the client renders into the right-click "tap for mana" menu: each entry is `{ "label": string, "produced": string, "cost_tap"?: bool, "cost_sacrifice"?: bool }`. Synthetic basic-land abilities are rendered into this list when the catalog has nothing registered (Forest → `[{"label": "Add {G}", "produced": "{G}", "cost_tap": true}]`). `scryfall_id` is stamped at deck-import time and lets the client resolve images via `GET /cards/{id}/image`. Omitted for placeholder cards (demo game seeded via `CMDCTRL_SEED_DEMO`). **[ADR 0078](decisions/0078-token-art.md)**: for a TOKEN, `scryfall_id` (when present) is a Scryfall **token printing** the server resolved from the token's name / type line / P/T / colors — an id chosen for its ART, not a card this player owns or could look up as a purchasable printing, and it may change between one process and the next if the underlying Scryfall dump refreshes (a token already on the battlefield keeps whatever id it was minted with). `is_token` (ADR 0078, omitempty) is `true` for a token (CR 111, the same printed-type-line test the CR 704.5d state-based action uses) and absent for a real card — public, so it survives the non-knower redaction exactly like `face_down_kind` and `phased_out` (`face_down_view_test.go`'s `redactedCardKeys` pins this). A pre-ADR-0078 client ignores the unknown key harmlessly. `oracle_id` (S14) is the Scryfall oracle-level identifier (stable across printings) and is the catalog lookup key — every printing of Lightning Bolt shares one `oracle_id`. `type_line` (e.g. "Legendary Creature — Human Wizard"), `power`, and `toughness` carry Scryfall data the client uses to filter creature-only UIs and label combat rows; all three omitted for placeholder cards. `battle_x` / `battle_y` are normalised positions in `[0, 1]` stamped by `set_battlefield_position` (S06+). **Sent for battlefield cards only, and for every one of them (#29)** — so an absent pair means "this card is not on the battlefield", never "this card is at the origin". They previously carried `omitempty` on a plain float and so vanished whenever the position was `(0, 0)`, which conflated a deliberate origin stamp with a permanent nobody had positioned. The origin is a routine value, not an exotic one: `set_battlefield_position` *clamps* rather than rejects, so every negative and NaN coordinate lands on exactly 0, and `x = 0` is "first in the row" for the client's within-row sort. Outside the battlefield the pair is omitted because it is meaningless there — zone exit clears it — and those zones are the bulk of every frame. Note this makes the wire honest rather than more expressive: `game.Card` carries no "positioned" flag, so "on the battlefield but never positioned" is not a state the server can distinguish from `(0, 0)`; such a card reports `(0, 0)`. Clients should keep an `?? 0` fallback — replays captured before this change omit the pair on battlefield cards too. `attacking_target` (player UUID) and `blocking_target` (attacker instance UUID) are set by `declare_attacker` / `declare_blocker` and cleared on zone exit and by `clear_combat`. Both omitted when not set. **#1706:** a creature that can block more than one attacker carries `block_capacity` (how many, when two or more — High Ground, Two-Headed Giant of Foriys, a monstrous Hundred-Handed One's 100) or `blocks_any_number: true` (Palace Guard), both omitted for the ordinary creature that blocks one; they are read off the effective characteristic and are public. When it blocks two or more, `blocking_targets` lists every attacker it blocks in declaration order (the first entry repeats `blocking_target`); it is omitted for an ordinary block, so a client reads `blocking_targets ?? [blocking_target]`. `type_line`, `power`, `toughness`, `attacking_target`, `blocking_target` added in S08. `attached_to` (S24, omitempty) is the CR 301.5c / CR 303.4 attachment relation for an Equipment or an Aura — a `TargetRefView` (`{ kind, id }`) naming the permanent (`kind: "card"`) or player (`kind: "player"`) this card is attached to. Omitted for every card attached to nothing. Shipped in ONE direction only: "what is attached to this creature" is derived client-side by partitioning the battlefield on this field, so the two directions cannot disagree. Not redacted — attachment is public battlefield state exactly like `attacking_target`. See [ADR 0036](decisions/0036-attachments.md). `restrictions` (S24, omitempty) is the CR 508.1c / 509.1b / 602.5 restriction set the engine computed for this permanent, as stable snake_case tokens: `"cant_attack"`, `"cant_block"`, `"cant_be_blocked"`, `"cant_activate"`, `"cant_activate_mana"`. Omitted for the permanent nothing is restricting, which is nearly all of them. Deliberately NOT folded into `abilities`: a restriction is not a keyword the permanent has, it is an effect something else has (Pacifism, Arrest, a Whispersilk Cloak), so it renders as a disabled control with a reason rather than as a keyword badge. The client READS it and derives nothing — who may attack is the server's decision and this is how it says so. Public battlefield state, redacted only for a card the viewer is not a `known_by` of (#95). See [ADR 0045](decisions/0045-combat-restrictions.md). `attack_target_restrictions` (#1794, [ADR 0106 §2](decisions/0106-five-small-seams-from-the-s50-rechecks.md), omitempty) is this creature's CR 508.1c restrictions on WHOM it may attack, each `{ player, planeswalkers?, source? }`: `player` is the seat it can't attack (its owner, read live, so a copy names its own owner), `planeswalkers` adds "or planeswalkers that player controls", and `source` names the card that imposes it — Xantcha, Sleeper Agent's "can't attack its owner or planeswalkers its owner controls". A row naming the creature's own controller is not sent, since no creature may attack its controller anyway (CR 506.2). The client draws the card's "CAN'T ATTACK <name>" chip from it and nothing else: which targets an attacker may pick is `legal_actions.sources[].attack_targets`, which leaves the owner out. Public, redacted like `restrictions`. `auto` (S14, omitempty) is true when the card's `oracle_id` is registered in the catalog (drives the gold-leaf AUTO badge). `target_mode` (S14, omitempty) is the announce-time target-prompt shape the client should show when casting: `"any"`, `"player"`, `"creature"`, `"permanent"`, `"stack_spell"`, `"stack_ability"`, `"stack_item"`, `"card_in_graveyard"`. The three stack values (#1211, CR 115.4) differ only in the sentence the picker writes — `"stack_spell"` a spell, `"stack_ability"` an activated or triggered ability, `"stack_item"` either — and all three are answered by an ordinary `{kind: "card", id}` ref: a picked ABILITY carries its STACK ITEM id (the id in `stack_items[i].id`), which for a spell item is already the card instance id, so no new ref kind arrives on the wire. Legality is `legal_targets` and never the mode. Empty ⇒ no prompt (cast fires immediately). Both fields redacted to zero for cards the viewer is not a `known_by` of. `target_cost_notes` (#746, omitempty) is a `[]string` of the printed clauses of the card's own cost modifiers whose price depends on its targets — Fireball's "This spell costs {1} more to cast for each target beyond the first", strive's "{1}{W} more" — stamped with the other cast clauses on cards in the viewer's own hand, command zone, castable graveyard, castable library top and (since #978) an exiled card the viewer's own grant opens. The X picker opens before targeting, and the auto-tap preview it reads is priced with no targets (so at the one-target price); the picker shows these clauses under its readout rather than a surcharge it cannot know yet ([ADR 0048](decisions/0048-cost-modification.md) addendum, open question 2). The cast itself is priced with its real targets. Absent for nearly every card; redacted with the other card-identity fields for a card the viewer is not a `known_by` of, and stripped with `legal_targets` and the other cast clauses from an opponent's revealed hand card. It ships on the snapshot rather than on the auto-tap preview response the ADR first named ([ADR 0048](decisions/0048-cost-modification.md) §15, amendment). `phyrexian_symbols` (#916, omitempty) is how many symbols in the card's cost carry CR 107.4's "or 2 life" option — 1 for Gitaxian Probe's `{U/P}`, 2 for Dismember's `{1}{B/P}{B/P}` — and is the ceiling on the `phyrexian_life` a `cast_spell` may claim. Shipped as a count rather than left to the client for the reason `activated_abilities[i].demands_x` is: a client re-deriving it would be a second parser of the mana-cost syntax. Stamped with the other cast clauses on cards in the viewer's own hand, command zone, castable graveyard, castable library top and (since #978) an exiled card the viewer's own grant opens, stripped from an opponent's revealed hand card, and redacted with `mana_cost` for a card the viewer is not a `known_by` of. `alternative_costs[i].phyrexian_symbols` answers the same question for an offer that replaces the printed cost, and `activated_abilities[i].phyrexian_symbols` for an activated ability's mana component.
- **TurnView**: `{ seq, number, active_seat, priority_holder, phase, step }` — `seq` is the 1-based identity of this turn and increments at every turn boundary; `number` remains the table-facing round shared by the seats in one rotation. Consumers that group per-turn state use `seq`; the UI continues to display `number`. `priority_holder` is the seat index (0-based) that currently holds priority within the step, OR `-1` (the `NoPriority` sentinel) during steps that don't grant priority. S13 lands the cursor with `priority_holder = -1` for `untap` and `cleanup` (CR 502.4 / 514.3); since #1501 it is also `-1` in `declare_blockers` while any defending seat is still declaring blockers (`block_pending_seats` non-empty), because CR 509.1 takes the declaration before anyone receives priority, and the active seat receives it when the last declaration completes; auto turn-based actions (auto-untap, auto-draw, auto-advance through cleanup) fire from the step-entry hook so the cursor never sits idle on a no-priority step. Added in S07; turn identity split in S41. Two omitempty fields carry CR 500.7 extra turns ([ADR 0059](decisions/0059-turn-machinery.md) Decision 11, #753): **`extra`** is true on a turn an effect created (Time Warp, Final Fortune) — `number` stays the round, so the board keeps "T3" and marks the turn "Extra turn" instead of numbering it (owner decision 1); **`extra_turns`** is the seat indices of the queued extra turns in the order they will be taken, next first (CR 500.7: the most recently created first), without the turns of players who have left (CR 800.4k: those never begin). Public: the effect that queued them was public. Three more carry the turn plan (ADR 0059 Decision 11, sub-PR 2b): **`phase_id`** identifies the phase in progress within this turn — the template's phases are 1 (beginning) through 5 (ending), and a phase an effect added (Relentless Assault's combat, Sphinx of the Second Sun's beginning phase) gets the next id, so two combats in one turn have two ids; **`phase_ordinal`** is the nth phase of its family this turn (2 in the second combat, 3 in the third main phase; precombat and postcombat main are one family, CR 505.1b); **`upcoming`** is the rest of the turn, `[{ step, phase_id }]` in order, including every added phase and step, and absent at cleanup. `first_strike_damage` appears in `upcoming` for every combat and is walked through when no combatant has first or double strike as it would begin (#717). All three are public — the turn's structure is public information — and omitted when zero or empty.
  Four omitempty fields join it during combat. **`block_decision_seats`** is the seat indices that still owe a block decision — since #1279, a DEFENDING seat whose block declaration is still open and that has a legal block to make; a seat leaves it once its declaration is complete, even with an untapped creature at home. **`block_pending_seats`** / **`blocks_declared_seats`** (#1279, [ADR 0045](decisions/0045-combat-restrictions.md) Decision 38) are where each defending seat's CR 509.1 declaration stands: still declaring, or finished (with or without blocks). A defending seat is in exactly one of the two; a seat nothing is attacking is in neither; both are absent outside `declare_blockers`. The engine completes a defender's declaration as the step begins when they have no legal block, when they send `finish_blocks`, when they `pass_priority` in the step (only possible for a declaring defender who holds priority, which since #1501 a live table never gives one — `priority_holder` is `-1` while anyone is declaring), and — for anyone still pending — as the step ends. Once every seat defending at that moment has finished, the declaration as a whole is over (#2021, ADR 0045 Decision 74): a seat that becomes a defending player afterwards (an attack reselected onto it, a creature put onto the battlefield attacking it) is listed in `blocks_declared_seats` and never asked to declare. This is the distinction an empty set of blockers cannot carry ("has not declared yet" vs "declared no blocks"), and what ninjutsu and "attacks and isn't blocked" wait on. Public, like `block_decision_seats`. **`attack_targets`** (S27) is what the ACTIVE player's creatures may attack right now, as `[{ kind: "player" | "planeswalker" | "battle", id, tax?, attack_limit? }]`: `id` is a seat id for a player and an instance id for a permanent, and the client must not re-derive "who protects which battle". **`tax`** (omitempty, [ADR 0080](decisions/0080-attack-taxes.md)) is the CR 508.1a price ONE creature pays to attack that target — `"{2}"` against a seat with Propaganda out, `"{2}{2}"` against one with Propaganda and Ghostly Prison, absent when attacking it is free. Server-priced, per creature; a declaration's real price is this once per attacking creature. **`attack_limit`** (omitempty, #1533, [ADR 0045](decisions/0045-combat-restrictions.md) Decision 46) is how many MORE creatures may be declared attacking that target this combat under a CR 508.1c count limit — Silent Arbiter's "no more than one creature can attack each combat", Crawlspace's "no more than two creatures can attack you each combat" — after the creatures already attacking: the smallest allowance among every limit that counts an attack on it. Absent when no limit counts that target (a Crawlspace protects its controller, not their planeswalkers or the other seats); `0` is sent, not dropped, when a limit is used up. Server-computed from the same count the declaration verbs refuse with, so `n` new attackers at that target are accepted exactly when `n <= attack_limit`; the client's "attack with all" picker caps its selection at it and never re-derives the rule.
- **StackItemView** (S13.1): `{ id, kind, controller, owner, source_card_id, label?, targets?, modes?, x_value?, distribution?, hold_priority?, split_second?, cant_be_countered?, damage_cant_be_prevented?, doubled_by?, doubled_by_name?, mana_spent?, colors_spent?, mana_spent_unknown? }` — announce-time metadata for one item on the stack. `kind` is `"spell"` (card lives in `stack` zone), `"activated"` or `"triggered"` (no card; source stays in its origin zone). `targets` is a list of `TargetRefView` slots: `{ kind: "player" | "card" | "self" | "none", id?: "<uuid>" }`. Resolution re-checks targets per CR 608.2b — if every targeted slot is illegal, the spell is "countered by game rules" and routes to its owner's graveyard. **#1553:** `cant_be_countered: true` marks a spell that can't be countered, either by its own printed "this spell can't be countered" (Supreme Verdict, Thrun) or because the mana that paid for it carried that rider (Cavern of Souls, Delighted Halfling, Boseiju, #1547). It is read from the engine's one counter gate, so it says exactly what a Counterspell would find; omitted when false and always absent on an ability item. Public. `damage_cant_be_prevented: true` ([ADR 0107](decisions/0107-state-triggers-rebound-disturb-and-damage-prevention.md) §5) marks a spell whose own text says its damage can't be prevented, under its condition as it stands now (Combust; Banefire with X of 5 or more). Read from the engine's prevention gate; omitted when false. Public. **#761:** `mana_spent` is how many mana paid for the spell and `colors_spent` the distinct COLOURS among them, in WUBRG order (colourless is not a colour, so it never appears there though it counts in `mana_spent`). Mana is spent face up, so both are public. `mana_spent_unknown: true` means the cast went through permissive mode or a strict-mode override: the engine never took the mana and has no record of what it was — render that as unknown, never as zero, because "nothing was spent" is a stronger claim and the one Vexing Bauble punishes. All three are absent on an ability item and on a copy of a spell (CR 707.10 — mana is not an object, so nothing was spent to cast the copy, and that zero is real). `PlayerView.commander_casts` (S13.1, omitempty) is the per-commander cast tax counter map keyed by commander UUID string. `CardView.damage_marked` (omitempty) drives the lethal-damage SBA. **#764:** a `TargetRefView` may additionally carry `slot` (the target CLAUSE it answered) and `mode` (the index into `modes` — the mode OCCURRENCE — whose clause list that is), both omitted at zero; and the item carries `mode_labels?: [string]`, the oracle bullet of each chosen mode in printed (resolution) order, CR 608.2c, and with repeats. The labels travel with the item because the caster's hand card is gone by the time the spell is on the stack, and `modes: [0, 2]` is not something a table can read.
- **ProtectionView** (`CardView.protection`, omitempty, #662): `{ printed, kind, value? }`, one per "protection from <quality>" on a permanent, parsed by the engine so no client or bot parses the token. `printed` is the quality as the card spells it. `kind` is `"color"`, `"card_type"`, `"subtype"`, `"everything"`, `"player"` (CR 702.16k; `value` is the chosen seat's id), `"mana_value_at_most"` (#2181, CR 702.16a / 202.3: `value` is the bound N as a decimal string, so Reaver Titan's "mana value 3 or less" is `{ printed: "mana value 3 or less", kind: "mana_value_at_most", value: "3" }`; a spell counts the X it was cast with, a token that is no copy is 0) or `"ring_bearer"` (#2145, CR 701.54: matches a source that is its controller's Ring-bearer when the check is made; no `value`). A client that meets a `kind` it does not know shows the `printed` text and nothing else.
- **Trigger doubling attribution** (#752, CR 603.2d): `StackItemView` may additionally carry `doubled_by?` (the public doubler permanent's UUID) and `doubled_by_name?` (its public card name). The fields are additive and omitted for ordinary spells, abilities, and undoubled triggers. A `trigger_prompt` or `pick_target` `PendingChoiceView` carries the same two optional fields, so the chooser can tell that the pending trigger is the additional trigger. The server projects these fields only from the engine's explicit trigger-doubler metadata; card identity in unrelated hidden zones remains subject to the normal per-viewer filter.
- **CardView.must_attack** (#1571, [ADR 0045](decisions/0045-combat-restrictions.md) Decision 51): `true` on a creature the active player owes an attack with right now — during `declare_attackers`, while the declaration could still obey a CR 508.1d requirement it does not (Zurgo, goad, Bident of Thassa), and this creature is one of the attacks that would answer it. Under a limit that lets only one of two such creatures attack, both are marked until one does. It clears the moment the active player's `pass_priority` would be accepted. Omitted otherwise. Public (it survives the redaction below): the requirement's sources are on the board. Server-computed from the same answer the legal-move enumerator gives; the client badges the creature and holds autopass, and never derives a requirement.
- **CardView.must_block** (#1597, [ADR 0045](decisions/0045-combat-restrictions.md) Decision 58): `true` on a creature a defending player owes a block with right now — during `declare_blockers`, while that player's declaration is pending and could still obey a CR 509.1c requirement it does not (Lure, Grand Melee, "blocks each combat if able", "must be blocked if able"), and this creature is one of the blocks the server wants: the same blocks `legal_moves` offers the seat as one `AlwaysLegal` "Block as required" move. It clears the moment that player's `pass_priority` would be accepted. Omitted otherwise. Public (it survives the redaction below): the requirement's sources are on the board. The client badges it and never derives a requirement; autopass already holds for a defender with a legal block (`block_decision_seats`).
- **CardView.goaded_by / goaders** (#1598, [ADR 0045](decisions/0045-combat-restrictions.md) Decision 53): a creature may be goaded by several players at once (CR 701.15c). `goaders` is every player whose goad is on it, oldest goad first, each ending as that player's next turn begins (CR 701.15a); omitted when it is not goaded. `goaded_by` keeps its S10 shape, one player ID, and is the most recent goader — always the last entry of `goaders` — so a client that reads only `goaded_by` still sees that the creature is goaded. Both are public and survive the face-down redaction below. Additive; a server before #1598 sends `goaded_by` alone.
- **CardView.known_by_you / face_down** (S13.5): per-viewer card visibility. `known_by_you` is true when the viewer is in the server-side KnownBy set for that card; the wire surfaces full printed characteristics in that case. When false, the wire keeps only instance_id / owner / controller / tapped / damage_marked / face_down / battlefield position / combat declarations (`attacking_target`, `attacking_target_kind`, `defending_player`, `blocking_target`, `blocking_targets`, `block_capacity`, `blocks_any_number`) / `goaded_by` / `goaders` / `must_attack` / `must_block` / `attached_to`, and zeroes everything read off the card itself — name, type line, Scryfall ID, P/T, counters, costs, faces, `auto`, `target_mode`, `mana_abilities`, `activated_abilities`, `restrictions`, `exile_play`, the hand-zone cast stamps, and the type-derived `summoning_sick` / `loyalty_activated` / `defense` / `protector_player` (#95; `face_down_view_test.go` pins this as an allowlist). The engine state stays observable, but identity doesn't leak. `face_down` is the visual flip flag (CR 406.3 / 708); the client draws a card back for a face-down card the viewer does not know, and the face for one it does. **ADR 0069** adds `face_down_kind` and `face_visible`. `face_down_kind` is WHY it is face down — `exiled` (CR 406.3, Necropotence), `foretold` (CR 702.143b), `hideaway` (CR 702.75a, [ADR 0091](decisions/0091-hideaway.md)), `permitted` (#1573 — Gonti, Night Minister and Outrageous Robbery's "exiles it face down; they may look at and play it", [ADR 0066's 2026-09-24 amendment](decisions/0066-granted-cast-and-play-permissions.md)), or one of the CR 708.2 permanent states `manifested` / `morphed` / `disguised` / `cloaked` — and it is PUBLIC, so it survives the redaction and labels the card back. `face_visible` is whether THIS viewer may look at the face: its controller for a CR 708.5 permanent, its owner for a foretold card (CR 702.143d), the holder of the cast permission over it for a `permitted` card (it arrives with the ordinary `exile_play`, `castable_here` and `cast_prices` for that seat alone; every other seat, the card's owner included, gets a card back and none of the three), the controller of the permanent that hid it for a `hideaway` card (and, by CR 406.3, anyone who has already looked — a hideaway land that changes hands hands the look over and the previous controller keeps theirs), nobody for a plain face-down exile (CR 406.3). A hidden SPELL that its linked ability has opened arrives with the ordinary `exile_play` stamp for the holder (`cost_override: "{0}"`, not `cast_only`, since hideaway says *play*); a hidden LAND is offered through the existing `may_cast` prompt and, on yes, played during the resolution as a land play (ADR 0091 decision 5). Since #1194 a spell CAST FACE DOWN (CR 708.4 — morph, megamorph, disguise) is one of these too while it is on the stack: it arrives as `face_down` with `face_down_kind: "morphed"` or `"disguised"`, the public 2/2 body, and `face_visible` true for its caster alone. It is stamped per-viewer beside `known_by_you`, equals `face_down && known_by_you`, and is what the client keys on to draw the real face plus a face-down badge. **The one exception to the redaction list above** is a face-down PERMANENT: CR 708.2 makes it a 2/2 colourless creature with no name, that body is public (an opponent has to see the 2/2 to block it), and so `type_line` (`"Creature"`), `power`, `toughness`, `colors`, `abilities`, `counters` and `summoning_sick` reach every viewer for it. When the effect that put it face down LISTED its characteristics (CR 708.2, #1270), that listed body is the public one: `type_line` is `"Artifact Creature — Cyberman"` for Cyber Conversion's victim and `"Land — Forest"` (no `power` / `toughness`) for Yedora's return, and `face_down_kind` is `"turned"` for both. No new field: the listing is read at layer 0, so the existing fields carry it. Everything that names the card underneath — `scryfall_id`, `mana_cost`, `faces`, `layout`, `auto`, `target_mode`, the ability lists — is still stripped from a non-knower, and most of it is never stamped at all because `CatalogKey` returns the empty key for a face-down permanent. Library cards have no knowers post-shuffle; opening-hand cards are known to their owner only; battlefield / stack / exile / graveyard / command-zone cards are public (all seated players are knowers). `ShuffleLibrary` clears every library card's KnownBy; `Mulligan` clears hand + library and re-grants the owner on the new opening hand. S14's `keepKnownInHandZone` view-filter path preserves revealed opponent-hand cards end-to-end: a seated viewer receives opponent hand zones containing only their `KnownByYou == true` entries (the `Count` stays accurate so the client can render the hidden remainder as face-down placeholders); the admin still sees a fully-hidden opponent hand via `hideZoneContents`, and a spectator (#1588) is a non-knower of anything a single seat does not know, so face-down cards read as backs to them too.
- **CardView.defending_player** (#1339, omitempty): on an attacking creature, the seat defending against its attack — the only seat whose creatures may block it (CR 802.4a): the player attacked, the controller of the planeswalker attacked, or the PROTECTOR of the battle attacked (CR 310.9d). Server-computed from the same function the block option generator and `declare_blocker(s)` read, so the client's block pickers offer exactly the blocks the server accepts; a block on any other attacker is refused with `illegal_block` / `not_defending`. Absent when the card is not attacking. A creature attacking a planeswalker or battle that has since left the battlefield keeps the seat that was defending it (#1364, CR 506.4c "it may be blocked", CR 802.2a), and its `attacking_target_kind` is absent because it attacks nothing — it may be blocked by that seat, and if unblocked it deals no damage. The same holds since #1376 when the planeswalker or battle is removed from combat WITHOUT leaving the battlefield (CR 506.4 — a control change, or phasing out): `attacking_target` is then the reserved id `00000000-0000-0000-0000-000000000506` (`game.AttackingNothing`), which names no seat or card, so the creature still reads as attacking but no arrow can resolve and no damage is dealt, and `defending_player` is the seat that was defending it before the removal. Public, like `attacking_target` it is derived from. A client reading a frame that predates the field falls back to `attacking_target` for a player attack (`defendingPlayerOf`, `client/src/lib/attackTargets.ts`).
- **CardView.class_level / solved** (S46, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md), both omitempty): the two DESIGNATIONS a permanent can have. `class_level` is a Class's CR 716.2 level — 1 for a Class nobody has levelled, up from there — and is present only for a Class on the battlefield, so the client renders its badge on presence rather than by parsing the type line. An **uncatalogued** Class carries it too: the level is engine state, not catalog state, exactly as an uncatalogued Saga still shows its lore counters. `solved` is a Case's CR 719.3 solved marker, absent rather than `false` for everything else. Both are PUBLIC — a level and a solved marker are visible across the table in paper — and both are cleared on the non-knower redaction with the other type-derived bits (`summoning_sick`, `loyalty_activated`, `defense`, `protector_player`), because a level says "Class" and a solved flag says "Case" as loudly as loyalty says "planeswalker"; CR 708.2 gives a face-down permanent no subtypes, so it is neither.

  **There is no field for "which printed abilities are active", and there does not need to be.** The designation gate is evaluated where an object becomes abilities, so an inactive ACTIVATED ability is already missing from `activated_abilities` — the same accessor the engine, the legal-move enumerator and the lobby read — and an inactive static or trigger has no per-ability representation on the wire to grey out. A station threshold needs nothing at all: charge counters ride the existing `counters` map, exactly as lore counters do.

- **CardView.token_text** (S46, [ADR 0083](decisions/0083-token-abilities.md), omitempty): a TOKEN's printed ability text, verbatim — `"When this token dies, you gain 1 life."`, with `\n` between printed lines. Absent for every printed card and for a vanilla token, which is what makes it additive: a pre-ADR-0083 client ignores the unknown key harmlessly. **It exists for the tokens ADR 0078's resolver still can't place.** [ADR 0078](decisions/0078-token-art.md) resolves most tokens to a Scryfall token printing at creation and stamps it onto `scryfall_id`; a token that resolves to nothing (no matching printing, or the dump the process loaded doesn't have one) keeps `scryfall_id` empty and renders through `Card.svelte`'s `.name-fallback` branch — the name on a grey rectangle. A mana or activated ability escapes that either way, because it reaches the player as a row in the right-click menu; a token's own TRIGGER ("when this token dies, create a 2/2 red Dragon") or STATIC has no control to hang text off, so without this field the words are nowhere on the client at all for a token that has no art. `EmblemView.text` is the same field for the same reason, one object over. **Derived on every read** from the token template's catalog entry (`game.TokenTextForCard`), never stored, so a wording fix reaches a game already in progress. Keyed on `CatalogKey` and not `CatalogAbilityKey`: a token silenced by Dress Down still shows what it PRINTS, exactly as the client keeps rendering a silenced card's oracle text. CR 707.2 rides along — a token that is a copy of a printed card has that card's oracle ID, answers empty here, and renders that card's own printing instead. Cleared on the non-knower redaction with the other catalog reads, though a token is known to every seat so the cell never fires in a real game.

- **CardView.ability_rows** (#2219, omitempty): `[{ kind, label }]`, the card's non-keyword abilities for the art tile's chips. `kind` is `"triggered"`, `"static"` or `"activated"`; the list is in that order. `label` is written by the server: a triggered row's `Key` and an activated row's `Label` (the names a stack item carries, ADR 0041 P9), a static slot's own label, each with the card's own name dropped from the front (`"Inferno Titan — 3 damage divided…"` reads `"3 damage divided…"`; a granted row keeps its grantor's name); a row the catalog gives no words is described by what it is (`"Replacement effect"`, `"Power/toughness effect"`). The client draws one chip per kind present with a count (⚡ ◆ ↻) after the keyword chips and lists the labels on hover or focus; **it never reads oracle text for them**. Built by `game.AbilityRowsOf` from the catalog through the same accessors the engine uses, so it is the card's **current** abilities: a CR 613.1f "loses all abilities" effect empties it, a layer-6 grant the layer pass let survive is in it (beside `granted_abilities`' text), an ADR 0071 designation gate decides a gated row, and a face-down permanent has none (CR 708.2). Off the battlefield it is what the card prints. Left out: keyword abilities (a trigger with a `Keyword`, the static synthesised from `PrintedKeywords`, and activated keyword abilities such as equip, crew and cycling — the keyword chips and the type line carry those) and mana abilities (ADR 0105's drop pip). An ability that only wears a keyword's frame around its own effect (`"Exhaust — {4}: Earthbend 4."`) is counted. Static rows come from `Static` and every other `CardDef` slot that holds a static ability of the object (cost modifiers, replacement effects, "can't gain life" and the rest); `game.TestEveryCardDefSlotIsClassifiedForAbilityRows` fails on a new slot until it is placed. Stamped only on the **battlefield and the hand**, the zones a tile is drawn in; absent for an uncatalogued card, which keeps `unimplemented`. Public on a card the viewer can see and cleared with the ability rows on the non-knower redaction. Additive: a client that ignores it shows no chips.
- **CardView.purpose** (ADR 0126 §6, omitempty): what the card does, as
  the catalog declares it. See
  [Declared purpose](#declared-purpose-what-a-spell-or-an-ability-does-adr-0126-6-2026-10-06).
  Hand, battlefield and stack; cleared with `ability_rows` on the
  non-knower redaction.

- **CardView.prepared** (#1328, [ADR 0090](decisions/0090-preparation-cards.md), omitempty): a permanent's CR 722.3a PREPARED designation. While it is set, the permanent's controller may cast the copy of its prepare spell that sits in exile — and that copy needs no new field: it is an ordinary card in `exile` whose `name` / `type_line` / `mana_cost` are the prepare spell's (face `1` active), carrying the `exile_play` stamp the client's exile button already reads, with `faces: [1]` and `player` naming the prepared permanent's controller. The stamp is DERIVED from the live permanent on every frame, so it disappears the moment the permanent leaves, is unprepared or its copy is cast; the copy itself leaves the `exile` zone view at the next state check. Public (the prepared creature and its copy are both visible across the table) and absent rather than `false` for every permanent that is not prepared; cleared on the non-knower redaction with `class_level` and `solved`. The client renders it through the same designation badge ("PREPARED").
- **CardView.doors** ([ADR 0103](decisions/0103-rooms.md), omitempty): a Room's two CR 709.5c designations, `{ "left": bool, "right": bool }` — `true` is an unlocked door. Present only for a face-up Room on the battlefield, so `false` inside it is a locked door and its absence means "not a Room on the battlefield". A locked half has no name, mana cost or rules text (CR 709.5), so a fully locked Room's `name` and `mana_cost` are empty: the client labels it from `faces`, which carries both halves. Public; cleared on the non-knower redaction with `class_level`. Each locked door's unlock is a `special_actions` row with `kind: "unlock"` and `door: "left"` or `"right"`, priced at that half's cost, with the server's own `available` timing answer.
- **CardView.fused** ([ADR 0103](decisions/0103-rooms.md), omitempty): for a split card with fuse in its owner's hand, the announce surface of the FUSED cast — `name` both halves, `mana_cost` both costs, and the same per-viewer target and price fields as a `faces` entry, the clauses being the left half's then the right half's. The client offers it as a third choice in the face picker and sends `fuse: true`. Stripped from a non-knower with `faces`. A split card out of play is its WHOLE card (CR 709.4): `name` is "Left // Right" and `mana_cost` both costs; `faces` carries each half.
- **CardView.harnessed** (#1321, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md) amendment 2026-09-23, omitempty): a permanent's CR 701.64 HARNESSED designation — the marker that switches its printed "∞ — [ability]" lines on (CR 702.186b). Unlike `class_level` / `solved` it carries no subtype gate: any permanent can print "Harness [this permanent]" (The Mind Stone is the first), so the field is read straight off the card once it is on the battlefield. Public (harnessed is visible across the table in paper) and absent rather than `false` for every permanent that is not harnessed; cleared on the non-knower redaction with `class_level`, `solved` and `prepared`. There is still no field for "which printed abilities are active" — an inactive `∞` trigger has no per-ability wire representation to grey out, exactly as an inactive Class or Case line has none.
- **CardView.ring_bearer** ([ADR 0114](decisions/0114-the-ring-tempts-you.md) §3, §9, #2076, omitempty): the permanent is its controller's Ring-bearer (CR 701.54b, 701.54e). Set when the Ring tempts that player and they choose it; it stops when another creature becomes their Ring-bearer, when another player gains control of it, or when it leaves the battlefield. A phased-out Ring-bearer keeps it (CR 702.26d). Public, read straight off the card, and kept on a face-down permanent and through the non-knower redaction: the choice was made in public and says nothing about which card it is. Absent rather than `false` otherwise.
- **CardView.monstrous** (#1700, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md) amendment 2026-09-28, omitempty): a permanent's CR 701.37b MONSTROUS designation, set by the monstrosity keyword action ("{cost}: Monstrosity N.") and kept until the permanent leaves the battlefield. It switches on the permanent's "as long as this creature is monstrous" lines and stops a second monstrosity from doing anything. Read straight off the card once it is on the battlefield, like `harnessed` — no card type owns monstrosity. Public and absent rather than `false` for a permanent that is not monstrous; cleared on the non-knower redaction with the other designations. The counters the action placed ride the ordinary `counters` map.
- **CardView.renowned** (#2049, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md) amendment 2026-10-09, omitempty): a permanent's CR 702.112b RENOWNED designation, set when one of its renown triggers resolves ("renown N": "when this creature deals combat damage to a player, if it isn't renowned, put N +1/+1 counters on it and it becomes renowned") and kept until the permanent leaves the battlefield. It switches on the permanent's "as long as this creature is renowned" lines and stops its renown from triggering again. Read straight off the card once it is on the battlefield, like `monstrous`. Public and absent rather than `false` for a permanent that is not renowned; cleared on the non-knower redaction with the other designations. The counters ride the ordinary `counters` map.
- **CardView.saddled** (#2695, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md) amendment 2026-10-08, omitempty): a Mount's CR 702.171 SADDLED designation, set by its saddle ability or by a spell or ability that says it becomes saddled, and kept only until the turn ends or the permanent leaves the battlefield. It switches on the Mount's "as long as it's saddled" lines and is what its "attacks while saddled" triggers read. Read straight off the card once it is on the battlefield, like `monstrous`. Public and absent rather than `false` for a permanent that is not saddled; cleared on the non-knower redaction with the other designations. Which creatures saddled it is engine state and is not sent.
- **`activated_abilities[i].saddle`** (#2695, omitempty): a Mount's saddle ability (CR 702.171a, "Saddle N") is crew's cost over OTHER untapped creatures, so it rides crew's fields. `crew_cost` is the saddle number, `crew_options` is the untapped creatures the controller controls with the Mount itself left out, the picks go back on `activate_ability` as `crew_ids`, and `saddle: true` tells the client to word the prompt "Saddle". It is activated at sorcery speed, which `cant_activate` reports like any other sorcery-speed ability. `saddle` is absent on every other ability, crew included.
- **CardView.suspected** (#2698, [ADR 0071](decisions/0071-designations-that-switch-abilities-on.md) amendment 2026-10-08, omitempty): a permanent's CR 701.60 SUSPECTED designation. A suspected creature has menace and can't block for as long as it is set, so the same card also carries `"menace"` in `abilities` and `"cant_block"` in `restrictions`; this field is the reason, for a badge. It stays on a permanent through a control change and is cleared when the permanent leaves the battlefield. Read straight off the card once it is on the battlefield, like `monstrous`. Public, absent rather than `false`, and, unlike the other designations, **kept on a face-down permanent**: it was given to the object in front of the table and says nothing about the hidden card. Additive; an older client ignores the key.
- **CardView.chosen_color / named_tribe** (#781, both omitempty): the answers a player gave to this permanent's "as this enters, choose a color" (CR 105.4) and "as this enters, choose a creature type" (CR 614.12) instructions — one uppercase colour letter (`"W"` / `"U"` / `"B"` / `"R"` / `"G"`) and one canonical creature type (`"Elf"`). Written by the engine onto `game.Card.ChosenColor` / `game.Card.NamedTribe` when the entry prompt is answered, and **cleared on every battlefield exit** (`game/zone.go`, `game/entry_tail.go`) — so a non-empty value means exactly "a permanent on the battlefield whose controller has answered", and `viewOfCard` copies both with no zone gate of its own. Absent for the overwhelming majority of cards, and absent in the window between the permanent entering and the prompt being answered; a consumer must read an absent value as the weaker outcome (no anthem, no mana), **never** as "any colour". **PUBLIC.** The choice is announced at the table and hidden from nobody, including the player who made it, and CR 607.2d is why it has to be on the card rather than only in the log: the linked ability says "creatures you control of the chosen color", and that set cannot be computed without the answer — an opponent could not tell which of their creatures a Heraldic Banner was pumping, and the controller had to remember what they named. Cleared on the non-knower redaction with the other type-derived bits (`class_level`, `solved`, `summoning_sick`, `loyalty_activated`), because `"Elf"` names Cavern of Souls and `"G"` names Coldsteel Heart as loudly as loyalty says "planeswalker"; CR 708.2 also leaves a face-down permanent with no such ability to have asked the question. The client renders both through one module (`client/src/lib/chosenValues.ts`) — a chip on the card tile's keyword row and a line in the hover panel's footer, with the linked clause in the tooltip. The engine event kinds (`EventColorChosen`, `EventCreatureTypeChosen`) are still absent from the public log; see #781.
- **CardView.chosen_option** (#1572, omitempty): the answer a player gave to this permanent's "as this enters, choose <A> or <B>" instruction (CR 614.12) when the options are words printed on the card — a Siege's anchor word (`"Khans"`, `"Dragons"`, `"Jeskai"`, `"Temur"`). Same lifecycle and the same PUBLIC / non-knower-redaction rules as `chosen_color` above: written onto `game.Card.ChosenOption` when the entry prompt (an ordinary `option_pick`, answered with `option_index`) is answered, cleared on every battlefield exit, absent in the window before the answer. It is the one chosen value that decides which of the permanent's OWN printed abilities exists ([ADR 0071](decisions/0071-designations-that-switch-abilities-on.md)'s designation gate): only the ability after the chosen anchor word appears in `activated_abilities` or acts at all, and before the answer neither does. The client does not render it yet.

- **GameView.phased_out / CardView.phased_out** (S46, [ADR 0084](decisions/0084-phasing.md), #1199): the CR 702.26 phased-out permanents. A shared, owner-less, public `ZoneView` beside `battlefield`, `stack` and `exile`, with `kind: "phased_out"`, and every card in it carrying `phased_out: true`.

  **A phased-out permanent is absent from `battlefield`, not flagged inside it**, and that is the whole point rather than a detail. CR 702.26b says a phased-out permanent "is treated as though it does not exist"; ADR 0084 makes that true of the engine by keeping it out of `Game.Battlefield.Cards`, and the wire follows so that the two surfaces that read the view — the bot's sixteen battlefield walks in `internal/aiseat`, and the positional stamps in `view.go` that index `view.Battlefield.Cards[i]` against `g.Battlefield.Cards[i]` — are correct without learning the rule.

  Phasing is **not** a zone change (CR 702.26d), so there is no `zone` log entry beside it and the permanent's `instance_id`, counters, damage, tapped state and `attached_to` are all unchanged across a phase cycle. `phased_out` on the card is redundant with the zone it arrived in and is carried anyway, so a client that one day renders phased-out permanents in place on the board has the bit without another wire change. It is **public** like `face_down_kind` and survives the non-knower redaction: everyone can see the board stop showing a permanent, and a viewer who cannot identify the card still has to be able to tell "phased out" from "died".

- **The engine event log is not on the wire.** S14 documented an `events` field on `GameView`, a rolling window of `EventView` entries (#139), but it was never added: no `GameView` in `server/internal/protocol/view.go` or `client/src/lib/protocol.ts` has ever carried one. The engine's log (`game.Game.Events`, kinds in `server/internal/game/events.go`) stays on the server. Game-state persistence saves it (`server/internal/game/snapshot.go`), and it reaches clients only through two table-visible projections built from it: `log` (LogEvent, below) and `reveals` (RevealView, below).

  Two engine event kinds about **prompts that changed hands when a player left the game** are deliberately server-side only and are **not** projected into `log`: `pending_choice_dropped` (#864) and `pending_choice_reassigned` (#902, CR 800.4g/h — `actor` the departed chooser, `target` the player who inherits, `source` the object, `label` the `PendingChoiceKind`). Both are diagnostics for a stalled table. What a player needs to see is already on the wire without them: the prompt itself, which the very next `snapshot` carries under its new `PendingChoiceView.chooser` — no new field, and no client change, because the picker already renders a prompt exactly when `chooser` is the viewer.
- **LogEvent** (S31, omitempty): `{ seq, kind, turn?, round?, step?, seat, target_seat?, card_id?, stack_item_id?, target?, amount?, old_zone?, new_zone?, combat?, combat_step?, unpaid?, sides?, results?, faces?, call?, wins?, seats?, choice?, label?, cause?, actor_is_host?, auto_answer_key?, text }` — one line of the **public game log** ([ADR 0033](decisions/0033-ai-bot-seat.md) §4). `turn` is the per-turn sequence identity carried forward from the latest step entry; `round` is present on `step` entries and is the table-facing number used in their rendered text. `GameView.log` is the last 200 table-visible events, **oldest first**, and it is a projection of the engine's own event log rather than a stored buffer — nothing on `game.Game` holds it, so it survives an undo, a snapshot restore and a deploy by riding `Game.Events`, which already does.

  `kind` is one of `step`, `cast`, `resolve`, `fizzle`, `counter`, `zone`, `draw`, `life`, `damage`, `attack`, `block`, `no_blocks`, `token`, `sacrifice`, `exert`, `eliminated`, `game_over`, `win_prevented`, `reveal`, `roll`, `flip`, `opening_roll`, `starting_player`, `table_roll`, `choose_color`, `choose_type`, `choose_player`, `choose_controller`, `choose_name`, `choose_option`, `control`, `special_action`, `activate_across`, `trigger`, `activate`, `cycle`, `counters`, `scry`, `surveil`, `discover`, `manifest_dread`, `saga_chapter`, `class_level`, `door_unlocked`, `door_locked`, `room_fully_unlocked`, `settings`, `spawn`, `transform`, `phase_out`, `phase_in`, `storm`, `turn_face_down`, `day_night`, `speed`, `citys_blessing`, `extra_turn`, `extra_turn_skipped`, `extra_phase`, `turn_ended`, `ring_tempted`, `auto_answer` — deliberately coarser than the engine's event kinds, because several engine events are one line to a reader and most engine events are no line at all. **Which** engine kinds get no line is no longer a matter of taste: `server/internal/protocol/log_event_kind_gate_test.go` (#984) reads every declared `game.EventKind` and fails unless the projection has an arm for it or the file lists it as a deliberate silence with a written reason.

  **Leaving and ending the game ([ADR 0057](decisions/0057-win-and-lose-by-effect.md) Decision 7, #749).** An `eliminated` entry carries `cause` — `"life"`, `"empty_draw"`, `"poison"`, `"commander_damage"`, `"effect"` or `"concede"` — and its text says why: "Alice lost the game (0 or less life)", "… (drew from an empty library)", "… (10 poison counters)", "… (commander damage)", "… (Pact of Negation)" with `card_id` the source of an effect loss, and "Alice conceded". `amount` is still 1 on a concession for older clients. A concession is ONE line: the engine's separate concede event is silent. A loss a "can't lose the game" effect stopped writes nothing (it would repeat on every check while a player sits at 0 life; `PlayerView.cant_lose` shows the state instead). A `game_over` entry closes a game that ended with a result: `seat` is the winner (`NoSeat` for a draw), `cause` is the outcome cause, `card_id` the winning source of an effect win — "Alice won the game (Felidar Sovereign)", "Alice won the game: every opponent has left", "The game is a draw". A `win_prevented` entry is an effect win a "can't win the game" gate stopped: `seat` would have won, `card_id` is the winning source and `target` the gate's source — "Alice would have won the game (Laboratory Maniac), but can't (Platinum Angel)". Once per prevented win.

  **An ability's `resolve` / `fizzle` (#1257).** For a SPELL, `card_id` is the spell and there is no `label`. For a triggered or activated ABILITY, which has no card on the stack, `card_id` is the ability's **source** and `label` the stack item's label — the same string the stack overlay printed for it ("Grapeshot — storm", "Sacrifice a creature: scry 1"). `text` names the ability by the label, prefixed with the source's name when the label does not already start with it: "Grapeshot — storm resolved", "Viscera Seer — Sacrifice a creature: scry 1 resolved", "Prodigal Pyromancer — {T}: deal 1 damage to any target was countered by game rules (no legal targets)". A copy of an ability (CR 707.10) carries its original's source and label, so it reads the same. A `label` on a `resolve` entry is therefore how a client tells an ability from a spell. **Redaction keys on the source's knowers, not on a dropped name**: a viewer who may not identify the source gets `card_id` and `label` both cleared and reads "an ability resolved". The source's id goes with the label because a source in a hidden zone (a hand, a library) has an instance id the zone filter never hands a non-knower, and the knower test has to be direct because a face-down permanent's name is empty on every view (CR 708.2), so the ordinary name-drop test sees nothing to drop while the label names the card outright. A source the view cannot account for at all — a token or copy that has ceased to exist, both public — keeps its line. Before #1257 these entries carried neither field and every ability rendered as "a card resolved".

  **Triggers and activations ([ADR 0119](decisions/0119-a-stack-you-can-follow.md) §5, S60).** Two kinds tell what goes on the stack as it goes there, not only when it resolves. Both are named and redacted exactly as an ability's `resolve` line is (above): `card_id` is the ability's **source**, `label` the stack item's label, and a viewer who may not identify the source gets both cleared.

  - `trigger` — a triggered ability triggered (CR 603.2) and was queued to go on the stack the next time a player would receive priority (CR 603.3): "Alice's trigger: Mulldrifter — draw two cards". `seat` is the ability's controller. The line is written as the ability triggers, which is the frame in which it reaches the stack except while its controller is ordering several (CR 603.3b). A harvested trigger and one announced by hand read the same. Consecutive identical lines (same controller, source and label, in the same step) collapse into one, with `amount` the count and " ×N" on the text; `amount` is absent on a trigger that happened once. Redacted: "Alice's ability triggered". The engine also emits its trigger event as a breadcrumb on every activation; that one names its source as the event's card and gets no line, so an activation is never told as a trigger. Because the discriminator is the engine event's existing shape, a restore point written before this change projects its triggers too.
  - `activate` — a player activated an ability that uses the stack (CR 602.2): "Alice activated Prodigal Sorcerer — {T}: deal 1 damage to any target". `seat` is the activator. Redacted: "Alice activated an ability". **Mana abilities get no line** (CR 605.3b: they do not use the stack, and every land tap would be one). An activation another line already tells gets none either: a cycling's `cycle` line stands for its activation (the projection skips the activation that follows a cycle of the same card), and an activation of another player's permanent is `activate_across`. An ability activated by hand through the sandbox's free-form `activate_ability` (no `ability_index`) emits no engine activation event, so it has no `activate` line; its `resolve` line still tells it.

  The text names an ability by its label, prefixed with the source's name unless the label already starts with the name or with a legendary's short name ("Tatyova — gain 1 life and draw a card", not "Tatyova, Benthic Druid — Tatyova — …"). This applies to the `resolve` and `fizzle` lines too.

  **Chosen values (#984).** `choose_color` (CR 105.4), `choose_type` and `choose_player` (CR 614.12) are the answers a player gives out loud to a "choose a ..." prompt — Coldsteel Heart's colour, Cavern of Souls' tribe, True-Name Nemesis' player. All three carry `card_id` (the card the answer was given for). `choice` carries the VALUE on the first two — the colour **letter** (`"G"`), the canonical creature type (`"Elf"`) — and is absent on `choose_player`, whose answer is a seat and therefore rides `target_seat` like every other player in the log. `text` renders it as "P1 chose green for Coldsteel Heart". `choice` is **redacted with the card's name**: the answer identifies the card as loudly as the name does (which is why `CardView.chosen_color` / `named_tribe` are stripped for a non-knower, #781), so a viewer who may not identify the card gets "P1 chose a color for a card" and no `choice` at all. **#1572:** `choose_option` is the same line for an "as this enters, choose <A> or <B>" anchor word (a Siege): `card_id` is the permanent, `choice` the word (`"Temur"`), redacted with the card's name, rendered "P1 chose Temur for Frostcliff Siege". **ADR 0102:** `choose_controller` is the answer to an `entry_controller` prompt — "enters under the control of an opponent of your choice". `card_id` is the permanent and the chosen seat rides `target_seat`, as on `choose_player`; `text` renders it "P1 chose P2 to control Captive Audience".

  **The Ring tempts you ([ADR 0114](decisions/0114-the-ring-tempts-you.md) §9, #2076).** A `ring_tempted` entry is one temptation (CR 701.54): `seat` is the player, `amount` how many times the Ring has now tempted them, `card_id` the creature chosen as their Ring-bearer (absent when they controlled none), and `cause` is `"forced"` when it was their only creature and the server chose it for them. `text` renders "The Ring tempts P1 (2) — P1 chooses Nazgûl as their Ring-bearer", "… — P1's only creature, Nazgûl, becomes their Ring-bearer", or "… — P1 controls no creature". The creature's name is redacted like any other `card_id`.

  **The six narrated silences (#1021).** Writing the log's deliberate silences down (#984) showed six of them to be gaps rather than decisions, and each is now a line. They are eight `kind`s for six decisions, because scry is not surveil and a Saga chapter is not a Class level — a client tones and filters by `kind`.

  - `control` — a permanent changed controller (CR 613.1b). `seat` is the player who GAINED control and `target_seat` the one who lost it, which is one sentence for a gain, an exchange (CR 701.12) and a duration expiring: "P1 gained control of Grizzly Bears from P2".
  - `special_action` — a CR 116.2 special action: foretell, suspend, plot, turning face up. `label` is the action as the card prints it ("Foretell {2}"). The card's zone move says only that a card left a hand for exile; this says which action it was.
  - `activate_across` — a player activated the "Any player may activate this ability" row of a permanent another player controls (CR 602.2, [ADR 0106 §1](decisions/0106-five-small-seams-from-the-s50-rechecks.md), #1793): "Bob activated Alice's Xantcha, Sleeper Agent". `actor` is the activator, `card_id` the permanent and `target_seat` its controller. Every other activation is an `activate` entry (below).
  - `auto_answer` — the server answered a prompt with its chooser's standing answer ([ADR 0127](decisions/0127-answering-repeated-prompts-for-you.md) §6): "Bob paid {1} for Rhystic Study (automatic)", "Bob didn't pay for Rhystic Study (automatic)", "Alice answered Yes to Consecrated Sphinx — draw two cards (automatic)". `seat` is the chooser and `card_id` the card that asked. `call` is the answer — `"pay"`, `"dont_pay"`, `"yes"` or `"no"` — and is public, as the answer is at a paper table; so is "(automatic)", because CR 732.1a asks that the table understand each player's shortcut. `label` is the cost paid on a `pay`, or the optional trigger's stack label on a `yes` / `no` about a trigger, and is redacted with the card's name: a viewer who may not identify the card reads "Bob paid for a card (automatic)". `auto_answer_key` is the key of the rule that answered, on the **chooser's own view only** (`FilterViewFor` clears it for everyone else); the chooser's notice uses it for "Ask me next time".
  - `cycle` — a cycling (CR 702.29b). It **replaces** the `zone` entry for the discard that paid the cost, the way a `sacrifice` entry replaces the zone move it causes, and keeps that entry's `seq`.
  - `counters` — the count of one counter kind on one card changed (CR 122). `label` is the kind (`"+1/+1"`), `amount` the count **after** the change — the engine's event carries no delta, so a placement and a removal are the same line with a different number, and `amount` ≤ 0 means the last one came off. **Loyalty and lore counters get no line**: a planeswalker's loyalty moves on every activation and every point of damage, both of which are already entries, and a lore counter's advance is the `saga_chapter` line below. The rule is `counterKindIsNarrated` in `log.go`.

    Counters on a **player** (poison, energy, experience, rad) are a different engine event — `player_counter_placed`, added by [ADR 0056](decisions/0056-infect-wither-toxic.md) Decision 5 — and get **no log entry yet**. It is a separate kind from `counter_placed` because the card-counter trigger helpers compare the event's target against card IDs, and it carries the **signed delta that landed** rather than the new total (the total is on `PlayerView.counters`). It is also the event the layer listener bumps on, which is what keeps a "corrupted" static reading an opponent's poison count from going stale. Its line — a `poison` entry reading "Alice got 3 poison counters (7/10)" — is [ADR 0056](decisions/0056-infect-wither-toxic.md) Decision 6 and lands with the client PR; until then `PlayerView.poison` carries the current total to every viewer on every frame, and the silence is recorded in `silentEventKinds` (`log_event_kind_gate_test.go`) with that reason.

    Counters on players also go through the **CR 614 replacement window** now, on the same `RepEventCounter` kind as counters on permanents, with `CounterPlayer` set instead of `CounterTarget`. That is what makes Vorinclex halve an opponent's poison and Lae'zel's "or on yourself" clause work. No action or view shape changed.
  - `scry` / `surveil` — a finished scry (CR 701.22) or surveil (CR 701.25). These carry **no `card_id` for anyone**, not even the spell that scried: the cards are hidden at both ends, and an entry with no card reference cannot leak one. `amount` is the public half — how many cards the table watched go to the bottom of the library, or into the graveyard. It is **not** the size of the scry ("scry 2"), which the engine's event does not carry.
  - `discover` — a finished discover (CR 701.57b, [ADR 0099](decisions/0099-discover.md)). `amount` is the N and `card_id` the discovered card, which the walk exiled face up; `card_id` is absent when nothing was found. The line comes when the card is settled: right after the `cast` line of the discovered spell, or once the card is in a hand (on "Put it into your hand", or on the pass that closes a "Cast it free" window).
  - `manifest_dread` — a player manifested dread (CR 701.62a, [ADR 0082](decisions/0082-casting-face-down-and-turning-face-up.md) amendment 2026-10-07). Carries no `card_id` for anyone: the manifested card is face down and known only to its controller, and the card put into the graveyard already has its `zone` line. Not emitted when the library showed no cards.
  - `saga_chapter` / `class_level` — a Saga reached a chapter (CR 714.2b) or a Class became a level (CR 716.2); `amount` is the chapter or the level.
  - `door_unlocked` / `door_locked` — a Room's door was unlocked (CR 709.5c: the unlock special action, an instruction, or the Room entering with the door of the half that was cast) or locked (CR 709.5g). `label` is the door's name. `room_fully_unlocked` — a Room got its second unlocked door (CR 709.5i); its own line by owner decision (ADR 0103). A fully locked Room has no name, so these lines name it by its two halves.
  - `settings` — the host or the admin changed a table setting ([ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.3, S35). `label` is the setting's key (`undo_limit`, `undo_scope`, `starting_life`, `commander_damage`, `bot_pace`, `allow_spawn`) and `choice` its **new** value as text — an int in decimal (`"-1"` is unlimited), a bool as `"true"`/`"false"`, an enum as its string. One line per field that actually moved; a no-op patch says nothing. `seat` is the host who changed it, or `-1` for the server admin, which the rendered text reads as "The admin". The only kind that is about the rules the game is being played under rather than about the game, which is exactly why it is written down: a budget that quietly halved mid-game is what this line prevents. It carries **no card**, so none of the card-identity redaction below applies to it.
  - `transform` — a permanent was turned over to its other face (CR 701.27a, S46, [ADR 0079](decisions/0079-transforming-a-permanent.md)). The entry's card name is the face it turned **into**, because `viewOfCard` reads the active face; `label` is the name of the face it turned **from**, which is the only place that name survives. A transform is **not** a zone change (CR 712.18 — the permanent doesn't become a new object), so there is no `zone` entry beside it and this is the only line the table gets. It is narrated rather than silent, unlike the other board-state changes, because a card physically turning over is announced out loud and a reader scrolling back wants to know when it happened.
  - `turn_face_down` — a permanent that was face up on the battlefield was turned face down (CR 708.2a, S46, [ADR 0082's 2026-09-23 amendment](decisions/0082-casting-face-down-and-turning-face-up.md), #1209). `card_id` is the permanent and `target` the object that did it (Ixidron, Cyber Conversion). Narrated for `phase_out`'s reason: turning face down is neither a zone change nor a transform (CR 701.27b says so in as many words), so no other line says it, and the board silently stops showing a card the table could read a second ago. The entry **names the source and not the permanent**, and that is not a redaction: a CR 708.2 object has no name for ANY viewer, its controller included (`CardView.Name` is the effective characteristic's, and the controller gets the art through `scryfall_id` / `face_visible` instead), so the line reads "Ixidron turned a card face down" for the whole table. The reverse direction is the next entry, and only for an effect's turn; the CR 116.2g special action's own `special_action` entry already carries that one.
  - `turn_face_up` — a face-down permanent was turned face up by an **effect** (CR 708.8 / 701.40b, [ADR 0082's second 2026-10-07 amendment](decisions/0082-casting-face-down-and-turning-face-up.md), #2590: Hauntwoods Shrieker, Zimone, Staff Room). `card_id` is the permanent, which is public again by the time the line is written (turning it up makes every seat a knower), and `target` the object that did it. The text reads "Zimone, Mystery Unraveler turned Grizzly Bears face up". The special action has no line of this kind.
  - `extra_turn` — an effect gave a player an extra turn (CR 500.7, [ADR 0059](decisions/0059-turn-machinery.md) Decision 11, #753). `seat` is the player who will take it and `card_id` the card whose effect created it; one entry per turn, so Time Stretch writes two. Narrated because it changes the turn structure: the next player to act is no longer the next seat. The turn itself begins with the ordinary `step` spine; `TurnView.extra` marks it.
  - `extra_turn_skipped` — a queued extra turn was skipped instead of beginning (CR 614.10, [ADR 0059](decisions/0059-turn-machinery.md) amendment, #2529; Trouble in Pairs). `seat` is the player who would have taken it and `card_id` the card whose effect created it. The text reads "Alice skips their extra turn (Time Warp)". Narrated because the `extra_turn` line promised a turn the turn bar will not move to.
  - `extra_phase` — an effect added phases or a step to the current turn (CR 500.8 / 500.9, [ADR 0059](decisions/0059-turn-machinery.md) Decision 11, #753). `seat` is the active player, whose turn gets them, `card_id` the card whose effect added them, `amount` how many, and `label` what was added: the phase kinds in the order they will occur, comma-separated (`"combat,main"`), or `"step:<step>"` for a single step (`"step:end"`, Y'shtola Rhul). The text reads "Alice's turn gets an additional combat phase and an additional main phase (Relentless Assault)". Narrated for `extra_turn`'s reason: a player about to pass priority out of a main phase has to know another combat is coming. The added phases themselves begin with the ordinary `step` spine; `TurnView.upcoming` shows them before they do.
  - `turn_ended` — an effect ended the turn (CR 724.1, [ADR 0059](decisions/0059-turn-machinery.md) amendment 2026-10-05, #2165): Sundial of the Infinite, Time Stop. `seat` is the active player, whose turn it was, and `card_id` the card whose effect ended it. The text reads "Alice's turn ends (Sundial of the Infinite)". Written once, as the process begins, so the `zone` lines for the stack being exiled follow it and read as part of it; the cleanup step it skips to then begins with the ordinary `step` line. A turn that ends normally has no such line.
  - `citys_blessing` — a player got the city's blessing (CR 702.131, [ADR 0096](decisions/0096-the-monarch-from-a-card-effect.md) amendment 2026-10-08, #2696). `seat` is the player and `card_id` the ascend permanent or spell whose check gave it. Narrated because the blessing arrives with no spell or ability of its own, so the marker beside the player's name would otherwise appear unexplained. One line per player per game.
  - `day_night` — the game became day or night (CR 731.1, [ADR 0132](decisions/0132-day-and-night.md), #2561). `label` is the new designation, `"day"` or `"night"`; there is no card or seat on the entry. Narrated because the untap-step check (CR 502.2) changes it with no spell or ability behind it, and a daybound permanent arriving does the same, so without a line every werewolf would turn over with nothing saying why. A game that gains its first designation gets the line too; only a flip from one to the other is what "whenever day becomes night or night becomes day" waits for.
  - `speed` — a player's speed changed (CR 702.179, [ADR 0138](decisions/0138-speed.md), #2122). `seat` is the player and `amount` the new speed, 1 to 4; the text is "Alice's speed is now 2", or "Alice has max speed" at 4. There is no card on the entry. Narrated because the start-your-engines state-based action gives a player speed 1 with nothing on the stack, and because reaching max speed switches abilities on across that player's board.
  - `phase_out` / `phase_in` — a permanent phased out or in (CR 702.26, S46, [ADR 0084](decisions/0084-phasing.md), #1199). Narrated for `transform`'s reason and one that is stronger here: phasing is **not** a zone change (CR 702.26d), so there is no `zone` entry beside it, and the board simply stops showing the permanent — which is indistinguishable from a permanent that died unless the log says which. Teferi's Protection removes a whole board this way, and a reader scrolling back has no other way to learn that it came back rather than never left.

  **`no_blocks` (#1279, #1500).** A defending player completed their CR 509.1 block declaration (`game.EventBlockersDeclared`) with zero blockers. The blocks a defender DID make are already `block` entries above, so this line is narrated only when `amount` on the underlying engine event is 0 — a defender who blocked with N creatures gets no second line about the same declaration, because the N `block` entries already said it. `seat` is the defender; the entry carries no `card_id` and no `target`, and `text` reads "P2 declares no blockers". It is public — the same sentence for every viewer, spectators included — because being asked to block and choosing not to is a fact the whole table watches happen, the same as an attack declaration.

  **`exert` ([ADR 0130](decisions/0130-exert.md) §5).** A player exerted a permanent (`game.EventExert`, CR 701.43a). `seat` is the player who exerted it and `card_id` the permanent; `text` reads "P1 exerted Oketra's Avenger". An exert as it attacks is paid as the attack declaration locks in, so its line comes just before that declaration's `attack` lines. The skipped untap it causes needs no line of its own: the card's `no_untap.next` already names the exerting player's untap step.

  **`storm` (S46, [ADR 0086](decisions/0086-storm-and-the-turns-cast-order.md), #1238).** A storm trigger resolved and settled on its count (CR 702.40a — in the pinned August 2026 edition storm is **702.40**; 702.39 is provoke). `card_id` is the storm spell, `amount` is how many *other* spells were cast before it this turn, by any player, and `text` reads "Grapeshot — storm count 3". The count is how many copies are about to be created, and an entry is emitted at zero too ("storm count 0" is the turn's first spell), so a reader can tell a trigger that found nothing to copy from one that never fired. `amount` is **not** redacted with the card name, unlike `counters` / `saga_chapter` / `class_level`: it is a fact about the turn's casts, which every seat watched, not a value read off the permanent — a viewer who may not identify the card reads "a card — storm count 3". This entry is the only line the table gets for the copies: a copy of a spell is *created*, not cast (CR 707.10), so it emits no engine event and produces no `cast` line, and the trigger's own `resolve` entry names the ability ("Grapeshot — storm resolved", #1257) but not the count. Since #1255 a storm spell COUNTERED in response to its own trigger is still copied, from last-known information (CR 608.2h), so a `counter` line for the spell can be followed by this entry and its copies.

  **`spawn` (S35, [ADR 0075](decisions/0075-table-settings-and-host-controls.md) §2.4).** The host or the admin put cards or tokens onto a live table from nowhere. `label` is the card or token name, `amount` the count, `new_zone` where they went and `target_seat` the seat whose zone it was; there is no `card_id`, because one entry covers every copy the request made. `text` reads "Luke (host) spawned 2 × Treasure onto Ana's battlefield". This line is the condition the feature ships under: a spawned Treasure is indistinguishable from an earned one on the board, so if the log does not say where it came from, nothing does.

    Two things about it are not like the other kinds. **A spawn into a hand or a library names the zone and NOT the card** ("Luke (host) spawned a card into Ana's hand"), and that redaction happens at BUILD time rather than per viewer — the entry carries no card reference for the knower predicate to key on, so the name is dropped for everyone, the spawner included. And **`actor_is_host` is stamped by the room, not by the projection**: the host is a property of `ws.Room`, not of the engine, so `protocol.StampHostOnLog` runs on the same pass that sets `PlayerView.is_host`. It is absent on a spawn by somebody who is not the host — the dev route (`POST /games/{id}/dev/spawn`) still lets anyone at a preview table spawn — and absent on an admin spawn, which has no seat at all and renders as "The admin spawned …".

  `label` is **redacted with the card's name**, exactly as `choice` is and for the same reason: "Foretell {2}" prices the card as loudly as an `alternative_costs` entry does, which `redactCardForViewer` already strips from a face-down card. So does the `amount` of a `counters`, `saga_chapter` or `class_level` entry, because `redactCardForViewer` strips `counters` from a non-knower too — a viewer who may not identify the card reads "a card's counters changed" and gets neither the kind nor the number.

  **Players are seat indices, not UUIDs.** `seat` is the responsible player (`-1` when there is none — an SBA life loss, a spell resolving with nobody to credit) and `target_seat` is the player acted on. `card_id` and `target` are card instance IDs; exactly one of `target_seat` / `target` is set on an entry that has a target at all. `step` rides only on `step` entries: every entry after one belongs to that step until the next, and a consumer that wants the step on every line carries it forward. All of this is wire-cost discipline — the struct repeats 200 times on every snapshot frame.

  `text` is the rendered line ("Aang cast Lightning Bolt", "Turn 13 (round 7) — Katara · precombat main": a step line counts turns and names the round, as the agent board header does, #2790) and is what a panel prints.

  **Stack item.** `stack_item_id` (ADR 0119 §3, omitempty) is on every `resolve`, `fizzle` and `counter` entry, for spells and abilities alike: the ID of the stack item that left the stack. For a spell it equals its card's ID; an ability's item has an ID of its own that no other field carries (an ability's `resolve` and `fizzle` `card_id` is its source permanent, and a countered ability's `target` is empty). It lets a client match the entry to the item that departed between two frames. It is additive, rides a new omitempty engine field, `game.Event.ResolvedStackItemID` (`events[].resolved_stack_item_id` in the snapshot; additive under v7, so no schema or protocol version change; `events[].stack_item_id` is untouched because stamping it would change frozen fixture values); an entry from before this field decodes without it. It is not redacted: the stack is public and an item ID names nothing the stack view does not.

  **Unpaid casts ([ADR 0118](decisions/0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) §2, #2188).** `unpaid` (omitempty) marks a `cast` entry whose `cast_spell` carried `force_cast: true`: the client's "Cast anyway (don't pay)" row, or the dock's Cast anyway after an `insufficient_mana` refusal. No mana was spent and the pool was left alone; life for Phyrexian symbols and every additional cost were still paid. Its `text` appends " without paying its mana cost" after the zone: "Alice cast Craw Wurm without paying its mana cost", "Alice cast Craw Wurm from exile without paying its mana cost". A permissive cast (no `strict`) is **not** marked, because that seat may be paying on paper. The flag is public, because the pool and the cost are public, so it is never redacted; the card's name is redacted per viewer as on any `cast` entry, and a face-down spell reads "Alice cast a card without paying its mana cost". The flag rides the engine's `game.Event.Unpaid` (`events[].unpaid` in the snapshot file), an additive field of schema 7.

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

- **Special actions** (CR 116.2 — foretell #658, suspend #659, turn face up #1194, [ADR 0062](decisions/0062-abilities-and-special-actions-from-the-hand.md) Decision 4 and [ADR 0082](decisions/0082-casting-face-down-and-turning-face-up.md) decision 9): a card in the viewer's OWN hand — and, since #1194, a FACE-DOWN PERMANENT the viewer controls — carries `special_actions`, an array of `{ kind, label, cost?, charged_cost?, available? }`. One row per CR 116.2 special action the card offers — `"Foretell {2}"`, `"Suspend 1—{R}"`, `"Plot {3}{U}"` — and the client fires the `special_action` verb straight from the row: there are no targets, no modes and no cost picker, because neither kind has a choice to make. `available` is the SERVER's per-kind timing answer for this frame, so a client greys the row rather than re-deriving a rule it would get backwards (foretell stays legal under split second, CR 702.61b; suspend does not, CR 702.62c). An unavailable row is greyed, never hidden: a player has to be able to see that the card has the keyword. Like `zone_abilities`, this is not public, and unlike it this one stays HAND-ONLY — foretell and suspend are announced out of a hand and nowhere else — `"Suspend 4—{U}"` names Ancestral Vision as loudly as its mana cost would — so the per-viewer redaction strips it with the rest of the cost surface. A kind the engine cannot carry out is never projected. The battlefield row is `turn_face_up` and nothing else: it is DERIVED from the face-down kind and the card underneath rather than declared by any card (CR 702.37b's morph cost for a morphed or disguised permanent, CR 701.40b's mana cost for a manifested or cloaked CREATURE card, and no row at all otherwise), so every face-up permanent on the board carries an empty list. **Since #1391 the top card of the viewer's OWN library carries the rows too**, when a permanent they control grants a special action there (Fblthp, Lost on the Range's plot). The rows sit on the projected card, so the owner sees them only while they can see it, and every other seat's frame drops them even when the top card is revealed — they are the owner's to take. A card may offer the same kind twice there (its own plot cost and its mana cost), and the client sends the row's `cost` back as the verb's `cost` param to say which. The client shows them as pills beside the library pile.

- **Sacrifice costs of N permanents** (#747, [ADR 0020](decisions/0020-activated-abilities.md) and [ADR 0021](decisions/0021-additional-costs.md) addenda): a sacrifice cost clause is shipped as `sacrifice_options` — a `LegalTargetsView` `{ cards[], min, max }` — in three places: `activated_abilities[i].sacrifice_options` (with `sacrifice_label`), `mana_abilities[i].sacrifice_options` (with `sacrifice_label`) and a hand card's `additional_cost.sacrifice_options` (with `additional_cost.label`). There is no separate count field. **`min` and `max` are the count and are always equal**: "Sacrifice a creature" is `1 / 1`, "Sacrifice two artifacts" is `2 / 2`, "Sacrifice five Treasures" is `5 / 5`. The answer rides back as `sacrifice_ids` on `activate_ability`, `activate_mana_ability` or `cast_spell`, and must name exactly `max` distinct permanents from `cards` (order does not matter). Anything else — too few, too many, one ID twice, a permanent the viewer does not control, one that does not match the clause, or the source itself when the cost also sacrifices the source — is refused with nothing paid. The N permanents leave the battlefield as one simultaneous event, so a "whenever a creature dies" watcher paid in with them sees every death (CR 603.10a). `cards` lists only permanents the viewer controls, from the non-targeting candidate walk (hexproof and shroud do not narrow it), without the source when the cost also sacrifices the source, and **in payment order**: tokens first, then lower mana value, then the ability's own source, then board order. That is the order the legal-move enumerator takes its one payment from (a sacrifice-N move names the first `max` of it), and the client's "Choose for me" button fills the picker with the first `max` too. A list shorter than `min` means the cost cannot be paid right now.

  **#1213: a count the ACTIVATOR announces.** `min` and `max` are no longer always equal. Two printed clauses vary:

  - **"Sacrifice one or more artifacts"** (Radiant Lotus) ships `min: 1` with `max: 0` — `LegalTargetsView`'s "no ceiling". Name any number at or above `min`; the board is the only bound.
  - **"Sacrifice X Treasures"** (Grim Hireling) ships `count_from_x: true`, and the count is the `x_value` the same message announces. Naming a different number of permanents is refused, not made cheap. Such an ability sets `demands_x: true` even when its mana component has no `{X}` at all — read the flag, never the cost string — and the announced X buys PERMANENTS rather than generic mana, so the mana owed is the same at every X.

  A fixed clause is byte-for-byte what it was: `min == max == N`. The bot's enumerator offers at most three counts for a variable clause ([docs/bot.md](bot.md)), which is policy rather than a rule; a client may name any number the bounds admit. Set-level restrictions ("with different names") still have no wire shape.

- **Return-a-permanent-to-hand costs** (#1213, [ADR 0073 amendment](decisions/0073-optional-additional-costs-and-the-cast-gate.md)): "Return a Forest you control to its owner's hand" (Quirion Ranger), "Return an artifact you control to its owner's hand" (Master Transmuter), "{1}, Return a land you control to its owner's hand" (Meloku the Clouded Mirror). `activated_abilities[i].return_label` is the clause as printed and `return_options` is a `LegalTargetsView` `{ cards[], min, max }` whose `min` and `max` are both the clause's count (1 for every printed card) and whose `cards` are the permanents that could pay right now, in the same payment order `sacrifice_options` uses — so the client reuses one picker and one "Choose for me". The answer rides back as `return_ids` on `activate_ability`: exactly `max` distinct permanents from `cards`. A TAPPED permanent is a legal pick (returning is not tapping) and so is the ability's own source when the clause admits it (Master Transmuter may return herself). An absent or empty list means the cost cannot be paid, and the engine refuses the activation (CR 118.3). The permanent goes to its OWNER's hand, not the activator's, and — being a cost — the move cannot pause: a commander returned this way takes its owner's hand rather than opening a CR 903.9 prompt (CR 601.2h / 602.2b). **#1227:** the clause may be a COMBAT one — ninjutsu's "Return an unblocked attacker you control to hand" (CR 702.49a) — in which case `cards` is empty outside the declare-blockers step and everything after it, and changes as blockers are declared. Nothing about the wire changes; a client that greys the row on an option list shorter than `min` (as both of this client's ability menus now do) is showing the player the same refusal the engine would give.
- **Delve** (CR 702.66, [ADR 0100](decisions/0100-delve-either-or-and-variable-sacrifice-costs.md) sub-PR 1): a card with delve carries `delve: { options: { cards[] }, max }` on its cast surface. `options.cards` are the caster's OWN graveyard cards that may be exiled — never the spell itself — in the server's payment order (lands first, then cards with no graveyard cast surface, then the rest, oldest first), which is also the order "Choose for me" fills from. `max` is the budget for the default announcement (this zone, X = 0, nothing tapped) and is a hint only; the picker's cap is the auto-tap preview's `delve_budget`, priced for the announcement as it stands. The picks ride `cast_spell` as `delve_ids`?: `["<uuid>", ...]`: zero up to the budget, distinct, each paying `{1}` of the generic mana in the total cost — after the cost modifiers and after the convoke taps, with X at the announced value, and never a coloured symbol (a "spend mana as though any colour" grant does not make one delvable). Exiling nothing is always legal. A list longer than the budget, a card named twice, a card not in the caster's graveyard, a card also named to `alt_cost_ids`, or `delve_ids` on a card without delve is refused as `bad_request` with nothing exiled and nothing paid. The cards are exiled at CR 601.2h with the spell already on the stack, so a "leaves your graveyard" trigger goes above it; a delved commander is asked about the command zone BEFORE anything is paid (#1397). The stack item's paid-cost record keeps the exiled objects (`PaidCost.Delved`), and the permanent the spell becomes keeps them as the cards "exiled with it" (CR 607.2q, ADR 0100 sub-PR 2) — Murktide Regent's counters, Soulflayer's keywords, Ethereal Forager's return — for as long as each is still in exile as that object. Delve does not change the spell's mana value. Since sub-PR 2, `delve` also appears on every card its viewer could cast while they control a permanent that says "Spells you cast have delve" (Teval, Arbiter of Virtue), with no change to the shape.
- **Either/or additional costs** (CR 601.2b, 601.2f–h, [ADR 0100](decisions/0100-delve-either-or-and-variable-sacrifice-costs.md) sub-PR 3): "As an additional cost to cast this spell, sacrifice an artifact or discard a card" (Demand Answers) ships as `additional_cost: { label, branches: [...] }`. Each entry of `branches` is an `AdditionalCostView` of its own — `discard_cards`, `sacrifice_options`, `mana_cost` ("{5}"), `pay_life` (a fixed "pay 3 life"), `blight` with `blight_options` — plus its `key` and `payable`: whether the viewer could pay it right now (enough other cards in hand, a permanent matching the clause, life of at least `pay_life` per CR 119.4, a creature to blight). Absent `payable` means the branch cannot be taken (CR 118.3). Mana is never asked, because CR 601.2g lets the caster tap afterwards. A cost with branches carries no components of its own. The caster names the branch on `cast_spell` as `cost_branch`?: `int`, an index into `branches`. It is **required** on a card with branches and refused (`bad_request`) on any other card, out of range, or for a branch that is not payable. It is never defaulted to 0. The chosen branch is then paid exactly like a mandatory cost: its cards ride `discard_ids` / `sacrifice_ids` / `blight_ids`, a payload for the other branch is refused, its mana joins the total at CR 601.2f next to a kicker's (so a cost modifier sees it), and its life is paid with the spell on the stack. A card whose every branch is unpayable is not castable (CR 601.2h: `castable_here` is false out of a graveyard or exile, and the legal move list offers no cast from hand). The stack item's paid-cost record keeps the branch (`PaidCost.CostBranch`, index + 1) and every card the additional cost discarded (`PaidCost.Discarded`, for Grab the Prize's "if the discarded card wasn't a land card"). A CR 707.10 copy keeps both. **Reveal and behold branches** (ADR 0100 amendment 2026-10-07): a branch that reveals a card from the viewer's hand ("reveal an Elf card from your hand or pay {3}") or beholds ("behold a Merfolk": a Merfolk you control, or reveal one from hand) ships `reveal: true` (plus `behold: true` for behold) and `reveal_options`, the cards that could pay it, hand cards first and then, for behold, permanents; present-and-empty means the branch is not payable. The one pick rides `cast_spell` as `reveal_ids` (exactly one id, refused when the announced branch has no such component or the card is not among the options). Paying shows a hand card to the table as an ordinary `reveal_cards` event and moves nothing.
- **Variable sacrifice costs on a cast** (CR 601.2b, 601.2f–h, 107.3a, [ADR 0100](decisions/0100-delve-either-or-and-variable-sacrifice-costs.md) sub-PR 4): "As an additional cost to cast this spell, sacrifice any number of creatures" (Vicious Betrayal), "you may sacrifice any number of creatures" (Torgaar, Famine Incarnate) and "you may sacrifice one or more creatures" (Plumb the Forbidden) all ship `additional_cost.sacrifice_options` with `min` 0 and `max` 0: an open count from zero, zero included, bounded only by `cards`. "Sacrifice X lands" (Devastating Summons) ships `count_from_x: true` and no `demands_x`: the number of permanents picked IS the X, sent as `x_value` on the same `cast_spell`, and the X prompt does not open. No new field — the existing `sacrifice_options` carries both shapes, and a fixed clause is unchanged (`min == max == N`). The whole `sacrifice_ids` list pays the variable clause: the engine holds a cast with one to no other sacrifice (effects.Register refuses the rest), so the width question a flat list raises has one answer. The count is recorded as `PaidCost.Sacrificed`, and a per-sacrifice discount ("This spell costs {2} less to cast for each creature sacrificed this way") reads it at CR 601.2f, so the auto-tap preview prices the cast from `?sacrifice_ids=` exactly as `cast_spell` charges it. Either shape is payable with an empty board, so it never makes a card uncastable. Bots are offered zero (when the floor allows it) and up to three positive counts, smallest first; for "sacrifice X … destroy X target …" the targets fix X and so the payment ([docs/bot.md](bot.md)).
- **Waterbend costs that are not a spell's** (#1310 / #1311, CR 701.67a, [ADR 0020 amendment 2026-09-23](decisions/0020-activated-abilities.md), [ADR 0073 amendment 2026-09-23](decisions/0073-optional-additional-costs-and-the-cast-gate.md)): "Waterbend [cost]" means pay the cost, and for each generic mana in it you may tap an untapped artifact or creature you control instead. Two surfaces carry it, both in the `TapCostView` shape a hand card's convoke / waterbend already ships as `tap_cost` (`{ key: "waterbend", label, options: { cards[] }, max, demands_x? }`), so one client picker serves all three. **An activated ability** — "Waterbend {8}: Transform Aang", Katara's "Waterbend {X}" — carries `activated_abilities[i].waterbend`: `options.cards` are the viewer's untapped artifacts and creatures that could pay right now (from the walk the engine validates against; the ability's own source is among them unless its cost also prints `{T}`, and a hexproof permanent is too, because a cost does not target), and `max` is how many — the waterbend's generic, capped by what the priced cost still charges; `max: 0` with `demands_x` means "as many as the X you announce". The ability's `mana_cost` is the WHOLE cost (`{8}`, `{X}`), because paying mana for all of it is always legal. The picks ride `activate_ability` as `waterbend_ids`: zero up to `max` distinct permanents, each paying `{1}`, the rest paid from the pool and the auto-tapper, which will not also tap a permanent named here (CR 118.3). Tapping is not the `{T}` symbol, so a creature that arrived this turn may pay (CR 302.6). The taps land with the ability already on the stack; a list that names a permanent that cannot pay, more than `max`, a permanent twice, or a permanent another component of the same cost spends is refused as `bad_request` with nothing tapped and nothing paid, and so is `waterbend_ids` on an ability without the clause. **A pay-or-counter prompt** — "Ward—Waterbend {4}" (The Unagi of Kyoshi Island) — is a `pay_unless` whose `PendingChoiceView` carries `tap_cost` beside `pay_cost`, listing the CHOOSER's waterbenders. The "Pay" answer is `resolve_choice { choice_id, apply: true, tap_ids: [...] }`; the taps pay part of the generic and the pool and auto-tap pay the rest. If the rest cannot be paid, the answer is a decline with NOTHING tapped. A malformed `tap_ids` — on a prompt with no `tap_cost`, on `apply: false`, or naming a permanent that cannot pay — is refused and the prompt stays open, so a client bug never costs the payer their spell.

- **Echo, and pay-unless payments that are not mana** (#1888, CR 702.30a / CR 118.12a / CR 702.24a, [ADR 0108 §5](decisions/0108-turn-scoped-effects-object-history-and-damage-shields.md)): an echo trigger asks its controller with the ordinary `pay_unless` prompt (held in the upkeep, #997). Most echo costs are mana and answer as before. "Echo—Discard a card", "Echo—Sacrifice two lands" and the cumulative upkeeps that discard or sacrifice carry `pay_cards: { action: "discard" | "sacrifice", count, options[] }` beside `pay_cost`, which then holds the payment's words ("Discard a card", "Sacrifice three lands"). `options` are instance IDs: the chooser's hand for a discard (sent to the chooser only; every other seat gets an empty list and the count), the permanents of the clause's kind the chooser controls for a sacrifice (in the payment order "Choose for me" uses). The "Pay" answer is `resolve_choice { choice_id, apply: true, card_ids: [...] }` naming exactly `count` distinct options; the cards are discarded or sacrificed as the cost and the permanent stays. `card_ids` on a mana prompt or on `apply: false`, the wrong count, a card named twice or one that cannot pay is refused and the prompt stays open; `apply: true` with no `card_ids` is a decline when the chooser cannot pay (fewer options than `count`) and is refused when they can. **CardView.echo_due** (omitempty) marks a battlefield permanent whose echo will trigger at its controller's next upkeep: it came under their control since the beginning of their most recent upkeep. It is derived from `Player.UpkeepsBegun` and `Card.ControlledSinceUpkeep`, never stored, and a face-down permanent never carries it (CR 708.2).
- **Computed life costs** (#1594, CR 601.2f–g via CR 602.2b, [ADR 0020 Decision 47](decisions/0020-activated-abilities.md)): `activated_abilities[i].life_cost` is the life an activation would charge the permanent's CONTROLLER right now, not a printed number. For a fixed "Pay 2 life" the two are the same; for War Room's "Pay life equal to the number of colors in your commanders' color identity" and Murderous Betrayal's "Pay half your life, rounded up" it is computed per snapshot by the same function the engine charges with, so a five-colour partner pair's War Room reads 5 and the same land under a mono-green commander reads 1. Zero is omitted, as before: a colourless commander (or none) makes War Room's cost free. No shape change and nothing to send back — the amount is fixed at announce and recorded on the stack item's paid cost; `activate_ability` carries no life field.
- **Energy costs** ([ADR 0129](decisions/0129-energy-getting-and-paying-it.md) §2 and §8, CR 107.14, CR 118.3): `activated_abilities[i].energy_cost` is a "Pay N {E}" component, the energy counters an activation removes from the activator, omitted at zero. `energy_cost_x: true` marks "Pay X {E}" (Sphinx of the Revelation): the announced `x_value` is added to `energy_cost`, `demands_x` is set, and the X may not exceed the activator's energy. When the controller has less energy than the printed part, `cant_activate` carries the engine's refusal, "Not enough energy (have 2, need 3)", the same text the error frame carries. Energy is charged under every payment posture: `strict`, permissive and `force_cast` decide only the mana. Nothing new is sent back; the engine reads the amount off the ability. The client draws {E} as an energy pip. Since ADR 0129 PR 2, `mana_abilities[i].energy_cost` is the same component on a mana ability (Aether Hub's "{T}, Pay {E}:"), with the same `cant_activate` when the controller is short, and the `/autotap` preview's `sources[i].energy` is the energy a planned source pays, so the client lists "Pay {E}" against it before the player confirms.
- **Energy paid while resolving** ([ADR 0129](decisions/0129-energy-getting-and-paying-it.md) §3, owner decision 3, CR 118.12, CR 118.12a): a `pay_unless` whose payment is energy ("you may pay {E}{E}. If you do", "sacrifice it unless you pay {E}") carries `pay_energy` (the amount; present at 0, since paying 0 {E} is a payment) beside `pay_cost`, which holds the symbols ("{E}{E}"). It is answered `{apply}` like any pay-unless; `card_ids` or `tap_ids` on it are refused. A "Pay" from a chooser with less energy than `pay_energy` is the decline (CR 118.3), so the client greys Pay with "Not enough energy (have 1, need 2)". Every energy prompt holds the step it was asked in. **`kind: "pay_amount"`** is "you may pay any amount of {E}" / "one or more {E}": `pay_amount: { min, max, goal?, unit }`, where `min` is the smallest payment (0 for "any amount", 1 for "one or more"), `max` the chooser's energy when asked, `goal` the amount the card is for (Harnessed Lightning: the damage that destroys its target; omitted when the card names none) and `unit` what one energy buys (`damage`, `counters`, `cards`, `power`, `tax` or `other`). The chooser answers `resolve_choice { choice_id, amount }` with 0 or a whole number in `min`..`max`, within their energy; anything else is refused and the prompt stays. It is routed by kind (paying nothing is 0) and **blocks the table**: the spell that asked is paused mid-resolution. A chooser who leaves pays nothing and the rest of the card runs with 0 (CR 800.4f). The legal-move enumerator offers 0, `min`, `goal` and `max`, each with `cost.energy`.
- **A printed ceiling on X** (#2581, CR 107.3a, CR 601.2b): `x_max` on the cast surface is the largest X the viewer may announce under the spell's printed "X can't be greater than <count>" — Open the Way's "the number of players in the game", Winter's Chill's "the number of snow lands you control" — counted now. It is per viewer (like `cast_prices`, never on the public half of a pile or a revealed hand), absent for every card with no printed ceiling, and present at 0 when X = 0 is the only announcement. `cast_spell` refuses an `x_value` above the count as `bad_request`, with nothing moved or paid; the count is read once, as X is announced, so a count that changes after the cast leaves the spell's X as announced. The bot enumerator never announces more, and a move's open X range (`value.max`) stops at it. The auto-tap preview reports the same number as `x_max`.
- **Energy in alternative and additional costs** ([ADR 0129](decisions/0129-energy-getting-and-paying-it.md) §5, PR 4, CR 107.14, CR 118.3, CR 118.9): `alternative_costs[i].energy` is the energy an offer charges (Nissa, Worldsoul Speaker's granted eight, Primal Prayers' one, Amped Raptor's "equal to its mana value"), omitted at zero. An offer the caster is short of energy for is not listed, and a claim of one is refused with "Not enough energy (have 2, need 8)". `cast_prices[i].energy` is the same on a price in the exile strip. A granted offer that casts "as though it had flash" (Primal Prayers, CR 601.3c) answers `timing_closed` for its own claim: the printed cost may be closed while the offer is open. **Replicate** (CR 702.56a) is an `optional_costs` entry keyed `replicate` and paid by naming its index once per payment, like a multikicker; `optional_costs[i].energy` is the energy each payment charges, `max_times` is capped at the payments the viewer's energy covers, and `energy_short: true` means it covers none, so the offer cannot be taken. A cast whose replicate payments and alternative cost together owe more energy than the caster has is refused. The energy is charged under every payment posture. The legal-move enumerator names it on `cost.energy`.
- **Exert costs** ([ADR 0130](decisions/0130-exert.md) §4, CR 701.43a, CR 602.2b): `activated_abilities[i].exert` and `mana_abilities[i].exert` (omitempty) mark an "Exert this creature" / "Exert this land" component (Steward of Solidarity, Arena of Glory). Activating the row exerts its source: it won't untap during the activator's next untap step, and the log has an `exert` line. It is always payable while the source is on the battlefield, even after an earlier exert this turn (CR 701.43b), so it never sets `cant_activate`, and nothing is sent back for it. The auto-tapper never pays an exert mana row, so a seat whose only way to pay is an exert row activates it itself with `activate_mana_ability`. In `legal_moves`, a move whose cost exerts its source carries `cost.exert: true`. The client draws an "exert" chip on the row.
- **Tap-other-permanent costs** (#758 / #759 / #1421, [ADR 0071 addendum 2026-09-23](decisions/0071-designations-that-switch-abilities-on.md), [ADR 0073 amendment 2026-09-24](decisions/0073-optional-additional-costs-and-the-cast-gate.md)): "Tap N untapped permanents you control" is one cost component shared by ordinary activated abilities and mana abilities. `activated_abilities[i]` and `mana_abilities[i]` carry the same `tap_others_label` plus `tap_others_options` `LegalTargetsView` (`{ cards[], min, max, count_from_x? }`); for a fixed clause `min` and `max` are both the printed count, and `cards` are the UNTAPPED permanents the viewer controls that could pay right now. The answer rides back as `tap_ids` on the corresponding `activate_ability` or `activate_mana_ability` payload: exactly the required number of distinct permanents from `cards`. The client reuses the same picker with the verb "Tap". This is not the `{T}` symbol, so a creature that arrived this turn is a legal pick (CR 302.6), and paying it does not target, so a hexproof creature is too (CR 601.2h). The source is left off when the clause says "another", or when the same ability also prints `{T}` (CR 118.3 — one permanent cannot pay both components). An absent or short list means the cost cannot be paid and the engine refuses the activation. Ordinary activated abilities record the payment in `PaidCost.TappedOthers`; station reads the tapped creature's power from it as the ability resolves, or its last-known power if it has left the battlefield (CR 608.2h). Mana abilities resolve immediately as usual.

  **#1421: "Tap X".** On an ordinary activated ability, `count_from_x: true` means `min` and `max` are `0 / 0` because the printed clause has no fixed bounds. The picker requires at least one useful pick and allows as many of the offered permanents as the player chooses; the number picked is sent as both the length of `tap_ids` and `x_value`. The engine requires those values to agree, then locks X onto the stack item for the effect to read at resolution. A variable tap clause sets `demands_x: true` even though it adds no generic mana, so clients must read the flag rather than parse the mana cost. A cost may not combine this clause with `{X}` mana or a second CountFromX component, since one announced X cannot represent two independently chosen values. It is also refused on a mana ability: CR 605.3b has no X-announcement step there. Fixed clauses remain byte-for-byte unchanged.

- **Discard costs on a MANA ability** (#1213): `mana_abilities[i]` carries `discard_cost_n` / `discard_cost_label` / `discard_cost_options` with exactly the meanings `activated_abilities[i]` gives them (#660), and the answer rides back as `discard_ids` on `activate_mana_ability`. Skirge Familiar's "Discard a card: Add {B}" is the one printed card. It is the same component with a second owner, so one client picker serves both ability kinds. The discard goes through the engine's one discard path, so `EventDiscardCard` fires per card and madness sees it; the AUTO-TAPPER never plans such a source, because which card to pitch is a decision the planner does not make. **#1369:** `discard_cost_options` reaches the controller's frame alone; every other viewer, spectators included, gets the count and the label.
- **Colour options on a MANA ability** (#1443, [ADR 0040](decisions/0040-mana-pipeline.md) amendment 2026-09-24): `mana_abilities[i].color_options`?: `string[][]` is, for each slot of the ability's output that asks for a colour, the colours that pick would offer the controller if the ability were activated now — one list per PICKING slot, in output order. A painland's `{R|W}` ships `[["R","W"]]`; Birds of Paradise under a mono-green commander `[["G","W","U","B","R"]]` (identity first, #843); Command Tower only the identity's colours (CR 903.4f); Mystic Gate's `{W|U}{W|U}` two lists. It is the SAME list, from the same server function, that the `mana_pick` prompt's `color_options` would carry — so a client may draw these at the card and send the answer back as `activate_mana_ability`'s `color` / `colors`, and nothing is tapped until the player has chosen. Clients render in the order sent. Absent for an ability that picks nothing (a Forest, Sol Ring's `{C}{C}`) and for one that adds no mana (`adds_no_mana`). A slot counts as picking by its PRINTED width, so a Command Tower narrowed to one colour still ships a one-entry list. For a computed output (Exotic Orchard) the list is what the board supports now; if the activation's own cost changes it, a named colour it no longer offers falls back to the `mana_pick`. **#2558:** `mana_abilities[i].different_colors: true` marks an ability that adds "N mana of different colors" (Firemind Vessel, `produced` `"{W|U|B|R|G:2}"`): `color_options` carries one list per mana, and the answer must name N DIFFERENT colours — a repeat is refused with `that colour is not one this mana ability can add` before anything is paid. A client offers only answers whose colours all differ. Absent for every other ability.
- **Exile-a-card costs on a MANA ability** (#1283, [ADR 0011](decisions/0011-mana-pool-and-auto-tapper.md) amendment 2026-09-23): `mana_abilities[i]` carries `exile_cost_n` / `exile_cost_label` / `exile_cost_options` — the discard triple's shape under its own names — and the answer rides back as `exile_ids` on `activate_mana_ability`: exactly `exile_cost_n` distinct cards from the options, none of them also named in `discard_ids`. Cadaverous Bloom's "Exile a card from your hand: Add {B}{B} or {G}{G}" is the printed card. It is NOT a discard: the card goes to exile through the ordinary zone change, no discard event fires and madness does not see it, which is why it has its own fields. Like the discard, the AUTO-TAPPER never plans such a source. Not to be confused with `exile_self` (#1228), which exiles the SOURCE and sends nothing. **#1297:** the options come with `exile_cost_zone` — `"hand"` or `"graveyard"`, the pile they are in — so the client resolves them out of the right zone; a clause that names no pile is the hand, and every mana ability that prints the component today is one. **#1369:** a HAND list is read out of the controller's hand, so it reaches the controller's frame alone; every other viewer, spectators included, gets `exile_cost_n`, `exile_cost_label` and `exile_cost_zone` and no list. A graveyard list (`exile_cost_zone: "graveyard"`) is read off a public pile and goes to every viewer.
- **"Discard X cards" on an ACTIVATED ability** (#2527, [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md)'s 2026-10-07 amendment): `discard_cost_count_from_x: true` beside `discard_cost_label` ("X cards") and `discard_cost_options` marks a discard cost whose count is the X the activator announces — Gix, Yawgmoth Praetor. `discard_cost_n` is absent (there is no fixed count) and `demands_x` is set. The picks ride back as `discard_ids` and their number IS `x_value`: the engine refuses an activation whose `discard_ids` length differs from `x_value`, whose cards are not distinct cards in the activator's own hand, or whose X is negative, with nothing paid. Zero cards at X=0 is a legal payment. The client always opens its picker (any number of the offered cards, none included) and skips the X stepper. `discard_cost_count_from_x` is public; `discard_cost_options` is the controller's hand and reaches the controller's frame alone, like every other hand-sourced list.
- **"Reveal X black cards from your hand" on an ACTIVATED ability** (#2598, [ADR 0020](decisions/0020-activated-abilities.md)'s 2026-10-08 amendment, Decisions 58–60): `activated_abilities[i].reveal_cost_label` ("X black cards") and `reveal_cost_options` mark a reveal cost, with `reveal_cost_count_from_x: true` when the count is the X the activator announces (the whole Martyr cycle) and `reveal_cost_n` the printed count otherwise; `demands_x` is set beside the X form. The picks ride back as `activate_ability.reveal_ids` and, for the X form, their number IS `x_value`: the engine refuses an activation whose `reveal_ids` length differs from the count (or from `x_value`), whose cards are not distinct cards in the activator's own hand matching the clause's colour, or that sends `reveal_ids` to an ability with no reveal cost, with nothing paid or revealed. Zero cards at X=0 is a legal payment. Revealing moves nothing (CR 701.20b): the cards stay in the hand, every seat learns them (`reveal` events, the same `reveals` window a cast's reveal uses), and no other component's pick is excluded from them. The client reuses its discard picker with the verb "Reveal", asked first, and skips the X stepper. `reveal_cost_label`, `reveal_cost_n` and `reveal_cost_count_from_x` are public; `reveal_cost_options` is the controller's hand and reaches the controller's frame alone.
- **"Discard a card with mana value X" on an ACTIVATED ability** (#2190, [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md)'s second 2026-10-07 amendment): `discard_cost_mana_value_x: true` beside `discard_cost_n: 1`, `discard_cost_label` and `discard_cost_options` marks a discard cost whose card's mana value is the announced X — Kozilek, the Great Distortion. `demands_x` is set but the client does not ask for X: the card picked IS the announcement, so it sends that card's mana value as `x_value` (the engine refuses any other, and a card whose cost it cannot read, with nothing paid) and narrows the ability's target clause by it (`mana_value_equals_x`, with `mana_values` counting {X} on a spell on the stack as the value chosen for it). `discard_cost_mana_value_x` is public; `discard_cost_options` is the controller's hand and reaches the controller's frame alone.
- **Library costs and random discards on an ACTIVATED ability** ([ADR 0109](decisions/0109-rule-gates-land-types-mana-and-cost-components.md) §7 and owner decision 3, #1902): three components, paid at announce with the rest of the cost. `top_cost_n` (the count; present marks the component), `top_cost_label` (the clause as printed, without the verb) and `top_cost_options` (every card in the viewer's own hand but the source, in hand order) describe "Put a card from your hand on top of your library"; the answer rides back as `top_ids`, and the client skips its picker when the options number exactly `top_cost_n`. The card goes to the top of its owner's library, its owner still knows it, and it is not a discard. `library_exile_cost_n` is "Exile the top N cards of your library": nothing to pick, nothing sent, and a library of fewer than N cards refuses the activation with `not enough cards in your library to pay that cost`. `discard_cost_random: true` beside `discard_cost_n` and `discard_cost_label` is "Discard N cards at random": nothing to pick, no `discard_cost_options`, no `discard_ids` (one is refused), and a hand that can't supply N cards after the rest of the cost refuses with `not enough cards in your hand to discard at random`. The engine draws the cards (CR 701.9b), and the library exile and the random discard are paid after every other cost (CR 601.2h); the client confirms either one before sending. A commander put on top of its library is asked CR 903.9b first (#1397); a commander exiled off the top or discarded at random is paid like any other card and then offered the command zone by the CR 903.9a state-based action (ADR 0115). **Privacy:** `top_cost_options` is read out of the controller's hand and reaches the controller's frame alone; every other viewer gets the count and the label.
- **Exile-N-cards costs on an ACTIVATED ability** (#1297, [ADR 0020](decisions/0020-activated-abilities.md) and [ADR 0073](decisions/0073-optional-additional-costs-and-the-cast-gate.md) amendments 2026-09-23): `activated_abilities[i]` (and `zone_abilities[i]`) carry the same four fields a mana ability's exile cost does — `exile_cost_n` (the count; present marks the component), `exile_cost_label` (the clause as printed, without the verb: `"two cards"`, `"a creature card"`), `exile_cost_options` (the matching cards in the viewer's OWN hand or graveyard that could pay right now, in pile order, the source excluded; absent when nothing can pay) and `exile_cost_zone` (`"hand"` or `"graveyard"`). The answer rides back as `exile_ids` on `activate_ability`. The cards are exiled at ANNOUNCE (CR 602.2b), with the rest of the cost and before the ability is on the stack, through the ordinary zone change with the cost-cannot-pause bit — so a commander exiled this way is exiled and then offered the command zone by the CR 903.9a state-based action (ADR 0115), no `EventDiscardCard` fires and madness does not see it. A graveyard card is not a hand payment, a hand card is not a graveyard payment, and a card that fails the clause's predicate (Tome Shredder's "an instant or sorcery card") is refused with nothing paid. The server records the exiled cards on the stack item's paid-cost record (`PaidCost.Exiled`, not on the wire), which is how an effect reads "the card exiled this way" (Holistic Wisdom). The client reuses the discard picker with the verb changed and the pile named, and skips it when the options number exactly `exile_cost_n`. **#1369:** when `exile_cost_zone` is `"hand"` the options reach the controller's frame alone — every other viewer, spectators included, gets the count, the label and the zone and no list, because the list's length and IDs are facts about a hidden hand. A graveyard list is read off a public pile and goes to every viewer ([ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)'s 2026-09-23 amendment).
- **Set rule on a sacrifice cost** (#2526, [ADR 0020](decisions/0020-activated-abilities.md) amendment 2026-10-07): "Sacrifice a Swamp and a Forest" (Jarad, Golgari Lich Lord) — a sacrifice clause whose parts are different kinds of permanent. Its `sacrifice_options` carry one extra field, `each_of`: `[{ label, cards[] }]`, one entry per printed part, each listing the candidates (from `cards`) that could fill it. `cards` is the union and `min` / `max` are the entry count (`2 / 2`), so a client that ignores `each_of` still sees an ordinary sacrifice of two. The answer is the usual `sacrifice_ids`, and must name exactly one permanent per entry, all distinct, each passing its entry — two Swamps are refused with nothing paid; a Swamp Forest can fill either entry but not both. `SacrificeCostModal` holds confirm until the picks fill every part. Absent on every clause without a set rule.
- **Exile-a-permanent costs** (#1600, [ADR 0020](decisions/0020-activated-abilities.md) amendment 2026-10-03): "Exile a creature you control" — The Soul Stone's harness ("{6}{B}, {T}, Exile a creature you control"), Altar of Bhaal, City of Shadows, and on a MANA ability Food Chain ("Exile a creature you control: Add X mana …"). `activated_abilities[i]` (and `mana_abilities[i]`) carry `exile_permanent_label` (the clause as printed, without the verb) and `exile_permanent_options`, a `LegalTargetsView` `{ cards[], min, max }` shaped exactly like `return_options`: `min` and `max` are the clause's count (1 on every printed card) and `cards` are the permanents the viewer controls that could pay right now, in the payment order `sacrifice_options` uses. The answer rides back as `exile_permanent_ids` on `activate_ability` or `activate_mana_ability`: exactly `max` distinct permanents from `cards`, none of them also sacrificed or returned by the same payment. An absent or short list means the cost cannot be paid (CR 118.3) and the engine refuses the activation. It is NOT `exile_ids`, which names CARDS in a hand or a graveyard, and it does not target, so a hexproof creature you control is on the list. The permanents are exiled at announce, through the ordinary zone change with the cost-cannot-pause bit: leaves-the-battlefield triggers see them, dies and sacrifice triggers do not, and on a mana ability those triggers reach the stack after the mana is in the pool (CR 605.3a). A commander named to the cost is asked about the command zone BEFORE anything is paid (#1397); the cost is paid either way. The AUTO-TAPPER never plans a mana ability with this component. The exiled permanents are recorded on the paid-cost record (`PaidCost.Exiled`, not on the wire), which is how Food Chain reads "the exiled creature's mana value" — as it last existed on the battlefield.
  - **Craft** (#2124, [ADR 0137](decisions/0137-craft.md), CR 702.167b): on a craft ability ("Craft with artifact {3}{W}{W}") the same `exile_permanent_options` may also list CARDS in the viewer's own graveyard, after the permanents — a material named without the word "card" may be either, and one payment may mix them. `min` / `max` are the clause's count ("Craft with two creatures" is 2). A client resolves the ids against the battlefield AND the viewer's graveyard; `exile_permanent_ids` carries both kinds back. No new field: a craft ability is otherwise an ordinary activated ability whose cost also exiles its source.
- **Private prompt text** ([ADR 0133](decisions/0133-opening-hand-actions.md)): a `pending_choices[i]` with `private_text: true` asks a question that names a card in the chooser's hidden hand — an opening-hand action ("Begin the game with Leyline of Sanctity on the battlefield?", a `confirm`). The chooser gets `source`, `reason`, `accept_label` and `decline_label`; every other viewer (spectators and admins included) gets the prompt, its `kind` and its `chooser`, with `reason` replaced by a neutral line and no `source` or labels. The table sees who the game is waiting on and not which card.
- **Set rules over a clause's targets** (#1559, CR 601.2c, [ADR 0019 amendment 2026-09-24](decisions/0019-structured-targeting.md)): any `LegalTargetsView` that describes a TARGET clause — a card's `legal_targets` and each entry of `clauses`, `modes[i].legal_targets` / `modes[i].clauses`, `alternative_costs[i].legal_targets`, `optional_costs[i].legal_targets` / `clauses`, `activated_abilities[i].legal_targets` / `clauses`, and a `pick_target` prompt's `pick_target` — may carry two more fields. **`different`** is `{ label, keys? }`: a rule no two picks of THIS clause may break by sharing a key. `label` completes "those targets must …" ("each have a different mana value", "be controlled by different players", "each have a different name") and is what the picker says in its banner; `keys` maps each legal card id to its key, and a card missing from it (an unreadable cost) collides with nothing. The client greys and refuses a candidate whose key a pick of the same clause already holds; the engine refuses a set that breaks the rule at announce with `bad_request` and the message `illegal target: those targets must <label>`, and re-judges the rule over the surviving picks at resolution (two survivors that now share a key are BOTH illegal). Player picks have no key. **`same`** (#1807, [ADR 0106 §5](decisions/0106-five-small-seams-from-the-s50-rechecks.md)) has the same `{ label, keys? }` shape and the opposite rule: every pick of THIS clause must share one key. Its only label so far is "come from a single graveyard", keyed on each card's owner. The client greys and refuses a candidate whose key differs from a pick of the same clause, so with nothing picked every candidate is open; a card missing from `keys` fits any group. The engine refuses a set with two keys at announce with `illegal target: those targets must <label>`, and offers a trigger or cast only when one key group can fill the clause's minimum. **`mana_value_at_most_x`** (with **`mana_values`**, card id → mana value) marks a clause bounded by the announced X — "with mana value X or less" (Agadeem's Awakening). Like `count_from_x`, the server builds the legal set before X is chosen, so it is a superset: the client drops each card whose `mana_values` entry exceeds the X it collected (a card with no entry meets no bound), and the engine refuses a pick above the announced X. **`mana_value_equals_x`** ("with mana value X", Lazav) narrows the same `mana_values` by an exact match. **`power_at_most_x`** and **`toughness_at_most_x`** (with **`powers`** / **`toughnesses`**, card id → the candidate's power or toughness now, counters included) are the same bound on another statistic (ADR 0109 §9). **`x_from_counters_removed`** says the X is the number of counters the activation's cost is removing — the sum of its `counter_counts` — rather than the `x_value` (Simic Manipulator). A spell's clause and an activated ability's clause can be X-bounded; a trigger's never is. Both fields are absent on every clause that prints neither, which is nearly all of them, and on cost-payment views (`sacrifice_options` and friends), which are not targeting.
- **Divided effects** (#1563, CR 601.2d / 700.2i, [ADR 0065 amendment 2026-09-28](decisions/0065-modal-and-multi-target-clauses.md)): any `LegalTargetsView` describing a TARGET clause — a card's `legal_targets` and `clauses[]`, `modes[i].legal_targets` / `clauses`, `activated_abilities[i].legal_targets` / `clauses`, and a `pick_target` prompt's `pick_target` — may carry **`divide`**: `{ total?, from_x?, double_from_x? }`, the amount the clause's effect is "divided as you choose" among its picks. `total` is a fixed amount (Fury's 4); with `from_x` the amount is the X the client collected (Fire Covenant), or twice X once X reaches `double_from_x` (Shatterskull Smashing's 6). A `pick_target` prompt's amount is always a fixed `total` — a trigger announces no X. The caster announces the split with the targets as **`distribution`**, target id → share, on `cast_spell`, `activate_ability` or the `pick_target` answer to `resolve_choice`. The engine refuses (`bad_request`) a division in which any target of the clause gets less than 1, the shares do not sum to the amount, a share names something that is not one of the clause's targets, or there are more targets than the amount (each must get at least 1) — the message after `game: invalid parameter:` says which. A lone target may omit `distribution`: it takes the whole amount. The settled division rides `StackItemView.distribution`, is carried by undo, snapshot and copy, and is honoured at resolution; a target that became illegal takes nothing and its share is **not** moved to the others (CR 608.2b). `divide` is absent on every clause that divides nothing. A card on the free-form S13.1 path (no structured clause) keeps sending `distribution` as unjudged data for the table, as before.
  **#1657 ([ADR 0065 amendment 2026-09-28 (b)](decisions/0065-modal-and-multi-target-clauses.md)):** `divide` gains **`up_to`** — "distribute UP TO that many" (Lathiel, the Bounteous Dawn): the shares may add up to LESS than the amount, never more, and each chosen target still gets at least 1; the engine refuses a sum over the amount. An amount the card reads off the board or off the claimed cost — Ureni's "X is the number of lands you control", Orca's "equal to its power", Avacyn's Judgment's "2, or X if its madness cost was paid" — arrives already resolved into `total` / `from_x` for the viewer (and, on an `alternative_costs[]` entry's `legal_targets`, for that cost): the client never sees the rule. The server fixes the amount when the division is announced — the cast, the activation, or the moment a trigger's `pick_target` prompt opens — so a prompt's `total` does not move while it is open. `activated_abilities[i].legal_targets` now carries `divide` for a single-clause divided ability too (Mogg Mob); it was omitted before #1657.
- **PendingChoiceView** (S14, omitempty): `{ id, kind, chooser, from_player, count, source, reason, options[], color_options? }` — one entry in the async effect-driven decision queue. `kind` is one of `discard_from_hand` (S14), `mana_pick` (S15), `replacement_order` / `optional_replacement` (S17), `damage_assignment` (S18), or `trigger_prompt` (S19, CR 603.5 optional "you may" triggered ability — chooser is the source's controller by default; `reason` is the prompt header text; `source` is the card that fired). `chooser` is the seat who picks (Thoughtseize — caster); `from_player` is the source zone owner (the discarder, or for `mana_pick` the controller of the mana source). `count` is how many cards must be picked (always 1 for `mana_pick` / `optional_replacement` / `trigger_prompt`). `source` is the card that produced the effect; `reason` is a human-readable tag (`"Thoughtseize"`, `"Add one mana of any color"`) for client copy. `options[]` is an array of `CardView` entries for `discard_from_hand` (omitted for non-card-list kinds); **ADR 0116:** a `discard_from_hand` entry also carries `eligible` (the instance IDs among `options[]` the chooser may pick: "you choose a nonland card from it") and `eligible_label` (what they are, as the card prints it: `"nonland card"`). `options[]` stays the whole revealed hand, which every seat may see because the reveal goes to the whole table (CR 701.20a), and `eligible` is filtered per viewer by the same rule as `options[]`. An entry with no `eligible` (one queued before ADR 0116) accepts every option. **#2115:** a `revealed_hand_pick` entry carries the same `options[]`, `eligible` and `eligible_label`, plus `choose_min` / `choose_max` (0 and `count` for "you may choose", `count` and `count` otherwise), `pick_destination` (`"discard"` or `"exile"`) and `pick_from_graveyard` (when set, `options[]` ends with `from_player`'s graveyard, every card of which is in `eligible`). All of it reaches every seat: the hand was revealed and a graveyard is public; `color_options` is the legal-color button list for `mana_pick`; `type_options` is the full creature-type vocabulary for `choose_creature_type`, materialised from the engine rather than stored on the choice. The chooser's client pops `ChoicePromptModal` and submits `resolve_choice` with the kind-appropriate payload. Distinct from `discard_pending` (S13.4 cleanup-step over-max, chooser == discarder). **S16.5 adds `copy_target`** — "you may have this permanent enter as a copy of ..." (Clone, Phyrexian Metamorph, Spark Double, Sakashima the Impostor). `options[]` carries the permanents that may be copied (public battlefield cards, unredacted); answered with the shared `{choice_id, card_ids}` payload, where an EMPTY list declines and the permanent enters as its own printed self. The permanent is still on the stack while the prompt is open — the answer decides what it enters AS, so its own ETB trigger has not fired yet either. **S21/S22 look-at kinds:** `scry`, `surveil` and `look_at_top` put the looked-at library cards in `options[]`, top-first. `put_in_library` (#996) puts the cards being placed there too — from a hand, a library, the battlefield or the stack — with the same chooser-only redaction, and adds `placement` (`"top"` / `"bottom"` / `"top_or_bottom"`), which stays public along with `count` — as do #1298's `top_count` and `top_depth`. A look at ANOTHER player's library (Jace's +2, Portent) projects the cards to the looker alone; the library's owner gets the prompt and its shape, not the cards. Both keywords are "look at", not "reveal" — only the chooser is a knower. The **count** stays visible to the table, which is correct: "scry 2" is a printed number, and `search_library` goes further still by dropping `options[]` outright for non-choosers, because the number of MATCHES is itself hidden information about a hidden zone. **S31 correction:** every kind now drops an option the viewer is not a knower of, rather than projecting it as a back. Redacting kept the `instance_id`, which is the right trade for a face-down permanent (the ID is already on that viewer's wire) and the wrong one for a card whose zone the seat projection has already stripped for that same viewer — a stable UUID for a specific card in a hidden zone is a correlation handle, since the card is drawn in private and cast in public under the same ID. The chooser always keeps their whole list, known or not; picking a back is a legal answer.
  **`commander_return`** ([ADR 0115](decisions/0115-commanders-die.md), CR 903.9a / CR 704.6d). "Your commander was put into a graveyard or into exile since the last state-based action check; put it into the command zone?", asked of the commander's OWNER (not its controller) as a state-based action. `source` is the commander, in its owner's graveyard or in exile; `reason` names it. Answered with `{apply}`: `true` moves it to its owner's command zone, `false` leaves it where it is, and it is not asked again until it is put into a graveyard or exile again. The entry carries one computed field, `playable_from_zone` (omitempty): whether the owner could cast or play the commander from the zone it is in now (a cast permission, flashback, escape, an adventurer on an adventure), ignoring timing and mana. It is recomputed on every view, never stored. While any `commander_return` prompt is open, no waiting trigger goes on the stack (CR 704.3); in the cleanup step a "yes" gives the active player priority (CR 514.3a) and a "no" does not. The prompt is plain data, so a table waiting on it is a restore point. A restore point that names a pending-choice kind the server does not know is refused and kept, as an unknown effect key is. Shipped switched off in ADR 0115 PR 2 and switched on in PR 3: since then every commander put into a graveyard or exile (destroyed, sacrificed, countered, discarded, milled, exiled, or paid as such a cost) is asked through this kind after it lands, and `optional_replacement` is left for a commander headed for a hand or a library (CR 903.9b). **ADR 0115 PR 5:** the CR 903.9b `optional_replacement` for a commander now sets `source` to the commander on every path (it already did for a cost; a battlefield exit left it empty), so the client can show the card. No field was added: `source`, `chooser` and `reason` of a `commander_return` are public on every seat's view, which is what lets the seats that are not deciding read "X is deciding about their commander" and the owner's prompt show the card. The line is client-side; nothing on the wire says "deciding".
  **`optional_replacement` facts and the shockland price** (#2390). An `optional_replacement` carries two computed fields (omitempty), recomputed on every view and never stored: **`dredge`** is N when the prompt offers a dredge (CR 702.52a) — the number of cards the yes mills — with the card it returns in `source`, in the chooser's graveyard; a grant that names no card (The Necrobloom's "land cards in your graveyard have dredge 2") leaves `source` empty, and the land is picked after the yes. **`playable_from_zone`** is set on CR 903.9b's commander question when the commander is headed for its owner's HAND, the one zone it can always be cast from (without the CR 903.8 tax); it is unset for a library. Since #2420 it is set on the question asked before a cost is paid too, when that cost returns the commander to its owner's hand (ninjutsu, "return a creature you control to its owner's hand", Daze's alternative cost). An `entry_pay_life` prompt's pay move in `legal_moves` now carries its life as `cost.life` (#547's MoveCost), beside the `pay_cost` text ("2 life"). Nothing else changed: the client renders these prompts as before; the heuristic bot reads the fields to answer them (docs/bot.md).

  **`entry_sacrifice` devour facts** (#2419). An `entry_sacrifice` that is CR 702.82a's devour carries **`devour`** (N, the +1/+1 counters each sacrificed creature buys), and **`devour_draw`** / **`devour_life`** (what the creature's own "for each creature it devoured" ability pays per creature, Skullmulcher and Marrow Chomper; omitted when zero). All three are computed on every view and omitted on the fixed-count sacrifice lands (Heart of Yavimaya) that share the kind. The client renders the prompt as before; the heuristic bot prices each creature against its counters.
  **`may_cast`** (S28 cascade; ADR 0099 discover). "You may cast it without paying its mana cost", asked during a resolution about one card. Answered with `{apply}`. The entry carries `may_cast_card` (the instance id of the offered card, which is in exile face up), `may_cast_keyword` (`cascade`, `discover`, `suspend` or `madness`; absent for an offer that is only its card's text, like hideaway) and `accept_label` / `decline_label` (the branch names, e.g. "Cast it free" / "Put it into your hand"). Accepting stamps a cast permission on the card; the cast itself is an ordinary `cast_spell` with `from_zone: "exile"`. A discover or cascade permission is flash-timed (CR 608.2g), capped at the mana value the keyword allows (a face over the cap is refused with `that spell's mana value is too high for this free cast`), and closes on the holder's next `pass_priority`: the card then goes into their hand (discover) or to the bottom of their library (cascade). If that move raises a trigger or a prompt, the pass is spent on it and priority stays with the holder.
- **LegalMoveView** (S31, omitempty): `{ type, player, params?, kind, label, source?, always_legal?, cost?, targets_stack?, idle_hint?, value? }` — one fully-specified thing the viewer's seat may do right now, produced by `server/internal/legal` ([ADR 0033](decisions/0033-ai-bot-seat.md) §1). `type` is the action type the move performs (`pass_priority`, `cast_spell`, `activate_ability`, `activate_mana_ability`, `declare_attacker`, `declare_blocker`, `declare_blockers`, `finish_blocks`, `resolve_choice`, `keep_hand`, `mulligan`, `discard_selection`, `special_action`, `roll_opening`, `choose_starting_player`) and `params` is **exactly** the `ActionPayload.params` object that performs it — a client fires a move by sending `{ type, player, params }` unaltered, and the server is contractually obliged to accept it. `kind` coarsely groups the move (`pass` / `land` / `cast` / `activate` / `mana` / `attack` / `block` / `finish_blocks` / `choice` / `mulligan` / `special_action` / `opening_roll`) so a consumer can ask "is there anything here but pass?" without parsing labels. `finish_blocks` (#1501) is a declaring defender's "No blocks" / "Done blocking": offered, always-legal and with no `source` or `params`, to every seat whose block declaration is still pending and owes no CR 509.1c requirement — it is NOT a pass, and the viewer holds no priority while it is on offer. `opening_roll` ([ADR 0121](decisions/0121-animated-dice.md) §4) is the only kind on offer while `opening_roll` is present in the view: `roll_opening` (always-legal, no `params`) to a seat in the current round that has not rolled, while nobody has won; and, to the winner alone, one `choose_starting_player` per seat still in the game (`params: { seat }`, labelled `"I go first"` for the winner's own seat, always-legal, and `"<name> goes first"` for each other). `host_roll_remaining` and a table roll are never listed. `label` is human-readable and menu-ready (`"Cast Lightning Bolt targeting Kess"`). `source` is the instance ID of the card the move is about, when there is one — this is the join key the client greys hand cards on.
  **Own seat only, always.** `legal_moves` names cards in a hand; another seat's list would leak exactly the hidden information every other redaction in this document protects. The server enumerates per seat into an unexported map and `FilterViewFor` hands back only the viewer's own entry; the admin (empty viewer ID), spectators and replay readers get nothing, which is the one place the admin's "empty viewer sees everything" convention deliberately does not apply. The unfiltered `GameView` that reaches the crash dump and the replay log carries no move list at all.
  **Empty is the normal case.** The list is populated only when the seat actually owes a decision — priority, a pending choice, a combat declaration, the mulligan window or a cleanup discard. A seat with nothing to do gets the field omitted, and a client must read an absent `legal_moves` as "no information", not as "nothing is legal": the safe reading when the field is missing is to stay permissive and let the server reject.
  **`always_legal`.** Present and `true` on a move the server cannot refuse whatever else happens between this frame and the click: `pass_priority`, and a pending choice's one unconditional answer (a `search_library` prompt's "fail to find", CR 701.23b; a `choose_cards` prompt's "choose nothing" when its `choose_min` is 0). Absent means "legal right now", which is what every entry in this list already promises — the flag is the stronger claim that it is still legal after the board has moved or after some other answer to the same prompt was rejected. Automated seats read it as the way out of a prompt they cannot otherwise answer: a seat owing a choice is enumerated that choice's answers and nothing else, so there is no pass to fall back on, and a bot with nowhere to go stops playing and holds the table (#544). Choice kinds that have no unconditional answer mark nothing.
  **`targets_stack`** (S35, #1307). Present and `true` on a `cast_spell` or `activate_ability` move when at least one of its chosen targets is an object already on the stack — a spell (a card in `stack`) or an activated / triggered ability (an entry in `stack_items` with no card behind it, CR 115.4). This is what lets a client tell a counterspell-shaped response (Counterspell, Stifle) apart from any other instant or ability it happens to be able to cast or activate right now, without re-deriving targeting rules client-side. Never set on `mana` or `land` moves — neither ever targets. On a modal move (Cryptic Command) the flag is per ANNOUNCEMENT: only the modes whose chosen targets are on the stack come back flagged, so a card with one counter mode and one burn mode ships both, correctly split.
  **`idle_hint`** (#1918, omitempty). A player-facing sentence on a legal `cast_spell` move that would do nothing on the board as it stands, for example `"Overloaded, this does nothing right now: there's no spell you don't control."` The move is still legal and the server still accepts it. CR 601.2c only asks a spell for legal targets, and an overloaded spell has none (CR 702.96a). The field is advice, like `cost`, and is never sent back. It is deliberately narrow. Only an alternative cost that turns "target" into "each" sets it, which today means overload. It is set only when nothing matches the clause that cost replaced. That clause is read **without** the targeting restrictions, because "each" is not targeting (CR 702.96b): an opponent's hexproof creature is something an overloaded Cyclonic Rift returns, so with one on the table the hint is absent. It is not a general "this does nothing" detector, and its absence promises nothing.
  **`cost`** (omitempty): `{ life?, phyrexian_life?, loyalty?, counters?: [{ card_id, counter, n }], mana?, hand?, energy? }`. It is the part of a move's price that `params` cannot name, because the engine reads it off the ability rather than off the payload (#74, #547). It is advice about the move, never sent back. `life` is life paid at announce. `phyrexian_life` (#1677, CR 107.4f) is the part of `life` that pays Phyrexian mana symbols instead of mana. It is already included in `life`, and it is in life points, 2 per symbol, where `params.phyrexian_life` counts symbols. It is broken out because it buys nothing: a policy may read a printed life cost as a sign of how much a move does, and life spent on a Phyrexian symbol is the same move at a different price. `loyalty` is a loyalty ability's signed counter delta on the source. `counters` (#625) lists the counters a "remove N counters" cost takes and which permanent they come off. That permanent is often not `source`: Heart of Kiran's alternative crew is paid with a planeswalker's loyalty counter. A policy should price the removal against that permanent, including the whole permanent when the counter is its last loyalty. **`mana`** is a mana cost string the move charges that `params` cannot name, on two kinds of move. On a `declare_attacker` move ([ADR 0080](decisions/0080-attack-taxes.md)) it is the CR 508.1a attack tax, `"{2}"` for an attack into Propaganda: an attack has no printed cost, so without this a bot prices an attack under Ghostly Prison exactly like a free one. On every `cast_spell` move ([ADR 0136](decisions/0136-planning-the-turns-mana.md) §2) it is the **total mana the cast charges** (CR 601.2f), which the card's printed `mana_cost` is not: the printed cost or the alternative cost the move claims (CR 118.9), plus the commander tax (CR 903.8) and the mana of the announced additional costs, with the board's increases and reductions applied, less what the move's delve payment exiles, and with X settled at the move's `x_value`. Tatyova, Benthic Druid cast from the command zone a second time is `"{5}{G}{U}"`; Blaze at X=3 is `"{3}{R}"`; a flashed-back Faithless Looting is `"{2}{R}"`. Phyrexian symbols the move pays with life are left out (they are in `life`), and the rest are written as their coloured half, because the move pays them with mana. A cast that charges no mana says `"{0}"`. It is the cost the server's payment check was run against when it offered the move, so a cast move always carries `cost`. **`hand`** (#1600) is how many cards a "Discard your hand" cost throws away if the move is made now (Lion's Eye Diamond, Null Brooch, Slate of Ancestry). Every other discard names its cards in `params.discard_ids`; this one names none, because the engine refuses ids for it, so without this a bot reads the whole hand as free. It is absent for an empty hand, which pays the cost. **`energy`** ([ADR 0129](decisions/0129-energy-getting-and-paying-it.md) §7, CR 107.14) is the energy counters the move removes from the seat: a "Pay N {E}" activation's N, or N + X for "Pay X {E}" at the move's `x_value`, and on a cast the energy its alternative cost and its replicate payments charge (ADR 0129 PR 4). The params name no amount, so without it a bot reads Aethertorch Renegade's "Pay eight {E}" as free. The move is only offered when the seat has the energy.
  **Capped, twice.** Target and mode expansion is capped at 12 concrete moves per source card (`legal.Options.MaxExpansionPerSource`), so a spell with thirty legal targets ships twelve of them. For a "search your library for up to N" prompt, and for a `choose_cards` prompt, that budget is spread ACROSS pick sizes rather than spent smallest-first, so a card that fetches a pair still offers pairs when the library holds more matching cards than the cap (#544), and a "discard two unless you discard a creature card" prompt still offers pairs when most single cards break its rule (#624). On top of that the wire projection caps the whole list at 48 moves: past that it *degrades* rather than truncates, keeping the first move of every `(source, kind, targets_stack)` tuple and dropping only the alternatives. `targets_stack` rides in that key, not just `(source, kind)`, so a capped board still keeps a modal card's counter mode alongside its non-counter mode rather than silently keeping only one. The invariant a client may rely on is therefore **"every card that has a legal move is represented by at least one entry"** — never "this is the complete set of targets". Targeting UI reads `CardView.legal_targets`, not this field. Since [ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) neither cap is silent: `legal_moves_truncated` says when the wire cap dropped an alternative, a [`legal_moves_request`](#legal_moves_request-client--server--adr-0122-61) returns the list before it, and [the cut report](#the-cut-report-adr-0122-62) says which cards the enumerator's own caps cut and by how much.
  **`value`** ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §6.2, omitempty): `{ kind, min?, max? }`, on a move whose answer is an open set the rules state rather than one point of it. Only two answers are: `kind: "card_name"` on every answer to a `choose_card_name` prompt, whose `params.card_name` may be any card name (CR 201.2) — the server validates a free one as it does a person's: trimmed, non-empty, at most 200 characters; and `kind: "x"` on a `cast_spell` or `activate_ability` move whose `params.x_value` may be any whole number from `min` to `max`, both present. The move offers `max`. Every value in the set is accepted with the move's other `params` unchanged, which is why an X fixed by something else in the move — a target count, a "mana value X or less" target, a division, a sacrifice or tap count, a delve or waterbend payment sized to it — carries no `value`. Advice about the move, like `cost`; never part of the payload.
- **LegalActionsView** ([ADR 0105](decisions/0105-legal-action-highlights.md), #1789, omitempty): `GameView.legal_actions` is `{ pass?, sources?: { [instance_id]: { kinds, moves, abilities?, mana_abilities?, special_actions?, zones?, faces?, attack_targets?, exert_on_attack?, blocks?, cast_idle_hint?, truncated? } } }`, a per-card digest of the viewer's own legal moves. It is what the client's legal-action highlights read: which of your cards you can do something with right now, and what.
  **Built from the uncapped list.** The server folds the enumeration into the digest **before** the 48-move wire cap degrades `legal_moves`. One enumeration feeds both fields, so they cannot disagree about whether a card has a move. The digest also stays exact past the cap, down to the ability row: a board whose `legal_moves` keeps one move per `(source, kind, targets_stack)` still lists every live ability ref, zone and face here.
  **Fields.** `pass` is `true` when the list holds a `pass_priority` move; a defender declaring blockers holds no priority and has none. `sources` is keyed by card instance ID; each entry says:
  - `kinds`: the distinct move kinds for the card, in enumeration order (`land`, `cast`, `activate`, `mana`, `special_action`, `attack`, `block`). Never `pass`, `choice`, `mulligan` or `opening_roll`; choice, mulligan and opening-roll answers are not digested, because their surfaces read `pending_choices`, the mulligan window and `opening_roll`.
  - `moves`: how many moves in the uncapped list involve the card.
  - `abilities`: the ADR 0093 `ref` of each live `activate_ability` row. It joins to `activated_abilities[].ref` or `zone_abilities[].ref` on the same card in the same frame.
  - `mana_abilities`: the same for `activate_mana_ability`, joining to `mana_abilities[].ref` or `zone_mana_abilities[].ref`. A basic land's intrinsic ability is `land:<colour>`.
  - `special_actions`: the live special-action kinds (`foretell`, `suspend`, `turn_face_up`, `plot`).
  - `zones`: the `from_zone` of each cast or land move (`hand`, `command`, `graveyard`, `exile`, `library`).
  - `faces`: the printed faces the card's cast moves cast (0 is the front).
  - `attack_targets`: the players, planeswalkers and battles the creature may attack.
  - `exert_on_attack` ([ADR 0130](decisions/0130-exert.md) §5, omitempty): `true` when the creature may be exerted as it attacks right now (CR 701.43d). The enumerator then offers each of its attacks twice, the plain `declare_attacker` and the same one with `"exert": true` (labelled "… and exert it (it won't untap during your next untap step)"), so this card's `moves` counts both. The client offers the choice from this field and never from oracle text.
  - `blocks`: the attackers the creature may block. A grouped block (the two creatures a menace attacker takes, #750) is one move with the first blocker as its `source`, but every creature in the group gets an entry, and the move counts toward each one's `moves`.
  - `truncated` ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §6.2): the enumerator's cuts for this card, each `{ cap, omitted, at_least? }` as in [the cut report](#the-cut-report-adr-0122-62) (no `source`: the key is the card). Absent when nothing was cut, which is nearly every card. `moves` counts what survived the caps; this says the card has more, and a `legal_moves_request` naming the card returns them.
  - `cast_idle_hint` (#1918): present only when **every** cast move for the card carries an `idle_hint`, and set to the first one. The card is castable, but no cast it has would do anything right now. The client draws a muted ring and shows the hint as the tooltip. If any cast lacks the hint (Counterflux's targeted half with a spell to point at, say), this field is absent and the card highlights as an ordinary castable card.

  Every list is de-duplicated and in enumeration order. Sources are written in enumeration order too, not sorted by ID.
  **Own seat only, always**, by the same construction as `legal_moves`: the server builds one digest per seat into an unexported map, and `FilterViewFor` hands back only the viewer's own. The admin (empty viewer ID), spectators and replay readers get none, and the unfiltered `GameView` carries none. It reveals nothing `legal_moves` does not already give the same viewer.
  **Absent means "no information".** The field is present only while the seat owes a decision in which it can do something (priority, a combat declaration), and costs 0 bytes on every other frame. A frame whose only moves are choice or mulligan answers carries no digest either. A client must read an absent digest as "highlight nothing", never as "nothing is legal", and must not dim anything because of it. On the busiest four-player priority frame the tests measure, it is about 1.6–2.2 KB, against a 4 KiB budget (`TestLegalActionsFrameBudget`).
- **PendingChoiceView — #742 additions:** `kind` may be `choose_color` ("choose a color", CR 105.4). `color_options` carries the legal colours (a subset of `W`/`U`/`B`/`R`/`G` — "a color other than blue" is four entries, and colorless is never offered); the answer is `resolve_choice { choice_id, color }`. **#780:** such a prompt also carries `color_purpose` — `"mana"`, `"benefit"`, `"harm"`, `"filter"` or `"protect"` — the card's own declaration of what it will do with the colour it is handed. Public (it is a reading of the card's printed text) and present on `choose_color` only. It exists because every one of the five colours is a legal answer, so nothing else on the prompt distinguishes a good one. **#986:** it now drives three things rather than one — the bot seat's policy scores by it (`aiseat/heuristic`), the ORDER of `color_options` is ranked by it server-side (see the addendum below), and the client's picker WORDS itself from it (`colorPromptCopy` in `client/src/lib/manaPick.ts`): "Everything of the color you choose is hit, including your own permanents" for `harm`, "This produces mana of the color you choose" for `mana`. A prompt with no declared purpose keeps the neutral wording, and so does a purpose a client does not recognise — a card the catalog has not annotated must read as vague, never as the wrong question. A `mana_pick` may now carry `color_amounts?: { [color]: int }` — "N mana of any one color" (Gilded Lotus) is one pick that adds that many tokens of the picked colour, and the amount may differ per colour (Nyx Lotus adds your devotion to the colour you pick). A colour missing from the map adds one; the field is absent on ordinary picks. **#2558:** an "N mana of different colors" activation (Firemind Vessel) answered without colours up front queues ONE `mana_pick` that is answered N times: after each answer the entry stays in the queue under a NEW `id`, with the answered colour removed from `color_options` and named in `reason` ("… — color 2 of 2 (not white)"), and no mana is added until the last answer, when all N arrive together. Answering a colour already chosen is refused.
- **PendingChoiceView — ADR 0127 additions (automatic answers, #1961):** a prompt that can take a standing answer carries `auto_answer_key` (opaque, derived by the server: the catalog row that asked, then the prompt's kind and ordinal; a prompt asked from the branch of an answered one extends that one's key), `auto_answer_card` (the asking card's name) and `auto_answer_prompt` (the question) — the display copies the client keeps beside the rule in Settings. `asked_by_hand` (`"no_mana"`, `"empty_library"`, `"loop"` or `"undone"`) marks a prompt that has a rule but is asked anyway, and says why: "Always pay" that auto-pay cannot cover (CR 118.3), "Always" with an empty library, the CR 732 loop notice, or an automatic answer the chooser undid. Such a prompt is never answered automatically afterwards. All four are omitempty and **the chooser's alone**: `FilterViewFor` clears them for every other seat and spectators. **Covered** are a `pay_unless` paid in mana only, with no X or Phyrexian symbol, that guards no spell and is owed in no step; a `trigger_prompt` whose ability has no target clause, no modes and no trade and is not a miracle reveal; and a `confirm` with the default Yes / No labels and no `life_cost`. Every other prompt — ward, Mana Leak, an upkeep pay-or-else, echo, a discard or sacrifice payment, a waterbend, a targeted or modal trigger, a cast, a life payment, every pick, `commander_return` — has no key and is always asked.
- **Automatic answers** ([ADR 0127](decisions/0127-answering-repeated-prompts-for-you.md) §4, #1961): after every commit (an action, a bundle, a lobby step), the room answers, one commit each, every open prompt whose chooser is a human seat still in the game with a rule for its `auto_answer_key`, in queue order, while the opening roll and the mulligan are over and no `loop_notice` stands. Each answer is a commit of its own: its own `seq`, replay line and undo entry, stamped with the **chooser** and exempt from their undo budget, so the chooser may `undo` it while it is the top entry. Undoing it reopens the prompt with `asked_by_hand: "undone"`. Only the last frame is broadcast, so a client sees `seq` jump by one per automatic answer. "Always pay" pays only when auto-pay can cover the cost from the floating pool and the auto-tapper (with strict payment it spends real mana or nothing), and the prompt is asked by hand otherwise. An automatic answer is not a player decision, so it does not restart the CR 732 loop run. It works with no socket connected.
- **PendingChoiceView — #752 additions:** `trigger_prompt` and `pick_target` may carry `doubled_by?` and `doubled_by_name?`, identifying the public permanent whose effect caused this additional trigger under CR 603.2d. The fields are omitted when no trigger doubler applies; they do not change the choice response payload.
- **PendingChoiceView — #744 additions:** `kind: "coin_call"` carries `coins` (one by default), optional `allow_stop`, `wins` (wins so far), and `max_useful_wins` (a bot hint, zero means unspecified). The chooser answers `resolve_choice { choice_id, call: "heads" | "tails" | "stop" }`; `stop` is legal only when offered. One call covers every coin in that instruction. Illegal calls leave the prompt and random stream untouched. Undo rewinds the stream: changing the call changes the displayed faces, not whether those flips win.
- **PendingChoiceView — #804 additions (CR 732):** `kind: "loop_shortcut"` is the loop breaker's shortcut proposal, queued to the controller of the repeating ability at the moment `GameView.loop_notice` goes up. `reason` carries the ability's label (`"<card> — <ability>"`), `loop_count` is how many times it has already resolved this turn, and `loop_max_iterations` is the largest answer the engine accepts (1000). The chooser answers `resolve_choice { choice_id, iterations }` — the table then resolves that ability exactly `iterations` more times with the notice down and automatic passing live, and the breaker raises the notice and re-offers the prompt on the last of them. `iterations: 0` is "stop here": the prompt clears and the table stays exactly where the breaker left it, notice standing and automatic passing suspended. The dispatcher routes this kind **by the pending choice's `kind`**, since its whole payload is an integer whose most meaningful value is the zero value. Unlike the `loop_notice` it belongs to, this prompt **blocks the table** — `advance_step`, `pass_priority` and `pass_turn` are refused while it is open (ADR 0018 §6), because the shortcut is proposed with the loop's trigger still on the stack. It is never queued to a seat that has left or been eliminated.
- **PendingChoiceView — #568 additions (CR 608.2):** `kind: "option_pick"` is "choose one of the following", asked while a spell or ability RESOLVES and frequently addressed to a seat that is not the effect's controller — Torment of Hailfire's "lose 3 life unless that player sacrifices a nonland permanent of their choice or discards a card", and the pile a Fact or Fiction chooser takes. `pick_options` carries one entry per branch in the card's printed order: `{ label, cards?: CardView[], life_cost?: int, player?: uuid }`.
- **Modes "that hasn't been chosen" — #1749 additions ([ADR 0097](decisions/0097-modes-that-havent-been-chosen.md)):** a `mode_pick` prompt may carry `mode_used_options?: string[]`, `mode_used_indexes?: int[]` and `mode_not_chosen?: "this_turn" | "ever"`. The two lists are the bullets this object's ability has ALREADY chosen ("choose one that hasn't been chosen this turn" — Gala Greeters, Monument to Endurance; "choose one that hasn't been chosen" — Silent Hallcreeper), in printed order, parallel to each other. They sit BESIDE `mode_options` / `mode_indexes` rather than inside them, so `mode_indexes` still means "what may be answered": a client merges the two by index to show every bullet, renders a used one disabled with "already chosen this turn" or "already chosen" (from `mode_not_chosen`), and never sends one — the server refuses it with `bad_request` and leaves the prompt open. A prompt's offer can SHRINK while it is open: when several instances of one ability are waiting at once, answering one takes its bullet off the others, and an instance left with nothing to choose is withdrawn from `pending_choices`. The same restriction on an activated ability shows on `activated_abilities[i].modes`: `modes.not_chosen` is `"this_turn"` or `"ever"`, and `modes.options[j].used: true` marks a bullet already chosen, read from the memory the activation gate refuses from — `activate_ability` naming a used bullet is `bad_request` with nothing paid. All the new fields are absent for every other modal card. A spell never carries them.
- **Escalate — #2126 (CR 702.120a):** `CardView.modes.escalate?: { label, mana_cost?, discard_cards?, tap_creatures?, tap_options?: string[], max_extra }` is the cost an escalate spell charges for EACH mode beyond the first. `mana_cost` joins the total the cast owes (CR 601.2f) once per extra mode, like a Spree bullet's cost; `discard_cards` cards ride `discard_ids` and `tap_creatures` untapped creatures ride `teamwork_ids`, `(modes - 1)` times the per-mode count each, in a cast the server validates to the exact length (a payment list of the wrong length is `bad_request`, and so is any mode count the caster cannot pay). `modes.max` is already clamped to `1 + max_extra`, the most extra modes the viewer can pay the discard or tap half for, so the picker never offers a count the server refuses; mana is not asked at announce (CR 601.2g). `tap_options` and `max_extra` are the asking seat's answer and are absent from a bystander's public copy, which keeps the printed `label`, `mana_cost`, `discard_cards` and `tap_creatures`. No new action field: nothing is sent for one mode. Absent on every other modal card.

  **`player` (#994, CR 800.4a)** is the SEAT a branch is about, and is present only on the prompts whose branches ARE players — "choose a player" / "choose an opponent" (#929) and True-Name Nemesis's as-enters sibling (#980). Absent on every other option. It exists because an option has to be able to name its subject: the engine re-checks an open prompt whenever the board moves under it, and a seat that carried its identity only as rendered text in `label` could not be matched back to a player, so a prompt went on offering somebody who had left the game. **An option list can therefore SHRINK while it is open** — a client holding an index across a broadcast must re-read the list rather than reuse the offset, and the server answers on the option, not the number. Public by CR 400.2, so unlike `cards` it rides no redaction and every viewer sees the same seats. The chooser answers `resolve_choice { choice_id, option_index }` — the INDEX, because an option is a consequence and not always a set of cards. The dispatcher routes this kind **by the pending choice's `kind`**, since zero (the first option) is both the field's zero value and the commonest answer. Every option listed is one the engine will accept: legality is enforced when the prompt is queued, so an option the chooser cannot take is never offered, and the FIRST option is the one `legal_moves` marks `always_legal`. `life_cost` is the same declaration `confirm` makes (#547) — a client holding only this payload must be able to tell "lose 3 life" from "discard a card". This kind **blocks the table**: the effect that asked it is paused mid-resolution (ADR 0018 §6, 2026-09-18 amendment).

  **A pile split is two prompts, not a kind.** Fact or Fiction sends a `choose_cards` prompt to the SPLITTER (an opponent) with `choose_min: 0` — an empty pile is a legal split — and, from its answer, an `option_pick` back to the controller whose two options each carry a pile in `cards`.

  **Redaction of a prompt's cards (#568, #549, PR #513).** One rule, applied to `options[]` and to every `pick_options[i].cards`. A viewer who is not a knower of a card does not receive it at all — it is dropped, not sent as a back, because a stable `instance_id` for a card in a hidden zone is a correlation handle. The CHOOSER keeps unknown cards as answerable backs only when the pool is their own material, or when it is another player's HAND (`discard_from_hand`, `revealed_hand_pick`), whose size is already public (CR 400.2). A prompt addressed to one seat over ANOTHER seat's library shows that seat only the cards a reveal made public — so a card of this family must reveal before it asks, and every printed one does. A client must therefore expect an option with a label and no `cards`, and keep its button live: that is the redaction working, not a missing render.

- **Public random outcomes — #744, #1480:** `log` kinds `roll` and `flip` aggregate one instruction into one entry, even when listeners emit other events between individual results. Rolls carry `sides` and `results: number[]`; flips carry `faces: ("heads" | "tails")[]`, and called flips also carry `call` and `wins` (omitted when zero). `text` is the complete rendered line. Results are public; source names retain normal knowledge filtering. No RNG keys, stream counters or source ordinals reach the gameplay view. A production start emits one source-less d20 entry per contender, then another per tied leader until one wins; these entries precede the first turn and therefore omit `turn`. Since [ADR 0121](decisions/0121-animated-dice.md) the opening roll also writes `opening_roll` and `starting_player` lines, all before any hand is dealt. `opening_roll` carries `label`: `"tie"` (`seats` tied on `results[0]` and roll again — "Alice and Carol tied with 17 and roll again"), `"won"` (`seat` won with `results[0]` and chooses — "Carol won the opening roll with 20"; no `results` when the others left before it rolled) or `"rolled_for"` (`seat`, the host, `-1` for the server admin, rolled for `seats` — "Bob rolled for Carol and Dave"). `starting_player` has `seat` the chooser and `target_seat` the seat chosen: "Carol chose to take the first turn", "Carol chose Bob to take the first turn". A `table_roll` line is a die or a coin rolled at the table for fun (`roll_table_die`, ADR 0121 §5), one line per roll and never grouped: `sides` (6 or 20) and `results`, or `faces` for a coin with no `sides`, and `roll_id`, the roll's number in this game (1 for the first). "Alice rolled a d20 at the table: 14", "Alice flipped a coin at the table: heads". An undo of an action made before the roll does not erase the line: the server writes it again after the restored history, under a new `seq` and the same `roll_id`, so a client keys the roll on `roll_id` and does not show it twice. The client shows their winner persistently during opening hands, shows fresh in-game outcomes in the reveal strip, silently primes reconnect history, and removes rewound cues on undo.

- **PendingChoiceView.color_options order (2026-09-17, [ADR 0040](decisions/0040-mana-pipeline.md) addendum):** for a `mana_pick`, `color_options` is the source's printed colour set with the chooser's commander colour identity **listed first** (read from the chooser's commander wherever it is, so casting the commander does not change it), then the remaining colours, each group in printed (WUBRG) order. An "any color" source (Birds of Paradise, Treasure) under a mono-green commander sends `["G","W","U","B","R"]`. Nothing is narrowed away, except for a source whose printed text says "any color in your commander's color identity" (Command Tower, Arcane Signet, Commander's Sphere, Path of Ancestry), which sends only the identity's colours. Clients render the buttons in the order sent. Every listed colour is a legal answer. **#844 (CR 903.4f, 2026-09-17):** those four sources send NO `mana_pick` at all when the chooser has no commander, or a commander whose colour identity is colourless — the ability adds no mana, so there is nothing to pick. `color_options` is therefore never empty on a `mana_pick`; a client that receives an empty one should render no prompt. **#986 (2026-09-19, [ADR 0033](decisions/0033-ai-bot-seat.md) amendment):** a `choose_color` prompt's `color_options` is ordered too, by the prompt's own `color_purpose` — `harm` by what the opposition loses NET of what the chooser loses, `filter` by what the opponents have on the battlefield, `protect` by the greatest POWER among the creatures other seats control of that colour, and `mana` / `benefit` / undeclared by what the chooser already controls. Ties keep the list's own order (WUBRG, or the card's narrowed list). Every term is a count of permanents on the battlefield, so the order is public information and identical for every viewer, and it never narrows the list — all five (or all four of "a color other than blue") are still offered. One function does it, `legal.OrderColorOptionsLocked`, and the bot seat's move list is built from the same call, so the first button and the first enumerated answer are always the same colour. Clients render in the order sent, here as everywhere. A `choose_color`'s `color_options` is never empty either, and for a sharper reason than the `mana_pick` rule above: an unanswerable prompt BLOCKS, so a card that narrowed its list to nothing (`"C"` alone, a typo'd letter) would stall the table rather than ask a bad question. `normaliseColorOptions` now falls back to all five when the narrowing drops every entry — CR 105.4 says the answer is one of the five — and logs one warning per process naming the card bug.

S03 does not yet apply visibility filtering — every client receives every
card's contents, including opponents' hands. Per-connection visibility is
an S04 concern (ships alongside authentication).

Clients should treat every received snapshot as the new authoritative
state; no partial merge is needed.

---

## Connection lifecycle (v0)

1. Client opens WebSocket to `/ws`, authenticated as described under
   [Connection lifecycle (S04)](#connection-lifecycle-s04) above.
2. Server accepts, stages the initial `snapshot` frame, and admits the
   client to the room's broadcast set.
3. Client and server exchange `action` / `snapshot` / `chat` / `ping`
   frames until either side closes.
4. The socket closes. What happens next depends on the close CODE, not
   on whether the closure was "expected" — see the table below.

This section was a stub from before S04 added authentication, restated
here now that [ADR 0044](decisions/0044-surviving-a-deploy.md) (S33,
"surviving a deploy") gave the close path, the reconnect ladder, and
the rewind case above real behaviour to describe.

### Close codes: which ones redial

| Code | Sender | Meaning | Client behaviour |
|---|---|---|---|
| 1000 (normal closure) | `hub.EvictGame` ("game deleted"), or a session revocation ("session revoked" — [ADR 0051](decisions/0051-user-database.md) decision 6) | The game is gone, or this credential no longer works. Nothing on the other end of a redial. | **Terminal.** `isTerminalClose` (`client/src/lib/ws.ts`) is true only for 1000. The client stops and renders an ended state; the backoff ladder never starts. |
| 1001 (going away) | `hub.Shutdown`, on every `systemctl restart` | The process is restarting. [ADR 0041](decisions/0041-game-persistence.md) persists and restores the room across it — the game is emphatically not gone. | **Reconnect.** Joins the same ladder as 1006. |
| 1006 (abnormal closure) | the browser, for any closure with no close frame — a network blip, a crash, or (from the browser's point of view) a rejected upgrade | Unknown cause; may or may not resolve itself. | **Reconnect.** |
| 4001 | `hub.RebindUserSessions` ("admin mode changed" — [ADR 0112](decisions/0112-signed-in-home-player-mode-and-one-decks-page.md) §2 item 5), when an allowlisted person switches admin mode or it lapses | This socket's admin bit is now wrong. The server never flips it on a live connection; it closes the socket so the upgrade decides the binding again. | **Ask `GET /me` first**, then reconnect at once, or stop. A reconnect in player mode binds only the person's own seat or spectator session; anything else is refused at the upgrade, which the browser sees only as a 1006, so a blind redial would loop forever. After `/me` answers, the client reconnects when it is an admin or holds its own seat or spectator session for this game, and otherwise stops and goes to the Lobby ("Player mode is on; this table isn't yours"). If `/me` does not answer, it waits a rung of the backoff ladder and asks again, never dialling in between (`lib/ws.ts`, `CLOSE_ADMIN_MODE_CHANGED`). |

1001 and 1006 used to be indistinguishable in practice: `hub.Shutdown`
wrote the close frame and then called `Close()` immediately after, so
whether the browser observed 1001 or 1006 depended on a race between
that write reaching the peer and the FIN that followed it — the same
shutdown binary could produce either code from one deploy to the next.
`closeAndDrop` (`server/internal/ws/hub.go`, #518) now waits up to
`closeGracePeriod` (250ms) for the write to actually leave, or for the
peer's own close response to land, before tearing the connection down
— so a deploy is reliably 1001.

### The reconnect ladder

A close that is not terminal moves `ConnectionStatus` to `reconnecting`
and schedules a redial. `reconnectDelayMs(attempt)` (exported from
`client/src/lib/ws.ts`) is exponential backoff with half-jitter,
doubling from a `RECONNECT_BASE_MS` (500ms) nominal delay up to a
`RECONNECT_CAP_MS` ceiling (30s) — attempt *n*'s delay lands uniformly
in `[nominal/2, nominal]` where `nominal = min(cap, base * 2^n)`. There
is no cap on the number of *attempts*; the delay itself is what caps
the ladder's growth. Each socket open resets `reconnectAttempt` (the
store a "reconnecting (attempt N)" banner reads) back to 0.

**The dead-session check (#1475).** After `DEAD_SESSION_CHECK_AFTER_FAILURES`
(3) consecutive failed dials in one run, the client asks `GET /me` —
the cheapest authenticated route, echoing the principal — whether the
session it is holding is still good, once per run of failures. A
definite 401 moves `ConnectionStatus` to `session_ended`, a fifth and
final state, and the ladder stops for good: nothing further will make
a dead credential valid. Anything else the probe can report — a 5xx, a
network failure, a timeout — is treated as "still trying" and the
ladder is unaffected, because a server that is only restarting must
never be mistaken for a revoked session. The connection banner renders
`session_ended` as "Your session ended — sign in again", linking to
Login, or to My games (`GET /me/games`, `POST /me/games/{id}/session`)
for a signed-in identity.

### `seq` is non-decreasing only WITHIN a generation, and a change means rewind

Restated from the `snapshot` frame's field table above, because it is
part of the connection contract, not just a field note: `payload.seq`
is guaranteed monotonically non-decreasing only *within one restore
generation* (`payload.generation`). [ADR 0041](decisions/0041-game-persistence.md)
persists a room's state so it survives `systemctl restart`, but a
restore point is only ever written from a state with no live Go
continuations — a game that ran on through states that were never
restorable rewinds, on restart, to the last state that could be
rebuilt exactly. That restore point's `seq` can be — and after any
nontrivial game, usually is — lower than the `seq` a connected client
had already rendered. `payload.generation` names this: 0 for a room
that has never been rebuilt from disk, and incremented by one every
time it is. See [ADR 0044](decisions/0044-surviving-a-deploy.md)
decision 5.

A client tracks the generation of the last snapshot it rendered.
Within an unchanged generation, the ordinary guard applies: a `seq`
lower than the highest one already seen is a dropped or out-of-order
frame and is ignored. A **generation change is a different case
entirely** and bypasses that guard outright — the frame is accepted
and rendered regardless of its `seq`, because the server is not
reporting disorder, it is reporting that this room was just rebuilt
from an earlier point on purpose. The client discards whatever it was
tracking (its seq watermark, the rendered snapshot) and starts fresh
from the new frame.

**The rewind toast.** A generation change is not itself shown to the
player — most restores land at or after the last thing a connected
client had already rendered (a clean shutdown, or a game that never
walked into an unrestorable continuation), and that is not a rewind
worth mentioning. The toast fires only when the generation change
*and* the new `seq` is lower than the highest `seq` this client had
already rendered — a real rewind, visible cards or actions
disappearing from the board — with the exact copy:

> The table was restored to an earlier point after a server restart

See `client/src/lib/ws.ts` (`dispatchFrame`'s `"snapshot"` case) and
`client/src/lib/connectionBanner.ts` (`REWIND_NOTICE`).

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
- **Spectator mode** (S11) — read-only connections.

### Auto-tap preview (S15)

`GET /games/{id}/auto-tap-preview?card=<uuid>&from_zone=<zone>&alternative_cost=<key>&optional_costs=<i>,<i>&tap_ids=<uuid>,<uuid>&face=<int>&x=<int>&phyrexian=<int>&ability=<int>&targets=<kind>:<uuid>,...&exile_ids=<uuid>,...&waterbend_ids=<uuid>,...&exclude=<uuid>,<uuid>...`
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
| `cost_branch` | no | ADR 0100 — the branch of an either/or additional cost being paid, an index into `additional_cost.branches`. Its `mana_cost` joins the total, so Lightning Axe previews at `{R}` with the discard branch and at `{5}{R}` with the mana one. Omitted on a card with branches, the quote carries no branch mana. |
| `tap_ids` | no | Comma-separated permanent UUIDs being tapped for convoke or waterbend. They pay part of the cost, and the plan must not tap them again for mana. |
| `delve_ids` | no | ADR 0100 — comma-separated graveyard card UUIDs exiled to delve (CR 702.66a). Each pays `{1}` of the generic, so the plan covers only the rest. The response's `delve_budget` is how many the announcement may name. |
| `face` | no | The printed face being cast (ADR 0034). A modal DFC's back face has its own mana cost. Defaults to 0, the front. |
| `x` | no | Caller-supplied X value for spells with `{X}` in their cost. Defaults to 0. |
| `ability` | no | Price the card's CR 602 activated ability at this index instead of its cast cost. Every cast-shaped param above is ignored on this branch — an ability is not a cast — except `tap_ids`, which here names the permanents tapped for a TapOthers cost (#1422). #1405: priced by `Game.PriceActivation`, the function `ActivateCatalogAbility` charges, so board activation modifiers (Boom Scholar) and the ability's own cost clause (the channel lands, Dragonfire Blade) are in the plan and the `missing` list. No commander tax. |
| `targets` | no | #1405, with `ability` only — the targets the activation announces, as comma-separated `card:<uuid>` / `player:<uuid>` entries. A price that reads the target (Dragonfire Blade's "{1} less for each color of the creature it targets") is previewed at that target's price. Omitted, the no-target price, which is what the activation charges when it names no target and what the X and Phyrexian pickers want, since they open before targeting. Any other kind, or a malformed UUID, is a 400. Slot and mode are not carried; no cost reads them. |
| `sacrifice_ids` / `discard_ids` | no | #1242 — the permanents and cards the cast names to its additional cost's sacrifice and discard. They do not change the price — except a per-sacrifice discount, which reads the sacrifice count (ADR 0100 §3: Torgaar, Famine Incarnate previews at `{B}{B}` with three `sacrifice_ids`); the plan must not ALSO spend them on mana (an Eldrazi Spawn offered to Village Rites, a Spirit Guide offered to Thrill of Possibility), exactly as `CastSpell`'s auto-tap will not. The same holds for `tap_ids`, which the plan now excludes too. ADR 0135 §4 (owner decision 3): a permanent named to a SACRIFICE may still be tapped for mana first (CR 601.2g before 601.2h), so a Llanowar Elves offered to Village Rites can pay part of the cost; no mana ability that sacrifices it is planned. |
| `alt_cost_ids` | no | ADR 0135 §4 — the permanents and cards the cast names to its alternative cost's card component, as on `cast_spell`. An emerge creature changes the PRICE (its mana value comes off the generic part), so Elder Deep-Fiend previews at `{1}{U}{U}` over a four-drop. The plan never spends a named card or permanent, except that one named to a sacrifice offer may tap for mana first, as with `sacrifice_ids`. |
| `exile_ids` / `waterbend_ids` | no | #1422, with `ability` only — the cards named to an "exile N cards" cost and the permanents tapped to waterbend, under the `activate_ability` payload's field names. With `ability`, `tap_ids` / `sacrifice_ids` / `discard_ids` name the activation's own TapOthers, sacrifice and discard payments. The plan never spends any of them, nor the ability's own source when its cost has `{T}` or sacrifices it: the preview excludes `game.ActivationAutoTapExclusions`, the set `ActivateCatalogAbility`'s auto-tap excludes, so it cannot plan Castle Vantress's own `{U}` for its `{2}{U}{U}, {T}` ability. Each waterbend tap also pays `{1}`, subtracted with `game.WaterbendReduced` before the Phyrexian strike, the activation's own order. |
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
  "cost": "{2}{R}{R}",
  "delve_budget": 0
}
```

**ADR 0100:** `delve_budget` is how many graveyard cards the
announcement may exile to delve — `CastPrice.DelveBudget`, the same
number `cast_spell` refuses a longer `delve_ids` list against. It is
priced for the announcement as it stands (X, the claimed costs, the
convoke taps), so the client's delve picker reads its cap from here.
Absent (zero) for a card with no delve and on the `ability` branch.

**#2581:** `x_max` is the spell's printed "X can't be greater than
<count>" (Open the Way's players in the game, Winter's Chill's snow
lands), counted now for the asking seat — the largest X `cast_spell`
accepts, and the same number as the card's own `x_max`. Absent for a
card with no printed ceiling and on the `ability` branch; present at 0
when X = 0 is the only announcement.

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
- **A disturb offer casts the card's BACK face** (#1855, [ADR 0107
  §4](decisions/0107-state-triggers-rebound-disturb-and-damage-prevention.md),
  CR 702.146a, 712.11a). A disturb card in its owner's graveyard is
  stamped front face up, as it sits there (CR 712.8a), with
  `castable_here`, `alternative_cost_required: true` and one
  `alternative_costs` entry, `{ key: "disturb", label: "Disturb {1}{U}",
  mana_cost: "{1}{U}" }`. That entry's `target_mode` / `legal_targets`
  are the BACK face's clause — a disturbed Aura's "enchant creature" —
  because the spell is the back face (CR 712.8c); the front face's own
  clause is not used. `cast_spell` claims it with `from_zone:
  "graveyard"`, `alternative_cost: "disturb"` and `face` 0 or 1: the
  server turns the card over itself, and refuses any other face. No new
  field.
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
  `not_before_seq`) carries the grant and no stamps, because there is
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

  Since #1474 it is also clear on a card whose exit an effect has
  paused on its owner's CR 903.9b prompt: a commander an effect is
  shuffling out of a graveyard into its owner's library, for example.
  (Since ADR 0115 an exile, Bojuka Bog's included, no longer pauses.) `cast_spell` refuses that
  card with the pending-prompt error until the owner answers, so the
  bit says the same. No field was added.

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

**`exile_play.any_type` (#1573).** "Mana of any TYPE can be spent to
cast it" (Hostage Taker, Gonti, Night Minister, Outrageous Robbery) is
wider than `any_color`: colorless is a type and not a color (CR
106.1b), so a `{C}` in the cost is payable with any mana too. When it
is set, `any_color` is set alongside it, so a client that reads only
the older key still says "any colour". Both are labels only — the
prices in `cast_prices` and the `castable_here` bit already reflect
the fold, because the server prices every cast through the one
pricer. A Phyrexian symbol does not fold (#1589, CR 107.4f): under
either label it stays in the `cast_prices` string as printed —
`{1}{B/P}{B/P}`, not `{3}` — and still counts toward
`phyrexian_symbols`, so the cast may claim it for 2 life with
`phyrexian_life`. When it is paid with mana instead, any mana pays it.

**`phyrexian_granted` (ADR 0131, #2531, omitempty)** sits beside
`phyrexian_symbols` on `CardView`, on `alternative_costs[i]` and on
`activated_abilities[i]`. When the viewer controls a permanent that
lets them pay 2 life for each `{B}` in a cost (K'rrik, Son of
Yawgmoth), `phyrexian_symbols` counts those symbols too — it stays
the ceiling on `phyrexian_life` — and `phyrexian_granted` says how
many of them are the grant's rather than printed Phyrexian symbols.
It counts the `{B}` half of a hybrid symbol and never generic mana,
`{C}` or the black mana a "spend only black mana on X" clause adds.
The price in `mana_cost`, `cast_prices` and the activated-ability row
stays as printed. The server never pays a granted symbol with life
unless the action claims it with `phyrexian_life`, auto-tap included
(`auto_tap` plans the cost as mana), so a client asks only when mana
falls short and keeps an always-available "Pay life for {B}…" entry
for the player who wants it. The same claim works on `declare_attacker`
and `declare_attackers` for a `{B}` attack tax, and the bot's legal
moves carry it as `params.phyrexian_life` and `cost.phyrexian_life`.
Since PR 2 the pair is also on `mana_abilities[i]` (a mana ability's own
mana component, answered with `phyrexian_life` on `activate_mana_ability`)
and on a `pay_unless` entry of `pending_choices` (answered with
`phyrexian_life` on `resolve_choice`), counted for the chooser. A move
label for a payment that spends life now reads "paying N life instead of
mana", because the life may buy a printed `{B/P}` or a granted `{B}`.

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

**A sacrifice as the price (#1727).** An offer whose card component is
a sacrifice — Dread Return's "Flashback—Sacrifice three creatures",
Fireblast's "you may sacrifice two Mountains rather than pay this
spell's mana cost", Demon of Death's Gate's "pay 6 life and sacrifice
three black creatures" — carries `sacrifice_options` instead of
`pay_options`: the same `LegalTargetsView` the additional cost's
`sacrifice_options` is, built by the same function. Its `cards` are the
caster's OWN permanents that match the clause (CR 701.21a), in payment
order (tokens first, then lower mana value), and `min` == `max` is the
count. The client opens its sacrifice picker on it. The picks still
ride `cast_spell` as `alt_cost_ids` — they pay the offer, which
vanishes when the caster declines it — and never as `sacrifice_ids`,
which is the additional cost's list. The server demands exactly the
count, each permanent named once, each on the battlefield under the
caster's control and matching the clause, and refuses anything else
as `bad_request` with nothing paid; a permanent named to both
`alt_cost_ids` and `sacrifice_ids` is refused the same way (CR 118.3).
The permanents are sacrificed at CR 601.2h with the spell already on
the stack, so their dies triggers resolve first, and a countered spell
leaves them dead. An offer whose clause the caster's board cannot fill
is not stamped at all, which is how a flashback that cannot be paid
leaves its card not `castable_here`. Per viewer, like `pay_options`.

**A discard as the price (ADR 0135 §2, #2412).** An offer whose
`pay_options` cards are DISCARDED rather than exiled or returned —
retrace's land card, Snag's "You may discard a Forest card rather than
pay this spell's mana cost", Foil's "an Island card and another card",
The Infamous Cruelclaw's granted "by discarding a card" — carries
`discards: true` (public, with the rest of the printed offer). The
picks ride `alt_cost_ids` and are discarded at CR 601.2h with the spell
on the stack. An offer with a set rule (Foil) also carries
`pay_options.each_of`, in the shape the sacrifice picker reads: one
group per printed part with the cards that could fill it, and the picks
must fill every group with a different card. Two non-Island cards are
refused as `bad_request` with nothing paid, and an offer whose hand
cannot fill the groups is not stamped.

**A tap as the price (ADR 0135 §1, #2030).** An offer whose card half
TAPS permanents — Orim's Cure's "If you control a Plains, you may tap an
untapped creature you control rather than pay this spell's mana cost",
Battle Screech's "Flashback—Tap three untapped white creatures you
control", Zahid's "pay {3}{U} and tap an untapped artifact you control"
— carries `tap_options` instead of `pay_options` or `sacrifice_options`:
the block an ability's `tap_others_options` ships, the UNTAPPED
permanents the viewer controls that match, in payment order, with `min`
and `max` both the count. `pay_label` is the clause ("an untapped
creature you control"). The client opens the tap picker ("Tap <pay_label>
to cast <card>"), even when exactly the count is offered, and the picks
ride `alt_cost_ids`. They are tapped at CR 601.2h with the spell on the
stack. A cost does not target, so a hexproof creature is listed, and it
is not the {T} symbol, so a creature that arrived this turn is too (CR
302.6). A tapped, opponent's or non-matching permanent, the wrong count,
or a permanent also named in `tap_ids` or `teamwork_ids` on the same cast
(CR 118.3) is refused as `bad_request` with nothing paid. The auto-tapper
never taps a named permanent for the offer's mana. An offer whose board
can't pay the count is not stamped. Per viewer, like `pay_options`.

**Awaken (ADR 0135 §3, #2411, CR 702.113).** An awaken offer is keyed
`awaken`, labelled as printed ("Awaken 4—{5}{B}{B}"), and adds a target:
its statement is the spell's own clauses followed by "target land you
control" (CR 702.113b: the land is a target only when the awaken cost is
paid). When that statement has more than one clause the offer carries
`clauses`, one `LegalTargetsView` per clause in printed order with its
own legal set, exactly as a card's `clauses` does (#764), and the client
walks them in order; a spell with no target of its own (Coastal
Discovery) has the land alone, in `legal_targets`. `clauses` is per
viewer and stripped from a public pile with `legal_targets`. The land
clause does not require a different object from the spell's own, so
Earthen Arms may name one land twice (CR 601.2c). `purpose.awaken_land`
is the N. A cast for the mana cost that names the land is refused as
`bad_request`, and so is an awaken cast that names no land. On
resolution the land gets its N +1/+1 counters first and then becomes a
0/0 Elemental creature with haste that is still a land, with no
duration.

**Emerge (ADR 0135 §4, #2416, CR 702.119).** An emerge offer is keyed
`emerge` and labelled as printed ("Emerge {5}{U}{U}", or "Emerge from
artifact {5}{B}{B}" for CR 702.119b's variant). It is a sacrifice of one
permanent, so it carries `sacrifice_options` (`min` == `max` == 1) and
`pay_label` ("a creature", "an artifact"), and two more fields:
`reduces_by_mana_value: true`, and `sacrifice_prices`, keyed by the
instance IDs in `sacrifice_options`, each `{ mana_value, price }`: the
candidate's mana value (a token that isn't a copy is 0) and the mana the
cast owes with it sacrificed ("{1}{U}{U}" over a four-drop), priced by
the server's one pricer, commander tax and the board's cost modifiers
included. The reduction comes off the generic part only (CR 118.7a),
after any increase and never below {0} (CR 601.2f); `mana_cost` stays
the printed emerge cost. `sacrifice_prices` is per viewer and stripped
from a public pile with `sacrifice_options`. The pick rides `cast_spell`
as `alt_cost_ids`, exactly one permanent. The auto-tap preview
(`GET /games/{id}/auto-tap-preview`) takes
the same pick as `alt_cost_ids` on its query string, so its `cost`,
`plan` and `missing` are for the reduced price. A permanent named to any
sacrifice cost on a cast (`alt_cost_ids` under a sacrifice offer, or
`sacrifice_ids`) may still be tapped for mana by the auto-tapper before
it is sacrificed (CR 601.2g before 601.2h), but no mana ability that
sacrifices it is planned. The sacrifice happens at CR 601.2h with the
spell on the stack, and the stack item's paid-cost record keeps the
objects the alternative cost paid with (`PaidCost.AltCostObjects`, kept
by a CR 707.10 copy and by the permanent the spell becomes).

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

  `key` is not unique on a card. "Kicker {R} and/or {W}" (CR 702.33b,
  #2153) is two offers that both carry `key: "kicker"`, each with
  `max_times` 1 and its own `label` and `mana_cost`; claim either, both
  (`[0, 1]`) or neither by `index`. No field was added for it.

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
  buyback and most kickers. **#1716:** not only gift — a kicker,
  teamwork or blight offer carries the trio too when paying it widens
  the clause (Bloodchief's Thirst's kicked "target creature or
  planeswalker", Too Evil to Stay Dead's teamwork lifting "mana value 4
  or less"). No new field; a client that already takes its clause from
  the first ticked offer carrying one needs no change. Per viewer, like
  every legal set.

- **`StackItemView.gift_to`** (player ID string) is who a spell on the
  stack promised its gift to; absent when it promised none. Public: a
  responder needs it, because a promised Long River's Pull counters
  any spell and an unpromised one only a creature spell.

### Teamwork and blight (#1703, ADR 0073 amendment 2026-09-28)

Two more optional costs, keyed `"teamwork"` (CR 702.194a) and `"blight"`
(CR 701.68a), announced by index in `optional_costs` like a kicker. What
is new is only what the caster names while paying. All four fields are
additive; a client that ignores them never announces either cost and casts
the card unpaid.

- **`cast_spell.teamwork_ids`** (`["<uuid>", ...]`) are the creatures tapped
  for an announced teamwork cost: any number of the caster's untapped
  creatures whose total **effective** power (`power` on the card view) is at
  least the teamwork number. Summoning-sick creatures may pay — this is not
  the `{T}` symbol. Refused, never ignored, when sent without the teamwork
  index; refused with `the tapped creatures' total power is below the
  teamwork number` when they fall short (including none named); a tapped
  creature, an opponent's creature, a repeat or a creature also in `tap_ids`
  is `bad_request`. The creatures are tapped with the spell on the stack and
  are never spent by `auto_tap`.

- **`cast_spell.blight_ids`** (`["<uuid>"]`, exactly one) is the creature an
  announced blight puts its N -1/-1 counters on — a creature the caster
  controls. One that dies of the counters is a legal choice and still pays.
  The counters are a COST: a replacement worded "if an effect would put"
  (Doubling Season) does not apply (CR 614.16); one that names no effect
  (Winding Constrictor, Vizier of Remedies) does.

- **`OptionalCostView`** gains `teamwork` (N) with `teamwork_options` (the
  viewer's untapped creatures; present-and-empty when their positive powers
  together do not reach N, so the client greys the toggle), and `blight` (N)
  with `blight_options` (the viewer's creatures; present-and-empty when they
  control none, CR 701.68b).

- The auto-tap preview (`GET /games/{id}/auto-tap-preview`) accepts
  `teamwork_ids` and `blight_ids` on its query string, with the meaning above:
  the creatures they name are not planned for mana.

- **Blight X** (#2174, CR 701.68a / 107.3a): "As an additional cost to cast
  this spell, blight X" (Soul Immolation) ships as `additional_cost:
  { blight_x: true, blight_x_max, blight_options, demands_x: true, label }`.
  The announced `x_value` is the number of -1/-1 counters, put on the ONE
  creature named in `blight_ids` (at X = 0 the creature may be omitted).
  `blight_x_max` is the greatest toughness among the viewer's creatures —
  absent at 0 — and an `x_value` above it is refused. Additive within
  schema v7.

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

  Absent, which is every ordinary hand cast, every command-zone cast
  and a Gravecrawler whose graveyard permission carries no price,
  means the printed cost is on the menu as usual. This field is a
  ZONE-and-payability answer only — see the `printed_cost_timing_closed`
  amendment below for the orthogonal "is now the moment" question.

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

## The cost picker's own clock: `printed_cost_timing_closed` and `timing_closed` (#1686, 2026-09-28)

Two additive fields, and a fix to a HAND-zone gap the ones above never
covered: before this, `castStampsFor` (the function behind
`alternative_costs`, `alternative_cost_required` and everything else
in this section) was called for a card in HAND with a hard-coded
`nil` permission — a hand's own printed text never needed one before
S42's granted permissions existed, since CR 601.2 already opens the
hand. Miracle (#1665) is the first keyword that grants a HAND-zone
offer (`CastPermission{Zone: hand, AltCostKey: "miracle", Timing:
flash}`), and under a `nil` grant `CastOffersForLocked` refuses a
`RequiresGrant` offer outright — so a live miracle grant was invisible
on the wire, in EVERY zone. Fixed by looking the permission up for
hand and the command zone exactly as the graveyard and library
branches already did; nothing changes for a card with no grant, which
stays every card that isn't a live miracle.

With the grant now visible, the second gap: `alternative_cost_required`
and the offer LIST are a ZONE-and-payability answer, deliberately
blind to timing (CR 307.1) — the exile strip's informational
`cast_prices` badge reads the same list while a sorcery is between
windows on purpose (#1389), so it must not go empty just because now
is not the moment. A live miracle grant needed a genuinely SECOND
signal: it opens its own claim at instant speed (CR 608.2g) and says
nothing about the printed one, which stays whatever timing the card
prints — a sorcery for Terminus. Drawn on another player's turn (or,
just as often, in the card's own owner's DRAW step — sorcery speed
needs a MAIN phase, and the draw step is not one), the printed claim
is closed while the miracle claim stays open, and before this pair of
fields the picker had no way to tell the two apart.

- **`CardView.printed_cost_timing_closed`** (bool, omitted when false)
  says the printed mana cost IS one of the prices this cast may claim
  (`alternative_cost_required` is false) but the engine will refuse it
  RIGHT NOW for timing. Stamped only when the card also carries at
  least one entry in `alternative_costs` — the picker never opens over
  a card with nothing else to offer, so a plain sorcery gains no stamp
  here regardless of the step, and the overwhelming majority of cards
  in every frame are untouched.
- **`AlternativeCostView.timing_closed`** (bool, omitted when false) is
  the same question asked of ONE offer: the engine will refuse this
  specific claim right now even though its zone and payability both
  check out. Every S22 keyword (overload, evoke, cleave, flashback,
  escape, warp) answers to the card's own printed timing or a wider
  per-player grant the same way the printed cost does, so this stays
  false for them; miracle's claim is the one with a clock of its own
  independent of the card's.

Both are derived from `game.CastTimingOpenLocked`, under the grant
narrowed to the specific claim being asked about
(`CastPermission.ForClaim`) — the same narrowing `CastSpell` applies
before its own timing check, so the picker and the announce path
cannot disagree. Both are public, the same reasoning
`ActivatedAbilityView.timing_closed` already gives: whose turn it is,
what is on the stack and the battlefield's own grants are all public
facts.

**Read them in the cost picker, and nowhere else.** `printedCostCastableNow`
and `castableAlternativeCostsOf` (`targeting.ts`) are the two new
readers; `printedCostClaimable` and `alternativeCostsOf` are
unchanged and still answer the zone-and-payability question alone,
because the zone browser's button label and `canCastFromHand`'s
tooltip both want that question specifically — a card whose sorcery
window happens to be shut right now is not "uncastable from this
zone", and conflating the two would print the wrong sentence in a
tooltip that already has the right one from `legal_moves`.

Additive on the wire (`v` unchanged): a client that ignores both
fields behaves exactly as it did before this amendment, for every
card that doesn't carry a granted alternative cost with its own clock
— which today is every card except a live miracle.

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
| `alternative_costs[i].sacrifice_options` | the permanents that can pay a sacrifice price (#1727) | the asking seat |
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
  `not_before_seq`, and `castable_here` is the "now" half.
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
  | warp | the printed cost | yes | from `not_before_seq` |
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
(with `not_before_seq`) and neither `castable_here` nor
`cast_prices`: the engine would not accept the cast at any price. The
client shows such a card dimmed with a "next turn" hint and no badge.

## `cast_prices` in the command zone (#2202, 2026-10-04)

Additive, `v` unmoved. `cast_prices` is stamped on the cards in a
seat's own **command zone** too, for the commander's price tag in the
client's castable-from-other-zones strip. Same shape and meaning as in
exile: what `game.PriceCastForEffect` will charge for a cast with
`from_zone: "command"`, cheapest first, priced with no targets and
X = 0. The cost carries the commander tax (CR 903.8, `{2}` per earlier
cast from the command zone) and every CR 601.2f cost modifier on top
of it, so a taxed commander under a Medallion reads `{2}{G}` + `{4}`
tax − `{1}` = `{5}{G}`. `printed` is true only when there is no tax and
no modifier moved the cost; the client draws no tag for that price.

Per viewer, like the exile prices: the commander's owner gets them and
nobody else does, spectators included. `castable_here` is still never
stamped in the command zone; whether a commander can be cast now is the
legal-move digest's answer (`legal_actions`, zone `"command"`), as it
was before.

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

## Granted abilities and ability refs (ADR 0093, #754, 2026-09-24)

Additive, `v` unmoved. An effect can give ANOTHER permanent an ability
(CR 113.10) — Cryptolith Rite's "Creatures you control have '{T}: Add
one mana of any color.'" — and the ability is the recipient's: its
controller activates it, `{T}` taps the recipient, and CR 302.6 reads
the recipient's summoning sickness. Three things reach the wire.

- **`ref` on every ability row.** Each `activated_abilities[i]`,
  `zone_abilities[i]`, `mana_abilities[i]` and `zone_mana_abilities[i]`
  carries `ref: string`, the row's stable name:
  - `own:<i>` — the permanent's own ability, counted in its full
    declared list (a designation gate opening does not renumber it);
  - `land:<colour>` — a CR 305.6 intrinsic land ability (`land:G`);
  - `grant:<bundle>:<i>:<n>` — ability `i` of the `n`th instance of a
    granted bundle.

  Granted rows always come AFTER the permanent's own and intrinsic rows,
  so a grant appearing never renumbers an existing row; a grant
  vanishing can renumber later ones. `activate_ability` and
  `activate_mana_ability` accept `ref` beside `ability_index` and check
  it against the row at that index BEFORE anything is validated or
  paid. A mismatch is refused with `that ability is no longer at that
  position — the board changed`, and the next snapshot carries fresh
  refs. An absent `ref` is accepted as before. The bot's legal moves
  carry `ref` in their `params`, so a seat that sends a move back
  unaltered sends it too.
- **`granted_by: {id, name}` on a granted row.** Absent on the
  permanent's own and intrinsic rows. `name` is the grantor's name when
  it is on the battlefield. A client's left-click counts a usable
  granted activated row like any other row: it activates it when it is
  the card's one usable row, and opens the ability popover when the
  card has two or more. A granted mana row goes to the mana picker, and
  the click never picks a silent default between two mana abilities
  (ADR 0093 Decision 8, as ADR 0117 §1 and its 2026-10-04 amendment
  amend it).
- **`granted_abilities: [{text, source_id?, source_name?}]` on the
  CardView.** The printed text of every ability another effect gave the
  permanent — layer-6 grants, and a copy's CR 707.9a grant (Phantasmal
  Image's sacrifice trigger), which has no `source_id`. This is the only
  place a granted TRIGGER, which has no row, is visible. Public, and
  cleared with the ability rows for a viewer who cannot see the card
  (a face-down permanent's non-controllers).

Identical grants are not merged: two grantors give two rows, each
labelled with its grantor (CR 113.2c; the working default until the
owner settles ADR 0093's open presentation question).

## Abilities any player may activate (ADR 0106 §1, #1793, 2026-10-01)

Additive, `v` unmoved. CR 602.2 lets an ability say who may activate
it, and "Any player may activate this ability" (Xantcha, Sleeper Agent,
Feral Hydra, Excavation) says everyone.

- **`any_player: true` on the `activated_abilities[i]` row.** Absent on
  every other row. A non-controller may send `activate_ability` for that
  row (`ability_index` and `ref` as usual); the server refuses any other
  row of a permanent the caller does not control with `game: caller does
  not control this card`. The activator pays the cost from their own
  pool, hand and board (CR 602.1a), the ability on the stack is theirs
  (CR 602.2a, 113.8), and "you" in its text is the activator (CR 109.5).
- **`opponents_only: true` / `owner_only: true` on the row** (2026-10-07,
  #1947). "Only your opponents may activate this ability" (Clergy of the
  Holy Nimbus) and "Only this creature's owner may activate this ability"
  (Personal Incarnation). Additive, absent on every other row. The
  server refuses the controller on an opponents-only row, and anyone but
  the owner on an owner-only one, with `game: caller does not control
  this card`. The row rides the per-seat copy below exactly as an
  `any_player` row does; the controller's own opponents-only row, and a
  thief's owner-only row, are the client's to grey from the card's
  `controller` and `owner`. `purpose` is never declared on these rows.
- **`grantor_only: true` and `activator` on a granted row** (2026-10-09,
  #1947). "Only you may activate this ability" inside an ability another
  effect granted (Martyrdom). `activator` is the player ID of the one
  player who may activate it: the "you" of the granting effect, whoever
  controls the permanent now. Additive, absent on every other row. The
  server refuses anyone else, the permanent's controller included, with
  `game: caller does not control this card`. The row rides the per-seat
  copy as an `any_player` row does; the client opens it to the viewer
  whose ID is `activator` and greys it for everyone else.
- **Each seat's copy is stamped for that seat.** For a seat that does
  not control the permanent, every `any_player` row is computed with that
  seat as the activator: `charged_mana_cost`, `life_cost`,
  `condition_unmet`, `timing_closed`, `exhausted`, `cant_activate`, and
  the cost-option lists (`sacrifice_options`, `discard_cost_options`, …)
  out of that seat's own board and hand. A hand-read list reaches its
  own seat alone, through the #1369 per-seat carrier; spectators get the
  controller's public row. The other rows on that copy are the
  controller's public ones.
- **`purpose: {draws?, controller_loses_life?}` on such a row.** What the
  row buys an activator who does not control the permanent, declared by
  the catalog (ADR 0106 §1 decision 8). The bot reads it; a client may
  ignore it. Since ADR 0126 the field is the one `purpose` object below,
  and may appear on any row; `controller_loses_life` is still declared
  on `any_player` rows alone.
- **The digest.** `legal_actions.sources` (ADR 0105) is built from the
  viewer's own move list, so another player's permanent appears there,
  for the viewer alone, exactly when the viewer may activate one of its
  `any_player` rows right now: `{"kinds": ["activate"], "abilities":
  ["own:0"], …}`. The move's `label` names the controller ("Xantcha,
  Sleeper Agent (controlled by Alice): …").
- **The log.** `activate_across` is a new `log` kind: "Bob activated
  Alice's Xantcha, Sleeper Agent", with `target_seat` the controller.
  An activation by the permanent's own controller is not narrated, as
  before.

## Declared purpose: what a spell or an ability does (ADR 0126 §6, 2026-10-06)

Additive, `v` unmoved, no snapshot change. A `purpose` object says what
a card, a mode, an alternative cost or an ability does, as printed
amounts. It is declared by hand on the card file in the catalog, like
the completeness mark, and never derived from oracle text at run time.
The engine never reads it: it is there for the heuristic bot, which
reads the view and nothing else (ADR 0033 §3), and could not otherwise
tell Wrath of God from Divination. A client may ignore it. A card or
row that declares nothing has no `purpose` key at all.

```json
"purpose": {
  "draws": 2, "controller_loses_life": 2, "discards": 1, "lands": 2,
  "tutors": 1, "self_mill_tutor": 1, "tokens": 2, "energy": 2,
  "sweep": {"matches": "creatures", "how": "destroy", "amount": 3,
            "amount_is_x": true, "opponents_only": true, "partial": true},
  "death_payoff": true,
  "discard_payoff": {"types": ["island", "pirate", "vehicle"],
                     "tokens": 1, "counters": 1, "damage_each_opponent": 1}
}
```

Every field is omitted when zero.

| Field | Meaning | Example |
|---|---|---|
| `draws` | cards its controller draws (on an `any_player` row, the activator) | Night's Whisper 2 |
| `controller_loses_life` | life the source's controller loses; `any_player` activated rows only (ADR 0106) | Xantcha 2 |
| `discards` | cards its controller discards on resolution (a discard paid as a cost is `additional_cost`'s) | Faithless Looting 2 |
| `lands` | land cards it puts onto the battlefield | Harrow 2 |
| `lands_untapped` | how many of those `lands` enter untapped, so their mana can be spent the turn it resolves ([ADR 0136](decisions/0136-planning-the-turns-mana.md) §2). Never more than `lands`; absent when they enter tapped. A land with its own enters-tapped clause still enters tapped | Harrow 2, Nature's Lore 1 |
| `tutors` | cards it searches out to hand or to the top of the library | Demonic Tutor 1 |
| `self_mill_tutor` | cards it searches out into the graveyard | Entomb 1 |
| `tokens` | tokens it creates for its controller | Big Score 2 |
| `energy` | energy counters it gives its controller ([ADR 0129](decisions/0129-energy-getting-and-paying-it.md) §7) | Glimmer of Genius 2 |
| `sweep` | present on a board wipe; see below | Wrath of God |
| `death_payoff` | on a triggered row: it pays out whenever a creature its controller controls dies | Blood Artist |
| `discard_payoff` | on a triggered row: it pays out whenever its controller discards a card it matches; see below | Mary Read and Anne Bonny |
| `pump` | on a triggered or activated row: `{power, toughness, keywords}` it gives its own source until end of turn ([ADR 0130](decisions/0130-exert.md), amendment of 2026-10-07) | an exert self-pump |
| `extra_combat` | additional combat phases it adds | Combat Celebrant 1 |
| `prevent_combat_damage_to_self` | on a triggered or activated row: it prevents all combat damage that would be dealt to its source this turn | Oketra's Avenger |
| `damage_to_creature` | damage it deals to one target creature | Glorybringer 4 |
| `damage_each_opponent` | damage it deals to each opponent | Resolute Survivors 1 |
| `life_gain` | life its controller gains | Resolute Survivors 1 |
| `awaken_land` | on an alternative cost only: the N of "Awaken N—[cost]", the +1/+1 counters it puts on a land its controller controls as that land becomes a 0/0 Elemental creature with haste ([ADR 0135](decisions/0135-alternative-costs-that-tap-discard-awaken-and-emerge.md) §3). The spell's own `purpose` still applies | Ruinous Path 4 |
| `extra_land_drops` | on the card only: the additional lands its controller may play, on each of their turns for a permanent and this turn for an instant or sorcery ([#2678](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2678)). On a permanent it is the same number as the static the engine runs | Oracle of Mul Daya 1 |
| `targets` | what happens to each target, one entry per target clause; see [Target entries](#target-entries-adr-0126-amendment-of-2026-10-08) | Sign in Blood |

An amount is the printed number. A card whose amount is X, or is
counted at resolution ("draw a card for each creature you control"),
does not declare it.

`sweep` describes the permanents a wipe removes:

- `matches`: `creatures`, `nonland_permanents`, `artifacts`,
  `enchantments`, `artifacts_and_enchantments`, `all_permanents`,
  `creatures_mana_value_3_or_less`, `creatures_mana_value_4_or_greater`.
- `how`: `destroy`, `exile`, `bounce` (to the owners' hands), `damage`,
  `minus` (−N/−N) or `sacrifice`.
- `amount`: the damage or the N, for `damage` and `minus`. Absent with
  `amount_is_x`, when the amount is the spell's or ability's X.
- `opponents_only`: only permanents its controller's opponents control
  ("creatures your opponents control", "target player controls").
- `partial`: some permanents of the class are spared by a condition
  `matches` does not name (nonwhite, without flying, power 4 or
  greater), so `matches` is an upper bound.

`discard_payoff` (ADR 0126's amendment of 2026-10-06) says which
discarded cards a triggered row pays on, and what it pays for each one:

- `any`: every card its controller discards ("a card", "one or more
  cards"). Otherwise `types` lists the card types and subtypes it pays
  on, lowercase as printed (`["island", "pirate", "vehicle"]`); a card
  with any one of them on its type line matches. A payoff has one of the
  two, never both.
- `tokens`: tokens it creates for its controller per card (Mary Read's
  tapped Treasure).
- `counters`: +1/+1 counters it puts on its source per card (Marauding
  Mako).
- `damage_each_opponent`: damage its source deals to each opponent per
  card (Glint-Horn Buccaneer).

It rides on `ability_rows[i].purpose` with the rest of the row, so a
hidden hand card and a face-down permanent never carry one.

Where it rides:

- **`CardView.purpose`**: the spell's resolution, or a permanent's
  enters effect. On the hand, the battlefield and the stack. Public on a
  card the viewer can see, and cleared with `ability_rows` for a viewer
  who cannot, so an opponent's hand card and a face-down object never
  carry one.
- **`modes.options[i].purpose`**: one bullet of a modal spell or
  ability (Farewell's "Exile all creatures."). A modal card declares its
  purposes here rather than on the card.
- **`alternative_costs[i].purpose`**: what the spell does when cast for
  that cost, where that differs from the card's own (overload turns
  Cyclonic Rift into a bounce sweep).
- **`activated_abilities[i].purpose`**: an activated row's own (a loot,
  Nevinyrral's Disk). It is no longer only on `any_player` rows.
- **`ability_rows[i].purpose`**: a triggered or activated row's, on the
  tile's row list (`death_payoff` on Blood Artist's trigger,
  `discard_payoff` on Mary Read and Anne Bonny's).

**`ability_rows[i].exert`** ([ADR 0130](decisions/0130-exert.md),
amendment of 2026-10-07) marks a triggered row as one of exert's:
`"linked"` on the "when you do" trigger linked to "You may exert this
creature as it attacks", and `"payoff"` on "Whenever you exert a
creature". Absent on every other row. The heuristic prices an exert by
these rows' purposes; a client may ignore it.

The modes and the offers are public with their labels, which say the
same thing in words, and hidden exactly when those are.

### Target entries (ADR 0126, amendment of 2026-10-08)

Additive, `v` unmoved, no snapshot change. Every amount above is the
**controller's**. `targets` says what happens **to each target** the
statement names, one entry per target clause, so a card whose value
depends on whom it is aimed at can say so: "Target player draws two
cards" gives two cards to whoever the move picks, which may be an
opponent.

```json
"purpose": {
  "targets": [{"slot": 0, "draws": 2, "life_loss": 2}]
}
```

| Field | Meaning | Example |
|---|---|---|
| `slot` | the target clause the entry describes, its index in the statement this `purpose` rides on (the card's own clause list, a mode's, an alternative cost's or an ability row's). It is the `slot` a move's `targets[i]` names, and on a modal cast `targets[i].mode` says which chosen bullet's statement that is. Always sent | 0 |
| `draws` | cards the target player draws | Sign in Blood 2 |
| `discards` | cards the target player discards on resolution | Prismari Command's loot 2 |
| `tokens` | tokens the target player creates | Prismari Command's Treasure 1 |
| `life_gain` | life the target player gains | |
| `life_loss` | life the target player loses (not damage) | Sign in Blood 2 |
| `damage` | damage dealt to the target, player or permanent | Lightning Bolt 3 |
| `returns` | what the target's **controller** is given when the target is removed (#2679), an object below; absent when nothing comes back | Rapid Hybridization's 3/3 |

`returns` holds printed amounts, each omitted when zero:

| Field | Meaning | Example |
|---|---|---|
| `creature_tokens` | creature tokens the target's controller creates | Rapid Hybridization, Beast Within, Stroke of Midnight 1 |
| `token_power`, `token_toughness` | each such token's printed power and toughness | 3 and 3; Stroke of Midnight 1 and 1 |
| `life_equal_to_power` | the target's controller gains life equal to the target's power, counted at resolution | Swords to Plowshares |
| `lands` | land cards the target's controller may put onto the battlefield | Path to Exile, Assassin's Trophy 1 |
| `lands_untapped` | how many of those enter untapped | Assassin's Trophy 1 |

Every amount but `slot` is omitted when zero, and `targets` is absent
when there are none. A clause with no entry says nothing about its
target. A `purpose` holding only `targets` is still sent. The server
refuses at boot an entry whose `slot` is not one of its statement's
clauses, one that names a slot twice or says nothing, a player amount on
a clause that cannot target a player, and `damage` on a clause that can
target neither a player nor a permanent, and a `returns` on a clause
that cannot target a permanent, creature tokens with no printed
toughness, or more untapped lands than lands. An amount is the printed
number: Blaze's X declares no entry. `damage_to_creature` above stays
for a row whose target is picked later, where no move names it.

## An equip row says it is one (#2449, 2026-10-07)

Additive, `v` unmoved.

- **`equip: true` on an `activated_abilities[i]` row** that is a CR
  702.6 equip ability: the row `effects.EquipAbility` builds, the same
  bit Leonin Shikari reads (#1208). Absent on every other row.

It is bot data. A bot's policy reads it to tell an equip from any other
"target creature you control" row, because moving an Equipment between
two of its own creatures buys only what the new host gains, and a free
equip priced like a pump was moved back and forth forever (#2449). The
client does not read it. Public with the row.

## A self-untap row says it is one (#2500, 2026-10-08)

Additive, `v` unmoved.

- **`untap_self: true` on an `activated_abilities[i]` row** whose whole
  effect is untapping its own source: Basalt Monolith's "{3}: Untap",
  Grim Monolith's "{4}: Untap". Declared by hand on the card file
  (`game.ActivatedAbility.UntapSelf`). Absent on every other row.

It is bot data. With the source's `mana_abilities` on the same card, a
policy can tell a self-untap that nets mana from one that only trades it,
and the heuristic prices the latter below passing so a Monolith is not
untapped in a loop (#2500). The client does not read it. Public with the
row.

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
