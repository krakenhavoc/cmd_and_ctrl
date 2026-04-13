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
| `CMDCTRL_DATA_DIR` | `./data` | Root directory for crash-recovery snapshots, the Scryfall dump, and the image cache. Set to empty (`CMDCTRL_DATA_DIR=""`) to disable disk writes entirely. |
| `CMDCTRL_ADMIN_TOKEN` | *(required)* | Shared admin secret for `POST /admin/login`. Must be at least 16 characters. |
| `CMDCTRL_SESSION_TTL` | `12h` | Session lifetime; any valid Go duration. |
| `CMDCTRL_SEED_DEMO` | unset | If `1`, seed the S03 4-player demo game at startup for the gamecli dev loop. |

Example:

```bash
CMDCTRL_ADMIN_TOKEN=$(head -c 24 /dev/urandom | base64) \
CMDCTRL_ADDR=:9090 CMDCTRL_DATA_DIR=/var/lib/cmdctrl \
make dev
```

## Endpoints

| Method | Path                      | Description |
|--------|---------------------------|-------------|
| GET    | /healthz                  | Liveness probe. Returns `200 ok`. |
| GET    | /ws                       | WebSocket endpoint speaking [protocol v0](../docs/protocol.md). Requires session (`?token=` or cookie) and `?game=<uuid>`. |
| POST   | /admin/login              | Exchange the admin token for an admin session. |
| POST   | /games                    | *(admin)* Create a new game. |
| GET    | /games                    | List games. |
| GET    | /games/{id}               | Game metadata. |
| POST   | /games/{id}/join          | *(public, invite-gated)* Claim a seat; returns a RolePlayer session. |
| POST   | /games/{id}/start         | Transition lobby → active. |
| GET    | /me                       | Echo the authenticated principal. |
| GET    | /cards/{id}               | Card metadata from the Scryfall index. |
| GET    | /cards/{id}/image         | Serve the card image (downloads + caches on miss). |

See [docs/lobby.md](../docs/lobby.md) for the lobby HTTP reference and
[ADR 0003](../docs/decisions/0003-auth-and-lobby.md) for the auth model.

On startup the server runs with an empty lobby; admins log in with
`CMDCTRL_ADMIN_TOKEN`, create games, and share invite URLs. Setting
`CMDCTRL_SEED_DEMO=1` additionally seeds a 4-player demo game so the
[gamecli](./cmd/gamecli/) can be driven without going through the
lobby flow.

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
│   │   └── view.go       #   GameView + ViewOfGame + FilterViewFor (per-viewer)
│   ├── actions/          # wire action types + Dispatch(game, action) router
│   │   └── actions.go
│   ├── ws/               # gorilla/websocket hub, Room, RoomManager, per-viewer broadcast
│   │   ├── hub.go        #   Hub + Client + UpgradeAuthorizer + broadcastToRoom
│   │   ├── room.go       #   Room.Apply returns GameView; serialises mutate+seq+capture+dump
│   │   ├── manager.go    #   RoomManager keyed by game UUID
│   │   └── e2e_test.go   #   scripted-turn golden test
│   ├── auth/             # pluggable Authenticator + MemoryAuthenticator + HTTP middleware
│   │   ├── auth.go       #   Authenticator interface, Principal, Role
│   │   ├── memory.go     #   S04 default: in-memory invite store
│   │   └── http.go       #   CredentialFromRequest, Middleware, ctx helpers
│   ├── lobby/            # GameMeta + invite flow + lobby HTTP handler
│   │   ├── lobby.go      #   Lobby registry over RoomManager
│   │   ├── http.go       #   /admin/login, /games*, /me routes
│   │   └── ws_authorizer.go # lobby.WSAuthorizer — session → (gameID, playerID) for the hub
│   └── cards/            # Scryfall index + disk-backed image cache + /cards routes
│       ├── index.go      #   streaming loader for the default_cards bulk dump
│       ├── images.go     #   on-demand fetch, sharded disk layout, per-id dedup
│       └── http.go       #   GET /cards/{id} + GET /cards/{id}/image
├── Makefile
├── .golangci.yml
└── go.mod
```

The `internal/` boundary is enforced by the Go compiler — nothing outside
`server/` can import these packages. That is intentional. Shared types go
in a future `pkg/` subtree only when they genuinely need to be shared.

## Crash recovery

After every successful action, `Room.Apply` writes the latest full
snapshot (unfiltered — crash-recovery carries full fidelity) to
`$CMDCTRL_DATA_DIR/games/<game-id>.json` via an atomic
`os.CreateTemp + os.Rename` sequence, so a crash mid-write never
leaves a partial file on disk. Restart-time reload is deferred to a
future sprint. Disable dumps entirely with `CMDCTRL_DATA_DIR=""`.

## Card data (S04)

The Scryfall default-cards bulk dump lives at
`$CMDCTRL_DATA_DIR/scryfall/default-cards.json`. Refresh with:

```bash
./scripts/scryfall-refresh.sh
```

…on a weekly cron. Images are cached on disk under
`$CMDCTRL_DATA_DIR/images/<aa>/<id>.<size>.jpg` and served via the
`/cards/{id}/image` route. The server is bootable without the dump —
card routes 404 until the first refresh lands.
