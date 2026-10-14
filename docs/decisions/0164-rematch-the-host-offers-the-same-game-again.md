# ADR 0164 — Rematch: the host offers the same game again

**Status:** Accepted · 2026-10-13 · Backlog
**Issue:** [#3285](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3285)
**Owner decisions:** the four answers of 2026-10-13 on the issue are quoted under [Owner decisions](#owner-decisions-2026-10-13). They are binding, and they are decided, so this ADR is Accepted on arrival.
**Numbering:** every remote head was listed (15 heads, `git ls-tree docs/decisions/` on each). The highest number anywhere was **0163**, so this is **0164**.
**Builds on:** [ADR 0075](0075-table-settings-and-host-controls.md) (the host and table settings), [ADR 0121](0121-animated-dice.md) (the opening roll that picks the first player), [ADR 0158](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) (Tables with History: the old game stays there), [ADR 0159](0159-per-table-discord-voice-rooms.md) (the pattern for state that rides the game view without entering the engine).

---

## Context

When a game ends, every player is looking at the results screen and the only way on is Back to lobby. To play the same four people again, someone makes a new table, sends a new link, and everyone picks their deck again. The usual wish is "same again".

## Owner decisions (2026-10-13)

1. **Who starts:** only the table's **host** can offer a rematch. The other humans accept or leave.
2. **Turn order:** **random**, the same as any new game.
3. **Expiry:** the offer lasts until **every other human has left the finished table's results screen**.
4. **Empty seats on a public table:** the rematch stays **private** to whoever accepted. The host can put a bot in an empty seat or leave it empty. Empty seats do not reopen to the lobby.

## Decision

### 1. A rematch is an ordinary lobby table, made when the host offers it

`POST /games/{id}/rematch`, from the host's seat session on an ended table, makes a **new game** (new id, new rows) and returns the host's seat session for it. The new table:

- has the old name with " (rematch)" added, and is created for the host's account;
- has the old table's settings (undo limit and scope, starting life, commander damage, bot pace, spawn), copied when they differ from a fresh table's;
- seats the host with the deck the host played, and a **bot seat for every bot** with the same name, tier, curated deck and decklist, so bots come back on their own. A bot that cannot be seated again (its tier is gone, the load cap is reached) is skipped; the host adds another from the table;
- is in the **lobby** state. The host starts it, adds a bot to an empty seat, or leaves a seat empty, with the routes and the lobby screen every table has. Starting opens the **opening roll** like any new game, so who goes first is random (decision 2). Nothing new is built for "start" or "fill with a bot";
- is **private** (decision 4). Both its invites are revoked as it is made and their plaintext is dropped, so no link to it exists, it has no listing, and nothing reopens it. Only the seats the offer carries over ever hold one.

The old game is untouched: it stays ended, keeps its replay and its place in History, and keeps its id.

"The host" is the table's host. A host who lost has passed hosting on, and when **every human lost** (the bots won) the engine has nobody left to host, so the **first human seat** offers. Otherwise a person playing bots would never get a rematch. The client mirrors the rule from the view alone. An admin session cannot offer one.

### 2. Everyone else is asked on the results screen

The offer rides the old game's view as `GameView.rematch`, the way the voice room does (ADR 0159 §6): the lobby sets it on the room, every capture stamps it, and nothing enters the engine, its snapshot or undo. `FilterViewFor` reduces it to the viewer's own part (`you`: `host`, `undecided`, `accepted`, `declined`) and a spectator sees none.

- `POST /games/{id}/rematch/accept` seats the caller at the new table with **their own deck** and answers with their seat session there, as a join does. The client swaps to it and goes to the new table.
- `POST /games/{id}/rematch/leave` records a no. It may still accept while the offer is open. After accepting, leaving is refused.
- An accept is refused (409) when the new table is full or has started, or the offer is closed. A refused accept changes nothing.

Bots are carried over by the lobby, not asked. An MCP agent seat is **not carried over**: its session lives in the client's process, which has no way to learn of a new table. It does not hold the offer open.

### 3. The offer lasts until every other human has left the results screen

The offer is **open** while at least one human seat other than the host has neither answered nor left. The lobby polls the hub (`Hub.ConnectedSeats`) once a second for a socket on each such seat, and a seat missing for 15 seconds has left, so a page reload does not close it. When no unanswered human is left, or every one has answered, the offer closes for good: the new table and whoever it already holds stay, and nobody else can be carried over. A host with no other human to ask closes it at once.

### 4. Decks

A started game no longer holds its decks, so `Start` keeps a deep copy of every seat's deck on the lobby entry (`Game.LobbyDecks`, fresh instance ids). They are dropped when the offer closes, and four hours after the game ended if nobody offered. They are in memory only: a table restored after a restart, or ended more than four hours ago, can no longer be rematched, and the route answers 409 with that sentence.

### 5. Client

On an ended table the dock's Back to lobby gains a **Rematch** button for the host. A human who is asked sees "<host> offered a rematch" with **Join the rematch** and **No thanks**. Offering and accepting swap the tab's session to the new seat and open the new table. The practice table never shows it.

## Consequences

- One new optional field on the game view (`rematch`), three HTTP routes, and one hub method. Old clients ignore the field and keep Back to lobby.
- The new table counts as a created game for the metrics and as a table for the capacity cap (its creation reserves load); the per-creator open-table cap and creation rate limit do not apply, since the host did not create it from the form.
- No new persistence: the new game is a normal game row. The offer itself lives in memory, so a restart during an offer leaves the new table with the host (and bots, and whoever accepted) and closes the offer.

## Out of scope

- Carrying an MCP agent seat over (it needs a way to tell the MCP process the new table).
- Reopening a rematch's empty seats to an invite or the lobby (decision 4).
- Rematching a table after a restart, or after four hours.
