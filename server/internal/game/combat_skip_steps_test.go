package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// CR 508.8 (#2899): "If no creatures are declared as attackers or put
// onto the battlefield attacking, skip the declare blockers step and
// the combat damage step." The steps are walked through, not entered
// (ADR 0059's 2026-10-12 amendment), so these tests read what BEGAN —
// EventStepBegan is the event "at the beginning of <step>" triggers
// are harvested from — not just where the cursor ended up.

// stepsBegunSince lists the steps that began after event sequence
// `since`, in order.
func stepsBegunSince(g *Game, since uint64) []Step {
	var out []Step
	for _, ev := range g.Events {
		if ev.Seq > since && ev.Kind == EventStepBegan {
			out = append(out, ev.Step)
		}
	}
	return out
}

func stepsContain(steps []Step, want Step) bool {
	for _, s := range steps {
		if s == want {
			return true
		}
	}
	return false
}

// eventHighWater is the Seq of the newest event, so a test can ask
// about "everything after this point".
func eventHighWater(g *Game) uint64 {
	var hw uint64
	for _, ev := range g.Events {
		if ev.Seq > hw {
			hw = ev.Seq
		}
	}
	return hw
}

// walkPastCombat advances to the postcombat main phase.
func walkPastCombat(t *testing.T, g *Game) {
	t.Helper()
	advanceIntoStep(t, g, StepPostcombatMain)
}

func TestNoAttackersSkipsDeclareBlockersAndCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	mark := eventHighWater(g)
	next, err := g.AdvanceStep()
	if err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if next.Step != StepEndCombat {
		t.Fatalf("after declare_attackers with no attackers the cursor is on %s, want end_combat", next.Step)
	}
	walkPastCombat(t, g)
	began := stepsBegunSince(g, mark)
	for _, s := range []Step{StepDeclareBlockers, StepFirstStrikeDamage, StepCombatDamage} {
		if stepsContain(began, s) {
			t.Errorf("%s began in a combat nothing attacked in (began: %v)", s, began)
		}
	}
	if !stepsContain(began, StepEndCombat) {
		t.Errorf("end_combat did not begin (began: %v) — CR 508.8 names two steps, not three", began)
	}
}

// The same through the priority loop: both seats passing in
// declare_attackers moves to the end of combat, and nobody is handed
// priority in a step that is not there.
func TestNoAttackersPassingPriorityLandsOnEndCombat(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for i := 0; i < 8 && g.Turn.Step == StepDeclareAttackers; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Turn.Step != StepEndCombat {
		t.Fatalf("step %s after the table passed in declare_attackers, want end_combat", g.Turn.Step)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("priority holder %d at end_combat, want the active seat %d", g.Turn.PriorityHolder, g.Turn.ActiveSeat)
	}
}

func TestAnAttackerKeepsDeclareBlockersAndCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	mark := eventHighWater(g)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	walkPastCombat(t, g)
	began := stepsBegunSince(g, mark)
	for _, s := range []Step{StepDeclareBlockers, StepCombatDamage, StepEndCombat} {
		if !stepsContain(began, s) {
			t.Errorf("%s did not begin in a combat with an attacker (began: %v)", s, began)
		}
	}
	if g.Seats[1].Life != 40-3 {
		t.Errorf("defender at %d life, want 37: the combat damage step ran", g.Seats[1].Life)
	}
}

// An attacker that is gone before blockers (killed, or removed from
// combat in the declare attackers step's priority window) WAS
// declared: CR 508.8 is about the declaration, so the steps stay.
func TestAnAttackerRemovedFromCombatStillKeepsTheSteps(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	g.WithWriteLock(func() {
		g.commitAttackDeclarationLocked()
		// Out of combat without leaving the battlefield.
		findBattlefieldCard(g, attacker).AttackingTarget = uuid.Nil
	})
	mark := eventHighWater(g)
	walkPastCombat(t, g)
	began := stepsBegunSince(g, mark)
	if !stepsContain(began, StepDeclareBlockers) || !stepsContain(began, StepCombatDamage) {
		t.Errorf("an attacker that was declared and then removed skipped the steps (began: %v)", began)
	}
}

