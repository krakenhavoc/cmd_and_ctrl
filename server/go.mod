module github.com/krakenhavoc/cmd_and_ctrl/server

// Minimum Go version is 1.22 — we rely on method-aware ServeMux (1.22+),
// log/slog (1.21+), and `for range N` (1.22+). Devcontainer ships the
// latest Go, but older collaborator installs stay supported.
go 1.22

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
)
