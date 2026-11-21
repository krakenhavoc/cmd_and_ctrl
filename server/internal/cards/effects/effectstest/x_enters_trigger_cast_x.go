package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_enters_trigger_cast_x_test.go — #1357. Before #1312
// (CastProvenance.X / Card.CastX(), CR 107.3m), an ETB trigger's Build
// had nowhere to read the X a spell was announced with — the
// resolving spell's StackItem is already gone by the time a trigger's
// Build runs — so The Goose Mother, Farmer Cotton, Springleaf Parade
// and Spiteful Banditry all moved their printed enters-TRIGGER effect
// into OnResolve instead, a beat before the permanent existed, with a
// declared caveat that the effect arrived without a trigger on the
// stack to respond to.
//
// Each test below proves the real shape now that #1312 has closed the
// seam: nothing happens as the spell resolves, the trigger reaches
// StackMeta as its own CR 603 object (waitForXTriggerThenSettle
// blocks there before letting it resolve), and only THEN does the
// printed effect happen, reading the X off Card.CastX().
// TestFlickeredFarmerCottonEntersTriggerReadsZeroCastX covers CR
// 107.3m's other half: a permanent that enters WITHOUT being cast —
// flickered, in this case — was never cast at all, so its trigger's
// CastX() is zero and its enters ability does nothing.

// waitForXTrigger passes priority until `source`'s ETB trigger
// reaches StackMeta and returns it — proving it is a real,
// independently respondable stack object rather than a side effect
// folded into the spell's own resolution. Fails the test if it never
// arrives (which is exactly what happens run against the pre-#1357
// shape: the token/damage effect lived in OnResolve, so no such
// trigger is ever queued at all).
func WaitForXTrigger(t *testing.T, g *game.Game, source uuid.UUID) *game.StackItem {
	t.Helper()
	for i := 0; i < 8; i++ {
		if item := TriggerOnStack(g, source); item != nil {
			return item
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority waiting for the enters trigger: %v", err)
		}
	}
	t.Fatal("the enters trigger never reached the stack")
	return nil
}
