// Command cmd_and_ctrl-bot is the S12.5 Discord slash-command bot.
// It runs as a second systemd unit next to the game server and
// exposes /cc-invite + /cc-games against an allow-listed set of
// guilds. Invite posting is done by calling back into the game
// server's admin HTTP API over loopback — the bot never mutates
// game state directly.
//
// Configuration is via environment variables. See
// server/internal/bot/config.go for the full list; all CMDCTRL_*
// names live there.
//
// This is PR 1 of the S12.5 bot slice: scaffolding only. The
// gateway connection and command handlers land in PR 3.
package main

import (
	"log/slog"
	"os"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bot"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg := bot.ConfigFromEnv()

	// Empty bot token is the canonical "feature off" signal. A
	// dev stack without a Discord app should still start this
	// binary in a smoke-test context and have it exit cleanly.
	if cfg.Disabled() {
		log.Info("bot disabled (CMDCTRL_DISCORD_BOT_TOKEN unset); exiting")
		return
	}

	if err := cfg.Validate(); err != nil {
		log.Error("bot config invalid", "error", err.Error())
		os.Exit(1)
	}

	log.Info("bot config loaded", "config", cfg.Redacted())

	// PR 3 will replace this with: open discordgo session,
	// register commands per allowed guild on ready, block on
	// SIGTERM. For PR 1 the binary prints its config and exits
	// so operators can smoke-test env wiring without a live
	// Discord app.
	log.Info("scaffolding exit (PR 1); gateway connection ships in PR 3")
}
