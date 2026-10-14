package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_died_subtypes_test.go — #3592. TurnTally.DiedSubtypes is the
// per-controller count of the creature types that died this turn
// (Brimaz, Blight of Oreskos), read as each creature last existed.

func diedWith(g *Game, player uuid.UUID, subtype string) int {
	var n int
	g.WithWriteLock(func() { n = g.DiedWithSubtypeThisTurn(player, subtype) })
	return n
}

func TestDiedSubtypesCountsPerControllerAndPerType(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	mine := diedPowerPush(g, me, "Phyrexian Token", "Token Artifact Creature — Phyrexian", 1, nil)
	theirs := diedPowerPush(g, opp, "Phyrexian Drone", "Creature — Phyrexian Horror", 2, nil)
	bear := diedPowerPush(g, me, "Bear", "Creature — Bear", 2, nil)
	survivor := diedPowerPush(g, me, "Phyrexian Survivor", "Creature — Phyrexian", 2, nil)

	diedPowerDestroy(t, g, mine, theirs, bear)

	if got := diedWith(g, me, "Phyrexian"); got != 1 {
		t.Errorf("Phyrexians that died under me = %d, want 1 (a token counts; an opponent's and a survivor do not)", got)
	}
	if got := diedWith(g, opp, "phyrexian"); got != 1 {
		t.Errorf("Phyrexians that died under the opponent = %d, want 1 (the type is case-insensitive)", got)
	}
	if got := diedWith(g, me, "Bear"); got != 1 {
		t.Errorf("Bears that died under me = %d, want 1", got)
	}
	if got := diedWith(g, g.Seats[2].ID, "Phyrexian"); got != 0 {
		t.Errorf("a seat nothing died under reads %d, want 0", got)
	}
	_ = survivor
}

func TestDiedSubtypesIsTheControllerItDiedUnderNotTheOwner(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, thief := g.Seats[0].ID, g.Seats[1].ID
	stolen := diedPowerPush(g, owner, "Phyrexian Thrall", "Creature — Phyrexian", 2, nil)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stolen {
				g.Battlefield.Cards[i].Controller = thief
			}
		}
	})

	diedPowerDestroy(t, g, stolen)

	if got := diedWith(g, thief, "Phyrexian"); got != 1 {
		t.Errorf("died under the thief = %d, want 1", got)
	}
	if got := diedWith(g, owner, "Phyrexian"); got != 0 {
		t.Errorf("died under the owner = %d, want 0 (CR 603.10a: the controller as it left)", got)
	}
}

func TestDiedSubtypesCountsAChangelingForEveryCreatureType(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	shifter := diedPowerPush(g, me, "Shifter", "Creature — Shapeshifter", 2, nil, KeywordChangeling)

	diedPowerDestroy(t, g, shifter)

	if got := diedWith(g, me, "Phyrexian"); got != 1 {
		t.Errorf("a changeling is a Phyrexian: got %d, want 1", got)
	}
	if got := diedWith(g, me, "Shapeshifter"); got != 1 {
		t.Errorf("a changeling counts once under its own type: got %d, want 1", got)
	}
	if got := diedWith(g, me, "Forest"); got != 0 {
		t.Errorf("Forest is not a creature type: got %d, want 0", got)
	}
}

func TestDiedSubtypesIgnoresAnythingThatDidNotDieAndStartsOverEachTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	exiled := diedPowerPush(g, me, "Phyrexian A", "Creature — Phyrexian", 1, nil)
	bounced := diedPowerPush(g, me, "Phyrexian B", "Creature — Phyrexian", 1, nil)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(exiled); err != nil {
			t.Fatal(err)
		}
		if err := g.BounceToHandForEffect(bounced); err != nil {
			t.Fatal(err)
		}
	})
	if got := diedWith(g, me, "Phyrexian"); got != 0 {
		t.Errorf("an exile and a bounce are not deaths: got %d, want 0", got)
	}

	dies := diedPowerPush(g, me, "Phyrexian C", "Creature — Phyrexian", 1, nil)
	diedPowerDestroy(t, g, dies)
	if got := diedWith(g, me, "Phyrexian"); got != 1 {
		t.Fatalf("died this turn = %d, want 1", got)
	}
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if got := diedWith(g, me, "Phyrexian"); got != 0 {
		t.Errorf("the tally must start over at the turn boundary: got %d, want 0", got)
	}
}

func TestDiedSubtypesSurvivesCloneTurnTally(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	diedPowerDestroy(t, g, diedPowerPush(g, me, "Phyrexian", "Creature — Phyrexian", 1, nil))
	var cloned TurnTally
	g.WithWriteLock(func() { cloned = cloneTurnTally(g.TurnTally) })
	if got := subtypeTallyCount(cloned.DiedSubtypes, me, "Phyrexian"); got != 1 {
		t.Errorf("the clone lost the tally: got %d, want 1", got)
	}
}
