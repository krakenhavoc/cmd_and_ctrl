package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b17.go — S82 SNC batch 17 helpers shared across test packages.

// SncSettle passes priority until the stack is empty, answering a target
// pick addressed to chooser with `pick` when there is one.
func SncSettle(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 12; i++ {
		AnswerOpenOrderingIfAny(t, g, chooser)
		if pick != uuid.Nil && LatestPickTarget(g, chooser) != nil {
			PickCard(t, g, chooser, pick)
			continue
		}
		if c := PayAmountPrompt(g, chooser); c != nil {
			if err := g.ResolvePayAmount(c.ID, chooser, c.PayAmount.Goal); err != nil {
				t.Fatalf("attack split: %v", err)
			}
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		PassPriorityAroundTable(t, g)
	}
}
