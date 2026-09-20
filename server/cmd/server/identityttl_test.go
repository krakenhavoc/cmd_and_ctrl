package main

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// CMDCTRL_IDENTITY_TTL (ADR 0051 decision 3). The fatal paths call
// os.Exit, so what is tested is the default and a parsed value.

func TestIdentityTTLDefaultsToThirtyDays(t *testing.T) {
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")
	t.Setenv("CMDCTRL_IDENTITY_TTL", "")
	t.Setenv("CMDCTRL_SESSION_TTL", "")

	cfg := loadConfig(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if cfg.IdentityTTL != 720*time.Hour {
		t.Errorf("IdentityTTL = %v, want 720h", cfg.IdentityTTL)
	}
	if cfg.SessionTTL != 12*time.Hour {
		t.Errorf("SessionTTL = %v, want 12h: the identity TTL must not move it", cfg.SessionTTL)
	}
}

func TestIdentityTTLIsReadSeparately(t *testing.T) {
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")
	t.Setenv("CMDCTRL_IDENTITY_TTL", "168h")
	t.Setenv("CMDCTRL_SESSION_TTL", "6h")

	cfg := loadConfig(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if cfg.IdentityTTL != 168*time.Hour || cfg.SessionTTL != 6*time.Hour {
		t.Errorf("IdentityTTL %v, SessionTTL %v; want 168h and 6h", cfg.IdentityTTL, cfg.SessionTTL)
	}
}

// With no database there is no revocation list, and the lobby gets a
// nil interface rather than a typed nil it cannot tell from a real one.
func TestNoDatabaseMeansNoRevocations(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := newRevocations(context.Background(), log, users.NoStore{})
	if r != nil {
		t.Fatalf("newRevocations(NoStore) = %v, want nil", r)
	}
	if lobbyRevoker(r) != nil {
		t.Error("lobbyRevoker(nil) is a non-nil interface; the routes would not 503")
	}
}
