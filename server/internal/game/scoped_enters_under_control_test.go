package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_enters_under_control_test.go — #6260, Gather Specimens: "If a
// creature would enter the battlefield under an opponent's control this
// turn, it enters under your control instead." CR 614.12 (a replacement
// that modifies how a permanent enters), CR 110.2a ("unless the effect
// states otherwise"), CR 616.1b (applied before the window's others).

func gatherSpecimens(t *testing.T, g *Game, me uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.CreaturesEnterUnderYourControlThisTurnForEffect(uuid.Nil, me, "test: Gather Specimens") {
			t.Fatal("the replacement was not registered")
		}
	})
}

func battlefieldControllerForTest(t *testing.T, g *Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var out uuid.UUID
	g.ReadSnapshot(func() {
		c, ok := g.battlefieldCardLocked(id)
		if !ok {
			t.Fatal("the creature is not on the battlefield")
		}
		out = c.Controller
	})
	return out
}

// An opponent's creature spell resolves under the caster of Gather
// Specimens' control.
func TestGatherSpecimensTakesAnOpponentsCreatureSpell(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, g.Turn.ActiveSeat, StepPrecombatMain)
	opp := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gatherSpecimens(t, g, me.ID)
	item := castSpellForControlTest(t, g, opp, "Creature — Bear")
	resolveTop(t, g)
	if got := battlefieldControllerForTest(t, g, item.SourceCardID); got != me.ID {
		t.Fatalf("the opponent's creature entered under %v, want the Gather Specimens player", got)
	}
}

// A creature put onto the battlefield by an opponent (a reanimation, a
// flicker) is taken too; my own creature and an opponent's noncreature
// permanent are not; and the effect ends with the turn.
func TestGatherSpecimensAnyEntryCreaturesOnlyThisTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gatherSpecimens(t, g, me.ID)

	bear := handCreatureFor(g, opp.ID, "Opponent Bear")
	putOntoBattlefieldTapped(t, g, opp.ID, bear)
	if got := battlefieldControllerForTest(t, g, bear); got != me.ID {
		t.Errorf("an opponent's creature put onto the battlefield entered under %v", got)
	}
	mine := handCreatureFor(g, me.ID, "My Bear")
	putOntoBattlefieldTapped(t, g, me.ID, mine)
	if got := battlefieldControllerForTest(t, g, mine); got != me.ID {
		t.Errorf("my own creature entered under %v", got)
	}
	rock := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	})
	putOntoBattlefieldTapped(t, g, opp.ID, rock)
	if got := battlefieldControllerForTest(t, g, rock); got != opp.ID {
		t.Errorf("an opponent's artifact entered under %v, want its owner", got)
	}

	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	later := handCreatureFor(g, opp.ID, "Later Bear")
	putOntoBattlefieldTapped(t, g, opp.ID, later)
	if got := battlefieldControllerForTest(t, g, later); got != opp.ID {
		t.Errorf("the replacement outlived its turn: the creature entered under %v", got)
	}
}
