# ADR 0126 — Bots that play their decks: the heuristic prices what a card does

**Status:** Proposed · 2026-10-06 · S66 — Bots that play their decks. Awaiting the owner's review. No weight changes until it is accepted.
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

`CombatValue`, which the combat planner uses to compare an attacker with a blocker, is not changed in this ADR (open question 3).

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
- **The combat planner.** `CombatValue` stays body-only (open question 3). How the bot attacks and blocks is not changed here.
- **The flat `ActivateBase` for a non-tap activated ability.** Konrad's mill is the visible case. It stays at 0.50 (open question 6). §6's purpose would let a later change price these rows.
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
| 2 | **Measurement.** The arena's per-contestant Play table and Cards section, `heuristic-baseline` and `BaselineConfig()` (with no new terms yet, so baseline and default are equal), and the baseline runs recorded in Measurements. No price changes. | 1 accepted |
| 3 | **Mana sources** (§2): the `ManaPerExtra` value on the battlefield and the ramp premium at cast time. Positions: the rock and elf positions. | 2 |
| 4 | **Permanents by what they do** (§3): the mana-value floor and `rowUtility`, in `permanentValue` and `creatureValue`. Positions: Rhystic Study, Viscera Seer, equip. | 2 |
| 5 | **The two windows** (§5): own second main and the end step before the bot's turn, `SpellFloor`, and the tapped-blocker price by timing. Positions: loot, tutor, cantrip. | 2 |
| 6 | **The purpose signal** (§6): `PurposeView`, `Spec.Purpose` and its row forms, the projection and redaction, `docs/protocol.md`, declarations for the curated decks and every catalog wipe, `TestCuratedDeckPurposes`, and the manual dump audit. No price changes. Touches `effects`, `game`, `protocol` and `decks`, and nothing under `aiseat/`. | 1 accepted |
| 7 | **Wipes, ramp spells and discard costs** (§4, §6's prices, §7's discard half). Positions: the two wipe positions and the two Windfall positions. | 3, 6 |
| 8 | **Sacrifice outlets** (§7's second half): dying anyway and death payoffs. Positions: the two sacrifice positions. | 4, 6 |
| 9 | **Exit.** The final run 1 and run 2 against A1 to A8, the Measurements rows, a "How the heuristic prices a card" section in `docs/bot.md` (replacing the pricing claims in its Known limitations), and the evidence on #2435. | 3–8 |

PRs 3, 4, 5 and 6 are independent of each other once PR 2 is in, so they can be built in parallel and merged in any order. Each is measured against the `develop` it merges into.

---

## Open questions for the owner

Each lists the recommendation first.

1. **How does the bot learn what a spell does?**
   - **Recommended:** a `Purpose` declared on the catalog `Spec` and its rows, projected onto the view (§6), with a manual audit over the Scryfall dump to find undeclared cards. This is faithful: the catalog is already where the engine's truth about a card lives, and ADR 0106 set the precedent of the bot reading declared data there.
   - Derive it on the server from the oracle text with a closed set of patterns. Every card is covered at once, but an English parser is a second, weaker reading of the card next to the catalog's, and it will disagree with the engine on exactly the odd cards.
   - No new signal: generic floors only. Board wipes then stay uncastable or always castable, never board-aware.
2. **Which cards get a purpose in S66?**
   - **Recommended:** the curated decks' cards in the priced classes, plus every catalog board wipe. A wipe without one is the only case where the bot harms itself.
   - The curated decks only.
   - Every catalog card in the classes. Several hundred cards, and better done as a catalog sweep later.
3. **Should a creature's ability rows count in combat (`CombatValue`) too?**
   - **Recommended:** not in S66. Count them in the board value only, measure, and open a follow-up issue if the logs show the bot trading a Blood Artist for a vanilla 2/2. Combat is the best-tuned part of the heuristic and has the most suite positions.
   - Yes, in PR 4, so that the board and combat agree.
4. **How strict is the strength bar?**
   - **Recommended:** A5 as written. At exit the new heuristic must beat the baseline outright (interval above the null over 96 games), and each step in between must not be detectably worse.
   - Only "not worse" at exit. Easier to meet, but it would let a change that makes the bot cast everything and play no better count as done.
5. **What happens to `heuristic-baseline` after S66?**
   - **Recommended:** keep it as the frozen reference until the next ADR that changes the heuristic's prices. That ADR then replaces it with a new frozen config. It costs one config function and one arena name.
   - Delete it at S66 exit.
6. **Konrad's flat-priced mill, and every non-tap activated ability at +0.50.**
   - **Recommended:** leave it in S66, and let A6 show whether black's lead survives once the other decks cast their cards. Price these rows by purpose in a later change if it does.
   - Lower `ActivateBase` now. That changes every activated ability in the catalog in the same PR as everything else, and the measurement would no longer say which change did what.
7. **Measure on today's decks or the rebalanced ones?**
   - **Recommended:** today's four decks, as the bar says. The pricing is then judged on the decks whose evidence motivated it, and #2436 is measured afterwards with the fixed heuristic, as the owner ordered.
   - On #2436's lists. That would mix the two changes' effects.
8. **The end-step window: cheap instants and tap abilities at the end of the seat before the bot (§5).**
   - **Recommended:** both §5 windows, the bot's own second main and the end step before its turn. That is when a human spends an instant they did not need, and it is the only window in which a tap loot costs nothing.
   - The second main only. Simpler, but the bot then casts every instant at sorcery speed and loots only on its own turn, giving up its blocker.

---

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

*Empty on purpose. PR 2 records the baseline: run 1's per-deck win rates, intervals and `never` counts, and run 2's baseline-against-baseline sanity row. Each later PR appends its row. Every row records: git SHA, `BaselineConfig()` and `DefaultConfig()` deltas, seeds, games, per-deck win rate with Wilson 95% interval, `never` counts, canary use rates, run 2's policy win rate with its interval, suite agreement overall and per tag, stalls, rejected moves, and turns p50.*
