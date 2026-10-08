# ADR 0136 — Speed: start your engines! and max speed

**Status:** Accepted · 2026-10-08 · S58 — Deck requests, October batch (tracker [#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077))
**Issues:** [#2122](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2122) (seam: speed, found landing the ADR 0116 pool, #2078). Deck request: Perilous Snare in the Sami Whammy deck ([#2190](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2190)). Registry row: `speed`.
**Owner decisions:** none needed. The rules fix every behaviour below. The one modelling choice (section 5, a player condition in each ability's existing predicate rather than a fifth designation kind) has no product effect and is explained there.
**Numbering:** 0135 (`0135-alternative-costs-that-tap-discard-awaken-and-emerge.md`) is the highest number on `origin/develop`. `git log --remotes -- 'docs/decisions/0136*'` after `git fetch origin` finds nothing on any remote head, and the 2026-10-08 sweep reserved 0136 for #2122 (claim comment on the issue).
**Builds on:** [ADR 0129](0129-energy-getting-and-paying-it.md) (a per-player value that rides the snapshot, the wire, the seat chip and the board text), [ADR 0071](0071-designations-that-switch-abilities-on.md) (the "[word] — [ability]" gate this does NOT reuse, and why), [ADR 0096](0096-the-monarch-from-a-card-effect.md) and `game/monarch.go` (inherent, sourceless triggered abilities that use the stack), [ADR 0132](0132-day-and-night.md) (a designation with an event, a layer bump, a log line, a client chip and a card vocabulary, all in one PR), [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator agrees with the engine), [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) decision 6 (a seam PR lands the cards it unblocks).

The seam lands in one PR: the rules, the snapshot, the wire, the client chip, the bot's board text and twelve proof cards. Nothing is deferred to a later PR except the other Commander-legal speed cards, which are listed on the registry row.

---

## Context

Aetherdrift's speed mechanic is on 40 Commander-legal cards in the 2026-10-05 Scryfall dump. Every one of them prints "Start your engines!" and nearly every one prints at least one "Max speed — [ability]". None could be catalogued: the engine had no per-player speed, nothing set it, nothing raised it, and no ability could ask whether its controller had max speed. Gastal Raider (#2122) and Perilous Snare (Sami Whammy, #2190) were waiting on it.

### The rules

Every rule below was read in the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026):

- **CR 702.179a:** "Start your engines! is a static ability. If a player controls a permanent with start your engines! and that player has no speed, their speed becomes 1. This is a state-based action." **CR 704.5aa** lists it among the state-based actions.
- **CR 702.179b:** "Players do not have speed until a rule or effect sets their speed to a specific value." No rule lowers a speed, so a player keeps theirs after the permanent that started it has gone.
- **CR 702.179c:** "If a player has no speed and they are instructed to increase their speed by a certain value, their speed becomes that value."
- **CR 702.179d:** "There is an inherent triggered ability associated with a player having 1 or more speed. This ability has no source and is controlled by that player. That ability is 'Whenever one or more opponents lose life during your turn, if your speed is less than 4, your speed increases by 1. This ability triggers only once each turn.'" It uses the stack, and its "if" is an intervening if (**CR 603.4**).
- **CR 702.179e:** "A player has max speed if their speed is 4." **CR 702.179f:** a player with no speed has speed 0 "for the purpose of an effect that refers to speed".
- **CR 702.178a:** "A max speed ability is a static ability. 'Max speed — [Ability]' means 'As long as your speed is 4, this object has [Ability].'" **CR 702.178b:** "If an ability granted by a max speed ability states which zones it functions from, the max speed ability that grants that ability functions from those zones." The glossary's "Max Speed" names whose speed: "that permanent's controller (or that card's owner, if it isn't on the battlefield)". So a Surveyor's "Max speed — {3}, Exile this card from your graveyard: Draw a card." works from the graveyard while its owner's speed is 4.
- **CR 119.4 and CR 120.3a:** paid life is lost life, and damage from a source without infect makes the player lose that much life. Infect damage, and damage to a player whose life total can't change, lose no life (#2105).

No Commander-legal card in the 2026-10-05 dump prints "increase your speed" outside the reminder text, or "your speed can't increase". Only CR 702.179c's verb is built for the first; the second is not built.

## Decision

### 1. One field on `Player`, one writer (`game/speed.go`)

`Player.Speed int` is the player's speed: 0 for none (every player at the start of a game), then 1 to `game.MaxSpeed` (4). `setSpeedLocked` is the only write. It refuses a value outside 0..4, does nothing when the value would not change, and emits `EventSpeedChanged` (`Actor` the player, `Amount` the new speed). `increaseSpeedLocked` raises by one and stops at 4; a player with no speed gets 1 (CR 702.179c). A player with no speed reads as 0 (CR 702.179f).

Readers, for effects (caller holds `g.mu`): `Game.SpeedOf(player) int` and `Game.HasMaxSpeed(player) bool`. They read `Player.Speed` and nothing else, so they are safe from inside a layer pass.

### 2. Start your engines! is a canonical keyword, and its state-based action reads it

`"start your engines!"` joins `canonicalKeywords` (exported as `game.KeywordStartYourEngines`) in the change that teaches the engine to honour it. The deck importer already lowercases Scryfall's `keywords` array and filters it against that table, so a deck-imported speed card that has no catalog entry starts its controller's speed too. A catalog card declares it in `PrintedKeywords`. The roadmap registry gains the keyword item `TestEveryCanonicalKeywordHasOneItem` asks for.

`startYourEnginesSBALocked` is CR 702.179a / 704.5aa. It runs in `stateBasedActionsLocked` after the token and prepare-copy sweeps: for each player still in the game with no speed, if a battlefield permanent they control has the keyword (`HasKeyword`, so a granted instance counts and a phased-out permanent, which is not on the battlefield slice, does not), their speed becomes 1 and the pass reports that it fired. A player who already has speed is untouched; one whose last speed permanent left keeps their speed (CR 702.179b), because nothing but this check and the trigger ever writes it.

### 3. The inherent trigger is a listener, as the monarch's are (`speedTriggers`)

The trigger has no source, so there is no card for the harvester to find it on. It rides the listener registry, registered right after `monarchTriggers` and for the same reason (`monarch.go`'s header): a card trigger and the speed trigger watching one event put the card's on `PendingTriggers` first, and CR 603.3b orders them.

`speedTriggers.OnEvent` reads two events, the two that carry a life loss (the `b04OpponentLostLife` pairing): `EventChangeLife` with a negative `Amount`, and `EventDealDamage` to a player with `DamageLifeLoss() > 0`. When the player who lost life is not the active player, and the active player is still in the game, has speed, has speed below 4 (the intervening if, checked as it triggers) and has not had the ability trigger this turn, the listener queues one item through `queueHarvestedTriggerLocked`:

- `Controller` and `Owner` the active player, `SourceCardID` `uuid.Nil` (no source; `StackOverlay` renders a sourceless item by its label, as it does the monarch's).
- `Label` "Speed — increase <name>'s speed".
- `Body` the keyed body `speed/increase`, with the player in `Params.Player`, so the item has no closure and an open trigger is a restore point (ADR 0041 P9).

"This ability triggers only once each turn" is `PlayerTurnTally.SpeedTriggered`, set as the item is queued, so a second loss in the same turn (another creature's combat damage, a second drain in the same resolution) does not trigger again even if the first item has not resolved. "One or more opponents lose life" is the same rule for a simultaneous loss: the first event in the batch triggers it, the rest find the flag set. The tally resets as the turn begins, with the rest of it.

The body is CR 603.4's second check: if the player's speed is still below 4 (a second trigger cannot exist, but a future "increase your speed" could have raised it), it increases by one. An eliminated player's trigger leaves the stack with them (CR 800.4a), as the monarch's does.

Every opponent is an opponent: Commander has no teams in this engine.

### 4. Speed is a layer input

`layerVersionBump` bumps on `EventSpeedChanged`, unconditionally, beside `EventMonarchChanged` and `EventDayNightChanged`. A static gated on max speed (section 5) otherwise keeps a cached characteristic from before the change, and Gastal Raider would be a 2/1 at speed 4 until something else moved. A speed changes at most four times a game per player, so there is no gate on the bump.

### 5. "Max speed — [ability]": the controller's speed, read in each ability's own predicate

ADR 0071's `Designation.Active(c Card)` is object-only by contract. It takes no `*Game`, so it can be called from inside the layer pass and from the trigger harvester's hot loop, and every designation it answers is state on the object (a Class's level, a Case's solved flag, a Room's doors). The Ring's count (ADR 0114 §2) is the one player-scoped gate, and it stays object-only because the count lives on the Ring emblem, which the gated abilities are printed on. Speed is different on both counts: it is the PLAYER's, and the abilities it switches on are printed on up to 40 different cards that the player controls, in any zone. Threading a `*Game` through `activeOnly` would change the signature of the four accessors and their 38 callers for one kind, and mirroring the speed onto every card the player controls would be a second copy to keep in step through every control change and zone move.

So max speed is read where each ability kind already reads the game, through one predicate, `Game.HasMaxSpeed(player)`, wrapped by one constructor per slot in `effects/speed.go` so a card file says `MaxSpeed…` once:

| Slot | Constructor | Where the condition goes | Result while speed < 4 |
|---|---|---|---|
| Static (layer 6 / 7c) | `MaxSpeedStatic(s)`, `MaxSpeedSelfPump(p, t)`, `MaxSpeedSelfKeywords(kw…)` | `StaticAbility.AppliesTo` | applies to nothing |
| Triggered | `MaxSpeedTrigger(t)` | `TriggeredAbility.AppliesTo` | never triggers (no prompt, no target) |
| Activated | `MaxSpeedActivated(a)` | `ActivatedAbility.Condition` (CR 602.1b slot), ANDed with any condition the ability already has | refused (`ErrConditionNotMet`), not offered by the enumerator, greyed in the popover with `condition_unmet` |
| Mana | `MaxSpeedMana(m)` | `ManaAbility.Condition` | refused; the auto-tapper skips it (`autotap.go` already reads `Condition`) |
| Replacement | `MaxSpeedReplacement(r)` | `ReplacementEffect.AppliesTo` | never applies |
| Cost modifier | `MaxSpeedCostModifier(m)` | `CostModifier.AppliesTo` (the source's controller) | changes no cost |

The controller is the source's controller on the battlefield and its owner elsewhere (CR 108.4a: a card in a graveyard has no controller, and "you" on it is its owner). `ActivatedAbility.Condition` is handed the activating player, which is the same player.

**The one observable difference from ADR 0071's gate** is the activated and mana rows: an unsatisfied designation makes the ability absent from the wire, and an unsatisfied condition shows it greyed. For max speed that is the better display: the label reads "Max speed — …", so the player sees what they are racing toward and why it is grey. In the rules the two differ only for an effect that asks whether an object HAS an activated ability (Necrotic Ooze, Experiment Kraj copy activated abilities). Such a copy carries the condition with it and reads its new controller's speed, which is what "as long as you have max speed" says on the copy. No speed card is affected.

**Layer invalidation** for the static row is section 4. Triggers, replacements, activations and cost modifiers are read when they are asked, so they need none.

### 6. The snapshot

Two additive fields within schema 7, both omitted at their zero value so every frozen fixture re-encodes byte for byte: `seats[].speed` (`Player.Speed`) and `turnTally.players{}.speedTriggered` (the once-per-turn flag). A file written before them restores as nobody having speed, which is every game before this change. A binary from before them, handed a file with them, drops them; that binary does not know the keyword either, and the next binary to open the game sets speed 1 again at the first state-based check for any player who controls a speed permanent, so the loss is at most the speed above 1. `Clone` copies `Player.Speed` by value and the turn tally already deep-copies its player map, so undo restores both. `TestSnapshotShapeIsRecorded` records the two paths.

### 7. The wire, the log and the client

- `PlayerView.speed` (`int`, omitted at 0): public, identical for every viewer (`FilterViewFor` does not touch it). The client reads `max speed` as `speed === 4`; no second field.
- The log gains `speed` (`LogSpeed`, `EventSpeedChanged`): "Alice's speed is now 2", or "Alice has max speed" at 4. Narrated because the state-based action sets speed to 1 with no spell or ability behind it, and because the speed rising is the thing a table of Aetherdrift decks watches for.
- **The seat chip.** `PlayerIdentity.svelte` shows a speed marker on every seat (self and opponents alike, since speed is public) while the seat has speed: a gauge icon and the number, styled as max speed at 4, with a title in plain words ("Speed 3 of 4. It rises once on each of your turns when an opponent loses life."). Nothing while a seat has no speed, which is every seat at a table with no speed cards. It is a display, not a stepper: speed has no sandbox control, because no rule lets it go down and the state-based action and the trigger are its only writers.
- `docs/protocol.md` documents `speed` and the `speed` log kind.

### 8. Bots

No new prompt, cost or move kind. A max-speed activated or mana ability is a `Condition`, which the enumerator already evaluates through the engine's own check (the #544 invariant), so a bot is never offered one the engine would refuse; `legal/speed_test.go` pins that. The speed trigger has no choice in it. The board text (`aiseat/boardtext`) prints the seat's speed after its player counters ("speed 3", "speed 4 (max speed)"), so a model seat and the MCP seat can see who is close to max speed. The heuristic does not price speed: it does not attack an opponent to raise its speed. That is a pricing question for ADR 0126 §8's tuning, not a rules one, and is listed under Out of scope.

### 9. Card-side vocabulary (`effects/speed.go`)

- `StartYourEngines` (the keyword token, for `PrintedKeywords`).
- `YourSpeed(g, player) int`, `YouHaveMaxSpeed(g, player) bool` for an effect that reads the number ("where X is your speed").
- The six `MaxSpeed…` constructors in section 5, plus `MaxSpeedCondition()` (the bare `ActivationCondition`, for a hand-built ability).

## Proof cards

Twelve, chosen deck cards first and then by EDHREC rank, so that every row of section 5's table has at least one card:

| Card | What it proves |
|---|---|
| Gastal Raider | the waiting card: the CR 702.179a keyword, and a max-speed static that is both a layer 7c pump and a layer 6 keyword grant |
| Perilous Snare | the deck card (#2190): a max-speed activated ability with a target and sorcery timing, beside an O-Ring exile |
| Muraganda Raceway | a max-speed mana ability beside an ungated one |
| Amonkhet Raceway | a max-speed targeted activated ability on a land |
| Vnwxt, Verbose Host | a max-speed replacement (draw two instead) |
| Avishkar Raceway | a max-speed activated ability with a discard cost |
| The Speed Demon | "where X is your speed" read at resolution |
| Racers' Scoreboard | a max-speed cost modifier |
| Starting Column | a max-speed sacrifice ability beside an any-colour mana ability |
| Aether Syphon | a max-speed triggered ability |
| Burnout Bashtronaut | a max-speed keyword grant (double strike) |
| Goblin Surveyor | a max-speed activated ability that works from the graveyard |

The other 28 Commander-legal speed cards are on the `speed` registry row's Waiting list. Most need nothing beyond this seam and are left for a card slice to keep this PR reviewable; the ones that need something else name it.

## Tests

- `game/speed_test.go`: a game starts with no speed; a start-your-engines permanent gives its controller speed 1 and nobody else any; speed is kept after the permanent leaves; an opponent losing life on your turn queues one sourceless trigger that raises the speed only when it resolves, your own loss queues none, and a second loss that turn queues none; damage that costs life counts; nothing happens off your turn; nothing triggers at 4 and the write refuses 5; the speed rises again on your next turn; CR 702.179c's no-speed-becomes-1; the change is an event and survives `Clone`.
- `cards/effects/speed_cards_test.go`: one test per proof card, each checking the max-speed half off at speed 3 and on at 4 (Gastal Raider's P/T and menace, Perilous Snare's refused and greyed activation then its counter, Muraganda Raceway's {C}{C}, Amonkhet Raceway's haste, Vnwxt's double draw, Avishkar Raceway's rummage, The Speed Demon's draw-and-lose X, Racers' Scoreboard's discount for its controller only, Starting Column's sacrifice, Aether Syphon's mill, Burnout Bashtronaut's double strike, Goblin Surveyor from the graveyard); every card declares the keyword; an uncatalogued card put through `deck.ToGameCard` starts its controller's speed.
- `legal/speed_test.go`: Perilous Snare's max-speed activation is not offered at speed 3 and is offered, and accepted, at 4.
- `protocol/speed_view_test.go`: `PlayerView.speed` is the same for every viewer and omitted for a player with none; the `speed` log line's text below and at max speed.
- `aiseat/boardtext`: the seat line prints "speed 3" and "speed 4 (max speed)", and nothing for no speed.
- Client: `playerIdentitySpeed.render.test.ts` (no chip without speed, a chip on your seat and an opponent's, max speed marked, the hover text) and `logKind.test.ts` (the `speed` kind is on both sides).
- The snapshot shape file, the drift test's classification of `Player.Speed`, the effect-key ledger (`speed/increase`), the oracle fixtures for the twelve cards and the roadmap tables carry the new state.

### Back-outs

Each change was reverted by itself, the named tests watched fail, and the change restored:

| Reverted | Fails |
|---|---|
| the SBA call in `stateBasedActionsLocked` | `TestStartYourEnginesGivesSpeedOne`, `TestGastalRaiderGivesSpeedAndIsBiggerAtMaxSpeed` |
| the `speedTriggers` registration | `TestSpeedRisesOnceOnYourTurnWhenAnOpponentLosesLife` |
| the once-per-turn flag check | `TestSpeedRisesOnceOnYourTurnWhenAnOpponentLosesLife` |
| "the loser is not the active player" | `TestSpeedRisesOnceOnYourTurnWhenAnOpponentLosesLife` |
| the below-4 check as it triggers | `TestSpeedStopsAtFour` |
| the damage arm (`DamageLifeLoss`) | `TestSpeedRisesFromDamageThatCostsLife` |
| CR 702.179c in `increaseSpeedLocked` | `TestIncreasingNoSpeedGivesOne` |
| the `EventSpeedChanged` arm in `layerVersionBump` | `TestGastalRaiderGivesSpeedAndIsBiggerAtMaxSpeed`, `TestBurnoutBashtronautDoubleStrikeAtMaxSpeed` |
| `Clone` copying `Speed` | `TestSpeedChangeIsAnEventAndSurvivesClone` |
| `MaxSpeedActivated`'s condition | `TestPerilousSnareCounterNeedsMaxSpeed`, `TestAmonkhetRacewayGrantsHasteAtMaxSpeed`, `TestStartingColumnSacrificeNeedsMaxSpeed`, `TestGoblinSurveyorDrawsFromTheGraveyardAtMaxSpeed`, `TestMaxSpeedActivationIsOfferedOnlyAtMaxSpeed` |
| `MaxSpeedMana`'s condition | `TestMuragandaRacewayTwoColorlessAtMaxSpeed` |
| `MaxSpeedStatic`'s gate | `TestGastalRaiderGivesSpeedAndIsBiggerAtMaxSpeed`, `TestBurnoutBashtronautDoubleStrikeAtMaxSpeed` |
| `MaxSpeedTrigger`'s gate | `TestAetherSyphonMillsOnlyAtMaxSpeed` |
| `MaxSpeedReplacement`'s gate | `TestVnwxtDrawsTwoOnlyAtMaxSpeed` |
| `MaxSpeedCostModifier`'s gate | `TestRacersScoreboardDiscountsOnlyAtMaxSpeed` |
| `start your engines!` in `canonicalKeywords` | `TestImportedUncataloguedSpeedCardStartsSpeed`, `TestEveryCanonicalKeywordHasOneItem` |
| `PlayerView.Speed` | `TestPlayerViewCarriesSpeedForEveryViewer` |
| the board text's speed phrase | `TestRenderPrintsSpeed` |

## Consequences

- Every Commander-legal speed card's speed half is now expressible, and a deck-imported speed card with no catalog entry still starts and raises its controller's speed.
- `stateBasedActionsLocked` gains one check. It walks the battlefield once per pass only for players with no speed, and returns at once at a table where nobody does, so it costs nothing at a table with no speed cards beyond one loop over the seats.
- `layerVersionBump` gains one unconditional arm.
- The listener list gains one entry. It does work only on a life-loss event during the turn of a player with speed below 4, so it is idle at most tables.

## Out of scope

- **"Increase your speed" and "your speed can't increase".** No Commander-legal card prints either outside reminder text. `increaseSpeedLocked` is the verb a future card would call, and already follows CR 702.179c (no speed becomes 1).
- **A sandbox control for speed.** Speed only goes up, and its two writers are rules the engine performs. A table that wants to correct it has undo.
- **The heuristic valuing speed** (section 8).
- **Two-Headed Giant and team play**, where "opponents" would differ. The engine has no team variant.
- **The other 28 cards** (the registry row's Waiting list, [#2711](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2711)).
