# ADR 0123 — Monitoring: metrics, logs, dashboards and alerts

**Status:** Proposed · 2026-10-04 · S63 — Monitoring: metrics, logs, dashboards and alerts.
**Issues:** [#2281](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2281) (this change, and S63's tracker); relates to [#598](https://github.com/krakenhavoc/cmd_and_ctrl/issues/598) items 2 and 3.
**Owner decisions:** the eight answers of 2026-10-04, quoted under [Owner decisions](#owner-decisions-2026-10-04). They are binding. This ADR also makes calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before PR 2 lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04 (`git fetch --all --prune`, then `docs/decisions/` listed on every remote head). The highest number on any branch was 0122, on `origin/develop`. This ADR takes **0123**.
**Builds on:** [ADR 0044](0044-surviving-a-deploy.md) (restore points, the shutdown census), [ADR 0051](0051-user-database.md) (users, games and seats in SQLite), [ADR 0017](0017-bug-report-button.md) §9 (redaction; the server never logs a request URL or query string), [ADR 0033](0033-ai-bot-seat.md) and [ADR 0052](0052-bot-decision-harness-and-eval.md) (the bot funnel's counters), [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) (agent seats).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery), in this repo and in [krakenhavoc/HomeLab](https://github.com/krakenhavoc/HomeLab).

---

## Context

The owner asked on 2026-10-04 for monitoring: "know all the common things and have logs of the metrics and dashboards so I can see how many active games there are, number of players total versus active currently, etc etc".

### What exists (checked on `origin/develop` at `3f0c0222`)

- **Logs.** The server and the Discord bot log JSON through `log/slog` at Info to stdout (`server/cmd/server/main.go:181`, `server/cmd/bot/main.go:25`), and systemd puts that in journald on each host. There is no level switch and no access log, on purpose (ADR 0017 §9). Useful lines already exist: `ws client connected` / `disconnected` with the hub's `total` (`ws/hub.go:450,762`), `table created` (`lobby/http.go:957`), `game restored` and `restore pass complete` with `abandoned` counts (`ws/persist.go:387,430`), and the shutdown census (`ws/shutdown_report.go`).
- **Metrics.** None. No Prometheus, expvar, pprof or OpenTelemetry in the code or `go.mod`. `/healthz` is a static `200 ok` (`main.go:447`). The only in-process counters are per bot seat (`aiseat.Stats`, `aiseat/model/metrics.go`), read by the admin-only `GET /games/{id}/bot/stats`.
- **Counts.** All of the owner's numbers are already cheap to compute:
  - rooms in memory: `ws.RoomManager.List()`;
  - live sockets: `Hub.Count()`, and each socket's `Binding` (`GameID`, `PlayerID`, `ReadOnly` for a spectator, `UserID`, `Admin`) at `ws/hub.go:116`;
  - game state and seat kinds: `lobby.GameMeta.State` (`lobby` / `active` / `ended`), `ArchivedAt`, `Practice`, and the seats' `IsBot`, `BotTier`, `IsAgent`, `UserID` (`lobby/lobby.go:89-213`);
  - history: the `users`, `games` (`state`, `started_at`, `ended_at`, `outcome`) and `seats` tables.
  `users.last_seen_at` is written only at Discord sign-in (`users/store_sql.go:124`), so it measures sign-ins, not play.
- **Hosts.** Two Proxmox VMs on one node, `cmd-and-ctrl` (prod, `cmd.labxp.io`) and `cmd-and-ctrl-dev`, each with the server on `127.0.0.1:8080` behind Caddy and a Cloudflare tunnel (`docs/environments.md`, `deploy/Caddyfile`). The HomeLab README lists "Prometheus, Grafana", but nothing is deployed: `scripts/monitoring/` holds only a README and Terraform defines no monitoring VM.
- **Nothing watches from outside.** #598 item 3: the prod server crash-looped from 6 August to 8 September 2026 and nobody noticed for a month.

### Owner decisions (2026-10-04)

From two rounds of questions the same day:

1. **Stack:** self-hosted in the HomeLab: Prometheus, Grafana and Loki on a new VM, not Grafana Cloud and not an in-app page only.
2. **Logs:** metrics and logs both, shipped and searchable.
3. **Alerts:** Discord alerts plus an uptime check.
4. **Scope:** games and players (active and total, connected now, spectators, bots and agents, games per day, new users), and also HTTP/WS and the Go runtime, host health, bot and model seats, and game engine health.
5. **Off-node watcher:** a GitHub Actions cron, since a monitor on the single Proxmox node cannot see the node go down. (Not a third-party heartbeat.)
6. **Grafana access:** LAN / VPN only. No public hostname, no Cloudflare tunnel.
7. **Retention:** metrics 1 year, logs 30 days.
8. **HomeLab half:** Claude writes the HomeLab PRs (Terraform VM, cloud-init, stack configuration) for the owner to review and apply. Claude never runs `terraform apply`.

---

## Decision

### 1. The shape

```
 cmd-and-ctrl (prod VM)                         cmd-and-ctrl-monitoring (new VM, LAN only)
 ┌──────────────────────────────┐               ┌───────────────────────────────────────────┐
 │ server :8080 (public mux)    │               │ Caddy :9091/:3101  push endpoints, basic  │
 │ server 127.0.0.1:9464 /metrics│◀─scrape─┐    │   auth per host ──▶ Prometheus  (1 y)     │
 │ journald                     │◀─read───┤    │                 ──▶ Loki        (30 d)     │
 │ textfile dir (backup age)    │◀─read───┤    │ Alertmanager ──▶ Discord webhook          │
 │ Grafana Alloy ───────────────┼─push────┼───▶│ blackbox_exporter ──▶ https://cmd…/healthz │
 └──────────────────────────────┘         │    │ Grafana :3000 (LAN)  dashboards from files │
 cmd-and-ctrl-dev: the same, env=dev ──────┘    │ config-sync timer ◀── this repo, main      │
                                                └───────────────────────────────────────────┘
 GitHub Actions cron (every 10 min) ──▶ both /healthz from the internet ──▶ Discord + an issue
```

- Each app VM runs **Grafana Alloy**, which scrapes the server's metrics listener on loopback, collects host metrics (its built-in node exporter), reads journald, and **pushes** everything to the monitoring VM. The app VMs open no new inbound port.
- The monitoring VM runs **Prometheus** (with its remote-write receiver), **Loki**, **Alertmanager**, **Grafana**, **blackbox_exporter**, and **Caddy** in front of the two push endpoints.
- What the app measures (metric names, dashboards, alert rules, the Alloy config) lives **in this repo**, next to the code it describes. The VM and its stack live **in HomeLab**.
- The **GitHub Actions cron** is the one watcher that does not share the node's fate.

### 2. The server: a metrics listener that is never public

- **`CMDCTRL_METRICS_ADDR`** (new). Unset means off, the default for dev loops and tests. CD sets it to `127.0.0.1:9464` on both hosts. The server **refuses to boot** on a non-loopback address: the only reader is Alloy on the same host, and a metrics page reachable from outside leaks the table counts to anyone.
- `/metrics` is served on **its own `http.Server`**, never on the public mux. A change to Caddy's `@api` matcher can therefore never expose it, and the public mux gains no route.
- The library is **`github.com/prometheus/client_golang`**, pinned to its current release when PR 2 is written, with `go mod why` for each new module in the PR description. Everything registers on **our own registry** (`metrics.Registry`), never the global default, so a test can build its own.
- The new package is **`server/internal/metrics`**. It imports only the Prometheus client. Event counters are package-level variables in it, and `ws`, `lobby`, `aiseat` and `db` import it. That is allowed under the `aiseat` import gate, which bans only `internal/game`. State gauges are **collectors** that `main.go` registers with the live `RoomManager`, `Hub`, `Lobby` and database. A collector reads at scrape time, so a gauge cannot drift from the truth.
- **`cmdctrl_build_info{commit}`** comes from `debug.ReadBuildInfo`'s `vcs.revision`, the way `botarena/report.go:284` reads it. The `env` and `host` labels are added by Alloy, not by the server.

### 3. What is measured

The names below are the contract. PRs 3 and 4 may add more but not rename these. Gauges marked *(collector)* are read at scrape time. Everything else is an event counter or histogram.

**Labels are closed sets, always.** No metric carries a user, player, seat or game ID, a name, a remote address, a URL or a free-text error. `TestMetricLabelsAreClosedSets` gathers the whole registry and fails on:
- a label name outside an allowlist;
- a value outside its declared set (route patterns come from the mux, action types from the `actions` enum);
- a metric or label name containing `id`, `user`, `name`, `ip`, `remote` or `url`.

This is how the privacy rules in ADR 0017 §9 and AGENTS.md §5 carry over to metrics, and it also bounds the number of series.

#### Games and players (the owner's ask)

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `cmdctrl_games` | gauge *(collector)* | `state` = `lobby\|active\|ended`, `archived` = `true\|false` | Tables in the lobby registry, practice tables excluded. |
| `cmdctrl_practice_games` | gauge *(collector)* | — | Open practice (tutorial) tables. |
| `cmdctrl_games_created_total` | counter | — | `POST /games` successes. |
| `cmdctrl_games_started_total` | counter | `seats` = `1`…`8` | Tables that left the lobby. |
| `cmdctrl_games_ended_total` | counter | `outcome` = `win\|draw\|closed` | `closed` is an admin close or abandonment (`outcome` NULL in migration 0007's terms). |
| `cmdctrl_game_duration_seconds` | histogram | — | Start to end, observed at end. |
| `cmdctrl_seats` | gauge *(collector)* | `kind` = `human\|bot\|agent` | Seats at active (non-practice) tables. |
| `cmdctrl_seats_connected` | gauge *(collector)* | `kind` = `human\|agent`, `account` = `signed_in\|guest` | Seats at active tables with at least one live socket now: **"players active currently"**. |
| `cmdctrl_spectators_connected` | gauge *(collector)* | — | Live read-only sockets. |
| `cmdctrl_users` | gauge *(collector)* | — | Registered accounts: **"players total"**. Read from SQLite and cached 60 s. |
| `cmdctrl_users_played` | gauge *(collector)* | `window` = `1d\|7d\|30d` | Distinct accounts seated at a table that was active in the window (seats joined to games on `started_at` / `ended_at`). It doesn't use `last_seen_at`, which counts only sign-ins. Cached 60 s. |
| `cmdctrl_users_created_total` | counter | — | First Discord sign-ins (new `users` rows). |

"Games per day" and "new users per day" are `increase(…[1d])` over the counters. History older than the metrics' year stays answerable from SQLite.

#### HTTP, WebSocket and runtime

| Metric | Type | Labels |
|---|---|---|
| `cmdctrl_http_requests_total` | counter | `route`, `method`, `code` = `2xx\|3xx\|4xx\|5xx` |
| `cmdctrl_http_request_seconds` | histogram | `route` (excluding `GET /ws`, which is long-lived) |
| `cmdctrl_ws_connections` | gauge *(collector)* | `role` = `seat\|spectator\|admin` |
| `cmdctrl_ws_connects_total` / `cmdctrl_ws_disconnects_total` | counter | `role` |
| `cmdctrl_ws_upgrade_rejections_total` | counter | `reason` (the hub's existing rejection cases, as a closed set) |
| `cmdctrl_ws_frames_total` | counter | `direction` = `in\|out`, `type` (protocol frame types) |
| `cmdctrl_ws_broadcast_seconds` | histogram | — |
| Go runtime and process | `collectors.NewGoCollector`, `NewProcessCollector` | — |

**`route` is the matched `ServeMux` pattern**, such as `POST /games/{id}/join`, never the path. The lobby is a second mux mounted at `/` (`main.go:585`), so the outer middleware would see `/` for every lobby route. The middleware therefore puts a small holder in the request context, and whichever mux matches last writes its pattern into it. A request no pattern matched is `route="unmatched"`. The route set is closed because the patterns are compile-time strings.

#### Game engine health

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `cmdctrl_actions_total` | counter | `type` (the `actions` enum), `seat_kind` = `human\|bot\|agent\|admin`, `result` = `applied\|rejected` | Every `Room.Apply` / `ApplyBundle` / `ApplyExternal`. |
| `cmdctrl_action_apply_seconds` | histogram | `result` | Time under the room lock. |
| `cmdctrl_effect_errors_total` | counter | — | `EventEffectError`s emitted: a card's effect threw mid-resolution. The engine survives it, so nothing else counts it. |
| `cmdctrl_restore_point_age_seconds` | gauge *(collector)* | — | The oldest restore point among active rooms: how much a deploy now would rewind (ADR 0044). |
| `cmdctrl_rooms_behind_restore_point` | gauge *(collector)* | — | Active rooms whose last applied action is not yet captured. |
| `cmdctrl_boot_restore_games` | gauge | `outcome` = `restored\|ended\|abandoned` | The last boot's restore pass, set once at boot. |
| `cmdctrl_boot_restore_degraded_cards` | gauge | — | Cards flagged `AbilitiesLostOnRestore` at the last boot. |
| `cmdctrl_db_size_bytes` | gauge *(collector)* | — | The live SQLite file. |
| `cmdctrl_db_backup_last_success_timestamp_seconds` | gauge | — | The in-process `VACUUM INTO` sweep. |

The shutdown census stays a log line. A process that is exiting is not scraped again, and Loki keeps the line.

#### Bot and model seats

These are totalled across every runner, from the counters `aiseat.Stats` and `aiseat/model/metrics.go` already keep. The per-game admin route is unchanged.

| Metric | Type | Labels |
|---|---|---|
| `cmdctrl_bot_decisions_total` | counter | `tier`, `layer` = `A\|B\|C` |
| `cmdctrl_bot_decision_seconds` | histogram | `tier` |
| `cmdctrl_bot_fallbacks_total` | counter | `tier`, `cause` (the decision log's closed fallback causes) |
| `cmdctrl_bot_model_calls_total` | counter | `tier`, `result` = `ok\|timeout\|error\|malformed\|out_of_range` |
| `cmdctrl_bot_model_call_seconds` | histogram | `tier` |
| `cmdctrl_bot_model_tokens_total` | counter | `tier`, `direction` = `prompt\|completion` |

Agent seats need nothing extra. They play over the WebSocket like anyone, so `cmdctrl_seats{kind="agent"}`, `cmdctrl_seats_connected{kind="agent"}` and `cmdctrl_actions_total{seat_kind="agent"}` cover them.

### 4. Host health and the backup's age

Alloy's built-in node exporter collects CPU, memory, load, filesystems (`/`, the data disk, the Docker partition where there is one) and network. Its systemd collector is filtered to `cmd-and-ctrl*`, `caddy`, `cloudflared` and `alloy`.

`scripts/backup-offsite.sh` gains two lines at the end. It writes `cmdctrl_offsite_backup_last_success_timestamp_seconds` and `cmdctrl_offsite_backup_last_exit_code` atomically, by write-then-rename, to `/var/lib/cmdctrl-metrics/backup.prom`. Alloy's textfile collector reads that directory. The directory is `root:cmdctrl 0775`, so the backup unit's `cmdctrl` user can write it with no new privilege.

### 5. Logs

- Alloy reads journald for `cmd-and-ctrl`, `cmd-and-ctrl-bot`, `cmd-and-ctrl-backup`, `caddy` and `cloudflared`, and pushes to Loki.
- **Loki labels are `env`, `host`, `unit` and `level` only.** `level` is parsed from the JSON line. `game_id`, `user_id`, `player` and `remote` stay in the line, searchable with `| json`, and never become labels. That bounds Loki's index the same way §3 bounds Prometheus's.
- **Retention is 30 days** (decision 7), enforced by Loki's compactor. The logs carry internal UUIDs and remote addresses already (`ws/hub.go:370`), so the 30 days is also the privacy bound. Grafana is LAN-only (decision 6), and nothing in Loki leaves the house.
- Nothing about what the server logs changes. ADR 0017 §9 (no URLs, no query strings), the redaction helpers, and "the log carries the count, never the IDs" all still hold, and Loki inherits them.

### 6. The monitoring VM (HomeLab)

- **Terraform:** `proxmox_virtual_environment_vm.cmd_and_ctrl_monitoring` in `terraform/deployments/lab/`, shaped like the existing `cmd_and_ctrl` VMs: 2 vCPU, 4 GiB, a 20 GB OS disk and an **80 GB data disk** for the stores. It gets a **fixed LAN address** set in Terraform, not DHCP, because two Alloy configs point at it and #598 item 1 is what a rotted DHCP lease costs.
- **Cloud-init** installs Docker and writes one Compose project under `/opt/monitoring`, with every image pinned by tag and digest:
  - `prometheus`, with `--web.enable-remote-write-receiver`, `--storage.tsdb.retention.time=1y` and a 60 GB `retention.size` backstop;
  - `alertmanager`;
  - `loki`, with a 720 h retention and the compactor on;
  - `grafana`;
  - `blackbox_exporter`;
  - `caddy`.

  All data lives on the data disk. Alloy also runs on this VM, with `env="monitoring"`, to report its own host metrics (disk above all) and the textfile the config-sync timer writes.
- **Push authentication:** Caddy terminates the two push endpoints (Prometheus remote write and Loki push), with **one basic-auth credential per app host**. The hashes come from sensitive tfvars. Prometheus and Loki listen only on the Compose network. The VM's firewall admits the LAN subnet on the push ports and Grafana's port, and nothing else.
- **Grafana** is on the LAN address, port 3000. Its admin password is a sensitive tfvar, and anonymous access is off. There is no tunnel and no public DNS (decision 6).
- **Alertmanager** sends to a Discord webhook, using Alertmanager's native `discord_configs`. The webhook URL is a sensitive tfvar and never appears in this repo.
- **Config from this repo:** a systemd timer (every 10 minutes) fetches `deploy/monitoring/` from `krakenhavoc/cmd_and_ctrl` at `main`. The repo is public, so no credential is needed. The timer then:
  1. runs `promtool check rules` and a JSON parse of every dashboard;
  2. if both pass, swaps the directory and reloads Prometheus (`/-/reload`) and Grafana's file provisioning;
  3. if either fails, keeps the old directory and writes `cmdctrl_monitoring_sync_last_success_timestamp_seconds` for the alert in §7.

  Only promoted, reviewed rules reach the monitor, and the monitor holds no trust in either app host or the runner.
- **Not backed up.** Grafana's state is all provisioned from files, so rebuilding the VM loses only history (metrics and logs). That is the accepted cost; see Consequences.

### 7. Alerts

The rules live in `deploy/monitoring/rules/*.yml`. Alertmanager groups by `alertname` and `env`, repeats every 4 h, and sends resolutions. Every message names the env. Dev alerts are `severity: warning`, and dev deploys restart the server often, so the dev thresholds are longer.

| Alert | Fires when | Severity |
|---|---|---|
| `SiteDown` | blackbox probe of `https://cmd.labxp.io/healthz` failing for 3 m (dev: 10 m). It goes out through Cloudflare, so it tests the tunnel too. | critical |
| `ServerNotReporting` | no samples from a server's metrics for 5 m (`absent_over_time`) | critical |
| `ServerCrashLoop` | `changes(process_start_time_seconds[15m]) > 3`. A deploy restarts it once. | critical |
| `UnitFailed` | a `cmd-and-ctrl*` unit in systemd state `failed`. This becomes possible once #598 item 2 bounds the restarts. | critical |
| `RestoreAbandonedGames` | `cmdctrl_boot_restore_games{outcome="abandoned"} > 0` within 15 m of a start | warning |
| `RestorePointLagging` | `cmdctrl_restore_point_age_seconds > 600` for 10 m | warning |
| `OffsiteBackupStale` | the backup's last success is older than 36 h, or its last exit was non-zero | warning |
| `DiskFilling` | any filesystem under 15% free for 15 m | warning |
| `MonitoringConfigSyncFailing` | the config-sync timer's last success is older than 1 h | warning |

`cmdctrl_effect_errors_total` and the bot model's timeout rate are dashboard panels, not alerts. Each is a reason to look, not to be paged.

### 8. The off-node watcher (GitHub Actions)

- `.github/workflows/uptime.yml` runs every 10 minutes on GitHub-hosted runners (the repo is public, so they are free). It curls both `/healthz` URLs from the internet, three tries 20 s apart. One success is up.
- **The state lives in an issue, not in Actions.** On the first down run for a site, the workflow opens an issue titled "`<env>` is down" with the label `outage`, and posts once to Discord. On the first up run, it comments the downtime, closes the issue and posts the recovery. A site that stays down produces no repeat messages.
- **The webhook is the Actions secret `CMDCTRL_ALERT_DISCORD_WEBHOOK`**. It can be the same channel Alertmanager uses.
- **Limits:**
  - GitHub runs scheduled workflows only from the default branch (`main`), so the watcher starts with the promotion that carries it.
  - GitHub may run a scheduled job several minutes late, or skip it under load. That is acceptable for the outage this exists for, which lasted a month.
  - It watches the sites. It does not watch the monitoring VM, which is LAN-only. That gap is under Consequences.

This closes #598 item 3. Item 2, bounded restarts, ships here too: this repo's half for the bot unit, and HomeLab's half for the server unit (PR H2). Items 1 and 4 stay on #598.

### 9. Dashboards

Dashboards are JSON in `deploy/monitoring/dashboards/`, provisioned read-only (`allowUiUpdates: false`). To change one, edit it in Grafana's scratch folder, export the JSON and open a PR. Each has an `env` variable defaulting to `prod`.

1. **Overview**, the owner's question at a glance:
   - games by state;
   - players connected now (human, agent, signed-in and guest), spectators and bot seats;
   - registered accounts against those who played in 1d, 7d and 30d;
   - games started and ended per day, game length and new users per day;
   - the build commit and uptime.
2. **Server:** HTTP rate, latency and errors by route, WebSocket connections and churn, frame rates, and Go runtime and process memory.
3. **Engine and bots:** actions a minute by type and seat kind, rejections, apply latency, effect errors, restore-point age, bot decisions by layer, model calls, timeouts and tokens.
4. **Hosts:** CPU, memory, disk, systemd unit state, the off-site backup's age, and the config sync's age.
5. **Logs:** Loki panels for errors and warnings by unit, WebSocket disconnect bursts, and restore and shutdown-census lines.

### 10. Deploying the agent (this repo's CD)

- A new CD step, **"Ensure monitoring agent"**, mirrors "Ensure off-site backup":
  1. installs Alloy, pinned, from Grafana's apt repository;
  2. writes `/etc/alloy/config.alloy` from `deploy/alloy/config.alloy`;
  3. writes `/etc/alloy/env` (`root:alloy 0640`) with the monitoring URL, this host's push credential and the `env` label;
  4. restarts Alloy only when either file changed.
- **Inputs:**
  - the Actions variable `CMDCTRL_MONITORING_URL` (the monitoring VM's fixed address);
  - the secret `CMDCTRL_MONITORING_PUSH_PASSWORD`, an environment-level value in `prod` and `dev`.
- **Missing values:** a missing value, or one still holding `REPLACE_ME`, is a `::warning::` and a stopped Alloy, **never a failed deploy**. The game must not depend on its monitor.
- The same step sets `CMDCTRL_METRICS_ADDR=127.0.0.1:9464` through `scripts/set-server-env.sh`, and creates `/var/lib/cmdctrl-metrics`.
- **Alloy's own UI** listens on `127.0.0.1:12345` only.

### 11. Tests and CI

- **Go:**
  - `TestMetricLabelsAreClosedSets` (§3);
  - the listener refuses a non-loopback address;
  - `/metrics` is absent from the public mux;
  - each collector against a built table: two seats, one connected guest, one spectator, one bot, and the expected `cmdctrl_seats_connected` and `cmdctrl_seats` series;
  - the route holder labels a lobby route with its own pattern, not `/`.
- **A CI job `monitoring-config`**, on PRs touching `deploy/monitoring/` or `deploy/alloy/`:
  - `promtool check rules` and `promtool test rules`, with one unit test per alert in `deploy/monitoring/rules/tests/`;
  - `alloy fmt` on the Alloy config;
  - a dashboard check: JSON parses, `uid`s are unique, and every PromQL `cmdctrl_*` name a panel uses is one the server registers.

  It runs the pinned upstream images on a GitHub-hosted runner. This job is the only Docker use, and it adds no Go cache volume.
- **The uptime workflow** gets a `workflow_dispatch` dry-run input that checks the URLs and prints what it would post, without posting.

---

## Delivery

Each cmd_and_ctrl PR goes into `develop` under Sprint S63 and Issue #2281. The HomeLab PRs follow that repo's own conventions and link #2281.

| PR | Repo | What | Needs | Parallel with |
|---|---|---|---|---|
| 1 | this | **This ADR**, the AGENTS.md §3 ADR range line, and the S63 section and index row in `docs/sprints.md`. Docs only. | — | anything |
| 2 | this | **Server foundation** (§2): `internal/metrics`, `client_golang`, the `CMDCTRL_METRICS_ADDR` listener and its loopback check, build info, the Go and process collectors, the HTTP middleware with the route holder, the label guard test, and AGENTS.md §3 and §5. | 1 | H1, 7 |
| 3 | this | **Games, players and WebSocket** (§3): the lobby, hub and database collectors, the game, user and WebSocket counters, and the database gauges. | 2 | 4, H1 |
| 4 | this | **Engine and bots** (§3): the action counters on all three apply paths, effect errors, the restore-point gauges, the boot restore gauges, and the bot and model totals. | 2 | 3, H1 |
| 5 | this | **The agent** (§4, §10): `deploy/alloy/config.alloy`, "Ensure monitoring agent", the backup textfile lines, `StartLimit*` on `deploy/cmd-and-ctrl-bot.service` (#598 item 2, this repo's half), and the monitoring section of `docs/environments.md`. | 2 | 3, 4, 6 |
| 6 | this | **Rules and dashboards** (§7, §9, §11): `deploy/monitoring/` with rules, rule tests, dashboards and the provisioning files, the `monitoring-config` CI job, and a new `docs/monitoring.md` (what each dashboard answers and how to change one). | 3, 4 | 5 |
| 7 | this | **The uptime watcher** (§8): `.github/workflows/uptime.yml`, the `outage` label, and the dry-run input. | 1 | anything |
| H1 | HomeLab | **The monitoring VM** (§6): Terraform VM, fixed address, firewall, cloud-init, the Compose project, Caddy push auth, the config-sync timer, and the sensitive tfvars declared with no values. | 1 | 2–7 |
| H2 | HomeLab | **`StartLimitIntervalSec` / `StartLimitBurst` on `cmd-and-ctrl.service`** in `templates/setup-cmd_and_ctrl.yaml.tftpl` (#598 item 2, the server's half). | — | anything |

### The owner's steps

These are the parts no agent can or should do:

1. **Discord:** create a webhook in the channel the alerts should go to.
2. **HomeLab:**
   - pick the monitoring VM's fixed address;
   - set the sensitive tfvars: Grafana admin password, the Discord webhook, and two push-credential hashes;
   - run `terraform apply` for H1, then H2.
3. **GitHub, cmd_and_ctrl:**
   - the variable `CMDCTRL_MONITORING_URL`;
   - the secret `CMDCTRL_MONITORING_PUSH_PASSWORD` in the `prod` and `dev` environments;
   - the repo secret `CMDCTRL_ALERT_DISCORD_WEBHOOK`.
4. **Promote to `main`.** The dashboards, the rules and the uptime cron all start from `main`.

The exit check: on cmd-dev, the Overview dashboard shows a table the owner opens, with its seat connected and a bot seat. Stopping the dev server's unit fires `ServerNotReporting` in Discord within about 5 minutes, and the alert resolves after a restart. One `uptime.yml` dry run prints both URLs as up. The evidence goes on #2281.

---

## Consequences

- **The owner's numbers become a page.** Active games, players connected now against registered and recent players, games a day, and new users are on the Overview dashboard, kept for a year, for both environments.
- **A dependency, a VM and an agent are added.** `client_golang` brings its transitive modules (`client_model`, `common`, `procfs`, `protobuf`). The monitoring VM is one more box to patch, and Alloy is one more unit on each app host. The game does not depend on any of them: no metrics address, a dead monitor, or missing CD values all leave the server serving.
- **The monitor is not watched from outside.** The cron watches the sites, not Grafana or Prometheus. If the monitoring VM dies on its own, alerts stop silently until the owner opens a dashboard. If the whole node dies, the cron still reports the sites down. This is the trade decision 5 made against a third-party heartbeat. `MonitoringConfigSyncFailing` covers only the sync.
- **The monitoring VM shares the node.** A node outage loses its alerts and history for the duration. The cron is the only thing that sees it.
- **History is not backed up.** Rebuilding the monitoring VM starts the year of metrics again. SQLite, which is backed up, still holds every game and user for long-range questions.
- **Thirty days of logs with internal UUIDs and remote addresses** now live in a second place, on the LAN only. That is the same data journald already holds, kept for a bounded time.
- **A small per-action cost.** One counter increment and one histogram observation per applied action, and a scrape every 30 s that takes each room's lock briefly. The room lock is already held for far longer per action.
- **#598 narrows** to items 1 (DHCP-pinned hosts) and 4 (the stateless runner).

## Out of scope

- **Tracing** (OpenTelemetry spans). Metrics and logs answer the questions asked. Tracing can be added later through Alloy without changing this design.
- **Client-side metrics** (browser errors, frame times, Web Vitals). The bug-report button covers client failures today.
- **pprof on the metrics listener.** It is easy to add later, but it is a debugging surface, not monitoring.
- **A public status page**, and any public route for counts. Decision 6 keeps all of it on the LAN.
- **Discord bot metrics** beyond its unit's state and its logs. The bot has no `/metrics`, since its traffic is a few commands a day.
- **Per-game or per-player dashboards.** The label rule (§3) forbids them in Prometheus. Loki can show one game's lines (`| json | game_id="…"`), and that is enough.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review.

1. **Push from Alloy, not pull from Prometheus** (§1). App hosts open no port, and their DHCP addresses never appear in the monitor's config.
2. **A separate loopback-only listener, with boot refused on any other address** (§2). The alternative was a `/metrics` route on the public mux behind admin auth.
3. **`client_golang` on our own registry** (§2), rather than a lighter library such as VictoriaMetrics' `metrics`. It is the standard, and it brings the Go and process collectors.
4. **The metric names and label sets in §3, with label sets closed and enforced by a test.** No IDs, even though some debugging would want them. Loki holds the IDs.
5. **"Played in the window" from seats and games, not `last_seen_at`** (§3), because `last_seen_at` counts sign-ins.
6. **A 30 s scrape interval.** About 8,000 active series across both environments and three hosts, at Prometheus's ~1.3 bytes a sample, is about 11 GB a year. That fits the 80 GB disk with room for Loki. A 15 s interval would double it.
7. **One Compose project on the monitoring VM, images pinned by digest** (§6), rather than distribution packages, which lag upstream, or one VM per component.
8. **Config pulled by the monitor from `main`** (§6), not pushed by CD. It needs no new SSH trust, and only promoted changes reach it. The cost: a dashboard change on `develop` is not live until promotion.
9. **Caddy basic auth per app host on the push endpoints, plus a firewall** (§6). The LAN is not the only boundary.
10. **The alert set and thresholds in §7**, including keeping effect errors and model timeouts off the pager. **One webhook for both environments**, with `env` in every message.
11. **Issue-as-state for the uptime cron** (§8), so a long outage posts twice: once down, once recovered.
12. **The monitoring VM is not backed up** (§6, Consequences).
13. **Bounded restarts on both units in this sprint** (#598 item 2), because `UnitFailed` needs a unit that can reach `failed`.
