package effectstest

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_prompts_restore_test.go — #4256: the outside-the-game picks, the
// discard-then prompts and the pay-unless prompts are keyed, so a table
// waiting on any of them is a restore point. Each test takes a STRICT
// restore (restoreRoundTrip with strict, which fails on a table that is
// not a restore point) while the prompt is open, then answers it on the
// restored table. The *UnknownKeyIsRefused tests are the rollback case:
// a binary that does not know the key refuses the file.

// refuseUnknownKey rewrites every snapshot prompt naming `key` to an
// unregistered key and checks the restore is refused.
func RefuseUnknownKey(t *testing.T, g *game.Game, key string) {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].PickThen == key {
			snap.PendingChoices[i].PickThen = "choose-cards/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatalf("the snapshot holds no prompt naming %q", key)
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown key: err = %v, want ErrUnknownEffectKey", err)
	}
}
