package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eorSettle passes priority until the stack is empty. Two or more of the
// damage triggers at once, aimed at different opponents, are not alike, so
// their controller is asked for an order; any order will do.
func EorSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		PassPriorityAroundTable(t, g)
		c := PendingOfKind(g, game.PendingChoiceTriggerOrder)
		if c == nil {
			return
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	t.Fatal("the triggers never settled")
}
