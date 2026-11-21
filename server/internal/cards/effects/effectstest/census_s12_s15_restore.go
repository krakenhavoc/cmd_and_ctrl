package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// strictRestoreWithCopyWalk takes a strict restore through JSON, after
// checking the open copy prompt was written as a copy walk (#4361 S15).
func StrictRestoreWithCopyWalk(t *testing.T, g *game.Game) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for _, c := range snap.PendingChoices {
		if c.PickThen == game.CopyWalkKey && c.CopyWalk != nil {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prompt in the snapshot is a written copy walk: %+v", snap.Continuations)
	}
	return RestoreRoundTrip(t, g, true)
}
