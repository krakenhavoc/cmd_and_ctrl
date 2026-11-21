package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// passUntilCopies passes priority until the trigger has resolved and
// made copies, and stops there with the cast window open.
func PassUntilCopies(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 12 && len(CardCopies(g)) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(CardCopies(g)) == 0 {
		t.Fatal("no copies appeared")
	}
}
