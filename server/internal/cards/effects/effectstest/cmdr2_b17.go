package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const Cmdr2B17HermitOracle = "80e9a42e-f4f4-4bfc-91a2-ad124737dca2"

// cmdr2B17NextUpkeepOfSeat0 walks to the next time seat 0's upkeep
// begins and settles what that put on the stack.
func Cmdr2B17NextUpkeepOfSeat0(t *testing.T, g *game.Game) {
	t.Helper()
	// Leave the current turn first.
	for i := 0; i < 400 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	PassPriorityAroundTable(t, g)
}
