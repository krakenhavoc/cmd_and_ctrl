package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// walkToEndStepOfSeat0 advances to seat 0's end step in a later round.
func WalkToEndStepOfSeat0(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 80; i++ {
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepEnd {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep iter %d: %v", i, err)
		}
	}
	t.Fatalf("never reached seat 0's end step")
}
