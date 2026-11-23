package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// broB01Activate funds `mana` and activates ability idx strictly.
func BroB01Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, idx int, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	return g.ActivateCatalogAbility(me.ID, src, idx, game.ActivateAbilityParams{Strict: true})
}
