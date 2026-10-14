package game

import (
	"testing"

	"github.com/google/uuid"
)

// combat_damage_hold_test.go — #3665: a combat damage step that asks a
// damage-assignment prompt deals none of its damage, and runs no
// state-based actions, until every prompt it asked is answered
// (CR 510.1, CR 510.2).

// openDamageAssignments is every open damage_assignment prompt.
func openDamageAssignments(g *Game) []*PendingChoice {
	var out []*PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceDamageAssignment {
			out = append(out, c)
		}
	}
	return out
}

// TestHeldCombatDamageAttackerLivesUntilItAssigns is the issue's table:
// a 2/2 blocked by two 1/1s. The blockers' damage used to land, and the
// attacker die, while its controller was still being asked to divide
// its damage. Now nothing is dealt until the answer, and the answer
// deals it all, so all three die in one sweep.
func TestHeldCombatDamageAttackerLivesUntilItAssigns(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	solemn := pushCombatant(t, g, me, "Solemn", 2, 2)
	servo1 := pushCombatant(t, g, opp, "Servo", 1, 1)
	servo2 := pushCombatant(t, g, opp, "Servo", 1, 1)
	declareAttacks(t, g, solemn)
	if err := block(t, g, servo1, solemn, servo2, solemn); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)

	prompts := openDamageAssignments(g)
	if len(prompts) != 1 {
		t.Fatalf("%d damage_assignment prompts, want 1", len(prompts))
	}
	if got := damageOn(g, solemn); got != 0 {
		t.Fatalf("the attacker has %d damage (-1 = gone) while its assignment is open, want 0 — CR 510.2", got)
	}
	if damageOn(g, servo1) != 0 || damageOn(g, servo2) != 0 {
		t.Fatal("blockers were dealt damage before the assignment was answered")
	}

	seq := lastSeq(g)
	if err := g.ResolveDamageAssignment(prompts[0].ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: servo1, Amount: 1}, {BlockerID: servo2, Amount: 1}}, 0); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}
	for _, id := range []uuid.UUID{solemn, servo1, servo2} {
		if findBattlefieldCard(g, id) != nil {
			t.Errorf("%s survived lethal combat damage", id)
		}
	}
	// One sweep: every combat damage event comes before the first death.
	dealt, died := 0, false
	for _, ev := range g.Events {
		if ev.Seq <= seq {
			continue
		}
		switch {
		case ev.Kind == EventDealDamage && ev.Combat:
			dealt++
			if died {
				t.Errorf("combat damage event %d came after a creature died", ev.Seq)
			}
		case ev.Kind == EventZoneMove && ev.OldZone == ZoneBattlefield:
			died = true
		}
	}
	if dealt != 4 {
		t.Errorf("%d combat damage events after the answer, want 4", dealt)
	}
}

// TestHeldCombatDamageWaitsForEveryPrompt: two multi-blocked attackers
// and an unblocked one. Answering the first prompt deals nothing — not
// even the unblocked attacker's damage — and the second deals it all.
func TestHeldCombatDamageWaitsForEveryPrompt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCombatant(t, g, me, "A", 2, 2)
	b := pushCombatant(t, g, me, "B", 2, 2)
	free := pushCombatant(t, g, me, "Free", 3, 3)
	a1 := pushCombatant(t, g, opp, "A1", 1, 1)
	a2 := pushCombatant(t, g, opp, "A2", 1, 1)
	b1 := pushCombatant(t, g, opp, "B1", 1, 1)
	b2 := pushCombatant(t, g, opp, "B2", 1, 1)
	declareAttacks(t, g, a, b, free)
	if err := block(t, g, a1, a, a2, a, b1, b, b2, b); err != nil {
		t.Fatal(err)
	}
	life := opp.Life
	advanceIntoStep(t, g, StepCombatDamage)

	prompts := openDamageAssignments(g)
	if len(prompts) != 2 {
		t.Fatalf("%d damage_assignment prompts, want 2", len(prompts))
	}
	answer := func(c *PendingChoice) {
		t.Helper()
		ids := c.DamageAssignment.BlockerIDs
		if err := g.ResolveDamageAssignment(c.ID, me.ID,
			[]DamageAssignmentEntry{{BlockerID: ids[0], Amount: 1}, {BlockerID: ids[1], Amount: 1}}, 0); err != nil {
			t.Fatalf("ResolveDamageAssignment: %v", err)
		}
	}
	answer(prompts[0])
	if opp.Life != life {
		t.Errorf("defender at %d after the first answer, want %d: the step is still being assigned", opp.Life, life)
	}
	for _, id := range []uuid.UUID{a, b, a1, a2, b1, b2} {
		if got := damageOn(g, id); got != 0 {
			t.Errorf("%s has %d damage (-1 = gone) with a prompt still open, want 0", id, got)
		}
	}
	answer(prompts[1])
	if opp.Life != life-3 {
		t.Errorf("defender at %d, want %d", opp.Life, life-3)
	}
	for _, id := range []uuid.UUID{a, b, a1, a2, b1, b2} {
		if findBattlefieldCard(g, id) != nil {
			t.Errorf("%s survived lethal combat damage", id)
		}
	}
	if findBattlefieldCard(g, free) == nil {
		t.Error("the unblocked attacker died")
	}
}

