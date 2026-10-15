# ADR 0176 — A movable stack panel

**Status:** Accepted · 2026-10-14 · Backlog
**Issue:** [#3429](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3429)
**Owner decisions:** the issue's "Wanted" list (2026-10-14) and its "Defaults chosen" section, which nobody changed. Both are binding.
**Numbering:** every remote head was listed (`git log --all --diff-filter=A --name-only -- docs/decisions/`, 15 heads). The highest number anywhere was **0173**; **0172** is taken on `docs/3401-companion-adr`. This is **0176**.
**Amended:** 2026-10-15 · [#3380](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3380) — the pile's default spot docks in the seam between the seats ([amendment below](#amendment-2026-10-15-the-default-spot-is-the-seam-3380); amends [ADR 0119](0119-a-stack-you-can-follow.md) §1's placement)
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

The saved spot is the top-left corner of the stack's box in every style. With nothing saved each style keeps the placement [ADR 0119](0119-a-stack-you-can-follow.md) gave it (the pile at the board's left edge beside the seam, the others centred across the table). With a spot saved the lane leaves that anchor for the spot: the pile keeps its card-sized width; `fan`, `spotlight` and `ribbon`, which are 1040px wide centred, come in to 640px (or the board's width less 32px), since a full-width lane could only move up and down. A pile folded to its tab sits at the saved spot and the tab is a handle too: a press that moves 3px or more drags and saves, a plain click unfolds it as before; reset is on the unfolded header. The pile's phone and short-board fallback to the docked card is unchanged, and there the docked card uses the same spot.

## Not decided here

An account-wide setting such as "always put the stack on the right" is a separate setting for later.

## Amendment 2026-10-15: the default spot is the seam (#3380)

**Owner decision:** on [#3380](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3380) the question of 2026-10-15 offered three defaults for a seat with nothing saved: (a) a fixed gutter, docking in the seam between the seats (the centre of the table, between the top and bottom rows) rather than over a corner seat; (b) over the least-active board; (c) no change. The owner answered "Let's test A in dev". This records (a) as the default to try on cmd-dev; it is a trial, and the owner may change it after playing with it.

**What changes.** Only the pile's default, the spot it takes when the seat has nothing saved. With two or more seats side by side in the top row (a 3- or 4-seat table, seated or spectating), the pile is centred horizontally on the seam between two of those seats, and stays centred vertically on the row seam as §6 and [ADR 0119](0119-a-stack-you-can-follow.md) §1 had it. On the quadrant grids (the default layout) that seam is the board's centre, where the four boards meet. In the row and focus layouts, whose top row has three seats, it is the seam nearest the centre, the left one on a tie. It is kept 12px inside the board on both sides. The strip and bottom-left fixture clamps of ADR 0119 §1 still apply; a fixture now counts only when it overlaps the pile's column, wherever that is.

**What does not.**

- A two-seat table keeps ADR 0119's spot, 12px from the left edge on the row seam. Its top row has one seat, so there is no seam between seats to dock in.
- A saved spot still wins: the default is only drawn while the seat has none, and reset (§5) returns to the new default.
- `fan`, `spotlight` and `ribbon` were already centred across the table on the row seam, and the compact docked card stays in the attention strip; neither moves.

**How.** `tableSeams` (lib/stackPile.ts) reads the row seam and the column seam from the board's seat slots, `pileLeft` centres the pile on the column seam, and `laneSpot` (lib/stackPosition.ts) draws the saved spot over that default. All three are pure and pinned by tests for 2-, 3- and 4-seat tables.

**Known cost.** The seam is a gap of 8px, and the pile is a card-sized panel, so it still covers the inner corners of the seats it sits between, and on a short board a tall pile overlaps them more. On a mock 4-seat board it covered about as many permanents as the left-edge spot did, spread over the corners of all four boards instead of over one seat. This is what the dev trial is for.
