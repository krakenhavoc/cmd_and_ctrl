# Monitoring: dashboards and alerts

The game server and both app hosts report to a monitoring VM in the
HomeLab: Prometheus for metrics (kept a year), Loki for logs (30 days),
Grafana for the dashboards and Alertmanager for the Discord alerts. The
design is [ADR 0123](decisions/0123-monitoring-metrics-logs-dashboards-and-alerts.md).
How each app host ships its data is in
[environments.md, Monitoring agent](environments.md#monitoring-agent).
This page is about what you look at and what wakes you up.

Grafana is on the monitoring VM's LAN address, port 3000. There is no
public hostname (ADR 0123 decision 6). The dashboards are in the
`cmd_and_ctrl` folder.

| What | Where in this repo | Read by |
|---|---|---|
| Alert rules | [`deploy/monitoring/rules/*.yml`](../deploy/monitoring/rules/) | Prometheus on the monitoring VM |
| Rule unit tests | [`deploy/monitoring/rules/tests/`](../deploy/monitoring/rules/tests/) | CI only, never synced |
| Dashboards | [`deploy/monitoring/dashboards/*.json`](../deploy/monitoring/dashboards/) | Grafana on the monitoring VM, read-only |
| Alloy config | [`deploy/alloy/config.alloy`](../deploy/alloy/config.alloy) | Alloy on each app host, installed by CD |

The monitoring VM fetches `deploy/monitoring/` from **`main`** every 10
minutes. It checks the rules with `promtool check rules` and parses every
dashboard. If both pass, it swaps the files in and reloads Prometheus and
Grafana. If either fails, it keeps the last good copy and
`MonitoringConfigSyncFailing` fires within the hour. So a change on
`develop` is not live until it is promoted.

## Dashboards

Every dashboard has an `env` picker at the top: `prod` (the default) or
`dev`. The ⋯ `cmd_and_ctrl` link in the top bar opens the others with the
same env and time range. Every panel has a description: hover its title.

| Dashboard | Answers |
|---|---|
| **Overview** (`cmdctrl-overview`) | How many games are running and waiting now. How many players are connected, as humans or agents, signed in or as guests, plus spectators, bot seats and practice tables. How many accounts exist and how many played in the last day, 7 days and 30 days. Games created, started and ended per day, how they ended, how long they last, and new accounts per day. Which commit is running, how long since it started, and whether the site and the server are up. |
| **Server** (`cmdctrl-server`) | HTTP requests per route and status class, 5xx by route, latency (overall and the slowest routes). WebSocket connections by role, connects and disconnects, refused upgrades by reason, frames by type, and broadcast time. The database's size and its last in-process backup. The process's memory, CPU, goroutines and open files. |
| **Engine and bots** (`cmdctrl-engine`) | Actions per minute by type and seat kind (human, bot, agent, admin), rejections, and how long a commit holds the room lock. Effect errors. Restore-point age and rooms behind theirs. What the last boot restored, found ended, abandoned or degraded. Bot decisions by layer and tier, decision time, and fallbacks by cause. Model calls by result, the timeout rate, call time and tokens. |
| **Hosts** (`cmdctrl-hosts`) | Whether the site's `/healthz` answers from outside. The off-site backup's age and last exit code. The config sync's age and last result. CPU, memory, load, disk free, disk I/O and network per host. The state of every watched systemd unit, and how often systemd restarted them. Pick `monitoring` in the env picker for the monitoring VM itself. |
| **Logs** (`cmdctrl-logs`) | Error and warning lines per minute by unit, and the lines themselves. WebSocket connects and disconnects per minute (a burst of disconnects with no deploy is a problem). What systemd said about our units. The shutdown census and the boot's restore pass. At the bottom, every line from one unit, filtered by any text, such as a `game_id`. |

Metrics carry no IDs, by design (ADR 0123 §3). To follow one game or one
person, use the Logs dashboard's stream with the ID as the search text,
or Grafana's Explore with `{env="prod", unit="cmd-and-ctrl.service"} | json | game_id="…"`.

Effect errors and the model timeout rate are panels, not alerts: each is
a reason to look, not to be paged.

## Alerts

All alerts go to one Discord channel, grouped by alert name and env. They
repeat every 4 hours while firing, and a message is sent when they
resolve. Every message names the env and, where there is one, the host.
On `dev` nothing is critical, and alerts a deploy could trip wait longer.

| Alert | Fires when | Severity | What to do |
|---|---|---|---|
| `SiteDown` | The monitoring VM's probe of `/healthz`, through Cloudflare, fails for 3 minutes (dev: 10). | prod critical, dev warning | If `ServerNotReporting` fired too, the server is down: see that row. If not, the server is fine and the path to it is broken: `systemctl status caddy cloudflared` on the host, then the Logs dashboard for those units. |
| `ServerNotReporting` | No metrics from the game server for 5 minutes: it is down while Alloy scrapes it, or nothing arrives from the host at all. | prod critical, dev warning | `systemctl status cmd-and-ctrl` and `journalctl -u cmd-and-ctrl -n 100` on the host. If the server is up, Alloy or the push path is down: see [Checking on it](environments.md#monitoring-agent). If the host does not answer, check the VM in Proxmox. |
| `ServerCrashLoop` | The server started 4 or more times in 15 minutes. A deploy starts it once. | prod critical, dev warning | The Logs dashboard's "What systemd says about the units" and `journalctl -u cmd-and-ctrl` show the panic or the fatal line. Fix forward, or promote a revert. |
| `UnitFailed` | `cmd-and-ctrl.service` or `cmd-and-ctrl-bot.service` is in `failed`: systemd hit the start limit (10 starts in 600 s) and gave up. The backup unit is left to `OffsiteBackupStale`. | prod critical (after 1 m), dev warning (after 5 m) | Read the unit's journal, fix the cause, then `sudo systemctl reset-failed <unit> && sudo systemctl restart <unit>` (see [Restarts and the start limit](environments.md#operating-notes)). |
| `RestoreAbandonedGames` | The last boot could not restore one or more tables from their restore points. It looks only in the 15 minutes after a start, then resolves by itself. | warning | The Logs dashboard's "Restores and the shutdown census" panel has one ERROR line per table with the reason. A file written by a newer build is kept for a roll-forward. The players of those tables lost their game. |
| `RestorePointLagging` | The oldest restore point among live tables is more than 10 minutes old, for 10 minutes. A deploy or crash now would rewind that table that far. | warning | Look for the table in the shutdown census lines of the next deploy, or hold the deploy. The cause is usually something on the stack the engine cannot capture yet ([ADR 0044](decisions/0044-surviving-a-deploy.md), [ADR 0041](decisions/0041-game-persistence.md)). |
| `OffsiteBackupStale` | The nightly off-site backup last succeeded more than 36 hours ago (`cause=stale`), or its last run exited non-zero (`cause=failed`). | warning | `journalctl -u cmd-and-ctrl-backup` and `systemctl list-timers cmd-and-ctrl-backup.timer` on the host. A deploy with missing R2 values disables the timer, which shows here as stale. Runbook: [Backups](environments.md#backups). |
| `DiskFilling` | A real filesystem (not tmpfs or overlay) on any host, the monitoring VM included, has under 15% free for 15 minutes. | warning | `df -h` and `sudo du -xh --max-depth=2 <mountpoint> \| sort -h \| tail` on the host. On an app host, look at the data directory (`replays/`, `bugreports/`, `restore/`, the Scryfall and image caches). On the monitoring VM, Prometheus's `retention.size` and Loki's retention are the levers (HomeLab). |
| `MonitoringConfigSyncFailing` | The monitoring VM has not applied `deploy/monitoring/` from `main` for more than an hour. It keeps serving the last good copy. | warning | Usually something on `main` that `promtool check rules` refuses or a dashboard that does not parse: run the checks below on `main`. Otherwise read the config-sync timer's journal on the monitoring VM. |

`SiteDown` and `ServerNotReporting` both depend on the monitoring VM.
If the whole Proxmox node is down, neither can fire, and the
[uptime watcher](environments.md#uptime-watcher), a GitHub Actions cron,
is what reports it. It does not curl the sites (Cloudflare challenges
GitHub runners). It watches a heartbeat the monitoring VM writes every 5
minutes, and opens a `` `monitoring` heartbeat lost `` issue and posts to
Discord when it is more than 20 minutes old.

## Changing a dashboard

The dashboards are provisioned read-only, so an edit in Grafana cannot be
saved over them. The JSON in this repo is the source.

1. Open the dashboard in Grafana, then **Edit → Save dashboard → Save as
   copy** into the **Scratch** folder. Change the copy until it is right.
2. Export it: **Export → Export as JSON**, with **Export the dashboard to
   use in another instance** (the external-sharing switch) **off**. With it
   on, datasources become `${DS_…}` inputs, which file provisioning
   cannot resolve and the CI check refuses.
3. Replace the file in `deploy/monitoring/dashboards/` with the export, then
   put back the original `uid` and `title` (the copy got new ones) and
   delete the top-level `id`. Keep the uid stable: links and bookmarks
   use it.
4. Run the checks below, open a PR into `develop`, and merge it.
5. It goes live when `develop` is promoted to `main`, within 10 minutes of
   the promotion. Delete the Scratch copy then.

A new dashboard follows the same steps, with a new uid starting
`cmdctrl-`, the `cmd_and_ctrl` tag, and an `env` variable. Name
datasources by uid only: `prometheus`, `loki` or `alertmanager`.

## Adding a metric

A metric starts in the server: read the package doc in
[`server/internal/metrics/doc.go`](../server/internal/metrics/doc.go)
(event counters vs. state collectors, and the closed label sets that
`TestMetricLabelsAreClosedSets` enforces). Once it is registered, a
dashboard or rule can use it. The name check below fails if a panel or a
rule names a `cmdctrl_` metric the server does not register, so a rename
in the server must update the dashboards in the same PR.

Metrics that other jobs write to the textfile directory
(`cmdctrl_offsite_backup_*`, `cmdctrl_monitoring_sync_*`) are not the
server's. They are listed by name in `textfileMetrics` in
[`server/cmd/server/monitoring_config_test.go`](../server/cmd/server/monitoring_config_test.go);
add a new one there.

## Adding an alert

1. Add the rule to a file in `deploy/monitoring/rules/`. Give it
   `severity: critical` or `warning`, and a `summary` and `description`
   that name `{{ $labels.env }}` and, where the series has one,
   `{{ $labels.host }}`. Two rules with the same name and the same labels
   fail `promtool check rules` (its duplicate-rule lint), so a prod/dev
   pair must differ in a label, such as `severity`.
2. Add a test group for it in `deploy/monitoring/rules/tests/`: at least
   one case where it fires, with its exact labels and annotations, and one
   where it does not.
3. Add a row to the alert table above with what to do when it fires.

## Checks

CI's `monitoring-config` job runs on any PR that touches
`deploy/monitoring/` or `deploy/alloy/`, with the pinned upstream images:

- `promtool check rules` and `promtool test rules` (`prom/prometheus:v3.15.0`, the monitoring VM's version);
- every dashboard parses as JSON;
- `alloy fmt --test` and `alloy validate` on the Alloy config (`grafana/alloy:v1.20.1`, the hosts' pin).

The server's tests (`server-test`) also run on those PRs, and two of them
read `deploy/monitoring/`:

- `TestMonitoringConfigNamesRegisteredMetrics`: every `cmdctrl_` name in a
  rule file or a dashboard is one the server registers, or a textfile
  metric. The server's names come from `Describe` on
  `metrics.Collectors()` and on every state collector `registerGameMetrics`
  registers, so a counter nothing has incremented still counts.
- `TestMonitoringDashboardsAreProvisionable`: every dashboard parses, has
  a unique `cmdctrl-` uid and title and an `env` variable, and names only
  the three datasource uids.

To run them before pushing, from the repo root:

```sh
docker run --rm -v "$(pwd):/repo:ro" -w /repo --entrypoint promtool prom/prometheus:v3.15.0 \
  check rules deploy/monitoring/rules/cmdctrl.yml
docker run --rm -v "$(pwd):/repo:ro" -w /repo --entrypoint promtool prom/prometheus:v3.15.0 \
  test rules deploy/monitoring/rules/tests/cmdctrl_test.yml
docker run --rm -v "$(pwd)/deploy/alloy:/alloy:ro" --entrypoint alloy grafana/alloy:v1.20.1 \
  fmt --test /alloy/config.alloy
scripts/go-docker.sh test ./cmd/server/ -run Monitoring
```
