// Command server is the cmd_and_ctrl game server. At S03 it exposes:
//
//	GET /healthz — liveness probe
//	GET /ws      — WebSocket endpoint speaking the v0 protocol
//
// On startup it seeds a default 4-player demo game that connected
// clients drive via the v0 action protocol. This singleton-game setup
// is a deliberate S03 shortcut; S04 replaces it with a lobby and
// per-session room management.
//
// Configuration is via environment variables:
//
//	CMDCTRL_ADDR     — listen address, default ":8080"
//	CMDCTRL_DATA_DIR — root for crash-recovery snapshots, default "./data"
//	                   set to empty string to disable disk writes entirely
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

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	addr := os.Getenv("CMDCTRL_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dataDir := os.Getenv("CMDCTRL_DATA_DIR")
	if dataDir == "" {
		// An explicit empty string via env would pass the first check
		// and disable writes; anything else (including unset) defaults
		// to "./data".
		if _, present := os.LookupEnv("CMDCTRL_DATA_DIR"); !present {
			dataDir = "./data"
		}
	}

	g := seedDemoGame(log)
	room := ws.NewRoom(g, log, dataDir)

	hub := ws.NewHub(log)
	hub.SetRoom(room)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /ws", hub.ServeWS)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("server listening", "addr", addr, "game_id", g.ID.String(), "data_dir", dataDir)
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

// seedDemoGame builds a 4-player Commander game in the active state,
// each seat with a 99-card filler library plus a placeholder
// commander. This exists so S03 clients can drive game state against
// a real Game without also having to implement a lobby. Seat names
// and player IDs are logged so the gamecli operator can read them
// from the server's stdout.
//
// S04 replaces this with a proper lobby.
//
// On any seeding failure we log with the structured slog handler and
// exit with a non-zero status rather than panic, so the server's
// startup logs stay consistent JSON and operators see a clean error
// instead of a stack trace.
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