// A creature put onto the battlefield attacking is never DECLARED
// (CR 506.3c) but CR 508.8 counts it: with nothing else attacking the
// steps still happen.
func TestPutOntoBattlefieldAttackingKeepsTheSteps(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ninja := pushHandCreature(t, g, me, "Ninja", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(ninja, HandEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			Attacking:  opp.ID,
		}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	if g.announcedAttacks[ninja] != true {
		t.Fatal("setup: the entry did not mark the permanent announced")
	}
	mark := eventHighWater(g)
	walkPastCombat(t, g)
	began := stepsBegunSince(g, mark)
	if !stepsContain(began, StepDeclareBlockers) || !stepsContain(began, StepCombatDamage) {
		t.Errorf("a creature put onto the battlefield attacking did not keep the steps (began: %v)", began)
	}
	if opp.Life != 40-2 {
		t.Errorf("defender at %d life, want 38: the entered creature was unblocked and dealt damage", opp.Life)
	}
}

// A creature put onto the battlefield "attacking" a defender that is
// no longer attackable is not attacking, so it does not keep the steps.
func TestPutOntoBattlefieldAttackingAnUnattackableDefenderDoesNotKeepTheSteps(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ninja := pushHandCreature(t, g, me, "Ninja", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(ninja, HandEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			Attacking:  uuid.New(),
		}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	mark := eventHighWater(g)
	walkPastCombat(t, g)
	if began := stepsBegunSince(g, mark); stepsContain(began, StepDeclareBlockers) || stepsContain(began, StepCombatDamage) {
		t.Errorf("a creature that attacks nothing kept the steps (began: %v)", began)
	}
}

// An additional combat phase asks again: the combat an effect added
// before the template's had an attacker and the template's does not,
// so only the second skips.
func TestAdditionalCombatWithNoAttackersSkipsTheSteps(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	advanceIntoStep(t, g, StepPrecombatMain)
	addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	mark := eventHighWater(g)
	// First combat: with the attacker.
	advanceIntoStep(t, g, StepEndCombat)
	first := stepsBegunSince(g, mark)
	if !stepsContain(first, StepDeclareBlockers) || !stepsContain(first, StepCombatDamage) {
		t.Fatalf("the first combat lost its steps (began: %v)", first)
	}
	// The template combat that follows: nobody attacks in it.
	mark = eventHighWater(g)
	advanceIntoStep(t, g, StepEnd)
	second := stepsBegunSince(g, mark)
	if stepsContain(second, StepDeclareBlockers) || stepsContain(second, StepCombatDamage) {
		t.Errorf("the second combat, with no attackers, kept the steps (began: %v)", second)
	}
	if !stepsContain(second, StepBeginCombat) || !stepsContain(second, StepEndCombat) {
		t.Errorf("the second combat lost its beginning or end (began: %v)", second)
	}
}

// The record travels with an undo clone and the persisted snapshot: an
// attacker that was declared and left still keeps the steps after a
// restore.
func TestCombatHadAttackersSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	g.WithWriteLock(func() {
		g.commitAttackDeclarationLocked()
		findBattlefieldCard(g, attacker).AttackingTarget = uuid.Nil
	})
	if !g.combatHadAttackers {
		t.Fatal("setup: the declaration did not record an attacker")
	}
	if !g.Clone().combatHadAttackers {
		t.Error("Clone dropped combatHadAttackers")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	restored, err := back.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !restored.combatHadAttackers {
		t.Error("the snapshot dropped combatHadAttackers")
	}
	// And it is spent with the combat.
	advanceIntoStep(t, restored, StepPostcombatMain)
	if restored.combatHadAttackers {
		t.Error("combatHadAttackers outlived the combat")
	}
}
