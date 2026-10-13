# ADR 0148 — Capacity limits for a public launch: a weighted table-load cap, a connection cap, creation rates, and a restore pass that respects them

**Status:** Accepted · 2026-10-12 · Backlog (operability, S70 load test)
**Issue:** [#2977](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2977); the measurements are on [#2978](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2978).
**Owner decision (2026-10-11):** don't grow the VM yet. A soft cap of about **30 games (~120 players)** is the target. Human tables cost less CPU than the four-bot tables measured on #2978, so bot seats weigh more against the cap than human seats. The restore pass on boot must respect it too.
**Numbering:** every ADR file on `origin/develop` and on all eleven remote heads was listed; the highest number anywhere was **0147**, so this is **0148**. The claim comment on #2977 reserved it.
**Builds on:** [ADR 0041](0041-game-persistence.md) and [ADR 0044](0044-surviving-a-deploy.md) (the restore pass), [ADR 0123](0123-monitoring-metrics-logs-dashboards-and-alerts.md) (metrics), [ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md) §3.4 (Live now), [ADR 0110](0110-remember-me.md) §5 item 4 (who may create a table).

---

## Context

Prod is one 2 vCPU / 4 GB VM behind a Cloudflare tunnel, and until this ADR the server accepted tables, sockets and creations without limit. The prod load test (#2978, 2026-10-11) ramped four-heuristic-bot tables, each with three spectator sockets:

| tables | healthz p95 | creates |
|---|---|---|
| 0 | 56 ms | |
| 20 | 155 ms | fine |
| 30 | 332 ms | fine, bots 14% slower per table |
| 31-36 | | time out; the process was killed at about 36 |

CPU saturates first (the bots' decision loops), then memory (about 40 MB per table on top of about 1.25 GB idle, mostly the card index) tips it over. The restart then took **about seven minutes**, and on boot the restore pass **brought back 29 of the 36 tables that had killed it**. With no cap that is a crash loop under real demand: one popular post makes a few dozen tables, the server dies, and the restart rebuilds exactly the load that killed it.

## Decision

### 1. What is capped: three things, one package

`server/internal/capacity` holds the limits, their defaults and their arithmetic, and nothing else. The lobby enforces the table-load cap and the creation rates; the hub enforces the connection cap. Every one is an environment variable, every one has a default, `0` switches a cap off, and a malformed value fails the boot naming the variable (a typo in a cap must not leave the server quietly uncapped).

| Variable | Default | Meaning |
|---|---|---|
| `CMDCTRL_MAX_TABLE_LOAD` | `180` | The weighted load ceiling (below). `0` is no cap. |
| `CMDCTRL_LOAD_WEIGHT_TABLE` | `2` | What an open table costs before anyone sits at it. |
| `CMDCTRL_LOAD_WEIGHT_HUMAN_SEAT` | `1` | What a human (or agent) seat adds. |
| `CMDCTRL_LOAD_WEIGHT_BOT_SEAT` | `2` | What a bot seat adds. |
| `CMDCTRL_MAX_WS_CONNECTIONS` | `400` | Live WebSocket connections. `0` is no cap. |
| `CMDCTRL_CREATE_ACCOUNT_BURST` / `_INTERVAL` | `3` / `30s` | Per-account table creation: a burst of 3, then one more per 30 s. (This is the rate `createLimit` always had; it is now configurable.) |
| `CMDCTRL_CREATE_IP_BURST` / `_INTERVAL` | `6` / `20s` | The same per client IP. |

### 2. The load: weighted, and why these weights

The load is the sum, over every unended, unarchived table, of `TableWeight + humans x HumanSeatWeight + bots x BotSeatWeight`. An ended table costs nothing (it runs nothing), and so does an archived one (its runners are stopped). A practice table is a table and costs like one (a human and a bot).

The defaults come from three constraints, and all three are met by one number:

- **The owner's target.** 30 games of four humans (120 players) is `30 x (2 + 4 x 1) = 180`: the cap's default, exactly.
- **The measured ceiling.** A four-bot table is `2 + 4 x 2 = 10`, so 180 holds **18** of them, 58 % of the ~31 the VM died at. The headroom is deliberate: the test's tables carried three spectator sockets each, a deploy rebuilds every room while serving, and CPU, not memory, was already degrading at 30.
- **Human tables are cheaper.** A bot seat is two human seats (an in-process runner with a legal-move enumerator for the whole game, which is what saturated the CPU); the per-table base of 2 carries the memory that humans and bots share (~40 MB each).

The weights are **tunable**. They are an estimate from one load test of one shape (all bots, heuristic tier, 3 spectators). The first real evidence a human-heavy night produces (CPU and memory per table in Grafana against `cmdctrl_capacity_used`) should move them, and no code change is needed: set the variables. Raising the VM (Terraform `memory_mb` / `cpu_cores`) and re-running the test is the other lever, and then the cap moves with it.

### 3. The cap is soft, and gates only new load

It refuses what would **introduce** load and never touches a table that exists:

| Refused at the cap | Never refused because of it |
|---|---|
| Creating a table (`POST /games`) | A human taking a seat at an existing table |
| Opening a practice table (`POST /games/practice`) | `Start`, and every action in a running game |
| Seating a **bot** at an unstarted table | A seated player reconnecting |
| A join or spectate while the connection cap is full | The restore pass for an admitted table |

So the load can sit above the cap (people keep sitting at tables that were admitted), and that is intended: the alternative, refusing a fourth human at a table three are already waiting at, is the cap hurting exactly the people it should leave alone. A create reserves its cost under the lobby lock from the check until its entry is registered, so a burst of concurrent creates cannot all pass a check made before any of them landed (`TestConcurrentCreatesCannotOvershootTheCap`: 40 racing creates against room for 5 make exactly 5).

### 4. The connection cap exempts seats and admins

`Hub.ServeWS` turns a socket away when the live count has reached `CMDCTRL_MAX_WS_CONNECTIONS` **and** the caller has no seat and is not an admin. A seated player is mid-game or reconnecting to a table that is already counted; an admin is the operator looking at a full server. The check runs before the initial snapshot is built, which is the expensive part of an admit, so a flood of refused sockets costs a handshake and one frame each. It is a soft cap (check and admit are not one atomic step), so a burst of simultaneous newcomers can overshoot it by the burst.

The front doors check it too: `POST /games/{id}/join`, `POST /join` and `POST /games/{id}/spectate` answer `server_full` when the count is at the cap, rather than seat someone whose socket would then be refused. The Discord OAuth callback join is not gated (a redirect flow has no place to show a message); its socket is refused at the hub like any other newcomer's.

### 5. The refusal is a stable code and a friendly sentence

An HTTP refusal is **503** with `{"error": "The server is full right now. Please try again in a few minutes.", "code": "server_full"}` and `Retry-After: 30`. The client matches the **code**, never the sentence, and owns the sentence it shows (`client/src/lib/serverFull.ts`), so a proxy that swallows the body still gets the friendly message and the four places that show a create or join failure cannot word it four ways.

A browser cannot read the body of a refused WebSocket handshake, so a socket refused at the cap is **accepted**, sent one `error` frame with the same code and message, and closed **1013** (try again later). 1013 is not terminal: the client keeps its backoff ladder, and the connection banner says "The server is full right now" instead of "Connection lost" until a snapshot arrives (`docs/protocol.md`).

Table-creation rate refusals stay what they were, **429** with `Retry-After`, and are counted.

### 6. The creation rates: per account, and per IP for everyone else

Creating a table needs a signed-in person (ADR 0110 §5 item 4), whose bucket is the existing per-account one, now configurable. The per-IP bucket is what bounds **guests**, who cannot create a table but do open practice tables (each a bot runner), and a machine minting accounts to get round the per-account bucket. It is spent by `POST /games` (the shared admin token, which is the Discord bot and an operator, is exempt from both, as before) and by `POST /games/practice`. Behind Cloudflare it keys on `CMDCTRL_TRUST_FORWARDED`'s hop like every other IP limiter.

### 7. The restore pass respects the cap

`RestoreFromDisk` was the crash loop's second half. It now asks the lobby, with every readable restore point in hand, which tables there is room for.

**Who comes back first**

1. A table with a **human** at it, before a table with only bots. A person is waiting at it; a bot table is the cheap thing to lose.
2. Within each group, the **most recently active** first (the restore point's capture time is the table's last committed move).
3. **First-fit** down that order: a table too big for what is left is deferred, but a smaller one behind it may still fit.

An ended table, an archived table and a table with no `games` row cost nothing and always come back (an ended table runs nothing, an archived one has its runners stopped, and a table with no row is dropped by the pass anyway).

**What happens to the rest: nothing.** A deferred table's restore point stays on disk, byte for byte, and its `games` and `seats` rows are untouched. They are **not** ended: `closeUnresidentRows` (#2987) keeps them out, by the same mechanism that keeps a `schema_too_new` roll-forward table. That is the point of the choice. Ending a deferred table throws away a game the server merely had no room for yet; leaving it costs a row that reads `active` with no room behind it until the next boot, which `DELETE /games/{id}` already handles (#2987's unresident delete). A later boot, or this one's successor with a raised cap, brings it back, and a player whose table is deferred sees what they see for any table the server has lost track of ("game not found"); the boot log names each by id, and `cmdctrl_boot_restore_deferred_games` counts them.

There is no runtime "admit the deferred ones" route. A deferred table comes back at the next restart. A hot-restore button is a possible follow-up if deferral turns out to happen often; the load test suggests it should happen only after a crash, which is the case it is for.

### 8. What an operator sees

- **Prometheus** (alongside the existing ones, [ADR 0123](0123-monitoring-metrics-logs-dashboards-and-alerts.md) §3): `cmdctrl_capacity_limit{resource}` and `cmdctrl_capacity_used{resource}` (`table_load`, `ws_connections`; a limit of 0 is "off"), `cmdctrl_capacity_refusals_total{reason}` (`table_load`, `bot_seat`, `ws_connections`, `join_connections`, `create_account`, `create_ip`), and `cmdctrl_boot_restore_deferred_games`. `cmdctrl_ws_upgrade_rejections_total` gains `reason="capacity"`.
- **Grafana Overview:** a Capacity row: load against its cap, connections against theirs, and requests turned away by reason.
- **Live now** ([ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md) §3.4): `GET /admin/live` gains `capacity` (`load`, `load_limit`, `connections`, `connection_limit`), shown as two tiles, "Table load 14 / 180" and "Connections 9 (no cap)".

## Consequences

- A public spike now ends in a polite message instead of a dead server, and a crash no longer rebuilds its own cause. The 7-minute restart is unchanged (it is the card index load), but it no longer ends in a second collapse.
- The default cap is conservative against the measured ceiling and exact against the owner's target; the first real traffic should tune the weights, not the code.
- The load can exceed the cap by the humans who sit at admitted tables. That is intended and bounded: at most `MaxPlayers` per admitted table.
- A deferred table is invisible to its players until the next boot. Acceptable for the case it exists for (post-crash); revisit with a hot-restore route if it ever happens in normal running.
- An alert on `cmdctrl_capacity_used / cmdctrl_capacity_limit` is deliberately not added: the right threshold comes from the first week of data.

## Back-out

Set `CMDCTRL_MAX_TABLE_LOAD=0` and `CMDCTRL_MAX_WS_CONNECTIONS=0` to turn the caps off with no deploy; the creation rates keep their old defaults. The restore pass admits everything when the load cap is off.
