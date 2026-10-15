package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// nth_main_phase_test.go pins ADR 0188 slice 6 (#3564): the
// AnchorNthMainPhase anchor (Decision 4, World at War's "after the second
// main phase this turn") and the delayed trigger bound to one phase of one
// turn (Decision 5, DelayedTrigger.OnPhaseID / OnSeq).

// phaseOrder collapses a walk to the order of the phase ids it visited.
func phaseOrder(ps []plannedPos) []int {
	var out []int
	for _, p := range ps {
		if len(out) == 0 || out[len(out)-1] != p.phaseID {
			out = append(out, p.phaseID)
		}
	}
	return out
}

func addNthMain(t *testing.T, g *Game, n int, kinds ...PhaseKind) []int {
	t.Helper()
	var ids []int
	g.WithWriteLock(func() {
		ids = g.AddPhasesForEffect(uuid.Nil, PhaseAnchor{Kind: AnchorNthMainPhase, N: n}, kinds...)
	})
	return ids
}

// advanceToPhase walks until the cursor is in the phase with this id.
func advanceToPhase(t *testing.T, g *Game, id int) {
	t.Helper()
	for i := 0; i < 80 && g.Turn.PhaseID != id; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward phase %d: %v", id, err)
		}
	}
	if g.Turn.PhaseID != id {
		t.Fatalf("never reached phase %d (at %+v)", id, g.Turn)
	}
}

// TestNthMainPhaseFromPrecombatMain: cast in the precombat main phase,
// the pair goes after the SECOND main phase: beginning is behind us, then
// main, combat, main, [combat, main], ending.
func TestNthMainPhaseFromPrecombatMain(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	ids := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain)
	if !reflect.DeepEqual(ids, []int{6, 7}) {
		t.Fatalf("ids %v, want [6 7]", ids)
	}
	got := phaseOrder(walkRestOfTurn(t, g))
	if want := []int{3, 4, 6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases walked %v, want %v (combat, main, [combat, main], ending)", got, want)
	}
}

// TestNthMainPhaseFromTheSecondMainPhase: in the second main phase it IS
// the anchor, the common case.
func TestNthMainPhaseFromTheSecondMainPhase(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPostcombatMain)
	ids := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain)
	if !reflect.DeepEqual(ids, []int{6, 7}) {
		t.Fatalf("ids %v, want [6 7]", ids)
	}
	got := phaseOrder(walkRestOfTurn(t, g))
	if want := []int{6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases walked %v, want %v", got, want)
	}
}

// TestNthMainPhaseFromTheUpkeep is the rebound case: a spell cast from
// exile in the upkeep anchors to that turn's second main phase.
func TestNthMainPhaseFromTheUpkeep(t *testing.T) {
	g := newActiveGame(t)
	if g.Turn.Step != StepUpkeep {
		t.Fatalf("setup: cursor %+v, want upkeep", g.Turn)
	}
	ids := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain)
	if !reflect.DeepEqual(ids, []int{6, 7}) {
		t.Fatalf("ids %v, want [6 7]", ids)
	}
	got := phaseOrder(walkRestOfTurn(t, g))
	if want := []int{1, 2, 3, 4, 6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases walked %v, want %v", got, want)
	}
}

// TestTwoNthMainCopiesInsertNewestFirst is the World at War ruling:
// several copies insert "after the original postcombat main phase and
// before the newest combat phase", so the second copy's pair comes first.
func TestTwoNthMainCopiesInsertNewestFirst(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	first := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain)
	second := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain)
	if !reflect.DeepEqual(first, []int{6, 7}) || !reflect.DeepEqual(second, []int{8, 9}) {
		t.Fatalf("ids %v then %v", first, second)
	}
	got := phaseOrder(walkRestOfTurn(t, g))
	if want := []int{3, 4, 8, 9, 6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases walked %v, want %v (the second copy's pair first)", got, want)
	}
}

