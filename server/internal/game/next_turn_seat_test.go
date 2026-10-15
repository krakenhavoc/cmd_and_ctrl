package game

import (
	"testing"

	"github.com/google/uuid"
)

// next_turn_seat_test.go — ADR 0188 Decision 9, #3566: NextTurnSeatForEffect
// names the seat whose turn really begins next. Every case here names the
// expected seat by hand AND walks the real rotation seam with PassTurn, so
// the read-only simulation and beginNextTurnLocked cannot drift apart.

// nextSeat reads the prediction under the read lock.
func nextSeat(t *testing.T, g *Game) int {
	t.Helper()
	seat, ok := g.NextTurnSeat()
	if !ok {
		t.Fatalf("NextTurnSeat has no answer on turn %d (seat %d, %s)", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
	return seat
}

// predictThenPass asserts the prediction is `want`, ends the turn, and
// asserts that seat is the one that took it.
func predictThenPass(t *testing.T, g *Game, want int) {
	t.Helper()
	if got := nextSeat(t, g); got != want {
		t.Fatalf("turn %d (seat %d, extra %v): next seat %d, want %d", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Extra, got, want)
	}
	passTurn(t, g)
	if g.Turn.ActiveSeat != want {
		t.Fatalf("the prediction said seat %d, the seam began seat %d's turn", want, g.Turn.ActiveSeat)
	}
}

// TestNextSeatDuringAnExtraTurnResumesFromTheInterruptedSeat is the live
// bot bug: seat 0 gives seat 2 an extra turn in a four-player game. In
// seat 2's extra turn the next turn is seat 1's, where a rotation from
// the ACTIVE seat names seat 3.
func TestNextSeatDuringAnExtraTurnResumesFromTheInterruptedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	takeExtraTurns(t, g, 2, 1)

	predictThenPass(t, g, 2) // seat 0's turn: the queued extra turn is next
	if !g.Turn.Extra {
		t.Fatalf("seat 2's turn should be the extra one")
	}
	predictThenPass(t, g, 1) // seat 2's EXTRA turn: seat 1, not seat 3
	predictThenPass(t, g, 2) // seat 1's normal turn: seat 2's normal turn
	predictThenPass(t, g, 3)
	predictThenPass(t, g, 0)
}

// TestNextSeatWithNothingQueuedIsTheNextSeat is the plain case.
func TestNextSeatWithNothingQueuedIsTheNextSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	for _, want := range []int{1, 2, 3, 0, 1} {
		predictThenPass(t, g, want)
	}
}

// TestNextSeatPassesASkipCharge: seat 1 owes a skip, so from seat 0 the
// next turn is seat 2's; the prediction writes nothing, so the charge is
// still owed until the seam spends it.
func TestNextSeatPassesASkipCharge(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 1, uuid.Nil, 1)

	if got := nextSeat(t, g); got != 2 {
		t.Fatalf("next seat %d, want 2 (seat 1 skips)", got)
	}
	if n := g.TurnSkipsFor(g.Seats[1].ID); n != 1 {
		t.Fatalf("asking for the next seat spent a charge: seat 1 owes %d, want 1", n)
	}
	predictThenPass(t, g, 2)
	predictThenPass(t, g, 3)
	predictThenPass(t, g, 0)
	predictThenPass(t, g, 1) // the charge is spent: seat 1 plays again
}

// TestNextSeatPassesEveryChargeOfTwoSkips: two skips skip two turns
// (CR 614.10a), so seat 1 is passed over on two successive rotations.
func TestNextSeatPassesEveryChargeOfTwoSkips(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 1, uuid.Nil, 2)

	for _, want := range []int{2, 3, 0, 2, 3, 0, 1} {
		predictThenPass(t, g, want)
	}
}

// TestNextSeatPassesADepartedSeat: seat 1 has left, so seat 0's next turn
// is seat 2's.
func TestNextSeatPassesADepartedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	for _, want := range []int{2, 3, 0, 2} {
		predictThenPass(t, g, want)
	}
}

// TestNextSeatSkipsTheQueuedTurnOfADepartedSeat: the queued extra turn of
// a seat that has left never begins, so the prediction goes past it.
func TestNextSeatSkipsTheQueuedTurnOfADepartedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	takeExtraTurns(t, g, 3, 1)
	if err := g.Concede(g.Seats[3].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	predictThenPass(t, g, 1)
	predictThenPass(t, g, 2)
}

// TestNextSeatStackedExtraTurns: the most recently created extra turn is
// taken first (CR 500.7), and rotation resumes from the interrupted seat
// once the stack is empty.
func TestNextSeatStackedExtraTurns(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	takeExtraTurns(t, g, 3, 1)
	takeExtraTurns(t, g, 2, 1) // created last, taken first
	takeExtraTurns(t, g, 3, 1)

	predictThenPass(t, g, 3)
	predictThenPass(t, g, 2)
	predictThenPass(t, g, 3)
	predictThenPass(t, g, 1) // the stack is empty: seat 0's rotation resumes
	predictThenPass(t, g, 2)
}

// TestNextSeatSkipsTheExtraTurnOfASeatOwingASkip: the Magosi ruling. The
// next turn seat 2 would begin is its extra turn, so that is the one its
// debt skips; the prediction names the seat after it, and the charge is
// gone by the time seat 2's normal turn comes.
func TestNextSeatSkipsTheExtraTurnOfASeatOwingASkip(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	takeExtraTurns(t, g, 2, 1)
	skipTurns(t, g, 2, uuid.Nil, 1)

	predictThenPass(t, g, 1) // seat 2's extra turn is skipped
	predictThenPass(t, g, 2) // and seat 2's normal turn is not
}

// TestNextSeatIsAbsentWhenTheGameIsNotInPlay: a game that has not
// started or has ended has no next turn to name.
func TestNextSeatIsAbsentWhenTheGameIsNotInPlay(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() { g.State = StateEnded })
	if seat, ok := g.NextTurnSeat(); ok {
		t.Fatalf("an ended game names seat %d as the next turn", seat)
	}
}
