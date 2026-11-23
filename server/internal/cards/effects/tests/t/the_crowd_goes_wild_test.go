package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theCrowdGoesWildOracle = "21302f37-0632-4202-9823-f08c5aec016e"

// TestTheCrowdGoesWildCountersUpToXAndGivesTrampleToCountered: with X =
// 2 two creatures get a +1/+1 counter, and trample lands on exactly the
// creatures that carry a counter, not the one that was left alone.
func TestTheCrowdGoesWildCountersUpToXAndGivesTrampleToCountered(t *testing.T) {
	effectstest.AssistCardSpec(t, theCrowdGoesWildOracle, "The Crowd Goes Wild")

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	bystander := effectstest.PushVanillaCreature(g, opp.ID, "Bystander", 2, 2)

	effectstest.AssistCardCastX(t, g, "The Crowd Goes Wild", "Sorcery", theCrowdGoesWildOracle, "{X}{G}", []string{"G", "C", "C"}, 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs} {
		if n := effectstest.FindBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]; n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, n)
		}
		if got := effectstest.EffectiveAbilities(t, g, id); !effectstest.ContainsString(got, "trample") {
			t.Errorf("%s abilities = %v, want trample until end of turn", name, got)
		}
	}
	if n := effectstest.FindBattlefieldCardForTest(g, bystander).Counters[game.CounterPlusOne]; n != 0 {
		t.Errorf("the untargeted creature has %d counters, want 0", n)
	}
	if got := effectstest.EffectiveAbilities(t, g, bystander); effectstest.ContainsString(got, "trample") {
		t.Errorf("the uncountered creature abilities = %v, want no trample", got)
	}
}
