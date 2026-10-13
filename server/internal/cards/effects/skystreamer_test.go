package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const skystreamerOracle = "aaf6d66b-67b1-43f1-a8bf-db83644f7d04"

// TestSkystreamerGainsFourForWhicheverPlayerIsPicked: "target player"
// is any player, so the gain lands on the picked opponent, not the
// caster.
func TestSkystreamerGainsFourForWhicheverPlayerIsPicked(t *testing.T) {
	assistCardSpec(t, skystreamerOracle, "Skystreamer")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	meBefore, oppBefore := me.Life, opp.Life

	id := castCatalogSpell(t, g, "Skystreamer", "Creature — Griffin", skystreamerOracle, nil)
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	assistCardSettle(t, g, me.ID, &pick, true)

	if opp.Life != oppBefore+4 {
		t.Errorf("picked player at %d, want %d", opp.Life, oppBefore+4)
	}
	if me.Life != meBefore {
		t.Errorf("caster at %d, want %d (the gain went to the pick)", me.Life, meBefore)
	}
	if got := effectiveAbilities(t, g, id); !containsString(got, "flying") {
		t.Errorf("Skystreamer abilities = %v, want flying", got)
	}
}
