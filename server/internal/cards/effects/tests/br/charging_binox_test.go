package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const chargingBinoxOracle = "0f10f9d1-f9f9-4356-9208-0f21f44bfb97"

// TestChargingBinoxIsAFullAssistTrampler is #3204's vanilla card: the
// catalog entry is Full, reads as having assist, and the creature that
// resolves has trample and nothing else.
func TestChargingBinoxIsAFullAssistTrampler(t *testing.T) {
	effectstest.AssistCardSpec(t, chargingBinoxOracle, "Charging Binox")

	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Charging Binox", "Creature — Beast", chargingBinoxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.AssistCardOnBattlefield(g, id) {
		t.Fatal("Charging Binox did not resolve onto the battlefield")
	}
	got := effectstest.EffectiveAbilities(t, g, id)
	if !effectstest.ContainsString(got, "trample") {
		t.Errorf("Charging Binox abilities = %v, want trample", got)
	}
	if effectstest.ContainsString(got, "haste") || effectstest.ContainsString(got, "flying") {
		t.Errorf("Charging Binox abilities = %v, want trample only", got)
	}
}
