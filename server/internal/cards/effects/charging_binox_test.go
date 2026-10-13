package effects

import "testing"

const chargingBinoxOracle = "0f10f9d1-f9f9-4356-9208-0f21f44bfb97"

// TestChargingBinoxIsAFullAssistTrampler is #3204's vanilla card: the
// catalog entry is Full, reads as having assist, and the creature that
// resolves has trample and nothing else.
func TestChargingBinoxIsAFullAssistTrampler(t *testing.T) {
	assistCardSpec(t, chargingBinoxOracle, "Charging Binox")

	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Charging Binox", "Creature — Beast", chargingBinoxOracle, nil)
	passPriorityAroundTable(t, g)

	if !assistCardOnBattlefield(g, id) {
		t.Fatal("Charging Binox did not resolve onto the battlefield")
	}
	got := effectiveAbilities(t, g, id)
	if !containsString(got, "trample") {
		t.Errorf("Charging Binox abilities = %v, want trample", got)
	}
	if containsString(got, "haste") || containsString(got, "flying") {
		t.Errorf("Charging Binox abilities = %v, want trample only", got)
	}
}
