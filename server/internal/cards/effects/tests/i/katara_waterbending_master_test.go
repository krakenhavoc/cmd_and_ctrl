package testsi

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kataraWaterbendingMasterOracle = "2e3af413-6706-4023-871a-e5e49da5beab"

// A spell cast on an opponent's turn is an experience counter; one cast
// on your own turn is not; an opponent casting is not.
func TestKataraWaterbendingMasterExperienceOnlyForYourOffTurnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)

	effectstest.AangAdvanceToMain(t, g, 0)
	effectstest.KataraCastInstant(t, g, me)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Fatalf("experience = %d after casting on an opponent's turn, want 1", got)
	}
	effectstest.KataraCastInstant(t, g, opp)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Fatalf("experience = %d after an OPPONENT cast, want still 1", got)
	}

	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.KataraCastInstant(t, g, me)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Errorf("experience = %d after casting on your own turn, want still 1", got)
	}
}

// Attacking with two experience counters: "yes" draws two, then the
// linked discard is the player's pick.
func TestKataraWaterbendingMasterAttackDrawsPerCounterThenDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	katara := effectstest.B12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)
	if err := g.AddPlayerCounterForEffect(me.ID, game.CounterExperience, 2); err != nil {
		t.Fatalf("seed experience: %v", err)
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	before := me.Hand.Size()

	effectstest.DeclareAttack(t, g, opp.ID, katara)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if got := me.Hand.Size(); got != before+2 {
		t.Fatalf("hand %d -> %d, want +2 before the discard", before, got)
	}
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Fatalf("the linked discard owes %d, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want net +1", before, got)
	}
}

// Declining draws nothing and discards nothing.
func TestKataraWaterbendingMasterDeclineDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	katara := effectstest.B12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)
	if err := g.AddPlayerCounterForEffect(me.ID, game.CounterExperience, 3); err != nil {
		t.Fatalf("seed experience: %v", err)
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	before := me.Hand.Size()

	effectstest.DeclareAttack(t, g, opp.ID, katara)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d after declining, want unchanged", before, got)
	}
	if got := effectstest.DiscardOwed(g, me.ID); got != 0 {
		t.Errorf("a discard is owed after declining: %d", got)
	}
}

// With no experience counters there is nothing to draw: no question,
// no discard.
func TestKataraWaterbendingMasterNoCountersAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	katara := effectstest.B12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)
	effectstest.AangAdvanceToMain(t, g, 1)
	before := me.Hand.Size()

	effectstest.DeclareAttack(t, g, opp.ID, katara)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Errorf("asked to draw with zero experience counters")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d, want unchanged", before, got)
	}
}
