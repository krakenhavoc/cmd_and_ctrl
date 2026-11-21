package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// advanceToDeclareAttackersOf walks AdvanceStep to `seat`'s
// declare-attackers step.
func AdvanceToDeclareAttackersOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepDeclareAttackers && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward declare attackers of seat %d: %v", seat, err)
		}
	}
	t.Fatalf("never reached declare attackers of seat %d", seat)
}

func HasRequirementRecord(g *game.Game) bool {
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddAttackRequirement {
				return true
			}
		}
	}
	return false
}
