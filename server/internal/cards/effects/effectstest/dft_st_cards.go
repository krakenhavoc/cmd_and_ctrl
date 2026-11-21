package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dftSTCycles puts the card into the active player's hand, cycles it
// with its cycling row at idx, and checks that it was discarded for a
// card.
func DftSTCycles(t *testing.T, name, typeLine, oracle string, idx int) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	SeedLibrary(me, "Drawn")
	id := PushCatalogHandCard(me, name, typeLine, oracle)
	hand := me.Hand.Size()
	FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: cycle: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Errorf("%s was not discarded to cycle", name)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("%s: hand is %d after cycling, want %d (one discarded, one drawn)", name, got, hand)
	}
}
