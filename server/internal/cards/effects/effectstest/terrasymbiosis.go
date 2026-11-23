package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tsPut puts n counters of kind on id, placed by `by`.
func TsPut(t *testing.T, g *game.Game, by, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(by, id, kind, n); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
	})
}
