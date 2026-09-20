// Command cmd_and_ctrl-bot is the S12.5 Discord slash-command bot.
// It runs as a second systemd unit next to the game server and
// exposes /cc-invite + /cc-games + /cc-end against an allow-listed
// set of guilds. Every command works by calling back into the game
// server's admin HTTP API over loopback — the bot never mutates
// game state directly.
//
// Configuration is via environment variables. See
// server/internal/bot/config.go for the full list; all CMDCTRL_*
// names live there.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

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

	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Error("discordgo.New", "error", err.Error())
		os.Exit(1)
	}

	// IntentsGuilds lets us observe GuildCreate so the bot can
	// register (or explicitly refuse to register) commands on a
	// guild that was added while the process was running. Not
	// privileged; no portal approval required.
	session.Identify.Intents = discordgo.IntentsGuilds

	client := bot.NewServerClient(cfg.ServerBaseURL, cfg.AdminToken)
	handler := bot.NewHandler(cfg, client, log)

	// Wire the interaction handler before Open — discordgo's
	// AddHandler is goroutine-safe but adding after Open risks
	// dropping events during the race window.
	session.AddHandler(handler.Dispatch)

	// Guild allow-list enforcement happens in two places:
	//  1. On Ready, register commands on every guild in the
	//     allow-list. Commands on non-listed guilds don't exist,
	//     period.
	//  2. On GuildCreate (bot added to a guild mid-run), register
	//     only if the new guild is on the allow-list. Non-listed
	//     guilds are logged and skipped — we do NOT auto-leave
	//     because an operator might be mid-setup and have not
	//     yet updated CMDCTRL_DISCORD_GUILD_IDS.
	session.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		log.Info("discord session ready", "username", s.State.User.Username)
		bot.RegisterCommands(s, cfg.AppID, cfg.GuildIDs, log)
	})
	session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if !cfg.GuildAllowed(g.ID) {
			log.Warn("bot is in a non-allowed guild; skipping registration", "guild", g.ID, "name", g.Name)
			return
		}
		bot.RegisterCommands(s, cfg.AppID, []string{g.ID}, log)
	})

	if err := session.Open(); err != nil {
		log.Error("discord session open", "error", err.Error())
		os.Exit(1)
	}
	log.Info("bot running; Ctrl-C or SIGTERM to stop")

	// Block until a shutdown signal arrives. systemd sends
	// SIGTERM on `systemctl stop`; a dev loop uses SIGINT.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info("shutdown signal; closing discord session")
	if err := session.Close(); err != nil {
		log.Warn("discord session close", "error", err.Error())
	}
}
