package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theCrowdGoesWildOracle = "21302f37-0632-4202-9823-f08c5aec016e"

// TestTheCrowdGoesWildCountersUpToXAndGivesTrampleToCountered: with X =
// 2 two creatures get a +1/+1 counter, and trample lands on exactly the
// creatures that carry a counter, not the one that was left alone.
func TestTheCrowdGoesWildCountersUpToXAndGivesTrampleToCountered(t *testing.T) {
	assistCardSpec(t, theCrowdGoesWildOracle, "The Crowd Goes Wild")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	bystander := pushVanillaCreature(g, opp.ID, "Bystander", 2, 2)

	assistCardCastX(t, g, "The Crowd Goes Wild", "Sorcery", theCrowdGoesWildOracle, "{X}{G}", []string{"G", "C", "C"}, 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs} {
		if n := findBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]; n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, n)
		}
		if got := effectiveAbilities(t, g, id); !containsString(got, "trample") {
			t.Errorf("%s abilities = %v, want trample until end of turn", name, got)
		}
	}
	if n := findBattlefieldCardForTest(g, bystander).Counters[game.CounterPlusOne]; n != 0 {
		t.Errorf("the untargeted creature has %d counters, want 0", n)
	}
	if got := effectiveAbilities(t, g, bystander); containsString(got, "trample") {
		t.Errorf("the uncountered creature abilities = %v, want no trample", got)
	}
}
