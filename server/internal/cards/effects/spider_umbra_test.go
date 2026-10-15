package effects

import "testing"

// Spider Umbra (#3654): +1/+1 and reach, and umbra armor.

func TestSpiderUmbraPumpsAndGivesReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Spider Umbra", spiderUmbraOracleID, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
	assertEffectiveKeyword(t, g, bear, "reach")
}

func TestSpiderUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Spider Umbra", spiderUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
