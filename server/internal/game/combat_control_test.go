package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// combat_control_test.go — #3999: a PlayerControl record for one combat
// phase (Secret of Bloodbending, CR 723.2), driven through
// BeginCombatControlForEffect and the step seams.

const bloodbendingName = "Secret of Bloodbending"

// scheduleCombatControl makes seat `controller` control seat
// `controlled` during that seat's next combat phase.
func scheduleCombatControl(t *testing.T, g *Game, controller, controlled int) uuid.UUID {
	t.Helper()
	src := uuid.New()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.BeginCombatControlForEffect(CombatControl{
			Controller: g.Seats[controller].ID,
			Controlled: g.Seats[controlled].ID,
			Source:     src,
			SourceName: bloodbendingName,
		})
	})
	if !ok {
		t.Fatalf("BeginCombatControlForEffect(%d controls %d) refused", controller, controlled)
	}
	return src
}

// The record waits through the caster's own combat and the controlled
// player's precombat main phase, applies from their beginning of combat
// step to their end of combat step, and is gone by their postcombat
// main phase.
func TestCombatControlCoversTheirNextCombatPhaseOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	src := scheduleCombatControl(t, g, 0, 1)

	sched := g.ScheduledPlayerControls()
	if len(sched) != 1 || sched[0].UntilKind() != ControlUntilCombat || sched[0].Stage != ControlAwaitingCombat {
		t.Fatalf("scheduled = %+v, want the one waiting combat record", sched)
	}
	if ev := turnControlEvents(g, EventPlayerControlScheduled); len(ev) != 1 || ev[0].Label != "combat" {
		t.Fatalf("EventPlayerControlScheduled = %+v, want one labelled combat", ev)
	}
	// Seat 0's own combat phase is not "their" combat phase.
	advanceToStepOfSeat(t, g, 0, StepDeclareAttackers)
	if got := controllerOfSeat(g, 1); got != 1 {
		t.Fatalf("seat 1 controlled by seat %d during seat 0's combat", got)
	}
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if got := controllerOfSeat(g, 1); got != 1 {
		t.Fatalf("seat 1 controlled by seat %d before their combat phase", got)
	}

	for _, step := range []Step{StepBeginCombat, StepDeclareAttackers, StepEndCombat} {
		advanceToStepOfSeat(t, g, 1, step)
		if got := controllerOfSeat(g, 1); got != 0 {
			t.Fatalf("seat 1 controlled by seat %d in %s, want seat 0", got, step)
		}
		if got := g.ControlUntilOfPlayer(g.Seats[1].ID); got != ControlUntilCombat || !got.PlaysForThem() {
			t.Errorf("ControlUntilOfPlayer in %s = %q, want combat", step, got)
		}
		if got := g.ControlStageOfPlayer(g.Seats[1].ID); got != ControlCombat {
			t.Errorf("ControlStageOfPlayer in %s = %q, want %q", step, got, ControlCombat)
		}
	}
	if !g.ControlsATurn(g.Seats[0].ID) {
		t.Error("ControlsATurn is false for the combat's controller")
	}
	if n := len(g.ScheduledPlayerControls()); n != 0 {
		t.Errorf("a begun record is still scheduled (%d)", n)
	}
	began := turnControlEvents(g, EventPlayerControlBegan)
	if len(began) != 1 || began[0].Source != src || began[0].Label != "combat" {
		t.Fatalf("EventPlayerControlBegan = %+v", began)
	}

	advanceToStepOfSeat(t, g, 1, StepPostcombatMain)
	if got := controllerOfSeat(g, 1); got != 1 {
		t.Fatalf("seat 1 still controlled by seat %d after combat", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left after the combat phase: %+v", g.PlayerControls)
	}
	if ev := turnControlEvents(g, EventPlayerControlEnded); len(ev) != 1 || ev[0].Label != "combat" {
		t.Errorf("EventPlayerControlEnded = %+v, want one labelled combat", ev)
	}
}

// An additional combat phase right after the controlled one is a new
// phase: the spent record does not reach it, and a record made during
// the first combat waits for it ("their next combat phase").
func TestCombatControlEndsWithItsPhaseAndAMidCombatRecordWaitsForTheNext(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleCombatControl(t, g, 0, 1)
	advanceToStepOfSeat(t, g, 1, StepDeclareAttackers)
	if controllerOfSeat(g, 1) != 0 {
		t.Fatal("setup: seat 1's first combat is not controlled")
	}
	var extra []int
	g.WithWriteLock(func() {
		extra = g.AddPhasesForEffect(uuid.Nil, PhaseAnchor{Kind: AnchorThisPhase}, PhaseKindCombat)
	})
	if len(extra) != 1 {
		t.Fatalf("AddPhasesForEffect = %v", extra)
	}
	first := g.Turn.PhaseID
	scheduleCombatControl(t, g, 2, 1)
	// Still seat 0's: the newer record waits for the next combat phase.
	if got := controllerOfSeat(g, 1); got != 0 {
		t.Fatalf("seat 1 controlled by seat %d mid-combat, want seat 0 still", got)
	}

	advanceToPhase(t, g, extra[0])
	if g.Turn.PhaseID == first || PhaseOf(g.Turn.Step) != PhaseCombat {
		t.Fatalf("not in the additional combat: %+v", g.Turn)
	}
	if got := controllerOfSeat(g, 1); got != 2 {
		t.Fatalf("seat 1 controlled by seat %d in the additional combat, want seat 2 (the second record)", got)
	}
	if len(g.PlayerControls) != 1 || g.PlayerControls[0].Controller != g.Seats[2].ID {
		t.Fatalf("records in the additional combat = %+v, want the second alone", g.PlayerControls)
	}
	advanceToStepOfSeat(t, g, 1, StepPostcombatMain)
	if len(g.PlayerControls) != 0 || controllerOfSeat(g, 1) != 1 {
		t.Errorf("after both combats: records %+v, seat 1 controlled by %d", g.PlayerControls, controllerOfSeat(g, 1))
	}
}

