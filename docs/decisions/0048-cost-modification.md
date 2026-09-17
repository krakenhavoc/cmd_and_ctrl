# ADR 0048 — Cost modification and the free-spell family

**Status:** accepted (S28)
**Extends:** [ADR 0011](0011-mana-pool-and-auto-tapper.md) (the cost
computation engine), [ADR 0021](0021-additional-costs.md) and
[ADR 0025](0025-alternative-costs.md) (the two cost slots that came
before this one), [ADR 0012](0012-layer-system.md) (the pattern this
deliberately does *not* use).

ADR numbering: scanned every remote branch per AGENTS.md §4 at the
time of writing — `0001`–`0004`, `0006`–`0023`, `0025`–`0028`,
`0031`–`0041` were taken; `0005`, `0024`, `0029`, `0030` are
permanently retired. `0042` was the first free number.

## Context

Six cards in the S28 scope make other people's spells more
expensive, three make the controller's own cheaper, and one sets a
floor:

> Spells cost {1} more to cast. — Sphere of Resistance
> Noncreature spells cost {1} more to cast. — Thalia, Thorn of Amethyst
> Artifact and enchantment spells your opponents cast cost {2} more. — Aura of Silence
> Each spell a player casts costs {1} more for each other spell that player has cast this turn. — Damping Sphere
> Instant and sorcery spells you cast cost {1} less to cast. — Goblin Electromancer
> Creature spells you cast cost {2} less to cast. — Heartless Summoning
> Creature spells you cast cost {1} less for each +1/+1 counter on Animar.
> As long as this artifact is untapped, each spell that would cost less than three mana to cast costs three mana. — Trinisphere

The engine already modelled four kinds of cost, and none of them is
this:

| kind | belongs to | landed |
| --- | --- | --- |
| mana cost | the spell | S15 |
| activated-ability cost | the ability | S21 sub-PR 2 |
| additional cost | the spell | S21 sub-PR 5 ([ADR 0021](0021-additional-costs.md)) |
| alternative cost | the spell | S22 ([ADR 0025](0025-alternative-costs.md)) |

Every one of those is a property of the object being cast or
activated, looked up by its oracle ID at announce. A cost modifier
is a property of **a permanent somebody else controls**, consulted
every time anybody announces anything. That inversion is the whole
of this ADR: the cast path stops being handed its cost and has to go
looking for the modifiers, the way the layer engine goes looking for
static abilities.

## Decisions

### 1. Not a layer, and not `Spec.Static`

The reflex is "a continuous effect from a battlefield permanent —
that is CR 613, that is `Spec.Static`". It isn't, and the signature
says so: `game.StaticAbility.Apply` takes a `*Characteristic` and a
target `*Card`, because every layer effect changes a characteristic
of an object on the battlefield.

A cost modifier changes neither. It changes what a player *pays* for
an object that is in their hand and has no `Characteristic` at all,
and CR 202.3c is explicit that the object's mana value is **not**
changed by any of it — a Goblin Electromancer does not turn
Counterspell into a one-drop for Imoti, Tarmogoyf or Tribute Mage.
There is no layer for the effect to sit in, and inventing one would
mean inventing a characteristic to hold it.

