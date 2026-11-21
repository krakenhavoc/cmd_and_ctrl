package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// replacement_resume_restore_test.go — #4253: a CR 616 prompt's resume
// frame is data (game/replacement_resume_keyed.go), so a table waiting on
// one is a restore point. Each test takes a STRICT restore (RestoreStrict,
// the boot path) while the prompt is open and answers it on the restored
// table, and checks what a player sees afterwards.

// strictRestoreWithFrame is restoreRoundTrip(strict) that also checks
// the open CR 616 prompt was written with its frame and tail key.
func StrictRestoreWithFrame(t *testing.T, g *game.Game, tail string) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for _, c := range snap.PendingChoices {
		if c.PickThen == tail {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prompt in the snapshot names the tail %q: %+v", tail, snap.Continuations)
	}
	return RestoreRoundTrip(t, g, true)
}
