# ADR 0122 — An agent at the table: a local MCP seat

**Status:** Proposed · 2026-10-04 · S62 — An agent at the table
**Issues:** [#2230](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2230) (this change, and S62's tracker).
**Owner decisions:** the five answers of 2026-10-04 on #2230, quoted under [Owner decisions](#owner-decisions-2026-10-04). They are binding. This ADR also makes calls the answers did not cover. They are listed under [Calls made here](#calls-made-here) so the owner can overturn any of them in review, before PR 2 lands.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-04. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 38 remote heads, then on all 182 local branches. The highest number on any of them is 0120, on `origin/develop`. 0121 is reserved for animated dice ([#2229](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2229), branch `docs/s61-adr-0121-animated-dice`, which had no ADR file committed and was not pushed when I swept). This ADR takes **0122**.
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) (the legal-move enumerator, the `Policy` type gate in §3, the funnel in §5), [ADR 0052](0052-bot-decision-harness-and-eval.md) (the funnel's measurements and the prompt's board rendering), [ADR 0044](0044-surviving-a-deploy.md) (HMAC session tokens that survive a deploy, the reconnect ladder), [ADR 0051](0051-user-database.md) (guest seats have no `user_id`), [ADR 0105](0105-legal-action-highlights.md) (the `legal_actions` digest), [ADR 0055](0055-loop-breaker.md) (the loop notice), [ADR 0075](0075-table-settings-and-host-controls.md) §2.1 (the host).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The owner wants to play Claude Code, Codex, or any other MCP client as an extra player at a real table. The client should play as a guest would, not as a privileged bot. The table should always know the seat is an agent.

Every claim below was checked on `origin/develop` at `1c2b9773`.

### Owner decisions (2026-10-04)

From #2230, verbatim:

1. **Shape:** an MCP server that sits in a normal human seat (option 1). It gets no special access: it sees the same filtered view and legal-move list any seat gets over the WebSocket.
2. **Sign-in:** a guest seat joined by invite link. No access to the owner's account.
3. **Disclosure:** the seat carries a visible "AI agent" badge for the table, recorded by the server.
4. **Auto-pass:** only real choices reach the model. Trivial windows (empty priority passes, forced moves, mana steps) are answered the way bot Layer A answers them.
5. **Hosting:** a local stdio binary that Claude Code or Codex launches. It connects out to the game server over the normal WebSocket, with no new server endpoint.

### What exists

**The wire already carries a seat's moves.** `GameView.legal_moves` is the closed list of what the viewer's own seat may do right now, straight out of `internal/legal` (`protocol/view.go:201`). Each entry is a `legal.Move` (`legal/legal.go:98`). Its `{type, player, params}` is exactly the action payload that performs it, so a client can send it back unaltered. `FilterViewFor` gives each seat its own list and no one else's. The field is empty unless the seat owes a decision: priority, a pending choice, a combat declaration, the mulligan window or a cleanup discard.

**The wire list is capped, and the cap degrades it.** The enumerator caps target and mode expansion at 12 moves per source (`legal.Options.MaxExpansionPerSource`), and the bot sees that cap too. On top of that, `legalMovesWireCap = 48` (`protocol/view.go:3988`) bounds the list on the wire. Past 48, `capLegalMoves` (`:4009`) keeps the **first** move of every `(source, kind, targets_stack)` tuple and drops the alternatives. Every card that has a move keeps one, but the choice between its targets is lost. That includes a creature's choice of which player to attack and a blocker's choice of which attacker to block. The bot is unaffected because it enumerates in-process at full fidelity (the comment at `:3985`). A WebSocket seat is not. Nothing on the wire says a list was degraded. It can be worked out, though: the `legal_actions` digest (ADR 0105, `protocol/legal_actions.go:109`) is built from the **uncapped** list, and `sources[id].moves` gives each card's uncapped move count. Its `attack_targets` and `blocks` are exact, and so is every card's `legal_targets`.

**There is no hello frame and no success acknowledgement.** A WebSocket connection is bound at the upgrade, from the session (`Authorization: Bearer`, cookie or `?token=`) and `?game=` / `?player=` (docs/protocol.md, "Connection lifecycle (S04)"). After that the client sends `action`, `chat` and `ping` frames. Nothing on the socket says what kind of client is connected. A failed action gets an `error` frame carrying the action's `id`. A successful one gets no reply of its own, only the next broadcast `snapshot`, whose `id` is always empty.

