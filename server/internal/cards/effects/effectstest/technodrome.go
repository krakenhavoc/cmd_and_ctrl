package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TechnodromeRestricted(t *testing.T, g *game.Game, id uuid.UUID) (cantAttack, cantBlock bool) {
	t.Helper()
	g.BumpLayerVersionForTest()
	g.WithWriteLock(g.RecomputeLayersIfStaleLocked)
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return game.Restricted(&c, game.CantAttack), game.Restricted(&c, game.CantBlock)
}
