# ADR 0122 — An agent at the table: a local MCP seat

**Status:** Proposed · 2026-10-04 · S62 — An agent at the table. Amended the same day with the owner's review answers (decisions 6–8), and again for the Go version (decision 8 as amended).
**Issues:** [#2230](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2230) (this change, and S62's tracker).
**Owner decisions:** the five answers of 2026-10-04 on #2230 and the three review answers of the same day, quoted under [Owner decisions](#owner-decisions-2026-10-04). They are binding. This ADR also makes calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before PR 2 lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 38 remote heads, then on all 182 local branches. The highest number on any of them was 0120, on `origin/develop`. 0121 was reserved for animated dice ([#2229](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2229)), and it has since landed on `develop`. This ADR takes **0122**.
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) (the legal-move enumerator, the `Policy` type gate in §3, the funnel in §5), [ADR 0052](0052-bot-decision-harness-and-eval.md) (the funnel's measurements and the prompt's board rendering), [ADR 0044](0044-surviving-a-deploy.md) (HMAC session tokens that survive a deploy, the reconnect ladder), [ADR 0051](0051-user-database.md) (guest seats have no `user_id`), [ADR 0105](0105-legal-action-highlights.md) (the `legal_actions` digest), [ADR 0055](0055-loop-breaker.md) (the loop notice), [ADR 0075](0075-table-settings-and-host-controls.md) §2.1 (the host), [ADR 0121](0121-animated-dice.md) (the opening roll the agent will also see).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The owner wants to play Claude Code, Codex, or any other MCP client as an extra player at a real table. The client should play as a guest would, not as a privileged bot. The table should always know the seat is an agent.

The first draft's claims were checked on `origin/develop` at `1c2b9773`. The review amendment's were checked at `dc4f1844`.

### Owner decisions (2026-10-04)

From #2230, verbatim:

1. **Shape:** an MCP server that sits in a normal human seat (option 1). It gets no special access: it sees the same filtered view and legal-move list any seat gets over the WebSocket.
2. **Sign-in:** a guest seat joined by invite link. No access to the owner's account.
3. **Disclosure:** the seat carries a visible "AI agent" badge for the table, recorded by the server.
4. **Auto-pass:** only real choices reach the model. Trivial windows (empty priority passes, forced moves, mana steps) are answered the way bot Layer A answers them.
5. **Hosting:** a local stdio binary that Claude Code or Codex launches. It connects out to the game server over the normal WebSocket, with no new server endpoint.

From the owner's review of the first draft, the same day:

6. **The move cap is fixed on the server.** When `capLegalMoves` trims the list, the wire carries an explicit truncation flag, and a seat can fetch its own full, uncapped legal-move list on demand: scoped to the seat, filtered like the rest of its view, never another seat's moves. The client-side variants go; the agent always chooses from the full closed list. *This overturns the first draft's call 12* (variants for attacks, blocks and one-target moves, worked out from the digest). The design is §6.1–§6.3. The request is a frame on the existing socket, not an HTTP route, so decision 5's "no new server endpoint" still holds for routes.
7. **A server acknowledgement for successful actions.** When an action is applied, the server echoes its id, additively on the wire, so a client knows exactly what landed. `act` reports acceptance from the ack, not from "the sequence advanced". *A new decision.* It replaces the first draft's inference from `seq` (§3). The design is §6.4.
8. **Upgrade Go and use the official SDK.** A delivery PR before the binary's moves the module to **Go 1.27**, and the binary uses `modelcontextprotocol/go-sdk`. *This overturns the first draft's call 1* (a hand-rolled stdio JSON-RPC server, chosen because the SDK needs Go 1.25). The owner's first answer named Go 1.25, the SDK's floor. The owner then chose 1.27, the newest release, because 1.25 no longer gets security fixes: Go supports only its two newest major releases, and 1.27 shipped on 2026-08-18, which left 1.26 and 1.27 supported. That settles the first amendment's call 1, which had flagged 1.25's support status and recommended 1.26. The design is §1 and §1.1, with the fallback rule for the linter in §1.1.

### What exists

**The wire already carries a seat's moves.** `GameView.legal_moves` is the closed list of what the viewer's own seat may do right now, straight out of `internal/legal` (`protocol/view.go:201`). Each entry is a `legal.Move` (`legal/legal.go:98`). Its `{type, player, params}` is exactly the action payload that performs it, so a client can send it back unaltered. `FilterViewFor` gives each seat its own list and no one else's. The field is empty unless the seat owes a decision: priority, a pending choice, a combat declaration, the mulligan window or a cleanup discard.

**The wire list is capped, the cap degrades it, and nothing says so.** On top of the enumerator's own caps (below), `legalMovesWireCap = 48` (`protocol/view.go:3988`) bounds the list on the wire. Past 48, `capLegalMoves` (`:4009`) keeps the **first** move of every `(source, kind, targets_stack)` tuple and drops the alternatives. Every card that has a move keeps one, but the choice between its targets is lost. That includes a creature's choice of which player to attack and a blocker's choice of which attacker to block. Nothing on the wire says a list was degraded. The `legal_actions` digest (ADR 0105, `protocol/legal_actions.go:109`) is built from the uncapped list, so a client can infer it by comparing counts, but no field states it. The in-process bot is unaffected: the runner calls `legal.EnumerateFor` directly (`aiseat/runner.go:463`) and never reads the wire.

**The enumerator has caps of its own.** On `dc4f1844`, `internal/legal` stops early in these places:

- `defaultMaxExpansionPerSource = 12` (`legal.go:422`): moves per source across targets, modes and cost choices.
- `defaultMaxX = 20` (`:423`).
- `maxEnumeratedVariableCounts = 3` (`abilities.go:880`).
- `subsetScanBudget = 64` (`choices.go:1213`).
- `creatureTypeAnswersCap = 5` (`choices.go:1451`).
- `cardNameAnswersCap = 5` (`choices.go:1501`).
- `maxEnumeratedCostPayments = 3` (`cast.go:488`).
- `maxEnumeratedRepeats = 3` (`cast.go:496`).

The bot lives with all of them, and none is reported. For the bot they are a cost bound, since Layer B scores each move on a cloned game. Some are policy judgements, for example "what is worth naming" for Pithing Needle.

**The browser is not hurt by the wire cap today.** Every browser reader of `legal_moves` asks a question the cap's key preserves:

- `classifyMove` (`client/src/lib/responseWindow.ts:90`) reads the kind, `targets_stack`, the sorcery window and the source's controller. That is everything autopass and `hasResponse` / `hasPlay` need.
- `canCastFromHand` (`timing.ts:237`) asks whether a card has any cast or land move.
- `hasPassMove` reads the digest first.
- Highlights read the uncapped digest, and `legalActions.ts:55` already marks the capped fallback as inexact.
- Targeting reads `CardView.legal_targets`, never the move list (docs/protocol.md on `legal_moves`).

So the browser needs no fetch. What it lacks is the flag. A future reader that wanted alternatives from `legal_moves` would be silently wrong, as the agent would have been.

**There is no hello frame and no success acknowledgement.** A WebSocket connection is bound at the upgrade, from the session (`Authorization: Bearer`, cookie or `?token=`) and `?game=` / `?player=` (docs/protocol.md, "Connection lifecycle (S04)"). After that the client sends `action`, `chat` and `ping` frames. Nothing on the socket says what kind of client is connected. A failed action gets an `error` frame carrying the action's `id`. A successful one gets no reply of its own, only the next broadcast `snapshot`, whose `id` is always empty.

`Client.handleAction` (`ws/hub.go:885`) runs on the connection's read goroutine. It calls `room.Apply`, then `broadcastToRoom`, which queues a filtered snapshot on every connection's `send` channel, this one included. A full channel disconnects the client rather than dropping a frame (`sendRaw`, `:1371`). The browser's frame switch logs any unknown `kind` as a protocol error (`client/src/lib/ws.ts:965`).

