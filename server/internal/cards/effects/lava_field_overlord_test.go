package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const lavaFieldOverlordOracle = "f19a8c7e-1552-4786-a403-dd51c5cb2a7d"

// TestLavaFieldOverlordDealsFourToAnOpposingCreature: the ETB picks a
// creature an opponent controls, and 4 damage kills a 4-toughness one.
func TestLavaFieldOverlordDealsFourToAnOpposingCreature(t *testing.T) {
	assistCardSpec(t, lavaFieldOverlordOracle, "Lava-Field Overlord")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := pushVanillaCreature(g, opp.ID, "Foe", 4, 4)

	id := castCatalogSpell(t, g, "Lava-Field Overlord", "Creature — Dragon", lavaFieldOverlordOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: foe}
	assistCardSettle(t, g, me.ID, &pick, true)

	if assistCardOnBattlefield(g, foe) {
		t.Error("the 4-toughness creature survived 4 damage")
	}
	if !assistCardOnBattlefield(g, id) {
		t.Error("Lava-Field Overlord is not on the battlefield")
	}
	if got := effectiveAbilities(t, g, id); !containsString(got, "flying") {
		t.Errorf("Lava-Field Overlord abilities = %v, want flying", got)
	}
}
