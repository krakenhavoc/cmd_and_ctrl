# ADR 0026 — Delayed triggers and return-from-exile (flicker)

**Status:** accepted (S22)
**Extends:** [ADR 0018](0018-triggers-on-the-stack.md) (triggered abilities use
the stack), [ADR 0013](0013-replacement-effects.md) (the CR 614 pipeline).

## Context

Flicker — "exile it, then return it to the battlefield" — is the theme
of the Azorius blink deck triaged in
[docs/decklists/aang-is-so-flashy.md](../decklists/aang-is-so-flashy.md),
and nine of that deck's cards were blocked on it. Two independent
pieces were missing, and the second is much bigger than the deck:

1. **Nothing returned a card to the battlefield from exile.**
   `ExileTarget` could send a permanent to exile; the exile zone was
   a one-way door for everything except the S21 impulse-play
   permission (ADR 0022), which lets you *cast* a card from exile —
   a different operation, gated on cost and timing.

2. **A delayed trigger had no shape at all.** Most blink cards read
   "exile it. At the beginning of the next end step, return that card
   to the battlefield." That second sentence is a CR 603.7 delayed
   triggered ability: an instruction created by a resolving spell,
   which then waits outside every zone until its condition is met,
   fires once, and ceases to exist.

`game.TriggeredAbility` cannot express (2). It is declared statically
on a catalog `Spec` and harvested by walking `g.Battlefield.Cards`, so
it needs a source permanent in play — and the defining property of a
delayed trigger is that whatever created it is usually gone by the
time it fires (an instant already in a graveyard; a creature that has
left). Delayed triggers are also far more common than blink: rebound,
"until your next turn" effects, "at the beginning of the next end
step, sacrifice it", and every "when you do, ..." reflexive trigger
want the same slot.

## Decisions

### 1. Delayed triggers are a queue on `Game`, not a card ability

```go
type DelayedTrigger struct {
    ID           uuid.UUID
    Controller   uuid.UUID
    SourceCardID uuid.UUID
    Label        string
    At           Step
    CreatedTurn  int
    Cards        []uuid.UUID
    Effect       func(g *Game, item *StackItem) error
}
```

on `Game.DelayedTriggers`, created by
`ScheduleDelayedTriggerForEffect` from inside a resolving effect.

Not a `TriggeredAbility` variant, because everything about the
dispatch differs: no source card on the battlefield, no `Watches`
pre-filter over the event log, fires exactly once, and dies when it
fires. Sharing the type would mean a `TriggeredAbility` with four of
its six fields meaningless.

### 2. It fires on step ENTRY, which is what makes "next" free

`runStepEntryHooksLocked` drains every queued trigger whose `At`
matches the step being entered, immediately before the step's own
turn-based actions.

The consequence is the whole rules content of the word "next" (CR
603.7c), obtained without any "was this created during the current
step" bookkeeping: an ability scheduled *during* an end step is queued
after that step's entry hook has already run, so it waits for the
following one. That timing is observable and it is exactly what blink
players rely on — a blink cast in an opponent's end step returns the
permanent a whole turn later, not moments afterwards. There is a test
pinning it.

### 3. It goes on the stack

A fired trigger becomes a `StackItem` on `PendingTriggers` — the same
queue the S19 harvester feeds — and resolves through the ordinary
APNAP drain. Applying the effect straight from the step hook would
deny every player their response window, which is the same mistake
ADR 0018 retired for ordinary triggers.

### 4. The payload rides on the item, not in the closure

`DelayedTrigger.Cards` is stamped onto the fired item's `Targets` as
`TargetCard` refs, and the `Effect` reads them back off `item`.

This is not cosmetic. A delayed trigger has to survive `Clone` /
`RestoreFrom` — a return that is still owed must be there after an
undo, and one whose creating spell was undone must not be. `Clone`
deep-copies the queue but **shares the `Effect` func**, exactly as
`cloneStackItem` already does for `StackItem.Effect`, on the contract
that the closure takes the live `*Game` and its own item at fire time
and captures neither. Keeping the payload in the struct rather than
in a captured variable is what lets a card's effect be a plain
package-level function with no captures at all
(`returnExiledCardsToOwners`).