**Joining.** `POST /games/{id}/join` takes `{invite_token, name}` (`lobby/http.go:806`) and mints a `player` session. A guest seat's session has no `user_id` and lasts `CMDCTRL_SESSION_TTL`, 12 h by default. Because the HMAC key survives a deploy (ADR 0044), so does the token. A seat can only be claimed before the game starts (409 after). The route shares the 1 req/s, burst 5, per-IP bucket with `/admin/login` (`http.go:331`). A seat needs a deck before the host can start: `POST /games/{id}/decks` takes a pasted list or a pre-built deck id. `GET /cards/{id}` serves Scryfall metadata, including oracle text, to any session. The WebSocket has no action rate limit, only a 64 KiB inbound frame cap (`ws/hub.go:38`). `checkOrigin` admits a request with no `Origin` header (`hub.go:194`), which is what a non-browser client sends.

**Where seat facts live.** `game.Player.IsBot` / `BotTier` / `BotDeck` ride the engine snapshot (`game/player.go:228`, `game/snapshot.go:600`) and are projected to `PlayerView.is_bot` (`protocol/view.go:1452`). The lobby reads them back from the engine on restore rather than giving them a column (`lobby/persist.go:371`, `loadEntry`: "the rest of SeatInfo … rides the engine snapshot on game.Player already"). `SnapshotSchemaVersion` is 7. An additive field is recorded with `-update-shape` and needs no bump (AGENTS.md §5).

**Layer A is a pure function, but it lives under the runner.** `rules.Resolve(aiseat.Input) Verdict` (`aiseat/rules/rules.go:108`) absorbs four kinds of window:

- `forced`: exactly one move.
- `mana-only`: the only alternatives to passing are mana abilities. Casts auto-tap, and the pool empties at the step boundary (CR 106.4).
- `same-land`: the only alternatives are copies of one land.
- `coin-call`: a heads-or-tails call with no stop.

Everything else escalates. `Resolve` imports `aiseat` for `aiseat.Input` (`aiseat/policy.go:36`), `PassIndex` and `CoinCall`. `aiseat` is the runner, and it imports `internal/game`, `internal/ws` and `internal/actions`. The type gate (`aiseat/heuristic/imports_test.go`) bans a **direct** import of `internal/game` from every package under `aiseat/`. It says why the ban cannot be transitive: `protocol` and `legal` import `game` themselves. The runner also stops a bot's pass while the CR 732 loop notice is up (`aiseat/runner.go:555`), because that pass is as automatic as a browser's autopass. `Resolve` does not check the notice itself.

**The board rendering is not importable as is.** The model tier's board text is `(*model.Policy).buildDelta` (`aiseat/model/prompt.go:156`). It is a method that renders the board and the moves together, reads the heuristic's ranking and `Policy.cfg`, and is unexported. The board half needs nothing but the view, the seat and two zone caps. ADR 0052 §6's glossary, recent-events block and window framing are not built yet.

**How much a seat decides.** ADR 0033 §5's update measured the funnel at four seats: 5,207 windows over two games and 7,810 over three, so about **650 windows per seat per game**. Layer A absorbed about 90% of them: `mana-only` 51.4%, `forced` about 38%, `same-land` 0.5%. That leaves about **60–65 windows per seat per game** for judgement. These are heuristic bots, and a model plays differently, so this is an order of magnitude, not a promise.

**The Go toolchain, and every place it is pinned.** I grepped the whole tree on `dc4f1844` for `golang:`, `go-version`, `go 1.`, `setup-go`, `GOTOOLCHAIN` and `golangci-lint`. The pins:

| Where | What it pins today |
|---|---|
| `server/go.mod:6` (and its comment, `:3-5`) | `go 1.22` |
| `.github/workflows/ci-cd.yml:165` | `go-version: "1.22"` in the `&setup-go` anchor of `server-check`, reused by alias in `server-test` (`:248`) and `build` (`:423`), which builds the deployed binaries (`:432`, `:437`) |
| `.github/workflows/ci-cd.yml:569` | `go-version: "1.22"` in `census-publish` |
| `.github/workflows/ci-cd.yml:209-227` | the lint step's comment and `GOTOOLCHAIN: auto`, which exists only because golangci-lint v1.64.8 needs Go ≥ 1.23 while the job runs 1.22 |
| `.github/workflows/e2e-nightly.yml:145`, `:239`, `:372`, `:682` | `go-version: "1.22"` in `bot-games`, `catalog-soak`, `bot-soak` and `playwright` |
| `scripts/go-docker.sh:15` | `GO_IMAGE` `golang:1.22` |
| `scripts/go-docker.sh:16` | `LINT_IMAGE` `golang:1.24` |
| `scripts/go-docker.sh:17` | `LINT_PKG` golangci-lint `v1.64.8` |
| `scripts/go-docker.sh:35-38`, `:95` | the usage text, and `GOTOOLCHAIN=auto` on lint |
| `server/Makefile:72-78` | `GOLANGCI_LINT` `v1.64.8` |
| `server/.golangci.yml` | golangci-lint v1 configuration format |
| `AGENTS.md` §5, "Go without a local toolchain" | `golang:1.22` (`golang:1.24` for golangci-lint) |

Not pins: `.devcontainer/devcontainer.json` installs the Go feature at `"version": "latest"`. `.devcontainer/Dockerfile` is a Node image. No Dockerfile builds the server: the deployed binaries are built by the `build` job above.

Comments that state the 1.22 pin as a fact, and that go stale with it:

- `aiseat/model/anthropic.go:18-30` and `aiseat/model/openai.go:27`: why those clients do not use an SDK.
- `game/cast_ban.go:92`, `game/cast_permission.go:273`, `game/player_statics.go:133` and `game/snapshot.go:798`: why there is no `omitzero`. `omitzero_tag_guard_test.go` enforces it, and it stays.

The toolchain releases on 2026-10-04, from the Go module proxy:

- Go 1.25.0 shipped 2025-08-08. The newest patch is 1.25.14.
- Go 1.26.0 shipped 2026-02-10. The newest patch is 1.26.8.
- **Go 1.27.0 shipped 2026-08-18.** Go's policy supports the two newest major releases, so **1.25 no longer receives security fixes**.

**golangci-lint and Go 1.27.** v1.64.8, the pin today, is the final v1 release, and it cannot lint a module that declares a newer Go than the one it was built with. The v2 line can. Each v2 release requires a minimum Go to `go run`:

- v2.4.0 to v2.8.0 need `go 1.24.0`.
- v2.10.0 to v2.12.x need `go 1.25.0`.
- v2.13.0 and later need `go 1.26.0`. The project's own `go.mod` comment says that floor is always "latest-1".

**Go 1.27 support arrived in v2.13.0** (published 2026-08-19, the day after Go 1.27.0). Its release notes list "go1.27 support (#6642)". v2.13.1, v2.13.2 and v2.14.0 (2026-09-24, the newest on the proxy on 2026-10-04) follow it. I checked it rather than read it: `golangci-lint@v2.14.0`, run under `golang:1.27` (go1.27.1) through `scripts/go-docker.sh` with an image override and only the shared volumes, reported "built with go1.27.1". It linted a scratch module declaring `go 1.27.0` with 0 issues. v2 reads a different configuration format, which `golangci-lint migrate` converts. Among other changes, `gofmt` and `goimports` move to a `formatters` section, and `disable-all` becomes `linters.default: none`.

**The SDK.** `github.com/modelcontextprotocol/go-sdk` v1.8.0 was published 2026-09-04 and declares `go 1.25.0`. I read its module source from the proxy. **It builds under Go 1.27.** I built a scratch module declaring `go 1.27.0` with a stdio server and one tool on `golang:1.27` (go1.27.1), through `scripts/go-docker.sh` and the shared volumes. `go mod tidy`, `go vet` and `go build` all passed, and the scratch module was deleted afterwards. `go list -deps` confirmed the linked set below.

- **Stdio.** `mcp.NewServer`, `mcp.AddTool[In, Out]` (`mcp/server.go:603`), whose handler type is `ToolHandlerFor[In, Out]` (`mcp/tool.go:57`), and `(*Server).Run(ctx, &mcp.StdioTransport{})` (`mcp/transport.go:128`). Inbound lines are capped by `MaxLineLength`, default 16 MiB.
- **Tests.** `mcp.NewInMemoryTransports()` (`:183`).
- **Protocol versions.** 2026-07-28, 2025-11-25, 2025-06-18 and 2025-03-26 (`mcp/shared.go:51-54`). Negotiation is the SDK's.
- **What a binary would link.** The `mcp` package imports the SDK's `auth` package, which imports `golang.org/x/oauth2`, so OAuth2 is linked even into a stdio-only binary. Also linked: `segmentio/encoding` (and `segmentio/asm`) for its JSON, `google/jsonschema-go`, `yosida95/uritemplate`, `golang.org/x/time/rate` and `golang.org/x/sync/errgroup`.
- **What the module graph adds without linking.** `golang-jwt/jwt/v5`, `google/go-cmp` and `golang.org/x/tools` are added to `go.sum`, and minimum version selection raises `golang.org/x/sys` from v0.30.0 to v0.41.0 for the whole server.

