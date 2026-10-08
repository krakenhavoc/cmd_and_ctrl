# ADR 0135 — Alternative costs that tap, discard, awaken and emerge

**Status:** Proposed · 2026-10-08 · S68 — Cost components and alternative costs (milestone 77)
**Issues:** [#2030](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2030) (alternative costs that tap untapped creatures you control: Prismatic Strands' flashback, Orim's Cure), [#2412](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2412) (alternative costs that discard cards: Snag, Foil), [#2411](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2411) (awaken, CR 702.113) and [#2416](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2416) (emerge, CR 702.119). The owner picked these four on 2026-10-08. Registry rows: `tap-creatures-alternative-cost`, `discard-alternative-cost` and `awaken`; emerge has no row yet.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-08. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote head: `origin/develop`, `origin/main`, `origin/cost-ledger`, `origin/docs/issue-audit`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/playmats`, `origin/fix/2611-marwyn-source-left`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default` and `pr/2326`. The highest number on any of them is 0134 (`0134-combat-you-can-watch.md`, on `origin/develop` and `origin/main`). There are no open pull requests. This ADR takes **0135**.
**Builds on:** [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) (strict payment, Cast anyway, granted alternative costs), [ADR 0129](0129-energy-getting-and-paying-it.md) (energy; its PR 4 adds `AlternativeCost.Energy` and a type filter on the granted offer, and is being built now), [ADR 0109](0109-rule-gates-land-types-mana-and-cost-components.md) (cost components, and a payment fact reaching the effect), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic's prices and `Purpose`), [ADR 0113](0113-small-seams-for-the-s58-deck-requests.md) §1 (`PaidCost.SacrificedObjects`), [ADR 0065](0065-modal-and-multi-target-clauses.md) (multi-clause target statements, #764) and [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) decision 6 (a PR lands every card its seam unblocks, each checked against its full text).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

S68 is about costs. Four of its seams are alternative costs (CR 118.9) that the engine can't charge yet, or that the registry says it can't:

- **Tap creatures instead.** "If you control a Plains, you may tap an untapped creature you control rather than pay this spell's mana cost" (Orim's Cure) and "Flashback—Tap an untapped white creature you control" (Prismatic Strands).
- **Discard instead.** "You may discard a Forest card rather than pay this spell's mana cost" (Snag) and "You may discard an Island card and another card rather than pay this spell's mana cost" (Foil).
- **Awaken.** A second price that also puts +1/+1 counters on a target land you control and makes it a 0/0 Elemental creature with haste.
- **Emerge.** A second price paid by sacrificing a creature, reduced by that creature's mana value.

Every claim below was checked on `origin/develop` at `62fa97c64`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, "effective as of September 25, 2026"). The issue numbers for awaken and emerge are right: awaken is **702.113** and emerge is **702.119** in that edition. Card lists come from the Scryfall dump of 2026-09-24, Commander-legal cards only, one per oracle ID. None of the cards below is catalogued.

### The rules

- **CR 118.9:** "An alternative cost is a cost listed in a spell's text, or applied to it from another effect, that its controller may pay rather than paying the spell's mana cost."
- **CR 118.9a:** "Only one alternative cost can be applied to any one spell as it's being cast." **CR 601.2b** says the same, and makes the choice part of the announcement: "the player announces their intentions to pay any or all of those costs".
- **CR 118.9c:** an alternative cost "doesn't change a spell's mana cost". **CR 118.9d:** "any additional costs, cost increases, and cost reductions that affect that spell are applied to that alternative cost." Commander tax is an additional cost (**CR 903.8**).
- **CR 118.3:** "A player can't pay a cost without having the necessary resources to pay it fully. For example, … a permanent that's already tapped can't be tapped to pay a cost."
- **CR 701.26a:** "To tap a permanent, turn it sideways from an upright position. Only untapped permanents can be tapped."
- **CR 302.6:** summoning sickness is about "A creature's activated ability with the tap symbol or the untap symbol in its activation cost" and attacking. Tapping a creature to pay a spell's cost is neither, so a creature that arrived this turn can pay. The rulings on Battle Screech and Sephara, Sky's Blade say so.
- **CR 702.34a:** "Flashback [cost]" means "You may cast this card from your graveyard … by paying [cost] rather than paying its mana cost" and "If the flashback cost was paid, exile this card instead of putting it anywhere else any time it would leave the stack."
- **CR 701.9a:** "To discard a card, move it from its owner's hand to that player's graveyard." A discard is not an exile; madness and "whenever you discard" see it.
- **CR 205.3i** and **305.6:** Forest, Island, Mountain, Plains and Swamp are land types. Snag's ruling: "You may discard any card with the land type Forest", including Tropical Island and Dryad Arbor.
- **CR 601.2c:** "A spell may require some targets only if an alternative or additional cost … was chosen for it; otherwise, the spell is cast as though it did not require those targets."
- **CR 601.2f:** "The total cost is the mana cost or alternative cost (as determined in rule 601.2b), plus all additional costs and cost increases, and minus all cost reductions. … It can't be reduced to less than {0}." **CR 601.2g:** "Mana abilities must be activated before costs are paid." **CR 601.2h:** "The player pays the total cost. … Partial payments are not allowed."
- **CR 118.7a:** "Effects that reduce a cost by an amount of generic mana affect only the generic mana component of that cost."
- **CR 702.113a:** "Awaken N—[cost]" means "You may pay [cost] rather than pay this spell's mana cost as you cast this spell" and "If this spell's awaken cost was paid, put N +1/+1 counters on target land you control. That land becomes a 0/0 Elemental creature with haste. It's still a land." **CR 702.113b:** the awaken target is chosen "only if that player chose to pay the spell's awaken cost. Otherwise the spell is cast as if it didn't have that target."
- **CR 608.2b:** a spell whose targets are all illegal doesn't resolve; one with some illegal targets "won't affect the illegal targets". **CR 608.2c:** instructions are followed "in the order written".
- **CR 611.2a:** "If no duration is stated, it lasts until the end of the game." **CR 613.1d** (layer 4, types), **613.4b** (layer 7b, set P/T) and **613.4c** (layer 7c, counters).
- **CR 702.119a:** "Emerge [cost]" means "You may cast this spell by paying [cost] and sacrificing a creature rather than paying its mana cost" and "If you chose to pay this spell's emerge cost, its total cost is reduced by an amount of generic mana equal to the sacrificed creature's mana value." **CR 702.119b:** "Emerge from [quality]" sacrifices "a [quality] permanent" instead. **CR 702.119c:** "You choose which permanent to sacrifice as you choose to pay a spell's emerge cost (see rule 601.2b), and you sacrifice that permanent as you pay the total cost (see rule 601.2h)."
- **CR 202.3e:** X in a permanent's mana cost is 0. The emerge rulings add: a token that isn't a copy has mana value 0, a creature worth more than the emerge cost leaves only its coloured part to pay, and "The creature chosen to be sacrificed is still on the battlefield as the cost of the emerge spell is determined and as you activate mana abilities to cast the emerge spell."
- **CR 707.10:** a copy of a spell copies its alternative costs, and "If an effect of the copy refers to objects used to pay its costs, it uses the objects used to pay the costs of the original".

### What exists

The registry rows for these seams are partly stale. What the code has today:

**The alternative-cost model.** `game.AlternativeCost` (`game/alternative_cost.go:44`) is a price plus clauses: `ManaCost`, `Life`, a `Condition` that gates the offer (Snuff Out's "If you control a Swamp", `effects.ControlsA`), a `Targets` rewrite of the target statement (cleave), `ClearsTargets` (overload), `FromZone` (flashback and escape set the graveyard), `ExileOnLeavingStack` (flashback's exile), the entry clauses of evoke and warp, and `Purpose` (ADR 0126 §6). Its card-shaped components are `ExileFromHand`, `DiscardFromHand`, `ReturnToHand`, `ExileFromGraveyard` and `Sacrifice`, each a `TargetSpec` whose `Min` is the count. `cardComponent()` (`:400`) returns the one present; `effects.Register` refuses an offer with two (`effects/registry.go:160`, "an alternative cost carries at most one"). The picks ride `CastSpellParams.AltCostIDs`. One payability predicate (`AlternativeCostPayableLocked`, `:628`), one per-card predicate (`altCostCardOKLocked`, `:680`), one validator (`validateAlternativeCostPaymentLocked`, `:707`), one candidate walk (`AltCostCandidatesLocked`, `:816`) and one payer (`payAlternativeCostLocked`, `:864`) serve the cast path, the view and the enumerator. The payer runs with the spell already on the stack.

**Discard is already a component.** `DiscardFromHand` arrived with retrace (#2528). Its payer is the one discard helper's cost path (`discardCardsLocked` with `DiscardCauseCost`), so madness and discard triggers see it. The view stamps its candidates on `pay_options` (`protocol/view.go:6118`), the enumerator walks it through `AltCostCandidatesLocked`, and `effects.Register` does not tie it to retrace. So Snag, Abolish, Flameshot and Outbreak can be declared today as an offer with an empty `ManaCost` and a `DiscardFromHand` spec for the land type. The registry's "can't discard cards" is wrong for them. Foil is the gap: "an Island card and another card" is two cards under two predicates.

**A set rule over picks.** `TargetSpec.EachOf` (`game/targets.go:598`, #2526) says that the picks of a sacrifice clause must fill a list of kinds one-to-one ("Sacrifice a Swamp and a Forest", Jarad). A kind is data (`SacrificeKind`: subtypes or card types, `sacrifice_cost.go:326`), and the matching is a small backtracking search (`assignSacrificeSet`). It reads battlefield cards only (`sacrificeSetFitsLocked`), and no kind matches "any card".

**Tapping others as a cost.** `game.TapOthersCost` (`tap_others_cost.go:66`, #758) is "tap N untapped [permanents] you control" on an activated ability: `Count`, a non-targeting `Filter`, `ExcludeSource` for "another", and `Label`. It has one options walk (`TapOthersOptionsForEffect`), one payability check, one validator (exactly N, distinct, controlled, untapped, matched without the targeting gate, no summoning-sickness check) and one payer that emits `EventTapCard` for each. Spells tap creatures in three other places: convoke and waterbend (`TapPermanentsCost`, which help pay a mana cost and never replace it, CR 702.51b), teamwork, and escalate's `AdditionalCost.TapCreatures` (Collective Effort). None is an alternative cost.

**Targets that exist only under a cost.** `AlternativeCost.Targets` replaces the spell's target statement when the offer is claimed, and `TargetSpecUnderAlternativeCost` applies it at announce, at the CR 608.2b recheck and in the view. A statement can have several clauses (`TargetSpec.Rest`, #764), clauses may share an object unless `Distinct` says otherwise, and a card reads its picks per clause with `Context.ClauseTarget(slot)`. `Context.PaidAltCost(key)` says which offer was claimed. A copy keeps `AltCost` (`spell_copy.go:383`).

**Turning a land into a creature for good.** Earthbend (`game/earthbend.go`, CR 701.66a) makes a land "a 0/0 land creature with haste" with one scoped-effect record: `AddTypesMod("Creature")`, `AddKeywordsMod("haste")` and `SetBasePTMods(0, 0)`, with `IndefiniteDuration()` pinned to the object (`PinnedTo`), so it ends when the land leaves (CR 400.7) and is plain data in a restore point. It puts counters through `AddCounterByForEffect` (the CR 614 placement window). It animates first and places counters second, and it adds a return trigger. Awaken prints the other order, adds the Elemental subtype (`AddSubtypesMod` exists) and has no return.

**Sacrifice as an alternative cost.** `AlternativeCost.Sacrifice` (#1727: Dread Return's flashback, Fireblast) validates and pays like the additional cost's sacrifice. Its permanents ride `AltCostIDs`, and `CastAutoTapExclusions` (`autotap_exclusions.go:50`) keeps every one of them away from the auto-tapper. `PaidCost.SacrificedObjects` records the additional cost's sacrifices; its comment says an alternative cost's sacrifice "is not recorded here … no printed card reads it". The cost pipeline (`costAfterModifiersLocked`, `mutations.go:2754`) applies the alternative-cost swap and the commander tax, then the board's increases and reductions (`CostQuery`, which already carries the announced sacrifice count for Torgaar), then convoke and delve. Nothing reduces a cost by the mana value of the permanent being sacrificed.

**Cast triggers.** "When you cast this spell" is `effects.WhenYouCastThisSpell` (`annihilator.go:133`), used by the Eldrazi titans.

**The bot.** The enumerator lists one move per claimable offer and caps the payments it offers at `maxEnumeratedCostPayments` (3), sorted cheapest first by the policy's `fuelValue` (`legal/cast.go:905`). It prices the whole cast once (`first.cost`) and checks affordability per payment only through the auto-tap exclusions (`:1355`). The heuristic charges every `alt_cost_ids` entry `fuelValue` (`heuristic/moves.go:534`): a hand card's value, or a permanent's whole value. For a permanent that is only tapped, that is the wrong price; `tapCreatureCost` (`windows.go:45`) is the right one.

**The client.** `AlternativeCostModal` lists the offers. The cast flow then opens the card picker for `pay_options` or the sacrifice picker for `sacrifice_options` (with `each_of` groups), and sends the picks as `alt_cost_ids` (`lib/targeting.ts:202`). An offer's own target clauses reach it as `clauses`.

**Undo.** A cast is one action. The room pushes a clone of the game before it (`ws/room.go:311`) and an undo restores that clone, so taps, discards, sacrifices and counters all come back together. A refused cast is validated before anything is paid and rolled back (`rollbackLocked`). New state needs only to be deep-copied in `clonePaidCost` and kept by a CR 707.10 copy.

### The cards

| Seam | Cards | Count |
|---|---|---:|
| Tap creatures (#2030) | Orim's Cure, Sivvi's Valor, Angelic Favor, Ramosian Rally, Lashknife ("If you control a Plains, you may tap an untapped creature you control"); Prismatic Strands ("Flashback—Tap an untapped white creature"); Battle Screech ("tap three untapped white creatures"), Group Project ("three untapped creatures"); Sephara, Sky's Blade ("pay {W} and tap four untapped creatures you control with flying"); Zahid, Djinn of the Lamp ("pay {3}{U} and tap an untapped artifact"); The Lady of Otaria ("tap three untapped Dwarves") | 11 |
| Discard (#2412) | Snag, Abolish, Flameshot, Outbreak (one land-type card); Foil (an Island card and another card) | 5 |
| Awaken (#2411) | Boiling Earth, Clutch of Currents, Coastal Discovery, Earthen Arms, Encircling Fissure, Mire's Malice, Ondu Rising, Part the Waterveil, Planar Outburst, Rising Miasma, Roil Spout, Ruinous Path, Rush of Ice, Scatter to the Winds, Sheer Drop | 15 |
| Emerge (#2416) | Abundant Maw, Adipose Offspring, Cresting Mosasaurus, Decimator of the Provinces, Distended Mindbender, Drownyard Behemoth, Elder Deep-Fiend, Herigast, Erupting Nullkite, It of the Horrid Swarm, Lashweed Lurker, Mockery of Nature, Twisted Riddlekeeper, Vexing Scuttler, Wretched Gryff; Crabomination (emerge from artifact) | 15 |

The registry's `Waiting` lists name Prismatic Strands and Sivvi's Valor (`tap-creatures-alternative-cost`), Snag (`discard-alternative-cost`), Encircling Fissure (`awaken`) and Distended Mindbender (`revealed-hand-pick-variants`). The rest come from the dump.

Out of this ADR's reach, and why:

- **Heirloom Epic** taps creatures "rather than pay that mana" of an activation, one per mana: that is convoke on an ability, not an alternative cost.
- **The Infamous Cruelclaw** casts the exiled card "by discarding a card rather than paying its mana cost" during its trigger's resolution. The price belongs to a cast permission (`CastPermission.AlternativeCostFor`), and the permission has only retrace's `DiscardLandCard` flag. It goes on the discard row's `Waiting` list with that reason, unless PR 2 finds the permission shape cheap enough to add.
- **Distended Mindbender** also needs one pick of two cards under two filters (`revealed-hand-pick-variants`, #2115). It stays on that row and moves to emerge's `Waiting` list as well.
- **Herigast** gives every creature spell you cast emerge, at a cost equal to its mana cost. That is ADR 0118 §3's granted offer with a type filter, which ADR 0129 PR 4 adds for Nissa, and a computed price. PR 6 lands it after both.

---

## Decision

### 1. Tapping creatures as an alternative cost (#2030)

**Declaration.** `AlternativeCost` gains a sixth card-shaped component, reusing the ability's struct rather than adding a seventh shape:

```go
// TapOthers is "tap N untapped [permanents] you control" as the price
// (CR 118.9): Orim's Cure's creature, Battle Screech's three white
// creatures, Zahid's artifact. The same struct an activated ability's
// cost carries (#758), so the options walk, the validator and the
// payer are the ones abilities already use. ExcludeSource is refused:
// the spell is not on the battlefield.
TapOthers *TapOthersCost
```

`cardComponent()` returns its `Filter`, the battlefield and its `Count`. `altCostCardOKLocked` adds the one check a sacrifice or a return does not need: the permanent is untapped (CR 118.3, 701.26a). `effects.Register` counts it in `altCostCardComponents`, so it can't sit beside another card component, and refuses `ExcludeSource` and the `CountFromX` form (no printed alternative cost taps X). Helpers in `effects/alternative_cost.go`, beside `SacrificeInstead` and `FlashbackSacrifice`: `TapInstead(n, label, condition, preds…)` (Orim's Cure is `TapInstead(1, "an untapped creature you control", ControlsA("Plains"), Creature())`), `FlashbackTap(n, label, preds…)`, and a mana argument for Sephara and Zahid (`TapInsteadPaying("{W}", 4, …)`).

**Payment.** `payAlternativeCostLocked` gains a `TapOthers` arm that calls `payTapOthersCostLocked`, so each permanent emits `EventTapCard` with the spell already on the stack, and a "whenever a creature becomes tapped" trigger goes above it. There is no summoning-sickness check (CR 302.6). The tapped permanents are in `AltCostIDs`, which `CastAutoTapExclusions` already excludes, so the auto-tapper can't also tap one of them for Sephara's {W} or Zahid's {3}{U}.

**Strict payment and Cast anyway.** `force_cast` waives only mana (ADR 0118 §2, ADR 0129 decision 4). The taps are always paid.

**Client.** `AlternativeCostView` gains `tap_options` (a `LegalTargetsView` with `min` = `max` = the count, from `TapOthersOptionsForEffect` with no source), and `pay_label` carries the clause. The cast flow opens the existing tap-others picker that station and Springleaf Drum use (#759), with the prompt "Tap {pay_label} to cast {card}", and sends the picks as `alt_cost_ids`. With exactly as many candidates as the count, the picker is still shown: tapping a blocker is a choice the player should see. The offer appears in `AlternativeCostModal` as "Tap an untapped creature you control" when the board can pay it, and not at all when it can't (#695).

**Bot.** The enumerator needs nothing new: the candidates come from the engine's walk, and affordability is checked per payment through the exclusions. The heuristic prices an `alt_cost_ids` entry by the offer's component. It reads which one from the card view (`tap_options` present), and charges a tapped permanent `tapCreatureCost` (a blocker, plus its attack before combat on its own turn), not `fuelValue`. `cheapestFuelFirst` orders tap candidates by the same price.

**Undo.** One cast, one undo entry. Undo untaps them with everything else.

### 2. Discarding cards as an alternative cost (#2412)

**One card.** No engine change. `effects.DiscardInstead(label, preds…)` builds `AlternativeCost{Key: "discard", ManaCost: "", DiscardFromHand: spec}`. Snag is `DiscardInstead("a Forest card", HasSubtype("Forest"))`. A subtype read from the card in hand matches nonbasic Forests and Dryad Arbor, as Snag's ruling says. PR 2's test proves a hand cast pays it end to end (retrace's tests cover a graveyard cast), and the registry row is corrected.

**Two cards under two rules (Foil).** `TargetSpec.EachOf` is widened in two ways:

- `SacrificeKind` gains `Any bool` ("another card"), so Foil's entries are `[{Label: "an Island card", Subtypes: ["Island"]}, {Label: "another card", Any: true}]`, with the spec's `Min = Max = 2` and the head predicate their union (any card).
- `sacrificeSetFitsLocked` reads the card from the zone the component names: the battlefield for a sacrifice, the hand for a discard. It is renamed `costSetFitsLocked`. `SacrificeKind` keeps its name, to keep the diff small.

"An Island card and another card" is satisfied by any two cards of which one is an Island card, which is what the one-to-one matching checks. Two Islands are fine; one Island and nothing else is not. `effects.Register` allows `EachOf` on `DiscardFromHand` as well as on a sacrifice clause.

**Payment.** Unchanged: one discard through `discardCardsLocked`'s cost path, with the spell on the stack, so madness (CR 702.35) and discard payoffs see every card.

**Client.** A discard offer already sets `pay_options`. PR 2 adds `each_of` to it, in the shape the sacrifice picker reads (`SacrificeSetGroupsForEffect`), and teaches the hand picker to show the two groups and refuse a set that doesn't match, as `SacrificeCostModal` does.

**Bot.** The heuristic already prices a hand card through `fuelValue`. PR 2 also credits `discardPayoff` for an alternative cost's discards, as it does for an additional cost's (ADR 0126's amendment of 2026-10-06), so Mary Read makes Snag's Forest cheaper. The enumerator's set search for `EachOf` is the sacrifice one, fed hand cards.

### 3. Awaken (#2411)

**Declaration.** Awaken is an ordinary printed offer whose target statement is the spell's own plus one clause:

```go
// effects.Awaken is "Awaken N—[cost]" (CR 702.113a). printed is the
// spell's own target statement, or nil for a spell with no target.
func Awaken(n int, cost string, printed *game.TargetSpec) game.AlternativeCost
```

It returns `AlternativeCost{Key: "awaken", Label: "Awaken N—[cost]", ManaCost: cost, Targets: Clauses(printed…, targetLandYouControl)}`. The land clause is the last one, and its slot is `printed.ClauseCount()`: slot 1 for Ruinous Path, slot 0 for Coastal Discovery. CR 702.113b and 601.2c are then the existing rewrite: the land is asked for only when the offer is claimed. The clause does not set `Distinct`, so Earthen Arms may put its own two counters and its awaken counters on the same land (CR 601.2c: one object may answer two instances of "target"). `effects.Register` refuses `Awaken` on a modal spell (none prints it) and refuses a `printed` that differs from the spec's own statement, so the two can't drift.

**Effect.** `effects.AwakenAfter(n, effect)` wraps the spell's effect: run it, then, if `ctx.PaidAltCost("awaken")` and `ctx.ClauseTarget(landSlot)` is still legal, call a new game primitive:

```go
// AwakenForEffect puts n +1/+1 counters on land, then makes it a 0/0
// Elemental creature with haste that is still a land (CR 702.113a),
// with no duration (CR 611.2a), pinned to the object (CR 400.7).
func (g *Game) AwakenForEffect(actor, source, land uuid.UUID, n int) error
```

It is earthbend's animation with three differences, each from the printed text:

1. **Counters first** (CR 608.2c). The counters go on while the land is not yet a creature, through `AddCounterByForEffect`'s CR 614 window. So Doubling Season, which reads "a permanent you control", doubles them, and Hardened Scales, which reads "a creature you control", does not. Earthbend animates first, because its text does.
2. **Elemental.** The record adds the Creature type, the Elemental subtype, haste and base 0/0 in one scoped effect with one timestamp, so the subtype ends with the animation (layers 4, 6 and 7b; the counters are 7c).
3. **No return trigger.**

The animation record shares earthbend's builder (`animateLandLocked(source, land, stamp, subtypes…)`), so earthbend's code becomes a call with no subtypes. If the counters' window pauses on a CR 616 ordering prompt, the animation waits for the resume, as earthbend's does. No new mod kind: the record uses existing kinds, so a restore point needs no bump.

**Resolution and CR 608.2b.** Nothing new. If the land clause is illegal and the printed one is not, the spell resolves without awakening. If every target is illegal, it doesn't resolve, which is the ruling for a spell with no other target (Coastal Discovery). A copy keeps `AltCost` and its targets (CR 707.10), so a copied awaken spell awakens the copy's land.

**Client.** The offer is in the picker as "Awaken 4—{5}{B}{B}". After it is claimed, the targeting flow walks the offer's `clauses` in order, so the player picks the creature, then "target land you control". No new component.

**Bot.** The enumerator walks the offer's clauses like any multi-clause cast. The heuristic needs a price for what awaken buys. `Purpose` gains `AwakenLand int` (the N), declared by the `Awaken` helper on the offer's `Purpose`, and priced as a hasty N/N creature less a fraction for putting a land in the way of creature removal. The weight is tuned under ADR 0126 §8.

### 4. Emerge (#2416)

**Declaration.** Emerge is a sacrifice offer with a cost reduction:

```go
// ReducedBySacrificedManaValue is emerge's "its total cost is reduced
// by an amount of generic mana equal to the sacrificed creature's
// mana value" (CR 702.119a). Only beside a Sacrifice of exactly one.
ReducedBySacrificedManaValue bool
```

`effects.Emerge(cost)` is `{Key: "emerge", ManaCost: cost, Sacrifice: one creature, ReducedBySacrificedManaValue: true}`. `effects.EmergeFrom(quality, cost, preds…)` is CR 702.119b (Crabomination: "an artifact"). `effects.Register` refuses the flag without a one-permanent sacrifice.

**The price.** CR 702.119c settles the creature at announce, and CR 601.2f totals the cost while it is still on the battlefield. So the reduction is a cost reduction in `costAfterModifiersLocked`, with the other reductions and after the increases and the commander tax. `CostQuery` gains `AltSacrificeManaValue int`, read from the named permanent's current mana value (CR 202.3; a token that isn't a copy is 0, X is 0). The engine subtracts it from the generic part only (CR 118.7a) and never below {0} (CR 601.2f). So Elder Deep-Fiend's {5}{U}{U} with a four-drop is {1}{U}{U}, and with a seven-drop it is {U}{U}. Under an effect that makes creature spells cost {1} more, the increase is added before the reduction is taken off. The price depends on the payment, so the price, the auto-tap preview and the auto-tapper all read `params.AltCostIDs`. That is already true of `effectiveCostLocked`'s inputs.

**Mana from the creature being sacrificed.** The ruling (CR 601.2g before 601.2h) lets the player tap the creature for mana and then sacrifice it. Today `CastAutoTapExclusions` keeps every `AltCostIDs` permanent away from the auto-tapper. This ADR narrows that for a sacrificed permanent: the planner may use its `{T}` mana abilities, and may not use one that sacrifices or exiles it (an Eldrazi Spawn named to emerge can't be cracked first). Question 3 asks whether that narrowing reaches every sacrifice cost on a cast or emerge alone.

**Payment.** Unchanged: the existing sacrifice payer, with the spell on the stack, as one simultaneous exit (dies triggers go above the spell). Then the new record: `PaidCost.AltCostObjects []ObjectRef` (`altCostObjects`, additive in schema 7, deep-copied in `clonePaidCost`, kept by a CR 707.10 copy) lists the objects any alternative cost's card component paid with, in the order named. `Context.AltCostPermanents()` reads them as they last existed (CR 608.2h). Adipose Offspring's "X, where X is the sacrificed creature's toughness" is its first reader. The `SacrificedObjects` comment is corrected.

**Cast triggers.** Eleven of the fifteen print "When you cast this spell". That is `WhenYouCastThisSpell`, unchanged. The creature is gone by then (the ruling's last point), which is the order CastSpell already pays and triggers in.

**Client.** An emerge offer sets `sacrifice_options` as Dread Return's does, plus `reduces_by_mana_value: true`. The sacrifice picker shows each creature's mana value and the price after it ("{1}{U}{U}"), from a new per-candidate `price` the view computes through `PriceCastForEffect`. The auto-tap preview then shows the reduced cost.

**Bot.** The enumerator prices each emerge payment on its own: the one-price shortcut (`first.cost`) is skipped for an offer with the flag. It ranks candidates by their value to the policy minus the generic mana they save, and keeps the cap's best affordable payments. Question 5 is about that order. The heuristic charges the sacrifice as it charges Dread Return's.

### 5. What does not change

- **CR 118.9a.** Each of these is one offer; `CastOffersForLocked` and `validateCastPathLocked` already allow one claim per cast. Jodah's granted {W}{U}{B}{R}{G} can't be combined with emerge or awaken.
- **Timing.** An offer carries no timing (ADR 0118 §3). Prismatic Strands is an instant from the graveyard; Battle Screech waits for a main phase.
- **Snapshots.** The only new field is `PaidCost.AltCostObjects`, recorded with `-update-shape`. The offer fields are catalog data, never serialised. `awaken`, `emerge`, `discard` and the tap keys are new `StackItem.AltCost` strings, read by key and ignored by a binary that doesn't know the card's offer. No fixture changes.
- **The closure ratchet.** `AlternativeCost` is reachable from the catalog only, not from `Game` (ADR 0118 §3).
- **ADR 0129 PR 4.** Energy adds `AlternativeCost.Energy`, a resource like `Life`, not a card component. This ADR's fields sit beside it, and both PRs edit `payAlternativeCostLocked`; the second to merge resolves a mechanical conflict.

---

## Delivery

Each PR goes into `develop`, Sprint S68, and lands its engine change and every card it unblocks together (ADR 0106 decision 6). Each checks every card against its full oracle text, flips its registry row, adds a closed-seam fragment under `docs/engine-seams/closed/`, moves any card that needs more onto that seam's `Waiting` list with the reason, and runs `go test ./internal/roadmap/ -update`. Each card PR runs the real-dump audits (`go test ./internal/decks/ -run RealDump`). PRs 3–5 change payment and costs, so they run the nightly E2E on their branch.

| PR | What | Cards | Needs |
|---|---|---|---|
| 1 | **This ADR.** Docs only. | — | — |
| 2 | **Discard** (§2): `DiscardInstead`; `SacrificeKind.Any`; `EachOf` on a hand discard; `each_of` on `pay_options` and the hand picker; the discard-payoff credit; the registry row corrected and closed. Issue #2412. | Snag, Abolish, Flameshot, Outbreak, Foil | 1 |
| 3 | **Tap** (§1): `AlternativeCost.TapOthers`, its arm in the payer and the per-card check, `tap_options`, the picker, the heuristic's tap price; `TapInstead`, `FlashbackTap`, the mana form. Issue #2030. | Orim's Cure, Sivvi's Valor, Angelic Favor, Ramosian Rally, Lashknife, Prismatic Strands, Battle Screech, Group Project, Sephara, Zahid, The Lady of Otaria (if its end-step condition exists; otherwise its own row) | 1 |
| 4 | **Awaken** (§3): `Awaken`, `AwakenAfter`, `AwakenForEffect` and the shared `animateLandLocked`, `Purpose.AwakenLand`; the `awaken` row closes. Issue #2411. | the 15 awaken cards | 1 |
| 5 | **Emerge** (§4): `ReducedBySacrificedManaValue`, `CostQuery.AltSacrificeManaValue`, per-payment pricing in the enumerator, the auto-tap exclusion narrowed, `PaidCost.AltCostObjects`, `reduces_by_mana_value` and per-candidate `price`; a new `emerge` row. Issue #2416. | the 13 emerge cards that need nothing else (all but Distended Mindbender and Herigast) | 1 |
| 6 | **Herigast** (§4): a granted emerge whose price is the spell's own mana cost, on the type filter ADR 0129 PR 4 adds. | Herigast, Erupting Nullkite | 5, ADR 0129 PR 4 |

**Order.** PRs 2–5 can run in parallel after this ADR. PRs 2 and 3 both touch `cardComponent` and the view's offer stamp, and PRs 3 and 5 both touch the cast pricing; the second to merge resolves a mechanical conflict. PR 6 waits on PR 5 and ADR 0129 PR 4.

## Consequences

- 44 cards become implementable, and Herigast after ADR 0129 PR 4. Four of them (Snag, Abolish, Flameshot and Outbreak) need no engine change at all.
- An alternative cost can tap, discard under a set rule, add a target clause through the existing rewrite, and reduce itself by what it sacrificed. Each is a field or a flag on the one `AlternativeCost` struct, read by the one predicate, validator and payer.
- A land can be awakened with the printed order of counters and animation, through the same one-record animation earthbend uses.
- Every alternative-cost payment is recorded on `PaidCost`, so a later card that reads "the sacrificed creature" or "the discarded card" needs no new field.
- The bot prices tapping as tapping, not as losing the creature, and chooses what to emerge by what it saves.

## Out of scope

- **Heirloom Epic** and any convoke-style tap on an activation.
- **The Infamous Cruelclaw**, unless PR 2 finds the cast permission's discard price cheap (see [The cards](#the-cards)).
- **Distended Mindbender's two-filter pick** (#2115).
- **Two card components on one offer** (Lunar Hatchling's escape). Question 2 offers it as the alternative to the `EachOf` widening.

---

## Questions for the owner

Each question lists the recommended option first. The recommendation is the most CR-faithful option in each case.

1. **How a tap alternative cost is declared (§1; CR 118.3, 302.6, 701.26a).**
   - **(a) Recommended:** reuse `TapOthersCost` as a card component on `AlternativeCost`. The options walk, validator and payer are the ones activated abilities already use, so untapped, "you control", no targeting and no summoning-sickness check are already right and tested.
   - (b) A new `TapPermanents *TargetSpec` field shaped like `Sacrifice`, with its own validator. It looks like its siblings, but it repeats #758's rules in a second place.
   - (c) Generalise teamwork's or escalate's `TapCreatures` count onto the offer. It is the least code, but it can't say "white", "with flying", "artifact" or "Dwarves", so six of the eleven cards wait.

2. **Foil's "an Island card and another card" (§2; CR 601.2h, 701.9a).**
   - **(a) Recommended:** widen `EachOf` to a hand discard and add an "any card" kind. It reuses #2526's one-to-one matching, the picker groups and the enumerator's set search, and it is exactly the printed rule.
   - (b) Allow several card components on one offer, each with its own pick list. It also unblocks Lunar Hatchling's two-part escape cost, but it changes the wire (`alt_cost_ids` becomes a list of lists), the view, the enumerator and the client, for one card here.
   - (c) Leave Foil on the `Waiting` list and ship the four one-card discards.

3. **Mana from a permanent you are sacrificing (§4; CR 601.2g, 601.2h, the emerge rulings).**
   - **(a) Recommended:** the auto-tapper may use the `{T}` mana abilities of a permanent named to any sacrifice cost on a cast (emerge, Dread Return's flashback, Fireblast, Village Rites), and never one that sacrifices or exiles it. That is what CR 601.2g allows every time, and it fixes the same gap in the costs that exist today.
   - (b) Only for emerge. It is the smallest change, but Dread Return and Village Rites keep refusing a payment a player could make at a real table.
   - (c) Keep today's exclusion. The player taps the creature by hand first and floats the mana. It is safe, but on a strict table the move list dims an emerge cast that only the sacrificed creature's mana can pay for.

4. **The awaken animation (§3; CR 702.113a, 608.2c, 611.2a).**
   - **(a) Recommended:** a game primitive, `AwakenForEffect`, that places the counters first and then animates, sharing earthbend's one-record animation with the Elemental subtype added and no return trigger. It follows the printed order, so Doubling Season applies and Hardened Scales doesn't, as at a real table.
   - (b) Reuse `EarthbendForEffect` with a flag that drops the return and adds the subtype. It is less code, but it animates first, so Hardened Scales would add a counter it shouldn't.
   - (c) Compose it on the card side from `AddCounterByForEffect` and `RegisterScopedEffectForEffect`. No new game function, but fifteen cards share a helper that knows about pinning and CR 616 pauses, which is engine knowledge in the effects package.

5. **Which creature the bot emerges (§4).**
   - **(a) Recommended:** price each payment on its own, rank candidates by their value to the policy minus the generic mana they save, and offer the best affordable ones up to the cap of three. The bot then sacrifices a spent creature with a high mana value, and never misses a cast that only one creature makes affordable.
   - (b) Keep the cheapest-creature-first order and price each payment on its own. It is simpler, but it offers the 1/1 token first even when a 6-drop would save six mana, and with a cap of three it can miss the one payment that is affordable.
   - (c) Offer every affordable payment, without a cap. It is complete, but the move list grows with the board.

6. **How the bot values what awaken buys (§3; ADR 0126 §6).**
   - **(a) Recommended:** a new `Purpose.AwakenLand` (the N), priced as a hasty N/N creature less a fraction for exposing a land to creature removal, tuned under ADR 0126 §8. The bot pays the awaken cost when the body is worth the extra mana.
   - (b) Treat it as one token (`Purpose.Tokens`). No new field, but a 6/6 and a 2/2 are worth the same, and the land's risk is ignored.
   - (c) Leave it unpriced. The bot never sees a reason to pay more, so it casts awaken spells for their mana cost only.

7. **What the delivery covers (Delivery; ADR 0106 decision 6).**
   - **(a) Recommended:** all six PRs: discard, tap, awaken and emerge with every card each unblocks, then Herigast after ADR 0129 PR 4. That is 44 cards, plus Herigast.
   - (b) PRs 2–5 only. Herigast stays on the emerge row's `Waiting` list until a later sprint.
   - (c) PRs 2 and 3 only (discard and tap, 16 cards). Awaken and emerge wait for a later sprint.
