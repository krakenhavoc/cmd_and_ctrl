package testsg

import (
	"testing"

	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Hyena Umbra (#3654): +1/+1 and first strike, and umbra armor.

func TestHyenaUmbraPumpsAndGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AuraCast(t, g, "Hyena Umbra", cardshu.HyenaUmbraOracleID, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
	effectstest.AssertEffectiveKeyword(t, g, bear, "first strike")
}

func TestHyenaUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Hyena Umbra", cardshu.HyenaUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