**The MCP clients on this workstation.** Claude Code 2.1.289: `claude mcp add [options] <name> <commandOrUrl> [args...]`, `-s/--scope local|user|project`, `-e KEY=value`, and `--` before the server's own flags. It also has `--allowedTools`, `--strict-mcp-config` and `--permission-mode`. codex-cli 0.160.0: `codex mcp add <NAME> -- <COMMAND>...` with `--env KEY=VALUE`, configuration in `~/.codex/config.toml`, and `-s/--sandbox read-only|workspace-write|danger-full-access`. Both were read from `--help` on this machine. Nothing else about either client was checked.

---

## Decision

### 1. One binary on the official SDK, with the SDK behind one file

The seat is `server/cmd/mcpseat` (a `main` of about 40 lines: flags, signal handling) over a library at `server/internal/mcpseat`. It is built with `make -C server build-mcpseat` to `server/bin/cmd_and_ctrl-mcpseat`. That target is `go build` like the others. On this workstation, which has no local Go, run `scripts/go-docker.sh go build -o bin/cmd_and_ctrl-mcpseat ./cmd/mcpseat`. It is not deployed. Like `boteval`, it is a local tool.

**It uses `github.com/modelcontextprotocol/go-sdk`, pinned at v1.8.0** (decision 8). It uses stdio only: `mcp.NewServer`, `mcp.AddTool` for each tool, and `Run` with `&mcp.StdioTransport{MaxLineLength: 1 << 20}`, since nothing a client sends this seat is near a megabyte. Protocol version negotiation, `ping`, cancellation (the SDK cancels a handler's context on `notifications/cancelled`, which ends a pending `wait_for_decision`) and the JSON-RPC framing are the SDK's. The binary does not use the HTTP transports, OAuth, sampling, resources or prompts. Stdout carries nothing but protocol, and every log line goes to stderr.

**The SDK sits behind one file.** `internal/mcpseat/transport.go` is the only file that imports the SDK. It does three things:

- It registers each tool. Each tool's input is a plain struct of ours, with `json` and `jsonschema` tags. Those are tags, not SDK types, and the SDK infers each tool's input schema from them.
- It converts each handler's result. Every handler has the shape `func(ctx context.Context, in X) (Result, error)`, where `Result` is ours: `{Text string, IsError bool}`. The file converts it to `*mcp.CallToolResult`.
- It keeps `clientInfo` from the initialize request for the badge (§7).

The tool handlers, the seat's state, the WebSocket client and the renderer never see an SDK type. A later SDK version, or a return to a hand-written transport, is a change to that file.

**What it costs the module.** The server module moves to Go 1.27 first (§1.1). The binary then links `x/oauth2`, `segmentio/encoding` and `segmentio/asm`, `jsonschema-go`, `uritemplate`, `x/time` and `x/sync`. `go.sum` gains `golang-jwt/jwt/v5`, `go-cmp` and `x/tools`, which are not linked. `x/sys` rises to v0.41.0 for every binary in the module. None of this reaches `cmd/server` or `cmd/bot`, because they do not import the SDK. PR 7 lists `go mod why` for each new module in its description, and checks with `go version -m bin/cmd_and_ctrl-server` that the server binary links none of them.

#### 1.1 The Go 1.27 upgrade

A PR of its own, before the binary's (Delivery PR 6). It changes every pin in the table under [What exists](#what-exists), and nothing else:

- **The module.** `server/go.mod` to `go 1.27.0`, with its comment saying why. It is the newest supported release (decision 8), above the MCP SDK's floor of 1.25, and it keeps method-aware `ServeMux`, `slog` and range-over-int as before.
- **CI.** Every `go-version: "1.22"` to `"1.27"`: the `&setup-go` anchor in `ci-cd.yml` and its aliases (`server-check`, `server-test`, and `build`, which builds the deployed binaries), `census-publish`, and the four `e2e-nightly.yml` jobs (`bot-games`, `catalog-soak`, `bot-soak`, `playwright`).
- **Lint.** golangci-lint to **v2.14.0**, the newest release, which runs on and lints Go 1.27 (checked above). It goes in `server/Makefile`'s `GOLANGCI_LINT` and `scripts/go-docker.sh`'s `LINT_PKG`, under its v2 module path `github.com/golangci/golangci-lint/v2/cmd/golangci-lint`. `server/.golangci.yml` is converted with `golangci-lint migrate`, keeping the same linters, the same exclusions for test files and the 6 m timeout. The diff is reviewed for any linter whose meaning changed.
- **The lint workaround.** Because the tool's floor (1.26) is below the module's Go, the lint step's `GOTOOLCHAIN: auto` and its comment block in `ci-cd.yml` go, and so does `-e GOTOOLCHAIN=auto` in `go-docker.sh`.
- **`scripts/go-docker.sh`.** `GO_IMAGE` to `golang:1.27`, and `LINT_IMAGE` to `golang:1.27` too: one image again. The usage text (`:35-38`) is updated. **The only cache volumes stay `cmdctrl-gomod`, `cmdctrl-gobuild` and `cmdctrl-golangci`**, as AGENTS.md §5 requires. golangci-lint v2 keeps its cache where v1 did, so the third volume is reused. `go-docker.sh prune` runs once after the switch, because the build cache of a new toolchain shares nothing with the old one.
- **The devcontainer.** The Go feature from `"latest"` to `"1.27"`, so a fresh container builds with what CI builds with.
- **AGENTS.md §5.** "golang:1.22 (golang:1.24 for golangci-lint)" becomes "golang:1.27 for both".
- **Stale comments.** The six listed under What exists are reworded. `anthropic.go` and `openai.go` keep their reasons (dependency weight) without the version argument. The four `omitzero` comments keep their rule and drop "CI's pinned 1.22", because the guard test still holds every toolchain to one encoding.
- **What the new toolchain checks.** Language and vet changes that apply once `go.mod` says 1.23 or later:
  - Go 1.23's synchronous timer channels. The only Stop-and-drain idiom in the tree is `lobby/practice.go:333`, and it is read for the new semantics.
  - `go vet`'s printf check for non-constant format strings (Go 1.24).
  - The `tests` analyzer (Go 1.24).
  - The `waitgroup` and `hostport` analyzers (Go 1.25).
  - Anything the 1.26 and 1.27 release notes add to the language, the standard library's defaults or `go vet`. The PR lists what it found.

  The PR fixes whatever these report, and runs the full test suite, the snapshot corpus test and the nightly workflows on its branch (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`).

**The fallback rule.** The PR pins the newest golangci-lint release whose notes say it supports the target Go, and verifies it by running it on that toolchain against the converted module. That is v2.14.0 for Go 1.27 today. If no release supports 1.27 when the PR is written (for example, if v2.14.0 turns out not to run cleanly on the real tree under 1.27, and no patch release fixes it), the PR **falls back to Go 1.26** everywhere above, which is still supported, with golangci-lint v2.13.2 or later. It says so in its description and in an amendment here. A linter that does not support the toolchain is never kept by turning lint off or by running it under a different Go than CI's.

### 2. What the binary may import, and why the line is real

The binary holds the line the bot holds (ADR 0033 §3): it sees only what the wire gives its seat. Here that line is physical before it is a type rule. The binary runs on the owner's workstation and holds a WebSocket to a server elsewhere, so there is no `*game.Game` in its process to read, whatever it imports.

The type rule is kept anyway, so the code says the same thing. `TestMCPSeatImportsNoServerState` in `internal/mcpseat` fails on a **direct** import of any of these, from `internal/mcpseat` or `cmd/mcpseat`:

- `internal/game`
- `internal/ws`
- `internal/lobby`
- `internal/actions`
- `internal/db`
- `internal/auth`
- `internal/aiseat/tiers`

It checks `Imports` and in-package `TestImports`. The end-to-end test (§10) is an external test package (`mcpseat_test`). It must stand a server up, so `XTestImports` is exempt, and the test says why. Like the `aiseat` gate, it has a control test that fails if the banned path stops matching anything.

Allowed:

- `internal/protocol` and `internal/legal`, the wire types. `LegalMoveView` is `legal.Move`.
- `internal/aiseat`, for `aiseat.Input` and `PassIndex` only.
- `internal/aiseat/rules` (§4).
- The extracted board renderer (§5).
- `internal/util/redact`.
- The SDK, in `transport.go` only (§1). A second gate test enforces that.

As with the bot, the import of `aiseat` transitively links `game` into the binary. `protocol` already does that for `gamecli`. A linked package is not a handle on the table.

### 3. The tools

Ten tools. Every one returns text sized for a model's context. None returns the session token, a raw frame or an action payload. Moves are named by **index into a numbered list the binary showed**, and every index refers to one window. That list is always the seat's full list (§6). The binary never builds an action from anything else.

| Tool | Input | Output |
|---|---|---|
| `join` | `invite_url` (string, required), `display_name` (string, default `"Agent"`), `deck` (optional: `{"id": "<pre-built deck id>"}` or `{"list": "<decklist text>"}`) | `game_id`, seat number, player id, table state, `resumed: true` when it reattached to a seat it already holds. On a refusal, the server's reason (wrong invite, full, started). |
| `set_deck` | `deck` as above | The deck name, or the server's validation problems. An unknown id answers with the ids this server has (the route's own 422). Before the game starts only. |
| `wait_for_decision` | `timeout_s` (1–50, default 25), `pass_until` (`"none"` default, or `"my_turn_or_stack"`, §4) | `status`: `decision`, `waiting` (the timeout passed with nothing to decide: call again), `not_started`, `eliminated` or `game_over` (with the outcome). With `decision`: `window` (an opaque token), the window's kind (`mulligan`, `priority`, `response`, `attack`, `block`, `choice:<kind>`), the compact view (§5), the numbered moves, and, since the last decision, the public log lines, chat lines and a one-line count of what the binary answered automatically. |
| `get_state` | `detail`: `compact` (default) or `full` | The view at that budget (§5). |
| `legal_moves` | `card` (optional: an instance id) | The current window's full numbered list, grouped by card, with kind, label, cost note and idle hint. With `card`, that card's moves expanded past the enumerator's caps (§6.2). Each card the enumerator still cut says so, with how many. |
| `card` | `ref`: an instance id this seat can see, or a card name on the table | Name, type line, mana cost, power and toughness, oracle text and the `unimplemented` flag, from `GET /cards/{scryfall_id}`, cached per id. A face-down or hidden card answers with what the seat may see, which is nothing more than the view's own text. |
| `act` | `window` (required), `move` (index, required), `value` (only for a move whose answer is an open set, §6.2) | `accepted` (the server's `ack`, §6.4), `rejected` with the server's code and message, `stale` with the new window, or `unknown` (the connection dropped before either arrived: read the state). |
| `say` | `text` (1–500 characters) | `sent`, or `rate_limited` with when to retry. |
| `concede` | `confirm: true` (required) | `conceded`. |
| `leave` | none | Disconnects and deletes the saved session. While the game is active and the seat is not eliminated, it refuses and says to concede first, because a vanished seat holds the table. |

**`act` is checked locally before it is sent.** If `window` is not the current one, because a snapshot moved the board while the model thought, the binary refuses with `stale` and returns the new window rather than sending a move that might mean something else now. Pending-choice option lists can shrink while open (docs/protocol.md on `player` options), so an index from an old list is never reused.

**How `act` reports.** The binary sends the move's `{type, player, params}` as an `action` frame with a fresh `id`. It reports `accepted` when an `ack` frame with that `id` arrives (§6.4), and `rejected` when an `error` frame with that `id` arrives. Nothing is inferred from other seats' snapshots. If the socket drops first, the answer is `unknown`, and the next `wait_for_decision` reads the state as it is. After three rejections in one window, `act` accepts only the window's `always_legal` move. The binary still never picks it on the model's behalf.

**Concede is out of band, as it is for the bot.** `legal` does not enumerate concession (`aiseat.Conceder`'s doc comment). Conceding is every player's right (CR 104.3a), so it gets its own tool, with a `confirm` flag so it cannot be an accident of argument order.

**No sandbox verbs and no improvisation.** The agent cannot `move_card`, `change_life`, `force_cast` or `undo`. None of those are in `legal_moves`, and decision 1 gives it no more than a seat's closed list. When a card it cast is `unimplemented` and resolves to nothing, the compact view says so, and the agent's recourse is the one a guest has without the sandbox: say in chat what the card should have done, so a human can resolve it. ADR 0033 §8's improvisation stays a bot feature.

### 4. Layer A, imported and not re-implemented

The binary calls `rules.Resolve` on every window, building `aiseat.Input{View, Seat, Moves}` from the frame. **`Moves` is always the full list:** when the frame says `legal_moves_truncated`, the binary fetches the full list (§6.1) before Layer A or the model sees anything. One copy of the rules means the agent's trivial windows are answered exactly as the bot's are, which is what decision 4 asks. When `Resolve` takes the window, the binary sends that move at once, with no `MinThink` delay, the way a browser's autopass does. It counts the move for the next decision's summary.

**Which windows are trivial**, i.e. never shown to the model:

| Layer A rule | Absorbed for the agent | Why |
|---|---|---|
| `forced`: exactly one legal move | yes | Nothing to decide. Mostly the pass-only window. |
| `mana-only`: the only alternatives to passing are mana abilities | yes | Casts auto-tap, and floating mana empties at the step boundary (CR 106.4). |
| `coin-call`: heads or tails with no stop | yes | Both calls have the same odds. |
| `no-moves` | yes (nothing is sent) | Degenerate. |
| `same-land`: the only alternatives are copies of one land | **no, escalates** | It plays a land, which is an action, not a pass. Holding a land drop for after combat (landfall) or for a discard outlet is a decision, and the owner's list ("empty priority passes, forced moves, mana steps") did not include it. It costs about one model call per turn. |

Whatever ADR 0121 PR 3 adds to Layer A for the opening roll applies here unchanged. Rolling with nothing else on offer is a `forced` window. Choosing who starts is a real choice and reaches the model.

One more rule comes from the runner, not from `Resolve`. **While `loop_notice` is up, nothing is passed automatically** (CR 732, ADR 0055). The notice is the server telling every client that a person has to ask for the next iteration, and here the agent is the person. **Combat declarations always escalate**: they carry no pass, so `Resolve` already sends them on unless exactly one move is offered.

**The owner may tune it only toward the model.** `--absorb=forced,mana-only,coin-call` lists the absorbed rules. A rule can be removed so more windows reach the model. A rule cannot be added, and there is no flag that answers anything Layer A would not. `same-land` can be switched on with the same flag if one call per turn turns out not to be worth it.

**`pass_until` is the model's own autopass, off by default.** With `wait_for_decision(pass_until: "my_turn_or_stack")`, the binary also passes a priority window on another player's turn when the stack is empty, nothing is owed and the window is not a combat declaration. It stops at once for anything on the stack, a pending choice, a combat declaration, the loop notice, or the start of the seat's own turn. The start of the turn also clears it, as the client's safety belt does (ADR 0009 §7). This is the convenience a human gets from the browser's autopass toggle. It is the model's choice for its own seat, and it never answers a choice. A model holding an instant for an opponent's end step leaves it off.

### 5. The compact view

**The board half of `buildDelta` moves to `internal/aiseat/boardtext`**, a pure `Render(view *protocol.GameView, seat string, opts Options) string`. It renders:

- the turn line;
- each seat, own first: life, hand and library counts, mana pool, end gates, emblems, battlefield, graveyard, own hand and command zone;
- the stack, bottom first, with ownership and targets;
- the seat's owed choices.

`model.buildDelta` calls it and keeps the move half. That PR is a pure refactor, and the prompt tests hold its bytes. The new package sits under `aiseat/`, so the existing import gate covers it from day one. The binary imports it, which makes the agent and the bot read the same board text. A fix to one is a fix to both. One thing is added for the agent, behind an `Options` flag the bot leaves off: an `(unimplemented: the engine does not run this card's text)` note on a card whose `CardView.unimplemented` is set.

**Budgets**, measured in UTF-8 bytes of the tool's text, at about 4 bytes per token:

| Output | Cap | How it is held |
|---|---|---|
| The board in `wait_for_decision` and `get_state(compact)` | 6,000 bytes (~1,500 tokens) | `MaxZoneCards` 24, as the bot uses. Over the cap, graveyards drop to counts first, then opponents' battlefields to the 12 highest-power permanents with "and N more", then log and chat lines from oldest. |
| The moves in `wait_for_decision` and `legal_moves` | **never cut** | Grouped by card: one header with the card and its cost, then one short line per alternative carrying its index. A cut list would be a partial list again (§6). On a large board this is the part that grows. The binary reports its size at game end (§9). |
| `get_state(full)` | 24,000 bytes (~6,000 tokens) | No zone caps, the last 24 public log lines, and every revealed card. |
| `card` | 2,000 bytes | One card. |

The budgets are enforced by the binary and tested (§10). A compact board is about the size of the bot's own board block, without its decklist.

### 6. What the server adds to the wire

One server PR (Delivery PR 5) adds four things to the protocol, all additive, all in `docs/protocol.md`, none in the snapshot: a flag on the view, a request frame and its reply, a report of the enumerator's own cuts, and an acknowledgement.

**Why, in terms of honesty.** A list with the alternatives silently dropped was never a list of the seat's real choices. It was a list of some of them, presented as all of them. A model choosing from it is not choosing among its legal moves. It is choosing among the moves a bandwidth bound happened to keep, without knowing a choice was made for it. Decision 1 promised the agent "the same legal-move list any seat gets", and a person at the table is never limited that way: they click any target the rules allow. The first draft's client-side variants patched three common shapes and left the rest partial. The fix belongs where the list is made, so every client that reads it can know whether it has the whole thing.

#### 6.1 The truncation flag and the full list on request

- **`GameView.legal_moves_truncated: true`**, own seat only, projected out of the per-seat map exactly like `legal_moves`. It is set when `capLegalMoves` dropped anything from this seat's list, and absent otherwise.
- **`legal_moves_request`** (client to server): `{"v": 0, "kind": "legal_moves_request", "id": "<uuid>", "payload": {"source"?: "<instance id>"}}`. It is answered for the connection's **bound seat only**, the same seat `FilterViewFor` filters for. A spectator, an admin with no `?player=`, or a seat that owes no decision gets an `error` with the request's `id`. There is no field naming a seat, so the request cannot ask about another seat's moves.
- **`legal_moves`** (server to client, the reply, same `id`): `{"seq", "generation", "moves": [...], "truncated"?: [...]}`. `moves` is the seat's enumeration with **no wire cap**, built under the room's lock by the same `legal.EnumerateLocked` call the view uses. `seq` and `generation` name the state it describes. A client discards a reply whose `seq` or `generation` is not the window it is deciding in, and asks again after the next snapshot. The moves carry the same knower redaction as the view's own list, because they are the same list: every card in them is one this seat may see.
- **Cost.** It is an enumeration the server already performs for every view, done once more for one seat. One request per connection is answered at a time, and a connection gets at most 4 per second. Past that, it gets an `error`, so a misbehaving client cannot turn the room's read lock into a loop.

#### 6.2 The enumerator's own cuts, reported and expandable

The wire cap is not the only place a list is cut (the eight caps under What exists). The same PR makes `internal/legal` report its cuts. Each cap site records what it would have offered and did not: the source card or the pending choice, the cap's name, and how many moves were left out. That record reaches the client in two places:

- The full-list reply carries `truncated: [{source | choice, cap, omitted}]`. So does the digest, as `legal_actions.sources[id].truncated`.
- A request with `source` re-enumerates that one card (a new `Source` filter on `legal.Options`) with every count cap raised to a hard ceiling of **512 moves**. Anything still cut is reported, never hidden.

**Two answers are open sets by the rules, and are not listed point by point.**

- **A card name.** CR 201.2 lets a player name any card. The enumerator's five suggestions and its "Pithing Needle" fallback stay as moves, but the move is also marked `value: {"kind": "card_name"}`. `act` may then supply any name, which the server already validates as it does a human's (trimmed, non-empty, at most 200 characters).
- **X.** The move is marked `value: {"kind": "x", "min": a, "max": b}`, from the enumerator's affordable range.

These are the only places `act` passes a value, and only inside a set the server stated. Creature types are a vocabulary of about 300 entries, so under the 512 ceiling they are listed in full.

**The runner does not use this, and that is deliberate.** The in-process bot never read the wire, so the 48-move cap and the request frame are nothing to it. It keeps its 12-per-source cap. It scores each move on a cloned game, so expansion multiplies its decision time, and `OrderTargets` (#687) already ranks a source's candidates so that the 12 it keeps are the ones that matter. What the bot gains is the report. The model tier's prompt can say which cards have more legal moves than it was shown, rather than "N further legal moves not listed" for the whole window. That is a one-line follow-up in `model/prompt.go`, not part of S62.

**The browser does not need the fetch** (What exists). Its readers ask only what the cap preserves. PR 5 adds the flag to `client/src/lib/protocol.ts` and leaves a comment at `legalActions.ts`'s fallback that the flag says when the fallback is inexact. No browser behaviour changes.

#### 6.3 Why a frame and not a route

A route would need its own authentication path and its own rate limit, and would answer outside the room's sequence. On the socket, the request already carries the connection's binding, and the reply can name the `seq` it describes. Decision 5 asked for no new server endpoint. This keeps that for HTTP and adds one request kind to the socket the seat already uses (decision 6).

#### 6.4 The acknowledgement

- **The frame.** `{"v": 0, "kind": "ack", "id": "<the action's id>", "payload": {"seq": <uint64>, "generation": <uint64>}}`, sent to the **originating connection only**, after every action the server applies. That covers ordinary actions through `room.Apply`, table settings through `ApplyExternal`, and `undo`. `seq` and `generation` are those of the state the action produced.
- **Ordering.** `handleAction` runs on the connection's read goroutine, and `broadcastToRoom` queues the snapshot on every connection's `send` channel before returning. The ack is queued on the same goroutine immediately after. So **on the originating connection the ack always follows the snapshot of the state it acknowledges**. A client that sees the ack already holds that state, or a newer one. Snapshots from other seats' actions may arrive between the two, and the ack's `seq` says which state was this action's.
- **Failure.** It is unchanged: an `error` frame with the action's `id`, and never an ack.
- **The browser.** It may use the ack, for example to clear a pending action's spinner on exactly the right frame instead of on the next snapshot. That is optional. What is not optional is that `ws.ts` handles the kind, because today it logs any unknown kind as a protocol error (`ws.ts:965`) and would do so on every action. PR 5 adds `case "ack"`, which records the frame in the protocol log at debug level and does nothing else. A browser holding an older cached client logs the unknown kind until it reloads. That is harmless, and the ADR accepts it rather than gating the frame per connection.

The ack and §6.1–§6.2 are in one PR (call 13). They change the same two files, `ws/hub.go` and `docs/protocol.md`, and are reviewed as one protocol change.

### 7. The badge: declared at join, one way, on the seat

**The agent declares itself in the join request.** No hello frame exists to carry a client kind, and a join field needs no new endpoint. So both join routes (`POST /games/{id}/join` and `POST /join`) accept an additive body field:

```json
{ "invite_token": "…", "name": "Claude", "agent": { "client": "claude-code" } }
```

`client` is the MCP client's `clientInfo.name` from `initialize`, cut to 32 characters of `[a-z0-9._-]`, or `"unknown"`. The binary always sends `agent`. It has no flag to leave it out.

**The server records it on the seat, and it never comes off.** `game.Player` gains `Agent bool` and `AgentClient string`, set by `JoinAs` before the seat is visible to anyone. Nothing clears them: no route, setting, reclaim or restore.

- **In the snapshot.** They are `isAgent` and `agentClient`, with `omitempty`. That is an additive change within schema 7, recorded with `-update-shape`, and copied in `clone.go`.
- **On restore.** `loadEntry` reads them back the way it reads `IsBot`, so **no database migration** is needed. A rollback reads the snapshot and drops two unknown keys.
- **On the wire.** `PlayerView` gains `is_agent` and `agent_client`, and so does `SeatInfo`, which is what the lobby list and the invite preview carry. All are public and identical for every viewer. Being an agent is a fact about the seat, like `is_bot`.

**Honest in the direction that matters.** The badge cannot be removed, so a seat that joined as an agent cannot later pass as a human. The reverse spoof, a human claiming to be an agent, harms nobody. What the server cannot do is tell a dishonest client from a person. A modified binary, or a script on a guest seat, could join without the field. The badge is a declaration by a cooperating client, and docs/mcp-seat.md says so. No technical check is offered as if it closed that gap.

**What the table sees.** The client renders an "AI agent" chip in the slot the bot chip uses (`PlayerIdentity.svelte:240-246`), so a seat is a bot, an agent or neither, never two. The chip also appears on the lobby seat list, the invite preview and the chat author line. Its tooltip reads "Played by an AI agent (<client>). It sees only what this seat sees." While the seat holds priority, the chip reads "thinking…", exactly as `botThinking` does for a bot (`:147`). That is derived from the priority holder, which every viewer already has, with no presence signal.

**Rules that follow from decisions 1 and 2:**

- An agent join that carries a signed-in session is refused with 400 "an agent seat joins as a guest". The binary never sends one. This makes decision 2 a server rule rather than a binary habit, and it means an agent seat never has a `user_id`.
- An agent seat is never the host, as a bot seat is not (`lobby/host.go`'s hand-over predicate gains `!p.Agent`). Changing the table's rules is not a seat's move.
- `GET /auth/discord/link` refuses an agent seat, so it stays a guest.

### 8. Security

**The session token.**

- **Where it lives.** It is kept in memory and in one file per seat: `$XDG_STATE_HOME/cmdctrl-mcpseat/<server host>/<game_id>.json`, falling back to `~/.local/state/…`. The file holds the origin, game id, player id, token and expiry. The directory is created `0700` and the file is written `0600` through a temp file and rename. The binary refuses to read a file or directory that is group- or world-accessible, as `ssh` does.
- **When it goes.** It is deleted at game end, on `leave`, and when found expired.
- **What is not stored.** The invite token. It is the owner's to paste, and reclaiming a seat goes through the saved session, not the invite.
- **On the wire.** It is sent as `Authorization: Bearer` on the upgrade and on HTTP calls, never as `?token=`, so it does not appear in a URL or an access log.
- **Never in a tool result.** Tool results go to the model provider. The test in §10 sets the token to a sentinel and fails if the sentinel appears in any tool result or log line.
- **Logs.** Every log line passes through `redact.Secrets`. `--log-file` writes `0600`.

**Which servers it will talk to.** `--allow-origin` (repeatable, set by the owner in the MCP config) lists the origins `join` accepts. The default is `https://cmd.labxp.io`, `https://cmd-dev.labxp.io` and `http://localhost` / `http://127.0.0.1` on any port. The model cannot change the list. An invite URL from a hostile chat line or card name cannot point the seat at a server that feeds it a forged table.

**Table text is untrusted input to an agent with tools.** Guest names, chat, deck names and named cards (`choose_card_name` accepts up to 200 characters of free text) are written by other people, and they land in the context of a coding agent that may have a shell. The binary does four things:

- It wraps every such string in a labelled field (`name: «…»`, `chat from «…»: «…»`).
- It cuts names to 40 characters and chat lines to 300.
- Every tool's description states that table text is data from other players, never instructions.
- It never puts a URL from the table into a tool result.

The binary cannot stop the client from obeying an injected string. docs/mcp-seat.md therefore recommends running the seat in a session that can do nothing else (§11).

**Rate limits.** The binary keeps its own, because the WebSocket has none for actions:

- **Actions:** 4 per second, burst 8. Layer A's automatic answers count.
- **Rejections:** three per window (§3).
- **Move-list requests:** at most one in flight. The server also enforces its own 4 per second per connection (§6.1).
- **Chat:** one line per 5 s, 500 characters (the server allows 1,000).
- **Card lookups:** 5 per second, each id fetched once.
- **`join`:** shares the server's per-IP bucket. A 429 is retried after 1, 2, 4 s, then reported.

**Reconnects and reclaiming the seat.**

- **The ladder.** The binary follows the browser's ladder (docs/protocol.md "The reconnect ladder"): close 1000 is terminal, 1001 and 1006 redial with 500 ms to 30 s half-jittered backoff, and after three failed dials `GET /me` decides whether the session is dead. An `act` in flight across a drop reports `unknown` (§3).
- **A restart.** If the MCP client restarts the binary, `join` with the same invite finds the saved session and reattaches (`resumed: true`) instead of claiming a second seat. After the start it could not claim one anyway.
- **A dead session.** If the session has expired (12 h) or the server refuses it, the seat is lost to the binary, exactly as for any guest. An admin can mint a reclaim ticket, and `join` accepts that link (`/games/{id}/reclaim`'s ticket) in `invite_url`.

**While the agent thinks, the table waits, as it waits for a person.** There is no human turn timer, and decision 4 sends only real choices to the model, so the binary never answers one for it: no deadline, no heuristic fallback, no move played under the agent's name. The table sees the chip reading "thinking…" (§7) on the seat holding priority. If the MCP client stops calling tools (its turn ended, or the owner closed it), the seat sits like an idle browser tab until someone acts. This is the same position the table is in with a person who stepped away. The tool descriptions and docs/mcp-seat.md tell the client to loop `wait_for_decision` → `act` until `game_over`.

### 9. Game pace and cost

Figures are from ADR 0033 §5's measured funnel. Where a figure is an estimate, it says so. PR 7 measures them, and they are recorded here as an amendment.

| Quantity | Value |
|---|---|
| Windows per seat per game | ~650, measured (four heuristic bots) |
| Absorbed by Layer A | ~90%, measured |
| Reaching the model | **~60–70 per game**: ~60–65 measured, plus ~one `same-land` window per turn |
| With `pass_until` on | fewer; unmeasured |
| Tool calls per decision | 2–4 (`wait_for_decision`, `act`, sometimes `card`, `legal_moves(card)` or `get_state`) |
| Context added per decision | ~2,000–2,500 tokens: a ≤1,500-token board, the moves, the model's reasoning and the `act` round trip. Estimate; a large board's full move list adds more |
| Context added per game | ~150k tokens. Estimate |
| Agent time per game | 15–40 s per decision gives **15–45 minutes** of the table waiting on the agent, on top of the game itself. Estimate |

**Tokens.** A long tool loop resends its whole context on each step. With prompt caching, most of that is cache reads, but it is still counted. Take 65 decisions, a ~25k-token client baseline, and ~2.5k added per decision. The model then reads about `65 × 25k + 2.5k × 65² / 2 ≈ 7M` input tokens per game, mostly cached, and writes about 40k. The context reaches ~175k by the end. That is near a 200k window, so expect one compaction per game, or use a long-context model.

What that costs depends on the model and the plan: dollars on an API key, a share of the usage limit on a subscription. This ADR states no price, because none was measured. To keep it down: use a fresh session per game, use the compact view (the default), and turn `pass_until` on when holding nothing at instant speed.

**What the binary reports.** At `game_over` it logs one line to stderr and returns the same in the tool result:

- windows seen, how many each rule absorbed, and how many were shown to the model;
- `act` calls and rejections;
- truncated windows, full-list requests, and the largest move list shown, in moves and bytes;
- bytes returned in tool results, an approximation of tokens;
- the time from each decision opening to its `act`.

The binary cannot see the model's own token count. The client reports that.

### 10. Tests

**The end-to-end test.** `internal/mcpseat/e2e_test.go` (external package) stands up the lobby handler and WebSocket hub on `httptest.NewServer`, the same stand-up `lobby/http_test.go` uses. It creates a two-seat table with a `random` bot on the whole-game tests' vanilla decks.

- **How it drives the seat.** Through the SDK's own client over `mcp.NewInMemoryTransports()`: `initialize`, `tools/list`, `join`, `set_deck`, then a scripted chooser that loops `wait_for_decision` → `act`, picking the pass or move 0, until `game_over` or a turn bound. The test, not a model, plays the agent's side.
- **A capped window.** One scripted board puts the seat past the 48-move cap, to exercise the truncation flag and the full-list request end to end.
- **What it asserts.** Every window the test is shown escalated under §4's table. Every `act` is answered by an ack or an error, never inferred. The badge is on `is_agent` in the bot's view of the table and in `GET /games`. No board text exceeds its budget. The sentinel token never appears in a tool result or log line. A stale `window` is refused.
- **Waits.** The test polls monotone predicates through one `waitFor` helper (AGENTS.md §5, #634) and has no `select` on `time.After`.
- **Cost.** It runs in every CI run, is sized to finish in seconds, and makes **no network model call**, because there is no model in it.

**Unit tests (binary).**

- The tool handlers, called directly with no SDK in the way. That is what keeping the SDK in one file buys.
- `transport.go`'s schema registration, for each tool.
- The import gates and their controls.
- Grouping the moves, and the `value` check against a stated set.
- The state file's permission refusal.
- The redaction of every log path.

**Server side, the wire (PR 5).**

- **The flag.** `legal_moves_truncated` is set exactly when `capLegalMoves` dropped a move, on the boards `legal_moves_test.go` already caps, and never for another seat.
- **The request.** It answers for the bound seat only, and refuses spectators, an unseated admin and a seat with no decision. Its moves equal the uncapped enumeration. It carries the room's `seq`. It is refused past 4 per second.
- **The cut report.** Every cap site in `legal` reports what it cut. A `source` request lifts the caps to 512 and reports anything still cut.
- **The ack.** It is sent for every applied action through all three apply paths, never for a refused one, and always after its own snapshot on the originating connection. Other connections never see it.
- **The browser.** `ws.ts` accepts `ack` without logging an error.

**Server side, the badge (PR 2).** The join field and its refusal with a signed-in session; the badge's persistence across `CaptureSnapshot` → restore → `loadEntry`; the snapshot shape file; `PlayerView` and `SeatInfo` carrying it for every viewer; the host hand-over skipping an agent; the Discord link refusing one.

**Client side.** A vitest for the chip, its "thinking…" state, and the bot chip and agent chip never both showing.

**The Go upgrade (PR 6)** is tested by the whole suite, the snapshot corpus, lint under v2.14.0 and the nightly workflows run on its branch, as §1.1 says.

### 11. Docs

A new `docs/mcp-seat.md`, linked from `docs/bot.md`'s introduction and from AGENTS.md §5, rather than a section of `docs/bot.md`. The bot guide is about seats the server runs. This one is about a seat the owner runs, so it carries different warnings. It covers:

- building the binary;
- the tools and the flags;
- the badge, and what it does and does not prove;
- the token file;
- the untrusted-text warning;
- the pace and cost of §9;
- the recipes below.

**Claude Code** (syntax checked against 2.1.289's `--help`):

```sh
claude mcp add -s user cmdctrl-seat -- \
  /home/luke/repos/cmd_and_ctrl/server/bin/cmd_and_ctrl-mcpseat \
  --allow-origin https://cmd-dev.labxp.io
```

To play, start a session in an empty directory that can use only the seat's tools, for example `claude --allowedTools "mcp__cmdctrl-seat__*"`. Then give it the invite link and "play this game to the end: loop wait_for_decision and act".

**Unverified for Claude Code:**

- the exact tool-name pattern Claude Code gives an MCP server whose name has a hyphen (check `/mcp`);
- which `--permission-mode` denies, rather than asks about, everything else;
- how long Claude Code waits on one tool call (the `MCP_TOOL_TIMEOUT` environment variable). This is why `wait_for_decision` defaults to 25 s and stops at 50.

**Codex** (syntax checked against codex-cli 0.160.0's `--help`):

```sh
codex mcp add cmdctrl-seat -- \
  /home/luke/repos/cmd_and_ctrl/server/bin/cmd_and_ctrl-mcpseat \
  --allow-origin https://cmd-dev.labxp.io
```

or, in `~/.codex/config.toml`:

```toml
[mcp_servers.cmdctrl-seat]
command = "/home/luke/repos/cmd_and_ctrl/server/bin/cmd_and_ctrl-mcpseat"
args = ["--allow-origin", "https://cmd-dev.labxp.io"]
```

To play, run `codex -s read-only` in an empty directory.

**Unverified for Codex:**

- the `[mcp_servers.<name>]` table shape beyond what `codex mcp add` writes (check the file it produces);
- the per-server tool timeout key. `tool_timeout_sec` is believed to exist, with a default of 60 s.

**Elsewhere:**

- AGENTS.md §3 gains `cmd/mcpseat/` and `internal/mcpseat/` in the layout.
- AGENTS.md §5 gains a short "MCP seat" entry: the make target, the flags, and the pointer to docs/mcp-seat.md.
- docs/protocol.md gains `PlayerView.is_agent` / `agent_client` (PR 2), and `legal_moves_truncated`, `legal_moves_request`, the `legal_moves` reply and `ack` (PR 5).
- docs/lobby.md gains the `agent` join field and its 400.

---

## Delivery

Each PR goes into `develop`, Sprint S62, Issue #2230.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **This ADR**, the AGENTS.md §3 ADR range line, and the S62 section in `docs/sprints.md`. Docs only. | — | anything |
| 2 | **Server: the badge** (§7). The `agent` join field on both join routes and its 400 with a signed-in session; `game.Player.Agent` / `AgentClient`, the snapshot keys and shape file, `clone.go`, `loadEntry`; `PlayerView` and `SeatInfo`; the host hand-over and Discord-link refusals; docs/protocol.md and docs/lobby.md; its tests (§10). | 1 | 4, 5, 6 |
| 3 | **Client: the chip** (§7) on the seat, lobby list, invite preview and chat, with its vitest. | 2 | 4, 5, 6 |
| 4 | **Refactor: `aiseat/boardtext`** (§5). Byte-identical bot prompts, held by the existing prompt tests. | 1 | 2, 3, 5, 6 |
| 5 | **Server: the wire** (§6). `legal_moves_truncated`; `legal_moves_request` and its reply; the cut report in `internal/legal` and the digest, the `Source` filter and the 512 ceiling; the open-set `value` markers; the `ack` frame on all three apply paths; `ws.ts`'s `case "ack"` and the protocol type; docs/protocol.md; its tests (§10). | 1 | 2, 3, 4, 6 |
| 6 | **Go 1.27 and golangci-lint v2.14.0** (§1.1), or Go 1.26 under the fallback rule. Every pin in the table under What exists, the v2 lint configuration, the devcontainer, AGENTS.md §5, the stale comments, and whatever the new vet analyzers report. Nightly workflows run on the branch before merge. | 1 | 2, 3, 4, 5 |
| 7 | **The binary and its tools** (§1–§5, §8, §9) on `modelcontextprotocol/go-sdk` v1.8.0. `internal/mcpseat` with the SDK in `transport.go` only, `cmd/mcpseat`, `make build-mcpseat`, the import gates, the end-to-end test and the unit tests (§10). `go mod why` for each new module in the description. | 2, 4, 5, 6 | 3 |
| 8 | **Docs** (§11). `docs/mcp-seat.md`, the AGENTS.md §3 layout and §5 entries, the pointer from docs/bot.md. Then one real game on cmd-dev, with its §9 numbers appended to this ADR as an amendment. | 7 | — |

PR 6 changes the toolchain under every other open branch. Each Go PR open at the time merges `develop` after it and reruns CI. If PRs 2, 4 or 5 merge first, they are simply built again by PR 6's CI.

After PR 8: the owner plays one game on cmd-dev with Claude Code in a seat. The other seats should see the chip, and the game should finish. The numbers go into this ADR, and #2230 is closed with that evidence.

## Consequences

- The owner can seat Claude Code, Codex or any MCP client at a real table as a guest, with the same view and moves a person in that seat has, minus the sandbox.
- Every agent seat is marked, permanently and publicly. The mark relies on a cooperating client, and the docs say so.
- A seat's move list now says when it is partial, and any seat can have the whole of it. That includes the enumerator's own cuts, which were invisible to the bot as well.
- Every applied action is acknowledged to the client that sent it, so no client has to infer success from the sequence number again.
- The server gains two snapshot keys and one join field. On the wire it gains three view fields, one request kind, two reply kinds and a digest field. It gains no HTTP route and no migration.
- The whole server module moves to Go 1.27, the newest supported release, and to golangci-lint v2. The binary links `x/oauth2` and four small modules through the SDK, and `x/sys` rises for every binary.
- The bot's board text moves to a package of its own, so the two seats read the same board.
- A table with an agent is slower than a table of bots, by roughly the agent's thinking time on 60–70 decisions.
- **A timing tell.** Layer A passes instantly, so an agent seat that holds priority for more than a moment has a real choice in front of it. Opponents can read that, as they can read a human's smart autopass. #1307's "considering…" chip hides the same tell for humans only partly. A uniform delay on the agent would hide it better, at the cost of slowing every trivial window. This ADR does not add one (call 7).

## Out of scope

- **A hosted or remote MCP server** (HTTP transport). Decision 5 is a local stdio binary, and the binary uses the SDK's stdio transport only.
- **Agent seats run by the server.** That is the bot (ADR 0033), and the bot keeps its tiers.
- **Improvisation and sandbox verbs for an agent** (§3).
- **A turn timer** for any seat, agent or human.
- **Several agent seats from one binary.** One binary process holds one seat. Two seats are two MCP server entries.
- **Agent-to-agent private channels.** Chat is the table's, and agents use it like anyone.
- **Raising `legalMovesWireCap` for the browser frame.** The cap stays as the browser's bandwidth bound. A seat that needs the rest asks for it (§6.1).
- **Lifting the bot's 12-per-source cap**, and the bot prompt's use of the cut report (§6.2). Both are follow-ups.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review. The first draft's call 1 (a hand-rolled transport) was overturned by decision 8, and its call 12 (client-side variants) by decision 6. The rest are renumbered.

1. **Go 1.27, with Go 1.26 as the fallback if no golangci-lint release supports 1.27 when PR 6 is written** (§1.1). Decision 8 fixed the version. The fallback rule is this ADR's.
2. **golangci-lint v2.14.0 with the configuration migrated by its own tool, one image for test and lint, and the devcontainer pinned to 1.27** (§1.1).
3. **The SDK at v1.8.0, stdio only, imported by `transport.go` alone, behind our own handler and result types** (§1, §2).
4. **`internal/mcpseat` plus a thin `cmd/mcpseat`, built by `make build-mcpseat`, not deployed** (§1).
5. **The import gate bans server-state packages from the binary and exempts only its external end-to-end test** (§2).
6. **`same-land` is not absorbed for the agent** (§4). Playing a land is an action, and its timing can matter. `--absorb` can turn it back on. **The owner may only remove Layer A rules, never add**, and **the loop notice stops every automatic pass** (CR 732), as it does for the bot.
7. **Layer A answers immediately, with no `MinThink`** (§4), as a browser's autopass does. The cost is a timing tell (Consequences).
8. **`pass_until` exists, off by default, chosen by the model per call, and cleared at the seat's own turn** (§4).
9. **The board renderer is extracted to `aiseat/boardtext` and shared with the bot** (§5), with an `unimplemented` note added for the agent.
10. **Budgets of 6,000 bytes for the board, 24,000 full and 2,000 per card, as plain text; the move list is never cut, only grouped** (§5).
11. **The full list is a request frame on the socket, answered for the bound seat only, one at a time and at most 4 per second, stamped with `seq` and `generation`** (§6.1, §6.3).
12. **The enumerator reports every cut it makes. A one-card request lifts its caps to 512. A card name and X are offered as stated open sets that `act` may fill** (§6.2). **The in-process runner keeps its own enumeration and its 12-per-source cap** (§6.2).
13. **The ack goes to the sender only, after its own snapshot, on all three apply paths. It is in the same PR as the move list, and the browser gets a no-op handler for it** (§6.4).
14. **Every `act` names its window, and a stale one is refused locally. After three rejections in a window, only the always-legal move is accepted from the model** (§3). The binary never plays it unasked.
15. **`concede` needs `confirm: true`. `leave` refuses on a live seat. No sandbox verbs and no improvisation for the agent** (§3).
16. **Oracle text on demand through `card` and the existing `GET /cards/{id}`** (§3), not inlined in every frame.
17. **The badge is declared in the join body as `agent: {client}`, from the MCP `clientInfo.name`, with no off switch in the binary, and lives on `game.Player` and the snapshot with no database column** (§7).
18. **An agent join with a signed-in session is refused. An agent seat is never host and cannot link Discord** (§7).
19. **The token file's place, its permissions, the Bearer header, and no token in any tool result or log line** (§8).
20. **`--allow-origin`, set by the owner, defaults to the two labxp hosts and localhost** (§8).
21. **Table text is wrapped, cut and labelled as untrusted** (§8). The docs recommend a session that can do nothing else.
22. **Client-side rate limits: 4 actions/s, one move-list request in flight, 1 chat line per 5 s, 5 card lookups/s** (§8).
23. **No deadline on the agent's decisions** (§8). The table waits as it does for a person, and the binary never plays a move the model did not choose.
24. **`join` reattaches a saved seat and accepts a reclaim link** (§3, §8).
25. **The docs go in a new `docs/mcp-seat.md`, not in `docs/bot.md`** (§11).

## Amendment (2026-10-04, #2263): distribution

§1 left the binary as something you build: `make -C server build-mcpseat`, or `go install github.com/krakenhavoc/cmd_and_ctrl/server/cmd/mcpseat@develop`. Both need Go 1.27, which a contributor who only wants an agent in a seat may not have. **The owner chose prebuilt binaries on a GitHub Release**, built by a release workflow. The alternatives were `go install` only, and builds the owner sends by hand.

**The workflow.** `.github/workflows/mcpseat-release.yml`, on GitHub-hosted runners. It tests `./internal/mcpseat/... ./cmd/mcpseat/...` first. Then it builds `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=… -X main.commit=…"` for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 and windows/amd64. Each target becomes one archive, `mcpseat-<version>-<os>-<arch>.tar.gz` (`.zip` for Windows), and the run writes one `SHA256SUMS` over all five. Go comes from the minor version in `server/go.mod`, at its newest patch, so there is no second pin to move.

- **On a pull request** that touches the binary, its library, `go.mod`/`go.sum` or the workflow itself, it stops there. The archives and `SHA256SUMS` are workflow artifacts, and nothing is attested or released.
- **On a pushed tag `mcpseat-vX.Y.Z`** (or `mcpseat-vX.Y.Z-<pre>`), it then attests each archive's build provenance with `actions/attest-build-provenance`, and creates a **draft** GitHub Release for the tag with the archives and `SHA256SUMS`. A tag with a `-` part is marked a prerelease. A tag that does not match the pattern is refused.
- **`workflow_dispatch`** takes a tag and does the same. It refuses a tag that does not exist or does not match. GitHub offers a manual run only for a workflow on the default branch, so this path works once the file reaches `main`. Until then, a pushed tag is how the release path is exercised.

**The tag scheme.** `mcpseat-v*`. The prefix keeps these tags clear of Go's `server/vX.Y.Z` submodule tags, which the toolchain would read as versions of the server module.

**Draft, then the owner publishes.** The owner pushes the tag. The workflow never tags, and never publishes. It leaves a draft that the owner reads and publishes by hand. A rerun refreshes the assets of its own draft. It refuses to touch a published release, so a fix is a new tag.

**Permissions.** The workflow is `contents: read`. Only the attest job holds `id-token: write` and `attestations: write`, and only the publish job holds `contents: write`. Neither compiles anything. A release build starts from a cold Go cache, so no other run's cache reaches a published binary.

**The version.** `mcpseat --version` (or `-version`) prints the stamped version and commit, the Go version and the platform, for a bug report to quote. A local build prints `dev` and `unknown`, unless `go install …@vX` or a git checkout supplies them through the Go build info.

**Verifying a download.** Check the archive against `SHA256SUMS` (`sha256sum -c SHA256SUMS --ignore-missing`, or `shasum -a 256 -c` on macOS). Or check its provenance: `gh attestation verify <archive> -R krakenhavoc/cmd_and_ctrl` proves the file was built by this repository's workflow, and names the commit.

**Which server a release joins.** A release is cut from a commit, and the binary speaks the wire that commit's server speaks. A release cut from `develop` joins cmd-dev. It joins cmd.labxp.io only once the server half of that commit has been promoted to `main`.

**Not included:**

- Code signing and macOS notarization, which need paid certificates. The docs say how to clear the quarantine flag (`xattr -d com.apple.quarantine mcpseat`).
- Homebrew or Scoop.
- Auto-update.