// A turn that ends mid-combat (the rotation seam, CR 724.1) ends the
// control with it: it is never live in a later combat.
func TestCombatControlEndsWhenTheTurnEndsMidCombat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleCombatControl(t, g, 0, 1)
	advanceToStepOfSeat(t, g, 1, StepDeclareAttackers)
	if controllerOfSeat(g, 1) != 0 {
		t.Fatal("setup: seat 1's combat is not controlled")
	}
	passTurn(t, g)
	if len(g.PlayerControls) != 0 {
		t.Fatalf("records left after the turn ended mid-combat: %+v", g.PlayerControls)
	}
	if n := len(turnControlEvents(g, EventPlayerControlEnded)); n != 1 {
		t.Errorf("EventPlayerControlEnded emitted %d times, want 1", n)
	}
	advanceToStepOfSeat(t, g, 1, StepDeclareAttackers)
	if got := controllerOfSeat(g, 1); got != 1 {
		t.Errorf("seat 1's next combat is controlled by seat %d", got)
	}
}

// A turn with no combat phase that begins (skipped by a turn ended
// before it) leaves the record waiting for the next one.
func TestCombatControlWaitsThroughATurnWhoseCombatNeverBegins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleCombatControl(t, g, 0, 1)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	passTurn(t, g)
	if n := len(g.ScheduledPlayerControls()); n != 1 {
		t.Fatalf("scheduled after a turn with no combat = %d, want the record still waiting", n)
	}
	advanceToStepOfSeat(t, g, 1, StepBeginCombat)
	if got := controllerOfSeat(g, 1); got != 0 {
		t.Errorf("seat 1's next combat is controlled by seat %d, want seat 0", got)
	}
}

// Leaving the game: a controlled player who leaves before their combat
// takes the record with them; a controller who leaves controls nobody,
// and the record is still spent on the phase.
func TestCombatControlLeavingTheGame(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleCombatControl(t, g, 0, 1)
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	g.WithWriteLock(g.prunePlayerControlsLocked)
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left after the controlled player left: %+v", g.PlayerControls)
	}

	g = newFourPlayerActiveGame(t)
	scheduleCombatControl(t, g, 3, 1)
	if err := g.Concede(g.Seats[3].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := len(g.ScheduledPlayerControls()); n != 0 {
		t.Errorf("a control whose controller left is still scheduled in the view (%d)", n)
	}
	advanceToStepOfSeat(t, g, 1, StepDeclareAttackers)
	if got := controllerOfSeat(g, 1); got != 1 {
		t.Errorf("seat 1 controlled by seat %d after the controller left", got)
	}
	if n := len(turnControlEvents(g, EventPlayerControlBegan)); n != 0 {
		t.Errorf("a began line for a controller who left (%d)", n)
	}
	advanceToStepOfSeat(t, g, 1, StepPostcombatMain)
	if len(g.PlayerControls) != 0 {
		t.Errorf("the record outlived its phase: %+v", g.PlayerControls)
	}
}

// The record is plain data: it survives a snapshot round trip waiting
// and live, and the restored live record ends with its phase.
func TestCombatControlSnapshotRoundTrip(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleCombatControl(t, g, 0, 1)
	scheduleCombatControl(t, g, 3, 2)
	advanceToStepOfSeat(t, g, 1, StepDeclareAttackers)

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	back, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !reflect.DeepEqual(back.PlayerControls, g.PlayerControls) {
		t.Fatalf("restored records\n %+v\nwant\n %+v", back.PlayerControls, g.PlayerControls)
	}
	if back.PlayerControls[0].Stage != ControlCombat || back.PlayerControls[1].Stage != ControlAwaitingCombat {
		t.Fatalf("stages = %q, %q", back.PlayerControls[0].Stage, back.PlayerControls[1].Stage)
	}
	if controllerOfSeat(back, 1) != 0 {
		t.Error("the restored game lost the live control")
	}
	advanceToStepOfSeat(t, back, 1, StepPostcombatMain)
	if controllerOfSeat(back, 1) != 1 || len(back.PlayerControls) != 1 {
		t.Errorf("restored game after combat: records %+v", back.PlayerControls)
	}
}

func TestBeginCombatControlRefusals(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if err := g.Concede(g.Seats[3].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	for _, cc := range []CombatControl{
		{Controller: g.Seats[0].ID, Controlled: g.Seats[3].ID},
		{Controller: g.Seats[3].ID, Controlled: g.Seats[1].ID},
		{Controller: g.Seats[1].ID, Controlled: g.Seats[1].ID},
	} {
		var ok bool
		g.WithWriteLock(func() { ok = g.BeginCombatControlForEffect(cc) })
		if ok {
			t.Errorf("BeginCombatControlForEffect(%+v) was accepted", cc)
		}
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records = %+v", g.PlayerControls)
	}
}

// PlaysForThem: the two lengths a controller plays the game for.
func TestControlUntilPlaysForThem(t *testing.T) {
	for u, want := range map[ControlUntil]bool{
		ControlUntilResolution: false,
		ControlUntilSearch:     false,
		ControlUntilTurn:       true,
		ControlUntilCombat:     true,
		"":                     false,
	} {
		if got := u.PlaysForThem(); got != want {
			t.Errorf("%q.PlaysForThem() = %v, want %v", u, got, want)
		}
	}
}
