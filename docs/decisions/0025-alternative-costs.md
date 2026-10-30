# ADR 0025 — Alternative costs to cast

**Status:** accepted (S22)
**Extends:** [ADR 0021](0021-additional-costs.md) (the *additional*
cost this is deliberately not), [ADR 0019](0019-structured-targeting.md)
(the clause a cost is now allowed to rewrite),
[ADR 0020](0020-activated-abilities.md) (the cost-as-a-struct model).

## Context

The [Aang triage](../decklists/aang-is-so-flashy.md) named this the
keystone gap, and it was right for a reason that isn't obvious from a
card count: **three cards already in the catalog were shipping with
their headline mode missing.**

- **Cyclonic Rift** was a two-mana single bounce. Nobody plays it for
  that. The reason it is a Commander staple is the {6}{U} overload —
  a one-sided board wipe at instant speed.
- **Vandalblast** was a one-mana Naturalize.
- **Slithermuse** could only be hard-cast for {2}{U}{U}, which is not
  how the deck uses it.

Plus a group of blocked cards — evoke, cleave, spree, foretell, plot,
warp — that all say the same thing in different words: *pay this
instead of the mana cost*.

[ADR 0021](0021-additional-costs.md)'s `AdditionalCost` is the other
kind: paid **alongside** the mana cost. Nothing in the engine could
express a cost that **replaces** it, and the two are not variations on
a theme — an additional cost adds to a number, an alternative cost
substitutes for it, and one of them can also rewrite the card's text.

## Decisions

### 1. A slice of structs on the Spec, keyed by a wire-stable name

```go
type AlternativeCost struct {
    Key              string       // "overload", "evoke", "cleave"
    Label            string       // "Overload {4}{R}"
    ManaCost         string       // paid INSTEAD of the printed cost
    Targets          *TargetSpec  // cleave: swap the clause
    ClearsTargets    bool         // overload: delete the clause
    SacrificeOnEntry bool         // evoke: the entry trigger
}
```

A **slice**, not a single struct: spree and the modal-cost cards offer
more than one, and widening later would churn every signature.

