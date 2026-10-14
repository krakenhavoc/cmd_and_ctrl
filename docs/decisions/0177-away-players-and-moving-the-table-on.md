# ADR 0177 — Away players: showing who is gone, and moving the table on without them

**Status:** Accepted · 2026-10-14 · slice 2 (server) built; see the [amendment](#amendment-2026-10-14--slice-2-the-server-3337)
**Issue:** [#3337](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3337) (found by the 2026-10-13 audit).
**Owner decisions:** the four answers of 2026-10-14 are quoted under [Owner decisions](#owner-decisions-2026-10-14). They are binding. Calls this ADR makes on its own are listed under [Calls made here](#calls-made-here) for the owner to overturn in review, before slice 2 lands.
**Numbering:** the AGENTS.md §4 sweep on 2026-10-14 (12 remote heads) found 0174 claimed by the Discord results feed (#3322), 0175 on `fix/2948-liliana-simultaneous-sacrifice` and 0176 on develop (the movable stack). This is **0177**, claimed on #3337.
**Builds on:**
- [ADR 0162](0162-discord-dm-notifications.md) §2: what "away" means, and the hub's presence (`ws/presence.go`, `Hub.SeatAway`).
- [ADR 0060](0060-leaving-the-game.md): leaving the game, CR 800.4a.
- The started-seat removal of #3093 (`lobby/seat_remove.go`).
- [ADR 0033](0033-ai-bot-seat.md): the bot seat and its runner.
- [ADR 0075](0075-table-settings-and-host-controls.md): table settings and the host.
- [ADR 0044](0044-surviving-a-deploy.md) decision 4: reconnecting, with reclaim as the backstop.
- [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md): agent seats.

---

## Context

One absent player stalls a four-player table, and nobody can tell why.

- **No presence on the wire.** The hub knows whether a seat has a connection, and whether its tab is hidden (ADR 0162 §2). Only the DM notifier reads that; the game view carries no per-seat signal.
- **Ambiguous "considering…" tag.** The only hint is the tag on the seat holding priority (`PlayerIdentity.svelte`), and it can't tell a player who is thinking from one who has left.
- **No way to move on.** A host can't concede for someone: removing a seat from a started game is admin-only (#3093) and blocks the person. There is no way to hand a seat to a bot mid-game. Bots are started only when the game starts (`StartBots` in `lobby.go`).

Strangers will join public tables during the pilot. A dropped or wandering player shouldn't end the game for three others.

## Owner decisions, 2026-10-14

1. **What the table can do:** hand the seat to a bot, or concede it for them, or keep waiting. (Recommended option, chosen.)
2. **Who decides:** the host. If the host is the one who is gone, any other present player. (Recommended option, chosen.)
3. **When:** after **3 minutes** with no open connection while the table is waiting on the seat. It's a table setting the host can change. (Recommended option, chosen.)
4. **A returning player can take the seat back** from the bot. (Recommended option, chosen.)

## Decisions

### 1. Presence goes on the wire

`PlayerView` gains `presence?: { state, since }`. It is omitted while the seat is present.

- **`state`:**
  - `"offline"` means no connection is bound to the seat.
  - `"hidden"` means every bound connection reports its tab hidden (ADR 0162 §2's `visibility` frame).
- **`since`:** when the seat last stopped being present (`Hub.SeatAway`'s `since`), as RFC 3339.

It is computed from the hub at send time, the way the room stamps other per-connection data on the view. So it isn't in the engine, the snapshot or the undo ring, and a restart starts every seat as "not seen leaving", the same as the notifier.

**Who sees it.** Every seat and every spectator, because who is at the table is public at a real table.

**Exempt seats.** Bots never carry it. Agent seats (ADR 0122) do: an agent whose process died is as stuck as a person who left.

**On screen.** The seat shows a chip:
- "offline 2m" for `offline`;
- "away 4m" for `hidden`, once it has lasted 2 minutes (ADR 0162's away time).

A seat holding priority while offline reads "offline", never "considering…".

### 2. When the table is offered a way on

A seat becomes **movable** when both of these hold:

- It has been **`offline` continuously for the table's `away_after`** (default 3 minutes).
- **The table is waiting on it.** "Waiting" uses the lobby's existing answer for the notifier (`notify.Tables`, ADR 0162 §2): the engine is waiting on that seat to hold priority, answer a prompt, mulligan, roll, declare, or discard.

A hidden tab never makes a seat movable. A player who alt-tabbed to Discord is still at the table, and the DM notifier is the right tool for them.

When a seat becomes movable, the deciders (decision 3) see a sheet in the action dock:

> **Bo has been offline for 3 minutes.** The table is waiting on them.
> [Hand Bo's seat to a bot] [Concede for Bo] [Keep waiting]

- **Keep waiting** hides the offer for another `away_after`. It is per seat, per absence, and the same for every decider.
- **Reconnecting withdraws the offer.** If the seat reconnects before anyone answers, the offer is gone.
- **One answer wins.** The first answer the server accepts is the one that happens; the others see the result in the log.

### 3. Who decides

- **The host decides.** That's the effective host of ADR 0075 (`CanManageTable`), when the host's own seat is present (not `offline`).
- **Otherwise, any present player decides.** If the host is offline, or the host is the movable seat, any seated, present, non-bot player may decide. Agent seats are included, as players.
- **The server admin** can always decide.
- **Re-checked when the answer arrives.** A decider who has since gone offline is refused.

### 4. Handing the seat to a bot

- **What runs it.** The seat is driven by the **heuristic** bot runner from that moment. It plays the player's own deck and board, and keeps their name, with a "(bot)" marker on the seat.
- **How it's recorded.** On the lobby's `GameMeta` seat as `StandIn: "heuristic"`. It isn't recorded as `IsBot`: the seat is still the person's, so the reclaim, statistics and DM code keep treating it as theirs.
- **How it starts.** `StartBots(room, [that seat])` starts the runner mid-game. On a restore, stand-in seats are restarted with the table's other bots.
- **Nothing in the engine changes.** The player is still in the game, still the owner and controller of their objects, so no rules question arises. The undo ring is untouched.
- **The log says:** "Ana handed Bo's seat to a bot (offline 3m)."

### 5. Conceding for them

The seat concedes as a player who gives up (CR 104.3a), through `game.Concede`, so CR 800.4a runs as for any concession: their objects leave, control effects end, and the turn moves on. If one player is left, the game ends (CR 104.2a).

Unlike #3093's admin removal, the seat is **not** marked `Removed` and the person is **not** blocked. They didn't do anything wrong, and if they come back they can watch the rest of the game.

The undo ring is emptied, as for #3093, because every entry on it is from before they left.

The log says: "Ana conceded for Bo (offline 3m)."

### 6. Taking the seat back

When the person's session reconnects to a stand-in seat:

- **They see a banner:** "A bot has been playing your seat since 21:14. [Take it back]".
- **Take it back** stops the runner at its next decision boundary (it never abandons an action half-sent) and clears `StandIn`. The log says: "Bo took their seat back."
- **Without that press,** the bot keeps playing. A player who only opens the table to look shouldn't take over by accident.

### 7. The table setting

`away_after` (ADR 0075 §2.2) is in minutes, 1–30, default 3. `0` means never offer. It's changeable at any time (§2.3), and a change applies to the clocks already running. Practice tables never offer.

## Calls made here

- **Offline only, not hidden, makes a seat movable.** Hidden shows as a chip but never offers a takeover (decision 2).
- **"Waiting" is the notifier's definition.** It's reused rather than redefined, so a DM and an offer can never disagree about whose move it is.
- **The bot is the heuristic tier.** Model tiers cost money (#2980) and would need the owner's budget. Random would throw the game.
- **`StandIn`, not `IsBot`.** The seat stays the person's (decision 4).
- **A concession for an away player is not a removal.** No `Removed` mark and no block (decision 5).
- **Take it back is a button, not automatic** (decision 6).
- **Statistics (#3370).** A game with a stand-in seat records `stand_in` on that seat's row. Whether it counts toward the person's win rate is the stats owner's call; this ADR only records it.
- **The offer is a dock sheet, not a modal.** The game behind it is still the game, as for every other prompt since ADR 0111.

## Out of scope

- Kicking a present player. That's #3093 and the moderation path.
- A vote. The owner chose host-else-anyone (decision 3).
- DMs to the away player. ADR 0162 already does that for a player who has opted in.
- Timeouts per priority window, chess clocks.

## Consequences

- One absent player no longer ends a game for three others. The table can play on with a bot, or play on without them.
- Everyone can see who is gone and for how long, instead of guessing at "considering…".
- A player who drops for five minutes can come back to their own seat.
- New surface to keep honest:
  - a per-seat presence field on the view;
  - three lobby actions: hand over, concede-away and keep waiting, plus take-back;
  - one table setting;
  - mid-game runner start and stop.

## Slices

| # | Slice | Model | Needs |
|---|---|---|---|
| 1 | This ADR | — | — |
| 2 | **Server:** `presence` stamped on the view from `Hub.SeatAway`; the `away_after` setting; the movable-seat clock (reusing the notifier's waiting-on-seat); `POST /games/{id}/seats/{player}/away` with `{"action": "stand_in" \| "concede" \| "wait"}` and the decider check; `StandIn` on `GameMeta`, runner start on hand-over and on restore, runner stop and `POST /games/{id}/seats/me/take-back`; the log lines; docs/protocol.md and docs/lobby.md | Opus | 1 |
| 3 | **Client:** the presence chip and "offline" over "considering…"; the offer sheet for deciders; the take-back banner; the `away_after` field in the table settings panel | Sonnet | 2 |
| 4 | **E2E:** a three-seat table where one page closes. After `away_after` (set to 1 minute) the host sees the offer, and each action does what it says; then the closed page reopens and takes its seat back | Sonnet | 2, 3 |

## Amendment 2026-10-14 — slice 2, the server (#3337)

Slice 2 is built. Where the decisions above left the mechanism open, it does this:

- **How the offer reaches the deciders.** A movable seat carries `presence.movable: true` on every frame. The client offers the sheet to the seats §3 names; the server checks the same rule when the answer arrives. `POST /games/{id}/seats/{player}/away` takes `stand_in`, `concede` or `wait`. A seat that is no longer movable answers 409, so the first answer accepted is the only one.
- **When frames go out.** The hub stamps presence on every frame, but it does not send one when a connection opens or closes. The lobby's away clock looks at every table every 5 seconds. It sends a table a frame at its current seq when a seat's presence or a movable mark changed. So a chip or an offer is at most 5 seconds late, and a reload costs the table one frame.
- **Offline since when.** For `offline`, `since` is when the seat's last connection closed, not when its tab went hidden before that. A seat the server has not seen leave since it started is away from the moment the server first looked, which is how the notifier counts it.
- **The stand-in runner.** It runs beside the table's own bots, through `StartStandIn` and `StopStandIn` on the bot host. `StartBots` replaces a game's whole set, which would restart the other bots. Take back cancels the runner and waits for it to exit. The runner never cancels between choosing a move and sending it.
- **A stand-in seat's own actions are refused** with the `stand_in` error until the player presses Take it back. Otherwise a player who reconnects and the bot would both drive the seat. The admin is not held to it.
- **Deciders must still be in the game.** A player who has been eliminated, or whose seat was removed, does not decide for the table.
- **Statistics.** The hand-over and the take-back are lines in the game's log (`EventAwaySeat`), so a statistics pass can tell which seats a bot played part of. Recording it on `seat_stats` is left to #3370.
