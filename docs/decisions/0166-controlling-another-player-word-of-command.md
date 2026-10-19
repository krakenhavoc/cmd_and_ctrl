# ADR 0166 — Controlling another player: Word of Command

**Status:** Accepted · 2026-10-13 · Backlog.
**Issues:** [#2164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2164) (the `control-another-player` seam). Requested card: Word of Command ([#2032](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2032), the Valgavoth deck, whose last manual card it is).
**Owner decisions behind it (2026-10-13):** the forced play is cast inside Word of Command's resolution, not offered as a later window, and it can't be declined while it can be played; the seam is built for Word of Command now, with Emrakul, the Promised End's turn-long control left for a follow-up. See [Owner decisions](#owner-decisions-2026-10-13).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-13. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 15 remote heads. The highest number anywhere is 0165, so this ADR takes **0166**. (It was drafted as 0164, which the rematch ADR took first.)
**Builds on:** [ADR 0010](0010-card-effect-catalog.md) §10 (the `PendingChoices` queue), [ADR 0018](0018-triggers-on-the-stack.md) §6 (prompts that block the table), [ADR 0033](0033-ai-bot-seat.md) (the legal-move enumerator), [ADR 0041](0041-game-persistence.md) (restore points), [ADR 0060](0060-leaving-the-game.md) (CR 800.4g), [ADR 0066](0066-granted-cast-and-play-permissions.md) (casting during a resolution) and [ADR 0116](0116-a-card-filter-on-the-revealed-hand-pick.md) (the revealed-hand pick).

---

## Context

Word of Command ({B}{B}, instant):

> Look at target opponent's hand and choose a card from it. You control that player until Word of Command finishes resolving. The player plays that card if able. While doing so, the player can activate mana abilities only if they're from lands that player controls and only if mana they produce is spent to activate other mana abilities of lands the player controls and/or to play that card. If the chosen card is cast as a spell, you control the player while that spell is resolving.

Every prompt, priority pass and action in `game/` is answered by the seat it belongs to. Nothing records that one player controls another, and nothing lets one seat act with another seat's resources. That is the whole of #2164's gap for this card.

### The rules

Checked against `MagicCompRules 20260925.txt`. Rulings are Scryfall's, read on 2026-10-13.

- **CR 723.2:** Word of Command and Opposition Agent control a player "for a limited duration". CR 723.1's turn-long form is Emrakul's.
- **CR 723.3:** only control of the player changes. Every object keeps its controller. The forced spell is the controlled player's spell.
- **CR 723.4:** what the controlled player could see, the controller sees too.
- **CR 723.5 / 723.5a:** the controller makes every choice and decision the controlled player is allowed to make or told to make, and pays that player's costs with that player's resources only.
- **CR 723.6:** the controlled player may still concede; the controller can't make them.
- **CR 723.8:** the controller keeps making their own decisions.
- **CR 608.2g:** a spell an effect tells a player to cast during resolution is cast by CR 601.2a–i, "except no player receives priority after it's cast". It becomes the topmost object and the resolving spell carries on.
- **CR 305.2b / 305.3:** a land can't be played, for any reason, without a land play left, or on another player's turn. Such an instruction is ignored.
- **Rulings:** "You must order your opponent to play the chosen card if it is possible to do so." "You do get to choose which lands get tapped." "If there is a non-mana way to cast a spell, as with Pitch Spells, you may require that way to be used." "You may Command your opponent to play a land (if they have not already done so this turn)." "To 'play a card' is to either cast a spell or to put a land onto the battlefield using the main phase special action."

### What exists

- **A resolution that waits on its own prompt.** A prompt queued while a resolution is open is stamped `midResolution` (`resolution_pause.go`, #1289). State-based actions and the trigger drain wait for it, and a blocking prompt stops every pass. That is CR 608.2g's frame: the cast can happen in the middle of Word's resolution and nobody gets priority in between.
- **The revealed-hand pick** (ADR 0116) chooses from another player's hand, with keyed continuations that survive a restore.
- **Actions name the seat they act for.** `actions.Action` has `Player` (the payload's seat) and `Caller` (the authenticated seat). A seated caller may not act for another seat (`ErrPlayerCallerMismatch`). Legal moves carry `Player` too.
- **Auto-tap exclusions** (`CastAutoTapExclusions`) are the one list both the engine's cast and the enumerator plan mana from.

## Decision

### 1. The control record

`Game.PlayerControls []PlayerControl` records each live player-controlling effect: `Controller`, `Controlled`, `Source` (Word of Command), the chosen `Card`, and the stage:

- **`ControlPlaying`** — from the pick until the chosen card is played or found unplayable. Word of Command is still resolving.
- **`ControlSpellResolving`** — the chosen card was cast. Control is dormant while the spell waits on the stack (the table gets priority as usual, and the controlled player answers for themselves). It applies again from the moment that spell begins to resolve until its resolution closes, then the record is removed.

`Game.ControllerOfPlayerLocked(p)` answers "who decides for p right now": p itself unless a record applies. The record is carried by restore points (ADR 0041) and in the snapshot. CR 723.1a's "last one wins" is the record list's order; with only Word of Command it never stacks.

### 2. Who decides

- **Actions.** In the dispatcher, a caller who controls `a.Player` acts as `a.Player` for every per-seat gate (priority holder, card controller, player-scoped verbs). A controlled player's own actions are refused with `ErrPlayerControlled`, except `concede` (CR 723.6). During `ControlPlaying` the only actions the controller may take as the controlled player are the forced play itself and the mana abilities of that player's lands (§4).
- **Prompts.** A prompt keeps its real `Chooser` (the controlled player), because every resolver reads the chooser as "whose permanents, whose hand". `resolve_choice` accepts an answer from the player who controls the chooser. The snapshot shows the prompt to the controller as theirs and to the controlled player as waiting on the controller, so the client needs no second routing.
- **Bots.** A bot that controls a player is offered that player's prompts and forced play as its own moves, with `Player` set to the controlled seat. A controlled bot is offered nothing.

### 3. Seeing what they see (CR 723.4)

"Look at target opponent's hand" makes Word's controller a knower of each card in that hand. It is a look, not a reveal: nobody else learns the hand (CR 701.20e). While a record applies, the controller is also shown every card the controlled player knows of. Cards outside the game stay private (CR 723.4's second sentence).

### 4. The forced play

Word of Command's resolution is a revealed-hand pick with two new variant fields: `Look` (only the chooser sees the hand) and `PickKeep` (the chosen card stays in the hand). Its continuation creates the record and queues a **`controlled_play`** prompt for Word's controller, naming the controlled player and the card. The prompt blocks the table and is stamped `midResolution`, so the resolution stays open until it is answered.

It has two answers:

- **Play it.** A `cast_spell` sent by the controller with `player` set to the controlled player and `controlled_play` set to the prompt's ID. The engine casts it by CR 608.2g: timing restrictions from the card's type are ignored (the cast happens during a resolution), its other restrictions are not, and the controller chooses everything (targets, modes, X, alternative or additional costs, which lands to tap). Mana may come only from the mana abilities of the controlled player's lands and from mana already in that player's pool. Every other source of theirs is excluded from auto-tap, and activating one is refused. The play is always paid for, whatever the table's strict-mana setting. The controller chooses which lands pay ("You do get to choose which lands get tapped") by picking them in the prompt; the rest are sent as `locked_sources`, so the auto-tapper leaves them alone. A **land** is played instead of cast, and only on the controlled player's turn with a land play left (CR 305.2b, 305.3). No player receives priority after the play. The prompt closes, Word finishes resolving, and the record moves to `ControlSpellResolving` (a spell) or is removed (a land).
- **It can't be played.** Accepted only when the card has no legal play under those restrictions. The engine asks the legal-move enumerator, through a probe `internal/legal` registers with `game` at init, so the answer the server accepts is the answer the bot sees. The record is removed and the resolution ends.

The prompt has no decline. While the card can be played, it must be (ruling).

### 5. While the spell resolves

The record is `ControlSpellResolving`. Every prompt the controlled player owes while that spell resolves is the controller's to answer (§2). The record goes when that resolution closes, or when the spell leaves the stack without resolving (countered, exiled).

### 6. Leaving the game

- **The controller leaves:** CR 800.4a ends the effect. The record is removed. An open `controlled_play` prompt is dropped and Word's resolution ends with nothing played.
- **The controlled player leaves:** the record is removed and the prompt is dropped (the card went with them).
- `controlled_play` gets the "drop" row in `choiceDepartureDecisions`, a "blocks" row in `choiceGateDecisions`, and `choiceMoves` offers the play moves, or the "can't be played" answer when there are none.

### 7. Out of scope

- **Emrakul, the Promised End** (CR 723.1: the whole of a turn). It reuses §1–§3, but the controller plays an entire turn, which needs the client to drive a second seat's board. Follow-up issue.
- **Opposition Agent** (CR 723.2's other card: control while searching). It reuses §1–§3 for a search prompt. Not requested.

## Owner decisions (2026-10-13)

1. **The forced play: in the resolution (recommended).** Not the grant window of ADR 0066, where the table gets priority before the cast and passing declines it. The play happens inside Word's resolution, and it can't be declined while the card can be played.
2. **Scope: Word of Command now (recommended).** The record and the routing are shared; Emrakul's turn-long control is a follow-up.

## Consequences

- Word of Command ships `full` and the Valgavoth deck (#2032) has no manual card left.
- `control-another-player` moves from missing to partial on the roadmap: Word of Command is implemented, Emrakul still waits.
- A second seat's prompt can now be answered by someone else. Every resolver still reads `Chooser` as the controlled player, so no resolver changes. The one new trust boundary is the dispatcher's "caller controls player" check, and it reads only the record.

## Delivery

One PR: the record and routing, the `controlled_play` prompt, the look-only pick, the land-only mana restriction, the enumerator's probe and moves, Word of Command, the client's prompt, and this ADR.

The client's prompt (`ControlledPlaySheet.svelte`) does not run the cast chain for a card in another player's hand. It lists the controller's legal moves that name the prompt (each a full cast: targets, modes, X, costs), with a land picker, and sends the one chosen. Those moves are never folded by the 48-move wire cap, and the enumerator expands this one card up to 200 ways, so no legal target is cut.

## Amendment (2026-10-16, #3998) — control while a player searches: Opposition Agent

§7 left out CR 723.2's other card:

> You control your opponents while they're searching their libraries.
> While an opponent is searching their library, they exile each card they find. You may play those cards for as long as they remain exiled, and you may spend mana as though it were mana of any color to cast them.

Nothing about it is resolution-long or turn-long. No stack item owns the control, and it lasts exactly as long as a search.

**The span of a search.** A search's choices are made while its `search_library` prompt is open (ADR 0015), and nothing else is decided in between: the prompt is mid-resolution and blocks the table. A search the rules decide alone (one forced take, or no match) opens no prompt, and there is no choice for anybody to make. So the prompt is the span.

**1. The record.** A `PlayerControl` with `Until: "search"` and stage `ControlSearching`, created by `queueSearchChoiceLocked` as the prompt is queued, one per Opposition Agent whose controller is an opponent of the searcher, oldest first (timestamp, CR 613.7). `Prompt` names the search prompt. `controlLiveLocked` reads it as live while that prompt is open **and** its Agent is still on the battlefield under the record's controller with the ability (`CatalogAbilityKey`, so one that lost its abilities controls nothing). `prunePlayerControlsLocked` drops it once the prompt is gone; a player leaving ends the control but never withdraws the prompt, which the searcher then answers for themselves. Only a search of the searcher's own library counts ("their libraries"): Bribery's search of another player's library is untouched. The record is plain data with a prompt ID and no continuation, so the closure ratchet (ADR 0041) is unchanged and a clone or undo taken mid-search is controlled the same way.

**2. Everything else is §2 and §3, unchanged.** The newest live record wins (CR 723.1a), so with two Agents the newer one's controller decides. The prompt keeps its real chooser and is shown to the controller with `on_behalf_of`. The controller's `resolve_choice` is re-addressed to the searcher, and the searcher's own is refused (`ErrPlayerControlled`). `FilterViewFor` makes the controller a knower of everything the searcher knows (the candidates and the searcher's hand, CR 723.4) for exactly as long as the record is live. The search prompt's option list goes to the answerer only, so the searcher sees "Alice is choosing" with no list, and a third player sees neither the list nor the hand. One narrowing in `routeControlledPlayer`: under a `"search"` control the controller may send only `resolve_choice` as the searcher (CR 723.5b). Nothing else is a choice the rules call for during a search.

**3. The found cards.** "They exile each card they find" replaces the destination. It is applied in `executeSearchTakeLocked`, as the cards are found, for every destination (hand, battlefield, top of library after the shuffle, graveyard, exile) and every count ("a card", "up to N", "any number", the "and/or graveyard" half). The search's `Dest` becomes exile, face up, and the cards go through the same exit route, so CR 614 and CR 903.9 still see the move. A reveal still happens, and so does the shuffle. `finishSearchLocked` grants the cards that landed to the newest Agent's controller with `CastPermission{Duration: WhileInZone, AnyColor: true}`. That is a play permission, so a land is played as that player's land drop under the usual timing. The affected player picks which replacement applies (CR 616.1), and that choice is made by their controller, which is why the newest Agent's controller is the one who gets the cards. The permission outlives the Agent. The effect's own continuation (`Then`) is handed no found cards, because they are not where it put them, which is how every "exile it instead" already reads (`landedInZoneLocked`). This applies to every search an opponent makes of their own library while an Agent is there, with or without a prompt, as the card says.

**4. Bots and the enumerator.** `legal.controlledChoiceMoves` already offers the controller the prompts a controlled player owes, enumerated as that player, so the search's answers reach the controller's `legal_moves` with `player` set to the searcher. In `aiseat`, `splitControlledTurnMoves` sends them down the controlled path, and `controlledSearchPolicy` answers them without a model. It takes the answer with the most cards, then the highest total mana value, and never fails to find when it may find. Every card found is the controller's to play and one fewer card in the searcher's library, so taking the most and the best is the controller's interest.

**5. Protocol.** `player_controls[].until` gains `"search"`. Nothing else on the wire changes, and nothing is added to the snapshot: `stage`, `until`, `source` and `prompt` are existing fields (an older binary reads the unknown stage as not live and prunes it).

**Out of scope.** CR 723.1a orders Agent records against an Emrakul turn record over the same player by when each record was made, not by the Agent's own timestamp: an Agent's search control made during an Emrakul turn always wins that search. Both orders give the search to one of the two controllers, and this one gives it to the player whose card is about searches.

## Amendment (2026-10-19, #4025) — which verbs the controller may send, per stage

`routeControlledPlayer` restricted the controller in three cases (a turn-long control, a search, Word of Command's forced play) and let every verb through in the rest. Word of Command's second stage, `ControlSpellResolving` (§5), was one of the rest: while the forced spell resolved, the controller could send `draw_card`, `move_card`, `tap`, `change_life` and every other sandbox verb as the controlled player, which moves that player's cards by hand. CR 723.5b says the controller makes the choices the rules or an object call for and no others, and ADR 0167 §4 closed the sandbox for the turn-long case for that reason. The same reasoning covers every stage.

**One table, default closed.** What the controller may send as the controlled player is now a function of the **stage of the newest live record over that player** (`Game.ControlStageOfPlayer`, the stage `ControllerOfPlayerLocked` already resolves by), read from `controlledStageVerbs` in `internal/actions/controlled_player.go`:

| Stage | Control | The controller may send as the controlled player |
| --- | --- | --- |
| `playing` | Word of Command, the forced play (§4) | `cast_spell` (the play, checked in the cast branch), `activate_mana_ability` (the engine refuses any source but their lands), `resolve_choice` |
| `spell_resolving` | Word of Command, the forced spell resolving (§5) | `resolve_choice` only. Nobody holds priority during a resolution, so no rules verb is a choice either |
| `searching` | Opposition Agent (the 2026-10-16 amendment) | `resolve_choice` only |
| `turn` | Emrakul, the Promised End (ADR 0167 §4) | every rules verb in `turnRulesVerbs` |

A stage with no row, or no live record, allows nothing, so a stage added later is closed until it is given a row. No row contains a `sandboxVerbs` entry. The controlled player's own verbs are still all refused but `concede` (§2), and a turn's controller still has the sandbox closed as themselves (ADR 0167 §4). Nothing on the wire changes.

**What did not change.** The rules verbs `spell_resolving` refuses are ones the controller could not usefully send anyway (no priority is held), so no legal play is lost: the enumerator offers the controller only the prompts the controlled player owes during a resolution.

**Tests.** `TestControlledVerbAllowedPerStage` tables every action type against every stage, with the allowed sets written out independently of the production table; `TestNoControlledStageAllowsASandboxVerb` and `TestTurnControlEverySandboxVerbIsRefusedAsTheControlledPlayer` pin the sandbox closure; and `controlled_verbs_test.go` drives the three Word of Command and Opposition Agent stages through the real cards and `actions.Dispatch`.

**The controller as themselves.** The table above governs what is sent *as the controlled player*. What the controller sends as themselves needs no new rule outside a turn-long control: every verb that names a card gates on the card's controller (`requireCardController`) and every player-scoped verb on `Caller == Player`, so a controller cannot move, tap, mark or reveal the controlled player's cards by hand in any stage. `assertControllerCannotTouchTheirCards` in `controlled_verbs_test.go` pins it for the forced play, the forced spell resolving and a search. The sandbox is therefore closed to the controller as themselves only during a whole-turn control (ADR 0167 §4), where they act for a seat for a full turn; closing it for a single resolution would only stop them editing their own table. The table-wide verbs with no gate for any seat (`announce_trigger` and others) are a pre-existing sandbox question outside control, filed as #4131.
