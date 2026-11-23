package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sauronPT is a battlefield permanent's power and toughness with its
// counters, after a fresh layer pass.
func SauronPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%v is not on the battlefield", id)
	}
	return c.CurrentPower(), c.CurrentToughness()
}
