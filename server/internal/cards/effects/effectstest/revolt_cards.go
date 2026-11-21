package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revoltFor makes a permanent `owner` controls leave the battlefield
// (a sacrificed land), the cheapest way to turn revolt on.
func RevoltFor(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	land := PushCatalogPermanent(g, owner, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(land); err != nil {
			t.Fatalf("setup sacrifice: %v", err)
		}
	})
}

func BattlefieldPlusOnes(g *game.Game, id uuid.UUID) int {
	return CounterOnBattlefieldCard(g, id, game.CounterPlusOne)
}
