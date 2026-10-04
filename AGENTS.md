# AGENTS.md — working on cmd_and_ctrl

This file is for AI coding agents (Claude Code, Codex, Cursor, etc.) operating in
this repository. Humans should read [PLAN.md](PLAN.md) first for the project
vision and phased roadmap; this file is about *how* to work, not *what* to build.

---

## 1. What this project is

A private, personal 4-player Magic: The Gathering Commander sandbox
with a Go game server and a TypeScript client. Personal-use only — not
a product. See [PLAN.md](PLAN.md) for scope, stack, and the Option B
("sandbox first, rules grafted in incrementally from S13+") decision.

Three hard problems, ranked: **game state and multiplayer sync** (Go
server, authoritative state, WebSocket broadcast), **UX polish** (the
whole point — Commander-specific affordances nothing else has), and —
long horizon — **incremental rules enforcement** (B→C track, starting
in S13+ after the sandbox is shipped).

---

## 2. Ground rules

1. **Read [PLAN.md](PLAN.md) before making architectural suggestions.** The
   central decision (reuse XMage vs build from scratch) is already made. Do not
   relitigate it without new information.
2. **This is a hobby project at ~10 hours/week.** Optimise for momentum and
   clarity, not enterprise rigour. No microservices, no k8s, no premature
   abstractions. One VPS. One database. One deployable per service.
3. **Personal use only changes the calculus.** Legal risk is low (Cockatrice and
   XMage exist). Scale is ≤8 users. Don't build for hypothetical public launch.
4. **UX polish is the entire point.** "Make it feel good" is core product
   work and deserves real care. Sandbox features (manual resolution,
   manual priority) are a deliberate design choice, not a shortcut.
