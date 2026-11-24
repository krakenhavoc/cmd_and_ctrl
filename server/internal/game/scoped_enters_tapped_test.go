package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_enters_tapped_test.go — #6093: ModCreaturesEnterTappedForOpponents,
// Cone of Cold's "until your next turn, creatures your opponents control
// enter tapped". The catalog half is in cards/effects.

func handCreatureFor(g *Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Seats[g.seatOfForTest(owner)].Hand.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: owner, Controller: owner,
	})
	return id
}

func (g *Game) seatOfForTest(id uuid.UUID) int {
	for i, p := range g.Seats {
		if p != nil && p.ID == id {
			return i
		}
	}
	return -1
}

func putOntoBattlefieldTapped(t *testing.T, g *Game, owner, id uuid.UUID) bool {
	t.Helper()
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: owner}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	c, ok := g.battlefieldCardLocked(id)
	if !ok {
		t.Fatal("the creature is not on the battlefield")
	}
	return c.Tapped
}

func TestCreaturesEnterTappedForOpponentsLastsUntilYourNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if !g.CreaturesEnterTappedForOpponentsForEffect(uuid.Nil, me.ID, g.UntilYourNextTurnDuration(me.ID), "test: enter tapped") {
			t.Fatal("the replacement was not registered")
		}
	})
	if !putOntoBattlefieldTapped(t, g, opp.ID, handCreatureFor(g, opp.ID, "Opponent Bear")) {
		t.Error("an opponent's creature entered untapped")
	}
	if putOntoBattlefieldTapped(t, g, me.ID, handCreatureFor(g, me.ID, "My Bear")) {
		t.Error("the controller's own creature entered tapped")
	}
	advanceToStepOfSeat(t, g, 2, StepPrecombatMain)
	if !putOntoBattlefieldTapped(t, g, opp.ID, handCreatureFor(g, opp.ID, "Later Bear")) {
		t.Error("an opponent's creature entered untapped on a later turn before the controller's next turn")
	}
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	if putOntoBattlefieldTapped(t, g, opp.ID, handCreatureFor(g, opp.ID, "Free Bear")) {
		t.Error("the replacement outlived its controller's next turn beginning")
	}
}
