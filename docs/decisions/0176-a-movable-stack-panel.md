# ADR 0176 — A movable stack panel

**Status:** Accepted · 2026-10-14 · Backlog
**Issue:** [#3429](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3429)
**Owner decisions:** the issue's "Wanted" list (2026-10-14) and its "Defaults chosen" section, which nobody changed. Both are binding.
**Numbering:** every remote head was listed (`git log --all --diff-filter=A --name-only -- docs/decisions/`, 15 heads). The highest number anywhere was **0173**; **0172** is taken on `docs/3401-companion-adr`. This is **0176**.
**Builds on:** [ADR 0159](0159-per-table-discord-voice-rooms.md) §6 and [ADR 0164](0164-rematch-the-host-offers-the-same-game-again.md) (state that rides the game view without entering the engine), [ADR 0119](0119-a-stack-you-can-follow.md) (the stack card and its docked place in the attention strip).

---

## Context

The stack card sits at the top left of the board, over the opponent's side. In a one-on-one game several triggers make it grow over the opponent's avatar and hide their life and counters.

## Decision

### 1. The stack can be dragged, in every style, and where it sits is the seat's own

The header of the stack is a drag handle (pointer events, so touch works): the docked `compact` card's, and the floating lane's for the `pile` (the default), `fan`, `spotlight` and `ribbon` styles ([ADR 0119](0119-a-stack-you-can-follow.md)). Each style's own placement stays the default. A placement belongs to the **seat** that made it and only changes that seat's view. A spectator can drag too, but nothing is saved for them.

### 2. A position is two fractions of the board

`{x, y}`, each from 0 to 1: the stack's top-left corner as a fraction of the board's width and height. One saved spot per seat applies to whichever style is showing, so changing the style keeps the place. A fraction survives a window resize. The client clamps what it draws so the stack stays fully on screen at its size now; the saved value is not rewritten by the clamp, so a window that grows again gives the card back its place.

### 3. It lives outside the engine, like `voice` and `rematch`

The room keeps one placement per seat (`ws.Room.SetStackPosition`) and stamps the map on every captured view. It is not in the snapshot, `undo` does not touch it, and it is dropped from the view once the game has ended. It is **not written to disk**: a server restart puts every card back at its default, which costs a player one drag. Persisting it would add a snapshot field for a cosmetic.

### 4. It goes up over HTTP and comes back on the viewer's own view

`PUT /games/{id}/stack-position` with `{x, y}` saves; `DELETE` resets. The seat is the caller's session, never a name in the request, so nobody can place another seat's card; a spectator or the admin token gets 403. A value outside [0, 1], a missing field or an unknown one is 400; a table that is over is 409. The write is one drag, not a stream: the client sends it on release, and the route is rate limited per caller.

The save is committed through the lobby's usual external-change path, so every client of that seat (a second tab, a reload, a reconnect) gets it on the next state frame as `GameView.stack_positions`. `FilterViewFor` reduces the map to the viewer's own entry; a spectator's view has none.

### 5. Reset

Double-clicking the header, or the "Reset stack position" button that shows once a placement exists, clears the saved spot. The button is the keyboard route. A drag starts only on the header, with the pointer captured there, so it never clicks a stack item, and a move under 3px is a click. The drag, save and reset are one helper (`lib/stackPlacement.svelte.ts`) used by the docked card and the lane.

### 6. How each style maps onto the spot

The saved spot is the top-left corner of the stack's box in every style. With nothing saved each style keeps the placement [ADR 0119](0119-a-stack-you-can-follow.md) gave it (the pile at the board's left edge beside the seam, the others centred across the table). With a spot saved the lane leaves that anchor for the spot: the pile keeps its card-sized width; `fan`, `spotlight` and `ribbon`, which are 1040px wide centred, come in to 640px (or the board's width less 32px), since a full-width lane could only move up and down. A pile folded to its tab sits at the saved spot but has no header to drag; unfold it first. The pile's phone and short-board fallback to the docked card is unchanged, and there the docked card uses the same spot.

## Not decided here

An account-wide setting such as "always put the stack on the right" is a separate setting for later.
