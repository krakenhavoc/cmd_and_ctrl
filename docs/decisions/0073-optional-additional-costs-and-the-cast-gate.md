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

---

## Note (2026-09-19, #719): the kicked record moved under `Card.Provenance`

§5 put the optional costs a spell was paid with onto the permanent it
becomes, as `Card.PaidOptionalCosts []int`, and gave the reason: an
entering permanent's own trigger cannot reach the stack item, so "when
this enters, **if it was kicked**" has nothing to read unless the
resolution path writes the fact down.

#653 reached the identical conclusion from the other direction a week
later, for "sacrifice it **unless it escaped**" (CR 702.138b), and
shipped `Card.Provenance CastProvenance{AltCost, FromZone}`. The two
fields had the same lifecycle, the same two clears (`MoveCard` and
`resetAsNewObjectLocked`), the same clone and snapshot handling, and
their stamps landed three lines apart in the same entry finisher. They
are one question — *how was the spell that became this permanent
cast?* — and CR 400.7d asks it once.

So #719's rebase folded the first into the second. `PaidOptionalCosts`
is now `Provenance.OptionalCosts`; the snapshot key `paidOptionalCosts`
is now `provenance.optionalCosts`. **Nothing else about §5 changes**:
the stamp is still made in the one moment that holds both the landed
permanent and the item, still before `EventETB` so the entry trigger
finds it, and still cleared on the way out (CR 400.7) so a reanimated
Gatekeeper of Malakir was not kicked.

The readers are untouched in shape — `CardKickedTimes`,
`CardPaidOptionalCost` and `OptionalCostTimesPaid` take the same
arguments and answer the same questions, and `ctx.WasKicked()` /
`ctx.KickedTimes()` still read the STACK ITEM for a resolving spell.
Only where the permanent-side ones look has changed.

A snapshot written before this note decodes `paidOptionalCosts` as an
unknown key and the permanent comes back unkicked. That is a one-deploy
window on a field that is a week old, and it errs weaker than printed.

---

## Amendment (2026-09-22, #1210): the ACTIVATION gate, and the chosen card name

§7 built one announce-time answer to "may this player cast this spell at
all?". This amendment builds its twin — "may this player activate this
ability at all?" (CR 602.5a) — in the same shape, in a new file beside it,
and adds the one piece of state §7 named as missing and then went without.

### Why a twin and not a widened `CastGateLocked`

A cast and an activation share the *rule* (CR 101.2, "can't beats may") and
share nothing else. A cast has a spell, a source zone, a face and an
announcement; an activation has a source OBJECT that stays where it is, an
ability INDEX inside that object, and a CR 605.1a mana/non-mana split that no
cast has. Widening `CastQuery` to carry both would have made every field
optional and every restriction start by asking which of the two it was
looking at — which is how Rule of Law comes to have an opinion about
Pithing Needle.

So: `game/activation_gate.go`, `Game.ActivationGateLocked`, and a
field-for-field mirror of §7's shapes, because the shape is the part that
was right.

### 1. `ActivationRestriction`, `ActivationQuery`, `CantActivateError`

```go
type ActivationAbility struct {
    Label string // the ability's printed label, the activation record's key
    Mana  bool   // CR 605.1a: this ability is a mana ability
}

type ActivationQuery struct {
    Game       *Game             // read-only
    Card       Card              // the object whose ability is being activated
    Controller uuid.UUID         // the player activating it
    Source     Card              // the permanent contributing the restriction
    FromZone   ZoneKind          // where that object is (CR 113.6)
    Ability    ActivationAbility // which ability, and whether it is a mana one
}

type ActivationRestriction struct {
    Label      string
    Forbids    func(q ActivationQuery) bool
    ActiveWhen Designation
}
```

`Card` / `Controller` / `Source` / `FromZone` mean exactly what `CastQuery`'s
do, down to the "you" split: `Controller` is the "you" of the ACTIVATION,
`Source.Controller` the "you" of the ABILITY. Linvala restricts everybody
else's creatures; Cursed Totem restricts everybody's, its controller
included.

