package testso

import (
	"testing"

	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Octopus Umbra (#3654): base 8/8, an optional attack trigger that taps
// a creature with power 8 or less, and umbra armor.

func TestOctopusUmbraSetsBase88AndTapsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	octopus := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AuraCast(t, g, "Octopus Umbra", cardso.OctopusUmbraOracleID, octopus)
	if p, tg := effectstest.EffectivePower(t, g, octopus), effectstest.EffectiveToughness(t, g, octopus); p != 8 || tg != 8 {
		t.Fatalf("enchanted creature is %d/%d, want 8/8", p, tg)
	}

	victim := effectstest.SeedBear(g, opp.ID)
	effectstest.DeclareAttack(t, g, opp.ID, octopus)
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, victim) {
		t.Error("the attack trigger did not tap the chosen creature")
	}
}

func TestOctopusUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Octopus Umbra", cardso.OctopusUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