Stamping the payload as target refs also earns the CR 608.2b
existence re-check for free, which is the behaviour a delayed return
wants: a card that is no longer where the trigger left it does
nothing rather than erroring.

The queue is projected onto the wire as `GameView.delayed_triggers`
(public information — the ability was announced when its source
resolved, and the cards it names sit in the shared exile zone), so
the client can show what is still owed and the replay log, which is
a stream of those snapshots, carries it.

### 5. A returned permanent is a NEW OBJECT

`ReturnFromExileToBattlefieldForEffect` mints a **fresh
`InstanceID`** and drops every scrap of the old object's battlefield
state: counters, marked damage, tap state, combat declarations, the
CR 613 timestamp, summoning-sickness history, and any lingering
exile-play permission. It returns the new ID; the pre-exile ID names
nothing afterwards.

This is CR 400.7, and it is the entire reason blink is played: it is
simultaneously an ETB engine and an answer, because counters and
auras fall off and a stolen creature goes home. Preserving the ID
would have been less code and a different card.

It is worth flagging as a deviation from an otherwise universal
engine invariant: everywhere else, `InstanceID` is the physical card
and survives every zone change. Returning from exile is the only
operation that re-mints one.

Both ETB paths fire on the return — `EventETB` (which the S19
harvester turns into stack items) and the direct `OnETB` catalog
hook. They are separate mechanisms and a card may use either, so a
return that fired only one would silently half-work.

### 6. The return runs through the CR 614 replacement pipeline

The entry is announced as a `RepEventMove` (exile → battlefield)
before the card is lifted out of exile, so Authority of the Consuls
taps the blinked creature and a self "enters tapped" replacement
still applies.

Running the pipeline *first* is deliberate: if it queues a CR 616
ordering prompt, nothing has moved, so the return is simply dropped
and the card stays in exile rather than being stranded between zones.
(No card in the catalog produces two competing replacements on one
entry today.)

The card's `Controller` is set to its destination controller in exile
*before* the pipeline runs, because a replacement asking "is the
entering permanent mine?" reads `Card.Controller`, and while a card
sits in exile that field still names whoever controlled it before it
left.

## Consequences

- `EventBeginEndStep` joins `EventBeginUpkeep` as a step-boundary
  event, so "at the beginning of your end step" is now expressible.
- Three cards from the Aang triage ship on this: **Y'shtola Rhul**
  (immediate blink on an end-step trigger), **Waterbender's
  Restoration** (delayed blink), **Cosmic Intervention** (turn-scoped
  replacement + delayed return).
- **Phelia, Exuberant Shepherd** is now blocked only on
  `EventAttack`, which is landing separately. When it does, Phelia is
  a trigger declaration over machinery that already exists.
- Delayed triggers are reusable well past flicker: rebound, "at the
  beginning of the next end step, sacrifice it" (token makers),
  "until your next turn" cleanups.

## Amendment, 2026-09-18: a delayed trigger may watch an EVENT (#663)

This reverses §1-2 for one case, and only for that case.

"When you next cast an instant or sorcery spell this turn, copy that
spell" (Doublecast, Galvanic Iteration, Teach by Example, Ral's -2,
Chandra's -2) is a delayed triggered ability whose condition is an
EVENT rather than a step. §1 said a delayed trigger is "not a
`TriggeredAbility` variant … no `Watches` pre-filter over the event
log", and §2 said firing on step entry is what gives "next" its
meaning for free. Both sentences are true of the step-conditioned
trigger and neither can express this card: there is no step to wait
for, and the thing being waited for is a cast that may never happen.

**Decision: extend `DelayedTrigger` with an event condition, rather
than build a second registry.**

```go
type DelayedTrigger struct {
    …
    On        []EventKind
    AppliesTo func(ev Event, dt *DelayedTrigger, g *Game) bool
    Optional  *TriggerOptionalPrompt
    Duration  *Duration
}
```

`At` and `On` are alternatives, and a trigger may carry both (nothing
in the catalog does yet). A trigger with `On` set is checked by ONE
hook in `triggerHarvester.OnEvent`, after the zone walks: the first
matching event fires it and removes it from the queue, in one place,
with no per-card special case anywhere.

