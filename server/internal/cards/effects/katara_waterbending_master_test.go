package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kataraWaterbendingMasterOracle = "2e3af413-6706-4023-871a-e5e49da5beab"

func kataraCastInstant(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Impulse", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// A spell cast on an opponent's turn is an experience counter; one cast
// on your own turn is not; an opponent casting is not.
func TestKataraWaterbendingMasterExperienceOnlyForYourOffTurnSpells(t *testing.T) {
	g := newCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)

	aangAdvanceToMain(t, g, 0)
	kataraCastInstant(t, g, me)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Fatalf("experience = %d after casting on an opponent's turn, want 1", got)
	}
	kataraCastInstant(t, g, opp)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Fatalf("experience = %d after an OPPONENT cast, want still 1", got)
	}

	aangAdvanceToMain(t, g, 1)
	kataraCastInstant(t, g, me)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Errorf("experience = %d after casting on your own turn, want still 1", got)
	}
}

// Attacking with two experience counters: "yes" draws two, then the
// linked discard is the player's pick.
func TestKataraWaterbendingMasterAttackDrawsPerCounterThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	katara := b12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)
	if err := g.AddPlayerCounterForEffect(me.ID, game.CounterExperience, 2); err != nil {
		t.Fatalf("seed experience: %v", err)
	}
	aangAdvanceToMain(t, g, 1)
	before := me.Hand.Size()

	declareAttack(t, g, opp.ID, katara)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if got := me.Hand.Size(); got != before+2 {
		t.Fatalf("hand %d -> %d, want +2 before the discard", before, got)
	}
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("the linked discard owes %d, want 1", got)
	}
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want net +1", before, got)
	}
}

// Declining draws nothing and discards nothing.
func TestKataraWaterbendingMasterDeclineDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	katara := b12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)
	if err := g.AddPlayerCounterForEffect(me.ID, game.CounterExperience, 3); err != nil {
		t.Fatalf("seed experience: %v", err)
	}
	aangAdvanceToMain(t, g, 1)
	before := me.Hand.Size()

	declareAttack(t, g, opp.ID, katara)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d after declining, want unchanged", before, got)
	}
	if got := discardOwed(g, me.ID); got != 0 {
		t.Errorf("a discard is owed after declining: %d", got)
	}
}

// With no experience counters there is nothing to draw: no question,
// no discard.
func TestKataraWaterbendingMasterNoCountersAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	katara := b12Push(g, me.ID, "Katara, Waterbending Master", "Legendary Creature — Human Warrior Ally", kataraWaterbendingMasterOracle, 1, 3)
	aangAdvanceToMain(t, g, 1)
	before := me.Hand.Size()

	declareAttack(t, g, opp.ID, katara)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Errorf("asked to draw with zero experience counters")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d, want unchanged", before, got)
	}
}
