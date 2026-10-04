module github.com/krakenhavoc/cmd_and_ctrl/server

// Go 1.27, the newest supported release (ADR 0122 §1.1, owner decision 8):
// Go supports only its two newest majors, and the MCP seat's SDK needs 1.25
// or later. We also rely on method-aware ServeMux (1.22+), log/slog (1.21+)
// and `for range N` (1.22+). CI, scripts/go-docker.sh and the devcontainer
// all pin the same version; move them together.
go 1.27.0

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	modernc.org/sqlite v1.36.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/exp v0.0.0-20230315142452-642cacee5cc0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	modernc.org/libc v1.61.13 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.2 // indirect
)
