package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Damage(t *testing.T, g *game.Game, src, victim uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, victim, n); err != nil {
			t.Fatal(err)
		}
	})
}

// damageAndSettle deals damage and lets the state-based actions kill
// what it killed, then lets the triggers that follow resolve.
func DamageAndSettle(t *testing.T, g *game.Game, src, victim uuid.UUID, n int) {
	t.Helper()
	Damage(t, g, src, victim, n)
	SettleStateBasedActions(t, g)
	PassPriorityAroundTable(t, g)
}
