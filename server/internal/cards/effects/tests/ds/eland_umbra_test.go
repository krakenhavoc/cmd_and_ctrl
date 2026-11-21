package testsds

import (
	"testing"

	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Eland Umbra (#3654): +0/+4, and umbra armor.

func TestElandUmbraAddsFourToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AuraCast(t, g, "Eland Umbra", cardse.ElandUmbraOracleID, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 2 || tg != 6 {
		t.Fatalf("enchanted bear is %d/%d, want 2/6", p, tg)
	}
}

func TestElandUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Eland Umbra", cardse.ElandUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
