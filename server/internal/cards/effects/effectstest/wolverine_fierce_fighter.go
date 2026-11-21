package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func DealDamageTo(t *testing.T, g *game.Game, source, target uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, target, n); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
}
