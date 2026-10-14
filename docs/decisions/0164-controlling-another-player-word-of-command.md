# ADR 0164 — Controlling another player: Word of Command

**Status:** Accepted · 2026-10-13 · Backlog.
**Issues:** [#2164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2164) (the `control-another-player` seam). Requested card: Word of Command ([#2032](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2032), the Valgavoth deck, whose last manual card it is).
**Owner decisions behind it (2026-10-13):** the forced play is cast inside Word of Command's resolution, not offered as a later window, and it can't be declined while it can be played; the seam is built for Word of Command now, with Emrakul, the Promised End's turn-long control left for a follow-up. See [Owner decisions](#owner-decisions-2026-10-13).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-13. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 15 remote heads. The highest number anywhere is 0163, so this ADR takes **0164**.
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
