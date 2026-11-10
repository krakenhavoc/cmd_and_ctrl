package game

import "testing"

// duration_until_next_end_step_test.go — "until the next end step" (Niko,
// Light of Hope): the window ends as the next end step of ANY turn
// begins, this turn's when it has not begun and the next turn's
// otherwise. Driven through the real step machine on a four-seat table.

func TestUntilNextEndStepEndsAsTheNextEndStepOfAnyTurnBegins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilNextEndStepDuration() })
	expired := func() bool {
		var e bool
		g.ReadSnapshot(func() { e = g.durationExpiredLocked(d, false) })
		return e
	}
	if d.Kind != UntilNextEndStep || !d.Known() {
		t.Fatalf("duration = %+v, want a known until-the-next-end-step window", d)
	}
	if expired() {
		t.Fatal("over as soon as it was made")
	}
	advanceToStepOfSeat(t, g, 1, StepPostcombatMain)
	if expired() {
		t.Fatal("over before the end step")
	}
	advanceToStepOfSeat(t, g, 1, StepEnd)
	if !expired() {
		t.Fatal("still running in the end step it should have closed at")
	}
}

func TestUntilNextEndStepMadeInAnEndStepWaitsForTheFollowingOne(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepEnd)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilNextEndStepDuration() })
	expired := func() bool {
		var e bool
		g.ReadSnapshot(func() { e = g.durationExpiredLocked(d, false) })
		return e
	}
	if expired() {
		t.Fatal("a window made in the end step is over at once")
	}
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if expired() {
		t.Fatal("over before the next end step")
	}
	advanceToStepOfSeat(t, g, 1, StepEnd)
	if !expired() {
		t.Error("the window outlived the next end step, another player's")
	}
}

func TestUntilNextEndStepKeepsItsPersistedIntAndNamesItself(t *testing.T) {
	if !UntilNextEndStep.Known() || UntilNextEndStep != UntilPlaneswalkAway+1 {
		t.Error("UntilNextEndStep does not follow UntilPlaneswalkAway")
	}
	if got := UntilNextEndStep.String(); got != "until the next end step" {
		t.Errorf("String() = %q", got)
	}
}
