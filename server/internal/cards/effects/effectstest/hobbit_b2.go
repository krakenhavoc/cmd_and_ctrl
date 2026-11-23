package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Dwarven Provisioner ----------------------------------------------

func HobActivateStrict(t *testing.T, g *game.Game, who, id uuid.UUID, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(who, id, 0, game.ActivateAbilityParams{Strict: true})
}
