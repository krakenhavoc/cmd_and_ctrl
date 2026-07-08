// Command server is the cmd_and_ctrl game server. As of S04 it exposes:
//
//	GET  /healthz          — liveness probe
//	POST /admin/login      — exchange admin token for an admin session
//	POST /games            — admin-only: create a new game
//	GET  /games            — authenticated: list games
//	GET  /games/{id}       — authenticated: game metadata
//	POST /games/{id}/join  — public (invite-gated): seat a player + mint session
//	POST /games/{id}/start — authenticated: transition lobby → active
//	GET  /me               — authenticated: echo principal
//	GET  /ws               — WebSocket endpoint speaking the v0 protocol
//
// Configuration is via environment variables:
//
//	CMDCTRL_ADDR          — listen address, default ":8080"
//	CMDCTRL_DATA_DIR      — root for crash-recovery snapshots, default "./data"
//	                        set to empty string to disable disk writes entirely
//	CMDCTRL_ADMIN_TOKEN   — shared secret for the admin login flow. REQUIRED.
//	                        Set to a long random string; anyone with this can
//	                        create games and list the whole lobby.
//	CMDCTRL_SESSION_TTL   — session lifetime, Go duration string (e.g. 12h).
//	                        Default 12h.
//	CMDCTRL_SEED_DEMO     — if "1", seed a 4-player demo game at startup.
//	                        Useful for the gamecli dev loop when you want a
//	                        ready-to-go room without going through the lobby.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	// Blank import: effects/wire.go's init() populates the S14
	// EffectResolver / ETBEffectHook / IsCatalogCard callbacks on
	// the game package. Without this import the catalog stays cold
	// and every card falls through to manual sandbox resolution.
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/envflag"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const (
	// scryfallPath is the relative location of the bulk dump inside
	// CMDCTRL_DATA_DIR. Matches scripts/scryfall-refresh.sh.
	scryfallSubdir      = "scryfall"
	scryfallDefaultFile = "default-cards.json"
	imagesSubdir        = "images"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	// Dev convenience: warn loudly when deck validation is being
	// skipped. Catches the case where the env var leaks into a
	// production deployment — a WARN on every boot is harder to
	// miss than a quiet behaviour change.
	if os.Getenv("CMDCTRL_DEV_SKIP_DECK_VALIDATION") != "" {
		log.Warn("CMDCTRL_DEV_SKIP_DECK_VALIDATION is set; deck validation is BYPASSED. Do not use in production.")
	}
	// Same family: the e2e suite relaxes the lobby rate limiters via
	// this knob (see lobby.Handler). Loud on boot for the same
	// leaked-into-prod reason.
	if envflag.Truthy(os.Getenv("CMDCTRL_DEV_RELAX_RATE_LIMITS")) {
		log.Warn("CMDCTRL_DEV_RELAX_RATE_LIMITS is set; lobby rate limits are EFFECTIVELY DISABLED. Do not use in production.")
	}

	cfg := loadConfig(log)

	// Auth + room manager are global singletons for the lifetime of
	// the process. They outlive individual games.
	authenticator := auth.NewMemoryAuthenticator()
	mgr := ws.NewRoomManager(log, cfg.DataDir)
	l := lobby.NewLobby(mgr)

	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&lobby.WSAuthorizer{Auth: authenticator})
	// Lobby HTTP mutations (join/deck/start) broadcast through the
	// hub so clients already on the game page see them immediately.
	l.SetStateBroadcaster(hub)
	if len(cfg.AllowedOrigins) > 0 {
		hub.SetAllowedOrigins(cfg.AllowedOrigins)
		log.Info("ws allowed-origins configured", "hosts", cfg.AllowedOrigins)
	}

	// Card index + image cache. The Scryfall bulk dump is optional
	// at startup: a missing file logs a warning and the /cards*
	// routes serve 404 / 503 until the next scryfall-refresh run.
	// This keeps the server bootable on a fresh deployment before
	// the cron has landed its first dump.
	cardIdx, imgCache := loadCardAssets(log, cfg.DataDir)

	// Optional demo game for the gamecli dev path. Creates a game
	// directly (bypassing the lobby's invite flow) so you can dial
	// the ws endpoint with any admin token and a ?game=<demo-id>
	// query param.
	if cfg.SeedDemo {
		g := seedDemoGame(log)
		mgr.Create(g)
		log.Info("seeded demo game", "id", g.ID.String())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /ws", hub.ServeWS)
	mux.Handle("/cards/", auth.Middleware(authenticator)(cards.Handler(cardIdx, imgCache)))
	discordCfg := discord.ConfigFromEnv()
	if discordCfg.Enabled() {
		log.Info("discord oauth enabled", "redirect_uri", discordCfg.RedirectURI)
	} else {
		log.Info("discord oauth disabled — set CMDCTRL_DISCORD_CLIENT_ID/SECRET/REDIRECT_URI to enable")
	}

	// Avatar cache lives under the same data dir as the scryfall
	// index and replay dumps. Empty DataDir disables the cache —
	// the /avatars endpoint will then 503 and the client falls
	// back to its initials placeholder.
	var avatarDir string
	if cfg.DataDir != "" {
		avatarDir = filepath.Join(cfg.DataDir, "avatars")
	}
	avatarCache := discord.NewAvatarCache(avatarDir, nil)

	mux.Handle("/", lobby.Handler(lobby.Config{
		Lobby:             l,
		Auth:              authenticator,
		AdminToken:        cfg.AdminToken,
		SessionTTL:        cfg.SessionTTL,
		Cards:             cardIdx,
		Evictor:           hub,
		Discord:           discordCfg,
		DiscordStateStore: discord.NewStateStore(),
		DiscordAvatars:    avatarCache,
	}))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		// Reap idle keep-alive connections so abandoned sockets don't
		// pin fds indefinitely. NO WriteTimeout on purpose: it would
		// sever long-lived WebSockets and replay streams mid-flight.
		IdleTimeout: 120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("server listening", "addr", cfg.Addr, "data_dir", cfg.DataDir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Order matters: close the HTTP listener first so no new WebSocket
	// upgrades can slip in after we've taken a snapshot of live clients.
	// ServeWS is a hijacking handler that returns immediately after
	// spawning its pumps, so srv.Shutdown itself completes quickly —
	// it's the act of closing the listener that we need.
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown", "err", err)
	}
	hub.Shutdown(shutdownCtx)
	log.Info("server stopped")
}

