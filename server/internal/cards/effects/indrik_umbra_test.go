package effects

import "testing"

// Indrik Umbra (#3654): +4/+4, first strike, every creature able to
// block it does so (Lure's requirement), and umbra armor.

func TestIndrikUmbraPumpsGivesFirstStrikeAndMakesEveryoneBlockIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	w1 := reqCreature(g, opp.ID, "Wall One", "", 0, 4)
	enchant(t, g, "Indrik Umbra", indrikUmbraOracleID, bear)

	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 6 || tg != 6 {
		t.Fatalf("enchanted bear is %d/%d, want 6/6", p, tg)
	}
	assertEffectiveKeyword(t, g, bear, "first strike")

	reqToBlockers(t, g, bear, other)
	reqRefusal(t, g.PassPriority())
	reqRefusal(t, reqBlock(t, g, other, w1))
}

func TestIndrikUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Indrik Umbra", indrikUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
