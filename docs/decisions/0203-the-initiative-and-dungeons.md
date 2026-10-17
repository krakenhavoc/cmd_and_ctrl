# ADR 0203 — The initiative and dungeons

**Status:** Accepted · 2026-10-17 · S58 — Deck requests, October batch
**Issue:** [#4096](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4096) (relates to [#4092](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4092), the Captain N'ghathrod deck, which asked for From the Catacombs).
**Numbering:** `git fetch origin --prune`, then every `docs/decisions/` file name in the history of every remote branch (`git log --remotes --name-only -- docs/decisions`, 18 heads). The highest in use is **0202**, and no open issue's claim comment reserves 0203. Reserved on #4096 before any code was written.
**Builds on:** [ADR 0096](0096-the-monarch-from-a-card-effect.md) (the monarch: a player designation whose inherent triggers ride the listener registry), [ADR 0064](0064-emblems.md) and [ADR 0114](0114-the-ring-tempts-you.md) (a rules object in the command zone whose abilities are catalog rows), [ADR 0041](0041-game-persistence.md) (keyed bodies, keyed prompt continuations, the closure ratchet, the effect-key ledger) and [ADR 0018](0018-triggers-on-the-stack.md) (triggered abilities and their targets).

---

## Context

The issue cited CR 725 for the initiative. In the pinned edition (September 25, 2026) CR 725 is the monarch; the initiative is **CR 726**:

- **726.2**: three inherent triggered abilities with no source, controlled by the player who had the initiative as they triggered — "At the beginning of the upkeep of the player who has the initiative, that player ventures into Undercity," "Whenever one or more creatures a player controls deal combat damage to the player who has the initiative, the controller of those creatures takes the initiative," and "Whenever a player takes the initiative, that player ventures into Undercity."
- **726.3**: one player at a time. **726.4**: when the holder leaves, the active player takes it at the same time (the next player in turn order if the active player is the one leaving). **726.5**: a holder told to take it again triggers the venture but gets no second designation.

Dungeons are **CR 309** and venturing **CR 701.49**: a dungeon card goes to its owner's command zone with a venture marker on its topmost room (309.4a); each room has a triggered room ability, "When you move your venture marker into this room, [effect]", controlled by the dungeon's owner (309.4c); a venture moves the marker along an arrow, the player choosing when two lead on (701.49b), or completes the dungeon from its last room and enters a new one (701.49c); a marker on the last room with no room ability of that dungeon still waiting to leave the stack completes it as a state-based action (309.6, CR 704.5t); "venture into [quality]" restricts the dungeon chosen (701.49d). The Undercity's own text adds "You can't enter this dungeon unless you 'venture into Undercity.'"

Before this ADR `Game.Initiative` was a bare marker written only by the sandbox's `set_initiative`; nothing read it. There was no dungeon, no venture and no completed-dungeon record.

## Decisions

### 1. The initiative is the monarch's shape

`game/initiative.go`. `takeInitiativeLocked` is the one write that gives a player the initiative; the card effect (`TakeInitiativeForEffect`, `effects.TakeTheInitiative`), the combat-damage trigger, CR 726.4's hand-on and the sandbox all go through it. It emits **`EventInitiativeTaken`** on **every** take, not only a change of holder, because CR 726.5 makes a re-take trigger the venture. `Actor` is the taker, `Target` the previous holder.

The three CR 726.2 abilities are a listener, `initiativeTriggers`, registered beside `monarchTriggers`: they have no source card for the harvester to find. Each queues a sourceless `StackItem` with a **keyed body** — `initiative/venture` (Params.Player ventures into Undercity) and `initiative/take` (Params.Player takes the initiative) — so the table gets a response window and a table with one waiting is a restore point. Controller: the player who had the initiative as it triggered.

"One or more creatures a player controls": the engine emits one combat damage event per creature, so the second and later events of one damage step find the first `initiative/take` trigger still on `PendingTriggers` (same taker, same controller) and add nothing. A first-strike step and the regular step are two batches with priority between them.

CR 726.4 runs on `EventPlayerEliminated` and is a take, so the new holder ventures. A player who leaves also takes their dungeon with them (CR 800.4a).

**The sandbox.** `set_initiative` with a player is the same take a card makes, so the player ventures — the sandbox action reaches a game state, it does not bypass what the state means (ADR 0096 §1). Clearing the marker is a correction and ventures nobody; it emits `EventInitiativeTaken` with no `Actor` so the layer pass re-reads.

### 2. A dungeon is data, and its rooms are catalog rows

`game/dungeon.go` holds the room graph as a `DungeonDef` registered by the catalog (`game.RegisterDungeon`): rooms in printed order, each with its arrows; `VentureIntoOnly` for the Undercity. `cards/effects/dungeons.go` declares a dungeon as a `DungeonSpec` and `registerDungeon` files both halves: the graph with the engine, and the room abilities as ordinary `TriggeredAbility` rows under the catalog key `dungeon:<key>` (into `defs`, not `registry`, so the card census never counts it — the Ring emblem's filing). Another dungeon is another `DungeonSpec`, not more code.

Each room ability watches **`EventVentureRoom`** (`Actor` the player, `CardID` the dungeon object, `Label` `"<dungeon>/<room>"`) for its own room. Because it is a catalog row, a room that targets (Forge, Trap!, Arena) declares its clause on the row and gets CR 603.3d and 608.2b from the trigger machinery, and its stack item names its catalog row (ADR 0041 P9), so it restores.

**The dungeon card is not a `Card` in a `Zone`.** It can't be cast, isn't a permanent and never leaves the command zone except by leaving the game (CR 309.2c), so a container every zone verb could resolve would only be a door to close. `Player.Dungeon` is `DungeonState{ID, Key, Room}`, and the harvester walks it as one more command-zone object (`harvestFromDungeonsLocked`, beside `harvestFromEmblemsLocked`), building a transient object from the state — only for `EventVentureRoom`, so it costs every other event one comparison. `Player.CompletedDungeons` is the completed keys, in order.

### 3. Venture

`(*Game).VentureForEffect(player, source, into)`: no dungeon — enter one (asked when more than one may be entered; `into` names the quality) with the marker on the topmost room; marker in any room but the last — follow an arrow (asked when two lead on); marker in the last room — complete the dungeon, then enter one. Every marker move emits `EventVentureRoom`.

**The choice is a keyed option pick** (`PendingChoiceOptionPick` with `ThenKey`): `venture/room` and `venture/dungeon`, registered with `RegisterOptionPickThen`. Each option carries the room or dungeon it names on two new hint fields, `ChoiceOption.Room` and `ChoiceOption.Dungeon` (wire `room` / `dungeon`), which the continuation reads rather than the index. No closure is stored, so a table waiting on the choice is a restore point and the closure ratchet does not move. A new `PendingChoice` kind was considered and not taken: the option pick already has the prompt, the enumerator's answers, the client's picker and the snapshot shape; the hint fields are all a bot or a client needs to know what a branch is.

**Completion** (`completeDungeonLocked`) removes the dungeon, appends its key to `CompletedDungeons` and emits **`EventDungeonCompleted`**. The CR 704.5t state-based action (`dungeonCompletionSBALocked`) completes a dungeon whose marker is on the last room unless a room ability of that dungeon object is on `PendingTriggers`, in `StackMeta`, or announcing (a pending choice whose source is the dungeon).

The layer pass bumps on `EventDungeonCompleted` and `EventInitiativeTaken`: "as long as you've completed a dungeon" and "if you have the initiative" are inputs no permanent moving stands in for.

### 4. Persistence and the wire

- **Snapshot** (additive under the current schema, recorded in `testdata/snapshot_shape/v9.txt`): `seats[].dungeon {id, key, room}`, `seats[].completedDungeons`, and the two option hint fields. A file without them restores a player in no dungeon who has completed none. A marker on a dungeon or room this binary does not register is refused at restore (`unknownDungeonKeys`, the ADR 0041 P10 refusal), because a marker nobody can advance is a table that cannot move.
- **Effect keys** (ledger): `body initiative/take`, `body initiative/venture`, `option venture/room`, `option venture/dungeon`.
- **Wire**: `PlayerView.dungeon` (`DungeonView{id, key, name, room, rooms[{key, name, text, leads_to}]}`) and `PlayerView.completed_dungeons` (names), public and identical for every viewer; `PickOptionView.room` / `.dungeon`. Log kinds `venture` ("Alice ventures into Undercity — Forge") and `dungeon_completed`; `EventInitiativeTaken` is silent (the resolving line says it, `GameView.initiative` carries the holder).

### 5. Hidden information

The dungeon card, its marker and the completed list are public. Lost Well's scry is the ordinary scry (only the scryer knows the cards); Secret Entrance reveals the land it takes; the Throne of the Dead Three reveals the top ten to every seat and shuffles after. Pinned by `initiative_view_test.go`.

### 6. Client and bot

`DungeonChip.svelte` on every seat's panel names the dungeon and room and opens a card listing the rooms, the current one and the ones the next venture may enter marked in words; the completed count rides the chip. The branch is the ordinary option-pick dialog, each option reading "Forge — Put two +1/+1 counters on target creature."

The heuristic prices a room by its key (`aiseat/heuristic/venture.go`): Forge only with a creature of its own (otherwise the required target is an enemy creature), Arena only with an enemy creature to goad, Trap! highest, then Catacombs, Archives, Stash and Lost Well.

### 7. Cards

From the Catacombs (requested), and seventeen more, all Full: Aarakocra Sneak, Avenging Hunter, Goliath Paladin, Underdark Explorer, Stirring Bard, Vicious Battlerager, Passageway Seer, Feywild Caretaker, Imoen, Mystic Trickster, Safana, Calimport Cutthroat, Bloodboil Sorcerer, Undermountain Adventurer, Dungeoneer's Pack, Loot Dispute, Explore the Underdark, White Plume Adventurer and Trailblazer's Torch. Card vocabulary in `cards/effects/initiative.go`: `TakeTheInitiative`, `WhenThisEntersYouTakeTheInitiative`, `IfYouHaveTheInitiative` (an intervening "if" with a footprint, `FootprintCondInitiative`), `YouHaveTheInitiative`, `YouCompletedADungeon`, `WheneverYouCompleteADungeon`.

## Consequences

- **Not built: the three Adventures in the Forgotten Realms dungeons** (Lost Mine of Phandelver, Dungeon of the Mad Mage, Tomb of Annihilation) and the "venture into the dungeon" cards that choose among them. A plain venture with only some of them filed would offer fewer dungeons than the rules give, so no plain-venture card ships until all three are rows. The machinery is in place: `VentureForEffect(player, source, "")` already asks among every dungeon not marked `VentureIntoOnly`. Filed as [#4100](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4100).
- `PutFromLibraryOntoBattlefield` gained `Counters` (the Throne's "with three +1/+1 counters on it"), seeded onto the entry.

**Back-outs.** Removing the `harvestFromDungeonsLocked` call fails every room test (`TestTakingTheInitiativeVenturesIntoSecretEntrance`'s search, `TestForgePutsTwoCountersOnTheTargetedCreature`); removing the `dungeonCompletionSBALocked` call fails `TestTheUndercityWalkCompletesAndStartsAgain` and `TestLootDisputeLoudRuckusOnCompletingADungeon`; dropping the pending-trigger dedup in `initiativeCombatDamageTriggerLocked` fails `TestCombatDamageToTheHolderTakesTheInitiativeOncePerBatch`; dropping the `EventInitiativeTaken` arm of the listener fails `TestTakingTheInitiativeAgainVenturesAgain`; dropping `initiativeLeftTheGameLocked` fails `TestTheInitiativePassesToTheActivePlayerWhenItsHolderLeaves`; dropping the upkeep arm fails `TestTheHolderVenturesAtTheBeginningOfTheirUpkeep`; dropping the `roomAbilityOutstandingLocked` exception fails `TestTheLastRoomStaysUntilItsAbilityLeavesTheStack`; dropping the trigger-time half of `IfYouHaveTheInitiative` fails `TestPassagewaySeerGrowsOnlyWithTheInitiative`; dropping the `ventureRoomValue` arm of the heuristic fails the four `venture_test.go` tests.
