package game

import (
	"testing"

	"github.com/google/uuid"
)

// put_attacking_test.go — #3143. PutOnBattlefieldAttackingForEffect makes a
// permanent already in play an attacker without declaring it one (CR 506.3c),
// after a control change of the same effect.

func putAttackingSteal(t *testing.T, g *Game, id, to uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), id, to, g.UntilEndOfTurnDuration(), "test — steal") {
			t.Fatal("GainControlForEffect refused")
		}
	})
}

// The headline: a creature taken from the defender and put into combat is
// tapped by the caller, attacks the player it was taken from, is never
// announced (no attack trigger), and connects unblocked.
func TestAStolenCreaturePutIntoCombatAttacksWithoutBeingDeclared(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	declared := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	theirs := pushCombatant(t, g, opp, "Ogre", 3, 3)
	declareAttacks(t, g, declared)

	putAttackingSteal(t, g, theirs, me.ID)
	var ok bool
	g.WithWriteLock(func() { ok = g.PutOnBattlefieldAttackingForEffect(theirs, opp.ID) })
	if !ok {
		t.Fatal("PutOnBattlefieldAttackingForEffect refused a creature I control")
	}
	c := findBattlefieldCard(g, theirs)
	if c == nil || c.Controller != me.ID {
		t.Fatalf("the Ogre is not mine: %+v", c)
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("AttackingTarget = %s, want %s", c.AttackingTarget, opp.ID)
	}
	if !g.announcedAttacks[theirs] {
		t.Error("not marked announced: the next lock-in would hand it an attack trigger (CR 506.3c)")
	}
	g.WithWriteLock(func() { g.commitAttackDeclarationLocked() })
	if n := attackEventsFor(g, theirs); n != 0 {
		t.Errorf("%d EventAttack for a creature put into combat, want 0", n)
	}

	before := opp.Life
	passUntilStep(t, g, StepCombatDamage)
	passUntilStep(t, g, StepEndCombat)
	if got := before - opp.Life; got != 5 {
		t.Errorf("defender lost %d life, want 5 (the Bears' 2 and the Ogre's 3)", got)
	}
}

// The control change has not been applied yet when the helper runs; the
// layer pass runs first, so the control change cannot take it out of combat
// afterwards (CR 506.4).
func TestPutIntoCombatSurvivesTheControlChangeBeingApplied(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushCombatant(t, g, opp, "Ogre", 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)

	putAttackingSteal(t, g, theirs, me.ID)
	g.WithWriteLock(func() {
		if !g.PutOnBattlefieldAttackingForEffect(theirs, opp.ID) {
			t.Fatal("refused")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c := findBattlefieldCard(g, theirs); c == nil || c.AttackingTarget != opp.ID {
		t.Errorf("a later layer pass dropped the attack: %+v", c)
	}
}

func TestPutIntoCombatRefusesWhatCannotBeAnAttacker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := pushCombatant(t, g, me, "Bears", 2, 2)
	already := pushCombatant(t, g, me, "Already", 2, 2)
	land := NewCard("Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	g.Battlefield.PushTop(land)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(already, opp.ID); err != nil {
		t.Fatal(err)
	}

	g.WithWriteLock(func() {
		if g.PutOnBattlefieldAttackingForEffect(land.InstanceID, opp.ID) {
			t.Error("a land was put into combat")
		}
		if g.PutOnBattlefieldAttackingForEffect(uuid.New(), opp.ID) {
			t.Error("a card that is not on the battlefield was put into combat")
		}
		if g.PutOnBattlefieldAttackingForEffect(already, opp.ID) {
			t.Error("an attacker that is already attacking was put into combat again")
		}
		if g.PutOnBattlefieldAttackingForEffect(creature, uuid.New()) {
			t.Error("a creature was set attacking something that does not exist")
		}
		if g.PutOnBattlefieldAttackingForEffect(creature, uuid.Nil) {
			t.Error("a creature was set attacking nothing")
		}
	})
	if c := findBattlefieldCard(g, creature); c == nil || c.AttackingTarget != uuid.Nil || g.announcedAttacks[creature] {
		t.Errorf("a refused put changed the creature: %+v", c)
	}
}
