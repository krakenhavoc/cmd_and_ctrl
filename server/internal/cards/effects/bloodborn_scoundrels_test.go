package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bloodbornScoundrelsOracle = "37c1de8d-30d6-4b08-9368-279cf66d3e85"

// TestBloodbornScoundrelsDrainsTheChosenOpponent: the ETB pick is an
// opponent, who loses 2 and the Scoundrels' controller gains 2.
func TestBloodbornScoundrelsDrainsTheChosenOpponent(t *testing.T) {
	assistCardSpec(t, bloodbornScoundrelsOracle, "Bloodborn Scoundrels")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before, oppBefore := me.Life, opp.Life

	castCatalogSpell(t, g, "Bloodborn Scoundrels", "Creature — Vampire Rogue", bloodbornScoundrelsOracle, nil)
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	assistCardSettle(t, g, me.ID, &pick, true)

	if opp.Life != oppBefore-2 {
		t.Errorf("target opponent at %d, want %d", opp.Life, oppBefore-2)
	}
	if me.Life != before+2 {
		t.Errorf("controller at %d, want %d", me.Life, before+2)
	}
}
