package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Abzan Battle Priest and Abzan Falconer --------------------------------

func RprB04ActivateOutlast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) error {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{})
}
