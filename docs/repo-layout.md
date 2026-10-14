# Repository layout

The full package tree. The short version is in [AGENTS.md](../AGENTS.md) §3.
The ADR index and numbering rules are in [decisions/README.md](decisions/README.md).

```
cmd_and_ctrl/
├── PLAN.md              # the charter: what C2 is, who the beta is for, scope, goals
├── AGENTS.md            # this file
├── Makefile             # top-level dev/test/lint targets
├── .devcontainer/       # Go + Node dev environment
├── server/              # Go game server (authoritative state, WebSocket + HTTP API)
│   ├── cmd/
│   │   ├── server/      # main package — serves :8080 with lobby + hub + cards routes
│   │   ├── mcpseat/     # local MCP seat: an MCP client plays a guest seat, badged as an AI agent (ADR 0122) — see docs/mcp-seat.md
│   │   ├── gamecli/     # dev WebSocket client for driving a game via v0 actions
│   │   ├── restorecheck/ # restore rehearsal and real-restore step: verify a restored tree, re-apply the account-deletion ledger (#3340)
│   │   └── snapshotscrub/ # strips player data from a restore point before it joins the snapshot corpus (#522)
│   ├── internal/
│   │   ├── game/        # authoritative domain: Game, Player, Zone, Card, Turn, mutations
│   │   │   └── testdata/snapshots/ # the restore-point fixture corpus: v<N>/ generated and frozen per schema version, real/ scrubbed from cmd-dev (#522)
│   │   ├── protocol/    # v0 wire format types + ViewOfGame + FilterViewFor
│   │   ├── actions/     # action type enum + Dispatch(Game, Action) router
│   │   ├── legal/       # legal-move enumerator — the closed move list a bot picks from and the client's timing lookup (ADR 0033 §1)
│   │   ├── ws/          # gorilla/websocket hub, Room, RoomManager, per-viewer broadcast
│   │   ├── mcpseat/     # the MCP seat's library: ten tools over stdio, Layer A imported from aiseat/rules, imports no server state (ADR 0122)
│   │   ├── aiseat/      # AI bot seats (S31): runner goroutine per bot, tiered policies, curated decks, announced improvisation — see docs/bot.md
│   │   │   └── boardtext/ # the board half of the bot prompt as a pure Render(view, seat, opts): shared with the MCP seat (ADR 0122 §5)
│   │   ├── auth/        # pluggable Authenticator interface + MemoryAuthenticator + HTTP middleware
│   │   ├── lobby/       # GameMeta registry, invite flow, lobby HTTP handler, WSAuthorizer, deck upload
│   │   ├── adminview/   # the admin views' read side: every SQL statement behind /admin/users* and /admin/games* (SELECT only, guarded by a test), their response types, and pure merges with the lobby's live tables; imports no lobby or ws (ADR 0124)
│   │   ├── cards/       # Scryfall index (streaming load) + disk-backed image cache + /cards routes
│   │   │   └── coverage/ # measures the live catalog; fails CI when the coverage docs or a card's Caveats stop being true
│   │   ├── catalog/     # /catalog routes (public) — what the engine automates + how completely (ADR 0042)
│   │   ├── roadmap/     # the curated registry of keywords, mechanics and engine seams behind the public roadmap; generates docs/engine-seams.md's open table (ADR 0092) and its closed list from docs/engine-seams/closed/ (#1461)
│   │   ├── playmat/     # a signed-in person's playmat (ADR 0128): decode + re-encode, SSRF-guarded link fetch, uuid-keyed files, users.playmat_id
│   │   ├── netguard/    # the outbound guard for an address a person chose: https only, public unicast checked at dial time (DNS rebinding), no redirects, timeout and body cap; AddrAllowed is shared with playmat (ADR 0182 §3)
│   │   ├── modelkeys/   # a signed-in person's own model key for their AI seats (ADR 0182, migration 0025): provider allowlist, sealed store, the seat client with the owner's limits and fixed failure sentences
│   │   ├── bugstore/    # bug-report artifacts: reporter screenshots (public, Camo-reachable) + pinned replays (admin-only)
│   │   ├── deck/        # decklist parsers (Moxfield, plain text) + Commander validation
│   │   ├── deckcoverage/ # a decklist's coverage report: every card bucketed manual / unreviewed / caveats / automated / no_effect (ADR 0095)
│   │   ├── deckrequests/ # which GitHub issue tracks each requested deck, and who asked when (ADR 0095, migration 0006)
│   │   ├── gamestats/   # a finished game's statistics record (ADR 0165): Compute from the final view + the whole log (protocol.FullLogOf), and the store behind migration 0023's three tables
│   │   ├── usersettings/ # a signed-in person's synced settings: one JSON object, 32 KiB cap, If-Match revision (ADR 0110 §4, migration 0008)
│   │   ├── tablesetups/  # the last setup each person started a table with: settings, bots, tablemates (ADR 0110 §5, migration 0008)
│   │   ├── moderation/  # the name and chat filter: a hand-kept denylist (denylist.txt) with normalisation, masking for chat, refusal for names, and the folded key a guest block uses; a test holds it clear of every card name (#2984, ADR 0075 §2.6)
│   │   ├── accountdelete/ # self-serve account deletion (ADR 0149, migration 0015): one SQL transaction, the tombstone, and Inventory — every column with its deletion treatment, held to the live schema by a test
│   │   ├── snapshotscrub/ # the scrubber behind cmd/snapshotscrub: generic-JSON rewrite, refuses snowflakes and emails (#522)
│   │   ├── capacity/    # the server's caps and their defaults: weighted table load, WebSocket connections, creation rates; policy and arithmetic only, the lobby and hub enforce them (ADR 0148)
│   │   ├── metrics/     # Prometheus metrics on our own registry, served on a loopback-only listener; route-pattern HTTP labels, closed label sets guarded by a test (ADR 0123)
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
├── scripts/             # scryfall-refresh.sh (weekly cron), backup-offsite.sh (nightly R2 backup), set-server-env.sh and ensure-monitoring-agent.sh (CD), ci-test-shard.sh (which packages each `server-test` shard runs, #2766) + one-off tools
├── deploy/monitoring/   # alert rules (+ promtool tests in rules/tests/) and Grafana dashboards the monitoring VM syncs from main (ADR 0123) — see docs/monitoring.md
├── data/                # runtime state (gitignored): Scryfall cache, snapshots, images
└── docs/
    ├── protocol.md      # v0 wire format spec
    ├── lobby.md         # lobby HTTP API reference
    ├── bot.md           # AI bot seat — user-facing guide (S31)
    ├── labels.md        # the contract labels, generated from client/src/lib/labels.ts (ADR 0125 §2)
    ├── engine-seams/closed/ # one fragment per closed seam; CI generates engine-seams.md's Closed list from them (#1461)
    ├── adding-cards.md  # the catalog card guide: recipes for cards, mechanics and engine seams (moved out of AGENTS.md §7, #1747)
    ├── sprints.md       # sprint plan
    └── decisions/       # ADRs; the index and numbering are in docs/decisions/README.md
```
