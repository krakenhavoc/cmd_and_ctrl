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

# vet + test + build
make
```

Default listen address is `:8080`. Override with `CMDCTRL_ADDR`:

```bash
CMDCTRL_ADDR=:9090 make dev
```

## Endpoints (S01)

| Method | Path      | Description |
|--------|-----------|-------------|
| GET    | /healthz  | Liveness probe. Returns `200 ok`. |
| GET    | /ws       | WebSocket endpoint speaking [protocol v0](../docs/protocol.md). |

## Layout

```
server/
├── cmd/
│   └── server/      # main package, HTTP server wiring
│       └── main.go
├── internal/
│   ├── protocol/    # wire format types (spec in docs/protocol.md)
│   └── ws/          # WebSocket hub and client handling
├── Makefile
├── .golangci.yml
└── go.mod
```

The `internal/` boundary is enforced by the Go compiler — nothing outside
`server/` can import these packages. That is intentional. Shared types go
in a future `pkg/` subtree only when they genuinely need to be shared.
