package testsds

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Eel Umbra (#3654): flash, +1/+1, and umbra armor.

func TestEelUmbraPrintsFlashAndPumps(t *testing.T) {
	spec, ok := effects.Lookup(cardse.EelUmbraOracleID)
	if !ok || !effectstest.HasKeyword(spec.PrintedKeywords, "flash") {
		t.Fatalf("Eel Umbra must print flash: %+v", spec.PrintedKeywords)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AuraCast(t, g, "Eel Umbra", cardse.EelUmbraOracleID, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
}

func TestEelUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Eel Umbra", cardse.EelUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
