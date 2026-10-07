# ADR 0129 — Energy: getting {E} and paying it

**Status:** Proposed · 2026-10-07 · S68 — Cost components and alternative costs (milestone 77)
**Issues:** [#1995](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1995) (paying energy as an activation cost: Consulate Surveillance, and Gonti's Aether Heart from the S58 deck requests, [#2033](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2033)). Registry row: `pay-energy-cost`.
**Owner decisions:** none yet. The six questions are under [Questions for the owner](#questions-for-the-owner). Each has a recommended option, listed first, and the sections below are written as if it were chosen. The other options are kept beside them.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-07. I ran `git fetch --all --prune` and listed `docs/decisions/` on all nine remote heads. The highest number on any of them is 0128 (`0128-playmats.md`). The one open pull request (#2516) adds no ADR. This ADR takes **0129**.
**Builds on:** [ADR 0008](0008-counter-mechanics.md) §2 (the `Player.Counters` map and the legacy `Poison` / `Energy` ints), [ADR 0056](0056-infect-wither-toxic.md) Decision 5 (counters on players go through the CR 614 window and emit `player_counter_placed`), [ADR 0109](0109-rule-gates-land-types-mana-and-cost-components.md) (cost components, and a payment fact reaching the effect), [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) (strict payment, Cast anyway, granted alternative costs), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic's pricing and the `purpose` signal), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (Ask, Always or Never on repeated prompts), [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator) and [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) decision 6 (a PR lands every card its seam unblocks, each checked against its full text).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

A player can already get energy. Decoction Module puts an energy counter on its controller each time a creature of theirs enters, and the seat shows the total. Nothing can spend it. `game.AbilityCost` removes counters from permanents but never from a player, so the 45 Commander-legal activated abilities that print "Pay {E}…:" can't be catalogued. Consulate Surveillance waits on it (its shield shipped with ADR 0108 PR 6, #1904), and so does Gonti's Aether Heart. Energy is also paid while abilities resolve ("you may pay {E}{E}. If you do", "sacrifice it unless you pay {E}"), and as an alternative cost.

Every claim below was checked on `origin/develop` at `0e20f52b1`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, "effective as of September 25, 2026").

### The rules

- **CR 107.14:** "The energy symbol is {E}. It represents one energy counter. To pay {E}, a player removes one energy counter from themselves." The glossary entry "Energy Symbol" says the same.
- **CR 122.1:** "A counter is a marker placed on an object or player that modifies its characteristics and/or interacts with a rule, ability, or effect. … Counters with the same name or description are interchangeable." Energy is a counter on a player. Unlike poison (CR 122.1f) and rad (CR 122.1i), no rule gives it an effect, a maximum or a state-based action.
- **CR 701.34a:** proliferate chooses "any number of permanents and/or players that have a counter", so an energy total proliferates.
- **CR 118.3:** "A player can't pay a cost without having the necessary resources to pay it fully." A player with three energy can't pay {E}{E}{E}{E}. **CR 601.2h** (through **CR 602.2b** for an ability): "Partial payments are not allowed. Unpayable costs can't be paid."
- **CR 107.3a:** when an activation cost or an additional cost has an X "and the value of X isn't defined by the text", the controller "chooses and announces the value of X as part of … activating the ability". This covers "Pay X {E}" (Sphinx of the Revelation, HELIOS One, Chthonian Nightmare). **CR 107.1b:** "You can't choose a negative number."
- **CR 602.1a:** "The activation cost is everything before the colon (:). An ability's activation cost must be paid by the player who is activating it."
- **CR 118.12:** "[A player] may [do something]. If [that player] [does, doesn't, or can't], [effect]." … "The action [do something] is a cost, paid when the spell or ability resolves." **CR 118.12a:** "[Do something] unless [a player does something else]" means the same as "[A player may do something else]. If [that player doesn't], [do something]." These cover every energy payment made while an ability resolves.
- **CR 603.12:** a reflexive trigger ("When you do, …", Guide of Souls, Saheeli, Radiant Creator) triggers on whether the action was taken during the resolution.
- **CR 605.1a:** an activated ability that could add mana, has no target and isn't a loyalty ability is a mana ability, whatever else its cost holds. "{T}, Pay {E}: Add one mana of any color" (Aether Hub) is one. **CR 605.3a** and **CR 117.1d:** a mana ability may be activated while paying a cost. **CR 118.3c:** "Activating mana abilities is not mandatory, even if paying a cost is."
- **CR 118.9** and **118.9a:** an alternative cost is paid "rather than paying the spell's mana cost", and "Only one alternative cost can be applied to any one spell as it's being cast." Nissa, Worldsoul Speaker ("You may pay eight {E} rather than pay the mana cost for permanent spells you cast"), Primal Prayers and Amped Raptor apply one.
- **CR 702.56a:** "Replicate [cost]" means "As an additional cost to cast this spell, you may pay [cost] any number of times". Reiterating Bolt's replicate cost is {E}{E}{E}. It is the only Commander-legal card whose spell takes energy as an additional cost (see [Cards](#the-cards)).
- **CR 702.6a** (equip), **702.84a** (unearth), **702.151a** (reconfigure) and **702.177a** (exhaust) each expand to an activated ability "[Cost]: …". Inventor's Axe, Salvation Colossus, Razorfield Ripper and Peema Trailblazer put energy in that cost.
- **CR 614.16** applies counter replacement effects to counters put "on a permanent" by an effect, so Doubling Season does not double energy. Izzet Generatorium and Aether Refinery print their own replacements for energy ("If you would get one or more {E}").

### What exists

**Where energy lives.** Energy is a player counter, stored twice and kept in step:

- `Player.Counters["energy"]` (`game/player.go:315-324`, `game.CounterEnergy` in `counter_types.go:83-86`) is the source of truth since S13.2 (ADR 0008 §2).
- `Player.Energy` (`player.go:95-98`) is the legacy int, mirrored by `mirrorLegacyPlayerCounterLocked` (`player_counters.go:202-217`) for the `set_energy` action and the wire. Its comment still says energy "has no rules effect until mechanics referencing it are implemented".

Poison is stored the same way (`Player.Poison` and `Counters["poison"]`). Both maps and both ints are in the snapshot: `seats[].counters` and `seats[].energy` (`testdata/snapshot_shape/v7.txt:2674`, `:2950`; `snapshot.go:578`, `:641`).

**Getting energy.** "You get {E}" is `AddPlayerCounterForEffect(controller, CounterEnergy, n)` (`decoction_module.go:52`). Since ADR 0056 Decision 5 that runs the CR 614 window (`RepEventCounter` with `CounterPlayer` set) and emits `EventPlayerCounterPlaced` with the delta that landed (`player_counters.go:53-200`). So a replacement on getting energy and a trigger on "whenever you get one or more {E}" both have an event to read today. Proliferate already reaches player counters (`proliferate.go:38`). Final Act already removes a player's energy as one counter kind among all, and Aetheric Amplifier already counts it.

**Paying for things.** `AbilityCost` (`game/activated.go:35-435`) has the CR 119.4 life components (`Life`, `LifeFrom`) and the counter components on permanents: `RemoveCounters` (`:203-243`, with `CounterRemovalCost` in `counter_cost.go:60`) and `AddCounter`. Its comment for `RemoveCounters` says the removal "is not a replaceable event". `ManaAbilityShape` (`effect_hooks.go:218-330`) has its own `LifeCost`, `ManaCost` and `RemoveCounters`. Nothing in either struct takes a counter off the activating player. `AbilityCost` is catalog data and is not in the snapshot. `PaidCost` (`paid_cost.go`) is, inside `stackMeta[].paid` and its siblings.

**Paying while resolving.** `effects.MayPay` (`effects/may_pay.go`) and `PayUnless` share one prompt, `PendingChoicePayUnless`. Its cost is a mana string. ADR 0108 §5 added `PayAction` (`game/pay_unless_action.go:48`) for the two non-mana payments echo and cumulative upkeep print: discard N cards and sacrifice N permanents. There is no prompt that asks a player for a number.

**The auto-tapper.** It plans a mana ability with a payable life cost in the "pain tier" since #2392 (`autotap.go:1224-1226`, `Pain` at `:901`), after every painless plan. It never plans one that adds a counter, sacrifices another permanent or makes restricted mana (`autotap.go:1164-1250`).

**The wire and the client.** `PlayerView` carries `energy` (`protocol/view.go:1503`) and the `counters` map (`:1610`). `PlayerIdentity.svelte` draws an energy marker when the total is above zero (`:501-505`) and sandbox steppers that send `add_player_counter` (`:183-184`, `:463-481`). `counterTypes.ts:77` gives energy a bolt glyph. `ActivatedAbilityView` (`view.go:3157-3260`) has a field for each cost component it shows (`life_cost`, `loyalty_cost`, `discard_cost_n`, …) and none for energy. The cost-symbol renderer (`client/src/lib/manaSymbol.ts:65-71`) knows W, U, B, R, G and C. It has no {E}.

**The bot.** The enumerator prices every non-mana payment on `legal.MoveCost` (`legal/legal.go:236-296`): `Life`, `PhyrexianLife`, `Loyalty`, `Counters`, `Mana` and `Hand`. It picks the largest affordable X for an X activation (`legal/abilities.go:302-330`). The heuristic prices life through `LifeCostValue` (`heuristic/score.go:454`) and counters on permanents by board value (`moves.go:112-125`). It never reads a seat's energy. The board text the model tiers and the MCP seat read prints each seat's life, hand, library and pool (`aiseat/boardtext/boardtext.go:102-106`), and no player counters, so a model seat can't see its own poison or energy.

### The cards

The Scryfall dump of 2026-09-24 has **142** Commander-legal cards with {E} in their oracle text. One, Decoction Module, is catalogued. Grouped by what each needs (a card can be in more than one group):

| Group | Cards | Examples |
|---|---:|---|
| "Pay {E}…" activated ability, fixed amount, not a mana ability | 45 | Consulate Surveillance, Gonti's Aether Heart, Whirler Virtuoso, Bristling Hydra, Aethertorch Renegade, Aetherworks Marvel, Aetherflux Conduit's fifty |
| Mana ability with an energy cost | 4 | Aether Hub, Servant of the Conduit, Solar Transformer, Conversion Apparatus |
| "Pay X {E}" activated ability | 3 | Sphinx of the Revelation, HELIOS One, Chthonian Nightmare |
| Energy in a keyword's cost | 4 | Inventor's Axe (equip), Salvation Colossus (unearth), Razorfield Ripper (reconfigure: "{2} or {E}{E}{E}"), Reiterating Bolt (replicate) |
| Energy as an alternative cost | 3 | Nissa, Worldsoul Speaker; Primal Prayers; Amped Raptor |
| Energy paid while resolving: "may pay", "unless you pay", "pay any amount of", "an amount of {E} equal to" | about 62 | Guide of Souls, Harnessed Lightning, Galvanic Discharge, Confiscation Coup, the Thriving cycle, Lathnu Hellion, Greenbelt Rampager |
| Gets energy and spends none | 22 | Glimmer of Genius, Attune with Aether, Rogue Refiner, the two Puzzleknots, Electrosiphon; and five that react to getting it: Aether Revolt, Brotherhood Scribe, Fabrication Module, Territorial Gorger, Izzet Generatorium |
| Reads "{E} you've paid or lost this turn" | 2 | Izzet Generatorium, Blaster Hulk |

The issue's count of 40 "Pay {E}…:" activation costs is 52 by this grouping: 45 fixed, 4 mana abilities and 3 with X. Peema Trailblazer's exhaust ability is in the 45.

No Commander-legal card says "As an additional cost to cast this spell, pay {E}". The only spell that takes energy as an additional cost is Reiterating Bolt, through replicate.

---

## Decision

### 1. Where a player's energy lives: the counter that already exists

**Energy stays a counter on the player**, `Player.Counters["energy"]`, as CR 107.14 and CR 122.1 describe it. Nothing new is stored. `Player.Energy` stays as the mirror, because `set_energy`, `PlayerView.energy` and the client's marker read it. The one place that writes the map, `setPlayerCounterLocked` with its mirror, is also the only place a payment writes. `player.go`'s comment that energy "has no rules effect" is corrected in PR 1.

Considered and rejected: **a dedicated `Player.Energy` as the source of truth.** It would take energy out of the counter map, so proliferate, Final Act, Aetheric Amplifier, Vorinclex and ADR 0056's replacement window would each need a special case. That contradicts CR 122.1, which makes energy an ordinary counter. **Retiring the legacy int** is a cleanup that touches the wire and the sandbox actions for no rules gain, so it is out of scope.

**Snapshot.** No change for getting or holding energy: `seats[].counters` and `seats[].energy` are already captured. §6's per-turn tally is the only new state, an additive field in schema 7.

### 2. Paying energy on an activated ability: an `Energy` cost component

`AbilityCost` gains two fields beside `Life`:

```go
// Energy is "Pay N {E}" (CR 107.14): remove N energy counters from the
// activating player (CR 602.1a). Zero means no such component.
Energy int

// EnergyX is "Pay X {E}": the announced X (CR 107.3a) is the amount.
// Energy is then the floor and is normally zero.
EnergyX bool
```

`ManaAbilityShape` gains `EnergyCost int` beside `LifeCost`, for Aether Hub and its kind. No mana ability prints "Pay X {E}", so it gets no X field. `effects.Register` refuses `EnergyX` on an ability with no other use for X, which none prints. It refuses a negative `Energy` too.

**Validation.** `activateCatalogAbilityLocked` and `ActivateManaAbility` check the player's energy before anything is paid. The check is part of the same up-front validation as life and counters on permanents, so a player short of energy taps nothing and loses nothing (CR 118.3, CR 601.2h). The refusal is a new sentinel, `ErrInsufficientEnergy`, with the text "Not enough energy (have 2, need 3)".

**X.** `DemandsX` counts `EnergyX`, so the view, the enumerator and the client ask for X exactly as for an {X} in mana. X may not exceed the player's energy (CR 118.3). If a card ever prints {X} in the mana as well as "Pay X {E}" (none does today; HELIOS One's mana is a fixed {3}), X is one number and both components charge it (CR 107.3a). `X` already rides the stack item, so an effect reading "draw X cards" needs nothing new.

