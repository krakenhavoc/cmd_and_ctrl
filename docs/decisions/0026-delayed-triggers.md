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
603.7c/d), obtained without any "was this created during the current
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
