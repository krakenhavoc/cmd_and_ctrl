# ADR 0126 — Bots that play their decks: the heuristic prices what a card does

**Status:** Accepted · 2026-10-06 · S66 — Bots that play their decks. The owner accepted it on 2026-10-06 and answered its eight open questions, all as recommended. The answers are recorded under [Owner decisions](#owner-decisions-2026-10-06). S66 closed on 2026-10-07 with the owner's [exit decisions](#exit-decisions-2026-10-07) on the acceptance bar.
**Issues:** [#2435](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2435) (this change). [#2436](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2436), the curated deck rebalance, waits on it. [#2437](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2437), a fifth curated deck, comes after both.
**Owner direction:** 2026-10-06, on #2435: fix the pricing before the rebalance, write an ADR before changing any weight, and measure it with [ADR 0052](0052-bot-decision-harness-and-eval.md)'s arena report on the curated decks, with the nightly gates green.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-06. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote head: 37 of them (`origin/develop`, `origin/main`, `pr/2326`, and 34 chore, docs, feat, fix, repro and wip branches). The highest number on any of them is 0125, on `origin/develop`, `origin/main` and `origin/feat/table-defaults-row-overlay`. This ADR takes **0126**.
**Amendments:** 2026-10-06, [discard payoffs](#amendment-2026-10-06-discard-payoffs) (accepted). 2026-10-08, [purposes that follow a mode's target, and damage priced by whether it kills](#amendment-2026-10-08-purposes-that-follow-a-modes-target-and-damage-priced-by-whether-it-kills) ([#2689](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2689)): accepted (owner answers 2026-10-08).
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

1. **The four-deck rotation.** `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1`. 64 games is 16 full rotations, so every deck sits in every chair 16 times. It reports per-deck win rates with intervals, the Cards table and the `never` counts. About 5 s a game, so about 6 minutes.
2. **New against baseline.** Two `heuristic` and two `heuristic-baseline` contestants, 48 games with the decks split one way and 48 with them swapped, so each policy plays each deck equally (96 games, 192 seat-games per policy). This measures strength, which run 1 cannot: four copies of one policy win 25% each by construction.
3. **The position suite.** `boteval suite run --policy heuristic`. Every gated position passes. Overall agreement and each tag's agreement are not lower than before the PR. Each sub-PR adds the positions for its class (below), labelled from its own arena decision logs, and gates them.
4. **The nightly gates.** `bot-games`, `bot-soak` (the heuristic gate's fixed seeds 101..120, the random soak of 100 games) and `catalog-soak` stay green on the PR's branch, run with `gh workflow run` before merge. If the catalog soak fails with an `EventEffectError` on a card the bot now casts for the first time, that is a real engine bug the pricing has uncovered. It is filed as its own issue and fixed, or the card's `Completeness` is corrected. It is not a reason to revert the price.

**Acceptance bar for S66** (the last sub-PR's run 1 and run 2, on today's four decks, before #2436):

| # | Measure | Today | Bar | Why this number |
|---|---|---|---|---|
| A1 | `never` count per deck (run 1) | 19–30 | **≤ 5 for every deck** | Some cards are legitimately held: a counterspell with nothing to counter, a removal spell with no target worth it, Kismet. Five per deck leaves room for those and nothing else. |
| A2 | Mana rocks and dorks (run 1) | used in 0% of games where offered | **used in ≥ 80% of the games in which each was offered** (amended for S66 on 2026-10-07: see [Exit decisions](#exit-decisions-2026-10-07)) | Casting a rock is almost always right. 80% leaves room for a turn where a land and a bigger spell use all the mana. |
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

### #2689 PR 5: target gifts across the catalog (2026-10-08)

Owner answer 5a: every catalog spell, mode or row that gives its target player a draw, a token, life or a loot declares a target entry (`Purpose.Targets`), so `PriceTargetPurposes` prices it as the seat's strength change instead of as a hit. Burn is not in this batch.

Declared (23 cards): Ancestral Vision, Atlantis Attacks (the Leviathan mode), Blessed Alliance (the life mode), Blood Pact, Bloodgift Demon, Cease // Desist (Cease), Cephalid Coliseum, Compulsive Research, Deep Analysis, Depth Defiler (the loot mode), Echocasting Symposium, Etched Oracle, Flame of Anor (the draw mode), Flumph, Forbidden Orchard, Insatiable Avarice (the drain-and-draw mode), Loran of the Third Path, Oona's Grace, Rise of the Eldrazi (slot 1 only), Scheming Silvertongue's Sign in Blood, Secret Rendezvous, Sublime Epiphany (the draw mode) and Wedding Ring. Flumph, Loran and Secret Rendezvous also declare the controller's own draw, because "you and target opponent each draw" is two gifts. Compulsive Research declares the printed two discards, not the one-land alternative.

Left undeclared, because no printed number says the amount:
- X, or an amount counted at resolution: Blue Sun's Zenith, Damnable Pact, Drown in Dreams, Heliod's Intervention, Inscription of Abundance (the greatest power), Kozilek's Command, Peer into the Abyss (half a library), Stroke of Genius.
- Generous Plunderer: the target opponent is chosen by a reflexive trigger, which has no row to carry a purpose.
- Treacherous Pit-Dweller: what it hands over is a creature.

The real-dump audit's "target gift" list goes from 30 cards to those 10. The gifts to the controller of a removal spell's target (Swords to Plowshares and the rest of #2679's list) cannot be declared here: their clause is a creature, and the guard refuses a player amount on one.

**Curated decks.** Only Loran of the Third Path (esper-control) gains an entry. `boteval arena --seats heuristic,heuristic,heuristic,heuristic --decks esper-control,izzet-aggro,mono-black-aristocrats,simic-ramp --games 64 --rotate --seed 1 --lockstep`, with the binaries built from `develop` at `50d5c34ec` and from this branch: 0 stalls in both, turns p50 13 in both, and wins esper-control 18 / 17, izzet-aggro 3 / 3, mono-black-aristocrats 21 / 21, simic-ramp 22 / 23. Nine games differ, and every one has a Loran in play: Loran's activation is taken 0 to 2 times in eight of them and 6 times in one, against 1 to 4 before, because aiming a symmetric draw at an opponent is no longer priced as an attack. Two more games differ without a Loran in either deck; a control run of the `develop` binary over seeds 1 to 3 differs from its own earlier run on seed 3 in the runner counters, so that is the residual tie-break nondeterminism, not this change. The suite is 41 of 41 before and after.
