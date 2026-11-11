# Commands

What to run, from the repo root. The Go rules are in [AGENTS.md](../AGENTS.md) §5.

## Dev environment

The devcontainer installs Go, Node, and the GitHub CLI. Ports 3000, 5173, and
8080 are forwarded.

## Commands

### Go without a local toolchain

The workstation has no local Go. `scripts/go-docker.sh` is THE way to run it: `scripts/go-docker.sh test ./internal/roadmap/...`, `vet`, `lint`, `go <args>`, `prune` (no arguments prints usage). It mounts the repo, uses `golang:1.27` for both tests and golangci-lint (at the version `server/Makefile` pins in `GOLANGCI_LINT_VERSION`), passes the exit status through (never pipe it through `tail`), chowns any root-owned files it created in the repo back to you on exit (#3701), and cleans the build cache first when it passes `CMDCTRL_GOCACHE_MAX_GB` (default 15). `make -C server docker-test docker-lint` wrap it. Env vars do not leak in: `CMDCTRL_SCRYFALL_DUMP=<path>` is bind-mounted read-only and rewritten to its in-container path, the `AISEAT_*` knobs are forwarded when set, and `GO_DOCKER_ENV="A B"` forwards any other names (see the script header; #2157).

**Never create any other cache volume**, and never hand-write a `docker run` for Go. The only three are `cmdctrl-gomod`, `cmdctrl-gobuild` and `cmdctrl-golangci`; per-PR volumes filled the Docker partition at 178 GB (#2035).

### Top-level
- `make help` — list targets
- `make server-dev` — run the Go server on :8080
- `make client-dev` — run the Vite dev server on :5173
- `make test` — server tests, client typecheck and client unit tests (Vitest)
- `make server-dev-skip-validation` — like `server-dev`, but skips deck validation so a small test deck imports; never in prod
- `make lint` — `go vet` + client ESLint + Prettier check
- `make cost-report` — what the project has cost: human-team estimate, Claude API list price, subscriptions. `make cost-push` publishes your Claude Code usage for the weekly report; see [scripts/cost/README.md](../scripts/cost/README.md)

### Server (Go, `server/`)
- `make -C server dev` — run locally on :8080 (seeds a 4-player demo game)
- `make -C server test` — `go test -race -cover ./...`
- `make -C server vet` — `go vet ./...`
- `make -C server lint` — golangci-lint at the version `server/Makefile` pins, via `go run`. With `TOOLS_DIR=<dir>` it instead `go install`s the pinned binary once under `<dir>/<go release>/` and runs that; CI's `server-check` caches that directory on a hosted runner, so lint and vulncheck start without downloading or compiling either tool (#4958). `make -C server tool-pins` prints both pins.
- `make -C server fmt` — `gofmt -s -w .`
- `make -C server build` — produces `server/bin/cmd_and_ctrl-server`
- `make -C server vulncheck` — govulncheck at `server/Makefile`'s `GOVULNCHECK` pin (`TOOLS_DIR` works as for `lint`): known vulnerabilities the server can reach, in its Go standard library and modules; CI runs it in `server-check` (#3335). Without local Go: `scripts/go-docker.sh go run "$(make -s -C server tool-pins | grep govulncheck)" ./...`
- `make -C server build-boteval` — produces `server/bin/boteval`, the AI-bot evaluation harness (#837). `boteval probe [--endpoint URL] [--model ID] [--think] [--max-tokens N] [--context N]` sends ONE request in the funnel's exact shape to a model endpoint and prints prompt/completion tokens vs the client-side estimate and the server's context size (`--context`, or read from Ollama's `/api/ps`), `finish_reason`, whether a `reasoning` field came back, the parsed index and copied label, and a verdict for each of the three failures that make a model seat play like a heuristic seat: truncation, thinking, and an answer that names no listed move (`INDEX: out of range`, #2196). Always exits 0. Not deployed; local tool.
- `make -C server build-mcpseat` — produces `server/bin/cmd_and_ctrl-mcpseat`, the local MCP seat (ADR 0122): a stdio MCP server that sits Claude Code, Codex or any MCP client in a guest seat, badged as an AI agent. Not deployed. Flags `--allow-origin`, `--absorb`, `--log-file`, `--state-dir`, `-v`; the session token lives in a `0600` file under `$XDG_STATE_HOME/cmdctrl-mcpseat/`. Guide, warnings and recipes: [mcp-seat.md](mcp-seat.md). `mcpseat --version` prints the stamped version and commit. Prebuilt binaries come from `.github/workflows/mcpseat-release.yml` ([ADR 0122](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md), amendment of 2026-10-04): a PR touching the binary builds the five archives and `SHA256SUMS` as artifacts; a pushed `mcpseat-vX.Y.Z` tag (the prefix keeps clear of Go's `server/vX` module tags) also attests them and opens a **draft** release. Only the owner pushes those tags and publishes the draft; an agent never does either.
- `cd server && go run ./cmd/gamecli -addr ws://localhost:8080/ws` — drive the demo game from a terminal; reads action JSON on stdin or via `-script path.json`

## Scheduled jobs

- Play-rate coverage: `cd server && CMDCTRL_SCRYFALL_DUMP=../data/scryfall/default-cards.json go run ./cmd/playrate [-top 500,1000,2000] [-list 1000]` prints how many of the top-N EDHREC-ranked cards the catalog holds and lists the missing ones in rank order (#3352). It needs the dump, so CI does not run it; the census block stays registry-only. Through Docker: `CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json scripts/go-docker.sh go run ./cmd/playrate > /tmp/playrate.txt`.
- Cron: `scripts/scryfall-refresh.sh` — weekly refresh of the Scryfall default-cards dump (suggested cron: `0 5 * * 0`). It also saves Scryfall's set list (`GET /sets`) beside the dump as `data/scryfall/sets.json`, atomically; a failed set-list download warns and keeps the last good copy. The server reads it at boot, if present, so the catalogue's set picker folds subsets (bonus sheets, promo sets) into their parent and leaves out promo, Secret Lair and joke sets (#4790); without it the picker lists every set on its own. When the dump is fresh enough to skip, the script still fetches a missing `sets.json`.
- Card roles: `cmd/cardroles` runs weekly right after `scripts/scryfall-refresh.sh` (the `scryfall-refresh` workflow, Sundays 05:00 UTC) and writes `data/scryfall/card-roles.json` (ADR 0217 §7): each card's roles (removal, ramp, tutor, ...) from Scryfall's community Oracle tags (Tagger). It downloads Scryfall's **Oracle Tags bulk file** (`oracle_tags` in the bulk-data catalog, about 6 MB gzipped, two requests, no search paging) with `User-Agent: cmd_and_ctrl-cardroles/1.0` and an `Accept` header, as [Scryfall's API rules](https://scryfall.com/docs/api) ask. The role-to-tag table is `roleTags` in `server/cmd/cardroles/roles.go`; a tag that is renamed or matches no card fails the run, writes nothing and leaves the last good file. The server reads the file at boot (`cards.Index.LoadRoles`), so the workflow restarts it afterwards; a missing file means no tag roles, and `effects.CardRoles` then answers from the catalog `Purpose` and `cards/roles_overrides.go` alone. The file is generated runtime data under the gitignored `data/`: never hand-edit it, never put it in the database. Locally: `scripts/go-docker.sh go run ./cmd/cardroles -data-dir ../data` (add `-in oracle-tags.jsonl.gz` to read a downloaded copy instead of fetching). CD ships the binary as `/opt/cmd_and_ctrl/server/bin/cmd_and_ctrl-cardroles`.
- Draft ratings: `scripts/draft-ratings-refresh.sh [--out DIR] [--keep-raw DIR] [--from-raw DIR] [SET ...]`, run by hand (not a cron: the files are checked in and embedded in the binary, so a refresh is a commit). It writes `server/internal/draftratings/data/<set>.json` for every draftable set of `GET /draft/sets` (or the sets named) from 17Lands' public card ratings, with a generic User-Agent, one request per set and a pause between them (ADR 0222 §7, #5703). A card needs 500 games in hand (17Lands publishes no win rate under that) and a set needs 40 such cards, so a new set gets a file only after a few days and a set 17Lands does not cover gets none. Credit to 17Lands is on every screen that shows the numbers. Run it again as sets mature.
- Off-site backup: `scripts/backup-offsite.sh`, run nightly by `deploy/cmd-and-ctrl-backup.service` + `.timer` (as `cmdctrl`, data dir read-only) on both hosts. It uses restic to back up the data dir to a per-host Cloudflare R2 bucket, taking `db/cmdctrl.backup.sqlite` but never the live db, and skipping the `scryfall/`, `images/` and `avatars/` caches. Credentials are in `/etc/cmd_and_ctrl/backup.env` (`root:cmdctrl 0640`, separate from the server's env), written by the CD step "Ensure off-site backup" from the `CMDCTRL_R2_*` and `CMDCTRL_RESTIC_PASSWORD` values in the `prod` / `dev` GitHub environments. A missing value, or one still containing `REPLACE_ME`, is a `::warning::` and a disabled timer, never a failed deploy. **The restic password must never change once a repository exists**; its recovery copy is the owner's password manager. Runbook, including the restore: [docs/environments.md](environments.md#backups) (#1031).

## Client

- `cd client && npm install` — first-time setup
- `npm run dev` — Vite dev server on :5173, proxies `/ws` to the Go server
- `npm run build` — type-check + production build into `client/dist/`
- `npm run check` — `svelte-check` typecheck only
- `npm run lint` — ESLint + Prettier check
- `npm test` — Vitest unit and render tests
- `npm run format` — auto-format