// TestNthMainCopyInTheThirdMainPhaseAddsNothing is the ruling: a World at
// War cast after the second main phase has ended creates no phases, even
// though a third main phase is in progress.
func TestNthMainCopyInTheThirdMainPhaseAddsNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPostcombatMain)
	if ids := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain); len(ids) != 2 {
		t.Fatalf("setup: ids %v", ids)
	}
	advanceToPhase(t, g, 7)
	if g.Turn.PhaseOrdinal != 3 {
		t.Fatalf("setup: in main phase %d, want 3", g.Turn.PhaseOrdinal)
	}
	planBefore := append([]PlannedStep(nil), g.TurnPlan...)
	events := len(g.Events)
	if ids := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain); ids != nil {
		t.Fatalf("a copy in the third main phase added %v", ids)
	}
	if !reflect.DeepEqual(g.TurnPlan, planBefore) {
		t.Errorf("plan changed: %+v", g.TurnPlan)
	}
	for _, ev := range g.Events[events:] {
		if ev.Kind == EventPhasesAdded {
			t.Errorf("EventPhasesAdded emitted for nothing added")
		}
	}
}

// TestNthMainCountsAnAddedMainPhase is CR 505.1b (ADR 0188 alternative E):
// after a Relentless Assault in the precombat main phase the turn's second
// main phase is the ADDED one, and "after the second main phase" goes
// there, not after the template's postcombat main.
func TestNthMainCountsAnAddedMainPhase(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	if ids := addPhases(t, g, AnchorThisMainPhase, PhaseKindCombat, PhaseKindMain); !reflect.DeepEqual(ids, []int{6, 7}) {
		t.Fatalf("setup: ids %v", ids)
	}
	// Plan: combat 6, main 7, combat 3, main 4. The second main is 7.
	ids := addNthMain(t, g, 2, PhaseKindCombat, PhaseKindMain)
	if !reflect.DeepEqual(ids, []int{8, 9}) {
		t.Fatalf("ids %v, want [8 9]", ids)
	}
	got := phaseOrder(walkRestOfTurn(t, g))
	if want := []int{6, 7, 8, 9, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases walked %v, want %v", got, want)
	}
}

// TestNthMainOfOneAfterItEndedAddsNothing: "the first main phase" has
// ended once the cursor is past it, and N beyond the plan finds nothing.
func TestNthMainOfOneAfterItEndedAddsNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepBeginCombat)
	if ids := addNthMain(t, g, 1, PhaseKindCombat); ids != nil {
		t.Errorf("after the first main phase ended, N=1 added %v", ids)
	}
	if ids := addNthMain(t, g, 3, PhaseKindCombat); ids != nil {
		t.Errorf("a third main phase that is not coming added %v", ids)
	}
	if ids := addNthMain(t, g, 0, PhaseKindCombat); ids != nil {
		t.Errorf("N=0 added %v", ids)
	}
	// And N=2 is still ahead.
	if ids := addNthMain(t, g, 2, PhaseKindCombat); len(ids) != 1 {
		t.Errorf("N=2 from the first combat added %v, want one combat", ids)
	}
}

// boundProbe schedules a beginning-of-combat trigger bound to phase id
// and returns a fire counter.
func boundProbe(t *testing.T, g *Game, phase int, fired *int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[g.Turn.ActiveSeat].ID,
			Label:      "probe — bound to one combat",
			At:         StepBeginCombat,
			OnPhaseID:  phase,
			Body: testBody(func(_ *Game, _ *StackItem) error {
				*fired++
				return nil
			}),
		})
	})
	return id
}

