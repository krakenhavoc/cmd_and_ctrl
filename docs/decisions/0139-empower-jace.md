# ADR 0139 — Empower Jace and the Jace planeswalker token

**Status:** Accepted · 2026-10-09 · Reality Fracture set (tracker [#2795](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2795)), outside docs/sprints.md at the owner's request
**Issues:** [#2796](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2796) (seam: Empower Jace and the Jace planeswalker token). Sibling: [#2797](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2797) (planeswalker statics), which some Empower Jace cards also need. Registry row: `empower-jace`.
**Owner decisions:** none needed. CR 701.71a fixes every behaviour below.
**Numbering:** 0139 was reserved for this ADR in the 2026-10-09 sweep (0138 was the highest on `origin/develop` and on every remote head), and recorded in #2796's claim comment.
**Builds on:** [ADR 0032](0032-planeswalkers.md) §8 (loyalty abilities are ordinary CR 602 activated abilities; a loyalty cost is not an effect), [ADR 0083](0083-token-abilities.md) (a token template's abilities live in the catalog under a token key), [ADR 0100](0100-delve-either-or-and-variable-sacrifice-costs.md) (behold as a branch of an either/or cost), [ADR 0063](0063-durations-and-control.md) (one duration model), [ADR 0026](0026-delayed-triggers.md)'s 2026-09-18 amendment (event-conditioned delayed triggers), [ADR 0033](0033-ai-bot-seat.md) §1 (the enumerator agrees with the engine).

The seam lands in one PR: the token, the keyword action, one engine fix it exposed, the enumerator, view, bot and client tests, and fifteen cards. The other twenty Empower Jace cards are listed on the registry row, with the ones that wait on #2797 named.

---

## Context

Reality Fracture (FRA, released 2026-10-02) prints Empower Jace on 35 cards in the 2026-10-05 Scryfall dump. None could be catalogued. The engine had loyalty abilities (`LoyaltyCost`, ADR 0032 §8) and the CR 704.5i state-based action, but no planeswalker token, and nothing that could "put N loyalty counters on a Jace token you control, creating it first if you have none".

### The rules

Read in the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026):

- **CR 701.71a:** "To empower Jace N means 'If you don't control a Jace planeswalker token, create a blue Jace planeswalker token with 0 loyalty, "[−1]: Surveil 1," and "[−3]: Draw a card." Choose a Jace planeswalker token you control. Put N loyalty counters on it.'"
- **CR 701.4a:** "'Behold a [quality]' means 'Reveal a [quality] card from your hand or choose a [quality] permanent you control on the battlefield.'" Countersculpt and Theorist's Sanctum behold a Jace.
- **CR 704.5i:** a planeswalker with loyalty 0 is put into its owner's graveyard. **CR 704.3:** state-based actions are checked only when a player would receive priority, never while a spell or ability is resolving.
- **CR 306.5b:** a planeswalker enters with loyalty counters equal to its printed loyalty. The token's is 0.

The printed token (Scryfall set `tfra`, oracle `a6a09b69-dd86-4fda-8971-b77a2a4abddb`) is named **Jace**, a non-legendary "Token Planeswalker — Jace", blue, loyalty 0.

Three readings follow from the text and are what the code does:

1. "A Jace planeswalker token" is any token that is a planeswalker with the Jace subtype: the one this action makes, a token copy of a Jace planeswalker card, or anything an effect turned into one. A Jace planeswalker CARD is not one, so Jace, Reality Sculptor's +1 never puts counters on himself.
2. The choice comes after the creation. If a replacement makes the creation produce two tokens (Doubling Season, CR 701.7b), the controller chooses one of them, and the other has 0 loyalty and goes at the next check.
3. The counters are put by an effect, so CR 614 counter replacements apply to them. Doubling Season both doubles the token and doubles the counters.

## Decision

### 1. The token is a catalog token template with loyalty abilities (`effects/empower_jace.go`)

`printedJaceToken` is an ADR 0083 template, slug `jace`, registered in `token_catalog.go`'s list:

- `Card`: name "Jace", type line "Token Planeswalker — Jace", colours `[U]`, `StartingLoyalty` 0.
- `Activated`: two rows, "−1: Surveil 1." (`LoyaltyCost(-1)`, the `Surveil` primitive) and "−3: Draw a card." (`LoyaltyCost(-3)`, `DrawCards`, `Purpose{Draws: 1}` for the heuristic).
- `Text`: the printed ability text, which reaches the board as `CardView.token_text`.

`effects.JaceToken()` hands out the template with the token key stamped on, as every other catalog token constructor does.

Nothing in `internal/game` needed to change for the token. `game.CatalogKey` answers a token with its key, so `ActivatedAbilitiesForCard`, `ActivateCatalogAbility` (the CR 606.2, 606.3 and 606.6 gates of ADR 0032 §8), the view's `activated_abilities` rows with `loyalty_cost`, `loyalty_activated`, the legal-move enumerator and the bots all reach the token's loyalty abilities through the route a printed planeswalker's take. This was the reason to use a template rather than a new token kind: there is one loyalty-ability path, and the token rides it. `TestRealDumpEveryTokenTemplateMatchesAPrintedToken` matches the template against the printed `tfra` token.

A printed loyalty of 0 means the CR 306.5b stamp (`stampStartingLoyaltyLocked`) puts nothing on the token. That is correct: the counters come from the instruction that made it (§2).

### 2. The keyword action is one primitive, `EmpowerJace{N | Count, Then}`

```go
EmpowerJace{N: 2}.Apply(ctx)                           // Empower Jace 2
EmpowerJace{Count: islandsYouControl}.Apply(ctx)       // Empower Jace X, X counted as it resolves
EmpowerJace{N: 6, Then: drawACard}.Apply(ctx)          // "Empower Jace 6. Draw a card."
```

`Apply` runs CR 701.71a's three sentences as one instruction for the resolving item's controller:

1. `JaceTokensControlledBy(g, player)` is the one walk for "a Jace planeswalker token you control": `IsJacePlaneswalkerToken` reads `IsToken`, `IsPlaneswalker` and `HasSubtype("Jace")` off the live object.
2. With none, it creates `JaceToken()` through `CreateTokensThenForEffect`, the engine's one creation path, so the CR 701.7b window applies, and continues from the creation's continuation (a creation can pause on a CR 616 ordering prompt).
3. It then walks again. One candidate gets the counters with no question. Two or more ask the controller with the existing own-permanents pick (`ChoosePermanents`, `PendingChoiceOwnPermanents`, exactly one). None (the creation was replaced away) places nothing.
4. The counters go on through `AddCounterByThenForEffect` with the controller as the placer (CR 701.71a: "you" put them). This is an effect, so the CR 614 pipeline applies, unlike paying a loyalty cost (ADR 0032 §8).
5. `Then` runs once the counters are on, whether or not any were placed, so "Empower Jace 6. Draw a card." draws after the choice is answered.

`Count` is evaluated once, as the action begins (CR 608.2h). A negative count is zero. "Empower Jace 0" still creates the token, which then has 0 loyalty and goes at the next check. That is the rules, not a special case (Jace, Reality Sculptor's +1 with no Islands).

**Why the 0-loyalty token is safe.** The token exists with no counters only inside the instruction, and state-based actions are not checked inside a resolution (CR 704.3). If the instruction pauses on its own prompt, #1289's resolution hold (`holdForOpenResolutionLocked`) keeps the state-based actions and the trigger drain waiting until the answer. So a Jace that is about to be chosen is never swept first. The Doubling Season test pins this: two 0-loyalty Jaces, a prompt, the chosen one ends at 12, and the other is gone only after the answer.

### 3. "Loyalty counters among Jaces you control" is `JaceLoyaltyAmong`

Jace, Reality Sculptor's 0 is "Activate only if there are twenty-five or more loyalty counters among Jaces you control". `JaceLoyaltyAmong(g, player)` sums loyalty counters on every permanent the player controls with the Jace subtype, tokens and cards alike. The card uses it as the ability's CR 602.1b `Condition`, so the row is greyed (`condition_unmet`), refused by the engine, and never offered to a bot until it holds.

### 4. Delayed triggers with a duration are also swept as a turn begins (`game/rotation.go`)

Jace, Reality Sculptor's −3 is "Until your next turn, whenever a creature attacks you or a planeswalker you control, it gets -5/-0 until end of turn": a CR 603.7b delayed trigger that repeats for a stated duration. The engine already had the parts (`DelayedTrigger.On`, `Condition`, `Repeats`, `Duration`), and the card schedules one with `DurationUntilYourNextTurn`.

The test found a gap. `clearExpiredDelayedTriggersLocked` ran only in the cleanup sweep, so an "until your next turn" trigger stayed queued, and visible on the wire's `delayed_triggers`, through the whole of its controller's next turn. The turn-began hook in `rotation.go` already ends "until your next turn" scoped statics, cast permissions and goads at that boundary (ADR 0063 Decision 2). It now also calls `clearExpiredDelayedTriggersLocked(false)`. `durationExpiredLocked` is still the one place that decides when a duration is over, and it was written for both call sites. No existing delayed trigger had an "until your next turn" duration, so nothing else changes.

### 5. The wire, the client and the bots

Nothing on the wire is new. The token reaches the board as an ordinary `CardView`: `counters.loyalty`, `activated_abilities` with `loyalty_cost` and `ref`, `loyalty_activated`, `token_text`. The client's click rule and popover treat it exactly like a printed planeswalker: a click activates the one payable ability, the popover lists both, and the manual loyalty rows an uncatalogued planeswalker gets are not added on top (`client/src/lib/jaceToken.test.ts`).

The "which Jace?" question is the existing `own_permanents` prompt, which the enumerator answers (`choiceMoves`), the bots answer, and the choice gate and the leave-game table already classify. If the answer names no Jace (every candidate left before it came), the instruction finishes with nothing placed.

### 6. Card-side vocabulary

- `effects.JaceToken()`, `IsJacePlaneswalkerToken`, `JaceTokensControlledBy`, `JaceLoyaltyAmong`, `JaceTokenSubtype` (`empower_jace.go`).
- `EmpowerJace{N, Count, Then}`.
- `YouCastYourFirstNoncreatureSpellThisTurn` (`triggers_common.go`), the "you" twin of Esper Sentinel's opponent predicate, for Plan for All Outcomes.
- "Behold a Jace or pay {N}" is `BeholdOrPay("a", "Jace", "{N}")` (ADR 0100), unchanged: `RevealCost.matches` reads `HasSubtype("Jace")`, so a Jace token is a legal behold.

The recipe is in [docs/adding-cards.md](../adding-cards.md#empower-jace-and-the-jace-token-adr-0139-2796-cr-70171).

## Proof cards

Fifteen cards, all `CompletenessFull`:

| Card | What it proves |
| --- | --- |
| No Admittance | Empower Jace after a targeted effect, on a sorcery |
| Countersculpt | beholding a Jace token (ADR 0100) and empowering the same token |
| Rewrite Regrets | after a reanimation from your graveyard |
| Campus Crier | from an ability that functions from the graveyard |
| Solve for Disappointment | as the continuation of an ADR 0116 revealed-hand pick |
| Plan for All Outcomes | from a "first noncreature spell each turn" trigger; its enters trigger is the owner's top-or-bottom choice |
| Mindseeker Oculus, Keeper of the Quiet Hour, Arcane Amphisbaena | from an enters trigger on the stack |
| Repurposed Enforcer | "Empower Jace X" counted as an attack trigger resolves |
| Protege's Awakening | `Then`: the draw waits for the choice |
| Academic Ascent, Tam's Resistance | after a pump; with "up to one" target and none chosen |
| Vraska's Final Mercy | on a modal bullet |
| Jace, Reality Sculptor | all three of his abilities (§2, §3, §4) |

## Tests

- `server/internal/cards/effects/empower_jace_test.go`: the token's characteristics and loyalty activations (CR 606.3, 606.6), creation with N counters, adding to an existing token, ignoring an opponent's token, the choice with two tokens, Doubling Season, Empower Jace 0, and one test per proof card (including the −3's duration ending as its controller's next turn begins).
- `server/internal/protocol/jace_token_view_test.go`: loyalty, both abilities with their `loyalty_cost`, and `token_text` reach the wire for the token.
- `server/internal/legal/jace_token_test.go`: the token's affordable loyalty abilities are offered and dispatched (#544), and the "which Jace?" prompt is answered.
- `server/internal/aiseat/jace_token_test.go`: the real heuristic answers the prompt and plays two turns past a table holding two Jace tokens.
- `client/src/lib/jaceToken.test.ts`: the board's click and popover rules on the token.

### Back-outs

Each line was reverted, the named test watched fail, and the line restored.

| Reverted | Failing test |
| --- | --- |
| `EmpowerJace.Apply` always creates a token (the "if you don't control one" check removed) | `TestEmpowerJaceAddsToTheJaceTokenYouControl`, `TestEmpowerJaceAsksWhichJaceWhenThereAreSeveral` |
| two or more Jaces: the first one gets the counters, no question | `TestEmpowerJaceAsksWhichJaceWhenThereAreSeveral`, `TestEmpowerJaceUnderDoublingSeasonPicksOneOfTheTwoJaces`, `legal.TestEmpowerJaceChoiceIsOfferedAndAccepted` |
| `printedJaceToken` left out of `tokenTemplates` | `TestJaceTokenIsABlueJacePlaneswalkerTokenWithTwoLoyaltyAbilities`, `TestJaceTokenLoyaltyAbilitiesActivate`, `protocol.TestJaceTokenShipsItsLoyaltyAbilitiesAndText` |
| `clearExpiredDelayedTriggersLocked(false)` removed from the turn-began hook | `TestJaceRealitySculptorMinusThreeShrinksAttackersUntilYourNextTurn` |
| `IsJacePlaneswalkerToken` without `IsToken()` | `TestJaceRealitySculptorPlusOneEmpowersATokenByIslands` (the counters went on Jace himself) |
| Jace, Reality Sculptor's 0 without its 25-counter condition | `TestJaceRealitySculptorZeroNeedsTwentyFiveAmongJaces` |

## Consequences

- Any future planeswalker token is a template with `LoyaltyCost` rows and needs no engine work.
- A delayed trigger whose duration ends as a turn begins now ends there, not at the following cleanup.

## Out of scope

- **Granted loyalty abilities, surviving 0 loyalty, loyalty-counter triggers, Jace's instant-speed activations:** #2797. Avatar of Burgeoning Echoes, Sanctum Lurker, Inspired Tethermage, Jace's Machinations and the ten "Way of the …" cards wait on it and are on the `empower-jace` row's Waiting list.
- **The remaining unblocked cards** (Fatehold Charm, Hexhaven Battalion, Overwrite the Multiverse, Theorist's Proxy, Theorist's Sanctum, Violent Echoes) are card work on this seam for the tracker's phase 2 (#2795).