**`Ability.Mana` is on the query, not on the call site**, and that is the one
decision in this amendment worth arguing. It would have been a line shorter to
have `ActivateManaAbility` skip the gate — except Pithing Needle exempts mana
abilities ("…can't be activated **unless they're mana abilities**") and Cursed
Totem does not ("Activated abilities of creatures can't be activated", full
stop). The exemption is printed on the *card*, so it belongs to the
restriction, and a call site that decided it would make Cursed Totem
unwritable without a second gate. Both paths call the one function; the
restriction asks.

`CantActivateError{Reason, Source}` wraps the existing `ErrCantActivate`, so
every `errors.Is(err, ErrCantActivate)` in the tree — the Arrest and Faith's
Fetters checks, the auto-tapper's, the client's error mapping — still matches
and needed no change. That is the `CantCastError` / `ErrCantCast` shape
exactly.

### 2. Callers: four, and the fourth is the one §7 did not have

- `ActivateCatalogAbility`, after the CR 113.6 zone check (so the ability is
  the one the view published) and **before the timing check**, X, targets and
  every cost. A refused activation costs nothing. Before timing because "can't
  be activated" is the answer that will still be true next turn, where
  `ErrSorcerySpeedRequired` will not; that order is ours, not the rules'.
- `ActivateManaAbility`, at the same point relative to its own gates — after
  the ability is resolved, before the exhaust record and the condition.
- `legal.abilityMovesForSource` and `legal.manaMoves` — #544: a bot is never
  offered a move the engine refuses.
- **`gatherTapSources`** (`game/autotap.go`). The fourth caller, and the one
  the cast gate has no equivalent of: the auto-tapper does not go through
  `ActivateManaAbility` at all — `materializePlanLocked` taps the permanent
  and mints its mana directly — so a plan built without asking would tap a
  Birds of Paradise under a Cursed Totem and produce mana the rule forbids.
  It sits beside the `CanActivateManaAbilities`, sickness and `Condition`
  checks already there, each of which is in that loop for the same reason.

**`ProducibleManaLocked` (CR 106.7) deliberately does NOT ask.** "Could
produce" is a question about what the ability would do *if it resolved*, not
about whether it can be activated; a Reflecting Pool beside a Cursed-Totem'd
Birds still sees {G}. #1183's exhaust narrowing is declared in that file as
the one exception, and this amendment does not add a second.

### 3. The view stamp: `cant_activate` grows a reason

`cant_activate` already existed — as one of the snake_case tokens in
`CardView.Restrictions`, which is the per-permanent Arrest bit and says
nothing about *which* ability or *why*. The board-wide restriction is a fact
about an ability ROW, so the reason lands there:
`ActivatedAbilityView.CantActivate` and `ManaAbilityView.CantActivate`, both
the printed clause, both stamped from the same `ActivationGateLocked` the
engine and the enumerator call, both absent on nearly every row.

It is a sibling of `condition_unmet` and `exhausted` rather than a third
spelling of them, for the reason those two are separate from each other: the
three recover differently and the client says so. A condition may hold again
next turn; an exhaust never does until the object is new; a restriction ends
when somebody kills the artifact.

### 4. `Card.ChosenName` — the piece §7 named and skipped

§7's own file says it: *"A BAN ON A CHOSEN CARD NAME. Meddling Mage and
Nevermore need a choose-a-card-name prompt, and the engine has `NamedTribe`
and `ChosenColor` but no name."* Pithing Needle, Phyrexian Revoker and
Sorcerous Spyglass want the same prompt on the activation side.

`Card.ChosenName string`, built on #1007's `ChosenPlayer` pattern and on S26's
`NamedTribe` before it:

- `PendingChoiceCardName` (`choose_card_name`), queued from the permanent's
  as-enters hook — the same declared simplification
  `creature_type_choice.go` argues at length, and for the same reason (the CR
  614 pipeline pauses only for a LAND, and Pithing Needle is an artifact).
- Answered with `{card_name: "…"}`. **Free text, and validated only for
  shape** — trimmed, non-empty, length-capped. There is no vocabulary to
  validate against: CR 201.2 lets a player name *any* card name, including
  one in no deck at the table and one the server has never seen. That is the
  one place this differs from `NamedTribe`, whose 345-word CR 205.3m list the
  engine does own.
