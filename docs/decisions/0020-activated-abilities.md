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

## Addendum (#625): counters as a cost

The "counters as a cost" line under *Out of scope* is now in scope, in
part. `AbilityCost.RemoveCounters` is a `game.CounterRemovalCost{Counter,
N, From}`: remove N counters as part of paying the cost. It covers three
printed shapes, and nothing else.

| Shape | Printed | Declared as |
|---|---|---|
| self | "Remove a gold counter from this artifact" | `From == nil` — `RemoveCountersFromThis("gold", 1)` |
| other permanent | "remove a loyalty counter from a planeswalker you control" | `From` is a `TargetSpec` — `RemoveCountersFrom("loyalty", 1, "a planeswalker you control", Planeswalker())` |
| any kind | "Remove a counter from a creature you control" | `Counter == ""` — `RemoveCountersFrom("", 1, "a creature you control", Creature())` |

**The choice is made at announce.** The permanent rides
`ActivateAbilityParams.CounterSourceIDs` (omitted for the self form) and,
for the any-kind form, the kind rides `CounterKind`, beside the
sacrifice and crew picks. No pending prompt is involved.

**§2 still holds: it is a predicate, not targeting.** `From` reuses
`TargetSpec` the way `SacrificeOther` does, and the engine matches it
with `specMatchLocked(..., false)`. A hexproof or shrouded permanent you
control pays. "You control" is enforced by the engine, not the spec.
`game.CounterCostOptionsForEffect` is the single candidate walk (built on
`SpecCandidatesForEffect`, not `LegalTargetsForEffect`) that the
protocol view and the move enumerator both read, so the client, the bots
and the engine agree on which permanent pays.

**§3 still holds: validate everything, then pay everything.**
`validateCounterRemovalCostLocked` runs in the validate block beside the
crew check: exactly one permanent for the other form, controlled by the
activator, matching `From`, holding at least N of the kind (else
`ErrInsufficientCounters`). A bad target, sacrifice or crew ID on the
same activation removes no counter. Payment order is now
mana → tap → crew → life → loyalty → **counters** → sacrifice. Counters
come off before sacrifices so that a self-form removal on a source that is
also sacrificed still finds the source.

Three rules come with the component and are enforced by the engine rather than by each card:

- **Not replaceable.** The removal goes through `applyCounterLocked`, not
  `AddCounterForEffect`. It is a cost, not an effect, so nothing that
  doubles or modifies counters applies. This is the loyalty cost's rule.
- **Not a loyalty activation.** Removing a planeswalker's loyalty counter
  to pay another permanent's cost does not stamp
  `LoyaltyActivatedThisTurn` and has no sorcery-speed window. The walker
  can still activate its own loyalty ability that turn. A walker paid
  down to 0 dies to the CR 704.5i state-based action the activation
  already runs, with the ability on the stack.
- **No timing of its own.** Heart of Kiran's crew-by-counter is instant
  speed, like crew.

**A 0/0 that pays with its last counter dies, and a `*` creature does
not, by the toughness rule in [ADR 0007 §7](0007-stack-foundation.md)
as amended by #683, not by a rule of the cost's.** The payment goes
through `applyCounterLocked` like any other counter removal, and the
state-based check the activation already runs applies that rule with
the ability on the stack. What the cost adds is that the case becomes
ordinary play: Mikaeus, the Lunarch paying his team pump with his last
counter, or Fain, the Broker spending a 0/0's last counter. A card cast
for X=0 never had a counter, so it is still skipped; that gap is noted
on the X cards, and declared to players on Mikaeus.

**"Rather than pay" on an activated ability is a second ability entry,
not an alternatives slot.** Heart of Kiran lists "Crew 3" and "Crew —
remove a loyalty counter from a planeswalker you control" as two
abilities with the same effect. The client already lists abilities
separately. What #259 requires is that the second entry has a real
cost, and it does: it cannot be activated without a planeswalker that
holds a counter.

`effects.Register` panics on `N <= 0`, and on an any-kind cost with
`N > 1`, because one kind choice cannot say how "remove two counters"
was paid when the two could be different kinds. `effects.Plus` merges
the field. Without that, `Plus(TapCost(), RemoveCountersFromThis(...))`
would silently drop the counter and the ability would be free.

Bots see the price on `legal.MoveCost.Counters` (`{card_id, counter,
n}`), priced against the permanent the counters come off, not the
move's source. That follows the #74 and #547 precedent: a cost the wire
payload cannot name rides the move.

**Still out of scope:**

- A removal **split across several permanents**: Iron Spider, Stark
  Upgrade's "Remove two +1/+1 counters from among artifacts you control".
  One permanent per payment cannot express it, and the card keeps its
  caveat.
- A cost that **adds** a counter: Devoted Druid's "Put a -1/-1 counter on
  this creature".
- Counter costs on **mana abilities** (`ManaAbilityCost`).

`docs/engine-seams.md`'s counter-cost row lists what is still waiting.