// TestPhaseBoundTriggerFiresOnlyAtItsCombat: Moraug's ruling. Resolved in
// the precombat main phase the added combat comes BEFORE the regular one;
// the trigger fires at the added combat and not at the regular one.
func TestPhaseBoundTriggerFiresOnlyAtItsCombat(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if len(ids) != 1 {
		t.Fatalf("ids %v", ids)
	}
	fired := 0
	if id := boundProbe(t, g, ids[0], &fired); id == uuid.Nil {
		t.Fatal("the schedule was refused")
	}
	advanceIntoStep(t, g, StepBeginCombat)
	if g.Turn.PhaseID != ids[0] {
		t.Fatalf("first combat is phase %d, want the added %d", g.Turn.PhaseID, ids[0])
	}
	settleStack(t, g)
	if fired != 1 {
		t.Fatalf("fired %d times at the added combat, want 1", fired)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a fired once-trigger stayed queued: %+v", g.DelayedTriggers)
	}
	advanceToPhase(t, g, 3)
	advanceIntoStep(t, g, StepEndCombat)
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("fired %d times after the regular combat, want still 1", fired)
	}
}

// TestPhaseBoundTriggerWaitsPastTheWrongCombat: bound to the SECOND of two
// added combats, it does not fire at the first.
func TestPhaseBoundTriggerWaitsPastTheWrongCombat(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPostcombatMain)
	older := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	newer := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	fired := 0
	boundProbe(t, g, older[0], &fired)
	advanceIntoStep(t, g, StepBeginCombat)
	if g.Turn.PhaseID != newer[0] {
		t.Fatalf("first combat is %d, want the newer %d", g.Turn.PhaseID, newer[0])
	}
	settleStack(t, g)
	if fired != 0 {
		t.Fatalf("fired at the combat it was not bound to")
	}
	advanceToPhase(t, g, older[0])
	settleStack(t, g)
	if fired != 1 {
		t.Fatalf("fired %d times at its own combat, want 1", fired)
	}
}

// TestScheduleRefusesAnUnreachablePhase: an id that is neither the phase
// in progress nor in the plan can never begin, so it is never queued.
func TestScheduleRefusesAnUnreachablePhase(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepBeginCombat)
	fired := 0
	for _, tc := range []struct {
		name  string
		phase int
	}{
		{"never minted", 99},
		{"already over (precombat main)", 2},
	} {
		if id := boundProbe(t, g, tc.phase, &fired); id != uuid.Nil {
			t.Errorf("%s: scheduled phase %d", tc.name, tc.phase)
		}
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a refused trigger was queued: %+v", g.DelayedTriggers)
	}
	// The phase in progress and a later one are both reachable.
	if id := boundProbe(t, g, g.Turn.PhaseID, &fired); id == uuid.Nil {
		t.Errorf("the phase in progress was refused")
	}
	if id := boundProbe(t, g, 4, &fired); id == uuid.Nil {
		t.Errorf("a phase still to come was refused")
	}
}

// TestPhaseBoundTriggerIsSweptWhenItsCombatNeverBegins: the turn is ended
// (CR 724.1) before the added combat. Phase ids restart every turn, so
// without OnSeq the next turn's own combat (id 3) would fire a trigger
// bound to a phase of the old turn. It is swept as the next turn begins.
func TestPhaseBoundTriggerIsSweptWhenItsCombatNeverBegins(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	fired := 0
	boundProbe(t, g, 3, &fired) // this turn's regular combat
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("setup: %d triggers queued", len(g.DelayedTriggers))
	}
	if g.DelayedTriggers[0].OnSeq != g.Turn.Seq {
		t.Fatalf("OnSeq %d, want the scheduling turn %d", g.DelayedTriggers[0].OnSeq, g.Turn.Seq)
	}
	passTurn(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("the unreachable bound trigger survived the turn change: %+v", g.DelayedTriggers)
	}
	advanceIntoStep(t, g, StepBeginCombat)
	settleStack(t, g)
	if fired != 0 {
		t.Errorf("a trigger bound to the old turn's combat fired in the next turn's")
	}
}

