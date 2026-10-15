package effects

import "testing"

// Snake Umbra (#3654): +1/+1, a may-draw on damage to an opponent, and
// umbra armor.

func TestSnakeUmbraPumpsAndDrawsOnDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := umbraCreature(g, me.ID, 2, 2)
	auraCast(t, g, "Snake Umbra", snakeUmbraOracleID, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tg)
	}

	// Damage to my own face draws nothing.
	emitNoncombatDamage(g, bear, me.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("Snake Umbra asked about damage dealt to its own controller")
	}

	before := me.Hand.Size()
	emitNoncombatDamage(g, bear, opp.ID, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestSnakeUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Snake Umbra", snakeUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
