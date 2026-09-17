package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
)

// decisionlogconfig_test.go covers the one config rule that is easy
// to get backwards: CMDCTRL_BOT_DECISION_LOG_MODE fails the boot only
// when the log is actually ON.
//
// A mode that this build does not know, sitting in an env file beside
// an unset log, changes nothing about how the server runs. Refusing
// to start over it is a server that will not come back after a
// rollback, which is a far worse outcome than the typo it is
// objecting to. With the log on, the same value IS worth failing the
// boot for: a deployment that asked for `model` records and silently
// got `escalated` ones finds out after a night of games.
//
// loadConfig calls os.Exit on the fatal paths, so what is tested here
// is the non-fatal half — the half the rule is about.

func TestDecisionLogModeIsIgnoredWhenTheLogIsOff(t *testing.T) {
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")
	t.Setenv("CMDCTRL_BOT_DECISION_LOG", "")
	t.Setenv("CMDCTRL_BOT_DECISION_LOG_MODE", "verbose-please")

	cfg := loadConfig(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if cfg.BotDecisionLog != "" {
		t.Fatalf("the log is meant to be off: %q", cfg.BotDecisionLog)
	}
	if cfg.BotDecisionLogMode != decisionlog.ModeEscalated {
		t.Errorf("mode %q, want the default %q", cfg.BotDecisionLogMode, decisionlog.ModeEscalated)
	}
}

func TestDecisionLogModeIsReadWhenTheLogIsOn(t *testing.T) {
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")
	t.Setenv("CMDCTRL_BOT_DECISION_LOG", t.TempDir())
	t.Setenv("CMDCTRL_BOT_DECISION_LOG_MODE", "model")

	cfg := loadConfig(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if cfg.BotDecisionLog == "" {
		t.Fatal("the log directory was not read")
	}
	if cfg.BotDecisionLogMode != decisionlog.ModeModel {
		t.Errorf("mode %q, want %q", cfg.BotDecisionLogMode, decisionlog.ModeModel)
	}
}

func TestDecisionLogDefaultsToOff(t *testing.T) {
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")

	cfg := loadConfig(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if cfg.BotDecisionLog != "" {
		t.Errorf("the decision log must be OFF unless an operator names a directory: %q", cfg.BotDecisionLog)
	}
	if cfg.BotDecisionLogMode != decisionlog.ModeEscalated {
		t.Errorf("mode %q, want %q", cfg.BotDecisionLogMode, decisionlog.ModeEscalated)
	}
}
