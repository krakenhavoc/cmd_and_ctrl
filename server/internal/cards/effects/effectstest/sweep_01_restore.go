package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep01Discard is sweep01Pick for a discard prompt, by key.
func Sweep01Discard(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := DiscardChoiceFor(g, chooser)
	if c == nil || c.PickThen != want {
		t.Fatalf("the open discard is %+v, want one keyed %q", c, want)
	}
	return c
}