**Keyed by name**, because unlike every other announce-time choice this
one has to survive the round trip. `Modes` are indexes into a list the
client can see; an alternative cost has to be claimable by the client
(`cast_spell`'s `alternative_cost`), recorded on the stack
(`StackItem.AltCost`), and read back at resolution
(`ctx.PaidAltCost("overload")`). A stable string is the only thing that
survives all three without a lookup table. `Register` panics on a blank
or duplicated key, so the failure is at boot rather than at a
mysteriously-rejected cast mid-game.

### 2. The cost is not the whole clause — three rewrites ride with it

This is the decision that matters, and the one a "just swap the mana
cost" implementation gets wrong.

Overload does not merely cost {6}{U}. It says *change "target" in its
text to "each"*. If the price swaps but the `TargetSpec` doesn't, the
announce gate still demands a target the spell no longer has, and worse,
the CR 608.2b re-check fizzles the spell the moment that target leaves —
so an overloaded Rift would be *counterable by removing one creature*.
`ClearsTargets` deletes the clause, and `TargetSpecUnderAlternativeCost`
is applied at announce **and** at the re-check so both judge the spell
under the text it was actually cast with.

Cleave is the same machinery pointed the other way: it *swaps in* a
wider clause rather than deleting one. Wash Away goes from "counter
target spell that wasn't cast from its owner's hand" to "counter target
spell", and both clauses are real — the printed one is enforced, not
waved through.

Evoke's rewrite is not a targeting one: *it's sacrificed when it
enters*. Modelled as a triggered ability (CR 702.74a) queued by the
engine as the permanent lands, **not** as a sacrifice folded into
resolution. The difference is the entire reason to evoke a Slithermuse:
the creature genuinely enters, the sacrifice uses the stack so opponents
get a window, and the leaves-the-battlefield trigger then goes on the
stack and draws. Folding it into resolution produces the same hand size
and the wrong stack.

The constructors (`Overload`, `Evoke`, `Cleave`) exist so a card file
cannot get half of this. A hand-written
`game.AlternativeCost{ManaCost: "{4}{R}"}` compiles, casts for four, and
still demands a target — a strictly worse Vandalblast that looks right.

### 3. Alternative replaces the mana cost and *only* the mana cost

Additional costs survive the swap: CR 601.2f is evaluated independently
of the cost chosen at 601.2b, so a card charging both charges both.
Commander tax likewise layers on top (CR 903.8 taxes whatever cost is
being paid), which is why `effectiveCostLocked` applies the tax *after*
the substitution rather than before.

The engine-side test that matters is the arithmetic one: a pool of
exactly {4}{R} must pay an overload and end empty. Six spent would mean
the printed {R} was charged on top — a bug that passes every test which
only checks the cast succeeded.

### 4. Claiming a cost the card doesn't offer is a rejection

Not a fall-back to the printed cost. Silently charging full price for a
cast the player meant to overload is the worst available failure: the
mana is gone, the board isn't wiped, and nothing says why. Same posture
as ADR 0021's "a card with no additional cost that arrives with
`discard_ids` is rejected rather than ignored". An overload arriving
*with* targets is rejected for the same reason — the client is confused
about which cost it is paying.

### 5. The picker opens first, and here that IS a rules-driven order

ADR 0021 was explicit that `DiscardCostModal` opening before targeting
is a UI choice, since the whole cast rides one message. The alternative
cost is different: the answer decides what the targeting prompt is
allowed to offer, so asking afterwards would mean re-opening it. The
chain is alternative cost → discard → sacrifice → X → modes →
targeting.

The picker lists **"Its mana cost"** as an ordinary first option and
defaults to it. An additional cost is a demand; an alternative cost is
an offer, and declining has to be as cheap as accepting.

### 6. The client stays ignorant of what the keywords mean

`AlternativeCostView` carries the target clause **the spell has under
that cost**, already resolved server-side. So the client's rule is: an
offer with no `target_mode` casts immediately, an offer with one enters
targeting on the legal set it carries. Nothing in `client/src` knows
what "overload" does, and the next keyword needs no client change.

This also fixes a castability trap: `canCastFromHand` greyed out a
Cyclonic Rift with nothing to bounce, because the *printed* clause had
an empty legal set. It now passes if **any** cost option's clause is
satisfiable.

### 7. Paid off the ADR 0021 debt rather than adding to it

ADR 0021 flagged that `xValue` / `discardIDs` / `sacrificeIDs` had
become three parallel positional parameters threaded through `begin` /
`beginForMode` / `continueCast`, and that a fourth should trigger a
bundle. This was the fourth. They are now one `CastChoices` object, and
`applyCastChoices` is the single place that knows the wire names.

## Consequences

- Cyclonic Rift, Vandalblast and Slithermuse ship with the mode people
  actually play them for. **Wash Away** is new.
- `StackItem` gains two fields, both of which are facts that stop being
  readable a moment after announce: `AltCost` and `CastFromZone`. The
  second is the one-field addition the triage asked for — CR 601.2a
  moves the card to the stack and nothing on it remembers where it came
  from, so Wash Away's clause needs it stamped. (The sibling request,
  a source zone on `EventCast` for Appa's "whenever you cast a spell
  from exile", is still open; this is the StackItem half, not the event
  half.)
- The stack overlay shows the cost paid, because it changes what the
  spell does — a responder needs to know whether the Rift on the stack
  is a wipe or a bounce.
- Airbend's piece 1 is now built. Pieces 2 and 3 (a permission with no
  expiry, exile-a-target-permanent-with-permission) are unchanged.

## Alternatives considered

**A boolean `IsAlternative` on `AdditionalCost`.** Rejected: the two
share a name and nothing else. An additional cost is mandatory,
unnamed, and never touches the card's text; an alternative cost is
optional, has to be named on the wire, and rewrites the clause. One
struct doing both would be a discriminated union pretending to be a
record.

**A parsed cost string ("overload:{4}{R}:each").** Same rejection ADR
0020 and ADR 0021 made: the shapes are few, and a cost mini-language
has to be maintained against a handful of cards.

**Overload as a second mode.** Tempting — `ModeSpec` already carries
per-option target clauses, and "choose one: target, or each" reads
close. Rejected because modes don't change the price. A player could
choose the overload mode and pay {R}, which is not a simplification but
a different and much stronger card. The Vandalblast comment in the
pre-S22 catalog said exactly this, and it was right.

