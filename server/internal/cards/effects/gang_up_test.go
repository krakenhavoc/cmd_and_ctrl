package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gangUpOracle = "dea311c3-9be3-4059-a206-5fbbdca4d6c9"

// TestGangUpDestroysACreatureWithPowerAtMostX: X = 2 destroys a 2-power
// creature, and a 3-power creature is refused at announcement.
func TestGangUpDestroysACreatureWithPowerAtMostX(t *testing.T) {
	assistCardSpec(t, gangUpOracle, "Gang Up")

	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := pushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Big", 3, 3)

	if _, err := assistCardCastXErr(t, g, "Gang Up", "Instant", gangUpOracle, "{X}{B}", []string{"B", "C", "C"}, 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}}); err == nil {
		t.Fatal("Gang Up at a 3-power creature with X = 2 was announced")
	}
	// The refused cast spent nothing, so the pool it was given still
	// pays for the legal cast below.
	assistCardCastX(t, g, "Gang Up", "Instant", gangUpOracle, "{X}{B}", nil, 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: small}})
	passPriorityAroundTable(t, g)

	if assistCardOnBattlefield(g, small) {
		t.Error("the 2-power creature survived Gang Up with X = 2")
	}
	if !assistCardOnBattlefield(g, big) {
		t.Error("the 3-power creature was destroyed without being a legal target")
	}
}
