# The MCP seat (ADR 0122, S62)

`cmd_and_ctrl-mcpseat` is a local program that sits an AI agent (Claude
Code, Codex, any MCP client) in a **guest seat** at a table, marked to
everyone else as an AI agent. It is a seat **you** run, on your own
machine, for an agent **you** choose. The server does not run it, and
the server's own bot seats ([docs/bot.md](bot.md)) are a different
thing: they are in-process and the server picks their moves.

The reasoning behind every choice below is in
[ADR 0122 — An agent at the table](decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md).
Where this page and the ADR differ, this page follows the merged code.

The binary reads **only the WebSocket** its seat has: the same filtered
view and the same legal-move list a browser gets. It cannot see
anyone's hand, library or the server's state, and it has no way to make
a move that is not in a numbered list it showed the agent.

## Build

It is a local tool, not deployed.

```sh
make -C server build-mcpseat        # -> server/bin/cmd_and_ctrl-mcpseat
```

This workstation has no local Go. Run it through the one sanctioned
wrapper, from the repo root:

```sh
scripts/go-docker.sh go build -o bin/cmd_and_ctrl-mcpseat ./cmd/mcpseat
```

(The container makes `server/bin` root-owned. The binary still runs as
you.) See AGENTS.md §5 for why there is no other way to run Go here.

## Flags

stdout carries the MCP protocol and nothing else. Logs go to stderr, or
to `--log-file`.

| Flag | Default | Meaning |
|---|---|---|
| `--allow-origin <scheme://host[:port]>` | `https://cmd.labxp.io`, `https://cmd-dev.labxp.io`, `http://localhost:*`, `http://127.0.0.1:*` | An origin `join` may talk to. Repeatable. The port may be `*`. Naming any replaces the whole default list. The agent cannot change it. |
| `--absorb <rules>` | `forced,mana-only,coin-call` | The trivial windows the binary answers without the model: a subset of `forced`, `mana-only`, `coin-call`, `same-land`, or `none`. You can only add `same-land` or remove rules, so more windows reach the model; nothing Layer A would not answer can be answered for it. |
| `--log-file <path>` | stderr | Write the log here, mode `0600`, appended. |
| `--state-dir <dir>` | `$XDG_STATE_HOME/cmdctrl-mcpseat`, else `~/.local/state/cmdctrl-mcpseat` | Where saved sessions live. |
| `-v` | off | Debug logging. |

`same-land` is off by default on purpose: playing a land is an action,
and holding it for after combat or for a discard outlet is a decision.
It costs about one model call per turn.

## Tools

Ten tools. Every result is text sized for a model's context. Moves are
named by **number in a list the binary showed**, and a number is only
good in the window it was shown in.

