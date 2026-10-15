package effects

import "testing"

// Eland Umbra (#3654): +0/+4, and umbra armor.

func TestElandUmbraAddsFourToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Eland Umbra", elandUmbraOracleID, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tg != 6 {
		t.Fatalf("enchanted bear is %d/%d, want 2/6", p, tg)
	}
}

func TestElandUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Eland Umbra", elandUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
