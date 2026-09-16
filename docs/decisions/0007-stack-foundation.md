# ADR 0007 — Stack foundation (S13.1)

**Status:** Accepted · 2026-04-21 · Sprint S13.1

## Context

S13 landed the priority + turn-based-action shape, but every "play a
card" still bypassed the stack: `play_card` moved hand → battlefield
in one mutation. Counters, instant-speed responses, hold-priority,
split-second, and the political stack-timing that defines Commander
all require the gap between "I cast this" and "it resolves." S13.1's
goal was to land that gap end-to-end — including activated abilities,
manual triggers with APNAP ordering, the four canonical Commander
SBAs, the per-commander cast tax, commander zone replacement, and
leaving-game stack cleanup.

Sub-sprint scope deliberately **excluded** the card-data layer:
auto-fire of triggered abilities from oracle text, auto-validation of
target legality at announce, auto-resolution of spell effects, the
mana pool, the replacement-effect engine generally, static-ability
continuous effects, and combat keywords. Those depend on a per-card
effect catalog (S14+).

## Decisions

### 1. Dual representation: `Game.Stack *Zone` + `Game.StackMeta map[uuid.UUID]*StackItem`

**Decision:** Keep the existing `Stack *Zone` zone holding spell
cards as the source of truth for "which cards are on the stack."
Add a parallel `StackMeta` map keyed by the item's ID for everything
else: kind, controller, owner, source card (for abilities), label,
targets, modes, X, distribution, hold-priority, split-second.

**Why dual representation vs. replacing `Stack *Zone` with a new
`StackZone` value type:**
- Snapshot, clone, restore, WS broadcast, and `FilterViewFor` all
  already understand `*Zone`. Replacing the zone shape would have
  meant touching every single one of those code paths for a
  non-feature change.
- For ability items there's no card on the stack (the source card
  stays put), so a single-shape "stack of cards" wouldn't have fit
  abilities anyway. The map gives us a uniform place for both.
- The wire surfaces both: existing `stack: ZoneView` keeps working
  for legacy clients; new `stack_items: StackItemView[]` is the
  primary cast/resolve UI surface and joins the two on `id`.

### 2. CastSpell as the new canonical play verb; PlayCard kept

**Decision:** `cast_spell` becomes the canonical "play a card from
hand" verb (CR 601). Lands skip the stack (CR 305 special action) and
route directly to battlefield. Other types go to the stack. The
caster retains priority (CR 117.3c) — `cast_spell` does not rotate
priority. Sorcery-speed gate (sorceries + lands) requires main phase
+ stack empty + active player. Instants bypass the gate.

`play_card` stays as the sandbox / admin direct-drop verb: token
creation, replay restore, "fix wedged state" cases. Two verbs over
one because casual play sometimes needs the direct drop and replays
predate the cast verb.

### 3. PassPriority resolves top-of-stack on wrap

**Decision:** Pre-S13.1, `PassPriority`'s wrap branch unconditionally
advanced the step. Post-S13.1, the wrap branch checks
`stackHasItemsLocked()`: non-empty → `resolveTopOfStackLocked()` runs
and priority returns to the active player at the same step (CR
117.3b / 608.1); empty → step advance (existing behaviour).
Eliminated-seat skip from S13 is preserved.

### 4. Target re-check at resolution (CR 608.2b), no announce-time validation

**Decision:** Targets are captured at announce time as `[]TargetRef`
with kinds `player / card / self / none`. The engine does NOT
validate legality at announce — sandbox players know what's
targetable. At resolution, the existence-only re-check fires:
- If the spell declared at least one targeted slot AND every targeted
  slot is now illegal (player no longer seated, card no longer in
  any tracked zone), the spell is "countered by game rules" and
  routes to the owner's graveyard without effect.
- If at least one target survives, the spell resolves normally;
  partial-target effects are resolved manually by the players (no
  auto-trim of the surviving subset).
- `Self` and `None` are fixed references and always-legal; they
  don't count toward the "had any targeted slot" signal.

### 5. APNAP queue for triggered abilities; no auto-fire from card events

**Decision:** Per CR 603.3b, multiple triggers waiting at a priority-
grant boundary must drain in active-player-non-active-player order,
with each affected player choosing the relative order of their own
simultaneous triggers. The implementation queues into
`Game.PendingTriggers` via `AnnounceTrigger`, drains via
`drainPendingTriggersAPNAPLocked()` at every priority-grant
boundary inside `runStateChecksLocked()`.

**Manual-only by design.** Auto-fire from card events (oracle text
parsing, "this triggers when X happens") needs a card-effect catalog
and is the bright line between S13.1 and S14+. Manual announcement
keeps the engine honest about what it knows: priority + ordering +
re-check, NOT effects.

### 6. Loyalty abilities resolve immediately (sandbox); proper stack item deferred

**Decision:** `ActivateLoyalty` applies the loyalty delta directly to
the planeswalker's `loyalty` counter and sets the once-per-turn flag
on `Game.LoyaltyActivatedThisTurn`. Sorcery-speed gated.

**Why immediate vs. proper stack item:** A real stack item means an
opponent could counter the loyalty ability. That's CR-correct but
sandbox-overkill — counters of loyalty abilities are vanishingly
rare in casual Commander, and the proper-stack version requires the
player to also choose the announce order with other simultaneous
abilities. The plan defers proper-stack loyalty to S14+ if the
playgroup actually exercises it.

