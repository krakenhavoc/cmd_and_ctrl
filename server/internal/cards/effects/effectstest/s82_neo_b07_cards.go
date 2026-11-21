package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// neoB07Channel puts the card in the active seat's hand in its precombat
// main phase, floats `mana`, and activates ability `idx` from the hand.
func NeoB07Channel(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, idx int,
	params game.ActivateAbilityParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	id := PushCatalogHandCard(me, name, typeLine, oracle)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	params.Strict = true
	return id, g.ActivateCatalogAbility(me.ID, id, idx, params)
}
