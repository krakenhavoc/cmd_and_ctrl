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
	github.com/modelcontextprotocol/go-sdk v1.8.0
	github.com/prometheus/client_golang v1.24.1
	modernc.org/sqlite v1.36.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/exp v0.0.0-20230315142452-642cacee5cc0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	modernc.org/libc v1.61.13 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.2 // indirect
)
