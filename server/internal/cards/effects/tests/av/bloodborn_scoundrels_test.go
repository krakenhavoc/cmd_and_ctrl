package testsav

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bloodbornScoundrelsOracle = "37c1de8d-30d6-4b08-9368-279cf66d3e85"

// TestBloodbornScoundrelsDrainsTheChosenOpponent: the ETB pick is an
// opponent, who loses 2 and the Scoundrels' controller gains 2.
func TestBloodbornScoundrelsDrainsTheChosenOpponent(t *testing.T) {
	effectstest.AssistCardSpec(t, bloodbornScoundrelsOracle, "Bloodborn Scoundrels")

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before, oppBefore := me.Life, opp.Life

	effectstest.CastCatalogSpell(t, g, "Bloodborn Scoundrels", "Creature — Vampire Rogue", bloodbornScoundrelsOracle, nil)
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)

	if opp.Life != oppBefore-2 {
		t.Errorf("target opponent at %d, want %d", opp.Life, oppBefore-2)
	}
	if me.Life != before+2 {
		t.Errorf("controller at %d, want %d", me.Life, before+2)
	}
}