**Joining.** `POST /games/{id}/join` takes `{invite_token, name}` (`lobby/http.go:806`) and mints a `player` session. A guest seat's session has no `user_id` and lasts `CMDCTRL_SESSION_TTL`, 12 h by default. Because the HMAC key survives a deploy (ADR 0044), so does the token. A seat can only be claimed before the game starts (409 after). The route shares the 1 req/s, burst 5, per-IP bucket with `/admin/login` (`http.go:331`). A seat needs a deck before the host can start: `POST /games/{id}/decks` takes a pasted list or a pre-built deck id. `GET /cards/{id}` serves Scryfall metadata, including oracle text, to any session. The WebSocket has no action rate limit, only a 64 KiB inbound frame cap (`ws/hub.go:38`). `checkOrigin` admits a request with no `Origin` header (`hub.go:194`), which is what a non-browser client sends.

**Where seat facts live.** `game.Player.IsBot` / `BotTier` / `BotDeck` ride the engine snapshot (`game/player.go:228`, `game/snapshot.go:600`) and are projected to `PlayerView.is_bot` (`protocol/view.go:1452`). The lobby reads them back from the engine on restore rather than giving them a column (`lobby/persist.go:371`, `loadEntry`: "the rest of SeatInfo … rides the engine snapshot on game.Player already"). `SnapshotSchemaVersion` is 7. An additive field is recorded with `-update-shape` and needs no bump (AGENTS.md §5).

**Layer A is a pure function, but it lives under the runner.** `rules.Resolve(aiseat.Input) Verdict` (`aiseat/rules/rules.go:108`) absorbs four kinds of window: `forced` (exactly one move), `mana-only` (the only alternatives to passing are mana abilities; casts auto-tap, and the pool empties at the step boundary, CR 106.4), `same-land` (the only alternatives are copies of one land) and `coin-call` (a heads-or-tails call with no stop). Everything else escalates. It imports `aiseat` for `aiseat.Input` (`aiseat/policy.go:36`), `PassIndex` and `CoinCall`. `aiseat` is the runner, and it imports `internal/game`, `internal/ws` and `internal/actions`. The type gate (`aiseat/heuristic/imports_test.go`) bans a **direct** import of `internal/game` from every package under `aiseat/`. It says why the ban cannot be transitive: `protocol` and `legal` import `game` themselves. The runner also stops a bot's pass while the CR 732 loop notice is up (`aiseat/runner.go:555`), because that pass is as automatic as a browser's autopass. `Resolve` does not check the notice itself.

**The board rendering is not importable as is.** The model tier's board text is `(*model.Policy).buildDelta` (`aiseat/model/prompt.go:156`). It is a method that renders the board and the moves together, reads the heuristic's ranking and `Policy.cfg`, and is unexported. The board half needs nothing but the view, the seat and two zone caps. ADR 0052 §6's glossary, recent-events block and window framing are not built yet.

**How much a seat decides.** ADR 0033 §5's update measured the funnel at four seats: 5,207 windows over two games and 7,810 over three, so about **650 windows per seat per game**. Layer A absorbed about 90% of them: `mana-only` 51.4%, `forced` about 38%, `same-land` 0.5%. That leaves about **60–65 windows per seat per game** for judgement. These are heuristic bots, and a model plays differently, so this is an order of magnitude, not a promise.

**The MCP libraries need a newer Go.** `server/go.mod` says `go 1.22`, and CI and `scripts/go-docker.sh` test on `golang:1.22`. The module has four direct dependencies. On 2026-10-04 the module proxy's latest releases were `github.com/modelcontextprotocol/go-sdk` v1.8.0 (`go 1.25.0`, eight direct requirements including `golang-jwt` and `x/oauth2`) and `github.com/mark3labs/mcp-go` v1.1.1 (`go 1.25.5`, seven, including `testify` and `spf13/cast`). Taking either one moves the whole server module to Go 1.25.

**The MCP clients on this workstation.** Claude Code 2.1.289: `claude mcp add [options] <name> <commandOrUrl> [args...]`, `-s/--scope local|user|project`, `-e KEY=value`, and `--` before the server's own flags. It also has `--allowedTools`, `--strict-mcp-config` and `--permission-mode`. codex-cli 0.160.0: `codex mcp add <NAME> -- <COMMAND>...` with `--env KEY=VALUE`, configuration in `~/.codex/config.toml`, and `-s/--sandbox read-only|workspace-write|danger-full-access`. Both were read from `--help` on this machine. Nothing else about either client was checked.

---

## Decision

### 1. One binary, a thin `main`, and a hand-rolled stdio server

The seat is `server/cmd/mcpseat` (a `main` of about 40 lines: flags, signal handling, stdio) over a library at `server/internal/mcpseat`. It is built with `make -C server build-mcpseat` to `server/bin/cmd_and_ctrl-mcpseat`. That target is `go build` like the others. On this workstation, which has no local Go, run `scripts/go-docker.sh go build -o bin/cmd_and_ctrl-mcpseat ./cmd/mcpseat`. It is not deployed. Like `boteval`, it is a local tool.

