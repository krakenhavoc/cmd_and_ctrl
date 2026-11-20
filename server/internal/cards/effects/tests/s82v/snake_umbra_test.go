package testss82v

import (
	"testing"

	cardssh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sh"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Snake Umbra (#3654): +1/+1, a may-draw on damage to an opponent, and
// umbra armor.

func TestSnakeUmbraPumpsAndDrawsOnDamageToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AuraCast(t, g, "Snake Umbra", cardssh.SnakeUmbraOracleID, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("enchanted bear is %d/%d, want 3/3", p, tg)
	}

	// Damage to my own face draws nothing.
	effectstest.EmitNoncombatDamage(g, bear, me.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("Snake Umbra asked about damage dealt to its own controller")
	}

	before := me.Hand.Size()
	effectstest.EmitNoncombatDamage(g, bear, opp.ID, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestSnakeUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Snake Umbra", cardssh.SnakeUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