- The wire carries `name_options` — the names of cards in PUBLIC zones (the
  battlefield, every graveyard, the stack) — as a *convenience* for the
  picker, never as the legal set. Public zones only, so the option list
  cannot leak a hidden card; the free-text entry is the general answer.
- Per INSTANCE, carried by clone and by the snapshot (`chosenName`,
  classified `carried` in `snapshot_drift_test.go`), cleared when the
  permanent leaves the battlefield (CR 400.7) by the same two clears
  `ChosenPlayer` uses.
- **NOT copiable** (CR 706.2), and that falls out of where it lives rather
  than out of a rule anybody has to remember: `CopiableValuesOf` projects
  printed characteristics and never looks here, so a Clone of a Pithing
  Needle names its own card as IT enters.
- `CardNameMatches(c Card, name string)` is the one comparison, and it asks
  every FACE (CR 201.2b: naming one half of a split or a modal DFC names the
  card), case-insensitively on trimmed strings.

**Meddling Mage becomes buildable** — a `CastRestriction` reading
`ChosenNameOf(q.Source.InstanceID)` and nothing new — which is the item §7
left open. It is not written here; this amendment supplies the prompt, not
the card.

### Scope, stated

Out, and named so the next issue does not have to rediscover them:

- **A restriction with a DURATION** (Stifle-style, "activated abilities can't
  be activated this turn"). Same answer §7 gave for Silence: it wants #755's
  registries. A third source slots into `ActivationGateLocked` without
  changing its signature; that is the extension point.
- **Split second** stays out here exactly as it stays out of the cast gate —
  it restricts taking an ACTION, and both activation paths keep their own
  check beside the gate call.
- **Loyalty abilities.** CR 606.1 makes them activated abilities and the gate
  covers them, but no card in the catalog restricts them as a class; when one
  arrives it is a bool on `ActivationAbility`, not a second gate.
- **The per-permanent bits stay.** `CantActivate` / `CantActivateMana`
  (Arrest, Faith's Fetters) are a restriction on ONE permanent put there by
  an Aura attached to it, live on `Characteristic.Restrictions`, and are
  checked where they always were. Folding them into the gate would have made
  every restriction walk the battlefield to answer a question layer 6 already
  answered.

## Note (2026-09-22, #1195): the timing read sits BESIDE the gate, not inside it

§7 lists three callers of `CastGateLocked`. A fourth question is asked at the
same three places and is deliberately **not** folded into that function:
CR 307.1's "may this player begin to cast this card right now", answered by
`game.CastTimingOpenLocked` (`server/internal/game/cast_timing.go`,
[ADR 0066](0066-granted-cast-and-play-permissions.md)'s 2026-09-22
amendment).

The two stay apart because they are different answers to the player.
`CastGateLocked` says a cast is **banned** — Rule of Law, Grafdigger's Cage,
a legendary sorcery with no legend out — and the wire spells that
`cant_cast`, a printed clause the client renders as a refusal.
`CastTimingOpenLocked` says a cast is not open **yet**, which is the ordinary
state of every sorcery in every hand on somebody else's turn. Folding the
second into the first would put a `cant_cast` clause on half the cards in
play and make the field meaningless.

What they share is the placement, and that is the part worth copying: one
function, called by `CastSpell` before any cost is paid, by
`legal.castMovesPayingOptional` so a bot is never offered a cast the engine
refuses, and by `protocol.castStampsFor` so `castable_here` is the engine's
answer rather than a rule the client reimplemented. §8's "bans with a
duration are out" is also still true of the gate — the duration-carrying
statements this note points at are timing statements, not bans, and they
carry ADR 0063's `game.Duration`.

---

## Note (2026-09-22, [#1212](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1212)): the mana joined the optional costs on `Card.Provenance`

The note above folded `Card.PaidOptionalCosts` into `Card.Provenance` and gave
the argument: an entering permanent's own trigger cannot reach the stack item, so
"when this enters, **if it was kicked**" has nothing to read unless the resolution
path writes the fact down, and *how was the spell that became this permanent
cast?* is one question CR 400.7d asks once.

The same argument reached the mana a fortnight later. "When this creature enters,
**if mana from a Treasure was spent to cast it**" (Hired Hexblade) and "…**if
`{R}` was spent to cast it**" (Gruul Scrapper) are the identical shape with a
different clause, and `StackItem.Paid.Mana` died on the stack exactly as
`Paid.OptionalCosts` used to. So `CastProvenance` gained `Mana []ManaToken` and
`ManaOnPaper bool`, stamped in the same `stampCastProvenanceLocked`, three lines
from `OptionalCosts`, cleared in the same two places, carried by the same clone
and snapshot. **Nothing about §5 changes**; it has a sibling now.

Two consequences for this ADR's own record, both small:

- `PaidCost.ManaSpent()` — the token slice — is `PaidCost.ManaTokens()`.
  `ManaSpent` is now the name of the VIEW both homes hand out
  (`game.ManaSpent`, ADR 0068's 2026-09-22 amendment §A2), and the six colour
  accessors on `PaidCost` delegate to it rather than reimplementing it.
- `stampCastProvenanceLocked`'s "nothing to say, leave it zero" branch is
  narrower: a cast that spent any mana, or that the engine waived, now says so.
  Only a genuinely free cast from hand still falls through, and the zero record
  is what "no mana was spent to cast it" reads.

§5's readers are untouched in shape — `CardKickedTimes`, `CardPaidOptionalCost`
and `ctx.WasKicked()` take the same arguments and answer the same questions.

---

## Amendment (2026-09-22, #1213): three more cost components

**Sprint:** S44 — mana and cost components. Tracker [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887).

Three rows of [engine-seams.md](../engine-seams.md) were the same gap wearing
three hats — a **cost component** the vocabulary could not express:

| Row | Cards waiting | What was missing |
|---|---|---|
| Return-a-permanent-to-hand cost component | Quirion Ranger, Master Transmuter, Wirewood Symbiote | `AbilityCost` has no "return a permanent you control to its owner's hand" |
| Variable-count sacrifice cost | Radiant Lotus, Grim Hireling | the sacrifice clause is a fixed count; `Register` refuses `Min != Max` and `CountFromX` |
| Discard cost on a non-activated-ability surface | Mox Diamond, Skirge Familiar | `ManaAbilityCost` has no discard component |

They land together because they are one vocabulary. This ADR owns it: §1 said
an optional cost is `AdditionalCost` with a flag rather than a type of its own
*"for the reason ADR 0021 gave the sacrifice component"*, and the same
reasoning decides all three below. Nothing already decided here changes.

### Decision 1 — a return-to-hand cost is `TapOthersCost` one verb over

`game.ReturnToHandCost` (`server/internal/game/return_cost.go`) is
`TapOthersCost`'s shape exactly: `Count`, a `Filter` reusing the `TargetSpec`
vocabulary, `ExcludeSource` for the printed word "another", and a `Label` that
reads like the card. One options walk, one payability predicate, one
validator, one payer — the #544 invariant `counter_cost.go` and
`tap_others_cost.go` already keep, so the engine, the protocol view's picker
and the legal-move enumerator cannot disagree about which permanent pays.

Not a new kind of thing, and deliberately not folded into an existing one:

- **Not `SacrificeOther` with a destination.** A sacrifice is
  `sacrificePermanentLocked` — `EventSacrifice`, a dies-trigger, a CR 701.17a
  reading. A return is an ordinary battlefield exit to a hand. Sharing the
  field would have made every existing reader of a sacrifice clause ask "but
  where does it go".
- **Not `BounceToHandForEffect`.** That is the EFFECT verb (Azorius Chancery's
  trigger, Chain of Vapor). Paying a cost is not an effect, and the difference
  is observable: CR 601.2h / 602.2b pay an announcement's costs as **one
  indivisible step**, so the move may not stop on a CR 903.9 prompt with the
  ability half announced. The payer therefore sets `zoneRoute.MustSettleNow`
  — the same bit, for the same reason, that the cost discard sets (`zone_route.go`,
  "a discard paid as a COST"). A commander returned to its owner's hand as a
  cost goes to the hand without asking; CR 903.9 is a *may*, and a cost that
  cannot ask falls back to the ordinary result.

Paid with the sacrifices, BEFORE the stack item is built, so a leaves-the-
battlefield trigger queued by the payment is drained by the closing
`runStateChecksLocked` and sits **above** the ability (CR 603.3b) — the order
`payCostSacrificesLocked` already establishes. The source may be a legal pick
when the filter admits it (Master Transmuter is an artifact and may return
itself), so the activation path drops its `source` pointer after the payment
exactly as it does after a sacrifice.

The component is declared on `AbilityCost` only. `AdditionalCost` and
`ManaAbilityCost` do **not** get a slot: no printed card pays a cast or a mana
ability with one, and ADR 0021's rule — a component earns its slot when a card
prints it — is the same rule that kept kicker out of this ADR's predecessor
for two sprints. `AbilityCost.Crew` and `ManaAbilityShape.AddCounter` show
both sides of that line; this one is on the "wait for the card" side.

### Decision 2 — a variable sacrifice count is an ANNOUNCEMENT, not a clause shape

`SacrificeCostCount` reads the count off the clause (#747: `Min == Max == N`).
The two printed variable forms are read off the ANNOUNCEMENT instead, and
`sacrifice_cost.go` grows one function that says which:

```go
SacrificeCostBounds(spec, x) (lo, hi int)   // hi == 0 means "no printed ceiling"
```

- **"Sacrifice one or more artifacts"** (Radiant Lotus) is `Min = 1`,
  `Max = 0`: a floor with no ceiling, settled by how many the activator names.
- **"Sacrifice X Treasures"** (Grim Hireling) is `CountFromX`: the bounds are
  the announced X on both sides, so naming a different number is a refused
  announcement rather than a cheap one.

`CountFromX` on a cost clause makes `AbilityCost.DemandsX()` true even when the
mana component has no `{X}` slot. That is the one sentence of `activated.go`
this amendment rewrites — *"X lives in the MANA component and nowhere else"* —
and it rewrites it the way the cast path already reads: `AdditionalCost.PayLifeX`
and waterbend's `TapPermanentsCost` are both an X outside the printed mana cost,
and Grim Hireling is the same thing on an ability. `XSlots()` is untouched and
still counts mana symbols only, so a `CountFromX` sacrifice adds **no generic
demand** — Grim Hireling's `{B}` stays `{B}` at every X.

**The count is recorded, not recomputed.** `PaidCost.Sacrificed` is how many
permanents the component actually took. Radiant Lotus's "three mana … for each
artifact sacrificed this way" is read at resolution, by which time the
artifacts are in graveyards and nothing on the board could count them — the
identical argument `CountersRemoved` was added under (#789), which is why it is
the same record and the neighbouring field rather than a second mechanism.
`Context.Sacrificed()` is the reader.

`effects.Register`'s guard narrows rather than disappears. Still refused, at
boot: a floor below one (a cost that can be paid with nothing is free),
`Max` below `Min` when a ceiling is printed, `AllowSame` (one permanent cannot
pay two sacrifices), `Players`, and — new — `CountFromX` **on a mana ability**,
which has no X to announce (CR 605.3b: there is no stack item to carry one).

**Last-known information is out of scope and stays on the row.** An effect that
reads the sacrificed permanents themselves ("the sacrificed creature's power")
needs LKI of a list, not a count; neither card on the row asks for it.

### Decision 3 — the discard component is declared once and owned by both ability kinds

`effects.ManaAbilityCost.DiscardCards *game.DiscardCost` is the component #660
put on `AbilityCost`, with the same validator and the same payer. This is not a
new decision so much as the fourth application of one already made three times
in this file's neighbourhood — `SacrificeOther`, `RemoveCounters`, `AddCounter`
and `TapOthers` are each one struct with two owners, for the stated reason that
"a card that printed it on a mana ability and on a CR 602 ability would be
paying one clause two ways otherwise".

It pays through `discardCardsLocked` with `DiscardCauseCost`: `EventDiscardCard`
fires once per card, the CR 614 window runs over the exit, madness sees it
(CR 702.35a), and CR 601.2h's indivisible step is expressed by `MustSettleNow`
on the route rather than by a second loop that knows not to prompt. There is no
new discard door, which is the whole point of having one.

**The auto-tapper never plans it.** `autoTapAbilityFor` already declines a life
cost, an add-a-counter cost and a tap-others cost on the same ground: the
planner spends no resource the player was not asked about, and which card to
pitch is a decision, not an inference. A hand-clicked Skirge Familiar is a mana
source; an auto-tapped one is not.

### What this does NOT decide

- **Mox Diamond.** Its discard is inside an entry REPLACEMENT — "you may
  discard a land card instead" — which is a card choice inside a replacement
  effect, a prompt seam rather than a cost component. It stays on the
  *Reveal-from-hand entry choice* row with the six reveal-lands.
- **Quirion Ranger and Wirewood Symbiote.** Both print "Activate only once each
  turn", and the engine counts an ability's RESOLUTIONS and TRIGGERS this turn
  (`TurnTally`) but not its ACTIVATIONS. Registering them without it would ship
  both cards stronger than printed (#259), so they stay on the
  *Per-source activations-this-turn count* row, which the return row already
  cross-references.
- **A variable TAP-others count.** "Tap X untapped artifacts you control"
  (Secluded Starforge) is the same announcement question one component over.
  #758 stated it out of scope and that row stays open; nothing here narrows it,
  and `SacrificeCostBounds` is the shape it should copy when it lands.

---

## Amendment (2026-09-23, [#1227](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1227)): what the returned permanent was attacking is a paid-cost fact

The #1213 amendment above shipped `AbilityCost.ReturnToHand` as
`TapOthersCost` one verb over. Ninjutsu (CR 702.49a) is the first card family
whose EFFECT has to know something about what the cost returned:

> Return an unblocked attacker you control to hand: Put this card onto the
> battlefield from your hand tapped and **attacking**.

and CR 702.49a's next sentence says attacking *the same player or planeswalker
that the returned creature was attacking*.

### Decision 4 — `PaidCost.ReturnedAttacking`

One field on the existing record, beside `CountersRemoved` (#789) and
`Sacrificed` (#1213), read back through `Context.ReturnedAttacking()`.

It is here rather than anywhere else because it is the same KIND of fact those
two are, and it is unrecomputable for a stronger reason than either. The
returned permanent's battlefield exit clears `Card.AttackingTarget`
(`zone.go`), and LKI carries no combat state — so one line after the bounce,
with the ability not yet even on the stack, nothing in the game can answer the
question. `payReturnToHandCostLocked` therefore reads the field BEFORE it
routes the card and hands the answer back to the announce path, which locks it
onto the stack item with the rest of the payment.

Three things that follow, and each is a decision rather than a detail:

- **It is recorded for every return cost**, not only ninjutsu's. Quirion
  Ranger's Forest is attacking nothing, so the field is `uuid.Nil` and the
  record simply says so. "The engine charged N" and "the card prints N" are
  different facts — `Sacrificed`'s own argument — and a record that only spoke
  up for the interesting case would make every reader ask which it was looking
  at.
- **It is the FIRST returned permanent that was attacking**, in the order the
  activator named them. No printed clause returns more than one, and the
  alternative — a slice, mirrored into the snapshot and the clone — would be
  shape for a card that does not exist.
- **It is not a bit on `ReturnToHandCost`.** The narrowing to "an unblocked
  attacker you control" is a PREDICATE on the clause's filter, so the one
  candidate walk (`ReturnToHandOptionsForEffect`) narrows the client's picker,
  the legal-move enumerator and the validator at once — the #544 invariant the
  #1213 amendment built the component around. A second bit on the component
  would have been a fourth reader with its own opinion.

`PaidCost.IsZero` grows the field so the sparse common case stays sparse, and
the record is copied by value like every other scalar on it.
