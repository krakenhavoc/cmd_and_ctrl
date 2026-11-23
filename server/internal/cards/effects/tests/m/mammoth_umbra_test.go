package testsm

import (
	"testing"

	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Mammoth Umbra (#3654): +3/+3 and vigilance, and umbra armor.

func TestMammothUmbraPumpsAndGivesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AuraCast(t, g, "Mammoth Umbra", cardslu.MammothUmbraOracleID, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Fatalf("enchanted bear is %d/%d, want 5/5", p, tg)
	}
	effectstest.AssertEffectiveKeyword(t, g, bear, "vigilance")
}

func TestMammothUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Mammoth Umbra", cardslu.MammothUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
