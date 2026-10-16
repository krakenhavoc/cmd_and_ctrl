# ADR 0201 — Restarting the game

**Status:** Accepted · 2026-10-16 · Backlog
**Issue:** [#3856](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3856) (owner decision, Ian, 2026-10-15: build it)
**Numbering:** every remote head was listed (`git log --remotes -- docs/decisions`, 97 remote branches) and the open issues' claim comments were searched. At the claim the highest number in use was **0199** (`feat/3955-cast-from-outside`) and 0200 was reserved on the issue; `feat/3952-land-mana-each` has since used **0200** for another ADR, so this one moved to **0201**, the first number no remote head uses.
**Builds on:** [ADR 0057](0057-win-and-lose-by-effect.md) (which listed restarting the game as out of scope), [ADR 0059](0059-turn-machinery.md) amendment 2026-10-05 (the CR 724.1 "end the turn" deferral this copies), [ADR 0121](0121-animated-dice.md) (the opening roll and the pregame procedure), [ADR 0155](0155-having-another-objects-activated-abilities.md) §8 (the CR 607.2a "exiled with" link on the card).

---

## Context

Karn Liberated's −14 reads: "Restart the game, leaving in exile all non-Aura permanent cards exiled with Karn. Then put those cards onto the battlefield under your control." Until now it was the one ability in the catalog left out on purpose, with a caveat saying the engine could not restart a game.

The rules (Comprehensive Rules, 2026-09-25 edition):

- **727.1** A restarted game immediately ends. Nobody wins, loses or draws. All players in that game when it ended start a new game following rule 103, except that **727.1a** the starting player is the player who controlled the spell or ability that restarted it.
- **727.2** Every Magic card involved in the game when it ended, phased-out permanents included, is in the new one, even a card brought in from outside the game. Ownership does not change.
- **727.3** Each player draws seven as the new game begins, so a player with fewer than seven cards in their library loses at the first state-based check of the first turn's upkeep, whatever mulligans they take.
- **727.4** The restarting effect finishes resolving just before the first turn's untap step. Its further instructions happen then, with no player holding priority; what they trigger goes on the stack the next time a player receives priority.
- **727.5** An effect may exempt cards: they are not in their owner's deck as the new game begins. **727.5a** An exempted commander does not begin in the command zone and is still that deck's commander.

## Decision

### 1. Three beats, like ending the turn

The game the ability is resolving in is about to stop existing, and the resolution frame still has its own tail to run against it (the resolved line, the CR 704.3 hold, the post-resolution priority grant). So the restart cannot happen inside the ability's effect. It has the shape CR 724.1 has in `end_turn.go`:

1. **The request.** `Game.RestartTheGameForEffect(restarter, source, exempt)` records a `GameRestart` on `Game.RestartPending` and parks priority. It must be the last thing the effect does.
2. **The boundary.** `runStateChecksLocked`, once nothing holds the resolution open, sees `RestartPending` and calls `finishRestartLocked`, which ends this game and begins the new one, and returns without running anything else of the boundary (there is nothing left to check in a game that no longer exists, and nobody has priority in a mulligan window). `settleResolutionLocked` runs the boundary for a pending restart, and `cursorMovedSince` reports a pending restart as a moved cursor so the resolution frame does not hand priority back.
3. **The tail.** CR 727.4's "then" waits on `Game.RestartTail` for the new game's mulligan window to close. `settleMulliganLocked` runs it after the CR 103.6 opening-hand offers and before the first step's entry hooks: the exempted cards enter together through `PutOntoBattlefieldTogetherThenForEffect` under the restarter's control, so their replacement effects ask what they ask and their enters triggers wait on `PendingTriggers` for the first player to receive priority. The first turn's untap step begins from the entry's continuation (`restartTailEntered`), so a card that stops to ask (a Clone's copy, a shockland's life) is answered before the turn begins.

Both fields are plain data, carried by `Clone`, `RestoreFrom` and the snapshot (additive in v9), and classified `carried` by the drift test.

### 2. The new game is built fresh and adopted in place

The room, the hub, the bot runners and every handler hold the `*Game` and its mutex, so the game cannot be swapped for another pointer. `buildRestartedGameLocked` builds a fresh `NewGame()` and `RestoreFrom` adopts it, the way an undo adopts a clone. Every game-scoped field therefore starts at its `NewGame` value: life, poison, energy, speed, emblems, the monarch and the initiative, day and night, the turn plan, delayed triggers, scoped effects, cast permissions, player statics, commander damage and tax, the per-turn tallies, the loop breaker, every prompt. A field added to `Game` later resets on a restart without anyone remembering to list it.

What is **carried** is the short named list in `buildRestartedGameLocked`: the table's identity and settings, the event log and its Seq and batch counters, the listener and built-in replacement configuration, the table's rolls, and the loop threshold. Each in-game seat is rebuilt by `restartedPlayer`, which carries the seat's identity and its own preferences (name, seat, Discord and bot identity, the agent badge, trigger-order mode, standing answers, `DeckImported`) and its outside-the-game containers (a companion still outside may be revealed again under CR 103.2b; the sideboard is still the pool). `TestRestartClassifiesEveryPlayerField` holds every `Player` field to one side of that line.

A few transient fields `RestoreFrom` leaves alone because an undo never needs them reset are reset by hand: the scripted opening hand, the library-top watch, the scoped-effect memo, the prompt keys, and `TurnTally.FirstEvent`, which is set to the end of the kept log so no "this turn" reader walks the old game.

### 3. The pregame is the one Start uses

`start()`'s pre-roll half (starting life and undo budget, the command zone made public, the game made active, the mulligan window opened) is now `beginPregameLocked`, and a restart runs it and then `dealAndBeginLocked(restarterSeat)`, the same deal-and-begin `Start` ends with. There is no opening roll: CR 727.1a names the starting player. The mulligan window, CR 103.5c's free first mulligan, the bottom-of-library prompts, companions (CR 103.2b) and the CR 103.6 opening-hand actions all run exactly as at a game's start, and the bots answer them through the same enumerator moves.

A seat that has left the game stays out (CR 727.1: only the players in the game when it ended start the new one). Its `Player` is kept as it was, for spectating, and its cards are not in the new game (CR 800.4a took them out when it left).

CR 727.3 is the one rule a normal start does not have: after the deal, a player whose library and hand together hold fewer than seven cards is marked `AttemptedEmptyDraw`, and loses at the next state-based check.

### 4. Every card is a new object with a new instance ID

CR 727.2's sweep walks every zone a card can be in (each seat's library, hand, graveyard and command zone, the battlefield, phased-out permanents, the stack and exile) and puts each card its in-game owner owns into that owner's library, or the command zone for a commander. Tokens, copies of spells, card copies and prepare copies are not cards and do not come back. Each card is normalised by `restartCards`: it goes through `MoveCard` from a battlefield, the one place every per-object fact is stripped (CR 400.7), which also splits a melded permanent and turns a double-faced card front face up; then a copy effect ends and the zone-bound facts go.