**Payment.** One helper, `payEnergyLocked(payer, n, source)`, removes the counters. It is the only path that pays energy: activated abilities, mana abilities, §3's prompts and §5's alternative costs all call it. It does three things:

1. It writes the map and the mirror through `setPlayerCounterLocked` and `mirrorLegacyPlayerCounterLocked`, as every other player counter does.
2. It opens **no** replacement window. CR 107.14 calls paying a removal. No printed card replaces removing counters from a player, and `RemoveCounters` on permanents pays the same way (`activated.go:216-219`).
3. It emits `EventPlayerCounterPlaced` with the negative delta, the payer as `Actor` and the source, through the existing `emitPlayerCounterDeltaLocked`. So the layer version bumps (a static that reads the total, Razorfield Ripper's "+X/+X where X is the amount of {E} you have", stays current), the log shows the change, and §6's tally sees it.

It is paid with the other non-mana costs, after validation and before the stack item settles, in the order the existing payment block already uses for life (CR 601.2h lets these be paid in any order).

**`PaidCost`.** No new field in PR 1. A fixed cost's amount is printed, and a variable one's is the announced X on the stack item. A resolution amount (§3) is carried by the prompt that asked for it. If a later card needs "the energy paid to activate it", `PaidCost.EnergyPaid` is the neighbour of `LifePaid`, additive in schema 7. ADR 0109's shared section notes that an older binary drops unknown keys inside `paid` silently, so the field would be written only by a card that tolerates it missing.

Considered and rejected: **widening `RemoveCounters` with a player form.** `CounterRemovalCost` is about permanents: `From` is a permanent query, `Among` splits across permanents, the any-kind form asks for a kind, and the payment names `CounterSourceIDs`. A player form would have to switch off most of that, and the payer can only ever be the activator (CR 602.1a), so there is nothing to choose. Energy is closer to life: a resource on the player, checked and paid as a number. **Putting {E} inside the `Mana` string** would make energy a symbol `ParseCost` accepts, and every reader of a mana cost would have to know it is not mana. Mana value, cost reduction, the auto-tapper, `spend only` and the strict mana gate would all be at risk. Both are listed in question 1.

### 3. Paying energy while an ability resolves

About 62 cards pay energy as an ability resolves. CR 118.12 and 118.12a make each one a cost paid on resolution, and the engine already has the prompt for that.

**A fixed amount** ("you may pay {E}{E}. If you do", "sacrifice it unless you pay {E}", "pay {E}{E}. If you can't") rides `PendingChoicePayUnless` with a new `PayAction` kind:

```go
// PayActionEnergy removes Count energy counters from the payer
// (CR 107.14). Nothing is named: the answer is yes or no.
PayActionEnergy PayActionKind = "energy"
```

`MayPay` and `PayUnless` take a `PayAction` as well as or instead of a mana `Cost`, as echo already does. "Pay" is offered only when the payer has the energy (CR 118.3), so a short player sees the decline alone. Greenbelt Rampager's mandatory "pay {E}{E}. If you can't" is the same prompt with no decline when the player can pay, and an automatic "can't" when they can't. A reflexive "When you do, …" (Guide of Souls, Saheeli, Radiant Creator) is the existing reflexive trigger behind the yes (CR 603.12). The rider closure sits on the prompt as it does today. A table holding one of these prompts is a restore point exactly when it is today, because the energy kind rides the same unexported `action` field (`pending_choice.go:944`) that discard and sacrifice ride.

ADR 0127's Ask, Always or Never covers these prompts with no extra work. They are `pay_unless` prompts from a card. "Always pay" pays when the energy is there and asks otherwise, which is ADR 0127 §5's rule for mana.

**A chosen amount** ("then you may pay any amount of {E}", "one or more {E}", Vault 112's "Pay any amount of {E}") needs a number. PR 3 adds a prompt, `PendingChoicePayAmount` (`pay_amount`). It carries `min`, `max` and a question. The max is the payer's energy, and the min is 0 or 1 as printed ("one or more" is 1, CR 107.1b). It is answered with `amount`, and the energy is paid through `payEnergyLocked` before the rider runs with the amount. The client answers it with a stepper in the action dock (ADR 0111 §2), with the total shown and Pay and Pay nothing (or Don't pay) buttons. The names are new labels (ADR 0125 §2). The heuristic answers it by `purpose` (§7), and the enumerator offers 0, the max and, when they differ, the smallest amount the card's own threshold names, so a 50-energy prompt is three moves and not 51. Question 3 asks whether this should be a new prompt or a long `option_pick`.

**"An amount of {E} equal to …"** (Confiscation Coup, Jolted Awake, Volatile Stormdrake, Behemoth of Vault 0) is a fixed amount the effect computes at resolution, so it is the `PayActionEnergy` prompt with that `Count`.

### 4. The cost pipeline: strict payment, the permissive posture and Cast anyway

Energy is never waived. Strict, permissive and `force_cast` decide whether **mana** is charged (ADR 0011 §1, ADR 0118 §2: "`force_cast` waives only the mana"). Energy is paid in all three, as life and every other non-mana component are today. A permissive seat tracking mana on paper still has its energy charged, and Cast anyway on Nissa's eight-energy alternative cost still pays the eight. Question 4 asks this, because the sandbox posture is a choice.

The sandbox is not lost. A player whose energy is wrong fixes it with the seat steppers or `set_energy`, which ADR 0111's table keeps as a sandbox control on the seat.

### 5. Energy in the auto-tapper, and in other costs

**Mana abilities with an energy cost** (Aether Hub, Servant of the Conduit, Solar Transformer, Conversion Apparatus) are planned by the auto-tapper in a new **energy tier**. A plan spends energy only when no plan without it pays the cost. It comes before the pain tier, so energy is spent before life. Within the tier, the plan spending the least energy wins. The auto-tap preview lists "Pay {E}" against the source, so the player sees it before confirming. The planner also checks the energy across the whole plan, not per source: two Aether Hubs with one energy between them pay for one coloured pip, not two. This follows #2392's pain tier. Aether Hub's whole purpose is to turn energy into colour. Without the tier, a strict table could never auto-pay a spell whose colour only the Hub provides, and the move list would dim it. Question 2 offers the other orderings, and leaving these sources to the player.

**Alternative costs** (PR 4). `AlternativeCost` and ADR 0118 §3's granted offer gain `Energy int`, paid through `payEnergyLocked` with the rest of the cast's non-mana costs. The offer is listed only when the player has the energy, through the same `AlternativeCostPayableLocked` filter that already gates a life cost. Nissa's offer is a standing static over "permanent spells you cast", which is ADR 0118 §3's seam with a type filter. Primal Prayers adds "as though it had flash" to the same offer. Amped Raptor's offer is the cast permission's own alternative cost (`CastPermission.AlternativeCostFor`) with a computed amount, "energy equal to its mana value".

**Replicate** (PR 4). Reiterating Bolt's replicate cost is an `AdditionalCost` with `Energy: 3` and a count announced at cast (CR 702.56a). The cast is refused when the count times three is more than the player's energy.

**Keyword costs** (PR 4). Equip, unearth and exhaust are activated abilities, so `Energy` on their `AbilityCost` is the whole change. Reconfigure's "{2} or {E}{E}{E}" is two ability rows, one per cost. That is the pattern the `RemoveCounters` comment describes for Heart of Kiran (`activated.go:225-230`).

### 6. Getting energy, and "paid or lost this turn"

"You get {E}" stays `AddPlayerCounterForEffect`, and "you get that many {E}" is the same call with a computed count. ADR 0056's window already lets Izzet Generatorium and Aether Refinery replace it (`RepEventCounter` with `CounterPlayer`, `CounterName == "energy"`, delta above zero). "Whenever you get one or more {E}" is a trigger on `EventPlayerCounterPlaced` with that name and a positive delta. One placement is one event, so "one or more" fires once per placement as printed.

**The tally** (PR 5). Izzet Generatorium ("Activate only if you've paid or lost four or more {E} this turn") and Blaster Hulk ("costs {1} less … for each {E} you've paid or lost this turn") read a per-turn total. `Player.EnergyPaidOrLostThisTurn int` counts every negative energy delta that lands, paid or removed by an effect, and is reset with the other per-turn tallies. It is an additive snapshot field (`seats[].energyPaidOrLostThisTurn`, recorded with `-update-shape`, no version bump). Only PR 5 needs it.

### 7. The bot: the enumerator and the heuristic

**The enumerator** (`legal/`). An activation with `Energy` is offered only when the seat has the energy, checked through the same function the engine validates with, as life is (ADR 0033 §1). For `EnergyX` the largest affordable X is bounded by the seat's energy, as well as by mana and `Options.MaxX`. `MoveCost` gains `Energy int` (`json:"energy,omitempty"`), the energy the move spends. Without it, a policy would read Aethertorch Renegade's "Pay eight {E}" as free. The `pay_unless` and `pay_amount` prompts are answered from the moves §3 lists.

**The heuristic.** Energy is priced as a resource with a flat weight, `Weights.Energy` (a starting value of 0.25 of a card in hand per counter, tuned under ADR 0126 §8's measurement rules).

- Spending: a move's `MoveCost.Energy` costs `Energy × n`, so a cheap repeatable sink is used when the effect is worth more than its counters.
- Getting: ADR 0126 §6's `purpose` gains an `energy` field, the energy a spell, an enters effect or an activated row gives its controller, so "you get {E}{E}" adds `Energy × 2` to `purposeValue`. It is declared by hand on each energy card, like the other purpose fields.
- Resolution amounts: a `pay_amount` rider declares its purpose per unit ("damage", "counters", "cards"), and the heuristic pays the smallest amount that reaches the goal it is already pricing. For Harnessed Lightning, that is the target's toughness. Otherwise it pays nothing.

Question 5 offers the alternatives: price energy by the best sink the bot controls, or leave it unpriced.

**The board text.** Each seat's line prints its non-zero player counters after the pool ("4 energy, 2 poison"), so a model seat and the MCP seat see what they can pay (`boardtext.go:102-106`). It is a shared change: the bot prompt and the MCP seat read the same `Render`.

### 8. What the client shows

- **The seat.** Unchanged. The marker shows the total and the steppers stay, as sandbox controls (ADR 0111's table). The marker becomes a live number the table can watch go down.
- **Costs.** `manaSymbol.ts` learns `E`: an energy glyph on the counter's own colour (`counterTypes.ts:77`, `#e0c050`), named "Energy" for tooltips and `costWords`. Every surface that renders a cost or a label with symbols then draws {E} as a pip rather than as text.
- **Activations.** `ActivatedAbilityView` and the mana-ability view gain `energy_cost` (`int`, omitted at zero) and `energy_cost_x` (`bool`). The popover row shows the cost with its {E} pips. When the seat is short, the row is greyed with `cant_activate` "Not enough energy (have 2, need 3)", the same server text as the refusal, so the ready ring, the click and the engine agree (ADR 0105). An X row's X stepper is capped at the seat's energy.
- **Prompts.** The `pay_unless` dialog names the energy ("Pay {E}{E}?"), and its Pay button is disabled with the reason when the seat is short. `pay_amount` is §3's stepper.
- **The log.** The existing `player_counter_placed` line covers both directions: "Alice gets 2 energy" and "Alice pays 3 energy (Bristling Hydra)". A payment carries the source, so it is told apart from Final Act's removal.
- `docs/protocol.md` documents `energy_cost`, `energy_cost_x`, `MoveCost.energy`, the `energy` purpose field and the `pay_amount` prompt.

---

## Delivery

Each PR lands its engine change and its cards together, test first. Each flips its registry row to implemented or partial, adds a closed-seam fragment under `docs/engine-seams/closed/`, puts any card that needs more on a `Waiting` list with the reason, and checks every card against its full oracle text (ADR 0106 decision 6). Card counts are estimates of the cards that need nothing else. A card needing an unrelated seam (Hightide Hermit's "as though it didn't have defender", Rex, Cyber-Hound's brain counters) goes on that seam's row.

| PR | Engine | Cards (est.) | Needs |
|---|---|---:|---|
| 1 | §1's comment fix; §2: `AbilityCost.Energy` / `EnergyX`, `payEnergyLocked`, `ErrInsufficientEnergy`, `DemandsX`; §7: `MoveCost.Energy`, the enumerator bound, `Weights.Energy`, the `energy` purpose field, the board-text counters; §8: `energy_cost`, `energy_cost_x`, the {E} glyph, the popover reason | about 45: the two waiting (Consulate Surveillance, Gonti's Aether Heart), about 30 more fixed and X activations, and the about 14 "get-only" cards (question 6) | — |
| 2 | §5's mana half: `ManaAbilityShape.EnergyCost`, the auto-tapper's energy tier, the preview line | 4 mana abilities, and the fixed activations PR 1 left for size | PR 1 |
| 3 | §3: `PayActionEnergy`, `MayPay` / `PayUnless` with an energy payment, `pay_amount` (engine, wire, dock stepper, enumerator, heuristic) | about 55 resolution-payment cards, split 3a (fixed amounts) and 3b (chosen amounts) if large | PR 1 |
| 4 | §5's other costs: `AlternativeCost.Energy`, the granted offer's type filter, replicate's `AdditionalCost.Energy`, the keyword rows | 7: Nissa, Primal Prayers, Amped Raptor, Reiterating Bolt, Inventor's Axe, Salvation Colossus, Razorfield Ripper | PR 1 |
| 5 | §6: `EnergyPaidOrLostThisTurn` (snapshot, additive), the reset, the "get one or more" trigger helper | about 6: Izzet Generatorium, Blaster Hulk, Aether Revolt, Brotherhood Scribe, Fabrication Module, Territorial Gorger | PR 1 |

**Order.** PR 1 first. PRs 2, 3 and 5 can then run in parallel. PR 4 follows PR 3. PRs 2 and 4 both touch the cost payment block. The second to merge rebases, and the conflict is mechanical.

**Docs.** PR 1 adds a "Paying energy" recipe to `docs/adding-cards.md`. Each PR updates the `pay-energy-cost` row's notes and adds rows for the families it does not land: `energy-resolution-payments`, `energy-alternative-costs` and `energy-paid-or-lost`, each pointing here. Every row goes through `go test ./internal/roadmap/ -update`, never by hand.

## Consequences

- Energy becomes a resource the engine charges. The 52 activation costs, the four mana abilities, about 62 resolution payments, three alternative costs and replicate become expressible, and about 120 cards become implementable across five PRs.
- One payment helper and one event. Every energy payment emits the same event, so the log, statics that read the total, and "paid or lost" readers see every payment the same way.
- No snapshot version bump. The only new state is one additive per-turn tally.
- The bot can spend energy, and a model seat can see it.
- The auto-tapper can spend energy, after every plan that doesn't.

## Out of scope

- **Retiring `Player.Energy`.** It is kept as a mirror (§1).
- **Counters other than energy on a player as a cost.** No Commander-legal card pays experience, poison or rad. `payEnergyLocked` is written so that a `PlayerCounterCost{Name, N}` could replace it if one ever does.
- **Two-Headed Giant.** Energy is not shared in it, and the engine has no team variant.
- **A dedicated energy panel.** The seat marker is enough at this scale (§8).

---

## Questions for the owner

Each question lists the recommended option first. The sections above are written as if every recommendation were chosen.

1. **How paying {E} is modelled on an activated ability (§2; CR 107.14, 602.1a).**
   - **(a) Recommended:** a dedicated `AbilityCost.Energy` / `EnergyX` component and `ManaAbilityShape.EnergyCost`, paid by one helper that removes the counters from the activator, checked up front like life. It is small, it matches CR 107.14's "removes one energy counter from themselves" exactly, and there is nothing for the player to choose.
   - (b) Widen `RemoveCounters` with a player form. It reuses one struct, but most of that struct (`From`, `Among`, the kind choice, `CounterSourceIDs`) has to be switched off for the player case, and every reader must learn the new branch.
   - (c) Put {E} in the `Mana` cost string. It needs the least new plumbing, but every mana-cost reader (mana value, reductions, the auto-tapper, the strict gate) would have to skip a symbol that is not mana.

2. **Mana abilities with an energy cost and the auto-tapper (§5; CR 605.3a, 118.3c).**
   - **(a) Recommended:** an energy tier, used only when no energy-free plan exists, before the pain tier, shown in the auto-tap preview. Aether Hub pays colours on a strict table with no extra click. Energy is spent before life, and only when needed.
   - (b) An energy tier after the pain tier. Life is spent before energy, which keeps energy for the cards' sinks but costs life the player may need more.
   - (c) Never planned. The player activates these by hand, like restricted mana. Nothing is ever spent without a click, but a spell only the Hub's colour can pay is dimmed until the player floats the mana.

3. **Resolution payments of a chosen amount (§3; CR 118.12, 107.1b).** "Pay any amount of {E}", "one or more {E}".
   - **(a) Recommended:** a new `pay_amount` prompt with `min` and `max`, answered by a stepper in the dock. A 20-energy total is one control, and the prompt can later serve "pay any amount of life" or mana.
   - (b) Reuse `option_pick` with one option per amount from 0 to the total. No new prompt kind or wire shape, but the list is as long as the energy total, and the bot's move list grows with it.

4. **Energy under the permissive posture and Cast anyway (§4; ADR 0118 §2).**
   - **(a) Recommended:** energy is always paid, like life. Only mana is waived. CR 118.3 holds for every component the engine can count, and the table's energy stays true.
   - (b) Energy is waived with mana when the payment is permissive or forced. It is more lenient on a sandbox table, but the energy total then drifts from what was paid, and nothing in the log says so.

5. **How the heuristic values energy (§7; ADR 0126).**
   - **(a) Recommended:** a flat weight per counter (`Weights.Energy`, starting at 0.25 of a card), charged when spent, credited through an `energy` purpose when gained, and tuned under ADR 0126 §8. It is simple and measurable, and the bot uses a sink when the effect is worth more than the counters.
   - (b) Price energy by the best sink the bot controls, so energy is worth 0 with no sink and more with Aetherworks Marvel out. Play is sharper, but there is one more board read per decision, and it needs the sinks' purposes declared before it can work.
   - (c) Leave energy unpriced. The bot spends it greedily on any sink. It needs no new weight, but the bot will burn eight energy on a weak effect while a better use waits.

6. **What the delivery covers (Delivery; ADR 0106 decision 6, ADR 0109 decision 4).**
   - **(a) Recommended:** all five PRs, each landing every card its seam unblocks, including the about 14 cards that only get energy (they need nothing new, and PR 1 tests the counter they add). About 120 of the 142 cards land, and energy decks become playable end to end.
   - (b) PRs 1 to 3 only: activation costs, mana abilities and resolution payments, which is about 105 cards. Alternative costs, replicate and the paid-or-lost tally go on registry rows. It is a shorter sprint, and the seven alt-cost and keyword cards and the two tally readers wait.
   - (c) PR 1 only, with the two waiting cards and the fixed activations, about 30 cards. It closes #1995 quickly and leaves every resolution payer, which is most of the energy cards, on the backlog.
