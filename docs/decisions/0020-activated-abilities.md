# ADR 0020 — Activated abilities in the catalog

**Status:** accepted (S21 sub-PR 2)
**Supersedes:** nothing. Extends [ADR 0018](0018-triggers-on-the-stack.md).

## Context

By the end of S20 the catalog could express four of the five ways a
card does something: spells resolve (`OnResolve`), triggers fire
(`Triggered`, S19), statics apply continuously (`Static`, S16) and
replacements rewrite events (`Replacements`, S17). The fifth —
CR 602 activated abilities — did not exist.

`Game.ActivateAbility` accepted a card ID and a free-text label and
pushed a stack item with no cost and no effect. Players paid the
cost by hand (tapping the permanent themselves, dragging the
sacrificed creature to the graveyard) and resolved the effect in
their heads. That was tolerable for "{T}: add mana" — which has its
own path anyway — and untenable for the S21 goal, because an
aristocrats deck IS its sacrifice outlets. "Sacrifice a creature:
deal 1 damage" isn't an effect with a cost attached; the cost is
half the card, and the dies-triggers it causes are the whole
strategy.

## Decisions

### 1. One shape, mirroring the rest of the catalog

`Spec.Activated []ActivatedAbility`, converted at the wire hook into
`game.ActivatedAbilityShape` — the same import-cycle dodge
`ManaAbilityShape` uses. An entry is a label, a cost, an optional
target clause, a sorcery-speed flag and an `Effect` closure.

The effect is the *same* `func(g *Game, item *StackItem) error`
S19 gave triggered abilities. That was the whole point of ADR 0018's
closure: resolution was already written. `resolveTopAbilityLocked`
runs the CR 608.2b re-check and then the effect, and it cannot tell
an activated item from a triggered one.

### 2. Costs are a struct, not a string

`AbilityCost{Tap, SacrificeSelf, SacrificeOther, Mana, Life}` —
additive fields rather than a parsed cost line. Parsing "{2}, {T},
Sacrifice a creature:" from oracle text would be a second
mini-language to maintain against the same 20-odd shapes the
constructors already cover (`TapCost()`, `SacrificeACreature()`,
`Plus(...)`).

`SacrificeOther` reuses `TargetSpec` for its "what may I sacrifice"
predicate, because the shape is identical. It is emphatically not
targeting: a sacrifice cost doesn't target, so hexproof and
"can't be the target of" never apply to it, and the client picks it
from a plain list instead of the board-click targeting flow. The
controller restriction (CR 701.21a) lives in the engine's cost
validation rather than the spec.

### 3. Validate everything, then pay everything

Every cost component is checked before any is paid. A tapped Krenko
fails without sacrificing; an illegal target fails without tapping.
This is the same discipline S21 sub-PR 1 applied to sacrifice-cost
mana abilities, and it matters more here because the costs are
heterogeneous — a partial payment would leave the board in a state
no rule describes.

Payment order within the commit is mana → tap → life → sacrifice.
Sacrifices go last because they move cards, which invalidates the
`*Card` pointer the tap needed.

### 4. Costs are paid at announce, so their triggers sit above

Paying a cost happens during activation (CR 601.2h), not at
resolution. So sacrificing a creature to Goblin Bombardment puts the
creature's dies-trigger on the stack *above* the Bombardment
ability, and it resolves first. Blood Artist drains before the
damage lands. The engine gets this for free by running the state
checks at the end of activation, after the item is already on the
stack — but the ordering is load-bearing for the whole archetype, so
there's a test that asserts it directly rather than trusting the
side effect.

### 5. The free-form path stays

`ActivateAbility` (label-only, no cost, no effect) still handles
non-catalog cards, which is most of the corpus. The action layer
picks the path by whether the payload names an `ability_index`. A
card gains a real ability by being added to the catalog, not by a
migration.

## Consequences

- Sac outlets work end to end: Goblin Bombardment, Carrion Feeder.
- Tap abilities enforce summoning sickness (CR 302.6), which the
  free-form path never did — Krenko can't tap the turn he lands.
  `CardView.summoning_sick` now rides the wire so the menu can grey
  the entry instead of failing the click.
- The activation menu is the existing right-click popover, with the
  activated abilities listed under the mana abilities.

## Out of scope

- **Mana abilities with a non-self sacrifice cost** (Ashnod's Altar:
  "Sacrifice a creature: Add {C}{C}"). It's a mana ability, so it
  must not use the stack (CR 605.3b), but its cost needs the
  sacrifice picker this ADR builds for stack-using abilities. It
  fits neither surface cleanly; it wants a third path where the
  activation collects a cost choice and then resolves immediately.
- **X in an activation cost** ({X}: …), **counters as a cost**
  (`Remove a +1/+1 counter:`), and **"activate only once each
  turn"** — all wanted by cards further down the S21 list, none
  needed by the three cards here.
- **Loyalty abilities** keep their own `ActivateLoyalty` path.
