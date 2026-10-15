package effects

import "testing"

// Eel Umbra (#3654): flash, +1/+1, and umbra armor.

func TestEelUmbraPrintsFlashAndPumps(t *testing.T) {
	spec, ok := Lookup(eelUmbraOracleID)
	if !ok || !hasKeyword(spec.PrintedKeywords, "flash") {
		t.Fatalf("Eel Umbra must print flash: %+v", spec.PrintedKeywords)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Eel Umbra", eelUmbraOracleID, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
}

func TestEelUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Eel Umbra", eelUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
