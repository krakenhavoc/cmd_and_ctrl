package testsi

import (
	"testing"

	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Indrik Umbra (#3654): +4/+4, first strike, every creature able to
// block it does so (Lure's requirement), and umbra armor.

func TestIndrikUmbraPumpsGivesFirstStrikeAndMakesEveryoneBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := effectstest.ReqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	w1 := effectstest.ReqCreature(g, opp.ID, "Wall One", "", 0, 4)
	effectstest.Enchant(t, g, "Indrik Umbra", cardshu.IndrikUmbraOracleID, bear)

	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 6 || tg != 6 {
		t.Fatalf("enchanted bear is %d/%d, want 6/6", p, tg)
	}
	effectstest.AssertEffectiveKeyword(t, g, bear, "first strike")

	effectstest.ReqToBlockers(t, g, bear, other)
	effectstest.ReqRefusal(t, g.PassPriority())
	effectstest.ReqRefusal(t, effectstest.ReqBlock(t, g, other, w1))
}

func TestIndrikUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Indrik Umbra", cardshu.IndrikUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
