package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// extra_upkeep_test.go pins #5405: AddUpkeepStepsAfterThisPhaseForEffect,
// "you get N additional upkeep steps after this phase" (Obeka, Splitter
// of Seconds). CR 500.10: each is a beginning phase with only its upkeep
// step, directly after the phase in progress; CR 500.8: newest first;
// CR 500.10a: "you get" adds nothing on another player's turn.

func addUpkeeps(t *testing.T, g *Game, you uuid.UUID, n int) []int {
	t.Helper()
	var ids []int
	g.WithWriteLock(func() { ids = g.AddUpkeepStepsAfterThisPhaseForEffect(uuid.Nil, you, n) })
	return ids
}

func TestAdditionalUpkeepStepsAfterThisPhase(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepBeginCombat)
	from := len(g.Events)
	upkeepsBefore := me.UpkeepsBegun
	ids := addUpkeeps(t, g, me.ID, 2)
	if len(ids) != 2 || ids[0] == ids[1] || ids[0] <= templatePhaseCount {
		t.Fatalf("added phase ids %v, want two fresh ones", ids)
	}
	var added *Event
	for i := range g.Events[from:] {
		if ev := g.Events[from+i]; ev.Kind == EventPhasesAdded {
			added = &ev
		}
	}
	if added == nil || added.Step != StepUpkeep || added.Amount != 2 || added.Actor != me.ID {
		t.Fatalf("EventPhasesAdded %+v, want step upkeep, amount 2, the active player", added)
	}

	walked := walkRestOfTurn(t, g)
	want := concatSteps(combatSteps[1:], []Step{StepUpkeep, StepUpkeep, StepPostcombatMain, StepEnd})
	if !reflect.DeepEqual(stepsOf(walked), want) {
		t.Fatalf("the turn walked %v\nwant %v (no untap or draw step: CR 500.10)", stepsOf(walked), want)
	}
	var upkeepPhases []int
	for _, p := range walked {
		if p.step == StepUpkeep {
			upkeepPhases = append(upkeepPhases, p.phaseID)
		}
	}
	if !reflect.DeepEqual(upkeepPhases, ids) {
		t.Errorf("upkeeps ran in phases %v, want %v: each its own beginning phase", upkeepPhases, ids)
	}
	began := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventBeginUpkeep && ev.Actor == me.ID {
			began++
		}
	}
	if began != 2 {
		t.Errorf("EventBeginUpkeep fired %d times for the added upkeeps, want 2", began)
	}
	if me.UpkeepsBegun != upkeepsBefore+2 {
		t.Errorf("UpkeepsBegun %d, want %d", me.UpkeepsBegun, upkeepsBefore+2)
	}
}

// The active player gets priority in each added upkeep (CR 503.1), and
// it reads as the second beginning phase and the second upkeep of the
// turn.
func TestAdditionalUpkeepGrantsPriority(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepEndCombat)
	addUpkeeps(t, g, me.ID, 1)
	seq := g.Turn.Seq
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != StepUpkeep || g.Turn.Seq != seq {
		t.Fatalf("cursor %+v, want the added upkeep of this turn", g.Turn)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("priority with seat %d, want the active player", g.Turn.PriorityHolder)
	}
	if g.Turn.PhaseOrdinal != 2 || g.Turn.StepOrdinal != 2 {
		t.Errorf("phase ordinal %d, step ordinal %d; want 2 and 2", g.Turn.PhaseOrdinal, g.Turn.StepOrdinal)
	}
}

// The Obeka ruling's example: upkeeps added during combat damage, then
// another combat added during end of combat: the newest (the combat)
// runs first, then the upkeeps, then the postcombat main phase.
func TestAdditionalUpkeepsAndALaterCombatRunNewestFirst(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepBeginCombat)
	addUpkeeps(t, g, me.ID, 2)
	addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	walked := walkRestOfTurn(t, g)
	want := concatSteps(combatSteps[1:], combatSteps, []Step{StepUpkeep, StepUpkeep, StepPostcombatMain, StepEnd})
	if !reflect.DeepEqual(stepsOf(walked), want) {
		t.Fatalf("the turn walked %v\nwant %v", stepsOf(walked), want)
	}
}

// Strict refusals: no steps for n < 1, and "you get" on another
// player's turn adds nothing (CR 500.10a). Neither emits an event.
func TestAdditionalUpkeepRefusals(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceIntoStep(t, g, StepBeginCombat)
	before := clonePlan(g.TurnPlan)
	from := len(g.Events)
	for _, tc := range []struct {
		name string
		you  uuid.UUID
		n    int
	}{
		{"zero steps", g.Seats[g.Turn.ActiveSeat].ID, 0},
		{"negative steps", g.Seats[g.Turn.ActiveSeat].ID, -2},
		{"another player's turn", opp.ID, 3},
	} {
		if ids := addUpkeeps(t, g, tc.you, tc.n); ids != nil {
			t.Errorf("%s: added %v", tc.name, ids)
		}
	}
	if !reflect.DeepEqual(g.TurnPlan, before) {
		t.Errorf("a refused call changed the plan: %+v", g.TurnPlan)
	}
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventPhasesAdded {
			t.Errorf("a refused call emitted %+v", ev)
		}
	}
}

// The queue is the turn plan, so a restore point carries it (additive:
// PlannedStep entries the snapshot already writes) and a restored game
// plays every upkeep still to come, and an undo clone does not see an
// addition made after it.
func TestAdditionalUpkeepsSurviveRestoreAndUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepEndCombat)
	undo := g.Clone()
	addUpkeeps(t, g, me.ID, 3)

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
	if !reflect.DeepEqual(restored.TurnPlan, g.TurnPlan) || restored.NextPhaseID != g.NextPhaseID {
		t.Fatalf("plan did not round-trip: %+v / %d", restored.TurnPlan, restored.NextPhaseID)
	}
	want := []Step{StepUpkeep, StepUpkeep, StepUpkeep, StepPostcombatMain, StepEnd}
	if got := stepsOf(walkRestOfTurn(t, restored)); !reflect.DeepEqual(got, want) {
		t.Errorf("restored game walked %v, want %v", got, want)
	}

	g.RestoreFrom(undo)
	if got := stepsOf(walkRestOfTurn(t, g)); !reflect.DeepEqual(got, []Step{StepPostcombatMain, StepEnd}) {
		t.Errorf("after undo the turn walked %v, want no added upkeep", got)
	}
}
