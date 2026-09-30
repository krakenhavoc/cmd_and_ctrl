# ADR 0100 — Delve, either/or additional costs, and a variable sacrifice count on a cast

**Status:** Accepted · 2026-09-30 · Post-S30 — Rolling deck-driven catalog growth. The owner answered the seven open questions on 2026-09-30; the answers are recorded under [Owner decisions](#owner-decisions-2026-09-30) and folded into the Decisions below. No engine code lands with this ADR.
**Issue:** [#1732](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1732), rows 1, 3 and 6 (Demand Answers, Treasure Cruise, Plumb the Forbidden). It relates to #296 and #1731.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune`, then listed `docs/decisions/` on every remote branch (37 heads). The highest number anywhere is **0098**. 0099, 0101 and 0102 are reserved for ADRs being written at the same time, so this one takes 0100.
**Builds on:**
- [ADR 0021](0021-additional-costs.md): the additional cost is validated at announce and paid once the spell is on the stack.
- [ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md): the payment plan, and the #1213 amendment's variable sacrifice count on an ability.
- [ADR 0020](0020-activated-abilities.md) §21: a variable count is announced and recorded.
- The one pricer (#696, `game.PriceCast`).
- Convoke and waterbend (`game.TapPermanentsCost`, S22).
- ADR 0013 §5af (#1397): a commander paid as a cost is asked before the payment.

---

## Context

Slice 296-a skipped six cards, and three of them wait on cost shapes the cast path cannot express:

| Card | Printed cost | Why it cannot be written |
|---|---|---|
| Treasure Cruise | Delve, {7}{U} | Nothing lets a graveyard card pay for generic mana. |
| Demand Answers | "As an additional cost to cast this spell, sacrifice an artifact or discard a card." | `game.AdditionalCost` requires every component it names. It cannot offer a choice. |
| Plumb the Forbidden | "As an additional cost to cast this spell, you may sacrifice one or more creatures." | `effects.Register` refuses a variable sacrifice clause on a cast. The comment in `validateAdditionalCostLocked` gives the reason: the flat `sacrifice_ids` list is walked in plan order, and a clause with no fixed width cannot be told apart from the next clause's payment. |

The other three rows of #1732 are out of scope: Veil of Summer's turn-scoped "can't be countered", Distant Melody's creature-type choice at resolution, and Grab the Prize's record of the discarded card. Grab the Prize is taken by sub-PR 3 after all (owner decision 6), because the record this ADR adds for the either/or cost is one field away from it.

### How many cards

Counted in the Scryfall dump of 2026-09-24. Art-series and placeholder printings are skipped, and each card is counted once by oracle ID. "Legal" means legal in Commander. No card below is in the catalog today. I checked every oracle ID against `cards/coverage/testdata/oracle/`.

- **Delve.** 30 cards have the keyword, and 28 of them are legal. One more legal card, Teval, Arbiter of Virtue, grants delve ("Spells you cast have delve"). Two legal delve cards also have {X} in their cost: Logic Knot and Empty the Pits.
- **Either/or additional cost.** 45 cards print "As an additional cost to cast this spell, *A* or *B*", where *A* and *B* are two cost verbs; 43 are legal. This leaves out the type disjunctions, which already work, such as "sacrifice an artifact or creature" (Deadly Dispute). The branches pair up like this:

  | Branches | Legal cards |
  |---|---|
  | mana or sacrifice | 10 |
  | discard or sacrifice | 5 (Dusk Mangler adds a third branch: pay 4 life) |
  | mana or discard | 3 |
  | reveal a card from your hand, or mana | 9 |
  | behold, or mana | 6 |
  | blight, or mana | 2 |
  | life or discard, life or sacrifice, life or mana | 1 each |
  | sacrifice or sacrifice (Lethal Throwdown) | 1 |
  | tap an artifact, exile two cards from your graveyard, forage, or "choose a creature you control or reveal a creature card" | 1 each |

- **Variable sacrifice as an additional cost.** 13 cards; 12 are legal. They come in three forms:
  - "Sacrifice X …": Devastating Summons, Eliminate the Competition and Immoral Bargain. None has {X} in its mana cost.
  - "Sacrifice any number of …": Corpse Cobble and Vicious Betrayal.
  - "You may sacrifice any number / one or more …": Plumb the Forbidden, Torgaar, Famine Incarnate, Dargo, the Shipwrecker, Devouring Greed, Devouring Rage, Rottenmouth Viper, and Extus's back face, Awaken the Blood Avatar.

  Four of the cards (Torgaar, Dargo, Rottenmouth and Extus) also say "this spell costs {N} less to cast for each permanent sacrificed this way".

The roadmap registry (`server/internal/roadmap/registry.go`) has no row for any of the three shapes. Its `set-level-sacrifice-cost` row is a different gap. So is the Marches' variable exile from hand: March of Swirling Mist ships with a caveat for it.

### What the rules say

- **CR 702.66a.** Delve means: "For each generic mana in this spell's total cost, you may exile a card from your graveyard rather than pay that mana."
- **CR 702.66b.** "The delve ability isn't an additional or alternative cost and applies only after the total cost of the spell with delve is determined."
- **CR 702.66c.** More than one instance of delve on the same spell is redundant.
- **CR 702.51a and 702.51b.** Convoke is worded the same way, one resource over: it taps creatures, and it can also pay coloured mana. That is why delve belongs next to convoke in the pricer, and not next to kicker or escape.
- **CR 607.2q.** A permanent spell that exiles cards while its cost is paid is linked to the permanent's ability that refers to cards "exiled with [this object]". "The second ability refers only to cards exiled to pay the cost of the spell that became that permanent." Murktide Regent, Soulflayer and Ethereal Forager read these linked cards.
- **CR 601.2b.** The player announces their intentions to pay alternative or additional costs, and announces the value of any variable cost. **CR 601.2f** determines the total cost; additional costs are added before cost increases and reductions. **CR 601.2h** pays the costs, and "Unpayable costs can't be paid."
- **CR 118.3.** A player can't pay a cost without the resources to pay it in full. For an either/or cost, this is what makes the other branch the only choice.
- **CR 118.8a and 118.8d.** Any number of additional costs may apply, and none of them changes the spell's mana value. **CR 202.3** says the same for delve: mana value comes from the mana cost.
- **CR 107.3a.** An X in an additional cost is announced while the spell is being cast. That is where Devastating Summons' "sacrifice X lands" gets its X.
- **CR 119.4.** Life can be paid only by a player whose life total is at least the amount.
- **CR 701.21a.** A player can sacrifice only a permanent they control.
- **CR 903.9.** A commander exiled from a graveyard by delve, or discarded or sacrificed for a branch, can go to the command zone instead.

---

## Decisions

The three shapes share one rule: **every one of them is settled by the announcement, and the price is read from the announcement in one place.** The auto-tap preview, the bot enumerator and `CastSpell` all price through `PriceCast`. A new cost shape that one of them priced for itself would reopen #696.

### 1. Delve is a payment, priced next to convoke

**Declaration.** `Spec.Delve bool` feeds `CardDef.Delve` and is read through `game.DelveFor(caster, card)`. It is a `Spec` slot, not a keyword token, for convoke's reason. Delve changes what may pay for the cost, so the cast path needs a bit it can ask. A token in `canonicalKeywords` would give delve to every deck-imported card whose effect is still manual. Open question 1 asks whether that is what the owner wants.

`DelveFor` is a function rather than a field read so that Teval's "Spells you cast have delve" can join it later. That grant would be a standing statement from a battlefield permanent, derived the way `standingCastPermissions` is, and not a keyword on a stack object. CR 702.66c makes a second source of delve do nothing, so `DelveFor` returns a bool and not a count.

**The options walk.** `Game.DelveOptionsForEffect(caster, castID)` returns the cards in the caster's own graveyard, except the card being cast. The card being cast is on the stack by CR 601.2a, but validation runs before it moves: Hogaak can be cast from the graveyard. Cards whose exit is paused are left out as well (#1445). This one walk is read by the validator, the view and the enumerator, which is the #544 rule every cost component follows.

**The budget is part of the price.** `CastPrice` grows one field:

```go
// DelveBudget is how many graveyard cards this announcement may exile to
// delve (CR 702.66a): the GENERIC mana in the total cost, after the cost
// modifiers and after what the announced convoke taps pay. 0 when the card
// has no delve.
DelveBudget int
```

`costAfterModifiersLocked` becomes: modifiers, then convoke or waterbend, then delve. Delve comes after the modifiers because CR 702.66b says it "applies only after the total cost … is determined". It comes after the taps because convoke can pay coloured symbols and delve can pay only generic ones. Taking the taps first lets convoke's matching put creatures on the coloured symbols, and delve then competes only for the generic that is left. The validator refuses a delve list longer than `DelveBudget`, as convoke refuses over-tapping. It does not truncate the list, because a card exiled for nothing is a card the player lost for nothing.

**Two traps, stated so the implementation does not fall into them.**

- **The "spend mana as though any colour" fold.** `spendAsThoughAny` runs inside `printedCostLocked` and turns coloured symbols into generic ones. That is correct for the mana solver, and wrong for delve: a Breeches-granted Murktide would otherwise delve away its {U}{U}. `ParsedCost` records how many symbols the fold moved (`FoldedColored`), and the budget is `Generic − FoldedColored`, floored at zero. A generic reduction can only lower the budget, which matches the rule, because in paper that reduction hits the generic part too.
- **X.** The budget folds X into generic, as `tapPermanentsAdjusted` already does. Logic Knot at X = 5 can delve five, and Empty the Pits at X = 3 can delve six. `CastPrice.Base` keeps the {X} slot, so the enumerator's X search still works (the #696 note on `Base`).

**Payment.** The cards are exiled at CR 601.2h, at the point where escape's graveyard exile is paid: after the spell is on the stack, and through `routeCardToZoneLocked` with `MustSettleNow`. Anything watching a graveyard sees the cards leave while the spell is on the stack. Murktide's own "whenever an instant or sorcery card leaves your graveyard" does not see its own delve, because Murktide is a spell at that moment, and that matches paper. A commander among the cards goes through the #1397 ask-first check: `askCostCommanderLocked` is handed the delve list with the rest of the moving cards.

**The record.** `PaidCost.Delved []ObjectRef` lists what was exiled to pay the cost, in the order named, with the object epoch. It is a list and not a count because CR 607.2q links the permanent to *those cards*. A card that has since left exile is a new object (CR 400.7) and no longer counts, and the epoch is what lets the engine see that. This is the same reasoning `PaidCost.Exiled` and `PaidTap` are built on.

**Readers.** They are split out as sub-PR 2 (§6):

- `CastCounts.Delved []Card` holds the cards still in exile as that object. `CountersPerDelved(kind, match)` is Murktide's entry clause, built as arithmetic over `CastCounts` like `XCounters`.
- `CastProvenance.Delved` carries the list onto the permanent (CR 400.7d, like `OptionalCosts`). Soulflayer's static and Ethereal Forager's attack trigger read it.

**What delve does not touch.**

- The mana value and `LocksXAtZero`. Delve is a payment, not a cost.
- The Phyrexian strike, which happens later, in `applyCastCostLocked`, and only on coloured symbols.
- The auto-tapper. It never chooses which cards to exile, for the reason it never chooses a discard: which card to spend is a decision, not an inference. It pays whatever `Total` is left after the player's delve.

### 2. Either/or is one mandatory cost with branches, and the branch is announced

**Declaration.** `game.AdditionalCost` grows `Either []AdditionalCost`. A cost with branches has no components of its own. Each branch is an ordinary `AdditionalCost` with a `Key`, which is what the card's resolution asks for, and a `Label`. Branches go in the existing mandatory slot, `Spec.AdditionalCost`, because the cost is mandatory and CR 601.2b only asks *which* one:

```go
AdditionalCost: EitherCost(
    SacrificeCost("an artifact", Artifact()).Keyed("sacrifice"),
    DiscardCost(1).Keyed("discard"),
), // Demand Answers
AdditionalCost: EitherCost(ManaAdditionalCost("{4}").Keyed("mana"),
    SacrificeCost("an artifact or creature", Or(Artifact(), Creature())).Keyed("sacrifice")), // Annihilating Glare
```

Two alternatives were rejected:

- **Two `OptionalCosts` with an "exactly one" rule.** `PaidCost.OptionalCosts` means "the caster chose to pay extra", and `ctx.WasKicked()` and `Card.Provenance` read it that way. A branch that is not optional would be a lie in that record. The enumerator would also have to learn a group constraint that the optional-cost sets do not have.
- **An alternative cost per branch.** CR 118.9a allows one alternative cost per spell. Demand Answers cast with flashback would then have no way to pay its sacrifice.

**Components a branch may carry in the first sub-PR:** `ManaCost`, `DiscardCards`, a fixed-count `Sacrifice`, a new fixed `PayLife int`, and `Blight`. `PayLife` is paid through `PayLifeForEffect` on the cost path, as `PayLifeX` is. `Blight` exists today; `Register` only has to stop requiring that it be optional and keyed `blight`. Together these cover 24 of the 43 legal cards. The rest need components the cast path does not have yet: reveal a card from your hand, behold, tap an untapped artifact, exile two cards from your graveyard, and forage. Forage is itself an either/or. Each of those arrives with its first card, as a new component that branches can use (ADR 0021 §1: a component earns its slot when a card prints it).

`Register` refuses:

- a branched cost that also has components of its own;
- fewer than two branches;
- a branch that is empty, `Optional`, `Repeat > 1`, or itself branched;
- a missing or duplicate branch `Key`;
- a variable sacrifice clause inside a branch. No printed card has one, and it would reopen §3's width question.

**The announcement.** `CastSpellParams.CostBranch *int`, sent on the wire as `cost_branch`, is an index into `Either`. It is **required** on a branched card, and a missing or out-of-range value is refused. It is not defaulted to zero, for the reason `Face` and `PhyrexianLife` give: silently casting at a different price from the one the player chose is the worst failure available, and for Annihilating Glare branch 0 is "pay {4}". On a card with no branches the value is refused if present, the same posture as `discard_ids` on a card with no discard.

**The plan.** `castCostPayments` takes the chosen branch as the plan's mandatory entry. The validator and the payer do not change at all, which is the point of ADR 0073's plan. The flat `discard_ids` and `sacrifice_ids` lists are walked against the branch that was announced. A payload for the other branch fails the exact-consumption check.

**The price.** A branch's `ManaCost` joins the total at CR 601.2f, next to the optional costs' mana and before the modifiers. There is one helper, `AdditionalCostMana(plan)`, which replaces `AddOptionalCostMana` and sums the mana across the whole plan. `printedCostLocked` calls it, and so does the bot enumerator, on the same announcement. So Thalia taxes Lash of the Balrog's "pay {4}" branch once, and a Ghalta-style reduction may reduce it.

**Payability.** `Game.AdditionalCostBranchPayableLocked(caster, castID, branch)` is modelled on `AlternativeCostPayableLocked` (#695). A branch is payable if it has enough cards in hand other than the spell, enough matching permanents, and life of at least `PayLife` (CR 119.4). Mana is deliberately not checked, because CR 601.2g lets the caster tap afterwards. The view, the enumerator and `CastSpell` all read it. A cast whose every branch is unpayable is `castable_here: false` (CR 601.2h: "Unpayable costs can't be paid"), and CR 118.3 is why the other branch is then the only choice.

**The record.** `PaidCost.CostBranch int` holds the chosen index plus one; 0 means there was no either/or cost, which keeps the common case empty. The reader is keyed: `ctx.PaidCostBranch("modified")` is Lethal Throwdown's "if the modified creature was sacrificed". A CR 707.10 copy keeps the value, as `OptionalCosts` does. It is not carried onto the permanent, because no permanent in the 43 reads it. If one does, it joins `CastProvenance` then.

**The discard record (owner decision 6).** Sub-PR 3 also adds `PaidCost.Discarded []uuid.UUID` (`discarded,omitempty`): the cards the additional cost discarded, in the order named, filled by `payAdditionalCostLocked` for a branch's discard and for a plain `DiscardCost` alike. The reader is `ctx.Discarded()`. It takes Grab the Prize (#1732, row 5): "if the discarded card wasn't a land card". The cards are read in the graveyard by instance ID, for the reason `PaidCost.Exiled` gives; a card that has since left the graveyard is read from its last-known record, because the printed clause asks about the card that was discarded, not about where it is now.

### 3. A variable sacrifice count on a cast: one variable clause, and it takes the whole list

The #1213 amendment solved the count for an ability. It left casts out because of the width problem described above. The fix is a rule, not a new wire shape:

> **A cast's payment plan may hold at most one variable sacrifice clause, and if it holds one, no other entry in the plan may carry a sacrifice.** The variable clause then consumes all of `sacrifice_ids`, bounded by `SacrificeCostBounds(spec, x)`.

No card in the dump breaks this rule. None of the 12 has a second sacrifice clause, a kicker that sacrifices, or a branch. `Register` enforces the rule across the mandatory slot, the branches and `OptionalCosts`, so a future card that breaks it is refused at boot rather than paid wrong.

Two alternatives were rejected. A structured per-entry payload (`payments: [{sacrifice_ids}, …]`) would rewrite a wire shape that three payers and the client share, to fix a problem no printed card has. A separately announced count would be redundant, since the list's length is the count, and redundancy is a second chance to disagree.

**The shapes.** `checkSacrificeClause` gains an `allowZero` argument, true only for a cast's additional cost:

- **"Sacrifice X …"** is `SacrificeXCost(label, preds…)`, which sets `CountFromX`. The X is `CastSpellParams.XValue`, the X that {X} spells and `PayLifeX` already use. CR 107.3i says every X on an object has the same value, so sharing it is correct. `Register` refuses `CountFromX` together with `PayLifeX` anyway, because no card prints both (the ADR 0073 Decision 9 posture). Eliminate the Competition's "destroy X target creatures" is the existing `CountFromX` target clause.
- **"Sacrifice any number of …"** and **"you may sacrifice any number / one or more …"** are all `SacrificeAnyNumberCost(label, preds…)`, with `Min 0` and `Max 0`. For these cards the printed "you may" and "any number" mean the same thing: sacrificing none is not paying. Plumb the Forbidden's "When you do" is a cast trigger that reads `PaidCost.Sacrificed ≥ 1`. Nothing about it needs the optional slot. The #1213 guard ("a cost that can be paid with nothing is free") stays for abilities, because it protects against a card-file mistake there. On a cast, zero is printed.

**The record** is the existing `PaidCost.Sacrificed`, read through `ctx.Sacrificed()`. Vicious Betrayal, Devouring Greed and Devouring Rage need nothing more.

**The price.** The per-sacrifice discount cards (Torgaar, Rottenmouth, Dargo and Extus) are self cost modifiers that read the announcement. `CostQuery` grows `Sacrificing int`, the number of permanents the announced variable clause names. The pricer fills it from `params.SacrificeIDs`, the way it already passes `XValue` and `Targets` (ADR 0048 addendum §13). The constructor is `CostsLessPerSacrificed("{2}")`. CR 601.2b announces the count before 601.2f totals the cost, so the modifier sees it. Because this goes through `PriceCast`, the preview and the enumerator price Torgaar with three creatures sacrificed exactly as `CastSpell` charges it.

**Out of scope.** Corpse Cobble's "X is the total power of the sacrificed creatures" needs last-known information for a list. That was named out of scope in #1213 Decision 2 as well. The cost can ship now, but the card waits (open question 5).

### 4. Wire, view and preview

| Where | Addition | Shape |
|---|---|---|
| `cast_spell` | `delve_ids` | `[uuid]`. The graveyard cards, in the order named. |
| `cast_spell` | `cost_branch` | `int`. Required on a branched card, refused on any other. |
| `CardView` | `delve` | `{options: LegalTargetsView, max}`. The options are `DelveOptionsForEffect`. `max` is the budget for the default announcement (from hand, X = 0, no taps), and is shown as a hint only. |
| `AdditionalCostView` | `branches` | `[AdditionalCostView]`. Each branch carries its `key`, `label`, `mana_cost`, `pay_life`, `discard_cards`, `sacrifice_options` and `payable`. |
| `AdditionalCostView.sacrifice_options` | (existing) | `min` / `max` of 0 / 0 is an open count. `count_from_x` is the X form; the sacrifice picker already reads it (ADR 0073 Decision 10). |
| `GET /games/{id}/auto-tap-preview` | `?delve_ids=`, `?cost_branch=` | The response grows `delve_budget`, which is `CastPrice.DelveBudget` for the announcement as it stands. |

The client reads `delve_budget` from the preview it already calls while a cast is being put together. So it never works out a budget from a mana string (#916's rule), and a change to X, the taps, a kicker or the branch updates the cap.

`docs/protocol.md` documents every row in the PR that adds it.

### 5. Client prompts

All three sit in the one cast chain, `handlePlayCard` (#874), and ride the `CastChoices` bundle in `targeting.ts` as `delveIDs` and `costBranch`. No parallel parameter is added.

- **The branch** is a radio group inside the alternative-cost picker, next to the optional-cost toggles. It is the same question ADR 0073 §9 put there: "what am I paying for this?", which CR 601.2b announces all at once. Unpayable branches are shown disabled. Choosing a branch opens that branch's own picker (`DiscardCostModal` or `SacrificeCostModal`), which the chain already has.
- **Variable sacrifice** is `SacrificeCostModal` with an open or X count. For the X form, the number of permanents picked is X, and the numeric X prompt is skipped, as tap-X does (ADR 0073 Decision 10). For a discount card, the running price comes from the preview.
- **Delve** is a card grid over the caster's graveyard: the escape-exile picker, capped at `delve_budget`. It opens after the X prompt and after the convoke picker, because both change the budget. It is skipped when the budget is zero or the graveyard is empty. Picking nothing is always legal. It has a **"Choose for me"** button (owner decision 2), as the sacrifice picker does (ADR 0020 §16): the button fills the budget with the first entries of `delve.options.cards`, which the server sends in the fuel order the bot pays from (§6), and it never confirms for the player.

### 6. The bot enumerator

`legal.castMovesForCard` prices every move through `PriceCastForEffect` with the full announcement, as it does today.

- **Delve.** For each X candidate it offers **two** payments (owner decision 3), both taken in `OrderCostFuel` order (the fuel price `aiseat/heuristic/fuel.go` already charges for escape): the **fewest** graveyard cards that make the cast affordable, and the **full budget**, `min(DelveBudget, graveyard size)` cards. Murktide Regent and Soulflayer want the second. When the two are the same set, one move is offered. If the pool already pays, the fewest is the empty payment. Two payments, not every subset, so the new dimension cannot use up the target and mode budget, which is the discipline the waterbend and discard payments follow. `affordableX` counts the delve budget when it searches, so a bot can reach Logic Knot at X = 3 with an empty pool and three cards in the graveyard.
- **Either/or.** One move per payable branch (at most three), each priced with its `cost_branch`. The existing payment search pays each branch's cards.
- **Variable sacrifice.** Zero when the floor allows it, then up to `maxEnumeratedVariableCounts` (3) positive counts, smallest first, with `sacrificePayments` choosing which permanents. For the X form, X is the count, so the X search and the payment are a single dimension.

`docs/bot.md` records the three policies.

### 7. Snapshot impact

Everything is additive within schema **v7**. There is no version bump.

- `PaidCost.Delved []ObjectRef` (`delved,omitempty`) and `PaidCost.CostBranch int` (`costBranch,omitempty`). `StackItem.Paid` is already carried. `clonePaidCost` deep-copies the new slice, and `PaidCost.IsZero` covers both fields. The shape file is updated with `-update-shape`.
- `CastProvenance.Delved` (sub-PR 2) is additive in the same way.
- A new board is added to `testdata/snapshots/v7/` beside the others: a delved Murktide Regent on the stack, with its cards in exile. The writer adds the file without changing any existing fixture (ADR 0041 phase 3, owner decision 6). A restore then has to find the linked cards by `ObjectRef`.
- The catalog slots (`Spec.Delve`, `AdditionalCost.Either`, `PayLife`, the clause shapes) are not snapshot data.
- `CastSpellParams` is not persisted. A cast parked on a commander's CR 903.9 answer (#1397) captures its parameters by value in `costCommanderFrame.announce`, so `delve_ids` and `cost_branch` come along with no change. That frame stays in the `census:ChoiceResumeFrames` class of the closure ratchet, and its ceiling does not change.

### 8. Delivery in sub-PRs

Each sub-PR is one PR into `develop`, with its own seam row in the roadmap registry and its first cards.

| # | Scope | First cards | Cards whose cost stops being the blocker |
|---|---|---|---|
| 1 | Delve: slot, options walk, `DelveBudget`, pricer step, validator, payer, `PaidCost.Delved`, view, wire, preview, client picker with "Choose for me", two bot payments | Treasure Cruise, Dig Through Time, Murderous Cut | 22 |
| 2 | Delve readers: `CastCounts.Delved`, `CountersPerDelved`, `CastProvenance.Delved`, and Teval's grant in `DelveFor` | Murktide Regent | 4 |
| 3 | Either/or: `Either`, `PayLife`, blight in a branch, `cost_branch`, `AdditionalCostMana`, branch payability, `PaidCost.CostBranch`, `PaidCost.Discarded`, view, client radio, bot | Demand Answers, Bone Shards, Lightning Axe, Grab the Prize | 24 + Grab the Prize |
| 4 | Variable sacrifice on a cast: the one-clause rule, `allowZero`, `SacrificeXCost`, `SacrificeAnyNumberCost`, `CostQuery.Sacrificing`, `CostsLessPerSacrificed` | Plumb the Forbidden, Vicious Betrayal, Torgaar | 11 |

Sub-PRs 1 and 3 are independent of each other. Sub-PR 4 depends on 3 only for `Register`'s cross-slot check, which is easier to write once branches exist. The owner fixed the order as 1 → 3 → 4 → 2 (owner decision 4).

Each card PR still triages its own card, because the cost is not always the card's only blocker. The next section lists the cases I know of.

---

## Cards unblocked, per shape

### Delve

- **Sub-PR 1 (22 cards):** Treasure Cruise, Dig Through Time, Murderous Cut, Become Immense, Gurmag Angler, Hooting Mandrills, Sultai Scavenger, Tombstalker, Shambling Attendants, Death Rattle, Dead Drop, Magmatic Sinkhole, Rite of Undoing, Set Adrift, Tasigur's Cruelty, Will of the Naga, Sibsig Muckdraggers, Tasigur, the Golden Fang, Logic Knot, Empty the Pits, Afterlife from the Loam and Sorcerous Squall.
- **Sub-PR 2 (4 cards):** Murktide Regent, Soulflayer and Ethereal Forager, which read the linked cards (CR 607.2q), and Teval, Arbiter of Virtue, which grants delve.
- **Still blocked by something else:**
  - Temporal Trespass needs extra turns (#753).
  - Hogaak, Arisen Necropolis needs "You can't spend mana to cast this spell"; its convoke and delve both work once sub-PR 1 lands.
  - Necropolis Fiend needs a variable "Exile X cards from your graveyard" on an ability, the question ADR 0073 Decision 7 left open.

### Either/or (sub-PR 3: 24 of the 43 legal cards)

- **Mana or sacrifice:** Annihilating Glare, Deadly Precision, Bayou Groff, Betrayer's Bargain, Eaten Alive, Lash of the Balrog, Louisoix's Sacrifice, Morkrut Behemoth, Spark Harvest, Stir Up Trouble.
- **Discard or sacrifice:** Demand Answers, Bone Shards, Minion Missile, Souls of the Lost. Dusk Mangler has a third branch, pay 4 life.
- **Mana or discard:** Lightning Axe, Pumpkin Bombardment, Titania, Rugged Rumbler. Titania's "Ward—Discard a card or pay {2}" is a separate either/or on a ward and is still missing.
- **Life:** Bitter Triumph, Final Payment, Redirect Lightning.
- **Sacrifice or sacrifice:** Lethal Throwdown. It also needs a "modified creature" predicate, which the engine does not have yet.
- **Blight or mana:** Bogslither's Embrace, Wild Unraveling.

The other 19 wait for their own component, and each becomes a new kind of branch when its first card arrives:

- **Reveal a card from your hand (9):** Daring Buccaneer, Flamekin Bladewhirl, Goldmeadow Stalwart, Sadistic Skymarcher, Silvergill Adept, Squeaking Pie Sneak, Surtland Elementalist, Thunderherd Migration, Wren's Run Vanquisher.
- **Behold (6):** Caustic Exhale, Kinsbaile Aspirant, Lys Alana Dignitary, Mudbutton Cursetosser, Silvergill Mentor, Soulbright Seeker.
- **Tap an untapped artifact:** Disruption Protocol.
- **Exile two cards from your graveyard:** Soaring Stoneglider.
- **Forage:** Feed the Cycle.
- **Choose a creature you control, or reveal a creature card:** Monstrous Emergence.

### Variable sacrifice (sub-PR 4: 11 of the 12 legal cards)

- **"Sacrifice X":** Devastating Summons, Eliminate the Competition, Immoral Bargain.
- **"Any number" or "one or more", read through the count:** Vicious Betrayal, Devouring Greed, Devouring Rage, Plumb the Forbidden. Plumb also needs a cast trigger and `CopySpell` with `Count`, and both already exist.
- **With a per-sacrifice discount:** Torgaar, Famine Incarnate, Rottenmouth Viper, Dargo, the Shipwrecker, and Awaken the Blood Avatar (Extus's back face).
- **Waiting on last-known information for a list:** Corpse Cobble.

---

## Consequences

- One pricer still answers every price. `DelveBudget` and the branch's mana are read from `CastPrice`, so the preview, the enumerator and `CastSpell` agree on the budget and the total.
- There is still one validator (`validateAdditionalCostLocked`) and one payer (`payAdditionalCostLocked`), and neither gets a second walk. The branch only decides which entry goes in the plan. The validator gains two checks: the fixed `PayLife` check (CR 119.4), and the rule that a variable clause, as the plan's only sacrifice entry, takes whatever is left of `sacrifice_ids`.
- #1213's comment ("a cast's sacrifice clause is always a FIXED count") and the `checkSacrificeClause` doc comment are rewritten in sub-PR 4.
- ADR 0021 and ADR 0073 each get a dated amendment pointing here when their sub-PR lands.
- `docs/adding-cards.md` gains a delve paragraph next to the convoke and waterbend (`Spec.TapCost`) note, and an either/or paragraph and a variable-sacrifice paragraph next to `SacrificeCost` and `SacrificeNCost`.

## Owner decisions (2026-09-30)

The owner answered all seven open questions on 2026-09-30. The Decisions above already read this way.

1. **Delve is a `Spec.Delve` slot**, not a keyword token, as recommended (§1). Only catalogued cards delve.
2. **The delve picker gets "Choose for me"**, which fills the budget in the server's fuel order, like the sacrifice picker (ADR 0020 §16). See §5.
3. **Bots are offered two delve payments**: the fewest cards that make the cast affordable, and the full budget. See §6.
4. **The sub-PR order is 1 → 3 → 4 → 2**: delve core, then either/or, then variable sacrifice, then the delve readers and Teval.
5. **Corpse Cobble waits.** Sub-PR 4 ships the cost and leaves the card out, until last-known information exists for a list of sacrificed permanents.
6. **Sub-PR 3 adds `PaidCost.Discarded` and takes Grab the Prize.** See §2.
7. **Each sub-PR adds its own registry row**, with its `Waiting` list and its closed-seam fragment. There is no separate rows PR.

## Open questions for the owner (answered)

The questions as they were put. The answers are above.


1. **Delve as a `Spec` slot or a keyword token.** I recommend `Spec.Delve`, like convoke, so only catalogued cards delve. The alternative is a `"delve"` token in `canonicalKeywords`. That would let every deck-imported delve card delve at once, with its effect still resolved by hand, the way split second works. Which do you want?
2. **"Choose for me" on the delve picker.** Should it fill the budget from the graveyard in the server's fuel order, as the sacrifice picker's button does (ADR 0020 §16), or should delve always be picked by hand?
3. **How many delve payments a bot considers.** One (the fewest cards that make the cast affordable) is the proposal. Should the bot also be offered the full-budget payment, which Murktide Regent and Soulflayer want, at the cost of one more move per cast?
4. **Order of the sub-PRs.** I propose 1 → 3 → 4 → 2: delve first, because it unblocks the most cards, and Murktide's linked readers last. Should variable sacrifice (Plumb the Forbidden, from the #1731 slice) go before the either/or instead?
5. **Corpse Cobble.** Ship the cost in sub-PR 4 and leave the card out until last-known information exists for a list (`PaidCost` gaining a `PaidTap`-style entry for each sacrificed permanent)? Or add that record in sub-PR 4 as well?
6. **Grab the Prize (#1732, row 5).** "If the discarded card wasn't a land card" needs `PaidCost.Discarded []uuid.UUID`, a record of the discard, which is one field next to `CostBranch`. Should sub-PR 3 add it and take Grab the Prize, or should it stay its own row?
7. **Registry rows now or later.** This ADR adds no registry rows, because the PR is a document only. Each sub-PR adds its own row with its `Waiting` list, as listed above. Would you rather have the three rows added now, in their own small PR, so the public roadmap shows the gap before the work lands?

## Amendment (2026-09-30): sub-PR 2 as built (#1732)

Sub-PR 2 is the last of the four, and with it the ADR's scope is done. It follows §1's "Readers" paragraph, with four details the paragraph did not settle:

- **One resolver.** `Game.DelvedCardsForEffect(refs)` turns a delve link into the cards still in exile as the objects delve put there (CR 400.7). `CastCounts.Delved`, Soulflayer's static and Ethereal Forager's trigger all read through it, so the rule is written once.
- **Last-known information carries the link.** `PermanentInfo.Delved` copies `CastProvenance.Delved` into the CR 608.2h record. Ethereal Forager's 2020-04-17 ruling is that its trigger still finds the cards after the Forager has left, and `ctx.SourcePermanent()` is where it finds them. Additive within v7, like `CastProvenance.Delved`; the new corpus board is `delve_linked_permanents.json`.
- **The layers are told when a linked card leaves exile.** `StaticAbility.DependsOnExile` is a fifth conditional invalidation, in the mould of `DependsOnHandSize`: a departure from exile to a hand, a library or the stack bumps nothing else, and Soulflayer would keep a keyword it had lost.
- **The grant is a `Spec` bit.** `Spec.SpellsYouCastHaveDelve` is read off the caster's permanents in `DelveForLocked`, keyed by `CatalogAbilityKey`, so a Teval that has lost its abilities grants nothing. CR 607.2q links only a delve ability printed on the spell, and every card that reads the link prints delve, so the payment record does not note where the delve came from.

A CR 707.10 copy of a delve spell still links to nothing, as sub-PR 1 decided. CR 707.10's sentence about "objects used to pay its costs" could be read to give a token copy of Murktide Regent the original's cards; no ruling says so, and reading it that way would be stronger than the engine can prove, so the weaker reading stands. Soulflayer ships with one caveat: the engine has no "hexproof from" keyword (ADR 0038 §6), so a creature card with one gives it nothing.

Temporal Trespass, listed above as waiting on extra turns, ships in this sub-PR, because ADR 0059's extra turns have landed. Still waiting on other text: Tasigur, the Golden Fang, Afterlife from the Loam and Sorcerous Squall (the Delve row); Hogaak, Arisen Necropolis ("You can't spend mana to cast this spell"); and Necropolis Fiend (a variable exile on an activated ability, the #1297 remainder).
