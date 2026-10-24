# ADR 0126 — Bots that play their decks: the heuristic prices what a card does

**Status:** Accepted · 2026-10-06 · S66 — Bots that play their decks. The owner accepted it on 2026-10-06 and answered its eight open questions, all as recommended. The answers are recorded under [Owner decisions](#owner-decisions-2026-10-06). S66 closed on 2026-10-07 with the owner's [exit decisions](#exit-decisions-2026-10-07) on the acceptance bar.
**Issues:** [#2435](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2435) (this change). [#2436](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2436), the curated deck rebalance, waits on it. [#2437](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2437), a fifth curated deck, comes after both.
**Owner direction:** 2026-10-06, on #2435: fix the pricing before the rebalance, write an ADR before changing any weight, and measure it with [ADR 0052](0052-bot-decision-harness-and-eval.md)'s arena report on the curated decks, with the nightly gates green.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-06. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote head: 37 of them (`origin/develop`, `origin/main`, `pr/2326`, and 34 chore, docs, feat, fix, repro and wip branches). The highest number on any of them is 0125, on `origin/develop`, `origin/main` and `origin/feat/table-defaults-row-overlay`. This ADR takes **0126**.
**Amendments:** 2026-10-06, [discard payoffs](#amendment-2026-10-06-discard-payoffs) (accepted). 2026-10-08, [purposes that follow a mode's target, and damage priced by whether it kills](#amendment-2026-10-08-purposes-that-follow-a-modes-target-and-damage-priced-by-whether-it-kills) ([#2689](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2689)): accepted (owner answers 2026-10-08). 2026-10-08, [what removal hands back](#amendment-2026-10-08-what-removal-hands-back) ([#2679](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2679)): accepted (owner decision on #2753). 2026-10-12, [the race sends the swing that ends the game soonest](#amendment-2026-10-12-the-race-sends-the-swing-that-ends-the-game-soonest) ([#2400](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2400)): accepted (owner decisions 2026-10-12). 2026-10-12, [cycling cards as cheaper discards, and the bot's own draw step](#amendment-2026-10-12-cycling-cards-as-cheaper-discards-and-the-bots-own-draw-step) ([#2457](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2457)): accepted (owner answers 2026-10-12). 2026-10-12, [the attrition horizon trades commanders](#amendment-2026-10-12-the-attrition-horizon-trades-commanders) ([#2474](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2474)): accepted (owner decisions 2026-10-12). 2026-10-12, [Izzet and Simic after the rebalance](#amendment-2026-10-12-izzet-and-simic-after-the-rebalance) ([#2480](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2480)): accepted (owner decisions 2026-10-12). 2026-10-12, [blocks that survive as a set](#amendment-2026-10-12-blocks-that-survive-as-a-set) ([#3066](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3066)): accepted (owner decisions 2026-10-12). 2026-10-12, [landfall payoffs](#amendment-2026-10-12-landfall-payoffs) ([#3072](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3072)): accepted (owner decisions 2026-10-13). 2026-10-13, [izzet-aggro rebuilt around Purphoros](#amendment-2026-10-13-izzet-aggro-rebuilt-around-purphoros) ([#3071](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3071)): accepted (owner decisions 2026-10-13). 2026-10-14, [landfall doublers, payoff casts and fetch lands](#amendment-2026-10-14-landfall-doublers-payoff-casts-and-fetch-lands) ([#3146](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3146)): accepted (owner, 2026-10-15).
**Purpose gains a field:** [ADR 0142](0142-declared-answers-on-catalog-abilities.md), 2026-10-09 — `Purpose.Answers`, what an activated row can do in response, on the wire as `purpose.answers`. It is not an amount, so a purpose that declares only answers is not priced (`PurposeView.Priced`).
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) (the seat, §3's type gate, §5's funnel), [ADR 0052](0052-bot-decision-harness-and-eval.md) (the arena, the position suite, the report block every bot PR carries), [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) §1 decision 8 (catalog-declared `purpose` on an activated row, read by the bot), [ADR 0037](0037-unimplemented-card-signal.md) (the `unimplemented` mark).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

### The problem

A precon evaluation on 2026-10-06 played 52 four-seat heuristic games on the four curated decks with `boteval arena` and the decision log on, against `develop` at `b7702518`. The heuristic never plays about a third of each deck.

| Deck | Non-land cards never cast or used | Of non-land cards |
|---|---:|---:|
| esper-control | 21 | 62 |
| izzet-aggro | 19 | 63 |
| mono-black-aristocrats | 30 | 64 |
| simic-ramp | 22 | 61 |

"Never" means offered in at least five decision windows and taken in none. Some of the cards, out of the windows they were offered in:

| Card | Used |
|---|---:|
| Sol Ring | 0 of 49 |
| Rhystic Study | 0 of 52 |
| Mary Read and Anne Bonny's loot | 0 of 171 |
| Entomb | 0 of 399 |
| Harrow | 0 of 279 |
| Viscera Seer (cast) | 0 of 88 |
| Unexpected Windfall | 0 of 141 |
| Deadly Dispute | 0 of 129 |

Every mana rock in all four decks was never cast: Sol Ring, Arcane Signet, the Talismans, the Signets, Mind Stone, Commander's Sphere, Thought Vessel, Worn Powerstone, Hedron Archive. So were Llanowar and Fyndhorn Elves, every enchantment engine (Rhystic Study, Phyrexian Arena, Impact Tremors, Bastion of Remembrance, Sanguine Bond, Dictate of Erebos, Elemental Bond, Guardian Project, Beastmaster Ascension), both Altars, and most of the cheap ramp, tutor and draw spells (Harrow, Farseek, Nature's Lore, Three Visits, Entomb, Demonic and Vampiric Tutor, Night's Whisper, Preordain).

The win rates are badly unbalanced as a result:

| Deck | Win rate, 52 games | Wilson 95% interval |
|---|---:|---|
| mono-black-aristocrats | 63% | 49.9%–75.2% |
| simic-ramp | 27% | 16.8%–40.3% |
| izzet-aggro | 6% | 2.0%–15.6% |
| esper-control | 4% | 1.1%–13.0% |

Black leads because the heuristic activates Syr Konrad's `{1}{B}` mill whenever it has spare mana. A non-tap activated ability is priced at a flat +0.50, so it is the one mana sink the bot always takes. The other decks' sinks are the cards above, which it never casts.

### Why: the pricing table

A cast is priced as what the card is worth once it resolves, minus `Weights.Hand` (1.20) for the card leaving the hand (`moves.go` `valueOfCast`, `fuel.go` `resolvedValue`, `score.go` `permanentValue`). In its own main phase with an empty stack the bot acts when the best move scores above `PassThreshold` (0.25). Anywhere else it needs `InstantThreshold` (1.50).

| Card shape | Price today | Result |
|---|---|---|
| Non-creature permanent with no target (rocks, enchantment and artifact engines, Equipment, Vehicles, Altars) | `Permanent` 1.20, or `ManaSource` 1.00 for a mana source, minus 1.20 | ≤ 0: never cast |
| Creature whose power + 0.45 × toughness + keywords is below about 1.6 (1/1s, 0/x, mana elves) | 0.90 (summoning sick) × that, minus 1.20 | never cast |
| Untargeted instant or sorcery of mana value 2 or less (tutors, cantrips, Rampant Growth) | 0.60 × mana value, minus 1.20 | never cast |
| Spell with a discard cost | a further 1.20 per card discarded, whatever the card | never cast |
| Untargeted tap ability on a creature (Mary Read's loot) | `ActivateBase` 0.50 minus 0.30 for the tapped blocker = 0.20 | never used |
| Sacrifice-outlet ability | 0.50 minus the sacrificed creature's whole value | effectively never |
| Untargeted 4-mana sorcery (a board wipe) | 0.60 × 4 − 1.20 = +1.20, whatever the board | always cast, even onto the bot's own winning board |

The root cause is one equation. On the battlefield a non-creature permanent is worth exactly what a card in hand is worth (1.20 against 1.20), so casting it is a wash, and a small creature or a cheap spell is worth less than the card. The evaluation has no idea what any of them do.

### What the heuristic may read

ADR 0033 §3 holds. A policy under `aiseat/` reads `protocol.GameView`, the seat's own filtered projection, and the `[]legal.Move` list, and imports nothing from `internal/game` (`heuristic/imports_test.go`). It does not read oracle text. So every price here comes from what `CardView` already carries, or from a new wire field this ADR names (§6).

What `CardView` carries today, and this ADR uses:

- `type_line`, `mana_cost`, `power`, `toughness`, `abilities` (keywords), `restrictions`, `is_commander`, `is_token`, `unimplemented`, `summoning_sick`, `tapped`, `counters`.
- `mana_abilities[]` with `produced` ("{C}{C}", "{W|U|B|R|G}"), `tap_cost` and `sacrifice_cost`. These are stamped on a card in hand as well as on the battlefield, so the bot can tell a Sol Ring in hand makes two mana.
- `ability_rows[]` (#2219): the card's non-keyword triggered, static and activated abilities, by kind, on the battlefield and in hand. Absent for an uncatalogued card. Mana and keyword abilities are left out.
- `activated_abilities[]` with `tap_cost`, `sacrifice_self`, `mana_cost`, `purpose` (ADR 0106: `draws`, `controller_loses_life`, any-player rows only today).
- `additional_cost` (`discard_cards`, sacrifice options), `alternative_costs`, `modes`, `target_mode` on the cast surface.
- `stack_items[]` with `targets`, so the bot can see an opponent's spell pointed at its creature.

What it does not carry: what an instant or sorcery does. Wrath of God and Divination are both untargeted sorceries with a mana value, and nothing on the wire tells them apart. A spell's effect is an `OnResolve` closure in the catalog (`cards/effects/wrath_of_god.go`), so it cannot be derived from the definition either. §6 adds a small declared signal for exactly that.

### What measures it

ADR 0052 gives three instruments, and this ADR uses all three:

- **The arena** (`boteval arena`, `internal/botarena`): win rate with a Wilson 95% interval against the null rate. Today it groups by policy, and four heuristic seats are one row. The per-deck win rates above were computed by hand from `games.jsonl`.
- **The position suite** (`aiseat/suite`): 22 labelled positions, run on every CI run, gated ones fail `go test`. None covers a mana rock, an engine, a loot, a sacrifice or a wipe.
- **The decision log**: the "offered versus taken" table above came from a scratch script over it.

---

## Decisions

### 1. Measure before touching a weight

The first code PR changes no price. It makes the arena report what this ADR is judged on:

1. **A per-contestant Play table.** One row per contestant (policy and deck) as well as per policy, with the same columns: decided seat-games, wins, win %, Wilson interval, null.
2. **A Cards section.** Per deck, every non-land card the deck's seats were offered (a `cast` or `activate` move naming it), with the windows it was offered in, the times it was taken, and the games in which it was offered and the games in which it was used at least once. It is tallied from the runner's observer, the same feed the decision log writes, so it needs no decision log on disk. A card offered in five or more windows and never taken is flagged `never`. The section ends with each deck's `never` count.
3. **A `heuristic-baseline` contestant.** The policy with every new term in this ADR switched off (§9). It is an arena name only. It is never offered in the lobby or in `GET /bot/options`.
4. **The baseline numbers.** The four-deck rotation and the new-versus-baseline run (§8) on `develop` as it stands, recorded in this ADR's [Measurements](#measurements).

### 2. Mana sources: priced by the mana they make and by how short the bot is

Ranked first, because a deck that does not ramp casts nothing else on curve.

**On the battlefield** (`permanentValue`, so the evaluation and the cast agree): a mana source is worth `ManaSource` for its first mana plus `ManaPerExtra` for each further mana its best repeatable ability makes. "Repeatable" means a tap ability with no sacrifice cost (`tap_cost` and not `sacrifice_cost` or `exile_self`). A source's amount is the count of symbols in `produced`. A choice such as `{W|U|B|R|G}` is one symbol, and `{C}{C}` is two. Sol Ring is worth 2.0 where today it is 1.0. A land is unchanged. A creature with a mana ability keeps its creature value and adds nothing here, because `Evaluate` already counts it as a mana source.

**At cast time only**, a ramp premium: `RampPerMana × min(amount, deficit)`. The deficit is how far the bot is from casting what it holds:

```
want    = the largest mana value among the cards in hand and the commander
          (with its tax), capped at RampWantCap (7)
sources = Σ amount over the bot's own mana sources, tapped or not
deficit = max(0, want − sources)
```

The premium is what the owner asked for. It is large while the bot is short of mana for its hand and curve, and it falls to nothing as the game goes on, because sources grow and the deficit closes. No separate turn decay is needed. A Sol Ring is still cast late (2.0 − 1.2 = +0.8), as it should be. An Arcane Signet late, with nothing in hand it cannot cast, is not (1.0 − 1.2), unless §5's leftover-mana window takes it.

Worked, with the starting weights in §9:

| Card, situation | Today | Proposed |
|---|---:|---:|
| Sol Ring, turn 1, a 5-drop in hand | −0.20 | 2.0 + 2 × 1.0 − 1.2 = **+2.80** |
| Arcane Signet, turn 2, deficit 3 | −0.20 | 1.0 + 1.0 − 1.2 = **+0.80** |
| Arcane Signet, turn 9, deficit 0 | −0.20 | **−0.20** |
| Llanowar Elves, turn 1, deficit 3 | +0.11 | 1.31 + 1.0 − 1.2 = **+1.11** |

A ramp spell (Rampant Growth, Harrow, Farseek, Three Visits, Nature's Lore) gets the same premium for the lands it puts onto the battlefield, but only once §6's `purpose.lands` says how many. Until then it falls under §5's floor.

Rituals and one-shot sources (Dark Ritual, Lotus Petal, a Treasure) stay as they are. Their value is the spell they let the bot cast this turn, which needs a plan the heuristic does not make (see [Out of scope](#out-of-scope)).

#### Amendment 2026-10-09: an idle late rock

The owner's decision on ADR 0136 PR 4b's question 2. The table's "Arcane Signet, turn 9" row priced a rock with no open deficit at −0.20, below even §5's `LeftoverThreshold`, so the sentence above ("unless §5's leftover-mana window takes it") never came true. ADR 0136 PR 4b found that 170 of the 249 turns run 1 ended with stranded mana were exactly that rock, left in hand with its mana unspent.

In the turn's last main-phase window (§5's first window: the bot's own last main phase, stack empty), a cast of a nonland permanent with a repeatable mana ability, no ramp premium (the deficit is closed) and only mana and taps for its cost is priced at `LeftoverThreshold` when its mana would otherwise go unused, that is, when no other move on offer is priced above `LeftoverThreshold`. The bot then casts it rather than pass. A rock's mana is still useful past `RampWantCap`: it pays a commander's growing tax and an X. If several such rocks are on offer, the one that makes the most mana goes first. The first main phase is unchanged, because combat and the draw may still give that mana a use, and so is every price above zero. The row now reads: Arcane Signet, turn 9, deficit 0: −0.20, and `LeftoverThreshold` (0.00, cast) in the turn's last main phase when nothing else on offer is priced above it.

It is `Config.IdleLateRocks`, on in `DefaultConfig()` and off in `BaselineConfig()`. It is a price, not part of ADR 0136's plan, so `heuristic-noplan` has it too. Measured in [ADR 0136](0136-planning-the-turns-mana.md#amendment-2026-10-09-a-rock-against-a-spell-over-two-turns)'s PR 4c, which also carries the same day's two-turn amendment.

### 3. Other permanents: priced by mana value and by what their rows say

Ranked second. This covers the enchantment and artifact engines, Equipment, Vehicles, Altars, and every small creature.

**A non-creature, non-mana, non-land permanent** (`permanentValue`'s default arm):

```
value = max(Permanent, PermanentPerMana × mana value) + rowUtility(card)
```

**A creature** (`creatureValue`, and so `CreatureValue`): its body as today, plus `rowUtility(card)` before the tapped, sick and restriction multipliers apply.

`rowUtility` reads `ability_rows`: `RowTriggered` for each triggered row, `RowStatic` for each static row, `RowActivated` for each activated row, counting at most `RowCap` (3) rows. Keyword and mana abilities are not rows, so a flier is not paid twice and a mana elf is priced by §2.

The rows are a signal of utility, not of its size. A row that is a drawback (Sulfuric Vortex hurting its own controller) reads as a positive. That is accepted, because the alternative is reading oracle text. A card that the engine does not implement has no rows and keeps ADR 0037's penalty. The mana-value floor does not apply to it either, so the bot is no keener to cast an uncatalogued card than it is today.

Because this is in `permanentValue`, the evaluation changes too. An opponent's Rhystic Study is now worth removing, and a Blood Artist is worth more than a vanilla 1/1 when the bot chooses what to sacrifice or block with. That consistency is deliberate (ADR 0033's amendment for #1013: one function prices a card both ways). It is also the main regression risk, and §8's suite run is what checks it.

`CombatValue`, which the combat planner uses to compare an attacker with a blocker, is not changed in this ADR (owner decision 3).

Worked:

| Card | Today | Proposed |
|---|---:|---:|
| Rhystic Study (MV 3, one triggered row) | 0.00 | 1.5 + 0.6 − 1.2 = **+0.90** |
| Impact Tremors (MV 2, one triggered row) | 0.00 | 1.2 + 0.6 − 1.2 = **+0.60** |
| Swiftfoot Boots (MV 2, one static row) | 0.00 | 1.2 + 0.5 − 1.2 = **+0.50** |
| Esper Sentinel (1/1, one triggered row) | 0.11 | 0.9 × (1.45 + 0.6) − 1.2 = **+0.65** |
| Viscera Seer (1/1, one activated row) | 0.11 | 0.9 × (1.45 + 0.4) − 1.2 = **+0.47** |

A Vehicle is priced as a non-creature artifact by this rule. Pricing its body when the bot can crew it is a refinement left out (see [Out of scope](#out-of-scope)).

### 4. Board wipes: cast only onto a board where they help

Ranked third. It does nothing for the never-cast count, but it is the one class where the bot actively hurts itself today.

A wipe is recognised by §6's `purpose.sweep`. Without that signal the heuristic cannot tell Wrath of God from Divination, so this part depends on §6.

A wipe's value is the change in the bot's own score, estimated the way `moves.go` estimates every move: take the seat evaluations, remove every permanent the sweep would remove from each seat's board total, recompute `ScoreEval`, and subtract the score before, then subtract `Hand` for the card.

```
wipeValue = ScoreEval(after) − ScoreEval(before) − Hand
```

`ScoreEval` already weighs the opposition by `OpponentMean` and `OpponentMax`, so a wipe that hits the table's leader is worth more, and a wipe that takes the bot's own board while the opponents have little is negative. Which permanents a sweep removes comes from `purpose.sweep`:

- `matches`: one of `creatures`, `nonland_permanents`, `artifacts`, `enchantments`, `artifacts_and_enchantments`, `all_permanents`. More values only when a curated card needs one.
- `how`: `destroy`, `exile`, `bounce`, `damage` or `minus` (−N/−N), with `amount` for the last two.
- A `destroy` sweep leaves a creature with `indestructible` in `abilities`. A `damage` or `minus` sweep leaves a creature whose toughness is above `amount`. A `bounce` sweep counts the permanent at half its value, because it comes back.

A modal wipe (Farewell) carries a `purpose` per mode, and an overloaded one (Damn, Cyclonic Rift) carries it on the overload's `alternative_costs` entry, so the move the bot is offered is priced by the mode or cost it names.

### 5. Spend what would be wasted: the leftover-mana and end-step windows

Ranked fourth. It covers the cheap untargeted spells, the tap abilities and the loots.

Mana empties between steps, and a tapped permanent untaps in its controller's untap step. So in two windows, a move that spends only mana and taps costs nothing that the bot would otherwise keep:

- **Own second main phase**, stack empty. This is the last sorcery-speed window of the turn.
- **The end step of the seat immediately before the bot** in turn order, stack empty. This is the last window before the bot untaps. Every mana source and every tap ability it has is spent for free here, which is when a human plays an end-of-turn Impulse, Vampiric Tutor or Entomb, or loots.

In these two windows the threshold for such a move is `LeftoverThreshold` (0.00) rather than `PassThreshold` or `InstantThreshold`. A move qualifies only if its costs are mana and tapping, plus the card itself for a cast. A move that also costs a life payment, a sacrifice, a discard, another card or a counter keeps the normal threshold.

Three price changes go with it:

- **A floor under every instant and sorcery:** `resolvedValue` is at least `SpellFloor` (1.30). That is 0.10 above the card it costs. A spell nobody can see the effect of is priced as a card that replaces itself and does a little more. It stays below `PassThreshold`, so cheap spells do not crowd out development in the first main phase, and it clears `LeftoverThreshold`. A spell with §6's purpose is priced by its purpose instead, and may clear the normal thresholds.
- **The tapped-blocker price depends on when the tap happens.** Today every tap of an untapped creature costs 0.30. In the end step before the bot's turn it costs nothing, because the creature untaps before any opponent attacks. In the bot's own first main phase, a creature that could attack costs its attack, priced as station already prices it (`moves.go`, #759). Elsewhere it stays 0.30.
- **Mary Read's loot** is then 0.50 in the end step before the bot's turn, which clears 0.00. With §6's `purpose` (`draws: 1, discards: 1`) it is priced by what it draws instead.

The second window needs no new signal. The view carries the active seat, the step and the seat order, and the "next live seat after the active one is me" test is three lines.

### 6. One new wire signal: a declared `purpose` for what a spell or an ability does

The signal follows the precedent of ADR 0106 §1 decision 8: catalog data the bot reads, declared on the card, never inferred from text.

**The type.** `protocol.ActivationPurposeView` becomes the one purpose type, `PurposeView`. Its existing fields keep their names and meaning, and the new ones are additive:

| Field | Meaning | Example |
|---|---|---|
| `draws` | cards its controller draws (exists) | Night's Whisper 2, a loot 1 |
| `controller_loses_life` | life the source's controller loses (exists, any-player rows) | — |
| `discards` | cards its controller discards on resolution | a loot 1, Faithless Looting 2 |
| `lands` | lands it puts onto the battlefield | Rampant Growth 1, Harrow 2 |
| `tutors` | cards it searches out to hand or to the top of the library | Demonic Tutor 1, Vampiric Tutor 1 |
| `self_mill_tutor` | cards it searches out into the graveyard | Entomb 1 |
| `tokens` | Treasure or other tokens it makes for its controller | Big Score 2 |
| `sweep` | `{matches, how, amount}` (§4) | Wrath of God `{creatures, destroy}` |
| `death_payoff` | on a triggered row: the row pays out whenever a creature of its controller's dies | Blood Artist, Zulaport Cutthroat |
| `discard_payoff` | on a triggered row: which discarded cards it pays on, and what it pays for each ([amendment of 2026-10-06](#amendment-2026-10-06-discard-payoffs)) | Mary Read and Anne Bonny, Marauding Mako |

**Where it is carried.** On `CardView` (the spell, or a permanent's enters-the-battlefield effect), on each `ModeOptionView`, on each `AlternativeCostView`, on `ActivatedAbilityView` (the existing field, now also for the controller's own rows), and on `AbilityRowView` (for `death_payoff`). It is projected in the same zones as `ability_rows`, hand and battlefield, and cleared for a viewer who may not see the card, like them. An opponent's hand card never has one. The stack's top card has one, so §7 can see a wipe coming.

**Where it comes from.** A `Purpose` field on `effects.Spec`, and on the mode, alternative-cost, activated and triggered rows. It is declared by hand, like `Completeness`. Nothing derives it at run time.

**What is declared, and the guard.** For S66: every card in the four curated decks in the classes this ADR prices (ramp, draw, loot, tutor, wipe, death payoff), and every catalog card that is a board wipe, because a wipe without a purpose is still cast onto any board. `TestCuratedDeckPurposes` (in `internal/decks`, no dump needed) fails when a curated deck's instant or sorcery has neither a `Purpose` nor an entry on a short, named list of spells whose value is their target (removal, counterspells). A manual test over the Scryfall dump, beside `realdump_manual_test.go`, lists catalog cards whose oracle text reads as a sweep, a draw, a tutor or a land search and that declare no purpose. It is a review aid that is not run in CI, because CI has no dump.

**What the heuristic does with it.** When a card or row has a purpose, `resolvedValue` uses the purpose instead of the mana-value proxy:

```
purposeValue = Hand × (draws + TutorWeight × tutors + SelfMillWeight × selfMillTutor)
             − DiscardWeight × discards
             + (ManaSource + RampPremium) × lands
             + TokenWeight × tokens
```

A sweep is priced by §4 instead.

**No snapshot change.** A purpose is catalog data projected into the view, not game state. Nothing new is captured, and `SnapshotSchemaVersion` does not move. Clients ignore the new fields. `docs/protocol.md` documents them.

**As built (PR 6, 2026-10-06).** Declaring every catalog wipe (owner decision 2) met wipes §4's three-field `sweep` cannot describe truthfully, so PR 6 made these additive choices. Each is catalog data and changes no price; PR 7 decides how to read them.

- `sweep` gained three optional fields: `amount_is_x` (Earthquake, Toxic Deluge, Crypt Rats), `opponents_only` ("creatures your opponents control", Cyclonic Rift's overload, a sweep that targets a player), and `partial` (the sweep spares some of its class by a condition the class does not name: nonwhite, without flying, power 4 or greater). With `partial`, `matches` is an upper bound.
- `how` gained `sacrifice` (All Is Dust, Tragic Arrogance, The Eternal Wanderer's −4).
- `matches` gained the two classes a curated card needs, Austere Command's `creatures_mana_value_3_or_less` and `creatures_mana_value_4_or_greater`, and nothing else. A wipe outside the curated decks declares the smallest listed class that holds what it removes, leaving out kinds no class names (planeswalkers, battles, lands), with `partial` when that class holds permanents it spares: Nevinyrral's Disk is `nonland_permanents`, partial.
- An amount that is X or counted at resolution is not declared (Pull from Tomorrow, Windfall, Shamanic Revelation). `TestCuratedDeckPurposes` names those curated spells on a second list, `noPrintedAmount`, beside the target list. Living Death is on it: it is a symmetric mass reanimation, and a sweep purpose would price the sacrifice and miss the return.
- A permanent's enters effect is the card's own `purpose` (Mulldrifter `{draws: 2}`, Wood Elves `{lands: 1}`), as §6 says. A triggered row declares `death_payoff`, and a saga chapter or an unlock trigger that sweeps declares its `sweep`. The activated rows of the curated decks' loots, tutors and land sacrifices declare theirs.

### 7. Discard costs and sacrifice outlets

Ranked fifth and sixth, because the cards are fewer.

**Discard-cost spells** (Unexpected Windfall, Big Score, Faithless Looting). The cost is priced by what each discarded card is worth to the bot, using `cardValue`, the price the discard-choice branch already uses. It is no longer a flat 1.20 per card. A late land in hand costs 0.30 and an early one 1.50. The enumerator already offers one payment per discard combination (`legal/cast.go`), so the bot picks the cheapest. The spell's own draws come from §6's `purpose`. Unexpected Windfall late, discarding a spare land: 2 × 1.2 (draws) + 2 × `TokenWeight` (Treasures) − 0.30 − 1.20 is comfortably positive. Early, discarding the only land in hand, it is not.

**Sacrifice outlets** (Viscera Seer, Carrion Feeder, Ashnod's Altar, Phyrexian Altar, Warren Soultrader, High Market, Vampiric Rites, Deadly Dispute's additional cost). A sacrifice costs the creature's value times the chance the bot would have kept it:

- **Dying anyway.** The creature is a target of an opponent's item on the stack (× `RemovalConfidence`, 0.80), or a declared sweep on the stack would remove it (× 1), or in the declare-blockers step it is in a combat it loses, by the combat planner's own trade arithmetic (× 1). Its sacrifice then costs only the part it would have kept. That makes "sacrifice it in response" the bot's play.
- **Death payoffs.** Each `death_payoff` row the bot controls adds `DeathPayoff` to every creature it sacrifices. With a Blood Artist and a Zulaport Cutthroat out, sacrificing a token to Viscera Seer is worth doing. Sacrificing a commander or a big creature still is not.

A mana ability with a sacrifice cost (the Altars) is used by the auto-tapper when a cast needs it, as today. The bot does not activate one for floating mana.

### 8. How it is measured, and the bar it must clear

Every sub-PR from §1 onward carries ADR 0052's report block in its description, plus this ADR's two tables. The runs:

1. **The four-deck rotation.** `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,mono-red-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1`. The aggro deck was `izzet-aggro` until [#3071](#amendment-2026-10-13-izzet-aggro-rebuilt-around-purphoros), and the measurements below that predate it name it so. 64 games is 16 full rotations, so every deck sits in every chair 16 times. It reports per-deck win rates with intervals, the Cards table and the `never` counts. About 5 s a game, so about 6 minutes.
2. **New against baseline.** Two `heuristic` and two `heuristic-baseline` contestants, 48 games with the decks split one way and 48 with them swapped, so each policy plays each deck equally (96 games, 192 seat-games per policy). This measures strength, which run 1 cannot: four copies of one policy win 25% each by construction.
3. **The position suite.** `boteval suite run --policy heuristic`. Every gated position passes. Overall agreement and each tag's agreement are not lower than before the PR. Each sub-PR adds the positions for its class (below), labelled from its own arena decision logs, and gates them.
4. **The nightly gates.** `bot-games`, `bot-soak` (the heuristic gate's fixed seeds 101..120, the random soak of 100 games) and `catalog-soak` stay green on the PR's branch, run with `gh workflow run` before merge. If the catalog soak fails with an `EventEffectError` on a card the bot now casts for the first time, that is a real engine bug the pricing has uncovered. It is filed as its own issue and fixed, or the card's `Completeness` is corrected. It is not a reason to revert the price.

**Acceptance bar for S66** (the last sub-PR's run 1 and run 2, on today's four decks, before #2436):

| # | Measure | Today | Bar | Why this number |
|---|---|---|---|---|
| A1 | `never` count per deck (run 1) | 19–30 | **≤ 5 for every deck** | Some cards are legitimately held: a counterspell with nothing to counter, a removal spell with no target worth it, Kismet. Five per deck leaves room for those and nothing else. |
| A2 | Mana rocks and dorks (run 1) | used in 0% of games where offered | **used in ≥ 80% of the games in which each was offered** (amended for S66 on 2026-10-07: see [Exit decisions](#exit-decisions-2026-10-07)) | Casting a rock is almost always right. 80% leaves room for a turn where a land and a bigger spell use all the mana. |
| A3 | The six canaries: Sol Ring, Rhystic Study, Faithless Looting (cast, for Mary Read's loot, see the note below), Entomb, Harrow, Viscera Seer (run 1) | 0 uses | **each used in ≥ 50% of the games in which it was offered** | They name the six classes. If one stays dead, its class is not fixed. |
| A4 | Wipes (suite) | always cast | **both new wipe positions pass, and are gated** | Arena games rarely set up the bad-wipe board, so a fixed position is the reliable test. |
| A5 | New against baseline (run 2) | n/a | **`beats null: yes` for `heuristic`**: the Wilson interval's lower bound above 25% | The claim is that the bot plays its deck better, so it must beat the old bot in a fair seating. With 192 seat-games, that needs about 31% or more (about 60 of 96 games). |
| A6 | Balance (run 1) | mono-black 63%, interval 49.9%–75.2% | **no deck's interval lies entirely above 50%** | The pricing must not widen black's lead. Making the decks even is #2436's job. This bar only says this ADR did not make it worse. |
| A7 | Suite | 22 positions | **every gated position passes; no tag's agreement falls** | ADR 0052 §1: a trade is not a win. |
| A8 | Health (runs 1 and 2) | 0 stalls | **0 stalls; rejected moves not above the baseline run's; turns p50 within ±3 of baseline** | More card classes in play means more chances of a move the engine refuses or a table that does not end. |

Each intermediate sub-PR must hold A7 and A8, and its run 2 must not be detectably worse: the `heuristic` interval's upper bound must not fall below 25%. Its own class's canaries should move, and the PR says by how much. A1 to A6 are judged at the end, because the classes interact. Casting rocks means more mana for engines, and the `never` count only falls far once all of them land.

If A5 or A6 fails at the end, the owner decides whether to ship, retune or narrow the scope, with the numbers in [Measurements](#measurements).

**Canary swap (2026-10-13, #3071 question 4; #3164).** Mary Read and Anne Bonny left every curated deck when #3071 rebuilt izzet-aggro as `mono-red-aggro` around Purphoros, so her loot canary was in no deck and the arena report marked it `not offered`. The owner replaced it with **Faithless Looting (cast)**, the loot in the rebuilt deck. The class stays "loot" and the bar stays 50%. Mary Read's loot is still priced by the discard-payoff amendment, but it is no longer measured as a canary.

**Positions added, one or more per sub-PR**, each harvested from that PR's run-1 decision logs with `boteval suite harvest`, labelled, and gated for `heuristic`:

- `cast-sol-ring-turn-one`, `cast-the-signet-when-short`, `cast-the-elf-turn-one`, `hold-the-signet-late` (pass and the land drop both accepted).
- `cast-rhystic-study`, `cast-viscera-seer`, `equip-before-attacking`.
- `loot-at-the-end-step-before-yours`, `tutor-at-the-end-step-before-yours`, `cantrip-with-leftover-mana`, `do-not-cantrip-before-the-three-drop`.
- `do-not-wrath-your-winning-board`, `wrath-a-losing-board`.
- `windfall-discarding-a-spare-land`, `do-not-windfall-away-the-last-land`.
- `sacrifice-the-creature-the-removal-targets`, `do-not-sacrifice-the-commander`.

The owner reviews each label in the PR that adds it, like every suite position.

### 9. The weights, and keeping the old bot runnable

Every new term is a `Config` or `Weights` field whose zero value is today's behaviour. `BaselineConfig()` returns `DefaultConfig()` with all of them zeroed. That is what the arena's `heuristic-baseline` contestant runs (§1), and a test asserts that `BaselineConfig()` ranks every suite position exactly as the policy did before S66.

Starting values. They are tuned in each sub-PR against the suite and run 2, and the PR records the values it ends on:

| Field | Start | Section |
|---|---:|---|
| `ManaPerExtra` | 1.00 | §2 |
| `RampPerMana` | 1.00 | §2 |
| `RampWantCap` | 7 | §2 |
| `PermanentPerMana` | 0.50 | §3 |
| `RowTriggered` / `RowStatic` / `RowActivated` | 0.60 / 0.50 / 0.40 | §3 |
| `RowCap` | 3 | §3 |
| `LeftoverThreshold` | 0.00, with `LeftoverWindows` on | §5 |
| `SpellFloor` | 1.30 | §5 |
| `TutorWeight` / `SelfMillWeight` | 1.00 / 0.50 | §6 |
| `DiscardWeight` | 0.60 | §6 |
| `TokenWeight` | 0.50 | §6 |
| `DeathPayoff` | 0.60 | §7 |

Nothing here touches Layer A (`aiseat/rules`), which never chooses a cast. The model tiers inherit the change, because the heuristic is their Layer B, their fallback, and the ranking that orders the moves their prompts show (`model/prompt.go`).

---

## Out of scope

- **The deck rebalance** (#2436) and **a fifth deck** (#2437). Both are measured with this pricing once it lands.
- **The combat planner.** `CombatValue` stays body-only (owner decision 3). How the bot attacks and blocks is not changed here.
- **The flat `ActivateBase` for a non-tap activated ability.** Konrad's mill is the visible case. It stays at 0.50 (owner decision 6). §6's purpose would let a later change price these rows.
- **Rituals and one-shot mana** (Dark Ritual, Lotus Petal, Treasures spent for a specific cast). These need a plan for the turn ("this mana lets me cast that"), which is lookahead, and ADR 0033 §3 rules out cloning a game.
- **Vehicles' bodies, Equipment's best host, Aura targets.** Priced as ordinary permanents here. The existing equip and attach pricing stays.
- **Holding up counterspells and instant-speed removal.** `InstantThreshold`'s "hold it" behaviour is unchanged outside §5's two windows, and §5 only covers moves that spend nothing but mana and taps.
- **Reading oracle text in the heuristic.** Ruled out by ADR 0033 §3's intent. §6 is the replacement.
- **A purpose for every catalog card.** S66 declares purposes for the curated decks and for every board wipe. Other cards use the generic floors, which already move them from "never" to "when the mana is spare".
- **Politics, threat assessment beyond `ScoreEval`, and learning.**

---

## Delivery

One lever per PR, as ADR 0052 asks. Each is separately measurable against run 1, run 2 and the suite, and each says which of A1 to A8 it moves.

| # | PR | Depends on |
|---|---|---|
| 1 | This ADR and the S66 sprint section | — |
| 2 | **Measurement.** The arena's per-contestant Play table and Cards section, `heuristic-baseline` and `BaselineConfig()` (with no new terms yet, so baseline and default are equal), and the baseline runs recorded in Measurements. No price changes. | 1 |
| 3 | **Mana sources** (§2): the `ManaPerExtra` value on the battlefield and the ramp premium at cast time. Positions: the rock and elf positions. | 2 |
| 4 | **Permanents by what they do** (§3): the mana-value floor and `rowUtility`, in `permanentValue` and `creatureValue`. Positions: Rhystic Study, Viscera Seer, equip. | 2 |
| 5 | **The two windows** (§5): own second main and the end step before the bot's turn, `SpellFloor`, and the tapped-blocker price by timing. Positions: loot, tutor, cantrip. | 2 |
| 6 | **The purpose signal** (§6): `PurposeView`, `Spec.Purpose` and its row forms, the projection and redaction, `docs/protocol.md`, declarations for the curated decks and every catalog wipe, `TestCuratedDeckPurposes`, and the manual dump audit. No price changes. Touches `effects`, `game`, `protocol` and `decks`, and nothing under `aiseat/`. | 1 |
| 7 | **Wipes, ramp spells and discard costs** (§4, §6's prices, §7's discard half). Positions: the two wipe positions and the two Windfall positions. | 3, 6 |
| 8 | **Sacrifice outlets** (§7's second half): dying anyway and death payoffs. Positions: the two sacrifice positions. | 4, 6 |
| 9 | **Exit.** The final run 1 and run 2 against A1 to A8, the Measurements rows, a "How the heuristic prices a card" section in `docs/bot.md` (replacing the pricing claims in its Known limitations), and the evidence on #2435. | 3–8 |

PRs 3, 4, 5 and 6 are independent of each other once PR 2 is in, so they can be built in parallel and merged in any order. Each is measured against the `develop` it merges into.

---

## Owner decisions (2026-10-06)

The owner answered the eight open questions on 2026-10-06. Questions 1, 2, 4 and 8 were answered directly. Questions 3, 5, 6 and 7 were taken as recommended; the owner was told and did not object. Every answer is the recommendation the draft gave, so no section above changed. These answers are binding on the Delivery PRs.

1. **How the bot learns what a spell does: a declared purpose.** The catalog declares `PurposeView` on the `Spec` and its modes, alternative costs and rows (§6), plus the manual audit over the Scryfall dump. Rejected: deriving it from oracle text on the server, and having no signal.
2. **Which cards get a purpose in S66: the curated decks' cards in the priced classes, plus every catalog board wipe.** Rejected: the curated decks only, and every card in the classes.
3. **Row utility in combat: not in S66.** `CombatValue` stays body-only, and ability rows count in the board value only (§3). If the decision logs show the bot making bad trades because of it, open a follow-up issue.
4. **The strength bar: A5 as written.** At exit the new heuristic must beat `heuristic-baseline` outright (its Wilson interval above the 25% null over 96 games). Each sub-PR before that must not be detectably worse (§8).
5. **`heuristic-baseline` after S66: kept.** It stays as the frozen reference until the next ADR that changes the heuristic's prices, which replaces it with a new frozen config (§9).
6. **Konrad's mill and the flat `ActivateBase`: left alone in S66.** Non-tap activated abilities stay at +0.50. Acceptance bar A6 decides whether black's lead needs a later change.
7. **Which decks are measured: today's four.** The acceptance bar is measured on the current curated decks. #2436 is measured afterwards with the fixed heuristic.
8. **The windows: both.** §5 applies in the bot's own second main phase and in the end step of the seat immediately before the bot's turn.

Later in S66, the owner made three more decisions on #2435:

- **A2 counts only offers made while the mana deficit is open** (2026-10-06). A late one-mana rock that §2 prices below zero on purpose, as in its "Arcane Signet, turn 9" row, does not count against the 80% bar. PR 3's run 1 had shown why: 221 of 287 unused one-mana rock and elf offers came with a deficit of 0 or less. PR 9 measures A2 this way.
- **The label review** (2026-10-06): ten of the thirteen S66 positions were approved as written. The other three are recorded under PR 7 in [Measurements](#measurements): `windfall-discarding-a-spare-land` led to the [discard-payoff amendment](#amendment-2026-10-06-discard-payoffs), and the other two keep accepted answers the heuristic gives.
- **Two follow-ups** came out of that review, outside S66: [#2457](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2457) (cycling cards as cheap discards, and the bot's own draw step as a spend window) and [#2458](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2458) (draw before deploying: a turn-mana lookahead for cantrips).

## Exit decisions (2026-10-07)

PR 9 measured the acceptance bar ([Measurements](#pr-9-the-exit-2026-10-06)). A1, A4, A6 and A7 passed; A2, A3, A5 and A8 failed. As §8 provides, the owner decided each on 2026-10-07. S66 ends with these decisions.

- **A6: black's lead goes to #2436.** No retune in S66. The `ActivateBase` experiment showed that Konrad's flat-priced mill is not what makes mono-black win, and that the frozen baseline wins as often on that deck. Owner decision 6 stands: `ActivateBase` stays at 0.50.
- **A2: amended for S66.** The bar was "used in ≥ 80% of the games in which each was offered", already narrowed by the owner on 2026-10-06 to offers made while the mana deficit is open. The amendment: in an open-deficit window, a rock or dork that is not cast must lose its window to a land drop or to another spell, never to a pass. PR 9's run 1 meets that: in all 571 such windows the bot played a land (275), cast another spell (295) or activated an ability (1). The 80% games-used rate (68% measured) is not an S66 bar. Casting the rock first and the spell a turn later needs a plan for the turn's mana, which is [#2458](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2458).
- **A3: accepted for S66, with Harrow as a follow-up.** Five of the six canaries pass. Harrow (30%) keeps the normal bar in the leftover windows because its land sacrifice is not a mana-and-taps cost. [#2469](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2469) treats a land sacrifice that the move's `purpose.lands` replaces as net mana for that rule. It is outside S66.
- **A5: accepted on the pooled result.** Over the 96 games §8 names, `heuristic`'s interval (22.2%–34.9%) does not clear 25%. Over 288 games on three seed blocks it is 165 of 574, 28.7% (25.2%–32.6%), and every block points the same way.
- **A8: accepted as known.** Every run-1 rejection was a pass racing a prompt that another seat's answer opened, which the runner recovers from. The two real defects behind the other rejections, [#2461](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2461) (auto-tap and bounce lands) and [#2462](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2462) (attacks after the active seat's pass), are fixed next, before #2436 is measured.

## Amendment (2026-10-06): discard payoffs

The owner approved this on #2435, after reviewing PR 7's position `windfall-discarding-a-spare-land`. The bot plays izzet-aggro there, with its commander Mary Read and Anne Bonny on the battlefield: "Whenever you discard an Island, Pirate, or Vehicle card, create a tapped Treasure token." The owner's answer is to discard the Island. The bot could not see why. §6 gave a triggered row `death_payoff` and nothing for a discard, and §7 priced a discarded card by `cardValue` alone, so the Island and the Mountain cost the same.

**The signal.** `PurposeView` gains `discard_payoff`, on a triggered row only, like `death_payoff`. It says which discarded cards the row pays on, and what it pays for each one, as printed amounts:

| Field | Meaning | Example |
|---|---|---|
| `any` | every card its controller discards | Marauding Mako |
| `types` | otherwise, the card types and subtypes it pays on, lowercase; a card with any one of them on its type line matches | Mary Read `["island", "pirate", "vehicle"]` |
| `tokens` | tokens it creates for its controller per card | Mary Read 1 |
| `counters` | +1/+1 counters it puts on its source per card | Marauding Mako 1 |
| `damage_each_opponent` | damage its source deals to each opponent per card | Glint-Horn Buccaneer 1 |

**How it is declared.** `game.Purpose.DiscardPayoff` on the card file's triggered row, with `TriggerWithPurpose`, by hand like every purpose. The registration guard refuses one anywhere but a triggered row, one with neither `any` nor `types` or with both, a type that is not one lowercase word, a negative amount, and one that pays nothing. It is projected on `ability_rows` with the row and cleared with it, so a hidden hand card and a face-down permanent never carry one. No snapshot change.

**How a discard is priced against it.** When the bot discards one of its own cards, every `discard_payoff` on a permanent it controls that matches the card pays:

```
payoff = TokenWeight × tokens
       + (Weights.Power + Weights.Toughness) × counters
       + DamageToOpponent × damage_each_opponent × live opponents
```

These are the units the policy already uses: a Treasure as a purpose's token, a +1/+1 counter as `counterRemovalValue` charges for losing one, a point of damage as an attack prices it. The discard costs `cardValue − payoff`. That applies wherever the policy prices a discard of its own card: §7's discard cost, the card it names on resolution (a loot, a rummage, "discard a card") and at cleanup, and the discards a purpose declares (a loot's `discards: 1`). For the last, the payoff is the payoff of the cards the bot would name, read from the hand it holds. `PriceDiscardPayoffs` switches it on in `DefaultConfig()` and is off in `BaselineConfig()`.

Worked, in the owner's position: late in the game a land in hand has a `cardValue` of 0.30. The Island makes a Treasure (0.50), so discarding it costs −0.20, against the Mountain's 0.30. Breeches, Brazen Plunderer is a Pirate and makes the same Treasure. It is a castable four-mana creature, though, and its `cardValue` is several points, so a 0.50 Treasure does not close the gap. It stays rejected.

**Which cards declare it.** Every curated card with a discard payoff:

- Mary Read and Anne Bonny: `types` Island, Pirate, Vehicle, `tokens: 1`.
- Marauding Mako and Scrounging Skyray: `any`, `counters: 1`.
- Magmakin Artillerist and Glint-Horn Buccaneer: `any`, `damage_each_opponent: 1`.
- Hashaton, Scarab's Fist: `types` creature, `tokens: 1`. The `{2}{U}` it asks for is not declared. A token is priced well under a 4/4, which leaves room for the mana.

`TestCuratedDeckPurposes` holds each to its declaration.

## Amendment (2026-10-08): purposes that follow a mode's target, and damage priced by whether it kills

**Status:** Accepted (owner answers 2026-10-08). The owner chose the recommended option, (a), on every question; see [Owner answers (2026-10-08)](#owner-answers-2026-10-08) at the end of this amendment. No code changed with it. The changes land in the PRs under its delivery plan.
**Issue:** [#2689](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2689). Related: [#2681](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2681) (Prismari's self loot and Treasure selection is never enumerated) and [#2457](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2457) (the bot's own draw step as a spend window).
**Amends:** §6 (what a purpose says, and who it is about) and the target half of the cast price. §4's sweep, §7, and the discard-payoff amendment are unchanged.
**Line numbers** are on `develop` at `8cda8c86d`. The issue cited `46f6b1e0c`; `aiseat/heuristic/moves.go` has not moved since, and `legal/legal.go`'s cap moved from `:554` to `:562`.

### The problem

**The evidence.** Review game 2 (`06e98afa`, 2026-10-08) was a two-seat table: the heuristic on izzet-aggro (Bot 1, life 34) against a Claude MCP seat on Esper (life 41). At decision-log seq 248, in the bot's turn-6 draw step, the bot had three lands and no creatures. Its hand was Angrath's Marauders, Goldspan Dragon, Solphim, Unexpected Windfall, Captain Lannery Storm, Malcolm and Prismari Command. Claude controlled Y'shtola, Night's Blessed, a 2/4 commander with no damage marked.

Prismari Command reads: "Choose two — • Prismari Command deals 2 damage to any target. • Target player draws two cards, then discards two cards. • Target player creates a Treasure token. • Destroy target artifact." The enumerator offered twelve casts. Every one of them was modes [0,1] or [0,2]; #2681 covers why. The trace priced them as follows:

| Mode 0's target | Mode 1's or 2's target | [0,1] | [0,2] |
|---|---|---:|---:|
| Y'shtola | Claude | **9.12** (chosen, index 5) | 9.12 (index 11) |
| Y'shtola | Bot 1 | 5.82 | 5.82 |
| Claude | Claude | 4.20 | 4.20 |
| Claude | Bot 1 | 0.90 | 0.90 |
| Bot 1 | Claude | 0.90 | 0.90 |
| Bot 1 | Bot 1 | −2.40 | −2.40 |

Pass was 0. The bot cast the top line: 2 damage to a 2/4, which kills nothing, and "Claude draws two, then discards two", which let Claude pitch two spare lands. It spent all three lands in the draw step, so it cast nothing in its main phase. Mary Read and Anne Bonny (4.17), Malcolm and Lannery Storm were all castable there.

**Why, in the code.** Every price in the table is one sum:

```
0.60 (3 × SpellPerMana 0.60, less the card's Hand 1.20)
+ a price per target
```

- **A player target is always priced as an attack.** `targetsValue` (`aiseat/heuristic/moves.go:781`) charges `SelfTargetPenalty` (−1.50, `:787`) for the bot itself, and pays `DamageToPlayer × leaderBoost` (+1.20 × 1.50 = +1.80, `:793`) for anyone else, whatever the mode does to them. Giving Claude two new cards or a Treasure is priced the same as burning Claude.
- **A creature target is always priced as removal.** `cardTargetValue` (`:820`–`:829`) prices an opponent's creature at `CreatureValue × RemovalConfidence × leaderBoost`, which is 5.60 × 0.80 × 1.50 = 6.72 for Y'shtola. Nothing asks whether 2 damage kills a 4-toughness creature.
- **The mode never enters the price.** Prismari declares no purpose on any mode (`cards/effects/prismari_command.go`). `castPurpose` (`purpose.go:98`) then falls back to the card, which has none, and `resolvedValueFor` (`fuel.go:217`–`:223`) uses the mana-value proxy. So [0,1] and [0,2] are the same number.
- **The signal could not say it.** `PurposeView`'s amounts are all the controller's: `draws` is "cards its controller draws" (`protocol/purpose_view.go:28`). Declaring Prismari's loot as `draws: 2, discards: 2` would add +1.20 to every cast whoever it targets, and keep the attack price on Claude. That is why `TestCuratedDeckPurposes` lists Prismari on `valueIsTheirTarget` with the class "its targets decide who draws and who gets the Treasure, by mode" (`internal/decks/purpose_test.go:93`). Sign in Blood is there for the same reason (`:99`).
- **The damage field exists, but nothing reads it for a spell.** `damage_to_creature` came with [ADR 0130's amendment of 2026-10-07](0130-exert.md#amendment-2026-10-07-exert-rows-and-what-they-declare) for exert rows. Only `exertGain` reads it, through `bestCreatureKill` (`exert.go:99`, `:119`), which checks `effectiveToughness(c) <= dmg`. Two catalog rows declare it (Glorybringer and Fervent Paincaster). No spell or mode does, and the cast price never looks.

**How wide it is.** These are greps, so they are upper bounds:

- 41 card files in `cards/effects` are modal and have a mode that targets a player.
- 265 files have a player or opponent target clause.
- About 280 name a damage effect and a creature or any-target clause.

The curated decks put it in play every game:

- izzet-aggro holds Lightning Bolt, Shock, Arc Trail, Fiery Temper, Abrade, Izzet Charm and Prismari Command. Every one of them prices any opposing creature as killed.
- mono-black-aristocrats holds Sign in Blood ("Target player draws two cards and loses 2 life"). By the same arithmetic, the bot always prefers to cast it at an opponent rather than at itself. That case is from the code, not from a logged game.

The hostile default is right for some effects: a discard (Kolaghan's Command), a mill, a life loss and damage are all bad for their target. It is wrong for a draw, a token, a life gain and a loot.

**What the rules say.** I checked every rule below against the Comprehensive Rules effective September 25, 2026.

- **CR 601.2c:** "if the spell uses the word 'target' in multiple places, the same object or player can be chosen once for each instance of the word 'target'". So Prismari's [1,2] with the bot as both targets, the play the review wanted, is a legal cast.
- **CR 120.6:** "If the total damage marked on a creature is greater than or equal to its toughness, that creature has been dealt lethal damage and is destroyed as a state-based action". **CR 704.5g** is that state-based action.
- **CR 514.2:** in the cleanup step, "all damage marked on permanents … is removed". Damage that does not kill is gone at end of turn.
- **CR 702.12b:** an indestructible permanent is not destroyed by lethal damage.
- **CR 702.2b:** a creature dealt damage by a source with deathtouch is destroyed.
- **CR 702.16e:** damage from a source of the protected quality is prevented.
- **CR 120.3c:** damage to a planeswalker removes that many loyalty counters. **CR 704.5i:** a planeswalker with loyalty 0 goes to the graveyard.
- **CR 120.3d:** damage from a source with wither or infect is dealt as −1/−1 counters, which stay.
- **CR 704.5a:** "If a player has 0 or less life, that player loses the game."

### Options considered

**A. How the catalog says who a purpose is about.**

1. **(Recommended) Per target clause.** `game.Purpose` gains a list of target purposes, each keyed by the `slot` of the target clause it describes within its statement: the card's own statement, a mode's, an alternative cost's or an activated row's. Each entry holds what happens *to that target*: `draws`, `discards`, `tokens`, `life_gain`, `life_loss` and `damage`. A move's target already names its clause: `targets[].slot`, and `targets[].mode`, an index into the move's `modes` (`legal/legal.go:962`–`:971`, ADR 0065 §2). So the heuristic can match every pick to its entry without a wire change to moves. This is the only option that can say Arc Trail ("2 damage to any target and 1 damage to another target") truthfully, and it follows CR 601.2c's one-target-per-instance model, which the engine already uses.
2. **A mode-level flag:** "this mode's amounts apply to its target player". It is smaller, and enough for Prismari and Sign in Blood. It cannot split a statement with two clauses, and it needs a separate damage field anyway.
3. **Inferring it from the clause kind:** a mode whose only target is a player has its amounts apply to that player. Rejected. §6's rule is "declared on the card, never inferred", and the inference is wrong for "Target opponent sacrifices a creature. You draw a card."
4. **Resolving the spell on a copy of the game and scoring the result.** Rejected. ADR 0033 §3 rules out cloning a game in a policy.

**B. How a gift to another seat is priced.**

1. **(Recommended) The bot's own formula, through the opposition weights.** For the bot as target, the entry is priced by `purposeValue`'s amount terms, as if the bot had cast an untargeted "you draw two". The discard payoffs and the cards it would discard come with that (`resolutionDiscardPayoff`). For an opponent, the same amounts make a strength change `x` for that seat. It is priced as the change in `ScoreEval` (`score.go:952`): `−(OpponentMean × x / n + OpponentMax × x)` when that seat is the strongest opponent, and only the mean term otherwise. That is how §4 prices a sweep, and it weighs a four-player table without a new constant. At a two-seat table it is −1.5x, which is what `LeaderBoost` (1.50) gives.
2. **A sign table with `LeaderBoost`**, as the issue proposed: +value on the bot, −value × `leaderBoost` on an opponent. Simpler, but it treats a gift to one of three opponents as costing the bot as much as the same gift to itself.

**C. How damage is priced.**

1. **(Recommended) By whether it kills, else nothing.** A creature target dies if it is not indestructible, the damage is not prevented by protection from the source, and either `damage ≥ toughness − damage_marked` (CR 120.6) or the source has deathtouch and `damage ≥ 1` (CR 702.2b). The combat planner's `kills` (`combat.go:127`) already makes these checks for combat. A creature that dies is priced as removal is today. One that survives is priced at `DamageChip × removal value`, with `DamageChip` 0.00 (Q3). The bot's own creature is priced negatively if it dies and 0 otherwise, rather than `OwnPermanentTarget`'s +0.40 pump guess. A planeswalker loses loyalty for good, so it is priced by the share of loyalty removed, and as killed at or above its loyalty (CR 120.3c, 704.5i). A battle keeps today's price (Q6).
2. **In proportion to damage over toughness.** Rejected. Marked damage is removed in cleanup (CR 514.2), so 2 of 4 points is worth nothing after this turn. The one exception is damage combined with a combat this turn, and that is the combat planner's business.

**D. Damage to a player with a declared amount.**

1. **(Recommended) Per point.** The amount is priced at `DamageToOpponent` (0.30, the unit the attack planner uses) per point, through B's opposition weights. It gets `LethalBonus` when the amount is at least the player's life and the player can lose to life (CR 704.5a). That replaces `targetsValue`'s `FinishLife` bet (`moves.go:800`–`:811`) with a reading, for a declared spell only. Damage to the bot itself is priced as the life lost, at `MarginalLife`.
2. **Today's flat `DamageToPlayer × leaderBoost` per target, whatever the amount.** Smaller, but a 2-damage mode stays worth 1.80 at a player on 41 life. At seq 248 that is enough to cast Prismari in the draw step anyway (see below).

**Undeclared cards keep today's price.** A target with no entry is priced by `targetsValue` exactly as now. Nothing is inferred.

### Recommendation

A1, B1, C1 and D1, in two pricing PRs behind two `Config` switches. `PriceTargetPurposes` covers A and B; `DamageByLethality` covers C and D. Both are on in `DefaultConfig()` and off in `BaselineConfig()` (§9), so each lever is measured alone, as ADR 0052 asks. Then declarations: every curated card in either class, and every catalog card whose spell, mode or row gives its target player something (Q5).

**Worked at seq 248**, with the recommended answers. Claude is the only opponent, so B1's weight is 1.5. Prismari declares mode 0 `damage: 2`, mode 1 `draws: 2, discards: 2` and mode 2 `tokens: 1`, each on slot 0. The cast is then purpose-priced, so the 1.80 mana proxy goes, and the card still costs Hand (1.20).

| Selection | Today | Amended |
|---|---:|---:|
| 2 to Y'shtola, loot to Claude | 9.12 | 0 − 1.80 − 1.20 = **−3.00** |
| 2 to Y'shtola, loot to Bot 1 | 5.82 | 0 + 1.20 − 1.20 = 0.00 |
| 2 to Claude, loot to Bot 1 | 0.90 | 0.90 + 1.20 − 1.20 = 0.90 |
| 2 to Claude, Treasure to Bot 1 | 0.90 | 0.90 + 0.50 − 1.20 = 0.20 |
| loot and Treasure to Bot 1 (after #2681) | not offered | 1.20 + 0.50 − 1.20 = 0.50 |

The bot's loot is 2 × 1.20 − 2 × `DiscardWeight` 0.60 = 1.20; Mary Read is not on the battlefield, so there is no discard payoff. The 2 damage to Claude is 2 × 0.30 × 1.5 = 0.90. Every Prismari cast is now below `InstantThreshold` (1.50), so the bot passes the draw step and casts Mary Read and Anne Bonny (4.17) in its main phase. That is the issue's accepted answer, and it needs no #2457 gate.

Under D2 instead, "2 to Claude, loot to Bot 1" is 1.80 + 1.20 − 1.20 = 1.80, which clears 1.50, and the bot still casts in its draw step. So D matters to the acceptance position.

### Snapshot and wire impact

- **No snapshot change.** A purpose is catalog data projected into the view, as §6 says. Nothing new is captured, `SnapshotSchemaVersion` stays 7, and `testdata/snapshot_shape/v7.txt` does not change.
- **Wire, additive.** `PurposeView` gains `targets`, omitted when empty: a list of `TargetPurposeView` `{slot, draws, discards, tokens, life_gain, life_loss, damage}`, every amount `omitempty`. It rides wherever `PurposeView` already does (the card, `ModeOptionView`, `AlternativeCostView`, `ActivatedAbilityView`), and it is projected and cleared with it. Clients ignore it. `docs/protocol.md` documents it, and `client/src/lib/protocol.ts` mirrors the type.
- **No move change.** `targets[].slot` and `targets[].mode` are on the wire already. The heuristic's private `targetRef` (`params.go:20`) gains the two fields it currently drops.
- **`game.Purpose` stays comparable.** `IsZero` compares it with `==` (`game/purpose.go:148`), so the new field is a pointer, as `DiscardPayoff` and `Pump` are. `plus` (a fused split spell) concatenates the two halves' lists, renumbering the second half's slots past the first's.
- **`damage_to_creature` stays.** It describes a triggered or activated row whose target is chosen later (exert's Glorybringer), where there is no move target to match. A `targets` entry describes a target the move names.
- **The registration guard** refuses:
  - an entry whose `slot` is not a target clause of its statement;
  - a player amount (`draws`, `discards`, `tokens`, `life_gain`, `life_loss`) on a clause that cannot target a player;
  - `damage` on a clause that can target nothing damage can be dealt to;
  - a negative amount;
  - an entry that says nothing.

### Delivery plan

| # | PR | Depends on | Acceptance |
|---|---|---|---|
| 1 | This amendment | — | docsguard |
| 2 | **The signal.** `game.Purpose`'s target list, `TargetPurposeView`, the projection, the guard, `docs/protocol.md` and `protocol.ts`. Declarations for the curated decks' cards in both classes: Prismari Command, Sign in Blood, Lightning Bolt, Shock, Arc Trail, Fiery Temper, Abrade, Izzet Charm's damage mode, and Blaze if X can be declared (otherwise it goes on `noPrintedAmount`). `TestCuratedDeckPurposes` requires an entry for each and drops their `valueIsTheirTarget` notes. The dump audit lists catalog cards whose text reads "target player draws / creates / gains", or "deals N damage to" a creature or any target, with no entry. **No price change.** Touches `effects`, `game`, `protocol`, `decks` and the client types, and nothing under `aiseat/`. | 1 | Unit tests for the guard and the projection. `TestCuratedDeckPurposes`. §8 run 1 and run 2 identical to `develop` apart from IDs and timings, as PR 6 showed. |
| 3 | **Target purposes priced** (A1, B1) behind `PriceTargetPurposes`. `targetRef` decodes `slot` and `mode`. A declared entry replaces `targetsValue`'s price for its pick, and a cast whose only declared amounts are target entries counts as purpose-priced, so the mana proxy goes. | 2 | Unit tests: Prismari's loot is worth more on the bot than on an opponent, its Treasure likewise, and Sign in Blood at itself beats an opponent on 30 life. §8 run 1 and run 2 under §8's sub-PR bar: no tag's agreement falls, and `heuristic`'s upper bound stays above 25%. A targeted run, `boteval arena --seats heuristic,heuristic,heuristic-baseline,heuristic-baseline --decks izzet-aggro,mono-black-aristocrats,izzet-aggro,mono-black-aristocrats --games 96 --rotate --lockstep`, reporting Prismari's and Sign in Blood's Cards rows, with who each was aimed at, read from the decision logs. |
| 4 | **Damage by whether it kills** (C1, D1) behind `DamageByLethality`. `cardTargetValue` takes the declared amount; the kill test is shared with `combat.go`'s. | 2 (3 for the acceptance position) | Unit tests: Shock at a 2/4 is about 0, at a 2/2 it is removal, at an indestructible 2/2 it is 0, at a 3/3 with 1 damage marked it is removal, and at a player on 2 life it gets the lethal bonus. §8 run 1 and run 2 as in PR 3. The izzet-aggro run, reporting burn aimed at creatures it killed, at creatures it did not, and at players. |
| 5 | **The catalog sweep** (owner answer 5): every catalog spell, mode or row that gives its target player something (a draw, a token, a life gain or a loot), then burn from the dump audit's list, in card batches with no price change. Each batch follows docs/adding-cards.md (its own oracle fixtures only, Completeness unchanged). | 2 | The dump audit's list shrinks. The real-dump audits pass. |

Suite positions are **proposals only**, for the owner to review in the PR that adds them, as every suite position is:

- **`prismari-draw-step-with-a-three-drop`** (PR 4, the issue's position):
  - Window: the bot's draw step; 3 lands untapped; Prismari Command in hand; the opponent has a 2/4; a castable 3-drop is in hand.
  - Accept: pass, or Prismari "loot and Treasure, both at the bot" once #2681 offers it.
  - Reject: any Prismari cast aimed at the 2/4 and the opponent.
- **`prismari-loot-yourself`** (PR 3):
  - Window: the bot's own second main phase; the opponent has no creatures; the bot has 3 lands and dead 5- and 7-drops in hand.
  - Accept: Prismari "loot and Treasure, both at the bot".
  - Reject: any selection that loots the opponent or gives them the Treasure.
- **`do-not-shock-the-two-four`** (PR 4):
  - Window: the bot's main phase; Shock in hand; the opponent's only creature is a 2/4.
  - Accept: pass, or Shock at the opponent.
  - Reject: Shock at the 2/4.
- **`sign-in-blood-yourself`** (PR 3):
  - Window: the bot's main phase; Sign in Blood in hand; the bot on 30 life; the opponent on 30.
  - Accept: Sign in Blood at the bot.
  - Reject: Sign in Blood at the opponent.

**Out of this amendment:** the draw-step gate (#2457), Prismari's missing selection and the labels that don't name the modes (#2681), a trigger whose target is picked later through `choices.go:194` (a follow-up once PR 3 shows the shape), wither and infect sources (CR 120.3d; no curated spell has either, so they keep today's price), and damage priced together with a combat this turn.

### Open questions for the owner (2026-10-08)

Each question lists the recommended option first.

1. **Q1. Where a target's amounts are declared.**
   - **(a) Recommended:** per target clause, keyed by `slot`, in a `targets` list on `Purpose` and `PurposeView`. Arc Trail's two clauses and Prismari's two modes are each described truthfully, and the move's existing `slot`/`mode` find the entry.
   - **(b)** A mode-level flag, "these amounts apply to the mode's target player", plus a mode-level `damage`. Smaller. Arc Trail and other two-clause statements stay undeclared at today's price.
   - **(c)** Inferred from the clause kind. This breaks §6's "declared, never inferred", and it is wrong for "target opponent sacrifices…, you draw".
2. **Q2. How a gift to an opponent is priced.**
   - **(a) Recommended:** the bot's own amount formula, turned into a change in that seat's strength and priced through `ScoreEval`'s opposition weights, as §4 prices a sweep. A gift to the strongest of three opponents costs 0.83 of its value, and a gift to another opponent 0.33.
   - **(b)** A sign flip times `LeaderBoost`: 1.5 for the leader, 1.0 for anyone else, at any table size. It is simpler, and it overcharges gifts at a four-player table. At two seats the two options agree.
3. **Q3. Damage that does not kill a creature.**
   - **(a) Recommended:** `DamageChip` 0.00, so it is worth nothing, because marked damage is removed in cleanup (CR 514.2).
   - **(b)** A small share, `DamageChip` 0.10 of the removal value, as a tie-break toward the bigger creature. Prismari's 2 at Y'shtola would then be 0.67, not 0.
   - **(c)** In proportion to damage over toughness. 2 at Y'shtola would be 3.36, still enough for a bad cast in a main phase.
4. **Q4. Declared damage at a player.**
   - **(a) Recommended:** per point at `DamageToOpponent` (0.30), through Q2's weights, with `LethalBonus` when the amount reaches the player's life (CR 704.5a). This replaces the `FinishLife` bet for declared spells only.
   - **(b)** Keep the flat `DamageToPlayer × leaderBoost` (1.80 for any amount). At seq 248, "2 to Claude and loot the bot" then prices 1.80, above `InstantThreshold`, and the draw-step cast stays.
5. **Q5. Which cards declare target purposes.**
   - **(a) Recommended:** the curated decks' cards in both classes (PR 2), then every catalog spell, mode or row that gives its target player something: a draw, a token, a life gain or a loot. Those are the sign errors (PR 5), about 40 modal files plus the non-modal "target player draws" spells. Burn across the catalog follows in batches from the dump audit. Undeclared burn keeps today's removal price.
   - **(b)** The curated decks only. Every other card keeps today's price.
   - **(c)** Every catalog card in both classes now: about 280 damage files and the player-target files, in PR 5's batches before PR 4 is measured.
6. **Q6. Planeswalker and battle targets of declared damage.**
   - **(a) Recommended:** a planeswalker is priced by the share of loyalty removed, and as killed at or above its loyalty (CR 120.3c, 704.5i). A battle keeps today's price.
   - **(b)** Both keep today's price, as removal whatever the amount. Smaller, and 2 damage to a 6-loyalty planeswalker stays priced as killing it.

#### Owner answers (2026-10-08)

The owner chose option (a), the recommended one, on every question. No section above changed. These answers bind the delivery PRs.

1. **Per target clause.** `Purpose` and `PurposeView` gain a `targets` list keyed by the clause's `slot` (A1).
2. **Through the table's weights.** A gift to an opponent is that seat's strength change, priced through `ScoreEval`'s `OpponentMean` and `OpponentMax` (B1).
3. **`DamageChip` 0.00.** Damage that does not kill a creature is worth nothing (C1; CR 514.2).
4. **Per point, plus lethal.** Declared damage at a player is priced at `DamageToOpponent` per point through those weights, with `LethalBonus` when it reaches the player's life (D1; CR 704.5a).
5. **The curated decks, then gifts.** PR 2 declares the curated decks' cards in both classes. PR 5 then declares every catalog spell, mode or row that gives its target player something, and burn follows from the dump audit. Undeclared burn keeps today's removal price.
6. **Planeswalkers by loyalty.** A planeswalker is priced by the share of loyalty removed, and as killed at or above its loyalty (CR 120.3c, 704.5i). A battle keeps today's price.

## Amendment 2026-10-08: what removal hands back

**Status:** Accepted (owner decision on #2753, 2026-10-08).
**Issue:** [#2679](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2679). PR: #2753.
**Amends:** the target half of the cast price, and the [amendment of 2026-10-08](#amendment-2026-10-08-purposes-that-follow-a-modes-target-and-damage-priced-by-whether-it-kills)'s target entries, which gain one field. Everything else is unchanged.

### The problem

A spell aimed at an opposing permanent is priced as removal, at the permanent's value × `RemovalConfidence` × `leaderBoost` (`cardTargetValue`, `damageCardValue`). Two kinds of removal give something back, and that price saw neither of them.

- **A commander comes back.** CR 903.9a: a commander put into a graveyard or exile may go to the command zone. CR 903.9b: one that would go to a hand or a library may go there instead. CR 903.8: each later cast from there costs {2} more. In review game 2 (`06e98afa`, seq 186), Chaos Warp on Y'shtola, a 2/4 commander, priced at 6.65 and beat casting the bot's own commander (4.17).
- **Some removal gives its target's controller a gift.** Rapid Hybridization, Pongify, Beast Within and Generous Gift give a 3/3. Stroke of Midnight gives a 1/1. Swords to Plowshares gives life equal to the target's power. Path to Exile and Assassin's Trophy give a basic land. In review game 1 (`8a9f18d7`, seq 66), Rapid Hybridization on a 1/1 priced at 3.45 and handed back a 3/3.

### Why the per-target entries cannot say it

The previous amendment's `targets` entries (#2689) describe what happens **to the target**. Their player amounts (`draws`, `discards`, `tokens`, `life_gain`, `life_loss`) belong to the player the clause picks. The registration guard refuses a player amount on a clause that cannot target a player. Removal's clause targets a creature or a permanent, and the gift goes to that permanent's **controller**, who is not a target. So `tokens: 1` on Rapid Hybridization's creature clause is refused, and it would mean the wrong thing if it were accepted. Inferring the gift from the card is ruled out by §6 ("declared, never inferred").

### Decision

**1. A declared return on the target entry.** `game.TargetPurpose` gains `Returns`, a `game.TargetReturn` of printed amounts:

| Field | Wire (`targets[].returns`) | Meaning | Example |
|---|---|---|---|
| `CreatureTokens` | `creature_tokens` | creature tokens the target's controller creates | Rapid Hybridization 1 |
| `TokenPower`, `TokenToughness` | `token_power`, `token_toughness` | each token's printed size | 3 and 3; Stroke of Midnight 1 and 1 |
| `LifeEqualToPower` | `life_equal_to_power` | the controller gains life equal to the target's power, counted at resolution, so a flag rather than a number | Swords to Plowshares |
| `Lands` | `lands` | land cards the controller may put onto the battlefield | Path to Exile, Assassin's Trophy 1 |
| `LandsUntapped` | `lands_untapped` | how many of those enter untapped | Assassin's Trophy 1, Path to Exile 0 |

It is additive on the wire, with every field omitted when zero, and there is no snapshot change. `docs/protocol.md` documents it and `client/src/lib/protocol.ts` mirrors it. An entry that holds only a return is not empty (`TargetPurpose.IsZero`).

**The guard** (`checkTargetPurposes`) refuses:

- a return on a clause that cannot target a permanent;
- a negative amount;
- creature tokens with no printed toughness, or a token size with no tokens;
- more untapped lands than lands.

`effects.RemovalReturning(slot, r)` builds the entry. It is declared on the eight cards listed above.

**Chaos Warp declares nothing.** What it gives back is whatever the owner reveals from the top of their library, and no printed amount says that.

**2. The price** (`NetRemoval`; on in `DefaultConfig`, off in `BaselineConfig`; `net_removal.go`).

- **An opposing commander**, owned by its opponent, is priced at `min(its value, 2 × CommanderTax + DamageToOpponent × power)` before `RemovalConfidence` and `leaderBoost`.
  - `2 × CommanderTax` is the tax counted for its {2}.
  - `DamageToOpponent × power` is the one turn of damage it does not deal while it is away.
  - This applies on the default removal path, and to declared damage that kills it or removes all of a planeswalker's loyalty.
  - It needs no catalog data, because `is_commander` is on the wire.
  - A commander the bot owns but an opponent controls keeps its full price, since it comes back to the bot.
- **A declared return** is valued the way the target's controller would value it:
  - a token at the `CreatureValue` of its printed body;
  - life at what `Strength` counts it;
  - a land at `ManaSource`.

  It is taken off the removal on the removal's own scale (× `RemovalConfidence` × `leaderBoost`), so the move is worth the removal less the gift. Only a target an opponent controls is netted. A return does not make the move purpose-priced, so the mana-value proxy stays.

**Worked:**

- Game 1 seq 66: Rapid Hybridization on the 1/1 goes from 3.45 to −1.77, and the bot passes.
- Game 2 seq 186: Chaos Warp on Y'shtola goes from 6.65 to 3.72, and the bot casts Mary Read and Anne Bonny (4.17).

### Measured

The run tables are under [Measurements, #2679](#2679-removal-priced-net-of-what-comes-back-2026-10-08).

- Run 2's seeds 1, 1001 and 2001 pooled are unchanged for `heuristic` (80 of 288 before and after).
- **Seed 2001** is the one seed where `heuristic` finished behind the baseline after the change (21 vs 27; it was 24 vs 24). Two more seeds put `heuristic` 6 and 2 games up. Over five seeds it went from 135 of 480 to 143 of 480 (28.1% to 29.8%, intervals overlapping). The owner reads the seed 2001 dip as noise against the five-seed pool.
- The counters it targets went to near zero:
  - gift removal cast on a target worth less than its gift: 23 to 0 in run 1, and 49 to 0 in run 2;
  - the listed removal on an opposing commander: 59 to 13, and 47 to 2.

### Open

- **Sweeps and edicts** still price a commander at its full body. A wipe's `ScoreEval` and an untargeted sacrifice do not read `is_commander`.
- **A non-removal targeted effect on an opposing commander** also gets the net price on the default path, because without an entry the heuristic cannot tell a destroy from a Pacifism-style aura or a tap. Such an effect keeps the commander on the battlefield, so the net price undervalues it. A target entry that says "this does not remove" is the fix, if a curated card needs it.

## Amendment 2026-10-12: the race sends the swing that ends the game soonest

**Status:** Accepted (owner decisions 2026-10-12). The owner accepted the objective and the commander tax, and kept the other two refinements as off switches; see [Owner decisions (2026-10-12)](#owner-decisions-2026-10-12) at the end of this amendment.
**Issue:** [#2400](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2400), filed by #2310 (PR #2399). PR: #3063. Sprint S67. Tracker #890.
**Amends:** the combat planner, which [Out of scope](#out-of-scope) left alone. The [#2675 knobs](#2675-2690-2676-combat-priced-by-what-combat-does-2026-10-08) set the precedent. The two planners it changes are the two-turn race (`raceWith`, `heuristic/race.go`) and the attrition horizon (`attritionLine`, `heuristic/attrition.go`). Neither has an ADR of its own: `docs/bot.md` documents them. Each change is a `Config` switch, and `BaselineConfig()` turns all of them off (§9).

### The problem

Several swings can qualify as a two-turn race, or as the first exchange of an attrition line. Both planners sent the **smallest**: the shortest prefix of the evasive-first attack order that qualified. Because of that, a more accurate count of the defender's free blocks made games longer, not shorter. A count that let a smaller swing qualify sent that smaller swing, and it was the bigger one that ended the game.

#2310 measured three refinements to the race's model of the defender and left all three out for this reason. All three are true to the rules:

1. A trampler's overflow past a free chump gets through (CR 702.19b).
2. The defender's commander is a free blocker only while its controller can pay the tax to cast it again (CR 903.8).
3. A commander that chumps under blockToSurvive's answer is back to block next turn (CR 903.9a). This one stalled both Wurm lockstep boards at the time, the failure #1548 hit.

### Decision

**1. The objective** (`RaceEndsSoonest`; on in `DefaultConfig`, off in `BaselineConfig`). Both planners try every prefix that qualifies and send the one that ends the game soonest. They rank on three keys (`swingKey`, `race.go`):

1. **The fewest turns to the kill the model sees.** For a race this is two, or one when the swing is lethal now. For an attrition line it is the number of exchanges before the race that ends it.
2. **The most damage the defender cannot avoid this turn.** This is the least of blockToSurvive's damage and the free answer's (`raceEval.sure`), the attrition horizon's `sure` damage.
3. **The smallest swing.** A tie keeps the swing tried first, so the reserve the crack-back check counted on is spent only when it buys something.

Inside the attrition search, the deeper levels only ask whether some line exists, so they keep the first one they find. The budget of 100 swings per decision is unchanged. When the budget runs out at the top level, the planner sends the best line found so far.

The objective is checked again at each declaration. A partly declared swing is a prefix of the same order, and a tie keeps the smaller swing, so the bot declares the plan one creature at a time and passes when it is complete (`TestRaceSendsTheSwingThatEndsSoonest`).

**2. The commander tax** (`FreeCommanderNeedsTax`; on in `DefaultConfig`, off in `BaselineConfig`). The defender's commander is a free blocker only while its mana sources, tapped or not, make the commander's mana value plus {2} for each earlier cast from the command zone (`canRecast`). The commander's own mana ability is not counted, since it is the card that died. The cast count is the public `commander_casts`, which the engine re-keys when the commander changes zones.

**3. Trample overflow past a free block** (`FreeTrampleOverflow`; off in both). The free answer's damage is the larger of the matching over the free blocks and `trampleBound` over the free blocks (`trampleBoundBy`). Both are lower bounds, as for `unblockedPower`.

**4. The recast chump** (`RecastChumpForNext`; off in both). Under blockToSurvive's answer, a defender's commander that dies blocking is back to block next turn. It is also out of the crack-back unless it has haste, the same split the free answer already makes (`recastCommanders`).

Switches 3 and 4 are built, tested and measured, and left off because they measured no better (below).

### Measured

The tables, and every seed that changes, are under [Measurements, #2400](#2400-the-race-sends-the-swing-that-ends-the-game-soonest-2026-10-12). In short, at `develop` `de452ebd6`, from non-race lockstep runs:

| | 4-seat mean | 2-seat mean | Wurm + Ogre | full Wurm |
|---|---:|---:|---|---|
| `develop` (smallest qualifying swing) | 16.08 | 15.08 | 5 / 6 | stalls (#2474) |
| the objective alone | 15.42 | 14.17 | 5 / 6 | stalls (#2474) |
| **the objective and the commander tax (the default)** | **15.25** | **13.75** | 5 / 6 | stalls (#2474) |

- **The objective alone shortens both gates**: −0.67 turns on the 4-seat gate and −0.92 on the 2-seat one. The objective fixes what #2310 found. With the old objective, every refinement left the gate as long or longer, as #2310 measured. With the new one, no refinement makes both gates longer.
- **The commander tax** is the one refinement that shortens both gates on top of the objective.
- **Trample overflow** shortens the 2-seat gate (13.67) and lengthens the 4-seat one (15.92).
- **The recast chump** changes nothing alone. With the tax it makes both gates longer (15.92 and 14.25).
- **Neither Wurm board stalls on the recast chump.** The Wurm-and-Ogre board ends on turns 5 and 8 under the old objective and on 5 and 6 under the new one. The full mirror stops at lives `[6 3]` with or without it. That is #2474's board, and it stops there on `develop` and on every variant.

### Options considered and rejected

- **Fewest turns alone.** Every race the model sees kills in two turns, unless the swing is lethal now, and the all-in check (`lethalPush`) usually takes those first. So "fewest turns" does not separate the swings that race. The damage the defender cannot avoid is what separates them.
- **Creatures traded as a third key**, preferring the exchange that kills more. It moved no gate seed. In the full Wurm mirror it sent both Wurms into their Wurm and an Ogre, a "2 for 2". The edge left was an Ogre, and the defender's commander traded it away for free, so the board stopped on `[6 6]` instead of `[6 3]`.
- **Turning all three refinements on**, as #2400 suggested once the objective changed. With all three the means are 15.92 and 13.75, behind the tax alone on the 4-seat gate.

### Open (#2400)

- **The objective ranks swings against one defender.** Both planners still try the opponents in focus order, focus first, and take the first one with a qualifying swing. At a 4-seat table the bot can send a three-exchange line at the focus seat while a one-exchange line waits against another seat (gate seeds 106 and 109). Ranking across defenders would override the focus order, the threat ranking the rest of the attack decision uses, so it is not part of this amendment.
- **#2474 is unchanged.** The full Wurm mirror stops on Bear and commander against a lone commander on 3 under every variant. That kill takes a commander trade and then two more turns, and no ranking of the swings that qualify can send a swing that does not qualify.

### Owner decisions (2026-10-12)

The owner accepted the amendment on 2026-10-12 and took the recommendation on all three questions. No section above changed.

1. **The objective is accepted.** `RaceEndsSoonest` is on in `DefaultConfig()`. It replaces "the smallest swing that qualifies" in the race and in the attrition horizon.
2. **The commander tax is accepted.** `FreeCommanderNeedsTax` is on in `DefaultConfig()`.
3. **Trample overflow and the recast chump stay as off switches.** `FreeTrampleOverflow` and `RecastChumpForNext` are kept, off in both `DefaultConfig()` and `BaselineConfig()`. #2474's board, a commander trade three turns deep, is where the recast question comes back.

## Amendment 2026-10-12: cycling cards as cheaper discards, and the bot's own draw step

**Status:** Accepted (owner answers 2026-10-12). The owner chose the recommended option, (a), on all three questions; see [Owner answers (2026-10-12)](#owner-answers-2026-10-12) at the end of this amendment. The code is in the PR that adds this section, behind two switches that are on in `DefaultConfig()` and off in `BaselineConfig()`, so either can be turned off without a revert.
**Issue:** [#2457](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2457), from the S66 label review of `do-not-windfall-away-the-last-land`, and the owner's comment on it of 2026-10-08 about the draw step.
**Amends:** §5 (which windows the bot spends in) and §7's discard half, with the [discard-payoff amendment](#amendment-2026-10-06-discard-payoffs)'s reach: wherever the policy prices a discard of its own card.

### The problem

**The position.** `do-not-windfall-away-the-last-land`: the bot's own draw step, round five, four lands, Mary Read and Anne Bonny on the battlefield. In hand: Unexpected Windfall, Windfall, Marauding Mako, Arcane Denial, Impulsive Pilferer and Myriad Landscape, the only land. The owner accepted two answers: pass, and Unexpected Windfall discarding the Mako ("if going to pay 2 to cycle mako might as well unexpected windfall the mako instead"). Discarding the Landscape is rejected.

On `develop` at `de452ebd6` the heuristic passes. Its prices: Unexpected Windfall discarding Arcane Denial +1.00, discarding the Mako or the Pilferer +0.65, Windfall +0.40, the Landscape −0.30; Mary Read's loot +0.80; cycling the Mako +0.50. Everything is under `InstantThreshold` (1.50), because the draw step is not one of §5's windows.

**Why the Mako is not the cheapest discard.** §7 charges `cardValue`. The Mako is a 1/1 with a triggered row, 2.05; Arcane Denial is an instant of mana value 2, 1.20. Mary Read pays a Treasure for the Mako, a Pirate (+0.50), and the Mako is still the dearer discard. Nothing on the wire said the Mako can turn itself into a card: its cycling row declared only `answers: ["value"]` ([ADR 0142](0142-declared-answers-on-catalog-abilities.md) S0), which has no amount.

**Why the draw step is not a window.** The owner on #2457 (2026-10-08): §5 assumes the spend "costs nothing the bot would otherwise keep", and that is false in the bot's own upkeep and draw step. A land tapped there stays tapped through both of its main phases. Review game 2 (`06e98afa`, seq 248) is the cost: Prismari Command cast in the draw step with all three lands, and nothing left for Mary Read and Anne Bonny in the main phase. The owner's gate, if the draw step is ever made a window: "an instant cast before the bot's first main phase, with an empty stack, has to beat the best main-phase use of the same mana. Otherwise it waits for main."

### Decision

**1. The signal.** The cycling helpers declare what the row turns the card into: `Cycling` declares `draws: 1`, and `Typecycling` (with `BasicLandcycling`) and `Transmute` declare `tutors: 1`. Every catalog card that uses them gets it, with no card file touched. It rides `zone_abilities[].purpose`, which only the card's owner is sent. Additive on the wire, no snapshot change. It does not change how cycling itself is priced: the heuristic reads a row's purpose for the battlefield only, so cycling stays `ActivateBase`.

**2. The price** (`SelfReplacingDiscards`, `aiseat/heuristic/self_replacing.go`). A nonland card whose hand row discards the card itself (`discard_self`), costs nothing else but mana, and declares a draw or a tutor is a cheaper discard:

```
replacement = Weights.Hand × (draws + TutorWeight × tutors)
spared      = SpellPerMana × the row's mana value
discard     = min(price, max(replacement, price − spared))
```

`price` is what the discard would cost without it: `cardValue` for a cost (§7), `handKeepValue` for a discard the bot chooses (#2691: cleanup, discard prompts, a loot's named cards, the discard-payoff estimate, a put from hand). The discard spares the mana cycling the card would cost, which is the owner's "you would pay {2} to cycle it anyway", on the policy's own value-per-mana proxy. It is never cheaper than the card cycling would find, which the bot gives up with it. A land is left alone: it is a land drop first, and `LastLandDiscard` and `DiscardLandFloor` already price that. A row whose cost asks for anything else (life, a sacrifice, other cards) is a different trade and is left alone.

In the position the Mako's discard falls from 2.05 to 1.20, so Unexpected Windfall discarding it is +1.50, above discarding Arcane Denial (+1.00). A 6/6 with cycling {2}{G} keeps its price less 1.80.

This is a valuation, not a consequence of the bot's other prices. By those, casting the Mako in a main phase (+0.65) beats cycling it (+0.50), so the bot would not cycle it, and a card that can also be cycled is, strictly, worth at least as much as the same card without. The rule takes the owner's reading of a cheap body with cycling: it is the deck's fodder. The owner confirmed that reading ([Q1](#questions-for-the-owner-answered-2026-10-12)).

**3. The draw step is not a spend window** (`WaitForMain`, `aiseat/heuristic/waitmain.go`). §5 keeps its two windows. Before its first main phase the premise is the other way round, so the bot does not spend there more readily; it spends there less. The owner's gate is applied as written. In the bot's own upkeep or draw step, with an empty stack, an instant-speed cast waits for the main phase. The same cast is still on offer there, beside every sorcery-speed use of the same mana, so the best main-phase use of that mana is never worth less than the cast is now. The reason reads `wait for the main phase: …`. A cast with something on the stack to answer is not held, and neither is an activated ability: the gate is for casts.

Without the gate the change in 2 would decide this position on a rounding error. +1.50 is `InstantThreshold` exactly, the sum came out at 1.5000000000000002, and the bot cast Unexpected Windfall in its draw step. With the gate it passes. Re-decided as a main phase on the same view, it casts Unexpected Windfall discarding the Mako.

**4. The label stays.** The issue asked for the Mako discard to become the only accepted move once both halves land: a cheap cycling discard and a draw step that is a spend window. Under this amendment the draw step is not a spend window, and the gate's answer in the draw step is pass, so the label keeps both answers ([Q3](#questions-for-the-owner-answered-2026-10-12)). `boteval suite run --policy heuristic` is 41 of 41 before and after. The only change in any position is this one's reason, from `nothing worth doing` to `wait for the main phase: cast spell (declared purpose) Cast Unexpected Windfall (+1.50)`.

The measurements are under [Measurements, #2457](#2457-self-replacing-discards-and-waiting-for-the-main-phase-2026-10-12).

### Questions for the owner (answered 2026-10-12)

Each lists the recommended option first.

1. **Q1. The price of a card that cycles.**
   - **(a) Recommended:** decision 2's rule: the cycling mana comes off, never below one card. A 1/1 with cycling costs one card; a big body keeps its price less the cycling mana.
   - **(b)** A flat cap at one card, `min(price, Hand × cards)`. Simpler, and it prices a castable bomb with cycling as one card, so a Windfall would pitch it.
   - **(c)** The signal only, no price. The Mako stays the dearer discard, and the label keeps its two answers for good.
2. **Q2. The bot's own upkeep and draw step.**
   - **(a) Recommended:** not a spend window, with the owner's gate on (`WaitForMain`). Run 1 had 530 such casts on `develop`, mostly removal and card draw; they move to the main phase.
   - **(b)** Not a spend window, and no gate. The bot keeps casting any instant priced over `InstantThreshold` there (the seq 248 shape), and this position's pick rests on the rounding in decision 3.
   - **(c)** A spend window (`LeftoverThreshold`) with the gate. The gate holds every cast such a window would let through, so this changes nothing over (a) for casts and lowers the bar for activations, the one kind the gate does not hold.
3. **Q3. The position's label.**
   - **(a) Recommended:** keep pass and the Mako discard, both accepted. The Mako discard is the right card; pass is the right timing.
   - **(b)** Pass only, which pins the gate, with a main-phase position for the Mako discard harvested from a later arena run.
   - **(c)** The Mako discard only, as the issue first asked. It contradicts the gate, so it needs Q2 (b).

#### Owner answers (2026-10-12)

The owner chose option (a), the recommended one, on every question. No section above changed. These answers are binding.

1. **The price as built.** A nonland card that can cycle into a card costs the cycling mana less as a discard, never less than one card (decision 2, `SelfReplacingDiscards`).
2. **The draw step is not a spend window.** §5 keeps its two windows, and the owner's gate stays on: an instant cast in the bot's own upkeep or draw step, with an empty stack, waits for the main phase (decision 3, `WaitForMain`).
3. **The label stays.** `do-not-windfall-away-the-last-land` keeps pass and the Mako discard, both accepted (decision 4).

## Amendment 2026-10-12: the attrition horizon trades commanders

**Status:** Accepted (owner decisions 2026-10-12). The owner accepted the amendment with both switches on, and had the defender's misblock filed as [#3066](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3066); see [Owner decisions (#2474, 2026-10-12)](#owner-decisions-2474-2026-10-12) at the end of this amendment.
**Issue:** [#2474](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2474), filed by #2472 (the #2462 fix). PR: #3065. Sprint S67. Tracker #890.
**Amends:** the attrition horizon (`planAttrition`, `attritionLine` and `racesFrom`, `heuristic/attrition.go`), which the [amendment of 2026-10-12](#amendment-2026-10-12-the-race-sends-the-swing-that-ends-the-game-soonest) also changed. The two-turn race does not change. Each change is a `Config` switch, on in `DefaultConfig()` and off in `BaselineConfig()` (§9).

### The problem

The full seed-1409 Wurm mirror stops on this board, under `develop` and every variant #2400 measured:

- **The edge seat:** a Bear 2/2 and its 3/3 commander, on 6 life.
- **The other seat:** its 3/3 commander, on 3 life.
- **Both seats:** 37 lands.

A player sees the kill. Swing both. The defender has to block the commander with its own, or take 5. Both commanders die and the Bear puts the defender on 1. Both are cast again, and the edge seat's is summoning sick (CR 302.6), so the edge seat waits a turn. Then the Bear and the commander swing into one blocker, and one of them gets through. On `develop` nobody attacks for the last 20 turns, and the game stops at the turn budget with lives `[6 3]`. #2472 skipped the board in `TestHeuristicWurmEdgeResolvesInsideTheTurnBudget` until this was fixed.

No two-turn race sees the kill, and the attrition horizon refused the first swing for two reasons:

1. **The defender's model only guessed.** `blockToSurvive` first takes every block the defender survives. Their commander on the Bear lives, so it takes that one. Then nothing is left for the bot's commander, 3 still connect, and the defender is dead on 3. The model cannot find a way to live from there, so it falls back to a guess (`defence.guessed`): every attacker blocked and dead, nothing of the defender's lost. An exchange counted that way is never acceptable.
2. **The bot's own commander counted as lost.** The horizon already counted the defender's commander as cast again (CR 903.9a): it is not a creature the defender loses, and it is back on the next board. The bot's commander was a creature lost for good, "because the search may not rely on recasting it". So commander for commander was one of mine for none of theirs.

### Decision

**1. The free answer where the model guessed** (`HorizonPlaysFreeAnswer`; on in `DefaultConfig`, off in `BaselineConfig`). Where `blockToSurvive` only guessed and the defender survives its free answer (#2310), the attrition line plays the free answer forward instead of the guess (`outcomeOf`). The free answer blocks with everything whose block costs the defender nothing, so on this board it is commander for commander and the Bear unblocked. The guess stood in for a way to live the greedy missed. The free answer is a way to live the model can name, and it loses the defender nothing for good. The sure damage does not change: it is still the least of the two answers' damage.

**2. The bot's commander cast again** (`HorizonRecastsCommander`; on in `DefaultConfig`, off in `BaselineConfig`). In a second pass of the search, a commander of the bot's that dies in an exchange is cast again (CR 903.9a) when the bot's mana sources, tapped or not, make its mana value plus {2} for each earlier cast (CR 903.8, `canRecast`, the same test the defender's commander gets under `FreeCommanderNeedsTax`).

- **It is not counted lost.** An exchange that kills it counts it as cast again, not as one of mine, and the reason says so: `attack: attrition — 0 of mine (1 cast again) for 0 of theirs and 2 sure now, a race within 2 exchange(s); crack-back 3 < my 6`.
- **It is back on the next board, summoning sick** (`afterExchange`). It can block, and it counts toward a race's next turn, but it is not an attacker on that board (`simAttackers`). One with haste can attack (CR 702.10). Its cast count goes up by one, so a second trade in the same line pays {2} more.
- **The line may wait a turn for it** (`racesFrom`, `afterWait`). When a board with a summoning-sick commander of mine has no lethal all-in, no race and no exchange, the line may pass that turn. The crack-back on it is every opponent swinging everything into every creature the bot has, the race's pessimistic count, and it must be less than the bot's life. After it, nothing is summoning sick. The waiting turn counts as an exchange, both for the horizon of three and for the turns key of `RaceEndsSoonest`.
- **It is a second pass, with a budget of its own**, and it runs only when the first pass finds no line (`planAttrition`). The first pass is the search as it was. So the switch changes no decision the horizon made before, and only adds a line where there was none. The worst case is two passes of 100 swings each.

On the issue's board, swinging both is then acceptable: 0 of mine (1 cast again) for 0 of theirs, 2 sure, crack-back 3 < 6. The next board has the defender on 1 and the bot's commander summoning sick. The Bear alone has nothing, so the line waits a turn, with a crack-back of 0, and then the Bear and the commander are lethal into one blocker. That is a line within 2 exchanges.

The two switches need each other here. With the free answer alone the bot's commander is still one of mine for none of theirs. With the recast alone the guess still counts the Bear dead as well.

### Compared with the recast chump

`RecastChumpForNext` (#2400, off) is the same question from the defender's side, inside the two-turn race: a defender's commander that dies chump-blocking under `blockToSurvive`'s answer is back to block next turn. The attrition horizon already counted the defender's commander that way. This amendment is the bot's side, and only inside the horizon. The race does not need it: a commander the bot casts again is summoning sick on the race's next turn, so the race is right to count it out of next. That is why the horizon has to be able to wait a turn.

The two are measured together below. The recast chump makes the same two gate seeds longer with this amendment as without it (106 and 211), and the full Wurm mirror ends on the same turns either way. It stays off.

### Measured

The tables, and every seed that changes, are under [Measurements, #2474](#2474-the-attrition-horizon-trades-commanders-2026-10-12). In short, at `develop` `2594c7de8`, from non-race lockstep runs:

| | 4-seat mean | 2-seat mean | Wurm + Ogre | full Wurm |
|---|---:|---:|---|---|
| `develop` | 15.25 | 13.75 | 5 / 6 | 26 / 26, stalls at `[6 3]` |
| both switches, the trade in the one pass | 15.92 | 13.75 | 5 / 6 | 7 / 6 |
| **both switches, the trade in a second pass (the default)** | **15.25** | **13.75** | 5 / 6 | **7 / 6** |

- **The default cashes the full Wurm mirror** on turn 7 with the edge on seat 0 and turn 6 with it on seat 1, and the board is no longer skipped.
- **It changes no gate result.** At `2594c7de8` all 24 gate games play move for move as on `develop`: wherever the second pass ran, it found no line. Rebased onto `032e1f138` (#3064), every gate game ends on the same turn with the same winner as on that `develop`, and one decision in one game (seed 105) takes a commander-trade line where `develop` passed.
- **With the trade in the one pass, the 4-seat gate is 0.67 turns longer.** Seeds 101, 103 and 111 lose a line the search used to find: more exchanges were acceptable at every depth, and the budget of 100 swings ran out before the search reached the old line. With the budget at 1000, seeds 101 and 103 end as on `develop`. In seed 109 the objective sends a commander-trade line, 6 sure within 2 exchanges, instead of 4 sure within 3, and the game ends a turn later with another winner.

### Options considered and rejected

- **The trade in the one pass.** It is the simpler search, and it lets the objective rank a commander-trade line against the others. It made the 4-seat gate 0.67 turns longer, mostly by spending the budget (above).
- **A bigger budget.** It would change every decision whose search runs out today, with or without this amendment. That is a separate measurement, not part of cashing one board.
- **Fixing the greedy in `blockToSurvive`** so that it retries from no blocks before it guesses. It would change the two-turn race's numbers on every board where the model guesses, not only the horizon's, so it was not built.
- **The bot's commander back and able to attack on the next board.** That is a turn early, and the line skips the crack-back on the turn the bot really waits.

### Open (#2474)

- **The full mirror ends on a misblock.** In both games the defender, on 3, blocks the Bear with its commander instead of the bot's commander, takes 3 and dies. So the games end on the first swing of the line. The block planner (`decideBlock`) scores each block on its own, and the Bear block scores higher (+6.90). Nothing checks that the blocks together leave the defender alive. With the right block the line takes two more turns. `TestAttritionWaitsForTheCommanderItCastAgain` pins those two turns at the decision, but no whole game in the tree plays them. The misblock is [#3066](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3066).
- **The second pass runs only when the first finds nothing.** A commander-trade line that ends sooner than the line the first pass finds is never sent.
- **The waiting turn is often not needed** ([#3080](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3080), closed by `HorizonRecastsSameTurn`). Once #3066 fixed the misblock, the full Wurm mirror showed the bot casting its commander again in the trade turn's second main phase, so it can attack on the next turn. The horizon now models that when the bot's untapped mana pays the tax after combat, and waits only when it does not; see [#3080's measurements](#3080-the-commander-cast-again-after-combat-2026-10-14).

### Questions for the owner (#2474, answered 2026-10-12)

1. **Accept the amendment, with both switches on in `DefaultConfig()`?** Recommended: yes. It cashes the full Wurm mirror and changes no gate result. The alternative is to keep the board skipped.
2. **Is the misblock its own issue?** Recommended: yes, filed separately, since it is the defender's block planner and not the attack planner. Fixing it changes every game where a defender facing lethal has to choose which attacker to block.

### Owner decisions (#2474, 2026-10-12)

The owner accepted the amendment on 2026-10-12 and took the recommendation on both questions. No section above changed.

1. **The commander-trade horizon is accepted.** `HorizonPlaysFreeAnswer` and `HorizonRecastsCommander` are both on in `DefaultConfig()`, and both off in `BaselineConfig()`. The full Wurm mirror's board is no longer skipped.
2. **The misblock is its own issue:** [#3066](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3066), in S67. It is the defender's block planner (`decideBlock`), not the attack planner, so it is not part of this amendment.

## Amendment 2026-10-12: Izzet and Simic after the rebalance

**Status:** Accepted (owner decisions 2026-10-12). The owner took the impulse price and the Fable and Monument swap, kept Simic and Esper as they are, and had Izzet's rebuild filed as [#3071](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3071) and the landfall payoff as [#3072](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3072); see [Owner decisions (#2480, 2026-10-12)](#owner-decisions-2480-2026-10-12) at the end of this amendment. The code is in the PR that adds this section, as two commits: the price change behind two switches (on in `DefaultConfig()`, off in `BaselineConfig()`), and the Izzet swap.
**Issue:** [#2480](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2480), the follow-up the owner left open when #2436 (PR #2475) merged. Sprint S67.
**Amends:** §6 (a new purpose field, `impulse`), §7 (what a Treasure given up to pay a cost costs), and the izzet-aggro list #2436 set. No other price and no other deck changes.

### The problem

The issue's win table is from 2026-10-07. Since then #3062 made lockstep deterministic, #3063 changed the race objective, #3064 added cheap cycling discards and the wait for the main phase, and #3065 added the commander-trade horizon. Re-measured on `develop` at `0e81b3af4`, with §8's run 1 (64 games, seed 1, `--rotate --lockstep`) and run 2:

| Deck | Run 1 wins of 64 | Win rate | Wilson 95% | Issue (2026-10-07) | `never` |
|---|---:|---:|---|---:|---|
| esper-control | 17 | 26.6% | 17.3%–38.5% | 29 | — |
| izzet-aggro | 1 | 1.6% | 0.3%–8.3% | 6 | Glint-Horn Buccaneer's, Marauding Mako's, Professional Face-Breaker's and Scrounging Skyray's activations |
| mono-black-aristocrats | 26 | 40.6% | 29.5%–52.9% | 21 | Ashnod's Altar, Blood Artist, Burnished Hart's activation |
| simic-ramp | 20 | 31.2% | 21.2%–43.4% | 8 | — |

Run 2: `heuristic` 66 of 192 seat-games, 34.4% (28.0%–41.3%); `heuristic-baseline` 30 of 192. Izzet won 1 of 48 with `heuristic` and 2 of 48 with `heuristic-baseline`.

**Simic is no longer behind** (31% in run 1, 13 of 48 in run 2), and Burnished Hart and Mind Stone have left its `never` list. **Izzet is further behind than the issue says.** It is out first or second in 48 of the 64 games, against 31 for esper-control, 23 for simic-ramp and 17 for mono-black-aristocrats.

### Why the cards go unused

The issue's option 1, worked from run 1's decision logs. Each activation's price in every window it was offered, against what the bot did instead:

| Card (deck) | Offered / taken | Price | Cause |
|---|---|---|---|
| Professional Face-Breaker, "Sacrifice a Treasure: exile the top card, you may play it this turn" (izzet) | 1,070 / 0 | −0.80 in every window, −0.50 with the Face-Breaker tapped | **The heuristic.** The row declared only `answers: ["value"]`, so it was the flat `ActivateBase` (0.50), and the Treasure it sacrifices cost its board value as a mana source (1.00), twice what §6 prices making one at (`TokenWeight`, 0.50). The 0.30 every untapped creature's activation pays did the rest. |
| Scrounging Skyray and Marauding Mako, cycling (izzet) | 121 / 0 and 70 / 0 | +0.50 (`ActivateBase`) | **Neither: the right play.** In the main phase, casting the creature, a discard payoff that grows, beats cycling it. In the upkeep and draw step +0.50 is under `InstantThreshold`. |
| Solphim, Mayhem Dominus, indestructible counter (izzet) | 385 / 2 | −1.1 to −7.7, and refused by the life floor in 81 windows | **Neither: the right play.** It is a `protect` answer (ADR 0142). It was taken in both windows where removal targeted Solphim. Everywhere else it is two cards and four life, or {R}{R}, for a counter nothing threatens. |
| Burnished Hart (simic) | 346 / 1 | −1.60 to −0.50; +1.04 the one time it was taken | **The heuristic, and not a small fix.** Two lands are priced at `ManaSource` each plus the ramp premium while the deficit is open, against the Hart's own body. By the time Simic has the Hart and {3} to spare, its deficit has closed. In 141 of the 150 windows in its own main phase with an empty stack it controlled a landfall payoff (127 of them Tatyova), so those two lands were two cards. No landfall payoff is priced at all. |
| Burnished Hart (mono-black) | 173 / 0 | −1.60 to −0.50 | **Neither.** Its body is sacrifice fodder for the death payoffs, and mono-black has no landfall payoff. |
| Mind Stone, draw (simic, izzet) | 161 / 3, 50 / 7 | +0.50 | **Neither.** It is no longer `never`. Its row declares only `answers: ["value"]`, so it is +0.50, under any cast in the main phase and under `InstantThreshold` in the upkeep and draw step. By the opponents' turns it is tapped for mana. A rock used as a rock is right. |
| Glint-Horn Buccaneer's rummage (izzet) | 7 / 0 | +1.50 | **Borderline.** It is priced exactly `InstantThreshold` in the combat steps, and the tie goes to passing. Two games. |

**What is not the cause.** No list change explains an unused activation. And none of these cards explains Izzet's record: Face-Breaker was on the battlefield with a Treasure in 13 of Izzet's 64 seat-games.

### Decision

**1. An impulse is a card the bot can use this turn** (`PriceImpulse`, `aiseat/heuristic/impulse.go`). `game.Purpose` and `PurposeView` gain `impulse`: the cards it exiles from the top of its controller's library that its controller may play this turn. It is additive on the wire with no snapshot change, and it is declared by hand like every purpose: Professional Face-Breaker's row declares `impulse: 1`, the only one in the curated decks. It is priced `Weights.Hand` per card in the bot's own main phase with an empty stack, when it has a land drop left or two mana from its lands and rocks, and at nothing anywhere else, where the card is lost at the end of the turn.

**2. A Treasure given up costs a token** (`OneShotManaAsToken`, the same file). A token whose every mana ability sacrifices it costs `TokenWeight` (0.50) when it is sacrificed to pay a cost, rather than its board value (1.00). The board value is unchanged, and so is everything else that reads it.

Together, Face-Breaker's row is 1.20 − 0.50 − 0.30 = **+0.40** in the bot's own main phase, over `PassThreshold`, and below zero everywhere else. Either switch alone leaves it below zero (`TestImpulseNeedsBothSwitches`).

**3. Izzet: Fable of the Mirror-Breaker and Monument to Endurance for Wash Away and Negate.** Both cards are graded full and legal. Monument declares a discard payoff (any card, one Treasure: the least of its three bullets, and the one the purpose can say). The issue's other card, The Indomitable, was measured and is not proposed. The bot casts it in 6 of the 12 games it is offered in and crews it in 2, as it crews Smuggler's Copter in 4 of 10, because a crew activation is priced at the flat `ActivateBase` less the tapped creature, +0.20, under `PassThreshold` ([Out of scope](#out-of-scope): Vehicles). Fable's Goblin makes a Treasure on every attack, its second chapter rummages, and the bot casts it in 9 of the 12 games it is offered in. Wash Away and Negate are the two narrowest counterspells.

**4. Simic: no change.** It is no longer behind (28.9% over two seed blocks at step 1), so neither its list nor a landfall price is proposed here. It falls to 21.1% (14.9%–29.0%) at step 2, in games the Izzet swap re-dealt, and that interval still holds 25%. If the owner takes the swap, Simic is the deck to re-measure first. **Esper: no trim.** Over two seed blocks (128 games) Esper and mono-black share the lead (36.7% and 32.8% before the swap, 35.9% and 39.1% after), neither interval lies above 50%, and Izzet's gap is not one a trim to Esper would close.

### Measured

The tables are under [Measurements, #2480](#2480-izzet-and-simic-after-the-rebalance-2026-10-12). In short, run 1 at seed 1 and run 2:

| Step | esper | izzet | mono-black | simic | Run 2 `heuristic` | Face-Breaker's row |
|---|---:|---:|---:|---:|---|---|
| `develop` `0e81b3af4` | 17 | 1 | 26 | 20 | 66 / 192, 34.4% (28.0%–41.3%) | 0 of 1,070 windows |
| 1, the impulse price | 17 | 1 | 24 | 22 | 64 / 192, 33.3% (27.0%–40.3%) | 20 of 935, in 7 of 13 games |
| 2, the Izzet swap | 20 | 3 | 25 | 16 | 56 / 192, 29.2% (23.2%–36.0%) | 18 of 335, in 6 of 11 games |

- **The impulse price works and moves little.** It changes 5 of run 1's 64 games and 4 of run 2's first half, and none of the second half, where no `heuristic` seat plays Izzet.
- **No swap moves Izzet beyond noise.** Pooled over every Izzet seat-game in run 1 at seeds 1 and 1001 and run 2, Izzet won 4 of 224 (1.8%, 0.7%–4.5%) before the swap and 10 of 224 (4.5%, 2.4%–8.0%) after. The issue's pair (The Indomitable and Monument) and a four-card set (Fable, Monument, Glorybringer and Third Path Iconoclast for Wash Away, Negate, Swan Song and Shock) are no different in run 1: 4 and 2 of 64, against the proposed pair's 3.
- **A deck change re-deals every game.** 55 of run 1's 64 games end differently after the swap, so a swap's numbers are a new sample, and the other decks move as much as Izzet does.
- **Health:** 0 stalls and 0 rejected moves in every run. Turns p50 is 13 in run 1 at every step. The suite is 41 of 41, every tag at 100%. No pinned arena digest moved.

### Options considered and rejected

- **The issue's pair, The Indomitable and Monument to Endurance.** See decision 3.
- **Four swaps.** No better than two in run 1 (2 of 64), and it cuts the deck further from its precon.
- **A landfall payoff purpose** (a `landfall_payoff` on a triggered row, like `discard_payoff`, so a land the bot puts onto the battlefield is worth Tatyova's card, a Cobra's mana or a Baloths token). It is the real cause of the Hart's price, and it would reprice every land Simic plays or fetches. It needs a new wire field and declarations on nine Simic cards. It is not small, and Simic is no longer behind. See question 4.
- **Pricing a crew by the attack it enables.** That is the Vehicles item in [Out of scope](#out-of-scope), and it is the reason The Indomitable was not proposed.

### Questions for the owner (#2480, answered 2026-10-12)

Each lists the recommended option first.

1. **The impulse price** (decisions 1 and 2).
   - **(a) Recommended:** accept, with both switches on in `DefaultConfig()`. It turns a dead activation into a used one and changes no strength number beyond noise.
   - **(b)** Reject. Revert the first commit. Face-Breaker stays `never`.
2. **The Izzet swap** (decision 3).
   - **(a) Recommended, weakly:** Fable of the Mirror-Breaker and Monument to Endurance for Wash Away and Negate. Both new cards are used, and the direction is up, but the change is within noise.
   - **(b)** The issue's pair: The Indomitable and Monument to Endurance. The bot rarely crews The Indomitable.
   - **(c)** No swap. Revert the second commit.
3. **Izzet's gap.** At 2% to 5% no swap of two to four cards closes it. In run 1 at seed 1, Izzet is out first or second in 48 of 64 games. Of the damage the game logs attribute to a deck's own cards, Izzet's is the least: its best source is Mary Read and Anne Bonny with 446, against Syr Konrad's 2,448 and Y'shtola's 1,324, which hit each opponent.
   - **(a) Recommended:** a new issue to rebuild izzet-aggro around what the heuristic plays well (cheap threats that make value when they enter or attack, and damage to each opponent, which is what Konrad and Y'shtola win with), measured over at least three seed blocks.
   - **(b)** Heuristic work first: the crew price above, and why Izzet's small attackers trade so badly.
   - **(c)** Accept Izzet as the weak deck for now.
4. **Simic and landfall.**
   - **(a) Recommended:** no change now. A follow-up issue for a landfall payoff purpose, to be measured against Simic's balance when it is built.
   - **(b)** Build the landfall price now, in its own amendment.
5. **Esper.**
   - **(a) Recommended:** no trim (decision 4).
   - **(b)** Trim Esper now (Consecrated Sphinx, Sheoldred, the two free counterspells, as #2475 listed).

### Owner decisions (#2480, 2026-10-12)

The owner accepted the amendment on 2026-10-12. They took the recommendation on questions 1, 2, 4 and 5, and chose a rebuild on question 3. No section above changed.

1. **The impulse price is accepted.** `PriceImpulse` and `OneShotManaAsToken` are both on in `DefaultConfig()`, and both off in `BaselineConfig()`.
2. **The Izzet swap is accepted:** Fable of the Mirror-Breaker and Monument to Endurance for Wash Away and Negate.
3. **Izzet is rebuilt:** [#3071](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3071), in S67. The rebuild may swap the commander if that makes sense, and may start from scratch.
4. **Simic does not change now.** The landfall payoff purpose is [#3072](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3072), in S67.
5. **Esper is not trimmed.**

## Amendment 2026-10-12: blocks that survive as a set

**Status:** Accepted (owner decisions 2026-10-12). The owner accepted the check with `BlockSetSurvives` on, and had the horizon's unneeded waiting turn filed as [#3080](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3080); see [Owner decisions (#3066, 2026-10-12)](#owner-decisions-3066-2026-10-12) at the end of this amendment. The code is in the PR that adds this section, behind one switch, on in `DefaultConfig()` and off in `BaselineConfig()` (§9).
**Issue:** [#3066](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3066), filed from #2474 (PR #3065). Sprint S67. Tracker #890.
**Amends:** the block planner (`decideBlock`, `heuristic/combat.go`), which [Out of scope](#out-of-scope) left alone, as the [#2675 knobs](#2675-2690-2676-combat-priced-by-what-combat-does-2026-10-08) and the two amendments of 2026-10-12 on the race and the attrition horizon did for the attack planner. No ADR covers the block planner, and `docs/bot.md` documents it.

### The problem

Blocks are declared one at a time, and `decideBlock` scores each block on its own: the damage it saves, the attacker it kills and the creatures it loses (#1548). Nothing checks that the blocks together leave the defender alive. The full seed-1409 Wurm mirror ends on this misblock ([Open (#2474)](#open-2474)):

- **The defender:** on 3, with its 3/3 commander as its only creature.
- **The attack:** a Bear 2/2 and a 3/3 commander.

The Bear block scores +6.90: it kills the Bear and loses nothing. The commander block scores +6.00: the two commanders trade. So the defender blocks the Bear, takes 3 from the commander and dies. Commander on commander takes 2 and lives. The line #2474 found then takes another turn, which no whole game in the tree played.

### Decision

**The blocks survive as a set** (`BlockSetSurvives`, `heuristic/blockset.go`; on in `DefaultConfig`, off in `BaselineConfig`).

1. **When it runs.** Before it scores the blocks on offer, the block planner asks whether finishing the declaration now kills the bot. If it does not, nothing changes: blocking more never lets more damage through, so every block on offer survives too.
2. **What kills.** The bot's life, unless it can't lose for life (ADR 0057); ten poison counters from infect and toxic (CR 104.3d, 702.90b, 702.164c); and the table's commander damage, 21 by default, from one commander (CR 903.10a). A trampler's overflow counts (`connects`, CR 702.19b), a double striker deals its damage twice (CR 702.4b), and a menace attacker with one blocker counts as unblocked (CR 702.110b).
3. **The search.** When finishing kills, each block on offer is asked whether some way of spending the bot's creatures that are still free after it survives (`survivable`). Each free creature blocks one attacker that still gets damage through, or nothing. Which creature may block which attacker is read off the block moves on offer, so the search never pairs creatures the enumerator would not let block. It stops at the first set that survives. It prunes a branch when its free creatures could not stop enough damage even if each stopped a whole attacker, and gives up after 2000 boards (`blockSearchBudget`). A block whose search gives up counts as unknown, and is neither chosen nor ruled out.
4. **The choice.** If some block on offer survives, the bot declares the best-scored such block, even at a score of zero or less, since finishing kills. If the best-scored block overall survives, that is the same block with the same reason as before. When the check changed the pick, the reason says `…, a set that survives (+6.00)`. If no block survives, or every search gave up, the decision is the old one.

On the issue's board, finishing now takes 5. The Bear block leaves no free creature and 3 through, which kills. The commander block leaves 2 through, which does not. So the bot blocks commander on commander.

### Measured

The tables, and every game that changes, are under [Measurements, #3066](#3066-blocks-that-survive-as-a-set-2026-10-12). In short, at `develop` `2b6f13831`, from non-race lockstep runs:

| | 4-seat mean | 2-seat mean | Wurm + Ogre | full Wurm | Run 1 wins, esper / izzet / black / simic |
|---|---:|---:|---|---|---|
| `develop` | 16.17 | 13.25 | 5 / 6 | 7 / 6, on the misblock | 18 / 5 / 25 / 16 |
| **`BlockSetSurvives`** | **16.08** | **13.33** | 5 / **7** | **8 / 7** | 18 / 5 / 25 / 16, every game the same |

- **The full Wurm mirror plays the line out.** The defender trades commanders and lives on 1. Both commanders are cast again, the bot's in its own second main phase. The Bear and the commander then swing into one blocker, and the defender, with no surviving block left, dies. The games end on turns 8 and 7, one turn later than the misblock, with the same winners.
- **The line is a turn shorter than the horizon planned.** The horizon counts the bot's commander summoning sick on the board after the trade and waits a turn for it (`afterExchange`, `afterWait`). The bot casts it again in the same turn's second main phase, so by its next turn the commander has been under its control since that turn began (CR 302.6), and no turn is spent waiting. `TestAttritionWaitsForTheCommanderItCastAgain` pins the decision on the board the horizon predicts, which the game never reaches.
- **2 of the 24 gate games change, each at a block the check changed, and neither changes its winner.** Seed 104 ends a turn sooner (13 → 12) and seed 202 a turn later (13 → 14). On the mirrors, Wurm + Ogre with the edge on seat 1 ends a turn later (6 → 7): the defender, on 4 with its commander, chumps the Ogre and takes 3 instead of trading commanders and taking 4. Every other game plays move for move as on `develop`.
- **Run 1 does not change.** All 64 games of §8's run 1 end on the same turn with the same winner, the same lives and the same number of decisions. Defenders facing lethal with a choice of blocks are rare at four seats with these decks.

### Options considered and rejected

- **Choosing the whole set at once**, declaring blocks from one search of every assignment. It would replace the planner's scores, which every gate game reads, rather than adding a check to them. The per-block declaration with a lookahead keeps today's choice wherever the bot survives, and the check costs nothing on a board where finishing survives.
- **Minimising the damage taken when finishing kills.** That is a different objective, which would change every chump the bot makes when it is not in lethal range. The check only rules out blocks that leave no way to live.
- **Checking only life.** Poison and commander damage kill as surely, and they are cheap to count from the public view (`PlayerView.poison`, `commander_damage`).

### Open (#3066)

- **The bot's own model of the defender does not change.** The race's `blockToSurvive` and the attrition horizon's free answer predict the defender's blocks with their own greedy, not with `decideBlock`. Where the two disagree the attacker misjudges a swing, as before.
- **The horizon's waiting turn is pessimistic.** See the second bullet under Measured. It costs a line a turn of the horizon's three, so a line that needs a wait and two more exchanges is never found. It is not part of this amendment: it is [#3080](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3080).

### Questions for the owner (#3066, answered 2026-10-12)

Each lists the recommended option first.

1. **Accept the check, on in `DefaultConfig()`?**
   - **(a) Recommended:** yes. It fixes the issue's misblock, the full Wurm mirror plays the line out, it moves no winner in any gate game or in run 1, and it runs only on a board where finishing the declaration kills.
   - **(b)** Keep it as an off switch. The full Wurm mirror keeps ending on the misblock.
2. **The horizon's waiting turn** (Open, second bullet).
   - **(a) Recommended:** file it as its own issue: the horizon counts a commander the bot casts again in its second main phase as able to attack on its next turn. It changes the horizon's search, which needs its own gate measurements.
   - **(b)** Leave it. The wait is pessimistic, so it only loses lines and never sends a bad one.

### Owner decisions (#3066, 2026-10-12)

The owner accepted the amendment on 2026-10-12 and took the recommendation on both questions. No section above changed.

1. **The set-level block check is accepted.** `BlockSetSurvives` is on in `DefaultConfig()`, and off in `BaselineConfig()`.
2. **The horizon's waiting turn is its own issue:** [#3080](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3080), in Backlog. It changes the attrition horizon's search, not the block planner, so it is not part of this amendment.

## Amendment 2026-10-12: landfall payoffs

**Status:** Accepted (owner decisions 2026-10-13). The owner accepted the landfall price with `PriceLandfallPayoffs` on, and had the open items filed as [#3146](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3146) and the random-bot game test's timeout as [#3147](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3147); see [Owner decisions (#3072, 2026-10-13)](#owner-decisions-3072-2026-10-13) at the end of this amendment. The code is in the PR that adds this section, behind one switch, on in `DefaultConfig()` and off in `BaselineConfig()`.
**Issue:** [#3072](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3072), filed from #2480's [owner decision 4](#owner-decisions-2480-2026-10-12). Sprint S67.
**Amends:** §6 (a new purpose field, `landfall_payoff`, on a triggered row) and what §6's `lands`, #2678's extra land drop and #2680's put from hand are worth. No other price and no deck changes.

### The problem

[#2480](#why-the-cards-go-unused) found Burnished Hart's activation offered 346 times on simic-ramp and taken once. Its two lands are priced at `ManaSource` each plus the ramp premium, against the Hart's own body, and by the time Simic has the Hart and {3} to spare its deficit has closed. In 141 of the 150 windows of its own main phase with an empty stack it controlled a landfall payoff, 127 of them Tatyova, so those two lands were two cards and two life. Nothing priced that. Re-measured on `develop` at `489b289aa` (run 1, both seed blocks), Simic's Hart activation was taken in 5 of 1,047 windows, and in 540 of the 559 seed-1 windows the decision log can read, Simic controlled a payoff.

### Decision

**1. A landfall row declares what it pays** (`game.LandfallPayoff`, `purpose.landfall_payoff` on the wire). The amounts are per land that enters under its controller's control: `draws`, `life_gain`, `tokens`, `mana`, `counters` (on its source), `energy` and `damage_each_opponent`. It is declared by hand, like every purpose, and only on a triggered row (`effects.Register` refuses it anywhere else, and refuses one that pays nothing). A card file builds the row with `effects.LandfallPaying(label, payoff, effect)` in place of `Landfall(label, effect)`. It is declared on Simic's Tatyova, Lotus Cobra, Tireless Tracker, Tireless Provisioner, Sporemound, Rampaging Baloths and Courser of Kruphix (`TestCuratedDeckPurposes` holds each to it), and on the catalog's other fixed-amount landfall rows: Architect of the Untamed, Emeria Angel, Koth of the Homestead, Koth, the Geomancer (the damage only), Maja, Nesting Dragon, Nissa, Worldsoul Speaker, both Omnaths that pay a fixed amount, Primeval Bounty, Roil Cartographer, Sabotender, Sazh's Chocobo, Scute Swarm, The Necrobloom, Undergrowth Champion and Zendikar's Roil. A payoff counted at resolution (Avenger of Zendikar's counter on each Plant) or conditioned on the lands (Field of the Dead) declares none.

**2. A land the bot puts onto the battlefield is worth what its payoffs pay** (`PriceLandfallPayoffs`, `aiseat/heuristic/landfall.go`). Per land, over the triggered rows of the bot's own permanents, in the units the policy already uses:

```
Weights.Hand × draws + MarginalLife(own life) × life_gain + TokenWeight × tokens
+ 0.5 × TokenWeight × mana + (Weights.Power + Weights.Toughness) × counters
+ Weights.Energy × energy + DamageToOpponent × damage_each_opponent × live opponents
```

It is added in the three places a land the bot puts down is priced: a declared purpose's `lands` (Burnished Hart's and Sakura-Tribe Elder's activations, Harrow, Cultivate, Solemn Simulacrum), the lands an extra land drop lets the bot play this turn (#2678), and a land put from hand (#2680). The ordinary land drop is not repriced: it is taken before any cast already. An opponent's payoff pays the opponent, so it adds nothing.

On turn 8 with seven lands, the Hart's activation is −1.20 without a payoff and **+1.44** with Tatyova out (two cards and two life), so the bot cracks it. With only Courser of Kruphix it is −0.96, and with an opponent's Tatyova −1.20, so it keeps the Hart (`TestHartIsActivatedWithTatyovaOut`, `TestHartIsKeptWhenTheLandsPayLittle`).

No weight is new. The one constant is `landfallManaShare` (0.5): a landfall mana empties at the end of the step, so it is half a Treasure.

### Measured

The tables are under [Measurements, #3072](#3072-landfall-payoffs-2026-10-12). In short, over run 1 at seeds 1 and 1001 (128 games) and run 2 at both seeds (384 `heuristic` seat-games):

| | Simic, run 1 | Simic, every `heuristic` seat-game | Run 2 `heuristic` | Hart activation, Simic run 1 |
|---|---|---|---|---|
| `develop` `489b289aa` | 28 / 128, 21.9% (15.6%–29.8%) | 61 / 224, 27.2% (21.8%–33.4%) | 134 / 384, 34.9% (30.3%–39.8%) | 5 of 1,047 windows, in 5 of 27 games |
| **`PriceLandfallPayoffs`** | **35 / 128, 27.3% (20.4%–35.6%)** | **70 / 224, 31.2% (25.5%–37.6%)** | 131 / 384, 34.1% (29.6%–39.0%) | **21 of 225 windows, in 21 of 26 games** |

- **The Hart is used.** It is taken in 21 of the 26 games it is offered in, against 5 of 27, and it is offered in far fewer windows because it is cracked early. In run 2 the `heuristic` Simic's Hart goes from 2 of 19 games to 12 of 17.
- **The other land-makers move with it.** Explosive Vegetation is cast in 50 of 67 games it is offered in (28 of 55 on `develop`), Sakura-Tribe Elder's sacrifice is used in 33 of 41 (17 of 42), and Harrow, Cultivate and Kodama's Reach are each cast a little more.
- **Simic's win rate rises, within noise.** Every interval overlaps the one before. Mono-black's run-1 share falls from 49 to 46 of 128 and Esper's from 45 to 42, in games the change re-dealt.
- **Run 2 is not detectably worse.** Half A (`heuristic` on Esper and Izzet, so no `heuristic` seat holds a payoff) plays every game as `develop` does at both seeds. Half B changes 34 and 38 of 48 games.
- **Health:** 0 stalls and 0 rejected moves in every run, turns p50 unchanged (13 and 14 in run 1). The suite is 41 of 41, every tag at 100%, on both.

### Options considered and rejected

- **Deriving the payoff from the effect** (a `Do(DrawCards{N: 1})` body read at registration). Several payoffs are hand-written closures (Tatyova's is), and §6 says purposes are declared, never derived. `LandfallPaying` is the one-line declaration instead.
- **Repricing the land drop.** It is already priced above every cast, so a payoff would change nothing.
- **Pricing Ancient Greenwarden's doubling.** It is a static, not a landfall row, and it pays nothing alone. See Open.
- **Declaring `lands: 1` on the fetch lands.** A fetch's activation is priced at the flat `ActivateBase` today and is cracked. A declared `lands` would make it pay for the land it sacrifices, and without a payoff it would fall under `PassThreshold`. That is a change to fetches, not to landfall.

### Open (#3072)

- **Ancient Greenwarden** doubles every landfall trigger and is in Simic's list. A doubler would need its own declaration (on the static, read with the payoffs).
- **Casting a payoff** is still priced by its body. Tatyova cast before a turn of land drops is worth the cards she will draw, which nothing prices. The commander is cast in 63 of 64 games anyway.
- **Fetch lands** (Simic has seven) are cracked at the flat `ActivateBase`, so a payoff does not make them more likely. They were not declared (see above).

All three are [#3146](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3146).

### Questions for the owner (#3072, answered 2026-10-13)

Each lists the recommended option first.

1. **Accept the landfall price, with `PriceLandfallPayoffs` on in `DefaultConfig()`?**
   - **(a) Recommended:** yes. The Hart and the other land-makers are used when a payoff is out, Simic's numbers move up within noise, and run 2 and the suite do not move.
   - **(b)** Keep it as an off switch. The declarations stay, and the bot prices lands as before.
2. **The open items** (Greenwarden's doubling, a payoff's own cast price, fetch lands).
   - **(a) Recommended:** one Backlog issue for all three, measured together, since each changes Simic alone.
   - **(b)** Leave them.

### Owner decisions (#3072, 2026-10-13)

The owner accepted the amendment on 2026-10-13 and took the recommendation on both questions. No section above changed.

1. **The landfall price is accepted.** `PriceLandfallPayoffs` is on in `DefaultConfig()`, and off in `BaselineConfig()`.
2. **The open items are one issue:** [#3146](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3146), in Backlog: Ancient Greenwarden's doubling, a payoff's own cast price, and the fetch lands.
3. **The random-bot game test's timeout is its own issue:** [#3147](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3147), in Backlog. `TestFourRandomBotsPlayToAWinner` hits its wall clock on `develop` as on this branch ([Measurements](#3072-landfall-payoffs-2026-10-12)), so it is not part of this amendment.

## Amendment 2026-10-13: izzet-aggro rebuilt around Purphoros

**Status:** Accepted (owner decisions 2026-10-13). The owner took the Purphoros list, had the deck's ID renamed to `mono-red-aggro` in the same PR, accepted Simic's run-1 rate, and had the canary and the token makers filed as [#3164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3164) and [#3165](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3165); see [Owner decisions (#3071, 2026-10-13)](#owner-decisions-3071-2026-10-13) at the end of this amendment. The code is in the PR that adds this section: the deck list, its new ID with migration 0018 and an alias for the old one, the test lists that name its cards, and the docs that describe it. No price, weight or switch changes.
**Issue:** [#3071](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3071), filed from #2480's [owner decision 3](#owner-decisions-2480-2026-10-12). Sprint S67.
**Amends:** the izzet-aggro list #2436 set and #2480 swapped two cards of. No other deck changes.

### The problem

#2480 found that no swap of two to four cards moves izzet-aggro beyond noise, and the owner asked for a rebuild that may change the commander. Re-measured on `develop` at `f1bf51ff6`, with #3072's landfall price, over four seed blocks of §8's run 1 (64 games each, `--rotate --lockstep`):

| Deck | Run 1, seeds 1, 1001, 2001, 3001 | Every `heuristic` seat-game (run 1 and run 2) |
|---|---|---|
| esper-control | 98 / 256, 38.3% (32.5%–44.4%) | 203 / 448, 45.3% (40.8%–49.9%) |
| izzet-aggro | **9 / 256, 3.5% (1.9%–6.5%)** | **21 / 448, 4.7% (3.1%–7.1%)** |
| mono-black-aristocrats | 81 / 256, 31.6% (26.3%–37.6%) | 169 / 448, 37.7% (33.4%–42.3%) |
| simic-ramp | 68 / 256, 26.6% (21.5%–32.3%) | 137 / 448, 30.6% (26.5%–35.0%) |

### Decision (proposed)

**1. The commander is Purphoros, God of the Forge, and the deck is mono-red.** Whenever another creature enters under its controller's control, Purphoros deals 2 damage to each opponent. It is indestructible and costs four, and the bot casts it in every one of the 253 run-1 games in which it could, so the deck's engine is on the table by turn four or five whatever the library deals. Damage to each opponent is what the two strongest decks win with (#2480 question 3: Syr Konrad and Y'shtola), and a four-seat table pays it three times.

**2. The 99 feeds that trigger.** It keeps 31 of the old list's non-basic cards and the Mountains, and the commander and 41 cards are new, every one graded `full` and legal in Commander; Path of Ancestry is still the deck's only minor-caveats card.

- **Damage to each opponent when a creature or an artifact enters:** Impact Tremors, Witty Roastmaster, Molten Gatekeeper, Agate Instigator, Fanatic of Mogis (once, as it enters), and the kept Weftstalker Ardent and Reckless Fireweaver (a Treasure is an artifact entering).
- **When a spell is cast:** Firebrand Archer, Kessig Flamebreather, Erebor Flamesmith, Coruscation Mage, Firespitter Whelp, Fiery Inscription, Brimstone Vandal, and the kept Guttersnipe.
- **More per hit:** Torbran, Thane of Red Fell and Hellrider.
- **Creatures that bring creatures:** Goblin Instigator, Beetleback Chief, Siege-Gang Commander, Legion Warboss, Goblin Rabblemaster, both Krenkos, Hanweir Garrison, Loyal Apprentice, Young Pyromancer, Person of Interest, Pashalik Mons, Fearless Liberator, Goblin Matron, and the token spells Hordeling Outburst, Krenko's Command and Dragon Fodder.
- **Treasure and card flow:** Charming Scoundrel, Patron of the Arts, Rapacious Dragon, Seize the Spoils, Strike It Rich, Ruby Medallion, Fire Diamond, and the kept Captain Lannery Storm, Generous Plunderer, Professional Face-Breaker, Fable of the Mirror-Breaker, Goldspan Dragon, Big Score, Unexpected Windfall, Faithless Looting and four rocks.
- **Removal:** the kept Lightning Bolt, Shock, Arc Trail, Abrade, Chaos Warp and Vandalblast, with Incinerate, Lightning Strike and Wild Slash.
- **Out:** Mary Read and Anne Bonny and the whole Pirate and discard plan (the Pirates, Marauding Mako, Scrounging Skyray, Glint-Horn Buccaneer, Magmakin and Ingenious Artillerist, Monument to Endurance, Smuggler's Copter), every blue card (the counterspells, Cyclonic Rift, Pongify, Rapid Hybridization, Preordain, Frantic Search, Mulldrifter, Ledger Shredder), Windfall and Wheel of Fortune, Blaze, Solphim and Ancient Copper Dragon, and the Izzet lands. Ten non-basic lands stay; the rest are 27 Mountains.

**3. The ID becomes `mono-red-aggro`** (owner decision 2, below), following the other mono-coloured decks' colour-archetype IDs; the name is now Forge and Fury and the summary is rewritten. A saved table setup and a player's last deck name a deck by ID, so migration 0018 rewrites `izzet-aggro` to `mono-red-aggro` in `users.last_deck` and in each bot of `table_setups`. `decks.Lookup` still answers `izzet-aggro` with the new deck (`legacyIDs`), for what the migration cannot reach: a guest's browser (which the client also translates as it reads it), an engine snapshot, an operator's script. `IDs()` and the pickers list only the new ID. The rename does not change a game: run 1 at seed 1 plays all 64 games move for move the same under either ID.

### Measured

The tables are under [Measurements, #3071](#3071-izzet-aggro-rebuilt-around-purphoros-2026-10-13). In short, over §8's run 1 and run 2 at seeds 1, 1001, 2001 and 3001:

| | esper, run 1 | izzet, run 1 | mono-black, run 1 | simic, run 1 | Izzet, every `heuristic` seat-game | Run 2 `heuristic` |
|---|---|---|---|---|---|---|
| `develop` `f1bf51ff6` | 98 / 256, 38.3% | 9 / 256, 3.5% (1.9%–6.5%) | 81 / 256, 31.6% | 68 / 256, 26.6% | 21 / 448, 4.7% | 274 / 768, 35.7% (32.4%–39.1%) |
| **the rebuild** | 66 / 256, 25.8% (20.8%–31.5%) | **77 / 256, 30.1% (24.8%–36.0%)** | 69 / 256, 27.0% (21.9%–32.7%) | 44 / 256, 17.2% (13.1%–22.3%) | **152 / 448, 33.9% (29.7%–38.4%)** | 289 / 768, 37.6% (34.3%–41.1%) |

- **Izzet's interval overlaps every other deck's.** It is 29.7% to 31.2% in each seed block, against 1.6% to 6.2% before.
- **Simic falls to 17.2% in run 1,** below the issue's 20%, and its interval (13.1%–22.3%) still holds 20%. Over every `heuristic` seat-game it is 24.3% (20.6%–28.5%), from 30.6%. The rebuild's wins come from all three decks, and Esper loses the most (38.3% to 25.8%), but Simic is the one that ends lowest. See question 3.
- **The commander is what moved it.** Scouted with run 1, the same engine cards under a blue-red commander, in a list with blue card flow, won 8% to 10%, and two lighter versions of this list won as much as it does (below). Only Purphoros in the command zone took the deck past 25%.
- **Run 2 is not worse.** `heuristic` wins 37.6% of its seat-games, against 35.7% on `develop`, and its lower bound clears 25% in every seed block.
- **The cards are used.** The `never` lists hold at most two of Izzet's rows per seed block (Pashalik Mons's sacrifice and Fearless Liberator's boast), against two or three before. Hordeling Outburst is cast in 9 of the 37 games it is offered in, and Krenko, Mob Boss's activation is taken in 8 of 30; see question 5.
- **Health:** 0 stalls, 0 rejected moves and 0 runner fallbacks in all 1,280 games. Turns p50 in run 1 is 12, from 13. The suite is 41 of 41, every tag at 100%.

**Scouted variants** (run 1 only, pooled over the seed blocks named; the list measured above is the last row):

| Variant | Seeds | esper | izzet | mono-black | simic |
|---|---|---|---|---|---|
| Vivi Ornitier (UR), a spells-and-tokens list on the same engine cards with blue card flow and counterspells | 1, 1001 | 38 / 128 | 10 / 128, 7.8% | 54 / 128 | 26 / 128 |
| Gandalf the Grey (UR), the same list with Vivi in the 99 | 1, 1001 | 40 / 128 | 13 / 128, 10.2% | 47 / 128 | 28 / 128 |
| Purphoros, the first draft (six of its cards never cast) | 1, 1001 | 37 / 128 | 37 / 128, 28.9% | 31 / 128 | 23 / 128, 18.0% |
| Purphoros, the proposed cards in a different order (so every game is dealt differently) | 1, 1001, 2001 | 55 / 192 | 52 / 192, 27.1% | 48 / 192 | 37 / 192, 19.3% |
| the same without Torbran, one more Mountain | 1, 1001, 2001 | 47 / 192 | 65 / 192, 33.9% | 43 / 192 | 37 / 192, 19.3% |
| the proposed list without Torbran, Hellrider, Witty Roastmaster and Molten Gatekeeper, for Thrill of Possibility, Ox of Agonas, Commander's Sphere and a Mountain | 1–3001 | 55 / 256 | 86 / 256, 33.6% | 75 / 256 | 40 / 256, 15.6% |
| the proposed list without Witty Roastmaster, Molten Gatekeeper and Agate Instigator, for three Mountains | 1–3001 | 78 / 256 | 77 / 256, 30.1% | 59 / 256 | 42 / 256, 16.4% |
| **proposed** | 1–3001 | 66 / 256 | **77 / 256, 30.1%** | 69 / 256 | 44 / 256, 17.2% |

A deck list's card order decides the shuffle, so a list in a new order is a new sample: the fourth row is the proposed cards in the scouting order, and it lands 3 points from the proposed row. Taking out the strongest cards did not make the deck weaker beyond that noise, and Simic sits between 15.6% and 19.3% under every Purphoros list.

### Options considered and rejected

- **Keeping a blue-red commander.** Vivi Ornitier and Gandalf the Grey, on the same engine cards, are 8% and 10%: better than 3.5%, and still far below the other decks. Saheeli, Jewel of Avishkar, the first choice, is not legal in Commander in the Scryfall dump.
- **A weaker Purphoros list,** to leave Simic above 20%. Two were measured over four seed blocks (above), and neither moved Izzet or Simic beyond noise.
- **More card swaps on the Mary Read list.** #2480 measured three sets and none moved it.

### Questions for the owner (#3071, answered 2026-10-13)

Each lists the recommended option first.

1. **The rebuild.**
   - **(a) Recommended:** accept the Purphoros list. Izzet's run-1 interval overlaps every other deck's, run 2 is not worse, and every card is graded `full`.
   - **(b)** Keep the deck blue-red with Gandalf the Grey's list, at about 10%.
   - **(c)** Reject. The deck stays as it is.
2. **The ID and the name.** The deck is mono-red with the ID `izzet-aggro` and the name Forge and Fury.
   - **(a) Recommended:** keep the ID now and file a follow-up to rename it (with a migration of saved table setups and last-deck records), so this PR changes the list and nothing that stores it.
   - **(b)** Rename the ID in this PR.
   - **(c)** Keep the ID for good; only the name says what the deck is.
3. **Simic at 17.2% in run 1** (13.1%–22.3%; 24.3% over every `heuristic` seat-game).
   - **(a) Recommended:** accept it as about 20%, and re-measure Simic when #3146 (Greenwarden's doubling, a payoff's cast price, fetch lands) is built, since it is the deck that issue strengthens.
   - **(b)** Hold the rebuild until #3146 lands.
   - **(c)** Look for a weaker Purphoros list. Two attempts did not move it.
4. **The A3 canary.** Mary Read and Anne Bonny's loot (§8, A3) is in no curated deck now, so the arena report shows it as not offered.
   - **(a) Recommended:** replace it with Faithless Looting (cast), the loot class in the rebuilt deck, in a follow-up.
   - **(b)** Drop the canary.
5. **Two token makers the heuristic undervalues.** Hordeling Outburst is cast in 9 of 37 games, and Krenko, Mob Boss's activation, which declares no token count because its X is counted at resolution, is taken in 8 of 30.
   - **(a) Recommended:** one Backlog issue to price token-making activations by the tokens and what they trigger, measured on this deck.
   - **(b)** Leave them.

### Owner decisions (#3071, 2026-10-13)

The owner accepted the amendment on 2026-10-13. They took the recommendation on questions 1, 3, 4 and 5, and option (b) on question 2. Decision 3 above was rewritten to match; no other section changed.

1. **The rebuild is accepted:** the Purphoros list, as measured.
2. **The ID is renamed in this PR,** to `mono-red-aggro`, with migration 0018 for saved table setups and last-deck records, and `izzet-aggro` kept as an alias that resolves to it (decision 3).
3. **Simic's 17.2% in run 1 is accepted** as about 20%. It is re-measured when [#3146](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3146) is built, which a comment on that issue records.
4. **The A3 loot canary becomes Faithless Looting (cast):** [#3164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3164), in Backlog.
5. **The token makers are one Backlog issue:** [#3165](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3165), pricing token-making spells and activations by what the tokens trigger.

## Amendment 2026-10-14: token makers priced by what the tokens trigger

**Status:** Proposed. The code is in the PR that adds this section, behind one switch, `PriceTokenPayoffs`, on in `DefaultConfig()` and off in `BaselineConfig()`.
**Issue:** [#3165](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3165), filed from #3071's [owner decision 5](#owner-decisions-3071-2026-10-13). Sprint S75.
**Amends:** §6 (three new purpose fields) and what §6's `tokens` are worth. No other price and no deck changes.

### The problem

On the rebuilt `mono-red-aggro` deck, run 1 at seeds 1, 1001, 2001 and 3001 found two token makers used much less than the deck's others. Hordeling Outburst (three Goblins) was cast in 9 of the 37 games it was offered in; Krenko's Command and Dragon Fodder, the same effect for two tokens, in 24 of 45 and 33 of 51. Krenko, Mob Boss's activation was taken in 8 of 30. A declared `tokens` is priced at `TokenWeight` (0.50) each, so Outburst was worth 1.5 before its card and its mana, and Krenko's row declares no count (X is counted at resolution), so it was priced at the flat `ActivateBase`. In this deck every creature that enters is also 2 damage to each opponent from Purphoros, and 1 from Impact Tremors and the other pingers, which nothing priced.

### Decision

**1. A creature-enter row declares what it pays** (`game.CreatureEnterPayoff`, `purpose.creature_enter_payoff`): landfall's fields, per creature that enters under its controller's control. Only on a triggered row (`effects.Register` refuses it elsewhere, and one that pays nothing). A card file wraps its trigger in `effects.CreatureEnterPaying`. Declared on Purphoros (2 damage), Impact Tremors, Warleader's Call, Witty Roastmaster, Molten Gatekeeper and Agate Instigator (1 damage each), and Hinterland Sanctifier, Lifecreed Duo and Dazzling Angel (1 life).

**2. A token maker says how many of its tokens are creatures** (`purpose.creature_tokens`, at most `tokens`), because a Treasure does not trigger Purphoros. Declared on Hordeling Outburst, Krenko's Command and Dragon Fodder. A count made at resolution declares `purpose.tokens_per_controlled`, a lowercase creature subtype: Krenko, Mob Boss is `goblin`, and the reader counts the Goblins the bot controls from the view when it prices the activation.

**3. A creature token is worth what the bot's own payoffs pay** (`PriceTokenPayoffs`, `aiseat/heuristic/token_payoff.go`), added to a declared purpose's value:

```
creature tokens x (sum over the bot's own creature-enter rows, in landfall's units)
+ TokenWeight x the Goblins counted, for a tokens_per_controlled row
```

An opponent's payoff pays the opponent, so it adds nothing. A row with `tokens_per_controlled` is now a priced purpose, so Krenko's activation is priced by its Goblins in place of `ActivateBase`; the tap is charged as before. A creature spell's own entry is not repriced. No weight is new.

With Purphoros out, Hordeling Outburst's cast price is 5.70 against 0.30 without (three Goblins, 2 damage, three opponents), and the bot casts it (`TestHordelingOutburstIsWorthMoreWithPurphorosOut`). Krenko's activation is priced by the Goblins it counts and by what they trigger (`TestKrenkoActivationIsPricedByItsGoblins`).

### Measured

The tables are under [Measurements, #3165](#3165-token-makers-priced-by-what-the-tokens-trigger-2026-10-14). The comparison is the same binary with the switch off (`develop`'s behaviour) against on, `--rotate --lockstep`, run 1 at seeds 1, 1001, 2001 and 3001 (256 games per deck) and run 2 at seeds 1 and 1001 in both halves:

| | `develop` | `PriceTokenPayoffs` |
|---|---|---|
| Hordeling Outburst cast, games used of offered | 10 of 38 | **33 of 35** |
| Krenko's Command cast | 20 of 42 | 37 of 43 |
| Dragon Fodder cast | 29 of 50 | 47 of 51 |
| Krenko, Mob Boss activated | 4 of 30 | **32 of 36** |
| mono-red-aggro, run 1 | 73 / 256, 28.5% (23.3%-34.3%) | **87 / 256, 34.0% (28.5%-40.0%)** |
| run 2 `heuristic`, all decks | 137 / 384, 35.7% (31.0%-40.6%) | 140 / 384, 36.5% (31.8%-41.4%) |

- **The token makers are used.** Every one is cast or used in about nine games in ten it is offered in.
- **The deck wins a little more, within noise.** The intervals overlap. The other decks' run-1 shares move by less than the interval (black 24.6% to 20.7%, esper 27.7% to 26.2%, simic unchanged), in games the change re-dealt.
- **Run 2 is not detectably worse**, and `heuristic` on mono-red-aggro is 31 / 96 to 36 / 96.
- **Health:** 0 stalls in all 16 runs, turns p50 11 to 12 on both, the suite 42 of 42 with every tag at 100%.

### Options considered and rejected

- **Pricing `tokens` higher.** It would raise Treasures and Clues too, which trigger nothing, and it does not see Purphoros.
- **Deriving the count from the effect.** Krenko's is a closure. `tokens_per_controlled` is declared by hand, as every purpose is.
- **Repricing a creature spell's own entry.** A creature cast is already priced by its body; the payoffs it triggers are a different change.

### Open (#3165)

- **A creature spell triggers the same payoffs** (Purphoros with a creature in hand) and nothing prices that.
- **Creature-token makers with fixed counts elsewhere in the catalog** (Siege-Gang Commander, Beetleback Chief and the other enters-effects) declare no `creature_tokens`; they keep their body price.
- **Pashalik Mons's sacrifice and Fearless Liberator's boast**, the deck's other `never` rows, are related and untouched.

## Amendment 2026-10-14: landfall doublers, payoff casts and fetch lands

**Status:** Accepted (owner, 2026-10-15). The code is in the PR that adds this section, behind three switches, each on in `DefaultConfig()` and off in `BaselineConfig()`.
**Issue:** [#3146](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3146), the open items of [#3072](#open-3072). Sprint S75.
**Amends:** the [landfall payoffs](#amendment-2026-10-12-landfall-payoffs) amendment (one new purpose field, `landfall_extra_triggers`, and `lands` declared on the fetch lands' activations). No weight is new and no deck changes.

### The problem

#3072 priced a land the bot puts down by what its landfall payoffs pay, and left three gaps, all in simic-ramp:

1. **Ancient Greenwarden** makes each landfall trigger fire twice, and the price counted each payoff once.
2. **A payoff's own cast** (Tireless Tracker, Lotus Cobra, Sporemound) was priced by its body alone, not by what it pays for the lands the bot puts down in the next turns.
3. **Fetch lands** (Simic has seven) were cracked at the flat `ActivateBase`, so a payoff on the board did not make them more likely to be cracked.

### Decision

**1. Doublers** (`PriceLandfallDoublers`). A card whose `TriggerDoubler` covers lands declares `landfall_extra_triggers` on the card: Ancient Greenwarden 1, and Elesh Norn, Mother of Machines 1, because her doubler covers every permanent. `effects.Register` refuses it off the card, and on a card with no `TriggerDoubler`. Each land pays the bot's payoffs (1 + the declared extras on its own permanents) times. An opponent's doubler doubles only the opponent's triggers, so it adds nothing.

**2. A payoff's cast** (`PriceLandfallCasts`). A permanent with a landfall row or a doubler adds the per-land gain it brings, `(board + own) × (k + extra) − board × k`, where `k` is the triggers per land already counted. That gain is multiplied by the lands the bot expects to put down:

```
expected = min(lands in hand, land drops left this turn)
         + min(lands left in hand + 0.4 × 2, land drops per turn × 2)
```

The second term covers the bot's next two turns, and 0.4 is `millLandShare`, a land in 0.4 of its draws. Tireless Tracker with two lands in hand and the drop used is 2.85 off and 3.85 on: two Clues more (`TestALandfallPayoffIsWorthTheLandsAhead`). Greenwarden with Tatyova out is worth her card and life again for each of those lands (`TestGreenwardenCastIsWorthTheDoubling`). The commander is not repriced: a command-zone card carries no ability rows on the wire, and Tatyova is cast in nearly every game anyway.

**3. Fetch lands** (`PriceFetchLandfall`). The fetch activations declare `lands` (1, or Myriad Landscape's 2) and `lands_untapped` where the land enters untapped. A fetch is a self-sacrificing land row that declares lands and nothing else. It is not priced as a purpose, because that would charge the sacrificed land and drop a fetch under `PassThreshold` with no payoff out. It keeps the flat `ActivateBase`, and:

- each land it puts down adds what the bot's payoffs pay (`landfallValue`);
- a land beyond the one it sacrifices is ramp (`ManaSource` and the ramp premium).

Misty Rainforest is +0.50 with no payoff out, as before, and +1.82 with Tatyova (`TestAFetchPaysTheLandfallPayoffs`). With the switch off, a fetch's declared `lands` is read nowhere.

### Measured

The tables are under [Measurements, #3146](#3146-landfall-doublers-payoff-casts-and-fetch-lands-2026-10-14). One tree was built twice, with the three switches off (`develop`'s behaviour) and on. Run 1 was at seeds 1, 1001, 2001 and 3001, with the rebuilt `mono-red-aggro` in the rotation, as #3071's owner decision 3 asked. Run 2's half B (`heuristic` on mono-black and Simic) was at seeds 1 and 1001, and half A at seed 1. Half A was kept small because the host is shared.

| | Switches off | Switches on |
|---|---|---|
| simic-ramp, run 1 (four seed blocks) | 43 / 256, 16.8% (12.7%–21.9%) | **49 / 256, 19.1% (14.8%–24.4%)** |
| Simic, every `heuristic` seat-game (run 1 and run 2 half B) | 77 / 448, 17.2% (14.0%–21.0%) | 76 / 448, 17.0% (13.8%–20.7%) |
| run 2 `heuristic`, the three runs | 97 / 288, 33.7% (28.5%–39.3%) | 93 / 288, 32.3% (27.2%–37.9%) |
| Tireless Tracker cast, run 1, games used of offered | 73 of 81 | **87 of 91** |
| Sporemound cast | 62 of 97 | 74 of 92 |
| a fetch cracked while a payoff is out (seed 1, decision logs), windows | 77 of 1,020 (7.5%) | **90 of 534 (16.9%)** |

- **The cards are used.** Tracker, Sporemound, Avenger of Zendikar (66 of 74 to 76 of 82) and Courser (64 of 71 to 72 of 76) are cast in more of the games they are offered in. With a payoff out, a fetch is cracked in more than twice the share of windows, and it sits uncracked for half as many windows.
- **Simic's win rate does not move beyond noise.** Run 1 rises from 16.8% to 19.1%, against the 17.2% (13.1%–22.3%) the owner accepted for #3071. Run 2's half B falls from 34 / 96 to 27 / 96. Over every `heuristic` seat-game it is flat. Every interval overlaps the one before it.
- **Run 2 is not detectably worse.** `heuristic`'s upper bound stays at 37.9%, above 25%. Half A changes 5 of 48 games, through mono-red-aggro's Myriad Landscape, whose second land is now ramp.
- **Health:** 0 stalls and 0 rejected moves in every run. Turns p50 is unchanged (11 to 12). The suite is 42 of 42.

### Options considered and rejected

- **Declaring `lands: 1` on a fetch and pricing it as a purpose.** It would charge the sacrificed land, and with no payoff out it would drop the fetch under `PassThreshold`, so the bot would stop cracking fetches.
- **Reading the doubler off the engine's `TriggerDoubler`.** A doubler is a closure, so the bot cannot read which events it covers. Every other purpose is also declared, never derived.
- **A longer horizon for the payoff's cast.** Two turns prices a Tracker at about two Clues. A longer horizon would put the payoffs above the ramp that finds the lands.

### Open (#3146)

- **The commander's cast.** A command-zone card has no ability rows on the wire, so Tatyova's cast is not repriced. Projecting the rows there is a protocol change, and it has not been measured.
- **Simic's balance** stays about where #3071 left it. Moving it needs more than landfall pricing.
- **Virtue of Knowledge** also doubles landfall, but it is an Adventure card whose card-slot purpose the bot reads for the spell, so it declares no `landfall_extra_triggers`.

## Amendment 2026-10-15: what a card does only when kicked or turned face up

**Status:** Accepted (a seam the coordinator filed, with the owner's brief). The code is in the PR that adds this section. No `Config` switch: the new fields are read by prices that already have their own (`PriceSweeps`, `Weights.Turn`), and `BaselineConfig()` prices neither.
**Issue:** [#3762](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3762), found landing [#3565](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3565) (the cards that skip your next turn). Sprint: none.
**Amends:** §6 (a declared purpose) and [ADR 0188](0188-skipped-turns-and-turn-bound-effects.md) Decision 10 (a skipped turn is priced).

### The problem

`game.Purpose` says what a spell or an ability does. It had no slot for what a card does only if an optional cost is paid, or only when it is turned face up, and two cards from #3565 showed what that costs:

- **Waterspout Elemental** ("Kicker {U} ... When this creature enters, if it was kicked, return all other creatures to their owners' hands and you skip your next turn") had to declare its `Sweep` on the card, because `TestRealDumpPurposeAudit` fails a catalog card that reads as a wipe and declares none. So the heuristic priced the **unkicked** cast as a mass bounce. Its skip had no home at all, so the bot never charged it.
- **Unstable Hulk** ("When this creature is turned face up, it gets +6/+6 ... You skip your next turn") could declare nothing: a face-down permanent has no text (CR 708.2a), so no purpose reaches the wire for it, and the skip is not something the card does when cast.

### Decision

**1. Two conditional purposes**, both on the card slot, both a `*Purpose` (a pointer, so `Purpose` stays comparable and `IsZero` still works):

```go
Purpose: game.Purpose{Kicked: &game.Purpose{Sweep: …, SkipsTurns: 1}}   // Waterspout Elemental
Purpose: game.Purpose{FaceUp: &game.Purpose{SkipsTurns: 1}}             // Unstable Hulk
```

The fields beside them are what the card does however it is cast, so an unkicked Elemental declares nothing and is priced as the 3/4 flyer it is.

**2. A reader recognises the condition from the move.** The wire carries them as `purpose.kicked` and `purpose.turned_face_up`. The heuristic reads `kicked` only when the cast move names an optional cost (`optional_costs` is non-empty), and adds it to whatever the cast's slot already declared (`castPurpose`, so the fuel, tax and plan readers agree). It reads `turned_face_up` only when pricing the `turn_face_up` special action, which adds `purposeValue` of it to the flat `SpecialActionValue`: the Hulk's turn costs `Weights.Turn`. `Priced()` does not count either field, so a card that declares only a conditional purpose keeps its mana-value proxy on every other path.

**3. Kicked is for one kicker.** `effects.Register` refuses `Kicked` unless the card has exactly one optional cost and it is a kicker. With two ("Kicker {R} and/or {W}") the view cannot say which was paid, and a multikicker's amounts are counted at resolution; those cards keep declaring nothing. It refuses `FaceUp` on a card with no face-down cast offer (Morph, Megamorph, Disguise). Both inner purposes must say something, may not nest, and may not declare `Targets`, `Aura`, `ExtraLandDrops`, `LandfallExtraTriggers` or `FlashCleanupSacrifice`, which belong to the card's own slot. Everything else is checked as a card-slot purpose is.

**4. A face-down permanent's controller is sent `turned_face_up`, and only that.** `CardPurposeOfPtr` still answers the zero purpose for a face-down object, except for `FaceUp` of the card underneath, read by oracle ID. `redactCardForViewer` clears `purpose` for a viewer who may not look at the face, so an opponent sees nothing, as before. A face-up Hulk also carries `turned_face_up`; nothing reads it there.

**5. The audit reads through them.** `Purpose.DeclaresSweep()` is true for a sweep declared directly or under `Kicked` or `FaceUp`, and `TestRealDumpPurposeAudit` uses it. A wipe that only a kicked cast makes still has to be declared, and now it can be declared where it is true.

### Cards

Waterspout Elemental moves its partial creature bounce and its skip to `Kicked`. Unstable Hulk declares its skip under `FaceUp`. The other catalog kicker cards were read against the dump. None declares a purpose that is true only when kicked; the ones with a kicked effect the Purpose model has a field for (Sunscape Battlemage's draw, Stormscape Battlemage's life) have two kickers, which decision 3 leaves alone, and the rest are counters, tokens counted at resolution or targeted effects the model has no field for.

### Not priced

The Hulk's +6/+6 and trample until end of turn. No purpose field prices a one-turn pump on a permanent being flipped (`Pump` is an exert row's), so turning the Hulk up now costs a turn and credits nothing. At `Weights.Turn` 6.00 against a flat 1.00 the bot will not turn a Hulk up, where before it flipped it for the flat value and skipped its next turn without ever pricing it. A flip for lethal is the case this loses; pricing it needs the attack the pump enables, and is a separate change.

### Options considered and rejected

- **Per-mode purposes.** Kicker is not a mode (CR 702.33), and the cards that choose among modes already declare per bullet.
- **A `Purpose` slot on the optional cost.** `AdditionalCost` is engine data the view projects as `OptionalCostView`; putting the bot's pricing there would make the engine own a bot concern, and the wipe still has to be found by the audit on the spec's purposes.
- **Leaving the sweep on the card and reviewing it as not-a-wipe.** `reviewedNotAWipe` says a card does not wipe. Waterspout does.

## Consequences

- About a third of each curated deck stops being dead. The decks can then be judged on their lists, which is what #2436 needs.
- The evaluation changes for every permanent with ability rows. That changes removal targets and sacrifice choices as well as casts, and the suite and run 2 are there to catch it.
- The bot plays faster tables. Expect turns p50 to fall a little and wall clock per game to rise a little, because more moves are priced per window. A8 bounds both.
- One wire type grows (`PurposeView`) and the catalog gains a hand-declared field. Like `Completeness`, it can be wrong. `TestCuratedDeckPurposes` and the dump audit are what keep it honest for the cards that matter.
- The arena reports per deck from now on, which #2436 and #2437 both need anyway.

## Rejected alternatives

- **Lowering `PassThreshold` or `Weights.Hand`.** That is one number for every card. It would make the bot cast the dead third, and also every bad card and every wipe, with no knowledge of which is which. It also moves every existing position at once.
- **Reading oracle text in the heuristic.** ADR 0033 §3 keeps policies on the view. The model tiers already get oracle text through `model.Config.Oracle`. A rule-based reader of English text is a parser nobody owns.
- **Simulating the cast on a cloned game.** Ruled out by ADR 0033 §3, and recorded at the top of `moves.go`.
- **Fixing the decks first.** The owner's ordering of 2026-10-06. A rebalance measured with a bot that never casts a rock would tune the decks around the bot's blind spots.

## Measurements

*Each PR appends its row. Every row records: git SHA, `BaselineConfig()` and `DefaultConfig()` deltas, seeds, games, per-deck win rate with Wilson 95% interval, `never` counts, canary use rates, run 2's policy win rate with its interval, suite agreement overall and per tag, stalls, rejected moves, and turns p50.*

### PR 2: the S66 baseline (2026-10-06)

`develop` at `61c6b02c1` plus PR 2's `74a1d4921`, which changes no price. `BaselineConfig()` equals `DefaultConfig()`: no ADR 0126 term exists yet, so the delta is empty. Arena runs use the concurrent schedule, turn budget 60.

**Run 1:** `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1`. 64 games, 0 stalls, 2 rejected moves, turns p50 14.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` | Non-land cards offered |
|---|---:|---:|---|---:|---:|
| esper-control | 4 | 6.2% | 2.5%–15.0% | 19 | 64 |
| izzet-aggro | 8 | 12.5% | 6.5%–22.8% | 22 | 72 |
| mono-black-aristocrats | 38 | 59.4% | 47.1%–70.5% | 29 | 69 |
| simic-ramp | 14 | 21.9% | 13.5%–33.4% | 24 | 64 |

A6 fails today as the ADR expected: mono-black's interval lies entirely above 25%, but not entirely above 50%.

Canaries (A3), games used out of games offered: Sol Ring 0 of 22 (esper), 0 of 15 (izzet), 0 of 18 (black), 0 of 20 (simic). Rhystic Study 0 of 15 (esper), 0 of 13 (simic). Mary Read and Anne Bonny's loot 0 of 62. Entomb 0 of 14. Harrow 0 of 28. Viscera Seer (cast) 0 of 20.

Mana rocks and dorks (A2): every rock in every deck is used in 0% of the games it was offered in. The only A2 cards above the 80% bar are creatures with a body big enough to clear today's price: Birds of Paradise 24 of 25, Ornithopter of Paradise 23 of 24, Delighted Halfling 18 of 19, Palladium Myr 21 of 24 (simic) and 9 of 10 (black). Llanowar, Fyndhorn and Elvish Mystic Elves are 0%.

**Run 2:** two `heuristic` and two `heuristic-baseline` contestants, `--games 48 --rotate --seed 1` twice: `heuristic` on esper-control and izzet-aggro and `heuristic-baseline` on mono-black-aristocrats and simic-ramp, then swapped. 96 games, 0 stalls.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves | Turns p50 |
|---|---:|---:|---:|---|---:|---:|
| heuristic | 192 | 48 | 25.0% | 19.4%–31.6% | 2 | 14 |
| heuristic-baseline | 192 | 48 | 25.0% | 19.4%–31.6% | 1 | 14 |

This is the sanity row: the two policies are one config, so each takes exactly the null. With the same seeds in both halves, each policy plays every deal once on each deck pair.

**Suite:** `boteval suite run --policy heuristic`: 22 of 22 agree, every tag at 100% (attack 4, block 4, cast 7, choice 1, combat 8, land 3, mulligan 6, removal 1). The answers are identical to `develop`'s, position by position.

**No price change:** a fixed-seed lockstep run (`--games 8 --rotate --seed 1 --lockstep` on the four decks) gives the same games before and after this PR: the same winner, turns, life totals and runner counters in every game.

### PR 3: mana sources (2026-10-06)

`develop` at `717d7ce51` plus PR 3's `5c457891d`. `DefaultConfig()` gains three terms and `BaselineConfig()` zeroes all three: `Weights.ManaPerExtra` 1.00, `RampPerMana` 1.00, `RampWantCap` 7, the §9 starting values unchanged. Arena runs use the concurrent schedule, turn budget 60.

Two readings of §2, both chosen as the closest to its intent:

- A source's amount is **net** of its ability's own mana cost, so a Signet's "{1}, {T}: Add {W}{U}" makes one mana, not two. Read literally ("the count of symbols in `produced`"), every Signet would be priced as a Sol Ring.
- `want` leaves out the card being cast, so a rock in a hand of nothing else closes no gap.

**Run 1:** the same command and seed as PR 2's. 64 games, 0 stalls, turns p50 14. There were 10 rejected moves, against PR 2's 2. Every one is `declare_attacker` refused with "action not legal in current step": a stale attack submitted after the step moved on, in the concurrent schedule, while the machine was heavily loaded by parallel runs. None is a cast or a mana source.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` (PR 2) |
|---|---:|---:|---|---:|
| esper-control | 2 | 3.1% | 0.9%–10.7% | 11 (19) |
| izzet-aggro | 5 | 7.8% | 3.4%–17.0% | 15 (22) |
| mono-black-aristocrats | 42 | 65.6% | 53.4%–76.1% | 24 (29) |
| simic-ramp | 15 | 23.4% | 14.7%–35.1% | 16 (24) |

A6 still fails, as it did at PR 2: mono-black's interval lies entirely above 25%, but not entirely above 50%.

Canaries (A3), games used out of games offered: Sol Ring 20 of 20 (esper), 12 of 12 (izzet), 17 of 17 (black), 16 of 17 (simic), all at PR 2's 0%. Rhystic Study, Mary Read's loot, Entomb, Harrow and Viscera Seer are still 0%; they belong to PRs 4 to 8.

Mana rocks and dorks (A2), games used out of games offered, against PR 2's 0% for every rock:

| Meets 80% | Below 80% |
|---|---|
| Sol Ring 94–100% in every deck; Hedron Archive 13 of 14; Worn Powerstone 15 of 16; and, as at PR 2, Palladium Myr, Birds of Paradise, Ornithopter of Paradise and Delighted Halfling | Arcane Signet 29–64%, the Signets 20–50%, the Talismans 20–53%, Mind Stone 11–44%, Thought Vessel 17–55%, Commander's Sphere 0–50%; Llanowar Elves 14 of 28, Fyndhorn Elves 7 of 18, Elvish Mystic 5 of 16 (all 0% at PR 2) |

Why the one-mana sources stop short: across the run, a (game, seat, card) offer of a one-mana rock or elf that was never used had a deficit of 0 or less at its first offer in 221 of 287 cases. The rock arrived after the bot's sources already covered everything it held. §2 prices that cast at 1.0 − 1.2 = −0.20 on purpose ("Arcane Signet, turn 9, deficit 0"), and §5's `LeftoverThreshold` (0.00) does not take a −0.20 move either. A2 counts every game in which the card was offered, so as written it cannot reach 80% for a one-mana source drawn late. The owner decides at PR 9 whether A2 should count only offers made while the deficit is open, or whether a late rock should be priced up.

**Run 2:** the same shape and seeds as PR 2's: two `heuristic` and two `heuristic-baseline` contestants, `--games 48 --rotate --seed 1` twice, the decks swapped between halves. 96 games, 0 stalls, turns p50 14.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves |
|---|---:|---:|---:|---|---:|
| heuristic | 192 | 55 | 28.6% | 22.7%–35.4% | 2 |
| heuristic-baseline | 192 | 41 | 21.4% | 16.1%–27.7% | 0 |

Not detectably worse: the `heuristic` interval's upper bound is 35.4%, above 25%. Per half: `heuristic` won 10 of 96 seat-games on esper and izzet against the baseline's black and simic, and 45 of 96 on black and simic against the baseline's esper and izzet.

**Suite:** four positions added and gated for `heuristic`: `cast-sol-ring-turn-one`, `cast-the-signet-when-short`, `cast-the-elf-turn-one` and `hold-the-signet-late`. Each comes from a run 1 decision log. `boteval suite run --policy heuristic` gives 26 of 26, every tag at 100% (attack 4, block 4, cast 11, choice 1, combat 8, land 4, mulligan 6, removal 1). `BaselineConfig()` passes priority in the first three positions instead of casting, which is the change they measure, and its rankings are recorded in `baseline_rankings.json`.

### PR 4: permanents by what they do (2026-10-06)

`develop` at `717d7ce51` plus PR 4's `b9cd11c3c`. `DefaultConfig()` against `BaselineConfig()`: `Weights.PermanentPerMana` 0.50, `RowTriggered` 0.60, `RowStatic` 0.50, `RowActivated` 0.40 and `RowCap` 3, all zero in the baseline. These are §9's starting values, unchanged. Arena runs use the concurrent schedule, turn budget 60.

**Run 1:** the same command and seed as PR 2's. 64 games, 0 stalls, 2 rejected moves (PR 2: 2), turns p50 13 (PR 2: 14).

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` | PR 2 `never` | Non-land cards offered |
|---|---:|---:|---|---:|---:|---:|
| esper-control | 3 | 4.7% | 1.6%–12.9% | 13 | 19 | 65 |
| izzet-aggro | 7 | 10.9% | 5.4%–20.9% | 17 | 22 | 73 |
| mono-black-aristocrats | 34 | 53.1% | 41.1%–64.8% | 23 | 29 | 71 |
| simic-ramp | 20 | 31.2% | 21.2%–43.4% | 15 | 24 | 65 |

The `never` count falls by 5 to 9 per deck, 26 in all. 31 cards left the lists, all of them the enchantment and artifact engines and the small utility creatures this PR prices: Esper Sentinel, Kismet, Land Tax, Phyrexian Arena, Rhystic Study and Smothering Tithe (esper); Bident of Thassa, Coastal Piracy, Impact Tremors, Impulsive Pilferer, Marauding Mako and Sulfuric Vortex (izzet); Bastion of Remembrance, Carrion Feeder, Dictate of Erebos, Exquisite Blood, Grave Pact, Phyrexian Arena, Sanguine Bond, Vampiric Rites, Viscera Seer and Zulaport Cutthroat (black); Beastmaster Ascension, Elemental Bond, Garruk's Uprising, Guardian Project, Rhystic Study, Sakura-Tribe Elder (cast and activate), Wayfarer's Bauble and Wood Elves (simic). Five joined: the activated abilities of creatures now cast for the first time, so now offered at all (Marauding Mako; Burnished Hart, Carrion Feeder, Vampiric Rites and Warren Soultrader's sacrifices). Those are §7's sacrifice outlets and PR 8's to price. What is left is §2's rocks and dorks, §5 and §6's cheap spells, loots and tutors, and the sacrifice outlets' activations.

Canaries (A3), games used out of games offered: Rhystic Study 12 of 16 (esper, 75%) and 17 of 20 (simic, 85%), from 0 of 15 and 0 of 13. Viscera Seer (cast) 22 of 23 (96%), from 0 of 20. Sol Ring, Mary Read and Anne Bonny's loot, Entomb and Harrow stay at 0, as expected: they are PRs 3, 5 and 7's classes. A2 is unchanged: every rock and the three Elves are still 0%.

A6: mono-black's interval (41.1%–64.8%) does not lie entirely above 50%, and its point estimate fell from 59.4% to 53.1%.

**Run 2:** the same shape as PR 2's. 96 games, 0 stalls.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves | Turns p50 |
|---|---:|---:|---:|---|---:|---:|
| heuristic | 192 | 51 | 26.6% | 20.8%–33.2% | 2 | 14 / 13 |
| heuristic-baseline | 192 | 45 | 23.4% | 18.0%–29.9% | 4 | 14 / 13 |

Not detectably worse: the `heuristic` interval's upper bound, 33.2%, is above 25%. Turns p50 is per half.

**Suite:** `boteval suite run --policy heuristic`: 24 of 24 agree, every tag at 100% (attack 4, block 4, cast 9, choice 1, combat 8, land 3, mulligan 6, removal 1). The 22 earlier positions answer as before. Two positions are new and gated: `cast-rhystic-study` and `cast-viscera-seer`, harvested from a four-game logged arena run on this branch (seed 101). `BaselineConfig()` misses both: it casts Pull from Tomorrow for one card in the first and passes in the second. The ADR's `equip-before-attacking` is a unit test instead (`TestEquipBeforeAttacking`), because none of the four curated decks holds an Equipment, so no arena window can be harvested for it.

**Re-measured on `develop` with PR 3.** PR 3 merged while this PR was open, so runs 1 and 2 were repeated on the merge of `develop` `62bc4ddec` into this branch. Same commands and seeds.

Run 1: 64 games, **1 stall**, 2 rejected moves, turns p50 12. The table compares with PR 3's own run 1.

| Deck | Wins | Win rate of decided | Wilson 95% interval | `never` | PR 3 `never` |
|---|---:|---:|---|---:|---:|
| esper-control | 2 | 3.2% | 0.9%–10.9% | 6 | 11 |
| izzet-aggro | 2 | 3.2% | 0.9%–10.9% | 9 | 15 |
| mono-black-aristocrats | 41 | 65.1% | 52.8%–75.7% | 15 | 24 |
| simic-ramp | 18 | 28.6% | 18.9%–40.7% | 7 | 16 |

Canaries: Rhystic Study 10 of 14 (71%) and 11 of 15 (73%). Viscera Seer 21 of 21. Sol Ring is at 100% in every deck (PR 3).

The stalled game is the Sanguine Bond and Exquisite Blood drain loop, which the bot now casts. The aristocrats seat was at 82 life and the others at 7, 2 and 5. The CR 732 loop breaker (ADR 0055) suspended automatic passing, and a bot table has nobody to step it on. Filed as #2450. It needs an ADR 0055 decision, not a price change.

A6: mono-black at 65.1% (52.8%–75.7%) is PR 3's number (65.6%, 53.4%–76.1%) unchanged. Its interval does lie entirely above 50%, both there and here. That is for the exit PR and #2436.

Run 2: `heuristic` 52 of 192, 27.1% (21.3%–33.8%). `heuristic-baseline` 44 of 192, 22.9% (17.5%–29.4%). 0 stalls, 3 and 3 rejected moves. Not detectably worse.

Suite on the merged tree: 28 of 28, every tag at 100% (attack 4, block 4, cast 13, choice 1, combat 8, land 4, mulligan 6, removal 1).

**Re-measured on `develop` with PRs 3, 5 and 6.** PR 5 (#2446) and PR 6 (#2448) merged next, so runs 1 and 2 were repeated on the merge of `develop` `816613cd` into this branch, with the same commands and seeds. PR 5 had also fixed the Goblin Sharpshooter untap. This branch's broader guard (`TapTarget` and `UntapTarget` do nothing for any permanent no longer on the battlefield) keeps PR 5's regression test, which covers the same case.

Run 1: 64 games, 1 stall, 4 rejected moves, turns p50 12.

| Deck | Wins | Win rate of decided | Wilson 95% interval | `never` |
|---|---:|---:|---|---:|
| esper-control | 1 | 1.6% | 0.3%–8.5% | 2 |
| izzet-aggro | 2 | 3.2% | 0.9%–10.9% | 7 |
| mono-black-aristocrats | 44 | 69.8% | 57.6%–79.8% | 11 |
| simic-ramp | 16 | 25.4% | 16.3%–37.3% | 2 |

Canaries, games used out of games offered:

| Canary | Deck | Used | Rate |
|---|---|---:|---:|
| Sol Ring | each of the four | 20 of 21, 16 of 16, 15 of 15, 20 of 21 | 95% to 100% |
| Rhystic Study | esper | 10 of 16 | 62% |
| Rhystic Study | simic | 20 of 24 | 83% |
| Mary Read and Anne Bonny's loot | izzet | 30 of 62 | 48% |
| Entomb | black | 11 of 15 | 73% |
| Harrow | simic | 0 of 29 | 0% |
| Viscera Seer | black | 21 of 22 | 95% |

The remaining `never` cards are mostly black's sacrifice outlets and their activations (PR 8), Harrow and the discard-cost spells (PR 7). The stall is again the aristocrats drain loop at the CR 732 loop breaker (#2450). The black seat was at 75 life. A6 still fails: black's interval lies entirely above 50%.

Run 2: `heuristic` won 55 of 192, 28.6% (22.7%–35.4%). `heuristic-baseline` won 41 of 192, 21.4% (16.1%–27.7%). 0 stalls; 1 and 0 rejected moves. Not detectably worse.

Suite: 31 of 31 agree, every tag at 100% (activate 1, attack 4, block 4, cast 15, choice 1, combat 8, land 4, leftover 3, mulligan 6, removal 1). Every position PRs 3, 4 and 5 added is gated for `heuristic`.

### PR 6: the purpose signal (2026-10-06)

`develop` at `717d7ce51` against PR 6's branch, and again at `62bc4ddec` (PR 3 merged) against the branch with `develop` merged in. No price changes: `BaselineConfig()` and `DefaultConfig()` are untouched, and the heuristic reads no new field (an any-player row's `Draws` and `ControllerLosesLife` are still the only purpose it reads, and they did not change on any card).

**Suite:** `boteval suite run --policy heuristic` with both binaries: 22 of 22 agree on `717d7ce51` and 26 of 26 on `62bc4ddec`, every tag at 100%, and the two JSON reports are identical apart from latency.

**Arena:** `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 8 --rotate --seed 1 --lockstep` with both binaries. `games.jsonl` and `summary.json` are identical apart from game IDs and timings: the same winner, turns and life totals in every game, the same runner counters (decisions, applied, rejected, passes) and Layer A meter per seat, and the same Cards table (offered and taken per card).

**Snapshot:** `TestSnapshotShapeIsRecorded` passes with `testdata/snapshot_shape/` unchanged; `SnapshotSchemaVersion` does not move.

### PR 5: the two windows (2026-10-06)

`DefaultConfig()` against `BaselineConfig()`: `LeftoverWindows` true (baseline false), `LeftoverThreshold` 0.00 (0), `SpellFloor` 1.30 (0), `TapByTiming` true (false). These are the starting values in §9, not changed by tuning. Measured twice:

- on `717d7ce51` (PR 2) plus PR 5 alone;
- on `277743d08` (PRs 3 and 6 merged) plus PR 5, the `develop` it merges into.

Arena runs use the concurrent schedule, turn budget 60.

Two readings of §5, both recorded in the PR:

- **The second main phase is a sorcery-speed window.** There the leftover bar is for sorcery-speed moves that tap no creature. An instant, a flash spell, an instant-speed ability, or any move that taps a creature waits for the end step before the bot's turn. The mana and the creature are both still untapped then. The first arena run without this rule did three things after combat: cast Entomb, cast Vampiric Tutor, and looted with Mary Read and Anne Bonny, giving up a blocker for a whole round.
- **`SpellFloor` is for untargeted spells only:** a spell with no target on the card, on any mode, or on the move. A targeted spell is priced by its targets, as before. A floor under it would fire removal at smaller creatures outside the windows. Out of scope says that must not change.

**Run 1:** the same command and seed as PR 2's.

On `277743d08` + PR 5: 64 games, 0 stalls, 1 rejected move, turns p50 12.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` (PR 3; PR 2) |
|---|---:|---:|---|---:|
| esper-control | 3 | 4.7% | 1.6%–12.9% | 6 (11; 19) |
| izzet-aggro | 4 | 6.2% | 2.5%–15.0% | 10 (15; 22) |
| mono-black-aristocrats | 44 | 68.8% | 56.6%–78.8% | 16 (24; 29) |
| simic-ramp | 13 | 20.3% | 12.3%–31.7% | 8 (16; 24) |

On `717d7ce51` + PR 5 alone: 64 games, 0 stalls, 1 rejected move, turns p50 13.

| Deck | Wins | Win rate | Wilson 95% interval | `never` |
|---|---:|---:|---|---:|
| esper-control | 3 | 4.7% | 1.6%–12.9% | 14 |
| izzet-aggro | 5 | 7.8% | 3.4%–17.0% | 18 |
| mono-black-aristocrats | 43 | 67.2% | 55.0%–77.4% | 23 |
| simic-ramp | 13 | 20.3% | 12.3%–31.7% | 13 |

This PR moves A6 the wrong way. Mono-black's interval now lies entirely above 50%. Three things in the black deck clear the leftover bar: its cheap creatures, its tutors, and Syr Konrad's untargeted mill in the end step before its turn. In four harvest games, 13 of 21 Konrad activations came from that window. Its opponents' engines still do not clear the bar; that is PR 4.

Canaries (A3), games used out of games offered, on `277743d08` + PR 5 (PR 5 alone in brackets; PR 2 was 0% for each):

| Canary | Games used | Meets A3 |
|---|---|:--:|
| Mary Read and Anne Bonny's loot | 32 of 62, 52% (27 of 61) | yes |
| Entomb | 13 of 19, 68% (16 of 18) | yes |
| Viscera Seer (cast) | 17 of 20, 85% (18 of 21) | yes |
| Sol Ring | 93–100% in every deck, from PR 3 | yes |
| Harrow | 0 of 27 | no |
| Rhystic Study | 0% | no |

- Viscera Seer's sacrifice ability is still never used (PR 8).
- Harrow does not move. Its additional cost sacrifices a land, so §5 keeps the normal bar for it, and it waits for §6's `purpose.lands` (PR 7).
- Rhystic Study is PR 4.

Mana rocks and dorks (A2), on `277743d08` + PR 5:

- Llanowar Elves 24 of 27, Fyndhorn Elves 23 of 24 and Elvish Mystic 16 of 18 now meet 80%. They were 14 of 28, 7 of 18 and 5 of 16 at PR 3. A cheap dork is cast with leftover mana in the second main phase.
- Worn Powerstone (13 of 16) meets 80%.
- The other one-mana and two-mana rocks are still below 80%, at 0–67%: a late rock is still priced below 0.00 (PR 3's note).

**Run 2:** the same shape and seeds as PR 2's. 96 games each, 0 stalls.

| Tree | Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves | Turns p50 |
|---|---|---:|---:|---:|---|---:|---:|
| `277743d08` + PR 5 | heuristic | 192 | 49 | 25.5% | 19.9%–32.1% | 2 | 13 |
| | heuristic-baseline | 192 | 47 | 24.5% | 18.9%–31.0% | 1 | 13 |
| `717d7ce51` + PR 5 | heuristic | 192 | 51 | 26.6% | 20.8%–33.2% | 2 | 15 / 14 |
| | heuristic-baseline | 192 | 45 | 23.4% | 18.0%–29.9% | 4 | 15 / 14 |

The result is not detectably worse in either tree: the `heuristic` interval's upper bound is 32.1% and 33.2%, both above 25%. PR 3 alone measured 28.6% (22.7%–35.4%). All three intervals overlap widely.

**Suite:** `boteval suite run --policy heuristic` on `277743d08` + PR 5 gives 29 of 29, every tag at 100%:

| Tag | Positions |
|---|---:|
| activate | 1 |
| attack | 4 |
| block | 4 |
| cast | 13 |
| choice | 1 |
| combat | 8 |
| land | 4 |
| leftover | 3 |
| mulligan | 6 |
| removal | 1 |

- Three new positions are gated for `heuristic`: `loot-at-the-end-step-before-yours`, `tutor-at-the-end-step-before-yours` and `cantrip-with-leftover-mana`. `BaselineConfig()` passes on all three, and its rankings are recorded.
- `do-not-cantrip-before-the-three-drop` is not added, because the harvest found no window where it was the unambiguous answer. `TestTheCantripDoesNotBeatTheThreeDrop` pins that behaviour in `heuristic` instead.

**Engine bug found:** the catalog soak on this branch failed on nightly seed 2026100601. Goblin Sharpshooter's own death triggers its untap. The untap then resolved against the card in exile and raised an effect error. The PR fixes it in `UntapTarget` (CR 400.7), with a regression test.

### PR 7: wipes, ramp spells and discard costs (2026-10-06)

`develop` at `6f73abab8` (PRs 2 to 6) merged into PR 7. `DefaultConfig()` against `BaselineConfig()`: `PricePurposes` true (baseline false), `TutorWeight` 1.00 (0), `SelfMillWeight` 0.50 (0), `DiscardWeight` 0.60 (0), `TokenWeight` 0.50 (0), `PriceSweeps` true (false), `DiscardCostByCard` true (false), `LastLandDiscard` 1.00 (0). The §9 starting values, unchanged by tuning. Arena runs use the concurrent schedule, turn budget 60.

Readings of the ADR, each the closest to its intent:

- **A sweep is priced against PR 6's shape as built.** `opponents_only` spares the bot's own permanents. `amount_is_x` reads the X the move announces. `sacrifice` removes everything it matches, indestructible or not. `destroy` spares an indestructible permanent. `damage` and `minus` spare a creature whose toughness is above the amount, and `damage` also spares an indestructible one. A `partial` sweep takes half of each matched permanent, because the view does not say which ones it spares. A `bounce` sweep takes half, but all of a token. An Aura goes with the permanent it enchants. Several chosen modes remove the union of what each removes.
- **PR 5's `SpellFloor` stays a floor under a declared purpose, and is not put under a priced sweep.** §5 puts the floor "under every instant and sorcery", and §6 says a purpose "may clear the normal thresholds", so the purpose lifts a spell and does not sink it. Without the floor, a tutor prices at 1.20 − 1.20 and Entomb at 0.60 − 1.20, and both would stop clearing §5's bar. A sweep gets no floor, because §4 exists to price a bad wipe below zero.
- **A permanent's purpose is its enters effect, added to its body** (§6, as built in PR 6): Wood Elves' land, Mulldrifter's two cards.
- **The bot's own activated row with a purpose is priced by that purpose instead of `ActivateBase`.** A row without one keeps `ActivateBase` (owner decision 6). A row that sacrifices its own source pays for the source. A row that taps a creature pays PR 5's `tapCreatureCost`, so a loot before combat costs the attack it replaces. Before PR 5 merged, without that price, the bot looted with Mary Read and Anne Bonny in its first main phase 414 times in run 1 instead of attacking with her.
- **The last land in hand costs more to discard.** §7's worked example says an early Windfall that discards the only land in hand is not positive, but its weights price it at +0.70. `LastLandDiscard` (1.00) charges the land drop it risks, on top of `cardValue`, while the bot has fewer than `LandsWanted` sources.

**Run 1:** the same command and seed as PR 2's, with `--decision-log`. 64 games, 0 stalls, turns p50 14. There were 16 rejected moves:

- 14 `declare_attacker` refused with "action not legal in current step": stale attacks in the concurrent schedule, which the decision log slows down.
- 1 Counterspell refused for "insufficient mana".
- 1 pass refused while a Rhystic Study prompt was unanswered. Rhystic Study is cast for the first time since PR 4.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` (PR 4, re-measured with PRs 3, 5 and 6) |
|---|---:|---:|---|---:|
| esper-control | 4 | 6.2% | 2.5%–15.0% | 1 (2) |
| izzet-aggro | 4 | 6.2% | 2.5%–15.0% | 3 (7) |
| mono-black-aristocrats | 41 | 64.1% | 51.8%–74.7% | 10 (11) |
| simic-ramp | 15 | 23.4% | 14.7%–35.1% | 2 (2) |

The `never` lists left are Commander's Sphere (esper), Lotus Petal and two creatures' activations (izzet), Tarmogoyf and Sakura-Tribe Elder's sacrifice (simic), and in black Altar's Reap, Village Rites, both Altars, Blood Artist and five sacrifice activations (PR 8). A6 fails, as it does on `develop` (PR 4's merged row: 69.8%, 57.6%–79.8%): black's interval lies entirely above 50%. This PR does not move it either way.

**This PR's classes**, games used out of games offered:

- **Wipes:** of 92 creature wipes cast, 1 was cast while the bot held the biggest creature board, by power plus toughness. On PR 3's run 1 decision log it was 13 of 88.
- **Ramp spells:** Farseek 20 of 27, Kodama's Reach 19 of 24, Nature's Lore 17 of 26, Cultivate 15 of 24, Rampant Growth 11 of 23, Three Visits 8 of 16, Harrow 6 of 25. Harrow sacrifices a land, so §5 keeps the normal bar for it, and late in the game two lands for a land and a card does not clear it.
- **Tutors**, mono-black then esper: Demonic Tutor 12 of 20 and 12 of 18, Vampiric Tutor 18 of 22 and 16 of 23, Diabolic Tutor 8 of 17 and 3 of 14. Also Buried Alive 11 of 15 and Entomb 15 of 18.
- **Discard costs:** Unexpected Windfall 15 of 17, Big Score 11 of 13.

Canaries (A3): Entomb 15 of 18 (83%), Mary Read and Anne Bonny's loot 41 of 62 (66%), Viscera Seer (cast) 23 of 23, Rhystic Study 8 of 14 (esper) and 18 of 22 (simic), Sol Ring 93–100%. Harrow is 6 of 25 (24%), up from 0% at PR 2.

A2, measured as the owner decided on #2435: over the games in which each rock or dork was offered while the bot's mana deficit was open, it was cast in 214 of 323 (66%).

- **At or above 80%:** Sol Ring 35 of 35, Llanowar Elves 14 of 14, Birds of Paradise, Delighted Halfling, Fyndhorn Elves, Elvish Mystic, Ornithopter of Paradise, Palladium Myr, Hedron Archive and Talisman of Curiosity.
- **Below 80%:** Worn Powerstone 10 of 13, Thought Vessel 12 of 21, Mind Stone 20 of 37, Arcane Signet 21 of 42, Commander's Sphere 7 of 22, the Signets 27–56%, and the other Talismans 33–54%.
- **Lower than before PR 4 merged** (75% on PR 5's tree): the rocks now compete for the same early mana with PR 4's engines, which are cast for the first time. This PR does not price rocks, and A2 is judged at PR 9.

**Run 2:** the same shape and seeds as PR 2's. 96 games, 0 stalls, turns p50 12 and 15 by half.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves |
|---|---:|---:|---:|---|---:|
| heuristic | 192 | 53 | 27.6% | 21.8%–34.3% | 2 |
| heuristic-baseline | 192 | 43 | 22.4% | 17.1%–28.8% | 0 |

Not detectably worse: the `heuristic` interval's upper bound is 34.3%, above 25%. By half, `heuristic` won 42 of 96 seat-games on black and simic against the baseline's esper and izzet, and 11 of 96 on esper and izzet against the baseline's black and simic. Both rejected moves are stale attacks.

**Suite:** four positions were added and gated for `heuristic`, all from this PR's run 1 decision logs, which were taken before PR 5 merged:

- `do-not-wrath-your-winning-board` and `wrath-a-losing-board`. A4 holds: both pass and are gated.
- `windfall-discarding-a-spare-land`. `BaselineConfig()` passes priority there instead of casting, which is the change it measures.
- `do-not-windfall-away-the-last-land`.

`boteval suite run --policy heuristic` gives 35 of 35, every tag at 100% (activate 1, attack 4, block 4, cast 19, choice 1, combat 8, discard 2, land 4, leftover 3, mulligan 6, removal 1, wipe 2). `baseline_rankings.json` records `BaselineConfig()`'s rankings for all four.

**Owner label review (2026-10-06).** The owner reviewed the thirteen S66 positions.

- **Approved as labelled, and now stamped `reviewer: krakenhavoc`, `reviewed_at: 2026-10-06`:** ten positions. They are `cast-sol-ring-turn-one`, `cast-the-signet-when-short`, `cast-the-elf-turn-one`, `hold-the-signet-late`, `loot-at-the-end-step-before-yours`, `tutor-at-the-end-step-before-yours`, `cast-rhystic-study`, `cast-viscera-seer`, `do-not-wrath-your-winning-board` and `wrath-a-losing-board`.
- **`do-not-windfall-away-the-last-land` is relabelled.** Accepted: Unexpected Windfall discarding Marauding Mako ("if going to pay 2 to cycle mako might as well unexpected windfall the mako instead"), and Pass. Rejected: discarding Myriad Landscape. It is stamped reviewed. The heuristic passes, which stays accepted. Cycling-aware discard pricing and a draw-step window go to a follow-up issue.
- **`cantrip-with-leftover-mana` (PR 5) keeps its label and is stamped reviewed.** The owner would rather cast Night's Whisper first. Since PR 4 the heuristic casts Exquisite Blood there, which the label accepts. Choosing the draw spell needs lookahead over the turn's mana, which goes to a follow-up issue.
- **`windfall-discarding-a-spare-land` keeps its label and has no `reviewed_at` yet.** Mary Read and Anne Bonny is on the battlefield there, so discarding the Island makes a Treasure, and the owner wants Island first. A follow-up PR adds a `discard_payoff` purpose as an amendment to this ADR, flips the label and stamps it.

### PR 8: sacrifice outlets (2026-10-06)

`develop` at `427d3f702` (PRs 3 to 7 merged) plus PR 8. `DefaultConfig()` against `BaselineConfig()`: `SacrificeDyingAnyway` true (baseline false) and `DeathPayoff` 0.60 (0). `DeathPayoff` is §9's starting value, not changed by tuning, and the removal discount is the existing `RemovalConfidence` (0.80). Arena runs use the concurrent schedule, turn budget 60.

Three readings of §7, each the closest to its stated intent:

- **The bar.** §7 says dying-anyway pricing "makes 'sacrifice it in response' the bot's play". A response is made outside a sorcery-speed window, where the bar is `InstantThreshold` (1.50). An outlet activation is worth `ActivateBase` (0.50), less the flat 0.30 that a non-tapping row on a creature already pays, so no dying-anyway price can clear 1.50. Sacrificing a permanent that is about to be lost spends nothing the bot would otherwise keep, which is §5's premise. So a move whose only non-mana cost is sacrificing permanents that are dying anyway clears `LeftoverThreshold` (0.00), in any window. The owner approved this reading on 2026-10-06.
- **What "dying anyway" reads.**
  - A spell's sweep is read from the slot its stack item names: the alternative cost, the modes, or the card. An ability's stack item does not say which row it came from, so a planeswalker's or a saga's sweep is not seen.
  - In combat, a creature counts as losing only when sacrificing it gives up nothing. An attacker loses when its blockers kill it, it kills none of them, and it puts no trample damage over. A blocker loses when it dies and every attacker it blocks dies or lives the same without it.
  - A chump in front of a trampler is not losing (CR 702.19d). A chump in front of any other attacker is, because the attacker stays blocked (CR 509.1h).
- **Payoffs.**
  - A payoff's own row is not counted towards its own sacrifice, because whether it sees its own death is text the wire does not carry.
  - A row that sacrifices its own source is left as PR 7 priced it.
  - A devoured creature is sacrificed, so devour counts the payoffs too.
  - The payoff is added for every creature sacrificed, as §7 says. For a creature that dies anyway, the triggers would have fired regardless, so this overstates the gain. Its only effect is to make the bot keener to sacrifice something it is losing.

**Run 1:** the same command and seed as PR 2's. 64 games, 0 stalls, 2 rejected moves, turns p50 13.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` | PR 7's `never` |
|---|---:|---:|---|---:|---:|
| esper-control | 3 | 4.7% | 1.6%–12.9% | 1 | 1 |
| izzet-aggro | 6 | 9.4% | 4.4%–19.0% | 3 | 3 |
| mono-black-aristocrats | 42 | 65.6% | 53.4%–76.1% | 4 | 10 |
| simic-ramp | 13 | 20.3% | 12.3%–31.7% | 2 | 2 |

Mono-black's `never` list is down to Ashnod's Altar, Phyrexian Altar, Blood Artist and Burnished Hart's activation.

Sacrifice outlets on PR 7's run 1 and on this one, as taken out of windows offered, then games used out of games offered:

| Card and move | PR 7 | PR 8 |
|---|---|---|
| Viscera Seer, activate | 0 of 2433; 0 of 21 | 11 of 2312; 8 of 22 |
| Carrion Feeder, activate | 0 of 2177; 0 of 13 | 9 of 1822; 4 of 13 |
| Vampiric Rites, activate | 0 of 1004; 0 of 8 | 3 of 603; 1 of 7 |
| Warren Soultrader, activate | 0 of 1821; 0 of 19 | 8 of 1781; 4 of 17 |
| Village Rites (cast) | 0 of 22 games | 11 of 19 games |
| Altar's Reap (cast) | 0 of 12 games | 8 of 14 games |
| Deadly Dispute (cast) | 13 of 22 games | 19 of 24 games |
| Ashnod's Altar, Phyrexian Altar (cast) | 0 of 15; 0 of 13 games | 0 of 17; 0 of 13 games |

- The outlets are now used, but sparingly. Their sacrifices are mostly into a sweep or a losing combat, or of a token while several payoffs are out.
- The Altars are still never cast. Each is a mana source whose ability sacrifices a creature, so §2 does not count it as repeatable and §7 does not price it. Their mana abilities are not activated for floating mana (§7).
- Blood Artist is never cast either (0 of 23 games). Its 0/1 body plus one triggered row is worth less than the card it costs. §7 prices a payoff while it is on the battlefield, not when it is cast.

**A6 still fails, and this PR does not move it.** Mono-black's interval (53.4%–76.1%) lies entirely above 50%. PR 7's run 1 had the same result: 64.1%, 51.8%–74.7%. Death payoffs did not push black higher in this run.

**Run 2:** the same shape and seeds as PR 2's. 96 games, 0 stalls, 2 rejected moves (both `heuristic`, half A), turns p50 15 and 12 by half.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval |
|---|---:|---:|---:|---|
| heuristic | 192 | 55 | 28.6% | 22.7%–35.4% |
| heuristic-baseline | 192 | 41 | 21.4% | 16.1%–27.7% |

It is not detectably worse: the `heuristic` interval's upper bound, 35.4%, is above 25%. By half, `heuristic` won 11 of 96 seat-games on esper and izzet, and 44 of 96 on black and simic.

The four rejected moves across runs 1 and 2 are two attacks submitted after the step moved on and two passes that raced a Rhystic Study prompt, all in the concurrent schedule. None is a sacrifice.

**Before PR 7 merged.** The same runs on `6f73abab8` plus this PR gave:

- run 1: black 66.7% (54.4%–77.1%), black's `never` count falling from 11 to 4, and 1 stall;
- run 2: `heuristic` 28.1% (22.2%–34.9%) against `heuristic-baseline` 21.9% (16.6%–28.2%).

The stall was seed 21, the Sanguine Bond and Exquisite Blood loop at the CR 732 loop breaker (#2450), with the black seat at 75 life.

**Suite:** `boteval suite run --policy heuristic` gives 37 of 37, every tag at 100%:

| Tag | Positions |
|---|---:|
| activate | 3 |
| attack | 4 |
| block | 4 |
| cast | 19 |
| choice | 1 |
| combat | 8 |
| discard | 2 |
| land | 4 |
| leftover | 3 |
| mulligan | 6 |
| removal | 1 |
| sacrifice | 2 |
| wipe | 2 |

Two positions are new and gated for `heuristic`, harvested from logged arena runs on this branch (seeds 101 and 201):

- `sacrifice-the-creature-the-removal-targets`. An opponent's Acidic Slime trigger targets one of three Phyrexian Wurms, and the bot sacrifices that Wurm to Viscera Seer. `BaselineConfig()` passes.
- `do-not-sacrifice-the-commander`. Syr Konrad, Zulaport Cutthroat and Bastion of Remembrance are out and an opponent is at 3. The bot sacrifices its summoning-sick Human Soldier token, which kills that opponent, and never Konrad. On PR 7's tree it casts Night's Whisper first, which the label also accepts. `BaselineConfig()` passes, which misses the kill.

**Owner label review (2026-10-06).** The owner approved both labels as written, and both are stamped `reviewer: krakenhavoc`, `reviewed_at: 2026-10-06`.

### Amendment: discard payoffs (2026-10-06)

`develop` at `427d3f702` (PRs 2 to 7) plus this PR's `76e22fc5e`. `DefaultConfig()` against `BaselineConfig()`: `PriceDiscardPayoffs` true (baseline false). No weight is new: a payoff is priced with PR 7's `TokenWeight` (0.50), `Weights.Power + Weights.Toughness` (1.45) per +1/+1 counter, and `DamageToOpponent` (0.30) per point to each live opponent. Arena runs use the concurrent schedule, turn budget 60.

**Suite:** `windfall-discarding-a-spare-land` is relabelled as the owner asked. The Island is the one accepted move. The Mountain is unlabelled. Breeches, Negate and Bident of Thassa stay rejected. Breeches is a Pirate and would make the same Treasure, but it is a castable four-mana creature against a spare land, and a 0.50 Treasure does not close that gap. The position is stamped `reviewer: krakenhavoc`, `reviewed_at: 2026-10-06`. Six positions hold a card that now declares a payoff: this one, `do-not-windfall-away-the-last-land`, `loot-at-the-end-step-before-yours`, `cantrip-with-leftover-mana`, `hold-the-signet-late` and `wrath-a-losing-board`. Each frozen view had the `discard_payoff` the server now projects stamped onto that row, and nothing else changed. `boteval suite run --policy heuristic` gives 35 of 35, every tag at 100% (activate 1, attack 4, block 4, cast 19, choice 1, combat 8, discard 2, land 4, leftover 3, mulligan 6, removal 1, wipe 2). The heuristic now picks the Island. `BaselineConfig()` ranks every position as recorded in `baseline_rankings.json`.

**Run 1:** the same command and seed as PR 2's, with `--decision-log`. 64 games, 0 stalls, 14 rejected moves (10 stale `declare_attacker`, 3 passes over an unanswered Rhystic Study prompt, 1 cast), turns p50 14.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` (PR 7) |
|---|---:|---:|---|---:|
| esper-control | 5 | 7.8% | 3.4%–17.0% | 1 (1) |
| izzet-aggro | 4 | 6.2% | 2.5%–15.0% | 3 (3) |
| mono-black-aristocrats | 40 | 62.5% | 50.3%–73.3% | 10 (10) |
| simic-ramp | 15 | 23.4% | 14.7%–35.1% | 2 (2) |

The `never` lists are PR 7's, card for card. A6 fails as it does on `develop`: black's interval lies entirely above 50%.

**This amendment's class:**

- Mary Read and Anne Bonny's loot (A3): 48 of 62 games, 77% (PR 7: 41 of 62, 66%). With an Island in hand, the loot is now also priced by the Treasure it makes.
- Discards made with Mary Read on the bot's own battlefield, from the run 1 decision log: a matching card (Island, Pirate or Vehicle) was offered in 72 discard windows and discarded in 22, 31%. On PR 7's run 1 log it was 8 of 44, 18%. Most of the windows where none was discarded offered only Pirate creatures, which the bot keeps. In the rest the bot was short of lands and kept the Island over a cheap rock or spell, which a land worth 1.50 in hand outweighs.
- The other payoff cards are cast as before: Marauding Mako 18 of 21 games, Scrounging Skyray 15 of 16, Glint-Horn Buccaneer 16 of 17, Magmakin Artillerist 11 of 13.

Other canaries: Sol Ring 90–100%, Rhystic Study 11 of 17 (esper) and 18 of 21 (simic), Entomb 13 of 18, Viscera Seer (cast) 22 of 22, Harrow 8 of 25.

**Run 2:** the same shape and seeds as PR 2's. 96 games, 0 stalls, turns p50 14 and 12 by half.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves |
|---|---:|---:|---:|---|---:|
| heuristic | 192 | 52 | 27.1% | 21.3%–33.8% | 2 |
| heuristic-baseline | 192 | 44 | 22.9% | 17.5%–29.4% | 0 |

Not detectably worse: the `heuristic` interval's upper bound is 33.8%, above 25%. PR 7 measured 53 of 192. By half, `heuristic` won 10 of 96 seat-games on esper and izzet against the baseline's black and simic, and 42 of 96 on black and simic against the baseline's esper and izzet.

**Re-measured on `develop` with PR 8.** PR 8 (#2459) merged while this PR was open, so the suite and runs 1 and 2 were repeated on the merge of `develop` `835bd475e` into this branch (`58464780b`), with the same commands and seeds. `BaselineConfig()` zeroes both PRs' terms: `SacrificeDyingAnyway`, `DeathPayoff` and `PriceDiscardPayoffs`.

Suite: 37 of 37 agree, every tag at 100% (activate 3, attack 4, block 4, cast 19, choice 1, combat 8, discard 2, land 4, leftover 3, mulligan 6, removal 1, sacrifice 2, wipe 2).

Run 1: 64 games, 0 stalls, 15 rejected moves (14 stale `declare_attacker`, 1 pass), turns p50 13.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` (PR 8) |
|---|---:|---:|---|---:|
| esper-control | 5 | 7.8% | 3.4%–17.0% | 1 (1) |
| izzet-aggro | 6 | 9.4% | 4.4%–19.0% | 3 (3) |
| mono-black-aristocrats | 40 | 62.5% | 50.3%–73.3% | 4 (4) |
| simic-ramp | 13 | 20.3% | 12.3%–31.7% | 2 (2) |

The `never` lists are PR 8's, card for card. A6 still fails: black's interval lies entirely above 50%.

- Mary Read and Anne Bonny's loot: 49 of 61 games, 80%.
- With Mary Read on the bot's own battlefield, a matching card was offered in 72 discard windows and discarded in 21, 29%. PR 7's log gave 18%.
- The payoff cards are cast as before: Marauding Mako 19 of 21 games, Scrounging Skyray 15 of 16, Glint-Horn Buccaneer 14 of 16, Magmakin Artillerist 10 of 12.
- Other canaries: Sol Ring 85–100%, Rhystic Study 10 of 17 (esper) and 18 of 22 (simic), Entomb 11 of 15, Viscera Seer (cast) 23 of 23, Harrow 8 of 26.

Run 2: 96 games, 0 stalls, turns p50 14 and 12 by half.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves |
|---|---:|---:|---:|---|---:|
| heuristic | 192 | 55 | 28.6% | 22.7%–35.4% | 3 |
| heuristic-baseline | 192 | 41 | 21.4% | 16.1%–27.7% | 0 |

Not detectably worse: the upper bound is 35.4%. PR 8 measured the same 55 of 192. By half, `heuristic` won 11 of 96 on esper and izzet, and 44 of 96 on black and simic.

### PR 9: the exit (2026-10-06)

`develop` at `e84f04622`: PRs 2 to 8 and the discard-payoff amendment, with nothing added. `DefaultConfig()` against `BaselineConfig()` is the union of the rows above:

- `ManaPerExtra` 1.00, `RampPerMana` 1.00, `RampWantCap` 7;
- `PermanentPerMana` 0.50, `RowTriggered` / `RowStatic` / `RowActivated` 0.60 / 0.50 / 0.40, `RowCap` 3;
- `LeftoverWindows` on, `LeftoverThreshold` 0.00, `SpellFloor` 1.30, `TapByTiming` on;
- `PricePurposes` on, `TutorWeight` 1.00, `SelfMillWeight` 0.50, `DiscardWeight` 0.60, `TokenWeight` 0.50, `PriceSweeps` on, `DiscardCostByCard` on, `LastLandDiscard` 1.00;
- `SacrificeDyingAnyway` on, `DeathPayoff` 0.60;
- `PriceDiscardPayoffs` on.

Every term is zero or off in `BaselineConfig()`. No weight moved from §9's starting values in S66. Arena runs use the concurrent schedule, turn budget 60.

**Run 1:** the command in §8, seed 1. Run ID `2026-10-06T23:54:04.143426041Z`. 64 games, 0 stalls, 4 rejected moves, turns p50 13.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | `never` (PR 2) | `never` cards |
|---|---:|---:|---|---:|---|
| esper-control | 5 | 7.8% | 3.4%–17.0% | 1 (19) | Commander's Sphere |
| izzet-aggro | 6 | 9.4% | 4.4%–19.0% | 3 (22) | Lotus Petal; Glint-Horn Buccaneer's and Professional Face-Breaker's activations |
| mono-black-aristocrats | 38 | 59.4% | 47.1%–70.5% | 4 (29) | Ashnod's Altar, Phyrexian Altar, Blood Artist; Burnished Hart's activation |
| simic-ramp | 15 | 23.4% | 14.7%–35.1% | 2 (24) | Tarmogoyf; Sakura-Tribe Elder's sacrifice |

Canaries (A3), games used out of games offered:

| Canary | Deck | Used | Rate | PR 2 |
|---|---|---:|---:|---:|
| Sol Ring | esper, izzet, black, simic | 23 of 24, 17 of 17, 14 of 14, 19 of 21 | 90%–100% | 0% |
| Rhystic Study | esper | 10 of 17 | 59% | 0% |
| Rhystic Study | simic | 18 of 23 | 78% | 0% |
| Mary Read and Anne Bonny's loot | izzet | 46 of 61 | 75% | 0% |
| Entomb | black | 11 of 16 | 69% | 0% |
| Harrow | simic | 8 of 27 | 30% | 0% |
| Viscera Seer (cast) | black | 23 of 23 | 100% | 0% |

Mana rocks and dorks (A2), measured as the owner decided on #2435 (below). The same command and seed were run again with `--decision-log --decision-log-mode all` (run ID `2026-10-07T00:00:13.141101708Z`, `run1log` below), because the deficit is read off each window's view. Over the games in which each rock or dork was offered while the bot's mana deficit was open, it was cast in **210 of 311 (68%)**. PR 7 measured 214 of 323 (66%) the same way.

- **At or above 80%:** Sol Ring 32 of 32, Llanowar Elves 14 of 14, Fyndhorn Elves 6 of 6, Birds of Paradise 11 of 11, Delighted Halfling 7 of 7, Palladium Myr 10 of 11, Hedron Archive 7 of 8, Elvish Mystic 5 of 6, Ornithopter of Paradise 8 of 10, Talisman of Curiosity 8 of 10.
- **Below 80%:** Worn Powerstone 10 of 13, Thought Vessel 13 of 20, Simic Signet 5 of 8, Mind Stone 21 of 35, Arcane Signet 20 of 40, Dimir Signet 7 of 14, Talisman of Creativity 6 of 13, Azorius Signet 4 of 9, Orzhov Signet 3 of 7, Izzet Signet 4 of 10, Commander's Sphere 7 of 21, Talisman of Progress 2 of 6.
- **Why they stop short.** In every one of the 571 windows where a rock was offered with the deficit open and not taken, the bot did something else in that window: it played a land in 275, cast another spell in 295, and activated an ability in 1. It never passed over one. The rock's own price there was positive (+0.80 for a two-mana Signet in 446 of them). By its next window, the land and the other spell had usually closed the deficit, and a rock with no deficit is priced below zero on purpose (§2's "Arcane Signet, turn 9" row). Casting the rock first and the spell a turn later is a turn plan, which the heuristic does not make.

**Run 2:** the shape in §8, seeds 1 to 48 in both halves. Run IDs `2026-10-06T23:54:04.114189398Z` (half A: `heuristic` on esper and izzet) and `2026-10-06T23:54:04.17840261Z` (half B, swapped). 96 games, 0 stalls, turns p50 14 and 12 by half.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | Rejected moves |
|---|---:|---:|---:|---|---:|
| heuristic | 192 | 54 | 28.1% | 22.2%–34.9% | 0 |
| heuristic-baseline | 192 | 42 | 21.9% | 16.6%–28.2% | 1 |

By half, `heuristic` won 10 of 96 seat-games on esper and izzet, and 44 of 96 on black and simic.

**Suite:** `boteval suite run --policy heuristic`: 37 of 37 agree, every tag at 100% (activate 3, attack 4, block 4, cast 19, choice 1, combat 8, discard 2, land 4, leftover 3, mulligan 6, removal 1, sacrifice 2, wipe 2). Every S66 position is gated for `heuristic`, and all thirteen labels the owner reviewed are stamped.

#### The acceptance bar

| # | Bar | Result | Evidence |
|---|---|:--:|---|
| A1 | `never` ≤ 5 for every deck (run 1) | **PASS** | 1, 3, 4 and 2 (PR 2: 19, 22, 29 and 24) |
| A2 | each rock and dork used in ≥ 80% of the games it was offered in, counting offers made while the deficit is open | **FAIL** | 210 of 311 overall (68%); 10 of 22 cards meet 80%, 12 do not |
| A3 | each canary used in ≥ 50% of the games it was offered in | **FAIL** | five of six pass (Sol Ring 90%–100%, Rhystic Study 59% and 78%, the loot 75%, Entomb 69%, Viscera Seer 100%); Harrow 8 of 27, 30% |
| A4 | both wipe positions pass, gated | **PASS** | `do-not-wrath-your-winning-board` and `wrath-a-losing-board` agree, gated for `heuristic` |
| A5 | `heuristic`'s run 2 interval lies above 25% | **FAIL** | 54 of 192, 28.1%, 22.2%–34.9% |
| A6 | no deck's run 1 interval lies entirely above 50% | **PASS, narrowly** | mono-black 47.1%–70.5%. Not robust: see below |
| A7 | every gated position passes; no tag falls | **PASS** | 37 of 37, every tag 100% |
| A8 | 0 stalls; rejected moves not above the baseline run's; turns p50 within ±3 | **FAIL on rejected moves** | 0 stalls in runs 1 and 2. Turns p50 13 (run 1) and 14 / 12 (run 2), against PR 2's 14. Rejected moves: run 2 has 0 for `heuristic` against 1 for `heuristic-baseline`, but run 1 has 4 against PR 2's 2 |

**A6 is not robust.** Run 1 again with the same seeds and the decision log on (`run1log`) gave mono-black 40 of 64, 62.5% (50.3%–73.3%), which fails. Run 1's shape on seeds 65 to 128 gave 38 of 64, 59.4% (47.1%–70.5%), which passes. Pooled over the three runs, mono-black won 116 of 192 seat-games, 60.4% (53.4%–67.1%): its interval lies entirely above 50%. Black's lead is the deck's, not S66's. `heuristic-baseline` on black won 30 of 48 seat-games, 62.5%, in run 2's half A, without casting a single enchantment engine, and PR 2's run 1 had black at 59.4%.

**What drives black's lead.** From run 1's per-seat card counts:

- Syr Konrad's mill: 139 activations in 33 of 56 games where it was offered. PR 2 had 222 in 44 of 61. Black won 16 of the 31 games (52%) in which Konrad did not mill.
- Sanguine Bond: black won 11 of the 13 games in which it cast it (85%), and 27 of the 51 in which it did not (53%). It cast both halves of the drain loop in 4 games and won all 4.
- Death payoffs: black won 28 of 43 games with one or two cast (65%), and 9 of 16 with three or more (56%).

These are correlations within a deck that wins most of its games either way; a longer game sees more of every card.

**An experiment, not committed.** `ActivateBase` 0.25 in place of 0.50, run 1's shape on seeds 1 to 64 and 65 to 128. At 0.25, a purposeless non-tap activation on an untapped creature prices at −0.05 and on a tapped one at 0.25, so it clears neither `PassThreshold` nor `LeftoverThreshold` on its own.

| Run 1 shape | `ActivateBase` 0.50 (develop) | `ActivateBase` 0.25 |
|---|---|---|
| Black, seeds 1–64 | 38 of 64, 59.4% | 42 of 64, 65.6% (53.4%–76.1%) |
| Black, seeds 65–128 | 38 of 64, 59.4% | 39 of 64, 60.9% (48.7%–71.9%) |
| Konrad mills, seeds 1–64 / 65–128 | 139 / 136 | 89 / 87 |
| `never`, esper / izzet / black / simic, seeds 1–64 | 1 / 3 / 4 / 2 | 2 / 9 / 7 / 3 |

Halving `ActivateBase` cut Konrad's mills by about a third and did not lower black's win rate, 81 of 128 against 76 of 128. It also added twelve entries to the `never` lists, every one an activated ability with no declared purpose: Mind Stone's in all four decks, Commander's Sphere's in izzet and black, Hedron Archive's, and Krenko, Mob Boss's, Marauding Mako's, Magmakin Artillerist's, Scrounging Skyray's and Solphim's. Konrad's mill is not what makes black win.

**Run 2 on fresh seeds.** Run 2's shape twice more, on seeds 49 to 96 and 97 to 144:

| Seeds | `heuristic` | Wilson 95% interval | `heuristic-baseline` |
|---|---|---|---|
| 1–48 (run 2) | 54 of 192, 28.1% | 22.2%–34.9% | 42 of 192, 21.9% |
| 49–96 | 58 of 192, 30.2% | 24.2%–37.0% | 38 of 192, 19.8% |
| 97–144 | 53 of 190, 27.9% | 22.0%–34.7% | 42 of 190, 22.1% |
| pooled | 165 of 574, 28.7% | **25.2%–32.6%** | 122 of 574, 21.3% |

Every block points the same way, and pooled over 288 games the `heuristic` interval lies above 25%. One game in the 97 to 144 block (seed 141) stalled at the CR 732 loop breaker after a drain loop eliminated two seats, the #2450 shape.

**Rejected moves.** Every rejection in these runs is one of three kinds:

- A pass decided just before another seat's answer opened a blocking prompt: Rhystic Study's "draw a card?" after the payer declined, or Sun Titan's trigger. The table is fine and the runner decides again. These exist because the bot now casts Rhystic Study. Run 1's four rejections are all of this kind.
- An attack declared after the active seat had passed priority in declare attackers. The enumerator still offers attacks then, and Layer A takes one as the only legal move. In `run1log`, 28 such attacks were accepted, in 23 of 64 games, and 6 were refused because the step had moved on. Filed as #2462.
- A cast refused for insufficient mana. The auto-tapper books one slot of a bounce land twice (Simic Growth Chamber's `{G}{U}` as `{U}{U}`), and a refused strict cast leaves the land tapped and its mana floating. Filed as #2461, with a reproduction.

**For the owner.** A1, A4 and A7 pass. A6 passes on the run §8 names, but not robustly. A2, A3, A5 and A8 fail, with the evidence above. A5 passes on the 288 pooled games. The owner decided them on 2026-10-07: see [Exit decisions](#exit-decisions-2026-10-07).

### #2436: the rebalanced decks (2026-10-07)

Not an S66 sub-PR: the A6 follow-up the owner left to #2436. `develop` at `5e4bb1dca` (with #2461's and #2462's fixes) plus the rebalanced lists, Y'shtola, Night's Blessed for Esper, and the purposes the incoming cards declare. `DefaultConfig()` and `BaselineConfig()` are PR 9's; no weight moved. Concurrent schedule, turn budget 60.

**Run 1:** the command in §8, seed 1. Run ID `2026-10-07T11:33:23.994475731Z`. 64 games, 0 stalls, 2 rejected moves (both a pass racing Rhystic Study's "draw a card?"), turns p50 15.

| Deck | Wins of 64 | Win rate | Wilson 95% interval | PR 9 | `never` (PR 9) | `never` cards |
|---|---:|---:|---|---:|---:|---|
| esper-control | 29 | 45.3% | 33.7%–57.4% | 5 | 0 (1) | — |
| izzet-aggro | 6 | 9.4% | 4.4%–19.0% | 6 | 3 (3) | Professional Face-Breaker's, Scrounging Skyray's and Solphim's activations |
| mono-black-aristocrats | 21 | 32.8% | 22.6%–45.0% | 38 | 4 (4) | Ashnod's Altar, Blood Artist, Commander's Sphere; Burnished Hart's activation |
| simic-ramp | 8 | 12.5% | 6.5%–22.8% | 15 | 2 (2) | Burnished Hart's and Mind Stone's activations |

A6 passes with room: no deck's interval lies above 50%, and mono-black's no longer lies above 25%. The issue's further mono-black trims (Demonic Tutor, Ancient Tomb, Cabal Coffers with Urborg, Sheoldred) were not made. Izzet's and Simic's intervals lie entirely below 25%.

**Run 2:** §8's shape, seeds 1 to 48 in both halves. Run IDs `2026-10-07T11:33:23.987898191Z` (half A: `heuristic` on esper and izzet) and `2026-10-07T11:33:24.033165755Z` (half B). 96 games, 0 stalls, 1 rejected move (`heuristic`, the same Rhystic Study race), turns p50 14 and 13.

| Policy | Seat-games | Wins | Win rate | Wilson 95% interval | PR 9 |
|---|---:|---:|---:|---|---|
| heuristic | 192 | 58 | 30.2% | 24.2%–37.0% | 54, 28.1% |
| heuristic-baseline | 192 | 38 | 19.8% | 14.8%–26.0% | 42, 21.9% |

By half, `heuristic` won 18 of 96 seat-games on esper and izzet (esper 17, izzet 1) and 40 of 96 on black and simic (black 23, simic 17).

### #2469: a land swap priced by what it nets (2026-10-08)

The owner chose, on 2026-10-08, a §6 price change for Harrow after #2649 gave its land sacrifice the leftover-window bar. A cast that sacrifices lands of the bot's own and whose `purpose.lands` is larger than the number sacrificed is priced as a swap plus a ramp spell (`NetLandSwaps`, `land_swap.go`): each sacrificed land costs its own value and is given back as an untapped land at `ManaSource`, and only the lands beyond those are ramp, with the ramp premium and `SpellFloor` under that net amount. Late in a game Harrow goes from −0.20 to +0.10 (an untapped land sacrificed) or +0.55 (a tapped one), at or above a Rampant Growth's +0.10. `BaselineConfig` turns it off. In the catalog only Harrow declares that shape: Roiling Regrowth, Cycle of Renewal and Entish Restoration declare no purpose (their lands enter tapped, which a purpose cannot yet say), and Crop Rotation swaps one land for one.

Before is `develop` at `c1391ff97` (#2649 merged), after is the branch; every run is `--rotate --lockstep` with the real dump, so a seed replays the same game until the price changes it. No run stalled.

| Run | Harrow before | Harrow after |
|---|---|---|
| simic-ramp ×4, seeds 1–40 | 19 / 49, 38.8% (26.4%–52.8%) | 20 / 49, 40.8% (28.2%–54.8%) |
| simic-ramp ×4, seeds 1000–1119 | 43 / 166, 25.9% (19.8%–33.1%) | 53 / 164, 32.3% (25.6%–39.8%) |
| simic-ramp ×4, pooled | 62 / 215, 28.8% (23.2%–35.2%) | 73 / 213, 34.3% (28.2%–40.9%) |
| Run 1 (§8, 64 games, seed 1) | 12 / 27, 44.4% (27.6%–62.7%) | 15 / 27, 55.6% (37.3%–72.4%) |
| Run 2 half B (`heuristic` on black and simic, 48 games) | 6 / 17, 35.3% | 7 / 17, 41.2% |

Win rates do not move beyond a game. Run 1: esper 27 and 27 of 64, izzet 3 and 2, black 22 and 24, simic 12 and 11. Run 2 half B: `heuristic` 42 of 96 seat-games both times (43.8%, 34.3%–53.7%), `heuristic-baseline` 6 of 96 both times; by deck, black 22 then 21, simic 20 then 21. The simic ×4 runs are at the null by construction. The other A2 and A3 rows are within one game in run 1 and run 2. In the simic ×4 runs, where a changed Harrow decision changes the rest of the game, they move by up to eight games either way along with their offered counts (Ornithopter of Paradise 137 of 168 to 145 of 177, Delighted Halfling 155 of 181 to 152 of 184), and the only row that crosses its bar is Delighted Halfling in the 40-game run, upward (76% to 82%). The suite is 37 of 37.

Run 1 now meets A3's 50% for Harrow, but the 160-game simic ×4 pool, at 34%, does not. A decision log of six games shows why the rest of the windows still pass: Harrow is priced positive in nearly every window it is offered in, and it is refused where the bar is `InstantThreshold`, with a trigger on the stack in the bot's own main phase or in its upkeep and draw steps, or it loses the main phase to a bigger cast that taps the bot out. It is rarely offered in the end step before the bot's turn, because the bot has spent its mana by then. Those are sequencing questions, not Harrow's price.

### #2675, #2690, #2676: combat priced by what combat does (2026-10-08)

Three fixes from the 2026-10-08 review games, each a knob that `BaselineConfig` turns off: `FocusNeedsValue` (the focus bonus only on an attack whose own value is positive, and an attacker with no power is not a blocker the defender must spend), `GangAwareAttacks` (an attack priced against any group of free blockers that kills it, a dying commander charged `CommanderTax`, and a creature tapped in the bot's first main phase charged its attack at `attackValue`), and `Weights.BlockOnlyBody` (0.10 per point of toughness for a token with no power and no abilities, in place of its body price in combat). This is combat, which §Out of scope left alone; `CombatValue` stays body-only (owner decision 3), since `BlockOnlyBody` reads only the body and the token flag.

Before is `develop` at `da81f844d`, after is the branch; every run is `--rotate --lockstep` with the real dump, seed 1, 0 stalls and 0 rejected moves in every run.

| Run | Contestant | Before | After |
|---|---|---|---|
| Run 1 (§8, 64 games) | esper-control | 27, 42.2% (30.9%–54.4%) | 25, 39.1% (28.1%–51.3%) |
| Run 1 | izzet-aggro | 2, 3.1% (0.9%–10.7%) | 4, 6.2% (2.5%–15.0%) |
| Run 1 | mono-black-aristocrats | 24, 37.5% (26.7%–49.7%) | 17, 26.6% (17.3%–38.5%) |
| Run 1 | simic-ramp | 11, 17.2% (9.9%–28.2%) | 18, 28.1% (18.6%–40.1%) |
| Run 2 (izzet and simic, 48 games) | heuristic | 24 / 96, 25.0% (17.4%–34.5%) | 28 / 96, 29.2% (21.0%–38.9%) |
| Run 2 | heuristic-baseline | 24 / 96, 25.0% (17.4%–34.5%) | 20 / 96, 20.8% (13.9%–30.0%) |

Run 1's turns p50 goes from 15 to 14; run 2's stays 12. The counters, from the decision logs (run 1, per game; run 2, the `heuristic` seats over 48 games):

| Counter | Run 1 before | after | Run 2 before | after |
|---|---:|---:|---:|---:|
| Attacks with a 0-power creature | 60 (0.94) | 5 (0.08) | 40 | 18 |
| Attackers blocked by two or more and lost | 41 of 76 | 7 of 23 | 7 of 17 | 1 of 7 |
| Losing blocks (the blocker dies, the attacker lives) | 186 | 161 | 155 | 118 |
| of them with a 0-power token | 3 | 5 | 0 | 1 |
| "No blocks" with an untapped 0-power token and an attack incoming | 5 | 1 | 0 | 0 |

The 0-power attacks left are all lethal pushes, two-turn races and attrition plans, which send every body by design. The curated decks make few 0-power tokens, so the chump counter barely moves; the issue's window is pinned by a unit test instead. The A3 canaries in run 1 move by at most two games: Mary Read's loot 44 of 61 to 49 of 61, Harrow 15 of 27 (56%) to 13 of 27 (48%), which crosses the bar downward by two games; the others stay above it. The suite is 41 of 41 before and after.

### #2677 and #2691: land searches by colour, discards by distance (2026-10-08)

Two card choices the 2026-10-08 review games found wrong, both local to the decision they fix (`card_choices.go`); `cardValue`, which the scry, the sacrifice, the fuel pricer and the cast-cost discard read, is unchanged.

- **#2677, `LandColorNeed` (0.30).** A library search scored every land at one flat `cardValue` and took the first. A land now adds `LandColorNeed` × its fit: for each colour its repeatable mana abilities make, 1/(1 + the bot's sources of it) when the hand or commander has a pip of it, and 0.1 when nothing does, so a dual beats a basic that meets the hand equally. The same score answers a `choose_cards` look at the library.
- **#2691, `DiscardByDistance`, `DistanceDiscount` (0.60), `DiscardSpellPerMana` (1.00), `DiscardLandFloor` (2.50).** A discard from the bot's own hand (cleanup, a discard prompt, a loot's or rummage's named cards, and the discard-payoff estimate that mirrors them) prices a nonland card by `handKeepValue`: an instant or sorcery at `DiscardSpellPerMana` per mana, a permanent at its `permanentValue`, either multiplied by `DistanceDiscount` per mana it is short (its mana value less sources and lands in hand, or its coloured pips less the sources and lands in hand of each colour, whichever is larger). A land is worth what it brings the rest of the hand closer to castable, never less than its `cardValue`, and never less than `DiscardLandFloor` while the bot has fewer than `RampWantCap` sources. The floor came from measurement: without it the bot discarded 243 lands in run 1 where it had discarded 142, and played no land on 27.1% of its own turns against 24.1%.

`BaselineConfig` zeroes all five. Before is `develop` at `36d0e9e4c` (with #2710's combat fixes), after is the branch merged onto it; every run is `--rotate --lockstep` with the real dump. No run stalled and no move was rejected. The counters come from a scratch observer on the runner's decision feed, not part of the change: a search counts when every option is a land and some option makes a colour the hand needs and has no source of while another does not; a discard counts when the nonland cards in hand differ in distance and at least one is short.

| Run | Measure | Before | After |
|---|---|---|---|
| Run 1 (§8, 64 games, seed 1) | search took a colour the hand needs | 27 / 64 | 55 / 55 |
| | search took a dual over a basic, need equal | 52 / 150 | 134 / 136 |
| | discard took the card furthest from castable | 0 / 56 | 17 / 43 |
| | own turns with no land played | 798 / 3267, 24.4% | 850 / 3267, 26.0% |
| | lands among the cards discarded | 133 / 581 | 205 / 633 |
| | turns p50 | 14 | 15 |
| Run 2, seed 1 | `heuristic` wins | 31 / 96, 32.3% (23.8%–42.2%) | 28 / 96, 29.2% (21.0%–38.9%) |
| | `heuristic-baseline` wins | 17 / 96 | 20 / 96 |
| Run 2, seed 1001 | `heuristic` wins | 24 / 96, 25.0% (17.4%–34.5%) | 24 / 96, 25.0% (17.4%–34.5%) |
| | `heuristic-baseline` wins | 24 / 96 | 24 / 96 |
| Run 2, pooled | `heuristic` wins | 55 / 192, 28.6% (22.7%–35.4%) | 52 / 192, 27.1% (21.3%–33.8%) |
| | `heuristic-baseline` wins | 41 / 192, 21.4% | 44 / 192, 22.9% |

Run 2 is `--seats heuristic,heuristic-baseline,heuristic,heuristic-baseline --decks izzet-aggro,izzet-aggro,simic-ramp,simic-ramp --games 48`, so each policy plays each deck 48 times; its turns p50 is 11 then 12 at seed 1 and 12 both times at seed 1001. In run 2 `heuristic` took the needed colour in 20 of 20 and 18 of 18 searches (8 of 21 and 6 of 19 before), and played no land on 18.3% and 16.9% of its turns (18.3% and 17.6% before). The pooled difference is three games of 192 either way. The same runs on the base before #2710 (`da81f844d`) gave 55 of 192 before and 55 after, with the baseline at 41 both times.

Run 1's deck shares moved: esper 26 to 20 of 64, izzet 5 to 7, black 18 to 25, simic 15 to 12. Four copies of one policy are zero-sum, so this measures no strength; run 2 is the strength measure. `never` counts are esper 0, izzet 2 to 3, black 4, simic 1 to 2. A2 and A3 rows move both ways with their offered counts, as in #2469's runs: in run 1, 17 of 41 rows met their bar before and 15 after; in run 2, 8 to 9 at seed 1 and 7 to 8 at seed 1001. Every A3 canary that met its bar in run 1 still does (Harrow, 11 of 26 before, is 12 of 27 after). One A2 row fell on both bases: Worn Powerstone, used in 18 of 20 games before and 8 of 14 after here, and 15 of 17 before and 9 of 13 after on the older base. Four run-1 games where black cast it before and not after (seeds 14, 22, 32 and 52) were replayed one by one with the decision log on, and the cause is not its price. In every window it is offered in, before and after, it is priced by the same rule (+2.8, +1.8 or +0.8 as the ramp deficit closes), and it is never discarded. It is the bot's last-choice play once the deficit closes, cast only in a main phase with nothing better to spend the mana on, and after the change that main phase comes later or not at all:

- Seed 22: the cleanup discards of turns 3 and 4 kept Exsanguinate and Ambition's Cost (pitched before) and pitched Rise of the Dark Realms, a nine-drop three mana short, and Dictate of Erebos. On turn 7 the bot cast Ambition's Cost (+2.4) where before, with neither card kept, it cast the Powerstone (+2.8).
- Seed 14: the table diverges from turn 3 on other seats' decisions. Demonic Tutor then takes Midnight Reaper, and from turn 6 every main phase has a castable creature or spell priced above the Powerstone's +0.8 (Midnight Reaper, Gray Merchant, Deadly Dispute, Sign in Blood, Read the Bones).
- Seed 32: black's decisions are the same as before, window by window, but the game ends at turn 11 instead of 16, before the late main phase in which it had cast the rock.
- Seed 52: black draws it on turn 6 instead of 8, taps out for its commander that turn, and an opponent's Wheel of Fortune discards its hand in the same round.

So the row measures how often black runs out of better plays before the game ends. The change gives it more of them (seeds 22 and 14), and it moves when the game ends and what the opponents do (seeds 32 and 52). Casting a card-draw spell or a creature over a tapped three-mana rock at seven mana is consistent with how §2 prices a rock once the deficit closes. Nothing in this change was adjusted for it.

The extra land discards in run 1 are late: in a 16-game diagnostic of the after build, 65 of 67 land discards came with seven or more mana sources on the battlefield. That is where the floor stops and a spare land is the right card to pitch, and it is what the rise in turns with no land played counts. A land drop offered and not taken stayed rare: 5 turns before and 1 after in run 1, none in run 2. The suite is 41 of 41 before and after, and no position's pick changed.

### Amendment PR 3: target purposes priced (2026-10-08)

`PriceTargetPurposes` (`target_purpose.go`), A1 and B1 of the [amendment of 2026-10-08](#amendment-2026-10-08-purposes-that-follow-a-modes-target-and-damage-priced-by-whether-it-kills). A player pick whose declared entry gives it cards, tokens or life is priced as that seat's strength change through `ScoreEval`, and the cast drops the mana proxy. Damage entries keep today's price until `DamageByLethality` (PR 4). `BaselineConfig` turns it off.

Before is `develop` at `32886c5cc` (PR 2 merged), after is the branch; every run is `--rotate --lockstep` with the real dump, 0 stalls in every run. Lockstep tie-breaks are not yet fully deterministic (#2730), so read the intervals.

| Run | Contestant | Before | After |
|---|---|---|---|
| Run 1 (§8, 64 games, seed 1) | esper-control | 26, 40.6% (29.5%–52.9%) | 23, 35.9% (25.3%–48.2%) |
| Run 1 | izzet-aggro | 5, 7.8% (3.4%–17.0%) | 4, 6.2% (2.5%–15.0%) |
| Run 1 | mono-black-aristocrats | 18, 28.1% (18.6%–40.1%) | 19, 29.7% (19.9%–41.8%) |
| Run 1 | simic-ramp | 15, 23.4% (14.7%–35.1%) | 18, 28.1% (18.6%–40.1%) |
| Run 2 (izzet and simic, 48 games, seed 1) | heuristic | 33 / 96, 34.4% (25.6%–44.3%) | 33 / 96, same |
| Run 2 | heuristic-baseline | 15 / 96, 15.6% | 15 / 96, same |
| Run 2, seed 1001 | heuristic | 31 / 96, 32.3% (23.8%–42.2%) | 31 / 96, same |
| Run 2, seed 1001 | heuristic-baseline | 17 / 96, 17.7% | 17 / 96, same |
| Targeted (izzet and black, 96 games) | heuristic | 64 / 192, 33.3% (27.0%–40.3%) | 64 / 192, same |
| Targeted | heuristic-baseline | 32 / 192, 16.7% | 32 / 192, same |

Run 1's turns p50 is 14 before and 15 after; run 2's is 12. A2 and A3 rows move by at most three games and none crosses its bar (Harrow stays below A3 at 11 of 26 and 10 of 23).

A head-to-head with the knob alone (today's heuristic against itself with `PriceTargetPurposes` off, a local build, izzet and simic, 48 games each at seeds 1 and 1001): with the knob 44 / 192, 22.9% (17.5%–29.4%); without 52 / 192, 27.1% (21.3%–33.8%). The intervals overlap and both contain the null; the direction is against the knob in both seeds, mostly on izzet at seed 1 (2 against 7 wins).

Who the cards were aimed at in the targeted run (after, from the decision logs; before, every cast aimed at an opponent by construction, as `heuristic-baseline`'s 23 Sign in Blood and 11 of 14 Prismari casts are):

| Contestant | Card | Aimed at | Casts |
|---|---|---|---:|
| heuristic | Sign in Blood | itself | 6 |
| heuristic | Sign in Blood | an opponent on 2 life (3) or 6 | 4 |
| heuristic | Prismari Command | 2 damage at an opponent, destroy an artifact | 8 |
| heuristic | Prismari Command | 2 damage at an opponent, loot itself | 3 |
| heuristic | Prismari Command | loot itself, destroy an artifact | 3 |
| heuristic | Prismari Command | loot itself, Treasure itself | 2 |
| heuristic-baseline | Prismari Command | 2 damage and loot, both at an opponent | 11 |
| heuristic-baseline | Prismari Command | 2 damage at an opponent, destroy an artifact | 3 |
| heuristic-baseline | Sign in Blood | an opponent | 23 |

`heuristic` cast Sign in Blood in fewer games (14 of 16 to 10 of 16; run 1, 11 of 12 to 6 of 11): at itself it is +0.96, not +2.40, so it loses more main phases to a creature. At seq 248 of review game 2, with PR 2's declarations put on the logged view, the chosen line moves from 2 at Y'shtola and loot Claude (9.12) to 2 at Y'shtola and loot the bot (6.72); the lines that loot Claude fall by 5.40 and those that give Claude the Treasure by 4.35. The cast stays above `InstantThreshold` because 2 damage at a 2/4 is still priced as removal (6.72), which is PR 4's to fix. The suite is 41 of 41 before and after.

### Amendment PR 4: damage priced by whether it kills, measured with PR 3 (2026-10-08)

The owner held PR 3 to measure it together with PR 4. `DamageByLethality` (`target_purpose.go`, `damageKills` shared with `combat.go`'s `kills`), C1 and D1 with owner answers 3, 4 and 6:
- A declared damage entry at a creature is removal if it kills and `DamageChip` (0.00) of removal if it does not.
- At a planeswalker it is the share of loyalty removed.
- At a player it is `DamageToOpponent` per point through the opposition weights, with `LethalBonus` at or above their life.

Arc Trail is now two clauses, 2 damage and 1 to another target, and declares both. This changes its moves: the second pick names slot 1. `BaselineConfig` turns the knob off.

Before is `develop` at `8ecec05ba`, and after is the branch with both knobs (develop merged in). Every run is `--rotate --lockstep` with the real dump, and every run had 0 stalls.

| Run | Contestant | Before | After |
|---|---|---|---|
| Run 1 (§8, 64 games, seed 1) | esper-control | 20, 31.2% (21.2%–43.4%) | 28, 43.8% (32.3%–55.9%) |
| Run 1 | izzet-aggro | 7, 10.9% (5.4%–20.9%) | 4, 6.2% (2.5%–15.0%) |
| Run 1 | mono-black-aristocrats | 25, 39.1% (28.1%–51.3%) | 20, 31.2% (21.2%–43.4%) |
| Run 1 | simic-ramp | 12, 18.8% (11.1%–30.0%) | 12, 18.8% (11.1%–30.0%) |
| Run 2 (izzet and simic, 48 games, seed 1) | heuristic | 32 / 96, 33.3% (24.7%–43.2%) | 37 / 96, 38.5% (29.4%–48.5%) |
| Run 2, seed 1 | heuristic-baseline | 16 / 96, 16.7% | 11 / 96, 11.5% |
| Run 2, seed 1001 | heuristic | 31 / 96, 32.3% (23.8%–42.2%) | 32 / 96, 33.3% (24.7%–43.2%) |
| Run 2, seed 1001 | heuristic-baseline | 17 / 96, 17.7% | 16 / 96, 16.7% |
| Targeted (izzet and black, 96 games) | heuristic | 62 / 192, 32.3% (26.1%–39.2%) | 57 / 192, 29.7% (23.7%–36.5%) |
| Targeted | heuristic-baseline | 34 / 192, 17.7% | 39 / 192, 20.3% |

Izzet-aggro over run 1 and both run 2 seeds: 20 of 160 before and 21 of 160 after. Run 1's turns p50 goes from 15 to 14, and run 2's stays at 12.

A2 and A3 rows move by up to 7 games, all on simic-ramp, which holds no declared target. Bars are crossed both ways:
- Run 1: Delighted Halfling and Rhystic Study fall below their bars, and Harrow rises above its bar (12 of 27 to 16 of 25).
- Run 2: Birds of Paradise and Ornithopter rise above their bars.
- Run 2, seed 1001: Delighted Halfling, Sol Ring and Rhystic Study rise above their bars.

That is a diverged game, not a price.

**Knob-alone head-to-head.** This is today's heuristic against a local build with both knobs off, on izzet and simic, 48 games each:
- seed 1: 25 against 23 wins
- seed 1001: 21 against 27 wins
- pooled: with the knobs 46 / 192, 24.0% (18.5%–30.5%); without 50 / 192, 26.0% (20.3%–32.7%)
- on izzet: 2 against 2 at seed 1, and 5 against 5 at seed 1001

**What the burn and the gifts were aimed at** (the targeted run's decision logs, after):

| Card | `heuristic` | `heuristic-baseline` |
|---|---|---|
| Lightning Bolt | 11 at creatures it killed | 5 killed, 6 at creatures that survived, 4 at players (1 lethal) |
| Shock | 14 killed | 2 killed, 8 survived, 4 at players |
| Fiery Temper | 16 killed, 2 survived | 10 killed, 11 survived, 2 at players |
| Izzet Charm (damage) | 19 killed | 6 killed, 11 survived |
| Arc Trail | 2 at an opponent with 1 killing a creature: 6; 2 at itself with 1 killing a creature: 4; lethal at an opponent: 1 | 3 killed, 6 survived, 3 at players only |
| Prismari Command | loot and Treasure at itself 5; loot itself with damage or artifact removal 7; 2 at an opponent with artifact removal 3 | 2 and loot both at an opponent 11 |
| Sign in Blood | at itself 7; at an opponent on 3 life or less 7 | at an opponent 21 |

Sign in Blood's cast rate (games used of games offered) is close to before:
- run 1: 11 of 11 before, 8 of 9 after
- targeted run: 16 of 17 before, 14 of 16 after

With PR 3 alone it was 6 of 11 and 10 of 16. PR 4 makes the burn and the bodies the bot would otherwise cast cheaper to hold, which leaves room in the main phase.

**Arc Trail at itself.** The 4 casts that put Arc Trail's 2 at the bot are the enumerator's doing. `legalStepSets` is a cartesian product capped at 12, in candidate order, and candidates are ordered by threat. So slot 0 takes the top-threat candidate in every offered set, and on some boards that candidate is the bot itself. In one logged window the only Arc Trail move offered was "2 at the bot, 1 at Fleshbag Marauder". This is a follow-up for the enumerator, like #2681. `heuristic-baseline` shows the same shape once.

**Seq 248 of review game 2**, with PR 2's declarations put on the logged view: every Prismari Command line is now below `InstantThreshold`. The best is "2 at Claude, loot the bot" at +0.90, and pass is 0, so the bot passes the draw step, as the amendment predicted. The table is in the PR. The suite is 41 of 41 before and after.

**Pins.** `TestArenaSeededGameIsTheSameGameAfterTheOpeningRollWindow` was re-pinned by hand (its fifth exception). The battle deck's Lightning Bolt at a player on 40 life is now held. With the knob off, the old digests still match.

**Real-dump audit.** The branch E2E's `realdump` job fails on `TestRealDumpPurposeAudit` over Eliminate the Impossible, a card from #2734 that reads as a wipe and declares no sweep. `develop` at `8ecec05ba` fails the same way. Arc Trail passes the audit.

### #2680 and #2678: puts from hand, own-permanent picks and extra land drops (2026-10-08)

Three things the 2026-10-08 review games found the heuristic pricing as nothing, or as the opposite of what they are (`puts.go`).

- **#2680, `PricePutsFromHand`.** "You may put a land card from your hand onto the battlefield" is a `choose_cards` over the bot's own hand, which `valueKeptInHand` prices as a discard of the named card, so the bot declined every one. The prompt now carries `choose_destination` (`battlefield`, or `battlefield_tapped`: an additive `PendingChoice.ChooseDestination`, recorded by the shape guard, withheld from non-choosers with the bounds), set by `PutFromHandOntoBattlefield`. A named land adds `ManaSource` (`TappedManaSource` when it enters tapped) and the ramp premium while the bot has fewer than `RampWantCap` sources, 0.3 of that after, plus `landColorFit`; any other permanent adds its resolved value.
- **#2680, `PriceOwnPermanentPicks`.** An `own_permanents` pick had no branch, so every answer scored 0 and the enumerator's cheapest-fuel-first order chose. A fixed-count pick now gives up what is worth least to keep: `permanentValue`, a land multiplied by the mana it makes and with its ability rows added. A pick whose count is the chooser's (Scapeshift, Tragic Arrogance's own leg) keeps the enumerator's order.
- **#2678, `PriceExtraLandDrops`, `ExtraLandDropRecurring` (0.50).** `purpose.extra_land_drops` is declared on every catalog card with `AdditionalLandPlays` (Register refuses a number that disagrees; `TestCuratedDeckPurposes` and `TestEveryExtraLandDropIsDeclared` hold the declarations), and on Explore. A cast adds `ManaSource` plus the ramp premium for each extra drop the bot holds a land for and could not otherwise play this turn, and, for a permanent, `ExtraLandDropRecurring` per drop while it has fewer than `RampWantCap` sources.

`BaselineConfig` zeroes all four. Before is `develop` at `c98668a15`, after is this branch; every run is `--rotate --lockstep` with the real dump and `--decision-log`. No run stalled and no move was rejected. The counters come from a scratch script over the decision logs.

| Run | Measure | Before | After |
|---|---|---|---|
| Run 1 (§8, 64 games, seed 1) | Uro's land put accepted | 0 / 72 | 81 / 81 |
| | Eureka Moment's land put accepted | 0 / 12 | 11 / 11 |
| | karoo returned itself | 78 / 96 | 0 / 43 |
| | karoo returned a tapped land when one was offered | 17 / 17 | 16 / 16 |
| | Oracle of Mul Daya cast, windows / games used of offered | 22 / 117, 22 / 28 | 25 / 82, 24 / 27 |
| | Exploration cast, windows / games used of offered | 15 / 92, 15 / 22 | 20 / 89, 18 / 25 |
| | land drop offered and not taken | 1 turn | 2 turns |
| | turns p50 | 15 | 14 |
| | simic-ramp wins | 12 / 64 | 21 / 64 |
| Run 2, seed 1 | `heuristic` wins | 32 / 96, 33.3% (24.7%–43.2%) | 30 / 96, 31.2% (22.9%–41.1%) |
| | `heuristic-baseline` wins | 16 / 96 | 18 / 96 |
| Run 2, seed 1001 | `heuristic` wins | 31 / 96, 32.3% (23.8%–42.2%) | 24 / 96, 25.0% (17.4%–34.5%) |
| | `heuristic-baseline` wins | 17 / 96 | 24 / 96 |
| Run 2, seed 2001 | `heuristic` wins | 23 / 96, 24.0% (16.5%–33.4%) | 26 / 96, 27.1% (19.2%–36.7%) |
| | `heuristic-baseline` wins | 25 / 96 | 22 / 96 |
| Run 2, pooled | `heuristic` wins | 86 / 288, 29.9% | 80 / 288, 27.8% |
| | `heuristic-baseline` wins | 58 / 288 | 64 / 288 |

Run 2 is `--seats heuristic-baseline,heuristic-baseline,heuristic,heuristic --decks izzet-aggro,simic-ramp,izzet-aggro,simic-ramp --games 48`, so each policy plays each deck 48 times per seed. The pooled difference is six games of 288, inside the run-to-run spread (seed 1001 moved seven games one way, seed 2001 three the other), and in no run does the baseline win more than `heuristic`. Turns p50 is 12 to 11 at seed 1 and 12 both times at seeds 1001 and 2001. Run 1's karoo returns before were the source itself in 78 of 96, because the enumerator offers the cheapest fuel first and a tapped karoo ties a tapped basic; after, with no other tapped land offered it returns an untapped land (27 times) rather than itself.

In run 1, every A3 canary meets its bar after (Harrow 12 / 27 before, 15 / 29 after, now meeting it). A2 rows meeting their bar fell from 6 to 4: Delighted Halfling (85% to 76%) and Ornithopter of Paradise (83% to 64%) in simic-ramp, whose early turns now also hold an Exploration or an Oracle priced above a body. In run 2 the met A2 and A3 rows went from 6 to 11 at seed 1, 8 to 10 at seed 1001 and 9 to 10 at seed 2001. The suite is 41 of 41 before and after, and no position's pick changed.

### #2679: removal priced net of what comes back (2026-10-08)

`NetRemoval` (on in `DefaultConfig`, off in `BaselineConfig`; `net_removal.go`), as decided in the [amendment of 2026-10-08, what removal hands back](#amendment-2026-10-08-what-removal-hands-back). An opposing commander its opponent owns is priced at `2 × CommanderTax + DamageToOpponent × power`, never above its value, because it returns from the command zone (CR 903.8, 903.9a, 903.9b). A target entry's new `returns` (a creature token of a printed size, life equal to the target's power, lands) is valued as the target's controller would value it and taken off the removal at `RemovalConfidence × leaderBoost`. The amendment's target entries could not say it: their amounts are the target's own, and the guard refuses a player amount on a clause that cannot target a player. So `TargetPurpose` gained `Returns`, additive on the wire as `targets[].returns`, declared on Rapid Hybridization, Pongify, Beast Within, Generous Gift, Stroke of Midnight, Swords to Plowshares, Path to Exile and Assassin's Trophy. Chaos Warp's return is a random card and declares nothing; its commander half is covered.

Review windows, re-ranked offline (game 1 with the declaration patched onto the logged view): game 1 seq 66, Rapid Hybridization on a 1/1 Archivist of Oghma, 3.45 to −1.77, so the bot passes. Game 2 seq 186, Chaos Warp on Y'shtola (a 2/4 commander), 6.65 to 3.72, below Mary Read and Anne Bonny from the command zone (4.17), which the bot now casts.

Before is `develop` at `50d5c34ec`, after is this branch; every run is `--rotate --lockstep` with the real dump. No run stalled. Counters come from a scratch pass over the decision logs; "worth less than the gift" compares the target's body (`CreatureValue`, or its board value) with the gift's value, both before `RemovalConfidence` and `leaderBoost`.

| Run | Measure | Before | After |
|---|---|---|---|
| Run 1 (64 games, seed 1) | gift-removal casts at a target worth less than its gift | 23 of 190 | 0 of 143 |
| | the listed removal (the eight, plus Chaos Warp) on an opposing commander | 59 | 13 |
| | any targeted cast at an opposing commander | 232 | 138 |
| | esper / izzet / black / simic wins | 18 / 3 / 21 / 22 | 21 / 1 / 22 / 20 |
| | turns p50 | 13 | 13 |
| Run 2, seed 1 | `heuristic` / `heuristic-baseline` wins | 29 / 19 | 29 / 19 |
| Run 2, seed 1001 | | 27 / 21 | 30 / 18 |
| Run 2, seed 2001 | | 24 / 24 | 21 / 27 |
| Run 2, seeds 1, 1001, 2001 | `heuristic` | 80 / 288, 27.8% (22.9%–33.2%) | 80 / 288, 27.8% (22.9%–33.2%) |
| | `heuristic` gift-removal casts worth less than the gift | 49 of 258 | 0 of 198 |
| | `heuristic` listed removal on an opposing commander | 47 | 2 |
| Run 2, seeds 3001 and 4001 (added) | `heuristic` / `heuristic-baseline` wins | 29 / 19, 26 / 22 | 35 / 13, 28 / 20 |
| Run 2, five seeds pooled | `heuristic` | 135 / 480, 28.1% (24.3%–32.3%) | 143 / 480, 29.8% (25.9%–34.0%) |
| | izzet-aggro under `heuristic` | 28 / 240 | 32 / 240 |

Seed 2001 is the one seed where `heuristic` finished behind the baseline after (21 to 27; it was 24 to 24 before), so two more seeds were run rather than any weight tuned. Pooled over five seeds `heuristic` is 8 games up and the baseline 8 down. Turns p50 is unchanged except seed 1001 (12 to 11). In run 1 every A3 canary meets its bar except Harrow (19 / 32, 59%, to 13 / 29, 45%); Harrow is not removal, and simic-ramp's games diverge once its Beast Within and Pongify casts change. A2 rows meeting their bar went from 6 to 8 of 31 (Delighted Halfling and Ornithopter of Paradise). The suite is 41 of 41 before and after, and no position's pick changed.

### #2693: the mulligan checks for something to cast (2026-10-08)

- **`KeepNeedsCast`, `KeepCastReach` (0).** `decideMulligan` counted lands only, so review game 2 kept two lands and five spells of three to seven mana and missed its next two land drops. A hand at `KeepMinLands` is now kept only if it holds a nonland spell whose mana value is at most its lands plus `KeepCastReach`, with its coloured pips made by those lands (a land that names no colour, such as a fetch land or Exotic Orchard, counts as any). Otherwise it takes the mulligan, but only while the mulligan is free (the engine's first redraws seven); a mulligan that costs a card keeps the old land-count rule. A reach of 1, the issue's "lands + 1", would still keep game 2's hand on Chaos Warp, so the default is 0.
- The arena has a new **Opening hands** section (`botarena/opening.go`): mulligans per keep, kept hands by size, and land drops missed on each seat's own turns 2–4, from the runner's observer.

`BaselineConfig` zeroes both. Before is this branch with `KeepNeedsCast` off, which takes the old code path exactly; every run is `--rotate --lockstep` with the real dump. No run stalled.

| Run | Measure | Before | After |
|---|---|---|---|
| Run 1 (§8, 64 games, seed 1) | mulligans per keep (all four decks) | 62 / 256, 0.24 | 67 / 256, 0.26 |
| | kept 7 / 6 | 246 / 10 | 246 / 10 |
| | land drops missed on own turns 2–4 | 77 / 768, 10.0% | 77 / 768, 10.0% |
| | turns p50 | 14 | 14 |
| | wins: esper / izzet / mono-black / simic | 15 / 5 / 23 / 21 | 15 / 4 / 25 / 20 |
| Run 2, seeds 1, 1001, 2001 pooled | `heuristic` wins | 86 / 288, 29.9% | 84 / 288, 29.2% |
| | `heuristic-baseline` wins | 58 / 288 | 60 / 288 |
| | `heuristic` izzet-aggro: wins, mulligans per keep, missed drops | 8 / 144, 0.25, 39 / 432 | 6 / 144, 0.26, 39 / 432 |
| | `heuristic` simic-ramp: wins, mulligans per keep, missed drops | 78 / 144, 0.23, 49 / 432 | 78 / 144, 0.28, 51 / 432 |
| | turns p50 | 11 | 11 |

Run 2 is `--seats heuristic,heuristic-baseline,heuristic,heuristic-baseline --decks izzet-aggro,izzet-aggro,simic-ramp,simic-ramp --games 48` per seed. The rule fires rarely: 5 extra mulligans in run 1's 256 seat-games and 9 in run 2's 288 `heuristic` seat-games, changing 5 and 7 games. In run 2's seven changed games the seat that took the new mulligan won two it had lost and lost two it had won. Every difference is inside its interval, and `heuristic` stays ahead of the baseline. In run 1 every A3 canary meets its bar before and after, and A2 rows move by at most two games. The suite is 41 of 41 before and after; it has no mulligan position.

### #2689 PR 5: target gifts across the catalog (2026-10-08)

Owner answer 5a: every catalog spell, mode or row that gives its target player a draw, a token, life or a loot declares a target entry (`Purpose.Targets`), so `PriceTargetPurposes` prices it as the seat's strength change instead of as a hit. Burn is not in this batch.

Declared (23 cards): Ancestral Vision, Atlantis Attacks (the Leviathan mode), Blessed Alliance (the life mode), Blood Pact, Bloodgift Demon, Cease // Desist (Cease), Cephalid Coliseum, Compulsive Research, Deep Analysis, Depth Defiler (the loot mode), Echocasting Symposium, Etched Oracle, Flame of Anor (the draw mode), Flumph, Forbidden Orchard, Insatiable Avarice (the drain-and-draw mode), Loran of the Third Path, Oona's Grace, Rise of the Eldrazi (slot 1 only), Scheming Silvertongue's Sign in Blood, Secret Rendezvous, Sublime Epiphany (the draw mode) and Wedding Ring. Flumph, Loran and Secret Rendezvous also declare the controller's own draw, because "you and target opponent each draw" is two gifts. Compulsive Research declares the printed two discards, not the one-land alternative.

Left undeclared, because no printed number says the amount:
- X, or an amount counted at resolution: Blue Sun's Zenith, Damnable Pact, Drown in Dreams, Heliod's Intervention, Inscription of Abundance (the greatest power), Kozilek's Command, Peer into the Abyss (half a library), Stroke of Genius.
- Generous Plunderer: the target opponent is chosen by a reflexive trigger, which has no row to carry a purpose.
- Treacherous Pit-Dweller: what it hands over is a creature.

The real-dump audit's "target gift" list goes from 30 cards to those 10. The gifts to the controller of a removal spell's target (Swords to Plowshares and the rest of #2679's list) cannot be declared here: their clause is a creature, and the guard refuses a player amount on one.

**Curated decks.** Only Loran of the Third Path (esper-control) gains an entry. `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1 --lockstep`, with the binaries built from `develop` at `50d5c34ec` and from this branch: 0 stalls in both, turns p50 13 in both, and wins esper-control 18 / 17, izzet-aggro 3 / 3, mono-black-aristocrats 21 / 21, simic-ramp 22 / 23. Nine games differ, and every one has a Loran in play: Loran's activation is taken 0 to 2 times in eight of them and 6 times in one, against 1 to 4 before, because aiming a symmetric draw at an opponent is no longer priced as an attack. Two more games differ without a Loran in either deck; a control run of the `develop` binary over seeds 1 to 3 differs from its own earlier run on seed 3 in the runner counters, so that is the residual tie-break nondeterminism, not this change. The suite is 41 of 41 before and after.

#### Burn batch 1 (2026-10-08)

The first burn batch of PR 5 declares a damage target entry (`DamageToTarget`, per clause and slot) on the dump audit's undeclared burn list, A through E in alphabetical order. No price code changed: declared burn goes through the existing `DamageByLethality` and `PriceTargetPurposes` code, and undeclared burn keeps today's price.

Declared (34 cards): Aethertorch Renegade (both abilities), Agate Assault (the damage mode), Annihilating Fire, Arc-Slogger, Archangel of Wrath (both triggers), Balduvian Trading Post, Ballista Watcher // Ballista Wielder (both faces), Barbarian Ring, Betrayer's Bargain, Blasting Station, Bonecrusher Giant // Stomp (Stomp), Boros Charm (the damage mode), Bot Bashing Time, Breya, Etherium Shaper (the player-or-planeswalker ability), Brutal Expulsion (the damage mode), Burn the Accursed (the 5 at the creature; the 2 to its controller is not a target clause), Call In a Professional, Carbonize, Cathartic Pyre (the damage mode), Collective Defiance (the creature and opponent modes), Combust, Consulate Turret, Cramped Vents // Access Maze (Cramped Vents' 6; the life gained from the excess is not declared), Cut In (slot 0 only), Dawnsire, Sunstar Dreadnought (100), Demonic Pact (the damage mode), Desert, Dracosaur Auxiliary, Dynavolt Tower, Eiganjo, Seat of the Empire (channel), Electrickery (the single-target cast; overload clears the clause), Elspeth's Smite, Engulfing Flames and Explosive Derailment (the damage mode).

Left undeclared (5):
- Arrow Storm: 4, or 5 with raid, read as it resolves.
- Burst Lightning: 2, or 4 kicked; the kicker is chosen at cast.
- Cinder Strike: 2, or 4 if blight was paid.
- Court of Ire: 2, or 7 if you are the monarch.
- Drakuseth, Maw of Flames: 4 to the first target and 3 to each other target of one clause, which one entry per clause cannot say.

The audit's burn list goes from 165 cards to 131: 34 declared, the 5 above left in A through E, and 126 from F on. Batch 2 starts at Fall of Cair Andros.

**Curated decks.** None of the four curated decks plays a card in this batch, so none gained an entry. `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1 --lockstep`, with the binaries built from `develop` at `a544704cb` and from this branch: 0 stalls in both, turns p50 14 in both, and wins esper-control 19 / 20, izzet-aggro 2 / 2, mono-black-aristocrats 22 / 22, simic-ramp 21 / 20. One game of 64 differs between the two. A second run of the `develop` binary differs from its own first run in 2 games and ends at 20 / 2 / 22 / 20, the branch's numbers, so the difference is the run-to-run tie-break nondeterminism of #2730, not this change. The suite is 41 of 41 before and after.

#### Burn batch 2 (2026-10-08)

The second burn batch of PR 5 declares a damage target entry on the audit's undeclared burn list from Fall of Cair Andros through Loot, the Pathfinder (F to L). No price code changed.

Declared (34 cards): Fall of Cair Andros (the 7-damage ability), Fanged Flames, Feed the Flames, Fervent Paincaster (the player-or-planeswalker ping), Fireblast, Firemaw Kavu (both triggers), Flame Jab, Flame-Blessed Bolt, Flames of the Blood Hand, Foundry Helix, Ghitu Slinger, Glassworks // Shattered Yard (Glassworks), Goblin Bombardment, Goblin Sharpshooter, Goblin Snowman, Grapeshot, Grim Lavamancer, Gut Shot, Idol of the Deep King // Sovereign's Macuahuitl (the Idol's trigger), Impractical Joke, Incendiary Flow, Incinerate, Insult // Injury (Injury, one entry per clause), Jaya Ballard, Task Mage (the 3-damage ability), Keldon Champion, Kolaghan's Command (the damage mode), Lava Coil, Lava Dart, Lesser Masticore, Lightning Axe, Lightning Helix and Lightning Strike (the 3 at the target; the life gained is not declared), Longhorn Sharpshooter, Loot, the Pathfinder (the 3-damage exhaust).

Left undeclared (7):
- Furystoke Giant: the 2 damage is an ability it grants to other creatures, and the audit and the wire's ability rows do not carry a target entry from a grant.
- Garruk Relentless: the creature deals damage back to him, which the fields cannot say.
- Glorybringer: already declared for the exert pricer as `DamageToCreature`; a target entry beside it would price the same hit twice. (Fervent Paincaster's exert ability is the same, and keeps its `DamageToCreature`.)
- Helicarrier Strike: 2, or 4 with teamwork, chosen at cast.
- Invasion of Tarkir: X plus 2, where X is the Dragons revealed.
- Land's Edge: the damage depends on the discarded card being a land.
- Lightning Surge: 4, or 6 with threshold, read as it resolves.

The audit's burn list goes from 131 cards to 97: 34 declared, leaving the 5 from batch 1, the 7 above and 85 from M on. Batch 3 starts at Mage il-Vec.

**Curated decks.** None of the four curated decks plays a card in this batch, so none gained an entry. `boteval arena` with the same command as batch 1, with binaries built from `develop` at `e0139b738` and from this branch: 0 stalls in both, turns p50 14 in both, and wins esper-control 19 / 20, izzet-aggro 2 / 2, mono-black-aristocrats 22 / 22, simic-ramp 21 / 20. A second run of the `develop` binary differs from its own first run in 1 game and ends at 20 / 2 / 22 / 20, the branch's numbers: the difference is #2730's run-to-run tie-break nondeterminism, not this change. The suite is 41 of 41 before and after. The regenerated `legal_actions_agreement.json` gains Goblin Bombardment's declared entry, the one card of the batch in that fixture.

#### Burn batch 3 (2026-10-09)

The third burn batch of PR 5 declares a damage target entry on the audit's undeclared burn list from Mage il-Vec through Rip Apart (M to R). No price code changed.

Declared (41 cards): Mage il-Vec, Magma Spray, Magmatic Galleon (the 5-damage enters trigger), Magmatic Sinkhole, Mardu Charm (the damage mode), Marjhan (the 1 damage to an attacker), Mayhem Devil, Meteor Storm, Mizzium Mortars (the single-target cast; overload clears the clause), Moonrager's Slash, Narset's Rebuke, Naya Charm (the damage mode), Nicol Bolas, God-Pharaoh (the 7-damage minus ability), Nine-Ringed Bo, Niv-Mizzet, Parun, Niv-Mizzet, the Firemind, Obliterating Bolt, Ogre Shaman, Omnath, Locus of Rage (the 3-damage trigger), Orcish Bowmasters (the damage; the amass is not declared), Orcish Hellraiser, Outpost Siege (the Dragons line), Pashalik Mons (the damage trigger), Phlage, Titan of Fire's Fury (the 3 at the target; the life gained is not declared), Pillar of Flame, Pinecone Strike (the damage mode), Pinpoint Avalanche, Pirate Ship, Prophetic Titan (the 4-damage mode), Pumpkin Bombardment, Puncture Blast, Puncturing Blow, Pyromania (both abilities), Rage of Purphoros, Ranger's Firebrand, Reduce to Ashes, Reiterating Bolt, Repulsor Blast (the 5 at the creature; the teamwork 2 to its controller is not a target clause, like Burn the Accursed), Ribbons of Night (the 4; the life and draw are not declared), Rift Bolt and Rip Apart (the damage mode).

Left undeclared (1):
- Razorgrass Ambush // Razorgrass Field: the spell face is a row of the shared modal-land table (`mdfc_lands.go`), which has no per-card purpose field, and this batch adds no fields.

The audit's burn list goes from 97 cards to 56: 41 declared, leaving the 12 from batches 1 and 2, Razorgrass Ambush and 43 from S on. Batch 4 starts at Savage Alliance.

**Curated decks.** None of the four curated decks plays a card in this batch, so none gained an entry. `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1 --lockstep`, with binaries built from `develop` at `a33469d45` and from this branch: 0 stalls in both, turns p50 14 in both, and the two reports are identical game for game (seed, turns, winner): esper-control 24, izzet-aggro 4, mono-black-aristocrats 26, simic-ramp 10 wins in both. The suite is 41 of 41 before and after. `TestLegalActionsAgreement` did not go stale, because none of the batch is in its fixture. `TestRealDumpPurposeAudit` already fails on `develop` for an unrelated reason (Call Forth the Tempest reads as a wipe with no Sweep); the burn list above is read from its output.

#### Burn batch 4 (2026-10-09)

The last burn batch of PR 5 declares a damage target entry on the audit's undeclared burn list from Savage Alliance through Zacama, Primal Calamity (S to Z). No price code changed.

Declared (33 cards): Savage Alliance (the 2 at a creature; its sweep mode keeps its Sweep), Scorching Dragonfire, Scorching Lava (the 2; the kicker rider changes no amount), Scorchmark, Serpentine Spike (2, 3 and 4 on slots 0, 1 and 2), Siege-Gang Commander, Skullcrack, Skysovereign, Consul Flagship, Smite the Deathless, Sorin Markov (the +2), Spikefield Hazard, Spring-Loaded Sawblades, Staggershock, Stormbind, Sudden Shock, Sunscorched Desert, Suplex (the damage mode), Sword of Fire and Ice (the 2; the draw is not declared), Synth Eradicator, Task Mage Assembly, Thornscape Battlemage (the kicked-with-{R} trigger), Thunderous Wrath, Ticking Gnomes, Touch of the Void, Trumpeting Carnosaur (the discard ability), Ugin, the Spirit Dragon (the +2), Walking Ballista, Warrior's Blades (the 3; the life gained is not declared), Wild Slash, Wilt in the Heat, Wrenn and Six (the -1), Yamabushi's Flame and Zacama, Primal Calamity (the 3-damage ability).

Left undeclared (10), which with the 13 from batches 1 to 3 is every undeclared card the audit still lists:
- Slaying Fire: 3, or 4 with adamant, read as it resolves.
- Stormscale Anarch: 2, or 4 if the card discarded at random was multicolored.
- Torch the Tower: 2, or 3 if bargained.
- Urza's Rage: 3, or 10 kicked.
- Unstable Footing: the 5 is dealt to a target chosen only when kicked, and that clause hangs on the kicker, not on the statement; `checkTargetPurposes` refuses slot 0 because the statement has no target clause.
- Three Bowls of Porridge: a modal activated ability. Only a spell's mode bullets carry a purpose the bot reads, and an activated row reads entries for its first mode only.
- Thornbite Staff: the 1 damage is an ability it grants to the equipped creature; the audit and the wire's ability rows do not carry an entry from a grant (like Furystoke Giant).
- Spiked Corridor // Torture Pit, Tibalt, Rakish Instigator and Weapons Manufacturing: the damage is a created token's own trigger (the Devil's 1, the Munitions' 2). The audit reads the card, not the token it makes, and a token's trigger is shared by every card that makes it, so it is not declared here.

The audit's burn list goes from 56 cards to 23: 33 declared. The 23 left are the 12 from batches 1 and 2, Razorgrass Ambush (batch 3) and the 10 above.

**Curated decks.** None of the four curated decks plays a card in this batch, so none gained an entry. `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1 --lockstep`, with binaries built from `develop` at `800eff6cb` and from this branch: 0 stalls in both, turns p50 14 in both, and wins esper-control 23 / 24, izzet-aggro 4 / 4, mono-black-aristocrats 27 / 26, simic-ramp 10 / 10. One game of 64 differs (the 48th, 16 turns against 13). A second run of the branch binary differs from its own first run in that same game and matches `develop` game for game, while two runs of the `develop` binary were identical, so the difference is #2730's run-to-run tie-break nondeterminism, not this change. The suite is 41 of 41 before and after. `TestLegalActionsAgreement` did not go stale. `TestRealDumpPurposeAudit` still fails on `develop` for Call Forth the Tempest (#2771), the only failure; the burn list above is read from its output.

### #2400: the race sends the swing that ends the game soonest (2026-10-12)

The [amendment of 2026-10-12](#amendment-2026-10-12-the-race-sends-the-swing-that-ends-the-game-soonest), accepted. `develop` is `de452ebd6`, after #2904 made lockstep deterministic. Every variant is that tree with the `Config` switches set as named, so `develop` is the row with all four off. The runs are the ones #2400's table used: non-race lockstep runs of `TestFourHeuristicBotsPlayToAWinner` (seeds 101–112) and `TestTwoHeuristicBotsPlayToAWinner` (seeds 201–212), with `AISEAT_HEURISTIC_GAMES=12 AISEAT_H2H_GAMES=12`, and the seed-1409 mirrors in `lockstep_game_test.go`. The full Wurm mirror is skipped on `develop` (#2474), so it was run with the skip lifted. A board is the turns with the edge on seat 0 / seat 1.

**Every variant.** The objective is `RaceEndsSoonest`. The refinements are T (`FreeTrampleOverflow`), X (`FreeCommanderNeedsTax`) and R (`RecastChumpForNext`).

| Objective | Refinements | 4-seat mean | 2-seat mean | Wurm + Ogre | full Wurm |
|---|---|---:|---:|---|---|
| smallest (`develop`) | — | 16.08 | 15.08 | 5 / 6 | 26 / 26, `[6 3]` |
| smallest | T | 17.08 | 15.25 | 9 / 9 | 26 / 26, `[6 3]` |
| smallest | X | 16.92 | 15.33 | 5 / 6 | 26 / 26, `[6 3]` |
| smallest | R | 16.08 | 15.08 | 5 / 8 | 26 / 26, `[6 3]` |
| smallest | T, X | 16.92 | 15.08 | 9 / 9 | 26 / 26, `[6 3]` |
| soonest | — | 15.42 | 14.17 | 5 / 6 | 26 / 26, `[6 3]` |
| soonest | T | 15.92 | 13.67 | 5 / 6 | 26 / 26, `[6 3]` |
| **soonest** | **X (the default)** | **15.25** | **13.75** | 5 / 6 | 26 / 26, `[6 3]` |
| soonest | R | 15.42 | 14.17 | 5 / 6 | 26 / 26, `[6 3]` |
| soonest | T, X | 16.00 | 13.25 | 5 / 6 | 26 / 26, `[6 3]` |
| soonest | X, R | 15.92 | 14.25 | 5 / 6 | 26 / 26, `[6 3]` |
| soonest | T, X, R | 15.92 | 13.75 | 5 / 6 | 26 / 26, `[6 3]` |

In every row, no gate game stalls, and heuristic beats random 12–0. The Drake mirror ends on 8 / 8 and Wurm + Bear on 1 / 2.

- **The old objective reproduces #2400's finding on today's `develop`.** Trample overflow adds a turn to the 4-seat gate, and the tax adds 0.84. Trample also takes Wurm + Ogre from 5 / 6 to 9 / 9.
- **The recast chump no longer stalls a Wurm board.** In #2310's measurement, at `a5285c5ee`, it stalled both at 26 turns. On `de452ebd6`, Wurm + Ogre ends on 5 / 8 with it under the old objective and on 5 / 6 under the new one. The full mirror stops at `[6 3]` with or without it. That is #2474's position: Bear and commander against a lone commander on 3, a kill three turns deep that neither planner sees. In the full mirror, R sends 8 attrition swings where the others send 1, and none of them connects.
- **The full mirror with creatures traded as a key**, the rejected third key, stops at `[6 6]`. The first exchange sent both Wurms, "2 of mine for 2 of theirs". The edge became an Ogre, and the defender's commander traded it away.

**Per seed** (turns / winner seat):

| Seed | `develop` | soonest | soonest + X |
|---|---|---|---|
| 101 | 12 / 1 | 12 / 1 | 12 / 1 |
| 102 | 11 / 3 | 11 / 3 | 11 / 3 |
| 103 | 18 / 1 | **12** / 1 | 12 / 1 |
| 104 | 15 / 0 | **13** / 0 | 13 / **2** |
| 105 | 22 / 3 | **17** / 3 | 17 / 3 |
| 106 | 11 / 2 | **13 / 1** | **16 / 0** |
| 107 | 24 / 3 | 24 / 3 | 24 / 3 |
| 108 | 17 / 2 | **16** / 2 | 16 / 2 |
| 109 | 17 / 1 | **18 / 3** | **13 / 2** |
| 110 | 16 / 3 | **15** / 3 | **14 / 0** |
| 111 | 14 / 2 | 14 / 2 | 14 / 2 |
| 112 | 16 / 0 | **20** / 0 | **21** / 0 |
| **4-seat mean** | **16.08** | **15.42** | **15.25** |
| 201 | 13 / 0 | **12** / 0 | 12 / 0 |
| 202 | 21 / 0 | **13** / 0 | 13 / 0 |
| 203 | 21 / 0 | **16** / 0 | 16 / 0 |
| 204 | 10 / 0 | **17** / 0 | 17 / 0 |
| 205 | 10 / 0 | 10 / 0 | 10 / 0 |
| 206 | 20 / 0 | 20 / 0 | **14 / 1** |
| 207 | 14 / 1 | **13** / 1 | 13 / 1 |
| 208 | 14 / 1 | 14 / 1 | 14 / 1 |
| 209 | 12 / 1 | 12 / 1 | 12 / 1 |
| 210 | 11 / 0 | **10** / 0 | 10 / 0 |
| 211 | 18 / 1 | 18 / **0** | **19 / 1** |
| 212 | 17 / 1 | **15** / 1 | 15 / 1 |
| **2-seat mean** | **15.08** | **14.17** | **13.75** |

Bold marks a change from the column to its left.

**Why the seeds change.** Each game was replayed with its move log, and the first move that differs was read. Past that move, the deals and draws meet a different board, so a later turn count is a different game, not the same decision repeated.

- **`develop` → soonest: 23 of 24 games reach a different move.** Only seed 102 plays the same game. In 22 of the 23, the first different move is the same swing grown bigger:
  - **20 are attrition swings** that add attackers for more sure damage, often in a shorter line. Seed 202: two Ogres (4 sure, a race within 3 exchanges) become two Ogres and the commander (7 sure, within 2), and the game ends on turn 13 instead of 21. Seed 203: a lone Drake (3 sure, within 3) becomes a Drake, a Wurm and two Ogres (11 sure, within 1), and it ends on 16 instead of 21. Seed 103: one Ogre (4 sure) becomes two (8 sure), and it ends on 12 instead of 18.
  - **2 are two-turn races.** Seed 111 goes from `8 now + 13 next` to `11 now + 13 next`. Seed 212 goes from `3 now + 6 next` to `5 now + 2 next`.
  - **The 23rd, seed 101, is a line found where `develop` found none.** Seat 2 attacks with a Bear for 5 sure where it had passed. The search used to give up when its budget ran out. Now it sends the best line found so far.

  Seeds 101, 107, 111, 205, 206, 208 and 209 end on the same turn. Seed 211 ends on the same turn with a different winner.
- **The three games that get longer.**
  - **Seed 204 (10 → 17).** The first swing is a Wurm and an Ogre for 11 sure, instead of the Wurm alone for 7. Two turns later the defender gang-blocks the Wurm with an Ogre and a Drake and kills it. The old line never offered that block, and the game grinds on.
  - **Seed 112 (16 → 20)** sends two Ogres for 8 sure instead of one for 4.
  - **Seed 106 (11 → 13)** sends the commander and a Bear for 5 sure instead of the commander alone for 3, and the winner changes.
- **soonest → soonest + X: 12 games reach a different move.** The tax acts only on a defender's commander, so every first difference is a commander the tax prices out. Its block stops being free, and a swing qualifies, or qualifies differently, that did not before.
  - **Seeds 206, 211 and 111:** an attrition swing where the bot had passed. Seed 206 ends on turn 14 instead of 20, and the other seat wins.
  - **Seeds 105 and 204:** an attrition swing becomes a two-turn race. In seed 204 it is the same three attackers, now `14 now + 4 next turn`.
  - **Seeds 112, 208 and 209:** the race's numbers grow.
  - **Seeds 104 and 110:** a per-creature attack becomes an attrition swing or a race.
  - **Seeds 106 and 109 (4 seats):** under soonest the bot had a line within 1 exchange against one opponent. With the tax, an opponent earlier in its order has a line too, within 3 exchanges, and the bot swings at that one instead. Both planners try defenders in focus order and take the first defender with a line. The objective ranks swings against one defender, not across defenders ([Open](#open-2400)).

**Other gates on the default** (soonest + X):

- `boteval suite run --policy heuristic`: 41 of 41, every tag 100%.
- `go test ./internal/aiseat/heuristic/ ./internal/aiseat/suite/...`: green. `TestBaselineConfigRanksTheSuiteAsBefore` holds, since `BaselineConfig()` turns all four switches off.
- `AISEAT_GAME_TESTS=1 AISEAT_HEURISTIC_GAMES=12 AISEAT_H2H_GAMES=12` over `./internal/aiseat/...` and `./internal/botarena/` with `-race`: green. The gate seeds under `-race` match the non-race runs above game for game. Heuristic vs random is 12–0. Decision latency over 2534 decisions is p50 105µs, p99 4.44ms, max 21.4ms (`-race`, a loaded box), under the 50ms p99 bar. #2310 measured 105µs, 5.25ms and 17.3ms on `develop` on its own box.
- Back-outs: with `RaceEndsSoonest` off, `TestRaceSendsTheSwingThatEndsSoonest` and two cases of `TestRaceCountsACommanderItsControllerCannotRecast` fail. With `FreeCommanderNeedsTax` off, the same two cases fail.

**#2310's tests.** Five of #2310's decision tests put the defender's commander on a board with no lands. With the tax counted, it could not be cast again, so it was not free, and the tests no longer measured what they were written for. Each board now gives the defender the two lands a {1}{R} commander needs (`withRecastMana`). `TestRaceStillBanksDamageOnlyALostCreatureStops` pins the lone commander's race, the smallest swing, so it runs with `RaceEndsSoonest` off.

### #2457: self-replacing discards and waiting for the main phase (2026-10-12)

The [amendment of 2026-10-12](#amendment-2026-10-12-cycling-cards-as-cheaper-discards-and-the-bots-own-draw-step), accepted. `DefaultConfig()` against `BaselineConfig()`: `SelfReplacingDiscards` true (baseline false) and `WaitForMain` true (false). No weight is new: the price reads `Weights.Hand`, `TutorWeight` and `SpellPerMana`.

Before is `develop` at `de452ebd6`, which has #2904's deterministic lockstep. After is this branch with both switches. "Discard price only" is a scratch build of this branch with `WaitForMain` off, to show each lever alone. Every run is `--rotate --lockstep` with the real dump. **0 stalls, 0 rejected moves and 0 runner fallbacks in every run.** Wall-clock times are not comparable: another agent's arena runs shared the machine.

**Run 1** (§8, 64 games, seed 1, with the decision log). The counters come from a scratch pass over the decision logs. A "cycler" is a card in the bot's hand with a `discard_self` hand row.

| Measure | Before | Discard price only | After |
|---|---:|---:|---:|
| games that differ from before (winner or turns) | — | 4 / 64 | 58 / 64 |
| wins: esper / izzet / mono-black / simic | 20 / 2 / 26 / 16 | 19 / 2 / 27 / 16 | 19 / 3 / 27 / 15 |
| turns p50 | 12 | 12 | 13 |
| casts in the bot's own upkeep or draw step, stack empty | 530 | 532 | **0** |
| discard choices (cleanup, loot, prompt) offering a cycler among two or more of the bot's cards: a nonland cycler named | 3 of 35 | 8 of 36 | **12 of 34** |
| cost discards offered with a cycler among the payments: a cycler paid | 0 of 20 | 0 of 14 | 1 of 15 |
| stranded turns (ADR 0136 P2) | 93 / 2,898, 3.2% | 97 / 2,923, 3.3% | 93 / 3,020, 3.1% |
| plan misses (ADR 0136 P6) | 45 / 1,049, 4.3% | 46 / 1,061, 4.3% | 62 / 1,303, 4.8% |
| decision p99 | 406 µs | 393 µs | 425 µs |

The 530 draw-step casts were mostly removal and card draw: Pull from Tomorrow 30, Withering Torment 29, Cyclonic Rift 26, Infernal Grasp 25, Go for the Throat 20, Utter End 20, Return of the Wildspeaker 18, Mortify 17, Murder 16 and Sphinx's Revelation 16 lead the list. They are now cast in the main phase, where the turn plan sees them, so planned windows rose from 1,049 to 1,303 and plan misses from 4.3% to 4.8%, under P6's 5%. Izzet-aggro's misses rose most (6.7% to 10.0%, 19 of 190). The discard price alone moves only 4 games of 64: cyclers are few (Marauding Mako, Scrounging Skyray and Magmakin Artillerist in izzet-aggro, and Raffine's Tower, a land, in esper-control).

Unexpected Windfall, the position's card, is cast less: 13 casts in 13 of 13 games offered before, 6 casts in 6 of 9 after. Before, 11 of the 13 were in the bot's own upkeep or draw step, and 10 of the 13 cards they discarded were lands, one of them a Myriad Landscape, the position's rejected answer. After, the bot reaches its main phase, plays that land as its land drop, and has no spare land left to pitch, so Windfall falls under the bar. None of the 6 casts after is in an upkeep or draw step with an empty stack.

`never` counts are unchanged except izzet-aggro, 3 to 4: Marauding Mako's cycling, taken once in 47 windows before, is taken in none of 66 after. Every A3 canary that met its bar still does, and Rhystic Study on simic-ramp reaches its bar (8 of 22, 36%, to 11 of 22, 50%). A2 rows meeting 80% of games offered: 7 before, 10 after.

**Run 2** (§8: two `heuristic` and two `heuristic-baseline`, `--games 48` with `heuristic` on esper-control and izzet-aggro, then 48 with the seats swapped). Seed 1 is the ADR's; seed 1001 was added to read the first seed's direction.

| Seed | Half | `heuristic` before | after | `heuristic-baseline` before | after |
|---|---|---:|---:|---:|---:|
| 1 | esper and izzet | 27 / 96 | 30 / 96 | 21 / 96 | 18 / 96 |
| 1 | black and simic | 41 / 96 | 36 / 96 | 7 / 96 | 12 / 96 |
| 1001 | esper and izzet | 27 / 96 | 29 / 96 | 21 / 96 | 19 / 96 |
| 1001 | black and simic | 38 / 96 | 36 / 96 | 10 / 96 | 12 / 96 |
| **1, pooled** | | 68 / 192, 35.4% (29.0%–42.4%) | **66 / 192, 34.4% (28.0%–41.3%)** | 28 / 192 | 30 / 192 |
| **both, pooled** | | 133 / 384, 34.6% (30.0%–39.5%) | **131 / 384, 34.1% (29.6%–39.0%)** | 59 / 384 | 61 / 384 |

Not detectably worse: the upper bound is far above 25%, and the lower bound still clears it, as before. By deck, over both seeds (96 seat-games each), `heuristic` won esper-control 52 to 56, izzet-aggro 2 to 3, mono-black-aristocrats 47 to 39 and simic-ramp 32 to 33. Mono-black's drop is about one standard error, mostly seed 1's half B (25 to 18); in run 1, a mirror, it is 26 to 27. Turns p50 move by at most one.

**Suite:** `boteval suite run --policy heuristic`, 41 of 41 before and after, every tag at 100%. The move, layer and reason are the same at every position but `do-not-windfall-away-the-last-land`, which passes for `wait for the main phase: cast spell (declared purpose) Cast Unexpected Windfall (+1.50)` where it passed for `nothing worth doing`. Its frozen view had the `purpose` the server now projects on the Mako's cycling row stamped on (`draws: 1`, `answers: ["value"]`), as the discard-payoff amendment did for its rows; nothing else changed. `BaselineConfig()` ranks every position as recorded in `baseline_rankings.json`.

**Pins.** `TestArenaSeededGameIsTheSameGameAfterTheOpeningRollWindow` was re-pinned by hand (its seventh exception). In seeds 31 and 107 the first difference is a Lightning Bolt the heuristic held through its upkeep and draw step: seed 31 casts it at the same Drake in its main phase, and seed 107 spends that mana on its commander. With `WaitForMain` off the old digests still match, whatever `SelfReplacingDiscards` is.

### #2474: the attrition horizon trades commanders (2026-10-12)

The [amendment of 2026-10-12](#amendment-2026-10-12-the-attrition-horizon-trades-commanders), accepted. `develop` is `2594c7de8`, with #2400's default (`RaceEndsSoonest` and `FreeCommanderNeedsTax` on). Every variant is the branch with the switches set as named, so `develop` is the row with both off. The runs are #2400's: non-race lockstep runs of `TestFourHeuristicBotsPlayToAWinner` (seeds 101–112) and `TestTwoHeuristicBotsPlayToAWinner` (seeds 201–212) with `AISEAT_HEURISTIC_GAMES=12 AISEAT_H2H_GAMES=12`, `TestHeuristicBeatsRandomHeadToHead`, and the seed-1409 mirrors in `lockstep_game_test.go`, the full Wurm mirror included now that its skip is gone. A board is the turns with the edge on seat 0 / seat 1.

F is `HorizonPlaysFreeAnswer`, H is `HorizonRecastsCommander`, and R is #2400's `RecastChumpForNext`. "One pass" is H with the trade allowed in the search's only pass, as first built; "second pass" is H as accepted.

| Switches | 4-seat mean | 2-seat mean | Wurm + Ogre | full Wurm |
|---|---:|---:|---|---|
| none (`develop`) | 15.25 | 13.75 | 5 / 6 | 26 / 26, `[6 3]` |
| F | 15.25 | 13.75 | 5 / 6 | 26 / 26, `[6 3]` |
| H, one pass | 15.92 | 13.75 | 5 / 6 | 26 / 26, `[6 3]` |
| F, H, one pass | 15.92 | 13.75 | 5 / 6 | 7 / 6 |
| H, second pass | 15.25 | 13.75 | 5 / 6 | 26 / 26, `[6 3]` |
| **F, H, second pass (the default)** | **15.25** | **13.75** | 5 / 6 | **7 / 6** |
| F, H, second pass, R | 15.92 | 14.25 | 5 / 6 | 7 / 6 |

In every row, no gate game stalls, and heuristic beats random 12–0. The Drake mirror ends on 8 / 8 and Wurm + Bear on 1 / 2.

- **F alone and H alone (second pass) change nothing.** Each needs the other on the Wurm board, and neither changes a gate game.
- **The default plays all 24 gate games move for move as `develop` does.** Compared with `AISEAT_DEBUG=1` move logs: 1710 to 2451 moves a game at 4 seats and 487 to 965 at 2 seats, all identical. Wherever the second pass ran, it found no line.
- **The full Wurm mirror** ends on turns 7 and 6 with lives `[6 0]` and `[0 6]`. On the stop board the edge seat sends its commander and the Bear: `attack: attrition — 0 of mine (1 cast again) for 0 of theirs and 2 sure now, a race within 2 exchange(s); crack-back 3 < my 6`. The defender blocks the Bear with its commander, takes 3 and dies ([Open](#open-2474)).
- **R** makes seeds 106 (16 → 24) and 211 (19 → 25) longer, as it does without this amendment (#2400's "soonest, X, R" row is 15.92 and 14.25). It stays off.

**Per seed, the trade in one pass** (turns / winner seat; every other seed matches `develop`):

| Seed | `develop` | F, H, one pass |
|---|---|---|
| 101 | 12 / 1 | **16 / 3** |
| 103 | 12 / 1 | **14** / 1 |
| 109 | 13 / 2 | **14 / 3** |
| 111 | 14 / 2 | **15** / 2 |

H alone in one pass changes the same four seeds the same way. The first move that differs, read off the move logs:

- **Seeds 101, 103 and 111 lose a line.** On `develop` the bot sends an attrition swing there: the commander and two Bears (101), an Ogre (103), a Bear (111). With the trade in the one pass it finds no line and falls back to a per-creature attack or a pass. Counting the bot's commander as cast again makes more exchanges acceptable at every depth, so the search spends its budget of 100 swings on them before it reaches the line it used to send. With the budget raised to 1000 as a check, seeds 101 and 103 end as on `develop` (12 / 1).
- **Seed 109** sends a commander-trade line, the Ogre, the commander and a Bear for 6 sure within 2 exchanges, where `develop` sent the Ogre for 4 sure within 3. The game ends a turn later with another winner.

**Rebased onto `032e1f138`.** #3064, the cycling-discard amendment above, merged while this was measured, and it moves `develop`'s gates. The default against that `develop`, with the same runs:

| | 4-seat mean | 2-seat mean | Wurm + Ogre | full Wurm |
|---|---:|---:|---|---|
| `develop` `032e1f138` | 16.17 | 13.25 | 5 / 6 | 26 / 26, `[6 3]` |
| **the default** | **16.17** | **13.25** | 5 / 6 | **7 / 6** |

Every gate game ends on the same turn with the same winner. Heuristic vs random is 12–0, the Drake mirror ends on 8 / 8 and Wurm + Bear on 1 / 2. One gate game plays differently. In seed 105, at move 1796, seat 3 sends its commander and a Bear at seat 1 (`0 of mine (1 cast again) for 0 of theirs and 2 sure now, a race within 2 exchange(s)`) where `develop` passed. Seat 1 chumps the commander with an Ogre instead of trading commanders. The game still ends on turn 15, and seat 3 still wins.

**Other gates on the default:**

- `boteval suite run --policy heuristic`: 41 of 41, every tag 100%, before and after the rebase.
- `go test ./internal/aiseat/heuristic/`: green. `TestBaselineConfigRanksTheSuiteAsBefore` holds, since `BaselineConfig()` turns both switches off.
- `AISEAT_GAME_TESTS=1 AISEAT_HEURISTIC_GAMES=12 AISEAT_H2H_GAMES=12 AISEAT_WALLCLOCK=600s` over `./internal/aiseat/...` and `./internal/botarena/` with `-race`: green. The gate seeds under `-race` match the non-race runs game for game, and heuristic vs random is 12–0. Decision latency over 2534 decisions is p50 96µs, p99 6.00ms and max 35.2ms (`-race`, on a box shared with another agent's test runs), under the 50ms p99 bar. Rebased onto `032e1f138` the run is green again, its gate seeds match too, and latency over 2389 decisions is p50 96µs, p99 7.74ms and max 33.7ms. #2400 measured 105µs, 4.44ms and 21.4ms. The second pass runs only in a decision whose first pass found no line, and wall-clock numbers on a shared box are not a reliable comparison.
- Back-outs: `TestAttritionTradesCommandersForTheKill` turns each switch off in turn and the bot sends nothing, and the full Wurm board stalls at `[6 3]` with either off (the F and H rows). `TestOutcomeOfTheCommanderTrade` reads the outcome under each of the four combinations.

### #3080: the commander cast again after combat (2026-10-14)

A follow-up to the #2474 amendment, behind `HorizonRecastsSameTurn` (on in `DefaultConfig()`, off in `BaselineConfig()`). In the pass that casts the bot's commander again, one that dies in an exchange is cast in the same turn's second main phase when the bot's mana sources still untapped at the attack pay its mana value plus its tax (every source of the bot's on a later turn the line plays forward, which starts with its untap step). It is then ready to attack on the next board (CR 302.6), so the line does not spend a turn waiting for it. When the untapped mana does not pay, the commander is back summoning sick as before. `TestAttritionRecastsInTheSameTurn` pins the issue's board: one exchange instead of two, and two with the lands tapped or the switch off.

Measured on `develop` `8624909a9`, the switch on against the same tree with it off. The gate runs are #2474's (non-race lockstep, `AISEAT_HEURISTIC_GAMES=12 AISEAT_H2H_GAMES=12`), and run 1 is §8's with `--lockstep`.

- **Gates.** 23 of 24 gate games end on the same turn with the same winner and lives. Seed 212 ends on turn 13 instead of 15, same winner (seat 1). Heuristic vs random is 12–0 both ways. The seed-1409 mirrors are unchanged: the Drake mirror 8 / 8, Wurm + Bear 1 / 2, Wurm + Ogre 5 / 7, the full Wurm mirror 8 / 7.
- **Run 1** (`--games 64 --rotate --lockstep --seed 1`, esper-control, mono-red-aggro, mono-black-aristocrats, simic-ramp): win counts identical both ways (20, 24, 10, 10 of 64), 0 stalls, turns p50 12. Two games differ, each two turns shorter with the switch on and the same winner (seeds 2 and 38).

### #2480: Izzet and Simic after the rebalance (2026-10-12)

The [amendment of 2026-10-12](#amendment-2026-10-12-izzet-and-simic-after-the-rebalance), accepted (owner decisions 2026-10-12). `develop` is `0e81b3af4`, with #3062's deterministic lockstep and #3063, #3064 and #3065's heuristic. Step 1 is the impulse price: `PriceImpulse` and `OneShotManaAsToken`, both true in `DefaultConfig()` and false in `BaselineConfig()`; no weight is new. Step 2 adds the Izzet swap. Every run is `--rotate --lockstep` with the real dump. Run 1 is §8's (four `heuristic` seats, 64 games); run 2 is §8's, `--games 48` with `heuristic` on esper-control and izzet-aggro (half A), then on mono-black-aristocrats and simic-ramp (half B). **0 stalls, 0 rejected moves and 0 runner fallbacks in every run.** The raw decision logs were deleted once the tables were extracted.

**Run 1, seed 1** (wins of 64, Wilson 95%):

| Deck | `develop` | Step 1 | Step 2 |
|---|---|---|---|
| esper-control | 17, 26.6% (17.3%–38.5%) | 17, 26.6% (17.3%–38.5%) | 20, 31.2% (21.2%–43.4%) |
| izzet-aggro | 1, 1.6% (0.3%–8.3%) | 1, 1.6% (0.3%–8.3%) | 3, 4.7% (1.6%–12.9%) |
| mono-black-aristocrats | 26, 40.6% (29.5%–52.9%) | 24, 37.5% (26.7%–49.7%) | 25, 39.1% (28.1%–51.3%) |
| simic-ramp | 20, 31.2% (21.2%–43.4%) | 22, 34.4% (23.9%–46.6%) | 16, 25.0% (16.0%–36.8%) |
| games that differ from the column to the left | — | 5 (2 winners) | 55 (36 winners) |
| turns p50 | 13 | 13 | 13 |

`never` lists. `develop`: izzet-aggro, Glint-Horn Buccaneer's, Marauding Mako's, Professional Face-Breaker's and Scrounging Skyray's activations; mono-black-aristocrats, Ashnod's Altar, Blood Artist and Burnished Hart's activation; esper-control and simic-ramp, none. Step 1: Face-Breaker's activation leaves the list (20 of 935 windows, 7 of 13 games). Step 2: the same as step 1, plus Burnished Hart's activation on simic-ramp (0 of 309 windows; it was 1 of 346 on `develop`). The A3 canaries that met their bar on `develop` meet it at both steps; Rhystic Study, below its bar on both decks on `develop` (11 of 23 and 10 of 24), meets it at step 2 (15 of 23 and 9 of 16).

**Run 1, seed 1001** (a second block, steps 1 and 2 only):

| Deck | Step 1 | Step 2 | Pooled with seed 1, step 1 | Pooled with seed 1, step 2 |
|---|---|---|---|---|
| esper-control | 30, 46.9% (35.2%–58.9%) | 26, 40.6% (29.5%–52.9%) | 47 / 128, 36.7% (28.9%–45.3%) | 46 / 128, 35.9% (28.1%–44.5%) |
| izzet-aggro | 1, 1.6% (0.3%–8.3%) | 2, 3.1% (0.9%–10.7%) | 2 / 128, 1.6% (0.4%–5.5%) | 5 / 128, 3.9% (1.7%–8.8%) |
| mono-black-aristocrats | 18, 28.1% (18.6%–40.1%) | 25, 39.1% (28.1%–51.3%) | 42 / 128, 32.8% (25.3%–41.3%) | 50 / 128, 39.1% (31.0%–47.7%) |
| simic-ramp | 15, 23.4% (14.7%–35.1%) | 11, 17.2% (9.9%–28.2%) | 37 / 128, 28.9% (21.8%–37.3%) | 27 / 128, 21.1% (14.9%–29.0%) |

`never` lists at seed 1001. Step 1: izzet-aggro, Glint-Horn Buccaneer's, Scrounging Skyray's and Solphim's activations (Face-Breaker's is not on it); mono-black-aristocrats, Ashnod's Altar and Blood Artist. Step 2: izzet-aggro, Glint-Horn Buccaneer's, Marauding Mako's, Scrounging Skyray's and Reflection of Kiki-Jiki's activations; mono-black-aristocrats, Ashnod's Altar, Blood Artist and Burnished Hart's activation. Simic's fall at step 2 (37 to 27 of 128) is in games the Izzet swap re-dealt, and its interval still holds 25%.

**Run 2, seed 1** (wins of 48 per deck):

| | `develop` | Step 1 | Step 2 |
|---|---|---|---|
| half A, `heuristic` esper / izzet | 29 / 1 | 28 / 0 | 17 / 4 |
| half A, `heuristic-baseline` black / simic | 14 / 4 | 13 / 7 | 15 / 12 |
| half B, `heuristic` black / simic | 23 / 13 | 23 / 13 | 20 / 15 |
| half B, `heuristic-baseline` esper / izzet | 10 / 2 | 10 / 2 | 12 / 1 |
| **`heuristic`, pooled** | **66 / 192, 34.4% (28.0%–41.3%)** | **64 / 192, 33.3% (27.0%–40.3%)** | **56 / 192, 29.2% (23.2%–36.0%)** |
| `heuristic-baseline`, pooled | 30 / 192 | 32 / 192 | 40 / 192 |
| games that differ, halves A / B | — | 4 / 0 | 43 / 36 |
| turns p50, halves A / B | 15 / 12 | 15 / 12 | 14 / 12 |

Not detectably worse at either step: the lower bound clears 25% at `develop` and step 1 and is 23.2% at step 2, with the upper bound far above 25%. Step 2's fall is mostly half A's Esper (28 to 17 wins), in games the swap re-dealt: it changes the Izzet library, so 43 of half A's 48 games are different games. Half B of step 1 plays every game as `develop` does, because no `heuristic` seat holds Face-Breaker there.

**Izzet, every seat-game** (run 1 at seeds 1 and 1001, and both halves of run 2): 4 of 224 at step 1 (1.8%, 0.7%–4.5%) and 10 of 224 at step 2 (4.5%, 2.4%–8.0%).

**Swap sets scouted with run 1, seed 1** (on step 1; wins of 64):

| Izzet in, out | esper | izzet | mono-black | simic | The new cards |
|---|---:|---:|---:|---:|---|
| The Indomitable and Monument to Endurance, for Wash Away and Negate (the issue's pair) | 20 | 4 | 23 | 17 | The Indomitable cast in 6 of 12 games offered, crewed in 2; Monument cast in 6 of 11 |
| **Fable of the Mirror-Breaker and Monument, for the same (step 2)** | 20 | 3 | 25 | 16 | Fable cast in 9 of 12, its Reflection's copy used in 2; Monument cast in 6 of 11 |
| Fable, Monument, Glorybringer and Third Path Iconoclast, for Wash Away, Negate, Swan Song and Shock | 16 | 2 | 31 | 15 | Glorybringer cast in 21 of 21, exerted in 4; Iconoclast cast in 12 of 13 |

**Why the cards went unused** (the counts in the [amendment](#why-the-cards-go-unused)) come from run 1's decision log on `develop`, read by a scratch pass: each window in which a card's activation was offered, the price the heuristic put on it, and the move it made instead. The landfall count reads the bot's own battlefield in the windows of its own main phase with an empty stack. "Out first or second" counts a seat whose last decision came on an earlier turn than at least two other seats' last decisions. The damage by source is read from the game-log tails the logged views carry, deduplicated, which cover about half of each game's log, so only the shares mean anything.

**Other gates:**

- `boteval suite run --policy heuristic`: 41 of 41 at step 2, every tag at 100%.
- `go test ./...`: green, including `TestImpulseIsTakenInTheOwnMainPhase`, `TestImpulseIsNotTakenWhenTheCardWouldBeLost`, `TestImpulseNeedsBothSwitches` and `TestBaselineConfigRanksTheSuiteAsBefore`. No pinned arena digest moved (`opening_roll_pin_test.go`), so none was re-pinned.
- With the dump, `go test ./internal/decks/`: green. `TestRealDumpDecksAreLegalCommanderDecks` now matches a double-faced card by its front face, as `Load`'s oracle-ID check does.

### #3066: blocks that survive as a set (2026-10-12)

The [amendment of 2026-10-12](#amendment-2026-10-12-blocks-that-survive-as-a-set), accepted. `develop` is `2b6f13831`, with #3065's commander-trade horizon and #2480's impulse price. The branch is that tree with `BlockSetSurvives` on, so `develop` is the row with it off. The runs are #2400's and #2474's: non-race lockstep runs of `TestFourHeuristicBotsPlayToAWinner` (seeds 101–112) and `TestTwoHeuristicBotsPlayToAWinner` (seeds 201–212) with `AISEAT_HEURISTIC_GAMES=12 AISEAT_H2H_GAMES=12`, `TestHeuristicBeatsRandomHeadToHead`, and the seed-1409 mirrors in `lockstep_game_test.go`, each with `AISEAT_DEBUG=1` so the move logs can be compared. Run 1 is §8's (four `heuristic` seats, 64 games, seed 1, `--rotate --lockstep`, the real dump). A board is the turns with the edge on seat 0 / seat 1.

| | 4-seat mean | 2-seat mean | Drake | Wurm + Bear | Wurm + Ogre | full Wurm |
|---|---:|---:|---|---|---|---|
| `develop` | 16.17 | 13.25 | 8 / 8 | 1 / 2 | 5 / 6 | 7 / 6 |
| **`BlockSetSurvives`** | **16.08** | **13.33** | 8 / 8 | 1 / 2 | 5 / **7** | **8 / 7** |

No gate game stalls, and heuristic beats random 12–0, on both.

**Per game that changes** (turns / winner seat). Every other game plays move for move as on `develop`: the move logs differ only in these five games, and in each the first different move is a block the check changed.

| Game | `develop` | `BlockSetSurvives` | The first different move |
|---|---|---|---|
| seed 104 | 13 / 2 | **12** / 2 | A defender with its commander as its only blocker faces a Wurm, an Ogre and a Bear. It chumps the Ogre (+3.65) where it blocked the Bear (+6.90), and lives through the combat that killed it on `develop`. |
| seed 202 | 13 / 0 | **14** / 0 | The defender's fourth block: its commander chumps the second Ogre (+3.65) where it joined a Bear on the first to kill it (+4.35), which let lethal through. |
| Wurm + Ogre, edge seat 1 | 6 / 1 | **7** / 1 | The defender, on 4 with its commander, faces an Ogre and a commander. It chumps the Ogre and takes 3 (+3.65) where it traded commanders and took 4 (+6.00). |
| full Wurm, edge seat 0 | 7 / 0 | **8** / 0 | The issue's board. The defender, on 3, trades commanders and takes 2 (+6.00) where it blocked the Bear and took 3 (+6.90). |
| full Wurm, edge seat 1 | 6 / 1 | **7** / 1 | The same board with the seats swapped. |

**The full Wurm mirror, move by move** (edge on seat 0; seat 1 is the mirror image). Round 7: seat 0 sends the commander and the Bear (`attack: attrition — 0 of mine (1 cast again) for 0 of theirs and 2 sure now, a race within 2 exchange(s); crack-back 3 < my 6`). Seat 1 blocks commander on commander (`block: chump to survive, a set that survives (+6.00)`), takes 2 from the Bear and is on 1. Both commanders go to the command zone, and seat 0 casts its own again in its second main phase. Seat 1 casts its own in its turn. Round 8: seat 0's commander is not summoning sick, so seat 0 sends the Bear and the commander at once (`attack: all-in for the kill`). Seat 1 has one blocker and two attackers, no block survives, and it dies on −2. The horizon's planned line had a waiting turn between the two swings, which the game does not need ([the amendment](#amendment-2026-10-12-blocks-that-survive-as-a-set)).

**Run 1, seed 1** (wins of 64): esper-control 18, izzet-aggro 5, mono-black-aristocrats 25, simic-ramp 16, on `develop` and on the branch. All 64 games end on the same turn with the same winner, the same lives and the same number of decisions for every seat. 0 stalls, turns p50 13, and the `never` lists are the same.

**Other gates on the branch:**

- `boteval suite run --policy heuristic`: 41 of 41, every tag at 100%, as on `develop`.
- `go test ./internal/aiseat/heuristic/ ./internal/aiseat/suite/...`: green. `TestBaselineConfigRanksTheSuiteAsBefore` holds, since `BaselineConfig()` turns the switch off. The new decision tests are in `heuristic/blockset_decide_test.go`: the issue's board (and, with the switch off, the old +6.90 Bear block), commander damage on 19 and on 18, the tenth poison counter from toxic, trample overflow, a set that needs two blockers, and a board where no set survives, which decides as with the switch off.
- `AISEAT_GAME_TESTS=1 AISEAT_HEURISTIC_GAMES=12 AISEAT_H2H_GAMES=12 AISEAT_WALLCLOCK=600s` over `./internal/aiseat/` with `-race`: green, and its gate seeds and mirrors match the non-race runs above game for game. `./internal/aiseat/...` and `./internal/botarena/` are green too. Decision latency over 2294 decisions is p50 112µs, p99 9.75ms and max 51.8ms (`-race`, with the other `-race` packages running beside it), under the 50ms p99 bar.
- **Latency without `-race`,** three runs of `TestHeuristicDecisionLatency` each, on the same box one after the other: `develop` p50 13.2–13.4µs, p99 1.10–1.14ms, max 5.0–5.5ms over 2389 decisions; the branch p50 13.8–14.2µs, p99 1.14–1.21ms, max 7.2–7.8ms over 2294 decisions. Its game (seed 4242) is one of the games a block changes, so the two are not the same decisions. The check runs only when finishing the declaration kills, and stops at the first set that survives.

### #3072: landfall payoffs (2026-10-12)

The [amendment of 2026-10-12](#amendment-2026-10-12-landfall-payoffs), accepted (owner decisions 2026-10-13). `develop` is `489b289aa`, with #3066's block check and #2480's impulse price. The branch is that tree with `PriceLandfallPayoffs` true in `DefaultConfig()` and false in `BaselineConfig()`, and the declarations; no weight is new. Every run is `--rotate --lockstep` with the real dump. Run 1 is §8's (four `heuristic` seats, 64 games); run 2 is §8's, `--games 48` with `heuristic` on esper-control and izzet-aggro (half A), then on mono-black-aristocrats and simic-ramp (half B). Each at seeds 1 and 1001. **0 stalls, 0 rejected moves and 0 runner fallbacks in every run.** The raw decision logs were deleted once the tables were extracted.

**Run 1** (wins of 64, Wilson 95%):

| Deck | Seed 1, `develop` | Seed 1, branch | Seed 1001, `develop` | Seed 1001, branch | Pooled, `develop` | Pooled, branch |
|---|---|---|---|---|---|---|
| esper-control | 18, 28.1% (18.6%–40.1%) | 18, 28.1% (18.6%–40.1%) | 27, 42.2% (30.9%–54.4%) | 24, 37.5% (26.7%–49.7%) | 45 / 128, 35.2% (27.4%–43.8%) | 42 / 128, 32.8% (25.3%–41.3%) |
| izzet-aggro | 5, 7.8% (3.4%–17.0%) | 4, 6.2% (2.5%–15.0%) | 1, 1.6% (0.3%–8.3%) | 1, 1.6% (0.3%–8.3%) | 6 / 128, 4.7% (2.2%–9.8%) | 5 / 128, 3.9% (1.7%–8.8%) |
| mono-black-aristocrats | 25, 39.1% (28.1%–51.3%) | 22, 34.4% (23.9%–46.6%) | 24, 37.5% (26.7%–49.7%) | 24, 37.5% (26.7%–49.7%) | 49 / 128, 38.3% (30.3%–46.9%) | 46 / 128, 35.9% (28.1%–44.5%) |
| simic-ramp | 16, 25.0% (16.0%–36.8%) | 20, 31.2% (21.2%–43.4%) | 12, 18.8% (11.1%–30.0%) | 15, 23.4% (14.7%–35.1%) | 28 / 128, 21.9% (15.6%–29.8%) | **35 / 128, 27.3% (20.4%–35.6%)** |
| games that differ from `develop` | — | 44 (20 winners) | — | 45 (15 winners) | | |
| turns p50 | 13 | 13 | 14 | 14 | | |

`never` lists are the same on both, apart from Izzet's at seed 1001, which Glint-Horn Buccaneer's and Reflection of Kiki-Jiki's activations leave on the branch (in games the change re-dealt). Burnished Hart's activation stays `never` on mono-black-aristocrats, which has no landfall payoff.

**Run 2** (wins of 48 per deck):

| | Seed 1, `develop` | Seed 1, branch | Seed 1001, `develop` | Seed 1001, branch |
|---|---|---|---|---|
| half A, `heuristic` esper / izzet | 17 / 5 | 17 / 5 | 27 / 3 | 27 / 3 |
| half A, `heuristic-baseline` black / simic | 13 / 13 | 13 / 13 | 7 / 11 | 7 / 11 |
| half B, `heuristic` black / simic | 25 / 20 | 20 / 20 | 24 / 13 | 24 / 15 |
| half B, `heuristic-baseline` esper / izzet | 2 / 1 | 6 / 2 | 10 / 1 | 6 / 3 |
| **`heuristic`** | **67 / 192, 34.9% (28.5%–41.9%)** | **62 / 192, 32.3% (26.1%–39.2%)** | **67 / 192, 34.9% (28.5%–41.9%)** | **69 / 192, 35.9% (29.5%–42.9%)** |
| games that differ, halves A / B | — | 0 / 34 | — | 0 / 38 |

Pooled over both seeds, `heuristic` is 134 / 384, 34.9% (30.3%–39.8%) on `develop` and 131 / 384, 34.1% (29.6%–39.0%) on the branch; `heuristic-baseline` 58 and 61 of 384. Half A plays every game as `develop` does, because no `heuristic` seat there holds a landfall payoff and `heuristic-baseline` has the switch off. Turns p50 is unchanged in every half.

**Simic, every `heuristic` seat-game** (run 1 at both seeds and run 2's half B): 61 / 224, 27.2% (21.8%–33.4%) on `develop`; 70 / 224, 31.2% (25.5%–37.6%) on the branch.

**Simic's land-makers** (`heuristic` seats; games used of games offered, and windows taken of windows offered):

| Card | Run 1, `develop` | Run 1, branch | Run 2 half B, `develop` | Run 2 half B, branch |
|---|---|---|---|---|
| Burnished Hart, activation | 5 of 27 games (5 of 1,047 windows) | **21 of 26 (21 of 225)** | 2 of 19 (2 of 786) | **12 of 17 (12 of 140)** |
| Sakura-Tribe Elder, activation | 17 of 42 | 33 of 41 | 13 of 31 | 23 of 26 |
| Explosive Vegetation | 28 of 55 | 50 of 67 | 20 of 37 | 29 of 37 |
| Harrow | 47 of 56 | 56 of 61 | 37 of 41 | 39 of 42 |
| Cultivate | 48 of 61 | 58 of 65 | 30 of 45 | 35 of 43 |
| Kodama's Reach | 34 of 47 | 49 of 58 | 36 of 46 | 39 of 45 |

By seed, in run 1: the Hart's activation went from 3 of 13 games (3 of 728 windows) to 9 of 11 (9 of 64) at seed 1, and from 2 of 14 (2 of 319) to 12 of 15 (12 of 161) at seed 1001. From run 1's decision logs at seed 1, Simic controlled a landfall payoff in 540 of the 559 Hart windows the logs carry on `develop` (2 taken); on the branch, in 36 of 49 (8 taken), and the activation's price in those windows had a median of +1.04 against −0.80 on `develop`.

**Other gates on the branch:**

- `boteval suite run --policy heuristic`: 41 of 41, every tag at 100%, as on `develop`.
- `go test -race ./...`: green, including `TestHartIsActivatedWithTatyovaOut`, `TestHartIsKeptWhenTheLandsPayLittle`, `TestHartWithTatyovaNeedsTheSwitch`, `TestARampSpellIsWorthMoreWithTatyovaOut`, `TestEveryLandfallAmountIsPriced`, `TestLandfallPayoffDeclarations`, `TestLandfallPayoffGuard`, `TestLandfallPayoffOnTheWire` and `TestBaselineConfigRanksTheSuiteAsBefore`. No pinned arena digest moved (`opening_roll_pin_test.go`).
- `AISEAT_GAME_TESTS=1` over `./internal/aiseat/` with `-race` and `-timeout=60m`: every game test passes but `TestFourRandomBotsPlayToAWinner`, whose three games (seeds 31000–31002) hit the 300 s wall clock at turns 19–20. It fails the same way on `develop` `489b289aa`, run alone on a quiet machine, and it seats the random policy, which reads no purpose. It is [#3147](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3147).

### #3071: izzet-aggro rebuilt around Purphoros (2026-10-13)

The [amendment of 2026-10-13](#amendment-2026-10-13-izzet-aggro-rebuilt-around-purphoros), accepted (owner decisions 2026-10-13). The runs below name the deck `izzet-aggro`: they were taken before the owner had the ID renamed to `mono-red-aggro`. `develop` is `f1bf51ff6`, with #3072's landfall price; the branch is that tree with the rebuilt list, and no price, weight or switch changes. Every run is `--rotate --lockstep` with the real dump. Run 1 is §8's (four `heuristic` seats, 64 games); run 2 is §8's, `--games 48` with `heuristic` on esper-control and izzet-aggro (half A), then on mono-black-aristocrats and simic-ramp (half B). Each at seeds 1, 1001, 2001 and 3001. **0 stalls, 0 rejected moves and 0 runner fallbacks in all 640 games on each side.** The raw run directories were deleted once the tables were extracted.

**Run 1** (wins of 64 per seed block, Wilson 95%):

| Deck | Seed 1, `develop` | Seed 1, rebuild | Seed 1001, `develop` | Seed 1001, rebuild | Seed 2001, `develop` | Seed 2001, rebuild | Seed 3001, `develop` | Seed 3001, rebuild | Pooled, `develop` | Pooled, rebuild |
|---|---|---|---|---|---|---|---|---|---|---|
| esper-control | 19, 29.7% (19.9%–41.8%) | 18, 28.1% (18.6%–40.1%) | 24, 37.5% (26.7%–49.7%) | 20, 31.2% (21.2%–43.4%) | 30, 46.9% (35.2%–58.9%) | 16, 25.0% (16.0%–36.8%) | 25, 39.1% (28.1%–51.3%) | 12, 18.8% (11.1%–30.0%) | 98 / 256, 38.3% (32.5%–44.4%) | 66 / 256, 25.8% (20.8%–31.5%) |
| izzet-aggro | 4, 6.2% (2.5%–15.0%) | 19, 29.7% (19.9%–41.8%) | 1, 1.6% (0.3%–8.3%) | 19, 29.7% (19.9%–41.8%) | 2, 3.1% (0.9%–10.7%) | 20, 31.2% (21.2%–43.4%) | 2, 3.1% (0.9%–10.7%) | 19, 29.7% (19.9%–41.8%) | 9 / 256, 3.5% (1.9%–6.5%) | 77 / 256, 30.1% (24.8%–36.0%) |
| mono-black-aristocrats | 21, 32.8% (22.6%–45.0%) | 13, 20.3% (12.3%–31.7%) | 24, 37.5% (26.7%–49.7%) | 17, 26.6% (17.3%–38.5%) | 17, 26.6% (17.3%–38.5%) | 16, 25.0% (16.0%–36.8%) | 19, 29.7% (19.9%–41.8%) | 23, 35.9% (25.3%–48.2%) | 81 / 256, 31.6% (26.3%–37.6%) | 69 / 256, 27.0% (21.9%–32.7%) |
| simic-ramp | 20, 31.2% (21.2%–43.4%) | 14, 21.9% (13.5%–33.4%) | 15, 23.4% (14.7%–35.1%) | 8, 12.5% (6.5%–22.8%) | 15, 23.4% (14.7%–35.1%) | 12, 18.8% (11.1%–30.0%) | 18, 28.1% (18.6%–40.1%) | 10, 15.6% (8.7%–26.4%) | 68 / 256, 26.6% (21.5%–32.3%) | 44 / 256, 17.2% (13.1%–22.3%) |
| turns p50 | 13 | 12 | 13 | 12 | 13 | 12 | 13 | 12 | | |
| not ended / rejected / fallbacks | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 / 0 | 0 / 0 / 0 | | |

`never` lists for izzet-aggro. `develop`: Glint-Horn Buccaneer's, Marauding Mako's and Scrounging Skyray's activations at seed 1; Mako's and Skyray's at 1001 and 3001; Glint-Horn's, Reflection of Kiki-Jiki's and Skyray's at 2001. Rebuild: none at seed 1; Pashalik Mons's sacrifice at 1001; Fearless Liberator's boast at 2001; both at 3001. The other decks' lists are the same on both sides, apart from rows that come and go with the games dealt (Warren Soultrader's and Hedron Archive's activations, Burnished Hart's on mono-black).

**Run 2** (wins of 48 per deck per half):

| | Seed 1, `develop` | Seed 1, rebuild | Seed 1001, `develop` | Seed 1001, rebuild | Seed 2001, `develop` | Seed 2001, rebuild | Seed 3001, `develop` | Seed 3001, rebuild |
|---|---|---|---|---|---|---|---|---|
| half A, `heuristic` esper / izzet | 17 / 5 | 21 / 17 | 27 / 3 | 24 / 16 | 30 / 3 | 19 / 22 | 31 / 1 | 19 / 20 |
| half A, `heuristic-baseline` black / simic | 13 / 13 | 5 / 5 | 7 / 11 | 3 / 5 | 11 / 4 | 5 / 2 | 8 / 8 | 4 / 5 |
| half B, `heuristic` black / simic | 20 / 20 | 19 / 13 | 24 / 15 | 17 / 17 | 24 / 16 | 19 / 20 | 20 / 18 | 11 / 15 |
| half B, `heuristic-baseline` esper / izzet | 6 / 2 | 5 / 11 | 6 / 3 | 6 / 8 | 6 / 2 | 4 / 5 | 9 / 1 | 6 / 16 |
| **`heuristic`** | 62 / 192, 32.3% (26.1%–39.2%) | 70 / 192, 36.5% (30.0%–43.5%) | 69 / 192, 35.9% (29.5%–42.9%) | 74 / 192, 38.5% (31.9%–45.6%) | 73 / 192, 38.0% (31.5%–45.1%) | 80 / 192, 41.7% (34.9%–48.7%) | 70 / 192, 36.5% (30.0%–43.5%) | 65 / 192, 33.9% (27.5%–40.8%) |
| **`heuristic-baseline`** | 34 / 192, 17.7% (13.0%–23.7%) | 26 / 192, 13.5% (9.4%–19.1%) | 27 / 192, 14.1% (9.8%–19.7%) | 22 / 192, 11.5% (7.7%–16.7%) | 23 / 192, 12.0% (8.1%–17.3%) | 16 / 192, 8.3% (5.2%–13.1%) | 26 / 192, 13.5% (9.4%–19.1%) | 31 / 192, 16.1% (11.6%–22.0%) |

Pooled over the four seed blocks, `heuristic` is 274 / 768, 35.7% (32.4%–39.1%) on `develop` and 289 / 768, 37.6% (34.3%–41.1%) on the branch; `heuristic-baseline` 110 and 95 of 768. Every game in every run is dealt differently on the branch, since the Izzet library changed. Turns p50 in run 2 falls by one to three turns in half A and up to two in half B.

**Every `heuristic` seat-game per deck** (run 1 and the `heuristic` half of run 2, four seed blocks):

| Deck | `develop` | Rebuild |
|---|---|---|
| esper-control | 203 / 448, 45.3% (40.8%–49.9%) | 149 / 448, 33.3% (29.1%–37.7%) |
| izzet-aggro | 21 / 448, 4.7% (3.1%–7.1%) | 152 / 448, 33.9% (29.7%–38.4%) |
| mono-black-aristocrats | 169 / 448, 37.7% (33.4%–42.3%) | 135 / 448, 30.1% (26.1%–34.5%) |
| simic-ramp | 137 / 448, 30.6% (26.5%–35.0%) | 109 / 448, 24.3% (20.6%–28.5%) |

**Izzet's cards in run 1** (games used of games offered, four seed blocks): Purphoros cast in 253 of 253; Siege-Gang Commander 47 of 54, Krenko, Mob Boss 63 of 69, Impact Tremors 44 of 48, Torbran 42 of 45, Fanatic of Mogis 46 of 47, Goldspan Dragon 36 of 36. The least used: Pashalik Mons's sacrifice 2 of 34, Fearless Liberator's boast 1 of 5, Krenko, Mob Boss's activation 8 of 30, Hordeling Outburst 9 of 37, Seize the Spoils 14 of 38.

**The scouted variants** in the amendment's table were each built as their own binary and run with run 1 only, at the seeds named. A list in another card order is another set of games, which is why the scouting order of the proposed cards is its own row.

**Other gates on the branch:**

- `boteval suite run --policy heuristic`: 41 of 41, every tag at 100%, as on `develop`. The positions are fixed boards, so a deck list does not change them.
- `go test -race ./...`: green. No pinned arena digest moved (`opening_roll_pin_test.go` seats the synthetic battle deck), so none was re-pinned.
- With the dump, `go test ./internal/decks/ ./internal/botarena/ ./internal/lobby/ ./internal/tablesetups/ ./cmd/boteval/`: green for izzet-aggro, including `TestLockstepRotatedSeedsReplay` (seeds 13 and 52 replay move for move), `TestRealDumpPurposeCuratedSpellList` and `TestRealDumpPurposeAudit`. `TestRealDumpDecksAreLegalCommanderDecks` and `TestTriageDecks` fail for mono-green-stompy, mono-white-weenie and boros-voltron, which this PR does not touch: the local dump (2026-09-24) calls a card in each not legal in Commander.
- **The rename does not change a game.** Run 1 at seed 1, built once with the deck's ID put back to `izzet-aggro` and once with `mono-red-aggro`, plays all 64 games the same (turns, every seat's life and decision count), and both give esper 18, the aggro deck 19, mono-black 13 and Simic 14 wins, as in the table above. The ID is not part of the deal; the card order is.
- **Migration 0018:** `TestMigration0018RenamesTheDeckInLastDeckRecords` and `TestMigration0018RenamesTheDeckInTableSetups` (a bot named after the old ID keeps its name; other bots, the setup's settings, tablemates, `game_id` and `updated_at` are unchanged). `TestLegacyIDsResolveToTheirReplacement` holds the alias, and the client's `lastDeck.test.ts` a guest's stored old ID.

### #3165: token makers priced by what the tokens trigger (2026-10-14)

The [amendment of 2026-10-14](#amendment-2026-10-14-token-makers-priced-by-what-the-tokens-trigger). The base is `develop` at `08d675cc1`. The branch is that tree with `PriceTokenPayoffs` true in `DefaultConfig()`, and the declarations; the comparison is one tree built twice, with the switch off and on, so the declarations alone change nothing. Every run is `--rotate --lockstep` with the real dump. Run 1 is §8's (four `heuristic` seats, 64 games) at seeds 1, 1001, 2001 and 3001; run 2 is §8's, `--games 48` with the decks split one way (half A: `heuristic` on esper-control and mono-red-aggro) and the other (half B), at seeds 1 and 1001. **0 stalls in every run.** The raw decision logs were deleted once the tables were extracted; the switch-off runs wrote none.

**Run 1** (wins of 64 per seed block):

| Deck | Seed 1, off | Seed 1, on | Seed 1001, off | Seed 1001, on | Seed 2001, off | Seed 2001, on | Seed 3001, off | Seed 3001, on | Pooled, off | Pooled, on |
|---|---|---|---|---|---|---|---|---|---|---|
| esper-control | 19 | 21 | 22 | 19 | 16 | 13 | 14 | 14 | 71 / 256, 27.7% (22.6%-33.5%) | 67 / 256, 26.2% (21.2%-31.9%) |
| mono-black-aristocrats | 13 | 10 | 17 | 15 | 12 | 9 | 21 | 19 | 63 / 256, 24.6% (19.7%-30.2%) | 53 / 256, 20.7% (16.2%-26.1%) |
| mono-red-aggro | 16 | 18 | 15 | 20 | 22 | 26 | 20 | 23 | 73 / 256, 28.5% (23.3%-34.3%) | **87 / 256, 34.0% (28.5%-40.0%)** |
| simic-ramp | 16 | 15 | 10 | 10 | 14 | 16 | 9 | 8 | 49 / 256, 19.1% (14.8%-24.4%) | 49 / 256, 19.1% (14.8%-24.4%) |
| turns p50 | 12 | 12 | 12 | 12 | 11 | 11 | 12 | 11 | | |

**Run 2** (`heuristic` wins, of 48 games per half):

| | Seed 1 half A | Seed 1 half B | Seed 1001 half A | Seed 1001 half B | Pooled `heuristic` | Pooled `heuristic-baseline` |
|---|---|---|---|---|---|---|
| off | 35 | 35 | 39 | 28 | 137 / 384, 35.7% (31.0%-40.6%) | 55 / 384, 14.3% (11.2%-18.2%) |
| on | 36 | 35 | 41 | 28 | 140 / 384, 36.5% (31.8%-41.4%) | 52 / 384, 13.5% (10.5%-17.3%) |

`heuristic` on mono-red-aggro in run 2: 31 / 96, 32.3% (23.8%-42.2%) off; 36 / 96, 37.5% (28.5%-47.5%) on.

**mono-red-aggro's cards, run 1** (games used of games offered, and windows taken of windows offered, pooled over the four seeds):

| Card | off | on |
|---|---|---|
| Hordeling Outburst, cast | 10 of 38 games (10 of 372 windows) | **33 of 35 (33 of 102)** |
| Krenko's Command, cast | 20 of 42 (20 of 341) | 37 of 43 (37 of 187) |
| Dragon Fodder, cast | 29 of 50 (29 of 512) | 47 of 51 (47 of 216) |
| Krenko, Mob Boss, activation | 4 of 30 (4 of 734) | **32 of 36 (61 of 242)** |
| Pashalik Mons, activation | 2 of 32 (2 of 338) | 2 of 35 (2 of 407) |

**Other gates on the branch:**

- `boteval suite run --policy heuristic`: 42 of 42, every tag at 100%.
- `go test -race ./...` and `golangci-lint`: green, including `TestHordelingOutburstIsWorthMoreWithPurphorosOut`, `TestTokenPayoffNeedsMineAndCreaturesAndTheSwitch`, `TestKrenkoActivationIsPricedByItsGoblins`, `TestCreatureEnterPayoffDeclarations`, `TestCreatureTokenDeclarations`, `TestCreatureEnterPayoffGuard` and `TestTokenPayoffsOnTheWire`.

### #3146: landfall doublers, payoff casts and fetch lands (2026-10-14)

The [amendment of 2026-10-14](#amendment-2026-10-14-landfall-doublers-payoff-casts-and-fetch-lands). The base is `develop` at `1687b44ee`. The branch is that tree with the three switches and the declarations. The comparison is one tree built twice, with `PriceLandfallDoublers`, `PriceLandfallCasts` and `PriceFetchLandfall` off and then on, so the declarations alone change nothing. Every run is `--rotate --lockstep` with the real dump. Run 1 is §8's at seeds 1, 1001, 2001 and 3001. Run 2 is §8's `--games 48`, with half B (`heuristic` on mono-black-aristocrats and simic-ramp) at seeds 1 and 1001 and half A at seed 1. **0 stalls and 0 rejected moves in every run.** A logged copy of run 1 at seed 1 replayed the unlogged games exactly, and was deleted once the tables were extracted.

**Run 1** (wins of 64 per seed block):

| Deck | Seed 1, off | Seed 1, on | Seed 1001, off | Seed 1001, on | Seed 2001, off | Seed 2001, on | Seed 3001, off | Seed 3001, on | Pooled, off | Pooled, on |
|---|---|---|---|---|---|---|---|---|---|---|
| esper-control | 20 | 18 | 20 | 16 | 12 | 16 | 15 | 12 | 67 / 256, 26.2% (21.2%–31.9%) | 62 / 256, 24.2% (19.4%–29.8%) |
| mono-red-aggro | 24 | 23 | 18 | 23 | 29 | 23 | 23 | 26 | 94 / 256, 36.7% (31.1%–42.8%) | 95 / 256, 37.1% (31.4%–43.2%) |
| mono-black-aristocrats | 10 | 11 | 14 | 11 | 11 | 13 | 17 | 15 | 52 / 256, 20.3% (15.8%–25.7%) | 50 / 256, 19.5% (15.1%–24.8%) |
| simic-ramp | 10 | 12 | 12 | 14 | 12 | 12 | 9 | 11 | 43 / 256, 16.8% (12.7%–21.9%) | **49 / 256, 19.1% (14.8%–24.4%)** |
| games that differ | — | 31 | — | 23 | — | 27 | — | 22 | | |
| turns p50 | 12 | 12 | 12 | 11 | 11 | 11 | 11 | 11 | | |

**Run 2** (`heuristic` wins, of 48 games per half):

| | Seed 1 half A | Seed 1 half B | Seed 1001 half B | `heuristic` | `heuristic` on Simic, half B |
|---|---|---|---|---|---|
| off | 38 (esper 17, red 21) | 31 (black 12, Simic 19) | 28 (black 13, Simic 15) | 97 / 288, 33.7% (28.5%–39.3%) | 34 / 96, 35.4% (26.6%–45.4%) |
| on | 38 (esper 17, red 21) | 28 (black 15, Simic 13) | 27 (black 13, Simic 14) | 93 / 288, 32.3% (27.2%–37.9%) | 27 / 96, 28.1% (20.1%–37.8%) |
| games that differ | 5 | 20 | 26 | | |

**Simic's landfall cards, run 1** (games used of games offered, four seed blocks):

| Card | off | on |
|---|---|---|
| Tireless Tracker, cast | 73 of 81 | **87 of 91** |
| Sporemound, cast | 62 of 97 | 74 of 92 |
| Avenger of Zendikar, cast | 66 of 74 | 76 of 82 |
| Courser of Kruphix, cast | 64 of 71 | 72 of 76 |
| Tireless Provisioner, cast | 64 of 71 | 65 of 68 |
| Lotus Cobra, cast | 63 of 72 | 64 of 70 |
| Rampaging Baloths, cast | 69 of 69 | 69 of 71 |
| Ancient Greenwarden, cast | 81 of 81 (85 of 186 windows) | 77 of 79 (78 of 150 windows) |
| Burnished Hart, activation | 21 of 35 | 16 of 28 |

The arena's Cards table leaves lands out, so the fetches are read from the decision logs of run 1 at seed 1, for the Simic seat. With a landfall payoff of its own out, a fetch was cracked in 77 of 1,020 windows it was offered in (30 of 36 games) with the switches off, and in 90 of 534 (35 of 36 games) with them on. With no payoff out, it was cracked in 170 of 1,315 and 154 of 882 (61 of 63 games on both). Payoff casts in those logs moved by about a turn at most: Greenwarden's median turn went from 8 to 7, and Courser's from 6 to 5.

**Other gates on the branch:**

- `boteval suite run --policy heuristic`: 42 of 42.
- `go test -race ./...` and `golangci-lint`: green, including `TestGreenwardenDoublesTheLandfallPrice`, `TestALandfallPayoffIsWorthTheLandsAhead`, `TestGreenwardenCastIsWorthTheDoubling`, `TestAFetchPaysTheLandfallPayoffs`, `TestLandfallDoublerDeclarations`, `TestLandfallDoublerGuard` and `TestFetchLandDeclarations`. With the dump, `TestCuratedDeckPurposes` and `TestRealDumpPurposeLandsUntapped` pass with the seven untapped fetches on `curatedLandsUntapped`.

## Amendment 2026-10-15 (b): the pump of a turn-face-up trigger, and bodies that leave

**Status:** Accepted. The code is in the PR that adds this section. Two `Config` switches, both on in `DefaultConfig()` and off in `BaselineConfig()`: `PriceFaceUpPump` and `PriceTemporaryBodies`.
**Issues:** [#3776](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3776) (the "Not priced" paragraph of the amendment above), [#3802](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3802). Sprint: none.

1. **`Purpose.FaceUp` may declare a `Pump`.** `Pump` was a triggered or activated row's field; the inner purpose of `FaceUp` is a turn-face-up trigger's effect, so it may carry one too (`effects.Register`: a `purposeOnFaceUp` slot, which `checkPump` also validates). Unstable Hulk declares `Pump{6, 6, trample}`. `Kicked` still may not.
2. **The flip is priced by the attack the pump enables.** `valueOfTurnFaceUp` adds, for a creature that can still attack this turn (untapped and not summoning-sick before attackers, or already attacking before damage), the attack value of the pumped creature less that of the plain one, per opponent, and takes the best. When the pump is what turns the swing into a kill (`lethalPushAs` true with the pumped creature and false without) it adds `LethalBonus`, so a flip for lethal outweighs the skipped turn's `Weights.Turn`. Out of that window the pump is worth nothing.
3. **A creature cast for evoke, warp, dash or blitz is a temporary body.** `valueOfCast` reads the alternative-cost key off the move and replaces the lasting body `resolvedValueFor` priced with what the cast buys: its enters effect (unchanged, the card's purpose), a haste attack this turn when it has one and it is still ahead, blitz's death draw, and the card dash and warp keep. Evoke keeps only the enters effect.


## Amendment 2026-10-24: a change until end of turn, priced by whose permanent it is

**Status:** Accepted. The code is in the PR that adds this section. One `Config` switch, `PriceTemporaryEffects`, with its weight `CombatGrant` (0.50), on in `DefaultConfig()` and off in `BaselineConfig()`; the arena's `heuristic-notemporary` is `DefaultConfig()` with it off.
**Issue:** [#4302](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4302), from real games. Sprint: S75 — Agents at the table: a current seat, quieter turns, sharper bots.
**Amends:** the amendment of 2026-10-08 (a target priced by what the purpose does to it).

### The problem

With Heliod, Sun-Crowned out, a heuristic seat ignored the creature in its hand and its castable commander, spent its mana on "{1}{W}: Another target creature gains lifelink until end of turn" aimed at an **opponent's** creature, and did nothing else that turn. Two causes, both in `targetsValue`'s sign convention, "a target is hit":

1. **The target.** Nothing on the wire said what the row does to its pick, so `cardTargetValue` priced an opposing creature as removed (its value × `RemovalConfidence` × `LeaderBoost`, about 2.5 for a 3/3) and the bot's own as `OwnPermanentTarget` (0.40). Lifelink on the opponent's creature was the best target on the list; `RemovalConfidence`'s own comment named the case ("an unrestricted 'target creature gets +3/+3' may target theirs").
2. **The ranking.** That price plus the flat `ActivateBase` (0.50) beat a two-drop's cast (+1.41) and a commander's (+2.41). Nothing was removed, so the same move was on offer at the same price in the next window, and the bot repeated it until its mana ran out: three activations off six Plains in the whole-engine repro.

### Decision

1. **A target entry may declare `UntilEndOfTurn`**, `TargetBenefit` or `TargetHarm`, with an optional `Pump` (the printed power and toughness, negative for a harm, and a benefit's keywords). `effects.BenefitUntilEOT` and `effects.HarmUntilEOT` build it. `effects.Register` refuses another value, a `Pump` without a direction, a benefit that takes power or toughness, a harm that adds any or grants a keyword, and a clause that cannot pick a permanent. The wire carries `until_end_of_turn` and `pump` on the entry.
2. **The heuristic prices such a permanent pick by side and by this turn** (`temporary_effect.go`): a benefit on an opponent's permanent, or a harm on the bot's own, costs the permanent's value (`hostWorth`, the misplaced Aura's price); a benefit on the bot's own creature is worth the attack it adds this turn (`pumpAttackGain`, shared with the turn-face-up pump: attack with the pump less attack without, `LethalBonus` when it makes the kill, haste letting a summoning-sick creature swing), plus `CombatGrant` while that creature is attacking or blocking before damage; anywhere else nothing. A harm is removal when its -N/-N kills (CR 704.5f, 704.5g), `CombatGrant` on a creature in combat when it does not, nothing elsewhere, and the old price when it prints no amount.
3. **No loop.** A benefit that adds nothing (every keyword it grants already on the creature, CR 702.15f) is below a pass, and so is a grant with no power or toughness that the same row already gave the same creature this turn: the policy remembers those per row and target for the turn, as `free_repeat.go` remembers a free activation, because a grant still on the stack shows nothing on the creature yet.
4. **Spare mana in combat.** Such a move on a creature in combat that costs only mana and taps clears `LeftoverThreshold` when the mana left after it still pays for the biggest card in the bot's hand or command zone its mana could cast; on another seat's turn all of it is spare. Otherwise it needs `InstantThreshold`, so the bot keeps the mana for its second main phase.
5. **The catalog declares it.** Every activated row whose text changes a target until end of turn declares a direction that matches its words: 74 rows, held by `TestUntilEndOfTurnRowsDeclareTheirDirection`. Exempt: Liliana of the Dark Realms (the sign is chosen at resolution), Soulbright Seeker (its third resolution adds mana) and Zack Fair (part of it lasts).

### Measurements

`boteval arena --lockstep --rotate --games 48 --seed 4302`, `heuristic` against `heuristic-notemporary`, two runs on mono-white-weenie, boros-voltron, mono-green-stompy and simic-ramp with the policies swapped between the decks, so each policy played each deck 48 times:

| | heuristic | heuristic-notemporary |
|---|---:|---:|
| all four decks | 52/192, 27.1% (21.3–33.8%) | 44/192, 22.9% (17.5–29.4%) |
| mono-white-weenie (Heliod, Mother of Runes) | 17/48, 35.4% | 13/48, 27.1% |
| boros-voltron (Mother of Runes, Rogue's Passage) | 18/48, 37.5% | 15/48, 31.2% |
| mono-green-stompy (Rogue's Passage) | 2/48 | 2/48 |
| simic-ramp (none) | 15/48 | 14/48 |

Heliod's activation: `heuristic-notemporary` took it in 68 of the 68 windows it was offered in, in all 8 games; `heuristic` in 15 of 260, in 3 of 9 games. Mother of Runes went from 39 activations in 360 windows to 8 in 603: out of combat it buys nothing. The intervals overlap, so the win rate is a direction, not a result. On the standard pool (mono-red-aggro, simic-ramp, esper-control, mono-black-aristocrats), which holds none of the declared cards, the same two runs gave identical per-deck wins and every game ended on the same turn: the change does nothing there.

### Not priced

- A protection grant aimed at the bot's own creature in answer to removal (Mother of Runes, Liliana the Faultless, Plaza of Heroes): `answers.go` prices only untargeted rows on the threatened creature.
- A pump or evasion grant with no printed amount before combat (+X/+X, "can't be blocked"): only `CombatGrant`, once the creature is attacking.
- Spells and triggered rows: they may declare the entry and are read the same way, but none does yet.
- Heliod's own lifegain trigger, which turns lifelink into +1/+1 counters.

## Amendment 2026-10-24 (b): spells, grants that save, and evasion before attacks

**Status:** Accepted. The code is in the PR that adds this section. Two `Config` switches, `PriceGrantSaves` and `PriceEvasionGrants`, on in `DefaultConfig()` and off in `BaselineConfig()`; the arena's `heuristic-nogrants` is `DefaultConfig()` with both off. The spells' declarations are read only through `PriceTemporaryEffects`, so `heuristic-notemporary` turns all of it off.
**Issue:** [#4312](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4312), the follow-up the amendment above left. Sprint: S75 — Agents at the table: a current seat, quieter turns, sharper bots.
**Amends:** the amendment above (a change until end of turn, priced by whose permanent it is).

### Decision

1. **Spells declare the entry.** Every catalog instant, sorcery and mode whose text changes a permanent it targets until end of turn declares `UntilEndOfTurn` on that target's entry: 106 clauses, held to their words by `TestUntilEndOfTurnSpellsDeclareTheirDirection` from the oracle fixtures. An amount decided at resolution (-X/-X, Tragic Slip's morbid -13/-13, Hatred's +X/+0) declares the direction only. A lasting rider that helps the same creature (a +1/+1 counter, a Role) does not stop a benefit and is left unpriced. A change that is partly a harm (Berserk's end-step destruction), the grant a card's own sweep spares (Fated Clash) and a grant that hurts the creature it helps (Martyrdom) declare nothing. A rider the purpose can say is declared beside the entry (a card drawn, a Treasure, life gained, energy).
2. **"Can't be blocked this turn" is declared.** `Pump.CantBeBlocked` (`cant_be_blocked` on the wire), built by `effects.CantBeBlockedUntilEOT`, on every row and spell that says it with no "except"; both audits hold the words to it. Protection from a quality chosen at resolution is the keyword `protection`.
3. **A grant that keeps the bot's permanent is priced as the save** (`grant_save.go`). With the stack about to remove that permanent, a benefit aimed at it is worth the permanent's value times the chance it is lost, `answerSave`'s price for an untargeted row, when it keeps it: hexproof, shroud or protection against what targets it and not a sweep (CR 608.2b, 702.11b, 702.18a, 702.16b), protection only when every targeting threat shares a colour; indestructible, undying and persist against destruction and lethal damage, half against a targeted removal that declares no damage (CR 702.12b, 702.93a, 702.79a); toughness against declared damage it outgrows. A save already on the stack for the same permanent is not made twice. One that costs only mana and taps clears `LeftoverThreshold` when its mana is spare.
4. **Protection names the colour of the removal.** The `protect` colour policy ranks an opponent's stack item aimed at one of the bot's permanents above every threat on the board.
5. **Evasion is priced before blockers.** With `PriceEvasionGrants`, `pumpAttackGain` puts the `cant_be_blocked` restriction on the pumped creature and credits a trample grant with the damage over the defender's biggest blocker at the defender's life value (CR 702.19b). After blockers are declared neither counts (CR 509.1h).

### Measurements

`boteval arena --lockstep --rotate --seed 4312 --games 400 --decks tricks-battle,tricks-battle` (two seats; the deck is BattleDeck with four Brute Force, three Dismember, four Mother of Runes, three Center Soul and a Rogue's Passage):

| run | heuristic | the other seat |
|---|---:|---:|
| against `heuristic-notemporary` | 285/400, 71.2% (66.6–75.5%) | 115/400, 28.7% |
| against `heuristic-nogrants` | 223/400, 55.8% (50.9–60.5%) | 177/400, 44.2% |

`heuristic-notemporary` casts Brute Force in 388 of 1648 windows it is offered in, most of them at an opponent's creature; `heuristic` in 357 of 8360. `heuristic-nogrants` never casts Center Soul (0 of 12776 windows); `heuristic` casts it in 104 of 10929, in answer to removal.

On the four curated decks #4302 measured (mono-white-weenie, boros-voltron, mono-green-stompy, simic-ramp; `--games 48 --seed 4302`, two runs with `heuristic` and `heuristic-nogrants` swapped between the decks), each policy won the same games on each deck: 13/48, 16/48, 4/48 and 15/48, 48/192 in all for both. Those decks hold Mother of Runes, Rogue's Passage and Boros Charm and no pump spell, and their opponents' removal rarely meets an untapped Mother: the change neither helps nor costs there. On tricks-battle the interval against `heuristic-nogrants` clears the 50% null by under a point, so that run is a direction; the one against `heuristic-notemporary` is a result.

### Not priced

- +X/+X before combat, and a grant whose amount is decided at resolution.
- Heliod's lifegain trigger, which turns lifelink into +1/+1 counters; haste letting a creature use its {T} abilities (CR 702.10c); Soulbright Seeker's third activation.
- Scry, learn, discover and the other riders a purpose cannot say.
