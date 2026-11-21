package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const magmaHellionOracle = "88f6575d-a6cb-4b2c-90dd-c6239f3c6e4d"

// TestMagmaHellionHasHasteAndTrampleWithAssist is #3204's keyword-only
// creature: Full, assist read off the catalog, and both printed
// keywords on the resolved creature.
func TestMagmaHellionHasHasteAndTrampleWithAssist(t *testing.T) {
	effectstest.AssistCardSpec(t, magmaHellionOracle, "Magma Hellion")

	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Magma Hellion", "Creature — Hellion", magmaHellionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.AssistCardOnBattlefield(g, id) {
		t.Fatal("Magma Hellion did not resolve onto the battlefield")
	}
	got := effectstest.EffectiveAbilities(t, g, id)
	if !effectstest.ContainsString(got, "haste") || !effectstest.ContainsString(got, "trample") {
		t.Errorf("Magma Hellion abilities = %v, want haste and trample", got)
	}
}
