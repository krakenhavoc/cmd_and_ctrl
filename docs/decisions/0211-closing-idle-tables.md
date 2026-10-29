# ADR 0211 — Closing idle tables: unstarted tables after a day, abandoned games after two

**Status:** Proposed · 2026-10-29 · S70 — Public beta: a table strangers can find. Building starts after the soft launch of 2026-10-30.
**Issue:** [#4562](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4562)
**Owner decisions:** the four answers of 2026-10-29 are quoted under [Owner decisions](#owner-decisions-2026-10-29). They are binding. Calls this ADR makes on its own are listed under [Calls made here](#calls-made-here), and the questions still open are at the [end](#open-questions-for-the-owner).
**Numbering:** checked on 2026-10-29. `develop` ends at 0210, and `git ls-tree` over every remote head's `docs/decisions/` shows no file above 0210. The heads checked are listed in the PR. This is **0211**.
**Builds on:**
- [ADR 0148](0148-capacity-limits-for-a-public-launch.md): the table-load cap, the restore pass and deferred tables.
- [ADR 0177](0177-away-players-and-moving-the-table-on.md): away seats and stand-in bots.
- [ADR 0162](0162-discord-dm-notifications.md): DM notifications, and opt-in asynchronous play as a goal.
- [ADR 0156](0156-a-front-door-site-shell-front-page-and-a-discord-gate.md)'s S77 amendment of 2026-10-16: the #mod-log.
- [ADR 0123](0123-monitoring-metrics-logs-dashboards-and-alerts.md) (metrics) and [ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md) (admin views).
- [ADR 0057](0057-win-and-lose-by-effect.md) decision 7: `games.outcome`, which tells a draw from an abandoned table.

---

## Context

The load cap (ADR 0148, `lobby/capacity.go`) adds up the cost of every unended, unarchived table. Nothing closes a normal table on its own, so a table nobody starts, or a game everyone walks away from, counts against the cap until an admin archives it. It survives restarts too: the restore pass brings it back. Each account may also hold three open lobby tables (`MaxOpenTablesPerCreator`), and an abandoned one keeps its slot. After a public launch, abandoned tables will fill the cap and the server will refuse real players.

What exists today:

- **Practice tables** are reaped after 20 minutes with no commit, or after 2 hours (`lobby/practice.go`, `watchPractice`). They have no rows and no restore point.
- **Voice rooms** are removed after 2 hours with no seat change while nobody is in the room (`voiceIdleAfter`, `lobby/voice_manager.go`). The table stays.
- **`closeUnresidentRows`** (`lobby/unresident.go`, #2987) marks a `lobby` or `active` row ended at boot when the restore pass did not bring its table back.
- **Admins** archive or delete by hand: `POST /games/{id}/archive`, `DELETE /games/{id}`, the Discord bot's `/c2-end` (which archives), and the buttons on `/admin/games`.

### How an admin close is recorded

Archiving is how an operator closes a table (`lobby/metrics.go`). `Lobby.SetArchived(id, true)`:

- counts the end once as `cmdctrl_games_ended_total{outcome="closed"}` if the engine still had the game running (`recordClosedLocked`);
- stamps `archived_at`, drops the table's reclaim tickets and stops its bot runners;
- writes the row through `persistGameLocked`, which also tells the voice manager and the #lfg manager. An archived table wants no voice room and no #lfg post, so both are removed (`voice_manager.go`, `lfg_board.go`).

The HTTP handler then evicts the table's sockets and posts "Table ended" to #mod-log (`ModLog.TableArchived`).

The engine's state is not changed. The row keeps `state` as it was, with `outcome` NULL and `winner_seat` NULL. Nothing goes through `syncStateLocked`'s end, so no statistics record is written (ADR 0165), no result is posted (ADR 0174), and no rematch is armed (ADR 0164). An archived table costs nothing against the cap (`entryCostLocked`), does not count toward its creator's three open tables (`OpenTablesCreatedBy`), shows as "archived" in My games with no rejoin link, and can be put back with `DELETE /games/{id}/archive`.

### Two gaps found while reading

1. **An archived table is hidden, but it may not be closed.** `FindByInvite` refuses an archived table, but `join`, `SpectateAs`, `Start`, `SetDeck` and `addBot` do not check `Archived()`, and neither does `WSAuthorizer.AuthorizeUpgrade`. As far as this reading goes, a player who still holds a session can open the table's URL, reconnect and keep playing, with the bots stopped. Sub-PR 1 starts with a test that shows whether this is so.
2. **Archived tables never leave memory.** An archived table keeps its room, undo ring included, in `l.games` until the process stops, and its restore point stays on disk, so every boot brings it back (ADR 0148 §7: an archived table "always comes back"). That is harmless for a few hand archives. It is not harmless for a sweep that may close dozens of tables a day.

## Owner decisions, 2026-10-29

1. An **unstarted lobby table** with no activity for **24 hours** is closed automatically.
2. A **started game** with **no move by any human for 48 hours** is ended with no winner, recorded the way an admin close is recorded.
3. A per-table **"slow game"** option (opt-in, about 7 days) for play-by-DM games is **deferred**. ADR 0162 names opt-in asynchronous play as a goal. The design leaves room for it (§9).
4. **Practice tables** keep their 20-minute reap.

## Decision

### 1. What counts as activity

There are two clocks, one per kind of table. Both are wall-clock times.

**A lobby table's clock** is reset by anything a seated person, or the table's host or creator, does to the table:

- joining a seat, leaving it, or being seated by a rematch accept;
- setting or changing a deck (`SetDeck`, `SetSeatDeckID`);
- the host adding or removing a bot, changing a table setting (over HTTP or the socket), transferring hosting, or switching voice, #lfg or results;
- the creator rotating an invite, or a seated person sending a DM invite;
- a chat line from a seated, non-bot connection;
- starting the game (which moves the table to the other clock).

It is not reset by a spectator, a preview, a socket opening or reconnecting, the playmat restamp (`playmat.go`), the voice manager's restamp (`voice_room.go`), the #lfg manager, or an admin.

**A started game's clock** is reset only by a **human move**: a commit the room accepts for a person's seat, at that person's request. That is:

- every action, bundle and undo the hub accepts from a seat-bound, non-admin connection whose seat is a person or an agent (ADR 0122). Passing priority is a move. So is the opening roll, a table die roll, a mulligan and a table-settings change.
- an away answer (hand over, concede for them, keep waiting) and a take-back (ADR 0177), which come over HTTP from a present player.

It is not reset by a bot's commit, a stand-in's commit, an admin's commit, the lobby's own commits, chat, a spectator, or a reconnect without an action.

**Where the stamp is taken.** `Room.noteActed` (ADR 0162 §2) already stamps the seat behind every commit, but it cannot be reused as is: a stand-in runner commits as the person's seat (`aiseat.Runner.apply` tags `Actor: r.seat`), and `Room.seatKind` reads a stand-in seat as human. So the room gains one more stamp beside it, `lastHumanAt`, taken in the same place under the same `actMu`, when the commit's actor is not uuid.Nil, the tag is not `Admin`, `Game.SeatDriver` says the seat is not a bot, and `Room.StandIn(seat)` is false. The hub already refuses a stand-in seat's own actions until Take it back, so every commit for a stand-in seat is the runner's.

The lobby side is a `touchLocked(entry)` call at each lobby entry point listed above, setting `entry.lastTouched`. It is explicit, not a hook in `applyLocked`, because `applyLocked` also carries system commits that must not count.

**A table's last activity** is the newest of `entry.lastTouched`, `room.LastHumanAt()` (and, for a lobby table, the last seated chat line), the persisted value below, and `created_at` (lobby) or `started_at` (started game).

**Where it survives a restart.** A new nullable column, `games.last_active_at` (Unix milliseconds, migration 0029). The sweep (§3) writes it when a table's in-memory value is newer than the stored one, so it is at most one sweep interval stale. Nothing on the commit path writes to the database. `loadEntry` reads it back on restore. A row with no value, which is every table that exists when this ships, starts at the time it is loaded, so the first deploy gives every existing table a full 24 or 48 hours rather than closing it at once.

The restore point is not used for this. It is written on every commit, so it would be free, but it is a snapshot-schema change for a lobby fact, and a deferred table's row can be read without it.

### 2. Bot-only tables

**What happens today.** A started game whose humans leave without conceding waits on them. The bots stop when the engine waits on a person, so the table burns no CPU, only cap, and the 48-hour clock closes it.

A table whose humans have all **conceded, lost or been handed to stand-ins** is different. The bots keep playing at the table's pace (`BotPaceNormal`: at least 700 ms a decision) until one of them wins. The loop breaker (ADR 0055, `game/loop_breaker.go`) bounds a bot-only game, but a four-bot Commander game can run for an hour or more, and a stall leaves it at rest, counting against the cap. The stand-in runners commit as people's seats, so a table played entirely by stand-ins looks alive to every clock that reads `lastActed`.

**What this ADR does.** A started game is **bot-only** when no seat still in the game is played by a person: every remaining seat is a bot or a stand-in. A bot-only game with **no socket bound to any human seat** (a conceded player watching the bots finish counts as someone at the table, a spectator does not) is closed after **1 hour**, through the same path as §4. The bot-only clock is in memory only. It starts when the sweep first sees the table bot-only and unwatched, and it resets when either stops being true. A restart starts it again, which costs at most one more hour.

The 48-hour clock applies as well, so a watched bot-only table with no human move still closes at 48 hours.

Whether a bot-only game should first be **played out at the fast pace** is open question 2.

### 3. The sweep

`Lobby.RunIdle(ctx)` ticks every **5 minutes** (`IdleSweepEvery`), started in `cmd/server/main.go` after `RestoreFromDisk`, beside `RunAway`. `TickIdle(now)` is exported for tests, which drive it with a fake clock as `TickAway` is.

Each tick, for every entry that is not practice, not archived, and in `lobby` or `active` state:

1. Work out the table's last activity (§1) and write `last_active_at` if it moved.
2. Work out its deadline: last activity plus `IdleLobbyAfter` or `IdleGameAfter`, or the bot-only clock plus `IdleBotsAfter`, whichever comes first.
3. Never close a table within **1 hour of it becoming resident** in this process (`entry.residentSince`, set by the boot restore, a hot restore of a deferred table, or a create). Wall-clock time counts while the server is down, so a long outage could otherwise close a table the moment the server comes back.
4. For a started game, warn at the deadline minus the warning lead (§6). Never close a started game within 1 hour of its warning.
5. Close what is due (§4).

Collecting the work happens under `l.mu`; the closes run after it is released, as `TickAway` does.

**At boot.** The restore pass and its admission are unchanged. The admission already puts the most recently active tables first (ADR 0148 §7), so a table past its clock is the last to take room. The sweep does not touch **deferred** tables: they are not resident, their rows stay exactly as they are (ADR 0148's promise), and when `POST /admin/capacity/restore-deferred` or a later boot brings one back, it gets the 1-hour residency grace like any restored table.

### 4. The close: an archive, with three additions

The sweep closes a table through `SetArchived`'s own steps, so it is recorded exactly as an admin close: `archived_at` set, `outcome` and `winner_seat` NULL, `cmdctrl_games_ended_total{outcome="closed"}` counted once for a running game, reclaim tickets dropped, bots stopped, voice room removed, #lfg post withdrawn, sockets evicted (`l.evictor`, as the practice reaper does). No statistics record, no result post and no rematch follow, for the reasons in Context.

The three additions:

- **The reason is recorded.** A new nullable column, `games.closed_reason` (migration 0029, with `last_active_at`): `idle_lobby`, `idle_game` or `idle_bots`. An admin archive leaves it NULL, and unarchiving clears it. My games and the admin views read it (open question 4).
- **A line in the table's log.** The close commits one `EventIdleClose` through `ApplyExternalClearingUndo` before archiving: "Closed: no moves for 48 hours." (or "no activity for 24 hours", or "only bots left for an hour"). The undo ring is emptied by the same commit, which frees most of a closed table's memory. Nobody is connected to see the line live; it is for the replay and for anyone who opens the table's history. The new event kind is additive within the snapshot version (docs/snapshots.md).
- **The table leaves memory after 7 days.** An idle-closed table keeps its room and restore point for **7 days** (`idleClosedKeep`), so an admin can unarchive it if someone asks. After that, the sweep unloads it: a new `RoomManager.Unload(id)` drops the room, the restore point and the snapshot but keeps the replay (unlike `Delete`, which removes the replay too), and the row is written `state = 'ended'` with `ended_at` set and `outcome` still NULL, which is how `closeUnresidentRows` records a table the server let go. My games and the admin views keep the row. Admin archives are not unloaded: that is a separate change (§10).

**Closed means closed.** Sub-PR 1 makes an archived table refuse, for anyone but an admin: `join`, `SpectateAs`, `Start`, `SetDeck`, `SetSeatDeckID`, `addBot`, a WebSocket upgrade, and any action. The refusal is `ErrGameArchived`, mapped to `409` with `code: "game_closed"`, and the client shows "This table was closed." This fixes gap 1 for admin archives too.

### 5. What closing touches

| Surface | What happens | Where |
|---|---|---|
| Load cap | The table costs 0 from the close on | `entryCostLocked` |
| Three open tables | The creator's slot is free again | `OpenTablesCreatedBy` |
| My games | The row stays. "archived" today; with open question 4, "Closed: no moves for 48 hours" | `mygames.go`, `client/src/lib/myGames.ts` |
| Results feed | Nothing is posted | `results_feed.go` runs only from an engine end |
| Statistics | No record | `recordStatsLocked` runs only from an engine end |
| Rematch | No offer: offers follow an engine end. A rematch's new table is an ordinary lobby table with its own 24-hour clock, and an accept to a closed one is refused (`rematch.go`) | |
| #lfg | The post is withdrawn | `lfg_board.go` |
| Voice room | Removed | `voice_manager.go` |
| Restore point | Kept 7 days, then removed with the room. The replay is kept | §4 |
| Invites | Rows kept (unarchive brings them back). `FindByInvite` already refuses; `join` and spectate refuse after sub-PR 1 | `lobby.go` |
| Reclaim tickets | Dropped | `SetArchived` |
| DM notifier, away clock | Skip archived tables already | `notify.go`, `away.go` |
| Admin views | Listed under archived, with the reason and the last activity time; Unarchive works for 7 days | `adminview`, `client/src/lib/adminViews.ts` |

### 6. Telling people

**A warning before a started game is closed: yes.** A game with real players in it is the one thing here that someone may want back. Twelve hours before the deadline (that is, after 36 hours with no human move), the sweep:

- commits an `EventIdleClose` warning line: "No one has moved for 36 hours. This game closes at 21:14 on Friday unless someone makes a move.";
- stamps `idle_close_at` on the game view (RFC 3339, omitted until warned), from which the table shows a one-line notice above the dock;
- sends a DM (ADR 0162) to each seated person still in the game who has opted in (open question 3 says which switch): "«Table name» will be closed at 21:14 on Friday because no one has moved for 36 hours. https://cmd.labxp.io/#/games/{id}".

Any human move clears the warning and the field. The lead is 12 hours, or a quarter of `IdleGameAfter` when that is shorter, so a short test value does not warn the moment a game starts. The warned flag is in memory: a restart between warning and close repeats the warning once, which is harmless.

**No warning for a lobby table or a bot-only table.** A lobby table costs a minute to make again, and a bot-only table has nobody in it.

**No #mod-log line.** #mod-log records what moderators did (ADR 0156); a timer closing a table is not a mod action, and at launch it could post dozens of lines a day that drown the reports. Each close logs one INFO line (`idle table closed`, with `game_id`, `reason` and the idle time), and the metrics below count them.

**Metrics** (ADR 0123 §3, beside the `cmdctrl_capacity_*` family in `metrics/capacity.go`):

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `cmdctrl_capacity_idle_after_seconds` | gauge | `kind` = `lobby`, `game`, `bots` | the configured threshold; 0 is off, as `cmdctrl_capacity_limit` reads |
| `cmdctrl_capacity_idle_closed_total` | counter | `kind` | tables the sweep closed |
| `cmdctrl_capacity_idle_warned_total` | counter | — | started games warned |
| `cmdctrl_capacity_idle_unloaded_total` | counter | — | closed tables unloaded after 7 days |

The DM rides the existing `cmdctrl_dm_notifications_total{kind, result}` with a new `kind="idle_warning"`. The Grafana Overview's Capacity row gains one panel: tables closed by kind. No alert: a close is the system working.

### 7. Configuration

Two settings for the owner's two thresholds, and a third for §2. They live in `internal/capacity` beside the caps, read by `FromEnv` the same way: a Go duration, unset means the default, `0` turns that cleanup off, and a malformed or negative value fails the boot naming the variable.

| Variable | Default | Meaning |
|---|---|---|
| `CMDCTRL_IDLE_LOBBY_TABLE_AFTER` | `24h` | close an unstarted table after this long with no activity |
| `CMDCTRL_IDLE_GAME_AFTER` | `48h` | close a started game after this long with no human move |
| `CMDCTRL_IDLE_BOTS_AFTER` | `1h` | close a bot-only, unwatched game after this long (open question 1) |

Documented in `docs/env.md` under the Capacity caps bullet. The ADR 0148 caps have no CD step today: they run on the binary's defaults. These get one, `Sync server env (idle cleanup)` in `ci-cd.yml`, reading repository or environment variables of the same names and always writing them, with empty meaning the default, the way the community variables are written. That lets the owner change a threshold, or turn one off, from GitHub without SSH, and lets the preview use short values.

The warning lead (12 hours), the residency grace (1 hour), the sweep interval (5 minutes) and the keep time (7 days) are constants.

### 8. Until this ships: archive by hand

Building starts after the soft launch on 2026-10-30. Until it ships, an admin clears abandoned tables once a day from `/admin/games` (or `/c2-end` in Discord):

- archive a **lobby** table created more than 24 hours ago that has nobody connected;
- archive an **active** game when nobody has been connected to it for about two days. Live now (ADR 0124 §3.4) shows sockets per table, and opening the table as admin shows the log's last line;
- leave a table alone when a player has said they are coming back.

Archive, never delete: archiving is reversible, frees the cap and the creator's slot, and is recorded the way this ADR records an idle close. Watch `cmdctrl_capacity_used{resource="table_load"}` against its limit on the Overview; the `CapacityHigh` alert fires at 80%. `docs/environments.md` gets this as a short runbook entry in sub-PR 4, and the entry is removed when the sweep ships.

### 9. Room for a slow game

The deferred "slow game" option (owner decision 3) is a table setting that lengthens one table's game clock to about 7 days. Nothing here blocks it:

- the sweep reads a table's game threshold through one function, `idleGameAfter(entry)`, which today returns the server's setting and later returns the table's when it is set;
- the warning DM of §6 is already what a play-by-DM game needs: a nudge before the table goes;
- the 12-hour lead scales by the same rule (a quarter of the threshold, at most 12 hours), which a 7-day clock may want longer; that is the slow-game ADR's call.

### 10. Out of scope

- Unloading admin archives after a keep time. The same `Unload` would serve, but it changes what an admin may expect to unarchive. A follow-up issue.
- Ranking the boot admission by `last_active_at` instead of the restore point's capture time. Possible once the column exists; not needed for this.
- Closing ended tables' rooms. An ended table already costs nothing and has no restore point.
- The slow-game setting (§9).

## Calls made here

- **Archive, not an engine end.** Ending the engine would write a statistics record, post a result and arm a rematch for a game nobody finished. The owner asked for the admin close's record, and the admin close is an archive.
- **Agents count as people** for every clock. An agent seat is a person's own process (ADR 0122), and ADR 0177 already treats it as a player. Open question 5 asks the owner to confirm.
- **Admin commits never count.** An admin poking a table to see what is wrong must not keep it open.
- **Chat counts for a lobby table, not for a started game.** The owner's word for a started game is "move". People chatting while they wait for a fourth player are at the table.
- **A client's automatic pass counts as a move.** The server cannot tell it from a click, so an open tab that auto-passes keeps its game open. That is a person's open tab, and the safety belt (ADR 0143 §4.2) clears Skip to my turn at the player's own main phase.
- **The clock counts while the server is down,** softened by the 1-hour residency grace, rather than recording downtime.
- **No warning for lobby or bot-only tables; no #mod-log line** (§6).
- **The sweep does not touch deferred tables** (§3).

## Consequences

- Abandoned tables stop holding the cap and their creators' slots. A table nobody starts is gone within a day, a walked-away game within two, and a game left to its bots within an hour or so.
- A player who returns to a game after 48 hours finds it closed, with a warning they may have missed if they had no DMs on. An admin can unarchive it for 7 days.
- Archived tables are closed for real (gap 1), for admin archives as well.
- Idle-closed tables leave memory and disk after 7 days, so the sweep does not trade a full cap for a full VM.
- New surface: two columns, one event kind, one view field, three settings, one CD step, four metrics, one DM message, and a sweep goroutine.

## Back-out

Set any of the three variables to `0` (through the repository variable, then a deploy, or in the host's env file and a restart) to turn that cleanup off. Tables already closed stay archived and can be unarchived from `/admin/games` for 7 days. The columns are nullable and unused with the sweep off.

## Implementation plan

Each sub-PR targets `develop`, references #4562, and adds or updates the tests named. Run with `scripts/go-docker.sh test ./internal/lobby/... ./internal/ws/...`.

| # | Sub-PR | Files | Tests | Model |
|---|---|---|---|---|
| 1 | **Closed means closed.** Archived tables refuse join, spectate, start, deck, bots, socket upgrade and actions for non-admins; `game_closed` code; client message | `lobby/lobby.go`, `lobby/ws_authorizer.go`, `lobby/http.go` (error map), `ws/hub.go` (action gate), `docs/lobby.md`, `docs/protocol.md`, `client/src/lib/` (refusal text) | A test per refused route on an archived table, written first to show the gap; an admin is still admitted; unarchive restores every route | Sonnet |
| 2 | **The clocks.** `Room.lastHumanAt` and `LastHumanAt`; `touchLocked` at the lobby entry points; seated chat stamp; migration 0029 (`last_active_at`, `closed_reason`); `GameRecord` and `SQLStore` columns; `loadEntry` reads them; `PRIVACY_SCHEMA_VERSION` bump | `ws/room.go`, `ws/away.go`, `ws/hub.go` (chat), `lobby/lobby.go`, `lobby/seat_remove.go`, `lobby/settings.go`, `lobby/host.go`, `lobby/away.go`, `lobby/store.go`, `lobby/store_sql.go`, `lobby/persist.go`, `db/migrations/0029_idle_tables.sql`, `client/src/lib/privacy.ts` | A bot's, a stand-in's and an admin's commit leave `LastHumanAt` alone, a person's and an agent's set it; each listed lobby entry point touches, a spectator, a playmat restamp and a voice restamp do not; the columns round-trip through the store and a restore | Opus |
| 3 | **The sweep and the close.** `lobby/idle.go` (`RunIdle`, `TickIdle`, deadlines, residency grace, bot-only clock, close, unload after 7 days); `capacity.Limits` gains the three durations and `FromEnv` reads them; `RoomManager.Unload`; `EventIdleClose`; metrics; `main.go` wiring; `docs/env.md`, `docs/monitoring.md`, `docs/snapshots.md`; the CD step | `lobby/idle.go`, `capacity/capacity.go`, `ws/manager.go`, `game/` (event kind), `metrics/capacity.go`, `cmd/server/main.go`, `.github/workflows/ci-cd.yml`, `deploy/monitoring/` (panel) | With a fake clock: a lobby table closes at 24 h and not at 23 h 55; a touch resets it; a game closes at 48 h after a warning, and a move clears the warning; a bot-only unwatched game closes at 1 h and a watching seat holds it; nothing closes within an hour of residency; `0` turns each off; a closed table costs 0, frees the creator's slot, records `closed_reason` and `outcome` NULL, counts `closed` once, posts no result and writes no stats; unload after 7 days keeps the replay and writes the row ended; `FromEnv` rejects a malformed duration | Opus |
| 4 | **Telling people.** `idle_close_at` on the view and the table notice; the warning DM (`notify` message kind); My games and admin views show the reason and the last activity; the launch-week runbook entry is replaced by the sweep's | `protocol/protocol.go`, `ws/` (stamp), `notify/message.go`, `notify/notifier.go`, `adminview/merge.go`, `lobby/mygames.go`, `client/src/` (table notice, `lib/myGames.ts`, `lib/adminViews.ts`), `docs/protocol.md`, `docs/environments.md` | The warning DM goes once, only to opted-in seated people still in the game, never to a bot, guest, agent or spectator; the view field appears at the warning and clears on a move; client unit tests for the notice and the My games label | Sonnet |
| 5 | **Bot-only play-out** (only if open question 2 is yes). A bot-only, unwatched game runs its bots at `BotPaceFast` until a person connects | `aiseat/runner.go`, `lobby/idle.go` | The pace switches on and off with the condition; a stalled bot-only game still closes at the 1-hour backstop | Opus |

The branch E2E is not needed: the sweep is server-side and driven by a fake clock in Go tests. Sub-PR 1 changes what a client sees on a refused socket, so it runs the branch E2E.

Until sub-PR 3 is deployed to production, §8 is the procedure.

## Open questions for the owner

1. **Close a bot-only game after an hour with nobody watching?** When every person at a started game has conceded, lost or been handed to a stand-in, and no person's seat has a connection, the bots play on at the table's pace. **Recommended: yes, after 1 hour, as a third setting (`CMDCTRL_IDLE_BOTS_AFTER`, `0` off).** It frees the cap and the CPU the same evening, and a stand-in player who comes back within the hour still gets their seat.
2. **Play a bot-only game out first?** The alternative to waiting is to switch its bots to the fast pace while nobody watches, so the game reaches a real result (the statistics record and the results post) within minutes, with the 1-hour close as the backstop for a stall. **Recommended: yes.** By the rules the game goes on among the remaining players (CR 800.4a), and finishing it is the faithful ending; closing it unfinished throws away a result the conceding players' statistics would otherwise get. It is sub-PR 5, and can follow the rest.
3. **Which DM switch covers the warning?** **Recommended: "Your turn".** A player who asked to be told when a game waits on them expects to be told it is about to close; a fourth switch would be off for everyone, which defeats the warning. "Table ready" and "Table filled" are about lobbies.
4. **Should My games say why a table closed?** Today an archived table reads "archived". **Recommended: yes, from `games.closed_reason`:** "Closed: no moves for 48 hours", "Closed: not started within a day", "Closed: only bots left". A player whose game vanished should not have to guess whether a moderator closed it.
5. **Do agent seats count as people?** **Recommended: yes,** as ADR 0177 already does. An agent's moves keep a game open the way a person's do. The cost is that an agent left running against bots keeps its table for as long as it runs, which a person playing bots can do too.
6. **Is 7 days the right window to unarchive an idle-closed table?** After it the table leaves memory and disk, and only the row and the replay remain. **Recommended: 7 days.** It covers a weekend and a "where did my game go" message, and it keeps memory bounded on a 4 GB VM.
