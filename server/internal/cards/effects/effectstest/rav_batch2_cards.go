package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ravB2Activate floats the given mana and activates a catalog ability.
func RavB2Activate(t *testing.T, g *game.Game, who, source uuid.UUID, index int, params game.ActivateAbilityParams, mana ...string) error {
	t.Helper()
	B06AddMana(g.PlayerByIDForEffect(who), mana...)
	params.Strict = true
	return g.ActivateCatalogAbility(who, source, index, params)
}

// ravB2MustActivate is ravB2Activate that fails the test on an error and
// lets the ability resolve.
func RavB2MustActivate(t *testing.T, g *game.Game, who, source uuid.UUID, index int, params game.ActivateAbilityParams, mana ...string) {
	t.Helper()
	if err := RavB2Activate(t, g, who, source, index, params, mana...); err != nil {
		t.Fatalf("activate ability %d: %v", index, err)
	}
	PassPriorityAroundTable(t, g)
}
