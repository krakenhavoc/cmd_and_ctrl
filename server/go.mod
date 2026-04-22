module github.com/krakenhavoc/cmd_and_ctrl/server

// Minimum Go version is 1.22 — we rely on method-aware ServeMux (1.22+),
// log/slog (1.21+), and `for range N` (1.22+). Devcontainer ships the
// latest Go, but older collaborator installs stay supported.
go 1.22

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
)

require (
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
)
