package game

import "testing"

// #2385: a turn that is ended (CR 724.1) skips its end step, so an
// "until your next end step" window made before it carries to the
// player's next end step that does begin (CR 611.2b, CR 724.1d).
func TestUntilNextEndStepSurvivesTheTurnBeingEnded(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	id, _ := grantUntilNextEndStep(t, g, 0)
	me := g.Seats[0]

	// A hand at the limit, so cleanup has nothing to wait on.
	for me.Hand.Size() > 7 {
		if _, err := me.Hand.PopTop(); err != nil {
			t.Fatalf("PopTop: %v", err)
		}
	}
	g.WithWriteLock(func() { g.EndTheTurnForEffect(me.ID) })
	g.SettleResolution()
	if g.Turn.ActiveSeat == 0 {
		t.Fatalf("the turn did not end: seat %d at %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("the ended turn's skipped end step closed the window")
	}
	for seat := 1; seat <= 3; seat++ {
		advanceToStepOfSeat(t, g, seat, StepEnd)
		if !castableFromExileBy(g, id, me.ID) {
			t.Fatalf("the window closed in seat %d's end step", seat)
		}
	}
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("the window closed before my next end step")
	}
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if castableFromExileBy(g, id, me.ID) {
		t.Error("the window outlived my next real end step")
	}
}
