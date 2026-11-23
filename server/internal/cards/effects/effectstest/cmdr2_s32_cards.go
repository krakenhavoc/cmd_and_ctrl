package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Otherworldly Escort ------------------------------------------------

func Cmdr2S32Sacrifice(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	g.RunStateChecksForTest()
	PassPriorityAroundTable(t, g)
}
