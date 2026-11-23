package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aroundTheTableToMain walks from seat 0's turn through the other three
// seats' turns to seat 0's next precombat main step, which is after its
// untap step.
func AroundTheTableToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for _, seat := range []int{1, 2, 3, 0} {
		AdvanceToMainOf(t, g, seat)
	}
}

// aroundTheTableToMainOf walks to the next precombat main step of `seat`,
// passing the turns of the others.
func AroundTheTableToMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 1; i <= len(g.Seats); i++ {
		AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+i)%len(g.Seats))
		if g.Turn.ActiveSeat == seat {
			return
		}
	}
}
