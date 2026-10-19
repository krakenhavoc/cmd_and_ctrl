# ADR 0167 — Controlling another player's turn: Emrakul, the Promised End

**Status:** Accepted · 2026-10-13 · Backlog.
**Issues:** [#2164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2164) (the `control-another-player` seam). Requested card: Emrakul, the Promised End ([#2034](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2034), the Jodi deck).
**Owner decisions behind it (2026-10-13):** this ADR lands now as its own docs PR, and the implementation follows on `develop` once [#3373](https://github.com/krakenhavoc/cmd_and_ctrl/pull/3373) has merged. The controller plays the turn through a "Playing as X" board swap, not through a legal-move sheet. See [Owner decisions](#owner-decisions-2026-10-13).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-13. I ran `git fetch --all --prune` and listed every file `docs/decisions/` has held on all 17 remote heads (`git log --remotes --name-only -- docs/decisions/`). `develop` ends at 0165. #3373's branch (`valgavoth-2032`) holds 0166, because it renumbered its Word of Command ADR from 0164 to 0166 that day. So this ADR takes **0167**.
**Builds on:** [ADR 0166](0166-controlling-another-player-word-of-command.md) (#3373: the `PlayerControls` record, `ControllerOfPlayerLocked`, routing, the view), [ADR 0059](0059-turn-machinery.md) Decisions 1, 5 and 6 (turn identity, extra turns, the rotation seam), [ADR 0060](0060-leaving-the-game.md) (CR 800.4), [ADR 0033](0033-ai-bot-seat.md) (the enumerator and bot seats), [ADR 0105](0105-legal-action-highlights.md) (the legal-action lookup), [ADR 0111](0111-action-dock.md) (the action dock), [ADR 0117](0117-click-to-act-and-a-per-colour-mana-stepper.md) (click to act) and [ADR 0127](0127-answering-repeated-prompts-for-you.md) (standing answers).

---

## Context

Emrakul, the Promised End ({13}, Legendary Creature — Eldrazi, 13/13):

> This spell costs {1} less to cast for each card type among cards in your graveyard.
> When you cast this spell, you gain control of target opponent during that player's next turn. After that turn, that player takes an extra turn.
> Flying, trample, protection from instants

It is the last manual card of #2034. Everything except the cast trigger already exists. The discount is a self cost modifier (`CostsLessEach`, `cards/effects/cost_modifier.go`) over the delirium count (`b16CardTypesInGraveyard`). "When you cast this spell" is `WhenYouCastThisSpell` (`cards/effects/annihilator.go`, used by Flayer of Loyalties and Ulamog). Protection from instants is a printed keyword the engine already enforces (Serra's Emissary's tests).

ADR 0166 built the control record for Word of Command, CR 723.2's limited form, and left Emrakul out (§7) because "the controller plays an entire turn, which needs the client to drive a second seat's board". This ADR covers that turn.

### The rules

Checked against `MagicCompRules 20260925.txt`. Rulings are Scryfall's (2025-01-24), read on 2026-10-13.

- **CR 723.1:** "This effect applies to the next turn that the affected player actually takes. The affected player is controlled during the entire turn; the effect doesn't end until the beginning of the next turn."
- **CR 723.1a:** when several effects control the same player, the last one created is the one that works. **723.1b:** if a turn is skipped, a pending effect waits for the next turn the player actually takes.
- **CR 723.3:** only control of the player changes. Objects keep their controllers, and the controlled player is still the active player.
- **CR 723.4:** the controller also sees what the controlled player can see about objects in the game (example: "that player's hand and the face of any face-down creatures they control"). Information about cards outside the game stays with the controlled player.
- **CR 723.5:** the controller makes every choice and decision the controlled player is allowed or told to make, "by the rules or by any objects". The examples include which spells they cast, what those spells target, which creatures attack, whom they attack and how they assign combat damage. **723.5a:** costs are paid with the controlled player's resources only. **723.5b:** the controller can't make choices that the rules or objects don't call for.
- **CR 723.6:** the controller can't make the controlled player concede. The controlled player may concede at any time.
- **CR 723.7:** an effect may restrict the controlled player's actions. Word of Command does; Emrakul doesn't, so the whole turn is the controller's to play.
- **CR 723.8:** the controller keeps making their own decisions.
- **CR 500.7:** extra turns are added "directly after the specified turn", and the most recently created is taken first.
- **CR 601.2i, 603.2:** the cast trigger triggers as the spell becomes cast and resolves above it.
- **CR 800.4a:** when a player leaves the game, "any effects which give that player control of any objects or players end". **800.4j:** a player who leaves during their own turn leaves that turn to finish without an active player. **800.4k:** a turn of a player who has left doesn't begin.
- **Rulings:**
  - "It resolves even if that spell is countered."
  - "The player you're controlling is still the active player during that turn."
  - "You can't make the affected player concede."
  - "You also can't make any choices or decisions for the player that would be called for by the tournament rules."
  - "If an effect causes another player to make decisions that the affected player would normally make (such as Master Warcraft does), that effect takes precedence."
  - "you can see all cards in the game that player can see. This includes cards in that player's hand, face-down cards that player controls, and any cards in that player's library the player may look at."
  - "Controlling a player doesn't allow you to look at that player's sideboard."
  - "you can use the affected player's resources only to pay that player's costs."
  - "If multiple players have cast Emrakul and targeted the same player, each ability's effect will create an extra turn."
  - "If the targeted player skips their next turn, you'll control the next turn the affected player actually takes, and the extra turn the player takes will be after that turn."
  - "If the targeted player loses the game while you control their turn during a multiplayer game, the extra turn never begins."

### What exists (develop plus #3373)

- **The record** (`game/player_control.go`, ADR 0166 §1). `Game.PlayerControls` is a list of `PlayerControl{Controller, Controlled, Source, SourceName, Card, Stage, Prompt}`. `controlLiveLocked` decides whether a record applies, and `ControllerOfPlayerLocked(p)` returns the newest live record's controller (CR 723.1a). The list is cloned (`clone.go`), restored (`RestoreFrom`) and snapshotted (`snapshot.go`, `playerControls`, additive in the v8 shape). `prunePlayerControlsLocked` runs after each resolution. `BeginControlledPlayForEffect` deletes every older record over the same player before adding its own.
- **Routing** (`actions/controlled_player.go`, `routeControlledPlayer`). A controlled caller may only concede. A controller's action that names the controlled player, or a `resolve_choice` for a prompt that player owes, runs with `Caller` and `Player` set to the controlled seat and `Actor` set to the sender. In `dispatch`, a `cast_spell` with `Actor != Caller` is refused unless it names a `controlled_play` prompt. While `LandManaOnly` applies (Word's forced play), only `cast_spell`, `activate_mana_ability` and `resolve_choice` pass.
- **Prompts and view.** A prompt keeps its real `Chooser`. `viewOfPendingChoices` shows it to the answerer with `on_behalf_of`. `GameView.player_controls` lists the live records, and `FilterViewFor` lets a controller know every card the controlled player knows (`isKnower`).
- **Legal moves** (`legal/controlled_play.go`). `newEnumerator` gives a controlled seat nothing. `controlledChoiceMoves` gives the controller the prompts the controlled player owes. `enumeratorAs` enumerates as another seat. `legal_moves` is the viewer's own list, capped at 48 by `capLegalMoves` (`protocol/view.go`) and built per seat by `legalMovesFor`.
- **Turns** (`game/rotation.go`, `game/extra_turns.go`). `beginNextTurnLocked` is the one rotation seam. `startTurnLocked` stamps the new turn, counts it (`noteTurnBegunLocked`) and runs `onTurnBeganLocked` before any step entry hook. Extra turns are a stack, `Game.ExtraTurns []ExtraTurn{Ref, Seat, Source}`. `TakeExtraTurnsForEffect` always pushes "after this one", and ADR 0059 Decision 5 says "if a card ever names another turn, the anchor is added then". `popExtraTurnLocked` drops a departed seat's turn (CR 800.4k) and opens the skip window (`RepEventExtraTurn`).
- **Undo** (`ws/room.go`). Each entry is stamped with the connection's seat (`ActionTag.Actor = c.playerID`, `ws/hub.go`), and `Room.Undo` refuses a seat whose stamp is not on the top entry (`ErrNotYourUndo`).
- **The client.** `Game.svelte` derives `viewerID` from the session, and nearly everything that acts reads it. That includes the `Board` prop (`seatPlacements(view.seats, viewerID)` in `lib/cardTypes.ts` puts that seat at the bottom, where `PlayerPanel` draws its `Hand`), `ActionDock`, the combat dock requests (`lib/combatDock.ts`), `AttackDeclarationModal`, the legal-action lookup (`legalActionsOf`, `lib/legalActions.ts`), click to act, and the autopass inputs (`components/board/priority.ts`, `autopassDecision.ts`, `priorityStops.ts`).

## Decision

### 1. A turn-long record

`PlayerControl` gains three fields, all plain data, so a table under one is still a restore point:

- `Until ControlUntil`, either `"resolution"` (Word of Command, the zero value, so existing records read back unchanged) or `"turn"` (Emrakul).
- `TurnSeq int`, the `Turn.Seq` of the controlled turn, set when it begins.
- `ExtraTurn *AnchoredExtraTurn{Ref, Floor}`, the extra turn owed after that turn. `Ref` is minted from `NextExtraRef` when the trigger resolves. `Floor` is `NextExtraRef` at the moment the controlled turn begins.

There are two new stages:

- **`ControlAwaitingTurn`:** created when the trigger resolves. It never applies (`controlLiveLocked` is false).
- **`ControlTurn`:** the controlled turn. Live while both players are in the game and `g.Turn.Seq == TurnSeq`.

`BeginTurnControlForEffect(TurnControl{Controller, Controlled, Source, SourceName})` is what Emrakul's trigger calls. It refuses a player who is not in the game and creates a record that waits. It deletes nothing: several Emrakul records over one player all wait for the same turn, the newest is the one that applies (CR 723.1a, already `ControllerOfPlayerLocked`'s order), and each owes its own extra turn (ruling). `BeginControlledPlayForEffect`'s clean-up narrows to Word records (`Until == "resolution"`), so a Word of Command cast during an Emrakul turn controls the player only while it resolves, and the Emrakul record applies again afterwards. A new event, `EventPlayerControlScheduled`, gives the log line "Alice will control Bob during Bob's next turn (Emrakul, the Promised End)".

### 2. When it begins and ends: the rotation seam

- **Begins.** In `startTurnLocked`, after `noteTurnBegunLocked` and before `onTurnBeganLocked`, every `ControlAwaitingTurn` record over the new active seat becomes `ControlTurn` with `TurnSeq = g.Turn.Seq` and `Floor = g.NextExtraRef`. That is before any step entry hook, so the untap step's choices ("you may choose not to untap") are already the controller's. `EventPlayerControlBegan` gives "Alice controls Bob this turn". The hook runs only when a turn really begins, which gives CR 723.1b and the skip ruling with no extra code. A skipped extra turn never reaches `startTurnLocked` (ADR 0059's #2529 amendment), so the record waits for the next turn that does begin. A record created during the controlled player's own turn waits for their next one, because only a turn that begins later can pick it up.
- **Ends.** First thing in `beginNextTurnLocked`, before `popExtraTurnLocked` chooses the next turn: every `ControlTurn` record with `TurnSeq == g.Turn.Seq` queues its extra turn (§3) and is removed. That is CR 723.1's "doesn't end until the beginning of the next turn", so the cleanup step, and a turn ended early by CR 724.1, are still controlled. `EventPlayerControlEnded` gives the log line.

`prunePlayerControlsLocked` leaves `"turn"` records alone except as §6 says. The rotation seam is the only place they end.

### 3. The extra turn, anchored to "that turn"

`TakeExtraTurnsForEffect` can't be used when the trigger resolves. It would put Bob's extra turn directly after the current turn, which is Alice's, and not after Bob's next turn. Emrakul is the card ADR 0059 Decision 5 was waiting for, and the anchor is added in `extra_turns.go` as `queueAnchoredExtraTurnLocked(seat, source, ref, floor)`. The end hook calls it while the controlled turn is still `g.Turn`. It inserts `ExtraTurn{Ref, Seat, Source}` into the stack directly below every entry whose `Ref > floor`, meaning the turns created during the controlled turn, which are "after this one" too and were created later. Everything else in the queue stays below it. That is CR 500.7 exactly: directly after the specified turn, and most recently created first. Like `TakeExtraTurnsForEffect`, it builds a fresh slice, never appending in place, because of the undo-snapshot aliasing trap documented there. It emits `EventExtraTurnAdded`, so the existing `extra_turn` log line and `TurnView.extra_turns` need nothing new.

It queues nothing for a controlled player who has left the game ("the extra turn never begins", CR 800.4k). It does queue the turn when the controller has left, or when a newer Emrakul overwrote the control: the extra turn is a separate part of the effect, and CR 800.4a ends only the control (§6). Several Emrakul records over the same player each queue their own turn.

### 4. Who decides: routing for a whole turn

`routeControlledPlayer` reads the kind of the live record over the target (a new `ControlUntilOfPlayerLocked`):

- **`"resolution"`:** unchanged from ADR 0166.
- **`"turn"`:** the controller may send, as the controlled player, every rules verb (the action types `internal/legal` enumerates, plus the dock's turn walk): `pass_priority`, `end_turn`, `cast_spell`, `activate_ability`, `activate_loyalty`, `activate_mana_ability`, `special_action`, `declare_attacker`, `declare_attackers`, `resolve_choice` and `discard_selection` (the cleanup discard, CR 514.1). The sandbox verbs (`draw_card`, `move_card`, `play_card`, `tap`, `untap`, `untap_all`, `change_life`, counters, `pass_turn`, `reveal_card`, `remove_from_combat`, `announce_trigger`, `sacrifice_permanent` and the `set_*` table verbs) are refused with `ErrPlayerControlled` for both players. That is CR 723.5b: they are not choices the rules or an object asks for. An admin connection (Caller `uuid.Nil`) keeps them for repairs, as today.

In `dispatch`'s `cast_spell` branch, "`Actor != Caller` is refused unless it names a `controlled_play` prompt" becomes "… unless it names one, or the live record is `"turn"`". In that case `requirePriorityHolder` runs on the controlled seat as it would for them. Costs come from the controlled seat because `castSpellLocked` already pays from `a.Player`, which is CR 723.5a. The controller casts the controlled player's commander from the command zone the same way, tax included. Their commander's "return to the command zone?" choice during that turn belongs to the controller (CR 723.5, a choice the rules call for).

The rest needs no change from ADR 0166:

- **Prompts.** `Chooser` stays the controlled player, and the controller answers.
- **Priority passes** and the turn-based actions they drive (untap and draw are automatic) need no change either. The active player is still Bob (CR 723.3). Alice passes for Bob, and gets priority for herself in turn order as usual (CR 723.8).
- **Combat.** The controller declares Bob's attackers, including attacks on the controller (CR 723.5's example). Goad and other attack requirements are checked against Bob as always. Blocks are declared by the defenders, who are not controlled.

Three neighbours of routing:

- **Undo.** Entries are stamped with the sending connection's seat. Alice's plays as Bob are Alice's to undo, from her own budget, and Bob can't rewind them (`ErrNotYourUndo`). Nothing changes.
- **Standing answers** (ADR 0127). A prompt whose chooser is controlled is never answered automatically: the rules are the controlled player's, and the decision isn't theirs. `autoAnswerRuleLocked` marks it with a new `AskedByHandControlled` reason. The trigger-order preference read is the answerer's (`ControllerOfPlayerLocked(chooser)`).
- **Concede.** As in ADR 0166: Bob may concede, and Alice can't concede for him (CR 723.6).

### 5. Seeing what they see (CR 723.4)

ADR 0166 §3's `isKnower` extension already does this for as long as the record is live. During the turn the controller sees Bob's hand, the faces of his face-down permanents, and what his search and look prompts show (prompt options pass through the same filter). It stops with the turn. Nothing writes the controller into `knowers`, because the rule is about visibility during control. The engine has no sideboard or outside-the-game zone, so the second sentence of 723.4 and the sideboard ruling have nothing to filter today. When a "card from outside the game" prompt is added, it must offer the controller nothing.

### 6. Leaving the game

- **The controller leaves, before or during the turn.** CR 800.4a ends the control: `controlLiveLocked` is already false, and Bob decides for himself from that moment. The record stays, inert, until the turn's end hook, which still queues the extra turn.
- **The controlled player leaves before the turn.** `prunePlayerControlsLocked` removes the record. Their turn never begins, so there is no control and no extra turn.
- **The controlled player leaves during the turn.** The turn finishes without an active player (CR 800.4j). The record is not live, and the end hook queues nothing (ruling, CR 800.4k).

### 7. Legal moves and bots

- **The controller's list** is their own moves plus, while a `"turn"` record over another seat is live, `enumeratorAs(controlled)`'s full run, with each move's `Player` set to the controlled seat. Only one seat holds priority at a time. Two lists can be non-empty together only for prompts, and the `player` field tells them apart. `capLegalMoves` caps per player, so the controller's own moves can't fold Bob's away.
- **A controlled seat** gets nothing (`newEnumerator`, unchanged). For a bot that means it does nothing all turn: the runner never wakes for an empty list, and its improvisations would be refused anyway (§4).
- **A bot controller** (controlling a human or a bot): the runner splits its moves by `Player`. Its own go to its policy as now. Moves for the controlled seat go to `controlledTurnPolicy`, a Filter-style wrapper in `aiseat` (`capability.go`'s pattern) that never calls a model. It passes priority, declares no attackers, and answers each prompt with the enumerator's first answer, or the declining answer when the move is `always_legal`. That is legal, never stalls the table, and never plays Bob's cards against his own interests or in the bot's favour. A policy that uses the turn well (attack Bob's enemies, burn his resources) is a follow-up. The bot's evaluation code assumes "my seat", and teaching it to price another seat's board is real design work.
- **The MCP seat** (ADR 0122) gets the same list, and `boardtext` labels the controlled seat's moves "as Bob".
- **Auto-pass and stops are client-side** (`priority.ts`), so the client answers whose apply: the controller's settings, for whichever of their two seats holds priority (§9). The controlled player's client passes nothing (§9). The server would refuse it anyway.

### 8. Protocol and view

- `PlayerControlView` gains `until` (`"resolution"` | `"turn"`) and `source_name`. `player_controls` keeps listing only live records, so `FilterViewFor`'s knower expansion stays correct.
- New `GameView.scheduled_player_controls`, public: the `ControlAwaitingTurn` records as `{controller, controlled, source_name}`. The turn strip reads "Next: Bob (Alice controls)" from it, and Bob's board warns him ahead of time. It grants no visibility.
- `legal_moves` keeps its shape. For a controller it now holds moves whose `player` is the controlled seat (§7).
- `on_behalf_of` and `chooser` are reused unchanged.
- `docs/protocol.md` documents all three. The snapshot gains `until`, `turnSeq` and `extraTurn` on `playerControls[]`, additive within v8 (`snapshot_shape/v8.txt`, `snapshot_drift_test.go`). An older binary restoring such a file drops the new fields and reads the record as a Word record whose stage it doesn't know. `controlLiveLocked` returns false for it and prune removes it, so the control and the extra turn are lost until the next deploy. That is weaker, never wrong, and the same trade as ADR 0059's 2a amendment, item 6.

### 9. Client: "Playing as Bob"

A new pure module, `lib/playingAs.ts` (unit-tested), computes from the view:

- `controlledByViewer(view, viewerID)`: the seat the viewer controls turn-long, or null.
- `controllerOfViewer(view, viewerID)`: who controls the viewer, or null.
- `actingSeatID(view, viewerID, pref)`: which of the two seats the board acts for.

**The controller's board.** `Game.svelte` derives `actingID = actingSeatID(…)`. When it is the controlled seat, the following read `actingID` where they read `viewerID` today:

- `Board`'s `viewerID` prop. `seatPlacements` then puts Bob at the bottom with his `Hand` in his `PlayerPanel`, face up, because `FilterViewFor` sends Alice his cards.
- `ActionDock`'s priority, pass and end-turn state.
- The combat dock requests and `AttackDeclarationModal`.
- `legalActionsOf(view, actingID)`, which now filters by `player`, so highlights, click to act and the mana stepper act for Bob.
- The `player` field of every `sendAction` they send.

These stay on `viewerID`:

- the settings sync (trigger order, standing answers)
- undo, the game menu and concede
- chat, fanfare and end cues
- table management, spawning, the opening roll and Discord.

`PlayingAsBanner.svelte` sits in the board's attention strip. It says "Playing as Bob, Emrakul, the Promised End" with a **View my board** toggle. Toggled, it says "Viewing your board, Bob's turn is yours" with **Play as Bob**. While Alice plays as Bob, the bottom seat gets an outline in Bob's seat colour, and the dock shows his name and avatar, so she always knows whose cards she is spending.

**The toggle.** Swapping is on as the controlled turn begins. After that, the board switches by itself only at a handoff: when the seat that owes the next decision changes between Alice and Bob (Alice gets priority for herself, or a prompt of her own opens). It never switches while the same seat is still deciding. The manual toggle always works, and the banner offers "Switch to Bob" or "Switch to you" when the other seat owes something. When the record ends, `actingID` falls back to `viewerID`.

**Auto-pass.** Alice's stops and smart autopass apply to whichever of her two seats holds priority. `priority.ts` is called with that seat as "me", so her usual "stop on my main phase" means Bob's main phase while she plays his turn. **End turn** (`end_turn`) works for Bob as the active player.

**The controlled player's board.** Bob's board is his own and unchanged. The banner reads "Alice controls you this turn (Emrakul, the Promised End)". The dock shows a waiting request with no actions, and the game menu keeps **Concede**. Autopass, hold priority and every automatic send stay off while `controllerOfViewer` is set. Prompts he owes show "Alice is choosing for you" (`on_behalf_of`). When `scheduled_player_controls` names him, his strip warns "Alice will control your next turn".

**Spectators** see the banner text only.

`ControlledPlaySheet` (Word of Command) is untouched.

## Out of scope

- **A bot that plays the controlled turn well.** §7's policy is passive on purpose. Follow-up issue.
- **Mindslaver and Sorin Markov's −7** ("control target player during that player's next turn", with no extra turn). This seam covers them with `ExtraTurn` nil. Not requested.
- **Opposition Agent** (CR 723.2, while searching). As ADR 0166 §7.
- **Two-Headed Giant** team control (ruling). Not a format this engine runs.

## Owner decisions (2026-10-13)

1. **Sequencing: ADR now, implementation after #3373.** This ADR lands as its own docs PR. The engine and client work follow on `develop` once #3373 (ADR 0166) has merged, because they extend its record, routing and view.
2. **Client: the "Playing as X" board swap.** During the controlled turn the controller's board gets a banner and a toggle. Their hand area, action dock, attack and block declarations and click to act act for the controlled seat, whose hand they see (CR 723.4). This was chosen over driving the turn from a legal-move sheet like `ControlledPlaySheet`. A whole turn is too many decisions for a list, and the board, dock and combat flow already know how to play a turn.

## Consequences

- Emrakul, the Promised End ships `full`, and #2034 has no manual card left.
- `control-another-player` closes: both waiting cards ship, and Opposition Agent is not requested.
- A whole turn can be played by someone other than its active player. The only trust boundary is still the dispatcher's routing, which reads only the record, and the sandbox verbs are closed to both players for that turn.
- `extra_turns.go` gains an anchored push, the first extra turn not taken "after this one".
- `Game.svelte` has two ids where it had one. Every new call site has to choose between "who I am" (`viewerID`) and "whom I act for" (`actingID`). `lib/playingAs.ts`'s tests and a render test of the swapped board keep that honest.

## Delivery

Two PRs, each shippable, after #3373:

1. **Engine, protocol and bots.** §1 to §8: the record's new fields and stages, the rotation hooks, the anchored extra turn, routing for a whole turn, the standing-answer and trigger-order reads, the controller's legal moves and per-player cap, `controlledTurnPolicy`, the view fields, `docs/protocol.md`, and the snapshot shape. Tests call `BeginTurnControlForEffect` directly. They cover the turn's start and end, CR 723.1a/1b, both leave cases, the extra turn's position against an extra turn created during the controlled turn, two Emrakul effects, a Word of Command inside an Emrakul turn, routing (rules verbs pass, sandbox verbs and the controlled seat refused, concede allowed), undo stamping, a bot controlling a human, a human controlling a bot, and a bot controlling a bot. Emrakul is **not** registered yet, so no table can reach a turn its humans can't play.
2. **Client and the card.** §9, plus Emrakul's `Spec` (the discount, the cast trigger, keywords), its effect key, the roadmap seam closed with its `docs/engine-seams/closed/` fragment, and `docs/adding-cards.md`'s recipe. Emrakul registers `full` here.

The card ships in the client PR, not behind a caveat in the engine PR. Without the swap, a human who cast Emrakul would hold a turn they had no way to play. Bob's seat would sit on priority until someone undid the trigger. A caveat can't fix a stall, and putting the card in the second PR costs nothing.

## Amendment (2026-10-16, #3857) — a control with no extra turn

§1 gave `BeginTurnControlForEffect` one shape: the turn-long record and the
extra turn it owes. Emrakul, the Promised End is the only card with the
second half. "You control target player during that player's next turn"
(Mindslaver, Worst Fears, Sorin Markov's −7, Construct a Cosmic Cube, and
each half of Cruel Entertainment) is the first half alone, and "Out of
scope" said this seam would cover it with `ExtraTurn` nil.

**The change.** `TurnControl` gains `NoExtraTurn bool`. When it is set,
`BeginTurnControlForEffect` mints no `NextExtraRef` and leaves
`PlayerControl.ExtraTurn` nil. Nothing else in the engine moves:

- `beginTurnControlsLocked` already guards `ExtraTurn != nil` before it
  stamps the `Floor`, and `endTurnControlsLocked` already skips the
  queueing for a nil `ExtraTurn`, so the record begins, applies and ends
  exactly as a turn-long record always did (CR 723.1, 723.1a, 723.1b).
- Routing, the view, the snapshot (`extraTurn` is already `omitempty`),
  the client's "Playing as" swap and `controlledTurnPolicy` read the
  record and never the extra turn, so they need nothing.
- Not minting a ref matters: a later control's `Floor` and a later extra
  turn's position (§3) are measured against `NextExtraRef`, and a ref
  burned by a control that owes nothing would be a hole in that
  sequence for no reason. `TestTurnControlWithoutAnExtraTurnOwesNone`
  pins both the missing extra turn and the unspent ref.

**The cards.** They share `cards/effects/turn_control_cards.go`
(`controlTargetPlayerNextTurn` for the one-target cards,
`controlPlayerNextTurn` for Cruel Entertainment's two directions). The
target is "target player" (any player; the engine refuses a player
targeting themselves and one who has left the game), except Construct a
Cosmic Cube, whose reflexive trigger says "target opponent". Cruel
Entertainment is two records, one each way, over two distinct targets;
both records wait for their own player's next turn, so neither player is
controlled for the other's turn, and neither owes an extra turn.

**Left out, filed as seams.** Opposition Agent needs a control scoped to
a search prompt plus an exile-and-play replacement on search results
(#3998); Secret of Bloodbending needs a control for one combat phase and
an optional waterbend additional cost (#3999); The Dominion Bracelet
needs a cost that exiles the granting Equipment from a granted ability
(#4000). The "Out of scope" list above is otherwise unchanged.


## Amendment (2026-10-19, #3999) — a control for one combat phase

Secret of Bloodbending: "You control target opponent during their next
combat phase. If this spell's additional cost was paid, you control that
player during their next turn instead." The paid half is the plain
turn-long control of the 2026-10-16 amendment. The unpaid half is a
control that lasts one combat phase (CR 723.2), which §1's two lengths
could not say.

**The record.** A third length, `Until "combat"`, with two stages of its
own, `ControlAwaitingCombat` and `ControlCombat` (`game/combat_control.go`).
`BeginCombatControlForEffect(CombatControl{…})` creates a waiting record,
refusing what `BeginTurnControlForEffect` refuses, and emits
`EventPlayerControlScheduled` labelled `combat`. No new field: `TurnSeq`
holds the turn the phase is in, so the snapshot shape is unchanged and an
older binary drops such a record the way §8 says it drops one it doesn't
know.

**When it begins and ends.** The step seams, where §2 used the rotation
seam:

- **Begins** as a combat phase of the controlled player's begins: the
  first step of a phase instance that really begins
  (`noteStepBegunLocked` now reports opening a phase) calls
  `beginCombatControlsLocked` before the step is announced, so its
  triggers and choices are already the controller's. An additional
  combat phase is a combat phase. A record made during a combat phase
  waits for the next one, and a turn whose combat phase never begins
  leaves it waiting.
- **Ends** as the cursor leaves that phase instance, in
  `advanceCursorLocked` beside CR 511.3's removal from combat: into the
  next phase, or into an additional combat phase straight after (a new
  `PhaseID`). A turn that ends during the phase (CR 724.1) never reaches
  that seam, so `closeTurnLocked` ends it too. Liveness also requires the
  cursor to be in a combat phase of `TurnSeq` with the controlled player
  active, so a record no seam ended is never live.
- **Leaving the game** follows §6: a waiting record whose controlled
  player left is pruned; a controller who left controls nobody, and the
  record is still spent on the phase.

**Who decides.** The controller plays the phase as §4 plays a turn.
`ControlUntil.PlaysForThem` (turn or combat) is the one predicate the cast
gate, the enumerator's controlled moves, the view's cast stamps, the bots'
split and `TurnControlledByLocked` read, so the "Playing as" swap, the
controller's `legal_moves` and `controlled_legal_actions`, and
`controlledTurnPolicy` all work unchanged. The verb row for
`ControlCombat` is `combatRulesVerbs`: the turn's rules verbs less
`end_turn`, which would walk the controlled player's turn on past the
control, and `discard_selection`, which is never in a combat phase. The
sandbox is closed to both players for the phase.

**Protocol and client.** `player_controls[].until` gains `"combat"`;
`scheduled_player_controls[]` gains `until` (absent reads as `"turn"`); the
three log lines carry `label: "combat"` and read "during Bob's next combat
phase", "this combat". The client's `lib/playingAs.ts` reads
`playsForSeat(until)` where it read `until === "turn"`, words its banner
"combat" for such a control, and hides End turn on the controlled seat's
board (`endTurnClosed`).

**The waterbend.** "As an additional cost to cast this spell, you may
waterbend {10}" is ADR 0073's optional cost with a waterbend clause; see
that ADR's amendment of the same date.

Tests: `game/combat_control_test.go`, `actions/controlled_combat_test.go`,
`actions/controlled_stage_test.go`, `aiseat/controlled_combat_test.go`,
`protocol/combat_control_view_test.go`, `cards/effects/secret_of_bloodbending_test.go`,
`client/src/lib/playingAs.test.ts`.

## Amendment (2026-10-19, #4154) — choosing another player's attackers or blockers

Master Warcraft: "Cast this spell only before attackers are declared.
You choose which creatures attack this turn. You choose which creatures
block this turn and how those creatures block." Odric, Master Tactician
("whenever Odric and at least three other creatures attack, you choose
which creatures block this combat and how those creatures block") and
Brutal Hordechief ("creatures your opponents control block this turn if
able, and you choose how those creatures block") and Melee ("you choose
which creatures block this combat and how those creatures block") say
the same about blocks. §4's ruling already names the card: "If an effect causes another
player to make decisions that the affected player would normally make
(such as Master Warcraft does), that effect takes precedence."

**The rulings** (Scryfall, read 2026-10-19): "the person who cast Master
Warcraft first chooses the complete group of creatures that are going to
attack. Then, for each of those creatures, the active player chooses who
or what it's going to attack"; "You choose attackers and make blocking
assignments regardless of whether it's your turn and regardless of
whether the creatures are attacking you. Your choices must be legal within
the normal rules for attacking and blocking"; "You can decide that a
creature won't block"; and Odric's "If there's a cost associated with
having a creature block and you choose for that creature to block, its
controller can choose to pay that cost or not. If that player decides to
not pay that cost, you must propose a new set of blocking creatures."
CR 508.1b has the active player announce what each attacker attacks, and
CR 508.1g-j make the costs to attack theirs.

**A separate record, projected as a stage.** This is not player control:
nobody controls the active player or the defenders, who keep every other
decision, their mana and what they can see. A new stage of the
`PlayerControls` list would have been wrong in three places at once: a
record per controlled player when one effect covers every player, a
lifetime (the turn) that is not when it applies (one declaration at a
time), and §5's visibility, which a declarer must not get. So the effect
is its own record, `Game.DeclarationControls []DeclarationControl
{Declarer, Source, SourceName, Attackers, Blockers, TurnSeq, ThisCombat}`
(`game/declaration_control.go`), and only while one of its declarations
is being made is it **projected** onto the player-control seam, as stage
`ControlDeclaring` with until `declarations`:

- over the active player in the declare attackers step, while the
  declarer's half of the attack declaration is open;
- over each defending player other than the declarer whose block
  declaration is pending.

`ControllerOfPlayerLocked` asks the projection before any record, which
is the ruling's precedence, and answers with whoever makes the
declarer's decisions (CR 723.5: a player controlling the declarer's turn
makes this choice too). `ControlStageOfPlayerLocked`,
`ControlUntilOfPlayerLocked` and `LivePlayerControlsLocked` follow it, so
routing, the enumerator's controlled moves, the bots' split and the
client's "Playing as" board all work as they do for a control, and
nothing else does: `ControlUntil.PlaysForThem` is false for
`declarations`, so the cast gate, the sandbox closure and the cast stamps
are untouched, and `FilterViewFor` adds no knower for it. The verb row
for `ControlDeclaring` (`actions/controlled_player.go`, `declaringVerbs`)
is `declare_attacker`, `declare_attackers`, `pass_priority`,
`declare_blocker`, `declare_blockers` and `finish_blocks`.

The newest record that applies wins, as CR 723.1a has it for control. A
"this turn" record ends at the rotation seam; a "this combat" one where a
combat-phase control ends (the cursor leaving the phase). A declarer who
has left the game chooses nothing.

**Attacks: two halves** (`Game.AttackHandoff`, cleared with combat):

1. *The declarer's.* The active player holds priority as always in this
   engine's declare attackers step, but the decisions are the declarer's.
   The declarer stages attackers with the ordinary verbs, naming a target
   for each as a proposal; eligibility, restrictions, count limits and
   requirements (CR 508.1c-d) are checked as always. Nothing is paid and
   nothing is exerted: those are costs, and costs are the active
   player's (exert refused with `ErrDeclarerCantExert`). The declarer's
   pass, sent as the active player, is not a pass: it runs the CR 508.1d
   checkpoint and finishes the half (`finishDeclarerAttacksLocked`). With
   no attackers the declaration is over and the active player may add
   none (`ErrAttackersChosen`).
2. *The active player's* (stage `targets`). They decide for themselves
   again: `declare_attacker` on a chosen attacker re-points it and charges
   nothing; on any other creature it is refused. They finish with
   `declare_attackers` naming every chosen attacker (`confirmAttackTargetsLocked`:
   targets, exerts and the whole attack tax at once), or with a pass,
   which accepts the targets as they stand and declines every cost. A
   creature may be left out (or is declined by the pass) only when it has
   no free target; a pass is refused (`ErrAttackCostsUndecided`) while a
   taxed attacker could be pointed somewhere free. A declined creature is
   taken out of the declaration (untapped, `Declined`), can't be chosen
   again this combat, asks no CR 508.1d requirement, and the declaration
   goes back to the declarer: Odric's "propose a new set". Because no
   cost is ever paid before the set is final, a declaration that goes
   back owes nobody a refund, which is the CR 733 reversal this engine
   could not otherwise make.

The lock-in (`commitAttackDeclarationLocked`) waits for both halves, so
the attack triggers name the targets the active player chose. A
sandbox `advance_step` out of an unfinished chosen declaration is refused
(`ErrAttackDeclarationOpen`), and a standing "end the turn by passing"
waits for the declarer and then accepts the targets.

**Blocks.** A defender's declaration is already a turn-based action with
priority parked (ADR 0045's #1501), so the declarer declares and finishes
each pending defender's declaration on their board, with the ordinary
verbs and under every restriction and requirement; their own stays
theirs. The engine has no cost to block, so Odric's third ruling has
nothing to ask yet.

**Moves and bots.** The declarer's enumeration adds, for each seat whose
declaration it makes, that seat's attack moves and a finishing pass
("Done choosing attackers"), or its block moves and `finish_blocks`, each
with `Player` set to the seat (`legal/declaration_control.go`). In the
declarer's half the attack moves carry no tax and no exert twin. In the
active player's half they are offered the re-points, the pass when the
engine takes it, and the confirmation when the targets cost something
they can pay. A bot declarer's moves go to `controlledTurnPolicy`: no
attackers and no blocks, or the required ones, which is legal and, for a
player choosing an opponent's attacks or the blocks against their own
attack, what they want. A bot in the active player's half uses
`attackTargetsPolicy`: accept, else confirm and pay, else a free target.

**Protocol and client.** `player_controls[].until` gains
`"declarations"`; `GameView.declaration_controls` (the records this turn)
and `attack_handoff` (`{stage?, declarer, source_name?, declined?}`) are
new and public. The client's `playsForSeat` includes `declarations`, so
the declarer's board swaps to the chosen seat with that seat's legal
actions and End turn closed, and the block verbs carry the seat. The
banner reads "Choosing Bob's attackers", "Alice is choosing your
blockers", "Alice chooses which creatures attack and block this turn";
the active player's half is a dock request (`lib/attackHandoff.ts`) with
Confirm attacks and, when something costs, Decline the costs; and
autopass holds in both halves (`owesChosenDeclaration`).

**The snapshot** gains `declarationControls` and `attackHandoff`,
additive in v9.

**Melee's other sentence**, "whenever a creature attacks and isn't
blocked this combat, untap it and remove it from combat", is a repeating
delayed trigger with a stated duration (CR 603.7b) on each defender's
completed declaration (`EventBlockersDeclared`, CR 509.3g), with an
each-condition naming every unblocked creature attacking that player, so
each one triggers on its own. Its duration is `UntilEndOfCombat`, and a
delayed trigger with that duration is now dropped as the cursor leaves
its combat phase (`clearEndOfCombatDelayedTriggersLocked`, beside the
combat-phase control's end), so an additional combat is untouched. The
existing sweeps ran only at the turn's end and start.

**Also noticed.** A declarer who leaves the game mid-way chooses nothing
more: the handoff is open from the step's start, and a half whose
declarer is gone is settled lazily (`settleOrphanedHandoffLocked`): what
they staged is taken back, untapped, and the active player declares for
themselves. A declaration with nothing that could attack is over as the
step begins (`beginChosenAttackDeclarationLocked`), so the declarer is
never asked to choose from nothing.

**Why not the issue's reading.** #4154 had the declarer choose each
attacker's defender too. The ruling and CR 508.1b give that to the
active player, and a declarer who could send every attacker at a third
player would make the card stronger than printed.

Tests: `game/declaration_control_test.go`, `actions/declaration_control_test.go`,
`actions/controlled_stage_test.go`, `aiseat/declaration_control_test.go`,
`protocol/declaration_control_view_test.go`,
`cards/effects/declaration_control_cards_test.go`,
`client/src/lib/playingAs.test.ts`, `client/src/lib/attackHandoff.test.ts`.
