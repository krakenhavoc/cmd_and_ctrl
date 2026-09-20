package aiseat

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// pace_internal_test.go pins runner.pacingNow — the ADR 0075 §2.2 /
// sub-PR 6 rule that maps a table's game.Settings.BotPace onto this
// decision window's MinThink/MaxThink. In-package because pacingNow
// is unexported: the rule is small enough to unit test directly
// against a Runner literal rather than timing a whole bot game.

// paceRoom builds a *ws.Room around a fresh game.Game, with no seats
// and no Start — pacingNow only reads g.Settings, so nothing else
// about the game needs to be real.
func paceRoom(t *testing.T) *ws.Room {
	t.Helper()
	g := game.NewGame()
	return ws.NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
}

func TestPacingNowMapsEveryPresetWhenFollowingTablePace(t *testing.T) {
	room := paceRoom(t)
	cfg := Config{FollowTablePace: true, MinThink: 999 * time.Hour, MaxThink: defaultMaxThink}
	r := &Runner{room: room, cfg: cfg}

	cases := []struct {
		pace    game.BotPace
		wantMin time.Duration
		wantMax time.Duration
	}{
		{game.BotPaceFast, 0, 2 * time.Second},
		{game.BotPaceNormal, defaultMinThink, defaultMaxThink},
		{game.BotPaceSlow, 2 * time.Second, 8 * time.Second},
	}
	for _, c := range cases {
		t.Run(string(c.pace), func(t *testing.T) {
			pace := c.pace
			if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &pace}); err != nil {
				t.Fatalf("UpdateSettings: %v", err)
			}
			gotMin, gotMax := r.pacingNow()
			if gotMin != c.wantMin || gotMax != c.wantMax {
				t.Errorf("pacingNow() = (%v, %v), want (%v, %v)", gotMin, gotMax, c.wantMin, c.wantMax)
			}
		})
	}
}

func TestPacingNowKeepsStrongsDeadlineAsAFloor(t *testing.T) {
	room := paceRoom(t)
	// Simulates ConfigFor(TierStrong): DefaultConfig with MaxThink
	// raised to the tier's own 5s deadline.
	cfg := DefaultConfig()
	cfg.MaxThink = 5 * time.Second
	r := &Runner{room: room, cfg: cfg}

	fast := game.BotPaceFast
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &fast}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if _, max := r.pacingNow(); max != 5*time.Second {
		t.Errorf("fast pace on a strong seat: MaxThink = %v, want 5s (the tier's own deadline, never shortened)", max)
	}

	slow := game.BotPaceSlow
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &slow}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if _, max := r.pacingNow(); max != 8*time.Second {
		t.Errorf("slow pace on a strong seat: MaxThink = %v, want 8s (slow's preset still wins when it is the larger)", max)
	}
}

func TestPacingNowUnsetBotPaceKeepsConfiguredValues(t *testing.T) {
	room := paceRoom(t)
	// A zero-value BotPace is not one of the three presets — e.g. a
	// Settings struct that was never defaulted. FollowTablePace must
	// not touch MinThink/MaxThink in that case.
	empty := game.BotPace("")
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	room.Game.Settings.BotPace = empty

	cfg := Config{FollowTablePace: true, MinThink: 111 * time.Millisecond, MaxThink: 3333 * time.Millisecond}
	r := &Runner{room: room, cfg: cfg}
	gotMin, gotMax := r.pacingNow()
	if gotMin != cfg.MinThink || gotMax != cfg.MaxThink {
		t.Errorf("pacingNow() with unset BotPace = (%v, %v), want cfg's own (%v, %v)", gotMin, gotMax, cfg.MinThink, cfg.MaxThink)
	}
}

func TestPacingNowIgnoredWhenFollowTablePaceIsOff(t *testing.T) {
	room := paceRoom(t)
	slow := game.BotPaceSlow
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &slow}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	// The shape NewManagerWithConfig and most whole-game tests use to
	// strip pacing entirely: FollowTablePace is the zero value
	// (false), so explicit Config values must win even though the
	// table is set to "slow".
	cfg := Config{MinThink: 0, MaxThink: time.Second}
	r := &Runner{room: room, cfg: cfg}
	gotMin, gotMax := r.pacingNow()
	if gotMin != 0 || gotMax != time.Second {
		t.Errorf("pacingNow() with FollowTablePace off = (%v, %v), want the configured (0, 1s) untouched by the table's slow preset", gotMin, gotMax)
	}
}

func TestPacingNowReadsAMidGameSettingsChange(t *testing.T) {
	// The point of reading on every decision rather than once at
	// Start: a host changing the pace mid-game must be live on this
	// seat's very next call, with no restart of the runner.
	room := paceRoom(t)
	cfg := Config{FollowTablePace: true, MinThink: defaultMinThink, MaxThink: defaultMaxThink}
	r := &Runner{room: room, cfg: cfg}

	fast := game.BotPaceFast
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &fast}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if min, max := r.pacingNow(); min != 0 || max != 2*time.Second {
		t.Fatalf("before change: pacingNow() = (%v, %v), want fast's (0, 2s)", min, max)
	}

	slow := game.BotPaceSlow
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &slow}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if min, max := r.pacingNow(); min != 2*time.Second || max != 8*time.Second {
		t.Errorf("after mid-game change: pacingNow() = (%v, %v), want slow's (2s, 8s) — the runner must not have cached the old preset", min, max)
	}
}