**Evoke as a sacrifice inside `OnResolve`.** Rejected for ADR 0021's
reason one zone over: it produces the right hand size and the wrong
stack. No response window, and the leave-trigger's ordering relative to
the sacrifice inverts.

## Known limitations

- **"Cast this turn" is not checked** on Wash Away. A spell on the
  stack was cast this turn in every situation the engine can currently
  produce, so the clause is satisfied by construction rather than by a
  timestamp. A card that parks a spell on the stack across turns would
  need a real cast-turn stamp.
- **Free alternative costs work but are untested by a real card.**
  `ManaCost: ""` parses to the zero cost, which is exactly right for
  "without paying its mana cost" — but nothing in the catalog exercises
  it yet.
- **Alternative costs that are not mana** (Force of Will's "exile a
  blue card and pay 1 life", Daze's "return an Island") have no shape
  here. `AlternativeCost` is a struct so they can arrive as fields, the
  same way `AdditionalCost` grew its `Sacrifice` clause.
- **Spree and the modal-cost cards** need a cost *per chosen mode*,
  which is a `ModeSpec` × `AlternativeCost` cross product this does not
  attempt. The slice shape is the half of it that exists.

## Amendment (2026-10-01, #1727): a sacrifice as an alternative cost

The "alternative costs that are not mana" limitation above was closed
piecemeal: S28 added `Life`, `ExileFromHand` and `ReturnToHand`, S29
`ExileFromGraveyard`. Each was one more card-shaped component on the
struct, read through one accessor (`cardComponent`) and named on the
wire in one list (`alt_cost_ids`). Dread Return's "Flashback—Sacrifice
three creatures" needed the one cost shape still missing, and that is
the whole of this amendment.

**Decision 1 — `AlternativeCost.Sacrifice` is the additional cost's
clause, not a new one.** It is a `*TargetSpec` over permanents with the
count on `Min` == `Max`, the shape `AdditionalCost.Sacrifice`,
`AbilityCost.SacrificeOther` and `ManaAbilityShape.SacrificeOther`
already carry (#747), built by the same `sacrificeSpec`. It is
validated by `validateSacrificeCostLocked` — exactly N, each named
once, each on the battlefield under the caster's control (CR 701.21a),
each matching the clause, nothing moved on any failure — and paid by
`payCostSacrificesLocked`, one simultaneous exit with an
`EventSacrifice` per permanent. So there is still one sacrifice
validator and one sacrifice payer in the engine. It is NOT a
`ReturnToHand` with a graveyard destination: that would emit no
sacrifice event, and Blood Artist would never see the three creatures.

**Decision 2 — the picks ride `alt_cost_ids`, not `sacrifice_ids`.**
The code had already answered this. `CastSpellParams.AltCostIDs` is
"the cards paid to the non-mana half of the claimed alternative cost",
kept separate from `SacrificeIDs` because the two pay different costs:
one vanishes when the caster declines the offer and one is charged
whichever price is paid. `SacrificeIDs` also has readers that must not
see an alternative cost's sacrifice — `PaidCost.Sacrificed`
(`ctx.Sacrificed()`) and the per-sacrifice discount
(`CostQuery.Sacrificing`) count the additional cost's list, and
`validateAdditionalCostLocked` refuses a non-empty list on a card with
no sacrifice clause. A permanent named to both lists is refused (CR
118.3).

**Decision 3 — the view ships `sacrifice_options`, not `pay_options`.**
The offer's options are built by `sacrificeCostOptions`, the function
the additional and optional costs' pickers read, so they come in
payment order with the clause's bounds, and the client opens its
existing `SacrificeCostModal` on them. An offer whose clause the
caster's board cannot fill is never stamped — `AlternativeCostPayableLocked`
counts the candidates, as it does for escape — so a flashback that
cannot be paid leaves the card not `castable_here`. The enumerator
reads the same `CardPaymentCount` and `AltCostCandidatesLocked` it reads
for escape and Daze, and candidates come in sacrifice payment order.

**Decision 4 — an alternative cost's cards are kept away from the
auto-tapper.** `CastAutoTapExclusions` now includes `AltCostIDs`. A
sacrifice price has no mana, but a cost increase gives it some (Dread
Return flashed back under Thalia owes {1}), and an Eldrazi Spawn named
to the sacrifice must not be cracked for that {1} first. The enumerator
checks affordability with the same exclusion.

**Decision 5 — two keyword constructors.** `effects.FlashbackSacrifice(n,
label, preds…)` keeps everything `Flashback` bundles (graveyard binding,
`ExileOnLeavingStack`) and swaps the price; `effects.SacrificeInstead(n,
label, life, preds…)` is the from-hand "you may sacrifice … rather than
pay this spell's mana cost", with an optional life half. `Register`
refuses a variable count on the component and an offer with two
card-shaped components.

Proof cards: Dread Return (adopted, now `full`), Lava Dart, Fireblast
and Demon of Death's Gate. Still not covered: emerge (CR 702.119 —
the sacrificed creature's mana value reduces the mana half), Firecat
Blitz's "Flashback—{R}{R}, Sacrifice X Mountains", and Cabal Therapy
(its front half needs a card-name choice at resolution).

## Amendment (2026-10-13, [#3172](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3172)): dash, freerunning, a flash surcharge, and what a payment remembers

Four small pieces landed with the "cav-b" caveat slice of the card-completeness tracker, each on a shape this ADR
already has.

**Dash (CR 702.109).** `AlternativeCost.Dash`, built only by `effects.Dash(cost)`, is the third sibling of
`SacrificeOnEntry` and `WarpExile`: a clause attached to the permanent's entry, applied in
`queueAltCostEntryTriggerLocked` (`game/dash.go`). It registers a layer-6 haste scoped effect for as long as the
creature remains on the battlefield, pinned to the object (`ForAsLongAsOnBattlefieldDuration`), and schedules a
CR 603.7 delayed trigger at the next end step (`dash/return`, a registered body) that returns the creature to its
owner's hand. The trigger records the creature's `ObjectEpoch` when it is scheduled and returns it only if it is still
that object (CR 400.7), so a dashed creature that died or was flickered stays where it is. The delayed trigger uses the
stack, so any player may respond. A granted offer may not carry `Dash`.

**Freerunning (CR 702.173).** `effects.Freerunning(cost, creatureType)` is an alternative cost with a `Condition`, shaped
like Spectacle. The condition reads the turn tally of who dealt combat damage to which player (`CombatDamagedPlayers`),
which already recorded the dealer's name and subtypes; it gains a third identity, "was a commander"
(`PlayersDealtCombatDamageThisTurnByCommander`), so an Assassin that died in the same combat still counts.

**Flash for more.** "You may cast this spell as though it had flash if you pay {N} more" is an offer priced at the
printed cost plus {N} with `AsThoughFlash` (`effects.FlashForMore`), the engine path ADR 0129 §5 built for Primal
Prayers. Casting for the printed cost stays sorcery-speed only (CR 601.3c). Modelling a surcharge as an alternative
cost means it cannot be combined with another alternative cost on the same spell; cost reductions apply to it as they
would to the printed cost plus {N}.

**What a payment remembers.** `PaidCost.Convoked` records the creatures tapped to convoke a spell, in the order named
(Lethal Scheme's "each creature that convoked this spell connives"); a copy of the spell does not carry it (CR 707.10).
`CastTally` gains `ManaSpent` and `ManaSpentLast`, the mana spent on spells this turn read from each spell's payment
record as it became cast, so "whenever you expend N" is a cast whose running total crossed N (`CastTally.Expended`,
`effects.WheneverYouExpend`). A payment the engine waived counts as no mana: weaker than printed, never stronger.

Back-out: remove the `Dash` clause and constructor and Ragavan reverts to caveated; each other piece is one constructor
or one field with the test named below.

Tests: `cards/effects/cav_b_dash_test.go`, `cav_b_test.go` (flash surcharge, freerunning, delirium discount, gift),
`cav_b_connive_test.go`, `cav_b_expend_test.go`.

## Amendment (2026-10-13, [#3225](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3225)): a per-offer X ceiling, for "as though it had flash if X is N or less"

Silver Scrutiny prints "You may cast this spell as though it had flash if X is 3 or less." CR 601.3c lets the caster
BEGIN the cast at instant speed and then requires the announced X to satisfy the condition. The two "as though it had
flash" shapes the engine had could not express that: a flat grant (`CastKindAsThoughFlash`) opens the window for every X,
and the offer that opens it for one claim (`AlternativeCost.AsThoughFlash`, `FlashForMore`) bounded nothing. Shipping
either would let a player draw ten cards at instant speed, which is stronger than printed, so the card shipped as a
sorcery with a declared caveat.

**`AlternativeCost.XCeiling int`** (zero is no ceiling) is the condition, carried on the offer. The offer is priced at the
card's own mana cost with `AsThoughFlash`, so claiming it costs nothing extra, and an X above the ceiling is refused
WHILE THAT OFFER IS CLAIMED. The ordinary sorcery-speed cast for the printed cost is not the offer and takes any X, which
is why the number lives on the offer and not on `Spec.XCeiling` (`game/x_ceiling.go`, #2581), which would bound
every cast of the spell. `effects.FlashIfXAtMost(cost, max)` builds it; hand-rolling the offer compiles and casts for
nothing extra, but forgets either half (the window or the bound).

**One function answers every reader.** `game.SpellXCeilingForOfferLocked(caster, key, offer)` is the smaller of the card's
printed ceiling and the offer's, and these now ask through it with the offer in hand:

- the announce gate in `castSpellLocked` (after CR 107.3b's free-cast lock, before the optional costs), refusing with
  `ErrInvalidParam`;
- the enumerator's `printedXCeiling` (`legal/x.go`), which folds it into the ceiling slot a cost already fills and only
  where the cost being searched has an {X}, so at instant speed no move announces X above it and no open range exceeds it;
- the view, which stamps it on the offer as `alternative_costs[].x_max` (already folded with the card's `x_max`; present
  only on an offer that carries one);
- the auto-tap preview, through `Game.SpellXCeilingForCast`, which resolves the claimed offer from the query's
  `alternative_cost` exactly as the cast does, so `x_max` in the preview is the claimed offer's.

The client's X picker takes `xCeilingFor(card, altCost)`: the claimed offer's `x_max`, else the card's. The offer's number
beats the card's because the server already folded the card's into it.

`Register` refuses a negative `XCeiling` and one on a cost with no {X} (the engine locks X at 0 there, so the number
would never be read). A granted offer may not carry one (`checkGrantedAlternativeCosts`): none has been needed, and the
seam has never been checked against it. `onlyMana` treats an offer with a ceiling as more than a price, so
`duplicatesListedPrice` never folds it into the printed row.

Silver Scrutiny is the card (`full`). Claiming the offer in a main phase is still the offer: X is bounded by the claim,
not by the speed it is cast at.

Back-out: pass `nil` for the offer at each of the four readers and the offer stops being bounded (the named test per
reader fails); drop the field and `FlashIfXAtMost`, and Silver Scrutiny returns to a sorcery with its caveat.

Tests: `cards/effects/flash_if_x_at_most_test.go` (gate, `Register` checks, the card), `legal/offer_x_ceiling_test.go`,
`protocol/offer_x_ceiling_view_test.go`, `lobby/auto_tap_preview_offer_x_ceiling_test.go`,
`client/src/lib/xCeilingForOffer.test.ts`.


## Amendment (2026-10-15, [#3772](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3772)): blitz, and casting from a graveyard using it

Blitz (CR 702.152a) is "you may cast this spell by paying [cost] rather than its mana cost. If you do, it gains haste
and 'When this permanent is put into a graveyard from the battlefield, draw a card.' Sacrifice it at the beginning of the
next end step." It is dash with a different ending, and it uses dash's shape.

**`AlternativeCost.Blitz`**, built only by `effects.Blitz(cost)` and its two variants, is a clause attached to the
permanent's entry, applied in `queueAltCostEntryTriggerLocked` (`game/blitz.go`):

- One layer-6 scoped effect for as long as the creature remains on the battlefield, pinned to the object, carries both
  `AddKeywordsMod("haste")` and `GrantAbilitiesMod(game.BlitzDrawGrant)`. The draw is a granted triggered ability
  (ADR 0093), so it is the creature's own: its controller draws, a later "loses all abilities" takes it, and it fires
  from last-known information as the creature dies. A creature that left and came back has neither (CR 400.7).
- A CR 603.7 delayed trigger at the next end step (`blitz/sacrifice`, a registered body) sacrifices it. It records the
  creature's `ObjectEpoch` when it is scheduled and acts only on that object. It also acts only while the trigger's
  controller still controls the creature (CR 701.17a: a player can sacrifice only a permanent they control), so a
  blitzed creature an opponent has stolen stays, with its haste and its draw.

**The draw is one shared bundle.** Every blitz creature gains the same ability, so `effects/blitz.go` files the bundle
once under `game.BlitzDrawGrant` ("blitz/draw") through `registerSharedGrant`, with the same checks and the same map a
`Spec.Grants` bundle gets. A per-card key would be the same ability under sixteen names, and the engine would have to
look up which one to name.

**The riders are exported.** Jaxis, the Troublemaker prints the same three clauses on an activated ability's token copy.
`Game.ApplyBlitzRidersForEffect(cardID, controller, label)` is the entry clause's body, and Jaxis calls it on the token.

**Price components.** `effects.BlitzDiscard(cost)` adds the existing `DiscardFromHand` component (Sabin, Master Monk's
"Discard a card"), and `effects.BlitzPayLife(cost, n)` adds `Life` (Tenacious Underdog). Both are the components Snag
and Timeline Culler already use, so the validator, the payability check, the view's `pay_options` with `discards` and
the client's picker needed nothing new.

**From the graveyard.** "You may cast this card from your graveyard using its blitz ability" is
`effects.BlitzFromHandOrGraveyard(offer)` plus `Spec.CastableZones: [graveyard]`, the shape Timeline Culler's warp
already uses: a `blitz` offer bound to the hand and a `blitz-graveyard` offer bound to the graveyard, at the same price
and label. Both bindings matter. The graveyard opens only through its own offer (rule 3 of `validateCastPathLocked`), so
the card cannot be cast from there for its mana cost. The hand offer is bound explicitly, because an unbound offer may be
claimed from any zone the card can be cast from, and that includes the graveyard this card opens. A plain `Blitz(cost)`
stays unbound, so a blitz creature cast from another zone by a permission (Muldrotha, an impulse exile) can still claim
it, as CR 118.9a allows.

A granted offer may not carry `Blitz` (`checkGrantedAlternativeCosts`), and `onlyMana` treats a blitz offer as more than
a price, as it does dash.

Not done: the bot prices a blitz offer as an ordinary price, as it already does dash, evoke and warp. It does not yet
weigh the creature it gives up at the end step.

Back-out: drop the `alt.Blitz` branch in `queueAltCostEntryTriggerLocked` and a blitzed creature is a discounted
creature that keeps no haste and never dies (`TestSabinBlitzHasteThenSacrificedAndDraws` fails). Drop the grant mod and
nothing draws. Drop the epoch check and a creature that died and came back under the same ID is sacrificed
(`TestBlitzedCreatureThatDiedAndCameBackIsNotSacrificed`). Drop the control check and a stolen one is sacrificed by its old
controller (`TestBlitzedCreatureStolenIsNotSacrificed`). Drop the hand binding and Sabin claims the hand offer from the
graveyard (`TestSabinCastsFromGraveyardOnlyUsingBlitz`).

Tests: `cards/effects/blitz_test.go` (the clause, the zones, the wire), `cards/effects/blitz_cards_test.go` (every card).

## Amendment (2026-10-30, [#4628](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4628)): mayhem, a graveyard price that needs a discard this turn

Mayhem (CR 702.187b) is "As long as you discarded this card this turn, you may cast it from your graveyard by paying
[cost] rather than paying its mana cost", an alternative cost (CR 601.2b, 601.2f-h). It is flashback's place and price
without flashback's exile, behind a condition that is about one object.

**The record.** `Game.DiscardedThisTurn(cardID)` is a cell of the turn tally the discard event writes. The tally's
listener, on `EventDiscardCard` with `NewZone` graveyard, calls `NoteDoneThisTurn(ev.CardID, "discarded")`. That is a
`TurnTally.Resolved` entry keyed by `ObjectTallyKey(card, epoch, label)`, and the epoch is read after the move has bumped
it (`zone.go`), so the key names the object that landed in the graveyard. Nothing else changes: `Resolved` is an existing
snapshot field, so there is no new shape. CR 400.7 comes with the key: a card that is discarded, leaves the graveyard and
returns is a new object, and a card that was milled, destroyed or countered was never discarded. The tally is flushed
with the turn, so "this turn" needs no clock of its own. A discard that a replacement sends somewhere other than a graveyard
records nothing, and nothing mayhem could cast is there.

**The permission.** It is the card's own and conditional, so it rides `Spec.CastableZonesIf` (ADR 0066 amendment
2026-10-13, Lightwheel Enhancements): `effects.MayhemCastableZone()` opens the owner's graveyard while `DiscardedThisTurn`
holds. Every consumer already asks `cardOpensZoneLocked`: the engine, the view's `castable_here`, the bot's enumerator
and the delve ordering. None of them needed a change.

**The price.** `effects.Mayhem(cost)` is a graveyard-bound offer, key `mayhem`, an on-disk identity. It has no
`ExileOnLeavingStack`: a mayhem spell goes to the graveyard when it resolves and is castable again the next time it is
discarded. The zone's bound offer makes the price mandatory out of the graveyard, as for flashback (rule 3 of
`validateCastPathLocked`), so the printed cost is not available there. Timing is the spell's own ("timing rules still
apply"): Swarm, Being of Bees has flash, Abomination does not.

**The gate on the price.** `AlternativeCost.RequiresDiscardedThisTurn` is checked by `validateCastPathLocked` beside
`RequiresGrant`. Without it the card's own zone condition would be the only guard, and rule 4 of that function lets a
stored permission that opens the graveyard (Yawgmoth's Will) claim any offer bound to the zone, so a milled card could be
cast for the cheap price. It is one flag read by the one gate the engine, the view and the enumerator share. It is a bool
rather than a `func(Game, Card)` so that `AlternativeCost` stays data and the closure ratchet does not move.

**Not done.** Mayhem without a cost (CR 702.187c, "You may play this card from your graveyard if you discarded it this
turn") has no card in the catalog and no door for a land: lands are played from a graveyard only under a permission
(Crucible of Worlds), not from a card's own declaration. It waits for its first card. The bot prices a mayhem offer as an
ordinary price, as it does flashback, and does not choose to discard so as to cast.

Back-out: drop the `NoteDoneThisTurn` line in the tally listener and nothing is ever recorded
(`TestMayhemDiscardedThisTurnCastsFromTheGraveyardForTheMayhemCost`). Drop the zone condition (return true) and a milled
card opens the zone (`TestMayhemMilledCardIsNotCastable`). Drop the `RequiresDiscardedThisTurn` check and the price can be
claimed under another permission (`TestMayhemPriceIsRefusedUnderAnotherPermissionWithoutADiscard`). Key the record by card
ID without the epoch and a card that came back is still "discarded" (`TestMayhemCardThatLeftTheGraveyardAndReturnedIsNew`).

Tests: `game/mayhem_test.go` (the record, the flush, the gate against stubs), `cards/effects/mayhem_cards_test.go` (the
seam against the real catalog), `cards/effects/mayhem_seven_cards_test.go` (every card), `legal/mayhem_test.go` (the
enumerator, dispatched), `protocol/mayhem_view_test.go` (`castable_here` and `alternative_costs`).

## Amendment (2026-10-30, [#4586](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4586)): sneak

CR 702.190a: "Sneak [cost]" means "Any time you could cast an instant during your declare blockers step, you may cast
this spell by paying [cost] and returning an unblocked creature you control to its owner's hand rather than paying this
spell's mana cost." CR 702.190b: a permanent spell whose sneak cost was paid enters tapped and attacking the same player,
planeswalker or battle as the returned creature (CR 506.3a: a noncreature permanent enters but is never attacking).
Marvel Super Heroes prints it (Elektra, Daughter of the Hand) and TMNT prints more.

It is ninjutsu (CR 702.49) moved from an activated ability onto a cast, and it reuses ninjutsu's pieces rather than
adding new ones:

- **The price** is an ordinary offer: `ManaCost` plus the `ReturnToHand` component Daze uses, narrowed by
  `effects.UnblockedAttacker`, the predicate ninjutsu's cost reads. The view's `pay_options`, the enumerator's
  `alt_cost_ids` and the announce validator share one candidate walk, so a blocked attacker, a creature at home and an
  attacker whose defender is still declaring (#1279) are offered by none of them. The return goes to the owner's hand
  through the existing payer, with the spell already on the stack.
- **The window** is one bit, `AlternativeCost.Sneak`, read by `CastTimingForOfferOpenLocked`, the function CastSpell,
  the enumerator, the view's per-offer `timing_closed` and `castable_here` already ask. A sneak claim opens instant-speed
  timing for itself, as `AsThoughFlash` does (CR 601.3c), and is shut everywhere but its caster's own declare blockers
  step (`sneakWindowOpenLocked`). A per-player "only as a sorcery" still shuts it, since the caster could not cast an
  instant then. The printed cost keeps the card's own timing, so the picker in combat shows the sneak row alone.
- **The defender** is read in the cast, before the return clears `Card.AttackingTarget`, and kept on the spell's
  `PaidCost.ReturnedAttacking`: ninjutsu's field, already in every snapshot and undo clone, so no schema change. The
  resolution seeds `EntersTapped` and `EntersAttacking` onto the permanent's entry event before the CR 614 pipeline
  (`applySneakEntryLocked`, beside the escape counters), and `stampEntryAttackerLocked` makes it an attacker that was
  never declared (CR 508.4) and is unblocked (CR 508.4d), the door ninjutsu and the attacking tokens share. A defender
  that has become unattackable leaves the creature tapped and not attacking (CR 506.3c), that door's existing posture.

An instant or sorcery with sneak takes the same constructor: the window and the cost apply and there is no entry.
`effects.Sneak(cost)` is the only way to declare it; the key `sneak` is an on-disk identity on `StackItem.AltCost`. A
granted offer may not carry `Sneak` (`checkGrantedAlternativeCosts`), and `onlyMana` treats it as more than a price.

Not done: a copy of a sneaked permanent spell copies the alternative cost paid (CR 707.10) but resolves through the
copy path (CR 707.10f), which does not read this entry clause, so the token enters untapped and not attacking. The bot prices a sneak offer as an ordinary price and does not weigh the attacker it
returns.

Back-out: drop the `Sneak` check in `CastTimingForOfferOpenLocked` and the offer is castable at instant speed in any step
(`TestSneakIsCastableOnlyInYourDeclareBlockersStep` fails). Drop the stamp in CastSpell and Elektra enters tapped and not
attacking (`TestElektraSneaksInTappedAndAttackingTheSamePlayer`). Drop `applySneakEntryLocked` and she enters untapped
and not attacking.

Tests: `cards/effects/sneak_test.go`.

## Amendment (2026-10-30, [#4641](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4641)): a "connives" event, for "whenever a creature you control connives"

Connive entered this ADR with Lethal Scheme (#3172, above) as a catalog helper that drew, discarded and countered but said
nothing to the rest of the table. Iron Monger, Sadistic Tycoon ("Whenever a creature you control connives, put a +1/+1
counter on each Villain you control") needs a thing to watch, so connive gets the event explore has.

`game.EventConnived` (`game/connive.go`) is emitted once per conniving permanent, by `Game.ConnivedForEffect(creature,
controller)`, when the process of CR 701.50a is complete (CR 701.50f): after the discard and the counter, on every branch,
an empty hand and a permanent that has left the battlefield included. `CardID` is the permanent, `Actor` its controller as
it connived. A permanent that left was told to connive as its last controller's (CR 701.50b), so the card side judges
"a creature you control" by `Actor`, not by where the creature is now (`effects.ACreatureYouControlConnived`,
`WheneverACreatureYouControlConnives`). Several permanents conniving together connive one after another (CR 701.50c),
which is how `ConniveEach` already ran, so each is its own event and the trigger fires once per creature. Connive 0 is
no connive (CR 701.50e): `ConniveX` with X of 0 returns before the draw and emits nothing. The four ways the catalog
connives each emit it from their own discard continuation: `ConniveEach`, `ConniveX`, Ledger Shredder and Cyclonus. The
event is silent in the public log (the draw, the discard and the counter are already lines) and persists nothing, so no
snapshot field and no closure change.

Back-out: remove the `ConnivedForEffect` call from `conniveDiscarded` and `TestIronMongerTriggersOnTheConniveNotTheCounter`,
`TestIronMongerTriggersOncePerConnivingCreature` and `TestIronMongerTriggersForACreatureThatLeftBeforeItConnived` fail.

Tests: `cards/effects/iron_monger_sadistic_tycoon_test.go`; the event assertions in `raffine_a_test.go`,
`slice_296e_test.go` (Ledger Shredder) and `cyclonus_the_saboteur_test.go`.
