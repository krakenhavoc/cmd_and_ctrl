# server

The `cmd_and_ctrl` game server. Authoritative in-memory game state, WebSocket
protocol to clients, written in Go. See the [top-level PLAN.md](../PLAN.md)
for the project vision and [AGENTS.md](../AGENTS.md) for working conventions.

## Running

```bash
# dev loop (no binary)
make dev

# build and run
make build
./bin/cmd_and_ctrl-server

# run tests
make test

# fmt-check + vet + test + build
make
```

## Environment variables

| Name | Default | Effect |
|---|---|---|
| `CMDCTRL_ADDR` | `:8080` | Listen address for the HTTP + WebSocket server. |
| `CMDCTRL_DATA_DIR` | `./data` | Root directory for crash-recovery snapshots. Set to empty (`CMDCTRL_DATA_DIR=""`) to disable disk writes entirely. |

Example:

```bash
CMDCTRL_ADDR=:9090 CMDCTRL_DATA_DIR=/var/lib/cmdctrl make dev
```

## Endpoints

| Method | Path      | Description |
|--------|-----------|-------------|
| GET    | /healthz  | Liveness probe. Returns `200 ok`. |
| GET    | /ws       | WebSocket endpoint speaking [protocol v0](../docs/protocol.md). |

On startup the server seeds a single 4-player demo game with synthetic
99-card decks. Player IDs are logged as structured JSON so the
[gamecli](./cmd/gamecli/) operator can read them from stdout. S04
replaces this demo game with a proper lobby.

## Driving the demo game with gamecli

```bash
# Start the server in one terminal:
make dev

# In another terminal, run a scripted turn:
go run ./cmd/gamecli \
  -addr ws://localhost:8080/ws \
  -script path/to/turn.json
```

Script format is one JSON object with an `actions` array; each entry
maps to a protocol v0 action frame. See
[docs/protocol.md](../docs/protocol.md) for the full action catalog.
Example:

```json
{
  "actions": [
    {"type": "draw_card", "player": "<player-uuid>"},
    {"type": "change_life", "player": "<player-uuid>", "params": {"delta": -5}},
    {"type": "pass_turn"}
  ]
}
```

`gamecli` can also read actions from stdin (one JSON object per line)
for interactive poking.

## Layout

```
server/
├── cmd/
│   ├── server/           # main: HTTP + WebSocket server with demo room
│   │   └── main.go
│   └── gamecli/          # dev WebSocket client for driving a room
│       └── main.go
├── internal/
│   ├── game/             # authoritative in-memory domain
│   │   ├── game.go       #   Game struct, lifecycle, Turn cursor
│   │   ├── mutations.go  #   13 S03 action methods (DrawCard, PlayCard, ...)
│   │   ├── player.go     #   Player, life, commander damage
│   │   ├── zone.go       #   Zone, ZoneKind, MoveCard
│   │   ├── card.go       #   Card instance
│   │   ├── turn.go       #   Phase/Step enums + turn advance machine
│   │   └── errors.go     #   sentinel errors
│   ├── protocol/         # v0 wire format types
│   │   ├── protocol.go   #   Frame, Kind, payloads
│   │   └── view.go       #   GameView projection + ViewOfGame builder
│   ├── actions/          # wire action types + Dispatch(game, action) router
│   │   └── actions.go
│   └── ws/               # gorilla/websocket hub, Room, crash recovery
│       ├── hub.go        #   Hub + Client + handleAction + classifyActionError
│       ├── room.go       #   Room.Apply serialises mutate + seq + capture
│       └── e2e_test.go   #   scripted-turn golden test
├── Makefile
├── .golangci.yml
└── go.mod
```

The `internal/` boundary is enforced by the Go compiler — nothing outside
`server/` can import these packages. That is intentional. Shared types go
in a future `pkg/` subtree only when they genuinely need to be shared.

## Crash recovery

After every successful action, `Room.Apply` writes the latest full
snapshot to `$CMDCTRL_DATA_DIR/games/<game-id>.json` via an atomic
`os.CreateTemp + os.Rename` sequence, so a crash mid-write never
leaves a partial file on disk. S03 writes dumps but does not reload
them on startup — that's S04 scope. Disable dumps entirely with
`CMDCTRL_DATA_DIR=""`.
