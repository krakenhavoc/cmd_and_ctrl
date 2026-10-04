# ADR 0115 — Commanders die (CR 903.9a)

**Status:** Proposed · 2026-10-03 · S58 — Deck requests, October batch (tracker [#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077))
**Issues:** [#2085](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2085) (this change), [#2076](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2076) and [ADR 0114](0114-the-ring-tempts-you.md) owner decision 4 (Sauron, Lord of the Rings, which found it).
**Owner decision behind it (2026-10-03):** fix this faithfully, for every trigger, with this ADR first. Sauron, Lord of the Rings ships before it with a caveat, and this work lifts that caveat.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-03. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 38 remote heads: `origin/develop`, `origin/main`, the `feat/s58-pr3-valgavoth`, `feat/s58-pr4-mishra` and `feat/s58-pr5-jodi` branches, and the other chore, docs, feat, fix, repro and wip branches. The highest number on any of them is 0114, so this ADR takes **0115**.
**Builds on:** [ADR 0013](0013-replacement-effects.md) (replacement effects; §5f, §5g and §5af are the CR 903.9 amendments this one narrows), [ADR 0018](0018-triggers-on-the-stack.md) (the trigger drain), [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator), [ADR 0041](0041-game-persistence.md) (snapshots) and [ADR 0060](0060-leaving-the-game.md) (CR 800.4).

This ADR was written plan-first. No engine, client or card code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The engine applies CR 903.9 to a commander headed for a graveyard or exile as a **replacement effect**. When the owner takes the offer, the commander goes straight to the command zone and never reaches the graveyard. So a commander that is destroyed or sacrificed never **dies**. "Whenever a creature dies" and "put into a graveyard" triggers do not see it, and neither do exile-matters triggers. Sauron, Lord of the Rings ("Whenever a commander an opponent controls dies, the Ring tempts you") found it while ADR 0114 was being written.

### The rules

Every number below was read in the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

- **CR 903.9.** "A commander may return to the command zone during a Commander game."
- **CR 903.9a.** "If a commander is in a graveyard or in exile and that object was put into that zone since the last time state-based actions were checked, its owner may put it into the command zone. This is a state-based action. See rule 704."
- **CR 903.9b.** "If a commander would be put into its owner's hand or library from anywhere, its owner may put it into the command zone instead. This replacement effect may apply more than once to the same event. This is an exception to rule 614.5."
- **CR 903.9c.** A melded or merged commander that uses the 903.9b replacement sends only the commander card to the command zone. It applies to 903.9b only, and the engine has neither meld nor merge.
- **CR 704.6d** repeats 903.9a in the list of variant state-based actions. It sits under **CR 704.6** ("Some variant games include additional state-based actions"), beside 704.6c (21 commander damage). It is not in the CR 704.5 list, and CR 704.3 performs every applicable state-based action, 704.5 and 704.6 alike, "simultaneously as a single event".
- **CR 704.3.** "Whenever a player would get priority … the game checks for any of the listed conditions for state-based actions, then performs all applicable state-based actions simultaneously as a single event. If any state-based actions are performed as a result of a check, the check is repeated; otherwise all triggered abilities that are waiting to be put on the stack are put on the stack, then the check is repeated." The same process runs in the cleanup step, where priority is given only if something was performed or a trigger is waiting (CR 514.3a).
- **CR 704.4.** State-based actions "pay no attention to what happens during the resolution of a spell or ability". A commander exiled and returned within one resolution (a blink) is never seen by 903.9a.
- **CR 700.4.** "The term dies means 'is put into a graveyard from the battlefield.'" Under 903.9a a destroyed commander is put into the graveyard, so it dies, whatever its owner does next. The Sauron ruling says so in words: it triggers "even if the owner of the commander that died chooses to return it to the command zone after it dies".
- **CR 603.10a.** Leaves-the-battlefield abilities "look back in time", so a dies trigger sees the commander as it last existed on the battlefield, controller included.
- **CR 603.6c.** "An ability that attempts to do something to the card that left the battlefield checks for it only in the first zone that it went to." A "when this dies, return it to its owner's hand" trigger on a commander finds it in the graveyard only if its owner leaves it there.
- **CR 400.7.** "An object that moves from one zone to another becomes a new object with no memory of, or relation to, its previous existence." The command-zone move out of a graveyard makes a new object as well, so a trigger that targeted the card in the graveyard loses it.
- **CR 903.3.** The commander designation "is not a characteristic of the object represented by the card; rather, it is an attribute of the card itself. The card retains this designation even when it changes zones." Its example: a face-down commander and a commander copying another card are still commanders, and "a permanent that's copying a commander … is not a commander". So the 903.9a check follows the card through every zone, face down included, and never applies to a token or a copy.
- **CR 903.8.** Commander tax counts "each previous time the player casting it has cast it from the command zone that game". **CR 903.10a.** A player loses after "21 or more combat damage by the same commander". By 903.3, "it" and "the same commander" name the card, so neither number resets when the card becomes a new object.

### What the engine does today

| Piece | On `origin/develop` at `bed5c73c` |
|---|---|
| The rule | `commanderZoneReplacement` (`server/internal/game/builtin_replacements.go:33-98`). It is an optional built-in that watches `EventZoneMove` and `EventDiscardCard`, applies to `ZoneGraveyard`, `ZoneExile`, `ZoneHand` and `ZoneLibrary`, and rewrites `NewZone` to `ZoneCommand`. It is installed for every game at `game.go:930`. |
| Where moves open the window | Effect exits go through `routeCardToZoneLocked` (`zone_route.go:482`, #529). Destroy, sacrifice and state-based exits go through `routeBattlefieldExitInBatchThenLocked` (`mutations.go:4973`) and land in `executeBattlefieldLeaveLocked` (`mutations.go:5100`). Its `case ZoneCommand` branch is the commander's landing. |
| How the owner is asked | Always a prompt. `queueOptionalReplacementPromptLocked` (`pending_choice.go:1660`) queues `PendingChoiceOptionalReplacement` with the question "Send commander to command zone instead?". `ResolveOptionalReplacement` (`pending_choice.go:1748`) answers it. The move **pauses** until the owner answers. There is no setting: `TableSettings` (`settings.go:73`) and the client's gameplay settings (`client/src/lib/settings.ts:139-248`) have nothing for it. |
| Costs | A commander moved to pay a cost is asked **before** the payment (#1397, ADR 0013 §5af): `askCostCommanderLocked` (`cost_commander_choice.go:227`) parks the announcement and replays it with the answer (`commanderZoneAnswer`, carried as `zoneRoute.commanderAnswer`). |
| What the pause costs | A paused exit leaves the card where it was until the answer. Because of that, a long list of guards exists. `zoneChangePausedLocked` (`pending_choice.go:4360`) is read by the doomed-permanent and Saga passes (`mutations.go:4185`, `:4307`), by the world rule (`entry_ordinal.go:139`), by `refusePausedCostCardsLocked` (`cost_commander_choice.go:181`, #1445, #1427, #1474) and its six callers, by the autotapper (`autotap.go:685`, `:831`), by `materializePlanLocked`, by random discard and by delve. A wrath that catches a commander takes it out "a beat late", outside the simultaneous batch (the `zone_route.go` header). |
| Who sees a death | No `EventDies` exists. A death is an `EventLTB` with `NewZone == ZoneGraveyard` (`events.go:371`), stamped with last-known information by `battlefieldExitLocked` (`battlefield_exit.go:41`). The catalog's helpers are `cardDied` (`effects/helpers.go:37`), `diedCreature` (`:360`), `permanentYouControlWasPutIntoAGraveyard` (`triggers_common.go:963`) and about twenty `b*Died` helpers. The engine-side readers are the turn tally (`turn_tally.go:934`) and the trigger doubler (`trigger_doubling.go:162`). Every one of them reads `NewZone`, which is `ZoneCommand` for a commander that took the offer. Exile and "from anywhere" triggers (`otherCardPutIntoExile`, `ranarSawAnExile`, `creatureCardPutIntoAGraveyard` and others) read the same field. |
| Commander tax and damage | `Player.CommanderCasts` (`player.go:257-268`, bumped at `mutations.go:1646`, read at `:2782`) and `Player.CommanderDamage` (`player.go:124-134`, written by `damage_tail.go:277`) are both keyed by the commander's **instance ID**. `MoveCard` (`zone.go:168`) keeps the ID and bumps `ObjectEpoch`, so a graveyard, exile or command-zone trip keeps the tally. A return from exile to the battlefield mints a new ID (`resetAsNewObjectLocked`, `entry_tail.go:284`), and both tallies reset. The comment at `player.go:124` calls that reset correct. By CR 903.3 it is not (see [decision 6](#6-commander-tax-and-commander-damage-follow-the-card)). |
| The state-based action loop | `runStateChecksLocked` (`mutations.go:4377`) runs `stateBasedActionsLocked` (`mutations.go:4020`) and then drains triggers, repeating until quiet. The pass is ordered: phase-in locks, layers, attachments (704.5m/n), counters (704.5q), player losses (704.5a–c, 704.6c), the doomed sweep (704.5f–i, 704.5k, battles) through `sweepDoomedPermanentsLocked` (`simultaneous.go:794`), Sagas (704.5s), the legend rule (704.5j, `legend_rule.go:132`), tokens (704.5d) and prepare copies. The legend rule is the precedent for a state-based action that asks a player: it queues a data-only prompt and the loop runs again after the answer. `holdForOpenResolutionLocked` (`resolution_pause.go:108`, #1289) is the precedent for holding the whole boundary. |
| Snapshots | Schema v7 (`snapshot.go:208`). The CR 903.9 prompt holds a `replacementResume` frame, which the census counts (`snapshot.go:2224`), so **a table waiting on the commander prompt is never a restore point today**. A legend-rule prompt is plain data and is one. |
| The wire and the client | `PendingChoiceView` (`protocol/view.go:304`) with `kind: "optional_replacement"`, answered by `resolve_choice` `{choice_id, apply}` (`actions/actions.go:1516`, `:1850`). The client renders it in `ChoicePromptModal.svelte:485` with plain Yes/No from `choiceDock.ts` (`copyFor` at `:134`, `answersFor` at `:253`). Nothing is commander-specific, and the battlefield path does not even set `Source`, so the card is not shown. The context menu has "Graveyard → command zone (CR 903.9)" (`contextMenu.logic.ts:1333-1351`), which sends `move_card` with `as_commander`. No e2e spec or tutorial step mentions any of it. |
| The bot | `legal/choices.go:196` offers yes and no. The heuristic scores yes 1 and no 0.5 (`aiseat/heuristic/choices.go:423`), so a bot **always** takes the command zone, including for an adventurer or escape commander it could cast from where it went. |

### Cards that carry the deviation

- **Caveats:** Rest in Peace ("A dying commander whose owner sends it to the command zone goes there instead of being exiled.") and Leyline of the Void (the same sentence for an opponent's commander). Sauron, Lord of the Rings is not on `develop` yet. ADR 0114 PR 3 ships it with "Doesn't trigger when an opponent's commander dies and its owner moves it to the command zone".
- **Comments that describe today's engine as the printed behaviour:** Agent of the Iron Throne ("never reaches a graveyard, so that death does not drain"), Phyrexian Rebirth (not counted as destroyed), The Battle of Bywater and `b17FoodPerCreatureYouControl`, God-Eternal Oketra, The Locust God, Underworld Cerberus and `b15ReturnListedCardsFromGraveyardToHand`.
- **Cards that already reason from 903.9a:** Liesa, Forgotten Archangel, Stone of Erech and Cosmic Intervention. Each reaches both printed outcomes by letting the owner order its replacement against the built-in (CR 616). The outcomes are right; the ordering prompt is an artifact of the built-in and goes away.

---

## Decision

### 1. Graveyard and exile go through the state-based action; hand and library stay a replacement

`commanderZoneReplacement` keeps its name and its prompt, and its `AppliesTo` narrows to `ZoneHand` and `ZoneLibrary` (CR 903.9b). That includes "may apply more than once to the same event", which the gather already allows. A commander headed for a graveyard or exile is moved there like any other card: it is destroyed, sacrificed, countered, discarded, milled or exiled, and every event and trigger of that move fires as printed.

Then a new step of `stateBasedActionsLocked`, `commanderReturnSBALocked` (new file `server/internal/game/commander_return.go`), applies CR 903.9a.

**"Put into that zone since the last time state-based actions were checked"** is a per-card mark: `Card.CommanderReturnDue bool`.

- **Set** when a card with `IsCommander` lands in a graveyard or in exile. `MoveCard` (`zone.go:168`) sets it, because every route ends there and it already knows the destination zone's kind, so no mover has to remember it.
- **Cleared** by the same `MoveCard` when the card leaves a graveyard or exile, whatever moves it. A commander exiled and returned within one resolution is never offered (CR 704.4).
- **Cleared** when the check below has looked at it.

**The check** runs **first** in the pass. It collects every marked card before anything else in the pass moves, so it reads the same board as the rest of the pass (CR 704.3). For each one:

- If its owner has left the game, nothing happens (CR 800.4a already took the card out).
- If the owner has an answer standing, that answer is applied at once (see [decision 2](#2-the-may-is-the-owners-every-time)).
- Otherwise a `PendingChoiceCommanderReturn` prompt (`"commander_return"`) is queued to the **owner**, not the controller, with `Source` set to the card. The prompt is plain data, like the legend rule's.

A commander put into a graveyard by a later step of the **same** pass (lethal damage, the legend rule) keeps its mark and is offered on the next pass. CR 704.3 runs that pass at once, because a pass that performed anything repeats.

**A "yes"** moves the card from its graveyard or exile to its owner's command zone with a plain `MoveCard`. That is a new object (CR 400.7), it emits `EventZoneMove` with `OldZone` set to graveyard or exile, and a "leaves your graveyard" trigger sees it (CR 603.10a). **A "no"** clears the mark and leaves the card where it is. It is not asked again unless it moves into a graveyard or exile again (a commander declined in a graveyard and later exiled by Bojuka Bog is offered again).

### 2. The "may" is the owner's, every time

The prompt goes to the owner for every commander the check finds, and the engine keeps no hidden default. Whether the owner may also set a standing answer is [open question 1](#1-a-prompt-every-time-or-a-standing-answer). The recommended answer keeps that standing answer in the client, which answers the prompt for the player, so the server's rule stays exact and nothing new is stored on a seat.

The prompt carries one server-computed fact, `playable_from_zone`: whether the owner could cast or play the card from where it now is. That covers a cast permission (`CastPermissionForLocked`, `cast_permission.go:882`) or a keyword that casts from that zone (escape, flashback, foretell, plot, an adventurer "on an adventure"). It is the zone half of `castableNow` (`protocol/view.go:5176`), ignoring timing and mana. The client shows it ("You could cast it from exile") and the bot reads it ([decision 4](#4-the-bot)).

The same prompt also serves the 903.9b replacement, through the existing `optional_replacement` path, once that path sets `Source` for the battlefield exit as the cost path already does (`cost_commander_choice.go:240`). The client then names the card in both.

### 3. Timing: the check finishes before anything else

CR 704.3 performs the state-based actions, the 903.9a choice included, before any waiting trigger goes on the stack. The engine has to match that, because the difference is visible. A "return target creature card from a graveyard" trigger must choose its target after the commander has gone, and a commander's own "when this dies, return it" trigger finds nothing once its owner has moved it (CR 603.6c).

So while any `commander_return` prompt is open, `runStateChecksLocked` stops after the pass that queued it: no trigger drain and no further pass. That is the `holdForOpenResolutionLocked` shape (#1289), and a new `holdForCommanderReturnLocked` is checked at the top of the loop and after each pass. Every other state-based action collected in that pass has already been performed. `ResolveCommanderReturn` applies the answer and calls `runStateChecksLocked` again, which repeats the check (the move was a state-based action performed) and then drains the triggers.

Several commanders (partners, or a wrath across the table) queue one prompt each, all at once. Each owner answers independently, and the hold lasts until the last answer. Nothing else can happen in between, so the moves are part of one check as far as anything can observe.

In the cleanup step, a "yes" counts as a state-based action performed for CR 514.3a, so players get priority. A "no" changes nothing on the board and does not count.

### 4. The bot

`legal` enumerates the new kind as yes and no, the same two moves it lists for `optional_replacement`. The heuristic answers yes, **unless** `playable_from_zone` is set, in which case it answers no. That way a bot does not send an adventurer, foretold or escape commander home when it could cast it from where it went. The model prompt names the card and the flag. Whether the bot should do more is [open question 2](#2-how-smart-should-the-bots-answer-be).

### 5. What goes away, and what stays

- **Goes:** the pause for graveyard and exile. Destroy, sacrifice, discard, mill, counter and exile no longer wait on the owner. The wrath batch takes a commander out with everything else, inside the simultaneous exit, so a Blood Artist sees every death in one batch.
- **Goes:** the CR 616 ordering prompt between the built-in and Rest in Peace, Leyline of the Void, Liesa, Stone of Erech, Cosmic Intervention and regeneration. There is no second replacement to order any more: the card's own replacement applies, and the state-based action offers the command zone afterwards.
- **Narrows:** #1397's ask-before-paying (`askCostCommanderLocked`) to cost moves whose destination is a hand or a library (ninjutsu's return, "return a creature you control to its owner's hand"). A sacrificed, discarded or exiled cost card is paid like any other and offered afterwards. The guards in `refusePausedCostCardsLocked` and the autotapper are unchanged. They key off an open prompt, so they now fire only for a bounce or a tuck.
- **Stays:** `zoneRoute`'s pause and resume for hand and library, and `MustSettleNow` for those destinations.
- **Changes meaning:** the sandbox `move_card` `as_commander` flag. A move to a graveyard or exile lands there and the state-based action asks the owner. With `as_commander`, the context-menu item "Graveyard → command zone" pre-answers yes: the commander dies, then goes home, all in one click. That is the admin override it always was.

### 6. Commander tax and commander damage follow the card

By CR 903.3, 903.8 and 903.10a, both tallies belong to the commander card, across every zone change. Today they follow the instance ID, which survives every move except a return from exile to the battlefield. Under decision 1 a blinked commander is never offered the command zone (CR 704.4), so every blink would reset its tax and start a fresh 21-damage clock.

The fix is the smallest one that holds. When `resetAsNewObjectLocked` mints a new ID for a card with `IsCommander`, it moves the owner's `CommanderCasts` entry and every player's `CommanderDamage` entry from the old ID to the new one. Nothing new is stored and the wire keys are unchanged, and the comment at `player.go:124` is corrected. It lands first ([PR 1](#delivery)), because PR 3 makes it common.

### 7. Tokens, copies, face-down commanders

- **A token** is never a commander (CR 903.3 names a card). **A copy** of a commander, whether a spell copy (`spell_copy.go:362` clears `IsCommander`), a token copy or a Clone, is not one either. None of them is marked or offered. A commander **copying** something else is still a commander and is offered.
- **A face-down commander** (Ixidron, a manifested commander, a foretold commander in exile) is still a commander (CR 903.3's example). It is marked and offered like any other. The prompt names the card to its owner only, which reveals nothing: a commander's identity is public from the start of the game (CR 903.6), and the command zone it would go to is public too. Other seats see "Alice is deciding about her commander".
- **Merged or melded** commanders (CR 903.3b–c, 903.9c) do not exist in the engine and are out of scope.

### 8. Snapshots, undo and restore

- `Card.CommanderReturnDue` is a new card field: `cardSnapshot.CommanderReturnDue bool` with JSON `commanderReturnDue,omitempty`. It is additive, so it stays within schema **v7**. The shape file is regenerated with `-update-shape`, the drift test and `clone.go` (a value copy) carry it, and no fixture is edited.
- `PendingChoiceCommanderReturn` is data only: `Chooser`, `Source`, `Reason`, no resume frame. A table waiting on it **is a restore point**, which today's prompt never is. The restore re-reads `playable_from_zone`, which is computed, not stored.
- New v7 corpus boards, added with `-write-corpus`, which writes new boards beside the existing ones: "a commander in a graveyard with the check due" and "a commander_return prompt open". Both are restored by `TestSnapshotCorpusRestores`.
- **Rollback.** A binary that does not know `commander_return` would restore an open prompt it cannot answer. So [PR 2](#delivery) ships the kind, its resolver, the field and the hold, with the step switched off, **one deploy before** [PR 3](#delivery) switches it on. A rollback by one deploy therefore lands on a binary that can answer the prompt. PR 2 also makes restore refuse a pending-choice kind the binary does not know, with `ErrUnknownEffectKey`, keeping the file, which is the rule effect keys already follow. That closes the gap for every later kind too.
- Undo inside an open prompt rewinds to the board with the commander in the graveyard and the prompt open, which is the board the owner saw.

### 9. Which cards change

- **Caveats that come off:** Rest in Peace, Leyline of the Void's second caveat, and Sauron, Lord of the Rings's caveat. If Sauron lands before PR 4, PR 4 lifts its caveat. If it lands after PR 3, it ships without one.
- **Comments rewritten, behaviour now as printed:** Agent of the Iron Throne (a commander's death drains), Phyrexian Rebirth (a destroyed commander counts), The Battle of Bywater, `b17FoodPerCreatureYouControl`, God-Eternal Oketra, The Locust God, Underworld Cerberus, `b15ReturnListedCardsFromGraveyardToHand`, Liesa, Stone of Erech, Cosmic Intervention.
- **Unchanged:** Sanctum of Eternity and every bounce or tuck (CR 903.9b), and Command Beacon (it moves a commander out of the command zone).
- Every "dies", "put into a graveyard" and exile-matters card in the catalog now sees commanders without a change of its own, because each one reads `EventLTB` and `EventZoneMove` and those now name the graveyard or exile.

### 10. Which tests change

About forty test files answer the CR 903.9 prompt today. The largest are:

- `game/commander_zone_routes_test.go` (11 of 15 tests), `game/discard_commander_test.go` (6 of 7), `game/sandbox_move_resume_test.go` (6 of 8);
- `effects/exit_payout_cards_test.go` (7 of 9), `effects/exiled_this_way_cards_test.go` (7 of 8), `effects/paused_tuck_continuations_test.go` (7 of 7), `effects/paused_exile_continuations_test.go` (5 of 5);
- four each in `game/exiled_this_way`, `paused_tuck_continuation`, `cost_commander_choice`, `milled_this_way`, `stale_commander_prompt` and `effects/milled_this_way_cards`, `exile_payout_cards`;
- about thirty files with one to three each.

The graveyard and exile cases are rewritten to the new order: the card lands, the trigger fires, then `commander_return` is answered. The hand and library cases keep their shape. Tests whose only subject was the graveyard or exile **pause** (paused exile continuations, stale prompts, paused exit costs and casts for a destroy or an exile) are deleted with the pause they tested. Each deletion is named in its PR, and the bounce and tuck versions stay.

New tests, at minimum:

- a destroyed commander fires a dies trigger and a Sauron-shaped "commander an opponent controls dies" trigger, whatever its owner answers;
- a wrath takes a commander in the same simultaneous batch;
- countered, discarded (Megrim), milled and exiled commanders land first and are offered after;
- Cloudshift on a commander asks nothing and keeps its tax and commander damage (PR 1);
- an adventurer commander gets `playable_from_zone` and the bot declines;
- Rest in Peace exiles a dying commander and then offers it, with no ordering prompt;
- a "target creature card in a graveyard" trigger chooses its target only after the answer;
- a commander's own "when this dies, return it to hand" finds nothing after a "yes";
- a decline is not asked again on the next pass, and a later graveyard-to-exile move asks again;
- the owner, not the controller, is asked for a stolen commander;
- CR 514.3a in cleanup;
- the restore of both new corpus boards;
- 903.9b bounce and tuck still prompt, and still apply more than once to one event.

---

## Delivery

Each PR targets `develop`, references S58 and #2085, and is green on its own.

| PR | What | Behaviour change |
|---|---|---|
| 1 | Commander tax and damage follow the card across a new object ([decision 6](#6-commander-tax-and-commander-damage-follow-the-card)); `player.go` comment fixed; blink tests. | A blinked commander keeps its tax and its damage. |
| 2 | Dormant plumbing: `Card.CommanderReturnDue` (snapshot, shape, clone), `PendingChoiceCommanderReturn` with `ResolveCommanderReturn` and the `runStateChecksLocked` hold, the wire view with `playable_from_zone`, `legal` and the heuristic, the client's copy and Yes/No for the kind, restore refusing unknown choice kinds, corpus boards, `docs/protocol.md`. The step is behind an unexported switch that is off. | None. |
| 3 | The switch: the step is on, `commanderZoneReplacement` narrows to hand and library, #1397 narrows to hand and library costs, `as_commander` pre-answers, the graveyard and exile tests are rewritten, the pause-only tests are deleted, the `zone_route.go` header and ADR 0013 §5f/§5af get an amendment pointer to this ADR. | Commanders die. |
| 4 | Cards: the caveats on Rest in Peace, Leyline of the Void and (if landed) Sauron, Lord of the Rings come off; the [decision 9](#9-which-cards-change) comments are rewritten; the oracle fixtures of those cards only are regenerated; `docs/adding-cards.md` gets one line ("a commander dies like any creature; CR 903.9a runs afterwards"). | Caveats lifted. |
| 5 | The owner's answer to [open question 1](#1-a-prompt-every-time-or-a-standing-answer) (if it adds a setting), and the prompt polish: the card is shown in both prompts, with the `playable_from_zone` line. | UX only. |

PR 2 must deploy before PR 3 ([decision 8](#8-snapshots-undo-and-restore)). PR 1 and PR 2 are independent of each other.

---

## Consequences

- Every catalog trigger that watches deaths, graveyards or exile sees commanders, with no per-card work.
- The commander rule stops pausing graveyard and exile moves. The largest source of mid-resolution pauses in the engine goes away, and so do the CR 616 ordering prompts it caused. A wrath is one simultaneous event again.
- A table waiting on the commander's owner becomes a restore point.
- One new prompt kind, one card field, and one hold in the state-based action loop.
- The owner sees the commander in the graveyard (or exile) for the moment it takes to answer. That is the paper experience, and it is what makes the dies triggers legible.

## Alternatives considered

- **Keep the replacement and emit a synthetic "dies" event for the command-zone move.** Rejected. The card never reaches the graveyard, so a trigger that acts on it there (CR 603.6c), Rest in Peace's exile, "put into a graveyard from anywhere" and the simultaneous batch would all still be wrong, each needing its own patch.
- **Graveyard only, with exile kept as a replacement.** Rejected. CR 903.9a names both zones, and exile-matters triggers have the same gap.
- **Do the check at the end of the pass instead of first.** Rejected. Collecting at the start reads the same board as the rest of the pass, and a commander killed in the pass is offered on the very next pass anyway.
- **Bump the snapshot schema to v8.** Rejected. Both additions are additive, new choice kinds have landed within v7 before, and shipping PR 2 a deploy before PR 3 covers the rollback that a bump would cover.
- **Key the tallies by a new stable card ID instead of re-keying.** Rejected for now. It is a new field on every card and in every snapshot. Re-keying at the one place that mints a new ID is enough, and a test pins it.

## Out of scope

- Meld and merge (CR 903.3b–c, 903.9c), which the engine does not have.
- A deeper bot policy ([open question 2](#2-how-smart-should-the-bots-answer-be), option C).
- Brawl (CR 903.12) beyond what already shares the Commander rules.

---

## Open questions for the owner

### 1. A prompt every time, or a standing answer?

CR 903.9a and 903.9b leave the choice to the owner each time. The issue asks whether the engine should ask or remember.

- **A. Ask every time.** No setting at all. The most prompts.
- **B. Ask, plus an opt-in standing answer (recommended).** Add a per-person setting, "Always send my commander to the command zone", synced with the account settings (ADR 0110 §4) and off by default. When it is on, the client answers both prompts (903.9a and 903.9b) with yes for that person. It still shows the prompt when `playable_from_zone` is set, because "always" never means "even when I could cast it from here". The server and the rule stay exact: the player's own client gives the player's own answer.
- **C. A server-side preference, on by default.** The engine moves the commander without asking unless the player turns it off. Fewest prompts, but the engine would be choosing for players who never opened the setting, and the state would live on the seat.

### 2. How smart should the bot's answer be?

- **A. Always yes**, as today.
- **B. Yes, unless `playable_from_zone` (recommended).** One flag the server already computes for the client, so the rule is a single line.
- **C. B, and also no when one of its own waiting triggers would use the card where it is**, such as the commander's own "when this dies, return it". It needs the bot to read the trigger queue for the card, which nothing in `aiseat` does today. Better play, larger change, and it can follow later.
