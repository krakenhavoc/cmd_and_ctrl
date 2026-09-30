package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// turn_plan_test.go pins ADR 0059 sub-PR 2b (#753): the turn plan
// (Decision 3), added phases and steps (Decision 4), ordinals and the
// attack history (Decision 8), and their clone and snapshot (Decision
// 10).

type plannedPos struct {
	step    Step
	phaseID int
}

// walkRestOfTurn advances step by step until the turn changes, and
// returns every step the cursor landed on (the one it started on
// excluded).
func walkRestOfTurn(t *testing.T, g *Game) []plannedPos {
	t.Helper()
	seq := g.Turn.Seq
	var out []plannedPos
	for i := 0; i < 80; i++ {
		before := g.Turn
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn == before {
			t.Fatalf("AdvanceStep did not move from %+v", before)
		}
		if g.Turn.Seq != seq {
			return out
		}
		out = append(out, plannedPos{g.Turn.Step, g.Turn.PhaseID})
	}
	t.Fatalf("turn never ended: %+v", out)
	return nil
}

func stepsOf(ps []plannedPos) []Step {
	out := make([]Step, len(ps))
	for i, p := range ps {
		out[i] = p.step
	}
	return out
}

func addPhases(t *testing.T, g *Game, anchor AnchorKind, kinds ...PhaseKind) []int {
	t.Helper()
	var ids []int
	g.WithWriteLock(func() {
		ids = g.AddPhasesForEffect(uuid.Nil, PhaseAnchor{Kind: anchor}, kinds...)
	})
	return ids
}

var combatSteps = []Step{StepBeginCombat, StepDeclareAttackers, StepDeclareBlockers, StepCombatDamage, StepEndCombat}

func concatSteps(parts ...[]Step) []Step {
	var out []Step
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// TestTurnPlanWalksTheTemplate: with nothing added, the plan is the
// fixed sequence it replaced, #717's first-strike step still walked
// through when nobody has first strike, and every template phase has
// its template id.
func TestTurnPlanWalksTheTemplate(t *testing.T) {
	g := newActiveGame(t)
	if g.Turn.Step != StepUpkeep || g.Turn.PhaseID != 1 {
		t.Fatalf("setup: cursor %+v, want upkeep of phase 1", g.Turn)
	}
	got := walkRestOfTurn(t, g)
	want := []plannedPos{
		{StepDraw, 1}, {StepPrecombatMain, 2},
		{StepBeginCombat, 3}, {StepDeclareAttackers, 3}, {StepDeclareBlockers, 3},
		{StepCombatDamage, 3}, {StepEndCombat, 3},
		{StepPostcombatMain, 4}, {StepEnd, 5},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("walked %+v\nwant   %+v", got, want)
	}
	if g.Turn.Step != StepUpkeep || g.Turn.PhaseID != 1 {
		t.Errorf("next turn begins at %+v, want upkeep of phase 1", g.Turn)
	}
	if len(g.TurnPlan) == 0 || g.TurnPlan[len(g.TurnPlan)-1].Step != StepCleanup {
		t.Errorf("the new turn's plan does not end at cleanup: %+v", g.TurnPlan)
	}
}

// TestRelentlessAssaultShapeFromPrecombatMain is ADR 0059 test 7:
// "after this main phase, an additional combat phase followed by an
// additional main phase", from the precombat main phase, gives main,
// combat, main (postcombat), combat, main (postcombat), ending — and
// the ordinals count the families.
func TestRelentlessAssaultShapeFromPrecombatMain(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	ids := addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain)
	if len(ids) != 2 || ids[0] != 6 || ids[1] != 7 {
		t.Fatalf("ids %v, want [6 7]", ids)
	}

	var ordinals []int
	seq := g.Turn.Seq
	var walked []plannedPos
	for g.Turn.Seq == seq {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Seq != seq {
			break
		}
		walked = append(walked, plannedPos{g.Turn.Step, g.Turn.PhaseID})
		if g.Turn.Step == StepBeginCombat || g.Turn.Step == StepPostcombatMain {
			ordinals = append(ordinals, g.Turn.PhaseOrdinal)
		}
	}
	wantSteps := concatSteps(combatSteps, []Step{StepPostcombatMain}, combatSteps, []Step{StepPostcombatMain, StepEnd})
	if !reflect.DeepEqual(stepsOf(walked), wantSteps) {
		t.Fatalf("walked %v\nwant   %v", stepsOf(walked), wantSteps)
	}
	if walked[0].phaseID != 6 || walked[5].phaseID != 7 || walked[6].phaseID != 3 || walked[11].phaseID != 4 {
		t.Errorf("phase ids along the walk: %+v", walked)
	}
	// combat 1, main 2, combat 2, main 3.
	if want := []int{1, 2, 2, 3}; !reflect.DeepEqual(ordinals, want) {
		t.Errorf("phase ordinals at each new phase %v, want %v", ordinals, want)
	}
}

