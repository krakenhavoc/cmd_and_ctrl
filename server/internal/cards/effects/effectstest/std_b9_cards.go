package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stdB9AddCounter puts a counter on a permanent through the engine.
func StdB9AddCounter(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, kind, n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	PassPriorityAroundTable(t, g)
}

// stdB9Kill destroys a permanent and settles the batched triggers.
func StdB9Kill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	g.RunStateChecksForTest()
	PassPriorityAroundTable(t, g)
}
