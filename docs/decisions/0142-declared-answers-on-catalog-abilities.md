# ADR 0142 — Declared answers on catalog abilities

**Status:** Accepted · 2026-10-09 · S60 — Table clarity: a stack you can follow. The owner answered all seven [questions](#questions-for-the-owner-answered-2026-10-09) on 2026-10-09, each with the recommended option (a). No code changed with it. The changes land in the PRs under [Delivery](#8-delivery).
**Issue:** [#2872](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2872). Related: [#2871](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2871) (smart auto-pass: "stop only if I can respond", and combat abilities counted during combat), [#2853](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2853) / PR #2856 (where `interacts` came from).
**Amends:** [ADR 0126](0126-bots-that-play-their-decks.md) §6 (`Purpose` gains one field) and [ADR 0009](0009-smart-priority-autopass.md)'s [#2853 amendment](0009-smart-priority-autopass.md#amendment-only-real-interaction-stops-you-2853) (`interacts` reads the declaration first). Both ADRs keep their decisions; this one changes where one input comes from.
**Numbering:** the AGENTS.md §4 sweep on 2026-10-09 (`git fetch --all --prune`, then `docs/decisions/` listed on every remote head: `origin/develop`, `origin/main`, `pr/2326`, `origin/cost-ledger`, `origin/docs/issue-audit`, `origin/feat/2862-bestow`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/one-cast-per-card-type`, `origin/feat/playmats`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default`). The highest number on any of them is 0141 (`0141-bestow.md` on `origin/feat/2862-bestow`). This ADR takes **0142**.
**Builds on:** [ADR 0126](0126-bots-that-play-their-decks.md) §6 (a declared `Purpose`, never inferred), [ADR 0009](0009-smart-priority-autopass.md) (smart auto-pass and its response classes), [ADR 0033](0033-ai-bot-seat.md) §3 (a policy reads the view and the move list only), [ADR 0052](0052-bot-decision-harness-and-eval.md) (how a bot change is measured), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (granted ability rows).

**Why a new ADR, not a fourth amendment to ADR 0126.** The field rides on `Purpose`, so it could be an ADR 0126 amendment. It is not one, for three reasons. Its first reader is smart auto-pass, not the bot, so it is an ADR 0009 change as much as an ADR 0126 one. It brings a ratchet test and a catalog sweep of about forty PRs' worth of rows, which is a programme, not a pricing tweak. And ADR 0126 is already 1,668 lines with three amendments, so a reader looking for "why does auto-pass stop for Albino Troll" would not find it there. Both ADRs get a one-line pointer here when the signal PR lands.

---

## Context

### The problem

[ADR 0009's #2853 amendment](0009-smart-priority-autopass.md#amendment-only-real-interaction-stops-you-2853) made smart auto-pass stop on an opponent's stack item only for real interaction. For an activated ability with no target, "real interaction" is the legal move's `interacts` flag, and `server/internal/legal/interacts.go` decides it by reading the ability:

1. its **cost**: sacrificing, exiling or returning a creature you control. The sacrificed object is read from the cost's printed label (`TargetSpec.Label`, "another creature", "a Goblin", "a Food"), because the filter itself is a `CardOK` closure;
2. its **declared purpose** (ADR 0126 §6): a pump, a combat-damage shield, damage to a creature, a sweep. Few activated rows declare one;
3. its **printed effect text**, the row's `Label` after the colon: about twenty phrases ("regenerate", "protection from", "prevent", "the next time a source", "monstrosity", "first strike"), a power/toughness regex, a pinging-sweep regex, "destroy" that is not "destroy this", a self-bounce and a blink.

The effect is a closure, so the text was the only description the enumerator had. That is the weakness the issue names. An audit at #2853 read the catalog's 706 untargeted instant-speed activated rows: 245 interact and 461 are value (`interacts_internal_test.go` pins a sample of each). Unusual wording can be misread either way:

- **A false yes is a pointless stop.** The first audit caught charge and storage counters ("put a charge counter" matched the counter rule), damage to each opponent, "destroy this enchantment", and Food, Treasure and land sacrifices. Each was fixed by narrowing a rule, and each narrowing risks the other error.
- **A false no is a missed window**, which is worse for the owner's goal. "Sacrifice this creature: …" rows are value to the text read (Sakura-Tribe Elder, Yavimaya Elder), though sacrificing a creature an opponent's spell targets is a standard answer. Combat grants ("gains flying", "can block an additional creature", "becomes an artifact creature until end of turn") are all "no" today; #2871 adds them as combat-only interaction, with another text read.

The owner's goal for auto-pass (2026-10-09): "make the autopass as smart as possible and as convenient as possible so most of the time players are not thinking why do I have to click to pass or thinking I missed my window to respond". A rule list over English is a parser nobody owns ([ADR 0126, Rejected alternatives](0126-bots-that-play-their-decks.md#rejected-alternatives)). It cannot get both halves right at once.

### What the bot cannot see

The bot has the same blind spot. A non-tap activated ability is priced at a flat `ActivateBase` (0.50; ADR 0126 owner decision 6). In a response window the bar is `InstantThreshold` (1.50). So the heuristic never regenerates Albino Troll in response to removal, never pumps a creature out of burn range, and never uses Selfless Spirit. It cannot tell a regeneration shield from Mind Stone's draw, because the wire carries neither. ADR 0126 §7's "dying anyway" price covers sacrifice outlets only, and only those whose cost the view names.

### What is already declared

`game.Purpose` (`server/internal/game/purpose.go`) is the catalog's declared description of what a spell, a mode, an alternative cost or an ability row does. It is declared by hand on the card file, never derived, projected into the view as `purpose`, and never read by the rules engine. It already carries the amounts the bot prices: draws, lands, tutors, tokens, sweeps, payoffs, `Pump`, `PreventCombatDamageToSelf`, `DamageToCreature` and per-target entries. It says **how much** an effect does, and only for the classes the bot prices. It does not say **what the ability answers**. A regeneration shield, hexproof, a fog, a sacrifice outlet and a "can't cast" lock have no amounts, so most interacting rows have nothing to declare there today.

---

## Decisions

### 1. The field: `Purpose.Answers`, a small declared set

`game.Purpose` gains one field:

```go
// Answers is what this ability (or spell) can do in response: the
// reasons a player would hold priority for it. Declared by hand, like
// every Purpose field. The zero value is "not declared", and the reader
// falls back to the printed-text read. AnswerValue declares "answers
// nothing".
Answers Answers
```

`Answers` is a bit set (`uint16`), so `Purpose` stays comparable and `IsZero` keeps working with `==`. The card file writes it with named constants:

```go
Purpose: game.Purpose{Answers: game.AnswerProtect},                       // Albino Troll: "{1}{G}: Regenerate this creature."
Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{...}},   // "{B}: This creature gets +1/+1 until end of turn."
Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},              // Mind Stone's "{1}, {T}, Sacrifice: Draw a card."
Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerValue}, // refused: value is alone
```

**The vocabulary.** Ten values in two tiers. The tier is fixed in code, per value. It is not declared.

| Value | Wire | Tier | Means | Examples |
|---|---|---|---|---|
| `AnswerProtect` | `protect` | stack | keeps a permanent of yours on the battlefield, or out of reach: regenerate, indestructible, hexproof, shroud, protection, phase out, a blink, returning itself to hand, granting persist or undying | Albino Troll, Selfless Spirit, Aethergeode Miner, Arcanis |
| `AnswerPump` | `pump` | stack | raises power or toughness: +N/+N, +1/+1 counters, monstrosity, adapt, a base power and toughness set until end of turn | Aetherwind Basker, Arbor Colossus, Allosaurus Shepherd |
| `AnswerPrevent` | `prevent` | stack | prevents or redirects damage, or sets a damage shield | Spore Frog, Opal-Eye, Aegis of Honor, Auriok Replica |
| `AnswerRemove` | `remove` | stack | removes, destroys, damages or shrinks other permanents with no target: damage to each creature, destroy all, −N/−N to each, an edict | Pestilence, Nevinyrral's Disk |
| `AnswerSacOutlet` | `sac_outlet` | stack | lets its controller sacrifice a creature at will, in its cost or its effect | Viscera Seer, Ashnod's Altar, Agency Coroner |
| `AnswerRestrict` | `restrict` | stack | stops what an opponent may do next: can't cast, can't activate | Ranger-Captain of Eos |
| `AnswerCombatGrant` | `combat_grant` | combat | grants a combat keyword or permission: flying, menace, trample, haste, lifelink, vigilance, reach, first strike, double strike, deathtouch, "can't be blocked", "can block an additional creature" | Afterthought Sentry, Endling, Anurid Swarmsnapper |
| `AnswerAnimate` | `animate` | combat | becomes a creature until end of turn: a creature land, a Vehicle that animates itself, crew | Foriysian Totem, Bespoke Battlewagon, every crew row |
| `AnswerMakesBlocker` | `makes_blocker` | combat | creates one or more creature tokens at instant speed | Dawn of Hope, Dragonkin Berserker |
| `AnswerValue` | `value` | — | declared: answers nothing. Draw, mana, ramp, a fetch, scry, a non-creature token, a counter that only counts | Mind Stone, Evolving Wilds, a Clue, cycling |

The brief's "protect, pump, sac outlet, prevent, remove or combat-grant" are all here. `restrict`, `animate` and `makes_blocker` cover the rows the audit's "yes" and #2871's combat list hold that none of those six describe. First strike, double strike and deathtouch move from the stack tier (today's text read counts them everywhere) to the combat tier. That is [Q2](#questions-for-the-owner-answered-2026-10-09).

**Rules for a declaration**, enforced by `effects.Register` at boot like every other `Purpose` check:

- `value` stands alone. It is refused beside any other answer.
- An answer is declared only where a reader exists (decision 2). Today that is an activated row, a granted activated row and a mana ability. A spell, a mode, an alternative cost or a triggered row with `Answers` is refused until a PR adds that slot's reader, so a declaration nothing reads fails at boot, as ADR 0126 §6 already requires for its fields.
- A row with a target clause and no untargeted mode does not declare one. Its move carries `has_targets`, which already stops you. The guard refuses it, so the sweep cannot waste effort there.
- A sorcery-speed row does not declare one either. It is never offered while a stack item is waiting.

**How it relates to the rest of `Purpose`.** `Answers` says **what kind** of answer, and the amounts say **how much**. They are declared side by side and say different things. A pump row declares `AnswerPump` and, where the bot prices it, its `Pump` amounts. A regeneration row has no amount, so it declares only `AnswerProtect`. Nothing derives one from the other: `Pump != nil` without `AnswerPump` is a declaration still to do, and decision 3's ratchet lists it. The one coupling is in the reader: decision 2's fallback keeps reading `Pump`, `PreventCombatDamageToSelf`, `DamageToCreature` and `Sweep` for an undeclared row, as `purposeInteracts` does today.

**Helpers declare for their rows.** A row built by a shared helper is declared in the helper, once. `cards/effects` has 46 functions that return an `ActivatedAbility`: monstrosity is `pump`, cycling, typecycling and basic landcycling are `value`, boast's rows are reviewed by effect. A crew row (`AbilityCost.Crew` set) is `animate`, and the guard requires it there. The card files that call a helper change nothing.

### 2. Where it lives, and who reads it

| Slot | Go field | Wire | Reader |
|---|---|---|---|
| Activated row (own) | `effects.ActivatedAbility.Purpose.Answers` → `game.ActivatedAbilityShape.Purpose.Answers` | `activated_abilities[].purpose.answers` | `legal.abilityInteracts`, the bot |
| Granted activated row (ADR 0093 bundle) | `AbilityGrant.Activated[].Purpose.Answers` | the same, on the granted row | the same |
| Token activated row | `token_catalog` `Activated[].Purpose.Answers` | the same | the same |
| Mana ability | `effects.ManaAbility.Answers` (only `sac_outlet` or `value`) | not on the wire | `legal.manaAbilityInteracts` |
| Spell, mode, alternative cost | `Spec.Purpose`, `ModeWithPurpose`, `CostWithPurpose` | `purpose.answers` | none yet: refused until [Q5](#questions-for-the-owner-answered-2026-10-09)'s follow-up |
| Triggered row | `TriggerWithPurpose` | `ability_rows[].purpose.answers` | none yet: refused until a bot PR reads an opponent's trigger on the stack |

An activated ability carried on a card instance (`Card.ActivatedAbilities`, built at run time by an effect) is not catalog data and has no declaration. It keeps the fallback. Decision 3 counts catalog rows only.

### 3. How `interacts` and `combat_interacts` read it

`legal/interacts.go` becomes one function with a declared path and a fallback path:

```go
// answersOf is what an untargeted activation can do in response.
func answersOf(ab game.ActivatedAbilityShape) (game.Answers, bool /*declared*/) {
	if a := ab.Purpose.Answers; a != 0 {
		return a, true
	}
	return fallbackAnswers(ab), false // today's cost, purpose and text reads, mapped onto the vocabulary
}

Interacts:       !hasTargets(targets) && answers.HasTier(game.TierStack),
CombatInteracts: !hasTargets(targets) && answers.HasTier(game.TierCombat), // #2871
```

- **A declaration wins outright.** A declared row is never read as text. A row declared `value` with a creature-sacrifice cost does not interact. The guard makes that safe: a row whose cost sacrifices, exiles or returns another creature must include `sac_outlet` or `protect` in its declaration, or it is refused at boot. That keeps today's cost rule true for every declared row, without a text read.
- **The fallback is today's code, unchanged**, wrapped to return the vocabulary: the cost read gives `sac_outlet`, the purpose read gives `pump`/`prevent`/`remove`, and the text read's phrases map to `protect`, `pump`, `prevent`, `remove` and `restrict`. #2871's combat text read (keyword grants, extra blocks, "becomes a creature") joins it as the fallback for the combat tier. If #2871 lands first, its text read moves into `fallbackAnswers` in the signal PR. If this lands first, #2871 reads `answersOf` from the start and adds only its fallback phrases.
- **The move flags do not change meaning.** `interacts` is still "this untargeted activation can answer the stack". `combat_interacts` is #2871's "…in a combat window". The client's classes in ADR 0009 are unchanged. Only the input is better.
- **Mana abilities** read `ManaAbility.Answers` first, then today's `sacrificeInteracts` on the cost label.

**The self-sacrifice case** ([Q4](#questions-for-the-owner-answered-2026-10-09)). "Sacrifice this creature: …" is an answer only when the creature is about to be lost. The enumerator knows when that is without a declaration. If the source is a creature, and it is a target of an item on the stack that its controller does not control, or it is attacking or blocking in a combat window, the move gets `interacts` (or `combat_interacts`) whatever its declaration says. That is read from game state, not text, so it is no part of the fallback, and the ratchet does not count it.

### 4. The ratchet: the fallback only shrinks

`TestAnswersFallbackOnlyShrinks`, in `internal/cards/effects` (it needs the registered catalog), modelled on `TestLegacyTriggerBuildsOnlyShrink`:

- **Scope:** every catalog activated row (own, granted bundle, token) that is not sorcery speed and has no target clause or an untargeted mode, and every mana ability with a sacrifice cost. Those are the rows the enumerator asks `answersOf` about.
- **The list:** `internal/cards/effects/testdata/answers_fallback.txt`, one line per undeclared row in scope: `<oracle_id>\t<card name>\t<ref>\t<label>`, where `<ref>` is ADR 0093's `own:<i>` / `grant:<bundle>:<i>:<n>` (or `token:<slug>:<i>`, `mana:<i>`).
- **It fails** on:
  - a row in scope that is neither declared nor listed. A new card's row must declare, so the fallback never grows;
  - a listed row that is now declared or gone. Delete the line, or run `-update-answers-fallback`, which only ever removes lines;
  - a line listed twice;
  - a list longer than `answersFallbackCeiling`, a constant in the test that each sweep PR lowers to the new count and nothing raises.
- **Bootstrap:** the signal PR writes the list whole, once, with the ceiling at its length. The #2853 audit counted 706 such rows; the PR records the measured number.

A second, non-failing file is the **review record**: `testdata/answers_disagreements.txt`, rewritten by `-update-answers-disagreements`. For every declared row whose declared tier differs from what `fallbackAnswers` would have said, it lists the row, the declaration and the fallback's verdict. A sweep PR regenerates it, so its diff shows exactly which verdicts the batch changed: a pointless stop removed, or a missed window closed. Reviewers read that diff, not 40 card files. A test checks it is current (the regenerate-and-compare shape of `TestLegacyTriggerBuildsOnlyShrink`), so a stale record fails CI.

`docs/adding-cards.md` gets a subsection under "Declaring what a card does": a row in scope declares `Answers`, the table above, and the guard's rules. A new card's PR that forgets fails the ratchet with a message that names the row and the doc section.

### 5. The sweep

**Order.** Helpers first, then the rows that interact (a misread there is a missed window), then the value rows, then the combat rows once #2871's reader exists.

| Phase | What | Rows (from the #2853 audit) | Batch size | PRs |
|---|---|---:|---:|---:|
| S0 | **Helpers and tokens.** The 46 `ActivatedAbility` helpers in `cards/effects` (cycling and its variants, monstrosity, boast, …), every crew row, and the token catalog's rows (Clue, Food, Treasure, Blood, Map, Powerstone). | measured in the signal PR | one PR | 1 |
| S1 | **The 245 interacting rows**, by answer kind, so a reviewer checks one kind at a time: pump and counters (about 85), prevent and redirect (about 80), protect (regenerate, indestructible, hexproof, phasing, blink, self-bounce; about 45), sacrifice outlets (about 35), remove and restrict (about 15). Some rows have two kinds; each goes in the batch of its first. | 245 less S0's | ≤ 40 rows | 6–7 |
| S2 | **The 461 value rows**, by card-file name, alphabetical, each declared `value` or, where the batch finds one, an answer the text read missed. | 461 less S0's | ≤ 80 rows | 6 |
| S3 | **Combat rows** once #2871's reader is in: the audit's 26 "no" rows that grant a combat keyword, an extra block or a body, and any that S2 declared `value` but that #2871 would count. | ~30 | one PR | 1 |
| S4 | **The fallback at zero** ([Q3](#questions-for-the-owner-answered-2026-10-09)): `answers_fallback.txt` is empty, the ceiling is 0, and the text read is retired as Q3 decides. | — | one PR | 1 |

**Sonnet-sized batches.** Every S1 and S2 batch is `tier:1-mechanical`: one answer kind (or one alphabetical slice), a recipe in `docs/adding-cards.md`, a sibling to copy, and the ratchet and disagreement record as its acceptance test. A batch touches at most about 60 card files and changes no behaviour except the verdicts the disagreement diff shows. The signal PR is `tier:3-design`, S0 is `tier:2-standard`, and the bot PR (decision 6) is `tier:3-design`. The batches touch only `Purpose` declarations, so they change no oracle fixture and no `Completeness`. Each still runs the real-dump audits (`go test ./internal/decks/ -run RealDump` with `CMDCTRL_SCRYFALL_DUMP`), as every catalog PR does.

**What a batch PR does when the declaration and the text read disagree.** It declares what the card does, never what the text read said, and lists each flip in the PR body under "Verdicts changed", from the disagreement diff. [Q6](#questions-for-the-owner-answered-2026-10-09) decides who reads that list before merge.

### 6. What the bot gains

The bot reads `activated_abilities[].purpose.answers` on the wire, like every purpose field. That stays inside ADR 0033 §3: no import of `internal/game`, no oracle text. One bot PR, behind a `Config` switch `PriceAnswers` that is on in `DefaultConfig()` and off in `BaselineConfig()` (ADR 0126 §9), measured with ADR 0052's report block.

- **Saving a creature in response.** In a response window, when an opponent's stack item targets one of the bot's creatures (or a declared sweep would remove it), an untargeted activation whose source is that creature is priced by what it saves:
  - `protect`: the creature's value × `RemovalConfidence`;
  - `pump`: the same, when the declared `Pump` toughness covers the item's declared damage (ADR 0126's per-target `damage`);
  - `prevent`: the same, against declared damage.

  Today all of these are the flat 0.50 under `InstantThreshold`, so the bot lets the creature die.
- **Sacrifice outlets the cost does not name.** ADR 0126 §7's "dying anyway" price applies to any `sac_outlet` row, not only those with a `sacrifice_cost` on the view.
- **Layer A can absorb more.** A response window where every non-pass move is declared `value` is a pass, with no model call. The model tiers spend their budget on windows that have an answer.
- **The prompt names the answer.** `boardtext` and the model prompt label a move "(answers: protect)", so the model tiers and the MCP seat ([ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md)) see what a human sees in the stop.

**Not covered:** pricing a combat grant (flying on a blocker) is the combat planner's business, and `CombatValue` stays body-only (ADR 0126 owner decision 3).

**Suite positions** are proposals for the owner to review in the bot PR, as every position is:

- `regenerate-the-targeted-troll`: an opponent's Murder targets the bot's Albino Troll; `{1}{G}` is open. Accept: regenerate. Reject: pass.
- `do-not-regenerate-into-nothing`: an opponent's Divination is on the stack. Accept: pass. Reject: regenerate.
- `sacrifice-the-elder-it-targets`: an opponent's Swords to Plowshares targets the bot's Sakura-Tribe Elder. Accept: sacrifice it for a land. Reject: pass.

### 7. Snapshot and wire impact

- **No snapshot change.** `Purpose` is catalog data projected into the view, not game state (ADR 0126 §6). Nothing new is captured. `SnapshotSchemaVersion` stays 7 and `testdata/snapshot_shape/v7.txt` does not change. `Answers` holds no func or interface, so `closure_fields.txt` does not change.
- **Wire, additive.** `PurposeView` gains `answers`: a list of the wire strings above in the table's order, omitted when not declared. `["value"]` is sent, so a reader can tell "declared value" from "undeclared". It rides wherever `purpose` already does on an activated row, and is projected and cleared with it. `docs/protocol.md` documents it under "Declared purpose", and `client/src/lib/protocol.ts` mirrors the type. Clients may ignore it.
- **No move change.** `interacts` keeps its name and meaning. `combat_interacts` is #2871's.
- **One trap, closed in the signal PR.** The heuristic treats a non-nil `purpose` as "purpose-priced" in places (`aiseat/heuristic/purpose.go`, the mode and alternative-cost lookups), and then drops the mana-value proxy. A row or a spell whose purpose holds only `answers` must not be priced as if it declared amounts that sum to zero. The signal PR gives `PurposeView` a `Priced()` method (any amount, sweep, payoff or target entry is set) and uses it at every such check. As ADR 0126 PR 6 did, it shows §8's run 1 and run 2 identical to `develop` apart from IDs and timings.

### 8. Delivery

| # | PR | Depends on | Acceptance |
|---|---|---|---|
| 1 | This ADR | — | docsguard; the owner's answers recorded here (done, 2026-10-09) |
| 2 | **The signal.** `game.Answers`, its constants and tiers, `Purpose.Answers`, `ManaAbility.Answers`, the guard (decision 1's rules and decision 3's cost coupling), `PurposeView.answers` and `Priced()`, `docs/protocol.md`, `protocol.ts`. `answersOf` and `fallbackAnswers` in `legal`. The fallback is today's code, so no verdict changes. The self-sacrifice rule if Q4 is (a). The ratchet, bootstrapped, and the empty disagreement record. `docs/adding-cards.md`. Pointer lines in ADR 0009 and ADR 0126. | 1 | Every `interacts` verdict on the catalog is unchanged (a test compares `answersOf` with the old `abilityInteracts` over every row in scope, then is deleted in S4). `interacts_internal_test.go` passes unchanged. Bot run 1 and run 2 identical. Branch E2E, because it touches the legal-move flags auto-pass reads. |
| 3 | **S0, helpers.** | 2 | The ratchet's list shrinks; the disagreement record shows every flip. |
| 4–10 | **S1 batches**, ≤ 40 rows each. | 2 (3 for helper-built rows) | As 3, plus the real-dump audits. |
| 11–16 | **S2 batches**, ≤ 80 rows each. Can run in parallel with S1 batches on different files. | 2 | As 3. |
| 17 | **The bot** (decision 6), behind `PriceAnswers`. | 2, and the S1 batches for the curated decks' cards | ADR 0052's report block, ADR 0126 §8's sub-PR bar, the three positions reviewed. |
| 18 | **S3, combat rows.** | #2871's reader, 2 | As 3. |
| 19 | **S4, the fallback at zero.** | all batches | `answers_fallback.txt` empty, ceiling 0; the text read retired per Q3. Branch E2E. |

The bot PR can land any time after PR 2. It is placed after S1 because what it prices are the rows S1 declares, and the curated decks' rows can be pulled into the first S1 batch to unblock it.

---

## Consequences

- Auto-pass stops for what a card does, as declared by whoever added the card, not for what twenty phrases guess. A misread is fixed on one card, in one line, and does not move the rule for every other card.
- Every verdict the sweep changes is in a reviewable diff, with no hunting through card files.
- New cards cost one more declaration each. The ratchet makes forgetting it a CI failure, not a silent fallback.
- The bot can answer removal with an ability for the first time, and Layer A can skip windows where nothing answers.
- About 15 sweep PRs of catalog churn. They are mechanical and parallel, and each one is small.

## Rejected alternatives

- **Keep growing the text read.** Every fix narrows or widens a rule for the whole catalog. That is the #2853 experience, and it cannot reach "no missed windows" without a parser.
- **Derive answers from the effect closure** (run it on a scratch game and diff the board). ADR 0033 §3 and `moves.go` rule out cloning a game for a policy, and the enumerator would pay that cost on every priority window.
- **Infer from existing `Purpose` amounts alone.** Most interacting rows (regenerate, hexproof, a fog, a lock) have no amount to declare, so the field would still be empty for them.
- **A single boolean `Interacts` on the row.** It fixes auto-pass and leaves the bot and #2871 nothing: neither can tell a regeneration shield from a combat grant. It is [Q1](#questions-for-the-owner-answered-2026-10-09)'s option (b).
- **Making every undeclared row interact now.** Conservative, but it brings back the pre-#2853 stop on every Mind Stone until the sweep finishes. The fallback is today's behaviour, so nothing gets worse while the sweep runs.

---

## Questions for the owner (answered 2026-10-09)

Each question lists the recommended option first. The owner chose (a), the recommended option, on all seven. No section above changed. The answers are recorded under [Owner answers](#owner-answers-2026-10-09) below and bind the delivery PRs.

1. **Q1. The vocabulary.**
   - **(a) Recommended:** the ten values in decision 1: six stack answers (`protect`, `pump`, `prevent`, `remove`, `sac_outlet`, `restrict`), three combat answers (`combat_grant`, `animate`, `makes_blocker`), and `value`. Enough for auto-pass's two tiers and for the bot to price a save, with a short table a card author can apply.
   - **(b)** One declared boolean per tier (`interacts`, `combat_interacts`). Smaller to declare, and it fixes auto-pass. The bot learns nothing it can price.
   - **(c)** A finer list, one value per mechanic (regenerate, indestructible, hexproof, protection, phasing, blink, …). More precise for the bot, but slower to sweep and easier to get wrong, and auto-pass gains nothing over (a).
2. **Q2. First strike, double strike and deathtouch grants.**
   - **(a) Recommended:** combat tier only (`combat_grant`). They change nothing about a spell on the stack outside combat. Today's text read counts them everywhere, so a creature with "{B}: gains deathtouch" stops you on every opponent spell.
   - **(b)** Keep them in the stack tier, as today.
3. **Q3. When the fallback reaches zero.**
   - **(a) Recommended:** delete the text read. An undeclared row (only a run-time row carried on a card instance can be one) counts as interacting, following ADR 0009 §3's "a false-positive stop over a false-negative skip". The cost read stays only for mana abilities, which have no other source.
   - **(b)** Keep the text read forever as the fallback for run-time rows.
   - **(c)** Delete it, and treat an undeclared row as value. Fewer stops, and a missed window for any run-time row that answers.
4. **Q4. "Sacrifice this creature: …" when the creature is threatened.**
   - **(a) Recommended:** the move interacts whenever its source creature is a target of an item on the stack its controller does not control, and combat-interacts while the creature is attacking or blocking. This is read from game state, not declared. It closes the Sakura-Tribe Elder missed window without stopping you for every spell.
   - **(b)** Never, as today. A self-sacrifice row is whatever its effect declares.
   - **(c)** Always: every self-sacrifice creature row is `sac_outlet`. Simple, but a Sakura-Tribe Elder on the battlefield then stops you on every opponent spell.
5. **Q5. Spells.**
   - **(a) Recommended:** this ADR defines the field for spells, modes and alternative costs, but refuses it there until a follow-up decides how a castable instant declared `value` (Opt, Brainstorm) should be classed. Today any castable instant stops you, and changing that is its own ADR 0009 decision.
   - **(b)** Include it now: an instant declared `value` is classed like an untargeted value ability (off by default), and every instant and flash card in the catalog joins the sweep. Fewer stops sooner, and roughly double the sweep.
6. **Q6. Who reads a batch's changed verdicts before merge.**
   - **(a) Recommended:** each sweep PR lists every flip under "Verdicts changed" (from the disagreement diff) and merges on green under the standing develop rule. The owner reads the lists afterwards and reopens any flip with a one-line issue.
   - **(b)** The owner reviews every batch's flips before it merges, as suite labels are reviewed. Safer, and it puts about 15 PRs in the owner's queue.
7. **Q7. How far the sweep goes.**
   - **(a) Recommended:** to zero: the 245 interacting rows, then all 461 value rows, then the combat rows. The issue's "done" is the ratchet at zero, and an undeclared value row is a misread waiting to happen.
   - **(b)** The 245 interacting rows and the combat rows only. Value rows stay on the fallback, which already says "no" for them, and the ratchet stops at about 460.

### Owner answers (2026-10-09)

1. **The ten values.** `Purpose.Answers` uses decision 1's vocabulary: six stack answers, three combat answers and `value` (Q1 (a)).
2. **Combat only.** Granting first strike, double strike or deathtouch is `combat_grant` and counts only in combat windows (Q2 (a)).
3. **Delete the text read at zero.** When the fallback reaches zero, S4 deletes the text read, and an undeclared row counts as interacting (ADR 0009 §3). The cost read stays for mana abilities only (Q3 (a)).
4. **Self-sacrifice when threatened.** A "Sacrifice this creature: …" move stops you when its creature is a target of an item on the stack that its controller does not control, or is attacking or blocking in a combat window. This is read from game state, not declared, and it lands in the signal PR (Q4 (a)).
5. **Spells later.** The field is refused on spells, modes and alternative costs. How a castable instant declared `value` is classed is a separate, later decision (Q5 (a)).
6. **Merge on green.** Each sweep PR lists its flips under "Verdicts changed", taken from the disagreement diff, and merges on green under the standing develop rule. The owner reads the lists afterwards (Q6 (a)).
7. **Sweep to zero.** The sweep covers the 245 interacting rows, then all 461 value rows, then the combat rows, until the ratchet reaches zero (Q7 (a)).

### Sweep rulings (owner, 2026-10-09)

Four rulings from the first S1 batches, for the rows the vocabulary table does not settle. They bind the remaining batches and any new card.

1. **Shrink effects are `remove`.** A row that gives a creature -X/-X or otherwise shrinks it (Flailing Manticore, Flailing Ogre, Flailing Soldier, Oona's Prowler) declares `remove`.
2. **Trades are `pump`.** A row that raises one of power and toughness while lowering the other (+2/-2, +1/-1, "+1/-1 or -1/+1": Multiform Wonder, Shipwreck Moray, Unliving Psychopath, Undulating Witness, Endling) declares `pump`.
3. **Turning off an opponent's protection is `restrict`.** A row that removes hexproof, indestructible or regeneration from what an opponent controls, or stops them regenerating (Arcane Lighthouse, Detection Tower, Shadowspear, Knight and Clergy of the Holy Nimbus), declares `restrict`.
4. **Modal, conditional and mixed rows declare the union.** A row whose modes or conditions can give different answers declares every answer any of them can give (Lost Jitte, Storm Elemental, Repeat Offender, Aerid Konstrari, Peema Trailblazer, Thran Weaponry, Gallia, and the mixed `protect|pump` and `protect|sac_outlet` rows).

Applied by analogy (owner, 2026-10-09): ruling 3 also covers a row that turns off its own controller's protection, so Glittering Lion and Glittering Lynx ("loses \"Prevent all damage that would be dealt to this creature\"", any player may activate) declare `restrict`.
