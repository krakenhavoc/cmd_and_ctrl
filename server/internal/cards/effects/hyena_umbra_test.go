package effects

import "testing"

// Hyena Umbra (#3654): +1/+1 and first strike, and umbra armor.

func TestHyenaUmbraPumpsAndGivesFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Hyena Umbra", hyenaUmbraOracleID, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
	assertEffectiveKeyword(t, g, bear, "first strike")
}

func TestHyenaUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Hyena Umbra", hyenaUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
