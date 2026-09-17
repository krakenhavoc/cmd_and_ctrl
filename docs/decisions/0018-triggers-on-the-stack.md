# ADR 0018 — Triggered abilities use the stack

**Status:** Implemented · 2026-09-02 · Branch `feat/s19-triggers-on-stack`
**Addendum:** 2026-09-17 · **Accepted** · [Trigger doubling (CR 603.2d)](#addendum-2026-09-17-trigger-doubling-cr-6032d--accepted) · tracked on [#752](https://github.com/krakenhavoc/cmd_and_ctrl/issues/752)

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

**Amendment (2026-09-17, #730): the looseness is `pay_unless` only.**
Nothing refused `advance_step`, or a priority pass that would advance
the step, while *any* prompt sat in `Game.PendingChoices` — so the
sandbox latitude this section granted Rhystic Study was in practice
granted to every prompt in the engine, and a sacrifice, a scry or a
CR 616 ordering could be walked past and then answered against a game
that had moved on (#701 and #725 each had to add a stale-frame guard
for exactly that; #651 reports the same shape for effect discards).

`Game.blockingChoiceLocked` (`server/internal/game/choice_gate.go`) now
gates `AdvanceStep`, `PassPriority` and `PassTurn`, which return a
`*ChoicePendingError` (wrapping `ErrChoicePending`, naming the choice's
ID and kind) while a blocking prompt is outstanding. It is
**deny-by-default with an explicit allowlist**, so a choice kind added
later blocks without its author having to know this rule exists. The
allowlist, verbatim:

- **`pay_unless` — does not block.** The reason this section already
  gives: the ability has resolved and left the stack, the question is
  addressed to a *different* player, and the answer spends from that
  player's pool or runs `OnDecline` — neither reads the step. Play
  continues, as it does in paper.

Every other kind **blocks**: `discard_from_hand`, `mana_pick`,
`replacement_order`, `optional_replacement`, `damage_assignment`,
`trigger_prompt`, `trigger_order`, `pick_target`, `sacrifice_choice`,
`scry`, `surveil`, `look_at_top`, `search_library`, `may_cast`,
`choose_protector`, `legend_rule`, `choose_color`, `confirm`,
`choose_cards`, `entry_pay_life`, `copy_target`,
`choose_creature_type` — and anything added after this line.

`mana_pick` is the near miss worth recording: it looks like background
bookkeeping, but an unanswered colour pick is mana that has *not*
entered the pool, `internal/legal` already offers a seat nothing else
while one is open, and the auto-tapper refuses to create one precisely
because its contract is "no further player decisions required". It
blocks.

What the gate does **not** do: answering a prompt (`resolve_choice`
and its kin), `concede`, chat, undo, and the admin context menu's raw
sandbox moves (`move_card`, `change_life`, `add_counter`,
`mark_damage` — ADR 0033 §8's list) are all untouched, so a table can
always be unstuck by hand. Casting and activating are not gated either;
`internal/legal` declines to offer them, and widening the refusal that
far is a bigger behaviour change than this bug needs.

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

## Addendum (2026-09-17): trigger doubling (CR 603.2d) — Accepted

**Status:** Accepted · 2026-09-17 · unscheduled (card-coverage audit, wave 2) · tracked on [#752](https://github.com/krakenhavoc/cmd_and_ctrl/issues/752)
**Numbering:** an addendum to ADR 0018, not a new ADR. No new number, and the ADR range line in AGENTS.md §3 is unchanged.
**Builds on:** this ADR (the harvester, `Build` / `Effect`, the APNAP drain, the `trigger_order` prompt),
[ADR 0049](0049-card-engine-seam-review.md) D2 (`OncePerBatch`, #587) and D4 (`CardDef`, #622),
the reflexive-trigger seam (#636, `game/reflexive.go`), the delayed-trigger queue
([ADR 0026](0026-delayed-triggers.md), `game/delayed.go`), the simultaneous-exit batch
(S23, `game/simultaneous.go`), [ADR 0041](0041-game-persistence.md) (snapshots) and
[ADR 0055](0055-loop-breaker.md) (the loop breaker's per-key count).
**Owner decisions:** three questions, answered 2026-09-17 and recorded at the
[end of this addendum](#decided-2026-09-17). The decisions below apply them:
extra instances are tagged on the stack overlay and on their prompt with no log
line (question 1), the first wave includes Elesh Norn, Mother of Machines and
Veyran, Voice of Duality with declared caveats (question 2), and each doubled
optional trigger gets its own yes/no prompt (question 3).

Line references are to `origin/develop` at `684f2786`.

### Context

#### What the rules ask for

CR 603.2d (Aug 7 2026 edition):

> An ability may state that a triggered ability triggers additional times. In
> this case, rather than simply determining that such an ability has triggered,
> determine how many times it should trigger, then that ability triggers that
> many times. An effect that states that an ability triggers additional times
> doesn't invoke itself repeatedly and doesn't apply to other effects that
> affect how many times an ability triggers. An effect that states a triggered
> ability of an object triggers additional times refers only to triggered
> abilities that object has, not to any delayed or reflexive triggered
> abilities (see rule 603.7 and rule 603.12) that may be created by abilities
> the object has.

The printed rulings on the cards this addendum is for (Scryfall, checked
2026-09-17) settle the rest. They say the same things across Panharmonicon,
Teysa Karlov, Yarok, Elesh Norn, Naban, Katara, Starfield Vocalist, Annie Joins
Up, Roaming Throne, Felix Five-Boots and Drivnod:

- **Not a copy.** "Doesn't copy the triggered ability; it just causes the
  ability to trigger twice." Modes and targets are chosen separately for each
  instance as it goes on the stack, and choices made on resolution are made
  separately too.
- **Doublers add up, they don't multiply.** Two Panharmonicons give three
  instances, three give four.
- **Simultaneity counts in both directions.** A permanent entering at the same
  time as Panharmonicon, Yarok, Elesh Norn, Naban, Katara or Starfield Vocalist
  (including the doubler itself) is doubled. A creature dying at the same time
  as Teysa or Drivnod (including Teysa herself) is doubled. If Teysa and a
  second Teysa meet through the legend rule, "an ability that triggers when it
  dies due to the legend rule triggers three times".
- **The cause is the event, not the words.** Teysa doubles "whenever an
  artifact is put into a graveyard from the battlefield" when an artifact
  creature dies, and doubles a "leaves the battlefield" trigger when the
  creature left by dying. She does not double "whenever you sacrifice a
  creature". Isshin doubles only conditions "directly related to attacking",
  not "whenever this creature becomes tapped". Felix doubles only triggers on
  the damage itself, not Ajani's Pridemate gaining from lifelink.
- **The creature is judged as it exists on the battlefield**, continuous
  effects included (a land that has become a creature dies and Teysa counts it;
  a Runeclaw Bear under Arcane Adaptation naming Wizard is a Wizard for Naban).
- **The count is fixed when the ability triggers.** Making a creature legendary
  after its ability triggered adds no instance under Annie Joins Up, and making
  it nonlegendary afterwards removes none.
- **Controller of the entering or dying object doesn't matter** unless the
  doubler prints "you control" (Naban's "a Wizard you control", Felix's "a
  creature you control"). The controller of the ability's *source* does: "a
  triggered ability of a permanent you control". Cloud, Midgar Mercenary is the
  one card here with no "you control" at all: "a triggered ability of Cloud or
  an Equipment attached to it".
- **Replacement effects and "as enters" choices are unaffected.** They are not
  triggers, and the engine already doesn't route them through the harvester.
- **Linked abilities:** extra instances are linked to the same second ability.
  "The exiled card" means every card any instance exiled.

#### Where triggers are put on the queue today

Every catalog trigger that comes from an event passes through
`dispatchTriggerLocked` (`game/triggers.go:364`). Its callers on develop:

| caller | line | what it harvests |
|---|---|---|
| `harvestFromZone` | `triggers.go:310` | battlefield watchers |
| `harvestCastFromStack` | `triggers.go:345` | `FromStack` abilities of the spell just cast (cascade) |
| `harvestLTB` | `triggers.go:449` | the leaving card's own triggers, off its LKI snapshot |
| `harvestSimultaneousExitLocked` | `simultaneous.go:147` | watchers that left earlier in the same batch |
| `QueueReflexiveTriggerForEffect` | `reflexive.go:175` | **CR 603.12 reflexive triggers** (shipped by #636 after the issue was written) |

Three more places build a triggered item without dispatch:

| site | line | what |
|---|---|---|
| `fireDelayedTriggersLocked` | `delayed.go:161`, queue at `:181` | CR 603.7 delayed triggers |
| `queueAltCostEntryTriggerLocked` | `alternative_cost.go:549`, queue at `:563` | evoke's "when this permanent enters, if its evoke cost was paid, its controller sacrifices it" (CR 702.74a). Warp's exile at `:557` is a delayed trigger. |
| `AnnounceTrigger` | `mutations.go:2093` | the sandbox's manual "trigger" button |

Three facts on develop shape the design:

1. **The issue's proposed check point would double reflexive triggers.**
   #752 suggests counting doublers inside `dispatchTriggerLocked`. Since #636,
   reflexive triggers enter that same function (`reflexive.go:175`), and CR
   603.2d excludes them by name.
2. **`OncePerBatch` would swallow the extra instance.** `dispatchTriggerLocked`
   returns at `triggers.go:365` when `triggerInFlightLocked` (`:554`) finds an
   item with the same source and key on `PendingTriggers`, on `StackMeta`, or
   waiting on a trigger or target prompt. A second dispatch for the same match
   finds the first one and returns.
3. **`harvestLTB` deletes the leaving card's LKI before the simultaneous-exit
   pass runs** (`defer delete` at `triggers.go:405`, and `OnEvent` calls
   `harvestSimultaneousExitLocked` after `harvestLTB`). A dying doubler's
   characteristics must therefore be read before that delete, or out of the
   batch copy (`simultaneous.go:78`).

#### What already exists

- `CardDef` (`game/carddef.go:31`): one precomputed catalog lookup, with a
  per-slot hook variable kept only where a test needs to stub it.
- `TriggeredAbility.Key` and `OncePerBatch` (`triggers.go:177-194`).
- The trigger and target prompts carry a value copy of the source, the LKI and
  the `Build` closure in `triggerResumeFrame` / `pickTargetFrame`
  (`pending_choice.go:604-625`), so each instance can own its own prompt.
- `seatNeedsTriggerOrder` (`mutations.go:2928`) already treats items with the
  same source and label as identical and skips the `trigger_order` prompt for
  them.
- Entries that go through `putOntoBattlefieldFromZoneLocked` move every card
  before announcing any of them (`battlefield_put.go:330-346`), so a watcher or
  doubler in the batch sees the others. Token creation announces each token as
  it is created (`effect_api.go:2098-2109`).
- The one catalog card that already declares the missing seam as a caveat:
  Cloud, Midgar Mercenary (`cards/effects/cloud_midgar_mercenary.go:34`).

### Decision 1 — A doubler is a catalog slot: `Spec.TriggerDoublers`

```go
// game/trigger_doubling.go
type TriggerDoubler struct {
	// Label names the static for tests and the attribution ("Panharmonicon").
	Label string
	// Applies reports whether this doubler adds one instance of the
	// ability described by q. Runs under g.mu in write mode and must not
	// take public locks or mutate state.
	Applies func(g *Game, q TriggerDoublingQuery) bool
}

type TriggerDoublingQuery struct {
	Event Event

	// The doubler as it is now (battlefield) or as it last existed (a
	// batch copy or the LKI of the card whose exit is being reported).
	Doubler    Card
	DoublerLKI Characteristic

	// The ability's source and the characteristics the harvester judged
	// it on (Effective() on the battlefield, the LKI snapshot for an LTB,
	// the batch copy for a simultaneous exit).
	Source    Card
	SourceLKI Characteristic
	// FromSpell is true for a FromStack ability of a spell being cast.
	// Every doubler except Echoes of Eternity prints "of a permanent".
	FromSpell bool
	Ability   *TriggeredAbility

	// Subject is the object the event is about (the entering, dying or
	// attacking permanent, the spell cast) with its characteristics as
	// the rulings require: live on the battlefield for an entry or an
	// attack, the LKI snapshot for an exit. Zero when the event has none.
	Subject    uuid.UUID
	SubjectLKI Characteristic
	HasSubject bool
}
```

`effects.Spec` gains `TriggerDoublers []game.TriggerDoubler`, `CardDef` gains
the same slot, and `buildDef` copies it. A `CatalogTriggerDoublers` hook
variable is added beside `CatalogTriggers`, because the engine tests in
`internal/game` stub doublers without importing the catalog (#622's rule for
when a per-slot variable earns its place).

**Why a predicate over the event rather than a declared cause enum.** The
wording varies more than an enum would (Gandalf's "a legendary permanent or an
artifact entering or leaving", Veyran's "you casting or copying", Felix's "a
creature you control dealing combat damage to a player", Cloud's "as long as
Cloud is equipped"). The common shapes become one-line helpers (Decision 6),
and a card with unusual wording writes its own predicate the way cards already
write `AppliesTo`.

**Why not a `StaticAbility`.** Statics are layer effects that the layer
engine recomputes. A doubler changes nothing in any layer. It is read at one
moment, by the harvester, and putting it into the layer system would make every
recompute carry a slot that is never applied.

**The count is `1 + the number of doublers whose Applies returns true`.** Each
doubler adds exactly one, so two Panharmonicons give three instances. A doubler
is never asked about its own contribution or another doubler's (the second
sentence of CR 603.2d), because doublers are not triggers and the scan never
looks at instances it has already added.

### Decision 2 — The count is taken at the harvest, not inside `dispatchTriggerLocked`

`dispatchTriggerLocked` is split in two:

```go
// harvestMatchLocked: the four harvest callers use this instead of
// dispatchTriggerLocked.
func (g *Game) harvestMatchLocked(pass *harvestPass, source Card, lki Characteristic, t TriggeredAbility, fromSpell bool) {
	if t.OncePerBatch && g.triggerInFlightLocked(source.InstanceID, t.Key) {
		return
	}
	extra := g.triggerDoublersLocked(pass, source, lki, t, fromSpell) // []doublerRef
	g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, doublerRef{})
	for _, d := range extra {
		g.dispatchTriggerInstanceLocked(pass.ev, source, lki, t, d)
	}
}

// dispatchTriggerLocked keeps its signature and meaning for the
// reflexive path: one instance, with the in-flight check.
func (g *Game) dispatchTriggerLocked(ev Event, source Card, lki Characteristic, t TriggeredAbility) {
	if t.OncePerBatch && g.triggerInFlightLocked(source.InstanceID, t.Key) {
		return
	}
	g.dispatchTriggerInstanceLocked(ev, source, lki, t, doublerRef{})
}
```

`dispatchTriggerInstanceLocked` is today's body from `triggers.go:368`
onward: the CR 603.3d empty-target drop, the optional prompt, the target pick,
`Build`, and the queue.

What this decides:

- **Reflexive triggers are never doubled**, because `QueueReflexiveTriggerForEffect`
  keeps calling `dispatchTriggerLocked`. **Delayed triggers are never doubled**,
  because `fireDelayedTriggersLocked` never reaches dispatch. Both are the CR
  603.2d exclusion, enforced by where the count is taken rather than by a flag
  someone could forget to set.
- **Each instance is a complete, independent dispatch.** It gets its own CR
  603.3d legal-target check, its own "you may" prompt, its own target pick, its
  own `Build` call and its own `EventTrigger`. That is the "choices are made
  separately for each instance" ruling, and it needs no new prompt kind.
- **The instances are queued one after another**, the first before the extras,
  so on `PendingTriggers` they are adjacent and in doubler order.
- **The manual `AnnounceTrigger` button is not doubled.** A player announcing
  by hand announces as many as they mean to.

#### Evoke's sacrifice trigger goes through the same count

CR 702.74a makes evoke's sacrifice "a triggered ability that functions on the
battlefield", triggered by the permanent entering, so Panharmonicon, Yarok,
Katara (for an Ally) and the like double it. `queueAltCostEntryTriggerLocked`
builds a `harvestPass` for a synthetic `EventETB` naming the card (the real
`EventETB` was emitted just before, at `mutations.go:1709` / `entry_choice.go:367`),
asks the same `triggerDoublersLocked`, and queues one sacrifice item per
instance. Nothing a player sees changes except the stack: the second sacrifice
resolves after the permanent is gone, and the existing `Effect` already does
nothing then (`alternative_cost.go:569-578`). Warp's exile stays a delayed
trigger and is not doubled (CR 702.185a, CR 603.2d).

#### Saga chapters are ordinary triggered abilities

Chapter abilities are catalog `TriggeredAbility` values watching
`EventSagaChapter` (`cards/effects/saga.go:51`). They reach the harvester like
any other ability, so a doubler keyed on the *source* (Katara, Annie Joins Up,
Roaming Throne, Echoes of Eternity) doubles them when the Saga fits its filter.
Doublers keyed on an *event* (entering, dying, attacking) do not, because the
chapter's event is a lore counter. For chapter I on a Saga's entry, this is the
engine's reading of CR 714.3a (the counter is put on "as it enters", a
replacement) and CR 714.2b (the trigger condition is the counter). Sub-PR 2
checks it against the rulings before any entry doubler ships. If a ruling says
otherwise, the entry helper gains the `EventSagaChapter` case whose counter came
from the entry, and nothing else moves.

### Decision 3 — Which doublers are asked, and as what

`OnEvent` builds a `harvestPass` once per event and passes it to all four
harvest functions:

```go
type harvestPass struct {
	ev         Event
	subject    uuid.UUID
	subjectLKI Characteristic
	hasSubject bool
	doublers   []doublerCandidate // nil until the first match needs them
	scanned    bool
}
```

1. **The subject is resolved eagerly, at the top of `OnEvent`.** For
   `EventLTB` it is the LKI snapshot of `ev.CardID`, read *before*
   `harvestLTB`'s delete (fact 3 in Context). For `EventETB`, `EventZoneMove`
   into the battlefield, `EventTokenCreated` and `EventAttack` it is the live
   battlefield card's `Effective()`. For `EventCast` it is the spell on the
   stack. For a combat `EventDealDamage` it is `ev.Source`, with `ev.Actor`
   already carrying its controller at emit time (`events.go`, the `Combat`
   field comment). This costs one map lookup or one battlefield find per event,
   and only for those kinds.
2. **Doubler candidates are collected lazily, on the first match of the
   event.** Most events match nothing, and most games have no doubler, so the
   scan runs only when a trigger is about to be dispatched, and then at most
   once per event. The candidates are:
   - every battlefield permanent whose `CatalogAbilityKey` has a non-empty
     `TriggerDoublers` slot (a permanent under a CR 613.1f "loses all
     abilities" effect has no key, so it is skipped: the ability-removal case
     from the issue);
   - every card in the open simultaneous-exit batch (`g.simultaneousExit`) that
     is no longer on the battlefield, read through `CatalogAbilityKey(copy)` on
     the batch copy, exactly as `harvestSimultaneousExitLocked` reads watchers;
   - for `EventLTB`, the leaving card itself when it is not in a batch: its
     `CatalogKey` from its destination zone, skipped when the captured LKI has
     `AbilitiesRemoved`, exactly as `harvestLTB` does.

   Candidates are de-duplicated by instance ID, and a batch copy or LKI wins
   over a destination-zone card.
3. **The candidate list is safe to reuse across the event's matches.** Nothing
   between the first match and the last one can change the battlefield:
   `Build` must not mutate state (Decision 2 of this ADR), and queueing a
   prompt or an item moves no card. A nested event (a listener emitting
   during dispatch) builds its own pass.

What this gives, against the rulings:

- **A doubler entering with the triggering permanents counts** whenever the
  engine's entry path puts them on the battlefield before announcing them
  (`battlefield_put.go:330-346`). Panharmonicon's own entry counts for an
  "whenever an artifact enters" watcher, because Panharmonicon is on the
  battlefield when its own `EventETB` fires. Where an entry path announces one
  permanent at a time (token creation), a doubler that enters later in the
  same effect misses the earlier ones. That is the gap watchers already have
  on that path, and this design adds no new one.
- **A doubler leaving in the same event counts**: Teysa dying alone (her own
  LKI), Teysa in a wipe (her batch copy, still visible when a later death in
  the batch is announced), and the legend-rule case (the dying Teysa's LKI
  plus the surviving one on the battlefield, giving three).
- **The count is taken once, when the ability triggers.** A doubler that
  leaves while an instance's prompt is open does not remove that instance, and
  a doubler that arrives later adds none (the Annie Joins Up ruling).

### Decision 4 — `OncePerBatch`: checked once per match, never per instance

The in-flight check runs once, before any instance is dispatched
(`harvestMatchLocked` above). Every instance of the match, including the extras,
is then queued without consulting it. The next event of the same batch finds
the first instance in flight and is declined as it is today, extras included.
Isshin plus Adeline with three attackers is exactly two Adeline triggers.

The check itself is unchanged, so its purpose for a later batch is unchanged.
Whatever replaces it later (see the follow-up below about a batch that
arrives while the previous batch's item is still on the stack) also runs once per
match, because it sits in the same place.

**Declared limitation: the doubler count comes from the batch's first event.**
The engine has no batch identity on events: "whenever one or more" is modelled
as "the first event of the batch, while nothing from it is in flight". If a
doubler's filter only matches a later member of the batch, the extra instance
is missed. Example: Naban plus a "whenever one or more creatures you control
enter" ability, when a non-Wizard's `EventETB` is announced before a Wizard's
in the same batch. The error only ever undercounts, so it is never stronger than
printed. Every doubler whose filter is the same for every member of a batch
(Isshin and Wulfgar for attacks, Teysa for "one or more creatures die",
Panharmonicon for an all-creature batch) is exact. The implementing PR adds a
test that pins this limitation, so fixing it later shows up as a deliberate test
change.

### Decision 5 — An extra instance is an ordinary item with one attribution field

```go
// on StackItem
// DoubledBy names the CR 603.2d effect this instance exists because
// of: the doubler's instance ID and its name as it was when the
// ability triggered. Zero on the first instance and on every item
// that was not doubled.
DoubledBy     uuid.UUID
DoubledByName string
```

- **Not a copy.** `IsCopy` stays false. An extra instance is the ability
  triggering again (the rulings), so anything that asks "was this a copy"
  answers no, and the ability-copy seam (Strionic Resonator) stays separate.
- **`Label` and `Key` are identical to the first instance.** That keeps the
  three things that already key on them correct with no change:
  `triggerInFlightLocked`, `seatNeedsTriggerOrder` (identical instances don't
  prompt for order, as two Bident draws don't today), and
  `TurnTally.Triggered` (`turn_tally.go:349`), which counts every instance
  because the ability did trigger that many times.
- **The frames carry it.** `triggerResumeFrame` and `pickTargetFrame` gain a
  `doubledBy doublerRef`, and the item `Build` returns after the prompt is
  answered gets the field stamped on it, as `targetSpec` is stamped today.
- **Clone and snapshot.** `cloneStackItem` copies both fields.
  `stackItemSnapshot` gains `doubledBy` / `doubledByName` (omitempty), restore
  reads them, and the snapshot round-trip test gains a doubled item. There is
  no new `Game` state: the count is computed and never stored, and the harvest
  pass lives on the call stack. A file written before the fields existed
  decodes as "not doubled", which is correct. No schema bump.
- **Undo.** An extra instance is on `PendingTriggers` or `StackMeta` or in a
  prompt frame, and those are already cloned. Undo restores it the way it
  restores any trigger.
- **Wire (owner-decided, [question 1](#decided-2026-09-17)).** The field
  reaches the client as two additive fields, `doubled_by` / `doubled_by_name`,
  on `StackItemView` and on the `trigger_prompt` / `pick_target`
  `PendingChoiceView`. The stack overlay entry and the prompt show a tag, for
  example "Mulldrifter — draw two cards · additional (Panharmonicon)". There is
  **no log line**: the public log records changes to the game's outcome and
  turn structure, not per-object annotations, and it shows no triggers today.
  There is no reveal-strip cue either. The engine field also serves tests,
  which assert attribution, and the server log.

**Linked abilities.** "The exiled card" of a doubled linked ability means every
card any instance exiled. The catalog's shared record, `b27ExiledWith` (which
reads the event log by source; Angel of Serenity and Duplicant use it),
already returns all of them. Sub-PR 2 greps the catalog for card code that
stores a linked result in a single per-source slot, and each hit is fixed or
declared on its card.

### Decision 6 — Cause helpers make each card one line

`cards/effects/trigger_doubling.go`, a vocabulary file (ADR 0049 D1):

| helper | printed cause | catalog examples |
|---|---|---|
| `DoublesEntering(filter)` | "If a [filter] entering causes a triggered ability of a permanent you control to trigger" | Panharmonicon (artifact or creature), Yarok / Elesh Norn / Starfield Vocalist (permanent), Naban (Wizard you control), Ancient Greenwarden (land), Traveling Chocobo (land or Bird you control) |
| `DoublesLeaving(filter, diesOnly)` | "… dying causes …" / "… leaving the battlefield causes …" | Teysa Karlov, Drivnod (creature, dies only); Gandalf's second half |
| `DoublesAttacking(filter)` | "If a creature [you control] attacking causes …" | Isshin (any creature), Wulfgar (creature you control), Windcrag Siege's Mardu mode |
| `DoublesCasting(filter)` | "If you casting … causes …" | Veyran (instant or sorcery, the cast half) |
| `DoublesCombatDamageToPlayer(filter)` | "If a creature you control dealing combat damage to a player causes …" | Felix Five-Boots |
| `DoublesAbilitiesOf(filter, opts)` | "If a triggered ability of a [filter] you control triggers" | Katara (Ally), Annie Joins Up (legendary creature), Harmonic Prodigy (Shaman or another Wizard), Roaming Throne (another creature of the chosen type, `NamedTribe`), Cloud (Cloud or an Equipment attached to it, while equipped), Echoes of Eternity (colorless spell or another colorless permanent) |

Every helper:

- **checks the controller**: `q.SourceLKI.Controller == q.DoublerLKI.Controller`.
  The one exception is an opt-out for Cloud, whose text has no "you control";
- **refuses `FromSpell`** unless the option allows spells (Echoes of Eternity
  only);
- **combines with `AnyOf`**, for Gandalf's "entering or leaving".

**Every helper path ships with a real card** (owner policy, 2026-09-17). A
helper or option is built in the PR that ships the first card using it, never
ahead of one. In the first wave that is every row above except three pieces,
which wait for their card: `DoublesLeaving`'s "leaving" variant and `AnyOf`
(Gandalf the White), and `DoublesAbilitiesOf`'s spells option (Echoes of
Eternity, waiting on #666). Until then `FromSpell` is refused by every helper
without an option to turn it off.

The event cases are pinned as follows:

- **Entering**: `EventETB`, `EventZoneMove` into the battlefield, and
  `EventTokenCreated`. CR 701.7a defines creating a token as putting it onto
  the battlefield, and one ability watches one of these kinds, never two for
  the same entry.
- **Dying**: `EventLTB` or `EventZoneMove` from the battlefield to a graveyard,
  where the subject's LKI is a creature. `EventSacrifice` does not match (the
  Teysa ruling). A "leaves the battlefield" trigger on a creature that died
  does match, because it watches the same `EventLTB`.
- **Attacking**: `EventAttack` only. `EventTapCard` from attacking does not match
  (the Isshin ruling). Tokens created "tapped and attacking" were never
  declared as attackers and emit no `EventAttack`, which is correct.
- **Combat damage to a player**: a combat `EventDealDamage` whose `Target` is a
  seat and whose `Actor` is the doubler's controller. Life gained from
  lifelink is `EventChangeLife` and does not match (the Felix ruling).
- **Casting**: `EventCast` whose `Actor` is the doubler's controller and whose
  subject matches the filter.

### Decision 7 — Scope, stated

In scope: every triggered ability the harvester dispatches from an event (the
four harvest sites, including `FromStack`), plus evoke's sacrifice trigger.

Out of scope:

- **Trigger suppression**: Elesh Norn's third paragraph ("Permanents entering
  don't cause abilities of permanents your opponents control to trigger") and
  Torpor Orb. This is the "Ability-suppression static" row in
  `docs/engine-seams.md`. It would sit at the same point in `harvestMatchLocked`
  (a suppressor makes the count zero before the doublers are asked), so this
  design leaves room for it, but it does not build it.
- **Copying an ability item**: Strionic Resonator (the "Ability-copy" row).
- **The "or copying" half of Veyran**, until a spell-copy event exists.
- **Echoes of Eternity's "copy it" half**, which is #666 for a permanent spell.
- **Delayed and reflexive triggers** (CR 603.2d) and the manual
  `AnnounceTrigger` button (Decision 2).
- **Prompt batching for many identical optional instances.** Owner-decided
  ([question 3](#decided-2026-09-17)): each doubled optional instance gets its
  own yes/no prompt. A grouped "how many of these N?" prompt is built only if
  play shows the prompt count matters, and it would need its own design.

No bot legal-move change: extra instances produce more prompts of kinds the
enumerator and the heuristic already answer (`trigger_prompt`, `pick_target`,
`trigger_order`).

### Consequences

- **Doubled triggers are rules-correct for every catalog trigger at once**,
  including ward (a catalog `TriggeredAbility`, so Katara doubles a warded
  Ally's ward trigger and the caster pays twice), Saga chapters under a
  source-keyed doubler, and evoke's sacrifice.
- **More prompts.** Every optional or targeted instance asks separately. With
  two Panharmonicons and a "you may" watcher, five tokens entering are fifteen
  prompts. The owner accepted this for the first release (question 3); grouping
  waits for evidence from play.
- **The loop breaker trips sooner on big doubled turns.** ADR 0055 counts
  resolutions per `(source, label)` since the last decision, and threshold 25
  was "unreachable by accident". Nine creatures entering together under two
  doublers, with one mandatory "whenever a creature enters" watcher, is 27
  resolutions with no prompt answered. The notice only suspends autopass, and
  the first pass clears it. This is accepted and recorded here, and the threshold
  is not changed for it.
- **Harvest cost**: at most one extra battlefield walk per event that matched
  a trigger (lazy, memoised per event). `BenchmarkHarvestETBEvent_80perms`
  (`cards/effects/bench_test.go:84`) is re-run in sub-PR 1 and the numbers go
  in the PR description.
- **`dispatchTriggerLocked` keeps its name and meaning** for the reflexive
  path. The harvest callers move to `harvestMatchLocked`.
- **Cloud, Midgar Mercenary loses its caveat** in sub-PR 2.
- **`docs/engine-seams.md`**: the "Trigger doubling" row moves to Closed when
  sub-PR 2 lands (the seam and its first real cards reach develop together). The cards still waiting on a second seam (Elesh Norn's
  suppression, Echoes' copy, Gandalf's flash permission, Ancient Greenwarden,
  Traveling Chocobo's #765) move to that other seam's row.

### Alternatives considered

- **Count inside `dispatchTriggerLocked`** (the issue's suggestion). Since
  #636 this would double reflexive triggers, which CR 603.2d excludes, and it
  would need a "don't double" flag that every future dispatch caller would have
  to remember. Rejected.
- **Declared causes (an enum `Entering | Dying | Attacking | …`) instead of a
  predicate.** It doesn't cover Gandalf, Cloud or Roaming Throne without a
  second escape hatch, and the helpers give the one-line cards the same
  brevity. Rejected.
- **A `Multiplier` field on the item** (one item, resolved N times). This
  breaks every ruling about separate choices: a separate target, a separate
  "you may", and separate responses (a Stifle removes one instance, not all
  of them). Rejected.
- **Cloning the built item N times after `Build`.** Build runs once, so
  per-instance targets and optional answers are impossible, and an item whose
  `Build` depends on announce-time state would be shared. Rejected in favour
  of N full dispatches.
- **Bypassing `OncePerBatch` by giving extras a different `Key`.** The next
  event of the batch would no longer see the extras as in flight, the ordering
  prompt would treat identical instances as different, and `TurnTally` would
  split one ability into two keys. Rejected.
- **Doublers as `StaticAbility` in the layer system.** Nothing is applied in a
  layer, and the harvester would have to read layer output that has no reason
  to exist. Rejected.
- **An eager per-game doubler index** kept up to date on every zone change.
  It is faster only for events that match a trigger, and it is a second source
  of truth for "what is on the battlefield" that undo and snapshot would have
  to carry. The lazy per-event scan is enough. Rejected for now.

### PR split

**Sub-PR 1 — engine, no cards.**
- `game/trigger_doubling.go`: `TriggerDoubler`, `TriggerDoublingQuery`,
  `harvestPass`, `triggerDoublersLocked`, `harvestMatchLocked`,
  `dispatchTriggerInstanceLocked`.
- `CardDef.TriggerDoublers`, `CatalogTriggerDoublers`, `Spec.TriggerDoublers`,
  `buildDef`.
- The four harvest sites move to `harvestMatchLocked`. `OnEvent` builds the
  pass and captures the subject before `harvestLTB`'s delete.
- Evoke's sacrifice trigger goes through the count.
- `StackItem.DoubledBy` / `DoubledByName`, the frames, `cloneStackItem`,
  `stackItemSnapshot`.
- Engine tests 1–14 below, with doublers stubbed through `CatalogTriggerDoublers`.
- The benchmark re-run.
- **Merges only with sub-PR 2.** Sub-PR 2 is opened against this PR's branch,
  and the two reach develop together, so no develop build carries the seam
  without a real card using it (owner policy, 2026-09-17).
- Checks: `go test ./internal/game/... ./internal/ws/... ./internal/aiseat/... ./internal/legal/...`,
  then `go test ./...` and `make lint`.

**Sub-PR 2 — vocabulary and reference cards.**
- `cards/effects/trigger_doubling.go` (Decision 6).
- Panharmonicon, Teysa Karlov and Isshin, Two Heavens as One, each with the
  rulings above as card tests. Cloud's caveat removed. Panharmonicon's tests
  include an evoked Mulldrifter, so the evoke-sacrifice path ships with a real
  card.
- The Saga chapter I reading checked (Decision 2) and the linked-state grep
  (Decision 5).
- Census regenerated (`go test ./internal/cards/coverage -update`), and the
  clone-baseline gate passes.
- The engine-seams row moves to Closed.

**Sub-PR 3 — wire and client** (owner question 1, option (a)).
- `StackItemView.doubled_by` / `doubled_by_name`, and the same on the
  `trigger_prompt` / `pick_target` `PendingChoiceView`. Update
  `docs/protocol.md` and `client/src/lib/protocol.ts`.
- The stack overlay and `ChoicePromptModal` show the tag
  "· additional (Panharmonicon)" after the item's label. No log kind and no
  reveal-strip cue.

**Card PRs** — the rest of the first wave (owner question 2, option (b)):
Katara, Naban, Yarok, Annie Joins Up, Starfield Vocalist, Harmonic Prodigy,
Roaming Throne, Felix Five-Boots and Wulfgar of Icewind Dale, plus Elesh Norn,
Mother of Machines and Veyran, Voice of Duality with declared caveats (see
[Cards](#cards)). Each card follows AGENTS.md: completeness declared, caveats
weaker than printed and never stronger.

### Test plan

Engine (`internal/game`, doublers stubbed):

1. **One doubler, mandatory trigger**: two items on `PendingTriggers`, same
   `Label`, the second with `DoubledBy` set, and two `EventTrigger`.
2. **Stacking**: two doublers give three instances, three give four.
3. **The query carries the right controllers**: `SourceLKI.Controller` is the
   player who controlled the source when it triggered, including a stolen
   creature that dies (read off its LKI), and `DoublerLKI.Controller` is the
   doubler's. The helpers' "you control" check is tested on the cards
   (test 15).
4. **Optional instances**: two `trigger_prompt`s. Declining one leaves the
   other, and each "yes" builds its own item.
5. **Targeted instances**: two `pick_target`s with independent picks. With an
   empty legal set, both are dropped with no prompt (CR 603.3d).
6. **`OncePerBatch`**: three `EventAttack` in one batch under one attack
   doubler gives exactly two instances. A second batch after both resolve
   triggers again.
7. **`OncePerBatch` limitation pinned**: a non-matching first event followed
   by a matching one gives one instance (Decision 4).
8. **Doubler dying alone**: a stubbed "creature dying" doubler on a creature
   destroyed by `DestroyPermanentForEffect`, with a dies-watcher on the
   battlefield, gives two watcher instances.
9. **Doubler in a wipe**: doubler, watcher and a third creature all destroyed by
   `DestroyPermanentsForEffect` give two instances per death for the watcher,
   including the death announced after the doubler's own.
10. **Legend rule**: a second copy of the doubler enters, one goes to the
    graveyard, and a dies-watcher of the dying copy triggers three times.
11. **Doubler entering in a batch**: `putOntoBattlefieldFromZoneLocked` with
    the doubler and two creatures, and a "whenever a creature enters" watcher,
    doubles all three announcements.
12. **Ability removal**: a doubler with `AbilitiesRemoved` adds nothing, both
    live and in LKI.
13. **Exclusions**: a reflexive trigger from a doubled source's resolution, a
    delayed trigger scheduled by a doubled source, a warp exile, and a manual
    `AnnounceTrigger` each queue exactly one item. Evoke's sacrifice under an
    entry doubler queues two.
14. **Clone and snapshot**: a doubled item on `PendingTriggers`, one on
    `StackMeta`, and one inside a `pick_target` frame survive
    `Clone`/`RestoreFrom` with `DoubledBy` intact. The snapshot round trip
    keeps the two fields, and a snapshot without them decodes as not doubled.

Cards (`internal/cards/effects`):

15. **Panharmonicon**: an ETB doubled; its own entry doubles an "artifact
    enters" watcher; a land entering is not doubled; a creature entering under
    an opponent's control doubles my watcher; an opponent's Panharmonicon
    doesn't double my watcher; "enters with a counter" is not doubled.
16. **Teysa Karlov**: a dies trigger doubled; "whenever you sacrifice" not
    doubled; "leaves the battlefield" of a dying creature doubled; an artifact
    creature dying doubles "whenever an artifact is put into a graveyard from
    the battlefield"; Teysa dying herself doubles Blood Artist.
17. **Isshin**: "whenever this attacks" doubled, and "becomes tapped" from
    attacking not doubled.
18. **Cloud, Midgar Mercenary**: an attached Equipment's combat-damage trigger
    doubled while equipped, and not doubled once the Equipment is moved.
19. **Elesh Norn, Mother of Machines**: an entry doubled for her controller's
    watcher, and an opponent's entry watcher still triggers (the declared
    caveat, pinned so the suppression seam changes it deliberately).
20. **Veyran, Voice of Duality**: a magecraft trigger from casting an instant
    doubled; the "copy" half is the declared caveat.
21. **Catalog soak** (#601) with the first-wave cards added: no stalls on the
    extra prompts.
22. **Clone-baseline gate**: no new duplicate bodies (the helpers exist so
    that holds).

Client (sub-PR 3): the overlay tag and the prompt tag are checked by hand
until #689, and any pure helper (the tag text) is unit-tested with vitest.

### Cards

Oracle text checked against the Scryfall dump on 2026-09-17. "Doubler only"
means that on a read of the full oracle text, every other clause already has an
engine seam.

**Doubler only, or doubler plus clauses that already exist:**

| card | the doubling text | other clauses |
|---|---|---|
| Panharmonicon | "If an artifact or creature entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time." | — |
| Teysa Karlov | "If a creature dying causes a triggered ability of a permanent you control to trigger, …" | "Creature tokens you control have vigilance and lifelink." |
| Isshin, Two Heavens as One | "If a creature attacking causes a triggered ability of a permanent you control to trigger, …" | — |
| Katara, the Fearless | "If a triggered ability of an Ally you control triggers, …" | — |
| Naban, Dean of Iteration | "If a Wizard you control entering causes a triggered ability of a permanent you control to trigger, …" | — |
| Yarok, the Desecrated | "If a permanent entering causes a triggered ability of a permanent you control to trigger, …" | deathtouch, lifelink |
| Annie Joins Up | "If a triggered ability of a legendary creature you control triggers, …" | ETB: 5 damage to target creature or planeswalker an opponent controls |
| Starfield Vocalist | "If a permanent entering the battlefield causes a triggered ability of a permanent you control to trigger, …" | warp {1}{U} |
| Harmonic Prodigy | "If a triggered ability of a Shaman or another Wizard you control triggers, …" | prowess, written as a cast trigger (no `prowess` helper exists yet) |
| Roaming Throne | "If a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time." | ward {2}, "as this creature enters, choose a creature type", "is the chosen type in addition" |
| Felix Five-Boots | "If a creature you control dealing combat damage to a player causes a triggered ability of a permanent you control to trigger, …" | menace, ward {2} |
| Wulfgar of Icewind Dale | "If a creature you control attacking causes a triggered ability of a permanent you control to trigger, …" | melee, written from `EventAttack` (the issue's borderline card) |

**Doubler plus a clause another seam owns:**

| card | the other clause | waits on |
|---|---|---|
| Elesh Norn, Mother of Machines | "Permanents entering don't cause abilities of permanents your opponents control to trigger." | ability-suppression static (no issue yet). **Ships in the first wave** with the suppression declared as a caveat (owner question 2) |
| Veyran, Voice of Duality | "cast **or copy**", both halves | a spell-copy event. **Ships in the first wave** with the copy half declared as a caveat (owner question 2) |
| Echoes of Eternity | "Whenever you cast a colorless spell, copy it." | #666 |
| Gandalf the White | "You may cast legendary spells and artifact spells as though they had flash." | the per-player flash row |
| Ancient Greenwarden | "You may play lands from your graveyard." | play from graveyard |
| Traveling Chocobo | "You may play lands and cast Bird spells from the top of your library." | #765 |
| Windcrag Siege | "As this enchantment enters, choose Mardu or Jeskai." | an as-enters mode choice (the issue calls it misattributed) |

**Noticed while checking, not in the issue's list:** Drivnod, Carnage Dominus
has Teysa's exact text. Its other ability ("{B/P}{B/P}, Exile three creature
cards from your graveyard: Put an indestructible counter on Drivnod.") is not
checked here, and the card goes to whichever batch issue owns it.

### Decided (2026-09-17)

Answered by the owner on 2026-09-17. The options are kept, the chosen one is
marked **(chosen)**, and the recommendation text is kept for the record.

1. **How does a player see that a trigger is an extra instance?** Today the
   public log doesn't show triggers at all (`protocol/log.go` projects no
   `EventTrigger`). Players see a trigger as its stack-overlay entry and, when
   it asks something, its prompt.
   - (a) **(chosen)** A tag on the stack-overlay entry and on its prompt, for example
     "Mulldrifter — draw two cards · additional (Panharmonicon)". This needs
     two additive wire fields and no log kind.
   - (b) (a) plus a public log line per extra instance.
   - (c) Nothing: the instances look identical.

   **Recommendation: (a).** A player's question is "why is this on the stack
   twice?" and they ask it while looking at the stack or the prompt, so the
   answer belongs there. A log line for extra instances only, in a log that
   shows no triggers at all, would stand out for the wrong reason. (c) makes
   Panharmonicon look like a bug.

   **Decision: (a).** No log line, under the cross-ADR log rule (the log
   records changes to the game's outcome and turn structure, not per-object
   annotations), and no reveal-strip cue. Applied in Decision 5 and sub-PR 3.

2. **Which cards ship in the first wave?**
   - (a) The twelve "doubler only" cards above, plus Cloud's caveat removal.
   - (b) **(chosen)** (a) plus Elesh Norn, Mother of Machines and Veyran, Voice of Duality,
     each with a declared caveat for the missing half. Both caveats are weaker
     than printed.
   - (c) The seam plus the three reference cards (Panharmonicon, Teysa,
     Isshin) and Cloud, leaving the rest to the batch issues.

   **Recommendation: (b).** Both omissions only weaken their controller, which
   AGENTS.md allows once declared, and Cloud already set that precedent. Elesh
   Norn's doubling half is still a strong commander on its own. The cost is
   that a player who picks Norn for the hoser half doesn't get it until the
   suppression seam exists, and the caveat has to say so plainly.

   **Decision: (b).** Elesh Norn and Veyran ship with their caveats declared
   (tests 19 and 20 pin them). Under the owner's cross-cutting policy every
   new seam path ships with at least one real card, so sub-PRs 1 and 2 reach
   develop together and helpers without a first-wave card wait for it
   (Decision 6).

3. **Many identical optional instances: one prompt each, or grouped?** Two
   Panharmonicons and a "you may" watcher turn five tokens entering into
   fifteen yes/no prompts.
   - (a) **(chosen)** One prompt per instance, as the rules describe. This ships with
     sub-PR 1 and needs no new prompt kind.
   - (b) Group untargeted identical optional instances (same source, same
     label, same event) into one "how many of these N?" prompt. This is a new
     prompt shape with enumerator and bot work.

   **Recommendation: (a) now, and (b) only if play shows the spam matters.**
   Tokens entering already give one prompt per token today, so doubling
   multiplies an existing cost without creating a new one. (b) is a real
   feature with its own ADR-sized questions (what happens when the instances
   differ in `NoLegalTarget`, and how bots answer a count).

   **Decision: (a).** One yes/no prompt per doubled optional instance.
   Grouping is considered only if play demands it.
