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
    ├── sprints.md       # sprint plan
    └── decisions/       # ADRs (0001 WS library … 0093 granted abilities) — see §4 on numbering
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
- `make -C server build-boteval` — produces `server/bin/boteval`, the AI-bot evaluation harness (#837). `boteval probe [--endpoint URL] [--model ID]` sends ONE request in the funnel's exact shape to a model endpoint and prints prompt/completion tokens vs the client-side estimate, `finish_reason`, whether a `reasoning` field came back, the parsed index, and a verdict for each of the two transport failures that make a model seat play like a heuristic seat. Always exits 0. Not deployed; local tool.
- `boteval suite run [--dir DIR] [--policy heuristic|assisted|strong] [--max-think 20s] [--parallel N] [--out report.json] [--md]` — asks a policy every labelled position in `server/internal/aiseat/suite/testdata/positions/` and reports agreement overall and per tag, plus reject-hits, malformed replies, out-of-range indices and timeouts. Exits non-zero on a gated miss. The `heuristic` run is also an ordinary Go test and runs on every CI run.
- `boteval suite harvest --from 'dir/*.decisions.jsonl' --to inbox/ [--escalated] [--disagree] [--fallback a,b] [--layer A,B,C] [--seat 0,1] [--tag block,attack] [--limit N] [--seed N]` — pulls candidate windows out of decision logs into an inbox of UNLABELLED positions. Deterministic under `--seed`.
- `boteval suite render --pos path/to/position.json [--deck ID]` — prints the exact prompt a model would see for one position and the move list with `<- accept / reject / heuristic / model@capture` markers. This is the labelling screen. See [docs/bot.md](docs/bot.md#position-suite).
- `boteval arena --seats a,b,c,d [--decks …] --games N --rotate --out DIR` — headless bot-vs-bot games with the report block ADR 0052 asks every bot PR to carry: win rate with a **Wilson 95% interval** against the table's null rate (1/seats), the funnel's layer/escalation/timeout counters, and decision + model-call latency tails. Rotation seats contestant `k` at position `(k+i)%n` in game `i`, so turn order cancels. A model tier with **no endpoint is refused, not downgraded** (an `assisted` seat with no client plays the heuristic under a model tier's name). A stall is **reported, not fatal**. Artifacts land in `<out>/<RFC3339 start>/`: `summary.md`, `summary.json`, `games.jsonl` (streamed per game), `decisions/`, `replays/`. Wall clock: ~0.2 s per two-seat heuristic game, ~3.5 s per four-seat curated-deck game, 5–15 min per game with one local-model seat. See [docs/bot.md](docs/bot.md#arena).
- `cd server && go run ./cmd/gamecli -addr ws://localhost:8080/ws` — drive the demo game from a terminal; reads action JSON on stdin or via `-script path.json`
- Endpoints: `GET /healthz`, `GET /ws` (protocol v0, see [docs/protocol.md](docs/protocol.md)), `POST /admin/login`, `/games*` lobby routes (see [docs/lobby.md](docs/lobby.md)), `GET /me/games` + `POST /me/games/{id}/session` (a signed-in user's games and seat reclaim, ADR 0051 sub-PR 4), `GET /me/tablemates` + `POST /games/{id}/invites/dm` (the people you have played with, and DMing one of them this table's existing invite link — ADR 0051 decisions 8 and 5, S34 sub-PR 6), `GET /auth/discord/link` (link Discord to a held seat), `/cards/*` image + metadata routes, `GET /catalog` + `GET /catalog/image/{id}` (signed-in session required — the card catalogue, [ADR 0042](docs/decisions/0042-card-catalog-page.md); `main.go` says why it is not public), `GET /roadmap` (public, no session — the engine roadmap: card names and caveats only, never art or oracle text, [ADR 0092](docs/decisions/0092-public-roadmap-and-site-portal.md))
- Env vars:
  - `CMDCTRL_ADDR` — listen addr (default `:8080`)
  - `CMDCTRL_DATA_DIR` — data root (default `./data`; empty string disables disk writes + card cache). Holds `db/cmdctrl.sqlite` (ADR 0051, S34 sub-PR 1 — the persistent user/game/deck store, `internal/db`) and its `db/cmdctrl.backup.sqlite` VACUUM INTO copy, alongside the existing `scryfall/`, `images/`, `avatars/`, `bugreports/`, `restore/`, `replays/` and `games/`. Since S34 sub-PR 3 the lobby's games, seats and invites are rows in that database, and invites are stored as hashes. `lobby/` holds only the `<id>.json.imported` files the one-time importer renamed and left for a rollback (docs/environments.md).
  - `CMDCTRL_DB_BACKUP_INTERVAL` — Go duration between the database's in-process `VACUUM INTO` backup sweeps (default `1h`; `<= 0` disables the sweep). The sweep writes `db/cmdctrl.backup.sqlite` beside the live file so a disk-level backup picks up a consistent copy. This is the same-disk copy only. The nightly off-node copy is `scripts/backup-offsite.sh` (restic to Cloudflare R2, #1031, ADR 0051 decision 1 as amended), run by `cmd-and-ctrl-backup.timer` on both hosts, not by this server. Runbook: [docs/environments.md](docs/environments.md#backups).
  - `CMDCTRL_ADMIN_TOKEN` — **required**. Shared admin secret for `POST /admin/login`. At least 16 characters.
  - `CMDCTRL_SESSION_TTL` — session lifetime as a Go duration (default `12h`). Applies to seat, spectator and admin sessions. The identity session a Discord sign-in mints uses `CMDCTRL_IDENTITY_TTL` instead.
  - `CMDCTRL_IDENTITY_TTL` — lifetime of the identity-only session the Discord callback mints from the login page, as a Go duration (default `720h`, 30 days; S34 sub-PR 7, [ADR 0051](docs/decisions/0051-user-database.md) decision 3). It is long because it is the credential a browser keeps across games. That is safe because it carries a `UserID` and can be revoked (`POST /logout/everywhere`, `POST /admin/users/{id}/revoke-sessions`, decision 6). Invalid or `<= 0`: the boot fails. Not provisioned by CD. The client's expiry timer re-arms past `setTimeout`'s ~24.8-day ceiling, so a 30-day session is not dropped early.
  - `CMDCTRL_SESSION_KEY` — HMAC-SHA256 key that signs session tokens (`auth.HMACAuthenticator`, #517, [ADR 0044](docs/decisions/0044-surviving-a-deploy.md) decision 3). Set: a token minted before a restart validates after it. **Unset: sessions fall back to the in-memory store, with a warning naming the variable on every boot, and every deploy logs everyone out.** Set but under 32 bytes, or equal to `CMDCTRL_ADMIN_TOKEN`: the boot fails. There is no default key. `Revoke` is advisory under HMAC, so `POST /logout` clears the cookie but does not kill a copied token before its expiry. A session with a `UserID` (a Discord sign-in, or a seat claimed from one) can be killed: `POST /logout/everywhere` or the admin's `POST /admin/users/{id}/revoke-sessions` moves `users.sessions_invalid_before`, and `auth.WithRevocation` refuses older tokens from an in-memory cache (ADR 0051 decision 6). Admin, guest and spectator sessions keep the advisory revoke. **Provisioned by CI/CD**: generated on each host the first time it is missing and never rewritten, since a new key logs every player out. See [docs/environments.md](docs/environments.md#secrets).
  - `CMDCTRL_IDENTITY_KEY` — AES-256-GCM key that encrypts Discord OAuth refresh tokens on the `identities` rows (`users.Sealer`, S34 sub-PR 2, [ADR 0051](docs/decisions/0051-user-database.md) decision 5). Same format as `CMDCTRL_SESSION_KEY`: a random string of at least 32 bytes, which the server hashes to the AES key. **Unset: Discord sign-in still works and still records the user, but the refresh token is discarded and `refresh_token` is NULL, with a warning naming the variable on every boot. It is never stored in the clear.** Set but under 32 bytes, or equal to `CMDCTRL_ADMIN_TOKEN` or `CMDCTRL_SESSION_KEY`: the boot fails. Rotating it makes stored refresh tokens unreadable (nothing reads them yet) and logs nobody out. **Provisioned by CI/CD** exactly like the session key: generated on each host the first time it is missing, never rewritten.
  - `CMDCTRL_DISCORD_BOT_TOKEN` — the Discord **bot** token, read by the SERVER for ADR 0051 decision 5's direct-message invites (`POST /games/{id}/invites/dm`, S34 sub-PR 6). Same secret the gateway bot holds; two units, two env files, one Actions secret. **Unset: the DM route answers 503 naming this variable and every other route is unchanged — it never fails open — and the boot log says which state it is in, once.** The route also needs an origin for the link it sends (`CMDCTRL_PUBLIC_BASE_URL`, falling back to `CMDCTRL_CLIENT_BASE_URL`); missing that is the same 503. **Provisioned by CI/CD on production only** ("Sync server env (Discord DM invites)"): one Discord application, and a preview box DMing real people from the same identity is the double-send the bot's prod-only rule exists to prevent.
  - `CMDCTRL_ALLOWED_ORIGINS` — comma-separated hostnames (or full URLs) permitted as cross-origin WebSocket callers. Same-origin is always allowed; unset = same-origin only.
  - `CMDCTRL_SEED_DEMO=1` — seed the S03 4-player demo game at startup for the gamecli dev loop
  - `CMDCTRL_DISCORD_CLIENT_ID` / `CMDCTRL_DISCORD_CLIENT_SECRET` / `CMDCTRL_DISCORD_REDIRECT_URI` — S12.5 OAuth credentials. Unset disables the Discord sign-in button (manual name entry still works).
  - `CMDCTRL_SECURE_COOKIES` — truthy sets the `Secure` attribute on the session cookie. Enable in any TLS deployment; leave unset for plain-HTTP local dev (a `Secure` cookie is never sent over http and would silently break login).
  - `CMDCTRL_TRUST_FORWARDED` — truthy keys the lobby rate limiter off the leftmost `X-Forwarded-For` hop instead of the socket `RemoteAddr`. Enable **only** when the server sits behind a trusted reverse proxy that sets the header; otherwise clients can spoof it to dodge limits.
  - `CMDCTRL_DEV_RELAX_RATE_LIMITS` — truthy effectively disables the lobby rate limiters. For the e2e suite and local load-y dev loops only; logs a loud warning on boot. **Never set in production.**
  - `CMDCTRL_GITHUB_TOKEN` — enables the in-app "report a bug" button (`POST /bugreport` files a GitHub issue). Use a fine-grained PAT with **Issues: write** on the one repo, nothing broader. Unset disables the feature and the client hides the button. See [ADR 0017](docs/decisions/0017-bug-report-button.md). **Reports are published to everyone who can read the repo, so credentials are redacted before they get there** — client (`client/src/lib/redact.ts`, applied to the protocol log, console capture and the submitted draft) and server (`server/internal/util/redact`, applied in `renderBugIssueBody`); never log a token-bearing URL raw, pass it through `redactURL` (#721, [ADR 0017 §9](docs/decisions/0017-bug-report-button.md); durable sessions #517 depend on it). **Provisioned by CI/CD**: the deploy job upserts it into `/etc/cmd_and_ctrl/env` (the env file the HomeLab cloud-init template writes and the systemd units read) from the `CMDCTRL_GITHUB_TOKEN` Actions secret, via `sudo scripts/set-server-env.sh`; to rotate, update the secret and rerun the deploy — no host access needed.
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
  - `CMDCTRL_BOT_DECISION_LOG` — directory for the per-game bot decision log (one JSONL line per decision window per bot seat: prompt, reply, heuristic ranking, fallback cause, latency). Empty (the default) is **off**. **Operator-only**: each record is the seat's own filtered view, but the file aggregates every bot seat at the table, so it is never served over HTTP and never attached to a bug report. Directory `0700`, files `0600`, 256 MiB per game. See [docs/bot.md](docs/bot.md#decision-log).
  - `CMDCTRL_BOT_DECISION_LOG_MODE` — `escalated` (default: full board view only for windows that left Layer A) | `all` | `model` (only the windows a model answered). An unrecognised value fails the boot **when the log is on**, like `CMDCTRL_BOT_MAX_THINK`; with the log off it is a warning, because refusing to start over a variable that changes nothing is a server that does not come back after a rollback.
- Cron: `scripts/scryfall-refresh.sh` — weekly refresh of the Scryfall default-cards dump (suggested cron: `0 5 * * 0`)
- Off-site backup: `scripts/backup-offsite.sh`, run nightly by `deploy/cmd-and-ctrl-backup.service` + `.timer` (as `cmdctrl`, data dir read-only) on both hosts. It uses restic to back up the data dir to a per-host Cloudflare R2 bucket, taking `db/cmdctrl.backup.sqlite` but never the live db, and skipping the `scryfall/`, `images/` and `avatars/` caches. Credentials are in `/etc/cmd_and_ctrl/backup.env` (`root:cmdctrl 0640`, separate from the server's env), written by the CD step "Ensure off-site backup" from the `CMDCTRL_R2_*` and `CMDCTRL_RESTIC_PASSWORD` values in the `prod` / `dev` GitHub environments. A missing value, or one still containing `REPLACE_ME`, is a `::warning::` and a disabled timer, never a failed deploy. **The restic password must never change once a repository exists**; its recovery copy is the owner's password manager. Runbook, including the restore: [docs/environments.md](docs/environments.md#backups) (#1031).

### Snapshot compatibility (#522)

A restore point written by yesterday's binary has to restore in today's. Three tests hold that line, and the rule they enforce is written next to `SnapshotSchemaVersion` in `server/internal/game/snapshot.go`: changes within a version are additive only, anything else is a bump, and fixtures are never edited.

- **The shape guard.** `TestSnapshotShapeIsRecorded` (`internal/game`) compares the snapshot structs' JSON shape with `internal/game/testdata/snapshot_shape/v<N>.txt`. If you add a snapshot field, record it in the same change: `cd server && go test ./internal/game -run TestSnapshotShapeIsRecorded -args -update-shape`. The tool refuses a non-additive change (a key renamed, removed or retyped) under the current number. Bump `SnapshotSchemaVersion`, say why in its comment, and rerun it. It writes `v<N+1>.txt` and leaves the old file frozen.
- **The fixture corpus.** `TestSnapshotCorpusRestores` (`internal/cards/effects`, because restoring tokens, Clones and Equipment needs the real catalog) restores every file under `internal/game/testdata/snapshots/` with `RestoreStrict`. It fails if any key or value in a fixture is missing from a fresh capture of the restored game, if a card comes back with fewer catalog abilities than the file recorded, or if the game will not round-trip or take an action.
  - `v<N>/` is the generated set for schema N. It is written once, when N is introduced: `cd server && go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus`. The writer never touches an existing **file** (ADR 0041 phase 3, owner decision 6): a board with no file yet is written beside the others, so a shape that becomes a restore point after its version was introduced can still be frozen without a bump. It fails if the current build would render an existing fixture differently, and tells you to bump instead. CI fails if the current version has no directory.
  - `real/` holds scrubbed restore points from cmd-dev. Add one with `cd server && go run ./cmd/snapshotscrub -in <restore/id.json> -out internal/game/testdata/snapshots/real/<name>.json`. The tool replaces names and player IDs, drops Discord IDs and avatar hashes, and refuses to write if a snowflake, an email or an original player name is left anywhere. It never overwrites a file.
  - **Never edit, regenerate or delete a fixture to make the test pass.** If a bump migrates an old shape, list the migrated paths in `corpusMigrations` in `snapshot_corpus_test.go`, with the reason. Nothing else may excuse a difference.
- **The ability check.** A card whose catalog entry lost abilities between the writing and the reading binary is restored anyway. It is flagged `Card.AbilitiesLostOnRestore`, which shows it as `manual` for the rest of the game, and the boot log has an ERROR line naming the game, the card and the counts. More abilities than captured is not a mismatch.
- **Effect keys (#1497, [ADR 0041](docs/decisions/0041-game-persistence.md) phase 3).** A `ScopedEffect` mod kind is an on-disk identity, like a token slug: never renamed, never reused. A restore point naming a kind (or a delayed-trigger / stack-item effect key) this binary cannot interpret is refused with `ErrUnknownEffectKey` and the file is KEPT — only a newer build can have written it, so it is the rollback case. That refusal is what lets the vocabulary grow within a schema version.
- **The closure ratchet.** `TestClosureFieldsReachableFromGame` (`internal/game`) lists every ROUTE from `Game` to something that can hold a func or an interface — every field of every reachable struct whose type reaches one, directly or through another struct — in `internal/game/testdata/closure_fields.txt`, each classified `rebuilt`, `keyed`, `transient`, `test-only` or `census:<Counter>`. A new route fails until it is classified, including a new field whose type is a struct already on the list (#1558): `cd server && go test ./internal/game -run TestClosureFieldsReachableFromGame -args -update-closure-fields` keeps the existing classes and marks new routes `unclassified`. The number of lines in each `census:` class and in `transient` is pinned by `closureClassCeilings` in the test and may only fall: phase 3's tiers delete lines and lower the ceiling; nothing raises one.

### AI bot seat (Go, `server/internal/aiseat/`)

- Runs **in process** with the game server — no separate binary, no socket. One goroutine per bot seat, started by `Lobby.Start` and by the lobby's restore path. User-facing guide: [docs/bot.md](docs/bot.md); architecture: [ADR 0033](docs/decisions/0033-ai-bot-seat.md).
- Endpoints: `GET /bot/options`, `POST /games/{id}/seats/bot`, `DELETE /games/{id}/seats/bot/{player_id}` — see [docs/lobby.md](docs/lobby.md).
- **`aiseat.Start` subscribes before it returns (#938).** The room subscription is taken on the CALLER's goroutine, so "Start returned" means "this seat is listening from now on" — the only ordering a caller can establish, since nobody schedules the runner goroutine. Taking it inside the loop left a window in which a commit reached every other seat and not this one; the first step still reads the live game, so no state was lost, but a seat that DECLINES a window parks until the next commit, and a commit that landed in the window is a wake that is never coming. Anything new that starts a runner keeps that property: do not move the `room.Subscribe()` back onto the runner's goroutine.
- **A wait in these tests must poll a MONOTONE predicate (#634).** A bot table moves on a goroutine the test does not schedule — one turn cycle of the two-seat runner fixture is ~95ms — so any condition the table merely passes through can be missed outright by a test goroutine that loses the processor for a cycle, and then either never comes true again (a 30s `waitFor` timeout) or comes true one turn late against the wrong board. `n == 1 land`, `ActiveSeat == 1` and any `== k` on a counter that keeps climbing are all that shape. Wait for a thing that has HAPPENED and stays happened — a `>=` on a monotone counter, a latch, or a tally over the runner's own decision log (`aiseat.Config.Observer`), which is a complete history rather than a sample. `waitFor` / `waitForRunner` / `waitForChan` in `runner_test.go` are the only wait primitives; the budget is a backstop, never an assertion (#848). `waitForChan` is for the handshakes that are a channel rather than a predicate — a policy reporting that it was asked, a blocking call reporting that it returned — and it exists so that a `select` on a channel against a `time.After` never has to be written here again (#1048). `internal/aiseat` now has no `select` on a channel against a `time.After` at all; what is left is two deliberate short sleeps inside negative assertions (`runner_test.go`'s 300ms block grace, `manager_decisionlog_test.go`'s 50ms after `Shutdown`), which a slow machine makes MORE likely to pass (#876).
- **The bot reads one `Spec`-level signal, and only one (#686).** `protocol.CardView.Unimplemented` — `game.Unimplemented(c)`, i.e. `c.NeedsEffect && !IsAutoCard(CatalogKey(c))`, the same bit behind the deck-upload summary and the stack overlay's `manual` chip — is what tells an `assisted` / `strong` seat that a spell it just cast resolved into silence and is a candidate for ADR 0033 §8 improvisation (`aiseat/model/improvise.go`). Two consequences for anyone changing coverage: a card that stops being flagged stops being improvised, silently; and the bit reaches the bot through the WIRE VIEW, not through the catalog, so `aiseat/` still imports nothing from `internal/game` (`heuristic/imports_test.go` enforces it). Nothing else in `aiseat/` reads a catalog signal at all — the whole point of ADR 0033 §3 is that a policy sees a filtered view and a move list.
- Env vars: the bot seat reads only the model-transport variables listed above at runtime: `CMDCTRL_OPENAI_ENDPOINT` / `_API_KEY` / `_SEND_THINK`, `CMDCTRL_ANTHROPIC_API_KEY` / `_ENDPOINT` (and `ANTHROPIC_API_KEY`), and `CMDCTRL_BOT_MODEL` / `_FRONTIER_MODEL` / `_MAX_THINK`, plus `CMDCTRL_BOT_IMPROVISE` (`0` turns §8 improvisation off; on by default for the model tiers) and the off-by-default `CMDCTRL_BOT_DECISION_LOG` / `_MODE`. They are parsed in `cmd/server/main.go` and `aiseat/model`. Everything else is per-seat request data or a compile-time default. Tier and deck come with the request.
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
- Commands: `/cc-invite [name]` (channel-visible invite URL), `/cc-games` (ephemeral list), and `/cc-end <game>` (confirm, then archive; #614).
- `/cc-end` is **host or admin** (S34 [#1044](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1044) added `games.created_by`; [#1098](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1098) wired it into `server/internal/bot/end.go`). "Admin" is a Discord user on `CMDCTRL_DISCORD_ADMIN_USER_IDS` or a guild member holding a role on `CMDCTRL_DISCORD_ADMIN_ROLE_IDS`. "Host" is the game's own **creator** — whoever called `POST /games` while signed in — checked via `GET /games/{id}/creator?discord_id=<snowflake>` (admin-only; the bot calls it with its own admin session, per `Handler.mayEnd`). The server never says who a game's creator actually is, to the bot or anyone else — that route answers only "does this one Discord id match", so the bot never learns a creator's identity by asking about someone else's. Games with no creator (an admin session created the table, or it was restored from a pre-ADR-0051 file import) always answer `false` there, so the two allowlists are the only route for those. Both allowlists unset **and** no creator still refuses everyone with an ephemeral message naming the two variables — it never fails open. Confirmation is an ephemeral Confirm/Cancel prompt naming the table, players and created time; only the invoker's clicks count, and the prompt expires 60s after it's shown (`confirmTTL` in `end.go`).
- Env vars (bot binary reads these; server binary does not yet — ADR 0051 Decision 5's DM invites, #613, will add the bot token to the server's env too):
  - `CMDCTRL_DISCORD_BOT_TOKEN` — **required**. Discord Developer Portal → Bot → Reset Token. In production: the Actions **secret** of the same name.
  - `CMDCTRL_DISCORD_APP_ID` — **required**. Application ID from the same portal. In production: the Actions **variable** of the same name.
  - `CMDCTRL_DISCORD_GUILD_IDS` — **required**. Comma-separated guild snowflakes; commands register only on these guilds and the bot rejects interactions from any other. In production: the Actions **variable** of the same name.
  - `CMDCTRL_ADMIN_TOKEN` — **required**. Same shared secret the server uses; the bot hits `POST /admin/login` + `POST /games` + `GET /games` + `POST /games/{id}/archive` over loopback. In production: copied by CD from `/etc/cmd_and_ctrl/env` on every deploy, never set separately.
  - `CMDCTRL_SERVER_BASE_URL` — default `http://127.0.0.1:8080`. Where the bot calls the admin API.
  - `CMDCTRL_CLIENT_BASE_URL` — default `https://cmd.labxp.io`. Used to compose the invite URL posted back to Discord.
  - `CMDCTRL_DISCORD_ADMIN_USER_IDS` / `CMDCTRL_DISCORD_ADMIN_ROLE_IDS` — **optional**, comma-separated snowflakes; who counts as **admin** for `/cc-end` (see above) — the game's own creator does not need either list, since that half of the check comes from `games.created_by` via the server, not from bot config. In production: the Actions **variables** of the same names, written by "Sync bot env" only when set — leaving them unset is a supported state, not a misconfiguration.
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
Comprehensive Rules effective August 7, 2026**, from
[magic.wizards.com/en/rules](https://magic.wizards.com/en/rules) (the TXT
download is `MagicCompRules 20260819.txt`; its text says "effective as of
August 7, 2026"). Check a number against that text before you write it. Do not
cite from memory: rule numbers move between editions. The June 2025 edition
re-sorted every keyword action in 701, so discard went from 701.8 to 701.9 and
reveal from 701.16 to 701.20. The 2026 edition added a new 310.8, which moved
the battle protector rules to 310.9.

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

The catalog at [server/internal/cards/effects/](server/internal/cards/effects/) is
opt-in per Scryfall `oracle_id`. A card that isn't in the catalog keeps its
pre-S14 manual sandbox behaviour; a card that is in the catalog resolves
automatically at the right stack boundary. Each card is one `init()` in its
own file — one-file-per-card keeps `git blame` clean and the merge-conflict
surface tiny.

### Recipe

1. **Find the oracle ID.** The Scryfall bulk dump at
   `data/scryfall/default-cards.json` (local only; refreshed weekly via the
   scheduled workflow) has every printing. One-liner:
   ```bash
   python3 -c "import json; d=json.load(open('data/scryfall/default-cards.json')); \
     print(next(c['oracle_id'] for c in d if c['name']=='CARD NAME'))"
   ```
   `oracle_id` (NOT `scryfall_id`) is the catalog key — stable across
   printings.

   **Skip placeholder printings.** The dump carries ~3,200 entries that
   are not playable cards: art-series cards (`layout: "art_series"`,
   `type_line: "Card // Card"`, and a name that is the real card's name
   **doubled** — `"Appa, Steadfast Guardian // Appa, Steadfast
   Guardian"`), plus `front_card` / token placeholders with
   `type_line: "Card"`. Any name match looser than `==` picks them up,
   and then an ordinary single-faced creature looks like a DFC. Filter
   `c['type_line'] not in ('Card', 'Card // Card')` before reading
   `layout`, `type_line`, `mana_cost` or `card_faces` off a printing.

2. **Pick a primitive composition.** See
   [server/internal/cards/effects/primitives.go](server/internal/cards/effects/primitives.go)
   for the primitives (22 in `primitives.go` as of S22, plus a few that
   live in their own files — `CreateTokenCopy` in `token_copy.go`,
   the flicker helpers in `flicker.go`). Most cards are 1-2 primitives
   sequenced. A
   card that can't be expressed with existing primitives either needs a new
   primitive (add it to `primitives.go`) or a new `*ForEffect` helper on
   `*Game` (under a lock caller already holds — follow the existing naming
   in [server/internal/game/effect_api.go](server/internal/game/effect_api.go)).

3. **Declare the target (S20).** If the card has a target, build a
   `Spec.Targets` from the constructors in
   [targets.go](server/internal/cards/effects/targets.go) so it reads
   like the oracle text:
   ```go
   Targets: TargetAny(),                                           // Lightning Bolt
   Targets: TargetCreature("target nonblack creature", NonBlack()), // Doom Blade
   Targets: TargetSpell("target noncreature spell", Noncreature()), // Negate
   Targets: TargetPlayer("target opponent", Opponent()),
   Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
   Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
   ```
   Predicates compose with `And` / `Or` / `Not`; add missing ones to
   `targets.go`, not to the card file. Multi-target clauses set the
   count on the same spec — `TargetCreature("two target nonartifact
   creatures", Not(Artifact())).WithCount(2, 2)`, `.WithCount(0, 2)`
   for "up to two", `.WithCount(1, 0)` for "any number" — and their
   `OnResolve` iterates `ctx.LegalTargets()` (or indexes
   `item.Targets` with `ctx.IsTargetLegal` per slot when the order
   matters, as in Arc Trail) so a target that left in response is
   skipped rather than erroring.

   **Target clauses (#764).** A count is one predicate chosen N
   times. When the slots have DIFFERENT predicates — Bite Down's
   "target creature you control" then "target creature or
   planeswalker you don't control" — they are separate CLAUSES, and a
   statement is an ordered list of them:
   ```go
   Targets: Clauses(
       TargetCreature("target creature you control", YouControl()),
       TargetPermanent("target creature or planeswalker you don't control",
           Or(Creature(), Planeswalker()), OpponentControls()),
   ),
   // "a SECOND target permanent you control" — must differ from the first:
   Targets: Clauses(
       TargetPermanent("target permanent you control", YouControl()),
       Distinct(TargetPermanent("a second target permanent you control", YouControl())),
   ),
   ```
   A `game.TargetSpec` **is** its first clause and hangs the rest off
   it (`Rest`), so a one-clause card, a cost-payment predicate
   (`SacrificeOther` and friends) and a mode's clause are all the same
   struct and the same walk — see
   [ADR 0065 §1](docs/decisions/0065-modal-and-multi-target-clauses.md).
   Each clause is enforced on its own at announce (CR 601.2c) and
   re-checked on its own at resolution (CR 608.2b), so a pair that
   fits the wrong slots is REFUSED rather than resolving to nothing.
   Read the slots back with `ctx.ClauseTarget(slot)` /
   `ctx.ClauseTargets(slot)`; `item.Targets` is still one flat list in
   announce order, so a positional reader keeps working. The engine computes the legal
   set for the client's picker on every snapshot, rejects an illegal
   pick at announce (`ErrIllegalTarget`, CR 601.2c), and re-runs the
   same predicate at resolution (CR 608.2b). Colour predicates read
   `Card.Colors` (Scryfall's computed colours; mana-cost fallback for
   fixtures). See [ADR 0019](docs/decisions/0019-structured-targeting.md).
   The legacy `TargetMode` string is derived from `Targets.Mode` —
   set it directly only for a card you deliberately leave on the
   free-form picker. Empty means no prompt.

   **Modal cards ("Choose one —", S20 sub-PR 4)** declare
   `Spec.Modes` instead of `Spec.Targets`, with the target clause on
   the option that has one:
   ```go
   Modes: ChooseOne(
       Mode("Exile target player's graveyard.", TargetPlayer("target player")),
       Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
       Mode("Each creature deals 1 damage to its controller."),
   ),
   // "Choose two —": ChooseN("Choose two", 2, 2, Mode(…), Mode(…), …)
   ```
   `OnResolve` is a run of `if ctx.HasMode(i) { … }` blocks in
   printed order (CR 608.2c). The engine validates the choice at
   announce and applies the chosen option's target clause exactly as
   it would a card-level one; the client shows a mode picker before
   targeting.

   **Modes (#764): one `ModeSpec`, three owners.** The same
   `game.ModeSpec` is read by `Spec.Modes` (a spell),
   `TriggeredAbility.Modes` (a trigger) and `ActivatedAbility.Modes`
   (an activated ability) — see
   [ADR 0065 §3](docs/decisions/0065-modal-and-multi-target-clauses.md).
   What differs is only WHEN the choice is made:

   - a **spell** announces its modes at CR 601.2b, with the cast;
   - an **activated ability** announces them at CR 602.2b, with the
     activation — one indivisible message, no prompt;
   - a **trigger** is put on the stack by the engine, so it asks:
     a `mode_pick` pending choice at CR 603.3c, after the "you may"
     prompt and before the CR 603.3d target pick. A bullet whose
     clause has no legal target is not offered, and if that leaves
     fewer than `Min` the trigger is removed (CR 603.3d).

   Every bullet targets if it wants to — the old "one targeted option
   per cast" panic in `Register` is gone — and each chosen occurrence
   gets its OWN target group. Constructors: `ChooseOne`, `ChooseN`,
   `ChooseOneOrMore` (Sublime Epiphany) and `ChooseNRepeating`
   (CR 700.2d, "you may choose the same mode more than once" — Mystic
   Confluence). A trigger or activated ability has no `OnResolve` to
   branch in, so declare each bullet's body on the option with
   `ModeDoing(label, targets, fn)`; the engine runs the chosen ones
   in announce order, once per occurrence. Inside a bullet, read its
   own targets with `ModeTarget(ctx, occurrence)` /
   `ctx.ModeTargets(occurrence)` — never `item.Targets[0]`, which
   belongs to whichever bullet was chosen first.

4. **Write the card file.** One file per card at
   `server/internal/cards/effects/<snake_name>.go`:
   ```go
   package effects

   import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

   // Card Name — "full oracle text quoted."
   //
   // S14 sandbox simplifications (if any — enters-tapped, auto-pick,
   // target-at-ETB deferrals, etc.). Be specific about WHAT is
   // deferred and to which future sprint.
   func init() {
       Register(Spec{
           OracleID:     "<uuid from step 1>",
           Name:         "Card Name",
           Completeness: CompletenessFull,
           TargetMode:   "<player / creature / ...>",
           OnResolve: func(item *game.StackItem, ctx *Context) error {
               // primitive composition here
               return nil
           },
       })
   }
   ```
   A permanent's printed "When ~ enters" trigger goes in
   `Spec.Triggered` watching `EventETB` (see "Adding a triggered
   ability" below) so it uses the stack and can be answered.
   `Spec.AsEnters` is only for CR 614.12 "As ~ enters, choose …"
   effects, which are not triggers and correctly happen off the
   stack; it was called `OnETB` and misused for both until #578.

   **Declare `Completeness`.** Since
   [ADR 0042](docs/decisions/0042-card-catalog-page.md) the prose
   simplification note above has a machine-readable twin, because the
   public catalogue page at `#/catalog` publishes it:

   ```go
   Completeness: CompletenessFull,        // everything printed happens
   // …or:
   Completeness: CompletenessCaveats,
   Caveats:      []string{"Cycling is not implemented — the land can only be played."},
   ```

   Four rules, and they are the whole contract:

   - **The zero value is `CompletenessUnreviewed`, and that is a legal
     thing to ship.** It publishes the card as unaudited, which is
     true, and nothing fails. Do not stamp `CompletenessFull` to tidy
     it up — a card falsely marked complete is the one outcome the
     field exists to prevent.
   - **`Caveats` is required with `CompletenessCaveats` and rejected
     without it.** `Register` panics either way, at boot.
   - **Write `Caveats` for a player, not for the next engineer.** One
     sentence, no engine vocabulary: "Flashback isn't implemented — the
     spell can only be cast from hand." The reason it is deferred, the
     sprint it lands in and the machinery it waits on all belong in
     the doc comment, where there is room. A test enforces the tone.
   - **A caveat goes stale the day someone else implements the
     mechanic**, in another PR, in a file you will never open. #412
     measured 24 of 87 declared simplifications already describing
     closed gaps; four more turned up in one September session
     (flashback, warp, the free cast, the surveil lands). So when you
     land a mechanic, `grep -ril "<mechanic>" server/internal/cards/effects/`
     and clear the caveats it just invalidated.
     `server/internal/cards/coverage` catches the part of that class a
     program can see: a caveat naming a mechanic the same card now
     declares fails the build outright, and one naming a mechanic some
     OTHER card already uses has to be pinned with a reason. It is a
     curated table of mechanic probes, not an analysis — a caveat
     about something not in the table is invisible to it, which is why
     the grep is still your job.

   Keep the prose note too. The field says *what*; the comment says
   *why*, and the comment is what stops the next person reopening a
   decision you already made.

   **Planeswalkers: leave `StartingLoyalty` alone.** Starting loyalty
   is printed card data, not card-effect data. The deck importer
   parses Scryfall's `loyalty` onto `game.Card.StartingLoyalty` and
   the engine stamps the counters on every battlefield entry, catalog
   entry or not — see
   [ADR 0032](docs/decisions/0032-planeswalkers.md). `Spec.StartingLoyalty`
   survives only as a fallback for cards that never go through deck
   import (tokens, fixtures); setting it on a real card is redundant
   at best. Loyalty *abilities* are still deferred — `AbilityCost` has
   no loyalty component, so don't invent one (see the deferral list
   below).

5. **Add a test case** in
   [cards_test.go](server/internal/cards/effects/cards_test.go). Use
   `newCatalogGame(t)` + `castCatalogSpell(t, g, name, typeLine, oracleID, targets)`
   + `passPriorityAroundTable(t, g)` and assert the resulting state. For
   an ETB trigger that call settles the spell and the trigger it queues;
   assert `triggerOnStack` between two calls when the response window is
   the point. An `AsEnters` choice fires inline during entry — no extra
   setup needed. For library tutors, seed needles via `pushLibraryCardForTest`
   (which uses `PushBottom` so the "first match" sandbox pick is
   deterministic).

6. **Add the card's oracle-text file.** Every catalogued card needs its
   own generated file, `server/internal/cards/coverage/testdata/oracle/<oracle_id>.json`
   (one per base oracle ID, whether or not the card has an activated
   ability). `TestAbilitiesMatchOracleText` checks ability labels
   against it, and `TestOracleFixtureCoversRegistry` fails in plain PR CI
   when a registered card has no file. Generate only your cards' files,
   never by hand, with the Scryfall dump (from `server/`):
   ```bash
   CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json \
     go test ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent \
     -update-oracle -oracle-ids=<oracle_id>,<oracle_id>
   ```
   Commit only the files for the cards you added. Two card PRs never
   share a file, so they no longer conflict here
   ([#1542](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1542)).
   Dropping `-oracle-ids` regenerates every file and deletes the files
   of cards that left the catalog. That is the nightly's job, and a
   card PR should not do it: if files for other cards change, your dump
   is older or newer than the one the fixture came from.

7. **Verify.** `cd server && go test ./internal/cards/effects/... ./internal/cards/coverage/...` and
   `gofmt -l internal/cards/effects/` should both be clean. The
   `TestNonCatalogSpellStaysSandbox` canary should still pass — it's the
   opt-in invariant.

8. **Manual smoke-test.** `CMDCTRL_DEV_SKIP_DECK_VALIDATION=1 make server-dev`
   plus a small deck (`make dev-skip-validation` target on the top-level
   Makefile) so the library is small enough to find your card quickly.
   Cast it, verify the AUTO badge renders, verify the effect resolves.

### Random effects (#744)

Use `Game.RollDiceForEffect(RandomDraw{Player, Source}, sides, n)`,
`FlipCoinsForEffect(draw, n)` for uncalled faces, and
`ChooseAtRandomForEffect(draw, ids, k)` for selections. These consume
keyed, persisted streams that undo rewinds; do not create a private RNG.
For a won/lost flip, queue `FlipCoinForEffect(CoinFlipSpec{Flipper,
Source, Coins, Question, Then})`. The player calls heads/tails through
`coin_call`; `AllowStop`, `Wins` and `MaxUsefulWins` support chains and
bot stop decisions. Continuations must capture immutable values and use
the `*Game` they receive, so undo operates on the restored game.
`WheneverYouRollDice` triggers once per instruction, including when
another roll trigger is still pending. See ADR 0054 and the random card
tests for compositions; clear any caveat made stale by the new mechanic.

### Adding a mana ability (S15+)

Mana abilities live on the same `Spec{}` struct via the optional
`ManaAbilities []ManaAbility` field. Used by Sol Ring, Arcane Signet,
Birds of Paradise today; basic lands fall back to a synthetic shape
the engine derives from `TypeLine` (no spec needed).

```go
func init() {
    Register(Spec{
        OracleID: "<uuid>",
        Name:     "Sol Ring",
        ManaAbilities: []ManaAbility{
            {
                Cost:     ManaAbilityCost{Tap: true},
                Produced: "{C}{C}",
                Label:    "Add {C}{C}",
            },
        },
    })
}
```

`Produced` is parsed by `game.ParseProducedMana`. Single-color slots
drop straight into the controller's pool when the ability fires;
multi-option slots use **pipe syntax** and queue a `mana_pick`
PendingChoice for the controller to resolve:

- `"{C}{C}"` — Sol Ring: two colorless slots.
- `"{W|U|B|R|G}"` — Birds of Paradise: one any-color slot, picker. The
  picker offers all five colours with the controller's commander
  identity listed first (owner decision, 2026-09-17). Every pipe gets
  that order; never narrow a card whose text just says "any color" or
  names its colours. The identity is read from the player's
  commander(s) in whatever zone they are in (CR 903.4a), not just the
  command zone.
- `"{W|U|B|R|G}"` + `NarrowToCommanderIdentity: true` — Arcane Signet,
  Command Tower, Commander's Sphere, Path of Ancestry: the engine
  intersects the pipe set with the controller's commander identity at
  activation time. Set the flag only when the printed text says "in your
  commander's color identity"; `TestOnlyCommanderIdentityCardsNarrow`
  and the dump-gated `TestNarrowToCommanderIdentityMatchesOracleText`
  hold the catalog to that. **CR 903.4f (#844):** that intersection may
  come back empty — the player has no commander, or a colourless one —
  and then the ability adds no mana at all: no token, no `mana_pick`,
  and the enumerator, the auto-tapper and the client's menu all stop
  offering it (`ManaAbilityAddsNoMana`). A commander with no colour data
  at all is a data gap rather than a colourless commander, and keeps the
  printed colours ([ADR 0040](docs/decisions/0040-mana-pipeline.md)
  #844 amendment).
- `"{W3|U3|B3|R3|G3}"` — Gilded Lotus (#742): ONE pick that adds three
  tokens of the picked colour. Use `OneColorOfAmount(n)`; see "Adding a
  choose-a-color card" below.

Mana abilities can carry cost components beyond `{T}`:

| Cost | Field | Card |
| --- | --- | --- |
| `{T}` | `ManaAbilityCost{Tap: true}` | Sol Ring |
| Sacrifice this | `ManaAbilityCost{Sacrifice: true}` | Lotus Petal, Treasure |
| Sacrifice another permanent | `ManaAbilityCost{SacrificeOther: SacrificeACreature().SacrificeOther}` | Ashnod's Altar, Phyrexian Altar |
| Sacrifice N permanents | `ManaAbilityCost{SacrificeOther: SacrificeN(2, "two creatures", Creature()).SacrificeOther}` | (none yet; #747) |
| Pay N life | `ManaAbilityCost{Life: 1}` | Mana Confluence |
| A mana cost | `ManaAbilityCost{Mana: "{1}"}` | the Signet cycle |
| Remove N counters | `ManaAbilityCost{RemoveCounters: RemoveCountersFromThis("charge", 1).RemoveCounters}` | Vivid Creek, Ramos |
| Remove any number of counters | `ManaAbilityCost{RemoveCounters: RemoveCountersXFromThis("storage", 0).RemoveCounters}` | Mage-Ring Network |
| Discard a card | `ManaAbilityCost{DiscardCards: DiscardACard().DiscardCards}` | Skirge Familiar |
| Exile a card from your hand | `ManaAbilityCost{ExileCards: ExileACardFromHand()}` | Cadaverous Bloom (#1283) |
| Exile this card from your hand | `ExileFromHandForMana("{R}")` (zone + cost together) | the Spirit Guides (#1228) |

"Exile a card from your hand" is NOT a discard with a different
destination: the card leaves through the one exit primitive, fires no
`EventDiscardCard` and is invisible to madness. It rides its own wire
triple (`exile_cost_n` / `exile_cost_label` / `exile_cost_options`,
answered with `exile_ids`) and the client reuses `DiscardCostModal`
with a different verb.

`SacrificeOther` takes a `*game.TargetSpec`, the same shape the CR 602
activated abilities use — build it with the `SacrificeACreature()` /
`SacrificeAPermanent()` helpers and take their `.SacrificeOther` field
rather than writing a spec by hand. The engine filters the candidate
set to the controller's own permanents (CR 701.21a), stamps it onto
`ManaAbilityView.SacrificeOptions`, and the client reuses
`SacrificeCostModal` to pick one. The chosen card comes back in the
`activate_mana_ability` payload as `sacrifice_ids`, and
`ManaAbilityParams.SacrificeIDs` carries it into the engine.

`ActivateManaAbility` validates every component before paying any of
them, so an illegal sacrifice choice leaves the source untapped. Mana
lands in the pool first and the dies-triggers go on the stack after
(CR 605.3b — a mana ability doesn't use the stack, but the sacrifice
still triggers), which is what makes Ashnod's Altar + a drain outlet
work.

Summoning sickness applies to any mana ability with a tap cost on a
creature source (CR 302.6) — Birds of Paradise, Palladium Myr. The
engine enforces it inside `ActivateManaAbility`; specs don't declare
it.

**Which of these the AUTO-TAPPER will plan (#1215).** `Sacrifice: true`
eats the SOURCE, names nothing and asks nothing, so the planner may pay
it — a board of Treasures funds a cast, and the `/autotap` preview, the
strict cast gate and the bot enumerator all say so. It plans such a
source LAST, behind every ordinary source and behind a frozen one:
cracking a Treasure for a generic pip a Mountain could have paid spends
a resource the player never agreed to spend. `SacrificeOther` asks WHICH
permanent dies and stays out of the plan entirely, with the life cost,
the add-a-counter cost, the tap-another cost, the discard cost and the
exile-a-card cost. Since #1242 a sacrifice cost with NO `{T}` (Gold,
Eldrazi Spawn, Eldrazi Scion) is plannable too — the executor cracks it
without tapping it, and a tapped one is still a source — and inside the
sacrifice tier a CREATURE the cost eats comes after a Treasure or a Gold.
What the planner still demands is that the ability cost the source
SOMETHING: a `{T}` or the source itself. Order the abilities so the
cheapest is FIRST; the planner takes one ability per permanent, in order
([ADR 0011](docs/decisions/0011-mana-pool-and-auto-tapper.md)
amendments 2026-09-22 and 2026-09-23).

**Counter costs (#789).** `ManaAbilityCost.RemoveCounters` is the SAME
`*game.CounterRemovalCost` a CR 602 ability's cost carries — one
component with two owners — so build it with the same constructors and
take their `.RemoveCounters` field: `RemoveCountersFromThis(kind, n)`,
`RemoveCountersXFromThis(kind, floor)` for "remove X / any number", and
`RemoveCountersFrom` / `RemoveCountersAmong` for the clauses that name
other permanents. The payment rides `ManaAbilityParams` with exactly the
fields `ActivateAbilityParams` uses (`CounterSourceIDs`,
`CounterCounts`, `CounterKind`), the view ships the same
`counter_cost_*` fields, and the client opens the same
`CounterCostModal`.

An ability whose OUTPUT depends on what the cost paid declares
`ProducedForPaid` instead of `Produced` — "Add {C} for each storage
counter removed this way" is
`ProducedForPaid: ProducedPerCounterRemoved("{C}")`. It is handed the
one `game.PaidCost` record, because by the time the mana is minted the
counters are gone.

The AUTO-TAPPER plans a counter-cost source only when it can both
DECIDE and AFFORD the cost: the counters must come off the source, the
kind and count must be printed, and the permanent must hold enough
right now. A Vivid land out of charge counters is not a mana source,
and a variable or any-kind cost is a decision the planner never makes.
Order the abilities so the free one is FIRST — the planner takes one
ability per permanent, in order, which is what keeps a Vivid land's
charge counters for a deliberate click.

**Reading the mana that paid (#761).** A spell that counts the mana
spent on it reads `effects.Context`, beside `PaidAltCost`:
`ctx.ColorsSpentCount()` (converge, CR 702.86),
`SunburstCounters(kind)` in `OnResolve` (sunburst, CR 702.44),
`AdamantSpent(ctx, "R", 3)` (adamant), and `ctx.NoManaSpent()` — or
`NoManaWasSpentToCast(g, spellID)` from a cast trigger — for "if no
mana was spent to cast it".

A converge or sunburst card must ALSO set `Spec.WantsDistinctColors`,
which makes the cast gate pay the generic half of the cost with colours
it has not spent yet. Without it the payment is colourless-first and the
card converges for less than the board allowed. Adamant deliberately
does not set it.

One rule covers every reader, and no card has to restate it: a payment
the engine WAIVED — permissive mode (the human default) or a
strict-mode override — answers "unknown", and unknown is always the
weaker-than-printed answer. Converge counts no colours, adamant does not
turn on, and "if no mana was spent" is false. Say so in a caveat, as
Painful Truths and Vexing Bauble do.

**Mana that does something when it's spent (#1547)** — "and that
spell can't be countered", "if that mana is spent on a creature
spell, it gains haste", "when that mana is spent to cast …, copy
that spell" — goes in `ManaAbility.SpendRiders`, built with the
constructors in
[mana_spend_rider.go](server/internal/cards/effects/mana_spend_rider.go):

```go
SpendRiders: []game.ManaSpendRider{SpentSpellCantBeCountered(ManaRestrictCast)},         // Cavern of Souls
SpendRiders: []game.ManaSpendRider{SpentCreatureGainsHaste()},                           // Hall of the Bandit Lord
SpendRiders: []game.ManaSpendRider{WhenManaSpent("Pyromancer's Goggles", game.ManaSpendTrigger{
    Label: "Pyromancer's Goggles — copy that spell", Effect: copyTheSpellYouJustCast,
}, ManaRestrictCast, ManaRestrictColor("R"), ManaRestrictAnyType("Instant", "Sorcery"))},
```

The trailing tags are the rider's filter — whether it FIRES — and are
not a spend restriction: Hall's {C} still pays for anything. A card
whose mana is also restricted (Cavern) declares both. Riders fire only
on a recorded payment, so every rider card carries the strict-mana
caveat. See ADR 0040's 2026-09-24 amendment.

For non-mana, non-static activated abilities (planeswalker +1/-1,
equip, cycling, etc.), wait — see the deferral list below.

### Adding a static ability (S16+)

Static abilities (anthems, type-changers, keyword grants, CDAs) live
on `Spec.Static []game.StaticAbility`. The layer engine recomputes
from scratch on every relevant event (battlefield zone change,
counter change); the wire-side `power` / `toughness` / `type_line` /
`abilities` fields reflect the post-layer effective characteristics.

```go
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
    Register(Spec{
        OracleID: "<uuid>",
        Name:     "Glorious Anthem",
        Static: []game.StaticAbility{
            {
                Layer:    game.Layer7PT,
                SubLayer: game.SubLayer7C_Modify,
                AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
                    return target.IsCreature() && target.Controller == source.Controller
                },
                Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
                    c.Power++
                    c.Toughness++
                },
            },
        },
    })
}
```

**Layer / SubLayer choices** (CR 613):

| What you're doing | Layer | SubLayer |
|---|---|---|
| Add a creature type / artifact / enchantment | `Layer4Type` | (ignored) |
| Grant a keyword (flying, trample, etc.) | `Layer6Ability` | (ignored) |
| **Remove** all abilities (Darksteel Mutation) | `Layer6Ability` + `RemovesAbilities: true` | (ignored) |
| Set P/T to a specific value (Tarmogoyf-style CDA) | `Layer7PT` | `SubLayer7A_CDA` |
| Modify P/T (+1/+1 anthem) | `Layer7PT` | `SubLayer7C_Modify` |
| +1/+1 / -1/-1 counter math | (don't — counter math stays in `CurrentPower`) | — |

**`AppliesTo` patterns:**
- "Creatures you control" — `target.IsCreature() && target.Controller == source.Controller`
- "OTHER X you control" — add `target.InstanceID != source.InstanceID`
- Has subtype X — read `target.Effective().Subtypes` (so type-add effects compose)
- Self-only (CDA) — `target.InstanceID == source.InstanceID`

**`Apply` patterns:**
- Anthem +1/+1 — `c.Power++; c.Toughness++`
- Type-add — append to `c.Types` after checking idempotency
- Keyword grant — `c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)`, which dedupes a redundant keyword and keeps every instance of a cumulative one (toxic, #748)
- CDA P/T — `c.Power = computed; c.Toughness = computed + 1`

**Tests** — see [anthem_test.go](server/internal/cards/effects/anthem_test.go) and [tarmogoyf_test.go](server/internal/cards/effects/tarmogoyf_test.go) for the layer-aware pattern. Use `pushBattlefieldCardWithTimestamp` (fires `EventZoneMove` so the listener stamps `EnteredBattlefieldAt` + bumps `layerVersion`); read effective characteristics via `effectivePower` / `effectiveToughness` / `effectiveTypes` / `effectiveAbilities` helpers.

**Don't bypass the printed/effective split:** if an effect needs to read another card's characteristic, use `target.Effective()` not `target.Power` / `target.TypeLine`. Reading printed values inside `AppliesTo` or `Apply` is a layer-ordering bug waiting to happen.

**A 7a or 7b effect makes the body REAL, and CR 704.5f then applies.** The importer writes `Toughness: 0` for a printed `*` (`strconv.Atoi("*")` fails), and the toughness state-based action skips a creature whose toughness the engine does not know. A 7a CDA or a 7b set is the engine knowing it: the pass stamps `Characteristic.PTDefined`, `Card.ToughnessIsKnown` reads it, and a Lord of Extinction with every graveyard empty dies like the 0/0 it is (#690). So a CDA you code must compute the number the card prints, 0 included — do not floor it at 1 to keep the creature alive. Everything else about the skip, including why a printed 0/0 with a printing behind it dies and a 0/0 token template does not, is on `Card.ToughnessIsKnown` and in [ADR 0007 §7](docs/decisions/0007-stack-foundation.md).

**Ability REMOVAL is a declaration, not something `Apply` does.** Set `RemovesAbilities: true` (and build it with `effects.LoseAllAbilities(keep…)`); the engine empties `Characteristic.Abilities` and stamps `AbilitiesRemoved` before your `Apply` runs, so `Apply` only has to append the keywords the same effect grants back. Clearing the slice by hand removes the keyword badges and leaves every catalogued activated, triggered, mana, static and replacement ability working underneath them, because those are read through the `Catalog*` hooks at use time — see [ADR 0046](docs/decisions/0046-layer-6-authoritative.md). If you are writing a NEW engine reader of a `Catalog*` hook that answers "what does this permanent do", key it with `game.CatalogAbilityKey`, not `game.CatalogKey`.

**Durations (CR 611.2, S38).** A continuous effect a spell or ability *creates* does not live on the battlefield — it goes in `Game.ScopedEffects` with a `Duration` on it, and the duration is plain data, never a closure. Four kinds, and one function (`durationExpiredLocked` in `server/internal/game/duration.go`) decides when any of them is over:

| Oracle text | Card-side builder | Ends |
|---|---|---|
| "until end of turn" | `DurationUntilEndOfTurn(ctx)` | that turn's cleanup step (CR 514.2) |
| "until your next turn" | `DurationUntilYourNextTurn(ctx, player)` | as that player's next turn begins, before untap — and when a departed player's turn *would have* begun (CR 800.4m) |
| "for as long as ~ remains on the battlefield" / "for as long as you control ~" | `DurationWhileSourceRemains(ctx, src)` / `DurationWhileYouControlSource(ctx, src, p)` | when the condition goes false, checked at the top of every layer pass (CR 611.2b) |
| no duration printed at all | `game.IndefiniteDuration()` | never (CR 611.2a) |

Reach for `BoostUntilEOT` / `GrantKeywordUntilEOT` for the first row and `ScopedEffectFor{Target|Match, Mods, Duration, Label}` for everything else — a DATA record over the closed `game.*Mod` vocabulary (`SetBasePTMods`, `AddSubtypesMod`, `RemoveTypesMod`, `SetControllerMod`, … in `game/scoped_effects.go`), which the snapshot carries, so a table holding one is still a restore point ([ADR 0041](docs/decisions/0041-game-persistence.md) phase 3, #1497). Every other until-end-of-turn builder (`RestrictUntilEOT`, `GrantAllCreatureTypesUntilEOT`, `BecomeCreatureUntilEOT` / crew) and prowess write the same record. So does a granted ABILITY for a duration, `GrantAbilitiesFor` (the `grantAbilities` mod; see "Granting an ability to another permanent" below). The closure-taking registry (`StaticForDuration`, `StaticUntilEOT`, `RegisterScopedStaticForEffect`, `Game.ScopedStatics`) was deleted in tier 3a, so nothing accepts a closure for one any more; an effect none of the mods can say is a new mod kind, not a closure. There is deliberately no `StaticUntilYourNextTurn` wrapper. A one-shot continuous effect from a resolving spell must pin its affected set at resolution (CR 611.2c): `ScopedEffectFor` does that itself — its `Match` is resolved once, into `(InstanceID, EnteredBattlefieldAt)` pairs, so a permanent flickered in response is correctly a new object (CR 400.7). An indefinite effect pinned to its object survives that object phasing out and in (CR 702.26d); a "for as long as" duration that tracks a source ends when the source phases out (CR 702.26f). The two "for as long as" builders return `(Duration, bool)` and the bool is load-bearing: CR 611.2b says an effect whose condition is already false as it would begin never begins, so register nothing. See [ADR 0063](docs/decisions/0063-durations-and-control.md) and [ADR 0035](docs/decisions/0035-until-end-of-turn-effects.md).

**Control from effects (CR 613.1b, CR 701.12, S38).** "Gain control of target permanent" is `GainControl{Target, Controller, Duration, Label}` and "exchange control" is `ExchangeControl{A, B}`. Both are layer-2 scoped effects — data records with one `setController` mod (#1497) — in the same bucket Mind Control's Aura uses, which is what makes control revert by itself (`Card.BaseController`) and makes two control effects sort by timestamp (CR 613.7) with no card-side work. Do NOT write `Card.Controller`. Three things ride along and are why the printed cards look the way they do: the permanent leaves combat (CR 506.4, declaration and announcement both), it is summoning-sick under its new controller however long it has been in play (CR 302.6 — which is why Act of Treason also grants haste), and ownership never changes (CR 108.3). An exchange is ONE effect: both objects are checked before either half is registered and the two halves share a timestamp, so it fails whole (CR 701.12b). `Controller` defaults to the effect's controller; pass it explicitly for "target opponent gains control of ~" — that card still waits on the choose-a-player prompt, not on this primitive.

### Granting an ability to another permanent (ADR 0093, #754)

"Creatures you control have '{T}: Add one mana of any color.'" is a
layer-6 grant of a catalog BUNDLE, never a closure on the recipient.
Declare the bundle in `Spec.Grants` (the #665 `AbilityGrant`, now with
`Mana` and a required `Text`) and name it from a static built with
`GrantAbilities(appliesTo, key)`:

```go
Grants: []AbilityGrant{{
    Key:  "cryptolith-rite/any-color",
    Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}},
    Text: "{T}: Add one mana of any color.",
}},
Static: []game.StaticAbility{GrantAbilities(creaturesYouControl, "cryptolith-rite/any-color")},
```

The engine does the rest: the grant lands on the recipient's
`Characteristic.GrantedAbilities` in the static's timestamp slot, the
ability readers return it after the recipient's own abilities with a
`grant:` ref, and "this creature" is the recipient everywhere (its
controller activates it, `{T}` taps it, CR 302.6 reads it). A removal
on the recipient takes an earlier grant and not a later one (CR
613.6); a removal on the GRANTOR takes the grant from everything
(CR 613.8a). A layer-6 grant is not copied (CR 707.2). `Register`
refuses a bundle ability with `ActiveWhen` (gate the grantor's static
instead) or a non-battlefield zone, and `TestEveryGrantKeyResolves`
refuses a grant naming an unregistered bundle or a bundle with a
`Static` slot.

The other constructors: `TribalAbilityGrant(TribeFilter{…}, key)` for
"All Slivers have …" / "Sliver creatures you control have …", and
`GrantAbilitiesToAttached(key)` for "Equipped creature has …" /
"Enchanted land has …". `AnyColorManaGrant(key)` and
`TapForManaGrant(key, produced, label, text)` build the common mana
bundles. The auto-tapper plans every acceptable mana ability of a
permanent (not only the first) and keeps a CREATURE's granted mana for
last, so an auto-paid cast does not tap your attackers; a land's
granted mana is an ordinary source. See `cryptolith_rite.go`,
`chromatic_lantern.go`, `necrotic_sliver.go` and `squirrel_nest.go`.

A granted TRIGGER goes in the bundle's `Triggered` slot, written with the
ordinary trigger shapes, and `source` in its `AppliesTo` / `Build` is the
HOST — so "this creature" is `ev.CardID == source.InstanceID`
(`ThisBecameTapped`, `ThisDied`), "you" is the host's controller, and the
trigger goes on the stack under the host's controller. "Only once each
turn" is `b11TriggeredThisTurn(g, source.InstanceID, label)` with a FIXED
label (the tally is per object and per label). A granted dies trigger
fires from last-known information, the ETB harvest sees a grant on the
creature that is entering, and a later ability removal on the host takes
the grant. Never write a granted trigger on the GRANTOR watching the
recipients: the controller, the removal and the LKI all come out wrong.
See `dionus_elvish_archdruid.go`, `agent_of_the_iron_throne.go` and
`thornbite_staff.go`.

A grant from a RESOLVING spell or ability ("until end of turn, target
creature gains '…'", Urza's Saga's "this Saga gains '…'") is the same
bundle, given with `GrantAbilitiesFor` (ADR 0093 PR 4, #1584):

```go
Grants: []AbilityGrant{{Key: feignDeathReturn, Triggered: …, Text: "When this creature dies, …"}},
OnResolve: func(_ *game.StackItem, ctx *Context) error {
    return GrantAbilitiesFor{Target: t, Keys: []string{feignDeathReturn}, Label: "Feign Death — …"}.Apply(ctx)
},
// "gets +2/+0 and gains …" is ONE effect: Also: []game.Mod{game.ModifyPTMod(2, 0)}
// no stated duration (CR 611.2a): Duration: g.PinnedTo(game.IndefiniteDuration(), id)
```

It registers ADR 0041 phase 3's ScopedEffect record with a
`grantAbilities` mod (`game.GrantAbilitiesMod`), which the layer pass
turns into the same layer-6 declaration a static makes — so every
reader, ref, removal and copy rule is the static grant's. It is data, so
the table stays a restore point. A zero `Duration` is "until end of
turn"; the affected set is pinned at resolution (CR 611.2c), so a
creature that dies and returns is a new object without the grant. Keys
must name a registered bundle with no `Static` slot: `Apply` refuses
one at resolution, `TestEveryDurationGrantKeyResolves` scans the
catalog's source for literal and constant keys, and a restore point
naming an unregistered bundle is refused with `ErrUnknownEffectKey`. A
"return it to the battlefield tapped [with a counter]" dies trigger is
`returnThisCreatureFromGraveyard`. See `feign_death.go`,
`fake_your_own_death.go`, `retraction_helix.go` and `urzas_saga.go`.
Still no shape: a duration that lasts "for as long as it has a <kind>
counter on it" (Ultima, Origin of Oblivion).

### Adding a replacement effect (S17+)

Replacement effects ("enters tapped", "if that would place counters,
place twice that many instead", "if a player would draw a card, that
player mills instead") live on the same `Spec{}` struct via the
optional `Replacements []game.ReplacementEffect` field. Used today by
Doubling Season, Hardened Scales, Kismet, Stasis, Gemstone Mine,
Fog, Stone of Erech.

**A discard goes through the exit primitive** (#853). Every discard
site — the CR 514.1 cleanup discard, the effect-discard continuation,
the revealed-hand leg, the random discard and the discard component of
an additional cost — shares `discardCardsLocked` (`server/internal/game/discard.go`),
which routes each card through `routeCardToZoneLocked` like every other
exit. So the CR 614 window opens on a discard, a discarded commander
gets the CR 903.9 offer, and a discard can PAUSE: `...ForEffect` returns
with the card still in hand and the rest of the batch (and the prompt's
`Then`) owed until the owner answers. The COST site is the exception —
CR 601.2h pays a spell's costs as one indivisible step, so it sets
`zoneRoute.MustSettleNow` and settles without asking, which means a
commander pitched to a cost goes to the graveyard.

**A tuck can pause, so read what LANDED** (#783). A library is a
CR 903.9 destination like every other, so "put it into its owner's
library" opens the window and can stop to ask a commander's owner about
the command zone. If your card has anything to do AFTER the tuck —
shuffle, reveal, scry, ask the next question, read the card's zone —
hand it over as a continuation (`TuckToLibraryThenForEffect`, or
`TuckCardsToLibraryThenForEffect` for a batch, which reports the cards
that really reached a library). `TuckToLibraryForEffect` stays
fire-and-forget and is right only when the tuck is the LAST instruction
on the card. A printed position ("on the bottom", "third from the top")
goes in `game.TuckOptions` so it rides the route and survives the
prompt — never reposition the card yourself on the next line. See
[ADR 0013 §5n](docs/decisions/0013-replacement-effects.md).

**A discard is its own replaceable event** (#650,
[ADR 0061](docs/decisions/0061-token-creation-and-discard-are-replaceable-events.md)).
The route opens `RepEventDiscard`, not a plain move, because what a
discard replacement watches for is the discard — so declare
`Watches: []game.EventKind{game.EventDiscardCard}` and check
`ev.Kind == game.RepEventDiscard`. It carries `DiscardPlayer`,
`DiscardCause` (`"effect"` / `"cost"` / `"cleanup"`) and the causing
`Source`, alongside the move payload your `Replace` rewrites
(`ev.NewZone`, `ev.NewZoneOwner`).

The CAUSE is the clause the rules draw, not "voluntary": an effect's
instruction, a cost (CR 601.2h, 602.2b, and the CR 118.12 "unless you
discard" branch), or the cleanup step's turn-based action (CR 514.1).
Library of Leng replaces `DiscardCauseEffect` only; madness replaces
every cause; the Obstinate Baloth shape reads the cause plus the
controller of `Source`. Build one with `DiscardBecomes{…}.Build()`
(`cards/effects/discard_replacements.go`) rather than by hand.

`EventDiscardCard` still fires wherever the card ends up (CR 701.8a
defines a discard by the move OUT of the hand), so a discard your
replacement redirects is still a discard for Megrim and friends, and a
discarded commander still gets the CR 903.9 offer. A COST discard
settles without asking, so an `Optional` replacement on one is skipped
un-applied — which is also the right answer, since costs are not
effects.

**Madness is one string** (#657, CR 702.35). Do not write either half
of it on a card:

```go
Register(Spec{
    Name:    "Fiery Temper",
    Madness: "{R}",       // and nothing else about the keyword
    ...
})
```

`buildDef` grows the CR 702.35a discard replacement
(`game.MadnessReplacement`) and the exile-zone trigger that offers the
cast (`game.MadnessTrigger`) from that one field, and appends them to
whatever the card declares itself — Big Game Hunter keeps its own ETB.
The cast the trigger offers is a per-instance `game.CastPermission`
priced at the madness cost, keyed `"madness"` and `TimingFlash`
(CR 608.2g), and declining puts the card into its owner's graveyard
(CR 702.35b). `Register` refuses an unparseable cost at boot. The
engine side, and the two declared simplifications it shares with
cascade, are in `server/internal/game/madness.go`.

Unlike static abilities, replacements fire **before** the event
happens — the pipeline constructs a `game.ReplacementEvent`, the
engine offers each applicable replacement a chance to mutate or
cancel it, then the underlying mutation runs (or is skipped, if
canceled). CR 616.1 iterative apply-loop, CR 614.5 once-per-event
tracking, and CR 616 affected-player-chooses-order are enforced
centrally.

```go
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
    Register(Spec{
        OracleID: "<uuid>",
        Name:     "Doubling Season",
        Replacements: []game.ReplacementEffect{
            {
                Watches: []game.EventKind{game.EventCounterPlaced},
                AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
                    if ev.Kind != game.RepEventCounter { return false }
                    target, ok := g.LookupCardForEffect(ev.CounterTarget)
                    if !ok { return false }
                    return target.Controller == src.Controller
                },
                Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
                    ev.CounterDelta *= 2
                    return nil
                },
                Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
                    return src.Controller
                },
                Label: "Doubling Season: double counters",
            },
        },
    })
}
```

**Kind picker** (CR 614):

| What the replacement watches | `ReplacementEventKind` | Relevant fields |
|---|---|---|
| Counter placement (+1/+1, loyalty, …) | `RepEventCounter` | `CounterTarget`, `CounterName`, `CounterDelta` |
| Zone motion (ETB, LTB, draw-as-move) | `RepEventMove` | `CardID`, `OldZone`, `NewZone`, `NewZoneOwner`, `EntersTapped`, `EntersWithCounters` |
| Card draw | `RepEventDraw` | `DrawPlayer`, `DrawCount` |
| Life total change | `RepEventLife` | `LifePlayer`, `LifeDelta` |
| Damage (combat and direct) | `RepEventDamage` | `DamageSource`, `DamageTarget`, `DamageAmount`, `IsCombatDamage` |
| Token creation (CR 701.7b) | `RepEventCreateTokens` | `TokenController`, `TokenGroups`, `TokenAttacking` |
| Discard (CR 701.8) | `RepEventDiscard` | `DiscardPlayer`, `DiscardCause`, `CardID`, `NewZone`, `NewZoneOwner` |
| Keyword action with a count — proliferate (CR 701.34), scry (CR 701.22), surveil (CR 701.25) | `RepEventKeywordAction` | `KeywordAction`, `KeywordActionCount`, `Actor`, `Source` |
| Mill amount (CR 701.17a) | `RepEventMill` | `MillPlayer`, `MillCount` |
| Mana produced (CR 106.12b) | `RepEventProduceMana` | `ManaPlayer`, `ManaSource`, `ManaColors`, `ManaFromTap` |
| Step entry (skip-step) | `RepEventStepTransition` | `StepTransitionStep`, `StepTransitionSeat` |

**Adding a kind to that table is five switches, not one** (#982). A
`ReplacementEventKind` has to be named in `eventKindMatches` (the watch
key), `affectedPlayerForEvent` (who CR 616.1 asks to order the
window), `applyResolvedReplacementEventLocked` (the resume), and both
terminal outcomes — `finishSettledReplacementLocked` for a cancelled
event and `abandonZoneRouteLocked` for one whose prompt is taken away.
Every one of those failures is silent, so
`TestEveryReplacementEventKindIsSwitchedOn`
([replacement_kind_gate_test.go](server/internal/game/replacement_kind_gate_test.go))
reads the switches out of the source and fails until each has an arm.
Naming a kind that owes NOTHING is a written arm, not an omission.

**Adding a plain `game.EventKind` owes the public log an answer** (#984).
`projectEvent` ([log.go](server/internal/protocol/log.go)) is the one
switch that decides whether the table is told about an event, and its
default arm is a silence — so a kind that should have produced a line
looks exactly like a kind that should not.
`TestEveryEventKindIsNarratedOrDeliberatelySilent`
([log_event_kind_gate_test.go](server/internal/protocol/log_event_kind_gate_test.go))
reads every declared kind and fails until each one either has an arm in
`projectEvent` or an entry in that file's `silentEventKinds` table with
a written reason. The two are exclusive and the test says so, so a kind
that grows a line has to lose its excuse. It was written because
`EventColorChosen`, `EventCreatureTypeChosen` and `EventPlayerChosen`
each shipped with a card, a view field and a test, and none of the three
ever reached the log.

Reading that table back is what #1021 did: six of its rows were gaps
rather than decisions, and a control change, a special action, a
cycling, a counter landing, a scry or surveil, and a Saga chapter or
Class level are lines now. Two rules came out of it and hold for the
next arm. **A value that identifies the card is redacted with the
card's name** — `choice`, `label` and a counter / chapter / level
`amount` all go when `redactLogForViewer` drops the name, because
`redactCardForViewer` already strips the same facts off the CardView
and a line that kept them would hand them straight back. And **a kind
whose changes are already a line somewhere else says so in a
predicate, not in a second table**: `counterKindIsNarrated`
([log.go](server/internal/protocol/log.go)) is why a loyalty tick and
a lore counter produce nothing, and it is an allowlist of SILENCES so
that a counter kind nobody has thought of yet gets a line rather than
a hole.

**`AppliesTo` patterns:**
- "Counters go on a creature you control" — `target.Controller == src.Controller && target.IsCreature()`
- "When a permanent enters the battlefield" — `ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield`
- Self-replacement (Gemstone Mine's three mining counters on its own ETB; every "this land enters tapped") — `ev.CardID == src.InstanceID`. This works even though the entering card is not on the battlefield yet: `gatherActiveReplacementsLocked` has a dedicated block for a card that is NOT on the battlefield, which passes the entering card itself as `src` ([replacements.go](server/internal/game/replacements.go), the `!g.Battlefield.Contains(ev.CardID)` branch). Prefer `SelfEntersTapped()` over an `AsEnters` tap — see the "enters tapped" note below.
- Opponents only (Kismet) — `controllerOf(ev.CardID) != src.Controller`

**`Replace` patterns:**
- Counter multiplier — `ev.CounterDelta *= 2` (Doubling Season)
- Counter addition — `ev.CounterDelta += 1` (Hardened Scales)
- Cancel — `ev.Cancel()` (Fog, Stasis)
- Redirect move — `ev.NewZone = game.ZoneExile` plus `ev.NewZoneOwner = uuid.Nil` (Stone of Erech)
- Enters-tapped — `ev.EntersTapped = true` (Kismet)
- Enters-with-counters — `ev.AddCounterAtETB("+1/+1", n)` (Gemstone Mine, Kalonian Hydra). For a count read from the CAST rather than from the board, use the declaration below instead.

**"Enters with N counters" where N comes from the CAST** (#1002,
CR 614.1c). Three families of "enters with counters" live in the
catalog and they are three slots, because they read three different
things:

| What N is | How to declare it |
|---|---|
| a printed number ("enters with three +1/+1 counters") | `Replacements: []game.ReplacementEffect{b10EntersWithCounters(kind, n, label)}` |
| a count off the BOARD ("…for each Zombie card in your graveyard") | `Replacements: []game.ReplacementEffect{b19EntersWithCountersCounted(kind, count, label)}` |
| a fact about the ANNOUNCEMENT (X, times kicked, colours spent) | `EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(kind)}` |

The first two are ordinary CR 614 self-replacements: everything they
need is reachable from `(g, src)` while the entry window is open. The
third is not — a replacement is handed the game, the source and the
event, and none of those carries the resolving stack item, which is
why thirteen X-creatures and sunburst used to put their counters on in
`OnResolve` a beat before the permanent existed and each declared a
caveat saying so. The engine now seeds the clause onto the entry event
from the `StackItem` that is right there
([game/entry_counters.go](server/internal/game/entry_counters.go)), one
line after escape's `applyAltCostEntryCountersLocked`, so a card file
declares arithmetic over `game.CastCounts` — `X`, `Kicked`,
`ColorsSpent` — and nothing else. Constructors:
`XCounters(kind)`, `CountersPerKick(kind, per)`,
`SunburstCounters(kind)` in
[cards/effects/entry_counters.go](server/internal/cards/effects/entry_counters.go).
Never build a `game.EntryCountersFromCast` by hand, for the reason
`mana_spent.go` gives: a card says what the card says and never
reaches for the payment record itself.

A permanent that did not come from a spell — reanimated, put onto the
battlefield, a token — enters with none, because the seeding site is
the spell's entry and nothing else (CR 107.3b).

Whatever seeds them, the settled map is **drained by one helper**,
`(*Game).applyEntryCountersLocked`
([game/entry_counters.go](server/internal/game/entry_counters.go)) —
never with a bare `range` over `ev.EntersWithCounters` (#1010). Each
kind opens its own `RepEventCounter` window, so with two KINDS on one
entry the drain order is the order those windows open, the order a
CR 616 prompt inside them is asked in, and the order the events land in
the log; Go randomises map iteration, so a bare range made all three
differ run to run. The order is canonical — counter name, ascending —
and is not a CR 616 choice: the window that produced the map has
already closed, and two kinds on one entry are one settled event with
two components.

**A token creation is a replaceable event** (#762,
[ADR 0061](docs/decisions/0061-token-creation-and-discard-are-replaceable-events.md)).
`RepEventCreateTokens` is opened once per creation **instruction**
(CR 701.7b), so "create two Treasures" is one event a doubler turns
into four Treasures. It carries GROUPS — a template, a count and the
creation's entry clause per KIND — because Academy Manufactor changes
*which* tokens are made and a bare count could not say so. Write a
doubler as `TokensDoubled(label)` (or `AnyPlayersTokensDoubled` for
Primal Vigor's symmetrical one); write anything else with
`ev.MultiplyTokens(n)`, `ev.ReplaceTokenKindsWhere(pred, templates…)`
and `ev.TokenTemplatesMatch(pred)` — never by reading `ev.TokenGroups`
directly.

Once that window settles, every token it makes goes through
`enterBattlefieldThroughPipelineLocked` — **the same entry primitive a
library search, an exile return and a reanimation use** (#478) — as a
`RepEventMove` into `ZoneBattlefield` with an empty `OldZone` (a token
comes from no zone, CR 111.1). So an
enters-tapped or enters-with-counters replacement you write for cards
covers tokens for free, and `fireETBHookLocked` runs for a token copy.
A creation CAN PAUSE — two different effects in the window is a CR 616
ordering prompt — and so can a single token's entry, so
`CreateTokensForEffect`'s returned IDs are EMPTY when it paused. If
your card's sentence continues past the tokens ("create a Treasure,
then sacrifice it"), hand that over as
`CreateTokensThenForEffect(spec, then)` rather than reading the slice
on the next line.

**A keyword action with a count is a replaceable event** (#976,
[ADR 0013 §5s](docs/decisions/0013-replacement-effects.md)).
"If you would proliferate, proliferate twice instead" (Tekuthal,
Inquiry Dominus) and "if you would scry, scry that many plus one
instead" replace the ACTION, not the counters it places or the cards
it looks at, so `RepEventKeywordAction` is opened once per
INSTRUCTION at the one entry point of each action — exactly the way
`RepEventCreateTokens` is opened once per creation instruction.

Write one with `KeywordActionBecomes(action, count, label)`
(`cards/effects/keyword_action_replacements.go`), or one of its named
wrappers — `ProliferateTwice(label)`, `ScryPlusOne(label)`,
`SurveilPlusOne(label)`. The
`count` function is applied to the count the EVENT carries, not the
printed one, which is what makes two of them compose the way CR 616.1
composes them. Declare nothing else: the helper writes
`Watches: []game.EventKind{game.EventKeywordAction}` (an
engine-internal watch sentinel, like `EventStepTransition` — nothing
logs it), narrows on the action, and scopes itself to the source's
controller, because every printed one says "if YOU would".

What the count MEANS is per action, and getting it wrong is the one
way to write this badly: for **proliferate** it is the number of
TIMES the whole action is taken (base 1 — the choice and all of it,
twice), for **scry** and **surveil** the number of CARDS (base N).
"Look at the top N cards of your library, then put them back in any
order" is NOT a keyword action and opens no window.

The window can PAUSE — a doubler and a "plus one" is a CR 616
ordering prompt, and ×2 then +1 differs from +1 then ×2 — so a
paused proliferate has placed no counters and a paused scry has
queued no prompt when the entry point returns;
`ScryThenForEffect`'s returned count is 0, the contract
`CreateTokensForEffect`'s empty ID slice already carries. Anything
after "then" still goes in the continuation (`Scry{Then: …}`), and it
runs on every terminal outcome, a cancelled action included: "scry 2,
then draw a card" draws whether or not the scry happened.

**The mill AMOUNT is a replaceable quantity too** (#569,
[ADR 0013 §5u](docs/decisions/0013-replacement-effects.md)). "If an
opponent would mill one or more cards, they mill twice that many cards
instead" (Bruvac the Grandiloquent) replaces the NUMBER, once, before
anything leaves the library, so `RepEventMill` is opened once per mill
INSTRUCTION — the same shape as the creation and the keyword action.

Do not confuse it with the per-card window, which is older and needs
nothing from you: every milled card already goes through the shared
exit primitive, so "if a card would be put into a graveyard from
anywhere, exile it instead" and CR 903.9 both see each of them. That
one is `graveyard_replacements.go`; this one is
`cards/effects/mill_replacements.go` —
`MillBecomes{Count, Scope, Label}`, with `OpponentsMillTwice(label)`
and `OpponentsMillPlus(n, label)` as the named wrappers.

The window opens only for something the rules call a mill: a
GRAVEYARD destination (CR 701.17a defines the keyword action by where
the cards go, so `MillToZone{To: game.ZoneExile}` is not a mill and
opens none) and a POSITIVE count (an unbounded `until` run names no
number to double). The count a replacement sees is the one the
INSTRUCTION named, not what the library can supply — CR 701.17b's
"mill as many as possible" clamp happens afterwards. A surveil's
graveyard leg is NOT a mill (CR 701.14a) and no mill replacement
touches it.

It can PAUSE, before any card is chosen, so `MillToZoneForEffect`'s
slice is empty when it did. If your card reads what was milled, use
`MillToZone{…, Then: …}` / `g.MillToZoneThenForEffect` — which you
should be doing anyway, for #893's reason.

**The MANA PRODUCED is a replaceable quantity too** (#1222,
[ADR 0013 §5ab](docs/decisions/0013-replacement-effects.md)). "If you
tap a permanent for mana, it produces twice as much of that mana
instead" (Mana Reflection, Nyxbloom Ancient) replaces the AMOUNT, so
`RepEventProduceMana` is opened once per production and before any of
it is in the pool. The family is
`cards/effects/mana_replacements.go` — `ManaProducedBecomes{Times,
Scope, Label}`, with `YouTapForTwiceAsMuchMana(label)` and
`YouTapForThriceAsMuchMana(label)` as the named wrappers.

Three things about it differ from the other amount events and all three
are the printed cards' doing:

- **It carries COLOURS, not a count.** `ev.ManaColors` is one entry per
  mana, because "twice as much of THAT mana" names the mana as well as
  the amount. Write `ev.MultiplyMana(n)` and nothing else — assigning
  `ManaColors` yourself could change the COLOURS, which CR 106.12b does
  not license.
- **`ev.ManaFromTap` is the printed condition**, not a convenience. Both
  cards say "if you TAP a permanent for mana" (CR 106.12a), so a spell's
  "Add {B}{B}{B}", a mana ability with no `{T}` and a triggered mana
  ability's own output are all productions and none of them is doubled.
  A card that really is symmetrical leaves the check out.
- **It never pauses.** CR 605.3b makes activating a mana ability one
  indivisible step with no priority window inside it, so every event of
  this kind sets `mustSettleNow` and the CR 616 ordering prompt is never
  asked. An `Optional` mana-production replacement would therefore be
  skipped un-applied; if you ever print one, that is the conversation to
  have first.

A pipe slot is replaced at the PICK, not at the activation — a Birds of
Paradise under Mana Reflection is ONE choice minting two of the chosen
colour — because that is the only moment the colour exists. The
auto-tapper plans with the replaced amount through the same predicate
(`producedManaPreviewLocked`), so a Mana-Reflected land really does pay
for two pips.

**The DRAW AMOUNT is a replaceable quantity too** (#1222, same section).
`RepEventDraw` has existed since S17; what #1222 added is
`ev.DrawCount`, whose base is always ONE because CR 121.2 makes "draw
three cards" three individual card draws. The family is
`cards/effects/draw_replacements.go` — `DrawBecomes{Count, Scope,
ExceptInOwnDrawStep, Label}`, with `YouDrawTwiceInstead(label)` and
`YouDrawTwiceInsteadExceptTheFirst(label)`.

The N cards a settled count asks for are drawn one at a time, so every
per-card payoff still fires per card — but they are ONE event, and the
window is not re-opened for them. That is what makes two Thought
Reflections draw FOUR rather than three, and it is why a cancel-style
draw replacement sharing the window takes the whole doubled draw rather
than one card of it (the declared simplification; no dredge card is
catalogued).

**"Except the first one you draw in each of your draw steps"** has no
per-draw-step tally behind it. Both cards that print it — Notion Thief
and Alhammarret's Archive — read it as "except ANY draw in that player's
own draw step" and carry the same `caveats` line. `drawnInOwnDrawStep`
is the one copy of the predicate; use it rather than writing the check
again.

**Two copies of your card will not prompt.** When every replacement
applicable to one event is the *same* declared effect — same catalog
entry, same slot in its `Replacements` slice, same controller — the
engine applies them all inline instead of asking the affected player
to order them, because every order is the same modification N times
(two Doubling Seasons are ×4, two Rhox Faithmenders are ×4,
[#792](https://github.com/krakenhavoc/cmd_and_ctrl/issues/792)). A
window with any *distinct* effect in it still prompts with everything
listed. Nothing to declare — but it does mean one thing is now on you:
**if your `Replace` writes its own source into the event** ("that
damage is dealt to *this* creature instead", "put the counter on
*this* creature instead"), two copies of your card are *not*
interchangeable and collapsing them would be wrong. No catalog card
does this yet; if yours is the first, say so on the PR rather than
shipping it quietly — the fix is a declared flag in the `PureCancel`
mould. See [ADR 0013 §5a](docs/decisions/0013-replacement-effects.md).

**A `may` is always offered, however many effects share the window.**
`Optional: true` queues a yes/no prompt for the effect's
controller before `Replace` runs, and that is now true on the
multi-effect paths too: an effect ordered alongside others by a CR 616
prompt pauses for its own question when the chain reaches it
([#847](https://github.com/krakenhavoc/cmd_and_ctrl/issues/847)), and
a window nobody is left to order — or one that cannot pause at all,
like a cost — skips it un-applied rather than firing it. So don't write
a `Replace` that assumes it only ever runs after a "yes"; it never runs
otherwise, but it may never run at all. See
[ADR 0013 §5h](docs/decisions/0013-replacement-effects.md).

**Tests** — see `server/internal/cards/effects/doubling_season_test.go` for the CR 616 ordering pattern (Doubling Season + Hardened Scales → the affected player picks order → `[HS, DS]` yields 4 counters, `[DS, HS]` yields 3). Use `pushBattlefieldCardWithTimestamp` to get the source on the battlefield + the listener to stamp `EnteredBattlefieldAt`; trigger the event with the public mutation (`AddCounter`, `DrawCard`, etc.) and assert on the resulting state plus any queued `PendingChoice`.

**Don't use the replacement pipeline when a primitive flag suffices.** "This card does X to a land it fetches" (Cultivate, Path to Exile, Solemn Simulacrum) is a self-contained card behavior, not a general replacement. Declare `TappedOnEntry: true` on the `SearchLibrary` primitive rather than a full `ReplacementEffect`. The generic pipeline is for effects that watch *other* cards' events.

> **History.** That `TappedOnEntry` flag used to be the *only* thing
> standing in for the pipeline on the search path, which is how a fetched
> fastland entered untapped
> ([#263](https://github.com/krakenhavoc/cmd_and_ctrl/issues/263),
> **fixed**). The search path — and the reanimation path, which had the
> same hole and was not in the issue — now both run
> `applyReplacementsLocked` before the card leaves its zone, and both
> fire `fireETBHookLocked`.

**An ENTRY can pause too, and the effect that asked for it waits**
(#478). A battlefield entry runs the CR 614 window before the card
leaves its old zone, and that window can stop to ask: a CR 616 ordering
prompt between two enters-tapped effects (Kismet plus Thalia, Heretic
Cathar), a shockland's "you may pay 2 life", Clone's "choose what to
copy", any "may". The library search, the exile return and the
reanimation are `entryResumable` now, so a fetched shockland IS offered
its payment and two replacements on one fetched Guildgate no longer eat
the card. What the effect still owed rides across the pause on
`ReplacementEvent.entryTail` — the library shuffle and
`EventSearchLibrary`, the caller's `Then`, and the CR 400.7 new object
an exile return mints — and the same resume every other paused entry
uses finishes it
([entry_tail.go](server/internal/game/entry_tail.go), [ADR 0013
§5o](docs/decisions/0013-replacement-effects.md)). For your card this
means the line after a fetch, a blink or a reanimation may run one
action later than the call; if you read the permanent's zone, its ID or
what arrived, use the effect's own continuation
(`SearchLibrarySpec.Then`, `ReturnFromExile.Then`) rather than the next
line.

**"Put … onto the battlefield" is one entry, and it can ask too**
(#1322, #1324, #1327; [ADR 0061 amendment
2026-09-23](docs/decisions/0061-token-creation-and-discard-are-replaceable-events.md)).
The hand / library / exile put batch
([entry_batch.go](server/internal/game/entry_batch.go)) runs each card's
window against the pre-entry board, one at a time; a card that asks
something stops the batch there, and the batch lands every card together
once the last answer is in. So a shockland Genesis Wave puts is offered
its life, and a Clone it puts is asked what to copy. Three rules for
card code:

- If your sentence continues past the put ("put the rest on the
  bottom", "if you put a Cave onto the battlefield this way"), use a
  Then door — `PutCardsFromLibraryOntoBattlefieldThenForEffect`,
  `PutFromHandOntoBattlefieldThenForEffect`, or the effects primitives
  `PutFromLibraryOntoBattlefield` / `PutFromHandOntoBattlefield`, which
  already do. The synchronous doors return nothing while a card waits.
- "Put both cards onto the battlefield" from two zones is
  `PutOntoBattlefieldTogetherThenForEffect` with a `game.BatchEntry` per
  card (Sword of Hearth and Home). Two calls in a row are two events, and
  CR 603.6a can tell: a trigger on the first card would not see the
  second.
- "If it entered under your control" after an exile return is
  `ReturnFromExile{…, Then: …}` (Phelia, Exuberant Shepherd); `entered`
  is the new object's ID, `uuid.Nil` when nothing came back.

The one entry that still cannot pause is the sandbox `move_card` verb.

**Regeneration is an engine built-in, not a card's replacement**
(#667, [ADR 0013 §5p](docs/decisions/0013-replacement-effects.md)).
"Regenerate target creature" is `effects.Regenerate{Target}`, and
that is the whole card side: it adds one shield
(`Card.RegenerationShields`, a count, cleared at cleanup and on the
way off the battlefield) and the rule lives in
`regenerationShieldReplacement`
(`server/internal/game/builtin_replacements.go`), which watches the
DESTROY `RepEventMove` and, when it applies, cancels the move, taps
the permanent, removes all damage from it, takes it out of combat and
spends one shield (CR 701.19a). Two shields never prompt — a built-in
is registered once per game, so two of them are one applicable
effect — and a shielded COMMANDER does prompt, because CR 903.9
applies to the same event and CR 616.1 gives its controller the order.

**"It can't be regenerated" is a rider on the destroy, not a keyword**
(CR 701.19c). Write `DestroyTarget{Target: id, CantBeRegenerated:
true}` or `DestroyAllMatching{Match: …, CantBeRegenerated: true}` on
every card whose oracle text prints the clause — Terminate, Mortify,
Putrefy, Pongify, Rapid Hybridization, Snuff Out, Damn, Damnation,
Wrath of God, Winds of Rath, Shatterstorm do — and leave it off the
printings that don't (Day of Judgment, Supreme Verdict, Vanquish the
Horde). The rider rides the route onto the event and gates the
built-in's `AppliesTo`, so an ignored shield is NOT spent
(CR 701.19c). `"regenerate"` is still not a keyword and is not in
`canonicalKeywords`: it is a keyword ACTION, and the closed keyword
list is for keyword abilities.

**What a shield does not stop**, and why each one is a separate
branch rather than one check: a sacrifice (CR 701.21a), a creature at
zero toughness (CR 704.5f), a planeswalker at zero loyalty
(CR 704.5i), a battle at zero defense (CR 704.5v/w), the legend rule,
an illegally attached Aura, an exile, a bounce. All of those take the
same battlefield exit a destruction does, so the exit carries a
declared `Destruction` flag — `destroyRoute` sets it,
`battlefieldExitRoute` does not — and the state-based-action sweep
tags each doomed permanent with the rule that doomed it
(`doomedPermanent`, `server/internal/game/simultaneous.go`).

### Adding a copy effect (S16.5+)

"You may have this creature enter as a copy of X" (Clone, Phyrexian
Metamorph, Spark Double, Sakashima the Impostor) is a replacement
effect with a picker inside it. Cards declare it through the shared
`EntersAsCopyOf` constructor in
[copy_effects.go](server/internal/cards/effects/copy_effects.go)
rather than building a `game.ReplacementEffect` by hand:

```go
Replacements: []game.ReplacementEffect{
    EntersAsCopyOf(
        "Phyrexian Metamorph",
        // what may be copied — evaluated when the prompt is built
        // AND again when the answer arrives.
        func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
            return copyCandidates(g, self, func(c game.Card) bool {
                return c.IsCreature() || c.IsArtifact()
            })
        },
        // the "except" clause — nil for a plain Clone.
        func(ev *game.ReplacementEvent, v *game.PrintedValues, g *game.Game, src *game.Card) {
            v.AddCardType("Artifact")
        },
    ),
},
```

**Where each kind of "except" clause goes:**

| Printed clause | Where it lands |
|---|---|
| "except its name is N" | `v.SetName("N")` |
| "except it's an artifact in addition to its other types" | `v.AddCardType("Artifact")` |
| "except it's legendary in addition to…" | `v.AddSupertype("Legendary")` |
| "except it isn't legendary" | `v.RemoveSupertype("Legendary")` |
| "except it enters with an additional +1/+1 counter" | `ev.AddCounterAtETB("+1/+1", 1)` |
| "except it enters with an additional loyalty counter" | `v.StartingLoyalty++` — NOT `AddCounterAtETB`; the CR 306.5b stamp refuses to run on a walker that already has loyalty counters |
| "except it's an Illusion in addition to its other types" | `v.AddSubtype("Illusion")` (CR 707.9b). An ADD: the copied Bear stays a Bear. The SET form ("except it's a 4/4 black Zombie") is `retypedTypeLine` in `token_copy.go`, not this |
| "except it has '\<ability\>'" | `v.GrantAbility("<card>/<what>")` (CR 707.9a), naming a bundle the card declared in `Spec.Grants` — see below |
| branch on what was copied | `v.HasCardType("Creature")` / `"Planeswalker"` / `v.HasSubtype("Illusion")` |

**Granting an ability (CR 707.9a, #665).** A granted ability is part
of the COPIABLE VALUES — a Clone copying a Phantasmal Image gets the
Image's sacrifice trigger — so it cannot be a closure on the
replacement. Declare it as catalog data on the card that grants it
and name it from the except clause:

```go
const phantasmalImageIllusionGrant = "phantasmal-image/illusion"

Grants: []AbilityGrant{{
    Key:       phantasmalImageIllusionGrant,
    Triggered: []game.TriggeredAbility{ /* …, or Static / Activated */ },
}},
Replacements: []game.ReplacementEffect{
    EntersAsCopyOf("Phantasmal Image", anyCreatureOnBattlefield,
        func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
            v.AddSubtype("Illusion")
            v.GrantAbility(phantasmalImageIllusionGrant)
        }),
},
```

The copy stores only the bundle's KEY, in `PrintedValues`, which is
what makes the grant copiable again and snapshot-safe. `Register`
files the bundle's own `game.CardDef` under `game.GrantKey(Key)` —
the same `defs` map cards use, the shape emblems already take — and
panics at boot on an empty key, a catalog-wide duplicate, or a bundle
with no abilities. Namespace the key with the granting card; it is as
permanent as an oracle ID, because a snapshot carries it.

Nothing else needs changing to make the grant WORK: `CatalogKey`
returns a composite `"<oracle_id>|grant:<name>"` and `catalogDef`
merges, so the harvester, the layer pass, the activation path and the
view all find it through the lookup they already used. Read a catalog
key as an identity rather than a lookup (deriving `EmblemKey`, or
`effects.Lookup` into the Spec registry) and you want
`game.BaseCatalogKey` first. Mana abilities have no grant slot. See
`server/internal/game/copy_grants.go` and the
[ADR 0043 amendment](docs/decisions/0043-copy-effects.md).

**What a copy brings, and what it does not.** `PrintedValues` is the
CR 707.2 copiable-value set: printed name, type line, mana cost,
colours, P/T, keywords, starting loyalty, layout / faces, and the
oracle ID — which is the load-bearing one, because every catalog hook
resolves through `CatalogKey`, so the copy inherits the copied card's
triggers, statics, replacements and abilities for free. It does NOT
bring counters, damage, status, or any other layer's effect. See
[ADR 0043](docs/decisions/0043-copy-effects.md).

**Don't reach for this for a token copy.** "Create a token that's a
copy of target creature" (Follow the Spirit, Kiki-Jiki) is
`CreateTokenCopy` in
[token_copy.go](server/internal/cards/effects/token_copy.go) — a
resolution-time effect that mints a new object, not a replacement of
something's own entry.

**Copying a SPELL is `CopySpell{StackID, Controller, Count,
ChooseNewTargets, Except}`** ([spell_copy.go](server/internal/cards/effects/spell_copy.go)),
CR 707.10 — Reverberate, Twincast, the Chain cycle, Double Major — and
`StackID` may be **this spell**, `ctx.Item.ID`. "That player may copy
this spell and may choose a new target for that copy" (the Chain
cycle) is a spell copying itself from inside its own resolution, which
works because the engine keeps the resolving item's metadata reachable
for the whole occurrence
([resolving_item.go](server/internal/game/resolving_item.go), #920) —
the item's `StackMeta` entry is deleted before `OnResolve` runs, and by
the time the copy question is answered the spell is already in a
graveyard, so the copy is built from last-known information. Set
`Controller` to the player the card says makes the copy (CR 707.10b);
it is **not** the copied spell's controller, and on the Chain cycle
that asymmetry is the card.

`Except` is the copy's "except …" clause (CR 707.10a) and takes the
same `*game.PrintedValues` an entering permanent's except clause does:
Double Major's "except it isn't legendary if the spell is legendary"
is `v.RemoveSupertype("Legendary")`, with no condition, because
removing an absent supertype does nothing. It is applied where the
copy is created, so the copy's characteristics are already edited when
protection and the CR 707.10c re-target prompt read them.

**A copy of a PERMANENT spell becomes a token as it resolves**
(CR 608.3f, CR 111.13), through #923's one token-creation path, so a
doubler doubles it, the creature's own enters abilities fire, and a
bounce cannot turn it into a card. Nothing on the card declares that —
a card that copies a creature spell needs no more than the target
clause. What is NOT copiable, and must stay off `PrintedValues`: the
costs paid (`PaidCost.OptionalCosts` — kicker rides the STACK ITEM
under CR 707.10b, not the characteristics), and whether the spell was
foretold (`StackItem.Foretold`, `Card.FaceDownKind`) — a copy was
never cast. A copy of an ABILITY is still unbuilt.

**Tests** — see
[copy_effects_test.go](server/internal/cards/effects/copy_effects_test.go).
The pattern is `castCatalogSpell` → `resolveWithCopyChoice(t, g,
pick)` (pass `uuid.Nil` to decline) → assert on the battlefield card.
Assert the OBSERVABLE result — name, P/T, types, counters, and for
anything with a catalog entry, its behaviour — because a copy that
rewrote only the display fields looks identical on the wire.

### Adding a combat-keyword card (S18+)

Creatures with printed combat keywords (flying, reach, deathtouch,
lifelink, trample, vigilance, first strike, double strike, menace,
defender, haste, flash) declare those keywords through a single
`Spec.PrintedKeywords []string` slot. The engine auto-generates a
Layer 6 `StaticAbility` at catalog load time that appends each
keyword to the card's own `Characteristic.Abilities`, so
battlefield-side consumers see the same surface Lord of Atlantis
uses for granted keywords. The same list feeds off-battlefield
reads (flash gating on a card in hand).

```go
import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
    Register(Spec{
        OracleID:        "<uuid>",
        Name:            "Serra Angel",
        PrintedKeywords: []string{"flying", "vigilance"},
    })
}
```

Keywords are bare strings, case-sensitive, matching the
canonicalised forms the engine expects. Canonical tokens:

| Token | Keyword |
|---|---|
| `"flying"` | Flying (CR 702.9) |
| `"reach"` | Reach (CR 702.17) |
| `"first strike"` | First strike (CR 702.7) |
| `"double strike"` | Double strike (CR 702.4) |
| `"deathtouch"` | Deathtouch (CR 702.2) |
| `"lifelink"` | Lifelink (CR 702.15) |
| `"trample"` | Trample (CR 702.19) |
| `"vigilance"` | Vigilance (CR 702.20) |
| `"menace"` | Menace (CR 702.111) |
| `"defender"` | Defender (CR 702.3) |
| `"haste"` | Haste (CR 702.10) |
| `"flash"` | Flash (CR 702.8) |
| `"hexproof"` | Hexproof (CR 702.11) — #353, targeting gate |
| `"shroud"` | Shroud (CR 702.18) — #353, targeting gate |
| `"indestructible"` | Indestructible (CR 702.12) — S25, destruction path |
| `"changeling"` | Changeling (CR 702.73) — S26, every creature type (`game.KeywordChangeling`) |
| `"plainswalk"`, `"islandwalk"`, `"swampwalk"`, `"mountainwalk"`, `"forestwalk"` | Landwalk (CR 702.14) — #705, block legality |
| `"nonbasic landwalk"` | Nonbasic landwalk (CR 702.14c) — #705, block legality |
| `"fear"` | Fear (CR 702.36b) — artifact or black blockers |
| `"intimidate"` | Intimidate (CR 702.13b) — artifact blockers or a shared color |
| `"shadow"` | Shadow (CR 702.28b) — attacker and blocker must both have it or both lack it |
| `"horsemanship"` | Horsemanship (CR 702.31b) — requires horsemanship on the blocker |
| `"skulk"` | Skulk (CR 702.118b) — blocker power cannot exceed attacker power |
| `"protection from <quality>"` | Protection (CR 702.16) — #662, all four DEBT checks. The one PARAMETERISED token; see "Protection" below before writing one |
| `"infect"` | Infect (CR 702.90) — #748, the damage tail: -1/-1 counters on a creature, poison on a player |
| `"wither"` | Wither (CR 702.80) — #748, the damage tail: -1/-1 counters on a creature |
| `"toxic N"` | Toxic (CR 702.164) — #748, N extra poison on combat damage to a player. Numbered AND cumulative: read it with `game.ToxicTotal`, never `HasKeyword`, and grant it through `game.AppendKeywordAbility` so a second instance adds up ([ADR 0056](docs/decisions/0056-infect-wither-toxic.md)) |
| `"prowess"` | Prowess (CR 702.108) — #706, the first TRIGGERED keyword in the table: `TriggersForCard` turns each instance on the effective ability list into one trigger (`game/prowess.go`). Cumulative like toxic, so grant it through `game.AppendKeywordAbility`. Never write a prowess trigger by hand — declare the token ([ADR 0014 amendment 2026-09-24](docs/decisions/0014-combat-keywords.md)) |
| `"split second"` | Split second (CR 702.61) — #1519, a SPELL's keyword: `castHasSplitSecond` (`game/split_second.go`) stamps `StackItem.SplitSecond` at announce, and while it is on the stack nobody casts or activates a non-mana ability. Declare it on an instant or sorcery exactly like flash; never pass the sandbox `SplitSecond` cast flag from a card ([ADR 0007 amendment 2026-09-24](docs/decisions/0007-stack-foundation.md)) |

**A keyword that is a trigger** has two shapes, and ADR 0014's
2026-09-24 amendment says which to use. A constructor on
`Spec.Triggered` (`Cascade()`, `Storm()`, `Ward(...)`) when the keyword
carries a parameter a bare token cannot hold or triggers from the
stack; a token here with an engine-side trigger (prowess) when it lives
on permanents, is granted and printed on tokens, and needs to work on a
card with no catalog entry. Either way the trigger carries its name in
`game.TriggeredAbility.Keyword`, which `cards/coverage` reads (#1258).

Hexproof, shroud, indestructible and changeling are not combat
keywords, but they ride the same `PrintedKeywords` slot and the same
`HasKeyword` reader. Their
consumers are `CanBeTargetedBy` (hexproof, shroud),
`DestroyPermanentForEffect` + the damage-driven creature SBAs
(indestructible — see `server/internal/game/indestructible.go` for
what it deliberately does *not* stop) and `HasAllCreatureTypes` in
`creature_types.go` (changeling — see "Adding a creature-type card"
below). The landwalk tokens are read by `Game.BlockPairRefusalLocked`
(`game/block_legality.go`, `game/landwalk.go`) against the defending
player's lands, by effective characteristics on both sides; the rarer
variants (snow swampwalk, legendary landwalk, desertwalk) join the
table with their first card.

The five evasion keywords in #825 use that same pair function. Read colors,
types and power from effective characteristics; shadow restricts both
directions, while horsemanship restricts only the attacker's blockers. The
legal enumerator and block-decision signal share the engine answer. The
bot's attack estimate reads the public view and never decides legality.

The table is closed on purpose: **a keyword joins it in the same
change that teaches the engine to honour it.** Declaring a token the
engine does not read puts a badge on the card that promises a rule
nothing enforces.

**Two protection-family keywords are deliberately outside the
table**, for reasons [ADR 0038](docs/decisions/0038-protection-style-keywords.md)
§7 sets out. *Ward* is a triggered ability, not a targeting
restriction, and it ships per-card via the `effects.Ward(WardMana(…))`
helper (S30) — it stays out of the table because the COST is a
parameter a bare token has nowhere to put. **A GRANTED ward is the
same thing, not a keyword grant** (#626): "equipped creature has
ward {1}", "Knights you control … have ward {1}" are
`effects.WardGranted(cost, label, grants)`, where `grants` is the
ordinary `StaticAbility.AppliesTo` predicate and the GRANTING object
carries the trigger — an Equipment from the battlefield, an emblem
from the command zone (CR 114.3). Nothing is written onto the warded
permanent, because there is nowhere on a `[]string` to put the cost.
`Ward` and `WardAttached` are both narrow cases of it, so the printed
and the granted ward share one trigger body and one payment path; when
a card grants a ward alongside an anthem, pass the SAME predicate value
to both halves so they cannot drift. *Protection* tests its
quality against the SOURCE of a spell or ability (CR 702.16b), which
the targeting choke point never receives; it is not implemented, and
it is tracked in #662 (an ADR comes first). A card that prints
protection ships without it and says so in `Caveats`, as Baneslayer
Angel, both Swords and Animar do.

**Layer-granted keywords still use `Spec.Static`.** Lord of Atlantis
grants `"islandwalk"` to *other* Merfolk via a Layer 6
`StaticAbility` (`TribalKeywordGrant` in `tribal.go`) — that pattern
stays, and since #705 the grant is enforced like any other keyword.
Stromkirk Captain grants `"first strike"` to the other Vampires you
control with the same builder, and Trailblazer's Boots grants
`"nonbasic landwalk"` to its equipped creature with
`GrantToAttached`. `PrintedKeywords` is only for the card's own printed
keywords.

**Tests** — assert `Effective().Abilities` contains the keyword
strings after the card is pushed to the battlefield. See
`server/internal/cards/effects/serra_angel_test.go` for the template.
Combat behaviour (flying block restriction, trample overflow, etc.)
is tested in `server/internal/game/combat_test.go` against
manufactured battlefield state — card-level tests just verify the
keyword strings are exposed.

**First strike and double strike change the TURN, not just the
damage.** A combat in which any attacking or blocking creature has
either keyword as combat damage would begin has TWO combat damage
steps (CR 506.1 / 510.4), and the engine models that as a real step:
`first_strike_damage` sits before `combat_damage` in `turnSequence`,
and a turn that does not need it walks straight through it without
entering it (`Game.stepExistsLocked`). Each step grants priority, so a
"whenever this deals combat damage" trigger from first-strike damage
goes on the stack and resolves BEFORE regular damage is dealt, and the
table can respond in between (#717). Who deals damage in the second
step is fixed as the first one begins —
`Game.firstStrikeStepParticipants`, one record, read by both steps; the
only keyword read left in the second step is the one CR 702.7c asks
for, "plus the ones that have double strike now" (#716). Nothing about
this is per card: declare the keyword and the turn structure follows.

**A creature is in combat until the end of combat step ENDS (#785,
CR 511.3).** `AttackingTarget` / `BlockingTarget` are stamped at
declaration and cleared by the cursor as it leaves `end_combat`, not
as it enters — so "each attacking creature" reads the whole attack
during that step (Aetherize, Settle the Wreckage, Aetherspouts), an
"activate only if … attacking" condition is true there, and "at end of
combat" triggers, which fire as the step BEGINS (CR 511.2), see the
attackers. There is one clear point, `clearCombatLocked`; the
`ClearCombat` verb, `PassTurn` and the eliminated-seat rotation call it
themselves because a turn that ends early never leaves the step.

**"Blocked" is a state, not a blocker count (#715, CR 509.1h).** An
attacker is blocked the moment the block declaration is locked in, and
it stays blocked for the rest of the combat however many creatures are
still blocking it — `Game.blockedAttackers`, written only by
`commitBlockDeclarationLocked` and read only by the damage steps
(`attackerBlockedLocked`). So a blocked attacker whose blockers all
died assigns no combat damage (CR 510.1c) unless it has trample, which
sends all of it to what the creature is attacking (CR 702.19d/e), and
block legality — menace's count included — is judged once, at the
declaration (CR 509.1b), and never re-checked at damage. Do not derive
"unblocked" from the live blocker count anywhere.

**`advance_step` passes priority until the step ends** (#914,
CR 117.4). The sandbox's skip-ahead button used to move the cursor
whatever was on the stack, so a trigger the step owed resolved after
the next step's turn-based actions — combat damage before an afflict,
regular damage before the first-strike damage triggers, a draw before
an upkeep trigger. A step cannot end with objects on the stack, so
`Game.AdvanceStep` now drives `PassPriority` for the caller until the
stack is empty and only then moves the cursor. An empty stack is
unchanged: one move, no pass. The drive stops — cursor where it is, no
error — on a blocking prompt one of its resolutions raised, on a
CR 726 loop notice, or on the game ending. A card whose trigger fires
in one step and pays off in the next needs nothing for this; write the
trigger and the turn structure is already right.

**Keyword behaviour is engine-side, not catalog-side.** You do not
write flying/trample/deathtouch logic in the card file. The combat
engine reads `HasKeyword(card, "flying")` and routes accordingly.
Card files declare the strings; the engine does the rest.

**Protection (CR 702.16, #662, [ADR 0072](docs/decisions/0072-protection.md)).**
The one keyword whose token carries a PARAMETER, so it is the one
keyword with a parser. Declare it like any other — in
`PrintedKeywords` for printed protection, in `GrantToAttached` /
`GrantKeywordUntilEOT` for a granted one — but write the token as
`"protection from <quality>"` with the quality spelled the way the
card prints it:

```go
PrintedKeywords: []string{"flying", "protection from Demons", "protection from Dragons"},
Static:          []game.StaticAbility{GrantToAttached("protection from red", "protection from blue")},
```

Three rules, and all three exist so the grammar keeps exactly one
owner (`server/internal/game/protection.go`):

1. **One printed clause can be two abilities.** "Protection from
   Demons and from Dragons" is two tokens (CR 702.16m), and each is
   checked on its own. Never write a joined one.
2. **The closed grammar is colours (`red`), card types (`artifacts`),
   creature subtypes (`Demons`), `everything` and `the chosen player`.**
   Anything else — "monocolored", "opponents" — parses as NOTHING, so
   the permanent gets no protection at all and the card keeps its ADR
   0037 unimplemented flag. That is the honest answer, not a bug: do not
   route round it with a bespoke static, declare the caveat.
3. **Never parse a token yourself.** `game.ProtectionQualities(card)`
   is the reader, `game.ProtectedFrom(card, chars)` is the predicate,
   and `game.ProtectionFromColor(letter)` is how a colour PICK
   (Mother of Runes, the `choose_color` prompt) becomes a token. The
   client and the bot read the parse off `CardView.Protection`
   instead, because the bot may not import `internal/game` at all.
4. **The player quality is a constant, not a string you type.**
   `game.ProtectionFromChosenPlayer` is the whole token for CR 702.16k
   (#980, True-Name Nemesis); pair it with
   `AsEnters: ChoosePlayerAsEnters(name, Players)` so something writes
   `Card.ChosenPlayer`, which is what the reader resolves the quality
   against. A hand-typed near-miss mints no token and the card ships
   looking finished and doing nothing.

The four checks are engine-side and a card opts into none of them:
targeting (`CanBeTargetedBy`, CR 702.16b), attachment
(`attachmentLegalLocked`, CR 702.16c-d), damage (the
`protectionPreventsDamageReplacement` built-in, CR 702.16e) and
blocking (`BlockPairRefusalLocked`, CR 702.16f). **The quality is
tested against the SOURCE OBJECT, never its controller** — a white
player's Lightning Bolt is red, and an Equipment's own ability is
colourless however red the creature wearing it is. The ONE exception is
CR 702.16k's player quality, which is a claim about the source's
controller precisely because it is not a claim about the source's
characteristics; it reads `Card.ChosenPlayer` off the PROTECTED
permanent, and an unanswered prompt leaves it protected from nobody.
If you add a new
targeting path, it has to name its source: pass a `game.TargetSource`
built with `SourceObject` (a live spell or permanent),
`SourceSnapshot` (a value copy, when the object may be gone by the
time the answer arrives) or `SourceChooser` (a cost payment, which
does not target — a DECLARATION, not an omission).

Player protection (Teferi's Protection, Leyline of Sanctity) is out
of scope: `game.Player` carries no ability slice.

### Adding a "can't" card (S24+)

"Can't attack or block" (Pacifism), "can't be blocked" (Whispersilk
Cloak), "this creature can't block" (Carrion Feeder) and "its
activated abilities can't be activated" (Arrest, Faith's Fetters) are
**not keyword grants** — do not append a string to
`Characteristic.Abilities` for them. They are bits in
`game.Restriction`, written into `Characteristic.Restrictions` by one
of three primitives:

```go
Static: []game.StaticAbility{
    RestrictAttached(game.CantAttackOrBlock),    // an Aura / Equipment, on its host
    RestrictSelf(game.CantBlock),                // printed on the permanent itself
},
// …or, from a spell or an activated ability's Effect:
RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked}.Apply(ctx)
```

Five bits: `CantAttack`, `CantBlock`, `CantBeBlocked`, `CantActivate`,
`CantActivateMana` (plus `CantAttackOrBlock` for the common pair).
The activation pair is split because Faith's Fetters says "unless
they're mana abilities" and Arrest does not.

**The engine reads them; you do not.** Declarations go through
`game.AttackerEligible` and `Game.CanBlockLocked` (the boolean form of
`Game.BlockPairRefusalLocked`, which also says why), activations through
`game.CanActivateAbilities` / `CanActivateManaAbilities`, and
`internal/legal` calls the same functions — that shared predicate is
the whole reason a bot is never offered a move the engine refuses
(#544). If you add a bit, add its gate AND its enumerator site in the
same PR, and assert `dispatchAll` over the enumerated moves.

Restrictions are checked **at declaration only** (CR 508.1c, 509.1b).
A creature pacified after attackers were declared keeps attacking.

[ADR 0045](docs/decisions/0045-combat-restrictions.md) has the
taxonomy, including what the vocabulary deliberately cannot say:
Silent Arbiter's and Crawlspace's count limits, which are set-shaped
predicates rather than bits — shipped in #1507 as `BlockRule.Limit`
(one more entry in `checkBlockDeclarationLocked`) and `game.AttackLimit`
(judged by both attack declaration verbs and the enumerator).

**Propaganda's attack cost used to be on that list and is not any
more** ([ADR 0080](docs/decisions/0080-attack-taxes.md), #1063). A
price is not a prohibition, so it is not a bit:

```go
AttackTaxes: []game.AttackTax{
    AttackTax("{2}", "Creatures can't attack you unless their controller pays {2} …"),
    // …or the count-scaled form, and the planeswalker clause:
    ProtectingPlaneswalkers(AttackTaxCounting(enchantmentsYouControl, "… {X}, where X is …")),
},
```

The "you" is **structural** — a tax protects its own controller and
nothing a card file writes can widen it to another seat. `ManaCost`
returns a cost STRING for ONE attacking creature; the engine asks it
once per attacker and concatenates, so three attackers into Propaganda
is `{2}{2}{2}`. `AttackTaxScope`'s zero value is the narrow "you", so a
card that forgets `ProtectingPlaneswalkers` under-taxes rather than
over-taxes.

**The engine charges it; you do not.** `Game.PriceAttackDeclaration`
is the one pricer, `DeclareAttackerWith` / `DeclareAttackersWith` pay
it at CR 508.1a before anything is staged, ALL OR NOTHING, and
`internal/legal` calls the same pricer so an unaffordable attack is
never offered (#544). It pays through the same
`payAbilityManaCostLocked` an activated ability uses, so `ManaTrigger`
fires for the taps and nothing about mana is duplicated.

### Attaching, and an ability whose source has gone (#812)

Two rules, each at one choke point, and no card file checks either.

**An attach that cannot happen does nothing** (CR 701.3b).
`game.AttachForEffect` is the only writer of `Card.AttachedTo` outside
the state-based action, and when the attachment is not on the
battlefield (only a permanent can be attached, CR 301.5c), the host is
not there, or the two are the same permanent, it emits
`EventAttachSkipped` with the reason on `Label` and returns **nil**. It
is not an `EventEffectError`: nothing failed, and the catalog soak
fails the nightly run on any effect error.

**An ability that attaches its SOURCE checks the source is still that
permanent.** `game.AttachSourceForEffect(item, host)` is the door for
equip (CR 702.6a) and for fortify and reconfigure when they arrive; it
asks `AbilitySourceGoneForEffect`, which is "not on the battlefield, OR
back with a different `Card.ObjectEpoch`" (CR 400.7 — a Loxodon
Warhammer bounced and replayed while its equip is on the stack keeps
its instance ID and is a new object). `StackItem.SourceEpoch` is the
announce-time reading, stamped by the two paths that build a
`StackItemActivated` item and by nothing else.

`AbilitySourceGoneForEffect` is NOT a general "did my source survive"
helper. CR 608.2 resolves an ability whether or not its source is
around, and almost every ability should carry on from last known
information — "{T}: this deals 2 damage to any target" deals its damage
from the graveyard. Consult it only where the effect genuinely cannot
be performed without the source as a permanent.

### Adding a block-rule card (S37+, #750)

"Can't be blocked except by Walls", "can't be blocked by creatures with
power 2 or less", "creatures with power less than this creature's power
can't block creatures you control", "can't be blocked by more than one
creature" are CR 509.1b block rules with a PARAMETER. They go in
`Spec.BlockRules`, built from
[block_rules.go](server/internal/cards/effects/block_rules.go): a
**scope** (whose creatures the rule binds, relative to this permanent)
plus a **rule**.

```go
BlockRules: []game.BlockRule{
    CantBeBlockedExceptBy(OnAttached(), OfCreatureType("Wall"), "Walls"),     // Prowler's Helm
    CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),   // Legolas Greenleaf
    CantBlockAttackers(PowerLessThanSource(), ControlledBySourceController(),
        "creatures with power less than Champion of Lambholt's can't block creatures its controller controls"),
    CantBeBlockedWhile(OnAttached(), PowerLE(3)),                             // Thieves' Tools
    MaxBlockers(OnAttached(), 1),                                             // Vorrac Battlehorns
    MinBlockers(OnSelf(), 3),                                                 // Rampaging Ceratops
},
```

- **The scope is never optional.** A rule is read off every permanent
  for every pair the engine checks, so a hand-written `Pair` that
  forgets "is this attacker mine?" binds the whole table.
- **The label is the printed parameter.** The player reads it in the
  refusal sentence ("can't be blocked except by Walls, and Llanowar
  Elves is not one").
- **Flat "can't block" / "can't be blocked" is still a `Restriction`
  bit** (`RestrictSelf`, `RestrictAttached`). A rule is for a clause
  with a parameter.
- **"This turn"** is `BlockRuleUntilEOT{Target: id, Rule: func(s
  BlockScope) game.BlockRule { … }}` (Gingerbrute, Departed Deckhand).
  It snapshots the set at resolution (CR 611.2c).
- **A token that prints one** declares it on its `tokenTemplate`'s
  `BlockRules` slot (Avatar Kuruk's Spirit).
- **"No more than N creatures can block each combat"** (Silent Arbiter,
  Dueling Grounds, Caverns of Despair) is `NoMoreThanNCanBlockEachCombat(n)`
  — `BlockRule.Limit`, judged over every block in the combat and refused
  as `declaration_limit` (#1507). It takes no scope: the printed line
  binds every creature at the table.
- **The attack half** — "no more than N creatures can attack each
  combat" / "…can attack you each combat" (Crawlspace) — is not a block
  rule. It goes in `Spec.AttackLimits`, built with
  `NoMoreThanNCanAttackEachCombat(n)` or
  `NoMoreThanNCanAttackYouEachCombat(n)` from
  [attack_limits.go](server/internal/cards/effects/attack_limits.go).
  "You" is the card's controller, the player — an attack on their
  planeswalker does not count. Both declaration verbs and the
  enumerator read the same check, so a bot is never offered a refused
  attack ([ADR 0045](docs/decisions/0045-combat-restrictions.md)
  Decisions 43-45).
- **The narrower and conditional variants** (#1534, Decision 47):
  "no more than N creatures can attack <this planeswalker>" (The
  Eternal Wanderer) is `NoMoreThanNCanAttackThisEachCombat(n)`; "as
  long as <this> is tapped, …" (Mirri, Weatherlight Duelist) wraps any
  limit as `AsLongAs(ThisIsTapped, …)`, read live; "each opponent
  can't block with more than N creatures this combat" is the
  `EachOpponentCantBlockWithMoreThanN{N, Label}` primitive in a
  trigger's effect, a turn-scoped `BlockRule.Limit` with
  `LimitPerDefender`, so each defending player is counted on their
  own.

Tests: [block_rules_test.go](server/internal/cards/effects/block_rules_test.go)
pins every shape through the verb, `legal.EnumerateFor` and
`block_decision_seats`;
[combat_limits_test.go](server/internal/cards/effects/combat_limits_test.go)
pins the whole-combat limits, checking the enumerator against the verb
on a clone for every candidate, and
[conditional_combat_limits_test.go](server/internal/cards/effects/conditional_combat_limits_test.go)
does the same for the #1534 variants, planeswalker targets included.

### Adding a triggered ability (S19+)

Triggered abilities ("when ~ enters", "when ~ dies", "at the
beginning of your upkeep") live on `Spec.Triggered
[]game.TriggeredAbility`. A per-game harvester listens to the event
log, runs `AppliesTo` for each watched event kind, and calls `Build`
to put an item on the stack. The item resolves — and its `Effect`
runs — only when every player has passed priority in succession,
so opponents can respond (counter the ability, remove the target,
sacrifice in response). See [ADR 0018](docs/decisions/0018-triggers-on-the-stack.md).

**Write the printed shape with a constructor** from
[triggers_common.go](server/internal/cards/effects/triggers_common.go)
(#579). The label is the whole stack label, "<card> — <what
happens>", and `Do(...)` sequences primitive values whose `Player` /
`Controller` field defaults to the item's controller:

```go
Triggered: []game.TriggeredAbility{
    WhenThisEnters("Mulldrifter — draw two cards", Do(DrawCards{N: 2})),
    Optional(WhenThisDies("Solemn Simulacrum — draw a card", Do(DrawCards{N: 1})),
        "Solemn Simulacrum — draw a card?"),
    AtYourUpkeep("Awakening Zone — create an Eldrazi Spawn", Do(CreateToken{Template: EldraziSpawnToken(), N: 1})),
    WheneverYouCast(Noncreature(), "Black Waltz No. 3 — 2 damage to each opponent",
        func(g *game.Game, item *game.StackItem) error { return damageToEachOpponent(g, item, 2) }),
},
```

The shapes: `WhenThisEnters`, `WhenThisDies`, `WhenThisEntersOrAttacks`,
`WheneverThisAttacks`, `AtYourUpkeep`, `AtEachUpkeep`, `AtYourEndStep`,
`AtYourPrecombatMain`, `WheneverYouCast(pred, …)`, `WheneverYouDraw`,
`WheneverAnOpponentDraws`, `Landfall`,
`WheneverAnotherCreatureEntersUnderYourControl`,
`WheneverACreatureYouControlDies`,
`WheneverThisDealsCombatDamageToAPlayer`, `WheneverYouGainLife`. A
condition with no shape yet is `On(game.EventX, when, label, effect)`
where `when` is any `AppliesTo`-shaped predicate — a named one
(`Self`, `ByYou`, `ByAnOpponent`, `AnyPlayer`, `ThisDied`,
`ThisAttacked`, `YouCast(pred)`, `AnOpponentCast(pred)`,
`LandEnteredUnderYourControl`, `ACreatureYouControlDied`, …, or
`AllOf(...)` of several) or a closure. One printed ability with two
conditions is `OnAny([]game.EventKind{…}, …)`. `Targeting(t, spec)`
adds a target clause; the effect then reads `item.Targets[0]`, so it
is a closure rather than `Do`. Add a missing shape or predicate to
`triggers_common.go`, not to the card file.

**Triggers in other zones (#925, CR 113.6):** a trigger watches from
the battlefield unless it says otherwise, and the ones that say
otherwise wrap the shape: `InGraveyard(Landfall(...))` is Bloodghast,
`WhenThisIsPutIntoYourGraveyardFromYourLibrary(...)` is Narcomoeba,
`InExile(AtYourUpkeep(...))` is suspend's countdown (#659). The
wrapper sets `game.TriggeredAbility.Zones`, which REPLACES the default
rather than adding to it — an ability printed to work from the
graveyard does not also fire from play, and that is the whole point:
"return this card from your graveyard" off a permanent has nothing to
return. "You" inside such a trigger is the card's **owner** (CR
108.4), because a card outside the battlefield has no controller; the
harvest stamps it, so `ByYou`, `Self` and `Landfall` all read as
printed and the stack item goes to the owner. Only the graveyard and
exile are walked — `effects.Register` panics at boot on any other
zone, and `ZoneStack` is `FromStack` (cascade). The per-event cost of
the battlefield walk is unchanged: `game.IndexTriggerZones` builds a
per-event-kind index at `Register`
([trigger_zones.go](server/internal/game/trigger_zones.go)), so an
event kind nothing declares costs one map lookup and no walk.

An effect that needs the item (targets, X, the source ID) or must
capture something off the event is a closure with the `Effect`
signature, exactly as before:

```go
WhenThisEnters("Mulldrifter — draw two cards",
    func(g *game.Game, item *game.StackItem) error {
        return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
    }),
```

Every constructor returns an ordinary `game.TriggeredAbility`. The
long form below is exactly what it builds, and is still the right
tool when `Build` itself has to do something — capture `ev.Actor` for
a PayUnless payer, read the event to decide whether to return nil:

```go
Triggered: []game.TriggeredAbility{{
    Watches: []game.EventKind{game.EventETB},
    AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
        return ev.CardID == source.InstanceID
    },
    Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
        return game.NewTriggeredItem(source, "Mulldrifter — draw two cards",
            func(g *game.Game, item *game.StackItem) error {
                return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
            })
    },
}},
```

**Combat declarations (#830, #859):** BOTH combat declarations
announce once, at their **lock-in** — the first priority boundary of
the step that staged them — so attack and block triggers are
harvested from the FINAL assignment and a creature re-pointed
mid-step never triggers what it left. `EventAttack` (one per declared
attacker, CR 508.1; `Target` is the defending player, planeswalker or
battle it ends on, read through `b17DefendingPlayer` for the player
behind it) comes from `commitAttackDeclarationLocked`; `EventBlock`
(one per blocker/attacker pair: "whenever this creature blocks",
"becomes blocked by a creature") and `EventBecomesBlocked` (one per
blocked attacker, CR 506.4: "becomes blocked", afflict) come from
`commitBlockDeclarationLocked`. A permanent PUT onto the battlefield
attacking (CR 506.3c) is not declared and announces nothing. See
[ADR 0045](docs/decisions/0045-combat-restrictions.md), amendment
Decisions 19-21 and 22.

**A trigger in the cleanup step gets priority (#661, CR 514.3a).**
Cleanup normally grants nobody priority, so a trigger queued there —
the hand-size discard is the usual one — used to wait for the next
player's upkeep. It doesn't now: if a state-based action is performed
or a trigger is waiting, the SBAs happen, the triggers go on the
stack, the active player gets priority **in** the cleanup step, and
once the stack is empty and everyone passes, **another cleanup step
begins** (hand size checked again, the CR 514.2 sweep run again, so an
"until end of turn" effect a cleanup trigger creates still ends this
turn). One exit decides it, `exitCleanupStepLocked`
(`server/internal/game/cleanup.go`), called from the cleanup step-entry
hook and from `DiscardSelection`'s resume. Nothing changes for a quiet
cleanup: it ends the turn in the same call it always did, so no new
auto-pass stop appears.

**"This turn" (#586):** anything a card asks about the current turn
is read off `Game.TurnTally`, never by walking `g.Events`:
`g.TurnTallyFor(player)` carries `LifeGained`, `LifeLost`, `CardsDrawn`,
`CreaturesDied`, `TokensCreated`, `PermanentsSacrificed`, `LandsEntered`,
`AttacksDeclared` and `CombatDamageToPlayers`; `g.TurnTally.CreaturesDied`
is the table-wide count; `g.ResolvedThisTurn(source, label)` and
`g.TriggeredThisTurn(source, label)` are the "once per turn" gates (an
empty label sums the source's abilities) and are per **object**, not
per card — a permanent that left the battlefield and came back this
turn answers zero, because CR 400.7 makes it a new object and the key
carries `Card.ObjectEpoch` (#936). The CR 726 loop breaker's
`LoopRun` / `LoopAllowance` share the same (source, label) pair and
stay per **card**, so a blink loop still trips the threshold; one key,
two projections, and `TurnTally`'s field comments say which reader
takes which;
`g.EnteredWithSubtypeThisTurn(player, subtype)` counts permanents that
entered under a player's control with a subtype, judged as they entered
rather than as they are now (a changeling counts for every creature
type; a type granted by another permanent's static at that moment is
not seen, so a card reading it declares that weaker gap);
`g.EnteredThisTurn(cardID)` is the same record's per-**object** cell —
"each green creature that **entered this turn**" (Oran-Rief), and the
"is the source itself one of them" half of an "another X entered this
turn" clause (Éowyn);
`g.PlayersDealtCombatDamageThisTurnByName(controller, name)` and
`g.PlayersDealtCombatDamageThisTurnBySubtype(controller, subtype)` are
the set of players a creature of yours hit in combat this turn, for the
target predicate of a "whenever … deals combat damage to a player …
**that player**" trigger, which is not handed the trigger's event
(Trygon Predator, Alela). Those three and the two subtype tallies are
recorded **as the event happens**, not read back later: the permanent
being asked about is usually gone by the time anything asks, and a
token is gone from every zone (CR 704.5d). A filtered question the tally
does not carry ("you sacrificed a *Food* this turn") ranges over
`g.EventsThisTurn()`, which is bounded at the real turn boundary — the
old upkeep-bounded scans missed the untap step, which is where #1009
finally bit. A counter the tally
should carry and does not is a field on `PlayerTurnTally` plus one case
in `turnTallyListener`, not a new scan.

**Never anchor a "this turn" question on `EventBeginUpkeep`.** The turn
begins at `onTurnBeganLocked`, which resets the tally *before* the untap
step; the upkeep event comes after it, and an untap-step trigger or
choice (ADR 0070) can put a permanent onto the battlefield or otherwise
act in between. Every such walk in the catalog is gone (#1009); the two
shapes that remain legitimate are a **cursor-bounded** walk (`ev.Seq`,
"what happened after this point") and a **most-recent-X** walk, neither
of which is a "this turn" question.

**Adding an activated ability (S21+):** put it in
`Spec.Activated`, one entry per printed ability, with the cost built
from the constructors in
[activated.go](server/internal/cards/effects/activated.go):

```go
Activated: []ActivatedAbility{{
    Label:   "Sacrifice a creature: deal 1 damage to any target",
    Cost:    SacrificeACreature(),          // or TapCost(), SacrificeThis(),
    Targets: TargetAny(),                   // ManaCost("{1}{B}"), PayLife(2),
    Effect: func(g *game.Game, item *game.StackItem) error {
        // Same contract as a triggered ability's item: never
        // capture a *Card; read the source via NewContext(g, item).
    },
}},
```

Compose multi-part costs with `Plus(ManaCost("{2}"), TapCost())`.
"This ability costs {1} less to activate …" is the ability's own
`CostModifiers` slot (#1296), not `Spec.CostModifiers` — the slot
prices that one ability wherever it functions (a channel land's hand
included), and `CostsLessForTheCardItTargets(label, ColorsOf)` is the
clause that reads the ability's target (Dragonfire Blade).
The engine validates every component before paying any of them, and
pays at announce — so a sacrifice cost's dies-triggers land on the
stack above the ability and resolve first. Mana abilities do NOT go
here (they skip the stack, CR 605.3b); they stay in `ManaAbilities`.
See [ADR 0020](docs/decisions/0020-activated-abilities.md).

**"Exile N cards from your graveyard" / "… from your hand" (#1297):**
`ExileFromGraveyard(n, label, match)` / `ExileFromHand(n, label, match)`,
composed with `Plus` — Grim Lavamancer is
`Plus(ManaCost("{R}"), TapCost(), ExileFromGraveyard(2, "two cards", nil))`,
Moorland Haunt passes `MatchCreature`. The activator picks the cards at
announce (`exile_ids`); they are exiled, not discarded, and the effect
reads which ones through `ctx.Exiled()` (Holistic Wisdom). Not
`ExileThis()`, which is the SOURCE. A variable count ("Exile X cards")
has no shape yet.

**An `{X}` in the cost:** put it in the mana component, read it back
with `ctx.X()`, and declare `XMatters: true` on the Spec (#810). The
engine still derives "this ability prompts for X" from the cost
string, so the view, the enumerator and the client can never disagree
with the card about whether there *is* an X; `XMatters` answers the
other question, which nothing can derive — does the card do anything
at X=0? "Look at the top X cards" does not, so the bot's enumerator
declines to offer it there (CR 732.2a; `internal/legal/x.go`). Leave
it unset only for a card with a fixed RIDER, something that happens
whatever X is — The Goose Mother's 2/2 flying body — and add an entry
to `xMattersAllowlist` in `effects/x_matters_guard_test.go` saying
what the rider is, because that source scan fails the build on a Spec
that reads `ctx.X()` without declaring:

```go
XMatters: true,                                              // every card below
Cost: Plus(ManaCost("{X}{X}"), TapCost(), SacrificeThis()),   // Treasure Vault
Cost: Plus(ManaCost("{X}"), TapCost(), MinX(1)),              // Helm of Obedience
```

`{X}{X}` is two slots, so X=3 costs six — the slot count comes off
`ParseCost` rather than a flag, which is what keeps a double-X cost
from silently charging half. `MinX(n)` is the printed floor: "X
can't be 0" is `MinX(1)`, and it is a real rule, not a hint — the
engine refuses an announcement below it and the enumerator declines
to offer the ability at all when the activator cannot reach the
floor. `Register` panics on a `MinX` with no `{X}` beside it.

X is announced as part of activating (CR 602.2b), **before any cost
is paid**, and locked onto the stack item: `item.XValue`, the same
slot a cast writes, so `ctx.X()` is the same accessor an X spell's
`OnResolve` uses. It cannot change afterwards, which is why "create
X Treasures" is a fact about the announcement rather than about how
much mana is around at resolution.

X lives in the MANA component and nowhere else. A cost with a
variable COUNT — Ruthless Technomancer's "Sacrifice X artifacts" —
is a different seam and is still open.

**A Phyrexian symbol in the cost (#787):** `{W/P}` and CR 107.4's ten
hybrid Phyrexian symbols (`{W/U/P}` … `{G/U/P}`) are ONE
`ColorRequirement` each — a set of colour options plus `Phyrexian` —
so there is no new symbol kind to declare and nothing for a card file
to write. The "or 2 life" half (CR 107.4c/f) is announced on the
CAST, as `CastSpellParams.PhyrexianLife`: the number of the cost's
Phyrexian symbols being paid with 2 life each, validated against what
the cost prints and against CR 119.4, paid through `PayLifeForEffect`.
An ACTIVATED ability announces the same thing the same way (#917):
`ActivateAbilityParams.PhyrexianLife`, the same wire name
`phyrexian_life`, through the same strike-and-pay helper, so Birthing
Pod's `{1}{G/P}` is `{1}` and two life for a player with no green. The
board asks the question in both chains (#916): the ceiling ships as
`phyrexian_symbols` on the card, on the chosen alternative cost and on
the ability, so **no client parses a mana string** to find out how
many symbols a cost prints.

**"Activate only if …" / "Activate only during your turn" (#743):**
the ability's `Condition`, a `func(g, controller, source) bool` built
from [activation_conditions.go](server/internal/cards/effects/activation_conditions.go)
(or `ControlsAtLeast`), the same shape a `ManaAbility.Condition`
takes:

```go
Condition: OpponentControlsAtLeast(4, MatchLand),   // Tectonic Edge
Condition: DuringYourTurn(),                        // Sanctum of Eternity
```

The engine checks it once, at activation, before X, targets or any
cost (`ErrConditionNotMet`, nothing paid), never at resolution
(CR 602.1b); the enumerator and the view (`condition_unmet`) read the
same closure. "Only as a sorcery" stays `SorcerySpeed: true` beside
it — Speaker of the Heavens sets both. Contract: read-only, runs under
`g.mu` (use `*ForEffect` accessors and `g.Turn` / `g.Seats`, never a
locking accessor), and reads only public information, because every
viewer receives the flag. Never drop a condition you can't express,
and never move it into `Effect`: the first is stronger than printed
(#259), the second charges the cost for nothing. "Activate only once
each turn" and boast still have no shape (the per-source activation
count in `docs/engine-seams.md`).

**Adding an additional cost to cast (S21 sub-PR 5):** "As an
additional cost to cast this spell, discard a card" goes in
`Spec.AdditionalCost`, not in `OnResolve`:

```go
AdditionalCost: DiscardCost(1),   // Thrill of Possibility, Big Score
```

The distinction is observable, which is why it's modelled: the cost
is paid to CAST the spell, so the discard happens with the spell
already on the stack and a discard payoff (Mary Read and Anne Bonny,
Marauding Mako) triggers ABOVE it and resolves first. It is also
paid whether or not the spell resolves — countering it doesn't give
the card back. Fold the discard into `OnResolve` and both of those
go wrong. The caster picks the cards in a client prompt that opens
before the X / mode / target prompts; they ride `cast_spell` as
`discard_ids` and the engine validates them at announce (the spell
itself is never a legal pick — CR 601.2a already moved it to the
stack). See [ADR 0021](docs/decisions/0021-additional-costs.md).

**Gift (CR 702.174, [ADR 0089](docs/decisions/0089-gift.md)):** one
field, and never a hand-rolled optional cost:

```go
Gift: GiftACard(),                                        // Dawn's Truce
Gift: GiftACard().Instead(TargetSpell("target spell")),   // Long River's Pull: "instead counter target spell"
```

The engine grows the rest — the "choose an opponent" cost (an ADR 0073
optional cost the client and the bot both offer), the gift itself BEFORE
`OnResolve` on an instant or sorcery, and the "when this enters, if the
gift was promised" trigger on a permanent. The card's own "if the gift
was promised" text reads `ctx.GiftPromised()` (or `source.GiftPromised()`
in a permanent's trigger). `.Instead(clause)` is the WHOLE target clause
of a promised cast, including one the unpromised spell does not have at
all (Valley Rally).

**Impulse exile (S21 sub-PR 6):** "exile the top card of that
player's library — until end of turn, you may cast that card" is
the `ExileTopWithPermission` primitive:

```go
ExileTopWithPermission{
    From:     victim,            // whose library
    GrantTo:  item.Controller,   // who may play it — usually not the owner
    N:        1,
    CastOnly: true,              // "you may CAST" (Ragavan); omit for "play" (Breeches)
    AnyColor: true,              // "spend mana as though it were mana of any color"
}.Apply(ctx)
```

The permission rides `Card.ExilePlay` and expires at end of turn.
`CastOnly` is not a detail: a land exiled by Ragavan is stranded,
because playing a land is not casting (CR 305.1), and the client
shows no button on it. Timing still applies on top — the grant says
you *may* play the card, not *when*. See
[ADR 0022](docs/decisions/0022-impulse-exile.md).

**The client has ONE cast entry point** (#874), `handlePlayCard` in
`Board.svelte`, and every surface that casts a card reaches it: the
hand, the graveyard's flashback button, and the exile pile's impulse
button, each passing the zone it came out of (and, for a grant that
names a face, that face). The prompts a cast owes the player — the
face picker, the alternative cost, the additional costs, X, the tap
cost, the modes, the targets — all hang off that one chain, so a
surface that dispatches `cast_spell` itself is not a shortcut, it is a
cast with every one of those questions silently answered "none". That
is exactly what the impulse button did until #874: a grant offering
the PRINTED cost could only ever announce X = 0. When you add a new
way to cast something, hand it to `handlePlayCard`; never build a
payload.

The same slot takes a **sacrifice** clause (S21):

```go
AdditionalCost: SacrificeCost("a creature", Creature()),           // Village Rites, Altar's Reap
AdditionalCost: SacrificeCost("an artifact or creature",           // Deadly Dispute
    Or(Artifact(), Creature())),
```

Identical reasoning one zone over: the creature dies with the spell
on the stack, so Blood Artist and Zulaport Cutthroat drain BEFORE the
cards are drawn, and countering the spell doesn't hand the creature
back. With nothing to sacrifice the spell is uncastable — the view
stamps `AdditionalCostView.SacrificeOptions` filtered to the caster's
own permanents (CR 701.21a), and an empty list is what
`canCastFromHand` greys the card on. The pick rides `cast_spell` as
`sacrifice_ids`, and the client reuses `SacrificeCostModal`, the same
picker the CR 602 abilities open.

`SacrificeCost` builds its spec with the shared `sacrificeSpec`
helper, so a sacrifice cost is validated by the same code whether it
hangs off a spell, an activated ability or a mana ability.

**An OPTIONAL additional cost — kicker, multikicker, buyback (ADR
0073, #664):** the same `game.AdditionalCost` with `Optional` set,
declared in `Spec.OptionalCosts` rather than `Spec.AdditionalCost`
because an optional cost has an INDEX — the announcement names
positions in that slice:

```go
OptionalCosts: []game.AdditionalCost{Kicker("{4}")},                            // Burst Lightning
OptionalCosts: []game.AdditionalCost{Multikicker("{G}", 20)},                   // Wolfbriar Elemental
OptionalCosts: []game.AdditionalCost{Buyback("{3}")},                           // Capsize
OptionalCosts: []game.AdditionalCost{BuybackSacrifice("a land", Land())},       // Constant Mists
OptionalCosts: []game.AdditionalCost{KickerSacrifice("a creature", Creature())},
```

Use the keyword constructor, never a hand-rolled
`game.AdditionalCost{Optional: true}` — for the reason `Flashback` has
one. The constructor carries the `Key` the ENGINE reads, and a
hand-rolled one compiles and then never returns a bought-back card to
hand. `Register` refuses the mistakes that would otherwise ship
quietly: an `Optional` cost in the mandatory slot, a missing or
duplicated `Key`, a repeatable cost that also demands cards or
permanents (every printed multikicker is mana), and an optional
`PayLifeX` (it would fight the mandatory cost for the shared `XValue`
slot).

Read the choice back at RESOLUTION with `ctx.WasKicked()` /
`ctx.KickedTimes()`, the same shape `ctx.PaidAltCost("overload")`
gives an overloaded spell:

```go
amount := 2
if ctx.WasKicked() { amount = 4 }        // Burst Lightning
```

Read it from a PERMANENT's own trigger with
`game.CardKickedTimes(*source)` — Gatekeeper of Malakir's "when this
enters, **if it was kicked**", Wolfbriar Elemental's count. Not the
stack item: it is out of `StackMeta` before the ETB event is emitted,
so the resolution path carries the record onto the permanent as
`Card.PaidOptionalCosts` (CR 400.7d) and that is what these read. It
is per-instance and cleared on the way out, so a kicked creature that
dies and is reanimated is not kicked.

**Buyback's return is the engine's, not the card's.** Declare the cost
and stop. `routeStackCardToGraveyardLocked` reads the paid record and
routes the resolving spell to its owner's hand through the same
stack-exit primitive flashback uses (CR 702.27a) — only on a
RESOLUTION, so a bought-back spell countered by game rules still goes
to the graveyard. A card that also returned itself in `OnResolve`
would be moving a card that is still on the stack.

**Still out:** escalate and entwine (their cost is per extra MODE,
which the index-list announcement cannot express), and "enters with a
counter for each time it was kicked" (Everflowing Chalice, Joraga
Warcaller) — that count is read during the CR 614 entry pipeline,
before the record reaches the permanent.

**"You can't cast …" and "cast this only if …" (ADR 0073, #760):**
one announce-time gate, two ways to reach it. A restriction a
PERMANENT imposes goes in `Spec.CastRestrictions`, built from the
constructors in
[cast_restriction.go](server/internal/cards/effects/cast_restriction.go):

```go
CastRestrictions: []game.CastRestriction{                                  // Rule of Law
    EachPlayerMaxSpellsPerTurn(1, "Rule of Law — each player can't cast more than one spell each turn."),
},
CastRestrictions: []game.CastRestriction{                                  // Grafdigger's Cage
    PlayersCantCastFrom("Grafdigger's Cage — players can't cast spells from graveyards or libraries.",
        game.ZoneGraveyard, game.ZoneLibrary),
},
```

A condition the SPELL prints goes in `Spec.CastCondition`, with its
printed clause beside it — `Register` refuses either half alone,
because the clause is the message the player is shown:

```go
CastCondition:      LegendarySorcery(),      // Urza's Ruinous Blast, CR 205.4e
CastConditionLabel: LegendarySorceryLabel,
```

Same contract an ability's `Condition` has: read-only, under `g.mu`,
public information only — the answer reaches every viewer as
`cant_cast` on the card. Checked once, at announce, and never at
resolution: a legendary creature that dies while the sorcery is on the
stack does not counter it. The engine reads them; you do not — one
function (`Game.CastGateLocked`) answers for `CastSpell`, the
legal-move enumerator and the view, which is what keeps a bot from
being offered a cast the engine refuses and the client from rendering
a button it would reject.

**Still out, and named in the ADR:** a ban with a DURATION (Silence,
Reflector Mage) wants #755's registries, and a ban on a chosen card
NAME (Meddling Mage, Nevermore) wants a choose-a-card-name prompt that
does not exist.

**"Each player sacrifices a creature of their choice" (S21):** use the
`EachPlayerSacrifices` primitive, not a loop over opponents:

```go
EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"}  // Grave Pact
EachPlayerSacrifices{Match: Creature(), Label: "a creature"}                          // Fleshbag Marauder
```

"Of their choice" is the rules content: it fans out one
`PendingChoiceSacrifice` per affected player, each addressed to that
player and offering only their own permanents, so nobody picks for
anyone else. `ExceptController` is the difference between "each other
player" / "each opponent" and "each player" (Fleshbag includes you, and
is a legal answer to its own trigger).

Two things this deliberately is **not**:

- **Not a targeting prompt.** The effect doesn't target, so hexproof,
  shroud, protection and "can't be the target" are all irrelevant, and
  it resolves fine when nobody has a creature. Reusing
  `PendingChoiceSacrifice` rather than `PendingChoicePickTarget` is
  what keeps those restrictions from leaking in.
- **Not optional.** A player with legal permanents must pick one; a
  player with none is skipped at queue time rather than prompted and
  allowed to decline.

Prompts are pruned in `executeBattlefieldLeaveLocked` — the single
choke point for a permanent leaving the battlefield — because an
outstanding choice *stops priority from passing*, so
`runStateChecksLocked` is exactly what does not run while one is
waiting. Put the re-check anywhere else and a creature that dies to a
drain mid-resolution leaves a prompt nobody can answer.

The client answers it through the shared `ChoicePromptModal` card grid
with the ordinary `{choice_id, card_ids}` payload; the dispatcher routes
that shape by the choice's kind, as it already does for `{apply}` and
`{order}`.

**Scry (S21):** use the `Scry` primitive. Anything the card says
*after* "then" goes in `Then`, not on the next line:

```go
Scry{Player: ctx.Controller(), N: 1}                       // Viscera Seer
Scry{Player: c, N: 2, Then: func(g *game.Game) error {     // Preordain: "Scry 2, then draw"
    return g.DrawNForEffect(c, 1)
}}
```

`Scry` only QUEUES a prompt — nothing moves until the player answers.
So a draw written as the statement after it resolves FIRST, which is
wrong twice over: it takes one of the cards the player is still
deciding about, and it leaves the prompt permanently unanswerable,
because that card is no longer in the library for the reorder to put
back. (That was a real bug in the first draft; there's a test pinning
the ordering.)

Two other things the engine handles so a card never has to:

- **Scry is "look at", not "reveal".** Only the scrying player is
  marked a knower, so the wire redacts the cards for every other seat.
  A copy-paste from `SearchLibraryForEffect`'s reveal path would mark
  every seat and hand the table the top of a library — a real
  information advantage, not a cosmetic slip.
- **An empty library is not an error.** The scry looks at nothing,
  queues no prompt, and `Then` still runs — the instruction after
  "then" isn't conditional on there having been cards to look at.

**Face-down objects (CR 406.3a / 708, [ADR 0069](docs/decisions/0069-face-down-objects.md)).**
A card is face down because of a *kind*, and the kind answers every
question about it. `Card.SetFaceDown(kind)` and `Card.ClearFaceDown()`
are the only writers of `FaceDown` + `FaceDownKind`; never set either
field directly. Seven kinds, in two families:

- **exile** — `FaceDownExiled` (CR 406.3: nobody may look, not even
  the player who exiled it — Necropotence) and `FaceDownForetold`
  (CR 702.143d: the OWNER may look). An exiled face-down card keeps
  its real characteristics and its catalog entry, because the cast out
  of exile needs them.
- **CR 708.2 permanents** — `FaceDownManifested`, `FaceDownMorphed`,
  `FaceDownDisguised`, `FaceDownCloaked` and `FaceDownTurned`
  (CR 708.5: the CONTROLLER may look). `Card.FaceDownIsPermanent()` is
  that partition, and for one of these the object **is** a 2/2
  colourless creature with no name, text, subtypes or mana cost —
  whatever the card underneath says. The first four are made by a
  KEYWORD; `FaceDownTurned` is CR 708.2a's "a face-up permanent is
  turned face down by a spell or ability" — Ixidron, Backslide, Cyber
  Conversion — and is a kind of its own because CR 708.7 asks WHICH
  rules put it there (#1209, [ADR 0082's 2026-09-23
  amendment](docs/decisions/0082-casting-face-down-and-turning-face-up.md)).

Four things a card therefore never has to do:

- **Who may look is written into `KnownBy`, not kept separately.** The
  face-down landing (`applyFaceDownLandingLocked`) REPLACES the
  knowledge set with the kind's answer, and skips
  `markCardKnownInZoneLocked` — exile and the battlefield are public
  ZONES, and that skip is the only thing that makes a face-down object
  private in one.
- **The 2/2 is layer 0.** `printedCharacteristic` returns it, so
  `Effective()`, targeting, "creature you control" predicates, combat,
  the SBAs and the wire all see a 2/2 with no further plumbing. Don't
  add a layer-1 override. `PrintedIsCreature` / `PrintedIsLand` stay
  the real card: they are the CR 707.2 copiable surface.
- **The catalog is silent.** `CatalogKey` returns the EMPTY key for a
  face-down permanent, so every `Catalog*` reader answers "no entry"
  (CR 708.2a: no text). A face-down permanent runs no trigger, static,
  replacement, activated or mana ability, fires no ETB hook and has no
  printed keywords. Turning it face up needs no restore step — the
  key simply answers again.
- **`MoveCard` clears the state on every zone change** (CR 400.7), so
  a new mover is covered by the rule and not by a code review; the
  DESTINATION sets it back if the destination is itself a face-down
  state. A face-down permanent that leaves the battlefield is revealed
  through the S22 reveal frame (CR 708.9), and `ManifestForEffect` is
  the primitive that makes one (CR 701.40a). The mechanics — the
  `turn_face_up` special action (CR 116.2g), the face-down cast
  (CR 708.4), morph, disguise and foretell themselves — shipped in
  #1194 and #658.

Turning a permanent that is already on the battlefield face down is
`Game.TurnFaceDownForEffect(source, ids...)` (#1209). It is variadic
because Ixidron is the batch: every permanent is turned over before
the first event goes out. Two refusals, both "nothing happens" in the
rules and so neither an error — CR 708.2b (a face-down permanent can't
be turned face down) and CR 712.16 (nor can a double-faced one) — and
a token is NOT refused, because no rule refuses one. It is not a new
object, so counters, damage, attachments and combat all ride through;
an Aura whose enchant restriction the 2/2 no longer meets is swept by
CR 704.5m with no code of its own. Whether such a permanent can be
turned face UP again is `TurnFaceUpOffer`'s, and the answer is the
CARD's: CR 702.37e and CR 702.168d key on the card having morph or
disguise, whatever put the permanent face down.

**Life changes: "that much life" comes from a continuation, never from
a read-back (#793).** A life change runs the CR 614 window (#482), so
it can pause on a CR 616 ordering prompt exactly the way damage can
(`life_tail.go` is `damage_tail.go`'s sibling; both land a settled event
in one place that the paused path and the unpaused path share). Reading
`p.Life` on the line after changing it therefore reads a total that has
not moved yet, and the card silently drains for nothing. Same lesson as
`Scry`'s `Then`, same shape:

```go
// "Target opponent loses X life. You gain life equal to the life lost this way."
g.ChangePlayerLifeThenForEffect(src, opp, -x, func(g *game.Game, applied int) error {
    return g.ChangePlayerLifeForEffect(src, me, -applied)  // applied is negative
})

// "EACH opponent loses X life. You gain life equal to the life lost this way."
g.LoseLifeEachThenForEffect(src, ctx.Opponents(), x, func(g *game.Game, lost int) error {
    return g.ChangePlayerLifeForEffect(src, me, lost)      // lost is positive
})
```

`applied` is the post-replacement amount, and it is `0` when the change
was replaced away ("your life total can't change") or the player has
left — conceded or eliminated, including while the change was waiting on
their CR 616 prompt (#808) — the continuation is told either way, so a batch never stalls on a
leg that moved nothing. The batch form is built on the single one; don't
write your own loop that waits. A card that only says "gain 3" keeps
using `GainLife` / `ChangePlayerLifeForEffect` and needs nothing.

**Damage has the same `Then` forms, for the same reason (#807).** A
damage event runs the CR 614 window too, so "deals N damage to each
opponent. You gain life equal to the damage dealt this way" (Creeping
Bloodsucker) is the life drain's twin and reads zero the same way:

```go
g.DealDamageToPlayerThenForEffect(src, opp, n, func(g *game.Game, dealt int) error { … })
g.DealDamageToCreatureThenForEffect(src, card, n, func(g *game.Game, dealt int) error { … })
g.DealDamageEachThenForEffect(src, ctx.Opponents(), n, func(g *game.Game, total int) error { … })
```

`dealt` is the post-replacement amount and `0` when nothing landed
(prevented, Fogged, or the target gone). The batch routes each target
as the kind of thing it is — player or permanent — the way the
`DealDamage` primitive does, and is built on the single forms. A card
that only deals damage keeps using `DealDamage` / the plain
`...ForEffect` calls. There is one lint for both halves:
`life_continuation_guard_test.go` fails on a `.Life` read after a life
change and on a `.Life` / `.DamageMarked` read after a damage call, in
the same function.

**Destroy clears damage only when it lands (#708).** Marked damage is
removed by the landed outcome of a battlefield exit — not by the
destroy entry points. A destruction a replacement rewrote
(regeneration, "exile it instead", indestructible) leaves
`DamageMarked` exactly where it was: damage stays until the cleanup
step (CR 514.2), and the replacement gets to read it. If you add a
replacement that removes damage — regeneration is the one the rules
name, CR 701.15a — it does that in its own `Replace`, not by leaning on
the destroy path.

**"For each X destroyed this way" comes from a continuation too
(#815).** A destruction can pause — a commander caught in a wipe stops
to answer CR 903.9 — so the number is not knowable on the line after
the sweep. `DestroyAllMatching`'s `Then` already receives it; what
changed is that the clause now runs from the sweep's continuation
(`g.DestroyPermanentsThenForEffect`), so it may run an action later,
and its two arguments finally describe the same set: `swept` is the
pre-move copies of the permanents that were actually DESTROYED and
`destroyed` is how many of them there were. Write the clause as
something that acts on what it is handed, not as the next line of the
card. The fire-and-forget `g.DestroyPermanentsForEffect(ids)` keeps
its `int` for a sweep nothing is waiting on; it cannot include a leg
that paused, so never read it as "destroyed this way".

What counts as destroyed is CR 701.7a — "move it from the battlefield
to its owner's graveyard". A permanent the CR 614 window saved is not
destroyed (it never left), and neither is one a replacement sent to
exile, a hand or a library instead (it left, but not to a graveyard).
A commander that takes CR 903.9's offer IS counted, which is the
engine's one declared exception and lives in
`destroyedThisWayLocked`. See
[ADR 0013 §5i](docs/decisions/0013-replacement-effects.md).

**"For each X exiled / returned this way" is a continuation too, and
"this way" means ARRIVED (#866).** `ExileAllMatching`, `BounceAllMatching`
and `ReturnAllToHand` behave exactly like `DestroyAllMatching`: give one
a `Then` and it runs from the sweep's continuation
(`g.ExileCardsThenForEffect` / `g.BounceCardsToHandThenForEffect`) with
the cards that actually reached the destination. CR 400.7 decides that
— a commander that took CR 903.9's offer went to the command zone, not
to exile or a hand, so it is not in the list. (Destroy is the one verb
that DOES count the command zone, and §5i says why.) The fire-and-forget
`g.ExileCardsForEffect(ids)` / `g.BounceCardsToHandForEffect(ids)` keep
their `int` for a sweep nothing is waiting on. See
[ADR 0013 §5k](docs/decisions/0013-replacement-effects.md).

**A ONE-CARD read-back uses the same `Then` (#870).** "Exile it with a
hit counter on it", "you may exile it. If you do, return a card",
Winds of Abandon's single-target mode: one card is not a smaller
problem, because the one leg is the one that can pause. So there is no
per-card exile path to keep in step — `ExileTarget{Target: id, Then:
func(ctx, exiled bool) error}` and `g.ExileCardThenForEffect` are
wrappers over the batch, and `exiled` is the batch's CR 400.7 answer.
`ExileTarget` with no `Then`, and the bare `g.ExileCardForEffect(id)`,
stay the fire-and-forget form: their `nil` means "no error", never "it
is in exile". A card that reads the move at all reaches for the `Then`.

**A mill reports what LANDED, through the same `Then` (#893).** A mill
opens the CR 614 window per card, so a commander coming off the top
stops to answer CR 903.9 and what was milled is not knowable on the
next line. `MillToZone{…, Then: func(ctx, milled []uuid.UUID) error}`
and `g.MillToZoneThenForEffect` are the read-back — `milled` is
CR 400.7's answer, the cards that ARRIVED in the destination, so a
commander that took the command zone and a card an "exile it instead"
replacement rewrote are not in it. `MillToZone` with no `Then`,
`MillCards` and `g.MillToZoneForEffect` stay fire-and-forget: they mill
AROUND a paused card rather than waiting for it, which is right when
nothing is waiting on the answer and wrong the moment anything reads
the result. "Exile the top N cards of your library" is the same
primitive with `To: game.ZoneExile`, and it is not a mill — no
`EventMill`, no mill payoff. See
[ADR 0013 §5l](docs/decisions/0013-replacement-effects.md).

**Every arrival in a graveyard goes through the window (#931).** "If a
card would be put into a graveyard from anywhere, exile it instead"
(Rest in Peace, Leyline of the Void) is one sentence whose whole
content is *anywhere*, so a graveyard arrival that moves the card by
hand breaks a card rather than merely skipping a prompt. All of them
route now: the battlefield exit, a mill, a discard, a countered spell,
`PutIntoGraveyardForEffect`, and — since #931 — a library SEARCH with
`Dest: game.ZoneGraveyard` (Entomb, Buried Alive) and SURVEIL's
graveyard leg. Two consequences for a card that uses them. A search
finishes from a continuation, so `SearchLibrary{…, Then}` is handed the
cards that ARRIVED where it aimed them (CR 400.7) and runs an action
later when a tutored commander stops to answer CR 903.9 — `found` is
not "what I picked". And "put a card from your hand into your
graveyard" that does NOT say *discard* is not a discard (CR 701.8a
defines one by the move out of the hand under that word): reach for
`PutIntoGraveyardForEffect`, never `discardCardsLocked`, or Megrim
fires off a card that never discarded. See
[ADR 0013 §5q](docs/decisions/0013-replacement-effects.md).

**A card that exiles and then USES the card hands the rest over
(#894).** "Exile it, then return it" (`Flicker`), "exile all creature
cards from graveyards, then put all cards exiled this way onto the
battlefield" (Living Death), "exile target creature, then its
controller searches" (Path to Exile): the second half belongs in
`ExileTarget.Then` — or, for a set, in `g.ExileCardsThenForEffect`'s
continuation, which also hands over the cards that really reached
exile. Written as the next line it runs while a commander's CR 903.9
prompt is still open, which at best asks the table two questions at
once and at worst LOSES the card: the return half finds nothing in
exile, finishes, and the commander lands there a moment later with
nothing left to move it. Gate the second half on the `exiled` /
landed answer when the card says "if you do" or acts on the exiled
card, and leave it ungated when it is a separate sentence (Path's
search happens either way). See
[ADR 0013 §5m](docs/decisions/0013-replacement-effects.md).

**A sacrifice is a batch too, and "sacrificed this way" is not
"destroyed this way" (#910).** `g.SacrificeAllThenForEffect(source,
ids, then)` sacrifices a set as ONE simultaneous exit and hands `then`
the permanents that were really sacrificed;
`g.SacrificeThenForEffect(source, id, then)` is the single-card
wrapper for "sacrifice a creature. If you do, …", and
`g.SacrificeAllForEffect(source, ids)` the fire-and-forget count for a
sweep nothing is waiting on. Reach for a `Then` form the moment a card
reads the result — "draw that many cards", "for each permanent
sacrificed this way" — because a sacrificed commander stops to answer
CR 903.9 and the number is not knowable on the next line. A sacrifice
is NOT a destruction (CR 701.17a: indestructible and regeneration do
not apply), and the two "this way" rules differ in one row: CR 701.7a
defines a destruction by the graveyard, so a permanent an "exile it
instead" replacement took was not destroyed, while CR 701.17a's
sacrifice is the controller's move OFF the battlefield and that same
permanent WAS sacrificed. `EventSacrifice` still fires before the move,
so a "whenever you sacrifice" payoff is unaffected either way. See
[ADR 0013 §5r](docs/decisions/0013-replacement-effects.md).

**A PROMPTED sacrifice returns a count of QUESTIONS, and nothing may
be gated on it (#1019).** `g.PlayerSacrificesForEffect` and
`g.EachPlayerSacrificesForEffect` do not sacrifice anything — they
QUEUE a prompt per seat and return how many seats were asked. The
permanent leaves when a player answers, which is one or more actions
later, so a clause written on the next line pays out before anybody
has chosen. Use a RUN whenever ANYTHING follows the prompt, the
ordering included: `g.EachPlayerSacrificesThenForEffect(source,
except, spec, reason, then)` for the APNAP fan-out,
`g.PlayersSacrificeThenForEffect(source, players, …)` for a named set
of seats, `g.PlayerSacrificesThenForEffect(source, player, spec,
reason, count, then)` for one seat asked N times, and
`effects.EachPlayerSacrifices.Then` card-side. One run is ONE printed
instruction however many prompts it takes, and `then` runs once —
after the last seat has answered AND the permanents they named have
finished moving, so a sacrificed commander’s CR 903.9 prompt holds it
too. It is handed `game.PromptedSacrifices`: `.Sacrificed(seat)` is
"if you sacrificed a creature this way", `.Count()` is "that many",
`.By(seat)` and `.Cards()` are the permanents themselves. A seat that
was asked and sacrificed nothing is IN the answer with an empty list;
a seat with nothing to sacrifice was never asked and is absent, and
the two read the same on purpose. For a plain "ask N times" with
nothing waiting, `g.PlayerSacrificesNForEffect` — never a hand-written
loop reading the count, which the payout lint now flags. See
[ADR 0013 §5x](docs/decisions/0013-replacement-effects.md).

**A PROMPTED discard is the same RUN, and the same rule (#1027).**
`g.QueueDiscardChoiceForEffect` discards nothing either — it queues a
question over the player's own hand and hands back the prompt's ID —
and a discard can pause for an extra action after the answer, because
a discarded commander is offered CR 903.9. So anything printed after a
discard goes in a run's continuation:
`g.PlayerDiscardsThenForEffect(prompt, then)` for one seat,
`g.PlayersDiscardThenForEffect(players, prompt, then)` for a named
set, `g.EachPlayerDiscardsThenForEffect(except, prompt, then)` for the
APNAP fan-out, and `g.EachPlayerDiscardsForEffect(except, prompt)` for
a fan-out with nothing waiting. The prompt you pass is the TEMPLATE —
its `Player` is ignored by the two multi-seat forms and stamped per
seat. `then` is handed `game.PromptedDiscards`, which reads exactly
like `PromptedSacrifices`: `.Count()` is "a card for each card
discarded this way", `.Discarded(seat)` is "if you discard a card this
way", `.By(seat)` and `.Cards()` are the cards.

What counts as discarded is CR 701.8a's move OUT of the hand, so a
madness card exiled instead of binned counts, a Library of Leng card
put on top of the library counts, and a leg the CR 614 window
cancelled does not. Do NOT measure it by reading the hand size back
across the prompt — that was `b39MayDiscardThenDraw` before #1027, and
while it happens to agree on every board today it is a SECOND reading
of a rule the engine already answers, which is the kind that drifts the
first time a replacement does something new.

`DiscardPrompt.Then` is the OTHER continuation and they are not
interchangeable: it is one LEG's own sentence ("each opponent discards
a card, THEN mills a card" — Vicious Rumors), it runs per seat, it is
handed `(g, seat, discarded)` because a fan-out copies one template,
and it still fires for an empty hand ("discard your hand, then draw
three"). The run's continuation is the rest of the INSTRUCTION and
runs once. A payoff written on the leg pays out per answer, which is
what Syphon Mind did. For the RANDOM discard (CR 701.8b) the
continuation is `g.DiscardRandomThenForEffect`. See
[ADR 0013 §5y](docs/decisions/0013-replacement-effects.md).

**"If it WAS a creature card" is a clause about the exiled card, so it
waits and it is gated (#911).** Cling to Dust, Scavenging Ooze and
Deluge of the Dead. Two facts at two moments: the card's TYPE is read
BEFORE the move (after it the card is in exile with none of its
battlefield-era layers — CR 608.2h), and the CLAUSE runs from the
continuation and only when the card ARRIVED in exile. "It" is the card
the first sentence moved, so with no exile there is no "it" and neither
branch runs — Cling to Dust's "Otherwise, you draw a card" is the same
conditional's other half, not a separate sentence. Write the whole
family as `ExileThenIfItWas{Target, Was: WasCreatureCard, Then,
Otherwise}` rather than by hand; `cards/effects/exile_payout_guard_test.go`
fails the build on a condition, after a fire-and-forget exile, that
reads a local assigned before it, and its allowlist is where a clause
that is genuinely ungated (Swords to Plowshares, Solitude) gets
recorded. See
[ADR 0013 §5t](docs/decisions/0013-replacement-effects.md).

**A spell whose own text moves it off the stack is not routed again
(#489).** "Exile Ascend from Avernus", Genesis Ultimatum's "exile
Genesis Ultimatum", "shuffle this into your library": the instruction
runs in `OnResolve`, which is before the resolution frame picks the
spell's destination, so `spellMovedItselfLocked` stops the frame from
moving a card its own effect has already placed (CR 608.2n — the spell
put into a graveyard is the one ON THE STACK). Nothing on the card side
is needed: write the self-move as an ordinary `ExileTarget` or tuck on
the spell's own ID and the frame leaves it alone. A resolution that
FAILS still runs the post-resolution state checks and the CR 117.3b
priority reset, and reports itself as an `EventEffectError` rather than
as a failed pass. See
[ADR 0013 §5t](docs/decisions/0013-replacement-effects.md).

**Never call a locking accessor inside a snapshot body (#877).**
Anything that runs inside `g.ReadSnapshot(func(){…})` or
`g.WithWriteLock(func(){…})` already holds `g.mu`, and `sync.RWMutex`
is not reentrant in either mode: a second `RLock` from the same
goroutine blocks the moment a writer is queued between the two, and a
second `Lock` hangs outright. Use the lock-free `*ForEffect` accessor
(`g.PlayerByIDForEffect`, not `g.PlayerByID`) or read the fields the
body can already see — in tests exactly as much as in the engine, since
every instance found so far was a test helper and one of them was the
aiseat flake that read as a slow machine (#848 / #876).
`game/snapshot_lock_guard_test.go` is the lint: it derives the
dangerous set from the engine's own sources and fails on any of them
called inside a snapshot body.

**A `Then` clause runs even when the prompt is never answered (#865).**
A paused exit whose prompt is taken away — its chooser conceded, or the
card left by another route while the question was open — reaches the
same continuation through `abandonZoneRouteLocked`, with nothing moved
and the leg counted as nothing. So write the clause to handle an empty
list; it will always run exactly once, and "the batch stalled" is not
one of the things that can happen to it. See
[ADR 0013 §5j](docs/decisions/0013-replacement-effects.md).

**A player who leaves takes their objects with them (#769, CR 800.4a).**
Conceding or losing removes every card that player OWNS from every zone —
battlefield, command zone, hand, library, graveyard, exile and the stack —
ends the control effects they were the source of, and exiles anything of
somebody else's they were still controlling. It is not a zone change: no
`EventZoneMove`, no `EventLTB`, no dies trigger. So an effect that stashed
an instance ID and looks it up later must handle `LookupCardForEffect`
returning `ok == false`, and a "for each creature you control" predicate
must not assume a player named earlier in the same resolution still has a
board. The one exception is the departure that ENDS the game, which keeps
the final board on purpose. See
[ADR 0060](docs/decisions/0060-leaving-the-game.md).

**And EVERY battlefield exit clears it, not just a destruction
(#816).** The clear lives in `MoveCard`'s one battlefield-exit cleanup
(`clearBattlefieldDamage`, permanent_damage.go), so a creature that is
exiled, bounced, tucked, milled, sacrificed or moved by hand leaves its
marked damage and its CR 702.2c deathtouch flag behind with everything
else CR 400.7 strips — the card in the new zone is a new object, and a
creature that comes back (replayed, reanimated, blinked) must not
arrive already damaged. Last
known information is unaffected: `snapshotLKILocked` runs while the
permanent is still on the battlefield, one line before the move, so a
dies-trigger reads the creature that died (the LKI `Characteristic` has
never carried marked damage, and the `source` card a trigger is handed
is the new object, which by CR 400.7 has none). Don't clear damage in a
card's effect: if your card leaves the battlefield, it is already done.

**Per-object state the ENGINE keeps is cleared by the other half of
that exit (#630).** `MoveCard` is a package-level function over two
zones, so it cannot reach a map on `Game` — and those maps are keyed by
instance ID, which survives a zone change. `battlefieldExitLocked`
(`server/internal/game/battlefield_exit.go`) is the Game-side half: it
takes the LKI snapshot and then forgets what the leaving OBJECT did —
`LoyaltyActivatedThisTurn` (CR 606.3, so a planeswalker bounced and
recast the same turn may activate again) and the combat announcement
maps. All three battlefield exits call it, and a new per-object
registry goes in it rather than growing a fourth clearing site. What
deliberately stays is `TurnTally`'s per-ability counts, and since #936
for a better reason than "it would break the loop breaker": the
card-facing gates (`Resolved`, `Triggered`) are keyed by
`ObjectTallyKey(source, Card.ObjectEpoch, label)`, so the returning
permanent reads a key nothing has written and its "only once each
turn" clause fires again with nothing deleted. `LoopRun` /
`LoopAllowance` keep the per-CARD `TallyKey`, because a blink loop
leaves and re-enters on every iteration and clearing their count there
would give ADR 0055's breaker an escape hatch. The epoch is bumped in
`MoveCard` next to the rest of CR 400.7's forgetting.

**Paying life is a cost, and a cost may not pause.** Use
`g.PayLifeForEffect(source, player, n)` for "pay N life" — a ward, a
shockland, an activation cost, "pay 2 life. If you do, draw". CR 119.4
makes the payment a life loss, so the window still runs and a life-loss
replacement still sees it; what the cost path adds is that it settles in
one step, because CR 601.2h pays a spell's costs as one indivisible step
and a half-paid cost cannot be rewound. A payment the window CANCELS
("your life total can't change") is not paid for free: CR 119.8 and
CR 614.17b say that cost can't be paid, so `PayLifeForEffect` returns
`ErrInvalidParam`. If you write a card that stops a player losing life,
also make the cost validators it reaches refuse the payment up front.
See [ADR 0013 §5b and §5e](docs/decisions/0013-replacement-effects.md).

The answer is `{bottom, top_order}` with `top_order` **top-first**, and
every looked-at card must appear in exactly one list: scry moves all of
them, so an answer that omits one is a client bug, not shorthand for
"leave it".

**"In any order" (#996, ADR 0088):** "put the rest on the bottom of your
library in any order", "put two cards from your hand on top of your
library in any order" and "its owner puts it on their choice of the top
or bottom of their library" are one prompt, `put_in_library`, reached
through `PutInLibraryInAnyOrder{Cards, From, Placement}`
([library_order.go](server/internal/cards/effects/library_order.go)). For
the common "look at N, take one, the rest on the bottom in any order"
sentence use `TakeRestOnBottomInAnyOrder` as the `TakeFromLibraryToHand`
`Then` — **not** `TakeRestOnBottomInRandomOrder`, which is only for cards
that print "a random order". Brainstorm's put-back is
`PutFromHandOnTopInAnyOrder`. Do not borrow `Scry` for a top-or-bottom
choice: scry is a keyword action, and borrowing it fires "whenever you
scry" payoffs. A position with no choice in it ("second from the top") is
`PutIntoLibrary{Depth: 2}`, and after a search it is
`SearchLibrary{ToTop: true, Depth: 3}`.

The same prompt covers four more shapes (#1298, ADR 0088's 2026-09-23
amendment), each with a sentence helper in the same file:
`LookAtLibraryThenPlace{Owner, N, Placement}` for a look at ANOTHER
player's library (Jace, the Mind Sculptor's +2, Portent — only the
looker learns the cards); `PutIntoLibraryAtDepthOrBottom{Card, Depth}`
for "its owner puts it second from the top or on the bottom" (Temporal
Cleansing — the OWNER chooses); `TopCount: 1` for "put one of those
cards on top and the rest on the bottom" (Cream of the Crop); and
`CounterToLibrary{StackID, Placement}` for "if that spell is countered
this way, put it on the top or bottom of its owner's library" (Hinder,
Spell Crumple). The last one asks FIRST and counters second; do not
write it as a counter to the top followed by a prompt.

**A rule about the chosen cards as a set (#624):** when a card-set
pick says something no count and no per-card list can ("discard two
cards unless you discard a creature card", "two lands that share a land
type"), put it on the prompt's `Validate`, never in `Then`:

```go
g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
    Chooser: c, Question: "Discard two cards unless you discard a creature card",
    Cards: hand, Min: 1, Max: 2, Zone: game.ZoneHand,
    Validate: func(picked []game.Card) bool {
        return len(picked) == 2 || picked[0].IsCreature()
    },
    Then: discardThem,
})
```

`SearchLibrarySpec.Validate` is the same hook for a search. Both run
before the prompt is dequeued, so a refused set comes back to the
player as an error (`ErrChoiceSetRejected` for a choose-cards prompt)
with the prompt still open. And `internal/legal` asks the same hook
(`ChooseCardsPickLegalLocked` / `SearchPickLegalLocked`) before it
offers a bot a set. A rule checked only inside `Then` is the #544
wedge: the enumerator offers the set, `Then` refuses it after the prompt
is gone, and the card resolves wrong with nothing left to retry.

`Validate` gets the picks as live `Card` values and no `*Game`, because
the enumerator calls it under the read lock. Anything else the rule
needs, like "or your whole hand if it has fewer than two", is a value
you capture when you queue the prompt, the same way `Cards`, `Min` and
`Max` are. It is never called for an empty pick, so a `Min: 0` prompt
always keeps "choose nothing". With `Min` above zero, don't queue a
prompt that no set can satisfy: nothing could answer it, and the
enumerator logs it rather than inventing an answer.

**A DISCARD says it through `DiscardPrompt`, not a raw pick** — the
prompt has the same `Validate`, and going through it is what keeps the
discard on the one discard path (CR 614 window, CR 903.9, madness,
`EventDiscardCard`) and on the RUN (#1027, above). Its floor is `Min`,
and that is the field the "unless" clauses need (#626):

```go
g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
    Player: who, Source: src, N: 2, Min: 1,   // two cards, or ONE creature card
    Question: "… discard two cards, unless you discard a creature card",
    Validate: func(picked []game.Card) bool {
        return len(picked) == 2 || (len(picked) == 1 && picked[0].IsCreature())
    },
})
```

Do NOT reach for `UpTo: true` to make room for the one-card answer.
`UpTo` drops the floor to ZERO, and `Validate` is never asked about an
empty pick — so the clause could be answered by discarding nothing.
That was a live bug on Compulsive Research until #626. `Min` is
clamped to the hand size, which is CR 701.8a's "as many as you can",
and the number your `Validate` compares against must be that same
clamped count, captured when you queue.

**"Put [it / a card from among them] onto the battlefield" off a
library (#745):** a reveal or a look followed by a put is not a search,
so never reach for `SearchLibrary` with a predicate (it emits
`EventSearchLibrary` and shuffles). Say the first half with the right
visibility, then hand the cards to `PutFromLibraryOntoBattlefield`
(`put_from_library.go`):

```go
looked := g.LookAtTopOfLibraryForEffect(controller, 8)   // "look at": only the looker knows
// revealed := g.RevealTopOfLibraryForEffect(...)         // "reveal": every seat knows
return PutFromLibraryOntoBattlefield{
    Cards: looked, Match: OfCreatureType("Dragon"),
    Max: 1, Optional: true,                 // "you may put a"; Max 0 is "any number"; All for "put all"
    Then: PutRestOnBottomInRandomOrder,     // or PutRestIntoGraveyard, or your own
}.Apply(ctx)
```

The prompt is asynchronous, so "the rest" goes in `Then` — it is the
only place that knows which cards were not chosen. Several picks enter
as one simultaneous batch (`PutCardsFromLibraryOntoBattlefieldForEffect`),
so don't loop the single-card move over them. The whole-sentence
shapes are named: `LookAtTopThenMayPutOntoBattlefield` (Ureni) and
`RevealUntilThenPutOntoBattlefield` (The Regalia).

**"Put it into your HAND" is the twin, and it is a different door
(#952).** `effects.TakeFromLibraryToHand` — same fields, same `Then`,
with a `Reveal` flag for "you may REVEAL a creature card from among
them" (only the TAKEN cards become public; the rest of a private look
stays private) — over the engine's
`Game.TakeFromLibraryToHandThenForEffect`. Whole sentences:
`LookAtTopThenMayTakeToHand` (Horn of the Mark) and
`RevealTopThenTakeToHand` (Goblin Ringleader). **Never `BounceToHand`
for this.** It appears to work only because the zone router finds a
card's zone by scan; "return it to its owner's hand" is not "put it
into your hand off the top of your library", the fire-and-forget form
drops a paused CR 903.9 leg from the accounting, and nothing watching a
bounce should see a library take. `TakeRestOnBottomInRandomOrder` and
`TakeRestIntoGraveyard` are the two rests.

"Put the rest on the bottom in a random order" anywhere else is
`g.PutOnBottomInRandomOrderForEffect(actor, from, ids)`, which draws
from the game's keyed RNG (`random_order` stream, ADR 0054) — never
`math/rand` — and repositions cards already in the library without a
zone change. `from` is the zone the effect left the cards in
(`ZoneLibrary` for a reveal or a look, `ZoneExile` for cascade): IDs
saved before a prompt may name cards that have since moved on, and
those are skipped rather than pulled back.

A token can end up in a library (Chaos Warp tucks one, and the engine
has no CR 704.5d sweep). It is not a card (CR 108.2) and can't change
zones again (CR 111.8), so the move refuses it and the helpers never
offer or stop on one. If your card moves a revealed card anywhere else
("otherwise put it into your hand"), skip a token with `IsToken`, as
Coiling Oracle and Risen Reef do.

**"This permanent enters tapped" (S21):** declare a self-replacement,
not an entry-hook tap:

```go
Replacements: []game.ReplacementEffect{SelfEntersTapped()},
```

The two are observably different, which is why the machinery exists: a
hook tap means the permanent enters UNTAPPED and is tapped a beat
later, emitting `EventTapCard`, so anything watching for a tap or for an
untapped permanent entering sees the wrong thing. A replacement emits
none. (Worn Powerstone used the workaround and said so in a comment; it
now uses the real thing, and the test pins the difference by counting
tap events rather than by checking `Tapped`, which both approaches
satisfy.)

This needed an engine change worth knowing about: the replacement
pipeline runs **pre-push**, so an entering card is not on the
battlefield and the ordinary catalog walk in
`gatherActiveReplacementsLocked` cannot find its own effect.
There is now a third gathering block that consults the ENTERING card's
own replacements, passing the card itself as `source` so an `AppliesTo`
comparing `ev.CardID` to `source.InstanceID` identifies "this
permanent". It is skipped for a card already on the battlefield, so a
permanent in play can never match both blocks and apply the same effect
twice.

Lands may carry triggers and mana abilities like any other permanent —
the ten-Temple cycle in `temples.go` combines all three (enters tapped,
an ETB scry trigger on the stack, pipe-syntax dual) and is written as a loop over a table, since
ten near-identical files is ten places to fix one mistake.

**"…unless" and "if you don't": a conditional tapland is one of three
clauses, and picking the wrong one is the usual mistake.** All three
are entry replacements on `Spec.Replacements`; what differs is WHO
decides and WHAT the decision costs:

| Printed text | Clause | Who decides |
|---|---|---|
| "enters tapped unless you control a Swamp" (checkland, battle land, bond land) | `SelfEntersTappedUnless(cond)` ([tapland_helpers.go](server/internal/cards/effects/tapland_helpers.go)) | nobody — the board does |
| "you may pay 2 life. If you don't, it enters tapped" (shockland) | `EntersTappedUnlessYouPayLife(name, 2)` ([shocklands.go](server/internal/cards/effects/shocklands.go)) | the player, for life |
| "you may reveal an Island or Swamp card from your hand. If you don't, it enters tapped" (reveal-land) | `EntersTappedUnlessYouRevealFromHand(name, clause, matches)` ([reveal_lands.go](server/internal/cards/effects/reveal_lands.go)) | the player, for nothing |

A CONDITION goes in `AppliesTo`, never inside `Replace`, so a land that
meets it contributes no applicable replacement at all — otherwise CR
616 asks the controller to order an effect that was always going to do
nothing. A DECISION is the opposite: the replacement is always
applicable and the question lives inside it, because a player may
decline even when they could say yes (bluffing an empty hand is a real
play, and the shockland at 20 life may still not want to pay).

The reveal clause is `game.EntryHandReveal{Matches, Min, Max, Question,
Then}` (#1198, [ADR 0013](docs/decisions/0013-replacement-effects.md)
§5z). `Matches` is a `func(game.Card) bool` over a card in HAND, so it
reads printed characteristics — `IsLandWithSubtype("island")` and
friends — and it must not take a `*Game` (it runs under the read lock
inside the bot enumerator as well as under the write lock on submit).
Revealing costs nothing and moves nothing (CR 701.20b): the card stays
in hand, the whole table becomes entitled to read it, and the engine
does the reveal itself. Do not write a prompt in the card file —
there is no per-card prompt code in this family, the same way there is
none in the shockland one.

**An alternative cast cost (S22):** "you may cast this spell for its
<keyword> cost **rather than** its mana cost" (CR 118.9) goes in
`Spec.AlternativeCosts`, built from the constructors in
[alternative_cost.go](server/internal/cards/effects/alternative_cost.go).
This is NOT `AdditionalCost`, which is a cost paid *alongside* the mana
cost:

```go
AlternativeCosts: []game.AlternativeCost{Overload("{4}{R}")},        // Vandalblast, Cyclonic Rift
AlternativeCosts: []game.AlternativeCost{Evoke("{3}{U}")},           // Slithermuse
AlternativeCosts: []game.AlternativeCost{                            // Wash Away
    Cleave("{1}{U}{U}", TargetSpell("target spell")),
},
```

One constructor per keyword rather than a generic builder, because each
keyword bundles a rewrite with its price: overload also **deletes** the
target clause (`ClearsTargets`), evoke also attaches the
sacrifice-on-entry trigger, cleave **swaps** the target clause for the
wider bracketed-words-removed one. Hand-rolling
`game.AlternativeCost{ManaCost: "{4}{R}"}` compiles, casts for four, and
still demands a target — a strictly worse Vandalblast that looks right.

The `Key` is the wire contract: it rides `cast_spell` as
`alternative_cost`, lands on `StackItem.AltCost`, and the card's
`OnResolve` branches on `ctx.PaidAltCost("overload")`. Keys must be
non-empty and unique per card; `Register` panics otherwise. Only
overload / evoke / cleave / flashback / warp / escape exist —
spree has no shape yet, and a card carrying it ships without it (say
so in the card comment). Preparation cards are a layout, not a cost —
see "Adding a preparation card" below. Foretell,
suspend and plot are not alternative costs at all: they are CR 116.2
special actions, declared in `Spec.SpecialActions` with
`effects.Foretell`, `effects.Suspend` and `effects.Plot` (ADR 0062
Decision 4).

**Casting from somewhere other than hand (S29):** a card whose text
opens another cast zone declares it in `Spec.CastableZones`, and the
price of that path rides `AlternativeCost.FromZone`:

```go
CastableZones:    []game.ZoneKind{game.ZoneGraveyard},          // Faithless Looting
AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
```

The zone is the **place** and the alternative cost is the **price**,
and they are checked independently. Hand is implicit and never has to
be listed — declaring the graveyard *adds* a path. An offer bound to a
zone can only be claimed from that zone, and a zone that has a bound
offer can only be cast from by claiming it (so Faithless Looting cannot
be flashed back for its printed `{R}`); a zone with no bound offer
charges the printed cost, which is Gravecrawler. `Register` panics on
an offer whose `FromZone` is not in `CastableZones`, because such an
offer is unclaimable.

Use the keyword constructor, never a hand-rolled `game.AlternativeCost`,
for the same reason overload and evoke have one: `Flashback` bundles
**three** things — the price, `FromZone: ZoneGraveyard`, and
`ExileOnLeavingStack`. The last is CR 702.34a's "exile this card
instead of putting it anywhere else any time it would leave the
stack", and it is a *replacement*, so it also catches a flashed-back
spell that fizzles and one answered by Hinder. A card that wrote the
cost by hand would flash back, land in the graveyard, and flash back
again every turn forever.

**Escape (S29)** is flashback's sibling and the place to look when a
cost needs a component the struct doesn't have yet. `Escape("{3}{B}",
5)` is "Escape—{3}{B}, Exile five other cards from your graveyard",
and `EscapeWithCounters("{5}{G}{G}", 4, 3)` adds CR 702.138c's "this
creature escapes with three +1/+1 counters on it".

Three things it added to `AlternativeCost`, all of them because escape
is a *price* rather than a permission:

- **`ExileFromGraveyard`** is the first cost component that names more
  than one card. The count is the spec's `Min` (== `Max`), the caster
  sends all of them in `alt_cost_ids`, and the engine demands exactly
  that many, all distinct, all in the caster's own graveyard. "Other"
  needs no clause of its own — CR 601.2a has already moved the spell to
  the stack by the time the cost is paid.
- **`EntersWithCounterName` / `EntersWithCounterCount`** hang the
  counters off the **cost**, not the card, so a reanimated or
  hard-cast Voracious Typhon enters as the 4/4 it prints. They ride the
  CR 614 entry pipeline, so Doubling Season doubles them.
- **No `ExileOnLeavingStack`.** This is the one to get right: an
  escaped card goes to the battlefield or the graveyard like any
  other, and escapes again next time. Copying flashback's constructor
  and swapping the key would ship a card that exiles itself, which is
  not what any escape card does.

**"Unless it escaped" — cast provenance (#653, CR 400.7d).** A
permanent remembers how the spell that became it was cast:
`Card.Provenance` is a `CastProvenance{AltCost, FromZone}` written at
the one place a spell becomes a permanent and cleared by `MoveCard`
when it leaves the battlefield (`game/cast_provenance.go`). Card code
reads `ctx.Escaped()` (CR 702.138b) or `ctx.CastProvenance()`, and
`SacrificeThisUnlessItEscaped("Phlage")` is the Titan cycle's whole
entry clause in one constructor.

**Which object you are asking matters, and this is the trap.**
`ctx.PaidAltCost("overload")` asks about the ITEM BEING RESOLVED —
right for an overloaded Cyclonic Rift, wrong for anything a PERMANENT
asks, because by then the item on the stack is the trigger and the
spell that paid escape finished resolving two steps ago.
`ctx.Escaped()` asks the source permanent. Get them the wrong way
round and the card compiles, casts, and answers "no" forever.

The record is per-ENTRY, exactly like `NamedTribe` and `ChosenColor`:
a Phlage that escaped, died and was reanimated is a new object that
did NOT escape (CR 400.7), which is the difference between the card
and an infinite loop. It grows a field when a card needs a fact the
permanent cannot re-derive — #664's "if it was kicked" is the next
one — and never a second record.

**Warp (S29)** is the other half of the same idea and the reason the
zone and the price are separate fields. `Warp("{R}")` is paid from
**hand**, so it needs no `CastableZones` at all — the discount is now,
the real card is later. Its constructor bundles `WarpExile`, which
schedules a CR 603.7 delayed trigger to exile the permanent at the
next end step and leaves a `game.CastPermission` behind carrying a
`NotBeforeTurn` floor for "on a later turn". The later cast is then an
ordinary cast from exile for the printed cost, through the button the
impulse-exile grant already renders.

Three zones are **not** card properties and must not be declared:
`ZoneCommand` (CR 903.4 grants that to the format); `ZoneExile`, whose
permission belongs to one exiled *instance*; and `ZoneLibrary`, whose
permission belongs to a POSITION (the top card) rather than to a card.
All three ride a granted permission instead — see the next paragraph.
Declare `ZoneExile` only when the card's own printed text grants the
cast to every copy, at any time, however it got there.

**Granted permissions (S42, [ADR 0066](docs/decisions/0066-granted-cast-and-play-permissions.md)):**
when an EFFECT rather than a card's own text opens a zone — Snapcaster
Mage giving flashback, Underworld Breach giving escape, Bolas's Citadel
opening the top of your library, impulse exile, airbend, warp, cascade
— it is one type, `game.CastPermission`, and there is deliberately no
second one. Two shapes, and which you want is decided by whether the
permission is a permanent's static ability or an effect that resolves:

```go
// A permanent's static ability: a STANDING permission, derived from
// the battlefield on every query, so two of them compose and one
// leaving cannot revoke the other's. No duration to expire.
CastPermissions: []game.CastPermission{{
    Zone:                    game.ZoneGraveyard,
    Filter:                  game.PermissionFilter{NonLandOnly: true},
    AltCostKey:              "escape",
    ExileOtherFromGraveyard: 3,
}},                                              // Underworld Breach
LibraryTopVisible: game.LibraryTopRevealed,      // Oracle of Mul Daya
CastPermissions:   []game.CastPermission{PlayFromTopOfYourLibrary(
    game.PermissionFilter{LandsOnly: true}, "Play a land from the top")},

// An effect that resolves: the set of card OBJECTS is locked NOW
// (CR 611.2c), so a card that reaches the graveyard afterwards has
// nothing.
GrantFlashbackToCard{Target: id}.Apply(ctx)            // Snapcaster
GrantCastFromYourGraveyard{                            // Past in Flames
    Filter: game.PermissionFilter{InstantOrSorceryOnly: true},
    AltCostKey: "flashback", ExileOnResolution: true,
}.Apply(ctx)
```

Five rules worth knowing before you write one. **The key is shared with
the printed keyword** (`"flashback"`, `"escape"`), because CR 702.34a's
"if the flashback cost was paid, exile it" and CR 702.138b's "escaped"
read the key however the permission arrived — and a card that both
prints and is granted the same key keeps its **printed** cost. **A
permission never opens the sorcery-speed gate** unless it sets
`Timing`; a sorcery in your graveyard is still a sorcery. **CR 400.7 is
free**: a permission names `{instance, epoch}`, so it ends the moment
its card leaves the zone by any route, and nothing has to clear it. **A
library permission needs its visibility half too** (`LibraryTopVisible`
— `LibraryTopOwner` for "you may look at the top card any time",
`LibraryTopRevealed` for "play with the top card revealed"): a card you
cannot see is a card you cannot play, and every printed card in the
family carries both clauses. **A permission may name FACES**
(`Faces []int`, ADR 0034): empty means it does not speak about faces
and the card's own `CastableFaces` decides; a list NARROWS to those and
no other, and a list of one is also the ANSWER — the caller's requested
face is ignored, because there is exactly one legal cast. A defeated
Siege's grant is `Faces: []int{1}` and CR 715.4's Adventure grant is
`Faces: []int{0}`, which is why it is a list: zero cannot mean both
"no opinion" and "the front face".

**Adventure cards (CR 715, #719, ADR 0034 step 6).** A card file writes
nothing for the lifecycle — the engine owns it
([game/adventure.go](server/internal/game/adventure.go)). What a card
file does is register **both faces**: the creature under the bare
oracle ID and the Adventure half under `"<oracle_id>#1"`, the same
composite keyspace a Siege's back face and the sixty MDFC land backs
live in. See
[foulmire_knight.go](server/internal/cards/effects/foulmire_knight.go).
The creature's entry is usually a bare `Completeness` declaration plus
its `PrintedKeywords` — without it the whole card wears the
"unimplemented" badge in hand, because the Adventure's text makes
`NeedsCatalogEffect` true for the card. **A TARGETED Adventure half is
no longer blocked** (#992): the view publishes the whole announce
surface — `target_mode`, `legal_targets`, `clauses`, the modes and the
price list — per castable face, and the client's face picker swaps the
chosen half's block in, so Stomp, Petty Theft and Swift End open a
target picker like any other spell. See
[bonecrusher_giant.go](server/internal/cards/effects/bonecrusher_giant.go).
Nothing about a card file changes for it: write the Adventure half's
`Targets` exactly as you would on a single-faced instant, under the
`"#1"` key. An uncatalogued adventure card is still unaffected — it has
no announce data on either face and resolves by hand.

**The window is a `game.Duration`** (#945,
[ADR 0063](docs/decisions/0063-durations-and-control.md)) — the same
vocabulary every continuous effect in the engine uses, swept through the
same `durationExpiredLocked`, so there is exactly one answer to "is this
still live" and one function that gives it
(`g.CastPermissionActiveForEffect(perm, player)`). Write the clause, not
a turn number:

```go
Duration: /* leave zero */                          // "until end of turn"
Duration: g.UntilEndOfYourNextTurnDuration(who),    // "until the end of your next turn"
Duration: game.WhileInZoneDuration(),               // "while it remains exiled"
NotBeforeTurn: g.Turn.Number + 1,                   // warp / foretell's "on a LATER turn" — a floor, not a duration
```

The zero value is "until end of turn", stamped against the current turn
by the one write path, so a card file that forgets gets the shortest
window rather than an unbounded grant. A permanent's `Spec.CastPermissions`
needs none at all: `standingCastPermissions` forces
`WhileInZone` on it, because a derived permission's duration is the
source's presence on the battlefield. And **never** schedule a delayed
trigger to shorten a grant — "until the end of your next turn" used to
need one and does not any more.

**One list of the prices a cast may claim (CR 118.9, #673):** "which
costs may this seat announce for this card out of this zone" has one
answer, `game.CastOffersForLocked`
([cast_zones.go](server/internal/game/cast_zones.go)). A nil entry is
the printed mana cost — present only when the cast path allows a claim
of nothing, which is what keeps a Faithless Looting in the graveyard
off its printed `{R}` — and the rest are the card's own zone-bound
offers plus the one a grant synthesises, in announce precedence and
filtered through `AlternativeCostPayableLocked` — the same #695
predicate the view's offer stamp and `CastSpell`'s own validator read
(Condition, CR 119.4's life, CR 601.2b's card component; mana is
deliberately NOT asked, because CR 601.2g lets the caster tap
afterwards). The bot enumerator walks it; a card file adds an offer and
every price surface follows. Do not re-derive "what can this be cast
for" anywhere else.

Since #1012 the VIEW reads it too — `protocol.stampCastOffers` — so
`alternative_costs` on the wire is that list verbatim, and the
projection decides only how each entry LOOKS. Two things fall out of
the same call and are derived nowhere else: `castable_here` is
`no cant_cast && at least one claimable price` (#1015 — an empty list
is a real "this card is not castable from here", not a card with no
offers), and `alternative_cost_required` is "the nil entry is missing",
i.e. the printed cost is not claimable from this zone. If you are
writing a second answer to either, you are writing the bug those two
issues were.

**A permission names an OBJECT, not a pile (#1022).** A `ScopeCards`
permission can name a card in ANOTHER seat's graveyard — Wrexial's
"cast target instant or sorcery card from that player's graveyard" —
and three surfaces used to answer "whose graveyard" by accident:
`CastSpell` resolved `from_zone: "graveyard"` to the caster's own pile,
the enumerator walked the caster's own, and the view asked each seat
about its own. All three ask `CastPermissionForLocked` now, and the
card is reachable wherever it sits **only** under a permission — a
card's own text (flashback, escape, Gravecrawler) opens its OWNER's
graveyard and nobody else's. One consequence worth knowing before you
touch it: a STANDING permission is refused over a pile its holder does
not own, because a permanent's printed text says "your graveyard" and
`PermissionFilter` has no clause for it.

**And the LIBRARY is the same sentence (#1035).** `permissionPositionOKLocked`
enforced CR 401.5's "the top card of your library" against the
permission HOLDER's own pile, which scoped every printed library clause
by accident and made a cross-seat permission open nothing at all — the
position rule reads the library the CARD is in now. Three things ride
on that and are easy to get wrong:

- `CastPermission.ZoneOwner` names the seat whose pile a STANDING
  permission is over, and it is the only way past the "your own pile"
  scoping. A DERIVED permission can never carry one — the catalog is
  static, `standingCastPermissionsLocked` zeroes it, and the view leans
  on that to rule out a foreign holder without walking the battlefield.
  "Each opponent's library" is one stored permission per opponent,
  granted at resolution.
- `CastPermission.SeesLibraryTop` is CR 401.5's LOOK for a cross-seat
  grant, because `CardDef.LibraryTopVisible` is per library and derived
  from the permanents its OWNER controls, and cannot express "you may
  look at the top card of THEIR library". Ask
  `LibraryTopVisibleToLocked`, never the enum: it is the one place both
  spellings are read, and `LibraryTopKnowersLocked` is built out of it
  so the seat that may cast the card is a seat the view has already
  made a knower of it.
- A grant that opens another seat's library top without the look opens
  nothing. A card you cannot see is a card you cannot play.

**The cast stamps are PER HOLDER (#1037).** `CardView.castOffers` is a
map of seat → `castStamps` and the exported fields carry only the
PUBLIC answer: the pile owner's own cast out of their own graveyard or
library, and nothing at all for exile, which has no owner.
`stampLegalTargets` stamps the public one, `stampGrantedPermissions`
files every other holder, `FilterViewFor` promotes exactly one. Two
seats may hold two permissions over one card and both get a picker. Do
not write a seat's answer into a `CardView` field — that is the bug
#1037 was, and `castable_here`'s meaning depends on it: public on a
per-seat pile, so a CLIENT must read the bit together with whether
`exile_play` (also resolved per viewer now) names them.

**`{X}` and a free cast (CR 107.3b, #831):** a spell with `{X}` in its
mana cost, cast while paying neither that cost nor an alternative cost
that includes `X`, has exactly one legal `X` and it is `0` — cascade's
`{0}` grant, a Siege's free cast and an offer priced `{R}` are all the
same answer. One predicate says so, `game.CastCost.LocksXAtZero`
([cast_cost.go](server/internal/game/cast_cost.go)), applied
once at CR 601.2b in `CastSpell`; a cost *reduction* never triggers it,
and a card file needs no flag for it.

**What a cast costs: one pricer, one entry point (#696).** Nothing
outside `internal/game` may re-derive a cast's price. `g.PriceCast`
(`g.PriceCastForEffect` under `g.mu`) takes the **announcement** — the
same `game.CastSpellParams` the cast would send — and answers a
`game.CastPrice`:

```go
price, err := g.PriceCast(playerID, card, game.CastSpellParams{
    FromZone: "graveyard", AlternativeCost: "flashback",
})
price.Paid   // the cost STRING this cast pays: "{2}{R}"
price.Base   // the total before convoke/waterbend spends against it
price.Total  // what the payment charges — what applyCastCostLocked demands
price.Card   // the card with the announced face materialised (ADR 0034)
```

It is the same walk `CastSpell` charges with, so it settles the CR
118.9 swap, a granted permission's flat override, the "spend mana as
though any colour" fold, the commander tax, the mana half of the
announced optional costs, the cost modifiers and the convoke
subtraction — in that order, once. The auto-tap preview endpoint
(`GET /games/{id}/auto-tap-preview`) and the bot enumerator both call
it; each of them used to keep a partial copy, and each copy disagreed
with the engine on every non-hand and alternative-cost cast. Add a
component to the price in `printedCostLocked` or
`costAfterModifiersLocked` and all four readers get it.

**An offer is offered only when it is payable (#695).**
`g.AlternativeCostPayableLocked(caster, castID, offer)` is the one
predicate behind "is this alternative cost on the table": its
`Condition`, CR 119.4's life (exactly N is payable — paying down to
zero is legal), and CR 601.2b's card component (enough matching cards
in the right zone, never the spell itself). The view's offer stamp,
the bot enumerator and `CastSpell`'s own validator all read it, so a
shown offer, an enumerated move and an accepted cast cannot disagree.
**Mana is deliberately not part of it**: CR 601.2g lets the caster tap
for it after the cost is chosen, which is what the auto-tapper is for.

**A delayed trigger (S22):** "at the beginning of the next end step,
<do X>" (CR 603.7) is `ScheduleDelayedTrigger`, not a closure that runs
now:

```go
ScheduleDelayedTrigger{
    Label: "Waterbender's Restoration — return the exiled creatures",
    Cards: exiled,                    // instance IDs, stamped onto the fired item's Targets
    Body:  returnExiledToOwnersBody,  // a registered game.BodyRef (delayed_bodies.go)
}.Apply(ctx)

ScheduleDelayedTrigger{              // a body with data: what a factory used to capture
    Label:  "Mana Drain — add {C} × 3", At: game.StepPrecombatMain, ControllerTurnOnly: true,
    Body:   manaDrainRefundBody,
    Params: game.EffectParams{Amount: mv},
}.Apply(ctx)
```

**What a delayed trigger does is DATA** ([ADR 0041](docs/decisions/0041-game-persistence.md)
phase 3, tier 2, #1497). `Body` is a `game.BodyRef` — a key into the
registry in `game/effect_bodies.go` — never a function, so a func
literal there does not compile. Register a new body ONCE, in
[delayed_bodies.go](server/internal/cards/effects/delayed_bodies.go),
as `game.SimpleDelayedBody("<area>/<name>", fn)` (or `game.DelayedBody`
for one that reads `game.EffectParams`: `Player`, `Object`, `Amount`,
`Cost`, `Name`, `Filter`), and append the key to the ledger with
`cd server && go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys`.
Keys are on-disk identities, like token slugs: never renamed, never
deleted — `game.EffectAlias(old, new)` keeps an old one resolving. A
table with a delayed trigger waiting is therefore a restore point, and
so is one whose fired trigger is on the stack.

`At` defaults to `game.StepEnd`; the queue is drained on step **entry**,
so an ability scheduled during an end step waits for the following one.
Set `ControllerTurnOnly: true` when the printed text says "at the
beginning of **your** next <step>" (Mana Drain): a matching step on
another player's turn then leaves the trigger queued. Leave it unset
for "the next turn's upkeep" (Arcane Denial), which the very next
upkeep at the table satisfies.
The instruction lives on the `Game`, not on a card — the spell that
created it is usually in a graveyard by the time it fires — and it goes
on the stack when the step begins, so every player gets a response
window. The body reads its payload off the item it is handed
(`item.Targets`, `item.Controller`) and its data off its params, and
captures nothing — it cannot.

**Event-conditioned delayed triggers (#663, CR 603.7b):** "When you
next cast an instant or sorcery spell this turn, copy that spell"
(Doublecast, Galvanic Iteration) waits for a THING TO HAPPEN rather
than for a step, and it is the same queue with a different condition —
`WhenYouNextCast(label, filter, body)`, or the general
`DelayedOnEvent{Label, On, Condition, CondParams, Body, Params}`, both in
[delayed_on_event.go](server/internal/cards/effects/delayed_on_event.go).
The condition is a registered `game.ConditionRef` and the spell filter is
DATA, a `game.CastFilter` — not a `CardPredicate`, which is a closure:

```go
return WhenYouNextCast("Doublecast — copy that spell",
    game.CastFilter{Types: []string{"Instant", "Sorcery"}}, copyTheSpellBody).Apply(ctx)
```

Four things it gets for free and must not re-implement. It fires
**once** and is removed (CR 603.7b), from one hook at the end of
`triggerHarvester.OnEvent`. It ends **with the turn** whether or not it
fired (CR 514.2), carrying ADR 0063's `Duration` and swept beside the
scoped statics — hand it a `Duration` only when the card says something
other than "this turn". It never sees the cast
that **created** it, because the `EventCast` of that spell was emitted
before the resolution that scheduled it. And the fired trigger goes
through `dispatchTriggerLocked`, the harvester's own dispatch, so the
CR 603.5 "you may", the CR 603.3d drop and the APNAP drain are the same
code an ETB uses. The triggering event's object rides on the item as
`Payload` — `ctx.PayloadCards()[0]` is "that spell" — so the body
reads it off the item and captures nothing. This reverses
[ADR 0026](docs/decisions/0026-delayed-triggers.md) §1-2 for this one
case; the 2026-09-18 amendment there is the record.

**A reflexive trigger (CR 603.12, #636):** "<do something>. **When
you do**, <do something else>" — Ziatora's fling, an Overlook land's
fetch, Invasion of Tarkir's damage. The second sentence is a trigger
created by the first one *while it resolves*, and it is
`ReflexiveTrigger`, applied from inside the parent's `Effect` once the
condition actually held:

```go
ReflexiveTrigger{
    Label:   "Ziatora, the Incinerator — damage equal to the sacrificed creature's power",
    Targets: TargetAny(),          // chosen when the trigger goes on the stack
    Cards:   []uuid.UUID{killed},  // the payload; read back with ctx.PayloadCards()
    Effect:  b29ZiatoraFling,      // a package-level func, NOT a closure
}.Apply(ctx)
```

`WhenYouDo(label, effect)` is the plain mandatory, untargeted case.
Both go through the harvester's own dispatch
(`Game.QueueReflexiveTriggerForEffect`), so the trigger gets a target
prompt, the CR 603.3d drop when nothing is legal, a "you may" if it
prints one, and a place on `PendingTriggers` — exactly as a harvested
trigger does, because by the time it is on the stack it is one.

Two rules, and both are why cards used to get this wrong by folding
the follow-up into the parent's effect:

- **It uses the stack, above the parent.** The table gets a response
  window between the two halves. Folding is the mistake ADR 0018
  retired for ordinary triggers.
- **Its targets are chosen when it goes on the stack**, not when the
  parent was announced — after the reveal, the sacrifice, the mill.
  A clause hung on the parent instead makes the controller pick
  before making the choice the trigger is about, which is what every
  folded card declared as a caveat.

"When you do" is conditional on the doing, and the `if` is the card's:
apply the trigger only on the branch where the thing happened. The
"you may" of "you MAY sacrifice a creature. When you do, …" belongs to
the *parent* — set `Optional` only when the reflexive sentence itself
says it.

**"Tap any number of … . When you tap one or more this way, …"** is
that shape with the choice in front of it, and it needs no new prompt
kind (#626, Teferi Akosa of Zhalfir's −3). The first sentence is an
ordinary `ChooseCardsPrompt` with `Min: 0` over the candidates and
`Zone: ZoneBattlefield`; its `Then` re-reads every pick before tapping
it (the board moves under an asynchronous prompt, and a creature that
is no longer an untapped one you control was not "tapped this way"),
counts what actually became tapped, and applies the `ReflexiveTrigger`
only when that count is at least one. Build the trigger's `Targets`
with the count in hand — "with mana value X or less" is a clause you
could not write on the parent, because X is not known until the taps
are in — and put the tapped creatures on `Cards` so the `Effect` can
read them back with `ctx.PayloadCards()` instead of closing over them.

**Mana from a spell (roadmap batch 01):** "Add {B}{B}{B}" on a SPELL
(Dark Ritual) or a non-mana ability (Mana Drain's refund) is the
`AddMana` primitive in `add_mana.go`, not a `ManaAbility` — a mana
ability never uses the stack (CR 605.3b) and these do, which is why
they can be countered and why Storm-Kiln Artist triggers on them:

```go
AddMana{Produced: "{B}{B}{B}"}.Apply(ctx)   // Player defaults to the controller
```

Pipe syntax queues the same colour pick a Birds activation does. The
mana empties with the pool at the end of the step (CR 106.4).

**Flicker (S22):** two shapes, and the difference is observable:

```go
Flicker{Target: id}                                  // "exile it, then return it" — one go
ExileTarget{Target: id} + ScheduleDelayedTrigger{…}  // "exile it. At the next end step, return it"
ReturnFromExile{Target: id, Tapped: true}            // "…return it tapped"
```

Either way the permanent returns as a **new object** — fresh
`InstanceID`, no counters, no damage, summoning-sick again — and
re-triggers every ETB it has. Leave `Controller` zero for "under its
owner's control"; set it only for "under your control". See
[flicker.go](server/internal/cards/effects/flicker.go).

**A plain token (#581):** `TokenCard("1/1 white Soldier")` — the
templates are rows in
[tokens_table.go](server/internal/cards/effects/tokens_table.go), keyed
the way the card prints them ("2/2 black Zombie", "1/1 blue Bird with
flying", "3/3 green Beast", "0/4 colorless Wall artifact with defender").
A token the table lacks is a new row, not a new constructor; a variant
(enters tapped, with counters) wraps the template in a `TokenSpec`.
`TestEveryTokenKeyResolves` fails on a key that is not in the table.

**A token that PRINTS AN ABILITY (#521, #1248, [ADR 0083](docs/decisions/0083-token-abilities.md)):**
not a row — a `tokenTemplate` in the catalog, because behaviour goes in
the catalog and data goes in the table. Declare the slug, the printed
characteristics, whichever of the four ability slots it uses, and the
printed text; add the builder to `tokenTemplates` in
[token_catalog.go](server/internal/cards/effects/token_catalog.go); give
it a named constructor beside the card that makes it.

```go
func PestToken() game.Card { return tokenFromCatalog(printedPestToken) }

func printedPestToken() tokenTemplate {
    return tokenTemplate{
        Slug: "pest",
        Card: game.Card{Name: "Pest", TypeLine: "Token Creature — Pest",
            Power: 1, Toughness: 1, Colors: []string{"B", "G"}},
        Triggered: []game.TriggeredAbility{
            WhenThisDies("Pest — you gain 1 life", func(g *game.Game, item *game.StackItem) error {
                return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
            }),
        },
        Text: "When this token dies, you gain 1 life.",
    }
}
```

The abilities are the SAME constructors a printed card uses — nothing
about a token is a dialect — and they are registered once at boot under
`game.TokenKey(slug)`, where the ordinary accessors find them through
`game.CatalogKey`'s token-key fallback. Five rules, each enforced at
boot by `checkTokenTemplate`: the template needs a slug, at least one
ability, and **printed text** (a token has no printing to fetch oracle
text from, so `Text` is the only thing that can tell the player what it
does — it ships as `CardView.token_text`); it must not declare an oracle
ID, and it must not leave a mana or activated ability on the `Card`
(that is a closure on the instance, and one of those on the battlefield
stops every restore point being written — #521, ADR 0041).

Two cards that print the SAME token share one template and one slug —
Beledros Witherbloom and Sedgemoor Witch both make `pest`. The slug is
an on-disk identity a snapshot carries, so rename one with the care a
database column gets, and qualify an ambiguous name
(`dragon-firebending`, `dragon-nesting`) rather than numbering it.
`tokenFromCatalog` takes the BUILDER, not the slug: a template may name
another token (the Goblin Shaman makes a Treasure) and a slug-keyed
registry would be a package variable whose initialiser reaches back into
itself.

**A token that's a copy (S22):** `CreateTokenCopy`, not a hand-written
template:

```go
CreateTokenCopy{Controller: item.Controller, Copy: cardID, N: 1,
    Except: func(t *game.Card) { /* "except it's a 4/4 black Zombie" */ }}.Apply(ctx)
```

`Copy` may be in **any** zone (graveyard for Hashaton, exile for
eternalize, battlefield for a Clone-style copy). The copied card's
oracle ID rides onto the token, so its triggered / static / mana /
activated abilities all come along for free — every one of those hooks
does a catalog lookup rather than reading a field. Two gaps worth
knowing: the copied card's `Spec.AsEnters` does **not** fire (its
`Triggered` `EventETB` abilities do), and per-instance state (counters,
`ExilePlay`, the cached characteristic) is deliberately not copied — CR
707.2. "Enters as a copy" for a real card (Clone) is a different thing
and still unimplemented: that is CR 613 layer 1, deferred to S16.5.

**Event picker.** Each row is the `when` for `On(kind, when, label, effect)`; the rows with a name in `triggers_common.go` are the constructors above.

| Trigger text | `Watches` | `AppliesTo` |
|---|---|---|
| "When ~ enters the battlefield" | `EventETB` | `ev.CardID == source.InstanceID` |
| "When ~ dies" | `EventLTB` | `cardDied(ev, source)` (graveyard-only; bounce / exile don't count) |
| "Whenever you sacrifice a permanent" | `EventSacrifice` | `ev.Actor == source.Controller` — fires while the permanent is still on the battlefield, before its `EventLTB` |
| "Whenever a player sacrifices a permanent" | `EventSacrifice` | `ev.CardID != uuid.Nil` (Mayhem Devil) — any player, any permanent type |
| "Whenever another creature dies" | `EventLTB` | `diedCreature(ev, g)` — resolves the dying card post-move; add `dead.Controller == source.Controller` for "you control", `!IsToken(dead)` for "nontoken" |
| "Whenever ~ attacks" | `EventAttack` | `attackDeclared(ev, source)` — `EventAttack` carries the attacking creature in `CardID`, exactly as `EventETB` carries the entering permanent |
| "Whenever a creature you control attacks" | `EventAttack` | `attackDeclaredByYou(ev, source.Controller)` — reads `ev.Actor` (the attacker's controller); fires **once per attacking creature**, so a three-creature alpha strike triggers three times. Add `ev.CardID != source.InstanceID` for "another". `ev.Target` is the defending player |
| "Whenever ~ enters or attacks" | `EventETB` + `EventAttack` on **one** ability | `ev.CardID == source.InstanceID` — one printed ability with two trigger conditions is one `TriggeredAbility` watching two kinds, not two declarations (Sun Titan) |
| "Whenever a spell or ability you control exiles one or more permanents" | `EventZoneMove` | `ev.OldZone == ZoneBattlefield && game.ExiledBySpellOrAbilityOf(ev, source.Controller)` plus `OncePerBatch` (Ranar the Ever-Watchful, #1320). Reads `ev.Cause` / `ev.CauseController`, which a routed move fills from the resolving item; a cost, a special action or a manual move names its own cause and never matches |
| "Whenever ~ becomes the target of a spell or ability" | `EventBecomesTarget` | `ev.CardID == source.InstanceID` — `CardID` repeats `Target` when the target is a card and is `uuid.Nil` for a player, so reading `CardID` is what keeps a player-targeting spell from matching. `ev.Actor` is the targeting player, `ev.Source` its source |
| "Whenever another creature you control becomes the target…" | `EventBecomesTarget` | `targetedAnotherCreatureYouControl(ev, source, g)` (Monk Gyatso) — excludes the source, checks the target is still on the battlefield, then reads its type and controller. Fires once per target **slot** (CR 115.3), at **announce** (CR 601.2c), so the trigger goes on the stack ABOVE the spell that targeted and resolves first — which is the whole card |
| "At the beginning of your upkeep" | `EventBeginUpkeep` | `ev.Actor == source.Controller` |
| "At the beginning of your end step" | `EventBeginEndStep` | `ev.Actor == source.Controller` — drop the check for "the beginning of the end step" (any player's) |
| "At the beginning of combat on your turn" / "your postcombat main phase" / "end of combat" (any step without a kind of its own) | `EventStepBegan` | `StepBegan(game.StepBeginCombat, true)` — or the constructors `AtBeginningOfYourCombat`, `AtYourPostcombatMain`, `AtEndOfYourCombat`, `AtYourStep(step, …)`, `AtEachStep(step, …)` (#588) |
| "Whenever you cast a creature spell" | `EventCast` | `ev.Actor == source.Controller` + `g.LookupCardForEffect(ev.CardID)` for the spell's type |
| "Whenever an opponent casts their first noncreature spell each turn" | `EventCast` | `g.CastTallyFor(ev.Actor).Noncreature == 1` (tally is bumped before the event fires) |
| "Whenever an opponent draws a card" | `EventDrawCard` | `ev.Actor != uuid.Nil && ev.Actor != source.Controller` — fires once per card |
| "Whenever ~ deals combat damage to a player" | `EventDealDamage` | `ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)` |
| "Whenever a creature you control deals combat damage to a player" | `EventDealDamage` | `combatDamageToPlayerBy(ev, source.Controller, g)` — checks `ev.Combat`, player target, creature source |
| "Whenever **one or more** creatures you control attack / enter / leave" (no object named) | the same kind as the per-creature wording | wrap the ability in `OncePerBatch(...)` (#587) — the engine emits one event per creature and declines the rest of the **batch** (see below). Without it the card ships **stronger** than printed |
| "Whenever **one or more** creatures you control deal combat damage to **a player**" / "whenever you attack **a player**" | the same kind as the per-creature wording | `OncePerBatchPerPlayer(...)`, or the ready-made `WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(creature, label, effect)` — once per **player**, not once per step (#784, CR 603.2c). The clause names an object, so the guard keys on it: three creatures hitting three opponents are three triggers, two hitting one opponent are one. A card whose stack label is computed per player (Breena, Nature's Will) sets a static `Key` and lets the guard supply the player |
| "Whenever a creature / land you control enters" (landfall) | `EventETB` | `enteredUnderYourControl(ev, source, g, false)` then `c.IsCreature()` / `c.IsLand()` (Impact Tremors, Tireless Provisioner) |
| "Whenever you create or sacrifice a token" | `EventTokenCreated` + `EventSacrifice` on one ability | `ev.Actor == source.Controller`, and for the sacrifice half `IsToken(LookupCardForEffect(ev.CardID))` — the sacrifice event fires **before** the zone move, so the token is still findable (Mirkwood Bats) |
| "When you lose control of ~" (Khârn the Betrayer) | `EventControlChanged` | `ThisChangedController` — `WhenYouLoseControlOfThis`. The event names the permanent in `CardID`, the player who LOST control in `Target` and the one who GAINED it in `Actor`; the item goes on the stack for `ev.Target`, because by the time the event lands the permanent belongs to somebody else. Emitted from the one materialise step at the end of the layer pass (#930), so a theft, an exchange, an Aura being destroyed and a duration expiring all reach it |
| "When you gain control of ~ from another player" (Risky Move) | `EventControlChanged` | `ThisChangedController` — `WhenYouGainControlOfThis`; the gaining player already controls the permanent, so the ordinary item is theirs |
| "Whenever an opponent gains control of a permanent you own" | `EventControlChanged` | `AnOpponentGainedControlOfAPermanentYouOwn` — `WheneverAnOpponentGainsControlOfAPermanentYouOwn`. The watcher is one permanent and the permanent that moved is another, linked by OWNERSHIP (CR 108.3), which no theft changes |
| "When this card becomes plotted" (Longhorn Sharpshooter, Aloe Alchemist) | `EventBecomesPlotted` | `WhenThisBecomesPlotted(label, effect)` — `Self`, watched from **exile** (`InExile`, #925). One emitter, `Game.PlotExiledCardForEffect`, so the plot special action and an "it becomes plotted" effect (Aven Interrupter) both fire it and a plain exile never does. `ev.Actor` is the plotter (the owner for the special action, the resolving item's controller for an effect), `ev.Source` what did it; the trigger is the card's OWNER's either way (CR 108.4). #1382 |
| "…its controller may draw" (Edric) | `EventDealDamage` | `ev.Actor` is the dealing creature's controller; use it for both `OptionalPrompt.Chooser` and the draw |

**What a batch is** (#829, CR 603.2c) — **a batch is every event the
engine emits between two points where play moves on: a stack item
beginning to resolve, and the turn cursor entering a new step.**
Nothing else opens one. So one resolution is one batch (a Cyclonic
Rift bouncing four creatures draws Dour Port-Mage one card), one
turn-based action is one batch however many engine calls the sandbox
splits it across (three `DeclareAttacker` clicks are one declaration
and one Adeline trigger), and the NEXT resolution is a new batch
however much of the last one is still on the stack (two Unsummons in
one turn draw two cards). The rule has no exception clause: the
first-strike and regular combat damage steps are two batches
(CR 510.4, #784), and since #717 they are two batches because they
are two real steps the cursor enters — so a first-striker and a
regular attacker connecting with the same player are two triggers.
The batch id is stamped on `Event.Batch` and the guard is
`oncePerBatchAllowsLocked`
([event_batch.go](server/internal/game/event_batch.go)) — one
counter, one guard, no per-card special cases. Known gap: two
SANDBOX-MANUAL mutations in a row with nothing resolving in between
share a batch.

**What the key counts** (#784, CR 603.2c's other half) — a batch is
WHEN; the ability's key is WHAT. A clause that NAMES AN OBJECT
triggers once for each of them in the batch: "deal combat damage to
**a player**", "attack **a player**". `TriggeredAbility.BatchKey`
reads that object off the event and the guard appends it to the
key, so the check is "(source, key, player) once per batch" —
`effects.OncePerBatchPerPlayer` is the one reading the catalog
uses. One key with two dimensions, one guard, no second dedupe path;
the last hand-rolled one (`TriggerInFlightForEffect`) is gone with
it.

**The two rules that matter:**

1. **`Build` builds; it does not resolve.** Do the work inside the
   `Effect` closure passed to `NewTriggeredItem`, never in `Build`
   itself. Applying the effect in `Build` skips the stack and denies
   every player their response window. The only game reads `Build`
   should do are the ones that pick targets.
2. **The `Effect` closure reads everything off `item` and the `g` it
   receives.** Don't capture `source *game.Card` (a pointer into a
   zone slice) or the `*game.Game` from `Build`'s arguments — undo
   restores a cloned game and the closure has to resolve against
   that one. `item.Controller`, `item.SourceCardID`, `item.Targets`
   carry what you need.

**Targeted triggers** declare the clause on the ability, exactly
like a spell's `Spec.Targets`:
```go
Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
```
The engine does the rest (S20): it computes the legal set when the
trigger fires and removes the trigger if the set is empty (CR
603.3d — no prompt at all), asks "you may" if there is one, then
queues a `pick_target` prompt the controller answers by clicking
the board. The chosen ref arrives in `item.Targets[0]` — your
`Effect` reads it from there (see `destroyChosenTargetTrigger` in
[acidic_slime.go](server/internal/cards/effects/acidic_slime.go)) —
and resolution re-checks it (CR 608.2b). Never pick a target
inside `Build`.

**"You may" triggers** set `OptionalPrompt: &game.TriggerOptionalPrompt{Question: "..."}`.
The harvester queues a yes/no `PendingChoice` instead of calling
`Build`; on "Yes", `Build` runs (via the target pick first, if the
trigger is targeted) and the item goes onto the stack.

**"Unless that player pays {N}"** (Rhystic Study, Smothering Tithe,
Esper Sentinel) is a `PayUnless` primitive the trigger's `Effect`
applies: it queues a `pay_unless` prompt for the taxed player and
returns; the "unless" consequence runs later as `OnDecline` when
they answer "Don't pay" — or "Pay" without the mana in pool +
untapped sources. Capture the payer's ID in `Build` (it's
`ev.Actor` for cast / draw events) and read the controller off the
`Context` inside `OnDecline`. See
[rhystic_study.go](server/internal/cards/effects/rhystic_study.go).

**Dies triggers** get the CR 603.10 last-known-information
characteristics as the third `Build` argument — the card is already
in the graveyard when `Build` runs, so read power / toughness /
types from `sourceLKI`, not `source`.

**"Where X is that creature's power"** (and any other read of the
event's permanent at RESOLUTION) is `ctx.TriggeringPermanent()`
(#1379, CR 608.2h): live while that object is still on the
battlefield, its last-known information — counters included — once
it has left, and never the new object a returned card became. Don't
capture a power in `Build` or look the card up by ID at resolution.
Check `info.Left` before acting ON the permanent: last-known
information is read, never written to. See
[ADR 0018's 2026-09-24 amendment](docs/decisions/0018-triggers-on-the-stack.md).
When that permanent is also the DAMAGE SOURCE ("it deals damage equal
to its power"), name it by object — `ref := ctx.Trigger().Object.Ref()`
then `DealDamage{SourceObject: &ref, …}` — so a departed source keeps
its lifelink and deathtouch and a returned card's new object is never
mistaken for it (#1396, [ADR 0056's 2026-09-24 amendment](docs/decisions/0056-infect-wither-toxic.md)).

**Tests** — `castCatalogSpell` + `passPriorityAroundTable` settles
the spell *and* the trigger it queues (the helper waits for
`Game.Stack`, `StackMeta`, and `PendingTriggers` to all empty).
Assert the trigger is on the stack with `triggerOnStack(g, cardID)`
before the second pass if the timing is the point of the test. For
optional triggers, `answerLatestTriggerPrompt` then
`passPriorityAroundTable` again. Upkeep triggers: `advanceToUpkeepOf`
then `passPriorityAroundTable`. See the S19 sections of
[cards_test.go](server/internal/cards/effects/cards_test.go).

### Adding a triggered MANA ability (#763)

Some triggers never reach the stack. **CR 605.1b:** a triggered
ability is a *mana ability* when it triggers off a mana ability
resolving, does not target, and could add mana — and **CR 605.4a**
then says a mana ability does not use the stack at all. It resolves
the instant the mana ability that triggered it has finished, with no
priority window for anybody.

That is the only way Wild Growth works: its extra `{G}` has to be in
the pool before the spell the land was tapped for is cast, and a stack
trigger would arrive long after.

So these live on **`Spec.ManaTriggers []game.ManaTrigger`**, never on
`Spec.Triggered`. The test is one line: **if the ability fires off a
permanent being tapped for mana, adds mana, and does not target, it
goes here.** An "add mana" trigger that fires on a CAST or an ATTACK
(Electro, Fire Nation Palace) is an ordinary stack trigger — CR 605.5a
— and stays in `Spec.Triggered`.

```go
ManaTriggers: []game.ManaTrigger{
    WheneverAttachedTapsForMana("Wild Growth — add an additional {G}", "{G}"),
},
```

Constructors live in
[mana_triggers.go](server/internal/cards/effects/mana_triggers.go):

| Printed clause | Constructor | Card |
| --- | --- | --- |
| "Whenever enchanted land is tapped for mana, its controller adds …" | `WheneverAttachedTapsForMana(label, produced)` | Wild Growth, Overgrowth, Fertile Ground |
| …with a computed output | `WheneverAttachedTapsForManaFunc(label, fn)` | Utopia Sprawl |
| "Whenever a player taps a land for mana, that player adds …" | `WheneverAPlayerTapsALandForMana(label, fn)` | Mana Flare, Heartbeat of Spring |
| "Whenever you tap a land for mana, add …" | `WheneverYouTapALandForMana(label, fn)` | Mirari's Wake, Zendikar Resurgent |

The output callbacks: `AddsFixedMana("{G}{G}")`,
`AddsOneManaOfAnyTypeProduced()` (reads `prod.Colors` — the produced
COLOUR, which is the other half of this seam) and
`AddsOneManaOfTheChosenColor()` (#742's stored `Card.ChosenColor`).
`Produced` returns the ordinary `ParseProducedMana` grammar, pipes
included, and returning `""` adds nothing — which is what an unchosen
colour must mean, never "any colour".

Six rules the engine applies for you, none of which a card declares:

- **It never touches the stack** and never queues a `PendingTrigger`.
- **It fires once per production**, from whichever of the three
  production sites knew the colour: the hand-clicked activation, the
  answered `mana_pick` (a dual land, a Birds) or the auto-tap executor.
- **Only a `{T}` fires it** (CR 106.12a). A sacrifice-cost mana ability
  and `AddMana` from a resolving spell (Dark Ritual) do not.
- **Triggered mana does not re-trigger.** A second Wild Growth does not
  see the first one's `{G}`.
- **A colour choice inside the trigger** prompts by hand and picks
  greedily against the cast under the auto-tapper, so an auto-tapped
  cast never stops on a prompt.
- **`ActiveWhen` and ability removal** work exactly as on
  `Spec.Triggered` — an Aura under Song of the Dryads has no trigger.

**Declared, and say so in the card comment:** the auto-tap PLANNER does
not count the extra mana ([ADR 0074](docs/decisions/0074-triggered-mana-abilities.md)
§7). It may tap one land more than it needed and the surplus floats
until the step ends — weaker than printed and safe. The mana that
arrives is always right.

Out of scope and still open: mana-production REPLACEMENT (CR 106.12b —
Nyxbloom Ancient, Mana Reflection) and turn-scoped mana triggers (High
Tide, Bubbling Muck, left to #663).

**Tests**: `pushAuraOnLand` + `tapForMana` in
[mana_trigger_cards_test.go](server/internal/cards/effects/mana_trigger_cards_test.go);
the engine rules themselves are in
[mana_trigger_test.go](server/internal/game/mana_trigger_test.go).

### Choices made at resolution (#796, #568)

Three shapes, all addressed by `Player` / `Chooser`, so "you may" and
"an opponent may" are the same call with a different seat.

**`MayChoice{Player, Question, YesLabel, NoLabel, LifeCost, OnYes,
OnNo}`** ([may_choice.go](server/internal/cards/effects/may_choice.go))
is the free yes/no a RESOLVING effect asks — "you may [do X]. If you
do, [Y]" where X is neither a search nor a cost the engine already
prompts for (Eden's sacrifice after the mill, Combustible Gearhulk's
question to its target). It is the existing `confirm` prompt underneath,
so it needs no new kind; `Player` defaults to the controller. Anything
printed after the decision goes in `OnYes`, not after `Apply` returns
— `Apply` only queues the prompt, exactly as `Scry.Then` exists.

**`PickOption{Player, From, Question, Options, Then}`**
([resolution_choice.go](server/internal/cards/effects/resolution_choice.go))
is "choose one of the following" over three or more branches, on the
new `option_pick` kind; `Then` gets the chosen INDEX. Build the option
list out of what the chooser can actually DO (CR 608.2) and put a
branch that always works FIRST — the enumerator marks that one
always-legal, and a prompt whose every branch can fail is a seat that
can be stuck (#544).

**`PileSplit{Splitter, Chooser, Owner, Cards, Then}`** is "an opponent
separates those cards into two piles; you take one" — two chained
prompts to two different seats, and no kind of its own. **Reveal the
cards first** (`RevealTopOfLibrary`): the splitter is being asked about
a zone that is not theirs, and protocol's `redactChoiceCards` shows
them only what is public, so an unrevealed pool reaches them as an
empty prompt.

**`ChoosePlayer{Chooser, Among, Except, Question, Then}`**
([choose_player.go](server/internal/cards/effects/choose_player.go))
is "choose a player" / "choose an opponent", also on `option_pick` —
one option per eligible seat, labelled with that seat's name.
`Among` is `Players`, `Opponents` or `OpponentsOf(id)`; `Except` is
how "choose a SECOND player" is spelled, fed from
`ctx.ChosenPlayers()`. `Then` reads the answer with
`ctx.ChosenPlayer()`, and it runs even when no question could be
asked — an empty pool or a chooser who has left — in which case
`ctx.ChosenPlayer()` is `uuid.Nil`, so **every branch checks before it
acts**.

**`ChoosePlayerAsEnters(label, pool)`** (same file) is the CR 614.12
form — "As this enters, choose a player" (True-Name Nemesis, Sawhorn
Nemesis), #980. It goes in `Spec.AsEnters`, and the difference from
`ChoosePlayer` is where the answer lives: not on one stack item for one
resolution, but on the permanent as `game.Card.ChosenPlayer`, read back
with `ChosenPlayerOf(g, sourceID)` for the rest of that permanent's
life. Same prompt underneath — the same `option_pick`, so the choice
gate, the enumerator, the wire and the CR 800.4a seat pruning all apply
unchanged.

Three rules for reading a stored player, and they are `ChosenColor`'s
verbatim: **read it LIVE on every check** (the answer arrives after the
permanent does, and a bounced permanent chooses again); **treat
`uuid.Nil` as nobody, never as everybody** (the window while the prompt
is open must apply to nothing, not to the whole table); and **do not
carry it into a copy** — CR 707.2, which you get for free because
`CopiableValuesOf` never looks at the field.

A chosen player is not a target in either form: it is named without
the stack, nothing may respond to it, and nothing re-checks it against
the board.

Branches take a `*Context` and are package-level functions capturing
scalars — never a `*game.Game` or a pointer into a zone, for
`StackItem.Effect`'s reason: an undo restores a clone and the branch
has to resolve against that one.

**A rule about the picked SET, not about each card** — "discard two
cards unless you discard a creature card", "any number of nonland
permanent cards with total mana value 4 or less from among them" — is
`Validate func(picked []game.Card) bool`, and it is enforced in exactly
one place: `checkChooseCardsPicksLocked`, which both `ResolveChooseCards`
and `internal/legal`'s `ChooseCardsPickLegalLocked` go through, so an
answer the bot is offered is an answer the resolver accepts. A refused
set comes back as `ErrChoiceSetRejected` with the prompt **still open**,
and it is never called for an empty pick, which keeps "choose nothing"
the answer a zero-floor prompt can always take. Every catalog primitive
that raises a card-set pick forwards the field verbatim —
`DiscardPrompt.Validate` (#624), `SearchLibrary.Validate` (#682),
`PutFromLibraryOntoBattlefield.Validate` and
`TakeFromLibraryToHand.Validate` (#998) — so **if you add another, add
the passthrough with it**. One behaviour rides along: a set rule turns
OFF the "the only legal answer is every candidate" shortcut, because
with a rule it is the rule and not the count that decides which subsets
are answers.

### Adding a `PendingChoiceKind` (#730, #794)

A new prompt kind owes two answers, and neither has a compiler behind
it. **One:** does an unanswered prompt of this kind stop the table?
Say so with a row in `choiceGateDecisions`
(`server/internal/game/choice_gate.go`), whose exported reader
`game.ChoiceBlocksTable(kind)` is the *only* predicate in the tree for
that question — the gated verbs ask it and so does `internal/legal`,
which is what keeps the bots and the engine from disagreeing the way
they did for the whole life of the allowlist (#794). Deny by default:
an unclassified kind blocks, and `pay_unless` is still the one kind
that does not (ADR 0018 §6). That row answers for the KIND; a live
PROMPT is asked through `(*game.Game).ChoicePromptBlocksTable`, which
adds two one-way narrowings on top of it, both DERIVED from the board
rather than declared by a card — `PendingChoice.GuardsStackItem`, a
prompt whose decline counters an object still on the stack (#951,
`counter_unless_paid.go`), and `PendingChoice.OwedInStep`, a prompt the
payer owes before the step it was asked in can end (#997,
`upkeep_pay_unless.go`). Both can only make a prompt block, never let
one through, and both lift by themselves when the thing they are about
has gone — which is what stops either being a wedge. There is no
card-level "please block" switch: #567 shipped one and the next two
cards with the same printed sentence both missed it (#997 removed it). **Two:** a case in
`choiceMoves`
(`server/internal/legal/choices.go`), or every seat owing one is
offered no answer *and* no pass — the #499 / #618 wedge that stopped
real tables on Door of Destinies and Cavern of Souls.

**Three (#902, #961):** a row in `choiceDepartureDecisions`
(`server/internal/game/leave_game.go`) — when the seat that owes this
prompt leaves the game, is the prompt **reassigned** to another player
(CR 800.4g: an object's choice that is not a cost) or **dropped** (its
own material, or a cost CR 800.4f says is simply not paid)? Deny by
default here, the opposite way round from the gate: an unclassified kind
is dropped, which is the pre-#902 behaviour and cannot wedge. The row's
second column is the **drop action**: when the prompt is dropped, does
the rule that ends it say what happens *instead*? `pay_unless` declares
`dropDecline` — CR 800.4f's cost is not paid, so the "unless" branch
runs, and Rhystic Study still draws — and `option_pick` declares
`dropDefault` (#1006), which runs the frame with `game.NoChoiceIndex`
so the rest of the card finishes even though the question ended
unanswered. Declare a new kind's action in the table, never in a
caller: the actions are performed in one place
(`runChoiceDropActionLocked`, `pending_choice.go`), reached both from
the departure sweep and from `dropChoiceLocked`, which is the door
every prune in the tree uses to withdraw a prompt. The table and its
reasoning are printed in the ADR 0060 amendments.

`TestEveryChoiceKindIsClassifiedAndEnumerated`
(`server/internal/legal/choice_gate_test.go`) reads the kind constants
out of `internal/game` and fails until the first two are done, naming
the kind and the file; `TestEveryChoiceKindHasAReassignmentDecision`
(`server/internal/game/leave_game_choices_test.go`) fails until the
third is. If either goes red on a kind you just added, that is the gate
working.

### Cumulative upkeep (#567, CR 702.24)

One constructor, `CumulativeUpkeep(label, cost)`
([cumulative_upkeep.go](server/internal/cards/effects/cumulative_upkeep.go)),
over primitives that already existed: an `AtYourUpkeep` trigger, the
counter primitive for the age counter (`game.CounterAge`), and
`PayUnless` for "sacrifice it unless you pay". The counter goes on
FIRST and the cost is then charged once **per counter** — built as
`strings.Repeat(cost, age)` at resolution, because a cumulative upkeep
of `{1}{U}` at three counters is three separate `{U}` symbols to pay
and not a number to multiply. `ParseCost` accumulates the repeated
string.

Two things that are not obvious:

- **The prompt blocks the table**, which no other `pay_unless` does.
  Use `effects.UpkeepPayUnless` (never `PayUnless`) for "at the
  beginning of your upkeep, pay or else": it goes through
  `Game.QueueUpkeepPayUnlessForEffect`, which anchors the prompt to the
  step it was raised in (`PendingChoice.OwedInStep`), and
  `game.ChoicePromptBlocksTable` holds the table there until it is
  answered — read by the engine gate and by `internal/legal` alike.
  ADR 0018 §6's latitude is Rhystic Study's: a question to a *different*
  player after the ability left the stack. This one asks the active
  player during their own upkeep, and the answer decides whether a
  permanent is still on the battlefield. The narrowing is one-way and
  per prompt — the `pay_unless` **kind** is unchanged, so Rhystic Study
  still plays as it did. Stasis and Pact of Negation are the same door
  (#997); nothing declares a halt on the card.
- **"Cumulative upkeep" is not a `canonicalKeywords` token**, for
  ward's reason (ward.go): the keyword carries a cost and a bare string
  in `Characteristic.Abilities` has nowhere to put one, so a token
  would tell the ADR 0037 coverage signal that every cumulative-upkeep
  card is implemented. The cost lives on the `Spec`.

Mana costs only. "Cumulative upkeep—Pay 2 life" (Glacial Chasm) and
"—Sacrifice a creature" (Phyrexian Soulgorger) are the same trigger
with a payment the pay-or-else prompt cannot parse; they wait for those
payment shapes rather than being approximated.

### The CR 726 loop breaker (#628)

Two permanents that trigger each other loop forever. The server never
blocks — one bounded unit of work per pass — but with autopass on for
every seat the table spins `pass → resolve → broadcast → pass` until
somebody finds the toggle. So the engine counts, and when the same
ability has resolved 25 times in one turn with **no player decision in
between** it raises `Game.LoopNotice` and one `EventLoopSuspected`.

What the notice does is suspend **automatic** passing — the client's
autopass `$effect` and the bot runner both hold — and nothing else.
Priority still rotates, `pass_priority` is still accepted, the trigger
is still on the stack. A human clicks "next" to step the loop on, or
casts something to end it. See
[ADR 0055](docs/decisions/0055-loop-breaker.md).

Three things to know if you touch priority, prompts or the tally:

- **Detection is one function**, `loopSuspectedLocked`
  (`server/internal/game/loop_breaker.go`), over `TurnTally.LoopRun` —
  `Resolved` restarted at each decision, keyed by
  `TallyKey(source, label)`: the CARD, deliberately, where the
  once-each-turn gates next to it are keyed per OBJECT (#936). A blink
  loop mints a new object every iteration and is still one loop. It
  counts ONE ability of ONE permanent in ONE turn, which is why four
  upkeep triggers from four players never approach it. The threshold is
  `DefaultLoopThreshold`; `Game.LoopThreshold` overrides it per game for
  tests. Do not add a second count.
- **A decision is anything but a pass**, and
  `notePlayerDecisionLocked` is the only thing that clears the run.
  Cast / attack / block notch through `turnTallyListener`; an
  activation notches in `ActivateCatalogAbility` (its announce emits
  `EventTrigger`, indistinguishable from a triggered one); an answered
  prompt notches in `dequeueChoiceLocked`. **A prompt the engine
  withdraws calls `dropChoiceLocked` instead** — the prune paths must
  not count as somebody deciding something, or a loop that queues and
  prunes a prompt each iteration never trips.
- **An activation does not clear its OWN run** (#810).
  `notePlayerActivationLocked` is `notePlayerDecisionLocked` with the
  activated ability's key kept, because an activation loop is a loop
  whose every iteration is a player decision — a free, repeatable
  ability re-offered the moment it resolves. Without the exception the
  run never got past 1 and the breaker never saw it. Every other key
  is still cleared: the decision was real.
- **A bare `pass_priority` is not a decision**, on purpose: if it were,
  the first manual "next" would clear the notice and four autopassing
  clients would spin the loop straight back up.

**The shortcut prompt (#804).** Raising the notice also asks the
repeating ability's controller "resolve it K more times, then stop?" —
`PendingChoiceLoopShortcut`, answered with a number. The answer is an
allowance on the tally (`TurnTally.LoopAllowance[key]`), spent one per
resolution; while it lasts the notice is down, so the client and the
bots pass normally with no code of their own, and the K-th resolution
raises the notice again and re-asks. `K = 0` is "stop here" and leaves
the table paused where the breaker put it. Two things to know if you
touch it: answering is a player decision, so the run is cleared and
`grantLoopShortcutLocked` puts *this key's* run back where the notice
found it — without that re-arm, "3 more" would mean 3 + the threshold;
and this prompt **blocks the table**, which the notice deliberately does
not, so `notePlayerDecisionLocked` withdraws a stale one rather than
leaving a wedge. CR 726.4's draw is still not built.

### Untapping in another player's untap step (#74)

"Untap all permanents you control during each other player's untap
step" (Seedborn Muse, Unwinding Clock, Drumbellower, Bender's
Waterskin, Quest for Renewal's second clause) is **not a trigger**,
and writing it as one is the mistake this field exists to prevent.
The untap step grants no priority (CR 502.4): nothing is announced,
nothing goes on the stack and there is nothing to respond to. The
clause widens the untap step's TURN-BASED ACTION — CR 502.3's "the
active player determines which permanents they control untap".

So it goes on `Spec.UntapStep []game.UntapStepPermission`, with the
constructors in
[untap_step.go](server/internal/cards/effects/untap_step.go):

```go
UntapStep: []game.UntapStepPermission{
    untapDuringEachOtherPlayersUntapStep(
        "Unwinding Clock — untap all artifacts you control",
        func(c game.Card) bool { return c.IsArtifact() }),
},
```

`AppliesTo` picks the STEP (every printed card in the family is
"each other player's", i.e. `activePlayer != source.Controller`, and
an intervening condition like Quest for Renewal's four quest counters
goes here too — it is continuous, so it is read at the instant the
step asks). `Untaps` picks the PERMANENTS, and is consulted only for
ones that are actually tapped.

The counterpart is `Spec.UntapStepRestrictions` (#751): self, attached
and filtered predicates keep permanents tapped during their controller's
own untap step. Read conditions after layers, especially power checks;
do not model them as layer-6 restriction bits. They do not stop a spell
from untapping a permanent or a Seedborn Muse permission on another
player's step. The constructors live in
[untap_restrictions.go](server/internal/cards/effects/untap_restrictions.go),
beside the permission helpers.

`Spec.UntapCaps` and `Spec.UntapOptOuts` (#826, [ADR 0070](docs/decisions/0070-untap-step-choices.md))
are the two clauses that make CR 502.3's *first* sentence a decision —
"players can't untap more than one land during their untap steps"
(Winter Orb, Static Orb, Winter Moon) and "you may choose not to untap
this during your untap step" (Rust Tick, Amber Prison). Constructors in
[untap_caps.go](server/internal/cards/effects/untap_caps.go). A cap is
a ceiling, not a restriction: it only asks when more permanents are
eligible than it allows, several caps compose (a chosen set has to
satisfy every one), and both families share ONE prompt, the
`untap_choice` kind. Caps and opt-outs are scoped by the engine to the
active player's own determination, because every printed card says
"during **their** untap steps" — so a Seedborn Muse untap on somebody
else's turn is uncapped, and the predicate never asks whose step it is.

**A turn-based action that can pause has ONE exit function.** The untap
step's is `exitUntapStepLocked`
([untap_choice.go](server/internal/game/untap_choice.go)), called from
the `StepUntap` case of the step-entry hook and from the prompt's
continuation; the cleanup step's is `exitCleanupStepLocked`
([cleanup.go](server/internal/game/cleanup.go)); the step ENTRY's is
`finishStepEntryLocked` (#710). Two sites that decide separately how a
step ends is how #661's discard path inherited a bug. `performUntapStepLocked`
returns whether it paused, and a paused step has untapped nothing and
moved no cursor — CR 502.3 is "determine, *then* untap them all
simultaneously", so the whole set untaps in one loop from the answer.

For one-shot effects use `DoesntUntapNextUntapStep` or `TapAndFreeze`.
`Player == uuid.Nil` follows the permanent's controller; a player ID
names that player's next untap step. Markers expire at that actual step,
even on an untapped permanent, survive skipped steps, and disappear on
zone changes. They are data on `Card`, not turn-scoped closures, so undo
and persisted snapshots retain them. Exert's action/cost remains separate
work. See [ADR 0058](docs/decisions/0058-doesnt-untap.md) and
[ADR 0070](docs/decisions/0070-untap-step-choices.md).

For "it doesn't untap during its controller's untap step **for as long
as** you control ~ / ~ remains tapped" (Ty Lee, Dungeon Geists, Rust
Tick), use `TapAndHoldWhileYouControlThis` or
`TapAndHoldWhileThisRemainsTapped` (#1313, ADR 0058's 2026-09-23
amendment). Each one records a HOLD: an `UntapSkip` whose `While` is a
CR 611.2 `Duration`. It is read at every untap step of the permanent's
controller, never used up by a step, and dropped once the duration ends.
The helpers tap the target whatever happens, and record the hold only if
the duration starts (CR 611.2b: a source already gone taps with no
lock). For any other duration, build it with a `Duration*` helper and
apply `DoesntUntapWhile`. Do not write this as an `UntapStepRestriction`
on the source that remembers its target. The hold lives on the target,
so a zone change, a copy and a snapshot already treat it correctly.

Two things to know when you touch the untap path at all:

- **Every untap goes through one primitive**
  (`Game.untapPermanentLocked`, `server/internal/game/untap.go`) and
  emits `EventUntapCard` when the permanent actually untaps — that is what
  makes Mesmeric Orb's
  "whenever a permanent becomes untapped" writable, and it must stay
  the only way `Tapped` goes false for a permanent on the
  battlefield. `MoveCard`'s battlefield-exit cleanup is not an untap
  (the card is no longer a permanent) and deliberately keeps its own
  write. A stun counter replaces any attempted untap of a tapped
  permanent with removal of one stun counter, including the sandbox
  buttons. A restricted or marked permanent never attempts to untap
  during the affected step, so it keeps its stun counters.
- **Untapping and summoning sickness are different questions.** The
  untap step clears `SummonedThisTurn` for the ACTIVE seat's
  permanents (CR 302.6 is about whose turn it is); a Seedborn Muse
  untap on somebody else's turn unquestionably untaps and just as
  unquestionably leaves your creatures sick. They were one loop
  before #74 only because the two sets were the same set.

### Phasing (#1199, CR 702.26)

A card that says **"phases out"** is two lines in the card file and
nothing else, because every consequence of the phrase is the engine's:

```go
// "Target creature phases out."
Effect: func(g *game.Game, item *game.StackItem) error {
    ctx := NewContext(g, item)
    return PhaseOut{Targets: legalTargetCards(item, g)}.Apply(ctx)
},

// "… phases out until this enchantment leaves the battlefield.
//  Tap that creature as it phases in this way."
return PhaseOutUntilLeaves{
    Targets:      legalTargetCards(item, g),
    Until:        source.InstanceID,
    TapOnPhaseIn: true,
}.Apply(ctx)
```

Both live in
[phasing.go](server/internal/cards/effects/phasing.go). **Pass the
whole target list in ONE call**, never one call per target: CR 702.26a
phases them out simultaneously, and an Equipment named beside the
creature it is attached to must not be dragged out twice (CR 702.26h).

**What you must NOT write, because the engine already does it**
(`game/phasing.go`, [ADR 0084](docs/decisions/0084-phasing.md)):

- the Auras, Equipment and Fortifications attached to the target go
  with it, transitively (CR 702.26g), and come back **still attached**;
- the permanent keeps its counters, its marked damage, its tapped
  state, its CR 613.7 timestamp and its `ObjectEpoch` (CR 702.26d) —
  phasing is **not a zone change**, so nothing goes through `MoveCard`
  and an exhaust ability stays spent across a phase cycle where a
  flicker would refresh it;
- it is removed from combat (CR 506.4);
- **no ETB, LTB or zone-change trigger fires**, at either end;
- it comes back during its controller's next untap step (CR 502.1),
  before that player untaps, with no delayed trigger to schedule.

**A phased-out permanent is not on the battlefield as far as your code
is concerned.** It is in `Game.PhasedOut`, out of
`g.Battlefield.Cards`, so `BattlefieldCardsForEffect`,
`legalTargetsLocked`, the layer pass, the SBA sweep, the trigger
harvester, `internal/legal` and the bot all skip it without asking —
CR 702.26b, true by construction rather than by 125 remembered
predicates. The only read surfaces that see one are
`g.PhasedOutCardsForEffect()` and `g.IsPhasedOutForEffect(id)`, and a
card should need neither. `FindCardZoneForEffect` deliberately answers
`nil`.

**The KEYWORD needs no `Spec` at all.** "Phasing" is in
`canonicalKeywords`, so the deck importer stamps it from Scryfall and
a printed-phasing permanent phases in and out on its own. A card that
GRANTS phasing (Shimmer's "each land of the chosen type has phasing",
an Aura's "enchanted permanent has phasing") is an ordinary layer-6
keyword grant — the engine reads the keyword off
`Effective().Abilities`, so a grant and a printing behave identically
and a permanent that has lost all abilities stops phasing.

**The wire** carries them in `GameView.phased_out`, a shared zone
beside `battlefield` / `stack` / `exile`, with `phased_out: true` on
each card; the client folds them back onto the controller's row,
dimmed and badged `PHASED`, with the click withheld.

Still missing, and named in the ADR rather than here: CR 702.26e /
702.26f's continuous-effect corners — a "gain control until end of
turn" whose object phases out keeps its `ScopedEffect` record
and applies again if the permanent returns inside the duration. No
catalogued card reaches it.

### Adding a preparation card (S46+, ADR 0090)

A preparation card (CR 722, Scryfall layout `prepare`) is two
registrations, the adventure shape — the permanent under the bare
oracle ID and its **prepare spell** under `"<oracle>#1"`:

```go
Register(Spec{OracleID: id, Name: "Skycoach Conductor",
    Replacements: []game.ReplacementEffect{SelfEntersPrepared()}})   // "enters prepared"
Register(Spec{OracleID: id + "#1", Name: "All Aboard", Targets: …, OnResolve: …})
```

The card never casts its prepare spell from hand (CR 722.3), and you do
not write the copy, the exile or the cast: the engine makes the CR 722.3c
copy in exile as the permanent becomes prepared, derives its controller's
permission to cast it, casts it as a copy that ceases to exist as it
leaves the stack, and unprepares the permanent as it is cast. The card
file says only WHEN the permanent becomes prepared —
`SelfEntersPrepared()` for "enters prepared", `BecomePrepared{Target}`
from a trigger or an ability — and what the prepare spell does. Read
`g.IsPreparedForEffect(id)` for "if this creature isn't prepared".

### Adding a hideaway card (S43+, ADR 0091)

Hideaway is two linked abilities (CR 607.2a), and a card file writes
both with shared words from
[hideaway.go](server/internal/cards/effects/hideaway.go):

```go
Triggered: []game.TriggeredAbility{Hideaway("Windbrisk Heights", 4)},  // CR 702.75a whole
Activated: []ActivatedAbility{{ …, Effect: func(g *game.Game, item *game.StackItem) error {
    if !condition(g, item.Controller) { return nil }                  // "… if <condition>" — asked at RESOLUTION
    return PlayHiddenCard{Source: HiddenRefOfActivation(item)}.Apply(NewContext(g, item))
}}},
```

`Hideaway` looks, asks for the one card, exiles it face down as
`FaceDownHidden` linked to the permanent OBJECT and bottoms the rest at
random; `PlayHiddenCard` grants the free play of the card THAT object
hid. The link is an object reference, so a triggered payoff captures
`game.ObjectRefOf(*source)` in its `Build` (Rabble Rousing) rather than
re-reading the source at resolution. The three Lorwyn lands are a table
in `hideaway_lands.go`; a new land of the same shape is a row.

### Adding a creature-type card (S26+)

Tribal cards come in three shapes, and the shared builders live in
[tribal.go](server/internal/cards/effects/tribal.go).

**A lord** ("Other Goblin creatures you control get +1/+1 and have
haste") is a `TribeFilter` plus one or two builders:

```go
goblins := TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}
Static: []game.StaticAbility{
    TribalAnthem(goblins, 1, 1),              // layer 7c
    TribalKeywordGrant(goblins, "haste"),     // layer 6
},
```

`Others` is the "other" in "other Goblins"; `YoursOnly` is the "you
control". **Read the printed card for the second one.** Half the
classic lords — Lord of Atlantis, Goblin King, Elvish Champion —
have no controller clause at all and buff the whole table, which is
a real and printed drawback. Inventing one is the most common way to
get a lord wrong, and the reason it is a named field rather than a
hand-written predicate.

**A named-tribe permanent** ("As this enters, choose a creature
type") carries the CR 614.12 prompt on `AsEnters` and reads the answer
back through `TribeFilter{Chosen: true}`:

```go
AsEnters: ChooseCreatureTypeAsEnters("Vanquisher's Banner"),
Static: []game.StaticAbility{TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 1, 1)},
```

The answer lands on `Card.NamedTribe` — per-instance state, carried
by the snapshot, cleared on battlefield-leave. Until the controller
answers, it is empty and the filter matches nothing; a static that
read an empty tribe as "everything" would be the dangerous
direction, so never write one. A mana ability whose restriction
names the chosen type uses `RestrictionsFunc`, not `Restrictions`
(see `ChosenTypeManaRestrictions`).

**Changeling** (CR 702.73a) is an enforced keyword since S26, so a
**vanilla changeling needs no catalog entry at all** — the deck
importer stamps it from Scryfall like any other printed keyword, and
`Card.HasSubtype` answers true for every creature type in every
zone. Only write a file when the card does something else too
(Irregular Cohort's token). A TOKEN declares it on the template's
`Keywords`: a keyword is plain data and always was, and it stays there
even for a token that has a catalog entry of its own (ADR 0083).

"Is every creature type" is a **layer-4 TYPE FACT**, not a keyword:
`Characteristic.AllCreatureTypes`, set by a grant (Maskwood Nexus, via
`AllCreatureTypesGrant`), by an until-end-of-turn grant
(`GrantAllCreatureTypesUntilEOT`), and by a printed changeling through
the one keyword→layer-4 projection in `printedCharacteristic`
(CR 702.73a is a characteristic-defining ability, so CR 613.2 applies
it before every other layer-4 effect). Read it with
`game.HasAllCreatureTypes`. The `changeling` keyword stays in
`Characteristic.Abilities` as the PRINTED source of the fact and as
the client's badge, and nothing else reads it.

It used to be the keyword alone, which put a layer-4 type where layer
6 could delete it — a creature that lost all its abilities stopped
being every creature type, against Maskwood Nexus' own 2021-02-05
ruling, and a later "is an Elk" left the keyword behind (#670, ADR
0067 §4). A grant declares **layer 4**, not the layer 6 an ability
grant would normally take: declared in layer 6 it gets
timestamp-ordered against every lord's keyword half, and a Goblin
Chieftain that entered first grants haste before the Bear became a
Goblin — while its +1/+1 lands correctly, because layer 7c runs after
all of layer 6. Half a working card.

Replace a subtype list with `Characteristic.SetSubtypes`, never by
assigning `c.Subtypes` — that helper is what clears the
every-creature-type fact, which is what "is an Elk" has to do
(CR 205.1b). An ADD (`c.Subtypes = append(c.Subtypes, "Swamp")`)
stays a bare append, because adding a type takes nothing away. And do
not append the ~345 entries of `game.AllCreatureTypes` to
`Characteristic.Subtypes`: it makes the wire type line unreadable
and every subtype loop quadratic, for a property one flag answers.

Ask "do these two creatures share a type" with
`game.SharesCreatureType`, never with a subtype-slice intersection —
the helper knows Forest on Dryad Arbor is a land type and that a
changeling shares nothing with a creature that has no creature type
at all. `OfCreatureType("Goblin")` is the targeting predicate.

**Tests** — `pushNamedTribePermanent` in
[tribal_test.go](server/internal/cards/effects/tribal_test.go) seeds
a permanent and answers its prompt in one call. Assert through
`effectivePower` / `effectiveAbilities` / `effectiveSubtypes` like
any other layer card.

### Adding a choose-a-color card (#742)

"Choose a color" (CR 105.4) is one prompt kind, `choose_color`, in two
forms, and the builders live in
[color_choice.go](server/internal/cards/effects/color_choice.go).

**Stored** ("As this enters, choose a color") copies the creature-type
pattern above: the prompt goes on `AsEnters`, the answer lands on
`Card.ChosenColor`, and the card's other abilities read it back.

```go
AsEnters: ChooseColorOtherThanAsEnters(game.ColorForMana, "Thriving Isle", "U"),
ManaAbilities: []ManaAbility{{
    Cost:         ManaAbilityCost{Tap: true},
    ProducedFunc: ProducedColorOrChosen("U"), // or ProducedChosenColor()
    Label:        "Add {U} or one mana of the chosen color",
}},
Static: []game.StaticAbility{ChosenColorAnthem(1, 0)}, // Heraldic Banner
```

**Every prompt declares a PURPOSE, and it is the first argument (#780).**
CR 105.4 makes all five colours a legal answer, so nothing about the
prompt says which one the card wants — an automated chooser with no
other information names its own main colour, which is right for
Coldsteel Heart and makes Wash Out a self-inflicted board wipe. Pick
the `game.ColorPurpose` that matches what happens to the colour named:

| Purpose | The chosen colour… | Cards |
|---|---|---|
| `game.ColorForMana` | is produced as mana | Coldsteel Heart, the Thriving lands, the Gates |
| `game.ColorForBenefit` | is helped, or survives | Heraldic Banner, Selective Obliteration |
| `game.ColorForHarm` | is punished, yours included | Wash Out |
| `game.ColorForFilter` | selects somebody else's cards | Oona, Queen of the Fae |
| `game.ColorForProtection` | is defended against | Mother of Runes, Story Circle |

It rides the prompt to the wire as `color_purpose` and the bot's policy
switches on it (`aiseat/heuristic/choices.go`, `colorChoiceValue`); the
engine itself never reads it. `color_purpose_guard_test.go` fails the
build for a prompt whose first argument is not one of those constants,
and for a card file that reaches `QueueColorChoiceForEffect` without
going through a builder.

The purpose has **three** readers, none of them the engine (#986): the
bot's policy scores by it (`colorChoiceValue`), the enumerator ORDERS
the answers by it, and the client's picker WORDS itself from it. The
ordering is one function — `legal.OrderColorOptionsLocked`
([color_order.go](server/internal/legal/color_order.go)) — called both
by `enumerator.colorAnswers` and by the `choose_color` projection in
`protocol.ViewOfGame`, so a bot's first offered answer and a human's
first button are the same colour. It reads battlefield counts and
nothing else (ADR 0033 §3), and it never narrows the list: an ordering
that dropped an option would be a rules change. Pick the purpose that
matches the card and all three follow.

Until the controller answers, the colour is empty, and every reader
must treat that as the weaker outcome: no mana, no anthem. Never read
an empty colour as "any colour". "A color other than blue" is just a
shorter option list, and colorless is never a colour.

**"Could produce" reads the choice (#782).** CR 106.7 is
`(*Game).ProducibleManaLocked` — the one function Exotic Orchard,
Reflecting Pool and Fellwar Stone ask — and it evaluates each mana
ability's `ProducedFunc` and runs the result through the same
`manaPickOptions` the activation does, so a chosen colour, a
commander-identity narrowing and a plain "any colour" all read exactly
as the tap would. Scryfall's `produced_mana` answers only for a card
with no catalog mana ability at all. A `ProducedFunc` that reads OTHER
permanents' producible mana must set
`ManaAbility.DerivesFromOtherSources` — that is the CR 106.7
recursion guard, `TestDerivedManaAbilitiesDeclareTheGuard` enforces it
both ways, and it is the only `ProducedFunc` shape "could produce"
skips.

**At resolution** ("Choose a color. …" inside a spell or ability) stores
nothing: `ChooseColorThen(purpose, g, chooser, source, question, then)` hands the
answer to a continuation that runs the rest of the effect (Wash Out,
Oona). The continuation receives the live `*Game`; rebuild the context
with `NewContext(g, item)` inside it. "Each player chooses a color" is a
chain: each answer's continuation asks the next player in APNAP order
(Selective Obliteration). Thread the answers through the chain as values
rather than mutating one shared map, so an undo cannot leak an answer
from an undone branch.

**"N mana of any one color"** is ONE pick minting N tokens, never N
pipe slots, which would let the player take N different colours. Write
it with the produced-mana grammar's per-colour count:
`OneColorOfAmount(3)` is `"{W3|U3|B3|R3|G3}"` (Gilded Lotus),
`ProducedOneColor(fn)` computes N at activation (Mona Lisa's power), and
a per-colour amount is `"{G4|U1}"` (Nyx Lotus's devotion). It works from
a spell or trigger too, through `AddManaForEffect`. That path offers the
printed colours with the commander's identity listed first, like a mana
ability; only printed "in your commander's color identity" text passes
`game.AddManaOptions{NarrowToCommanderIdentity: true}` to
`AddManaWithOptionsForEffect` (or sets `AddMana.NarrowToCommanderIdentity`),
the effect-side twin of the mana ability's flag. **The auto-tapper plans
such a source (#779)**: it offers the solver one candidate per colour,
they are mutually exclusive, and the plan carries the colour through to
the executor — so a Gilded Lotus funds `{3}{U}{U}` beside two Islands
and never funds `{W}{U}` alone, and the surplus floats
([ADR 0040](docs/decisions/0040-mana-pipeline.md) #779 addendum).

**Tests**: `pushChosenColorPermanent` and `answerColor` in
[color_choice_cards_test.go](server/internal/cards/effects/color_choice_cards_test.go).

### Shared vocabulary, and the clone gate

The catalog is one package and its helpers are one vocabulary. The
September 2026 review ([Discussion #557](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/557))
found the biggest cost in the tree was card-side copy-paste that grew
because batch authors were told never to touch shared files; that rule
is gone. In its place:

- **Shared code lives in mechanic-named files, and those files are
  append-only.** A trigger shape or condition goes in
  `triggers_common.go`; a card predicate or an effect body used by
  more than one card goes in `helpers.go` (or a `predicates_<mechanic>.go`
  / `effects_<mechanic>.go` beside it); a token is a row in
  `tokens_table.go`, or — if it prints an ability of its own — a
  `tokenTemplate` in the catalog (ADR 0083). Add a function; never change an existing one's
  behaviour in a card PR. Two PRs that both append to the same file
  merge cleanly.
- **No batch prefixes.** A helper is named for what it says
  (`instantOrSorceryCastByYou`), not for the batch that first needed
  it. The `bNN` names still in the tree are the promotion pass's
  backlog (#583), not a convention to follow.
- **Grep before you write.** `grep -n "func .*CastByYou" *.go` before
  writing a "whenever you cast" predicate; the third copy of a helper
  is how the catalog got to ~8,700 redundant lines.
- **The gate.** `TestNoNewExactClonesInTheCatalog`
  (`server/internal/cards/coverage`) fails a PR that introduces a new
  byte-identical function or closure body of six or more lines, and
  names both copies. **It also fails a PR that grows a group the
  baseline already knows about** — a third, fourth or tenth copy of a
  body already listed there is not free (#786): the gate compares the
  measured copy count against the baseline's `<copies>` column as a
  floor, not just a hash lookup, so an already-known hash with more
  members than recorded still fails. Fix either case by calling the
  one that exists, or by naming one shared helper and calling it
  twice. The baseline (`coverage/testdata/clone_baseline.txt`) records
  the duplicates that predate the gate; regenerate it with
  `go test ./internal/cards/coverage/ -update` when a PR removes some
  or shrinks a group's count, never add or edit a line by hand. **It
  is keyed on the body hash and the set of declaring files, never on a
  line number** (#895): each row is `<hash> <lines> <copies> <files>`,
  where `<lines>` is the body's LENGTH. So an edit above a listed
  closure does not rewrite the file, and a baseline diff in your PR
  means the set of duplicates — or a group's size — really changed.
  The `file.go:closure@line` locations are still printed in the
  failure report, where a human wants them.

### Winning, losing, and "can't lose" (S40, ADR 0057)

A card that says "you win the game" or "<player> loses the game" calls
the primitive and **returns its error** — nothing else:

```go
return WinTheGame{}.Apply(ctx)                    // "you win the game" (Felidar Sovereign)
return LoseTheGame{Player: target}.Apply(ctx)     // "that player loses the game" (Strixhaven Stadium)
```

Both are immediate (CR 104.2b, CR 104.3e), and both return
`game.ErrStopResolution` when the rest of the effect must not happen —
the game ended, or the resolving item's own controller left. The engine
treats that error as a clean stop, so a card never swallows it, and a
card **never** calls the rotation or the game-over check itself. A
draw-replacement win is `WinInsteadOfDrawingFromAnEmptyLibrary(name)`
(Laboratory Maniac), which cancels the draw before it tries to win.

"You can't lose the game" / "your opponents can't win the game" on a
permanent is a declaration, not code:

```go
GameEndGates: YouCantLoseOpponentsCantWin(),   // Platinum Angel, Herald of Eternal Dawn
GameEndGates: YouCantWinOpponentsCantLose(),   // Abyssal Persecutor
```

The engine reads it off the battlefield through `CatalogAbilityKey` at
every loss and every win, so two copies compose and an ability-stripped
copy gates nothing. The "this turn" version from a spell is
`CantLoseAndOpponentsCantWinThisTurn{Label: name}` (Angel's Grace),
stored on the caster as a `PlayerStatic` with an until-end-of-turn
duration. Concession is never gated, and the last player standing always
wins (CR 104.2a). See [ADR 0057](docs/decisions/0057-win-and-lose-by-effect.md)
and its 2026-09-24 amendment.

### When NOT to add a catalog entry

The registry of known seams — what is missing, which cards wait on
it, which are already tracked — is `server/internal/roadmap/registry.go`
([ADR 0092](docs/decisions/0092-public-roadmap-and-site-portal.md)),
rendered into the open table of [docs/engine-seams.md](docs/engine-seams.md).
Check it before triaging a skip as "needs machinery". In the batch PR,
append each skipped card's name to its seam's `Waiting` list in the
registry (adding a seam entry if there is none), then regenerate the
table with `go test ./internal/roadmap/ -update` — never edit the table
by hand; `TestSeamsTableIsCurrent` fails a PR whose table and registry
disagree (Discussion #559 item 6). The same registry feeds the public
roadmap page, so its `Summary` and `Missing` sentences are held to the
Caveats tone rule, and a seam you close in an engine PR is flipped to
implemented there.

**Closing a seam: add a fragment, never edit the Closed list.** The
Closed seams list in `docs/engine-seams.md` is generated from one file
per closure in `docs/engine-seams/closed/` (#1461, Discussion #1231
Option A), because a hand-written list that every engine PR prepended
to made any two open PRs conflict. Your engine PR adds
`docs/engine-seams/closed/<issue>-<slug>.md`:

```
---
title: "Phasing"
date: 2026-09-23
issues: [1199]
pr: 1251
---
**Phasing** (#1199, [ADR 0084](decisions/0084-phasing.md)) — what closed, and what is still open.
```

`title` is double-quoted and matches the bold heading the body starts
with (no `- ` list marker; the generator adds it); `issues` is a list;
`pr` is optional (add it once the PR exists, or leave it out). Links
are relative to `docs/`. Say which issue you mean rather than "the row
above": new fragments sort by date, not by where you would have put
them. Do **not** regenerate the list in a PR — CI's `census-publish`
job does it on every push to `develop` and `main`, as it does for the
census, and a regenerated block in your diff is the conflict this
exists to remove. `TestClosedSeamsAreCurrent` skips a stale list
outside CI but always fails a malformed fragment or an entry written
into the list by hand (the next refresh would delete it). To look at
the result locally, run
`go test ./internal/roadmap/ -run TestClosedSeamsAreCurrent -update-closed`
and then discard the change to `docs/engine-seams.md`.

- **Activated abilities whose cost has no component** — `AbilityCost`
  carries tap-this, sacrifice-this, sacrifice-another (since #747
  **N of them**: `SacrificeN(2, "two artifacts", Artifact())` for an
  ability, `SacrificeNCost(2, "two creatures", Creature())` for an
  additional cost to cast; a fixed count only, `Register` refuses
  "sacrifice X" and "one or more", and a restriction on the set
  ("with different names") has no shape either), mana, life,
  and since S27 **loyalty** (`LoyaltyCost(n)`, [ADR 0032](docs/decisions/0032-planeswalkers.md) §8)
  and **crew** (`CrewCost(n)`), and since #625 **counter removal**
  (`RemoveCountersFromThis(kind, n)` for "from this",
  `RemoveCountersFrom(kind, n, "a planeswalker you control", preds…)`
  for another permanent you control, kind `""` for "a counter" of any
  kind, `RemoveCountersXFromThis(kind, floor)` for "remove X / any
  number", and `RemoveCountersAmong(kind, n, label, preds…)` for "from
  among …" — with kind `""` there too since #943, the any-kind split
  (Tekuthal's "three counters from among other artifacts, creatures,
  and planeswalkers you control"), where the activator names a kind
  per permanent as well as a count and the payment sends
  `counter_kinds` beside `counter_source_ids` only when the kinds
  actually differ — [ADR 0020](docs/decisions/0020-activated-abilities.md) addendum)
  ([activated.go](server/internal/game/activated.go)) and nothing else.
  Equip needs no component of its own — `EquipAbility("{2}")` is a mana
  cost plus a target clause. **"Rather than pay" on an activated
  ability** is not an alternatives slot either: write it as a **second
  ability entry** with the same effect and its own real cost, which is
  what Heart of Kiran does ("Crew 3" and "Crew — remove a loyalty counter
  from a planeswalker you control"). The second entry must have a cost
  that can actually go unpaid, or it is the #259 mistake below. Still
  no shape: cycling and **convoke / waterbend on an ACTIVATED ability**
  — don't invent one. (Two things that used to be on this list are not
  any more: a counter removal **split across several permanents**
  shipped with #789 and #943's any-kind form, and a cost that **adds**
  a counter — `AddCounterToThis(kind, n)`, Devoted Druid — with #789.) (Convoke and waterbend on a *spell* do have one since
  S22: `Spec.TapCost`, built with `Convoke()` / `Waterbend("{X}")`. The
  activated-ability seam is separate and still open — Katara, Water
  Tribe's Hope is the card waiting on it.) (Ordinary activated abilities built from
  those components are fine since S21: see `Spec.Activated`
  above.) Shipping a card with a cost the engine
  can't express simply omitted makes it **stronger than printed**, which
  is the wrong direction for a simplification:
  [#259](https://github.com/krakenhavoc/cmd_and_ctrl/issues/259) was that
  mistake reaching the catalog (Waterbender's Restoration shipped as a
  two-mana mass blink) and is now **closed**, but it stood for a sprint
  and was caught by writing a decklist doc rather than by a test. If the
  cost has no shape, leave the CARD out.
- **Triggers on events the engine doesn't emit yet** ("whenever a creature enters under an opponent's control", landfall-with-a-target) — check
  [events.go](server/internal/game/events.go) for an `EventKind`
  first. If there isn't one, the event plumbing is the PR, not the
  card. Two things that used to be on this list are not any more:
  **attack declarations** (`EventAttack`) and **"becomes the target of a
  spell or ability"** (`EventBecomesTarget`), both S22 — see the event
  picker above.
- **Cost-replacement effects** (Trinisphere, Thalia, Spellshift, Kambal)
  touch the S15 cost engine rather than the S17 event pipeline. They
  land with S28.
- ~~**Cards that add a layer dependency, or that ability removal gets
  wrong**~~ — no longer a blocker, and the hold list is released
  (2026-09-18, [ADR 0067](docs/decisions/0067-layer-dependency-ordering.md),
  [#668](https://github.com/krakenhavoc/cmd_and_ctrl/issues/668) /
  [#669](https://github.com/krakenhavoc/cmd_and_ctrl/issues/669) /
  [#670](https://github.com/krakenhavoc/cmd_and_ctrl/issues/670)). The
  layer engine orders layer 4 by CR 613.8 dependency, an ability
  removal applies in its own layer and reaches forwards only
  (CR 613.6), and a "becomes a basic land type" effect removes only
  the land's own rules text in layer 4 (CR 305.7,
  `effects.SetsBasicLandType`). **Magus of the Moon** shipped with
  that change. **Arcane Adaptation** (#401), **Leyline of
  Transformation** (#396), **Encroaching Mycosynth** (#401),
  **Yavimaya, Cradle of Growth** (#294) and **Prismatic Omen** (#396)
  are unblocked: a layer-4 type-add whose "applies to" reads a card
  type or subtype is the resolved case now, not the broken one.

  Two things a card in that shape still has to get right. A layer-4
  effect that REPLACES the subtype list calls
  `Characteristic.SetSubtypes`; a static that spans layers and should
  survive its own source being silenced declares
  `ContinuesAfterRemoval` on the later-layer halves (ADR 0067 §2).
  Adding a dependency-ordered bucket other than layer 4 needs a
  catalogued pair that wants it and a benchmark — the ADR has the
  numbers.
- ~~**Cards that need a pick-from-zone UI**~~ — no longer a blocker.
  S20 shipped structured targeting and S18.5 the zone browser, so
  "target card in your graveyard" is a real target clause:
  `TargetCardInGraveyard(label, preds…)`
  ([targets.go](server/internal/cards/effects/targets.go)), answered by
  clicking the card in the zone browser. Eternal Witness and Sun Titan
  both use it. The S14 "auto-pick the top of the graveyard" fallback is
  only for cards that never declared a clause.

### Adding a trigger doubler (#752)

Declare `Spec.TriggerDoublers` using `DoublesEntering`, `DoublesDying`,
`DoublesAttacking`, or `DoublesAbilitiesOf` in
[trigger_doubling.go](server/internal/cards/effects/trigger_doubling.go).
Set the declaration's `Label` to the card's printed name for the stack and
prompt attribution. The cause helpers filter the event's subject; the
source helper filters the permanent whose ability triggered. Those are
different objects, and death predicates must use battlefield last-known
characteristics. See Panharmonicon, Teysa Karlov, Isshin and Cloud for examples.

The harvester creates independent instances: each gets its own optional
choice and targets. Do not copy an existing stack item or double inside a
card's `Build`/`Effect`. Delayed, reflexive and manual triggers are excluded.
Two matching doublers add two instances, giving three total. Tests should
exercise the actual card and check controller restrictions and a negative
cause, not just the helper predicate. The `OncePerBatch` first-event
limitation and remaining card wave are tracked in
[ADR 0018's addendum](docs/decisions/0018-triggers-on-the-stack.md#addendum-2026-09-17-trigger-doubling-cr-6032d--accepted).

### Emblems (#623)

"You get an emblem with [ability]" (CR 114) is one `Spec` slot plus a
one-line ability. Declare the emblem next to the ability that makes
it, and make the ability's whole effect `CreateEmblem{}` — it names
nothing, because the emblem it creates is this card's:

```go
Register(Spec{
    OracleID: "05e6b243-…",
    Name:     "Elspeth, Sun's Champion",
    Emblem: &EmblemSpec{
        Label:  "Elspeth, Sun's Champion emblem",   // "<card> emblem"
        Text:   "Creatures you control get +2/+2 and have flying.",
        Static: []game.StaticAbility{ /* … */ },   // and/or Triggered
    },
    Activated: []ActivatedAbility{{
        Label:  "−7: You get an emblem with \"…\"",
        Cost:   LoyaltyCost(-7),
        Effect: func(g *game.Game, item *game.StackItem) error {
            return CreateEmblem{}.Apply(NewContext(g, item))
        },
    }},
})
```

An emblem's abilities are written in **exactly** the vocabulary a
permanent's are: `game.StaticAbility` with a layer and a sub-layer
(the emblem object is the `source`, so "creatures you control" is the
same `target.Controller == source.Controller` an anthem uses), and the
ordinary trigger constructors (`Targeting(WheneverYouDraw(…), spec)`,
`WardGranted(…)` for an emblem that grants ward — Teferi Akosa of
Zhalfir's "Knights you control get +1/+0 and have ward {1}" is one
`TribeFilter` feeding an anthem static and a ward trigger).
There is no emblem dialect, because `effects.Register` files a second
`game.CardDef` under `game.EmblemKey(OracleID)` and the emblem object
reaches the layer pass and the harvester through the same
`CatalogStaticAbilities` / `CatalogTriggers` hooks a battlefield
permanent does. See [ADR 0064](docs/decisions/0064-emblems.md).

Three things to know:

- **`Register` panics** on an `EmblemSpec` with no `Label`, no `Text`,
  or no abilities at all. An emblem whose printed ability the engine
  cannot express yet is NOT declared with an empty `Static` — leave
  the ultimate omitted and say why in `Caveats`, as Wrenn and Six does
  for retrace (#652). ADR 0032 still holds: a −7 that costs seven
  loyalty and delivers a chip that does nothing is the lie the
  omission exists to avoid.
- **Nothing removes an emblem**, and nothing can name one. It is not a
  permanent, not a card, never a legal target, and there is no move,
  route or admin verb that reaches it — `Player.Emblems` is a second
  command-zone slice and `ZoneRef` is `{Kind, Owner}`, so
  `{command, owner}` always resolves to the commander pile. The one
  exit is CR 800.4a, its owner leaving the game.
- **The wire is `PlayerView.emblems[]`**, public and unredacted, with
  the label and text read from the catalog on every projection. The
  board draws chips beside the player identity; the command-zone pile
  stays commander-only.

### Designations: Class levels, solved Cases, station thresholds (#757, #759)

A **designation** is a marker a permanent has on the battlefield that
switches some of its own printed abilities on — a Class's level
(CR 716.2), a Case being solved (CR 719.3), a station card's charge
counters (CR 721.2), and later a Room's unlocked door (CR 709.5). Four
printed mechanics, one gate:
[ADR 0071](docs/decisions/0071-designations-that-switch-abilities-on.md).

**Write the gate, never an `if` inside the ability.** Every entry in
`Spec.Static`, `Spec.Triggered`, `Spec.Activated` and
`Spec.CostModifiers` may carry `ActiveWhen`; the constructors are
`Level(n)`, `Solved()` and `AtChargeCounters(n)` in
[designations.go](server/internal/cards/effects/designations.go), with
`AtLevel(n, trigger)` and `WhenSolved(trigger)` for the common
trigger case. An ability whose gate is unsatisfied **does not exist**:
it is not gathered by the layer pass, not matched by the trigger
harvester, not offered by the activation path or the legal-move
enumerator, and not on the wire. A predicate inside `AppliesTo` is a
weaker and different statement — it would still have prompted for a
target — so do not write one.

**Use the constructors for the lifecycle, too.** `LevelUp(n, cost)`
is the whole "{cost}: Level N" ability, carrying CR 716.2d's sorcery
timing and CR 716.2a's "only from level N-1"; `ToSolve(label, cond)`
is the whole "To solve —" clause, an end-step trigger whose condition
is re-checked on resolution (CR 603.4). `SpacecraftAt(n, p, t)` and
`ThresholdKeywords(n, kw…)` are the two station threshold shapes.

Three things follow from the designation living on `game.Card`
(`ClassLevel`, `Solved`) rather than in `Counters`: nothing
proliferates or doubles it, a copy does not take it (CR 716.2c,
719.3b), and it is cleared when the permanent leaves the battlefield
(CR 400.7). A new designation needs a kind, an arm in
`Designation.Active`, and a layer-version bump on the event that
changes it — nothing else.

### Abilities from the hand (#660)

An activated ability declares **where it functions** (CR 113.6) with
`ActivatedAbility.Zones []game.ZoneKind`. Nil — nearly every ability —
means the battlefield. Cycling declares `{game.ZoneHand}`, and the
graveyard activations behind it (Reassembling Skeleton, Drownyard
Temple) will declare `{game.ZoneGraveyard}`. There is **one**
activation path with a zone dimension, not a hand fork:
`ActivateCatalogAbility` finds its source in whatever zone holds it,
treats the card's OWNER as "you" off the battlefield (CR 108.4), and
refuses a mismatch with `ErrActivationZoneNotAllowed` before anything
is validated or paid. See
[ADR 0062](docs/decisions/0062-abilities-and-special-actions-from-the-hand.md).

Write cycling with the constructors, never by hand:

```go
Activated: []ActivatedAbility{Cycling("{3}")},                 // Cycling {3}
Activated: []ActivatedAbility{BasicLandcycling("{2}")},        // Basic landcycling {2}
Activated: []ActivatedAbility{Typecycling("Plainscycling", "{2}",
    "a Plains card", IsLandWithSubtype("plains"))},
```

They stamp the zone, the `DiscardSelf` cost and the `Cycling` bit that
makes `EventCycle` fire (CR 702.29b). A battlefield watcher of that
event is `WheneverYouCycle(label, effect)` — Astral Slide, Drake Haven.
**"When you cycle THIS card" does not work yet**: the card is in the
graveyard by then (CR 702.29c) and the harvester has no scan that finds
it there (ADR 0062 Decision 7). Declare that half as a caveat, as
Magmakin Artillerist does.

Two discard cost components sit on `AbilityCost`. `DiscardThis()` is
cycling's and is only legal on a hand ability. `DiscardACard()` /
`DiscardCardsMatching(n, label, match)` / `DiscardN(n, label)` are the
general clause — Fauna Shaman's "Discard a creature card" — whose picks
the activator names at announce (`discard_ids`), like a sacrifice cost's.
Both pay through the one discard helper with cause COST, so every
discard payoff sees them and none of them can pause (CR 601.2h /
602.2b).

`effects.Register` panics at boot on a non-battlefield ability that
declares a tap, sacrifice-this, crew or loyalty component: none of them
has a permanent to pay with. That is the check to read if a new hand
ability refuses to boot.

### Special actions from the hand (#658, #659)

A CR 116.2 special action is **not** an ability and **not** a cast: it
uses no stack, there is no announce and nothing to respond to. There
is ONE verb for all of them — `special_action {card_id, kind}` — and
one engine entry, `game.PerformSpecialAction`, with a per-kind timing
table beside the kinds. See
[ADR 0062](docs/decisions/0062-abilities-and-special-actions-from-the-hand.md)
Decision 4.

| kind | window | split second |
|---|---|---|
| `foretell` (CR 702.143a) | any time you have priority during **your** turn | **legal** (CR 702.61b) |
| `suspend` (CR 702.62a) | any time you could begin to **cast** the card — sorcery timing for a sorcery, instant timing for an instant | **illegal** (CR 702.62c imports it) |

That asymmetry is the one thing here that is easy to get wrong and
invisible when you do. `legal/special_actions.go` deliberately does
NOT open with the `if g.SplitSecondActive { return }` that
`legal/cast.go` and `legal/abilities.go` open with; it asks
`SpecialActionTimingOKLocked` per kind, which is the same function the
engine refuses with, so the enumerator and the engine cannot drift.

Declare one on the card with the keyword constructors, never by hand:

```go
SpecialActions: []game.SpecialAction{Foretell("{1}{U}")},   // Foretell {2}, cast later for {1}{U}
SpecialActions: []game.SpecialAction{Suspend(1, "{R}")},    // Suspend 1—{R}
```

`effects.Register` refuses a kind the engine cannot carry out, an
unparseable cost, and a suspend with no time counters, at boot.

Both keywords ride models that already exist and neither adds a
second one: the later cast is a per-instance `game.CastPermission`
(ADR 0066) scoped to that one card object, foretell's exile is
`Card.FaceDownKind = foretold` (ADR 0069, viewers = the owner), and
suspend's countdown is TWO triggered abilities that declare
`Zones: {ZoneExile}` (#925) — the upkeep counter removal and, since
#990, "when the last time counter is removed", which watches the
counter event so Clockspinning and Vampire Hexmage end a countdown
the same way an upkeep does. A card-level `CastableZones: exile`
declaration is the WRONG shape for either and was retired on #659: it
opens exile for every copy of the card, at any time, however the copy
got there — so a Path to Exile'd Rift Bolt would be castable.

### Adding a `Spec` slot (#622)

The engine reads the catalog through one precomputed `game.CardDef`
per card, built at `Register`. A new slot is four edits: the field on
`effects.Spec`, the field on `game.CardDef`
([carddef.go](server/internal/game/carddef.go)), one line in
`effects.buildDef` ([carddef.go](server/internal/cards/effects/carddef.go)),
and the engine call site that reads it. Add a per-slot
`game.CatalogX` variable only if a game-package test needs to stub
that slot without importing the catalog; the existing ones default to
reading the `CardDef` and are not set by the catalog any more.

### When in doubt

[ADR 0010](docs/decisions/0010-card-effect-catalog.md) captures every
architectural decision and sandbox simplification the catalog was built
around. Start there.

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