So the engine **derives** the answer, which is the same call
`NoMaxHandSize` made for the same reason (#338): the cast path asks
the battlefield for every modifier in play and prices the spell
through them. Nothing is written anywhere, so nothing has to be
unwound when the permanent leaves, and two Spheres of Resistance
with one of them leaving comes out right with no bookkeeping.

`Spec.CostModifiers []game.CostModifier`, wired through
`CatalogCostModifiers` exactly as `Static`, `Replacements` and
`Triggered` are.

### 2. Three kinds, applied in three passes

```go
const (
    CostIncrease  CostModifierKind = iota  // "costs {1} more"
    CostReduction                          // "costs {2} less"
    CostFloor                              // Trinisphere
)
```

CR 601.2f fixes the order, and the order is observable. Sphere of
Resistance plus Goblin Electromancer on a Lightning Bolt is
{R} → {1}{R} → {R}: the increase is what gives the reduction a
generic symbol to eat. Run the reduction first and the Bolt costs
{1}{R} — a wrong answer that every "did it get cheaper?" test
passes. The engine makes two separate passes over the same modifier
list rather than one pass with a sort, because the clamping in pass
two has to see the whole of pass one.

`CostFloor` is a third pass rather than a conditional increase
because Trinisphere's own ruling puts it after everything: "each
spell that would cost less than three" can only mean the price the
spell actually ends up at.

### 3. The floor is zero generic mana, and colours are untouched

A reduction spends **generic** mana and stops at zero. It never
touches a coloured requirement, an overshoot does not carry into the
next reduction, and there is no credit. Heartless Summoning on a {B}
Carrion Feeder leaves {B}.

This is the single most mis-implemented rule in the family, so
`reduceGeneric` is the only function in the engine that knows how to
spend a reduction, and hybrid / Phyrexian requirements go through it
untouched — they are coloured requirements with an escape hatch, not
generic mana.

(The S28 brief asked for "a floor of {1} per CR 117.13". There is no
such rule — CR 117 is timing and priority — and a {1} floor would
make Heartless Summoning leave a {1}{B} creature at {1}{B}, which is
wrong on the card. The rule implemented is CR 601.2f's: generic
only, floor zero.)

### 4. A modifier that can't be priced refuses the cast

`CostModifier.Amount` is a `func(CostQuery) int`, because three of
the ten cards scale with the board (Animar counts counters, Damping
Sphere counts spells). A hook that returns a negative number returns
`ErrCostModifier` and the **cast is rejected** — no clamp, no
fall-back to the printed cost.

This is #289's lesson applied one cost component over. That bug let
an unparseable cost silently make ~1,047 cards free, and the fix was
to refuse rather than to guess. A reducer with a sign error is the
same failure wearing a different hat: the thing that must never
happen on this path is a spell that comes out cheaper than it
should. The card stays in hand and the player is told which
permanent misbehaved.

### 5. Layered after the alternative cost and the commander tax,
### before convoke

`effectiveCostLocked`'s order is now:

1. printed mana cost, or the alternative cost claimed at announce
   (CR 601.2e / CR 118.9);
2. an exile-play grant's cost override (airbend);
3. commander tax (CR 903.8);
4. **cost modifiers — increases, reductions, floor (CR 601.2f)**;
5. convoke / waterbend subtraction.

Steps 1–3 settle what the spell "would cost", which is the number
every modifier is written against: Thalia taxes an overloaded
spell's *overload* cost, and Trinisphere looks at the taxed
commander's total rather than the corner of the card. Step 5 stays
last because tapping creatures is a way of *paying* a total cost,
and CR 601.2f settles the total before anything is paid against it.

### 6. Every pricing surface goes through one function

Three call sites priced casts before this sprint and all three now
consult the modifiers:

- `effectiveCostLocked` — the cast itself and the auto-tapper;
- `GET /games/{id}/autotap-preview` (`internal/lobby/http.go`) — or
  the preview would plan a tap for the printed price and the cast
  that follows would come up short, exactly when the player most
  needs the two to agree;
- the S31 legal-move enumerator (`internal/legal/cast.go`) — or the
  bot enumerates casts the engine then rejects for insufficient
  mana, and a bot that keeps picking rejected moves stalls.

The last two reach it through `ApplyCostModifiers` /
`ApplyCostModifiersForEffect`, the RLock-wrapped and
already-under-the-lock pair, following `AutoTapForCost*`'s existing
convention.

## Consequences

**Good.** Every cost-modifying card in the game is now a data
declaration in a card file — predicate plus amount — with no engine
change. The nine-card S28 group and the whole Stax archetype behind
it (Winter Orb aside, which is a different mechanism) are catalog
work from here.

**The cost.** Every cast now walks the battlefield. That is O(board)
per cast with a nil-return fast path for cards that declare nothing,
which at a four-player table is tens of pointer comparisons — but it
is a new per-cast cost where there was none, and it is paid on the
snapshot path too because the legal-move enumerator prices every
enumerable cast.

**Not covered.** Modifiers from outside the battlefield (there are
none in scope; the scan is one `for` loop to widen), modifiers that
reduce coloured requirements (no card does — CR 601.2f is why), and
per-player duration-scoped modifiers like Will Kenrith's "until your
next turn, that player's spells cost {2} less", which needs a
floating registry keyed by player rather than a battlefield source.
[ADR 0035](0035-until-end-of-turn-effects.md)'s turn-scoped registry
is the shape that would host it; its clock is one turn too short.

---

## Addendum: the rest of S28

Three more decisions landed in the same sprint. They are recorded here
rather than in ADRs of their own because each is an EXTENSION of
machinery that already had one — and saying so is the point: two of
the four cost mechanisms the sprint's issue asked for already existed.

### 7. The free-spell family extends `AlternativeCost`, it does not
### replace it

Force of Will, Solitude, Snuff Out, Daze and Fierce Guardianship all
say "rather than pay this spell's mana cost", which is
[ADR 0025](0025-alternative-costs.md)'s CR 118.9 clause verbatim — the
same one overload, evoke and cleave have ridden since S22. What they
needed was not a parallel mechanism but two new components on the
existing struct:

- **`Condition`** — "if you control a commander" (the Commander
  Legends cycle), "if you control a Swamp" (Snuff Out). Checked at
  announce AND in the view, so an offer the caster cannot take is
  never shown rather than shown and rejected.
- **Non-mana payments** — `Life`, `ExileFromHand`, `ReturnToHand`,
  named on the wire by a new `alt_cost_ids`. Validated together
  before any of them is paid, so a player at 1 life with no blue card
  gets a rejected cast rather than a dead player and a spell still in
  hand.

The payments are charged in the CR 601.2h window, with the spell
**already on the stack** — the same window the additional cost uses
and for the same reason. Countering a Force of Will does not refund
the pitched card; a Daze returns its Island before the spell it is
answering resolves. A resolution-time implementation would get both
backwards and would make Force of Will a strictly better card than
the printed one.

`alt_cost_ids` is deliberately a separate wire field from
`discard_ids` / `sacrifice_ids` rather than a shared "cards paid"
list: those pay ADDITIONAL costs, which are charged whichever cost
the caster chose, while this one is part of the alternative cost and
vanishes when the offer is declined. One list would make the two
indistinguishable.

Pact of Negation is in the sprint's "free-cast" group and needs none
of this — its printed cost really is {0}. What it has is a debt, and
the debt is an ordinary CR 603.7 delayed trigger
([ADR 0026](0026-delayed-triggers.md)).

### 8. Cascade needed a triggered ability of a SPELL

CR 702.85a fires "when you cast this spell" — while the source is on
the stack, which is not where the S19 trigger harvester looks
([ADR 0018](0018-triggers-on-the-stack.md) scans the battlefield,
because that is where abilities exist under CR 113.6).

`TriggeredAbility.FromStack` is an **opt-in flag** plus a scan
narrowed to `EventCast` and to the one card the event names. A
blanket stack scan was the obvious alternative and is wrong: a
creature spell on the stack carries its permanent's declared triggers
with it, so "whenever another creature enters" would start firing
from the stack — a turn early, from a zone where the ability does not
exist.

### 9. Cascade's free cast is a GRANT, not an inline cast

Printed cascade casts the exiled card then and there, inside the
trigger's resolution, ignoring timing. The engine instead stamps the
impulse-exile permission ([ADR 0022](0022-impulse-exile.md)) with a
`{0}` cost override and lets the player cast it with an ordinary
`cast_spell`.

The reason is the announce path, not laziness: a cast needs targets,
modes and X collected from the player, and there is no frame for a
half-validated cast inside a resolution.
`CastSpellParams.Face` documents why `PendingChoice` is deliberately
not that frame and should not grow into it.

The trade is weaker than printed in the case that matters — a
cascaded sorcery on an opponent's turn cannot be cast at all — and
the one way it could have been STRONGER is closed: an accepted but
uncast hit goes to the bottom of its owner's library at the next end
step, via a delayed trigger. Without that, "yes" would park the card
in exile permanently available, which is a real upgrade on a keyword
that gives you exactly one window.

The random bottoming draws from `g.rng`, the game's own persisted
`*rand.PCG`, so a replayed game bottoms the pile the way the live one
did. `math/rand` here would desynchronise every snapshot after a
cascade.

### 10. `PendingChoiceMayCast` rather than a `{0}` pay-unless

"You may cast it without paying its mana cost" is a yes/no asked
during a resolution, which `PendingChoicePayUnless` can already
express — pass a cost of `{0}` and the payment always succeeds.

It is still the wrong prompt, because the prompt is copy as much as
it is control flow: pay-unless's entire client vocabulary is
"Pay {2}" / "Don't pay", and a dialog reading "Pay {0}?" about a free
spell is how a player answers the wrong question. The new kind shares
the `{apply: bool}` payload with the other four yes/no kinds and
carries the offered card (`MayCastCard`) so the prompt can show it
rather than name it in a sentence.

## What S28 did not ship

Named in the sprint issue, deliberately left out, each for a
structural reason rather than for time:

- **Urza's Incubator** — "as this enters, choose a creature type"
  needs a creature-type choice, which is a new `PendingChoice` kind
  and a new client picker for one card.
- **Will Kenrith** — its −2 is a cost reduction scoped to a PLAYER
  and to a duration ("until your next turn") rather than to a
  battlefield source. See the "not covered" note above.
- **Etali, Primal Conqueror** — a transform DFC whose ETB casts an
  unbounded number of exiled spells for free; the free-cast half is
  the grant this sprint built, but the "any number of spells" loop
  and the transform side are a different sprint.
- **Cabal Therapy** — its flashback is an alternative cost paid from
  the GRAVEYARD, and `CastSpellParams.FromZone` still has no
  graveyard case (S29's alt-cast-paths work). The front half also
  needs "choose a nonland card name", which has no prompt.
- **Phyrexian mana** — `ParsedCost` has carried `Phyrexian` and
  `HasPhyrexian` since S15 (mana_cost.go parses `{W/P}`), but the
  solver always pays the coloured half. Paying 2 life instead is a
  per-symbol choice at announce, which means a new announce-time
  payload and a per-symbol picker; it is the one item here that is
  purely additive to this ADR's machinery rather than blocked by
  something else.

---

## Addendum (2026-09-17): a spell's own cost modifier, affinity, and targets in `CostQuery` (#746)

**Status:** Accepted · 2026-09-17 · tracked on
[#746](https://github.com/krakenhavoc/cmd_and_ctrl/issues/746). This
status covers this section only. §1–§10 are unchanged and stay accepted,
except that §16 below widens §2's increase pass to coloured mana. The
section narrows the "Modifiers from outside the battlefield" line under
**Not covered** to the gaps listed in *Out of scope* below.
**Decided by the lead on the owner's standing guidance (2026-09-17):**

1. **The preview catches up in #696.** The engine and enumerator PRs ship
   now. Self modifiers and a `face` parameter join #696's list. §15 adds
   the cast's targets to that list, which follows from item 2.
2. **The X picker prices at one target** and shows the printed surcharge
   clause as a note (§15).
3. **Coloured cost increases are in scope.** The increase pass adds
   strive's coloured symbols (§16), and Call the Coppercoats ships with
   Fireball (§17).

The options considered are kept in [Decided questions](#decided-questions)
at the end of the section.

### Context

§1 built cost modifiers as statics of **battlefield** permanents. A spell
that changes **its own** cost has nowhere to declare it:

- "This spell costs {1} less to cast for each creature on the
  battlefield" (Blasphemous Act, Vanquish the Horde)
- affinity for artifacts, CR 702.41a: "This spell costs {1} less to cast
  for each [text] you control" (Thought Monitor, Myr Enforcer)
- "This spell costs {X} less to cast, where X is the total power of
  creatures you control" (Ghalta, Primal Hunger)
- "This spell costs {1} more to cast for each target beyond the first"
  (Fireball), and strive's "{1}{W} more … for each target beyond the
  first" (Call the Coppercoats)
- "This spell costs {7} less to cast if it targets a spell or ability
  that targets a creature you control with power 7 or greater" (Not of
  This World)

Checked on develop at `bcac391`:

- **Battlefield only.** `activeCostModifiersLocked`
  (`server/internal/game/cost_modifier.go:230`) walks only
  `g.Battlefield.Cards` (`:235-240`). Its comment calls widening it "a
  one-line change" (`:223-227`), and **Not covered** above says the same
  (`:189-190`). That widening would be wrong. A Sphere of Resistance in a
  hand would start taxing spells. A card's own modifier needs its own
  slot next to `CardDef.CostModifiers` (`game/carddef.go:59`, read at
  `:180`), per AGENTS.md §7, "Adding a `Spec` slot".
- **One pricing function.** `applyCostModifiersLocked`
  (`cost_modifier.go:258`) serves all three pricing surfaces (§6):
  - the cast (`mutations.go:1319`, inside `effectiveCostLocked` at `:1290`);
  - the enumerator (`legal/cast.go:151`);
  - the auto-tap preview (`lobby/http.go:1017`).
- **No targets.** `CostQuery` (`cost_modifier.go:113-152`) has Card,
  Controller, Source, FromZone, XValue and Cost. Targets are announced
  before the total cost is determined (CR 601.2c, then 601.2f). The cast
  path validates them first (`mutations.go:702`, before pricing at `:877`
  and `:889`), and `CastSpellParams.Targets` exists
  (`mutations.go:272`). No modifier can see them.
- **The enumerator prices once, before targets.** It prices a cast before
  it expands modes and targets (`legal/cast.go:140-165`), and only for
  hand and command-zone casts (`:147-150`).
- **The preview prices the printed cost.** It prices without targets
  and without a face. It keeps its own copy of the printed-cost,
  commander-tax and zone logic (`lobby/http.go:994-1010`, #696), but its
  modifier pass is the shared `ApplyCostModifiers` (`:1017`), called with
  `Card` set.
- **Increases are generic only.** §2's increase pass is
  `cost.Generic += n` (`cost_modifier.go:275`). Nothing can add a
  coloured symbol.
- **A shipped caveat.** Blasphemous Act says "The cost reduction is
  missing" (`cards/effects/blasphemous_act.go:29`).

### Decisions

#### 11. `Spec.SelfCostModifiers`: a separate slot, read only for the spell being cast

`Spec.SelfCostModifiers []game.CostModifier` becomes
`CardDef.SelfCostModifiers`, with one line in `effects.buildDef` and one
reader, `game.SelfCostModifiersFor(card Card)`. That reader looks the card
up by `CatalogKey(card)`. It is the four-edit slot recipe in AGENTS.md §7.
No per-slot `Catalog…` variable is added until a game-package test needs
to stub it.

A separate slot, not a flag on `CostModifier`, because the question "does
this modifier apply to other spells?" is answered by where the card
declares it. A modifier in `CostModifiers` is read from the battlefield
and never from a hand. A modifier in `SelfCostModifiers` is read only
while its own card is being priced, and never from the battlefield. A
card can have both. Mycosynth Golem has affinity for artifacts itself and
also grants affinity to artifact creature spells from the battlefield.

Card files reuse the existing constructors (`CostsLess`, `CostsLessEach`,
`CostsMore`, `CostsMoreEach`). The slot, not the constructor, is what
makes a modifier apply to its own spell. Two new constructors:

- `AffinityFor(label string, match CardPredicate)` is `CostsLessEach`
  counting the caster's permanents that match (CR 702.41a). Each
  `AffinityFor` entry is its own modifier, so a card with two instances
  applies both (CR 702.41b). No card needs two yet, and the rule costs
  nothing extra.
- `CostsMorePerTargetBeyondFirst(per string, label string)` is Fireball
  (`"{1}"`) and strive (`"{1}{W}"`). It reads targets (§13), and it adds
  `per` for each target beyond the first, and nothing for zero or one.
  `per` may hold coloured symbols (§16). Its amount never goes negative,
  so an unpriced zero-target query cannot trip §4's refusal.

`effects.Register` panics when `SelfCostModifiers` contains a
`CostFloor`. No printed card sets a floor on its own cost, and an
untested kind should not be declarable.

#### 12. Self modifiers join the same three passes, bound to the card as it is cast

`activeCostModifiersLocked` takes the query and appends the cast card's
self modifiers after the battlefield ones. §2's increase, reduction and
floor passes, §3's generic-only floor and §4's refusal of a negative
amount apply unchanged, with the increase pass widened by §16. Two
consequences:

- **The order stays CR 601.2f's.** Every increase, from the battlefield
  and from the spell, is added before any reduction. Ghalta, Primal Hunger
  under a Sphere of Resistance goes from {10}{G}{G} to {11}{G}{G} before
  its power reduction runs. The reduction then spends generic mana only.
  With enough power it reaches {G}{G} and stops there, because
  `reduceGeneric` never touches a coloured requirement.
- **Self modifiers come after battlefield modifiers within a pass.** That
  is for determinism. It changes no result today: reductions clamp and
  add up, so their order does not matter unless an `Amount` reads
  `q.Cost`, and no catalog modifier does.

The bound source is **the card being cast, with `Controller` set to the
caster**. CR 601.2a makes the caster the spell's controller, and a card in
a hand carries no controller of its own. Setting it keeps a `YourSpell()`
predicate true, where it would otherwise be quietly false.

**Zones.** CR 113.6d: an ability that "modifies what that particular
object costs to cast functions on the stack". So a self modifier applies
whatever zone the spell is cast from: hand, command zone, graveyard or
exile. The engine prices before it moves the card, so at pricing time
the card is still in its source zone. Counting helpers therefore
**exclude `q.Card` itself**. Under CR 601.2a the card is already on the
stack when the total cost is determined. "For each card in your
graveyard", read during a flashback cast, must not count the card being
cast.

**Alternative costs.** CR 118.9d applies increases and reductions to an
alternative cost. `printedCostLocked` has already swapped in the
alternative cost, the exile-grant override and the commander tax before
modifiers run (§5). A Fireball cast without paying its mana cost still
pays its per-target surcharge, and a Blasphemous Act cast with an
alternative cost still gets its reduction.

**Faces.** `CatalogKey(q.Card)` reads the face the cast stamped
(`mutations.go:530`), and the enumerator stamps the same face per
castable face (`legal/cast.go`, `CastableFaces` and `SetFace`). An MDFC
or split card uses the self modifiers of the face being cast and no
others.

#### 13. `CostQuery.Targets`, shown only to modifiers that declare they read it

`CostQuery.Targets []TargetRef` is filled from `params.Targets` in
`effectiveCostLocked`. A new field, `CostModifier.ReadsTargets bool`,
controls who sees it: `amountFor` passes the targets to a modifier that
sets the flag, and nil to every other modifier.

The flag does two jobs:

- **It tells the enumerator whether it can price before expanding
  targets** (§14).
- **It keeps the engine and the enumerator in agreement even when a card
  gets the flag wrong.** A modifier that reads `q.Targets` without the
  flag sees nil in both places. The card's own test catches it. The other
  failure mode, where the engine sees targets and the enumerator does
  not, would offer bots casts the engine refuses (#544).

Battlefield modifiers get the field too. "Spells your opponents cast that
target a Merfolk you control cost {2} more" (Kopala, Warden of Waves)
needs no further change. Target-reading constructors set the flag
themselves: `CostsMorePerTargetBeyondFirst`, and a `TargetsMatch(pred)`
cost predicate paired with a `CostsLessIf…` constructor for Not of This
World.

#### 14. The enumerator prices per target set only when something reads targets

A new read, `g.CastPriceReadsTargetsForEffect(card)`, returns true when
the card's self modifiers, or any battlefield modifier, set
`ReadsTargets`. It ignores `AppliesTo`, so it errs toward true.

- **False** (every card and board today): nothing changes. The cast is
  priced once, before modes and targets, and the early affordability
  return stays.
- **True:** there is no early return. After modes and targets are
  expanded (`legal/cast.go:209-225`), each (mode, target set) is priced
  with its targets, and `affordableX` runs on that price. An unaffordable
  set is skipped **without spending expansion budget**, so a Fireball
  that is affordable at one target is not crowded out by the three-target
  sets ahead of it. Each move carries its own X.

A nil-targets price cannot be used as a gate, because target-reading
reductions exist. Not of This World costs {7} at nil targets and {0} with
a qualifying target. A nil-targets check would hide it from a player
who has less than seven mana, even when the cast is free.

Only hand and command-zone casts are enumerated (#673). A graveyard or
exile cast of a self-modified spell is not offered to bots, like every
other such cast.

#### 15. The preview catches up in #696, and the X picker prices at one target

The auto-tap preview has no targets and no face on its query string. It
copies the printed-cost, commander-tax and zone logic rather than calling
`effectiveCostLocked` (`lobby/http.go:994-1010`, #696). The technical rule
is that the preview prices through the same function as the cast, with
whatever targets and face it is given, and with nil targets when it has
none.

**Sequencing (lead decision, option (c)).** The engine and enumerator PRs
ship without waiting for #696, and the preview's copy is not extended.
The preview's modifier pass is already the shared `ApplyCostModifiers`
with `Card` set (`http.go:1017`). Once §12 binds self modifiers inside
`activeCostModifiersLocked`, the preview sees them with no change, for a
single-faced card cast from the hand or the command zone. That covers
every card in the engine PR. What stays wrong goes on #696's list, and
is fixed there:

- **a `face` parameter**, so an MDFC or split card's preview uses the
  self modifiers of the face being cast, not the default face's;
- **the cast's targets**, so a multi-target Fireball or strive cast, or a
  target-reading reduction such as Not of This World, is previewed at
  its real price;
- the graveyard and exile zones and the alternative cost, which #696
  already lists, now with self modifiers applied to them (CR 113.6d,
  CR 118.9d).

**The X picker (lead decision, option (a)).** `XCostModal.svelte` opens
before targeting (ADR 0021 §3), so it cannot know the target count. It
prices at one target and says so:

- The preview it calls has no targets, so it prices at nil targets. For a
  per-target surcharge that is exactly the one-target price, because
  `CostsMorePerTargetBeyondFirst` adds nothing for zero or one target.
  No change to the preview's pricing is needed.
- The preview response gains `cost_notes []string`: the `Label` of each
  of the card's own self modifiers that sets `ReadsTargets`, in
  declaration order, from the cast branch only. It is empty for every
  other card. The label is the printed clause ("This spell costs {1}
  more to cast for each target beyond the first"). The response is
  already per caller and per card, so no per-viewer stripping is needed.
  Like the pricing, it reads the default face until #696 adds `face`.
  Every target-priced card in §17 is single-faced.
- The modal shows each note under the affordable/missing readout. The
  readout's price is the one-target price, and the note tells the player
  that more targets cost more.

A `CastSpell` with more targets than the player can pay for is refused by
the engine, as today.

#### 16. Coloured cost increases (strive)

Lead decision, option (a). Strive (CR 207.2c ability word) reads "This
spell costs {1}{W} more to cast for each target beyond the first" (Call
the Coppercoats), and every card in the cycle adds a coloured symbol per
extra target. §2's increase pass adds only generic mana.

- **`CostModifier.Unit ParsedCost`.** When it is set on a `CostIncrease`,
  `Amount` counts units, and the increase pass adds `Amount` copies of
  `Unit`: the generic part to `cost.Generic`, and one fresh
  `ColorRequirement` per coloured symbol per unit to `cost.Required`. A
  zero `Unit` (no generic, no coloured symbols) is today's behaviour:
  `Amount` is generic mana. Every modifier in the catalog keeps working
  unchanged.
- **The constructor parses once.** `CostsMorePerTargetBeyondFirst` parses
  `per` with `game.ParseCost` and panics if it fails, so a typo fails at
  init and not at a cast. Fireball's `"{1}"` goes through the same field.
- **`effects.Register` panics** on a `Unit` that has `{X}`, a Phyrexian
  or hybrid symbol, or that sits on a `CostReduction` or `CostFloor`. No
  printed increase needs any of these, and §3 keeps reductions generic
  only.
- **Order and floor are unchanged.** Coloured increases are added in the
  increase pass, before any reduction (CR 601.2f). `reduceGeneric` never
  touches a coloured requirement, so a Goblin Electromancer on a
  two-target Call the Coppercoats removes generic mana only, and both
  {W} symbols stay. `raiseToMinimum` (`cost_modifier.go:391`) already
  counts coloured requirements toward mana value, so a Trinisphere sees
  the added symbols.
- **Nothing else changes.** The solver, the auto-tapper and convoke
  already pay a `Required` list of any length. An alternative cost gets
  the coloured surcharge too (CR 118.9d, §12).

#### 17. Cards

- **Blasphemous Act** drops its caveat.
- **Engine PR:** Thought Monitor (`AffinityFor` artifacts), Myr Enforcer,
  Ghalta, Primal Hunger, Vanquish the Horde.
- **Enumerator PR (§14):** Fireball and Call the Coppercoats. A
  target-priced card must not ship before the enumerator can price it.
  The rest of the strive cycle becomes catalog work.
- Not of This World, Ancient Stone Idol, Voyage Home, Sapling Nursery,
  Mycosynth Golem and Thrumming Hivepool become catalog work. Earthquake
  Dragon and Metalwork Colossus also wait on #655.

### Out of scope

- **Coloured reductions**, such as "costs {W}{B} less" (CR 118.7c): §3
  still holds. Coloured *increases* are in (§16).
- **A note for a battlefield modifier that reads targets** (Kopala,
  Warden of Waves) in the X picker. §15's `cost_notes` lists the card's
  own self modifiers only.
- **Cost modifiers on activated abilities**: the separate "Cost
  modification for activated abilities" row in `docs/engine-seams.md`.
- **Per-player, duration-scoped modifiers** (Will Kenrith): still
  **Not covered**.
- **Showing a modified price on a card in hand.** Nothing in the client
  shows a modified price today, including under a Sphere.

### Consequences

- About a dozen cards become catalog work. Blasphemous Act becomes whole.
- Each cast adds one more `CardDef` read, for its own self modifiers.
  The early return for a cast with no modifiers stays.
- The enumerator does more work only on boards where a target-reading
  modifier exists, and at most one price per target set it already
  expands.
- `CostQuery` gains a field that most modifiers never see. That is on
  purpose (§13).
- `CostModifier` gains `Unit`, which every existing modifier leaves zero
  (§16). Strive's cycle becomes catalog work.
- **Until #696 lands, the auto-tap preview is wrong for some casts of a
  self-modified spell:** a back face that has its own self modifiers, a
  multi-target Fireball or strive cast, and a graveyard or exile cast.
  The preview only matters after a strict-mana refusal. The cast, the
  bots and the census are right from the engine PR on.
- The X picker shows a one-target price and a note for Fireball-style
  cards. It never shows an exact multi-target price before targeting.

### Alternatives considered

- **Widening `activeCostModifiersLocked` to every zone.** Sphere of
  Resistance in a hand would tax the table. Filtering on
  "source == q.Card" inside the walk makes every cast scan every hand and
  graveyard. Rejected.
- **A `SelfOnly bool` on `CostModifier` in the existing slot.** Every
  battlefield walk would have to remember to skip self-only modifiers. A
  slot makes the wrong read impossible. Rejected.
- **Always pricing per target set in the enumerator.** This drops the
  early return for every unaffordable targeted spell in every hand, on
  every snapshot, to serve a handful of cards. Rejected in favour of §14's
  flag.
- **Showing all modifiers the targets.** A modifier that reads targets
  without saying so would split the engine from the enumerator, and
  nothing would catch it. Rejected (§13).

### Implementation plan

1. **Engine PR.**
   - the `SelfCostModifiers` slot (the four §7 edits);
   - binding inside `activeCostModifiersLocked`;
   - `CostQuery.Targets`, `CostModifier.ReadsTargets` and the hiding in
     `amountFor`;
   - `effectiveCostLocked` passes the targets;
   - `CostModifier.Unit` and the coloured increase pass (§16);
   - the registration guards, `AffinityFor` and `CostsMorePerTargetBeyondFirst`;
   - the §17 engine-PR cards, and the census regenerated.

   No card that reads targets is registered in this PR. The target and
   coloured-increase paths are tested with fixture cards.
2. **Enumerator PR.**
   - `CastPriceReadsTargetsForEffect` and pricing per target set;
   - `cost_notes` on the auto-tap preview response, documented in
     `docs/protocol.md`;
   - Fireball and Call the Coppercoats, and the census regenerated;
   - `docs/engine-seams.md`'s "A spell's own cost modifier" row moved to
     Closed.
3. **Client PR.** `cost_notes` on `AutoTapPreview` in `api.ts`, the note
   under `XCostModal`'s readout, and vitest cases.
4. **Preview, in #696** (lead decision): the `face` parameter and the
   cast's targets join #696's list, and the preview then prices self
   modifiers through the shared cast-price entry point #696 builds.

### Test plan

- **Floor.** Ghalta with more total power than its generic cost costs
  exactly {G}{G}, not less, and never touches a colour.
- **Ordering.** Ghalta under Sphere of Resistance, and a noncreature
  self-reducer under Thalia: every increase lands before any reduction
  (the §2 Lightning Bolt case, with the reducer on the spell).
- **A Sphere in hand taxes nothing.** A Sphere of Resistance in the
  caster's hand, and one in an opponent's, leaves a Lightning Bolt at {R}.
- **A self modifier on the battlefield does nothing.** Thought Monitor on
  the battlefield does not reduce another artifact spell.
- **Zones.**
  - A command-zone cast pays the tax, then gets the reduction.
  - A flashback-style graveyard cast of a "for each card in your
    graveyard" reducer does not count itself.
  - An alternative-cost cast is reduced (CR 118.9d).
- **Faces.** An MDFC whose front has a self reduction and whose back has
  none: the back costs its printed cost, and the front is reduced.
- **Targets.**
  - Fireball at one, two and three targets costs {X}{R}, {X}{1}{R} and
    {X}{2}{R}.
  - A modifier without `ReadsTargets` sees nil targets in both the engine
    and the enumerator.
- **Coloured increases (§16).**
  - Call the Coppercoats at one, two and three targets costs {2}{W},
    {3}{W}{W} and {4}{W}{W}{W}.
  - Under Goblin Electromancer, a two-target cast costs {2}{W}{W}: the
    reduction spends generic mana only, and both {W} stay.
  - Under Sphere of Resistance and Goblin Electromancer, a two-target
    cast costs {3}{W}{W}: both increases land before the reduction.
  - A pool with enough generic mana but one {W} short refuses the
    two-target cast, and the auto-tapper taps a white source for the
    added {W}.
  - `Register` panics on a `Unit` with `{X}`, a hybrid or a Phyrexian
    symbol, and on a `Unit` on a reduction or a floor.
  - Every catalog modifier registers with a zero `Unit`.
- **Refusal.** A self modifier returning a negative number refuses the
  cast (§4), and names the card.
- **Enumerator parity.**
  - For each card above, every cast move the enumerator offers is
    accepted by `CastSpell` with `auto_tap`.
  - A board with no target-reading modifier offers the same moves as
    before.
  - Fireball offers the one-target set when only that one is affordable,
    even behind unaffordable larger sets.
- **Preview.**
  - Blasphemous Act in hand, with creatures on the battlefield, is
    previewed at its reduced price, through the shared
    `ApplyCostModifiers`.
  - Fireball's preview at X = 3 prices {3}{R} and carries its
    `cost_notes` clause. Lightning Bolt's preview carries no
    `cost_notes`.
- **Client.** `XCostModal` renders each `cost_notes` entry under the
  readout, and renders nothing extra when the list is absent.
- **Catalog soak** with the new cards.

### Decided questions

Answered on 2026-09-17 by the lead, on the owner's standing guidance. The
chosen option is marked **(chosen)**; the recommendation text is kept for
the record.

1. **How does the auto-tap preview catch up?** The preview's copy of the
   price logic will not see self modifiers. When a strict-mana cast of
   Blasphemous Act is refused and the player clicks "Auto-tap & cast", the
   modal can show the printed {8}{R} as missing and keep the button
   disabled, even though the cast itself would succeed.
   - **(a)** The self-modifier work waits for #696's shared pricing entry
     point, so the preview is fixed once.
   - **(b)** Ship the engine and enumerator PRs now. Extend the preview's
     copy with self modifiers as a stopgap, and delete it when #696 lands.
   - **(c) (chosen)** Ship the engine and enumerator PRs now. Leave the
     preview wrong for self-modified spells, add them to #696's list
     (with a `face` parameter), and fix it there.

   **Recommendation: (c).** The cast, the bots and the census are right
   from day one. The preview only matters after a strict-mana refusal. And
   (b) grows the duplicate code #696 exists to remove.

   **Checked when recording the decision:** the question overstates the
   gap. The preview's modifier pass is already the shared
   `ApplyCostModifiers` (`lobby/http.go:1017`), so a single-faced card
   such as Blasphemous Act cast from hand is previewed at its reduced
   price once the engine PR lands. What (c) leaves for #696 is the face,
   the targets, and the non-hand zones and alternative costs #696 already
   lists. Applied in §15, *Consequences* and the implementation plan.
2. **What does the X picker show for a spell whose price depends on
   targets?** Fireball's X picker opens before targeting (ADR 0021 §3), so
   the per-target surcharge is not known yet.
   - **(a) (chosen)** Price at one target, and show the modifier's
     printed clause under the readout ("costs {1} more for each target
     beyond the first").
   - **(b)** For such cards, collect targets before X, so the readout is
     exact.
   - **(c)** Price at one target with no note.

   **Recommendation: (a).** It is honest and cheap, and it keeps one cast
   flow. (b) reorders the client flow for a handful of cards.

   Applied in §15 (`cost_notes` on the preview response), the
   implementation plan and the test plan.
3. **Coloured cost increases (strive): in this work?** §2's
   `CostIncrease` adds generic mana only. Strive's "{1}{W} more for each
   target beyond the first" adds a coloured requirement. Call the
   Coppercoats is on #746's first list, and strive is an ability word
   (CR 207.2c) across a cycle of cards.
   - **(a) (chosen)** In. `CostsMorePerTargetBeyondFirst` takes a mana
     string, and the increase pass adds its coloured symbols as
     `Required` entries. Reductions stay generic-only (§3).
   - **(b)** Out. Fireball ships, strive gets its own seam row, and Call
     the Coppercoats waits.

   **Recommendation: (a).** It is one field on the increase pass. The
   solver already pays coloured requirements, and a Sphere or Thalia test
   pins the order.

   Applied in §16, §17, *Consequences*, the implementation plan and the
   test plan.
