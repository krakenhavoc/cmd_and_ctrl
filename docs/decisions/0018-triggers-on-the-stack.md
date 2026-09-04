# ADR 0018 — Triggered abilities use the stack

**Status:** Implemented · 2026-09-02 · Branch `feat/s19-triggers-on-stack`

## Context

S19 sub-PRs 1–5 shipped the auto-fire trigger pipeline: a per-game
`triggerHarvester` listener walks the battlefield on every event,
matches `Spec.Triggered` declarations, and calls `Build` to produce
a `StackItem`. Twelve catalog cards (ETB, dies, upkeep) ride it.

What they actually did, though, was apply the effect *inside*
`Build` and return `nil`. Mulldrifter drew two cards the instant
the creature resolved; Phyrexian Arena took its life and card as
the upkeep step began; Reclamation Sage destroyed the artifact the
moment its controller clicked "Yes". Nothing ever reached
`PendingTriggers` or `StackMeta`. The card files called this a
sandbox simplification "deferred to a future engine-completeness
sweep", and the engine had the matching gap: `resolveTopAbilityLocked`
deleted the item and emitted `resolve` with no way to run anything.

That shortcut cost three things:

- **No response window.** A trigger that resolves at announce time
  can't be answered — no counter-the-ability, no sacrifice in
  response, no bounce-the-target. S19's own exit criterion
  ("advance to upkeep with Phyrexian Arena → trigger goes on stack
  automatically") wasn't met.
- **No CR 608.2b re-check.** Targets picked at announce time were
  acted on immediately, so "target left before resolution" could
  never happen — and neither could the fizzle a correct engine
  produces.
- **A migration debt that grows per card.** Sub-PRs 6 and 7 add
  eight more trigger cards. Every one written in the inline pattern
  is one more to rewrite later.

The stack itself already supported ability items (S13.1 manual
`AnnounceTrigger`, `ActivateAbility`, the LIFO `Seq` fix in #202,
`CounterAbility`). Only the resolution effect was missing.

## Decisions

### 1. `StackItem.Effect` — a resolution callback on the item

```go
Effect func(g *Game, item *StackItem) error
```

`resolveTopAbilityLocked` now mirrors the spell path: remove the
item from `StackMeta`, run `spellAllTargetsIllegalLocked` (every
targeted slot illegal → `EventFizzle`, no effect), emit
`EventResolve`, then run `Effect`. An error from `Effect` surfaces
as `EventEffectError` and does not wedge the stack — the ability
has ceased to exist either way (CR 608.2m). `nil` keeps the S13.1
manual-sandbox meaning: the players resolve it by hand.

Why a closure on the item rather than routing through the
`EffectResolver` hook by oracle ID + ability index: the trigger's
announce-time context (which event fired it, the CR 603.10 LKI
snapshot, the auto-picked target) is already captured at `Build`
time, and a closure is the cheapest way to carry it to resolution.
The `triggerResumeFrame` on `PendingChoice` set the precedent for
func-valued state in the game struct. Games aren't serialised (the
protocol package projects its own views; undo goes through
`Clone`), so a func field costs nothing on the wire.

Two rules follow from undo: `cloneStackItem` copies `Effect`, and
the callback **takes the live `*Game`** instead of capturing one,
so an item restored from a snapshot resolves against the restored
game. For the same reason it must not capture `*Card` pointers into
zone slices — it reads controller / source / targets off the
`item` it receives. `NewTriggeredItem(source, label, effect)` is
the one-liner that gets the shape right.

### 2. `Build` builds; it does not resolve

`TriggeredAbility.Build` returns the item and **must not mutate
game state**. Targets are chosen here (CR 603.3d — when the
ability is put on the stack) and stamped onto `item.Targets`, so
the stack overlay shows them and the resolve-time re-check applies.
A targeted trigger with no legal target returns `nil` and never
reaches the stack, which is also what CR 603.3d says.

All twelve S19 cards were migrated. The dies-trigger and upkeep
cards were mechanical; Acidic Slime and Reclamation Sage share a
`destroyTargetTrigger` helper; Eternal Witness carries its
top-of-graveyard pick as a target and lets `ReturnFromGraveyard`
report `ErrCardNotFound` if the card was exiled in response.

### 3. Every priority boundary drains — including the ones that
### didn't before

A harvested trigger goes `PendingTriggers → StackMeta` via
`drainPendingTriggersAPNAPLocked`, which ran only from the
priority-wrap and step-advance paths. Two paths queued triggers and
then left them stranded until the next wrap — where an empty stack
would advance the step *before* draining, putting the trigger a
step late:

- `ResolveTriggerPrompt` (answering "Yes") now calls
  `runStateChecksLocked` — answering is the moment the ability is
  put on the stack (CR 603.3).
