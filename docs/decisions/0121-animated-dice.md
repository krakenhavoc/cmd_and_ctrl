# ADR 0121 — Animated dice: an opening roll at the table, dice you can watch, and a Roll a die action

**Status:** Accepted · 2026-10-04 · S61 — Dice you can watch (milestone; issue [#2229](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2229))
**Issues:** [#2229](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2229), which is both this change and the sprint's only issue. No separate S61 tracker exists.
**Owner decisions:** the four answers of 2026-10-04 recorded on #2229, quoted under [Owner decisions](#owner-decisions-2026-10-04). They are binding. This ADR also makes twenty smaller calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before the PR that implements each one lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 38 remote heads, including `origin/develop`, `origin/main`, the open PR branches `feat/s60-0119-pr7-fuller-log`, `fix/s58-0115-pr4-lift-caveats` and `feat/1548-heuristic-gang-blocks`, and the card, chore, docs, feat, fix, repro and wip branches. I also listed the 180 local branches in this checkout. The highest number on any of them is 0120 (`origin/develop`). No branch has 0121, so this ADR takes **0121**.
**Amends:** [ADR 0054](0054-dice-rolls-and-coin-flips.md) owner decision 3 ("a roll or flip shows a log line and a reveal-strip cue, with no animation"). The owner reversed it on 2026-10-04 (decision 4 below). The log line and the strip cue stay; an animation is added. #1486's "the winner of the d20 roll goes first" (`starting_player.go`) is replaced by a choice (decision 2).
**Builds on:** [ADR 0054](0054-dice-rolls-and-coin-flips.md) (keyed RNG streams, rewind, the roll and flip events and log), [ADR 0075](0075-table-settings-and-host-controls.md) §2.1 and §2.3 (the host, and the no-undo path for table actions), [ADR 0076](0076-tutorial.md) §2 (the practice table starts with seat 0 and no roll), [ADR 0111](0111-action-dock.md) (dock requests, the ⋯ menu, labels as a contract), [ADR 0119](0119-a-stack-you-can-follow.md) (the pile, the linger and the attention strip share the screen with the dice), [ADR 0120](0120-expand-a-players-board.md) §3 (the board-anchor helper and the z ladder), [ADR 0033](0033-ai-bot-seat.md) (the legal-move enumerator and bot tiers).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

Every claim below was checked on `origin/develop` at `4d6a1f1a`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

### Owner decisions (2026-10-04)

The request on #2229: "Make dice rolls visible and fun: an animated die that the whole table sees, and rolls that never lock other players out." The owner added: "Multiple players can roll dice at once so it doesn't lock out the other players while someone is rolling", and the result should be "visible to all players naturally".

1. **The opening roll happens at the table, before mulligans.** When the game starts, each seat gets a Roll button. Players roll whenever they like, all at the same time, and each result animates for everyone. Bots roll for themselves. Tied leaders roll again.
2. **The winner chooses who starts (CR 103.1).** Today the winner simply goes first.
3. **No timer for a player who doesn't roll.** The host gets a "roll for everyone left" button.
4. **The animation covers all three:** the opening roll; every die a card rolls and every coin flip (amending ADR 0054 owner decision 3); and a new "Roll a die" table action (d6, d20 or coin, any time, never blocking, logged).

### The rules

- **CR 103.1:** "At the start of a game, the players determine which one of them will choose who takes the first turn. In the first game of a match … the players may use any mutually agreeable method (flipping a coin, rolling dice, etc.) to do so." The roll picks a **chooser**, not the starting player. "The player chosen to take the first turn is the starting player."
- **CR 103.2c and 903.6:** the commander goes to the command zone, then each player shuffles. **CR 103.3:** "After the starting player has been determined and any additional steps performed, each player shuffles their deck." **CR 903.7:** "Once the starting player has been determined, each player sets their life total to 40 and draws a hand of seven cards." **CR 103.5:** each player then draws their starting hand, and the mulligan declarations begin with the starting player. So in paper nobody has seen a hand when the choice is made.
- **CR 706.1, 706.1a, 706.1b:** an *effect* instructs a player to roll; a dN has N equally likely results; players "may agree to use an alternate method for rolling a die, including a digital substitute, as long as the method used has the same number of equally likely outcomes". The server's keyed draw is that substitute. The animation is presentation of a result already drawn.
- **CR 705.1, 705.2:** a coin has two equally likely sides. A flip that only cares about heads or tails has no winner or loser and no call.
- A die rolled at the table for fun is not an instruction from an effect (CR 706.1). It is not a game action, and nothing in the game may trigger on it.

### What exists

**The opening roll is automatic and runs after the deal.** `Game.start` (`server/internal/game/game.go:1102`) shuffles every library, deals seven cards to each hand (`:1130-1156`), and only then, when `rollForFirst` is set, calls `rollStartingSeatLocked` (`:1158-1160`). That function (`starting_player.go:11`) rolls a d20 per seat through `RollDiceForEffect(RandomDraw{Player: seat}, 20, 1)`, under `g.mu`, and `chooseStartingSeat` (`:21`) rerolls tied leaders until one remains. The winner takes turn 1 (`newStartingTurn`, `turn.go:206`). Nobody presses anything.

**Callers.** `StartWithFirstPlayerRoll` (`game.go:1076`) is called by the lobby's start (`lobby/lobby.go:1014`), the demo seed (`cmd/server/main.go:1234`), the bot arena (`botarena/arena.go:479`) and the probe (`cmd/boteval/probe.go:469`). The practice table calls plain `Start(nil)`, so seat 0, the human, takes turn 1 (`lobby/practice.go:206`, ADR 0076).

**The RNG.** `randForLocked` (`rng.go:82`) seeds one ChaCha8 per operation from `HMAC-SHA256(key, kind ‖ player ‖ source ‖ turn ‖ counter)` (`rngSeed`, `:180`). The counters are per stream and scoped to `Turn.Seq` (`rngTurnIndexLocked`, `:149`). The opening shuffle and the opening rolls both draw at turn index 0, on different streams: `("shuffle", seat, nil)` and `("roll", seat, nil)`. Clone and `RestoreFrom` carry the counters, so undo rewinds randomness (ADR 0054 Decision 4).

**Events and the log.** `RollDiceForEffect` (`effect_api_random.go:10`) emits one `EventRollDie` per die and `FlipCoinsForEffect` (`:28`) one `EventFlipCoin` per coin, grouped by `BatchSeq`. A called flip is a `coin_call` pending choice (`:91`, `:105`). The public log folds a batch into one `LogRoll` or `LogFlip` entry (`protocol/log.go:166-167`, collapse at `:795`) with `sides`, `results`, `faces`, `call` and `wins` (`:507-511`). The text is built by `renderRandomLogText` (`protocol/random_log.go`): "Alice rolled a d20 for Hoarding Ogre: 14", "Alice called heads for X: tails — lost".

**The client.** `reveals.ts` cues each new `roll` or `flip` entry once per `seq` (`trackRandomEvents`, `:84`), primes the window on a reconnect (`primeRandomEvents`, `:76`), and forgets rewound entries after an undo. `RevealBanner.svelte` draws the cue in the attention strip ("rolled" or "flipped", `:165`). `startingPlayer.ts` finds the opening roll by a heuristic: a `roll` entry with `sides: 20`, no `card_id` and no `turn`. `Game.svelte` shows it as the `.opening-roll` pill in the strip's `opening hand decisions` banner (`:1733`) and in the mulligan sheet (`:1948`).

**The pre-game window.** `MulligansOpen` is set at Start (`game.go:1167`). While it is true, step entry hooks do nothing (`game.go:1741`), the legal enumerator offers only keep and mulligan (`legal/legal.go:488`), and `KeepHand` closes the window when every seat has kept (`mutations.go:9029`). The mulligan is simultaneous.

**Undo and table actions.** `Room.Apply` (`ws/room.go:254`) clones before each action and pushes an undo entry. `Room.ApplyExternal` (`:381`) records none; the hub routes `set_table_settings` there after a host-or-admin gate (`ws/hub.go:979`, `Room.CanManageTable`, `ws/host.go:122`). `RestoreFrom` **truncates the event log** to the clone's length (`clone.go`, the `g.Events = src.Events` block). So an undo of an action taken *before* an `ApplyExternal` action removes that later action's log line too.

**Animation settings.** `settings.animations` has a master `enabled` (off at first load when the OS asks for reduced motion), a `speed` multiplier and per-effect toggles (`cardDraw`, `cardPlay`, … `damagePopups`) (`client/src/lib/settings.ts:43-60`, defaults `:340-352`). `SETTINGS_VERSION` is 17 (`:290`).

**The screen.** The z ladder inside `.board` (ADR 0120 §2, ADR 0119): panel chrome ≤ 10, the expanded board 32, arrows 35–36, the stack lane and pile 38, the attention strip 40, the dock 55, card-local menus 60–70, modals 200, the hover zoom 300. ADR 0119's pile sits at the left edge; its linger shows a departed item for 1.5 s.

**Contracts.** Every e2e spec that starts a table (11 files call `startGameAs`, `tests-e2e/tests/lobby-api.ts:84`) then waits for `dialog "keep or mulligan your hand"` or sends `keep_hand` from the admin. `full-game.spec.ts:82-101` and `attack-all-318.spec.ts:94-97` already tolerate a random starting seat. `tutorial-layout.spec.ts` uses the practice table.

### What the code contradicts

1. **The hands are dealt before the roll.** Under CR 103.3, 103.5 and 903.7 the shuffle and the draw come after the starting player is determined. While the winner only went first this was invisible. Once the winner *chooses* (decision 2), a winner who has seen their seven cards chooses with information paper never gives. §1 moves the deal after the choice.
2. **The winner goes first with no choice** (`starting_player.go`), which decision 2 replaces.
3. **Undo drops the log line of a later table action** (the truncation above). It already happens to a settings change. For a table roll it would erase a roll the whole table watched (§5).
4. **The mulligan is simultaneous**, where CR 103.5 has the starting player declare first and the others in turn order. Out of scope here; see [Out of scope](#out-of-scope).

---

## Decision

### 1. The opening roll is a pre-game window

**Where it lives.** The game is `StateActive` with `MulligansOpen` true, as at Start today, plus a new `Game.OpeningRoll *OpeningRoll`, non-nil while the roll is open. A new `State` value was considered and rejected: 129 non-test call sites read `StateActive`, among them the bot runner's exit (`aiseat/runner.go:428`), the lobby's started-at bookkeeping (`lobby/persist.go:170`) and the client's routing, and every one would need a decision. `MulligansOpen` already means "active, but the first turn has not begun", and every guard that reads it (step hooks, instant windows, the untap choice) keeps the turn machinery parked for free.

```go
type OpeningRoll struct {
    Rounds  []OpeningRollRound // round 1 is every seat; each later round is the tied leaders
    Chooser int                // the seat that chooses; -1 while rolling
}
type OpeningRollRound struct {
    Seats []int             // seats that roll in this round
    Rolls []OpeningRollDie  // in the order they were rolled
}
type OpeningRollDie struct {
    Seat, Result int
    By           int // the seat that pressed the button: the roller, or the host (§3)
}
```

**The order.** A new entry point, `StartWithOpeningRoll(r)`, does what Start does today up to and including life totals and undo budgets, then opens the roll. It does **not** shuffle or deal. The turn cursor stays parked at the pre-game turn: `Turn.Seq` 0, step untap, `PriorityHolder` `NoPriority`. When the chooser chooses (§2), the engine, in one action:

1. sets `StartingSeat` and `Turn = newStartingTurn(chosen)`, and notes the turn begun, as `start` does today;
2. shuffles each library and deals seven, exactly the code `start` runs today, moved (CR 103.3, 103.5, 903.7). The shuffle still draws at turn index 0 on `("shuffle", seat, nil)`, before the cursor moves to `Seq` 1, so for a given key every library comes out as it does today;
3. clears `OpeningRoll`. `MulligansOpen` stays true, and the mulligan proceeds as it does now.

**A round.** A seat in the current round's `Seats` with no die in its `Rolls` may roll once. Its die is `RollDiceForEffect(RandomDraw{Player: seat}, 20, 1)`, unchanged: the stream `("roll", seat, nil)` at turn index 0, its counter advancing once per die. When every seat in the round has rolled, the engine finds the highest result. One leader becomes `Chooser`. Several leaders open a new round with just those seats (decision 1). Results are public the moment they are rolled, as in paper.

**The same winner either way.** A seat's k-th opening die is fixed by the key, the seat and k. It does not depend on when the seat rolls, on who rolls first, or on whether the host pressed the button for it. So:
- the interactive roll and the automatic one find the same chooser for the same key;
- the host's button (§3) cannot change anyone's result, only when it is shown;
- nothing is decided in state before a die is rolled. The result is drawn when the action is applied, as ADR 0054 Decision 6 requires of a called flip.

**The automatic path stays.** `StartWithFirstPlayerRoll` is reimplemented on the same window: open it, roll every contender in seat order round by round, have the chooser choose themselves, deal. That is what the arena (`botarena`), the probe (`boteval`) and the demo seed keep calling. Because the streams are independent and the shuffle stays at turn index 0, a seeded arena game is the same game it is today: same rolls, same starting seat, same libraries. The order of events in its log changes (rolls before draws). The practice table keeps calling `Start(nil)` and never opens the window (ADR 0076).

**A seat that leaves during the roll.** Concede is allowed while the roll is open. The seat is eliminated without the active-player departure path (there is no active player yet; CR 103.8 has not happened). It is struck from every round, and the standings are recomputed from the latest round that still has a seat in it: a single leader among the remaining seats chooses, several reroll. If one seat is left, the game ends as it does today (last standing).

**Nothing else happens while the roll is open.** `actions.Dispatch` gains an **allowlist** gate: while `OpeningRoll` is non-nil, only `roll_opening`, `host_roll_remaining`, `choose_starting_player`, `roll_table_die`, `concede`, `set_table_settings` and `set_undo_limit` are dispatched. Every other type, and every type added later, is refused with a new `ErrOpeningRollOpen` ("the opening roll is not finished"). An allowlist fails closed. `KeepHand` and `Mulligan` also refuse with it, as a second guard, because a hand does not exist yet. The legal enumerator returns only the roll moves (§4) while the window is open, ahead of its `MulligansOpen` branch.

**Undo.** Nothing done while the window is open mints an undo entry, a concede included (§3). The first undo entry of a game is taken by the first ordinary action after the deal, so no undo, the admin's included, can reach back into the roll. Without this, undoing a concede made mid-roll would restore the roll as it stood before the concede and silently take back every die rolled since.

**Persistence.** `gameSnapshot` gains `openingRoll,omitempty` (rounds, seats, rolls with seat, result and by; chooser). It is additive under schema v7: the shape is recorded with `-update-shape` in the same PR, and a new generated corpus fixture (a 4-seat table mid-roll after a tie) is added beside the others (the writer never touches an existing file, AGENTS.md §5). The drift test marks `OpeningRoll` `carried`. `cloneLocked` deep-copies it and `RestoreFrom` restores it. A restored game mid-roll continues where it stopped, and its next die is the one it would have rolled.

**Rollback.** A binary from before this change reads such a file, ignores the unknown key, and finds an active game with empty hands and the mulligan open. That table is broken until the host ends it. The alternative, a schema bump, abandons every live game on rollback (ADR 0041 Decision 5), and a table caught mid-roll by a rollback deploy is rare. ADR 0054 Decision 3 accepted the same trade.

### 2. The winner chooses

`choose_starting_player {player, params: {seat}}` is player-scoped and accepted only from `Chooser` (or the admin). `seat` is any seat that is not eliminated, the chooser's own included (CR 103.1: "who takes the first turn"). The engine then runs the three steps of §1. There is no confirm step and no undo: as at a paper table, "you go first" is final. The dock makes the choice one deliberate click (§6).

A chooser who never chooses holds the table, as a player who never passes priority does. There is no timer (decision 3) and no host override of the choice; see [Calls made here](#calls-made-here), item 6.

### 3. The wire

**Actions.** All four are in `actions.Type`, listed in `docs/protocol.md`, and dispatched under the §1 gate.

| Action | Who | Params | Effect |
|---|---|---|---|
| `roll_opening` | the seat (player-scoped) or the admin | — | One d20 for that seat in the current round. Refused unless the seat is in the round and has not rolled in it (`ErrNotYourRoll`). |
| `host_roll_remaining` | the host or the admin (`Room.CanManageTable`, checked in the hub like `set_table_settings`) | — | Rolls, in seat order, for every seat in the **current** round that has not rolled, with `By` set to the host's seat. A tie it causes opens a new round with its own Roll buttons. Refused when no seat is left to roll. |
| `choose_starting_player` | the chooser (player-scoped) or the admin | `{seat}` | §2. |
| `roll_table_die` | any seated player (player-scoped) | `{die: "d6" \| "d20" \| "coin"}` | §5. Legal at any time while the game is active, during the roll and the mulligan included. |

**No undo entry.** One predicate, `actions.MintsNoUndo(g, t)`, is true for these four, for the two settings verbs, and for every action while the opening roll is open (in practice, `concede`). The hub and the bot runner, the two places that call `Room.Apply`, route a type it names through `Room.ApplyExternal` instead. Rolls are simultaneous: an undo entry for each would bury every seat's real last action under other seats' rolls and make "undo your own most recent action" refuse.

**View fields.**

```json
"opening_roll": {
  "rounds": [
    { "seats": [0, 1, 2, 3], "rolls": [ {"seat": 2, "result": 17}, {"seat": 0, "result": 17},
                                        {"seat": 1, "result": 9, "by": 0}, {"seat": 3, "result": 4} ] },
    { "seats": [0, 2], "rolls": [ {"seat": 2, "result": 11} ] }
  ]
}
```

`GameView.opening_roll` is present only while the window is open. `by` is present only when the host rolled for the seat. `chooser` (a seat) is present once one leader remains; in the example, seat 0 has not rerolled yet, so it is absent. Everything in it is public. `starting_seat` keeps its meaning and its shape, and a client ignores it while `opening_roll` is present (it reads 0 then).

**Events and log.** All additive and `omitempty`, recorded in the shape file under v7, and each new event kind is projected (the kind gate test lists none of them as silent):

| Log kind | Text | From |
|---|---|---|
| `roll` (existing) | "Alice rolled a d20: 17" | each opening die, unchanged |
| `opening_roll` | "Alice and Carol tied with 17 and roll again" · "Carol won the opening roll with 20" · "Bob rolled for Carol and Dave" | a round ending, and the host's button (`seats` lists the seats, `results` the high result) |
| `starting_player` | "Carol chose to take the first turn" · "Carol chose Bob to take the first turn" | the choice (`seat` the chooser, `target_seat` the chosen) |
| `table_roll` | "Alice rolled a d20 at the table: 14" · "Alice flipped a coin at the table: heads" | §5 (`sides` 6 or 20, `results` or `faces`, `roll_id`) |

`seats []int` and `roll_id` are new `LogEvent` fields. `startingPlayer.ts` stops guessing from source-less d20s and reads the `starting_player` entry, so a table d20 rolled before turn 1 can never be taken for the opening roll.

### 4. Bots

**The enumerator** (`legal`). While the window is open, a seat gets:
- `roll_opening`, `AlwaysLegal`, when it is in the current round and has not rolled;
- when it is the chooser, one `choose_starting_player` per seat that is not eliminated, its own marked `AlwaysLegal`;
- nothing otherwise.

It never offers `host_roll_remaining` (a bot is never the host, `ws/host.go:33`) or `roll_table_die` (it changes nothing in the game, and a random-tier bot would roll every window). `dispatchAll` accepts every offered move (#499, #544).

**The policies.** Layer A (`aiseat/rules`) absorbs a lone `roll_opening`, and absorbs the choice as the chooser itself, so a model tier never spends a call on either. The heuristic does the same. The random tier picks uniformly, as it does in every window, so a random bot may hand the first turn to someone else. Going first is the usual choice in Commander, and nothing a bot sees in its view before the deal (no hand exists) could argue otherwise.

**Pacing.** A bot rolls and chooses on its normal clock (`MinThink`, by the table's bot speed), so three bots at normal speed roll within about a second of Start, concurrently with the people. `TestManagerPlaysALobbySeatedTable`, which seats bots through the lobby, now passes through the window, and is the whole-game check that bots roll and choose. The arena, the probe, the whole-game tests that start through `StartWithFirstPlayerRoll` and the lockstep runner keep the automatic path and never see the window.

### 5. The Roll a die table action

**What it is.** `roll_table_die` rolls a d6 or a d20, or flips a coin, for the seat that asked. It is not a game action: no effect instructed it (CR 706.1), so it emits **no** `EventRollDie` or `EventFlipCoin`, and no "whenever you roll one or more dice" or "whenever you win a coin flip" ability can see it. It emits a new `EventTableRoll`, which no watcher matches. It changes nothing but the log. It is legal at any time while the game is active (decision 4), mints no undo entry (§3), and never opens a pending choice, so nobody waits for anybody.

**Its own stream, outside the turn counters.** The die is drawn from `rngSeed(key, "table", seat, nil, 0, Game.tableRollNext)`, then `tableRollNext` is incremented. It does not go through `randForLocked` and never touches `rngCounters`:
- a card roll, a shuffle and a pick draw exactly what they would have drawn had no table roll happened, in the same turn or any other;
- the HMAC key keeps it unpredictable, as every other draw is (ADR 0054 Decision 2);
- `tableRollNext` is per game, never reset, and **carried forward** across undo the way `Settings` are (`RestoreFrom` leaves it), so a table roll after an undo is a fresh roll, not a replay of one already seen. It is persisted (`tableRollNext,omitempty`, additive under v7) so a restored game does not repeat earlier results.

The draw lives in `rng.go`, so `TestNoDirectRandomSource` holds, and a known-answer test pins it.

**An undo does not erase it.** The game keeps the last 32 `EventTableRoll` events it emitted in a ring that `RestoreFrom` does not restore. After `RestoreFrom` truncates the log, it re-emits each of those whose `Seq` is past the restored end, with its payload unchanged. The roll keeps its `roll_id` (its `tableRollNext` value), so the client animates it once, whatever `seq` it has now. The line moves to after the restored history, which is still after everything that happened before it. The ring is not persisted: after a restore there is no undo stack to cross.

**Rate.** The hub accepts one `roll_table_die` per seat per 2 s and refuses the next with "wait for your last roll to land". The log ring holds 200 entries (`PublicLogMax`), and a held-down button must not push the game's history out of it.

### 6. The table during the opening roll

**The dock.** For a seat that has to roll, a dock request (ADR 0111 §1, rule 1: something you owe) opens a non-modal dialog **`roll for the first turn`** whose primary is **`Roll`** (Enter). Its question line reads "Roll a d20. The highest roll chooses who goes first." or, in a later round, "You tied with 17. Roll again." The roll request never blocks another seat's dock.

For the host, once the host has rolled or has nothing to roll, the request becomes "Waiting for Bob and Dave to roll" with a secondary **`Roll for everyone left`**. It is not a primary, so Enter never presses it. If the host still has to roll, the same secondary sits beside `Roll`.

For the chooser, a sheet **`choose who takes the first turn`** lists the seats in turn order, each a button named **"<name> goes first"**, the chooser's own named **"I go first"**. The line above reads "You won the roll with 20. Choose who takes the first turn." There is no primary, so Enter chooses nothing; a click does. Everyone else sees "Carol is choosing who goes first" in the dock's status line.

`next`, Pass turn and the toggles are disabled while the window is open, as during the mulligan.

**The attention strip.** The `opening hand decisions` banner becomes **`opening roll`** while the window is open: one chip per seat in its colour, with its result, "rolling…" or "—" (not in this round), the round's tied leaders outlined, and the chooser marked once known. Strip cues are not raised for opening dice, because this banner already shows each one. After the choice the banner goes back to `opening hand decisions`, and its `.opening-roll` pill reads the `starting_player` entry: "Carol won the d20 roll with 20 and goes first" (the existing text when the winner chose themselves) or "Carol won the d20 roll with 20 and chose Bob to go first".

**Spectators** see the banner and the dice and have no dock. The board is empty while the window is open: no hand has been dealt.

### 7. The animation

**One component, three sources.** A new `DiceLayer.svelte`, mounted in `Board` beside `CombatArrows`, draws every die and coin from one queue. It reads new `roll` and `flip` log entries (card rolls and opening dice), `table_roll` entries, and the `opening_roll` view. The server's result is always known before anything moves: the client never draws a random number, and an animation starts only from the frame that carries its result.

**Duration.**
- **Tumble:** 900 ms × `animations.speed`. A d6 is a cube and a d20 a flat icosahedron outline, both in the roller's seat colour. The faces shown while it tumbles come from a fixed sequence seeded by the entry's `seq` (or `roll_id`), not from `Math.random`, so every viewer sees the same tumble and it lands on the server's number.
- **Hold:** the settled result stays for 1.6 s. This is reading time and does not scale with `speed`, the rule ADR 0119 §3 uses for the linger.
- **Out:** a 200 ms fade. The opening roll's dice do not fade; they stay settled by their seats until the choice is made, as the standings.
- **Coin:** a coin spins on its vertical axis for the same 900 ms and lands on the face. A called flip (ADR 0054's `coin_call`) shows the call above the coin ("called heads") during the spin and **won** or **lost** under it at the hold.

**Where it is drawn: at the roller's seat.** Each die settles beside the roller's avatar (`[data-seat-id]`, found through ADR 0120's `boardAnchor` helper, so an expanded board's copy is used when one is open), on the side of the avatar toward the board's centre. Four seats rolling at once are four places that never overlap, so concurrent rolls never queue behind each other (the owner's "multiple players can roll dice at once"). A seat whose avatar is not on screen gets its die at the attention strip's cue instead. The layer is **z 41**: above the strip (40), the pile (38) and the expanded board (32), below the dock (55), menus and modals. It is `aria-hidden` and takes no pointer events, so it never covers a click.

**Batches.** One log entry is one animation: "Alice rolled 2d12 for Reckless Endeavor" tumbles two dice side by side and settles them together. Up to six dice are drawn; a larger batch draws six and a "+N" chip, and the strip cue lists every result. Yusri's five coins spin as a row.

**Bursts.** A seat with several entries arriving between two frames plays them in log order, each `max(400 ms, 2.5 s / n)`, at most three queued. Older ones are dropped; the log and the strip cue keep them. This is ADR 0119 §3's burst rule.

**Display only.** Nothing waits for a die. The server never delays a broadcast, the bots do not pause for it, a prompt that follows a roll (a `coin_call`'s next flip, Reckless Endeavor's choose-one) opens at once over a tumbling die, and input stays live. That is the rule of ADR 0053 and ADR 0119 §3, and it is what keeps a roll from locking anyone out.

**Alongside ADR 0119.** A card's roll usually happens during its resolution. The pile's linger shows that item resolving on the left for 1.5 s while the die tumbles at the roller's seat; the two are in different places and run together. ADR 0119 §2's stack hold is unchanged: the roll happens after the hold, when the item resolves. The strip's text cue (`RevealBanner`) for a roll is raised when the die settles rather than when the frame arrives, so the text never gives the number away mid-tumble. With motion off it is raised at once.

**Reduced motion and the toggle.** A new `animations.dice` per-effect toggle, default on, labelled in Settings → Animations "Dice and coins: animate rolls and flips". It sits under the master `enabled` switch, which is off at first load when the OS asks for reduced motion. With either off there is no tumble and no spin: the settled die or coin appears at once in the same place for the same 1.6 s hold, and fades. The result is information, so it shows either way; only the motion is gated (ADR 0119 §3). The new field is filled by the settings merge, so it needs no migration.

**Reconnects, replays and undo.** The first frame after mounting, a reconnect or a replay toggle primes every entry already in the window without animating it (`primeRandomEvents`, extended to `table_roll`; opening dice already rolled are drawn settled). A card roll that an undo rewinds and the redo rolls again animates again; it is the same result (ADR 0054 Decision 4). A table roll re-emitted after an undo does not (§5).

**Announced.** The strip's live region reads each settled roll once: "Alice rolled a d20: 14", "Alice flipped a coin at the table: heads", "Carol won the opening roll with 20".

### 8. Settings, labels and logging

**Settings.** One new field, `animations.dice` (§7). No table setting: the opening roll is how every real table starts (decision 1), and a host who wants it over fast has the button (decision 3).

**New accessible names** (AGENTS.md §5, labels are a contract):
- `dialog "roll for the first turn"`, its primary `Roll`, and the secondary `Roll for everyone left`;
- `dialog "choose who takes the first turn"` and its buttons "<name> goes first" and "I go first";
- the strip banner `opening roll`;
- the ⋯ menu items `Roll a d6`, `Roll a d20` and `Flip a coin` (ADR 0111 decision 3's menu). They are disabled for 2 s after the viewer's own table roll, matching the hub's rate.

**Names that stay.** `region "actions"`, `region "attention"`, `dialog "keep or mulligan your hand"`, `Keep hand`, `Mulligan`, `next`, `Pass turn`, `more actions`, and the `opening hand decisions` banner after the choice.

**Logging.** The four log kinds of §3. `docs/protocol.md` lists them, the four actions, `opening_roll` and the two new `LogEvent` fields. `client/src/lib/protocol.ts` mirrors them. The log panel gets an icon for `opening_roll`, `starting_player` and `table_roll`.

### 9. Tests and contracts

**Go, `internal/game`.**
- The window: every seat rolls in any order and the same chooser results as `StartWithFirstPlayerRoll` for the same key (a property over many keys); a tie opens a round with only the leaders; a seat cannot roll twice in a round or roll outside it; only the chooser can choose; hands are empty before the choice and seven after; the libraries equal the automatic path's for the same key; a concede recomputes the standings, including the chooser's; the host's roll rolls only the current round's missing seats, with `By`.
- The gate: a table-driven test sends every `actions.Type` except the allowlist while the window is open and expects `ErrOpeningRollOpen`, so a new action type fails it until it is classified.
- Persistence: capture mid-roll, restore, and the next die equals the live game's; the shape file records `openingRoll` and `tableRollNext`; the drift rows; clone and `RestoreFrom` carry `OpeningRoll`.
- The table stream: a table roll leaves `rngCounters` unchanged, and the next card roll, shuffle and pick are the same with or without it; `RestoreFrom` neither rewinds `tableRollNext` nor loses a table roll's log line; no `EventRollDie` or `EventFlipCoin` is emitted, and a "whenever you roll" watcher does not fire; a known answer; 20,000 d20s and 10,000 coins under a fixed key are fair within a fixed tolerance, as ADR 0054 item 8.
- Existing tests that pin the event order of `StartWithFirstPlayerRoll` are re-pinned for rolls-before-draws, and the PR lists them.

**Go, elsewhere.**
- `ws`: none of the four verbs mints an undo entry; `host_roll_remaining` is refused for a non-host; a second table roll inside 2 s is refused; an undo of A's action after B's table roll keeps B's line.
- `legal`: the moves of §4, one `AlwaysLegal` each, every one accepted by `dispatchAll`.
- `aiseat`: Layer A and the heuristic roll and choose themselves, and agree; the random tier answers; `TestManagerPlaysALobbySeatedTable` finishes.
- `botarena`: a pinned seeded game has the same starting seat, the same opening hands and the same winner as before this change.
- `protocol`: the view, and the text of each log kind.

**Vitest.** As pure functions: the dice plan (one animation per entry, the six-die cap and "+N", the burst schedule and its cap of three, the deterministic tumble sequence, the settle time, no motion with the toggle or the master off); priming; anchoring and its strip fallback; the opening-roll view model (standings, ties, the chooser, which request each role gets); `startingPlayer.ts` reading `starting_player` and never a table d20. Render tests for the two dialogs, their names and buttons, the strip banner, and the ⋯ menu items.

**Playwright.** `startGameAs` gains an opt-out-able finish: after `POST /games/{id}/start` it sends `host_roll_remaining` as the admin until a chooser exists, then the chooser's `choose_starting_player` for themselves, and returns the starting seat. Every existing spec keeps its flow unchanged with that. `full-game.spec.ts` opts out and plays the roll through the UI: both players press `Roll`, each sees the other's result in the `opening roll` banner, the winner chooses the other player, and that player takes turn 1. One new spec, `table-roll-2229.spec.ts`, rolls a d20 from the ⋯ menu during the opening roll and during a turn, and both pages show the log line. Ties are not forced in Playwright (they are random); the Go tests cover them. Every client PR, and the PR that switches the lobby, runs the nightly E2E on its branch before it merges (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), and the nightly runs on `develop` after the last one.

---

## Delivery

Each PR goes into `develop`, Sprint S61, Issue #2229.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **This ADR,** the S61 section and index row in `docs/sprints.md`, the AGENTS.md §3 ADR range line, and an "Amended by ADR 0121" line on ADR 0054. Docs only. | — | — |
| 2 | **Server: the opening roll window** (§1–§3). `OpeningRoll`, `StartWithOpeningRoll`, the deal moved after the choice, `StartWithFirstPlayerRoll` on the same window, the three opening verbs, the dispatch allowlist and `ErrOpeningRollOpen`, concede during the roll, `actions.MintsNoUndo` in the hub and the runner, the host gate, the view field, the events and log kinds, the snapshot field, shape, corpus fixture and drift rows, `docs/protocol.md`, the Go tests. The lobby still calls `StartWithFirstPlayerRoll`, so no table changes yet. | 1 | 4 |
| 3 | **Bots** (§4). The enumerator's moves, Layer A, the heuristic, the random tier, and the Go tests. | 2 | 4 |
| 4 | **Client: the dice layer for card rolls and flips** (§7). `DiceLayer.svelte`, the queue and plan, the die and the coin, seat anchoring, batches and bursts, the cue raised on settle, `animations.dice` and its Settings row, the announcer, the vitest, the nightly E2E on the branch. Needs nothing new from the server. | 1 | 2, 3 |
| 5 | **Client and lobby: the opening roll at the table** (§6, §8). The dock requests and the chooser's sheet, the strip banner, opening dice in the layer, `startingPlayer.ts` on `starting_player`, the e2e helper and `full-game.spec.ts`, and the lobby's start switched to `StartWithOpeningRoll`. The vitest, the nightly E2E on the branch. | 2, 3, 4 | 6 |
| 6 | **Server and client: Roll a die** (§5). The `table` stream, `tableRollNext`, `roll_table_die`, `EventTableRoll` and its re-emission after an undo, the `table_roll` log kind, the hub's rate, the ⋯ menu items, table rolls in the layer, `table-roll-2229.spec.ts`, the Go tests and vitest, the nightly E2E on the branch. | 2, 4 | 5 |

PR 5 is the one that changes what a person sees at Start, so it lands only after the bots can roll (PR 3); otherwise a dev table with a bot would wait for a die nobody casts. After PR 6: run the nightly E2E on `develop`, then close #2229 with evidence: the run, and a cmd-dev table of two people and a bot in which all three roll at once, a tie rerolls if one comes up, the winner hands the first turn to someone else, a Hoarding Ogre's d20 tumbles at its controller's seat, and a table d20 rolled mid-turn shows in both logs.

## Consequences

- A game starts the way a paper game does: everyone rolls, the high roll chooses, and only then is anything shuffled or drawn. Nobody chooses with their hand in view.
- Starting a table takes a few seconds longer: everyone presses Roll, and the winner chooses. Bots do both on their own clock, and the host can roll for anyone who has not.
- The automatic path stays for the arena, the probe and the demo, and a seeded game there is unchanged.
- Every die and coin in the game, and a table die, tumbles for everyone at the roller's seat, for under 3 s, without slowing anything.
- A table roll is visibly not part of the game: it triggers nothing, is drawn from its own stream, cannot be undone, and keeps its log line through other players' undos.
- One new pre-game field in the snapshot and one counter, both additive under v7. A rollback binary breaks a table caught mid-roll.
- Six new accessible names and one settings field. The existing names stay.

## Out of scope

- **Mulligans in turn order** (CR 103.5: the starting player declares first). The mulligan stays simultaneous. It is a separate change, and it now has a starting player to start from.
- **The London mulligan's bottom-N** (the sheet's "simplified London" note). Unchanged.
- **Power Play** (CR 103.1c) and any other card that sets the starting player.
- **A timer** for a seat that does not roll or a chooser who does not choose (decision 3).
- **Other dice** (d4, d8, d10, d12, d100) as table rolls. Decision 4 named d6, d20 and a coin. Card rolls of any size already animate (§7).
- **Keeping table-action log lines through undo in general.** §5 does it for table rolls. A settings change still loses its line when an earlier action is undone; that is a separate fix.
- **Sound.** The audio settings exist, but a dice sound is not part of this change.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review:

1. **The window is `StateActive` plus `OpeningRoll`, not a new `State`** (§1). Fewer places change, and `MulligansOpen` already parks the turn machinery.
2. **The shuffle and the deal move after the choice** (§1, CR 103.3, 103.5, 903.7). The streams keep every library identical to today's for the same key.
3. **The interactive and automatic paths are one state machine,** and the automatic one stays for the arena, the probe and the demo (§1).
4. **An allowlist gate** in `actions.Dispatch` refuses every other action type while the roll is open, new ones included (§1).
5. **Concede is allowed during the roll,** and the standings are recomputed without the seat (§1).
6. **No host override of the choice.** A chooser who never chooses holds the table like a player who never passes; the admin can act for the seat. The alternative was a host button "Let <winner> go first", which overrides the winner's CR 103.1 choice (§2).
7. **The choice is final**: no confirm, no undo (§2).
8. **`host_roll_remaining` rolls the current round only.** A tie it causes gives the tied seats their own Roll buttons (§3).
9. **None of the four verbs mints an undo entry, and neither does a concede during the roll,** decided by one predicate (§3).
10. **Bots choose themselves;** Layer A absorbs both the roll and the choice; the enumerator never offers the host's button or a table roll (§4).
11. **A table roll is not a game roll:** no `EventRollDie` or `EventFlipCoin`, nothing triggers, its own `table` stream and per-game counter outside the turn counters, carried forward across undo (§5).
12. **A table roll's log line survives an undo** by re-emission, keyed by `roll_id` so it is not animated twice (§5).
13. **One table roll per seat per 2 s,** enforced in the hub (§5).
14. **The roll and the choice are dock requests;** "Roll for everyone left" is a secondary, never Enter; the choice has no primary (§6).
15. **Opening dice raise no strip cues;** the `opening roll` banner shows them (§6).
16. **Dice are drawn at the roller's seat,** at z 41, with the strip as the fallback (§7).
17. **900 ms tumble scaled by speed, 1.6 s hold unscaled, 200 ms fade;** the tumble's faces are seeded by the entry, never `Math.random` (§7).
18. **Six dice at most, then "+N"; bursts follow ADR 0119's rule** (§7).
19. **The strip's text cue waits for the die to settle** (§7).
20. **An `animations.dice` toggle,** default on, under the master switch. The result shows with motion off (§7).