The alternative considered and rejected was a `TurnScopedTriggers`
registry modelled on `TurnScopedReplacements` / the scoped statics
([ADR 0063](0063-durations-and-control.md)), which would have left this
ADR intact. It was rejected because it
would have been a second queue holding the same four things this one
holds (a controller, a source, a label, an effect), snapshotted twice,
cloned twice, and drained by a second dispatch — and the one real
difference between the two is a predicate. A recorded decision is
worth reversing in writing; it is not worth duplicating a queue to
avoid reversing.

### Why the properties of §2-4 survive

- **"Next" is still free, by a different mechanism.** A step-conditioned
  trigger gets it from the step-entry hook having already run. An
  event-conditioned one gets it from *when it is created*: the trigger
  is created by a resolving spell, and the `EventCast` of the spell
  that created it was emitted before that resolution began. A
  Doublecast cannot copy itself, and nothing has to remember that it
  must not.
- **It fires exactly once (CR 603.7b).** The queue is rewritten before
  the first item is dispatched, exactly as `fireDelayedTriggersLocked`
  rewrites it, so a trigger cannot see its own follow-on events. Two
  Doublecasts in one turn are two entries and both fire on the same
  cast — two copies, which is the printed outcome.
- **It uses the stack, through the harvester's own dispatch.** §3 said
  a fired delayed trigger becomes a `StackItem` on `PendingTriggers`.
  The event-conditioned one goes further and enters
  `dispatchTriggerLocked` — the same function `harvestFromZone` calls,
  the same one `QueueReflexiveTriggerForEffect` calls since the 2026-09-17
  amendment — so the CR 603.3d drop, the CR 603.5 "you may" and the
  APNAP drain are the harvester's and not a second implementation.
- **The payload still rides on the item.** §4 holds unchanged for
  `Cards`; what is new is that the TRIGGERING EVENT'S object rides too,
  as `StackItem.Payload`, which is how "copy THAT spell" names the
  spell without the `Effect` closing over it.

### Duration (CR 514.2)

"This turn" is a real clause and it is the default: an event-conditioned
trigger created with no explicit duration is stamped with
`g.UntilEndOfTurnDuration()` and swept by `sweepTurnEndLocked`, beside
the scoped statics and the turn-scoped replacements, whether or not it
ever fired.

It is the SAME `Duration` [ADR 0063](0063-durations-and-control.md)
gave the scoped statics, and deliberately so: "until end of turn" means
the same thing to a delayed trigger as it does to a Giant Growth,
`durationExpiredLocked` is the one function in the engine that decides
when a duration is over, and a private int here would have been a
second answer to that question — the exact shape ADR 0063 retired when
it replaced `ScopedStatic.ExpiresAfterTurn`. A card wanting "until your
next turn" needs no new machinery.

A step-conditioned trigger is unaffected: its `Duration` is nil, which
means "no duration", and that is what "at the beginning of the NEXT end
step" needs when it is scheduled during an end step.

### Persistence

`On` and `Duration` are data and are carried by `Clone`,
`snapshot.go` and the wire view. `AppliesTo` and `Optional` are
closures and are dropped, like `Effect` beside them — a restored
trigger with no `Effect` is inert either way, and the whole trigger is
already counted once in `ContinuationCensus.DelayedTriggerEffects`, so
nothing is double-counted and nothing is silently lost.

### Citation fix

§2 above cited "CR 603.7c/d" for "next". The rule that a delayed
trigger fires only once is **CR 603.7b**; CR 603.7c is the one that
gives "the next time" its meaning. The §2 text now cites 603.7c, and
this amendment cites 603.7b for the fires-once property.

## Amendment, 2026-09-17: the reflexive sibling is NOT this slot (#636)

The Context above says "every 'when you do, ...' reflexive trigger"
wants the delayed-trigger slot. It does not, and #636 builds the other
thing instead — `Game.QueueReflexiveTriggerForEffect` plus
`effects.ReflexiveTrigger`, in `reflexive.go` beside this file's
`delayed.go`.

