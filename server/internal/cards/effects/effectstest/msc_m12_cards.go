package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mscM12Settle resolves the stack, answering the CR 603.3b order prompt the
// two Vulture rows raise when they trigger together (their order is moot).
func MscM12Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		if TriggerOrderPrompt(g) != nil {
			AnswerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		PassPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
}