Every card then gets a **new instance ID**. The event log is kept, and much of the engine walks it by card ID: Approach of the Second Sun's "this game", the CR 607 exile record, Clone's look-back. A reused ID would let the old game's history speak for a card in the new one. Source ordinals carry over to the new IDs, and the RNG key is derived afresh (`restartRNGKey`), so the new game's shuffles are not the old game's streams replayed from turn 0, and an undo of the restart that is redone derives the same key, as any undone shuffle does.

The exempted cards are swept into the new game's exile instead, also with new IDs, and those IDs are the tail's. An exempted commander keeps `IsCommander` (727.5a).

### 5. The log begins again

`EventGameRestarted` is emitted once, as the new game begins and before the deal: Actor the restarter, Source and Label the card, Amount the number of exempted cards. The source is an object of the old game, so its name rides `PublicNames` and the line names it for every viewer. The public log's fold resets its ring at that event (`logRing.reset`), so the restart line (`game_restarted`) is always the first line of the new game's log. The old game's lines are dropped from the view rather than kept: every card they name is a new object now, and they could only say "a card". They stay in the event log, the replay and a bug report.

The client's mulligan sheet reads that line for its first-turn headline in place of the opening roll's ("Alice restarted the game with Karn Liberated and goes first").

### 6. Karn Liberated

The +4 and the −3 now exile through `ExileLinkedForEffect` with the activating Karn object (`Context.SourceRef`), so the cards carry `Card.ExiledWith`. The −14 reads `ExiledWithForEffect` for that object and exempts the non-Aura permanent cards among them. An Aura, an instant or a sorcery exiled with Karn is not exempt and goes back to its owner's deck, as the printed text says. A Karn that left and came back, or a second Karn Liberated, is a different object with no claim on the first one's exiles (CR 400.7). The card is `CompletenessFull`.

### 7. Undo and replays

A restart happens inside an ordinary commit, so the room's undo entry for it holds the old game's clone and an undo restores the old game whole (`RestoreFrom` adopts every field the restart replaced). The replay records the commit like any other: its line is the new game's snapshot.

## Consequences

- A `*Player` held across a restart is the old game's, exactly as one held across an undo is. Nothing in the server holds one across a commit; the room, the lobby and the bot runners all go through player IDs, which a restart keeps.
- The engine runs state-based actions at its own boundaries, not at every priority grant, so the CR 727.3 loss lands at the first check of the new game's first turn rather than at a fixed moment of its upkeep.
- An older binary reading a snapshot taken between the mulligans and the tail drops `restartTail` and begins the first turn without the exempted cards, which stay in exile: weaker than the card, never stronger.

### What this does not do

- **CR 103.6 ordering.** The opening-hand actions (a Leyline) are offered as prompts when the mulligan window closes, and the tail's cards enter at the same moment. A Leyline whose prompt is answered after that enters after Karn's cards rather than before them. No printed card cares which came first.
- **CR 727.6 / 727.7.** There are no subgames and no limited range of influence in the engine.
