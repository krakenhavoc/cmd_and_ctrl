package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bringDownOracle = "05d2b89b-acc4-4289-95a3-a09ebb23a5d6"

// TestBringDownDestroysACreatureWithPowerFourOrGreater: a 4-power
// creature is destroyed, and a 3-power creature is refused at the cast.
func TestBringDownDestroysACreatureWithPowerFourOrGreater(t *testing.T) {
	assistCardSpec(t, bringDownOracle, "Bring Down")

	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := pushVanillaCreature(g, opp.ID, "Small", 3, 3)
	big := pushVanillaCreature(g, opp.ID, "Big", 4, 4)

	if err := castCatalogSpellErr(t, g, "Bring Down", "Sorcery", bringDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: small}}); err == nil {
		t.Fatal("Bring Down at a 3-power creature was cast")
	}
	castCatalogSpell(t, g, "Bring Down", "Sorcery", bringDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)

	if assistCardOnBattlefield(g, big) {
		t.Error("the 4-power creature survived Bring Down")
	}
	if !assistCardOnBattlefield(g, small) {
		t.Error("the 3-power creature was destroyed")
	}
}
