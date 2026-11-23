package testsl

import (
	"testing"

	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

func TestLionUmbraPumpsAndGivesReachAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.ModifiedBear(g, me.ID)
	effectstest.AuraCast(t, g, "Lion Umbra", cardskr.LionUmbraOracleID, bear)
	// 2/2 with a +1/+1 counter, then +3/+3. The counter is layer 7d,
	// which only the post-layer characteristics (CurrentPower) include.
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Fatalf("enchanted bear is %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AssertEffectiveKeyword(t, g, bear, "reach")
	effectstest.AssertEffectiveKeyword(t, g, bear, "vigilance")
}

func TestLionUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.ModifiedBear(g, me.ID)
	umbra := effectstest.AuraCast(t, g, "Lion Umbra", cardskr.LionUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
