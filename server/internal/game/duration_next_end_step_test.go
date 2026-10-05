package game

import (
	"testing"

	"github.com/google/uuid"
)

// duration_next_end_step_test.go — #2373: "until your next end step",
// Ob Nixilis, Captive Kingpin's window. It ends as the NEXT end step of
// the named player's begins, which is this turn's when the grant is
// made on their own turn before the end step, and their next turn's
// otherwise. Driven through the real step machine, with a four-seat
// table so "an opponent's end step" means something.

// grantUntilNextEndStep drops an exiled instant for seat's player,
// stamped by the constructor under test, and returns its ID.
func grantUntilNextEndStep(t *testing.T, g *Game, seat int) (card uuid.UUID, d Duration) {
	t.Helper()
	p := g.Seats[seat]
	g.WithWriteLock(func() { d = g.UntilYourNextEndStepDuration(p.ID) })
	id := exiledWithGrant(t, g, p, "Impulsed Card", CastPermission{Player: p.ID, Duration: d})
	return id, d
}

func TestUntilNextEndStepFromYourMainEndsAtThisTurnsEndStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	id, d := grantUntilNextEndStep(t, g, 0)
	me := g.Seats[0]

	var turns int
	g.ReadSnapshot(func() { turns = g.turnsBegunForLocked(me.ID) })
	if d.Kind != UntilYourNextEndStep || d.ExpiresAtTurnsBegun != turns {
		t.Fatalf("duration = %+v, want this turn (%d) as the closing turn", d, turns)
	}
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("the window closed before the end step")
	}
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if castableFromExileBy(g, id, me.ID) {
		t.Error("the window is still open in the end step it should have closed at")
	}
	if n := len(me.CastPermissions); n != 0 {
		t.Errorf("the end-step sweep left %d permissions behind", n)
	}
}

func TestUntilNextEndStepFromAnOpponentsTurnSurvivesTheirEndStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	id, d := grantUntilNextEndStep(t, g, 0)
	me := g.Seats[0]

	var turns int
	g.ReadSnapshot(func() { turns = g.turnsBegunForLocked(me.ID) })
	if d.ExpiresAtTurnsBegun != turns+1 {
		t.Fatalf("closing turn = %d, want my NEXT turn (%d)", d.ExpiresAtTurnsBegun, turns+1)
	}
	advanceToStepOfSeat(t, g, 1, StepEnd)
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("an opponent's end step closed my window")
	}
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("the window closed before my next end step")
	}
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if castableFromExileBy(g, id, me.ID) {
		t.Error("the window outlived my next end step")
	}
}

func TestUntilNextEndStepFromYourEndStepWaitsForTheNextOne(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepEnd)
	id, d := grantUntilNextEndStep(t, g, 0)
	me := g.Seats[0]

	var turns int
	g.ReadSnapshot(func() { turns = g.turnsBegunForLocked(me.ID) })
	if d.ExpiresAtTurnsBegun != turns+1 {
		t.Fatalf("closing turn = %d, want my next turn (%d)", d.ExpiresAtTurnsBegun, turns+1)
	}
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("a window made in the end step is closed at once")
	}
	// Through the rest of this turn, three opponents' turns and their
	// end steps.
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
		t.Error("the window outlived my next end step")
	}
}

func TestUntilNextEndStepSurvivesCloneAndSnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	id, d := grantUntilNextEndStep(t, g, 0)
	me := g.Seats[0]

	for _, tc := range []struct {
		name string
		make func(t *testing.T) *Game
	}{
		{"clone", func(t *testing.T) *Game { return g.Clone() }},
		{"snapshot round-trip", func(t *testing.T) *Game {
			restored, err := g.CaptureSnapshot().Restore()
			if err != nil {
				t.Fatalf("Restore: %v", err)
			}
			return restored
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.make(t)
			got := exilePlayOf(out, id).Duration
			if !got.Equal(d) {
				t.Fatalf("duration came back as %+v, want %+v", got, d)
			}
			if !castableFromExileBy(out, id, me.ID) {
				t.Error("the restored window is not live")
			}
			advanceToStepOfSeat(t, out, 0, StepEnd)
			if castableFromExileBy(out, id, me.ID) {
				t.Error("the restored window outlived the end step")
			}
		})
	}
}