// TestAddedPhasesAfterTheSameAnchorRunNewestFirst is CR 500.8's "most
// recently created occurs first": two "after this phase" combats added
// during one combat run in the reverse of the order they were added.
func TestAddedPhasesAfterTheSameAnchorRunNewestFirst(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepBeginCombat)
	first := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	second := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	walked := walkRestOfTurn(t, g)
	var combatIDs []int
	for _, p := range walked {
		if p.step == StepBeginCombat {
			combatIDs = append(combatIDs, p.phaseID)
		}
	}
	if want := []int{second[0], first[0]}; !reflect.DeepEqual(combatIDs, want) {
		t.Fatalf("added combats began in order %v, want %v (newest first)", combatIDs, want)
	}
	// And no main phase between the combats: end of combat goes
	// straight to the next beginning of combat (the Aurelia ruling).
	for i, p := range walked {
		if p.step == StepEndCombat && i+1 < len(walked) && walked[i+1].step == StepPostcombatMain {
			if i+1 != len(walked)-2 {
				t.Errorf("a main phase came between added combats: %v", stepsOf(walked))
			}
		}
	}
}

// TestAnchorThisMainPhaseOutsideAMainPhaseAddsNothing: the Relentless
// Assault ruling — resolving outside a main phase creates no phases.
func TestAnchorThisMainPhaseOutsideAMainPhaseAddsNothing(t *testing.T) {
	g := newActiveGame(t)
	before := append([]PlannedStep(nil), g.TurnPlan...)
	events := len(g.Events)
	if ids := addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain); ids != nil {
		t.Fatalf("added %v from the upkeep", ids)
	}
	if !reflect.DeepEqual(g.TurnPlan, before) {
		t.Errorf("plan changed: %+v", g.TurnPlan)
	}
	for _, ev := range g.Events[events:] {
		if ev.Kind == EventPhasesAdded {
			t.Errorf("EventPhasesAdded emitted for nothing added")
		}
	}
}

// TestAddedMainPhaseIsPostcombat is ADR 0059 test 9 (CR 505.1a): an
// added main phase is a postcombat main phase, so the precombat-main
// announcement happens once, and postcombat main begins twice.
func TestAddedMainPhaseIsPostcombat(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	start := len(g.Events)
	addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain)
	walkRestOfTurn(t, g)
	precombat, postcombat := 0, 0
	for _, ev := range g.Events[start:] {
		switch {
		case ev.Kind == EventBeginPrecombatMain:
			precombat++
		case ev.Kind == EventStepBegan && ev.Step == StepPostcombatMain:
			postcombat++
		}
	}
	if precombat != 0 {
		t.Errorf("EventBeginPrecombatMain fired %d more times in the turn", precombat)
	}
	if postcombat != 2 {
		t.Errorf("postcombat main began %d times, want 2", postcombat)
	}
}

