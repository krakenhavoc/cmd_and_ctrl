# ADR 0168 — The London mulligan: cards to the bottom after keep, and the free first mulligan

**Status:** Accepted · 2026-10-13 · S72 — Smoother table play
**Issue:** [#2799](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2799) (found by the shuffle-fairness work, #2787 and #2788).
**Numbering:** first claimed as **0166** on #2799, which collided with Word of Command's 0166 when that landed on develop; 0167 was taken on another head. The AGENTS.md §4 sweep on 2026-10-13 (13 remote heads) found **0167** the highest, so this is **0168**, re-claimed on #2799.
**Builds on:** the turn-ordered mulligan window of `game/mulligan_order.go` (#2237), [ADR 0088](0088-ordered-library-placement.md) (ordering a pile headed to the bottom of a library), the chained-choice prompts Brainstorm uses (`effects/library_order.go`), and [ADR 0133](0133-opening-hand-actions.md) (prompts raised as the window closes).

---

## Context

### The rules

From the Comprehensive Rules effective September 25, 2026:

- **CR 103.5.** Each player draws their starting hand size, normally seven. To take a mulligan, a player shuffles their hand into their library, draws a new hand of their starting hand size, "then puts a number of those cards equal to the number of times that player has taken a mulligan on the bottom of their library in any order". A player who keeps may take no further mulligans. "A player can take mulligans until their opening hand would be zero cards, after which they may not take further mulligans."
- **CR 103.5c** (and **CR 800.6**). "In a multiplayer game and in any Brawl game, the first mulligan a player takes doesn't count toward the number of cards that player will put on the bottom of their library or the number of mulligans that player may take."
- **CR 100.1a / 100.1b.** A two-player game begins with only two players; a multiplayer game begins with more than two.

### What the engine does today

- `game.Mulligan(player, handSize)` shuffles the hand away and draws `handSize` cards. Nothing goes to the bottom, ever.
- The client's Mulligan button always sends `hand_size: 7`, so a human's mulligans are free and unlimited in effect.
- The legal-move list offers bots and the MCP seat "Mulligan to 7, 6, 5…", a shrinking hand, so bots pay a cost humans don't.
- The turn order is right already: `mulligan_order.go` (#2237) has each seat decide in turn from the starting player, round after round.

## Decisions

### 1. Inside the window, a mulligan always draws seven

While `MulligansOpen`, `Mulligan` draws the starting hand size (seven) whatever `hand_size` says. It is refused with `ErrNoMoreMulligans` once the seat's counted mulligans reach seven, the point where the opening hand would be zero cards (CR 103.5). The free first mulligan of a multiplayer game does not count toward that limit (CR 103.5c).

Outside the window, `mulligan` with a hand size stays what it was: the ⋯ menu's sandbox "Mulligan to N", which is not a rules action.

### 2. The cards go to the bottom when the seat keeps

N is the seat's mulligans taken, less one in a game that began with more than two players. When a seat keeps with N > 0, the engine asks it to put N cards from its hand on the bottom of its library, in an order it chooses.

The rule's own order bottoms as part of each mulligan, before the next keep-or-mulligan declaration. Bottoming once, at keep, is what Arena and MTGO do. It leads to the same games: either way the player looks at the same seven cards, chooses the same N to give up, and the cards bottomed after an earlier mulligan are shuffled away by the next one. The difference is only that a player who mulligans again does not first sort cards that are about to be shuffled away.

### 3. Two prompts, the Brainstorm pair

The keep raises the same two prompts Brainstorm's put-back uses, both private to the seat (its hand is a hidden zone):

1. A `choose_cards` prompt over the hand: exactly N cards. Skipped when N is the whole hand. It declares its destination, `library_bottom` (a new value of the open `ChooseDestination` vocabulary), so the client's sheet says where the cards go instead of the generic "what happens to them is the card's business".
2. A `put_in_library` prompt with the bottom lane (ADR 0088) to order them. Skipped for a single card, which has no order.

Bots answer both through their existing scoring: a `choose_cards` over the bot's own hand keeps the most valuable cards (`valueKeptInHand`), and the library order prompt is scored as it already is. The MCP seat gets both prompts through the legal-move list, unchanged.

### 4. The window waits for the bottoms; the other seats don't

A seat that keeps is done deciding, so the next seat's decision is offered at once and the table does not wait on a sorting player. The window closes, and CR 103.6's opening-hand actions are offered (ADR 0133), only once every live seat has kept **and** finished bottoming. A per-seat flag, set at keep and cleared when the cards are on the bottom, says which seats are still bottoming.

A seat that leaves mid-prompt drops its prompts and no longer gates the window, as for a seat that leaves before deciding.

### 5. What the table sees

- The seat's view gains `mulligan_bottom`: how many cards a keep would put on the bottom now. The opening-hand sheet says it under the buttons ("Keep: put 1 card on the bottom"), and so does the mulligan's cost line.
- The legal-move labels say the same: "Keep hand (put 1 on the bottom)" and "Mulligan (draw 7, then put 2 on the bottom)". The mulligan move's `hand_size` parameter is now the opening hand that mulligan leads to (seven less what it would bottom), which is what the bot's keep rules reason about. The engine ignores it inside the window.
- The bottomed cards are face down to everyone. Their owner knows them, as ADR 0088 decision 3 says for any ordered placement.

## Calls made here (for owner review)

- **A two-player Commander table gets no free mulligan by default.** CR 103.5c gives it to multiplayer games only, and a two-player game is not one (CR 100.1). Many playgroups grant it in every Commander game, so it is a host table setting (see the amendment below).
- **Bottoming at keep** rather than after each mulligan (decision 2).

## Amendment, 2026-10-13: the free first mulligan as a table setting (#3387)

The owner's call: the default follows the CR, and the host can turn on a free first mulligan for a two-player table.

- **The setting.** `TableSettings.FreeFirstMulligan` (wire `free_first_mulligan`, ADR 0075 §2.2), off by default. When it is on, `freeMulligansLocked` is 1 whatever the seat count, so everything that reads the count follows it: the keep's bottom count, the mulligan limit, the view's `mulligan_bottom`, the legal-move labels and `hand_size`, and the bots' keep rules. A multiplayer table has the free mulligan by the rules either way, so the setting changes nothing there.
- **Fixed at Start.** The mulligan window opens at Start, and a cost that changed while a seat was deciding would be unfair. A change after Start is refused with `ErrFreeMulliganLocked` (HTTP 422), the same rule and status as starting life, and the panel disables the control with the same "already been dealt" hint.
- **Where it shows.** The table settings panel has a checkbox for it. The log says "made the first mulligan free" or "made the first mulligan cost a card". A saved setup carries it, and an LFG post lists it when it is on. The client's mulligan cost line treats the first mulligan as free when the table has more than two seats or the setting is on.
- **Snapshot.** An additive field, `settings.freeFirstMulligan`, omitted when false. An older snapshot reads as off, which is the rules. No schema bump (the `snapshot_shape` rule for additive fields).

## Out of scope

- CR 103.5b, actions "any time you could mulligan" (Serum Powder). Nothing in the catalog prints one.
- Vanguard hand modifiers (CR 103.5a) and the shared team turns option (CR 103.5d).
- A different starting hand size. Nothing in the catalog changes it, so the engine keeps seven as a constant.

## Snapshot impact

None persisted. `MulligansTaken` and `MulligansOpen` are already in the snapshot. The bottoming flag is dropped, not persisted: it is set only while the two prompts are open, and their continuations are counted in `ContinuationCensus.ChoiceResumeFrames`, so a table waiting on them is not a restore point. A deploy in that moment restores to before the keep, and the seat decides again.

## Delivery

One PR: this ADR, the engine (count, limit, prompts, the window's close), the view field, the legal-move labels and parameter, the bot's keep rules, the client's copy, and tests for each.
