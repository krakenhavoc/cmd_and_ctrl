package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arnSetCombat marks `id` as attacking `attacking` and/or blocking
// `blocking` without walking the combat steps, for the spells and
// abilities that read who is in combat as they resolve.
func ArnSetCombat(g *game.Game, id, attacking, blocking uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = attacking
				g.Battlefield.Cards[i].BlockingTarget = blocking
			}
		}
	})
}

// arnTap taps a permanent the way an effect would.
func ArnTap(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
}

func ArnTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}