// TestAdditionalEndStep is CR 500.9 (Y'shtola Rhul): an end step added
// after the end step begins as a second end step, with its own
// announcement and ordinal, and the turn then goes to cleanup.
func TestAdditionalEndStep(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepEnd)
	if !g.IsFirstStepOfItsKindForEffect() || g.Turn.StepOrdinal != 1 {
		t.Fatalf("first end step reads ordinal %d", g.Turn.StepOrdinal)
	}
	var ok, wrong bool
	g.WithWriteLock(func() {
		wrong = g.AddStepAfterCurrentForEffect(uuid.Nil, StepUpkeep)
		ok = g.AddStepAfterCurrentForEffect(uuid.Nil, StepEnd)
	})
	if wrong {
		t.Errorf("an upkeep step was added to the ending phase")
	}
	if !ok {
		t.Fatalf("AddStepAfterCurrentForEffect(end) = false")
	}
	start := len(g.Events)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != StepEnd || g.Turn.StepOrdinal != 2 || g.IsFirstStepOfItsKindForEffect() {
		t.Fatalf("after the first end step: %+v", g.Turn)
	}
	ends := 0
	for _, ev := range g.Events[start:] {
		if ev.Kind == EventBeginEndStep {
			ends++
		}
	}
	if ends != 1 {
		t.Errorf("the added end step announced %d times, want 1", ends)
	}
	seq := g.Turn.Seq
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Seq == seq {
		t.Errorf("the turn did not end after the added end step: %+v", g.Turn)
	}
}

// TestAddedBeginningPhaseUntapsButIsNotANewTurn is Sphinx of the
// Second Sun's ruling: the added untap step untaps, but a creature that
// came under its controller's control this turn stays summoning sick,
// and the per-turn tally is not reset.
func TestAddedBeginningPhaseUntapsButIsNotANewTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushCreatureFor(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepPostcombatMain)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
		g.bumpPlayerTally(me.ID, func(p *PlayerTurnTally) { p.CardsDrawn += 5 })
	})
	addPhases(t, g, AnchorThisPhase, PhaseKindBeginning)
	seq, begun := g.Turn.Seq, me.TurnsBegun
	advanceIntoStep(t, g, StepUpkeep)
	if g.Turn.Seq != seq || me.TurnsBegun != begun {
		t.Fatalf("the added beginning phase began a new turn: %+v", g.Turn)
	}
	c, _ := g.LookupCardForEffect(id)
	if c.Tapped {
		t.Errorf("the added untap step did not untap")
	}
	if !c.SummonedThisTurn {
		t.Errorf("the added untap step cured summoning sickness (CR 302.6)")
	}
	if g.TurnTallyFor(me.ID).CardsDrawn < 5 {
		t.Errorf("the per-turn tally was reset by an added beginning phase")
	}
	if g.Turn.StepOrdinal != 2 || g.Turn.PhaseOrdinal != 2 {
		t.Errorf("second upkeep reads step ordinal %d, phase ordinal %d; want 2 and 2", g.Turn.StepOrdinal, g.Turn.PhaseOrdinal)
	}
}

// TestSecondCombatStartsClean: combat state from the first combat is
// gone in the second (CR 511.3), the attack history remembers both
// (ADR 0059 test 13), and IsFirstCombatPhase is true only in the first
// (test 10).
func TestSecondCombatStartsClean(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	attacker := pushCreatureFor(t, g, me, 3, 3)
	blocker := pushCreatureFor(t, g, them, 1, 1)
	advanceIntoStep(t, g, StepBeginCombat)
	if !g.IsFirstCombatPhaseForEffect() {
		t.Fatalf("first combat reads IsFirstCombatPhase false: %+v", g.Turn)
	}
	addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceIntoStep(t, g, StepEndCombat)
	firstCombat := g.Turn.PhaseID
	if g.TimesAttackedThisTurn(attacker) != 1 {
		t.Fatalf("TimesAttackedThisTurn after the first combat = %d", g.TimesAttackedThisTurn(attacker))
	}
	advanceIntoStep(t, g, StepBeginCombat)
	if g.Turn.PhaseID == firstCombat || g.IsFirstCombatPhaseForEffect() || g.Turn.PhaseOrdinal != 2 {
		t.Fatalf("second combat: %+v", g.Turn)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
				t.Errorf("%s still in combat from the first combat", c.Name)
			}
		}
		if len(g.blockedAttackers) != 0 || len(g.announcedAttacks) != 0 || g.attacksDeclared || len(g.firstStrikeStepParticipants) != 0 {
			t.Errorf("combat bookkeeping survived into the second combat")
		}
	})
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(attacker) })
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if n := g.TimesAttackedThisTurn(attacker); n != 2 {
		t.Errorf("TimesAttackedThisTurn = %d, want 2", n)
	}
	if got := g.AttackedPlayersThisTurn(attacker); len(got) != 1 || got[0] != them.ID {
		t.Errorf("AttackedPlayersThisTurn = %v, want [%v]", got, them.ID)
	}
	// A new object has not attacked (CR 400.7).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].ObjectEpoch++
			}
		}
	})
	if g.AttackedThisTurn(attacker) {
		t.Errorf("a new object reads as having attacked")
	}
}

