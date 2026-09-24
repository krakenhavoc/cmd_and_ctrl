package ws

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skip_tally_test.go pins ADR 0041 phase 3's P7 tally (#1497): every
// capture writeRestorePointLocked skips is counted by census kind, the
// run of consecutive skips resets when a restore point is written, and
// the shutdown census line carries all three numbers — which is the
// evidence a shutdown that happens to find every table idle cannot
// give.
func TestTheSkipTallyCountsByKindAndResetsTheRun(t *testing.T) {
	room, g, seat0 := newShutdownTestRoom(t)
	holdAnAbility := func() error {
		g.WithWriteLock(func() {
			id := uuid.New()
			g.StackMeta = map[uuid.UUID]*game.StackItem{id: {
				ID: id, Kind: game.StackItemActivated, Controller: seat0,
				Label:  "the tally test ability",
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}}
		})
		return nil
	}
	clear := func() error {
		g.WithWriteLock(func() { g.StackMeta = nil })
		return nil
	}

	for i := 0; i < 3; i++ {
		if _, _, err := room.Apply(seat0, holdAnAbility); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	}
	rep := room.ShutdownReport()
	if got := rep.SkippedByKind["stackEffects"]; got != 3 {
		t.Errorf("skipped stackEffects = %d, want 3 (by kind: %v)", got, rep.SkippedByKind)
	}
	if rep.SkipRun != 3 || rep.LongestSkipRun != 3 {
		t.Errorf("run = %d, longest = %d, want 3 and 3", rep.SkipRun, rep.LongestSkipRun)
	}

	// A clean action writes a restore point and ends the run; the
	// per-kind totals and the longest run are kept.
	if _, _, err := room.Apply(seat0, clear); err != nil {
		t.Fatalf("Apply clear: %v", err)
	}
	rep = room.ShutdownReport()
	if rep.SkipRun != 0 {
		t.Errorf("run = %d after a restore point was written, want 0", rep.SkipRun)
	}
	if rep.LongestSkipRun != 3 || rep.SkippedByKind["stackEffects"] != 3 {
		t.Errorf("longest = %d, by kind %v; want 3 and stackEffects 3 kept", rep.LongestSkipRun, rep.SkippedByKind)
	}

	var buf bytes.Buffer
	LogShutdownCensus(slog.New(slog.NewTextHandler(&buf, nil)), []*Room{room}, nil)
	out := buf.String()
	for _, want := range []string{"skipped_by_kind", "stackEffects:3", "longest_skip_run=3", "skip_run=0"} {
		if !strings.Contains(out, want) {
			t.Errorf("the shutdown census line does not carry %q:\n%s", want, out)
		}
	}
}
