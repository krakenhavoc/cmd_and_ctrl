# ADR 0007 — Stack foundation (S13.1)

**Status:** Accepted · 2026-04-21 · Sprint S13.1
**Amended:** 2026-09-16 · Branch `fix/zero-toughness-683` — §7's placeholder-creature exemption ([#683](https://github.com/krakenhavoc/cmd_and_ctrl/issues/683))
**Amended:** 2026-09-19 · Branch `fix/690-691-cda-toughness-sba-and-x-zero` — §7's exemption becomes one predicate, `Card.ToughnessIsKnown` ([#690](https://github.com/krakenhavoc/cmd_and_ctrl/issues/690), [#691](https://github.com/krakenhavoc/cmd_and_ctrl/issues/691))

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
APNAP trigger drain (CR 704.3 + 603.3b) and runs at every priority-
grant boundary. Bounded at 32 iterations as a safety belt against
unintended SBA / trigger ping-pong.

**Placeholder-creature exemption.** A creature with `Toughness == 0`
AND no `Counters` is treated as "stats unparseable" (matches the
existing `Card.Power` documentation note about Mortivore-style "*"
cards and the demo seed) and is NOT destroyed by the toughness-0
SBA. A printed >0 creature reduced to ≤0 by counters does die. New
`Card.CurrentToughness` mirror of `CurrentPower` does the math.