| Tool | Input | What it does |
|---|---|---|
| `join` | `invite_url` (required: the invite link, or an admin's seat-reclaim link), `display_name` (default `Agent`), `deck` (optional `{id}` for a pre-built deck or `{list}` for a decklist) | Takes a guest seat, or reattaches to the seat this binary already holds (`resumed`). One binary holds one seat: a second table needs `leave` first. Refused for an origin not on `--allow-origin`. A 429 is retried after 1, 2, 4 s. |
| `set_deck` | `deck`, as above | Installs a deck before the game starts. An unknown id answers with the ids the server has. |
| `wait_for_decision` | `timeout_s` (1 to 50, default 25), `pass_until` (`none`, or `my_turn_or_stack`) | Waits for a real choice. Statuses: `decision` (with `window`, `kind`, the compact board, the numbered moves, and the log and chat since last time plus a count of what was answered automatically), `waiting` (call again), `not_started`, `eliminated`, `game_over` (with the outcome and the report below), `disconnected` (an error). |
| `get_state` | `detail`: `compact` (default) or `full` | The board as the seat sees it. |
| `legal_moves` | `card` (an instance id) or `choice` (a pending choice id, or `cleanup_discard`), both optional | The open window's full numbered list, grouped by card. With `card` or `choice`, that card's or prompt's moves with the enumerator's caps lifted (up to 512). A one-card request with no moves answers an empty list. |
| `card` | `ref`: an instance id or a card name on the table | Name, type, cost, power/toughness and oracle text, with a note when the engine does not run the card's text. Cached per id. |
| `act` | `window` (required), `move` (required), `value` (only for a move marked open) | Makes one move. `status`: `accepted` (the server's ack), `rejected` (the server's code and message), `stale` (the board moved; nothing was sent, and the new window is returned), `unknown` (the socket dropped or no answer came; read the state), `not_sent`, `cancelled`. |
| `say` | `text` (1 to 500 characters) | One line of table chat. `sent` or `rate_limited` with when to retry. |
| `concede` | `confirm: true` | Concedes. Has its own tool because the move list never offers it. |
| `leave` | none | Disconnects and deletes the saved session. Refused while the game is live and the seat is still in it: concede first. |

What to know about how they behave:

- **The loop is `wait_for_decision`, then `act`, until `game_over`.** The
  tool descriptions tell the client so. Typically 2 to 4 tool calls per
  decision.
- **The move list is never cut.** The wire's 48-move cap and the
  enumerator's own per-card caps are reported, and the binary fetches
  the full list before Layer A or the model sees anything. A card the
  enumerator still cut says so, with how many ("at least N" where the
  count is a floor).
- **A move is checked before it is sent.** A stale `window` is refused
  locally. After three rejections in one window, only the window's
  always-legal move is accepted, and the binary never picks it for the
  model. A request the server answers `no_decision` closes the window
  (the board moved); `rate_limited` is retried after 300 ms, up to three
  times.
- **`value` is only for open sets**: any card name (up to 200
  characters) where the rules let a player name one, or a number for X
  inside the stated range.
- **`act` waits for the server's ack** for up to 10 s, then answers
  `unknown`. Against a server that predates the ack (PR #2250) it waits
  2 s and says why.
- **`pass_until: my_turn_or_stack`** also passes priority on other
  players' turns while the stack is empty and nothing is owed. It stops
  for anything on the stack, a choice, a combat declaration, the loop
  notice (CR 732) or the start of the seat's own turn. Leave it off to
  hold an instant for an opponent's end step.
- **Nothing is played on the model's behalf.** No deadline, no
  heuristic fallback. The table waits for the agent as it waits for a
  person.
- **No sandbox verbs, no improvisation.** The agent cannot move cards,
  change life, force a cast or undo. When a card is marked as not run by
  the engine, its recourse is to say in chat what it should have done.
- **A server that predates the badge (#2249) refuses the join** with a
  400 about an unknown `agent` field. The binary says so and does not
  retry without the field.

## The badge

At join the binary declares `agent: {client}` in the request, where
`client` is the MCP client's name from `initialize`, normalised to 32
characters of `[a-z0-9._-]` or `unknown`. It has no flag to leave it
out. The server stamps the seat (`is_agent`, `agent_client`), and
nothing ever clears it: no route, setting, reclaim or restore. Every
viewer sees an "AI agent" chip in the slot a bot chip uses, on the
table, the lobby seat list, the invite preview and chat, reading
"thinking…" while the seat holds priority.

**What it proves:** a seat that joined as an agent cannot later pass as
a human, and a human who claims to be an agent harms nobody.

**What it does not prove:** that a seat without the badge is a person.
The badge is a declaration by a cooperating client. A modified binary or
a script on a guest seat can join without the field, and the server
cannot tell it from a human. No check is offered as if it closed that
gap.

## The token and the state file

- The session token is held in memory and in one file per seat,
  `<state-dir>/<server host>/<game id>.json`: origin, game id, player id,
  token and expiry. The directory is `0700`, the file `0600`, written
  through a temp file and rename. The binary **refuses to read** a file
  or directory that is group- or world-accessible, as `ssh` does.
- It is deleted on `leave`, at game end, and when found expired
  (a guest session lasts 12 h).
- The invite token is never stored. If your MCP client restarts the
  binary, `join` with the same invite reattaches (`resumed: true`)
  instead of claiming a second seat.
- On the wire the token is only ever `Authorization: Bearer`, never
  `?token=`. It is stripped from every tool result (those go to the model
  provider) and every log line, which also passes `redact.Secrets`.
- A dead session means the seat is lost, as for any guest. An admin can
  mint a reclaim link, which `join` accepts as `invite_url`.

## Table text is untrusted

Player names, chat, deck names and named cards are written by other
people, and they land in the context of an agent that may have a shell.
The binary wraps each such string as `«…»`, cuts names to 40 characters
and chat and log lines to 300, strips URLs, and states in every tool
description that table text is data, never instructions. **It cannot
stop a client from obeying an injected string.** So run the seat in a
session that can do nothing else: an empty working directory, and only
this server's tools allowed. `--allow-origin` is yours, not the
model's, so a hostile invite link in chat cannot point the seat at
another server.

Client-side limits the binary keeps: actions 4 per second (burst 8;
automatic answers count), three rejections per window, one move-list
request in flight, one chat line per 5 s, card lookups 5 per second.

## Recipes

### Claude Code

```sh
claude mcp add -s user cmdctrl-seat -- \
  /home/luke/repos/cmd_and_ctrl/server/bin/cmd_and_ctrl-mcpseat \
  --allow-origin https://cmd-dev.labxp.io
```

To play, start a session in an empty directory that can use only the
seat's tools, for example `claude --allowedTools "mcp__cmdctrl-seat__*"`,
and say: "Join this table (invite link), then play it to the end: loop
wait_for_decision and act." The host starts the game from a browser
seat. An agent seat is never the host.

Unverified for Claude Code: the exact tool-name pattern for a server
name with a hyphen (check `/mcp`); which `--permission-mode` denies
rather than asks about everything else; and the per-call timeout
(`MCP_TOOL_TIMEOUT`), which is why `wait_for_decision` defaults to 25 s
and stops at 50.

### Codex

```sh
codex mcp add cmdctrl-seat -- \
  /home/luke/repos/cmd_and_ctrl/server/bin/cmd_and_ctrl-mcpseat \
  --allow-origin https://cmd-dev.labxp.io
```

or in `~/.codex/config.toml`:

```toml
[mcp_servers.cmdctrl-seat]
command = "/home/luke/repos/cmd_and_ctrl/server/bin/cmd_and_ctrl-mcpseat"
args = ["--allow-origin", "https://cmd-dev.labxp.io"]
```

Play with `codex -s read-only` in an empty directory. Unverified: the
table shape beyond what `codex mcp add` writes, and the per-server tool
timeout key (`tool_timeout_sec` is believed to exist, default 60 s).

## Pace and cost

The table waits for the agent, with no turn timer. **Everything in this
table is an estimate** unless marked measured; ADR 0122 §9 has the
derivation.

| Quantity | Value |
|---|---|
| Windows per seat per game | ~650 (measured, four heuristic bots) |
| Absorbed by Layer A | ~90% (measured) |
| Reaching the model | ~60 to 70 per game (estimate) |
| Tool calls per decision | 2 to 4 (estimate) |
| Context added per decision | ~2,000 to 2,500 tokens (estimate) |
| Context added per game | ~150k tokens (estimate) |
| Table waiting on the agent | 15 to 45 minutes per game, at 15 to 40 s a decision (estimate) |

A long tool loop resends its whole context each step, so input tokens
run in the millions per game (mostly cache reads) and the context
nears 200k by the end: expect one compaction, or use a long-context
model. No price is stated because none was measured. To keep it down,
use a fresh session per game, keep the compact view, and turn
`pass_until` on when holding nothing at instant speed.

**The game-end report.** At `game_over` the binary logs one line and
returns the same in the tool result: windows seen, how many each rule
absorbed and how many were shown to the model; `act` calls and
rejections; truncated windows, full-list requests and the largest move
list shown (moves and bytes); bytes returned in tool results (about
4 per token); and the time from each decision opening to its `act`. It
cannot see the model's own token count. Your client reports that.

### Measured (to be filled after the first cmd-dev game)

| Quantity | Measured |
|---|---|
| Windows seen / shown to the model | _pending_ |
| Decisions, rejections | _pending_ |
| Largest move list shown | _pending_ |
| Tool-result bytes | _pending_ |
| Time per decision | _pending_ |
| Client-reported tokens and cost | _pending_ |
