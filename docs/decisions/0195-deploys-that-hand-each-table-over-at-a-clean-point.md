# ADR 0195 — Deploys that hand each table over at a clean point: a short two-process overlap, not hours of blue/green

**Status:** Accepted (2026-10-15) · proposed 2026-10-14 · Backlog (operability). The owner's decisions are at the end.
**Issue:** [#3039](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3039). Related, both open: [#3037](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3037) (measure what a prod deploy costs a live table) and [#3038](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3038) (wait for a clean restore point before restarting prod).
**Numbering:** checked on 2026-10-14 across every remote head (`develop`, `main`, `cost-ledger`, `docs/adr-demonstrate-a-loop`, `docs/issue-audit`, `feat/3536-graveyard-exile-cast-cost`, `feat/3546-target-crewed-or-saddled`, `feat/3552-pilot-token-makers`, `feat/3560-skip-extra-turns-ugins-nexus`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2947-resolution-copy-casts`, `fix/2950-darkness-replacement-continuation`, `fix/3080-commander-trade-wait`, `fix/3724-access-maze-x`, `fix/caddy-reload-admin-off`, `issue-3016-reuse-trigger-order`, `perf/clone-3136`, `wip/836-one-click-default`, `pr/2326`) and the open PRs. `develop` ends at 0192; 0193 is #3699 (`docs/adr-demonstrate-a-loop`) and 0194 is #3730 (`perf/clone-3136`). 0190 is a gap below merged numbers that #3699 and #3730 both left alone as possibly reserved, so this takes **0195**.
**Supersedes in part:** [ADR 0044](0044-surviving-a-deploy.md) decision 1 ("cold restore, not drain … revisit only if the topology gains a second node"). Decisions 2–8 stand.
**Builds on:** [ADR 0041](0041-game-persistence.md) (restore points, decision 5's version-skew refusal), [ADR 0044](0044-surviving-a-deploy.md) (HMAC sessions, close 1001, restore generation, no undo across a restore), [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (VM size, per-table memory, the cap and its hot restore route), [ADR 0023](0023-develop-environment.md) (dev rehearses prod), [ADR 0159](0159-per-table-discord-voice-rooms.md) and [ADR 0161](0161-an-lfg-board-in-discord.md) (boot reconcilers), [ADR 0181](0181-deploy-credentials-fail-closed.md) (deploy runners).

---

## Context

### The problem

Every production deploy restarts the one game server. A live table pays three costs:

1. **A gap.** The socket closes with 1001, the client backs off and redials, and nothing answers until the new process has loaded its card index and run the restore pass. ADR 0148's load test saw a seven-minute restart under load; a normal boot is shorter, but nobody has measured it on prod (#3037).
2. **A rewind.** A table whose last restore point is behind its live `seq` (it holds a continuation, most often an open prompt: `ChoiceResumeFrames`, ADR 0044 decision 6 and the #3039 design note) comes back at that restore point. Players replay what they did.
3. **Lost undo.** The undo stack is in-memory clones and is not persisted (ADR 0044 decision 6), so every table loses its undo history.

Players also lose the table outright when the restore point is from a newer schema (rollback) or when the restore pass defers it under the cap (ADR 0148 §7).

Public beta makes deploys more visible: more tables are live at any given hour, and prod deploys follow every `develop` → `main` promotion.

### Today's flow, from the repo

- **Topology** (`docs/environments.md`): prod and the dev preview are two identical 2 vCPU / 4 GB VMs on one Proxmox node. Each runs `cmd-and-ctrl.service` as `cmdctrl`, bound to `127.0.0.1:8080`, data in `/var/lib/cmd_and_ctrl/data` on a dedicated disk, env in `/etc/cmd_and_ctrl/env`. The server's systemd unit comes from the HomeLab cloud-init template; only the Discord bot's unit (`deploy/cmd-and-ctrl-bot.service`) is installed by CD from this repo.
- **Front door:** Cloudflare tunnel → Caddy `:80` (`deploy/Caddyfile`) → `reverse_proxy 127.0.0.1:8080` for the API paths (`/ws`, `/games/*`, `/admin/*`, …), `file_server` from `/var/www/cmdctrl-client` for the rest. Caddy runs with `admin off`, so CD applies a new Caddyfile with `systemctl restart caddy`, which drops every proxied WebSocket. Caddy trusts loopback's forwarded headers since #3354, and the server reads `CF-Connecting-IP` for per-IP limits.
- **Deploy** (`.github/workflows/ci-cd.yml`, `cmdctrl-cd-dev` on `develop`, `cmdctrl-cd-prod` on `main`, the same steps): `scripts/keep-previous-release.sh` sets the live release aside (#3339) → rsync the server binary over the live one → (prod only) the bot binary and unit → `rsync --delete` the client into the web root → maybe the Caddyfile → sync env values with `scripts/set-server-env.sh` → `systemctl reset-failed; systemctl restart cmd-and-ctrl` → restart the Discord bot (prod only) → check `GET /config` reports the right env → `scripts/smoke-release.sh` → backup and monitoring setup. `rollback-prod.yml` runs `scripts/rollback-release.sh`, which swaps the `.prev` release back and restarts, and does not touch the database.
- **Shutdown** (`server/cmd/server/main.go`): on SIGTERM, `srv.Shutdown`, `hub.Shutdown` closes every socket with 1001, and `ws.LogShutdownCensus` logs each room's `ShutdownReport` (`SeqBehind`, restore-point age). Nothing is written at exit; the restore point is written on every `Apply` when the game is restorable (`captureLocked`).
- **Boot:** `db.Open` migrates SQLite (`<data>/db/cmdctrl.sqlite`, WAL, `busy_timeout(5000)`) and refuses a database newer than the binary (`ErrSchemaTooNew`). Then the card index loads, and `Lobby.RestoreFromDisk` restores every `<data>/restore/<id>.json` the cap admits, with its `<data>/lobby/<id>.json` metadata, relaunching bot runners. The voice and #lfg managers then run a **boot reconcile that adopts the rooms and posts tables claim and deletes the ones no table claims**.
- **Restore points:** `SnapshotSchemaVersion` is 9, `minRestorableSchema` is 1. A file from a newer schema is refused and kept (ADR 0041 decision 5). The restore generation travels in the file and on the snapshot frame (ADR 0044 decision 5). ADR 0148's amendment added `POST /admin/capacity/restore-deferred`, which restores a table from disk **into a running server**, skipping any table that already has a room. That is the receiving half of a hand-off, already in production.
- **Sessions:** HMAC-signed with `CMDCTRL_SESSION_KEY` from the shared env file (ADR 0044 decision 3), so any process on the host validates them. Revocation is a database timestamp (ADR 0051 decision 6).
- **Memory** (ADR 0148): about 1.25 GB idle, mostly the card index, plus about 40 MB per table (#3730 measures ~21 MB of that as the undo ring). The cap targets about 30 tables.

## Options

Each option is judged on the same seven concerns: WebSocket sessions, the lobby, the database and migrations, Discord singletons, snapshot schema versions, bots and MCP seats, and what the VM can hold.

### A. Stay cold, gate the restart on a clean moment (#3038)

Poll a readiness summary built from `ShutdownReport`, restart when every room is clean, give up after a deadline.

- Cheapest. No change to any of the seven concerns: it is today's restart, timed better.
- **It does not scale with tables.** The gate needs *every* table clean at once. With 20 tables, each mid-prompt some share of the time, the joint chance shrinks fast, so the deadline usually fires and the gate becomes a delay. #3038 already names the race (clean when polled, dirty a second later).
- The gap (boot time) and the lost undo remain for every table.

### B. In-process drain with an old-binary listener hand-off

The old process passes its listening socket to the new one (systemd socket activation, or `SO_REUSEPORT` on the same address, "tableflip" style), stops accepting, and keeps serving the connections it already holds until their games end.

- **WebSockets:** existing sockets stay on the old process. Any *new* socket for an old game (a reconnect after a blip, a spectator, a refresh) lands on the new process, which does not have the game. So B needs routing by game id anyway, and becomes C or D below. On its own it is not an option; it is the cut-over mechanism for one.
- What it does contribute: `SO_REUSEPORT` lets two processes share `127.0.0.1:8080`, so Caddy, the smoke script, `LOCAL_PORT` and the bot's `CMDCTRL_SERVER_BASE_URL` keep one address. Closing the old listener sends every new connection to the new process. A connection still in the closed socket's accept queue is reset; the client's backoff and Caddy's dial retry cover it.

### C. Literal blue/green: two systemd instances, routing by game id, old games finish on the old binary

`cmd-and-ctrl@blue` and `@green` on two ports. New tables go to the newest; each existing game's traffic goes to the instance that owns it until the game ends or a maximum age expires, after which the old instance falls back to a cold restart of its remaining tables.

- **Routing:** the game id is in `/ws?game=<id>` and `/games/{id}/…`. Caddy cannot route on a changing set of ids without a config change, and a config change is a Caddy restart (admin is off) that drops every socket. So the router must be dynamic and outside Caddy: a small router process, or the newest instance proxying unowned games to the old one over loopback (`httputil.ReverseProxy` carries WebSocket upgrades). Invite links carry no instance, so the router must hold the id → instance map.
- **WebSockets:** best case for players: no blip at all on an old table.
- **Lobby:** `Lobby.games` is in-memory per process. The lobby list, `/c2-games`, the admin views and the Live-now view each see half the tables for hours unless every reader merges two processes.
- **Database and migrations:** both processes write one SQLite file (fine in WAL with the busy timeout), but the old binary runs for **hours** against a schema the new one migrated. Every migration would have to stay compatible with the previous binary for that long, and `ErrSchemaTooNew` means the old instance can never restart once the new one migrates.
- **Discord singletons:** the voice and #lfg boot reconcilers delete rooms and posts that no table in *their* process claims. A second process booting would delete the other's voice rooms and posts. Every reconciler, sweeper (admin mode, reports) and the hourly `VACUUM INTO` backup would need one owner across processes for hours.
- **Snapshot schema:** each process writes only its own games' files; restore at boot must skip files the other owns. The old binary never reads a newer file, so ADR 0041 decision 5 is not hit, except on rollback.
- **Client assets:** CD does `rsync --delete` into one web root. A player who refreshes on an old table gets the new client against the old server, for hours. Fixing that means versioned client directories and routing them too.
- **Bots and MCP seats:** in-process runners stay with their game; MCP seats are WebSocket clients and follow routing.
- **Memory:** two idle processes are ~2.5 GB of a 4 GB VM for hours. `develop` deploys several times a day, so dev would stack a third generation unless a deploy forces the oldest over, which is the cold path again.
- **Fixes don't reach running games.** Many deploys are card and engine fixes; under C the table that hit the bug keeps it until it ends.

### D. Per-table hand-off at a clean restore point, with a short overlap (recommended)

Start the new process beside the old. Once it is booted and ready, cut new connections over to it. Then the old process hands each table across **at that table's own clean point** (its restore point is at its live `seq`), and exits when it holds none. A table that is never clean within a deadline is handed over at its last restore point (today's rewind, for that table only).

One hand-off, under the room's lock:
1. The old room is clean (`SeqBehind == 0`, no continuation). If not, wait for the next commit and test again.
2. It refuses further actions with a retryable "table is moving" error, stops its bot runners, and confirms its restore point is the live `seq`.
3. It tells the new process (control socket below), which restores that one file with the hot-restore path (`RestoreRoomsAdmitting` for one id plus `adoptRestored`) and relaunches the bots.
4. The old room closes its sockets with 1001 and unregisters. The client redials, lands on the new process, and gets a snapshot at the same `seq`.

Per concern:

- **WebSockets:** one reconnect per table, no rewind, no boot gap (the new process booted before anything moved). The restore generation bumps, so the client discards and re-renders (ADR 0044 decision 5), which is correct. Sessions validate on both sides (shared HMAC key). A dedicated close code ("moved, reconnect now") would skip the first ~500 ms of backoff; optional.
- **Routing:** only during the overlap, and only for tables not yet moved. The new process holds the shared address (`SO_REUSEPORT`, option B's contribution); a request for a game it does not have, but the old process does, is proxied to the old process's private unix socket. No Caddy change.
- **Lobby:** lobby-phase tables (not started) are trivially clean and move first. The new process lists what it holds; for the minutes of the overlap the lobby list can miss a few tables, which then appear as they arrive. Acceptable, or merge the old process's list over the control socket.
- **Database and migrations:** the new binary migrates at boot while the old one serves for **minutes**, not hours. The rule becomes: a migration must not break the previous binary while both run (add, don't rename or drop in the same release). A release that can't keep it marks itself and the deploy falls back to today's cold restart. `ErrSchemaTooNew` is unchanged: the old process never reopens the database.
- **Discord singletons and sweepers:** one owner, by a lock file (`flock` on `<data>/run/singleton.lock`). The old process holds it while it serves alone; at cut-over it stops its reconcilers and sweepers and releases it; the new process starts them **after the last table has arrived**, so its boot reconcile sees every table and deletes nothing that is live. Reconcilers are idempotent (`voice_manager.go`: "a missed or repeated notify costs nothing"), so a few minutes without one only delays a voice permission. The results feed and DM invites are per game or per request and need no owner. The separate Discord bot unit is untouched; its admin calls reach whichever process holds `:8080`.
- **Snapshot schema:** hand-off is one-way, old → new, the direction the fixture corpus already guarantees (ADR 0044 decision 7). A newer file is never offered to an older binary. Rollback stays a cold restart (`rollback-prod.yml`), with ADR 0041 decision 5's refusal as today.
- **Bots and MCP seats:** runners stop on the old side and relaunch on the new one, as the restore pass already does. MCP seats reconnect like any client.
- **Capacity:** a hand-off is not new load and is admitted without the cap check; new creations on the new process count the old process's remaining load (read over the control socket), so the sum stays under ADR 0148's cap.
- **Client assets:** the new client lands at the start of the overlap, as today. Its skew with the old server lasts minutes and only until each table moves.
- **Memory:** peak is two idle processes plus every table once: ~2.5 GB + ~1.2 GB at 30 tables, against 4 GB. Tight; #3037 should measure RSS, and the old process should return freed memory as tables leave (`debug.FreeOSMemory` after each hand-off). See open question 3.
- **Undo:** each table still loses its undo history at its hand-off (ADR 0044 decision 6). The ring is in-memory clones; carrying it is out of scope here.
- **Fixes reach running games** within minutes, which C cannot do.

Control channel: each process listens on a private unix socket (`/run/cmd_and_ctrl/<release>.sock`, `cmdctrl` only, never on the public mux) for `ready`, `drain`, `tables`, `handoff/{id}` and `load`. CD talks to it over ssh, the processes talk to each other.

### E. Cold restart that replays the actions since the restore point

Record every input after the last restore point (actions, prompt answers, the RNG position) and replay them at boot, so even a mid-prompt table comes back exactly. This removes rewinds on a *crash* as well, which no deploy-only option does. But it needs a fully deterministic engine replay of continuation-holding states and a journal on every action, and it does nothing for the boot gap or undo. It belongs with the remaining census work (making `ChoiceResumeFrames` data), not with deploys. Noted, not chosen.

### F. Shared state and failover (the #3039 design note)

Postgres for users, lobby and restore points, ownership leases with the restore generation as the fencing token, routing by owner, two VMs. It makes blue/green a planned failover and survives a powered-off host. It needs a second database and a second VM, which AGENTS.md §2 and §8 rule out without an ADR, for a scale PLAN.md does not name. The design note's own recommendation stands: not yet. D is shaped so it doesn't block F: the hand-off verbs (`drain`, `handoff/{id}`) are the planned-handoff half of F.

### G. Outside the host

A Cloudflare load balancer or Worker routing by game id, or two VMs behind them: a new hosted service and a second VPS. Rejected under AGENTS.md §2.

## Decision

**Adopt D.** A deploy starts the new release beside the old one, cuts new connections over once the new process is ready, hands each table over at its own clean point, and lets the old process exit when it holds nothing or the deadline passes. Running games finish on the **new** version, after one reconnect, without a rewind in the common case. That is a deliberate change from #3039's title: keeping the old binary alive for the life of a Commander game (C) costs hours of double memory on a 4 GB VM, hours-long migration compatibility, two-process Discord reconcilers and versioned client assets, and it keeps fixes away from the games that need them. D gets most of C's benefit for a few minutes of overlap.

Rules that hold throughout:

1. **Today's cold restart is always the fallback.** Any failure (the new process isn't ready in time, a hand-off errors, the control socket is gone, the release is marked "cold") ends in `systemctl restart`, which works exactly as now. A `force` input on the deploy skips straight to it.
2. **At most two processes, for at most the deadline.** A deploy that finds an overlap still running first forces the older one's remaining tables over.
3. **One-way hand-off.** Old → new only. Rollback is a cold restart.
4. **One singleton owner,** by the lock file, and the new owner reconciles only once every table has arrived.
5. **Dev rehearses it.** `cmdctrl-cd-dev` runs the same steps (ADR 0023), so every `develop` deploy exercises the hand-off.

ADR 0044 decision 1's reasoning ("nowhere to drain to") is answered by a second *process*, not a second node. Decision 1 is superseded only in that part once this is accepted.

## Phased plan

**Phase 0: measure (#3037).** Prod boot time, `SeqBehind` per table at real deploys (the shutdown census already logs it), how long tables stay unclean, and RSS idle and per table. This sets the deadline and confirms the memory budget. If #3037 shows rewinds are rare and boots are seconds, stop here and do #3038 or nothing.

**Phase 1: the server pieces, no deploy change** (one PR each, all testable in-process with two `RoomManager`s over one temp data dir):
1. `Room.HandOff`: wait for clean under the lock, refuse actions with a retryable "moving" error, stop runners, return the restore-point seq. Deadline and force variants.
2. Single-table hot restore: `RestoreRoomsAdmitting` and `adoptRestored` for one id, admitted outside the cap.
3. The unix control socket and its verbs; the readiness bit; the old process's load in the new one's cap.
4. `SO_REUSEPORT` on the public listener and a `drain` that closes it; the proxy for not-yet-moved games.
5. The singleton lock around the voice and #lfg managers, the sweepers and the backup sweep; metrics listener retried until the old process frees `127.0.0.1:9464`, with a `release` label on build info.

**Phase 2: the deploy.**
1. A template unit (`cmd-and-ctrl@.service`, instance = release id) with a per-release binary directory, so the old binary isn't overwritten mid-run. The unit comes from HomeLab today; see open question 5.
2. CD: install the release under its own directory; start the new instance; wait for `ready`; `drain` the old one; poll `tables` until zero or the deadline; force the rest; stop the old instance; smoke as today. `keep-previous-release.sh` and `rollback-release.sh` learn the per-release layout.
3. A migration check: a release carrying a migration that renames or drops is marked cold (a marker in the migration file, read by CD), and CI fails a breaking migration without the marker.
4. Runbook in `deploy/README.md` and `docs/environments.md`, including `force`.

**Phase 3: polish,** each optional: a "moved" close code with immediate redial; the lobby list merged across the overlap; an alert on an overlap that outlives its deadline.

**Not planned:** long-lived old versions (C), journal replay (E), shared state (F). Each is a new ADR if wanted.

## Consequences

- A routine deploy costs a table one reconnect and its undo history, at a moment when nothing is mid-prompt. The boot gap disappears from the player's view.
- The deploy gets longer (up to the deadline) and has more moving parts. The cold path stays and is exercised by `force` and by every failure.
- The server gains a private control surface. It is a unix socket owned by `cmdctrl`, so only that user and root reach it.
- Migrations gain a rule: compatible with the previous binary for one overlap, or marked cold.
- The VM runs near its memory limit for minutes per deploy until #3037 says otherwise.

## Back-out

Remove the hand-off steps from CD and restart as today. The server pieces are inert without a `drain` (one process holds the lock and the listener, as now).

## Owner decisions (2026-10-15)

The owner approved the recommendations on 2026-10-15. Each one is accepted as written:

1. **The goal.** Running games move to the new version at a clean point, after one reconnect, instead of #3039's
   literal "finish on the old version". **Accepted.**
2. **The deadline.** 10 minutes, then tune from #3037. **Accepted.**
3. **Memory.** Wait for #3037's RSS numbers before raising prod memory, and expect to raise it. **Accepted.**
4. **Routing.** Share `127.0.0.1:8080` with `SO_REUSEPORT` and keep Caddy unchanged. **Accepted.**
5. **The server's systemd unit.** Move it into this repo and install it from CD. **Accepted.**
6. **Migrations.** Adopt "compatible with the previous binary, or marked cold" with a CI check. **Accepted.**
7. **#3038.** Skip it if #3037 justifies this ADR; its readiness summary is phase 1 item 3. **Accepted.**
8. **Undo across a hand-off.** Accept that it is lost, as on every restart today. **Accepted.**
