package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDeck3655MogDanceMakesAMoogleAndCountersTheMoogles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mog := effectstest.PushCatalogPermanent(g, me.ID, "Mog, Moogle Warrior", "Legendary Creature — Moogle Warrior", effectstest.Deck3655MogOracle, false)
	myCreature := effectstest.HandCardForTest(me, "My Bear", "Creature — Bear", "")
	theirSpell := effectstest.HandCardForTest(opp, "Their Bolt", "Instant", "")
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()

	effectstest.AdvanceToStepOf(t, g, 0, game.StepEnd)
	if effectstest.TriggerOnStack(g, mog) == nil {
		t.Fatal("no Dance trigger at my end step")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		switch p.ID {
		case me.ID:
			effectstest.AnswerMayChoice(t, g, p.ID, true)
			effectstest.AnswerChooseCards(t, g, p.ID, myCreature)
		case opp.ID:
			effectstest.AnswerMayChoice(t, g, p.ID, true)
			effectstest.AnswerChooseCards(t, g, p.ID, theirSpell)
		default:
			effectstest.AnswerMayChoice(t, g, p.ID, false)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != myHand || opp.Hand.Size() != theirHand {
		t.Errorf("hands %d and %d, want %d and %d (one discarded, one drawn)", me.Hand.Size(), opp.Hand.Size(), myHand, theirHand)
	}
	var moogles, counters int
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Moogle") {
			moogles++
			counters += c.Counters[game.CounterPlusOne]
		}
	}
	if moogles != 2 {
		t.Errorf("Moogles I control = %d, want Mog and one token", moogles)
	}
	if counters != 2 {
		t.Errorf("+1/+1 counters across my Moogles = %d, want 2 (a noncreature card was discarded)", counters)
	}
}
