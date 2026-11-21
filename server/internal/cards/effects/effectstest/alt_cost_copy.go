package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// endTheTurn walks to the end step and lets its delayed triggers resolve.
// The original's and the token's trigger at the same time, so the order
// prompt between them is answered as offered.
func EndTheTurn(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceTo(t, g, game.StepEnd)
	for i := 0; i < 8; i++ {
		if TriggerOrderPrompt(g) != nil {
			AnswerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if StackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return
		}
		PassPriorityAroundTable(t, g)
	}
	t.Fatal("the end step did not settle")
}