5. **We do not use XMage.** This was evaluated and rejected in S01 — see
   [PLAN.md §2.2](PLAN.md#22-why-not-option-a-the-s01-discovery). The entire
   backend is an original Go game server; the rules engine (eventually,
   S13+) is grown incrementally on top of it. No Java anywhere in the stack.

---

## 3. Repo layout (evolving)

```
cmd_and_ctrl/
├── PLAN.md              # vision, roadmap, open decisions
├── AGENTS.md            # this file
├── Makefile             # top-level dev/test/lint targets
├── .devcontainer/       # Go + Node dev environment
├── server/              # Go game server (authoritative state, WebSocket + HTTP API)
│   ├── cmd/
│   │   ├── server/      # main package — serves :8080 with lobby + hub + cards routes
│   │   ├── gamecli/     # dev WebSocket client for driving a game via v0 actions
│   │   └── snapshotscrub/ # strips player data from a restore point before it joins the snapshot corpus (#522)
│   ├── internal/
│   │   ├── game/        # authoritative domain: Game, Player, Zone, Card, Turn, mutations
│   │   │   └── testdata/snapshots/ # the restore-point fixture corpus: v<N>/ generated and frozen per schema version, real/ scrubbed from cmd-dev (#522)
│   │   ├── protocol/    # v0 wire format types + ViewOfGame + FilterViewFor
│   │   ├── actions/     # action type enum + Dispatch(Game, Action) router
│   │   ├── legal/       # legal-move enumerator — the closed move list a bot picks from and the client's timing lookup (ADR 0033 §1)
│   │   ├── ws/          # gorilla/websocket hub, Room, RoomManager, per-viewer broadcast
│   │   ├── aiseat/      # AI bot seats (S31): runner goroutine per bot, tiered policies, curated decks, announced improvisation — see docs/bot.md
│   │   ├── auth/        # pluggable Authenticator interface + MemoryAuthenticator + HTTP middleware
│   │   ├── lobby/       # GameMeta registry, invite flow, lobby HTTP handler, WSAuthorizer, deck upload
│   │   ├── cards/       # Scryfall index (streaming load) + disk-backed image cache + /cards routes
│   │   │   └── coverage/ # measures the live catalog; fails CI when the coverage docs or a card's Caveats stop being true
│   │   ├── catalog/     # /catalog routes (signed-in) — what the engine automates + how completely (ADR 0042)
│   │   ├── roadmap/     # the curated registry of keywords, mechanics and engine seams behind the public roadmap; generates docs/engine-seams.md's open table (ADR 0092) and its closed list from docs/engine-seams/closed/ (#1461)
│   │   ├── bugstore/    # bug-report artifacts: reporter screenshots (public, Camo-reachable) + pinned replays (admin-only)
│   │   ├── deck/        # decklist parsers (Moxfield, plain text) + Commander validation
│   │   ├── deckcoverage/ # a decklist's coverage report: every card bucketed manual / unreviewed / caveats / automated / no_effect (ADR 0095)
│   │   ├── deckrequests/ # which GitHub issue tracks each requested deck, and who asked when (ADR 0095, migration 0006)
│   │   ├── usersettings/ # a signed-in person's synced settings: one JSON object, 32 KiB cap, If-Match revision (ADR 0110 §4, migration 0008)
│   │   ├── tablesetups/  # the last setup each person started a table with: settings, bots, tablemates (ADR 0110 §5, migration 0008)
│   │   ├── snapshotscrub/ # the scrubber behind cmd/snapshotscrub: generic-JSON rewrite, refuses snowflakes and emails (#522)
│   │   └── db/          # persistent SQLite store (ADR 0051): open/WAL/migrate/backup (S34 sub-PR 1); users/games/decks land in later sub-PRs
│   ├── Makefile
│   └── .golangci.yml
├── client/              # TypeScript + Svelte 5 + Vite (PixiJS arrives in S05)
│   ├── src/
│   │   ├── App.svelte   # router shell
│   │   ├── main.ts
│   │   ├── app.css
│   │   ├── lib/         # protocol types, WebSocket client, session, api, hash router
│   │   └── routes/      # Login / Lobby / Join / Game views
│   ├── index.html
│   ├── vite.config.ts
│   ├── svelte.config.js
│   ├── tsconfig.json
│   ├── eslint.config.js
│   └── package.json
├── scripts/             # scryfall-refresh.sh (weekly cron), backup-offsite.sh (nightly R2 backup), set-server-env.sh (CD) + one-off tools
├── data/                # runtime state (gitignored): Scryfall cache, snapshots, images
└── docs/
    ├── protocol.md      # v0 wire format spec
    ├── lobby.md         # lobby HTTP API reference
    ├── bot.md           # AI bot seat — user-facing guide (S31)
    ├── engine-seams/closed/ # one fragment per closed seam; CI generates engine-seams.md's Closed list from them (#1461)
    ├── adding-cards.md  # the catalog card guide: recipes for cards, mechanics and engine seams (moved out of AGENTS.md §7, #1747)
    ├── sprints.md       # sprint plan
    └── decisions/       # ADRs (0001 WS library … 0108 turn-scoped effects, object history and damage shields; 0109 rule gates, land types, mana and cost components; 0110 remember me: durable sign-in, account settings, admins and saved setups; 0111 the action dock; 0112 signed-in home, player mode and one decks page; 0113 small seams for the S58 deck requests; 0114 the Ring tempts you; 0115 commanders die (CR 903.9a); 0116 a card filter on the revealed-hand pick; 0117 click to act and a per-colour mana stepper; 0118 strict payment by default, Cast anyway, and alternative costs; 0119 a stack you can follow; 0120 expand a player's board; 0121 animated dice: an opening roll at the table, dice you can watch, and a Roll a die action) — see §4 on numbering
```

When you create a new top-level directory, add it here.

---

## 4. Sprint and tracking discipline

### Branches and environments

`develop` is the integration branch; `main` is the release branch.
Feature branches PR into **`develop`**, which auto-deploys to
`https://cmd-dev.labxp.io`. Promotion to production is a `develop` →
`main` PR. `main` is still the repo's default branch, so **always pass
`--base develop` to `gh pr create`** — a feature PR into `main` fails
the required `promotion-guard` check. Full matrix and host runbook:
[docs/environments.md](docs/environments.md); rationale:
[ADR 0023](docs/decisions/0023-develop-environment.md).

Two rules that are easy to get wrong:

- **Dev-only features are gated server-side.** `CMDCTRL_ENV=dev`
  is the only switch that can enable one, and `appenv.LoadFeatures`
  refuses to enable anything outside a dev deployment. Wrap every
  dev-only route in `requireDev`. Hiding a control behind a
  client-side `if` is not gating it — spawning cards and acting as
  another seat are cheats on a live table.
- **Never deploy or restart the Discord bot from `develop`.** Two bot
  processes on the same application answer every slash command twice.
  CD enforces it; don't hand-install the bot unit on the dev box.

Work is organised into 2-week sprints tracked in:

- **Project board:** https://github.com/users/krakenhavoc/projects/5
- **Milestones:** https://github.com/krakenhavoc/cmd_and_ctrl/milestones
- **Sprint plan with task checklists:** [docs/sprints.md](docs/sprints.md)

Each sprint has one tracking issue (`#1` through `#12`) whose body is the
checklist of sub-tasks. Commits reference the sprint issue number.

**Every commit and pull request must reference the sprint and issue it relates
to.** This is non-negotiable — it's how we keep a part-time, multi-month project
coherent.

### Picking an ADR number

**Check every branch, not just the one you are on.** ADR files live in `docs/decisions/` and the
number is in the filename *and* the H1, so two branches that both grab "the next number" collide
silently. They do **not** conflict at merge time: the filenames differ, so git merges both cleanly
and neither author learns. The collision is invisible in the diff and invisible in the merge, and
only shows up when somebody lists the directory — by which point the number is in commit messages,
issue bodies and cross-links in other ADRs.

`TestADRNumbersAreUniqueAndMatchTheirHeading` (`server/internal/docsguard`) fails CI on a duplicate
number and on an H1 that disagrees with its filename. It exists because asking authors to check was
tried first and three collisions reached `main` anyway. If it fires, renumber the ADR with fewer
inbound links (`git grep` its filename), fix its H1, and update the ADR range line in §3.

```bash
git fetch --all --prune
for b in $(git branch -r --format='%(refname:short)' | grep -v HEAD); do
  git ls-tree --name-only "$b" docs/decisions/
done | sed 's|.*/||' | cut -d- -f1 | sort -u
```

Take the first number that does not appear, and say in the PR body which branches you checked.
Numbers are **not** reused when an ADR is renumbered or abandoned — `0005`, `0024`, `0029` and
`0030` are permanently unused for exactly that reason. If a number you already used turns out to be
taken, renumber **your** file (title, filename and every inbound link) rather than asking the other
branch to move; the one that merges first keeps the number.

### Commit message format

```
<type>(<scope>): <subject>

<body>

Sprint: S<NN> — <sprint name>
Issue: #<issue_number>
```

Example:

```
feat(server): add websocket ping/pong round-trip

Minimal gorilla/websocket hub accepting JSON frames on /ws. Replies to
v0 ping frames with pong carrying server_time. No game state yet.

Sprint: S01 — Go server + client scaffold
Issue: #1
```

`<type>` is one of: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `spike`.
`<scope>` is the top-level dir being touched (`server`, `client`, `docs`, etc.).

### Pull request format

PR descriptions must include:

```
## Summary
<1-3 bullets>

## Sprint
S<NN> — <sprint name>

## Issues
Closes #<n>, relates to #<n>

## Test plan
<bulleted checklist>
```

If a PR does not belong to the active sprint, say so explicitly and justify it.

Feature PRs target `develop`, so GitHub does not apply their closing keywords
directly (it only does that for PRs targeting the default branch). After a
`develop` → `main` promotion merges, `main-promotion-issue-close.yml` maps the
promoted squash commits back to their original PRs and honors line-leading
`Closes`, `Fixes`, and `Resolves` directives from the `## Issues` section. Keep
each closing directive explicit in that section; prose such as "does not close"
and references outside it are deliberately ignored.

---

## 5. Commands you'll actually run

*(Populate as the project takes shape. Empty sections are fine — don't invent.)*

### Dev environment

The devcontainer installs Go, Node, and the GitHub CLI. Ports 3000, 5173, and
8080 are forwarded. (The Java/Maven features remain installed for now but are
unused — they can be removed in a later cleanup PR.)

### Go without a local toolchain

The workstation has no local Go. `scripts/go-docker.sh` is THE way to run it: `scripts/go-docker.sh test ./internal/roadmap/...`, `vet`, `lint`, `go <args>`, `prune` (no arguments prints usage). It mounts the repo, uses `golang:1.22` (`golang:1.24` for golangci-lint), passes the exit status through (never pipe it through `tail`), and cleans the build cache first when it passes `CMDCTRL_GOCACHE_MAX_GB` (default 15). `make -C server docker-test docker-lint` wrap it.

**Never create any other cache volume**, and never hand-write a `docker run` for Go. The only three are `cmdctrl-gomod`, `cmdctrl-gobuild` and `cmdctrl-golangci`; per-PR volumes filled the Docker partition at 178 GB (#2035).

### Top-level
- `make help` — list targets
- `make server-dev` — run the Go server on :8080
- `make client-dev` — run the Vite dev server on :5173
- `make test` — server tests + client typecheck
- `make lint` — `go vet` + client ESLint + Prettier check

### Server (Go, `server/`)
- `make -C server dev` — run locally on :8080 (seeds a 4-player demo game)
- `make -C server test` — `go test -race -cover ./...`
- `make -C server vet` — `go vet ./...`
- `make -C server fmt` — `gofmt -s -w .`
- `make -C server build` — produces `server/bin/cmd_and_ctrl-server`
- `make -C server build-boteval` — produces `server/bin/boteval`, the AI-bot evaluation harness (#837). `boteval probe [--endpoint URL] [--model ID] [--think] [--max-tokens N]` sends ONE request in the funnel's exact shape to a model endpoint and prints prompt/completion tokens vs the client-side estimate, `finish_reason`, whether a `reasoning` field came back, the parsed index and copied label, and a verdict for each of the three failures that make a model seat play like a heuristic seat: truncation, thinking, and an answer that names no listed move (`INDEX: out of range`, #2196). Always exits 0. Not deployed; local tool.
- `boteval suite run [--dir DIR] [--policy heuristic|assisted|strong] [--max-think 20s] [--think] [--max-tokens N] [--note text] [--parallel N] [--out report.json] [--md]` — asks a policy every labelled position in `server/internal/aiseat/suite/testdata/positions/` and reports agreement overall and per tag, plus reject-hits, malformed replies, out-of-range indices and timeouts. Exits non-zero on a gated miss. The `heuristic` run is also an ordinary Go test and runs on every CI run.
- `boteval suite harvest --from 'dir/*.decisions.jsonl' --to inbox/ [--escalated] [--disagree] [--fallback a,b] [--layer A,B,C] [--seat 0,1] [--tag block,attack] [--limit N] [--seed N]` — pulls candidate windows out of decision logs into an inbox of UNLABELLED positions. Deterministic under `--seed`.
- `boteval suite render --pos path/to/position.json [--deck ID]` — prints the exact prompt a model would see for one position and the move list with `<- accept / reject / heuristic / model@capture` markers. This is the labelling screen. See [docs/bot.md](docs/bot.md#position-suite).
- `boteval arena --seats a,b,c,d [--decks …] --games N --rotate --out DIR [--think] [--max-tokens N]` — headless bot-vs-bot games with the report block ADR 0052 asks every bot PR to carry: win rate with a **Wilson 95% interval** against the table's null rate (1/seats), the funnel's layer/escalation/timeout counters, and decision + model-call latency tails. Rotation seats contestant `k` at position `(k+i)%n` in game `i`, so turn order cancels. A model tier with **no endpoint is refused, not downgraded** (an `assisted` seat with no client plays the heuristic under a model tier's name). A stall is **reported, not fatal**. Artifacts land in `<out>/<RFC3339 start>/`: `summary.md`, `summary.json`, `games.jsonl` (streamed per game), `decisions/`, `replays/`. Wall clock: ~0.2 s per two-seat heuristic game, ~3.5 s per four-seat curated-deck game, 5–15 min per game with one local-model seat. See [docs/bot.md](docs/bot.md#arena).
- `cd server && go run ./cmd/gamecli -addr ws://localhost:8080/ws` — drive the demo game from a terminal; reads action JSON on stdin or via `-script path.json`
- Endpoints: `GET /healthz`, `GET /ws` (protocol v0, see [docs/protocol.md](docs/protocol.md)), `POST /admin/login`, `/games*` lobby routes (see [docs/lobby.md](docs/lobby.md)) — including `POST /games` (any signed-in person or the admin token; a person is the table's creator and is held to 3 open tables and 1 creation per 30 s, the token to neither; `"setup": "last"` applies their last setup — [ADR 0110](docs/decisions/0110-remember-me.md) §5), `POST /games/{id}/setup` (the caller's last setup on an unstarted table: host, creator or admin; skipped bots are named, never downgraded), and `POST /games/practice` + `POST /games/{id}/practice/leave`, the tutorial's practice table (any session; unlisted, never persisted, reaped when idle; [ADR 0076](docs/decisions/0076-tutorial.md)) — `GET /me/setup` + `GET /me/last-deck` (the setup captured when a table you created starts, and the deck you last seated, for the create form and the deck panel to pre-fill; ADR 0110 §5), `GET /me/games` + `POST /me/games/{id}/session` (a signed-in user's games and seat reclaim, ADR 0051 sub-PR 4), `PUT /me/admin-mode` (an allowlisted person switches admin mode on for 12 hours, or off; everyone else 403; `GET /me` adds `admin_allowed`, `admin_mode` and `admin_mode_ends_at` — [ADR 0112](docs/decisions/0112-signed-in-home-player-mode-and-one-decks-page.md) §2), `POST /me/session` (a signed-in session's cookie reinstall and renewal on use, ADR 0110 §1 items 5 and 6), `GET /me/decks` (+ `/{id}/coverage`, `PATCH`, `DELETE`: the saved deck library with coverage computed on read, ADR 0110 §6), `POST /me/decks` (a signed-in person saves a checked link or pasted list to that library: a link the check's cache holds fetches nothing, one that must be fetched spends a `POST /deck-coverage` token, a name already used replaces that deck (`replaced`), and a new deck past 200 is a 409; [ADR 0112](docs/decisions/0112-signed-in-home-player-mode-and-one-decks-page.md) §3), `GET /me/tablemates` + `POST /games/{id}/invites/dm` (the people you have played with, and DMing one of them this table's existing invite link — ADR 0051 decisions 8 and 5, S34 sub-PR 6), `GET` + `PUT /me/settings` (a signed-in user's per-person settings on their account, `If-Match` revision, 32 KiB, 1 write/s per person; 403 for everyone else, who keep browser-only settings — [ADR 0110](docs/decisions/0110-remember-me.md) §4), `GET /auth/discord/link` (link Discord to a held seat), `/cards/*` image + metadata routes, `GET /catalog` + `GET /catalog/image/{id}` (signed-in session required — the card catalogue, [ADR 0042](docs/decisions/0042-card-catalog-page.md); `main.go` says why it is not public), `GET /roadmap` (public, no session — the engine roadmap: card names and caveats only, never art or oracle text, [ADR 0092](docs/decisions/0092-public-roadmap-and-site-portal.md)), `POST /deck-coverage` (public, no session, per-IP limited — a Moxfield/Archidekt link or pasted list bucketed into `manual` / `unreviewed` / `caveats` / `automated` / `no_effect`; names, oracle IDs and caveats only) and `POST /deck-requests` (a Discord-signed-in user, or the bot's admin session naming the member — files or joins a GitHub issue listing the deck's missing cards, 3 asks per person per 24 h, none for an admin (allowlisted and in admin mode; through the bot, only when the named member's own account is one), #2052; a person may name one of their saved decks as `deck_id` instead of a link or a list, ADR 0112 §3 item 5; [ADR 0095](docs/decisions/0095-deck-coverage-and-deck-requests.md), [docs/lobby.md](docs/lobby.md))
- Env vars:
  - `CMDCTRL_ADDR` — listen addr (default `:8080`)
  - `CMDCTRL_DATA_DIR` — data root (default `./data`; empty string disables disk writes + card cache). Holds `db/cmdctrl.sqlite` (ADR 0051, S34 sub-PR 1 — the persistent user/game/deck store, `internal/db`) and its `db/cmdctrl.backup.sqlite` VACUUM INTO copy, alongside the existing `scryfall/`, `images/`, `avatars/`, `bugreports/`, `restore/`, `replays/` and `games/`. Since S34 sub-PR 3 the lobby's games, seats and invites are rows in that database, and invites are stored as hashes. `lobby/` holds only the `<id>.json.imported` files the one-time importer renamed and left for a rollback (docs/environments.md).
  - `CMDCTRL_DB_BACKUP_INTERVAL` — Go duration between the database's in-process `VACUUM INTO` backup sweeps (default `1h`; `<= 0` disables the sweep). The sweep writes `db/cmdctrl.backup.sqlite` beside the live file so a disk-level backup picks up a consistent copy. This is the same-disk copy only. The nightly off-node copy is `scripts/backup-offsite.sh` (restic to Cloudflare R2, #1031, ADR 0051 decision 1 as amended), run by `cmd-and-ctrl-backup.timer` on both hosts, not by this server. Runbook: [docs/environments.md](docs/environments.md#backups).
  - `CMDCTRL_ADMIN_TOKEN` — **required**. Shared admin secret for `POST /admin/login`. At least 16 characters.
  - `CMDCTRL_DISCORD_ADMIN_USER_IDS` — the admin allowlist ([ADR 0110](docs/decisions/0110-remember-me.md) §3, owner answer 3): comma-separated Discord user snowflakes, the same variable the Discord bot reads for `/c2-end`, so one list covers both. Being on the list makes a person *able* to switch admin mode on, and it starts off ([ADR 0112](docs/decisions/0112-signed-in-home-player-mode-and-one-decks-page.md) §2): each listed person is in **player mode**, where every admin route and gate answers exactly as for a non-admin, until they `PUT /me/admin-mode` `{"on": true}`; admin mode lapses after 12 hours (a once-a-minute sweeper closes their admin sockets with 4001), and logout-everywhere or an admin revoke ends it. A failed boot load of the modes logs an ERROR and leaves everyone in player mode. In admin mode, a signed-in session (identity, seat or spectator) whose Discord ID is on the list is an admin exactly like the shared token: every admin route, every host-or-admin gate, and full WebSocket parity — any game, optionally as any seat, each binding logged at Info with the user ID (owner answer 4). An allowlisted person still joins a table as their own Discord identity and plays as their own seat; the admin context menu's card overrides work from that seat. Admin is computed per request (`lobby.Config.isAdmin`) and never stored in a token; `GET /me` reports it as `admin`. It **requires a user**: a reclaim ticket carrying a listed seat's Discord ID is not admin, and on a server with no database the list grants nothing (the boot log says so). Three bot-only paths stay on the shared token alone (`isServerCredential`): the bot's rate buckets, filing a deck request for a named member, and DMing a raw snowflake. Empty or unset means only the token is admin. A malformed entry fails the boot; the log carries the count, never the IDs. `TestRoleAdminIsComparedOnlyInTheAdminPredicates` fails on any new bare `auth.RoleAdmin` comparison, `TestAllowlistAndModeAreAskedOnlyInAdminsGo` on any `AdminList.Has` or `AdminModes.On` outside `admins.go`, and `TestEveryAdminCallSiteIsInTheSameAnswerTable` on a new admin call site that `TestPlayerModeAnswersExactlyAsANonAdmin` does not ask. **Provisioned by CI/CD on both hosts**, from the Actions variable of the same name (an environment-level value in `dev` or `prod` overrides a repo-level one), by "Sync server env (admin allowlist)" — **always written, empty included**, so clearing the variable removes every allowlisted admin on the next deploy. The step refuses a malformed value before writing it, leaving the running server as it was.
  - `CMDCTRL_SESSION_TTL` — lifetime of a session with no user, as a Go duration (default `12h`): guest seats, guest spectators, admin-token sessions and reclaim-ticket sessions. Those cannot be revoked, so they stay short. Every session with a user follows `CMDCTRL_IDENTITY_TTL` instead ([ADR 0110](docs/decisions/0110-remember-me.md) §1).
  - `CMDCTRL_IDENTITY_TTL` — lifetime of every session that carries a user, as a Go duration (default `720h`, 30 days; S34 sub-PR 7, [ADR 0051](docs/decisions/0051-user-database.md) decision 3, widened by [ADR 0110](docs/decisions/0110-remember-me.md) §1 in S55). A Discord sign-in (the login page, the invite link's callback, linking Discord to a seat) gets the full lifetime. A session minted **from** a signed-in session (joining by link or code, spectating, `POST /me/games/{id}/session`, the practice table, a reclaim redeemed by the seat's own user) inherits that session's expiry to the millisecond, so joining a table never extends a sign-in (`issueFor`, `server/internal/lobby/session_ttl.go`). It is long because it is the credential a browser keeps across games. That is safe because every such session carries a `UserID` and can be revoked (`POST /logout/everywhere`, `POST /admin/users/{id}/revoke-sessions`, decision 6). **Renewal on use** (`POST /me/session`, ADR 0110 §1 item 5, owner answer 1): a signed-in session more than half spent (from its `issued_at` to its `expires_at`) is re-issued for a fresh `CMDCTRL_IDENTITY_TTL` with the same principal. The client calls it at page load and at each session's half-life, so someone who plays at least once per half-lifetime is never asked to sign in again. Revocation is how a renewed session ends. The same route reinstalls the signed-in session the client keeps aside in `localStorage["cmdctrl.identity"]` while it holds an admin-token or ticket session (§1 item 6). See [docs/lobby.md](docs/lobby.md#post-mesession-adr-0110-1-items-5-and-6-s55). Invalid or `<= 0`: the boot fails. Not provisioned by CD. The client's expiry timer re-arms past `setTimeout`'s ~24.8-day ceiling, so a 30-day session is not dropped early.
  - `CMDCTRL_SESSION_KEY` — HMAC-SHA256 key that signs session tokens (`auth.HMACAuthenticator`, #517, [ADR 0044](docs/decisions/0044-surviving-a-deploy.md) decision 3). Set: a token minted before a restart validates after it. **Unset: sessions fall back to the in-memory store, with a warning naming the variable on every boot, and every deploy logs everyone out.** Set but under 32 bytes, or equal to `CMDCTRL_ADMIN_TOKEN`: the boot fails. There is no default key. `Revoke` is advisory under HMAC, so `POST /logout` clears the cookie but does not kill a copied token before its expiry. A session with a `UserID` (a Discord sign-in, or a seat claimed from one) can be killed: `POST /logout/everywhere` or the admin's `POST /admin/users/{id}/revoke-sessions` moves `users.sessions_invalid_before`, and `auth.WithRevocation` refuses older tokens from an in-memory cache (ADR 0051 decision 6). Admin, guest and spectator sessions keep the advisory revoke. **Provisioned by CI/CD**: generated on each host the first time it is missing and never rewritten, since a new key logs every player out. See [docs/environments.md](docs/environments.md#secrets).
  - `CMDCTRL_IDENTITY_KEY` — AES-256-GCM key that encrypts Discord OAuth refresh tokens on the `identities` rows (`users.Sealer`, S34 sub-PR 2, [ADR 0051](docs/decisions/0051-user-database.md) decision 5). Same format as `CMDCTRL_SESSION_KEY`: a random string of at least 32 bytes, which the server hashes to the AES key. **Unset: Discord sign-in still works and still records the user, but the refresh token is discarded and `refresh_token` is NULL, with a warning naming the variable on every boot. It is never stored in the clear.** Set but under 32 bytes, or equal to `CMDCTRL_ADMIN_TOKEN` or `CMDCTRL_SESSION_KEY`: the boot fails. Rotating it makes stored refresh tokens unreadable (nothing reads them yet) and logs nobody out. **Provisioned by CI/CD** exactly like the session key: generated on each host the first time it is missing, never rewritten.
  - `CMDCTRL_DISCORD_BOT_TOKEN` — the Discord **bot** token, read by the SERVER for ADR 0051 decision 5's direct-message invites (`POST /games/{id}/invites/dm`, S34 sub-PR 6). Same secret the gateway bot holds; two units, two env files, one Actions secret. **Unset: the DM route answers 503 naming this variable and every other route is unchanged — it never fails open — and the boot log says which state it is in, once.** The route also needs an origin for the link it sends (`CMDCTRL_PUBLIC_BASE_URL`, falling back to `CMDCTRL_CLIENT_BASE_URL`); missing that is the same 503. **Provisioned by CI/CD on production only** ("Sync server env (Discord DM invites)"): one Discord application, and a preview box DMing real people from the same identity is the double-send the bot's prod-only rule exists to prevent.
  - `CMDCTRL_ALLOWED_ORIGINS` — comma-separated hostnames (or full URLs) permitted as cross-origin WebSocket callers. Same-origin is always allowed; unset = same-origin only.
  - `CMDCTRL_SEED_DEMO=1` — seed the S03 4-player demo game at startup for the gamecli dev loop
  - `CMDCTRL_DISCORD_CLIENT_ID` / `CMDCTRL_DISCORD_CLIENT_SECRET` / `CMDCTRL_DISCORD_REDIRECT_URI` — S12.5 OAuth credentials. Unset disables the Discord sign-in button (manual name entry still works).
  - `CMDCTRL_SECURE_COOKIES` — truthy sets the `Secure` attribute on the session cookie. Enable in any TLS deployment; leave unset for plain-HTTP local dev (a `Secure` cookie is never sent over http and would silently break login).
  - `CMDCTRL_TRUST_FORWARDED` — truthy keys the lobby rate limiter off the leftmost `X-Forwarded-For` hop instead of the socket `RemoteAddr`. Enable **only** when the server sits behind a trusted reverse proxy that sets the header; otherwise clients can spoof it to dodge limits.
  - `CMDCTRL_DEV_RELAX_RATE_LIMITS` — truthy effectively disables the lobby rate limiters. For the e2e suite and local load-y dev loops only; logs a loud warning on boot. **Never set in production.**
  - `CMDCTRL_GITHUB_TOKEN` — enables the in-app "report a bug" button (`POST /bugreport` files a GitHub issue) and ADR 0095's deck requests (`POST /deck-requests` files or comments on a deck-request issue; it also needs `CMDCTRL_DATA_DIR`, and answers 503 naming whichever is missing). Use a fine-grained PAT with **Issues: write** on the one repo, nothing broader. Unset disables both and the client hides the bug button. See [ADR 0017](docs/decisions/0017-bug-report-button.md). **Reports are published to everyone who can read the repo, so credentials are redacted before they get there** — client (`client/src/lib/redact.ts`, applied to the protocol log, console capture and the submitted draft) and server (`server/internal/util/redact`, applied in `renderBugIssueBody`); never log a token-bearing URL raw, pass it through `redactURL` (#721, [ADR 0017 §9](docs/decisions/0017-bug-report-button.md); durable sessions #517 depend on it). **Provisioned by CI/CD**: the deploy job upserts it into `/etc/cmd_and_ctrl/env` (the env file the HomeLab cloud-init template writes and the systemd units read) from the `CMDCTRL_GITHUB_TOKEN` Actions secret, via `sudo scripts/set-server-env.sh`; to rotate, update the secret and rerun the deploy — no host access needed.
  - `CMDCTRL_GITHUB_REPO` — `owner/name` slug issues are filed against (default `krakenhavoc/cmd_and_ctrl`).
  - `CMDCTRL_PUBLIC_BASE_URL` — the origin this server is reachable at from the public internet (e.g. `https://cmd.labxp.io`). Needed for bug-report **screenshots**: GitHub renders an issue image by fetching it through its Camo proxy, so the URL in the issue body has to be absolute and publicly resolvable. Falls back to `CMDCTRL_CLIENT_BASE_URL`; with neither set (or no `CMDCTRL_DATA_DIR`) attachments are disabled, `/bugreport/config` reports `attachments:false`, the modal hides its file picker, and text reports keep working. **No default on purpose** — a wrong origin produces issues full of broken images, which is worse than a deploy that doesn't offer upload. Provisioned by CI/CD from the `CMDCTRL_PUBLIC_BASE_URL` repo variable. See [ADR 0017 §6](docs/decisions/0017-bug-report-button.md).
  - `CMDCTRL_OPENAI_ENDPOINT` — an OpenAI-compatible `/v1/chat/completions` server for the model-backed bot tiers (`assisted`, `strong`): Ollama, LM Studio, llama.cpp's server, vLLM. Takes a URL (e.g. `http://192.168.1.18:11434`), or `1` for a stock Ollama on this machine. If both this and an Anthropic key are set, this one wins. Unset by default.
  - `CMDCTRL_OPENAI_API_KEY` — optional bearer key for that endpoint; most local servers need none.
  - `CMDCTRL_OPENAI_SEND_THINK` — `0` stops the client sending Ollama's `think: false`. Set it only for a server that rejects the field. A hybrid-thinking model (qwen3, …) left on its default spends the whole deadline thinking and answers nothing.
  - `CMDCTRL_ANTHROPIC_API_KEY` — the hosted alternative for the model-backed bot tiers. Falls back to `ANTHROPIC_API_KEY` when unset. **No model endpoint at all is a supported configuration, not a broken one.** `random` and `heuristic` work as normal. Since [#514](https://github.com/krakenhavoc/cmd_and_ctrl/pull/514), `assisted` and `strong` report `available:false` in `GET /bot/options`, with a reason the picker shows, and a request to seat one is a **422**. The server never silently downgrades them. A seat already running when its model goes away keeps playing on the rules filter plus the heuristic. See the bot-seat section below and [docs/bot.md](docs/bot.md#the-model-endpoint).
  - `CMDCTRL_ANTHROPIC_ENDPOINT` — overrides the Messages API URL (default `https://api.anthropic.com/v1/messages`). For pointing a dev stack at a stub; there is no reason to set it in production.
  - `CMDCTRL_BOT_MODEL` — model id for routine bot windows. **Required with `CMDCTRL_OPENAI_ENDPOINT`**: use the name the server serves, e.g. what you `ollama pull`ed. Without it the Anthropic default ids are sent and the endpoint answers 404. Optional for Anthropic, which defaults to `claude-haiku-4-5`.
  - `CMDCTRL_BOT_FRONTIER_MODEL` — model id for escalated windows. Defaults to `CMDCTRL_BOT_MODEL` when that is set, otherwise to the Anthropic default, `claude-opus-5`. One model in both slots is supported, and it means `strong` and `assisted` ask the same model.
  - `CMDCTRL_BOT_MAX_THINK` — Go duration; the model tiers' hard think deadline. It only ever widens a tier's own deadline (2s `assisted`, 5s `strong`). It defaults to `20s` when `CMDCTRL_OPENAI_ENDPOINT` is set, and to the tier deadlines otherwise. An unparseable or non-positive value fails the boot.
  - `CMDCTRL_BOT_THINK` — `1` lets the model tiers think before each decision (an experiment, #2196; off by default). It also raises the reply budget to 8000 tokens and, unless `CMDCTRL_BOT_MAX_THINK` is set, the deadline to 120s. See [docs/bot.md](docs/bot.md#letting-the-model-think-2196).
  - `CMDCTRL_BOT_MAX_TOKENS` — the model tiers' reply budget per decision, in tokens. Unset keeps 128 (routine) / 256 (escalated), or 8000 with `CMDCTRL_BOT_THINK` on. A malformed value fails the boot.
  - `CMDCTRL_BOT_DECISION_LOG` — directory for the per-game bot decision log (one JSONL line per decision window per bot seat: prompt, reply, heuristic ranking, fallback cause, latency). Empty (the default) is **off**. **Operator-only**: each record is the seat's own filtered view, but the file aggregates every bot seat at the table, so it is never served over HTTP and never attached to a bug report. Directory `0700`, files `0600`, 256 MiB per game. See [docs/bot.md](docs/bot.md#decision-log).
  - `CMDCTRL_BOT_DECISION_LOG_MODE` — `escalated` (default: full board view only for windows that left Layer A) | `all` | `model` (only the windows a model answered). An unrecognised value fails the boot **when the log is on**, like `CMDCTRL_BOT_MAX_THINK`; with the log off it is a warning, because refusing to start over a variable that changes nothing is a server that does not come back after a rollback.
- Cron: `scripts/scryfall-refresh.sh` — weekly refresh of the Scryfall default-cards dump (suggested cron: `0 5 * * 0`)
- Off-site backup: `scripts/backup-offsite.sh`, run nightly by `deploy/cmd-and-ctrl-backup.service` + `.timer` (as `cmdctrl`, data dir read-only) on both hosts. It uses restic to back up the data dir to a per-host Cloudflare R2 bucket, taking `db/cmdctrl.backup.sqlite` but never the live db, and skipping the `scryfall/`, `images/` and `avatars/` caches. Credentials are in `/etc/cmd_and_ctrl/backup.env` (`root:cmdctrl 0640`, separate from the server's env), written by the CD step "Ensure off-site backup" from the `CMDCTRL_R2_*` and `CMDCTRL_RESTIC_PASSWORD` values in the `prod` / `dev` GitHub environments. A missing value, or one still containing `REPLACE_ME`, is a `::warning::` and a disabled timer, never a failed deploy. **The restic password must never change once a repository exists**; its recovery copy is the owner's password manager. Runbook, including the restore: [docs/environments.md](docs/environments.md#backups) (#1031).

### Snapshot compatibility (#522)

A restore point written by yesterday's binary has to restore in today's. Three tests hold that line, and the rule they enforce is written next to `SnapshotSchemaVersion` in `server/internal/game/snapshot.go`: changes within a version are additive only, anything else is a bump, and fixtures are never edited.

- **The shape guard.** `TestSnapshotShapeIsRecorded` (`internal/game`) compares the snapshot structs' JSON shape with `internal/game/testdata/snapshot_shape/v<N>.txt`. If you add a snapshot field, record it in the same change: `cd server && go test ./internal/game -run TestSnapshotShapeIsRecorded -args -update-shape`. The tool refuses a non-additive change (a key renamed, removed or retyped) under the current number. Bump `SnapshotSchemaVersion`, say why in its comment, and rerun it. It writes `v<N+1>.txt` and leaves the old file frozen.
- **The fixture corpus.** `TestSnapshotCorpusRestores` (`internal/cards/effects`, because restoring tokens, Clones and Equipment needs the real catalog) restores every file under `internal/game/testdata/snapshots/` with `RestoreStrict`. It fails if any key or value in a fixture is missing from a fresh capture of the restored game, if a card comes back with fewer catalog abilities than the file recorded, or if the game will not round-trip or take an action.
  - `v<N>/` is the generated set for schema N. It is written once, when N is introduced: `cd server && go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus`. The writer never touches an existing **file** (ADR 0041 phase 3, owner decision 6): a board with no file yet is written beside the others, so a shape that becomes a restore point after its version was introduced can still be frozen without a bump. It fails unless every existing fixture is a subset of a fresh render of its board (the restore guard's own comparison, so keys added later in the version are fine and a key lost, retyped or changed is not), and tells you to bump instead (#1801). `TestSnapshotCorpusBoardsStillMatchTheirFixtures` runs that check on every CI run, and also fails on a registered board with no fixture yet. CI fails if the current version has no directory.
  - `real/` holds scrubbed restore points from cmd-dev. Add one with `cd server && go run ./cmd/snapshotscrub -in <restore/id.json> -out internal/game/testdata/snapshots/real/<name>.json`. The tool replaces names and player IDs, drops Discord IDs and avatar hashes, and refuses to write if a snowflake, an email or an original player name is left anywhere. It never overwrites a file.
  - **Never edit, regenerate or delete a fixture to make the test pass.** If a bump migrates an old shape, list the migrated paths in `corpusMigrations` in `snapshot_corpus_test.go`, with the reason. Nothing else may excuse a difference.
- **The ability check.** A card whose catalog entry lost abilities between the writing and the reading binary is restored anyway. It is flagged `Card.AbilitiesLostOnRestore`, which shows it as `manual` for the rest of the game, and the boot log has an ERROR line naming the game, the card and the counts. More abilities than captured is not a mismatch.
- **Effect keys (#1497, [ADR 0041](docs/decisions/0041-game-persistence.md) phase 3).** A `ScopedEffect` mod kind is an on-disk identity, like a token slug: never renamed, never reused. A restore point naming a kind (or a delayed-trigger / stack-item effect key) this binary cannot interpret is refused with `ErrUnknownEffectKey` and the file is KEPT — only a newer build can have written it, so it is the rollback case. That refusal is what lets the vocabulary grow within a schema version.
- **Abilities on the stack (#1497, ADR 0041 P9, tier 4).** An ability's stack item names the catalog row it came from: `Body: "catalog/activated"` or `"catalog/triggered"`, plus `Params.Ability`, an `AbilityRef` of `{key, slot, ref, name}`. The `ref` is ADR 0093's `own:<i>` or `grant:<bundle>:<i>:<n>`, and `name` is the row's label (an activated row's `Label`, a triggered row's `Key`). `ActivateCatalogAbility` stamps every catalog activated ability. The trigger harvest stamps a triggered row that DECLARES its `Effect` (slice 4-2): the registry gives each row its identity as it files the definition (`game.IdentifyCatalogRows`), so no card file names its own row. A table with either waiting — queued or on the stack — is a restore point. Restore rebuilds the item's effect, target clause and mode clause from that row, in the running binary; a `TargetsFrom` clause is built again from the item's carried `Trigger` and its source. Three outcomes, set by the owner's answers of 2026-09-25:
  - The row at the ref has the same label: restored.
  - It does not, and exactly one row in the same list has that label (a deploy reordered the card's abilities): that row is used and the ref is rewritten.
  - No row, or more than one, has the label: the game is restored anyway. The item stays on the stack as a manual item with no effect, like a sandbox-announced one. Its source card is flagged `Card.AbilitiesLostOnRestore`, and the boot log has one ERROR line naming the game, the card, the label and the ref (`GameSnapshot.LostStackAbilities`). Never abandon the game and never drop the item.

  A binary from before slice 4-2 does not register `catalog/triggered`, so it refuses a file that names it (the rollback case). An ability carried on a card instance (`Card.ActivatedAbilities`) is never stamped and is still counted by the census. Neither is a trigger whose row computes its effect in a hand-written `Build` with no `Effect` beside it, nor one whose `TargetsFrom` reads the board (`TriggeredAbility.TargetsFromReadsBoard`): such rows would be listed in `server/internal/cards/effects/testdata/legacy_trigger_builds.txt`, which is EMPTY since tier 4-final and which `TestLegacyTriggerBuildsOnlyShrink` holds to shrinking — it fails on a listed row that no longer needs listing (delete the line, or run `cd server && go test ./internal/cards/effects -run TestLegacyTriggerBuildsOnlyShrink -args -update-legacy-triggers`, which never adds one) and on a new hand-written `Build` that is not listed (declare the `Effect` instead). A `Build` that makes a keyed item (`game.NewKeyedTriggeredItem` over a tier-2 body, as suspend and madness do) is data already and is not listed. Since this change, every stack-item field (on the stack, queued, or in `lastKnownStack`) is refused if this binary does not know it, for a file of the current schema. So a new stack-item field is refused by every binary from here on and needs no bump. `lastKnownStack`, the CR 608.2h record of spells countered this turn, is carried too. The census counts a stack item once, in `IntrinsicAbilityCards`, when its effect is a closure with no `Body` or when it has a target or mode clause with neither an oracle ID nor an ability ref behind it. In production the only such item is an ability carried on a card instance, which is that counter's own subject; a test that hand-builds an item (`newTriggeredItemForTest` in `internal/game`) lands there too. `StackTargetSpecs` and `StackEffects` are both retired (tier 4-final, owner decision 2026-09-25): `game.NewTriggeredItem(source, label)` takes no effect, because what a stack item does is always a catalog row or a registered body.
- **The closure ratchet.** `TestClosureFieldsReachableFromGame` (`internal/game`) lists every ROUTE from `Game` to something that can hold a func or an interface — every field of every reachable struct whose type reaches one, directly or through another struct — in `internal/game/testdata/closure_fields.txt`, each classified `rebuilt`, `keyed`, `transient`, `test-only` or `census:<Counter>`. A new route fails until it is classified, including a new field whose type is a struct already on the list (#1558): `cd server && go test ./internal/game -run TestClosureFieldsReachableFromGame -args -update-closure-fields` keeps the existing classes and marks new routes `unclassified`. The number of lines in each `census:` class and in `transient` is pinned by `closureClassCeilings` in the test and may only fall: phase 3's tiers delete lines and lower the ceiling; nothing raises one. The one exception is ADR 0041 P11: when a tier retires a counter, its lines may move to the class of the route they still have, in the same PR, and the PR lists those lines and the new ceiling.

### AI bot seat (Go, `server/internal/aiseat/`)

- Runs **in process** with the game server — no separate binary, no socket. One goroutine per bot seat, started by `Lobby.Start` and by the lobby's restore path. User-facing guide: [docs/bot.md](docs/bot.md); architecture: [ADR 0033](docs/decisions/0033-ai-bot-seat.md).
- Endpoints: `GET /bot/options`, `POST /games/{id}/seats/bot`, `DELETE /games/{id}/seats/bot/{player_id}` — see [docs/lobby.md](docs/lobby.md).
- **`aiseat.Start` subscribes before it returns (#938).** The room subscription is taken on the CALLER's goroutine, so "Start returned" means "this seat is listening from now on" — the only ordering a caller can establish, since nobody schedules the runner goroutine. Taking it inside the loop left a window in which a commit reached every other seat and not this one; the first step still reads the live game, so no state was lost, but a seat that DECLINES a window parks until the next commit, and a commit that landed in the window is a wake that is never coming. Anything new that starts a runner keeps that property: do not move the `room.Subscribe()` back onto the runner's goroutine.
- **A wait in these tests must poll a MONOTONE predicate (#634).** A bot table moves on a goroutine the test does not schedule — one turn cycle of the two-seat runner fixture is ~95ms — so any condition the table merely passes through can be missed outright by a test goroutine that loses the processor for a cycle, and then either never comes true again (a 30s `waitFor` timeout) or comes true one turn late against the wrong board. `n == 1 land`, `ActiveSeat == 1` and any `== k` on a counter that keeps climbing are all that shape. Wait for a thing that has HAPPENED and stays happened — a `>=` on a monotone counter, a latch, or a tally over the runner's own decision log (`aiseat.Config.Observer`), which is a complete history rather than a sample. `waitFor` / `waitForRunner` / `waitForChan` in `runner_test.go` are the only wait primitives; the budget is a backstop, never an assertion (#848). `waitForChan` is for the handshakes that are a channel rather than a predicate — a policy reporting that it was asked, a blocking call reporting that it returned — and it exists so that a `select` on a channel against a `time.After` never has to be written here again (#1048). `internal/aiseat` now has no `select` on a channel against a `time.After` at all; what is left is two deliberate short sleeps inside negative assertions (`runner_test.go`'s 300ms block grace, `manager_decisionlog_test.go`'s 50ms after `Shutdown`), which a slow machine makes MORE likely to pass (#876).
- **The bot reads one `Spec`-level signal, and only one (#686).** `protocol.CardView.Unimplemented` — `game.Unimplemented(c)`, i.e. `c.NeedsEffect && !IsAutoCard(CatalogKey(c))`, the same bit behind the deck-upload summary and the stack overlay's `manual` chip — is what tells an `assisted` / `strong` seat that a spell it just cast resolved into silence and is a candidate for ADR 0033 §8 improvisation (`aiseat/model/improvise.go`). Two consequences for anyone changing coverage: a card that stops being flagged stops being improvised, silently; and the bit reaches the bot through the WIRE VIEW, not through the catalog, so `aiseat/` still imports nothing from `internal/game` (`heuristic/imports_test.go` enforces it). Nothing else in `aiseat/` reads a catalog signal at all — the whole point of ADR 0033 §3 is that a policy sees a filtered view and a move list.
- Env vars: the bot seat reads only the model-transport variables listed above at runtime: `CMDCTRL_OPENAI_ENDPOINT` / `_API_KEY` / `_SEND_THINK`, `CMDCTRL_ANTHROPIC_API_KEY` / `_ENDPOINT` (and `ANTHROPIC_API_KEY`), and `CMDCTRL_BOT_MODEL` / `_FRONTIER_MODEL` / `_MAX_THINK` / `_THINK` / `_MAX_TOKENS`, plus `CMDCTRL_BOT_IMPROVISE` (`0` turns §8 improvisation off; on by default for the model tiers) and the off-by-default `CMDCTRL_BOT_DECISION_LOG` / `_MODE`. They are parsed in `cmd/server/main.go` and `aiseat/model`. Everything else is per-seat request data or a compile-time default. Tier and deck come with the request.
- **Pacing is per-decision, not per-seat (ADR 0075 §2.2, sub-PR 6).** `aiseat.Config`'s base pacing is still `MinThink` 700ms, `MaxThink` 2s (5s for `strong`, which `CMDCTRL_BOT_MAX_THINK` can widen for the model tiers) — but that is only the floor now. When `Config.FollowTablePace` is set (true for every production Config: `DefaultConfig`, `ConfigFor`, every tier's `RunnerConfig`), the runner re-reads `g.Settings.BotPace` — a `game.TableSettings` field a host can change mid-game — through `TableSettingsSnapshot()` before EVERY decision and remaps `MinThink`/`MaxThink` to that preset (`fast` 0/2s, `normal` 700ms/2s, `slow` 2s/8s), so a pace change is live on a bot's very next move with no seat restart. `MaxThink` is always the larger of the preset and the Config's own value, so `strong`'s longer deadline is never shortened by a `fast` table. `FollowTablePace` defaults to **off**, which is what keeps a hand-built `Config` — the zero value tests and `NewManagerWithConfig` use to strip `MinThink` to 0 for a fast whole-game test — in charge of its own numbers regardless of what `BotPace` the game carries.
- **The arena lives OUTSIDE `aiseat/`**, at `server/internal/botarena/`, and that is not a style choice: `heuristic/imports_test.go` bans `internal/game` from every subpackage of `aiseat/` — including their `_test.go` files, over Imports, TestImports *and* XTestImports — because a **policy** holding authoritative state could read an opponent's hand. An arena has to hold the `*game.Game` and the `*ws.Room`, so it sits above the ban and hands each policy nothing but the filtered `aiseat.Input` a runner would. `botarena.BattleDeck` is the whole-game tests' deck moved here verbatim; the copy in `heuristic_game_test.go` stays where it is, because those tests may not import this package.
- **The whole-game tests are gated off by default** and the package owns the longest tests in the tree (CI runs `go test` with a 30m timeout because of them):
  - `AISEAT_GAME_TESTS=1` — the master gate. Without it every whole-game test in `internal/aiseat` skips. The nightly `bot-games` job runs `go test ./internal/aiseat/... -race -timeout 30m -skip 'TestFourRandomBotsPlayToAWinner|TestFourHeuristicBotsPlayToAWinner'`, then the random-table step. The catalog soak is its own job, `catalog-soak` (#1456; it needs `CMDCTRL_SCRYFALL_DUMP` from disk, which is why it is the one bot job that stays self-hosted while `bot-games` and `bot-soak` run on GitHub-hosted runners). The two skipped tests are the `bot-soak` job's, which runs them without `-race` (`.github/workflows/e2e-nightly.yml`).
  - `AISEAT_HEURISTIC_GAMES=N` / `AISEAT_H2H_GAMES=N` — widen the four-heuristic and heuristic-vs-random samples (defaults 3 and small, for CI). The nightly `bot-soak` job sets `AISEAT_HEURISTIC_GAMES=20`, which is S31 exit criterion 2 (#685); the seeds are fixed at 101..120, so it plays the same twenty tables every night. It sets no `AISEAT_H2H_GAMES`. Both tests play **lockstep** since #1409 (`playLockstepGame` in `heuristic_game_test.go`: the seats' ordinary act-loop called in seat order on one goroutine, through `aiseat.NewStepped` / `Runner.Step` since #1503), so a seed replays move for move — and `boteval arena --lockstep` plays the same schedule; the soak keeps the production one-goroutine-per-seat schedule, because concurrent liveness is what it tests (discussion #1390).
  - `AISEAT_HEURISTIC_SCHEDULE=concurrent` — plays the heuristic gate the old way, one runner goroutine per seat. For comparing the two schedules; the nightly does not set it.
  - `AISEAT_RANDOM_GAMES=N` / `AISEAT_RANDOM_SEED=<uint64>` — `TestFourRandomBotsPlayToAWinner`, S31's four-`random`-bots test: N consecutive games to a winner, default 3, from base seed 31000. The nightly runs it at `AISEAT_RANDOM_GAMES=20` in its own step without `-race`.
  - `AISEAT_REPLAY_DIR=<path>` — where that test writes each game's replay JSONL. It sets the location only, not whether replays are kept. Unset uses a temp dir. Either way a passing game's replay is deleted and a failing one kept, since one game is hundreds of MiB. The nightly points it at the workspace and uploads it on failure.
  - `AISEAT_FUNNEL_GAMES=N` — widen the Layer A absorption / model-funnel whole-game run.
  - `AISEAT_SOAK_GAMES=N` — independent of the master gate; runs `TestRandomBotSoak` for N games. `AISEAT_SOAK_POLICY=random|heuristic|mixed` picks what fills the seats (default `random`), and `AISEAT_SOAK_SEED=<uint64>` pins the base seed (default: the clock). Game *i* uses `base+i`, and the run logs the base seed and the replay command before the first game, so a green artifact still says which hundred tables were played. A stall prints its own seed too. The nightly `bot-soak` job sets `AISEAT_SOAK_GAMES=100` — S31 exit criterion 4 (#685) — with `AISEAT_SOAK_SEED` derived from the UTC date times 1000, so consecutive nights fuzz disjoint tables rather than overlapping in 99 of 100 games.
  - `CMDCTRL_PRINTED_CACHE_CHECKS=0` — set it when **profiling or benchmarking** a test binary (the 3-game soak at `AISEAT_SOAK_SEED=20260923000`, `GOMAXPROCS=2` is the usual recipe). In every test binary each hit of `game`'s printed-characteristic cache is rebuilt and compared ([#1498](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1498), `game/printed_cache.go`), which is the whole suite doubling as the cache's invalidation test — and exactly the cost the cache saves, so a profile taken with the check on shows it (`checkPrintedEntry`). It cannot turn the check on in production.
  - `AISEAT_DEBUG=1` — per-move log in the runner tests.
  - `AISEAT_LIVE_MODEL_TESTS=1` — **the only test in the tree that talks to a model over a network** (`aiseat/model/improvise_live_test.go`, #686): it sends one real improvisation call and checks the bundle against the rail the runner would put it through. Needs a transport too (`CMDCTRL_ANTHROPIC_API_KEY`, or `CMDCTRL_OPENAI_ENDPOINT` plus `CMDCTRL_BOT_MODEL`) and skips without one. **No workflow sets it and none should**: CI has no key, and a test whose result depends on a model's judgement is not a gate. Everything else in that package runs against `model.FakeClient`.
  - `AISEAT_DECISION_LOG=<dir>` — writes a per-game decision log for every game played through `playGame`/`playGameIn` (the same writer the server uses, `aiseat/decisionlog`). Unset (the default, and CI) costs nothing: the runner builds no event without an observer. `AISEAT_DECISION_LOG_MODE` takes the same three values as the server variable. This is how the position corpus gets harvested, and what `TestDecisionLogReplaysOffline` uses to prove a recorded window re-decides identically offline. **Mind the disk**: a four-seat game writes a few thousand windows and a four-player board view is ~50 KiB, so one game is 100–250 MiB. `escalated` only saves anything for policies that run Layer A (the `heuristic` TIER, `rules.NewFilter(heuristic.New(), …)`); the bare `heuristic.New()` the whole-game tests seat reports every window as Layer B and keeps them all in full.
  - `AISEAT_CATALOG_GAMES=N` / `AISEAT_CATALOG_SEED=<uint64>` — the catalog soak (`TestCatalogSoak`, #601): N four-bot games on decks dealt from the catalog itself rather than from the hand-written vanilla decks the other whole-game tests use. The seed defaults to one derived from the UTC date, so each night deals new decks and coverage accumulates; the log prints the base seed, and every failure prints the seed to replay. Needs `CMDCTRL_SCRYFALL_DUMP` as well (a `Spec` carries an oracle ID and a name, not a type line or a mana cost) and skips without it. It **fails on any `EventEffectError`** — a card whose primitive threw mid-resolution, which the engine logs and survives, so nothing else in the tree goes red over it.
  - `AISEAT_CATALOG_POLICY=random|heuristic|mixed` (#1456) — what fills the catalog soak's seats, modelled on `AISEAT_SOAK_POLICY` (the random soak, below): `random` (default, unchanged from before this variable existed) is the widest exploration of the catalog; `heuristic` seats `heuristic.New()` at every seat, which builds board states — blockers up, equipment attached, auras out — that `random` rarely sets up and most catalog cards actually need to be reached at all; `mixed` alternates by seat index. An unrecognised value fails the test outright rather than falling back silently, because a catalog soak run is expensive enough that a misspelled flag deserves a loud failure. The nightly's `catalog-soak` job sets `heuristic`.
  - `AISEAT_CATALOG_REPORT=<path>` — where the catalog soak writes its report JSON: `{"cards": [...], "census": {...}}`. `cards` is the per-card table (cast / resolved / entered / triggered / errored, per oracle ID) it always wrote. `census` is ADR 0041 P7's measurement (#1497, #1558 item 4) — the catalog soak captures a restore point (`game.Game.CaptureSnapshot`, the same call `ws.Room.writeRestorePointLocked` makes) after every observed action of every game in the run, and tallies what it found: `actions` and `captured` (how many of those captures were restorable), `by_kind` (blocked captures per census kind, summed over the whole run — `game.ContinuationCensus.Kinds`'s keys), and `longest_run_by_kind` (the longest run of CONSECUTIVE actions any one kind stayed the blocker — the number the shutdown census can't give, since it only ever reads one instant). It is not a gate: the shutdown census (#524) reads one instant per deploy and every deploy so far has caught an idle table, so this is the stand-in evidence P7 asks for to order phase 3's remaining tiers. Capture is sampled at the poll rate the stall detector already uses, so a burst of actions inside one poll tick coalesces into one capture — an undercount that can only make a stale run look shorter than it was, never longer. The nightly uploads the whole file as an artifact on every run, green included: the useful half of `cards` is the list of cards no bot game reached, and the useful half of `census` is whether tier 4 (stack effects and target specs, by a wide margin in every run so far) is worth hurrying.
  - `AISEAT_STALL` / `AISEAT_WALLCLOCK` — Go durations. Since [#685](https://github.com/krakenhavoc/cmd_and_ctrl/issues/685) **every bot-table budget in the package reads them**, so there is one pair of knobs for a loaded runner and no literal left to edit:
    - `TestFourRandomBotsPlay` — stall detector and wall clock, defaults `15s` / `300s`.
    - `TestCatalogSoak` — the same, same defaults.
    - `TestRandomBotSoak` — per-game stall detector and wall clock, defaults `3s` / `60s`.
    - `playGame` / `playGameIn`, which is every other whole-game test (four-heuristic, random-to-a-winner, head-to-head, latency, funnel, model, life-cost, decision-log replay) — stall detector, default `5s`.
    - `TestManagerPlaysALobbySeatedTable` — `AISEAT_WALLCLOCK` as its wait for the table to finish, default `300s`. Its wait for the runners to EXIT is no longer a budget at all: that is `waitForRunner` now (#1048).

    **`AISEAT_WALLCLOCK` is a floor on the `playGame` / `playGameIn` wall clock, not a replacement.** Each caller passes its own — 120s for a 50-turn heuristic table, 25 turns for the decision-log replay — and those numbers say something about the game being played, so the variable may raise them and never cuts them. Everywhere else it is the value. Unset (the default, and PR CI) changes nothing anywhere.

    Raise them on a loaded runner rather than editing a test. The nightly `bot-soak` job sets `AISEAT_STALL=30s` and `AISEAT_WALLCLOCK=300s`, and both are `workflow_dispatch` inputs.
  - `CMDCTRL_SCRYFALL_DUMP=<path>` — gates the manual bot-deck test that validates the four curated decks against the real Scryfall dump. Also used by other packages.

### Discord bot (Go, `server/cmd/bot/`)
- Separate binary from the game server; runs as `cmd-and-ctrl-bot.service` on the prod VPS. See [docs/decisions/0004-discord-identity.md](docs/decisions/0004-discord-identity.md).
- `make -C server build-bot` — produces `server/bin/cmd_and_ctrl-bot`
- Commands: `/c2-invite [name]` (channel-visible invite URL), `/c2-invite-dm <user> [game] [name]` (DM someone an invite; #613), `/c2-games` (ephemeral list), `/c2-end <game>` (confirm, then archive; #614), `/c2-deck-check [link]` and `/c2-deck-req [link]` (ADR 0095 §4 — deck coverage checks and deck requests; with no link, a paste modal, per the ADR's 2026-09-25 amendment).
- **Registration is a bulk overwrite** ([#1631](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1631), ADR 0095): `RegisterCommands` calls `ApplicationCommandBulkOverwrite` once per guild instead of creating one command at a time, so a stale registration (the old `cc-` names, or any command removed from `commandDefinitions`) disappears on the first boot of a new binary rather than needing a manual cleanup step.
- **Deferred replies.** `/c2-deck-check` and `/c2-deck-req` are the bot's first deferred interactions: both defer with `InteractionResponseDeferredChannelMessageWithSource`, ephemeral, giving about 20s instead of the other commands' 4s HTTP budget (a deck fetch plus a GitHub call can run long). A deferred response's visibility is fixed at defer time — `WebhookEdit` has no `Flags` field — so a command whose final reply must be channel-visible (`/c2-deck-req` on `filed` or `joined`) defers ephemerally, deletes that placeholder once the result is known, and posts the visible half as a follow-up message instead of an edit.
- `/c2-deck-check <link>` calls the public `POST /deck-coverage` with the bot's admin session (its own, larger rate bucket) and replies ephemerally with the bucket counts, up to ~15 `manual` card names, a link to the site's full report, and — when there is something to request — a **Request these cards** button. The button's custom ID carries the deck link URL-encoded; when that would exceed Discord's 100-character cap it stores the link server-side instead (same in-memory-with-TTL trade-off as `/c2-end`'s confirmations: a bot restart drops it, run the check again) and carries a short token. Pressing it runs the same request flow as `/c2-deck-req`, attributed to whoever clicked.
- **Pasted lists (ADR 0095, amendment 2026-09-25).** Moxfield blocks the server (a Cloudflare 403 on every endpoint), so both deck commands take `link` as optional. With no link the bot answers at once with a modal (`InteractionResponseModal`, custom ID `c2-deck-paste:check` or `c2-deck-paste:req`, one required paragraph input of up to 4000 characters). A modal must be the first response, so the deferral happens on submit: `Dispatch` handles `InteractionModalSubmit` behind the same guild allow-list and runs the same deferred flow with `{text}`. A pasted check's "Request these cards" button always carries a token into the in-memory store (which holds a link or a list), and its full-report line links `#/deck-check` plainly. A server error with `hint: "paste_list"` (any Moxfield fetch error) is shown with "Run `/c2-deck-check` with no link to paste it."
- `/c2-deck-req <link>` calls `POST /deck-requests` with the bot's admin session and `requester: {discord_id, display_name}` (`Member.Nick`, then `User.GlobalName`, then `User.Username`). `filed` and `joined` (a fresh comment) reply in the channel with the issue link; `joined` with `already_requested`, `nothing_to_add`, `rate_limited` and every error reply ephemerally.
- `/c2-invite-dm <user> [game] [name]` (#613) is a thin client of the server's `POST /games/{id}/invites/dm`: the SERVER opens the DM with its own `CMDCTRL_DISCORD_BOT_TOKEN`, and the bot never sends one from its gateway session. The bot calls it with its admin session and `discord_id` (the route accepts a raw snowflake from an admin session only). With no `game` it creates a table like `/c2-invite` (the invoker hosts) and DMs its invite; with a `game` (id or name prefix, autocompletes) it DMs that table's invite, but only for its creator or a `/c2-end` admin, because the server trusts the bot's admin session and the bot is the only gate. It defers (ephemeral, 20s budget) and answers ephemerally; a bot user is refused. A **503** from the route (server has no bot token or public origin) becomes a clear "DM invites are not set up" message that names `CMDCTRL_DISCORD_BOT_TOKEN`, and a just-created game's link is handed back to the invoker when the DM fails. It needs the server's `CMDCTRL_DISCORD_BOT_TOKEN`, which CD provisions on production only, so on develop it always answers that 503 message.
- `/c2-end` is **host or admin** (S34 [#1044](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1044) added `games.created_by`; [#1098](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1098) wired it into `server/internal/bot/end.go`). "Admin" is a Discord user on `CMDCTRL_DISCORD_ADMIN_USER_IDS` or a guild member holding a role on `CMDCTRL_DISCORD_ADMIN_ROLE_IDS`. "Host" is the game's own **creator** — whoever called `POST /games` while signed in — checked via `GET /games/{id}/creator?discord_id=<snowflake>` (admin-only; the bot calls it with its own admin session, per `Handler.mayEnd`). The server never says who a game's creator actually is, to the bot or anyone else — that route answers only "does this one Discord id match", so the bot never learns a creator's identity by asking about someone else's. Games with no creator (an admin session created the table, or it was restored from a pre-ADR-0051 file import) always answer `false` there, so the two allowlists are the only route for those. Both allowlists unset **and** no creator still refuses everyone with an ephemeral message naming the two variables — it never fails open. Confirmation is an ephemeral Confirm/Cancel prompt naming the table, players and created time; only the invoker's clicks count, and the prompt expires 60s after it's shown (`confirmTTL` in `end.go`).
- Env vars (the bot binary reads these; the server also reads `CMDCTRL_DISCORD_BOT_TOKEN`, for DM invites (#613), and `CMDCTRL_DISCORD_ADMIN_USER_IDS`, as its admin allowlist (ADR 0110 §3)):
  - `CMDCTRL_DISCORD_BOT_TOKEN` — **required**. Discord Developer Portal → Bot → Reset Token. In production: the Actions **secret** of the same name.
  - `CMDCTRL_DISCORD_APP_ID` — **required**. Application ID from the same portal. In production: the Actions **variable** of the same name.
  - `CMDCTRL_DISCORD_GUILD_IDS` — **required**. Comma-separated guild snowflakes; commands register only on these guilds and the bot rejects interactions from any other. In production: the Actions **variable** of the same name.
  - `CMDCTRL_ADMIN_TOKEN` — **required**. Same shared secret the server uses; the bot hits `POST /admin/login` + `POST /games` + `GET /games` + `POST /games/{id}/archive` over loopback. In production: copied by CD from `/etc/cmd_and_ctrl/env` on every deploy, never set separately.
  - `CMDCTRL_SERVER_BASE_URL` — default `http://127.0.0.1:8080`. Where the bot calls the admin API.
  - `CMDCTRL_CLIENT_BASE_URL` — default `https://cmd.labxp.io`. Used to compose the invite URL posted back to Discord.
  - `CMDCTRL_DISCORD_ADMIN_USER_IDS` / `CMDCTRL_DISCORD_ADMIN_ROLE_IDS` — **optional**, comma-separated snowflakes; who counts as **admin** for `/c2-end` (see above) — the game's own creator does not need either list, since that half of the check comes from `games.created_by` via the server, not from bot config. The user-ID list is **also the server's admin allowlist** (ADR 0110 §3; see the server env vars above), so a user listed here may switch admin mode on at the site (ADR 0112 §2) — the bot itself has no player mode; role IDs stay bot-only, because the server cannot see guild roles. In production: the Actions **variables** of the same names, written to `bot.env` by "Sync bot env" on **every** deploy, an unset one as an empty value, so clearing a variable removes it from the bot as well — leaving them unset is a supported state, not a misconfiguration.
- Unset `CMDCTRL_DISCORD_BOT_TOKEN` disables the bot (binary exits 0 after logging `bot disabled`). Convenient for dev stacks without a registered Discord app.
- Bot secrets live in a dedicated env file, `/etc/cmd_and_ctrl/bot.env` (`root:cmdctrl-bot`, mode `0640`), not the server's env file — ADR 0004 §6 explains why. The CD "Sync bot env" step writes it on production only; never hand-edit it. Host setup and verification: [deploy/README.md](deploy/README.md#discord-bot-production-only). Once the `CMDCTRL_DISCORD_BOT_TOKEN` secret is set, CD emits a `::warning::` (never a failure) for a production host that cannot run the bot: no `cmdctrl-bot` user or group, or the unit not enabled or not active after its restart.
- Each allowed guild authorizes the app with the `bot applications.commands` scopes and `permissions=0` (ADR 0004, revised 2026-09-16): the bot user must be a guild member to open DMs for #613. Install URL: [deploy/README.md](deploy/README.md#discord-scopes).

### Client (TypeScript + Svelte 5 + Vite, `client/`)
- `cd client && npm install` — first-time setup
- `npm run dev` — Vite dev server on :5173, proxies `/ws` to the Go server
- `npm run build` — type-check + production build into `client/dist/`
- `npm run check` — `svelte-check` typecheck only
- `npm run lint` — ESLint + Prettier check
- `npm run format` — auto-format
- **Labels are a contract** ([ADR 0076](docs/decisions/0076-tutorial.md) §2.4, [ADR 0111](docs/decisions/0111-action-dock.md) §10). The action dock's and the table's `aria-label`s and dialog names are what the e2e suite selects on and what the tutorial anchors to: `region "actions"`, `region "attention"`, `turn and phase indicator`, `group "priority controls"`, `group "declare attackers"`, `group "declare blockers"`, each dock request's dialog name (`keep or mulligan your hand`, `Discard N card`, `Select target for X`, a trigger's reason), and the button names `next`, `Pass turn`, `Keep hand`, `more actions`. Renaming one is a breaking change: change it only with the specs and tutorial steps that read it, in the same PR, and run the nightly Playwright job on the branch (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`), because PR CI does not run it.

---

## 6. When you're unsure

1. Re-read [PLAN.md](PLAN.md) section 2 (architectural decision) and section 6
   (roadmap).
2. Check the current sprint in [docs/sprints.md](docs/sprints.md) — the scope
   there is authoritative for "what should I be working on right now".
3. Look for an open decision in [PLAN.md](PLAN.md) section 7. If the question is
   listed there, surface it to the user rather than guessing.
4. Prefer a spike (time-boxed, throwaway) over speculative architecture.

### Rules citations

`CR NNN.Nx` in code, tests and docs means the **Magic: The Gathering
Comprehensive Rules effective September 25, 2026**, from
[magic.wizards.com/en/rules](https://magic.wizards.com/en/rules) (the TXT
download is `MagicCompRules 20260925.txt`; its text says "effective as of
September 25, 2026"). Check a number against that text before you write it. Do not
cite from memory: rule numbers move between editions. The June 2025 edition
re-sorted every keyword action in 701, so discard went from 701.8 to 701.9 and
reveal from 701.16 to 701.20. The 2026 edition added a new 310.8, which moved
the battle protector rules to 310.9. The September 2026 edition added a new
506.6 ("attacks a player alone"), which moved "had to attack" to 506.7 and the
combat timing rules for spells (old 506.7–506.7g) to 506.8–506.8g.

To move the pin to a newer edition, do it in one PR of its own. Download the
new TXT. For every section the tree cites, compare the rule's text in the two
editions, and renumber by matching content, never by adding to the number.
Then update the date and file name above.

A one-line grep does not find every citation. Comments wrap, so "CR" often
ends one line and the number starts the next. List the sections with a search
that crosses line breaks:

```sh
rg -U -o -N --no-filename 'CR\s*(//|\*|#)?\s*[0-9]{3}\.[0-9]+[a-z]?' . < /dev/null \
  | grep -oE '[0-9]{3}\.[0-9]+[a-z]?' | sort -u
```

Some citations have no "CR" in front at all: the second number in
"CR 305.1, 116.2a", and rule tables in comments such as the state-based
action list in `mutations.go`. Once you know which numbers moved, search for
each old number on its own too (`git grep -nw "701\.19"`), then read each hit.

---

## 7. Adding a catalog card (S14+)

The full guide moved to [docs/adding-cards.md](docs/adding-cards.md) (#1747).
**Read it before adding or changing any catalog card or engine seam.** It is
long on purpose; open the subsection you need, not the whole file.

Rules that must not be missed:

- **One file per card**, `server/internal/cards/effects/<snake_name>.go`, one
  `init()` each. The catalog is opt-in per Scryfall `oracle_id` (never
  `scryfall_id`).
- **Declare `Completeness` truthfully.** The zero value is `Unreviewed` and is a
  legal thing to ship; never stamp `Full` to tidy it. `Caveats` are written for
  a player, one sentence, no engine vocabulary.
- **Generate only your cards' oracle fixtures** (`-update-oracle -oracle-ids=...`);
  never regenerate them all in a card PR.
- **Stronger than printed is the wrong direction.** If the engine can't express
  a cost or clause, leave the card out rather than ship it without.
- **Grep for shared helpers first.** The clone gate
  (`TestNoNewExactClonesInTheCatalog`) fails a new duplicate body of six or more
  lines. Shared code goes in mechanic-named, append-only files.
- **Skipped cards** go on the seam's `Waiting` list in
  `server/internal/roadmap/registry.go`, then `go test ./internal/roadmap/ -update`.
  Never edit the docs table by hand.

Subsections of [docs/adding-cards.md](docs/adding-cards.md):

  - [Recipe](docs/adding-cards.md#recipe)
  - [Random effects (#744)](docs/adding-cards.md#random-effects-744)
  - [Adding a mana ability (S15+)](docs/adding-cards.md#adding-a-mana-ability-s15)
  - [Adding a static ability (S16+)](docs/adding-cards.md#adding-a-static-ability-s16)
  - [Granting an ability to another permanent (ADR 0093, #754)](docs/adding-cards.md#granting-an-ability-to-another-permanent-adr-0093-754)
  - [Alternative costs for every spell you cast (ADR 0118, #2163)](docs/adding-cards.md#alternative-costs-for-every-spell-you-cast-adr-0118-2163)
  - [Abilities any player may activate (ADR 0106, #1793)](docs/adding-cards.md#abilities-any-player-may-activate-adr-0106-1793)
  - [Adding a replacement effect (S17+)](docs/adding-cards.md#adding-a-replacement-effect-s17)
  - [Adding a copy effect (S16.5+)](docs/adding-cards.md#adding-a-copy-effect-s165)
  - [Adding a combat-keyword card (S18+)](docs/adding-cards.md#adding-a-combat-keyword-card-s18)
  - [Adding a "can't" card (S24+)](docs/adding-cards.md#adding-a-cant-card-s24)
  - ["Players can't play lands" (ADR 0109 §4, #1895)](docs/adding-cards.md#players-cant-play-lands-adr-0109-4-1895)
  - ["Cards in graveyards can't be targeted" (ADR 0109 §6, #1885)](docs/adding-cards.md#cards-in-graveyards-cant-be-targeted-adr-0109-6-1885)
  - ["Spells you control can't be countered" (ADR 0106, #1806)](docs/adding-cards.md#spells-you-control-cant-be-countered-adr-0106-1806)
  - [Attaching, and an ability whose source has gone (#812)](docs/adding-cards.md#attaching-and-an-ability-whose-source-has-gone-812)
  - [Adding a block-rule card (S37+, #750)](docs/adding-cards.md#adding-a-block-rule-card-s37-750)
  - [Adding a triggered ability (S19+)](docs/adding-cards.md#adding-a-triggered-ability-s19)
  - [Adding a triggered MANA ability (#763)](docs/adding-cards.md#adding-a-triggered-mana-ability-763)
  - [State triggers (ADR 0107, #1858)](docs/adding-cards.md#state-triggers-adr-0107-1858)
  - [Choices made at resolution (#796, #568)](docs/adding-cards.md#choices-made-at-resolution-796-568)
  - ["Reveals their hand. You choose a … card from it" (ADR 0116, #2078)](docs/adding-cards.md#reveals-their-hand-you-choose-a--card-from-it-adr-0116-2078)
  - [Adding a `PendingChoiceKind` (#730, #794)](docs/adding-cards.md#adding-a-pendingchoicekind-730-794)
  - [Cumulative upkeep (#567, CR 702.24)](docs/adding-cards.md#cumulative-upkeep-567-cr-70224)
  - [Echo (ADR 0108 §5, #1888, CR 702.30)](docs/adding-cards.md#echo-adr-0108-5-1888-cr-70230)
  - [The CR 732 loop breaker (#628)](docs/adding-cards.md#the-cr-732-loop-breaker-628)
  - [Untapping in another player's untap step (#74)](docs/adding-cards.md#untapping-in-another-players-untap-step-74)
  - [Phasing (#1199, CR 702.26)](docs/adding-cards.md#phasing-1199-cr-70226)
  - [Adding a preparation card (S46+, ADR 0090)](docs/adding-cards.md#adding-a-preparation-card-s46-adr-0090)
  - [Adding a hideaway card (S43+, ADR 0091)](docs/adding-cards.md#adding-a-hideaway-card-s43-adr-0091)
  - [Adding a creature-type card (S26+)](docs/adding-cards.md#adding-a-creature-type-card-s26)
  - [Adding a choose-a-color card (#742)](docs/adding-cards.md#adding-a-choose-a-color-card-742)
  - [Shared vocabulary, and the clone gate](docs/adding-cards.md#shared-vocabulary-and-the-clone-gate)
  - [Winning, losing, and "can't lose" (S40, ADR 0057)](docs/adding-cards.md#winning-losing-and-cant-lose-s40-adr-0057)
  - [When NOT to add a catalog entry](docs/adding-cards.md#when-not-to-add-a-catalog-entry)
  - [Adding a trigger doubler (#752)](docs/adding-cards.md#adding-a-trigger-doubler-752)
  - [Emblems (#623)](docs/adding-cards.md#emblems-623)
  - [The Ring tempts you (ADR 0114, #2076)](docs/adding-cards.md#the-ring-tempts-you-adr-0114-2076)
  - [Designations: Class levels, solved Cases, station thresholds (#757, #759)](docs/adding-cards.md#designations-class-levels-solved-cases-station-thresholds-757-759)
  - [Adding a Room or a split card (ADR 0103, #1756)](docs/adding-cards.md#adding-a-room-or-a-split-card-adr-0103-1756)
  - [Abilities from the hand (#660)](docs/adding-cards.md#abilities-from-the-hand-660)
  - [Special actions from the hand (#658, #659)](docs/adding-cards.md#special-actions-from-the-hand-658-659)
  - [Adding a `Spec` slot (#622)](docs/adding-cards.md#adding-a-spec-slot-622)
  - [When in doubt](docs/adding-cards.md#when-in-doubt)

---

## 8. Things to explicitly *not* do

- Do not add CI/CD beyond basic lint + test until there's code to protect.
- Do not introduce a database, auth provider, or payment anything without
  discussion. Shared password is fine for now (see [PLAN.md §4](PLAN.md#4-tech-stack)).
- Do not attempt a "full rules engine" sprint. Rules enforcement grows
  incrementally in the S13+ B→C track, one mechanic at a time, driven by
  real games. See [PLAN.md §2.3](PLAN.md#23-why-not-option-c-straight-custom-rules-engine).
- Do not reintroduce a dependency on XMage or any JVM tooling. That was
  evaluated and rejected in S01.
- Do not generate card art, card text, or anything else that would pull this
  project out of "private, personal use" territory.
- Do not create commits or PRs that do not reference a sprint (see section 4).
