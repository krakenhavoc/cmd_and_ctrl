package effects

import "testing"

// Boar Umbra (#3654): +3/+3, and umbra armor.

func TestBoarUmbraPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Boar Umbra", boarUmbraOracleID, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Fatalf("enchanted bear is %d/%d, want 5/5", p, tg)
	}
}

func TestBoarUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Boar Umbra", boarUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
