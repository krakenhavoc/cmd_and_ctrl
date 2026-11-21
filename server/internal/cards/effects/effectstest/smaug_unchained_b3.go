package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sub3Destroy destroys a permanent and lets every trigger resolve.
func Sub3Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	MonSettle(t, g)
}

// sub3TriggersFrom counts the triggered items from one source waiting or
// on the stack.
func Sub3TriggersFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			n++
		}
	}
	return n
}
