package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tapForTest taps a battlefield permanent through the public
// mutation, which is the same path a cost payment takes.
func TapForTest(t *testing.T, g *game.Game, ids ...uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		if err := g.TapCard(id, true); err != nil {
			t.Fatalf("TapCard: %v", err)
		}
	}
}
