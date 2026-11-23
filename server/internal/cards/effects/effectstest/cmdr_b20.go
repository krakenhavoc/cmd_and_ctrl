package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// advanceDecliningTriggerPrompts walks the turn cursor to the declare
// attackers step of `seat`, answering "no" to any optional trigger.
func AdvanceDecliningTriggerPrompts(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepDeclareAttackers && g.Turn.ActiveSeat == seat {
			return
		}
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
				if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, false); err != nil {
					t.Fatal(err)
				}
				break
			}
		}
		if _, err := g.AdvanceStep(); err != nil {
			PassPriorityAroundTable(t, g)
		}
	}
	t.Fatalf("never reached declare attackers of seat %d", seat)
}