### 7. Five SBAs in one loop, with the placeholder-creature exemption

**Decision:** `stateBasedActionsLocked()` runs one pass of all five
canonical CR 704.5 SBAs:
- 704.5a: 0 or less life → eliminated
- 704.5b: tried to draw from empty library → eliminated (flag set
  in `drawCardLocked`, drained on next SBA pass)
- 704.5f: creature with current toughness ≤ 0 → destroyed
- 704.5g: creature with damage marked ≥ current toughness →
  destroyed
- 704.6c / 903.10a: 21 commander damage from a single source →
  eliminated

`runStateChecksLocked()` is the loop wrapper that pairs SBAs with
APNAP trigger drain (CR 704.4 + 603.3b) and runs at every priority-
grant boundary. Bounded at 32 iterations as a safety belt against
unintended SBA / trigger ping-pong.

**Placeholder-creature exemption.** A creature with `Toughness == 0`
AND no `Counters` is treated as "stats unparseable" (matches the
existing `Card.Power` documentation note about Mortivore-style "*"
cards and the demo seed) and is NOT destroyed by the toughness-0
SBA. A printed >0 creature reduced to ≤0 by counters does die. New
`Card.CurrentToughness` mirror of `CurrentPower` does the math.

Counter-specific SBAs (planeswalker loyalty 0, battle defense 0,
+1/+1 -1/-1 cancel, poison ≥ 10, saga final chapter) belong to
S13.2 and inherit the same loop.

### 8. Per-commander cast tax (CR 903.8); per-commander damage deferred

**Decision:** New `Player.CommanderCasts map[uuid.UUID]int` keyed by
the commander's instance ID. `CastSpell` increments
`CommanderCasts[id]` when `FromZone == "command"`. Sandbox: the
engine doesn't enforce the +{2}-per-cast surcharge; the counter is
the affordance for players to track the tax themselves.

**Per-commander damage** (the partner-pair fix the prior plan
flagged) stays on the existing per-opponent map for now — the
partner-pair playtest is the thing that would force the migration,
and it's not on the table yet. The `CommanderCasts` map proves the
keyed-by-instance shape works; damage can adopt the same shape when
needed.

### 9. Commander zone replacement (CR 903.9) as an explicit player choice

**Decision:** `move_card` action gains an `as_commander: bool` flag.
When true AND the card is a commander AND the destination is
graveyard / exile / hand / library, the move is rewritten to the
owner's command zone. Non-commander cards with the flag set keep
their requested destination.

**Why explicit-flag vs. automatic engine transform:** CR 903.9 is a
"may" replacement effect — the owner *chooses* whether to route to
command. Automatic redirect would force the choice. The flag makes
the choice the player's, surfaced via the move-card menu's "send to
command zone" option.

### 10. Leaving-game cleanup: spells exile, abilities + triggers vanish, targets re-check

**Decision:** `cleanupStackForEliminatedLocked` runs as part of
`eliminatePlayerLocked` (used by both `Concede` and SBA-driven
eliminations). Spell items move from `Game.Stack` to `Game.Exile`
(closest sandbox analogue to CR 800.4a "cease to exist"); ability
items + pending triggers controlled by the leaver disappear.

**Targets are NOT scrubbed** at cleanup time. The existing target
re-check at resolution (decision 4) already turns those slots
illegal and counters-by-game-rules at the moment the mechanic
actually matters. Less work, same outcome.

## Consequences

- Cast / resolve / counter / SBA / APNAP all work end-to-end with
  full announce-time data capture. Lightning Bolt → stack with target
  → both pass priority → SBA loop fires → resolves to graveyard. Same
  loop handles Counterspell.
- The wire grew `stack_items[]`, `pending_triggers[]`,
  `split_second_active`, `commander_casts{}`, `damage_marked`. All
  optional / omitempty so legacy clients keep working — they just
  miss the new affordances.
- `runStateChecksLocked()` is now the canonical "things changed,
  re-check the world" routine. Future feature work that mutates
  battlefield / life / library state should call it once at the end
  of the mutation; the loop is bounded and safe to invoke liberally.
- The placeholder-creature exemption is a documented contract
  pinned by `TestS131SBAPlaceholderCreatureSurvives`; future
  refactors of the SBA loop must preserve it.
- Heavier announce-time UX (CastDialog with target picker / X input,
  AbilityDialog for `activate_ability` + `announce_trigger` flows,
  inline mark-damage on creature tiles) is follow-up work — the wire
  shape already accepts the data, so the basic cast loop and
  counter affordance are fully testable today.
- `play_card` stays for sandbox / admin / replay; a future cleanup
  could remove it once we're confident nothing depends on the
  bypass-the-stack behaviour.

## Out of scope (S14+)

- Auto-fire of triggered abilities from card events
- Auto-validation of target legality at announce
- Auto-resolution of spell effects
- Mana pool modeling (CR 106)
- Replacement-effect engine generally (CR 614)
- Static-ability continuous effects (CR 604)
- Combat keyword effects
- Per-commander damage map (deferred until partner-pair playtest)