// TestHeldCombatDamageTrampleCountsDamageAssignedThisStep: CR 702.19b's
// lethal damage takes into account "damage from other creatures that's
// being assigned during the same combat damage step". That damage is
// held, not marked, while the prompt is open, so the check reads it
// from the hold.
func TestHeldCombatDamageTrampleCountsDamageAssignedThisStep(t *testing.T) {
	withMultiBlockStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := pushCombatant(t, g, me, "One", 3, 3)
	trampler := pushCombatant(t, g, me, "Trampler", 4, 4, "trample")
	wall := pushMultiBlocker(t, g, opp, "Wall", blockTwoOracle, 0, 4)
	chump := pushCombatant(t, g, opp, "Chump", 1, 1)
	declareAttacks(t, g, one, trampler)
	if err := block(t, g, wall, one, wall, trampler, chump, trampler); err != nil {
		t.Fatal(err)
	}
	life := opp.Life
	advanceIntoStep(t, g, StepCombatDamage)

	prompts := openDamageAssignments(g)
	if len(prompts) != 1 {
		t.Fatalf("%d damage_assignment prompts, want 1", len(prompts))
	}
	// One's 3 is assigned to the wall, so 1 more is lethal for it.
	if err := g.ResolveDamageAssignment(prompts[0].ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: wall, Amount: 1}, {BlockerID: chump, Amount: 1}}, 2); err != nil {
		t.Fatalf("trample past a wall already assigned 3 of its 4: %v", err)
	}
	if opp.Life != life-2 {
		t.Errorf("defender at %d, want %d", opp.Life, life-2)
	}
	if findBattlefieldCard(g, wall) != nil || findBattlefieldCard(g, chump) != nil {
		t.Error("a blocker survived lethal damage")
	}
}

// TestHeldCombatDamageIsNotARestorePoint: the hold is not carried by
// the persisted snapshot, so a table waiting on it is not a restore
// point; undo carries it.
func TestHeldCombatDamageIsNotARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Atk", 2, 2)
	b1 := pushCombatant(t, g, opp, "B1", 1, 1)
	b2 := pushCombatant(t, g, opp, "B2", 1, 1)
	declareAttacks(t, g, atk)
	if err := block(t, g, b1, atk, b2, atk); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if g.heldCombatDamage == nil {
		t.Fatal("no damage held behind the prompt")
	}
	if snap := g.CaptureSnapshot(); snap.Restorable() {
		t.Error("a table holding combat damage behind its prompts was written as a restore point")
	}
	clone := g.Clone()
	if clone.heldCombatDamage != g.heldCombatDamage {
		t.Error("Clone dropped the held combat damage")
	}
}

// TestHeldCombatDamageSettlesWhenTheChooserLeaves: the attacker's
// controller concedes with the prompt open. Their prompt and their
// attacker go (CR 800.4a), and the hold is settled rather than left
// behind.
func TestHeldCombatDamageSettlesWhenTheChooserLeaves(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Atk", 2, 2)
	b1 := pushCombatant(t, g, opp, "B1", 1, 1)
	b2 := pushCombatant(t, g, opp, "B2", 1, 1)
	declareAttacks(t, g, atk)
	if err := block(t, g, b1, atk, b2, atk); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	g.WithWriteLock(func() {
		if g.heldCombatDamage != nil {
			t.Error("the hold outlived the prompt it waited on")
		}
		if len(openDamageAssignments(g)) != 0 {
			t.Error("the departed player's damage_assignment prompt is still open")
		}
	})
	if damageOn(g, b1) != 0 || damageOn(g, b2) != 0 {
		t.Error("the departed attacker's damage was dealt")
	}
}