*Amended 2026-09-16 (#683).* The exemption no longer covers a
creature that has **lost** its counters. A printed 0/0 (Hangarback
Walker, Mikaeus, the Lunarch, an X hydra) whose last counter is
removed, or cancelled under CR 704.5q, looked exactly like a
placeholder and survived as a 0/0 nothing could kill. Taking a card's
counters from some to none sets `Card.LostLastCounter`, and the
exemption does not apply to a flagged card, so CR 704.5f puts it into
the graveyard. The flag is per object: it clears wherever `Counters` is
reset for a new object, and removing a counter from a card that has
none does not set it. One exception keeps the placeholder
convention intact: a creature whose printed toughness is not a number
(`*`, `1+*`, `?`) imports with `Card.VariableToughness`, set per face
and carried by copy effects and token copies, and stays exempt after
losing its last counter, because its 0 is the import stand-in and not
a printed 0. A creature that never had a counter (a card cast for
X=0, a declined Clone, the demo seed) is exempt as before. A restore
point written before this change has no `VariableToughness`, so
restore recomputes it from the card's Scryfall printing without a
schema bump. When the printing is unknown (a token, no dump loaded),
the flag is set for any 0 toughness, which keeps the exemption as it
was before (`game/snapshot_backfill.go`).

*Amended 2026-09-19 (#690, #691).* The exemption is no longer a shape
the SBA loop matches on; it is one predicate, `Card.ToughnessIsKnown`,
and it is the only gate CR 704.5f has. The question it answers is
"does the engine know what this object's toughness IS", because that,
and not "is the number 0", is what the convention was ever about.
Decided in order:

1. A nonzero `Toughness` or any counter — never in question.
2. `Characteristic.PTDefined`: a layer 7a or 7b effect DEFINED the
   body in the current pass, so the layer system has computed what
   the stand-in was standing in for. A coded characteristic-defining
   ability is out of the exemption and an uncoded `*` stays in it,
   which is #690: Consuming Aberration and Lord of Extinction print
   `*`, this engine sizes them, and empty graveyards make them real
   0/0s. Set in `applyOneEffectLocked` for the 7a and 7b buckets only
   — 7c modifies, 7d counts counters and 7e switches, and all three
   need a number something else defined — and deliberately excluded
   from `sameCharacteristic`, because the CR 613.8 dependency probe
   asks what an effect does to an object, not what the pass did.
3. Otherwise `VariableToughness` keeps the exemption, unchanged from
   the 2026-09-16 amendment.
4. `LostLastCounter`, unchanged, and now earning its keep for the
   objects branch 5 does not reach.
5. A printing behind the object (`ScryfallID`). Branch 3 already took
   every printing whose toughness did not parse, so what is left is a
   real printed 0/0 and CR 704.5f applies: a Hangarback Walker cast
   for X=0, a Wildwood Scourge that entered with no counters, a
   declined Clone. This is #691. **The engine does not refuse an X=0
   cast** — CR 601.2b makes X a number the caster announces and 0 is
   legal — it just does not survive one. The bot enumerator already
   declines to OFFER X=0 for a card that declares `XMatters`
   (`internal/legal/x.go`), which is a separate rule about what is
   worth putting in front of a player.

What is still exempt is what the convention was always for: objects
with no printing, no computed body and no counter history — a test
fixture that typed a creature line without a body, a 0/0 token
template. The old "placeholder" framing is retired with it; the
engine has no placeholder cards any more, it has objects it has
printed data for and objects it does not.

Skipping is still skipping the whole creature, and that was a second
bug: the CR 704.5g lethal-damage and CR 702.2c deathtouch checks sit
below the same `continue`, so a coded `*` creature could not be killed
by damage either. Both checks now run for every object the predicate
answers yes for.

One importer line backs it up (`deck.variableToughness`): a MISSING
printed toughness on a creature is stamped `VariableToughness` rather
than read as a printed 0. Every creature prints a toughness, so a
record the dump reader cannot read must not become a death sentence
under branch 5; a joined multi-face type line is exempt, because
Scryfall leaves that toughness empty on purpose and `SetFace(0)`
overwrites the flag from the face.

Counter-specific SBAs (planeswalker loyalty 0, battle defense 0,
+1/+1 -1/-1 cancel, poison ≥ 10, saga final chapter) belong to
S13.2 and inherit the same loop.

*Amended 2026-09-23 (#1289): the loop does not run inside a
resolution, including a resolution paused on one of its own prompts.*
CR 704.3 checks state-based actions "whenever a player would receive
priority", and nobody receives priority while a spell or ability is
resolving. "Runs at every priority-grant boundary" was true of the
call sites and false of one of them. `passPriorityLocked` runs the
loop straight after `resolveTopOfStackLocked` returns, and a
resolution that asks a question returns before it has finished: the
prompt is queued, the function carries on, and the rest of the
resolution runs later from the answer. So the sweep ran in the middle
of the resolution. On a Doubling Season plus Hardened Scales board an
earthbend makes a bare land a 0/0 creature and its counter placement
then waits on the CR 616 ordering prompt; the sweep killed the land to
CR 704.5f, the earthbend return brought it back as a new object, and
the counters had nothing to land on. A fresh amass Army died the same
way, and with a doubled creation both Armies died before either could
be chosen.

The decision is one rule with one enforcement point
(`game/resolution_pause.go`):

1. **A resolution is open** from the moment an item begins to resolve
   (`beginResolutionLocked`, called by `resolveTopOfStackLocked` and
   `resolveTopAbilityLocked`) until its CR 704.3 boundary runs.
   `Game.resolutionOpen` records it.
2. **A prompt queued while a resolution is open is stamped**
   `PendingChoice.midResolution` by `QueueChoiceForEffect`, the one
   place every prompt is queued. Two families are left unstamped,
   because the rules place them after the resolution: putting a
   triggered ability on the stack (the CR 603.3b order, the CR 603.5
   optional yes/no, the CR 603.3c mode pick and the CR 603.3d target
   pick, recognised by their resume frames), and the CR 726 loop
   shortcut. Holding the boundary for a trigger's target pick would
   pick targets before state-based actions, which is the order #809
   fixed. A `pick_target` for a spell copy (CR 707.10c) is part of the
   resolution that made the copy and is stamped.
3. **`runStateChecksLocked` holds** while a resolution function is
   still on the Go stack (`Game.resolutionDepth`), or while a stamped
   prompt that blocks the table is open. It holds the whole boundary,
   not only the sweep. CR 603.3 puts triggered abilities on the stack
   "the next time a player would receive priority", which is the same
   moment, and the #809 target refresh and the #864 eliminated-chooser
   sweep belong to it too. When the hold does not apply it closes the
   resolution and runs as before.
4. **The answer that finishes the resolution runs the boundary.** Most
   answer paths already end in `runStateChecksLocked`; it now finds
   nothing to hold for and runs. `actions.Dispatch` calls
   `Game.SettleResolution` after every action as the backstop for the
   paths that do not (a colour pick, a blocking pay-unless), and
   `Concede` settles too, because a departure can drop the last prompt
   a resolution was waiting on.

**Why this cannot wedge a table.** Only a stamped prompt that
**blocks the table** holds (`ChoicePromptBlocksTable`, ADR 0018 §6). A
blocking prompt already refuses every pass, so holding the boundary
behind it stops nothing that was not already stopped, and every
blocking kind is enumerated with an answer for the seat that owes it
(`legal.choiceMoves`). A `pay_unless` asked of somebody else after the
ability has left the stack does not block, so it does not hold, and
the table plays on around it with state-based actions running as §6
decided. A seat that leaves has its prompts settled by ADR 0018 §6's
departure table: a dropped prompt holds nothing, and a reassigned one
(CR 800.4g) holds until its new, seated chooser answers.
`TestAPausedResolutionReleasesWhenItsChooserLeaves` runs that for every
classified kind and exercises both columns.

**Persistence.** `resolutionOpen` and `midResolution` are carried by
Clone (an undo across the answer is paused again) and by the
persisted snapshot (`resolutionOpen` / `midResolution`, omitempty, no
schema bump; a file from before restores unpaused, which is the old
behaviour). `resolutionDepth` is zero between actions and is not
copied: a clone taken inside a resolution must not inherit a function
it is not running.

**One test changed its setup.** `TestEliminatedDuringOwnResolutionReachesDecidedState`
(#864) killed its player by calling `runStateChecksLocked` inside the
resolving ability's own effect. That call now holds, correctly, so the
test takes the player out with `eliminatePlayerLocked` instead and
keeps both of its assertions.

**Out of scope.** The hold is about stack resolutions. A prompt raised
while a spell is being cast or an ability activated, or by a
turn-based action, is not stamped and behaves as before. CR 704.3
applies there too in principle, but no card was found that is wrong
because of it.

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
- The exemption is a documented contract pinned by
  `TestS131SBAPlaceholderCreatureSurvives` and, since #690/#691, by
  `game/toughness_known_test.go` and
  `cards/effects/printed_zero_body_test.go`. Future refactors of the
  SBA loop must preserve it — and must ask `Card.ToughnessIsKnown`
  rather than re-deriving the answer from `Toughness == 0`.
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
