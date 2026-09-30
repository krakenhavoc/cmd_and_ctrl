package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

// turn_plan_preplan_restore_test.go: a restore point written before
// the turn plan (ADR 0059 sub-PR 2b, #753) comes back with the turn's
// ordinals, not with them counting again from zero.

// stripTurnPlan turns a capture into what a binary from before the
// turn plan wrote: no turnPlan or nextPhaseId, no phase fields on the
// cursor, and no step or phase counts in the turn tally.
func stripTurnPlan(t *testing.T, snap *GameSnapshot) *GameSnapshot {
	t.Helper()
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	delete(generic, "turnPlan")
	delete(generic, "nextPhaseId")
	turn, ok := generic["turn"].(map[string]any)
	if !ok {
		t.Fatal(`no "turn" in the capture`)
	}
	for _, k := range []string{"PhaseID", "PhaseOrdinal", "StepOrdinal"} {
		if _, ok := turn[k]; !ok {
			t.Fatalf("turn.%s is not in the capture: this helper no longer strips what the plan added", k)
		}
		delete(turn, k)
	}
	tally, ok := generic["turnTally"].(map[string]any)
	if !ok {
		t.Fatal(`no "turnTally" in the capture`)
	}
	for _, k := range []string{"stepsBegun", "phasesBegun", "phaseStarted"} {
		if _, ok := tally[k]; !ok {
			t.Fatalf("turnTally.%s is not in the capture: this helper no longer strips what the plan added", k)
		}
		delete(tally, k)
	}
	legacyRaw, err := json.Marshal(generic)
	if err != nil {
		t.Fatal(err)
	}
	var legacy GameSnapshot
	if err := json.Unmarshal(legacyRaw, &legacy); err != nil {
		t.Fatal(err)
	}
	return &legacy
}

func assertSameOrdinals(t *testing.T, where string, got, want *Game) {
	t.Helper()
	if got.Turn.Step != want.Turn.Step || got.Turn.PhaseID != want.Turn.PhaseID ||
		got.Turn.PhaseOrdinal != want.Turn.PhaseOrdinal || got.Turn.StepOrdinal != want.Turn.StepOrdinal {
		t.Errorf("%s: restored cursor %s phase %d, phase ordinal %d, step ordinal %d; live %s phase %d, phase ordinal %d, step ordinal %d",
			where, got.Turn.Step, got.Turn.PhaseID, got.Turn.PhaseOrdinal, got.Turn.StepOrdinal,
			want.Turn.Step, want.Turn.PhaseID, want.Turn.PhaseOrdinal, want.Turn.StepOrdinal)
	}
	gt, wt := got.TurnTally, want.TurnTally
	if !reflect.DeepEqual(gt.StepsBegun, wt.StepsBegun) || !reflect.DeepEqual(gt.PhasesBegun, wt.PhasesBegun) || gt.PhaseStarted != wt.PhaseStarted {
		t.Errorf("%s: restored counts steps %v phases %v started %d; live steps %v phases %v started %d",
			where, gt.StepsBegun, gt.PhasesBegun, gt.PhaseStarted, wt.StepsBegun, wt.PhasesBegun, wt.PhaseStarted)
	}
}

// TestPrePlanRestoreInTheSecondMainPhaseKeepsTheOrdinals: a pre-plan
// restore point in the postcombat main phase has the ordinals of a
// live game at the same step, and a combat added from there ("after
// this main phase", Relentless Assault) is the SECOND combat of the
// turn, so "if it's the first combat phase" (Karlach) reads false.
// Before the fix the restored counts started at zero and that combat
// read as the first.
func TestPrePlanRestoreInTheSecondMainPhaseKeepsTheOrdinals(t *testing.T) {
	live := newActiveGame(t)
	advanceIntoStep(t, live, StepPostcombatMain)
	if live.Turn.PhaseOrdinal != 2 {
		t.Fatalf("setup: live postcombat main has phase ordinal %d, want 2", live.Turn.PhaseOrdinal)
	}

	restored, err := stripTurnPlan(t, live.CaptureSnapshot()).Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	assertSameOrdinals(t, "at the restore point", restored, live)

	for _, g := range []*Game{live, restored} {
		addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain)
		advanceIntoStep(t, g, StepBeginCombat)
	}
	assertSameOrdinals(t, "in the added combat", restored, live)
	var liveFirst, restoredFirst bool
	live.WithWriteLock(func() { liveFirst = live.IsFirstCombatPhaseForEffect() })
	restored.WithWriteLock(func() { restoredFirst = restored.IsFirstCombatPhaseForEffect() })
	if liveFirst || restoredFirst {
		t.Errorf("the added combat reads as the first combat of the turn: live %v, restored %v", liveFirst, restoredFirst)
	}
	if restored.Turn.PhaseOrdinal != 2 {
		t.Errorf("the added combat's phase ordinal after a pre-plan restore is %d, want 2", restored.Turn.PhaseOrdinal)
	}
}

// TestPrePlanRestoreOrdinalsMatchALiveGameAtEveryStep walks one turn
// and restores a pre-plan copy at every step that gives priority: each
// has the ordinals and counts the live game has there.
func TestPrePlanRestoreOrdinalsMatchALiveGameAtEveryStep(t *testing.T) {
	live := newActiveGame(t)
	seq := live.Turn.Seq
	for live.Turn.Seq == seq {
		restored, err := stripTurnPlan(t, live.CaptureSnapshot()).Restore()
		if err != nil {
			t.Fatalf("Restore at %s: %v", live.Turn.Step, err)
		}
		assertSameOrdinals(t, string(live.Turn.Step), restored, live)
		if _, err := live.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// TestPlanFileRestoreKeepsItsOwnOrdinals: a file written WITH the plan
// is never re-derived — its counts are the record, even where they
// differ from the template (a second combat).
func TestPlanFileRestoreKeepsItsOwnOrdinals(t *testing.T) {
	live := newActiveGame(t)
	advanceIntoStep(t, live, StepPrecombatMain)
	addPhases(t, live, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain)
	advanceIntoStep(t, live, StepEndCombat)
	advanceIntoStep(t, live, StepPostcombatMain)
	advanceIntoStep(t, live, StepBeginCombat)
	if live.Turn.PhaseOrdinal != 2 {
		t.Fatalf("setup: second combat has phase ordinal %d, want 2", live.Turn.PhaseOrdinal)
	}
	raw, err := json.Marshal(live.CaptureSnapshot())
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
	assertSameOrdinals(t, "second combat", restored, live)
}
