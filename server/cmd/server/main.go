// Command server is the cmd_and_ctrl game server. At S01 it exposes:
//
//	GET /healthz — liveness probe
//	GET /ws      — WebSocket endpoint speaking the v0 protocol
//
// Configuration is via environment variables:
//
//	CMDCTRL_ADDR — listen address, default ":8080"
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	addr := os.Getenv("CMDCTRL_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	hub := ws.NewHub(log)

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
		log.Info("server listening", "addr", addr)
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
