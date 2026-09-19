# ADR 0073 — Optional additional costs and the announce-time cast gate

**Status:** accepted
**Date:** 2026-09-18
**Sprint:** S42 — Casting from non-hand zones: granted permissions and alternative costs
**Issues:** [#664](https://github.com/krakenhavoc/cmd_and_ctrl/issues/664) (optional additional costs — kicker and buyback, including non-mana), [#760](https://github.com/krakenhavoc/cmd_and_ctrl/issues/760) (one announce-time cast gate). Tracker [#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885).

**Numbering:** every remote branch was swept with the AGENTS.md §4 loop on
2026-09-18 — `git ls-remote --heads origin`, then `git ls-tree -r
--name-only <ref> docs/decisions/` over every fetched head. `0066`–`0072`
were taken (granted cast permissions, layer dependency ordering, the mana
spent on a spell, face-down objects, untap-step choices, designations,
protection); `0073` was the first free number, and the sweep was repeated
immediately before the push. `0005`, `0024`, `0029` and `0030` stay
permanently unused.

## Why these two are one ADR

They are the two halves of **CR 601.2b**, the announce step, and they are the
last two questions the cast path cannot ask.

- #664 is "may I pay *more* than the cost?" — a choice the caster makes while
  proposing the spell, which changes what the spell *does* rather than only
  what it costs.
- #760 is "may I cast this *at all*?" — a refusal the board or the card
  itself imposes at the same moment.

Both are read at announce, both must be visible to the bot enumerator and to
the client before the cast is dispatched, and both are facts that have to
survive onto the stack item. Shipping them apart would mean touching
`CastSpell`'s announce block, `legal.castMovesForCard`, `stampLegalTargets`
and the client's cast chain twice in the same sprint, with two chances to
disagree about the order the questions are asked in.

## Context

### Optional additional costs (#664)

`game.AdditionalCost` (`server/internal/game/additional_cost.go:27-65`) has
three components — `DiscardCards`, `Sacrifice *TargetSpec`, `PayLifeX` — and
every one of them is mandatory. [ADR 0021](0021-additional-costs.md) left the
optional ones out twice, on purpose: at `:118-121` and again at `:167-169`,
"they need a *choice* of whether to pay, which changes the spell's effect,
not just its price — that's a mode-plus-cost shape and wants its own
design."

Nothing in the tree says `kicker` or `buyback`. The consequences:

- **Constant Mists** ("Buyback—Sacrifice a land. Prevent all combat damage
  that would be dealt this turn.") is a {1}{G} Fog without it. Its buyback is
  **non-mana**, so it needs the choice on a `Sacrifice` component, not a new
  mana-only mechanism.
- **Kicker** (CR 702.33) reaches far more cards than buyback, and its payload
  is read in three different places: by the resolving spell (Burst
  Lightning), by an entering permanent's own trigger (Gatekeeper of Malakir's
  "when this enters, **if it was kicked**"), and by a count (Wolfbriar
  Elemental's multikicker, CR 702.33d).
- `docs/engine-seams.md` records 15 skipped cards against the row and the
  2026-09-16 audit attributes 32 / 47.

The three existing cost seams the design has to compose with:

| seam | file | what it settles |
|---|---|---|
| mandatory additional cost | `additional_cost.go` | paid **alongside** the mana cost, CR 601.2f |
| alternative cost | `alternative_cost.go` | paid **instead of** the mana cost, CR 118.9 |
| cost modifier | `cost_modifier.go` | increases and reductions applied **after** both, ADR 0048 |

### The cast gate (#760)

`Game.CastSpell` is the only entry point for a cast, and none of its gates
asks "is this player forbidden from casting this spell?". No `CantCast`,
`CastRestriction`, `CastCondition` or `ErrCantCast` exists. `restrictions.go`
carries per-permanent bits (`CantAttack`, `CantActivate`, …) and nothing that
speaks about a player.

Three separate kinds of "no" are waiting on one gate:

1. **Statics on permanents.** Rule of Law ("each player can't cast more than
   one spell each turn"), Grafdigger's Cage ("players can't cast spells from
   graveyards or libraries"), Rakdos, Lord of Riots ("you can't cast creature
   spells unless an opponent lost life this turn"). Each is a static ability
   of a battlefield permanent, read at cast time with the caster and the face
   already chosen — which is `CatalogCostModifiers` / `CostQuery`
   (`cost_modifier.go:197` / `:113`) exactly, one field different.
2. **A spell's own condition.** CR 307.6's legendary sorcery ("you may cast
   this spell only if you control a legendary creature or planeswalker",
   Urza's Ruinous Blast), and the "cast only if" family generally.
3. **Bans with a duration** created by a resolving spell (Silence's "this
   turn", Reflector Mage's "until your next turn").

The two consumers that must agree with `CastSpell` or the client lies: the
bot enumerator `legal.castMoves` (`legal/cast.go:35`, which already keeps a
twin of the split-second check at `:37` for exactly this reason) and the
view's `CastableHere` stamp (`protocol/view.go:849-865`).

## Decisions

### 1. `Optional` is a flag on the existing component, not a new kind of cost

`game.AdditionalCost` gains four fields and keeps all three of its existing
components:

```go
Optional bool     // CR 601.2b: the caster chooses whether to pay this
Key      string   // "kicker", "multikicker", "buyback" — the wire and record identity
ManaCost string   // the mana half of the clause: "{4}", "{1}{W}", ""
Repeat   int      // CR 702.33d multikicker: the maximum number of times it may be paid
```

`DiscardCards`, `Sacrifice` and `PayLifeX` are unchanged and are what makes
Constant Mists' "Buyback—Sacrifice a land" a one-line card: it is the
`Sacrifice` component the engine has had since S21, with `Optional: true` on
it. A mandatory cost and an optional one are validated by the same function
and paid by the same function; the flag is the only difference, which is the
whole point of putting it there rather than inventing an `OptionalCost` type
that would have to grow its own copy of every component.

`ManaCost` is genuinely new — every mandatory additional cost in the catalog
is non-mana, and kicker is mana — and it belongs on the component because CR
601.2f adds additional costs to the total *cost*, not to a side channel.

**A card declares its optional costs in `Spec.OptionalCosts []game.AdditionalCost`,
a second slot beside the existing `Spec.AdditionalCost`.** One slot would
have been tidier and was rejected: `AdditionalCostFor` is read by the engine,
the enumerator, the view and ~20 card files as "the one mandatory cost", and
an optional cost must be *indexable* (the announcement names which ones were
paid). `Register` cross-checks the two slots so they cannot disagree — an
entry in `OptionalCosts` without `Optional: true` panics at boot, and so does
an `Optional: true` in the mandatory slot.

### 2. The choice is announced at CR 601.2b, as a list of indices

`CastSpellParams.OptionalCosts []int` — indices into the card's
`OptionalCosts` list. **Paying one N times is naming its index N times**, so
multikicker is `[0, 0, 0]` and needs no second field. `PaidCost.OptionalCosts`
records the same list, normalised to ascending order, and `KickedTimes()` is
`len(filter(indices))`.

Validated in `CastSpell` immediately after the mode check and **before**
targets, which is CR 601.2b before CR 601.2c and matters for its own sake: a
kicked spell may legally choose different targets from an unkicked one, so
the targeting gate has to be judged under the cost actually being announced.

Rejected: a map of key → count. A repeated index is one slice, orders
deterministically, clones by `append`, and serialises as `[]int` with no
custom marshaller — and a map would have let the same optional cost be named
under two spellings.

### 3. The mana half joins the cost where CR 601.2f puts it

`printedCostLocked` adds the chosen optional costs' `ManaCost` after the
alternative-cost swap and the commander tax, and **before** the cost
modifiers — so Thalia taxes a kicked spell once, and Trinisphere reads the
kicked total. One helper, `optionalCostManaLocked`, and the bot enumerator
calls it on the same inputs, so the price a bot is offered and the price the
engine charges cannot drift.

### 4. The non-mana half is paid at CR 601.2h with the mandatory cost

The caster's `discard_ids` and `sacrifice_ids` stay flat wire lists. The
engine builds **one ordered payment plan** — the mandatory cost first, then
each chosen optional cost in index order, once per payment — and walks the
flat lists against it. `validateAdditionalCostLocked` and
`payAdditionalCostLocked` take the plan instead of a single cost, so there is
still exactly one validator and one payer, and the ordering discipline
(validate everything, then pay, with the spell already on the stack) is
untouched. A Blood Artist still sees Gatekeeper of Malakir's kicker sacrifice
while the Gatekeeper is on the stack.

**A repeatable optional cost may only carry mana.** `Register` panics on a
`Repeat > 1` entry with a `Sacrifice`, `DiscardCards` or `PayLifeX`
component. This is not a simplification: every printed multikicker is a mana
cost, and the alternative — N independent card-shaped payments per cast —
would need a wire shape nothing asks for. `PayLifeX` is refused on an
optional cost outright, because it rides the shared `XValue` slot and two
claimants on one number is a bug waiting to be written.

### 5. "Was it kicked" is a paid-cost fact, read in exactly one place per consumer

- **At resolution**, `ctx.WasKicked()` / `ctx.KickedTimes()` read
  `item.Paid.OptionalCosts`. Both are sugar over the generic
  `ctx.OptionalCostTimes(key)`; a card that prints a differently-named
  optional cost uses the generic form and needs no new accessor.
- **At entry**, a permanent's own trigger cannot reach the stack item — the
  item is out of `StackMeta` before the ETB event is emitted, and a
  `TriggeredAbility.AppliesTo` receives only the game, the source and the
  event. So the resolution path **carries the record onto the permanent**:
  `Card.PaidOptionalCosts []int`, stamped in `resolveTopOfStackLocked` right
  after `MoveCard` and before the `EventZoneMove` / `EventETB` pair, which is
  the one moment that holds both the landed permanent and the item.
  `game.CardKickedTimes(card)` is the read, and Gatekeeper of Malakir and
  Wolfbriar Elemental are the two consumers.

  The field is per-instance and is cleared with the rest of them in
  `entry_tail.go`'s CR 400.7 reset, so a Gatekeeper that dies and is
  reanimated is not kicked.

This is the narrow, honest slice of #653's cast-provenance seam: the *paid
optional costs* travel, and `AltCost` still does not. Say so on #653 rather
than pretending the seam closed.

### 6. Buyback replaces the resolution destination, and only that one

CR 702.27b: "if the buyback cost was paid, put this card into its owner's
hand as it resolves instead of putting it into that player's graveyard."

`routeStackCardToGraveyardLocked` already takes the stack item because
flashback replaces the destination there (CR 702.34a). It gains a third
argument, `resolved bool`, and buyback is two lines inside it:

```go
if resolved && item.BuybackPaid() { r.Dst, r.DstOwner = ZoneHand, c.Owner }
```

`resolved` is true only at the resolution call site. The fizzle call site
("countered by game rules", CR 608.2b) and the defensive no-meta path pass
false, which is the rule and not a convenience: buyback says *as it
resolves*, so a bought-back Capsize whose target left in response still goes
to the graveyard, and a countered one does too (a counterspell never reaches
this function at all).

**Flashback wins over buyback** when a card somehow has both, because CR
702.34a replaces *every* exit and buyback replaces one of them. No printed
card has both; the precedence is written down so the branch order is a
decision rather than an accident.

The destination still goes through `routeCardToZoneLocked` — the #529 shared
stack-exit primitive — so CR 903.9's commander choice still applies to a
bought-back commander.

### 7. One gate function, two callers, one view stamp

```go
func (g *Game) CastGateLocked(caster uuid.UUID, card Card, zone ZoneKind, params CastSpellParams) error
```

`card` is the **face-materialised** copy (ADR 0034 stamps the face before any
announce gate runs), so the face rides the card rather than a fourth
parameter that could disagree with it.

It consults two sources, in this order:

1. **`CastRestriction`s contributed by battlefield permanents**, collected
   through `CatalogAbilityKey` so a permanent under a CR 613.1f
   ability-removing effect stops restricting — the same key
   `CostModifiersForCard` uses, for the same reason. The shape is
   `cost_modifier.go`'s with one field:

   ```go
   type CastRestriction struct {
       Label      string                 // "Rule of Law — each player can't cast more than one spell each turn"
       Forbids    func(q CastQuery) bool
       ActiveWhen Designation
   }
   ```

   `CastQuery` carries `Game`, `Card`, `Controller`, `Source`, `FromZone` and
   `OptionalCosts` — the same read-only-by-value discipline `CostQuery` has,
   for the same reason.

2. **The spell's own condition**, `Spec.CastCondition func(g *Game,
   controller uuid.UUID, card Card) bool`. The legendary-sorcery helper
   `LegendarySorcery()` is one call of it (CR 307.6); Urza's Ruinous Blast is
   the proof card.

The gate returns `*CantCastError`, which wraps the new `ErrCantCast` sentinel
and carries the printed clause that refused it, so the client's toast names
the card rather than saying "invalid parameter". `ws.classifyActionError`
maps it the way it maps `ErrUnparseableCost`.

**Three call sites, one function:**

- `CastSpell`, at the point ADR 0066 named
  (`cast_permission.go:591-594`'s "#760 hook"): after the face, the source
  zone, the permission and the announce-time choices are settled, and before
  any cost is paid. Every free cast ends here too — cascade, a granted
  permission, an impulse grant — so CR 101.2's "can't beats may" falls out of
  the placement rather than needing a rule of its own.
- `legal.castMovesForCard`, once, so a bot is never offered a banned cast.
- `protocol.stampLegalTargets`, so `CastableHere` and the hand's cast button
  are already grey.

**Split second stays out of the gate.** It was tempting to absorb the
enumerator's twin at `legal/cast.go:37`, and it is wrong: split second
restricts *taking an action* (it blocks activations too) and correctly does
**not** block a land play, which is a special action (CR 702.61b). Folding it
into a per-card *cast* gate would either change the land rule or make the
gate carry an exception that has nothing to do with casting. The twin stays;
the gate is added beside it.

### 8. Cast counts, durations and what is deliberately not here

- Rule of Law reads the existing `CastTally{Total, Noncreature}`
  (`game.go:795-798`). **No nonartifact count is added** — Ethersworn
  Canonist is not in this PR's proof set and a count with no reader is a
  field to keep correct for nothing. It is one line when its card arrives.
- **Bans with a duration are out.** Silence and Reflector Mage need the
  turn-scoped and permanent-duration registries (#755); the gate's first
  source list is statics and conditions only. The `CastRestriction` shape
  takes a duration-carrying third source without changing its signature, and
  that is the extension point.
- **"Can't cast" for a chosen card name is out.** Meddling Mage and Nevermore
  need a choose-a-card-name prompt, which does not exist (`NamedTribe` and
  `ChosenColor` do; a name does not). Grafdigger's Cage is the zone-shaped
  proof card instead.
- **Suspend** (#655, #659) is not wired here, but the gate is a method on
  `*Game` taking a card and a zone, so a special action can call it. Nothing
  about it is bound to `CastSpell`.

### 9. Bots and the client

**Bots.** `castMovesForCard` enumerates the unkicked cast and, when the seat
can afford it, the cast with each single optional cost paid — and, for a
repeatable one, up to `min(Repeat, 3)` payments. Bounded on purpose: the
expansion is already modes × targets × cost payments, and an unbounded
multikicker would multiply it by the seat's mana. The policy is written in
`docs/bot.md`: *a bot considers paying each optional cost at most once, and a
repeatable one up to three times.* A non-mana optional cost is enumerated
with the same payment search the mandatory one uses.

**Client.** The optional costs are offered as toggles (and a stepper, for a
repeatable one) **inside the alternative-cost picker**, not as a fourth modal
in the chain. They are the same question asked at the same moment — "what am
I paying for this?" — and CR 601.2b announces them together. A card with
optional costs and no alternative costs opens the same modal with only the
add-ons showing. The choice rides the existing `CastChoices` bundle
(`targeting.ts`), which is what ADR 0021's "known debt" note asked for.

## Consequences

- Constant Mists, Capsize, Burst Lightning, Gatekeeper of Malakir, Wolfbriar
  Elemental and Rite of Replication ship; so do Urza's Ruinous Blast, Rakdos,
  Lord of Riots, Grafdigger's Cage and Rule of Law.
- `docs/engine-seams.md`: "Optional additional costs (kicker, buyback)",
  "\"Can't cast\" restriction gate" and "Legendary-sorcery cast-condition
  hook" move to Closed.
- ADR 0021's two "kicker is out" paragraphs get a dated amendment pointing
  here.
- `StackItem.Paid` is classified `carried` in `snapshot_drift_test.go`
  already; `PaidCost.OptionalCosts` and `Card.PaidOptionalCosts` join it, and
  `clonePaidCost` deep-copies the slice.
- Escalate and entwine are the same family and are **not** covered:
  escalate's cost is per *extra mode*, which is a mode-and-cost product the
  `[]int` announcement cannot express. When they arrive they extend §2 rather
  than replacing it.

## Alternatives considered

**A separate `OptionalCost` type keyed like alternative costs.** Rejected:
Constant Mists' buyback is a `Sacrifice` clause, and a new type would have
had to grow `DiscardCards`, `Sacrifice` and a validator that were already
written. The flag reuses all three.

**Kicker as a mode.** [ADR 0065](0065-modal-and-multi-target-clauses.md) is
explicit that "if kicked" is not a mode — a mode is a choice among printed
bullets with no price attached, and the engine's mode machinery has no cost
component. They share the announce step and nothing else.

**Buyback as a replacement effect registered at cast time.** Rejected: the
destination is already chosen in one function that already reads the stack
item for flashback. A second mechanism to answer the same question in the
same place is how flashback-and-buyback interactions become undefined.

**A `CantCast` bit on `Restriction`.** Rejected: the existing bits are
per-*permanent* and the restriction is per-*cast* — it has to see the spell,
the zone and the caster, which a bit on a battlefield card cannot. The
`CostQuery` shape already carries exactly those.

**Making the gate return a bool.** Rejected: the client has to say *which*
card refused the cast, and a bool would have meant a second call to find out.
