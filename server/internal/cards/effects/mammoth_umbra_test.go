package effects

import "testing"

// Mammoth Umbra (#3654): +3/+3 and vigilance, and umbra armor.

func TestMammothUmbraPumpsAndGivesVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Mammoth Umbra", mammothUmbraOracleID, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Fatalf("enchanted bear is %d/%d, want 5/5", p, tg)
	}
	assertEffectiveKeyword(t, g, bear, "vigilance")
}

func TestMammothUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Mammoth Umbra", mammothUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
