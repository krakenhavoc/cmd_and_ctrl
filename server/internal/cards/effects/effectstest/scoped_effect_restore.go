package effectstest

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scoped_effect_restore_test.go pins what ADR 0041 phase 3's first
// slice (#1497) is for: a card whose continuous effect used to be a
// closure-bearing ScopedStatic — and so froze its table's restore
// point for as long as the effect lived — now leaves a game that IS a
// restore point, and the effect is still there, and still ends, in the
// restored game.
//
// Every test asserts what a player can see (P/T, types, subtypes)
// across a real capture → JSON → restore, not merely that a record was
// written.

// restoreRoundTrip captures g, sends it through JSON and restores it.
// strict demands a full restore point (RestoreStrict, the boot path);
// otherwise Restore is used and the census is only reported.
func RestoreRoundTrip(t *testing.T, g *game.Game, strict bool) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if strict && !snap.Restorable() {
		t.Fatalf("the game is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var restored *game.Game
	if strict {
		restored, err = decoded.RestoreStrict()
	} else {
		restored, err = decoded.Restore()
	}
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	return restored
}
