# ADR 0042 — Cost modification (CR 601.2f)

**Status:** accepted (S28 sub-PR 1)
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
