package aiseat_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// two_seat_test.go — #1096. The soak now deals two seats as well as
// four (AISEAT_SOAK_SEATS), but a fuzzer is a tripwire, not a
// specification: it tells you something broke, at a seed, after a
// hundred games. These are the two shapes the issue actually names,
// pinned deterministically, in one test that needs no bot, no runner
// and no soak:
//
//   - the PRIORITY CURSOR has exactly two positions. There is no third
//     seat for it to pass through, so "everyone has passed in
//     succession" (CR 117.4) is reached after two passes rather than
//     four, and the pass that reaches it must advance the step rather
//     than hand priority back to the player who just passed.
//   - the TURN WRAPS between the same two seats. `Turn.Number` counts
//     ROUNDS rather than turns (see game/turn.go), so at two seats it
//     advances every SECOND turn, and the seat that follows seat 1 is
//     seat 0 — the rotation the four-seat soak can never exercise,
//     because four-minus-two is not two.
//
// Deliberately in this package rather than in internal/game: it is the
// aiseat soak's table that these facts are about, `newRoom` is the
// harness that builds it, and a two-seat table built any other way
// would be pinning a different room than the one being fuzzed.

// advanceUntilPriority walks the step machine forward until some seat
// holds priority, so the assertions below start from a step that
// grants it (Untap and Cleanup do not).
func advanceUntilPriority(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if g.Turn.PriorityHolder != game.NoPriority {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("no step granted priority within 40 advances")
}

// TestTwoSeatPriorityCursorHasExactlyTwoPositions pins the first
// shape: with an empty stack, one pass hands priority to the only
// opponent and the next completes the round.
func TestTwoSeatPriorityCursorHasExactlyTwoPositions(t *testing.T) {
	room := newRoom(t, 2, 4242)
	g := room.Game
	if len(g.Seats) != 2 {
		t.Fatalf("newRoom(2) dealt %d seats", len(g.Seats))
	}
	advanceUntilPriority(t, g)

	active := g.Turn.ActiveSeat
	if g.Turn.PriorityHolder != active {
		t.Fatalf("priority starts with seat %d, want the active seat %d (CR 117.3a)",
			g.Turn.PriorityHolder, active)
	}
	other := 1 - active
	step, turn := g.Turn.Step, g.Turn.Number

	if err := g.PassPriority(); err != nil {
		t.Fatalf("first pass: %v", err)
	}
	if got := g.Turn.PriorityHolder; got != other {
		t.Fatalf("after one pass priority is with seat %d, want the only opponent, seat %d", got, other)
	}
	if g.Turn.Step != step {
		t.Fatalf("the first pass advanced the step to %s; only a completed round may (CR 117.4)", g.Turn.Step)
	}

	// The second pass completes the round: two seats, two passes.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("second pass: %v", err)
	}
	if g.Turn.Step == step && g.Turn.Number == turn {
		t.Errorf("after both seats passed in succession the game is still in %s of turn %d; "+
			"the round never completed", step, turn)
	}
}

// TestTwoSeatTurnWrapsBetweenTheSameTwoSeats pins the second: the
// rotation alternates, and Turn.Number — a ROUND counter — advances
// every second turn rather than every turn.
func TestTwoSeatTurnWrapsBetweenTheSameTwoSeats(t *testing.T) {
	room := newRoom(t, 2, 4242)
	g := room.Game
	if len(g.Seats) != 2 {
		t.Fatalf("newRoom(2) dealt %d seats", len(g.Seats))
	}

	// Four consecutive turns is two full rounds, which is the smallest
	// sample that can tell "alternates" from "sticks" and "counts
	// rounds" from "counts turns".
	type observed struct {
		seat  int
		round int
	}
	seen := []observed{{g.Turn.ActiveSeat, g.Turn.Number}}
	for len(seen) < 4 {
		before := g.Turn.ActiveSeat
		for i := 0; g.Turn.ActiveSeat == before; i++ {
			if i >= 400 {
				t.Fatalf("the turn never left seat %d", before)
			}
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		seen = append(seen, observed{g.Turn.ActiveSeat, g.Turn.Number})
	}

	for i := 1; i < len(seen); i++ {
		if seen[i].seat == seen[i-1].seat {
			t.Fatalf("turn %d and %d were both seat %d; with two seats the rotation must alternate",
				i-1, i, seen[i].seat)
		}
		if seen[i].seat != 1-seen[i-1].seat {
			t.Fatalf("seat %d followed seat %d; at two seats the only successor is the other one",
				seen[i].seat, seen[i-1].seat)
		}
	}
	// Rounds: seats 0,1 are round N; the next 0,1 are round N+1.
	if seen[1].round != seen[0].round {
		t.Errorf("the round number moved from %d to %d inside one round (seat %d → seat %d); "+
			"Turn.Number counts rounds, not turns",
			seen[0].round, seen[1].round, seen[0].seat, seen[1].seat)
	}
	if seen[2].round != seen[0].round+1 {
		t.Errorf("after a full rotation the round number is %d, want %d",
			seen[2].round, seen[0].round+1)
	}
}
