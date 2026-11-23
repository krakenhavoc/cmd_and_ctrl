package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ringTemptTimes has the Ring tempt `player` n times. Each time must be
// a forced choice or no choice: the test arranges at most one creature.
func RingTemptTimes(t *testing.T, g *game.Game, player uuid.UUID, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		RingTempt(t, g, player)
		if RingPrompt(g, player) != nil {
			t.Fatal("setup: the tempt asked for a Ring-bearer; arrange one creature")
		}
	}
}

// ringSettle passes priority until the stack is empty, ordering any
// simultaneous triggers in the order offered (CR 603.3b). It stops at
// any other open prompt, which the caller answers.
func RingSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		PassPriorityAroundTable(t, g)
		order := LatestChoiceOfKind(g, game.PendingChoiceTriggerOrder)
		if order == nil {
			return
		}
		if err := g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	t.Fatal("the table never settled")
}
