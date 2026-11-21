package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cacklingCounterpartOracle is declared in token_copy_end_step_grant_test.go.
const ElectroduplicateOracle = "112f2b3f-32e7-40b1-b80e-0b99184a840c"

// passThroughEndStepTriggers resolves the end step's triggers, ordering
// them when two arrive at once.
func PassThroughEndStepTriggers(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		PassPriorityAroundTable(t, g)
		c := PendingOfKind(g, game.PendingChoiceTriggerOrder)
		if c == nil {
			return
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
}
