package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declared_trigger_restore_test.go — ADR 0041 P9, tier 4-2 (#1497):
// real cards whose triggers the constructors now declare, waiting on
// the stack, restored through JSON as a deploy restores them, and
// resolved in the restored game. The corpus holds the files; these hold
// the behaviour.

// restoreThroughJSON is a capture written and read back, as the boot
// path reads one.
func RestoreThroughJSON(t *testing.T, g *game.Game) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	restored, err := DecodeRestorePoint(t, MustMarshal(t, snap)).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}
