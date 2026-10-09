# ADR 0132 — Day and night, daybound and nightbound

**Status:** Accepted · 2026-10-07 · S41 — Turn machinery and per-turn accounting (the issue has no milestone; this is the nearest tracker, [#884](https://github.com/krakenhavoc/cmd_and_ctrl/issues/884))
**Issues:** [#2561](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2561) (seam: day and night, found building The Celestus for [#2190](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2190)). Follow-up: [#2586](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2586) (the remaining 37 day and night cards).
**Owner decisions:** none needed. Every behaviour below is fixed by the Comprehensive Rules; no product question came up.
**Numbering:** 0131 (`0131-krrik-pay-life-for-black-mana.md`) is the highest number on any branch. Checked with `git log --all --name-only -- docs/decisions/013*` after `git fetch origin`, which covers every remote head: only 0130 and 0131 appear above 0129. No open issue's claim comment reserves 0132.
**Builds on:** [ADR 0079](0079-transforming-a-permanent.md) (the transform verb this one must not bypass, and its out-of-scope item 3, which calls day and night "a turn-machinery change, not a transform one"), [ADR 0084](0084-phasing.md) (phasing is the first action of the untap step and the day/night check the second), [ADR 0034](0034-multi-face-cards.md) (faces, `SetFace`, and per-face keywords from [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) §4), [ADR 0059](0059-turn-machinery.md) (the rotation seam), [ADR 0096](0096-the-monarch-from-a-card-effect.md) (a game-wide designation with an event, the shape this copies).

The seam work landed in one PR: rules, snapshot, wire, client chip and 13 proof cards. Nothing below is deferred to a later PR except the remaining cards (#2586).

---

## Context

The engine had no day/night designation. Nothing tracked whether it was day or night, "it becomes day" and "it becomes night" were not instructions, and "whenever day becomes night or night becomes day" had no event to watch. The Celestus needs all three, and so do 49 other Commander-legal cards (50 in all in the September 2026 Scryfall dump): 14 with an "if it's neither day nor night, it becomes day as this enters" clause, a day/night trigger or a day/night condition, and 35 daybound // nightbound double-faced cards. They were waiting on a row that did not exist (ADR 0079 listed it as an out-of-scope deferral).

### The rules

Every rule was read in the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026). **The issue and the task brief cite day and night as CR 726. In this edition 726 is the initiative; day and night is CR 731.** (ADR 0079 already cited 731, and ADR 0084 cites 502.2.) The text of the rules this ADR depends on:

- **CR 731.1:** "Day and night are designations that the game itself can have. The game starts with neither designation. 'It becomes day' and 'it becomes night' refer to the game gaining the day or night designation. It can become day or night through the daybound and nightbound keyword abilities (see rule 702.145). Other effects can also make it day or night. Once it has become day or night, the game will have exactly one of those designations from that point forward."
- **CR 731.1a:** "The phrases 'day becomes night' and 'night becomes day' refer to the game losing the first designation and gaining the second one."
- **CR 731.2 / 502.2 / 703.4b:** as the second part of the untap step, immediately after phasing, "if it's day and the previous turn's active player didn't cast any spells during that turn, it becomes night. If it's night, and previous turn's active player cast two or more spells during the previous turn, it becomes day." (731.2a, 731.2b.) "If it's neither day nor night, this check doesn't happen and it remains neither." (731.2c.) "This turn-based action doesn't use the stack."
- **CR 502.2a / 731.2a / 731.2b** also give a shared-team-turns variant. This engine has no team turns; see [Out of scope](#out-of-scope).
- **CR 702.145b:** daybound "represents three static abilities. 'Daybound' means 'If it is night and this permanent is represented by a double-faced card, it enters transformed,' 'As it becomes night, if this permanent is front face up, transform it,' and 'This permanent can't transform except due to its daybound ability.'"
- **CR 702.145c:** "Any time a player controls a permanent that is front face up with daybound and it's night, that player transforms that permanent. This happens immediately and isn't a state-based action."
- **CR 702.145d:** "Any time a player controls a permanent with daybound, if it's neither day nor night, it becomes day."
- **CR 702.145e / 702.145f:** nightbound is the mirror: "As it becomes day, if this permanent is back face up, transform it", "This permanent can't transform except due to its nightbound ability", and "Any time a player controls a permanent that is back face up with nightbound and it's day, that player transforms that permanent."
- **CR 702.145g:** "Any time a player controls a permanent with nightbound, if it's neither day nor night and there are no permanents with daybound on the battlefield, it becomes night."
- **CR 712.18** (ADR 0079): a transform leaves the same object. A permanent that enters transformed never transformed.
- **CR 712.9:** only a permanent represented by a double-faced card can transform. The CR 702.145b example is a Clone of Bird Admirer: "that permanent can't transform because it isn't represented by a double-faced card."

## Decision

### 1. One field on `Game`, one writer (`game/daynight.go`)

`Game.DayNight` is a `DayNightState{Designation, PrevTurnSpells, PrevTurnKnown}`. `Designation` is `""` (neither, the start of every game), `"day"` or `"night"`. `becomeDayNightLocked` is the only write. It returns at once when the designation would not change (a second "it becomes day" is nothing, CR 731.1), sets it, emits `EventDayNightChanged`, and settles the board (section 4).

The event carries the NEW designation in `Label` and `Amount` 1 when the game was already day or night and has flipped, 0 when it has just gained its first. CR 731.1a makes that distinction the rule: the first designation is a gain, not "night becomes day", so a Celestus or a Brimstone Vandal must not trigger on the game's first designation. `game.DayNightFlipped(ev)` is the predicate; cards never read the fields.

Reads: `IsDay`, `IsNight`, `DayNightDesignation` (caller holds `g.mu`, like `CastTallyFor`). Effects: `BecomeDayForEffect`, `BecomeNightForEffect`, `BecomeDayIfNeitherForEffect` ("if it's neither day nor night, it becomes day" must not turn night into day), `ToggleDayNightForEffect` (The Celestus: night to day, otherwise night).

### 2. The turn-based check reads a number the engine had already thrown away

CR 502.2 asks, at the start of a turn, what the PREVIOUS turn's active player cast. By then `SpellsCastThisTurn` has been reset by `onTurnBeganLocked`. `recordPrevTurnSpellsLocked` copies `CastTallyFor(activePlayer).Total` into `DayNight.PrevTurnSpells` from `beginNextTurnLocked`, beside `recordLastTurnAttacksLocked` and for the same reason, before the rotation replaces `g.Turn`. Copies are not cast (CR 707.10), so `CastTally.Total` is the number the rule counts. `PrevTurnKnown` is false until a turn has ended, so the first turn of the game has nothing to read.

`dayNightTurnCheckLocked` is the check. It runs from `performUntapStepLocked` straight after `performPhasingLocked` and before the untap set is built: CR 502.1, 502.2, 502.3 in that order, and the same placement ADR 0084 chose for phasing (everything ahead of the untap step's one pause runs exactly once). An extra turn is a turn; the player who took it is the "previous turn's active player" for the turn after it.

An eliminated player's tally stays in the map by player ID, so a turn they started and lost still counts the spells they cast.

### 3. Daybound and nightbound are canonical keywords

`daybound` and `nightbound` join `canonicalKeywords` in this change, the same rule every keyword follows ("a keyword joins the table in the same change that teaches the engine to honour it"). The deck importer stamps each from the face's own text (`printedKeywordsForFace`, ADR 0107 §4), so a deck-imported werewolf works with no catalog entry, and `Face.Keywords` gives the front face `daybound` and the back face `nightbound`. A catalog card declares them in `PrintedKeywords` like any other keyword. The roadmap registry gains a keyword item for the pair (`TestEveryCanonicalKeywordHasOneItem` fails without it).

### 4. `settleDayNightLocked` is the "any time" half of CR 702.145

One function, called at the moments the board and the designation can have moved against each other: after every designation change, after a permanent enters the battlefield, and after a phase-in. It does four things, in this order:

1. If neither: a permanent with daybound makes it day (CR 702.145d); failing that, a permanent with nightbound and no daybound permanent on the battlefield makes it night (702.145g).
2. If night: every front-face-up daybound permanent that `CanTransform` transforms (702.145c).
3. If day: every back-face-up nightbound permanent that `CanTransform` transforms (702.145f).

It is a plain loop, not a state-based action ("This happens immediately and isn't a state-based action"). `CanTransform` is the filter for "represented by a double-faced card" (CR 712.9), so a token copy of a werewolf keeps its daybound ability (it can start the clock) and never turns over.

### 5. A daybound permanent that enters at night enters transformed (`dayNightListener`)

CR 702.145b's first ability is an "enters transformed" replacement. A listener on `EventZoneMove` into the battlefield, registered right after `layerVersionBump` and ahead of the trigger harvester, turns an entering front-face-up daybound permanent over with `SetFace(1)` when it is night, with the layer cache dropped and no `EventTransform` (a permanent that enters transformed never transformed, CR 712.18, and ADR 0079 decision 5 says the same for exile-and-return). The move event is emitted after the card is on the battlefield and before its `EventETB`, so ETB triggers are harvested off the face that entered. Infestation Expert cast at night makes Infested Werewolf's two Insects, not Infestation Expert's one.

It is a listener and not a CR 614 pipeline entry because the pipeline exists to change an entering permanent before it arrives and ask questions about it, and this changes only which face is up. The consequence, recorded as a gap: a back face's own self-replacement (an "enters tapped" clause on a nightbound face) would not apply to a card that enters transformed. No nightbound face prints one.

### 6. Nothing else transforms a werewolf

CR 702.145b and 702.145e: "This permanent can't transform except due to its daybound / nightbound ability." `TransformPermanentForEffect` returns nil and does nothing for a permanent with either keyword, exactly as CR 701.27c treats an object that can't transform (no error, no `EventEffectError`). The day/night rules themselves go through `transformInPlaceLocked`, the unguarded body of the same verb (EventTransform, the layer-cache drop, the "as this transforms into" hook), so a werewolf's turn-over is the same event as any transform and a "whenever this transforms" trigger sees it.

`ExileAndReturnTransformedForEffect` is not guarded. A daybound permanent exiled and returned "transformed" is a new object that enters under CR 702.145b's replacement, which only matters at night; no card does it.

### 7. The snapshot

`GameSnapshot` gains an optional `dayNight` object, a `*DayNightState` that is nil (key absent) for a game that never had a designation and has not ended a turn. The key is absent from every frozen fixture, which re-encode byte for byte. This is additive within the current schema, for the reason OpeningRoll's key was: a file written before it decodes as neither day nor night with no previous turn, which is every game before this change; and a binary from before it, handed a file with the key, drops it. That binary does not know daybound, so a game it restores has werewolves on whichever face they were on and no designation, and the next binary to open it sets the designation from the board at the first entry or flip. `TestSnapshotShapeIsRecorded` records the four new paths (`dayNight`, `.designation`, `.prevTurnSpells`, `.prevTurnKnown`). `Clone` and `RestoreFrom` carry the struct by value, so undo restores the clock with the board.

### 8. The wire and the client

`GameView.day_night` is `"day"` or `"night"`, omitted while neither, public and identical for every viewer (`FilterViewFor` copies it). The log gains `day_night` (`EventDayNightChanged`, "It becomes night"), narrated because the untap-step check changes the designation with no spell or ability behind it, so without a line every werewolf would turn over with nothing saying why. The log kind gate (`TestEveryEventKindIsNarratedOrDeliberatelySilent`) requires either a narration or a written silence; this is a narration.

The client shows one chip in the action dock's header beside the turn line (`PhaseDisplay`, `dayNight.ts`): "☀ Day" or "☾ Night" with a title that says the rule in plain words, nothing while the game has neither (most tables never use a day/night card, so a "neither" chip would be noise). `protocol.ts` gains `day_night` and the `day_night` log kind, and `gameLog.ts` tones it like the turn-over it causes. The chip reads the designation off the same view every seat gets.

### 9. Card-side vocabulary (`effects/daynight.go`)

`BecomeDay{}`, `BecomeNight{}`, `ToggleDayNight{}` (primitives); `BecomesDayAsEnters()` (an `AsEnters` hook, off the stack, because the clause is an "as enters" one, CR 614.12); `WheneverDayBecomesNightOrNightBecomesDay(label, effect)` (an `On` over `EventDayNightChanged` with `DayNightFlipped`); `ItsNight(g)` / `ItsDay(g)` and `ItsNightCost()` (a `CostPredicate`, for Moonrager's Slash). `plusOneCounterOnChosenTargets` is the shared body Ivy Lane Denizen and Sunrise Cavalier were each spelling out.

### 10. Bots

The Celestus's toggle is an ordinary sorcery-speed activated ability, so the enumerator already offers it and the dispatcher already accepts it; this PR pins both (`legal/day_night_toggle_test.go`: offered in the sorcery window and accepted, withheld in the end step, withheld without the mana). It carries no declared purpose, so the heuristic prices it at the flat `ActivateBase`, the posture ADR 0126 §6 (owner decision 6) gives every activated ability with no purpose. `aiseat/day_night_game_test.go` plays a bot with a Celestus for six turns and requires that it toggles, that the flip's trigger resolves (the bot's life goes up) and that no move is rejected.

The trigger's optional loot is a `MayChoice` (a yes/no `PendingChoiceConfirm`) followed by a discard prompt, both of which the bot runner already answers.

What the bot does NOT do is reason about the toggle: it does not know that turning night to day turns its own werewolves back, or that flipping the designation helps an opponent's Brimstone Vandal. That is a pricing question for a future change (a declared `dayNight` purpose beside `draws` and `tokens` would carry it), not a rules one.

## Proof cards

Thirteen, all `CompletenessFull`:

| Card | What it proves |
|---|---|
| The Celestus | the as-enters clause, the sorcery-speed toggle, the flip trigger with a linked optional loot, firing off the untap-step check |
| Brimstone Vandal | the trigger fires on a flip and not on the first designation |
| Firmament Sage, Obsessive Astronomer | the same trigger with a draw and a choose-up-to-two discard-then-draw |
| Sunrise Cavalier | a targeted flip trigger |
| Moonrager's Slash | a cost reduction that reads the designation |
| Olivia's Midnight Ambush | a resolution-time "if it's night … instead" |
| Into the Night | "it becomes night" as a spell effect, with the discard-then-draw |
| Fearful Villager, Harvesttide Infiltrator, Bird Admirer | daybound // nightbound through the deck importer: both faces carry their own keywords, the permanent follows the designation, and an effect cannot transform it |
| Village Watch // Village Reavers | a back-face static (a tribal keyword grant) that appears when night turns it over |
| Infestation Expert // Infested Werewolf | a card cast at night enters on its back face and it is the back face's "enters" that triggers; turning over later triggers nothing |

The other 37 Commander-legal cards that use day or night are on the `day-and-night` roadmap row's Waiting list (#2586). Most are honest on today's engine and need only a slice builder.

## Tests

- `game/daynight_test.go`: starts neither; the first designation is not a flip; a repeat is no event; "if neither" leaves night alone; the toggle table; the untap-step check over eight cases on the real turn rotation (including "neither stays neither" and "only the active player's spells count"); the first turn has nothing to read; Clone/RestoreFrom; daybound entering at neither makes day; entering at night enters transformed with no `EventTransform`; werewolves follow the designation and stay the same object; an instruction cannot transform one; the full untap-step flow; a non-double-faced daybound permanent starts the clock and never turns over; nightbound alone makes night.
- `cards/effects/day_night_*_test.go`, `werewolf_cards_test.go`: one test per proof card, the importer path for the werewolves, a layer-cache test for a static that reads the designation (a test-only probe spec).
- `legal/day_night_toggle_test.go`, `aiseat/day_night_game_test.go`: section 10.
- `protocol`: the `day_night` field in the `FilterViewFor` field test and the protocol doc test; `client`: `phaseDisplay.dayNight.render.test.ts`.
- The snapshot shape file and the populated-game snapshot test carry the new key.

### Back-outs

Each change was reverted by itself and the named test watched fail, then restored:

| Reverted | Fails |
|---|---|
| the check call in `performUntapStepLocked` | `TestUntapStepCheckReadsThePreviousTurn`, `TestUntapStepTurnsTheWerewolvesOver` |
| `recordPrevTurnSpellsLocked` in `beginNextTurnLocked` | `TestUntapStepCheckReadsThePreviousTurn` |
| the transform guard | `TestOnlyTheDayNightRulesTransformAWerewolf` |
| the `dayNightListener` registration | `TestDayboundEnteringAtNeitherMakesItDay`, `TestDayboundEnteringAtNightEntersTransformed` |
| the entry flip's night test | `TestDayboundEnteringAtNightEntersTransformed` |
| `Amount` 1 only on a flip | `TestFirstDesignationIsNotAFlip`, `TestBrimstoneVandalPingsOnAFlipAndNotOnTheFirstDesignation` |
| the `EventDayNightChanged` arm in `layerVersionBump` | `TestADayNightStaticFollowsTheDesignation` |
| the "if neither" guard | `TestBecomeDayIfNeitherLeavesNightAlone` |
| the settle at the end of `becomeDayNightLocked` | `TestWerewolvesFollowTheDesignation` |
| `daybound` in the keyword table | `TestImportedWerewolfFacesCarryTheirOwnKeywords` |
| The Celestus's `SorcerySpeed` | `TestCelestusToggleIsSorcerySpeed`, `TestCelestusToggleIsWithheldOutsideTheSorceryWindow` |
| Moonrager's Slash's discount | `TestMoonragersSlashCostsTwoLessAtNight` |
| Olivia's Midnight Ambush's night branch | `TestOliviasMidnightAmbushShrinksByTwoByDayAndThirteenAtNight` |
| Village Reavers' `YoursOnly` | `TestVillageReaversGivesHasteToYourWolvesAndWerewolves` |

## Consequences

- A deck-imported werewolf now participates in day and night with no catalog entry. Before this change `Daybound` and `Nightbound` were dropped by the importer as unknown keywords and the card played as a plain double-faced creature that nothing could turn over; now an existing "transform" effect (Moonmist) does nothing to it, as the rules say.
- `TransformPermanentForEffect` gained a guard. It is the only behaviour change to an existing verb, and it is a no-op for every card that does not have `daybound` or `nightbound`.
- The first card on the Waiting list that needs more than a slice builder will say so in its own PR; none of the 37 is blocked on this seam.

## Out of scope

- **A keyword granted to a permanent that did not print it, and a copy of a werewolf.** CR 702.145c/f say "any time"; this engine settles at the moments the designation or the battlefield changes. A permanent that gains daybound mid-game (no card does) is noticed at the next entry or flip. The Clone example in CR 702.145b is already right: a copy is not represented by a double-faced card.
- **Shared team turns** (CR 502.2a, 731.2a/b): no team mode exists.
- **"It stops being day or night"** (The Weekly Princess, not Commander-legal): no verb. Once it is day or night it stays one (CR 731.1).
- **A nightbound face's self-replacement on entering transformed** (section 5): no face prints one.
- **Pricing the toggle in the heuristic** (section 10).
- **The other 37 cards** (#2586).
- **Enters-or-transforms-into triggers, the "as this transforms into Curse of Leeches" attach (built: [ADR 0079](0079-transforming-a-permanent.md), amendment 2026-10-09, decision 11), and Tovolar, Dire Overlord's "it becomes night. Then transform any number of Human Werewolves"** (a permanent with daybound can't be transformed by the instruction, CR 702.145b, so the second sentence does nothing to a real werewolf): each is a slice's decision, listed under #2586.
