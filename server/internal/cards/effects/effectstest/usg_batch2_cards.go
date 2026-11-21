package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// usgB2Activate floats `mana` for `who`, then activates ability 0 of `id`
// with strict payment, paying any sacrifice with `sac`.
func UsgB2Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, mana string, targets []game.TargetRef, sac ...uuid.UUID) error {
	t.Helper()
	FloatForTest(g, who, mana)
	return g.ActivateCatalogAbility(who.ID, id, 0, game.ActivateAbilityParams{Targets: targets, SacrificeIDs: sac, Strict: true})
}