The two look alike from a card file (both are created by a resolving
effect, both outlive the source, both go on the stack) and differ in
the one place that decides the design: **when the condition is met.**
A delayed trigger's condition is a future step, so it has to sit in a
queue on the `Game` until that step arrives, which is why
`DelayedTriggers` exists and why it is snapshotted. A CR 603.12
reflexive trigger's condition was met by the resolution that created
it — that is what "when you DO" means — so it goes on
`PendingTriggers` immediately and reaches the stack at the next
priority boundary, above its parent. Putting it in the delayed queue
would have needed a step it does not have, and would have skipped the
CR 603.3d target pick, which is the half of the rule the folded cards
were actually getting wrong.

What it does inherit from §3 and §4, unchanged: it uses the stack
(same `PendingTriggers` → `StackMeta` path, same response window), and
its payload rides on the item rather than in the closure — on
`StackItem.Payload` rather than on `Targets`, because a reflexive
trigger's `Targets` is its own target clause and is rewritten when the
pick comes in. Everything else is the S19/S20 harvester's dispatch,
reused rather than re-implemented: one `dispatchTriggerLocked`, so
there is one place the CR 603.3d drop, the CR 608.2b re-check and the
APNAP drain can be got right.

## Amendment, 2026-09-21: an event-conditioned trigger may watch an OBJECT, and may outlive the turn (#1178)

The 2026-09-18 amendment above built `DelayedTrigger.On` for one shape
— "when you next CAST an instant or sorcery spell **this turn**" — and
stamped `UntilEndOfTurn` on any event-conditioned trigger that named no
duration, "because every printed one says *this turn*". Earthbend is
the first that does not.

> *When it dies or is exiled, return it to the battlefield tapped.*

Three things follow, and all three are recorded in full in
[ADR 0081](0081-earthbend-and-object-keyed-delayed-triggers.md)
Decision 5; what belongs here is what changed about **this** slot:

1. **The condition watches an OBJECT, not an actor.** `AppliesTo` reads
   the trigger's own `Cards` payload and matches `EventLTB` whose
   `CardID` is that object and whose `NewZone` is a graveyard or exile.
   `NewZone` is where the card ACTUALLY went, after the CR 614 window
   on the move settled, so a replacement that exiles a dying permanent
   still satisfies "dies **or is exiled**".
2. **It can carry a non-turn duration.** `Duration` is `Indefinite`
   pinned to the object (ADR 0063's `Game.PinnedTo`), so the trigger is
   owed for as long as the permanent stands and is swept by
   `clearExpiredDelayedTriggersLocked` the moment the permanent leaves
   by a route the condition does not name. The pin is the garbage
   collector, not the rule — the rule is CR 400.7, which makes the
   returning card a new object the trigger no longer names.
3. **Identity may be DERIVED.** `DelayedTrigger.ID` is minted when left
   `uuid.Nil`, and a caller that must not queue two triggers about the
   same object can compute one instead — earthbend hashes
   `{instance, EnteredBattlefieldAt}`. No schema change, and the queue
   stays plain data that clones and snapshots by value.

Everything §2-4 promised still holds, for the reason the 2026-09-18
amendment gives: the dispatch is still `dispatchTriggerLocked`, so the
CR 603.5 "you may", the CR 603.3d drop, the CR 608.2b re-check and the
APNAP drain are the harvester's and not a second copy.

## What this deliberately does not do

- **No extra steps.** Y'shtola's "there is an additional end step
  after this step" is not implemented; `Turn.advance` walks a fixed
  twelve-step sequence by index and inserting a step is turn-machinery
  work.
- **No alternative cast costs.** Cosmic Intervention ships without
  foretell and Waterbender's Restoration without a costed waterbend
  {X}, for the same reason Vandalblast and Cyclonic Rift ship without
  overload — "pay X *instead of* the mana cost" has no expression yet.
- **No indestructible, no devotion.** Thassa, Deep-Dwelling has the
  same immediate-blink end-step trigger Y'shtola has, and would have
  been a fourth card, but its "as long as your devotion to blue is
  less than five, Thassa isn't a creature" clause is genuinely
  inexpressible: `Card.IsCreature()` reads `Card.TypeLine` directly,
  not `Effective().Types`, so a Layer 4 type-changing effect is inert
  for combat, SBAs and targeting. Shipping Thassa would have meant
  shipping a permanently-a-creature Thassa and calling it a
  simplification, which is a different card rather than a smaller one.