**It speaks MCP over stdio with about 400 lines of its own code, not a library.** MCP's stdio transport is newline-delimited JSON-RPC 2.0. A tools-only server needs six messages:

- `initialize`: answer with the protocol version, `capabilities: {tools: {}}` and `serverInfo`, and keep the client's `clientInfo`.
- `notifications/initialized`.
- `ping`.
- `tools/list`.
- `tools/call`.
- `notifications/cancelled`, which ends a pending `wait_for_decision` early.

Anything else gets JSON-RPC's `-32601`. It answers with the version the client asked for when that is one it implements (`2025-06-18` and `2025-03-26`, the revisions it is written against), and otherwise with its newest. That is the spec's negotiation rule. A newer revision is added only after someone has read it. Results are `content: [{type: "text"}]` with `isError` on a refusal. Stdout carries nothing but protocol, and every log line goes to stderr.

Why not a library: both maintained Go SDKs require Go 1.25, and this module is on 1.22. One local tool would move the whole server, CI and the Docker image to a new toolchain. It would also bring in JWT, OAuth2 and JSON-schema machinery the stdio path never uses. AGENTS.md §2 asks for no premature abstraction. Six messages over one pipe are less code to review than the dependency's changelog. If the module moves to Go 1.25 for its own reasons, swapping in `modelcontextprotocol/go-sdk` is one PR, because the transport is one file behind the tool handlers.

### 2. What the binary may import, and why the line is real

The binary holds the line the bot holds (ADR 0033 §3): it sees only what the wire gives its seat. Here that line is physical before it is a type rule. The binary runs on the owner's workstation and holds a WebSocket to a server elsewhere, so there is no `*game.Game` in its process to read, whatever it imports.

The type rule is kept anyway, so the code says the same thing. `TestMCPSeatImportsNoServerState` in `internal/mcpseat` fails on a **direct** import of `internal/game`, `internal/ws`, `internal/lobby`, `internal/actions`, `internal/db`, `internal/auth` or `internal/aiseat/tiers` from `internal/mcpseat` or `cmd/mcpseat`. That covers `Imports` and in-package `TestImports`. The end-to-end test (§10) is an external test package (`mcpseat_test`). It must stand a server up, so `XTestImports` is exempt, and the test says why. Like the `aiseat` gate, it has a control test that fails if the banned path stops matching anything.

Allowed: `internal/protocol` and `internal/legal` (the wire types: `LegalMoveView` is `legal.Move`), `internal/aiseat` (for `aiseat.Input` and `PassIndex` only), `internal/aiseat/rules` (§4), the extracted board renderer (§5) and `internal/util/redact`. As with the bot, the import of `aiseat` transitively links `game` into the binary. `protocol` already does that for `gamecli`. A linked package is not a handle on the table.

### 3. The tools

Ten tools. Every one returns text sized for a model's context. None returns the session token, a raw frame or an action payload. Moves are named by **index into a numbered list the binary showed**, and every index refers to one window. The binary never builds an action from anything else (variants in §6 come from the same enumeration).

| Tool | Input | Output |
|---|---|---|
| `join` | `invite_url` (string, required), `display_name` (string, default `"Agent"`), `deck` (optional: `{"id": "<pre-built deck id>"}` or `{"list": "<decklist text>"}`) | `game_id`, seat number, player id, table state, `resumed: true` when it reattached to a seat it already holds. On a refusal, the server's reason (wrong invite, full, started). |
| `set_deck` | `deck` as above | The deck name, or the server's validation problems. An unknown id answers with the ids this server has (the route's own 422). Before the game starts only. |
| `wait_for_decision` | `timeout_s` (1–50, default 25), `pass_until` (`"none"` default, or `"my_turn_or_stack"`, §4) | `status`: `decision`, `waiting` (the timeout passed with nothing to decide: call again), `not_started`, `eliminated` or `game_over` (with the outcome). With `decision`: `window` (an opaque token), the window's kind (`mulligan`, `priority`, `response`, `attack`, `block`, `choice:<kind>`), the compact view (§5), the numbered moves, and, since the last decision, the public log lines, chat lines and a one-line count of what the binary answered automatically. |
| `get_state` | `detail`: `compact` (default) or `full` | The view at that budget (§5). |
| `legal_moves` | none | The current window's numbered moves with kind, label, cost note and idle hint, and, per degraded card, its variants (§6). |
| `card` | `ref`: an instance id this seat can see, or a card name on the table | Name, type line, mana cost, power and toughness, oracle text and the `unimplemented` flag, from `GET /cards/{scryfall_id}`, cached per id. A face-down or hidden card answers with what the seat may see, which is nothing more than the view's own text. |
| `act` | `window` (required), `move` (index, required), `variant` (optional, §6) | `accepted`, `rejected` with the server's code and message, or `stale` with the new window. |
| `say` | `text` (1–500 characters) | `sent`, or `rate_limited` with when to retry. |
| `concede` | `confirm: true` (required) | `conceded`. |
| `leave` | none | Disconnects and deletes the saved session. While the game is active and the seat is not eliminated, it refuses and says to concede first, because a vanished seat holds the table. |