- `MoveCardByID` (the sandbox move_card verb) and the land branch
  of `CastSpell` now call `runStateChecksLocked`. Both are special
  actions after which the actor keeps priority (CR 116.3c), and
  CR 117.5 puts SBAs + the trigger drain at exactly that boundary.
  A catalog creature dropped straight onto the battlefield gets its
  ETB trigger on the stack immediately.

### 4. Wire order is by `Seq`

`viewOfStackItemsInStackOrder` appended ability items in map
iteration order after the spell cards. With abilities on the stack
every turn, that's visibly random. It now sorts by insertion `Seq`
— the same order the engine resolves in — with stable ties keeping
the legacy spell-cards-first behaviour for zero-`Seq` snapshots.

### 5. Client: autopass passes through triggers; autopass off is full control

The S13.6 auto-pass gate treated "stack non-empty" as "a card is
on `Game.Stack`", so ability items didn't count and the
conventional (autopass-off) path would sail past every trigger —
Counterspell in hand or not. The gate now stops for ability items
exactly as it does for spells. The two modes mirror Arena:

- **Autopass on** (the session toggle): every priority window
  passes, triggers included — "get me through this turn".
- **Autopass off**: full control. Every spell *and* every trigger
  on the stack is a window the viewer answers by hand. Step stops
  and the smart-skip predicate still govern empty-stack windows
  as before; they never skip a non-empty stack.

No smart-skip escape for triggers: a player who wants routine
upkeep triggers to fly by turns autopass on.

### 6. "Unless that player pays" is a pending choice queued at resolution

Rhystic Study, Smothering Tithe and Esper Sentinel (sub-PR 6) ask
a *different* player a question when the trigger resolves. That is
`PendingChoicePayUnless`: the trigger's `Effect` queues it for the
taxed player (`PayUnless` primitive) and returns; the ability has
resolved and left the stack. The answer arrives as
`resolve_choice {apply}` like the other yes/no kinds. "Pay" spends
from the chooser's pool, auto-tapping their untapped sources first
via the S15 planner (no exclusions); a "Pay" they can't cover
degrades to a decline. Decline runs the card's consequence
(`OnDecline`) against a fresh `Context` for the original item.

Sandbox looseness, accepted: priority isn't gated on the pending
prompt, so play can continue while a Rhystic tax is unanswered.
It matches how the table plays it in paper ("you paying for that?"
while the next spell is already being cast) and avoids a modal
lockstep across four browsers.

Two more drains fell out of the sub-PR 6 cards: `CastSpell` (the
caster gets priority right after casting, CR 117.3c — Rhystic's
trigger must be on the stack by then) and the sandbox `draw_card`
verb (Tithe / Sphinx watch draws). `Game.SpellsCastThisTurn` is
bumped before `EventCast` fires so "first noncreature spell each
turn" reads `Noncreature == 1` for the spell that triggered it.

## Out of scope (explicit deferrals)

- **Treasure's sac-for-mana** is inert until S21 ships sacrifice
  as a cost component; Smothering Tithe makes countable artifacts.
- **"You may draw"** on Rhystic Study is treated as "draw".

- **Modal triggers** ("draw a card or gain 3 life") still need a
  `ModePrompt` slot; nothing in the catalog needs one yet.
- ~~**Trigger ordering UI** for one player's simultaneous triggers
  (CR 603.3b)~~ — shipped in sub-PR 8: `PendingChoiceTriggerOrder`
  holds the APNAP drain until each seat with ≥2 *differing*
  triggers has ordered them (identical ones auto-order; the
  submitted order is resolution order). A trigger arriving while a
  prompt is open re-asks with the full list.
- ~~**Cast / combat-damage triggers**~~ — shipped in sub-PRs 6 and 7
  in the `NewTriggeredItem` shape.
- ~~**Stack overlay art for ability items.**~~ Shipped in sub-PR 8:
  the overlay resolves `source_card_id` against battlefield / exile
  / graveyards for art, title and target names.
- **`PlayCard`** (hand → battlefield sandbox drop, pre-S13.1) still
  emits no ETB at all. Left alone; `move_card` is the verb the
  client uses.

## Consequences

- The test harness `passPriorityAroundTable` now waits for
  `Game.Stack`, `StackMeta`, and `PendingTriggers` to all empty, so
  "cast creature → ETB trigger → effect" settles in one call. Tests
  for optional triggers answer the prompt and call it again.
- The Playwright S19 suite resolves triggers through each player's
  own "next" button (`resolveStack`) rather than an admin
  `pass_priority`, so a browser that auto-passed can't be
  double-passed into a step advance.
- Engine coverage: effect runs on resolution not build; fizzle when
  every target is gone; effect error surfaces and clears the item;
  effect survives `Clone`/`RestoreFrom`; sandbox move drains onto
  the stack. Catalog coverage: countering Mulldrifter's trigger
  draws nothing; Reclamation Sage fizzles when its target leaves in
  response and does not re-pick; two dies triggers from one wrath
  stack APNAP.
