package game

import (
	"testing"

	"github.com/google/uuid"
)

// seam6098_combat_records_test.go — #6098: two combat records.
//
//   - EventLTB's AttackingAlone / BlockingAlone (CR 506.5, exit_alone.go),
//     judged against the whole simultaneous event.
//   - The CR 508.1d record of which creatures had to attack this combat
//     (had_to_attack.go).

// A lone attacker that dies reports attacking alone; so does its lone
// blocker report blocking alone.
func TestLTBAttackingAloneAndBlockingAlone(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	blockAfterLockIn(t, g, attacker, blocker)

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(blocker); err != nil {
			t.Fatalf("destroy blocker: %v", err)
		}
	})
	if ev := ltbSince(t, g, seq, blocker); !ev.BlockingAlone || ev.AttackingAlone {
		t.Errorf("blocker: BlockingAlone=%v AttackingAlone=%v, want true/false", ev.BlockingAlone, ev.AttackingAlone)
	}
	seq = lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(attacker); err != nil {
			t.Fatalf("destroy attacker: %v", err)
		}
	})
	if ev := ltbSince(t, g, seq, attacker); !ev.AttackingAlone || ev.BlockingAlone {
		t.Errorf("attacker: AttackingAlone=%v BlockingAlone=%v, want true/false", ev.AttackingAlone, ev.BlockingAlone)
	}
}

// Two attackers: one destroyed on its own was not attacking alone. A
// creature not in combat is neither.
func TestLTBNotAloneWithAnotherAttacker(t *testing.T) {
	g := newActiveGame(t)
	a1 := pushCombatant(t, g, g.Seats[0], "A1", 2, 2)
	a2 := pushCombatant(t, g, g.Seats[0], "A2", 2, 2)
	bystander := pushCombatant(t, g, g.Seats[1], "Bystander", 1, 1)
	declareAttacks(t, g, a1, a2)

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(a1); err != nil {
			t.Fatalf("destroy: %v", err)
		}
		if err := g.DestroyPermanentForEffect(bystander); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if ev := ltbSince(t, g, seq, a1); ev.AttackingAlone {
		t.Error("an attacker beside another attacker reported attacking alone")
	}
	if ev := ltbSince(t, g, seq, bystander); ev.AttackingAlone || ev.BlockingAlone {
		t.Error("a creature out of combat reported attacking or blocking alone")
	}
	// The survivor, now the only attacker, is attacking alone when it goes.
	seq = lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(a2); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if ev := ltbSince(t, g, seq, a2); !ev.AttackingAlone {
		t.Error("the last attacker left did not report attacking alone")
	}
}

// Two attackers destroyed in ONE event: neither was attacking alone, even
// though the second leaves after the first has gone (CR 506.5, 603.10a).
func TestLTBSimultaneousExitIsNotAlone(t *testing.T) {
	g := newActiveGame(t)
	a1 := pushCombatant(t, g, g.Seats[0], "A1", 2, 2)
	a2 := pushCombatant(t, g, g.Seats[0], "A2", 2, 2)
	declareAttacks(t, g, a1, a2)

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if n := g.DestroyPermanentsForEffect([]uuid.UUID{a1, a2}); n != 2 {
			t.Fatalf("destroyed %d, want 2", n)
		}
	})
	for _, id := range []uuid.UUID{a1, a2} {
		if ev := ltbSince(t, g, seq, id); ev.AttackingAlone {
			t.Errorf("%s reported attacking alone in a two-creature wipe", id)
		}
	}
}

// A goaded creature had to attack; one with no requirement did not; and a
// creature goaded only after the declaration never had to. The record goes
// with the combat.
func TestHadToAttackThisCombatRecordsTheDeclaration(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goaded := pushCombatant(t, g, me, "Goaded", 2, 2)
	free := pushCombatant(t, g, me, "Free", 2, 2)
	late := pushCombatant(t, g, me, "Late", 2, 2)
	goad(t, g, goaded, opp.ID)
	if g.HadToAttackThisCombatForEffect(goaded) {
		t.Fatal("recorded before any declaration")
	}
	declareAttacks(t, g, goaded, free, late)
	goad(t, g, late, opp.ID)
	if !g.HadToAttackThisCombatForEffect(goaded) {
		t.Error("the goaded attacker is not recorded as having had to attack")
	}
	if g.HadToAttackThisCombatForEffect(free) {
		t.Error("an attacker with no requirement is recorded as having had to attack")
	}
	if g.HadToAttackThisCombatForEffect(late) {
		t.Error("a creature goaded after the declaration is recorded as having had to attack")
	}
	// The record survives a clone and a snapshot round trip.
	var clone *Game
	g.WithWriteLock(func() { clone = g.cloneLocked() })
	if !clone.HadToAttackThisCombatForEffect(goaded) {
		t.Error("the undo clone dropped the record")
	}
	_, restored := roundTrip(t, g)
	if !restored.HadToAttackThisCombatForEffect(goaded) {
		t.Error("the snapshot round trip dropped the record")
	}
	passUntilStep(t, g, StepPostcombatMain)
	if g.HadToAttackThisCombatForEffect(goaded) {
		t.Error("the record outlived the combat")
	}
}

// A plain "attacks each combat if able" requirement counts too, and a
// creature that leaves the battlefield is a new object that never had to
// attack (CR 400.7).
func TestHadToAttackThisCombatAttacksEachCombatAndForgetsOnExit(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	must := mustAttackCreature(t, g, me, 1)
	declareAttacks(t, g, must)
	if !g.HadToAttackThisCombatForEffect(must) {
		t.Fatal("a creature that attacks each combat if able is not recorded")
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(must); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if g.HadToAttackThisCombatForEffect(must) {
		t.Error("the record outlived the object")
	}
}