// TestPhaseBoundTriggerIsSweptWhenTheTurnIsEnded is the same sweep through
// CR 724.1's real "end the turn": the added combat is still to come when
// the turn ends, never begins, and its trigger is gone at the next turn.
func TestPhaseBoundTriggerIsSweptWhenTheTurnIsEnded(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	active := g.Seats[0]
	advanceIntoStep(t, g, StepPrecombatMain)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	fired := 0
	if id := boundProbe(t, g, ids[0], &fired); id == uuid.Nil {
		t.Fatal("the schedule was refused")
	}
	turn := g.Turn.Seq
	castEndTurnSpell(t, g, active)
	passUntilStackEmpty(t, g)
	if g.Turn.Seq != turn+1 {
		t.Fatalf("the turn did not end")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("the trigger bound to a combat that never began survived: %+v", g.DelayedTriggers)
	}
	if fired != 0 {
		t.Errorf("the trigger fired although its combat never began")
	}
}

// TestPhaseBoundFireTestChecksTheTurn: a trigger whose OnSeq is stale
// never fires even if the sweep did not remove it, because the fire test
// checks OnSeq too (belt and braces for a restore that skips the sweep).
func TestPhaseBoundFireTestChecksTheTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	fired := 0
	boundProbe(t, g, 3, &fired)
	g.DelayedTriggers[0].OnSeq = g.Turn.Seq - 1 // as if it were left over from the turn before
	advanceIntoStep(t, g, StepBeginCombat)
	settleStack(t, g)
	if fired != 0 {
		t.Errorf("fired in a turn it was not scheduled in")
	}
}

// TestPhaseBoundTriggerCloneAndSnapshot: the binding clones, and comes
// back from a restore point with the same phase and turn.
func TestPhaseBoundTriggerCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID, Label: "probe", At: StepBeginCombat, OnPhaseID: ids[0],
			Body: carriedTestBodyKey,
		})
	})
	want := g.DelayedTriggers[0]

	cl := g.Clone()
	if got := cl.DelayedTriggers[0]; got.OnPhaseID != ids[0] || got.OnSeq != want.OnSeq {
		t.Errorf("clone lost the binding: %+v", got)
	}
	cl.DelayedTriggers[0].OnPhaseID = 0
	if g.DelayedTriggers[0].OnPhaseID != ids[0] {
		t.Errorf("the clone aliases the live trigger")
	}

	_, restored := roundTrip(t, g)
	if len(restored.DelayedTriggers) != 1 {
		t.Fatalf("restored %d triggers", len(restored.DelayedTriggers))
	}
	if got := restored.DelayedTriggers[0]; got.OnPhaseID != ids[0] || got.OnSeq != want.OnSeq {
		t.Errorf("restore lost the binding: %+v", got)
	}
}

// TestDurationsAcrossTwoCombats is ADR 0188 Decision 6: "until end of
// combat" from the first combat has ended when the second begins, and
// "until end of turn" lasts through both, to the cleanup step.
func TestDurationsAcrossTwoCombats(t *testing.T) {
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepBeginCombat)
	first := g.Turn.PhaseID
	if ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat); len(ids) != 1 {
		t.Fatalf("setup: ids %v", ids)
	}
	if !combatShield(t, g, 0, thisCombat(t, g)) {
		t.Fatal("the until-end-of-combat shield was not registered")
	}
	var turnLong Duration
	g.WithWriteLock(func() { turnLong = g.UntilEndOfTurnDuration() })
	if !combatShield(t, g, 1, turnLong) {
		t.Fatal("the until-end-of-turn shield was not registered")
	}
	if n := scopedEffectCount(g); n != 2 {
		t.Fatalf("%d records during the first combat, want 2", n)
	}
	for i := 0; i < 20 && !(g.Turn.PhaseID != first && g.Turn.Step == StepBeginCombat); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if g.Turn.PhaseID == first {
		t.Fatalf("never reached the second combat")
	}
	if n := scopedEffectCount(g); n != 1 {
		t.Fatalf("%d records in the second combat, want only the until-end-of-turn one", n)
	}
	advanceIntoStep(t, g, StepEnd)
	if n := scopedEffectCount(g); n != 1 {
		t.Fatalf("%d records in the end step, want the until-end-of-turn one still in force", n)
	}
}
