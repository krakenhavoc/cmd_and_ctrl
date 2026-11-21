package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_st.go — Lost Caverns of Ixalan helpers shared across test packages.

// LciSTActivateFromHand seeds a catalog card into the active seat's hand
// in its main phase, floats `mana`, and activates the hand ability at
// `index`, returning the card's id.
func LciSTActivateFromHand(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, index int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	id := PushCatalogHandCard(me, name, typeLine, oracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, index, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: activate hand ability %d: %v", name, index, err)
	}
	return id
}
