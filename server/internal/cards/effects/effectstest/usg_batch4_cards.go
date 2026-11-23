package effectstest

import (
	"testing"
)

// usgB4Cycles cycles a card from the active seat's hand for {2} and checks
// it was replaced by one draw.
func UsgB4Cycles(t *testing.T, name, typeLine, oracle string) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	id, _ := CycleFromHand(t, g, name, typeLine, oracle, "{C}{C}")
	PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Errorf("%s: the cycled card is not in the graveyard", name)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("%s: hand %d → %d, want the cycled card replaced by one draw", name, hand, me.Hand.Size())
	}
}
