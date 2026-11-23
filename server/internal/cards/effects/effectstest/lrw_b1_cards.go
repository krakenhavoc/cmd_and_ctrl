package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lrwB1Activate activates row `row` of a catalog permanent.
func LrwB1Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, row int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who.ID, id, row, params); err != nil {
		t.Fatalf("activate row %d: %v", row, err)
	}
}
