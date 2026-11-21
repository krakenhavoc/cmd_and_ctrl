package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ons1Cycle cycles a card from the active seat's hand for `mana` and
// checks it was discarded and that resolving the ability drew a card.
func Ons1Cycle(t *testing.T, g *game.Game, name, typeLine, oracle, mana string) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id, _ := CycleFromHand(t, g, name, typeLine, oracle, mana)
	if !me.Graveyard.Contains(id) {
		t.Fatalf("%s was not discarded to cycle", name)
	}
	before := len(me.Hand.Cards)
	PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("%s: hand %d -> %d, want one card drawn", name, before, got)
	}
}
