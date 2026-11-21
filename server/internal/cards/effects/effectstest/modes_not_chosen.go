package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modePickChoicesFor is every open mode_pick prompt for a chooser, in
// queue order.
func ModePickChoicesFor(g *game.Game, chooser uuid.UUID) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceModePick && c.Chooser == chooser {
			out = append(out, c)
		}
	}
	return out
}

func AnswerMode(t *testing.T, g *game.Game, c *game.PendingChoice, chooser uuid.UUID, mode int) {
	t.Helper()
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
}
