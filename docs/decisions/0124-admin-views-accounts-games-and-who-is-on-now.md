# ADR 0124 — Admin views: accounts, games and who is on now

**Status:** Accepted · 2026-10-05 · S64 — Admin views: accounts, games and who is on now. The owner accepted it unchanged in review on #2297 the same day, Calls made here included.
**Issues:** [#2296](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2296) (this change, and S64's tracker); relates to [#2281](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2281) (ADR 0123's dashboards, whose tiles link here).
**Owner decisions:** the three answers of 2026-10-05, quoted under [Owner decisions](#owner-decisions-2026-10-05). They are binding. This ADR also makes calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before PR 2 lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-05. I ran `git fetch --all --prune` and listed every file any remote branch has had under `docs/decisions/` (`git log --remotes --name-only -- docs/decisions/`, a superset of the tips). The 37 remote heads are `origin/develop`, `origin/main` and 35 chore, docs, feat, fix, repro and wip branches. The highest number on any of them is 0123, on `origin/develop`. The one open PR (#2162) adds no ADR. This ADR takes **0124**.
**Amends:** [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) call 17 (an agent seat gains a database column, §6). [ADR 0112](0112-signed-in-home-player-mode-and-one-decks-page.md) §2 item 8 (an admin who opens `#/admin` lands on the admin views, §7) and §3 item 7 (the sign-in return route also accepts an admin view, §7). [ADR 0076](0076-tutorial.md) §2.2 only as far as the admin views go: practice tables are listed there (§4), and nowhere else changes.
**Builds on:** [ADR 0112](0112-signed-in-home-player-mode-and-one-decks-page.md) §2 (admin mode, and the answer table every admin call site must join), [ADR 0110](0110-remember-me.md) §3 (admin is checked on every request, never stored in a token), [ADR 0051](0051-user-database.md) (users, identities, games, seats, decks), [ADR 0044](0044-surviving-a-deploy.md) decision 3 (HMAC sessions, so there is no session list), [ADR 0017](0017-bug-report-button.md) §9 (the server never logs a URL's query string), [ADR 0092](0092-public-roadmap-and-site-portal.md) (hash routing and the site header), [ADR 0095](0095-deck-coverage-and-deck-requests.md) (deck requests), [ADR 0122](0122-an-agent-at-the-table-a-local-mcp-seat.md) (agent seats), [ADR 0123](0123-monitoring-metrics-logs-dashboards-and-alerts.md) (the metrics, the scrape-time accessors and the Overview dashboard).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

On 2026-10-05, looking at the new Grafana Overview (ADR 0123), the owner asked to make it "smarter": click the Registered accounts tile and see the accounts behind the number, and so on for the other tiles. Today the only way to see who or what is behind a number is to query prod's database by hand. Metrics cannot carry it, because ADR 0123 §3 bans IDs and names from metric labels, and `TestMetricLabelsAreClosedSets` enforces that.

### What exists (checked on `origin/develop` at `12fa8e77`)

**Admin gating.**
- `requireAdmin` (`server/internal/lobby/admins.go:232`) is every admin route's gate: any valid session, then `isAdmin`, else 403 `{"error":"admin only"}`. A missing or bad credential is the auth middleware's 401. It logs one Info line, `admin action`, with `r.Method+" "+r.URL.Path` (`admins.go:241`): the path, never the query string.
- `isAdmin` is the shared token, or an allowlisted person in admin mode (`isAdminPrincipal`, `admins.go:206`). An allowlisted person in player mode gets exactly what a non-admin gets.
- Four guard tests hold that line:
  - `TestRoleAdminIsComparedOnlyInTheAdminPredicates` (`lobby/admins_test.go:369`);
  - `TestAllowlistAndModeAreAskedOnlyInAdminsGo` (`lobby/player_mode_test.go:555`);
  - `TestEveryAdminCallSiteIsInTheSameAnswerTable` (`player_mode_test.go:1154`). Its census, `adminCallSites` (`:714`), finds every `Handle` call inside `Handler` whose handler goes through `requireAdmin`, by its literal pattern, and every function that calls `isAdmin`, `isAdminPrincipal` or `requireAdmin`. It parses only the `lobby` package;
  - `TestPlayerModeAnswersExactlyAsANonAdmin` (`:1125`), which asks each site as a non-admin, in player mode and in admin mode.
- `TestRequireAdminOnEveryAdminRoute` (`admins_test.go:442`) asks each admin route as a guest, a reclaim ticket, an unlisted person, an allowlisted person in player mode, and an admin.
- Existing admin routes under `/admin/`: `POST /admin/login` and `POST /admin/users/{id}/revoke-sessions` (`lobby/http.go:352,686`). `/admin/*` is already in Caddy's `@api` matcher (`deploy/Caddyfile:40`), the service worker's `API_PATH` (`client/src/sw/service-worker.js:86`) and Vite's proxy (`client/vite.config.ts:34`). A new route under it needs no new prefix anywhere.
- The admin actions this ADR links to: archive and unarchive (`POST`/`DELETE /games/{id}/archive`, `http.go:407-408`; archiving a running table is how `/c2-end` ends it), the replay (`GET /games/{id}/replay`, `:490`), and revoke (`adminRevokeUserSessions`, `lobby/revocation.go:98`).

**What the database holds** (migrations `0002`-`0009`; every time is Unix milliseconds):
- `users`: `id`, `display_name`, `avatar_url` (a same-origin `/avatars/<snowflake>/<hash>.png` path, `users/users.go:42`), `created_at`, `last_seen_at`, `sessions_invalid_before`, `last_deck`, `admin_mode_at`. `last_seen_at` is written only at a Discord sign-in (`users/store_sql.go:129`), so it measures sign-ins, not play.
- `identities`: `provider`, `subject` (the Discord snowflake), `user_id`, `display_name`, `avatar_hash`, `refresh_token` (sealed, ADR 0051 decision 5), `scopes`, `linked_at`, `refreshed_at`.
- `games`: `id`, `name`, `created_by`, `state`, `created_at`, `started_at`, `ended_at`, `archived_at`, `winner_seat`, `host_player_id`, `host_discord_id`, `outcome` (`win`, `draw`, or NULL for unknown, an admin close, or not ended; migration `0007`).
- `seats`: `game_id`, `seat`, `player_id`, `user_id`, `guest_name`, `bot_tier`, `deck_id`, `deck_name`, `pending_discord_id`. **No agent column.** ADR 0122 call 17 kept the agent badge on `game.Player` and the snapshot only, and `loadEntry` reads it back from there (`lobby/persist.go:374`). A table whose restore point is gone has lost its agent badge.
- `decks`: `id`, `owner_id`, `name`, `source_format`, `source_text`, `commanders`, `card_count`, `created_at`, `updated_at`, `source_url`.
- `deck_requests` (`deck_key`, `issue_number`, `issue_url`, `created_at`) and `deck_request_asks` (`deck_key`, `requester`, `at`). `requester` is `discord:<snowflake>` (migration `0006`), so an account's asks are found through its identity.
- `user_settings`, `table_setups`, `invites` (hashed), `schema_migrations`.
- **Practice tables have no row.** They are never persisted (ADR 0076 §2.2, `lobby/practice.go`).

**What is only in memory.**
- The lobby's tables: `GameMeta` and `SeatInfo` (`lobby/lobby.go:77,148`), with each seat's `IsBot`, `BotTier`, `IsAgent`, `AgentClient` (`:189`), `UserID` (`:202`, never on the wire), and the table's `Practice` flag (`:130`). A table is in memory while its room is: from creation, or from a restore point at boot (`RestoreFromDisk`, `persist.go:277`).
- The hub's sockets: each `Client` (`ws/hub.go:739`) has `gameID`, `playerID`, `readOnly`, `admin`, `userID` and `issuedAt`. It has **no connection time**. `Binding` (`hub.go:117`) carries the same fields.
- ADR 0123's scrape-time accessors: `Hub.MetricsSockets()` (`hub.go:297`) copies every socket's game, seat and role under the hub's read lock only, and `Lobby.MetricsTables()` (`lobby/metrics.go:71`) copies every table's state, archived and practice flags and seat kinds under `l.mu` only. `Lobby.MetricsLiveUsers()` (`:99`) lists the accounts seated at a running table. The tables collector (`metrics/tables.go:144`) reads the hub, then the lobby, never holding both. These types carry no names and no user IDs, by ADR 0123's rule.

**How the Overview's numbers are defined** (`metrics/tables.go`, `users/metrics.go`):
- *Active games*: tables in state `active`, not archived, not practice. *Waiting in the lobby*: state `lobby`, not archived.
- *Players connected*: seats at active, non-archived, non-practice tables with at least one live seat socket, human or agent.
- *Spectators*: every live read-only socket. *Bot seats*: bot seats at active tables. *Practice tables*: open practice tables.
- *Registered accounts*: `users` rows. *Accounts that played* in 1d, 7d or 30d: accounts seated at a started table whose end (`ended_at`, else `archived_at`) is in the window (`SQLStore.LastPlayed`, `users/metrics.go:31`), plus accounts seated at a running table now.

**The client.**
- Hash routing (`client/src/lib/router.ts`). `#/admin` is already a route: `adminLogin`, the shared token's form (`router.ts:103`). ADR 0112 §2 item 8 sends the token's session and any allowlisted person from it to `#/lobby` (`lib/signedInHome.ts:62`).
- `GET /me` reports `admin` (the effective answer), `admin_allowed` and `admin_mode`. `lib/admin.ts` keeps them on the session, and `isAdmin(session)` reads `admin`.
- `SiteHeader.svelte` has the site nav (`:170`): Lobby, My games, Decks, Catalog, Roadmap, Home. Nothing in it is admin-only.
- The decks page's sign-in return (`AFTER_SIGN_IN_KEY`, `lib/decksPage.ts:124`) accepts only the decks route (`isDecksReturn`, `:131`).

**The Overview dashboard** (`deploy/monitoring/dashboards/overview.json`, uid `cmdctrl-overview`) has one template variable, `env`, a custom `prod,dev`. No panel has a data link. The monitoring VM syncs this directory from `main` (ADR 0123 §6). The dashboards' Go checks are in `server/cmd/server/monitoring_config_test.go` (`TestMonitoringDashboardsAreProvisionable`, `:159`). The blackbox probe series carry each site's URL as `instance` (`probe_success{job="blackbox", env="prod", instance="https://cmd.labxp.io/healthz"}`, `deploy/monitoring/rules/tests/cmdctrl_test.yml:20`).

### Owner decisions (2026-10-05)

From #2296:

1. **Where:** admin pages in the app, reading the live database, behind admin mode (ADR 0112). Grafana's tiles link to them. Not a database copy in Grafana, and not an admin API read by a Grafana plugin.
2. **Views:**
   - **Accounts:** every account.
   - **One account's detail:** their games, saved decks, deck requests and sign-in state.
   - **Games:** every table, with its seats and outcome.
   - **Live now:** who is connected, to which table, as player or spectator, plus bot and agent seats.
3. **Read-only first.** Existing admin actions (archive or end a table, revoke sessions) are linked from the right rows, not rebuilt.

The brief that came with the decisions spelled out each view's columns: accounts with name, avatar, first seen, last sign-in, games played and last played; an account's games, saved decks, deck requests and sign-in state, where sessions are stateless (ADR 0044 decision 3), so there is no session list and the view shows `last_seen_at` and `sessions_invalid_before` and links the existing revoke; every table's state, created, started and ended times, outcome and winner, archived and practice flags, and its seats (who, guest name, bot tier, agent and agent client); and who is connected now, to which table, as seat or spectator, plus bot and agent seats, as the drill-down for "Players connected". All four views ship in this sprint.

---

## Decision

### 1. The shape

```
 Grafana (LAN) ── data link ──▶ https://cmd[-dev].labxp.io/#/admin/<view>?<filter>
                                        │ the client (admin mode only)
                                        ▼
 GET /admin/users, /admin/users/{id}, /admin/games, /admin/games/{id}, /admin/live
   requireAdmin (lobby.Handler) ──▶ lobby/admin_views.go handlers
        │                                   │
        ▼                                   ▼
   internal/adminview (SELECT only)    live overlay: Hub.LiveSockets(), then Lobby.LiveTables()
        │                              (one lock at a time, as ADR 0123's accessors)
        ▼
   SQLite (the live database)
```

- **Five read-only routes**, all `GET`, all behind `requireAdmin`, all registered in `lobby.Handler` with literal patterns. The census in `adminCallSites` therefore finds each one, and `TestEveryAdminCallSiteIsInTheSameAnswerTable` fails until the answer table asks it (§8).
- **The database is the record; memory is the overlay.** History comes from SQL. What only memory knows (who is connected, a table's live state, practice tables, the agent badge of a seat whose row predates §6's column) is merged on top.
- **The client gets four pages** under `#/admin/…` (§7), and the Overview's tiles link to them (§9).

### 2. Routes

Every route answers JSON. Times are Unix milliseconds, as on `/me/games`, and absent when unknown. Each response has `generated_at`.

| Route | Answers | Filters | Bound |
|---|---|---|---|
| `GET /admin/users` | every account, one row each (§3.1) | `played=1d\|7d\|30d` | 1,000 rows, then `truncated: true` |
| `GET /admin/users/{id}` | one account: its row, sign-in state, games, decks, deck requests (§3.2) | — | games 500, deck requests 100, each with its own `…_truncated`; decks are capped at 200 by the library |
| `GET /admin/games` | tables, newest first (§3.3) | `state=lobby\|active\|ended`, `archived=true\|false`, `practice=exclude\|include\|only`, `user=<account id>` | `limit` (default 50, max 200) and an opaque `cursor`; `next_cursor` when there is more |
| `GET /admin/games/{id}` | one table: the same row, plus its live connections (§3.3) | — | — |
| `GET /admin/live` | who is connected now, by table, with totals (§3.4) | — | — |

- **Who may call them:** the shared token, or an allowlisted person in admin mode: `requireAdmin`, unchanged. **Everyone else gets what a non-admin gets from any admin route:** 401 with no valid session, and 403 `{"error":"admin only"}` with one. An allowlisted person in player mode gets that same 403. It is 403 rather than 404 because the existing admin routes answer 403, and the same-answer table compares answers, so a new kind of refusal would be the odd one out. The routes' existence is no secret: they are in this ADR and in `docs/lobby.md`.
- **Unknown filter values are a 400** that names the parameter. A missing parameter means "any". `practice` defaults to `exclude`, so the unfiltered list counts what the Grafana tiles count.
- **Paths:** `/admin/users`, matching the existing `POST /admin/users/{id}/revoke-sessions` and the `users` table. The client says "Accounts", the owner's word.
- **No database:** `/admin/live` still works, from memory. The other four answer 503 `admin views need the user database, and this server has none`, as `PUT /me/admin-mode` does.
- **No new rate limit.** Only admins reach the handlers, every query is bounded, and Live now's polling is bounded by the client (§7).
- **Metrics:** the five patterns join `cmdctrl_http_requests_total{route}` on their own, because the route label comes from the mux's registered patterns (ADR 0123 amendment, "The route label").

### 3. What each route returns

The field lists below are the whole response. A field not listed is not served, and §8's allowlist test pins that.

#### 3.1 An account row (`GET /admin/users`, and the top of `/admin/users/{id}`)

| Field | From |
|---|---|
| `id` | `users.id` (an internal UUID, which the revoke route already takes) |
| `name` | `users.display_name` |
| `avatar_url` | `users.avatar_url`, when set |
| `first_seen_at` | `users.created_at`: the first Discord sign-in |
| `last_sign_in_at` | `users.last_seen_at` |
| `games_played` | distinct tables the account holds a seat at that have started (`started_at` set), in any state |
| `last_played_at` | the latest end (`ended_at`, else `archived_at`) of those tables |
| `playing_now` | seated at a running table in memory now (the `MetricsLiveUsers` rule) |

- **`played=<window>`** keeps an account whose `last_played_at` falls in the window or that is `playing_now`. That is exactly `cmdctrl_users_played{window}`, so the page and the tile count the same accounts. §8 tests the two against each other.
- **Sorted** by `playing_now`, then `last_played_at`, then `last_sign_in_at`, newest first. The client re-sorts on any column, since the whole list is on the page.
- **One query** with `LEFT JOIN seats` and `LEFT JOIN games` and `GROUP BY users.id`, plus the in-memory live list. No query per row.

#### 3.2 One account (`GET /admin/users/{id}`)

- **`account`**: the row in §3.1.
- **`sign_in`**:
  - `last_sign_in_at` (`users.last_seen_at`);
  - `discord_linked_at` (`identities.linked_at`);
  - `sessions_invalid_before` (`users.sessions_invalid_before`, absent while 0);
  - `revoke_path`: `/admin/users/{id}/revoke-sessions`.

  Sessions are HMAC tokens with no server-side row (ADR 0044 decision 3), so there is no list of sessions, and the view says so in one line. `sessions_invalid_before` is the only per-person session state there is: a session issued at or before it is refused.
- **`games`**: the account's tables, newest first, each a §3.3 row with `their_seat` (the seat number they hold). Two queries: the account's seats with their games, then every seat at those tables, the way `SQLStore.SeatsOfUser` does it (`lobby/store_mygames.go:48`).
- **`decks`**: the saved library, newest first: `id`, `name`, `format` (`source_format`), `source_url` when set, `commanders`, `card_count`, `created_at`, `updated_at`. **Not `source_text`**, the list itself, and no coverage report, which costs a catalog pass per deck. The owner can open a list from its link.
- **`deck_requests`**: the account's asks, newest first: `deck_key`, `asked_at`, `issue_number` and `issue_url` (the issue now tracking that deck). One query joins `deck_request_asks.requester` to `'discord:' || identities.subject` for the account. The snowflake is compared in SQL and never leaves it.
- **Not shown:** synced settings, the last table setup, the last deck, admin mode and allowlist membership. ADR 0112's "the allowlist is never served" stands, and none of those were asked for.

#### 3.3 A table row (`GET /admin/games`, `/admin/games/{id}`, and an account's games)

| Field | From |
|---|---|
| `id`, `name`, `state` | `games`; for a table in memory, `state` is the lobby's cached state, which is fresher |
| `created_at`, `started_at`, `ended_at`, `archived_at` | `games` |
| `outcome` | `win` or `draw` from `games.outcome`; `closed` for a table that ended, or was archived after it started, with no outcome recorded (migration 0007's NULL, the same rule as `cmdctrl_games_ended_total{outcome="closed"}`); absent for a table that has not ended |
| `winner_seat` | `games.winner_seat`, when set; the client labels it from `seats` |
| `creator` | `{id, name}` from `games.created_by` joined to `users`; absent for a table the token created |
| `practice` | from memory; a practice table has no row |
| `loaded` | the table's room is in memory now. A row in state `lobby` or `active` that is not loaded is a table that did not come back after a restart. Today nothing shows those. |
| `seats` | below |
| `spectators_connected` | loaded tables only |

Each seat:

| Field | From |
|---|---|
| `seat` | `seats.seat` |
| `kind` | `human`, `bot` or `agent`: the same three kinds as `cmdctrl_seats{kind}` |
| `account` | `{id, name, avatar_url}` when `seats.user_id` is set |
| `guest_name` | `seats.guest_name`, when there is no account |
| `discord_pending` | `true` for a Discord seat whose person has not signed in again (`seats.pending_discord_id` set). The snowflake itself is not served. |
| `bot_tier` | `seats.bot_tier`, for a bot |
| `agent_client` | the normalised MCP client name (`NormalizeAgentClient`), for an agent: from memory for a loaded table, else from `seats.agent_client` (§6) |
| `deck_name` | `seats.deck_name` |
| `host` | `seats.player_id = games.host_player_id` |
| `connected` | the number of live sockets bound to this seat; loaded tables only |

- `player_id` is not served. Nothing in the views needs it.
- **The list query** is keyset-paginated on `(created_at, id)`, newest first, with the filters in SQL. The seats for the page come in one more query (`WHERE game_id IN (…)`, joined to `users` for names). The loaded tables' overlay is one `Lobby.LiveTables()` copy (§5).
- **Practice tables** come only from memory. `practice=only` lists them (newest first, never more than a handful, unpaginated). `include` puts them first on the first page.
- **`/admin/games/{id}`** adds `connections`: each live socket at the table, as `{kind: seat|spectator|admin, seat, account, since}`. `account` is absent for a guest or the token, and `since` is the connection time (§5). It answers 404 for an ID in neither the database nor memory.

#### 3.4 Live now (`GET /admin/live`)

- **`tables`:** every table that has a live socket, plus every running, non-archived table (a bots-only table has no socket). Each has `id`, `name`, `state`, `practice`, `archived` and:
  - `seats`: as in §3.3, with `connected` and `since` (the earliest of the seat's sockets). Bot seats are listed with their tier, and agent seats with their client.
  - `spectators`: one entry per read-only socket, `{account, since}`. `account` is absent for a guest spectator, who is shown as "guest spectator". A spectator socket carries no name, and none is added.
  - `admins`: one entry per admin-bound socket, `{account, as_seat, since}`. `account` is absent for the token, and `as_seat` is set when the admin is bound to a seat.
- **`unbound_sockets`**: a count of sockets whose game the lobby does not hold. It should be 0, and a non-zero count is worth a look.
- **`totals`**: `players_connected`, `spectators`, `bot_seats`, `practice_tables` and `admin_views`. The first four are computed by the **same function** as the Overview's tiles (§5), from the same two copies, so they can differ from Grafana only by the time between a scrape and a page load. The page puts each total beside the rows it counts. "Players connected" counts running tables only, as the tile does, while the page also lists the people waiting at lobby tables.

### 4. What personal data the views show, and what they never show

| Shown | Never shown |
|---|---|
| Account display names and avatars | Discord snowflakes as a field of their own. The avatar path contains one, as it does everywhere the site shows an avatar today, and the seat list (`SeatInfo.discord_id`) already serves it at a table. |
| Account UUIDs (the revoke route's key) and game UUIDs | Refresh tokens, sealed or not, and `scopes` |
| Guest names, bot tiers, agent clients | Invite tokens, invite hashes, reclaim tickets |
| Deck names, commanders, links and counts | Deck lists (`source_text`) |
| Deck request keys and issue links | `requester` values (`discord:<snowflake>`) |
| Sign-in and revocation times | IP addresses and remote addresses, which the database does not hold and the views do not add |
| | Emails: none are stored (the Discord scope is `identify`) |
| | Synced settings, table setups, admin mode, allowlist membership |

**Practice tables are listed** in the admin views, unlike in `GET /games` (ADR 0076 §2.2 keeps them out of the lobby list, the admin's included). The Overview has a Practice tables tile, and its drill-down has to show them. Nothing else about practice tables changes.

**Logging.** `requireAdmin` already writes one `admin action` line per request with the method and path. The path carries at most a game or account UUID, which the logs already carry, and never the query string, so a filter is never logged (ADR 0017 §9). The handlers add no log line of their own, except an ERROR on a failed query, which carries the error and no values.

### 5. Where the code lives

- **`server/internal/adminview`** (new): the read-only store and the response types.
  - `adminview.Store` with `Accounts`, `Account`, `Games`, `Game`. Every SQL statement for the views is in this one package.
  - It takes the server's `*db.DB` and **only reads**. A guard test (§8) fails if any non-test file calls `ExecContext` or `BeginTx`, or passes a query string that does not start with `SELECT` or `WITH`.
  - It imports `internal/db` and `uuid`, and not `lobby` or `ws`. The live overlay is passed in as plain values.
  - `adminview.Merge…` functions are pure: rows plus overlay in, response out. They are unit-tested without a database or a hub.
- **`server/internal/lobby/admin_views.go`**: the five handlers. They sit in `lobby` so `requireAdmin` and the answer-table census see them. Each takes the overlay, runs the store, merges and writes.
- **`lobby.Config` gains `AdminViews adminview.Store`** (nil: the 503 above) and **`LiveSockets`**, an interface the hub implements. `main.go` wires both, as it wires `AdminSockets` today.
- **`ws.Hub.LiveSockets() []ws.LiveSocket`** (new), beside `MetricsSockets`. It copies each socket's `GameID`, `PlayerID`, `UserID`, `ReadOnly`, `Admin` and `ConnectedAt` under the hub's read lock only. `Client` gains `connectedAt`, set when the hub registers it. `MetricsSockets` stays as it is: the metrics package carries no user IDs, by ADR 0123's rule.
- **`lobby.Lobby.LiveTables() []lobby.LiveTable`** (new), beside `MetricsTables`. It copies each table's ID, name, cached state, archived and practice flags, and each seat's number, player ID, kind, user ID, name, bot tier, agent client and host flag, under `l.mu` only. Like `MetricsTables`, it reads the cached lifecycle state, takes no room or game lock, and writes nothing.
- **Never two locks.** A handler calls `LiveSockets()`, then `LiveTables()`, then the database, each released before the next, in the same order as the metrics collector. The two copies are not one instant, and a socket that connects between them can be counted against a table read after it. That is the skew ADR 0123 already accepts for a gauge, and the next refresh is right.
- **One definition of the totals.** The counting loop in `tablesCollector.Collect` moves into a pure `metrics.Tally(tables []metrics.Table, sockets []metrics.Socket) metrics.Counts`. The collector calls it to emit its gauges, and `/admin/live` calls it, with `LiveTable` and `LiveSocket` converted to the metrics types, for its totals. The collector's output is unchanged, and its existing tests hold it.
- **Names for live rows** come from the seats (`SeatInfo.DisplayName`, `Name`) and, for spectators and admins with an account, from one `SELECT id, display_name, avatar_url FROM users WHERE id IN (…)`.

### 6. The agent badge gets a column

Owner decision 2 asks for the agent and its client on every table, but today the badge lives only in the engine snapshot (ADR 0122 call 17). A table whose restore point is gone, which is any ended table after a while, shows its agent seat as a guest.

- **Migration `0010_seat_agent.sql`:** `ALTER TABLE seats ADD COLUMN agent_client TEXT`. NULL means not an agent. The value is the normalised client name, never empty for an agent (an empty name is `unknown`). The migration is additive, like `0004` and `0007`.
- **`seatRecords`** (`persist.go:72`) writes it from `SeatInfo.AgentClient` when `IsAgent`. The seat list is already rewritten at every join, so a new agent seat lands with its join.
- **`loadEntry`** keeps the engine as the authority for a restored table. It reads the column only when the snapshot's player is missing, which cannot happen today. The snapshot shape does not change.
- **No backfill.** Rows written before the migration have no value. A table still in memory shows its badge from memory, and §3.3 says which source each seat used.

### 7. The client

**Routes** (`lib/router.ts`). A new route `{ name: "adminViews", view, … }`:

| Hash | View |
|---|---|
| `#/admin/live` | Live now |
| `#/admin/games`, with `?state=`, `?archived=`, `?practice=` | Games |
| `#/admin/games/<id>` | One table |
| `#/admin/accounts`, with `?played=1d\|7d\|30d` and `?sort=` | Accounts |
| `#/admin/accounts/<id>` | One account |

- **`#/admin` alone stays the token's form** (`adminLogin`), with one change to ADR 0112 §2 item 8: a session that is an admin now (the token's own, or an allowlisted person in admin mode) goes from `#/admin` to `#/admin/live`, not `#/lobby`. An allowlisted person in player mode still goes to `#/lobby`.
- **`sort`** is client-only (`last_played`, `first_seen`, `name`, `games`). The server never sees it.
- **One page component**, `routes/Admin.svelte`, with a tab strip (`navigation "admin views"`: Live now, Games, Accounts) and one child per view under `lib/components/admin/`. The pure parts (parsing and building the filter query, the seat label, the outcome line, the relative times, the poll schedule) live in `lib/adminViews.ts`.
- **Rows link on.** An account opens its detail. A table opens its detail, and from there **Open table** (`#/games/<id>`, where an admin's socket gets the full view, as today). A seat with an account opens that account.
- **The linked actions**, each the existing route with the existing confirmation copy from the Lobby, never a new server action:
  - **Archive** (`POST /games/{id}/archive`) on a table that is not archived. On a running table this ends it, as `/c2-end` does, and the confirm says so.
  - **Unarchive** (`DELETE /games/{id}/archive`) on an archived one.
  - **Replay** (`GET /games/{id}/replay`) on an ended table.
  - **Revoke sessions** (`POST /admin/users/{id}/revoke-sessions`) on an account, with a confirm that names the person and says they stay signed out until they sign in with Discord again.

  Delete is not offered here. It stays on the Lobby's admin list, where it is today.

**Who sees what.**
- **The header** gets an **Admin** link in the site nav, last, for a session where `isAdmin(session)` holds: the token, or admin mode on. It goes to `#/admin/live`. It goes away when admin mode lapses, through the same store. Home (`#/home`) stays the public site map and gets no admin card.
- **A signed-out visitor** on an admin route goes to `#/login`, as on every gated route. Before that, the route is stored in `sessionStorage["cmdctrl.afterSignIn"]`, and `oauthCompleteTarget` returns there after sign-in. `isDecksReturn` widens to `isReturnRoute`, which accepts the decks route and the admin views and nothing else (ADR 0112 §3 item 7). A Grafana link opened in a signed-out browser therefore lands on the page it named after sign-in, or on the message below.
- **A session that is not an admin** sees a message and no data. The page does not fetch when `isAdmin(session)` is false.
  - An allowlisted person in player mode: "Admin mode is off. Switch it on to see this page." with the same switch the account menu has. After the switch the page loads.
  - Anyone else: "This page is for admins."
- **A stale answer** (the page fetched, and the server said 403) shows the same message and asks `/me` again (`refreshAdminStatus`). The server is the gate, and a stale `admin: true` only costs one refused request.

**Refreshing.**
- **Live now** polls `GET /admin/live` every 10 seconds while the tab is visible. It stops when the tab is hidden and fetches once when it is visible again. After 10 minutes with no pointer or key input on the page, it pauses and shows "Paused. Resume", so a tab left open overnight does not write a log line every 10 seconds.
- **Games and Accounts** load once and have a Refresh button.
- **Times** show as "4 min ago", with the exact local time in the tooltip.

**Labels.** The views add four names to the labels contract, because the nightly e2e spec in PR 5 selects on them: `navigation "admin views"` with its links `Live now`, `Games` and `Accounts`; `region "live now"`; `table "games"`; `table "accounts"`.

### 8. Tests

**Server.**
- **The answer table.** The census makes each of the five patterns a site. Each gets a probe in `sameAnswerProbes` with `differs: true`, and `TestPlayerModeAnswersExactlyAsANonAdmin` asks it from all three sessions. `TestRequireAdminOnEveryAdminRoute` gains the five routes, with 401 without a session, 403 for a guest, a reclaim ticket, an unlisted person and player mode, and 200 for the token and admin mode.
- **The field allowlist,** `TestAdminViewsServeOnlyTheirFields`. It seeds a database with a sealed refresh token, an identity, invites, a reclaim ticket, a deck with its list, a deck request and an agent seat. Then it asks each route as an admin and walks the JSON:
  - every key path must be on the route's pinned list, so a new field is a deliberate edit;
  - no key may contain `token`, `secret`, `refresh`, `password`, `ip`, `remote`, `addr`, `invite`, `subject`, `snowflake`, `discord_id`, `scope`, `email` or `requester`;
  - the seeded snowflake may appear only inside an `avatar_url` value. The invite tokens, the reclaim ticket, the deck list and the refresh token may not appear at all.
- **The store against a seeded SQLite file** (`internal/adminview`), using the real migrations:
  - `games_played` counts started tables only, and `last_played_at` follows the `ended_at`-else-`archived_at` rule;
  - `played=7d` returns exactly the accounts `users.SQLStore.LastPlayed` returns for the same instant, plus the live list;
  - keyset pages never repeat or skip a table, including tables with the same `created_at`;
  - each seat kind maps right: account, guest, pending Discord, bot, agent (from the column);
  - the `outcome` mapping, and the deck-request join through `identities`;
  - an account's games use two queries, whatever the number of games.
- **SELECT only,** `TestAdminViewStoreOnlyReads`: it parses the package and fails on `ExecContext`, `BeginTx`, or a query literal that does not start with `SELECT` or `WITH`.
- **Live totals equal the tiles.** A table with two seats, one connected guest, one spectator, one bot and one agent, as in ADR 0123's collector test, and a practice table. `/admin/live`'s totals must equal what `metrics.NewTablesCollector` reports over the same hub and lobby. `metrics.Tally` has its own table test, and the collector's existing tests stay green unchanged.
- **Locks.** `LiveSockets` and `LiveTables` are each tested to copy and release. `/admin/live` runs under `-race` while sockets connect and tables start.
- **No query string in the log.** A `GET /admin/games?state=active` writes an `admin action` line whose `action` has no `?`.
- **Migration 0010.** `seatRecords` writes `agent_client` for an agent seat and NULL otherwise. `loadEntry` still takes the badge from the snapshot.

**Client (vitest).**
- `parseHash` for each admin hash, with its filters, and `#/admin` alone still `adminLogin`.
- Filter round trip: a filter built into a hash parses back to itself, and an unknown value is dropped, never sent.
- `routeRedirect` and the return route: signed out goes to `#/login` with the route stored; an admin on `#/admin` goes to `#/admin/live`; player mode on `#/admin` goes to `#/lobby`; `isReturnRoute` refuses anything but the decks page and the admin views.
- The pure helpers: the seat label (account, guest, pending Discord, bot tier, agent client), the outcome line, relative times, and the poll schedule (visible, hidden, idle pause, resume).
- A render test per view from a fixture response, and one for the player-mode message, which must render no row.

**The dashboard links,** `TestOverviewTilesLinkToTheAdminViews` in `server/cmd/server/monitoring_config_test.go`:
- every panel named in §9's table has at least one data link;
- each link's URL starts with `${site}/#/admin/`, names one of the four list views, and uses only that view's parameters and their closed values (or a `${__field.labels.…}` placeholder for a label of the panel's own query);
- `site` is a hidden variable over the `prometheus` datasource that reads `probe_success` with `env="$env"`.

**E2E.** One nightly Playwright spec: the token session opens `#/admin/live`, sees the table the spec created with its seat connected, opens Games, and finds the same table.

### 9. Grafana links

**The host follows `env`.** A new hidden template variable on the Overview:

```json
{
  "name": "site", "type": "query", "hide": 2, "refresh": 2,
  "datasource": { "type": "prometheus", "uid": "prometheus" },
  "query": { "query": "label_values(probe_success{job=\"blackbox\", env=\"$env\"}, instance)", "refId": "site" },
  "regex": "/^(https:\\/\\/[^\\/]+)\\//"
}
```

It takes the site's origin from the URL the blackbox exporter already probes for that env: `https://cmd.labxp.io` for prod and `https://cmd-dev.labxp.io` for dev. Nothing in this repo names a host twice, and a renamed site changes in one place, the probe target in HomeLab.

**The links.** Each is a data link (`fieldConfig.defaults.links`, or an override for one query), titled "Open in cmd_and_ctrl", opening a new tab:

| Panel (id) | Link |
|---|---|
| Active games (2) | `${site}/#/admin/games?state=active&archived=false` |
| Waiting in the lobby (3) | `${site}/#/admin/games?state=lobby&archived=false` |
| Players connected (4) | `${site}/#/admin/live` |
| Spectators (5) | `${site}/#/admin/live` |
| Bot seats (6) | `${site}/#/admin/live` |
| Practice tables (7) | `${site}/#/admin/games?practice=only` |
| Players connected, by kind and account (8) | `${site}/#/admin/live` |
| Games by state (9) | `${site}/#/admin/games?state=${__field.labels.state}&archived=false`; the practice series, by an override on its query: `?practice=only` |
| Registered accounts (11) | `${site}/#/admin/accounts` |
| Accounts that played (12) | `${site}/#/admin/accounts?played=${__field.labels.window}` |
| Seats at active tables, by kind (13) | `${site}/#/admin/live` |
| Games created, started and ended per day (15) | `${site}/#/admin/games` |
| New accounts per day (17) | `${site}/#/admin/accounts?sort=first_seen` |

- **Panel 12's three queries keep their label.** `max(cmdctrl_users_played{env="$env", window="1d"})` becomes `max by (window) (…)`, and likewise for 7d and 30d, so each bar can link to its own window. The bars show the same numbers.
- **The links ship in `deploy/monitoring/dashboards/overview.json`.** The monitoring VM syncs that directory from `main`, so they appear at the next promotion, and the pages they open must be on `main` by then too (Delivery).
- **The admin still signs in to the site.** Grafana is LAN-only and the site is public, so a link opens the public site in the admin's own browser, which needs a session in admin mode to see anything (§7). Grafana passes no credential, and the link carries none.

---

## Delivery

Each PR goes into `develop` under Sprint S64 and Issue #2296.

| PR | What | Size | Needs | Parallel with |
|---|---|---|---|---|
| 1 | **This ADR**, the AGENTS.md §3 ADR range line, and the S64 section and index row in `docs/sprints.md`. Docs only. | S | — | anything |
| 2 | **The lobby's half** (§5, §6): migration `0010_seat_agent.sql`, `seatRecords` and `loadEntry` for the agent column, and `Lobby.LiveTables` with its lock test. | S | 1 | 7 |
| 3 | **Server: accounts and games** (§2, §3.1–3.3, §4, §5): `internal/adminview` (store, types, merge functions), `lobby/admin_views.go` with `GET /admin/users`, `/admin/users/{id}`, `/admin/games`, `/admin/games/{id}`, `Config.AdminViews`; the answer-table probes, the field allowlist, the SELECT-only guard and the store tests (§8); `docs/lobby.md`; AGENTS.md §3 (the new package) and §5 (the endpoints). | L | 2 | 4, 7 |
| 4 | **Server: Live now** (§3.4, §5): `Client.connectedAt`, `Hub.LiveSockets`, `Config.LiveSockets`, `metrics.Tally` and the collector's move onto it, `GET /admin/live`, the totals and lock tests; `docs/lobby.md` and AGENTS.md §5. | M | 2 | 3, 7 |
| 5 | **Client: the admin pages** (§7): the route and its parsing, `routes/Admin.svelte` and the three list views and two detail views, the header link, the not-an-admin message, `isReturnRoute` and the `#/admin` change, the linked actions, Live now's polling, the vitest and render tests, the e2e spec (`tests-e2e/tests/admin-views.spec.ts`, run on the branch with `gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), and the four labels in AGENTS.md §5's contract. | L | 3, 4 | 7 |
| 6 | **Docs and the exit check:** `docs/monitoring.md` gains "from a tile to the rows behind it"; the exit check below, with its evidence on #2296. | S | 5, 7 | — |
| 7 | **Grafana links** (§9): the `site` variable, the data links and panel 12's queries in `overview.json`, and `TestOverviewTilesLinkToTheAdminViews`. | S | 1 | 2–5 |

PRs 2 and 7 can start at once. PRs 3 and 4 start once PR 2 has merged, since both read `Lobby.LiveTables` and PR 3 reads `agent_client`. They run in parallel. They both add to `lobby.Config` and `main.go`, and whichever merges second rebases; the conflict is mechanical. PR 7 also touches `overview.json`, which `fix/s63-2281-instant-stat-panels` changes too, and the second to merge rebases. PR 7 may merge before PR 5, because nothing reads the dashboards from `develop`; the links only go live with the promotion that carries PR 5 as well. PR 5 is the largest, and it may be split by view (Live now; Games; Accounts) if review wants smaller pieces, with the route and shell in the first.

### Exit criteria

1. This ADR is accepted, and every PR in the Delivery table has merged into `develop`.
2. On cmd-dev, an allowlisted person in admin mode opens each of the four views, and in player mode sees the message and no data. Signed out, a link to `#/admin/games` comes back to that page after sign-in.
3. With one table running on cmd-dev (one human seat connected, one bot, one spectator), Live now's totals match the Overview's tiles for `env=dev`, and the Players connected tile's link opens Live now on cmd-dev.
4. Registered accounts and Accounts that played (7d) in Grafana match the row counts of `#/admin/accounts` and `#/admin/accounts?played=7d`.
5. The evidence is posted on #2296, and the work reaches `main` with the next promotion, which is when the Grafana links go live.

---

## Consequences

- **Every tile has a drill-down.** The number in Grafana and the rows behind it are one click apart, and the two are defined by the same code where they count the same thing.
- **The views are as fresh as the database.** No copy, no export and no second store. A query runs per page load. At ≤8 people and hundreds of tables, each is a few milliseconds.
- **Admins can see every account's games, decks and deck requests.** Before this, the same was possible only with a shell on the host. The view is limited to the token and allowlisted people in admin mode, and every load is an `admin action` log line naming who looked and at which path.
- **Live now's polling writes log lines.** At most one per 10 seconds per open, visible tab, and none after 10 minutes without input.
- **`seats` gains `agent_client`.** ADR 0122's "no migration" no longer holds for the badge, and the badge now survives the restore point. Rows from before the migration stay without it.
- **One small change for every admin.** An admin who opens `#/admin` now lands on Live now instead of the Lobby.
- **Practice tables become visible to admins**, in the admin views only.
- **The Grafana link host depends on the blackbox probe.** If HomeLab stops probing a site, its links lose their host until the probe returns.

## Out of scope

- **Any new admin action**: banning, renaming, merging accounts, editing a table, deleting from these pages. Owner decision 3. The existing actions are linked.
- **A session list.** HMAC sessions have no rows (ADR 0044 decision 3). Adding a session table is a separate decision.
- **Per-person activity history** beyond tables and deck requests (page views, sign-in history, IPs). Nothing records it, and these views add no recording.
- **Card-level views of a game** (the board, the log). Open table and Replay already do that.
- **Coverage reports for an account's decks.**
- **Grafana panels that list rows.** Owner decision 1 puts rows in the app, not in Grafana.
- **Search across accounts on the server.** The list is small, and the page filters it as you type.
- **Exports** (CSV and the like).

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review.

1. **Server paths `/admin/users…`, `/admin/games…` and `/admin/live`, under the existing `/admin/` prefix** (§2). `users` matches the existing revoke route. The client says "Accounts".
2. **A table-detail route and page** (`/admin/games/{id}`, `#/admin/games/<id>`) beside the four views, as the target of a row click and a stable deep link (§2, §3.3).
3. **403 for a non-admin and for player mode**, the same as every existing admin route, not 404 (§2).
4. **The bounds:** accounts unpaginated up to 1,000; games keyset-paginated, 50 a page up to 200; an account's games capped at 500 and its deck requests at 100 (§2).
5. **The unfiltered games list excludes practice tables**, so it counts what the tiles count. `practice=only` lists them (§2, §3.3).
6. **The field lists in §3 are the whole response, pinned by a test**, and no snowflake, deck list, requester, token, scope or address is served (§3, §4, §8).
7. **No deck lists and no coverage in an account's view**; name, commanders, link and card count only (§3.2).
8. **Admin mode and allowlist membership are not shown** on an account (§3.2). ADR 0112's "never served" stands.
9. **Practice tables are listed in the admin views**, though `GET /games` hides them from admins too (§4).
10. **A new package `internal/adminview` for the SQL, handlers in `lobby`**, so the existing answer-table census covers the routes with no change to the census (§5).
11. **`metrics.Tally` as the one definition of the totals**, shared by the collector and Live now (§5).
12. **`Client.connectedAt` and `Hub.LiveSockets`**, a sibling of `MetricsSockets`, rather than widening the metrics types with user IDs (§5).
13. **Migration `0010`: `seats.agent_client`**, amending ADR 0122 call 17, with no backfill (§6).
14. **`#/admin` alone stays the token's form, and sends an admin to `#/admin/live`** (§7), amending ADR 0112 §2 item 8.
15. **The sign-in return route accepts the admin views** (§7), amending ADR 0112 §3 item 7.
16. **An Admin link in the header for admins only, and no card on Home** (§7).
17. **Live now polls every 10 seconds while visible and pauses after 10 minutes idle** (§7), rather than a push channel. Every load stays logged by `requireAdmin`.
18. **Four labels join the contract** for one nightly e2e spec (§7, §8).
19. **The Grafana link host comes from the blackbox probe's `instance`** through a hidden `site` variable (§9), rather than a second hand-kept list of hosts.
20. **Panel 12's queries keep their `window` label**, so each bar links to its own window (§9).
21. **No new rate limit on the admin reads** (§2).

## Amendment, 2026-10-07: a second action, Remove playmat (#2501)

Decision 3 ("read-only first") linked the existing actions from their rows. The account view now carries a second one beside Revoke sessions: **Remove playmat**, which calls the route ADR 0128 §9 already added, `DELETE /admin/users/{id}/playmat`. No new route, no new rate limit, and the census is unchanged.

- **The read side.** `GET /admin/users/{id}` serves `playmat_url`, the same-origin `/playmats/<uuid>` path or absent for none. It sits at the top level of the account body, not on the account row, so the accounts list does not grow a field. The handler fills it from the playmat service (`AccountRows.PlaymatURL`); the store reads no new column. `TestAdminViewsServeOnlyTheirFields` pins it.
- **The client.** The Playmat card shows a small thumbnail of the current mat, so the admin sees what they are removing, loaded through the same token-bearing `playmatSrc` the table uses. The button opens a confirmation that names the person, as Revoke sessions does, and a successful removal reloads the account. With no mat there is a line saying so and no button.
- **Still true.** The views stay read-only apart from the linked actions; the person can upload another mat, so this removes an image and does not ban the feature.
