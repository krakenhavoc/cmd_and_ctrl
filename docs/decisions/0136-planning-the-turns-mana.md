# ADR 0136 — Planning the turn's mana: rock first, draw first, two spells over one

**Status:** Accepted (owner answers 2026-10-08) · 2026-10-08 · S67 — Bots round 3: no stalls, sharper sequencing.
**Owner decisions:** the owner answered this ADR's ten questions on 2026-10-08, each with the recommended option. The answers are listed under [Owner answers](#owner-answers-2026-10-08) and are binding. The options not chosen are kept under [Questions for the owner (answered)](#questions-for-the-owner-answered).
**Issues:** [#2458](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2458) (this change: draw before deploying, rock first, two spells over one). It must not conflict with [#2668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2668) (hold instant-speed ramp for the end step before your turn); §5 says how the two fit. Under owner answer 6, #2668 is delivered by this ADR's PR 5.
**Owner direction:** 2026-10-08, on #2458: a design pass before any implementation.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-08. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote head: `origin/develop`, `origin/main`, `origin/cost-ledger`, `origin/docs/issue-audit`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/playmats`, `origin/fix/2545-breeches-flake`, `origin/fix/2611-marwyn-source-left`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default` and `pr/2326`. The highest number on any of them is 0135 (`0135-alternative-costs-that-tap-discard-awaken-and-emerge.md`, on `origin/develop`, `origin/main` and `origin/fix/2545-breeches-flake`). The one open pull request, #2700, adds no ADR. This ADR takes **0136**.
**Builds on:** [ADR 0126](0126-bots-that-play-their-decks.md) (the prices, §2's ramp premium, §5's leftover windows, §6's `purpose`, §8's measurement, and the [exit decision on A2](0126-bots-that-play-their-decks.md#exit-decisions-2026-10-07) that sends rock-first here), [ADR 0033](0033-ai-bot-seat.md) §3 (a policy reads the view and the move list, never the game), [ADR 0052](0052-bot-decision-harness-and-eval.md) (the arena and the position suite).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

**Why a new ADR rather than an ADR 0126 amendment.** ADR 0126 prices one move at a time and lists the turn plan under its [Out of scope](0126-bots-that-play-their-decks.md#out-of-scope) ("this mana lets me cast that ... is lookahead"). This ADR adds a new step to the decision, a search over sets of moves with a mana model, and it has its own measurement and its own config switch. None of ADR 0126's prices change. It is also being written while an ADR 0126 amendment for #2689 is in progress, and a separate file avoids two authors editing one.

---

## Owner answers (2026-10-08)

The owner chose (a), the recommended option, on every question. The sections below are written to these answers, and they bind the Delivery PRs.

1. **The enumerator stamps each cast move's total mana cost** as `legal.MoveCost.mana` (Q1, §2).
2. **The bot decides whether a set is payable** with its own mana model over the view (Q2, §2).
3. **A plan holds only casts whose costs are mana and the card**, in the bot's own main phase with an empty stack (Q3, §1).
4. **`PurposeView` gains `lands_untapped`**, declared on simic-ramp's Harrow, Nature's Lore and Three Visits (Q4, §2).
5. **Draw before deploying is an order only**, with no new weight (Q5, §4).
6. **The plan holds instants for the end step before the bot's turn, and replaces #2668's rule** (Q6, §5). #2668 is delivered by this ADR's PR 5: if #2668 has not landed by then, PR 5 is its implementation and closes it; if it has, PR 5 replaces its rule and keeps its two positions gated.
7. **The commander tax is weighed through the mana budget**, and the forward charge stays a flat `CommanderTax × 0.5` per cast (Q7, §6).
8. **The plan is measured against `heuristic-noplan`**, and `heuristic-baseline` stays the frozen pre-S66 policy (Q8, §8).
9. **The acceptance bar is P1–P7 as written** (Q9, §8).
10. **`cantrip-with-leftover-mana` is kept as it is**, and `cantrip-before-the-permanent` is added with Night's Whisper the only accepted move (Q10, §8).

## Context

### The problem

The heuristic decides one move at a time. In its main phase it takes the highest-priced move that clears the bar, then sees the next window after that move resolves. It never asks what the rest of the mana will buy. So it casts the single most valuable spell even when a cheaper spell first, or two spells, would use the same mana for more.

The owner's review game on 2026-10-08 (game `8a9f18d7`, 1v1: the heuristic on simic-ramp against a Claude MCP seat) shows it on five turns in a row. Every value below was re-ranked on `develop` at `a5f7d2639` from the windows in the game's decision log. The prices have not moved since the game.

| seq | Turn | Mana | Today's cast (value) | Then | Better, same mana | Value |
|---|---|---:|---|---|---|---:|
| 133 | T3 | 3 | Ornithopter of Paradise {2} (+1.96) | 1 mana unused | Arcane Signet {2} (+0.80), tap it, then Ornithopter | +2.76 |
| 226 | T5 | 4 | Uro, Titan of Nature's Wrath {1}{G}{U} (+7.71) | 1 unused | Signet, then Uro (5 of 5 mana) | +8.51 |
| 292 | T6 | 6 | Tatyova, Benthic Druid from the command zone, no tax {3}{G}{U} (+3.71) | 1 unused | Signet, then Tatyova (7 of 7) | +4.51 |
| 402 | T8 | 8 | Tatyova with {2} tax, 7 mana (+3.71) | 1 unused | Oracle of Mul Daya {3}{G} (+2.76) + Harmonize {2}{G}{G} (+2.40) (8 of 8) | +5.16 |
| 475 | T9 | 9 | Tatyova with {4} tax, 9 mana (+3.71) | 0 unused | Oracle + Harmonize (8 of 9) | +5.16 |

The result: Arcane Signet and Oracle of Mul Daya sat in hand from turn 1 to turn 10. Tatyova died after the casts at 292 and 402 and was recast each time.

Seq 180 (T4, 4 mana) is in the issue too. The bot cast Explosive Vegetation (+2.80) into open blue mana, where Signet and Rampant Growth (+0.80 each, +1.60) give the same two mana of ramp. A plan does not change that choice, and it should not: Explosive Vegetation is one card for the ramp that the pair spends two cards on. The counterspell risk is the real argument for the pair, and it is out of scope here.

The S66 label review found the same shape in a sorcery-speed hand (#2458's opening post, position `cantrip-with-leftover-mana`). The bot is in its second main phase with five Swamps, Ashnod's Altar, Bastion of Remembrance, Vampiric Tutor, Exquisite Blood and Night's Whisper. The owner wants Night's Whisper first: "More cards hopefully gives a chance to draw into something and ability to double spell with altar or bastion". The position's view was frozen before ADR 0126 PR 6 declared purposes, so Night's Whisper prices at the floor (+0.10) there. With its declared purpose (`draws: 2`) it prices +1.20, against Exquisite Blood's +1.90 and Bastion's +1.50. Night's Whisper and Bastion together are +2.70 in the same five mana.

ADR 0126's exit decision on A2 (2026-10-07) sent rock-first here. In S66's final run, mana rocks and dorks were cast in 68% of the games in which they were offered with an open mana deficit, against a bar of 80%. In every such window the rock lost to a land drop or another spell, never to a pass. Casting the rock first and the spell after it is exactly the plan this ADR adds.

### Why: the code

- **One move at a time.** `decideGeneral` (`aiseat/heuristic/heuristic.go:789`) prices each move with `valueOf` and takes the best one over the bar. Nothing compares a set of moves.
- **Mana is free.** `valueOfCast` (`moves.go:480`) prices what the card does and what it costs in cards, life and sacrifices. The mana it spends is not priced, so a 7-mana cast and a 2-mana cast of the same value tie. That is right for one move and wrong for a turn.
- **The commander tax is never weighed against what else the mana buys.** A cast from the command zone is charged a flat `CommanderTax × 0.5` (`moves.go:603`) for the {2} it adds to the next cast (CR 903.8). The tax paid now is mana, and mana is free, so Tatyova prices +3.71 at no tax, {2} and {4} alike.
- **The ramp premium is per card.** `rampFor` (`moves.go:634`) prices each new source against the deficit on its own. Two sources in one turn would each claim the whole deficit.
- **The bot does not know how much mana it has.** `SeatEval.UntappedMana` (`score.go:306`) counts untapped mana permanents, not mana, and no code asks whether a second spell is still payable after the first.

### What the heuristic may read

ADR 0033 §3 holds. Everything the plan needs is on the view today except two facts, and §2 names those:

- The bot's untapped permanents, their `mana_abilities[]` (`produced`, `tap_cost`, `sacrifice_cost`, `exile_self`, `mana_cost` for a filter or a Signet, `color_options`, `adds_no_mana`, `condition_unmet`), `summoning_sick` and `abilities` (haste).
- `mana_pool` and `commander_casts` on the bot's `PlayerView`, and `mana_cost` and `purpose` on each card in hand.
- The move list: one cast move per payable way to cast each card, with its X, alternative cost and additional-cost choices in `Params`.

Missing: the total mana cost a cast move charges after the tax, cost reductions and increases (CR 601.2f), and whether the lands a ramp spell puts onto the battlefield enter untapped.

### What this ADR does not change

ADR 0126's prices, thresholds and windows. Combat. Every window that is not the bot's own main phase with an empty stack. The model tiers' prompts.

---

## Options considered

1. **A targeted rock-first rule.** If a castable mana rock that can be tapped this turn is in hand, and the top move is still payable after casting it, cast the rock first. Small, and it fixes seq 133, 226 and 292. It does nothing for seq 402 and 475 (two spells over one), the cantrip position, or the tax.
2. **A pairwise lookahead.** Compare the top move with the best feasible pair. It fixes all five windows in the table, but a pair is an arbitrary limit. Three spells in a turn are common in Commander, and seq 475's alternatives include a three-card set.
3. **A turn plan (recommended).** A bounded search over sets of the sorcery-speed casts on offer, with a mana model built from the view. The set worth most is the plan; the bot makes the plan's first move by a fixed order (mana first, then draws, then the rest). It covers every case above, prices the tax paid now by what else the mana would buy, and gives #2668 a natural home (§5). It costs a mana model, which is the bulk of the work.
4. **Ask the engine.** The enumerator stamps on each cast move which other cast moves are still payable after it, using its own `canPayExcluding`. That check is exact for pairs and knows the engine's payment rules, but it cannot see mana that a rock cast first would add, it is pairwise only, and it costs O(n²) payment checks on every enumeration for every seat. It is kept as a fallback in [Q2](#questions-for-the-owner-answered).
5. **Simulate the turn on a cloned game.** Ruled out by ADR 0033 §3 and listed as rejected in ADR 0126.

---

## Decision

### 1. When the plan runs

In the bot's own main phase with an empty stack (`state.sorcerySpeed`), first or second. Nowhere else. The land drop is unchanged: it is priced at `LandValue` (8.00), above any plan, so the bot plays its land first and then plans the rest.

**Members.** A plan's members are cast moves whose only costs are mana and the card itself: `costsOnlyManaAndTaps` (`windows.go`) without the tap part, and no sacrifice, discard or other card cost in `Params`. A move with a non-mana cost still competes, as a plan of one. A card with several cast moves (one per X, per alternative cost, per land Harrow could sacrifice) contributes at most one of them to a plan. Only moves the bot is offered in this window are considered, so a plan never names a cast the engine has not already found legal on its own.

**Size.** The candidates are the `PlanMaxCards` (10) cards whose best single move prices highest. That is at most 1,024 sets. The search prunes any set whose total mana exceeds what the bot can make, and it stops at the context deadline with the best set so far, like `decideGeneral` (ADR 0033 §10).

### 2. The mana model

A small, pure function under `aiseat/heuristic`, reading only the view:

- **What the bot has.** Its mana pool, plus one entry per untapped permanent it controls with a repeatable mana ability it can activate now. "Repeatable" is `repeatableMana`'s test (`score.go:768`): a tap ability with no sacrifice or exile cost. "Now" excludes a creature that is summoning sick and lacks haste (CR 302.6), an ability with `adds_no_mana`, `condition_unmet` or a `cant_activate` reason. Each entry is the set of symbols it can make, from `produced`, or from `color_options` when `produced` is empty (Exotic Orchard). An entry the view cannot size counts as one mana of no colour. A filter ability (a Signet, Flooded Grove) is an entry that consumes one generic mana and makes its output. A mana ability with a life cost or a damage clause (Mana Confluence, City of Brass) counts as a source; the plan does not price that life, just as a single cast does not today.
- **What each move costs.** The total mana cost the move charges (CR 601.2f): printed cost, X, the commander tax and any reduction or increase. This ADR asks the enumerator to stamp it on the move ([Q1](#questions-for-the-owner-answered)), because it is computed there already.
- **What each member adds.** A mana rock or a hasty mana creature, once it has resolved, adds its entry for the members after it. A ramp spell's lands add mana this turn only if they enter untapped, which the view does not say today ([Q4](#questions-for-the-owner-answered)).
- **Feasible.** A set is feasible if, in the plan's order (§4), each member's cost can be paid from what is left, coloured symbols first and then generic. A hybrid symbol may be paid by either colour. A Phyrexian symbol is paid with mana here; its life is already priced by `valueOf`. A set of one is always feasible: the engine offered the move, and its answer outranks the model's.

The model is advisory. If it is wrong, the second member is not offered in the next window, and the bot re-plans with what it has. The worst case is the one-spell turn it plays today. PR 2's arena counts these misses (§8).

### 3. What a plan is worth

```
value(S) = Σ valueOf(m) for m in S
         − Σ rampPremium(m)                 (each member's own premium out)
         + RampPerMana × min(Σ amount(m), deficit)   (one shared premium back)
```

`valueOf` is ADR 0126's price, unchanged. The ramp premium is recomputed once for the set, against the deficit measured over the cards **not** in the plan, so two rocks in one turn do not both claim the whole gap, and a rock whose only purpose was to reach a spell the plan already casts closes nothing.

**The plan** is the feasible set with the highest value. Ties go to the smaller set, so the bot keeps cards in hand rather than spending them for nothing, and then to today's order. Taking the best set means every member adds value: a rock priced below zero (Arcane Signet at seq 402, −0.20) is in the plan only when the set is worth more with it than without it.

**The decision.** When the plan holds more than one member and is worth more than the best single move, the bot makes the plan's first move (§4) if the plan's value clears the window's bar (`PassThreshold`, or `LeftoverThreshold` in ADR 0126 §5's windows). Otherwise `decideGeneral` decides exactly as today. A plan of one is today's choice.

The plan is not stored. It is rebuilt in every window from the view, so it adds nothing to the little the policy keeps between decisions (the aggression rotation and the concede counter). After the Signet resolves, the next window plans again with the Signet on the battlefield and finds the Ornithopter. If a draw finds something better, the next plan uses it.

### 4. The order: mana first, then draws, then the rest

The plan says which spells; the order says which one now:

1. **Mana first.** A member that adds mana this turn (a rock, a hasty dork, a ramp spell whose lands enter untapped) goes first, so its mana is there for the rest. This is rock-first.
2. **Then draws.** A member whose purpose draws or tutors goes next, cheapest first. This is draw before deploying: the card it finds is in hand when the next window plans the rest, and may replace part of it. That is the owner's reason for Night's Whisper first.
3. **Then the rest**, highest value first.

The order changes no price. It only picks which member of an already chosen plan is cast first ([Q5](#questions-for-the-owner-answered) offers a draw premium as well).

### 5. Instants: cast now, or hold for the end step (#2668)

ADR 0126 §5 already holds an instant-speed move out of the bot's second main phase, because the mana it leaves untapped is still there in the end step before the bot's turn (`leftoverEligible`, `windows.go:144`). #2668 asks for the same thing for instant-speed ramp in any main phase: Harrow belongs in the end step before the bot's turn, where the mana it spends would otherwise go unused, and today it is rarely still castable there because a bigger main-phase cast has tapped the bot out.

In a plan, an instant-speed member is **held** unless one of these holds:

- it adds mana that a later member of the plan uses (§4 item 1), or
- it draws or tutors and the plan has mana left after it, so the card it finds can still be cast this turn.

A held member stays in the plan. Its mana is reserved: no other member may use it, and it counts in the plan's value. It is not cast now. If every remaining member is held, the bot passes, and ADR 0126 §5 casts the held spell in the end step before its turn as it does today.

So #2668 becomes one case of the plan, with no second rule beside it. In #2668's main-phase case, Harrow against a bigger spell that taps the bot out, the plan compares {the big spell} with {a smaller spell, Harrow held}, both on the same mana. Whichever of #2668 and this ADR's PR 5 lands first, the other keeps #2668's two gated positions, `harrow-at-the-end-step-before-yours` and `develop-before-harrow`, passing. If #2668 ships first as a standalone rule, PR 5 replaces it with this one and keeps the positions. Under owner answer 6, #2668 is delivered by PR 5.

### 6. The commander tax

The tax paid now is mana, and the plan prices mana as a budget, so it is weighed against what else that mana buys. At seq 402 and 475, Tatyova's 7 and 9 mana lose to Oracle and Harmonize because the same mana buys more, which is the trade a player makes.

The forward-looking half, `CommanderTax × 0.5` per cast from the command zone (`moves.go:603`), stays flat. Each cast adds the same {2} to the next one (CR 903.8), whatever the tax is now, so a flat charge is the faithful one. Whether to recast a commander that keeps dying to the same removal is a threat question, and it is out of scope.

### 7. What the bot reports

`Rank` is unchanged: each move keeps its own price, so the model tiers' prompts and the decision log's candidate list read as today. The decision's reason names the plan, for example `plan: Arcane Signet → Ornithopter of Paradise (+2.76)`. The decision log's trace gains an additive `plan` field listing the members in order, held members marked. Decision log records are JSON lines read by `boteval`, and an added field is ignored by older readers.

### 8. Config, and the bar it must clear

**Config.** `PlanTurnMana bool` and `PlanMaxCards int` on `heuristic.Config`, on in `DefaultConfig()`, off in `BaselineConfig()`. The zero value is today's policy. An arena contestant `heuristic-noplan` is `DefaultConfig()` with `PlanTurnMana` off, so a run can measure the plan alone ([Q8](#questions-for-the-owner-answered)). `heuristic-baseline` stays the frozen pre-S66 policy (ADR 0126 owner decision 5): this ADR changes no price.

**The runs** are ADR 0126 §8's, with `heuristic-noplan` in place of `heuristic-baseline` in run 2:

1. **Run 1, the four-deck rotation** on today's curated decks: `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks <the four> --games 64 --rotate --seed 1`.
2. **Run 2, plan against no plan:** two `heuristic` and two `heuristic-noplan`, 48 games each way round the decks (96 games).
3. **#2668's run:** simic-ramp ×4, `--rotate --lockstep`, seeds 1–40 and 1000–1119, for Harrow's use.
4. **The suite** and **the nightly gates**, as ADR 0126 §8 items 3 and 4.

**New arena numbers** (PR 2, before any behaviour change), per contestant and deck:

- **Stranded mana:** in each of the bot's own turns, the mana it could still make at its last pass in its last main phase, and whether a cast was on offer then. Reported as the share of turns that ended with 2 or more mana unspent and a cast still on offer.
- **Plan misses:** windows in which the previous window's plan named a next member that is not offered now, though no other seat acted in between.

**Acceptance** ([Q9](#questions-for-the-owner-answered)):

| # | Measure | Bar |
|---|---|---|
| P1 | Rocks and dorks offered with an open deficit (ADR 0126's A2, as PR 9 measured it; 68% at S66's exit) | used in **≥ 80%** of the games in which each was offered (run 1) |
| P2 | Turns ending with ≥ 2 mana unspent and a cast on offer (run 1) | **at most half** of `heuristic-noplan`'s share |
| P3 | Plan against no plan (run 2) | not detectably worse: the `heuristic` interval's upper bound is not below 25%. The PR reports whether its lower bound clears 25% |
| P4 | Harrow (run 3) | use rises against `develop` on the same seeds, with no win-rate regression (#2668's acceptance) |
| P5 | Suite | every gated position passes, including #2668's two and this ADR's new ones; no tag's agreement falls |
| P6 | Health (runs 1–3) | 0 stalls; rejected moves not above `heuristic-noplan`'s; turns p50 within ±3; plan misses under 5% of planned windows |
| P7 | Speed | the heuristic's decision p99 within 5 ms of `heuristic-noplan`'s |

**Positions** (proposals; the owner reviews every label, and none is gated before that):

- `rock-before-the-two-drop` (seq 133): T3, three lands untapped (Island, Field of the Dead, Exotic Orchard), Ornithopter of Paradise and Arcane Signet in hand. Accept "Cast Arcane Signet". Reject "Cast Ornithopter of Paradise".
- `signet-then-uro` (seq 226): T5, four lands including a Forest, Uro and Arcane Signet in hand. Accept "Cast Arcane Signet". Reject "Cast Uro, Titan of Nature's Wrath".
- `signet-then-the-commander` (seq 292): T6, six mana, Tatyova in the command zone with no tax. Accept "Cast Arcane Signet". Reject "Cast Tatyova, Benthic Druid from the command zone".
- `two-spells-over-the-taxed-commander` (seq 402): T8, eight mana, Tatyova with {2} tax, Oracle of Mul Daya and Harmonize in hand. Accept "Cast Oracle of Mul Daya" and "Cast Harmonize". Reject "Cast Tatyova, Benthic Druid from the command zone".
- `cantrip-before-the-permanent`: the `cantrip-with-leftover-mana` board, re-captured from a game on a view that carries the declared purposes. Accept "Cast Night's Whisper" only, as the owner asked on #2458 ([Q10](#questions-for-the-owner-answered)).

---

## Worked: the evidence windows under the plan

Prices are today's (re-ranked on `a5f7d2639`). Mana counts use the model in §2.

| Window | Best single move | The plan, in order | Plan value | First move |
|---|---:|---|---:|---|
| seq 133 | Ornithopter +1.96 | Arcane Signet, Ornithopter (3 lands + the Signet's 1 = 4 mana, 4 spent) | +2.76 | Arcane Signet |
| seq 180 | Explosive Vegetation +2.80 | Explosive Vegetation (the best pair, Signet + Rampant Growth, is +1.60) | +2.80 | unchanged |
| seq 226 | Uro +7.71 | Signet, Uro (5 of 5) | +8.51 | Arcane Signet |
| seq 292 | Tatyova +3.71 | Signet, Tatyova (7 of 7) | +4.51 | Arcane Signet |
| seq 402 | Tatyova +3.71 | Oracle, Harmonize (8 of 8). Adding the Signet makes it 10 of 9 | +5.16 | Oracle of Mul Daya |
| seq 475 | Tatyova +3.71 | Oracle, Harmonize (8 of 9). With the Signet it is 10 of 10 but worth +4.96, so the Signet is left out | +5.16 | Oracle of Mul Daya |
| cantrip position, with today's purposes | Exquisite Blood +1.90 | Night's Whisper, Bastion of Remembrance (5 of 5) | +2.70 | Night's Whisper (draws first) |
| `develop-before-harrow` (gated) | Peregrine Drake +3.70 | Peregrine Drake. Harrow, then Eureka Moment on the lands it nets, is +2.00 | +3.70 | unchanged |

In each window where today's choice changes, the first move is one the issue or the owner named. The two windows the plan should leave alone are left alone. The ramp premium in these rows is the shared one from §3; at seq 133 and 226 the deficit (to Avenger of Zendikar's 7) is large enough that sharing it changes nothing.

---

## Snapshot and wire impact

- **Snapshot:** none. A plan is computed from the view in each window and never stored. `SnapshotSchemaVersion` stays at 7 and no shape file changes.
- **The move list** (owner answer 1): `legal.MoveCost` gains an additive `mana` string on cast moves, the total mana cost the move charges (CR 601.2f), such as `"{5}{G}{U}"` for Tatyova with {2} tax. Absent on every other kind. Clients ignore it. `docs/protocol.md` documents it.
- **`PurposeView`** (owner answer 4): an additive `lands_untapped` count, declared on the curated ramp spells whose lands enter untapped (simic-ramp's Harrow, Nature's Lore and Three Visits), and checked by `TestCuratedDeckPurposes`. Catalog data projected into the view, like every purpose; no snapshot change.
- **The decision log:** an additive `plan` field in the trace (§7).
- **The arena report:** additive columns (§8).

---

## Delivery

One lever per PR, as ADR 0052 asks. Every PR is in S67 (or the bot sprint that follows it), targets `develop`, and carries ADR 0052's report block. PRs 4 and 5 change the bot's decisions, so they run the nightly E2E on their branch.

| PR | What | Acceptance | Needs |
|---|---|---|---|
| 1 | **This ADR.** Docs only. | docsguard | — |
| 2 | **Measurement.** The arena's stranded-mana and plan-miss numbers, the `heuristic-noplan` contestant (equal to `heuristic` until PR 4), and runs 1 and 3 on `develop` recorded here as the baseline. No behaviour change. | Unit tests for the two tallies on hand-built observer feeds; the baseline rows in [Measurements](#measurements) | 1 |
| 3 | **The two wire facts.** The cast move's total mana cost (Q1) and `lands_untapped` with its curated declarations (Q4), `docs/protocol.md`. Touches `legal`, `protocol`, `effects` and `decks`, and nothing under `aiseat/`. | Enumerator tests: the stamped cost for a taxed commander, an X spell, a cost reduction and an alternative cost; `TestCuratedDeckPurposes` | 1 |
| 4 | **The plan** (§1–§4, §6, §7): the mana model, the set search, the shared ramp premium, the order, the reason and the trace, `PlanTurnMana` on. Instants are not held yet. | Unit tests for the mana model over the six evidence views (feasibility, filter lands, a sick dork, an Orchard); runs 1 and 2 against P1–P3, P5–P7; the four seq positions and the cantrip position proposed | 2, 3 |
| 5 | **Held instants** (§5). Delivers #2668 (owner answer 6). | Run 3 against P4; #2668's two positions stay gated; runs 1 and 2 hold P3 and P5–P7 | 4 |
| 6 | **Exit.** A "How the heuristic plans a turn" section in `docs/bot.md`, the Measurements rows, and the evidence on #2458. | — | 4, 5 |

PRs 2 and 3 can be built in parallel. The owner's answers leave this plan as it was drafted. PR 5 also delivers #2668 (owner answer 6): its `## Issues` section closes #2668 if that issue is still open, and otherwise says which rule it replaced.

---

## Measurements

PRs 4 and 5 add their rows under PR 2's baseline.

### PR 2: the baseline, before any plan (2026-10-08)

On this branch, cut from `develop` at `e2b731dd3`. All runs use `--rotate --lockstep` and the real dump. No plan is made yet, so `planned windows`, `checked` and `plan misses` are 0 in every run, and `heuristic-noplan` plays as `heuristic`. **0 stalls in every run.**

**Run 1:** `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --lockstep --seed 1`. 64 games, turns p50 15.

| Deck | Won | Own turns | Stranded | Stranded % | Idle | Mean unspent |
|---|---:|---:|---:|---:|---:|---:|
| esper-control | 27 / 64 | 869 | 136 | 15.7% | 306 | 1.73 |
| izzet-aggro | 2 / 64 | 755 | 39 | 5.2% | 143 | 1.01 |
| mono-black-aristocrats | 24 / 64 | 872 | 81 | 9.3% | 185 | 1.06 |
| simic-ramp | 11 / 64 | 794 | 47 | 5.9% | 149 | 0.91 |
| **all** | | **3,290** | **303** | **9.2%** | **783** | **1.19** |

P1's baseline, A2 in run 1: 7 of the 31 rock and dork rows reach 80% (Sol Ring on all four decks, Worn Powerstone, Birds of Paradise and Delighted Halfling). Pooled, they were used in 267 of 628 seat-games in which they were offered (42.5%). The Signets, Talismans, Mind Stone, Arcane Signet and Commander's Sphere sit at 0% to 38%. Decision p99 is 507 µs.

**Run 2:** two `heuristic` and two `heuristic-noplan`, `--games 48 --rotate --lockstep --seed 1` twice: half A seats `heuristic` on esper-control and izzet-aggro and `heuristic-noplan` on mono-black-aristocrats and simic-ramp, and half B swaps them. 96 games, turns p50 15 in both halves.

| Contestant | Won (pooled) | 95% CI | Own turns | Stranded | Stranded % | Mean unspent | Decision p99, half A / B |
|---|---:|---|---:|---:|---:|---:|---|
| heuristic | 49 / 192, 25.5% | 19.9%–32.1% | 2,484 | 240 | 9.7% | 1.20 | 456 µs / 554 µs |
| heuristic-noplan | 47 / 192, 24.5% | 18.9%–31.0% | 2,480 | 238 | 9.6% | 1.18 | 457 µs / 458 µs |

The two contestants are the same policy here, so the gap is the deck split, which the swap cancels: by deck the two halves agree within one game and a few turns (esper 115 stranded of 670 turns in both halves).

**Run 3:** simic-ramp ×4, `--games 40 --seed 1` and `--games 120 --seed 1000`. 160 games, turns p50 12. Harrow used in 76 of 218 seat-games offered (34.9%). Stranded: 293 of 6,563 own turns (4.5%), mean unspent 0.76.

**Unchanged decisions.** `boteval suite run --policy heuristic` gives 41 of 41 (100%) on this branch and on `develop`, with the same move, layer and reason string at every position. Run 1 on `develop` at `e2b731dd3` gives the same winner and the same turn count in all 64 games. Its Cards table differs from this branch's by one window at a time on a few cards (Day of Judgment against Wrath of God, Sheoldred, Commander's Sphere). A second `develop` run differs from the first in the same way (Damnation against Day of Judgment), so this is run-to-run noise in how lockstep games break ties between equally priced cards, not this PR.

---

## Out of scope

- **Counterspell risk** (seq 180): one big spell into open blue mana against two smaller ones. The plan makes it possible to weigh later; this ADR does not.
- **Rituals and one-shot mana** (Dark Ritual, Lotus Petal, Treasures). They are not repeatable sources, so the mana model leaves them out, as ADR 0126 §2 does.
- **Activated abilities as plan members** (Konrad's mill, equip, a loot). They keep their single-move prices ([Q3](#questions-for-the-owner-answered)).
- **Casting a draw spell before the land drop** when a land is in hand. The land is still played first.
- **Recasting a commander that keeps dying.** A threat question, not a mana one.
- **The model tiers.** They inherit the plan through Layer B. Showing the plan in their prompts is a later change.

## Consequences

- Rocks are cast on the turn they are drawn when they help the curve, and the bot spends its mana on two or three spells when that is worth more than one. The ramp deck's dead cards (Signet, Oracle) should start being played.
- The bot weighs the commander's tax against the rest of its hand.
- #2668's rule lives inside the plan rather than beside it.
- The heuristic gains its first model of its own mana. It is an approximation of the engine's auto-tapper, and the plan-miss counter is what says how good the approximation is.
- Each main-phase decision searches up to 1,024 sets. P7 bounds the cost.

---

## Questions for the owner (answered)

Each question lists the recommended option first. The recommendation is the most CR-faithful option where that applies. The owner chose the recommended option for all ten on 2026-10-08 ([Owner answers](#owner-answers-2026-10-08)). The questions are kept with the options not chosen.

1. **Where a cast's mana cost comes from (§2; CR 601.2f, 903.8).**
   - **(a) Recommended:** the enumerator stamps the total mana cost each cast move charges, as an additive `mana` string on `legal.MoveCost`. It already computes this when it checks the move is payable, so the plan sees the same cost the engine charges: the commander tax, X, cost reductions and increases (Thalia, a Medallion), and an alternative cost.
   - (b) The bot computes it from the card's printed `mana_cost`, plus {2} per entry in `commander_casts`, plus X. No wire change, but every cost reduction and increase is invisible to the plan, so it plans casts that do not fit or misses ones that do.

   **Answered: (a), as recommended (owner answer 1).**

2. **Who decides whether a set is payable (§2).**
   - **(a) Recommended:** a mana model in the bot, from the view, as §2 describes. It can see a rock's mana arriving mid-turn, and a wrong answer costs only a plan miss, which the arena counts.
   - (b) The enumerator answers it: each cast move lists the other cast moves still payable after it, from the engine's own payment check. Exact for pairs, but pairwise only, blind to a rock's new mana, and O(n²) payment checks on every enumeration for every seat.
   - (c) (a) now, and (b) added for pairs only if run 1's plan misses exceed 5% of planned windows.

   **Answered: (a), as recommended (owner answer 2).**

3. **What can be in a plan (§1).**
   - **(a) Recommended:** casts whose only costs are mana and the card, in the bot's own main phase with an empty stack. Anything else competes as a plan of one, priced as today.
   - (b) Also activated abilities whose only cost is mana (Konrad's mill, an equip). The plan then decides between a sink and a spell, but the flat `ActivateBase` (ADR 0126 owner decision 6) would compete as if it were a real price.
   - (c) Also casts with sacrifice or discard costs. Two members could then want the same card or creature, and the search has to track that.

   **Answered: (a), as recommended (owner answer 3).**

4. **Lands a ramp spell puts onto the battlefield (§2, §4).**
   - **(a) Recommended:** add `lands_untapped` to `PurposeView` and declare it on the curated ramp spells whose lands enter untapped (simic-ramp's Harrow, Nature's Lore and Three Visits). The plan can then cast Harrow first and spend its lands, as a player would.
   - (b) Treat every such land as entering tapped. No wire change, and the plan never counts their mana this turn, so it under-plans Harrow's turns.

   **Answered: (a), as recommended (owner answer 4).**

5. **Draw before deploying (§4).**
   - **(a) Recommended:** an order only. Within the chosen plan, draw and tutor members come after mana and before everything else. No new weight, so nothing that does not already fit the plan is cast for its draw.
   - (b) The order, plus a `DrawOption` premium for each card drawn while the plan still has mana left after the draw: the chance the card is castable this turn. It would let a cantrip beat a bigger spell that the plan prefers on price, but it is a new weight to tune.
   - (c) No draw rule: highest value first. Night's Whisper is cast second, after Bastion, and the card it draws comes too late to change the turn.

   **Answered: (a), as recommended (owner answer 5).**

6. **Instants and #2668 (§5).**
   - **(a) Recommended:** the plan holds an instant-speed member for the end step before the bot's turn unless its mana or its draw is used this turn, and this replaces #2668's rule (or is #2668's implementation if it has not landed).
   - (b) The plan casts instants like sorceries, and #2668 stays a separate rule. Two rules then decide the same Harrow window, and the second to land must be checked against the first.

   **Answered: (a), as recommended (owner answer 6).**

7. **The commander tax (§6; CR 903.8).**
   - **(a) Recommended:** the plan's mana budget weighs the tax paid now, and the forward charge stays a flat `CommanderTax × 0.5` per cast, because each cast adds the same {2} to the next.
   - (b) Also scale the forward charge with the casts already made. It would make the bot recast its commander less as the game goes on, but the {2} each cast adds does not grow.
   - (c) Price each mana spent at a small weight in single-move ranking too, outside the plan. It changes every existing position at once.

   **Answered: (a), as recommended (owner answer 7).**

8. **The contestant the plan is measured against (§8).**
   - **(a) Recommended:** a new arena name, `heuristic-noplan` (today's `DefaultConfig()` with the plan off). `heuristic-baseline` stays the frozen pre-S66 policy, as owner decision 5 of ADR 0126 says, because this ADR changes no price.
   - (b) Re-freeze `heuristic-baseline` at today's `DefaultConfig()`. One name fewer, but the S66 reference is lost.

   **Answered: (a), as recommended (owner answer 8).**

9. **The acceptance bar (§8).**
   - **(a) Recommended:** P1–P7 as written: A2's original 80%, stranded mana halved, run 2 not detectably worse, and the health and speed bounds.
   - (b) Also require run 2's lower bound above 25%: the plan must beat `heuristic-noplan` outright. A sequencing change may be worth too little per game to show in 96 games, so this bar could fail on a change that is right.

   **Answered: (a), as recommended (owner answer 9).**

10. **The cantrip position (§8; #2458's opening post).**
    - **(a) Recommended:** keep `cantrip-with-leftover-mana` as it is (every cast accepted, pass rejected), and add `cantrip-before-the-permanent`, re-captured on a view that carries today's purposes, with Night's Whisper the only accepted move.
    - (b) Narrow `cantrip-with-leftover-mana` itself to Night's Whisper. Its frozen view has no purpose on Night's Whisper (+0.10), so the plan picks Exquisite Blood there, and it passes only with Q5 (b)'s premium.

   **Answered: (a), as recommended (owner answer 10).**
