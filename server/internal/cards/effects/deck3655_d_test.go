package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_d_test.go — Mog, Moogle Warrior from the Terra, Herald of
// Hope deck request (#3655).

const deck3655MogOracle = "f18cce8c-8fb6-49a7-8f79-2529c0ee3ecc"

func TestDeck3655MogDanceMakesAMoogleAndCountersTheMoogles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mog := pushCatalogPermanent(g, me.ID, "Mog, Moogle Warrior", "Legendary Creature — Moogle Warrior", deck3655MogOracle, false)
	myCreature := handCardForTest(me, "My Bear", "Creature — Bear", "")
	theirSpell := handCardForTest(opp, "Their Bolt", "Instant", "")
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()

	advanceToStepOf(t, g, 0, game.StepEnd)
	if triggerOnStack(g, mog) == nil {
		t.Fatal("no Dance trigger at my end step")
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		switch p.ID {
		case me.ID:
			answerMayChoice(t, g, p.ID, true)
			answerChooseCards(t, g, p.ID, myCreature)
		case opp.ID:
			answerMayChoice(t, g, p.ID, true)
			answerChooseCards(t, g, p.ID, theirSpell)
		default:
			answerMayChoice(t, g, p.ID, false)
		}
	}
	passPriorityAroundTable(t, g)

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