**`act` is checked locally before it is sent.** If `window` is not the current one, because a snapshot moved the board while the model thought, the binary refuses with `stale` and returns the new window rather than sending a move that might mean something else now. Pending-choice option lists can shrink while open (docs/protocol.md on `player` options), so an index from an old list is never reused. The binary sends the move's `{type, player, params}` as an `action` frame with a fresh `id`. It reports `rejected` if an `error` frame with that `id` arrives, and `accepted` once a snapshot advances `seq` with no such error. The protocol has no success acknowledgement, so an error that arrives after another seat's snapshot is reported on the next `wait_for_decision` and not lost. After three rejections in one window, `act` takes only the window's `always_legal` move. The binary still never picks it on the model's behalf.

**Concede is out of band, as it is for the bot.** `legal` does not enumerate concession (`aiseat.Conceder`'s doc comment). Conceding is every player's right (CR 104.3a), so it gets its own tool, with a `confirm` flag so it cannot be an accident of argument order.

**No sandbox verbs and no improvisation.** The agent cannot `move_card`, `change_life`, `force_cast` or `undo`. None of those are in `legal_moves`, and decision 1 gives it no more than a seat's closed list. When a card it cast is `unimplemented` and resolves to nothing, the compact view says so, and the agent's recourse is the one a guest has without the sandbox: say in chat what the card should have done, so a human can resolve it. ADR 0033 §8's improvisation stays a bot feature.

### 4. Layer A, imported and not re-implemented

The binary calls `rules.Resolve` on every window, building `aiseat.Input{View, Seat, Moves}` from the frame. One copy of the rules means the agent's trivial windows are answered exactly as the bot's are, which is what decision 4 asks. When `Resolve` takes the window, the binary sends that move at once, with no `MinThink` delay, the way a browser's autopass does. It counts the move for the next decision's summary.

**Which windows are trivial**, i.e. never shown to the model:

| Layer A rule | Absorbed for the agent | Why |
|---|---|---|
| `forced` — exactly one legal move | yes | Nothing to decide. Mostly the pass-only window. |
| `mana-only` — the only alternatives to passing are mana abilities | yes | Casts auto-tap, and floating mana empties at the step boundary (CR 106.4). |
| `coin-call` — heads or tails with no stop | yes | Both calls have the same odds. |
| `no-moves` | yes (nothing is sent) | Degenerate. |
| `same-land` — the only alternatives are copies of one land | **no, escalates** | It plays a land, which is an action, not a pass. Holding a land drop for after combat (landfall) or for a discard outlet is a decision, and the owner's list ("empty priority passes, forced moves, mana steps") did not include it. It costs about one model call per turn. |

One more rule comes from the runner, not from `Resolve`. **While `loop_notice` is up, nothing is passed automatically** (CR 732, ADR 0055): the notice is the server telling every client that a person has to ask for the next iteration, and here the agent is the person. **Combat declarations always escalate**: they carry no pass, so `Resolve` already sends them on unless exactly one move is offered.

**Layer A survives the wire cap, for the rules the agent uses.** The cap removes only alternatives within one `(source, kind, targets_stack)` tuple. It never removes a card, and never removes a kind from a card. A `forced` window has one move, and a `coin-call` window offers two answers, so neither is ever near 48. `mana-only` reads only which kinds are present, and the cap keeps every kind. So for those three, the verdict on a capped list is the verdict on the uncapped one. `same-land` is the exception: it compares land-move labels, and a capped list can drop the second face of one modal double-faced land, which would turn an escalation into an absorption. The agent does not absorb `same-land` (the table above), and the test notes the case for anyone who switches it back on. PR 4 pins the rest with a test that runs both lists through `Resolve` for every window of a recorded four-player game.

**The owner may tune it only toward the model.** `--absorb=forced,mana-only,coin-call` lists the absorbed rules. A rule can be removed so more windows reach the model. A rule cannot be added, and there is no flag that answers anything Layer A would not. `same-land` can be switched on with the same flag if one call per turn turns out not to be worth it.

**`pass_until` is the model's own autopass, off by default.** With `wait_for_decision(pass_until: "my_turn_or_stack")`, the binary also passes a priority window on another player's turn when the stack is empty, nothing is owed and the window is not a combat declaration. It stops at once for: anything on the stack, a pending choice, a combat declaration, the loop notice, and the start of the seat's own turn. The last one also clears it, as the client's safety belt does (ADR 0009 §7). This is the convenience a human gets from the browser's autopass toggle. It is the model's choice for its own seat, and it never answers a choice. A model holding an instant for an opponent's end step leaves it off.

### 5. The compact view

**The board half of `buildDelta` moves to `internal/aiseat/boardtext`**, a pure `Render(view *protocol.GameView, seat string, opts Options) string`. It renders the turn line, each seat (own first: life, hand and library counts, mana pool, end gates, emblems, battlefield, graveyard, own hand and command zone), the stack bottom-first with ownership and targets, and the seat's owed choices. `model.buildDelta` calls it and keeps the move half. That PR is a pure refactor and the prompt tests hold its bytes. The new package sits under `aiseat/`, so the existing import gate covers it from day one. The binary imports it, which makes the agent and the bot read the same board text. A fix to one is a fix to both.

Two things are added for the agent, both behind `Options` flags the bot leaves off: an `(unimplemented: the engine does not run this card's text)` note on a card whose `CardView.unimplemented` is set, and the `legal_actions` digest's degraded-card counts (§6).

**Budgets**, measured in UTF-8 bytes of the tool's text, at about 4 bytes per token:

| Output | Cap | How it is held |
|---|---|---|
| `wait_for_decision` and `get_state(compact)` | 6,000 bytes (~1,500 tokens) | `MaxZoneCards` 24, as the bot uses. Over the cap, graveyards drop to counts first, then opponents' battlefields to the 12 highest-power permanents with "and N more", then log and chat lines from oldest. Every move line stays: at most 48 on the wire, at about 60 bytes each. |
| `get_state(full)` | 24,000 bytes (~6,000 tokens) | No zone caps, the last 24 public log lines, and every revealed card. |
| `card` | 2,000 bytes | One card. |

The budget is enforced by the binary and tested (§10). A compact frame is about the size of the bot's own board block, without its decklist.

### 6. When the wire list is capped

The binary works out degradation per card. A card is degraded when `legal_actions.sources[id].moves` is greater than the number of `legal_moves` entries whose `source` is that card. Then:

- **Attacks.** The digest's `attack_targets` lists every player, planeswalker and battle the creature may attack. `act(window, move, variant: {"attack_target": "<id>"})` swaps the `target` of that card's kept `declare_attacker` move for one on the list.
- **Blocks.** The digest's `blocks` lists every attacker the creature may block. A `variant: {"block_attacker": "<id>"}` swaps the `attacker` of a kept single-blocker `declare_blocker` move.
- **One-target spells and abilities.** When the kept move's source has a single target clause with `min = max = 1`, `variant: {"target": "<id>"}` swaps that one target for another in the clause's `legal_targets` (`players` or `cards`).
- **Everything else** (several clauses, divided damage, X-scaled counts, modes) gets no variant. The compact view says "Card X: N further uses not listed", so the model knows the list is partial.

Each variant is checked locally against the enumerated set it came from before anything is sent, and the server validates every action as usual. The list stays closed: a variant is another member of a set the server enumerated for this seat, not a free-form action. Nobody has measured how often the 48 cap fires at a real table. The binary counts degraded windows and variants used, and reports the counts at game end (§8).

### 7. The badge: declared at join, one way, on the seat

**The agent declares itself in the join request.** No hello frame exists to carry a client kind, and decision 5 rules out a new endpoint. So both join routes (`POST /games/{id}/join` and `POST /join`) accept an additive body field:

```json
{ "invite_token": "…", "name": "Claude", "agent": { "client": "claude-code" } }
```

`client` is the MCP client's `clientInfo.name` from `initialize`, cut to 32 characters of `[a-z0-9._-]`, or `"unknown"`. The binary always sends `agent`. It has no flag to leave it out.

**The server records it on the seat, and it never comes off.** `game.Player` gains `Agent bool` and `AgentClient string`, set by `JoinAs` before the seat is visible to anyone. Nothing clears them: no route, setting, reclaim or restore. They ride the snapshot as `isAgent` and `agentClient` with `omitempty`. That is an additive change within schema 7, recorded with `-update-shape`, and copied in `clone.go`. `loadEntry` reads them back the way it reads `IsBot`, so **no database migration** is needed, and a rollback reads the snapshot and drops two unknown keys. `PlayerView` gains `is_agent` and `agent_client`, and so does `SeatInfo`, which is what the lobby list and the invite preview carry. All are public and identical for every viewer. Being an agent is a fact about the seat, like `is_bot`.

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

**Rate limits.** These are kept by the binary, because the WebSocket has none and the binary should not be the first client to need one:

- **Actions:** 4 per second, burst 8. Layer A's automatic answers count.
- **Rejections:** three per window (§3).
- **Chat:** one line per 5 s, 500 characters (the server allows 1,000).
- **Card lookups:** 5 per second, each id fetched once.
- **`join`:** shares the server's per-IP bucket. A 429 is retried after 1, 2, 4 s, then reported.

**Reconnects and reclaiming the seat.**

- **The ladder.** The binary follows the browser's ladder (docs/protocol.md "The reconnect ladder"): close 1000 is terminal, 1001 and 1006 redial with 500 ms to 30 s half-jittered backoff, and after three failed dials `GET /me` decides whether the session is dead.
- **A restart.** If the MCP client restarts the binary, `join` with the same invite finds the saved session and reattaches (`resumed: true`) instead of claiming a second seat. After the start it could not claim one anyway.
- **A dead session.** If the session has expired (12 h) or the server refuses it, the seat is lost to the binary, exactly as for any guest. An admin can mint a reclaim ticket, and `join` accepts that link (`/games/{id}/reclaim`'s ticket) in `invite_url`.

**While the agent thinks, the table waits, as it waits for a person.** There is no human turn timer, and decision 4 sends only real choices to the model, so the binary never answers one for it: no deadline, no heuristic fallback, no move played under the agent's name. The table sees the chip reading "thinking…" (§7) on the seat holding priority. If the MCP client stops calling tools (its turn ended, or the owner closed it), the seat sits like an idle browser tab until someone acts. This is the same position the table is in with a person who stepped away. The tool descriptions and docs/mcp-seat.md tell the client to loop `wait_for_decision` → `act` until `game_over`.

### 9. Game pace and cost

Figures are from ADR 0033 §5's measured funnel. Where a figure is an estimate, it says so. PR 5 measures them, and they are recorded here as an amendment.

| Quantity | Value |
|---|---|
| Windows per seat per game | ~650, measured (four heuristic bots) |
| Absorbed by Layer A | ~90%, measured |
| Reaching the model | **~60–70 per game**: ~60–65 measured, plus ~one `same-land` window per turn |
| With `pass_until` on | fewer; unmeasured |
| Tool calls per decision | 2–4 (`wait_for_decision`, `act`, sometimes `card` or `get_state`) |
| Context added per decision | ~2,000–2,500 tokens: a ≤1,500-token frame, the model's reasoning and the `act` round trip. Estimate |
| Context added per game | ~150k tokens. Estimate |
| Agent time per game | 15–40 s per decision gives **15–45 minutes** of the table waiting on the agent, on top of the game itself. Estimate |

**Tokens.** A long tool loop resends its whole context on each step. With prompt caching, most of that is cache reads, but it is still counted. Take 65 decisions, a ~25k-token client baseline, and ~2.5k added per decision. The model then reads about `65 × 25k + 2.5k × 65² / 2 ≈ 7M` input tokens per game, mostly cached, and writes about 40k. The context reaches ~175k by the end. That is near a 200k window, so expect one compaction per game, or use a long-context model. What that costs depends on the model and the plan: dollars on an API key, a share of the usage limit on a subscription. This ADR states no price, because none was measured. To keep it down: use a fresh session per game, use the compact view (the default), and turn `pass_until` on when holding nothing at instant speed.

**What the binary reports.** At `game_over` it logs one line to stderr and returns the same in the tool result: windows seen, absorbed by rule, shown to the model, `act` calls, rejections, degraded windows, variants used, bytes returned in tool results (an approximation of tokens), and the time from each decision opening to its `act`. The binary cannot see the model's own token count. The client reports that.

### 10. Tests

- **The end-to-end test.** `internal/mcpseat/e2e_test.go` (external package) stands up the lobby handler and WebSocket hub on `httptest.NewServer`, the same stand-up `lobby/http_test.go` uses. It creates a two-seat table with a `random` bot on the whole-game tests' vanilla decks, then drives the seat through its JSON-RPC interface over in-memory pipes. The test, not a model, plays the agent's side: `initialize`, `tools/list`, `join`, `set_deck`, then a scripted chooser that loops `wait_for_decision` → `act`, picking the pass or move 0 and sometimes a variant, until `game_over` or a turn bound.
- **What it asserts.** Every window the test is shown escalated under §4's table. The badge is on `is_agent` in the bot's view of the table and in `GET /games`. No tool result exceeds its budget. The sentinel token never appears in a tool result or log line. A stale `window` is refused.
- **Waits.** The test polls monotone predicates through one `waitFor` helper (AGENTS.md §5, #634) and has no `select` on `time.After`.
- **Cost.** It runs in every CI run. It is sized to finish in seconds and makes **no network model call**, because there is no model in it.
- **Unit tests.** The JSON-RPC subset (malformed lines, unknown methods, version negotiation, cancellation of a pending wait); the import gate and its control; §4's cap invariance over a recorded game; §6's degradation detection and variant checks against fixtures cut from `legal_moves_test.go`'s capped boards; the state file's permission refusal; the redaction of every log path.
- **Server side.** The join field and its refusal with a signed-in session; the badge's persistence across `CaptureSnapshot` → restore → `loadEntry`; the snapshot shape file; `PlayerView` and `SeatInfo` carrying it for every viewer; the host hand-over skipping an agent; the Discord link refusing one.
- **Client side.** A vitest for the chip, its "thinking…" state, and the bot chip and agent chip never both showing.

### 11. Docs

A new `docs/mcp-seat.md`, linked from `docs/bot.md`'s introduction and from AGENTS.md §5, rather than a section of `docs/bot.md`. The bot guide is about seats the server runs. This one is about a seat the owner runs, so it carries different warnings. It covers: building the binary; the tools; the flags; the badge and what it does and does not prove; the token file; the untrusted-text warning; the pace and cost of §9; and these recipes.

**Claude Code** (syntax checked against 2.1.289's `--help`):

```sh
claude mcp add -s user cmdctrl-seat -- \
  /home/luke/repos/cmd_and_ctrl/server/bin/cmd_and_ctrl-mcpseat \
  --allow-origin https://cmd-dev.labxp.io
```

To play, start a session in an empty directory that can use only the seat's tools, for example `claude --allowedTools "mcp__cmdctrl-seat__*"`. Then give it the invite link and "play this game to the end: loop wait_for_decision and act". **Unverified:** the exact tool-name pattern Claude Code gives an MCP server whose name has a hyphen (check `/mcp`); which `--permission-mode` denies, rather than asks about, everything else; how long Claude Code waits on one tool call (the `MCP_TOOL_TIMEOUT` environment variable). That last one is why `wait_for_decision` defaults to 25 s and stops at 50.

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

To play, run `codex -s read-only` in an empty directory. **Unverified:** the `[mcp_servers.<name>]` table shape beyond what `codex mcp add` writes (check the file it produces), and the per-server tool timeout key (`tool_timeout_sec` is believed to exist, default 60 s).

AGENTS.md §3 gains `cmd/mcpseat/` and `internal/mcpseat/` in the layout. §5 gains a short "MCP seat" entry: the make target, the flags, and the pointer to docs/mcp-seat.md. docs/protocol.md gains `PlayerView.is_agent` / `agent_client`, and docs/lobby.md the `agent` join field and its 400.

---

## Delivery

Each PR goes into `develop`, Sprint S62, Issue #2230.

| PR | What | Needs | Parallel with |
|---|---|---|---|
| 1 | **This ADR**, the AGENTS.md §3 ADR range line, and the S62 section in `docs/sprints.md`. Docs only. | — | anything |
| 2 | **Server: the badge** (§7). The `agent` join field on both join routes and its 400 with a signed-in session; `game.Player.Agent` / `AgentClient`, the snapshot keys and shape file, `clone.go`, `loadEntry`; `PlayerView` and `SeatInfo`; the host hand-over and Discord-link refusals; docs/protocol.md and docs/lobby.md; the server tests in §10. | 1 | 4 |
| 3 | **Client: the chip** (§7) on the seat, lobby list, invite preview and chat, with its vitest. | 2 | 4, 5 |
| 4 | **Refactor: `aiseat/boardtext`** (§5). Byte-identical bot prompts, held by the existing prompt tests, plus the cap-invariance test for Layer A (§4). | 1 | 2, 3 |
| 5 | **The binary and its tools** (§1–§6, §8, §9). `internal/mcpseat`, `cmd/mcpseat`, `make build-mcpseat`, the import gate, the end-to-end test and the unit tests (§10). | 2, 4 | 3 |
| 6 | **Docs** (§11). `docs/mcp-seat.md`, the AGENTS.md §3 layout and §5 entries, the pointer from docs/bot.md. Then one real game on cmd-dev, with its §9 numbers appended to this ADR as an amendment. | 5 | — |

After PR 6: the owner plays one game on cmd-dev with Claude Code in a seat. The other seats should see the chip, and the game should finish. The numbers go into this ADR, and #2230 is closed with that evidence.

## Consequences

- The owner can seat Claude Code, Codex or any MCP client at a real table as a guest, with the same view and moves a person in that seat has, minus the sandbox.
- Every agent seat is marked, permanently and publicly. The mark relies on a cooperating client, and the docs say so.
- The server gains two snapshot keys, two wire fields and one join field. It gains no endpoint, no frame kind and no migration.
- The bot's board text moves to a package of its own, so the two seats read the same board.
- A table with an agent is slower than a table of bots, by roughly the agent's thinking time on 60–70 decisions.
- Layer A passes instantly, so an agent seat that holds priority for more than a moment has a real choice in front of it. Opponents can read that, as they can read a human's smart autopass. #1307's "considering…" chip hides the same tell for humans only partly. Giving the agent a uniform delay would hide it better and slow every trivial window; this ADR does not (call 8).
- The server module stays on Go 1.22. The binary carries a small MCP transport of its own, which has to be kept up with the spec's stdio revisions by hand.

## Out of scope

- **A hosted or remote MCP server** (HTTP transport). Decision 5 is a local stdio binary.
- **Agent seats run by the server.** That is the bot (ADR 0033), and the bot keeps its tiers.
- **Improvisation and sandbox verbs for an agent** (§3).
- **A turn timer** for any seat, agent or human.
- **Several agent seats from one binary.** One binary process holds one seat. Two seats are two MCP server entries.
- **Agent-to-agent private channels.** Chat is the table's, and agents use it like anyone.
- **Raising `legalMovesWireCap`, or per-seat wire caps.** §6 works within the cap. If PR 6's count shows variants are not enough, that is its own change.

## Calls made here

The owner's answers did not settle these. Each is decided above, and each can be overturned in review:

1. **Hand-rolled stdio JSON-RPC, no MCP library** (§1). Both Go SDKs need Go 1.25, and the module is on 1.22.
2. **`internal/mcpseat` plus a thin `cmd/mcpseat`, built by `make build-mcpseat`, not deployed** (§1).
3. **The import gate bans server-state packages from the binary and exempts only its external end-to-end test** (§2).
4. **`same-land` is not absorbed for the agent** (§4). Playing a land is an action, and its timing can matter. `--absorb` can turn it back on.
5. **The owner may only remove Layer A rules, never add** (§4). Nothing answers a window Layer A would not.
6. **The loop notice stops every automatic pass** (§4, CR 732), as it does for the bot.
7. **`pass_until` exists, off by default, chosen by the model per call, and cleared at the seat's own turn** (§4).
8. **Layer A answers immediately, with no `MinThink`** (§4), as a browser's autopass does. The cost is a timing tell (Consequences).
9. **The board renderer is extracted to `aiseat/boardtext` and shared with the bot** (§5), with an `unimplemented` note added for the agent.
10. **Budgets of 6,000 bytes compact, 24,000 full and 2,000 per card, as plain text only** (§5). No structured duplicate of the same data.
11. **Oracle text on demand through `card` and the existing `GET /cards/{id}`** (§3), not inlined in every frame.
12. **Degraded lists get variants for attacks, blocks and one-target moves, from the digest and `legal_targets`, and nothing else** (§6).
13. **Every `act` names its window, and a stale one is refused locally** (§3).
14. **After three rejections in a window, only the always-legal move is accepted from the model** (§3). The binary never plays it unasked.
15. **`concede` needs `confirm: true`. `leave` refuses on a live seat** (§3).
16. **No sandbox verbs and no improvisation for the agent** (§3).
17. **The badge is declared in the join body as `agent: {client}`, from the MCP `clientInfo.name`, with no off switch in the binary** (§7).
18. **The badge lives on `game.Player` and the snapshot, with no database column** (§7), as `IsBot`'s companions do.
19. **An agent join with a signed-in session is refused** (§7), so decision 2 is enforced by the server.
20. **An agent seat is never host and cannot link Discord** (§7).
21. **The token file's place, its permissions, the Bearer header, and no token in any tool result or log line** (§8).
22. **`--allow-origin`, set by the owner, defaults to the two labxp hosts and localhost** (§8).
23. **Table text is wrapped, cut and labelled as untrusted** (§8). The docs recommend a session that can do nothing else.
24. **Client-side rate limits: 4 actions/s, 1 chat line per 5 s, 5 card lookups/s** (§8).
25. **No deadline on the agent's decisions** (§8). The table waits as it does for a person, and the binary never plays a move the model did not choose.
26. **`join` reattaches a saved seat and accepts a reclaim link** (§3, §8).
27. **The docs go in a new `docs/mcp-seat.md`, not in `docs/bot.md`** (§11).