// config is the minimal set of env-derived values main() needs. Kept
// in its own struct so that validation (in loadConfig) is a single
// call site rather than scattered os.Getenv checks.
type config struct {
	Addr       string
	DataDir    string
	AdminToken string
	SessionTTL time.Duration
	SeedDemo   bool
	// AllowedOrigins is the cross-origin hostname allow-list passed to
	// the hub's WebSocket CheckOrigin. Same-origin is always allowed
	// (no config needed). Set CMDCTRL_ALLOWED_ORIGINS to a comma-
	// separated list for LAN clients reaching the server from a
	// different host/port than the one it binds on.
	AllowedOrigins []string
}

// loadConfig pulls the server's env vars, applies defaults, and
// exits (with a user-facing log message) if a required value is
// missing. This runs BEFORE the server does any work so misconfig
// shows up immediately instead of 500-ing the first login attempt.
func loadConfig(log *slog.Logger) config {
	c := config{
		Addr:       envOr("CMDCTRL_ADDR", ":8080"),
		AdminToken: os.Getenv("CMDCTRL_ADMIN_TOKEN"),
		SeedDemo:   os.Getenv("CMDCTRL_SEED_DEMO") == "1",
		SessionTTL: 12 * time.Hour,
	}

	if raw := os.Getenv("CMDCTRL_SESSION_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			log.Error("CMDCTRL_SESSION_TTL invalid", "value", raw, "err", err)
			os.Exit(1)
		}
		if d <= 0 {
			log.Error("CMDCTRL_SESSION_TTL must be positive", "value", raw)
			os.Exit(1)
		}
		c.SessionTTL = d
	}

	// CMDCTRL_DATA_DIR: honour an explicit empty string (disables
	// disk writes). Unset → "./data".
	if _, present := os.LookupEnv("CMDCTRL_DATA_DIR"); present {
		c.DataDir = os.Getenv("CMDCTRL_DATA_DIR")
	} else {
		c.DataDir = "./data"
	}

	if raw := os.Getenv("CMDCTRL_ALLOWED_ORIGINS"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if s = strings.TrimSpace(s); s != "" {
				c.AllowedOrigins = append(c.AllowedOrigins, s)
			}
		}
	}

	if c.AdminToken == "" {
		log.Error("CMDCTRL_ADMIN_TOKEN is required — set it to a long random string before starting the server")
		os.Exit(1)
	}
	if len(c.AdminToken) < 16 {
		log.Error("CMDCTRL_ADMIN_TOKEN is too short — use at least 16 characters")
		os.Exit(1)
	}
	return c
}

func envOr(key, dflt string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return dflt
}

// loadCardAssets boots the card index and image cache from the data
// dir. A missing bulk dump is logged as a warning, not a fatal: the
// server stays usable for lobby-only operations and starts serving
// card images as soon as the next scryfall-refresh writes the file.
//
// Empty DataDir (disk persistence fully disabled) returns a nil
// index and cache; the /cards routes handle that with 404 / 503.
func loadCardAssets(log *slog.Logger, dataDir string) (*cards.Index, *cards.ImageCache) {
	if dataDir == "" {
		log.Warn("CMDCTRL_DATA_DIR is empty; card index and image cache disabled")
		return nil, nil
	}
	idx := cards.NewIndex()
	dumpPath := filepath.Join(dataDir, scryfallSubdir, scryfallDefaultFile)
	if n, err := idx.Load(dumpPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn("scryfall dump not found; run scripts/scryfall-refresh.sh", "path", dumpPath)
		} else {
			log.Error("scryfall dump load failed", "path", dumpPath, "err", err)
		}
	} else {
		log.Info("scryfall index loaded", "cards", n, "path", dumpPath)
	}

	cache, err := cards.NewImageCache(filepath.Join(dataDir, imagesSubdir))
	if err != nil {
		log.Error("image cache init failed", "err", err)
		return idx, nil
	}
	return idx, cache
}

// seedDemoGame builds a 4-player Commander game in the active state,
// each seat with a 99-card filler library plus a placeholder
// commander. This exists so the gamecli dev path has a target
// without going through the lobby flow — turned on via
// CMDCTRL_SEED_DEMO=1.
func seedDemoGame(log *slog.Logger) *game.Game {
	g := game.NewGame()
	names := []string{"Alice", "Bob", "Carol", "Dave"}
	for i, name := range names {
		deck := make([]game.Card, 0, 100)
		deck = append(deck, game.NewCommander(fmt.Sprintf("Demo Commander %d", i+1), uuid.Nil))
		for j := range 99 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %03d", j+1), uuid.Nil))
		}
		p, err := g.AddPlayer(name, deck)
		if err != nil {
			log.Error("seedDemoGame AddPlayer failed", "name", name, "err", err)
			os.Exit(1)
		}
		log.Info("demo player seated", "name", name, "seat", p.Seat, "id", p.ID.String())
	}
	if err := g.Start(nil); err != nil {
		log.Error("seedDemoGame Start failed", "err", err)
		os.Exit(1)
	}
	return g
}
