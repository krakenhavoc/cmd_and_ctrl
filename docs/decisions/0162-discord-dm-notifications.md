# ADR 0162 — Discord DM notifications

**Status:** Proposed · 2026-10-13 · Backlog
**Issue:** [#3320](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3320)
**Owner decisions:** the two answers of 2026-10-13 are quoted under [Owner decisions](#owner-decisions-2026-10-13). They are binding. Calls this ADR makes on its own are listed under [Calls made here](#calls-made-here) for the owner to overturn in review, before slice 2 lands.
**Numbering:** checked with ADR 0161's sweep (12 remote heads, highest 0160). ADR 0161 took 0161; this is **0162**.
**Builds on:** [ADR 0051](0051-user-database.md) decision 5 (`discord.Bot.SendDM`), [ADR 0105](0105-legal-action-highlights.md) (which seat has something to do), [ADR 0110](0110-remember-me.md) §4 (the client's synced settings), [ADR 0123](0123-monitoring-metrics-logs-dashboards-and-alerts.md) (metrics), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the notice and the deletion inventory), [ADR 0161](0161-an-lfg-board-in-discord.md) (a table that fills from #lfg).

---

## Context

A Commander game stalls when the player it is waiting on has gone to make tea. Nobody else can move, and the only signal today is the table itself, in a tab the player may have closed. Slower, asynchronous play is not possible at all unless people sit and watch.

Every signed-in player has a Discord account, and the server can already DM one: `discord.Bot.SendDM` opens a DM channel and posts with the bot token (ADR 0051 decision 5), which is how `/c2-invite-dm` and the in-app DM invite work. The bot shares the CMD & CTRL server with its members, which is what Discord requires before a bot may DM someone, unless the person has turned off DMs from server members.

What the server knows:

- **Who must act.** The engine knows which seat it is waiting on: the priority holder, or the seat that owes a choice (a mulligan, a target, a search, attackers or blockers). The `legal_moves` frame (ADR 0105) is computed from the same state.
- **Who is connected.** The hub tracks every socket by game and seat (`Hub.LiveSockets`). It does not know whether the tab is in front: the client listens for `visibilitychange` only to pace its heartbeat.

## Decision

### 1. Off until the player opts in

Owner answer 1.

Three kinds of DM, each a separate switch in **Settings → Notifications**, all off by default:

| Kind | Sent when |
|---|---|
| **Your turn** | the game has been waiting on you for 2 minutes while you are away (§2) |
| **Table ready** | a table you sit at has filled (every seat taken) or started, while you are away |
| **Table filled** (hosts) | a table you host has filled, while you are away; the natural partner of an #lfg post (ADR 0161) |

The in-app **DM invite** (ADR 0051 decision 5) is unchanged: someone chose to send it, and it is not a notification.

**The nudge.** The first time a signed-in player is seated at a table while the feature is on and every switch is off, the table shows a one-line suggestion: "Get a Discord DM when it's your turn?" with **Turn on** and **Not now**. Dismissing it is remembered in the client's synced settings (ADR 0110 §4) and it never shows again.

### 2. Away, and how often

Owner answer 2: two minutes.

- **Away** means, for that seat, **no open socket**, or every open socket reports its tab hidden.
- **Hidden** is new: the client sends `{"type": "visibility", "visible": false|true}` on `visibilitychange` and once on connect. The hub keeps the latest value per client. An older client that never sends it counts as visible, so it never gets a DM it would not have before.
- **Your turn** fires when the engine has been waiting on that seat **continuously for 2 minutes** and the seat is away for the whole of that time. Acting, or coming back, resets the clock.
- **At most one DM per player per table per 10 minutes**, across every kind. A player who stays away through several decisions gets one DM, not one per priority pass.
- **Never** to a bot, an agent seat (ADR 0122), a guest, a spectator, or for a practice table.

### 3. The message

Plain text, one line and a link, nothing from the game's hidden information:

> It's your turn at **«Table name»** (turn 7). https://cmd.labxp.io/#/games/{id}

> **«Table name»** is full and ready to start. https://cmd.labxp.io/#/games/{id}

The link opens the table; a signed-in player's seat is rejoined the usual way (`POST /me/games/{id}/session`). Table names have already passed the ADR 0075 §2.6 filter.

### 4. Who sends, and failures

- **The server sends**, through `discord.Bot.SendDM`, from a small notifier goroutine fed by the lobby: the server is where the engine, the hub and the preferences are. The gateway bot is not involved.
- **Discord refuses the DM** (`50007`, "Cannot send messages to this user": DMs from server members off, the bot blocked, or the person left the server): the notifier records `dm_failed_at` and stops sending to that person. Settings shows "Discord refused the last message. Check that you allow DMs from the CMD & CTRL server's members." Turning any switch off and on again clears it.
- **429** waits `retry_after` (capped) once; anything else is logged at WARN and dropped. A notification is never retried late: a "your turn" that arrives after you have moved is worse than none.
- **No bot token** (every environment but production): the feature is off, the Notifications section is hidden, and the routes answer 503.

### 5. Storage and API

- **A new table**, not the settings blob. ADR 0110 §4 keeps `user_settings` opaque to the server on purpose, and the server must read these.
  `notification_prefs(user_id TEXT PRIMARY KEY REFERENCES users(id), turn INTEGER NOT NULL DEFAULT 0, table_ready INTEGER NOT NULL DEFAULT 0, table_filled INTEGER NOT NULL DEFAULT 0, dm_failed_at INTEGER, updated_at INTEGER NOT NULL)`.
- **Rate-limit state** (the 2-minute clocks and the 10-minute windows) lives in memory. A restart forgets it, which at worst sends one extra DM.
- `GET /me/notifications` and `PUT /me/notifications` (the three booleans). The response also carries `available` (the bot token is set) and `dm_failed` (§4).
- **Account deletion** (ADR 0149) deletes the row; the inventory lists it.
- **Metrics** (ADR 0123): `cmdctrl_dm_notifications_total{kind, result}` with `result` = `sent`, `refused`, `error`.

### 6. Privacy

The notice gains a short paragraph: if you turn on notifications, CMD & CTRL asks Discord's bot to send you direct messages naming the table and its turn number; the setting is stored with your account and deleted with it; Discord holds the messages under its own policy.

## Consequences

- **Slow tables keep moving,** and asynchronous play becomes possible for people who opt in.
- **Nobody is messaged who did not ask.** The price of off-by-default is that few people will find it; the nudge is the answer.
- **The client gains one frame** (`visibility`). The protocol doc lists it.
- **Production only,** like DM invites.

## Owner decisions, 2026-10-13

1. **Default:** off, opt in per kind, with a one-time suggestion at the table. Recommended option, chosen.
2. **Away rule:** "split the difference at 2 mins" between the offered 3 and 1 minutes.

## Calls made here

- **One DM per player per table per 10 minutes.** The owner split the away time between the offered 3 and 1 minutes; the rate window is split the same way, between the offered 15 and 5.
- **Hidden counts as away.** A tab left open on another screen is the commonest way to miss a turn; a socket alone would not catch it.
- **No late retries.**
- **Three kinds, not more.** Game over, chat mentions and spectator joins are left out until someone asks.

## Slices

| # | Slice | Model | Needs |
|---|---|---|---|
| 1 | This ADR | — | — |
| 2 | **Server:** migration and the prefs store, `GET`/`PUT /me/notifications`, the `visibility` frame in the hub, the notifier (waiting-on-seat clock, away rule, rate window, `SendDM`, 50007 handling), metrics, account-deletion inventory; docs/protocol.md and docs/lobby.md | Opus | 1 |
| 3 | **Client:** send `visibility`; Settings → Notifications (three switches, the refused hint); the one-time nudge at the table; privacy paragraph | Sonnet | 2 |
