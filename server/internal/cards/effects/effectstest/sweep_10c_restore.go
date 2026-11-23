package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_10c_restore_test.go — #4253 slice 10c: the card-name choice, the
// choose-a-source prompt, the land-mana walk and the draw-instead `done`
// are data, so a table waiting on one of them is a restore point. Each test
// takes a STRICT restore (RestoreStrict, the boot path) while the prompt
// is open, checks the snapshot wrote the prompt with its key, and answers
// it on the restored table.

// restoreAtKeyedPrompt strict-restores g after checking that an open prompt
// in its snapshot names the key (PickThen), which is what makes a binary
// without the key refuse the file.
func RestoreAtKeyedPrompt(t *testing.T, g *game.Game, key string) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for _, c := range snap.PendingChoices {
		if c.PickThen == key {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prompt in the snapshot names %q: %+v", key, snap.Continuations)
	}
	if snap.Continuations.ChoiceResumeFrames != 0 {
		t.Fatalf("the open prompt still counts a closure frame: %+v", snap.Continuations)
	}
	return RestoreRoundTrip(t, g, true)
}
