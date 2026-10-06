# ADR 0126 — Bots that play their decks: the heuristic prices what a card does

**Status:** Accepted · 2026-10-06 · S66 — Bots that play their decks. The owner accepted it on 2026-10-06 and answered its eight open questions, all as recommended. The answers are recorded under [Owner decisions](#owner-decisions-2026-10-06).
**Issues:** [#2435](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2435) (this change). [#2436](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2436), the curated deck rebalance, waits on it. [#2437](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2437), a fifth curated deck, comes after both.
**Owner direction:** 2026-10-06, on #2435: fix the pricing before the rebalance, write an ADR before changing any weight, and measure it with [ADR 0052](0052-bot-decision-harness-and-eval.md)'s arena report on the curated decks, with the nightly gates green.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-06. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote head: 37 of them (`origin/develop`, `origin/main`, `pr/2326`, and 34 chore, docs, feat, fix, repro and wip branches). The highest number on any of them is 0125, on `origin/develop`, `origin/main` and `origin/feat/table-defaults-row-overlay`. This ADR takes **0126**.
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

1. **The four-deck rotation.** `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1`. 64 games is 16 full rotations, so every deck sits in every chair 16 times. It reports per-deck win rates with intervals, the Cards table and the `never` counts. About 5 s a game, so about 6 minutes.
2. **New against baseline.** Two `heuristic` and two `heuristic-baseline` contestants, 48 games with the decks split one way and 48 with them swapped, so each policy plays each deck equally (96 games, 192 seat-games per policy). This measures strength, which run 1 cannot: four copies of one policy win 25% each by construction.
3. **The position suite.** `boteval suite run --policy heuristic`. Every gated position passes. Overall agreement and each tag's agreement are not lower than before the PR. Each sub-PR adds the positions for its class (below), labelled from its own arena decision logs, and gates them.
4. **The nightly gates.** `bot-games`, `bot-soak` (the heuristic gate's fixed seeds 101..120, the random soak of 100 games) and `catalog-soak` stay green on the PR's branch, run with `gh workflow run` before merge. If the catalog soak fails with an `EventEffectError` on a card the bot now casts for the first time, that is a real engine bug the pricing has uncovered. It is filed as its own issue and fixed, or the card's `Completeness` is corrected. It is not a reason to revert the price.

**Acceptance bar for S66** (the last sub-PR's run 1 and run 2, on today's four decks, before #2436):

| # | Measure | Today | Bar | Why this number |
|---|---|---|---|---|
| A1 | `never` count per deck (run 1) | 19–30 | **≤ 5 for every deck** | Some cards are legitimately held: a counterspell with nothing to counter, a removal spell with no target worth it, Kismet. Five per deck leaves room for those and nothing else. |
| A2 | Mana rocks and dorks (run 1) | used in 0% of games where offered | **used in ≥ 80% of the games in which each was offered** | Casting a rock is almost always right. 80% leaves room for a turn where a land and a bigger spell use all the mana. |
| A3 | The six canaries: Sol Ring, Rhystic Study, Mary Read's loot, Entomb, Harrow, Viscera Seer (run 1) | 0 uses | **each used in ≥ 50% of the games in which it was offered** | They name the six classes. If one stays dead, its class is not fixed. |
| A4 | Wipes (suite) | always cast | **both new wipe positions pass, and are gated** | Arena games rarely set up the bad-wipe board, so a fixed position is the reliable test. |
| A5 | New against baseline (run 2) | n/a | **`beats null: yes` for `heuristic`**: the Wilson interval's lower bound above 25% | The claim is that the bot plays its deck better, so it must beat the old bot in a fair seating. With 192 seat-games, that needs about 31% or more (about 60 of 96 games). |
| A6 | Balance (run 1) | mono-black 63%, interval 49.9%–75.2% | **no deck's interval lies entirely above 50%** | The pricing must not widen black's lead. Making the decks even is #2436's job. This bar only says this ADR did not make it worse. |
| A7 | Suite | 22 positions | **every gated position passes; no tag's agreement falls** | ADR 0052 §1: a trade is not a win. |
| A8 | Health (runs 1 and 2) | 0 stalls | **0 stalls; rejected moves not above the baseline run's; turns p50 within ±3 of baseline** | More card classes in play means more chances of a move the engine refuses or a table that does not end. |

Each intermediate sub-PR must hold A7 and A8, and its run 2 must not be detectably worse: the `heuristic` interval's upper bound must not fall below 25%. Its own class's canaries should move, and the PR says by how much. A1 to A6 are judged at the end, because the classes interact. Casting rocks means more mana for engines, and the `never` count only falls far once all of them land.

If A5 or A6 fails at the end, the owner decides whether to ship, retune or narrow the scope, with the numbers in [Measurements](#measurements).

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
