package effects

import "testing"

const magmaHellionOracle = "88f6575d-a6cb-4b2c-90dd-c6239f3c6e4d"

// TestMagmaHellionHasHasteAndTrampleWithAssist is #3204's keyword-only
// creature: Full, assist read off the catalog, and both printed
// keywords on the resolved creature.
func TestMagmaHellionHasHasteAndTrampleWithAssist(t *testing.T) {
	assistCardSpec(t, magmaHellionOracle, "Magma Hellion")

	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Magma Hellion", "Creature — Hellion", magmaHellionOracle, nil)
	passPriorityAroundTable(t, g)

	if !assistCardOnBattlefield(g, id) {
		t.Fatal("Magma Hellion did not resolve onto the battlefield")
	}
	got := effectiveAbilities(t, g, id)
	if !containsString(got, "haste") || !containsString(got, "trample") {
		t.Errorf("Magma Hellion abilities = %v, want haste and trample", got)
	}
}
