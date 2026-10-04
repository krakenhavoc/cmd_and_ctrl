package main

import (
	"io"
	"log/slog"
	"testing"
)

// botthink_test.go: the thinking experiment (#2196) is OFF unless an
// operator turns it on, and loadConfig reads both of its variables.
// A malformed value exits the boot (model.ThinkingFromEnv's error),
// which is the half that cannot be tested in-process.

func TestBotThinkingDefaultsToOff(t *testing.T) {
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")
	t.Setenv("CMDCTRL_BOT_THINK", "")
	t.Setenv("CMDCTRL_BOT_MAX_TOKENS", "")

	cfg := loadConfig(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if cfg.BotThink || cfg.BotMaxTokens != 0 {
		t.Errorf("think %v max tokens %d; both must be off by default", cfg.BotThink, cfg.BotMaxTokens)
	}
}

func TestBotThinkingIsRead(t *testing.T) {
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "a-sufficiently-long-admin-token")
	t.Setenv("CMDCTRL_DATA_DIR", "")
	t.Setenv("CMDCTRL_BOT_THINK", "1")
	t.Setenv("CMDCTRL_BOT_MAX_TOKENS", "6000")

	cfg := loadConfig(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if !cfg.BotThink || cfg.BotMaxTokens != 6000 {
		t.Errorf("think %v max tokens %d, want true / 6000", cfg.BotThink, cfg.BotMaxTokens)
	}
}
