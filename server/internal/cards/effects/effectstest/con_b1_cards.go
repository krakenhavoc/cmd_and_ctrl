package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conB1Activate floats `mana` and activates row `row` of `id` under the
// strict mana gate.
func ConB1Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, row int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	FloatForTest(g, who, mana)
	params.Strict = true
	return g.ActivateCatalogAbility(who.ID, id, row, params)
}
