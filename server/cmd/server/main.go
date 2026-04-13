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
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg := loadConfig(log)

	// Auth + room manager are global singletons for the lifetime of
	// the process. They outlive individual games.
	authenticator := auth.NewMemoryAuthenticator()
	mgr := ws.NewRoomManager(log, cfg.DataDir)
	l := lobby.NewLobby(mgr)

	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&lobby.WSAuthorizer{Auth: authenticator})

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
	mux.Handle("/", lobby.Handler(lobby.Config{
		Lobby:      l,
		Auth:       authenticator,
		AdminToken: cfg.AdminToken,
		SessionTTL: cfg.SessionTTL,
	}))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
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
