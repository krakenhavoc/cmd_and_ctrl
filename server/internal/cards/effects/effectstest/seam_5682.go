package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seam_5682.go — die-roll test helpers (#5682) shared across test packages.

// S5682Rolls lists the dice announced after the first `from` events, and
// fails on one that is not a d<sides> from `source` by `roller`.
func S5682Rolls(t *testing.T, g *game.Game, from int, roller, source uuid.UUID, sides int) []int {
	t.Helper()
	var out []int
	g.WithWriteLock(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind != game.EventRollDie {
				continue
			}
			if ev.Actor != roller || ev.Source != source || ev.Sides != sides || ev.Amount < 1 || ev.Amount > sides {
				t.Errorf("roll %+v, want a d%d by the roller from the source", ev, sides)
			}
			out = append(out, ev.Amount)
		}
	})
	return out
}
