package game

import (
	"testing"
)

// duration_end_of_combat_next_turn_test.go pins #6093's duration: "until
// end of combat on your next turn" (Brazen Cannonade) outlives the turn it
// is made in and every other seat's turn, lasts through the beginning,
// first main phase and combat of its player's next turn, and is over once
// that combat phase has ended (CR 511.3). Driven through the real step
// machine on a four-seat table.

func endOfCombatNextTurnExpired(g *Game, d Duration) bool {
	var expired bool
	g.ReadSnapshot(func() { expired = g.durationExpiredLocked(d, false) })
	return expired
}

func TestUntilEndOfCombatOnYourNextTurnEndsAsThatCombatEnds(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilEndOfCombatOnYourNextTurnDuration(me.ID) })
	if d.Kind != UntilEndOfCombatOnYourNextTurn || !d.Known() {
		t.Fatalf("duration = %+v, want a known until-end-of-combat-on-your-next-turn window", d)
	}
	bear := pushScopedTestCreature(g, me.ID, 2, 2)
	registerDurationMarker(g, bear, d, "until end of combat on your next turn")
	stored := func() int {
		n := 0
		g.ReadSnapshot(func() {
			for _, e := range g.ScopedEffects {
				if e.Duration.Kind == UntilEndOfCombatOnYourNextTurn {
					n++
				}
			}
		})
		return n
	}
	advanceToStepOfSeat(t, g, 0, StepEnd)
	if endOfCombatNextTurnExpired(g, d) {
		t.Fatal("the window closed in the turn it was made in")
	}
	g.WithWriteLock(func() {
		if g.durationExpiredLocked(d, true) {
			t.Fatal("the window closed at the cleanup of the turn it was made in")
		}
	})
	for seat := 1; seat <= 3; seat++ {
		advanceToStepOfSeat(t, g, seat, StepPostcombatMain)
		if endOfCombatNextTurnExpired(g, d) {
			t.Fatalf("the window closed during seat %d's turn", seat)
		}
	}
	// The turn-start sweep of a stored record runs before the new turn's
	// tally is emptied, with the last turn's combat still in it: a record
	// with this duration must survive it.
	for _, step := range []Step{StepUpkeep, StepDraw, StepPrecombatMain, StepBeginCombat, StepEndCombat} {
		advanceToStepOfSeat(t, g, 0, step)
		if endOfCombatNextTurnExpired(g, d) {
			t.Fatalf("the window closed at %s of its player's next turn, before that combat ended", step)
		}
		if stored() != 1 {
			t.Fatalf("the stored record was swept by %s of its player's next turn", step)
		}
	}
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	if !endOfCombatNextTurnExpired(g, d) {
		t.Fatal("the window is still open in the postcombat main phase of its player's next turn")
	}
}

// Made on another seat's turn, "your next turn" is still the player's
// next turn, and any later turn of theirs is past it.
func TestUntilEndOfCombatOnYourNextTurnMadeOnAnotherTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	advanceToStepOfSeat(t, g, 2, StepPrecombatMain)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilEndOfCombatOnYourNextTurnDuration(me.ID) })
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	if endOfCombatNextTurnExpired(g, d) {
		t.Fatal("the window closed before its player's next combat")
	}
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if !endOfCombatNextTurnExpired(g, d) {
		t.Fatal("the window outlived its player's next turn")
	}
}

// The kind keeps its persisted int, names itself, and refuses a stamp
// that belongs to another kind.
func TestUntilEndOfCombatOnYourNextTurnKeepsItsPersistedInt(t *testing.T) {
	// #6229 appended UntilACreatureSpellIsCast after it.
	if !UntilEndOfCombatOnYourNextTurn.Known() || UntilEndOfCombatOnYourNextTurn != ThisTurnAndNextTurn+1 || UntilEndOfCombatOnYourNextTurn+1 != UntilACreatureSpellIsCast {
		t.Error("UntilEndOfCombatOnYourNextTurn is not right after ThisTurnAndNextTurn, followed by UntilACreatureSpellIsCast")
	}
	if int(UntilEndOfCombatOnYourNextTurn) != 14 {
		t.Errorf("UntilEndOfCombatOnYourNextTurn persists as %d, want 14", UntilEndOfCombatOnYourNextTurn)
	}
	if got := UntilEndOfCombatOnYourNextTurn.String(); got != "until end of combat on your next turn" {
		t.Errorf("String() = %q", got)
	}
	if (Duration{Kind: UntilEndOfCombatOnYourNextTurn, AfterTurn: 3}).Known() {
		t.Error("a next-turn stamp on this kind is accepted")
	}
}
