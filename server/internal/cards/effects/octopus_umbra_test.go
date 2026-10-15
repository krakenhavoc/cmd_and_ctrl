package effects

import "testing"

// Octopus Umbra (#3654): base 8/8, an optional attack trigger that taps
// a creature with power 8 or less, and umbra armor.

func TestOctopusUmbraSetsBase88AndTapsOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	octopus := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Octopus Umbra", octopusUmbraOracleID, octopus)
	if p, tg := effectivePower(t, g, octopus), effectiveToughness(t, g, octopus); p != 8 || tg != 8 {
		t.Fatalf("enchanted creature is %d/%d, want 8/8", p, tg)
	}

	victim := seedBear(g, opp.ID)
	declareAttack(t, g, opp.ID, octopus)
	pickIfAsked(t, g, me.ID, victim)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) {
		t.Error("the attack trigger did not tap the chosen creature")
	}
}

func TestOctopusUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Octopus Umbra", octopusUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
