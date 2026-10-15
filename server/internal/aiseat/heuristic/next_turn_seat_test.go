package heuristic

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// next_turn_seat_test.go — ADR 0188 Decision 10, #3566. The heuristic's
// nextTurnSeat used to rotate from the ACTIVE seat, so during seat 2's
// extra turn after seat 0's (four seats) it named seat 3 where the engine
// begins seat 1's turn. These views are shaped as the engine emits them
// (the engine side is pinned in game/next_turn_seat_test.go and
// protocol/next_seat_view_test.go).

func fourSeatView(active int, step string, next *int) protocol.GameView {
	var v protocol.GameView
	for i := 0; i < 4; i++ {
		v.Seats = append(v.Seats, protocol.PlayerView{ID: nsSeat(i).String(), Seat: i, Life: 40})
	}
	v.Turn.ActiveSeat = active
	v.Turn.Step = step
	v.Turn.NextSeat = next
	return v
}

func nsSeat(i int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i))
}

func intp(i int) *int { return &i }

// TestNextTurnSeatReadsTheEnginesAnswerInAnExtraTurn is the live bug:
// seat 2's extra turn after seat 0's, next_seat 1.
func TestNextTurnSeatReadsTheEnginesAnswerInAnExtraTurn(t *testing.T) {
	v := fourSeatView(2, "end", intp(1))
	v.Turn.Extra = true
	if got := nextTurnSeat(&v); got != 1 {
		t.Fatalf("nextTurnSeat = %d in seat 2's extra turn, want seat 1", got)
	}
	// The fallback for a view without the field is the walk from the
	// active seat, which is exactly the bug; it only serves a view an
	// older server built.
	v.Turn.NextSeat = nil
	if got := nextTurnSeat(&v); got != 3 {
		t.Fatalf("fallback nextTurnSeat = %d, want 3 (the walk from the active seat)", got)
	}
}

// TestBeforeMyUntapInTheExtraTurnsEndStep: the bot at seat 1 is told its
// turn is next at the end of seat 2's extra turn, and the bot at seat 3 is
// not.
func TestBeforeMyUntapInTheExtraTurnsEndStep(t *testing.T) {
	p := New()
	for _, c := range []struct {
		me   int
		want bool
	}{{1, true}, {3, false}, {0, false}, {2, false}} {
		v := fourSeatView(2, "end", intp(1))
		v.Turn.Extra = true
		st := p.newState(aiseat.Input{Seat: nsSeat(c.me), View: v})
		if st.beforeMyUntap != c.want {
			t.Errorf("seat %d: beforeMyUntap = %v, want %v", c.me, st.beforeMyUntap, c.want)
		}
	}
}

// TestBeforeMyUntapWhenMyExtraTurnIsNext: in seat 0's end step, with seat
// 0's own extra turn next, seat 0's untap is the next thing to happen.
func TestBeforeMyUntapWhenMyExtraTurnIsNext(t *testing.T) {
	p := New()
	v := fourSeatView(0, "end", intp(0))
	v.Turn.ExtraTurns = []int{0}
	st := p.newState(aiseat.Input{Seat: nsSeat(0), View: v})
	if !st.beforeMyUntap {
		t.Error("beforeMyUntap is false in seat 0's end step with seat 0's extra turn next")
	}
	if !st.leftover {
		t.Error("the leftover window is closed in the last window before the bot untaps")
	}
	// Not the end step: not the window.
	v = fourSeatView(0, "postcombat_main", intp(0))
	st = p.newState(aiseat.Input{Seat: nsSeat(0), View: v})
	if st.beforeMyUntap {
		t.Error("beforeMyUntap is true outside the end step")
	}
}

// TestNextTurnSeatNamesTheSeatAfterASkip: seat 1 owes a skip, so the
// engine's next_seat from seat 0 is seat 2; the bot at seat 1 is not
// "before my untap" even though it is next in seat order.
func TestNextTurnSeatNamesTheSeatAfterASkip(t *testing.T) {
	p := New()
	v := fourSeatView(0, "end", intp(2))
	v.Seats[1].TurnSkips = 1
	if got := nextTurnSeat(&v); got != 2 {
		t.Fatalf("nextTurnSeat = %d, want 2", got)
	}
	if st := p.newState(aiseat.Input{Seat: nsSeat(1), View: v}); st.beforeMyUntap {
		t.Error("the skipping seat is told its untap is next")
	}
	if st := p.newState(aiseat.Input{Seat: nsSeat(2), View: v}); !st.beforeMyUntap {
		t.Error("seat 2 is not told its untap is next")
	}
}

// TestNextTurnSeatOutOfRangeIsUnknown: a next_seat that names no seat in
// the view is "cannot say", not a guess.
func TestNextTurnSeatOutOfRangeIsUnknown(t *testing.T) {
	v := fourSeatView(0, "end", intp(9))
	if got := nextTurnSeat(&v); got != -1 {
		t.Fatalf("nextTurnSeat = %d for next_seat 9, want -1", got)
	}
}
