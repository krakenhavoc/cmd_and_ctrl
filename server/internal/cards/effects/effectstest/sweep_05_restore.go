package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_05_restore_test.go — #4253's closure sweep, batch 05 (path_to_exile
// through selective_obliteration): the option picks, card picks and discards
// that carried a Go closure are keyed now, so a table waiting on any of them
// is a restore point. Each test opens the prompt, takes a STRICT restore
// through JSON (restoreRoundTrip with strict, which fails on a table that is
// not a restore point) and answers on the restored table.

// sweep05OptionKey returns the open option pick of `chooser`, failing the
// test when it names a key other than `want`.
func Sweep05OptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}
