package testscr

import (
	"testing"

	cardsdi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/di"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Drake Umbra (#3654): +3/+3 and flying, and umbra armor.

func TestDrakeUmbraPumpsAndGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AuraCast(t, g, "Drake Umbra", cardsdi.DrakeUmbraOracleID, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Fatalf("enchanted bear is %d/%d, want 5/5", p, tg)
	}
	effectstest.AssertEffectiveKeyword(t, g, bear, "flying")
}

func TestDrakeUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Drake Umbra", cardsdi.DrakeUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