// TestEndCombatPhaseEndsOnlyThisCombat: CR 724.2's "end the combat
// phase" skips out of THIS combat phase and into the next phase, which
// may be an added combat.
func TestEndCombatPhaseEndsOnlyThisCombat(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepDeclareAttackers)
	added := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	g.WithWriteLock(func() { g.EndCombatPhaseForEffect() })
	if g.Turn.Step != StepBeginCombat || g.Turn.PhaseID != added[0] {
		t.Fatalf("after ending the first combat: %+v, want the added combat's beginning", g.Turn)
	}
}

// TestTurnPlanCloneAndUndo is ADR 0059 test 18: an undo taken before
// the phases were added restores the plan without them.
func TestTurnPlanCloneAndUndo(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	snap := g.Clone()
	addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain)
	if len(g.TurnPlan) == len(snap.TurnPlan) {
		t.Fatalf("setup: no phases added")
	}
	g.RestoreFrom(snap)
	walked := walkRestOfTurn(t, g)
	want := concatSteps(combatSteps, []Step{StepPostcombatMain, StepEnd})
	if !reflect.DeepEqual(stepsOf(walked), want) {
		t.Fatalf("after undo the turn walked %v, want %v", stepsOf(walked), want)
	}
}

// TestTurnPlanSnapshotRoundTrip is ADR 0059 test 15: a game restored
// with an added phase still to come plays it, and a snapshot written
// before the plan existed restores with the template's tail.
func TestTurnPlanSnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain)

	snap := g.CaptureSnapshot()
	raw, err := json.Marshal(snap)
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
	if !reflect.DeepEqual(restored.TurnPlan, g.TurnPlan) || restored.NextPhaseID != g.NextPhaseID {
		t.Fatalf("plan did not round-trip: %+v / %d", restored.TurnPlan, restored.NextPhaseID)
	}
	walked := walkRestOfTurn(t, restored)
	want := concatSteps(combatSteps, []Step{StepPostcombatMain}, combatSteps, []Step{StepPostcombatMain, StepEnd})
	if !reflect.DeepEqual(stepsOf(walked), want) {
		t.Fatalf("restored game walked %v\nwant %v", stepsOf(walked), want)
	}

	// A file from before the plan: no turnPlan, no nextPhaseId and no
	// phase fields on the cursor.
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	delete(generic, "turnPlan")
	delete(generic, "nextPhaseId")
	if turn, ok := generic["turn"].(map[string]any); ok {
		delete(turn, "PhaseID")
		delete(turn, "PhaseOrdinal")
		delete(turn, "StepOrdinal")
	}
	legacyRaw, _ := json.Marshal(generic)
	var legacy GameSnapshot
	if err := json.Unmarshal(legacyRaw, &legacy); err != nil {
		t.Fatal(err)
	}
	old, err := legacy.Restore()
	if err != nil {
		t.Fatalf("Restore legacy: %v", err)
	}
	if old.Turn.PhaseID != 2 {
		t.Errorf("legacy restore phase id %d, want the template's 2", old.Turn.PhaseID)
	}
	if want := templateTailAfter(StepPrecombatMain); !reflect.DeepEqual(old.TurnPlan, want) {
		t.Errorf("legacy restore plan %+v, want the template tail", old.TurnPlan)
	}
}

// TestHandMovedCursorRebuildsThePlan: a cursor moved by something
// other than the plan (a test setting Turn.Step) is followed by the
// template from where it stands, as the fixed walk used to be.
func TestHandMovedCursorRebuildsThePlan(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.Turn.Step = StepDeclareBlockers
		g.Turn.Phase = PhaseCombat
	})
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != StepCombatDamage || g.Turn.PhaseID != 3 {
		t.Fatalf("after a hand-moved declare blockers: %+v", g.Turn)
	}
}
