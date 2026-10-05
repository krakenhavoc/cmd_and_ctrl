package game

import (
	"errors"
	"testing"
)

// mulligan_order_test.go — CR 103.5, issue #2237: the opening-hand
// decisions go in turn order, starting with the starting player, and a
// second round is only the seats that mulliganed.

// startMulliganTable seats n players, runs the opening roll and has the
// chooser give the first turn to `start`, which leaves the mulligan
// window open with hands dealt.
func startMulliganTable(t *testing.T, n, start int) *Game {
	t.Helper()
	g := openOpeningRollForTest(t, n, 11, 29)
	rollUntilChooser(t, g, func(_ int, seats []int) []int { return seats })
	chooser := g.OpeningRoll.Chooser
	if err := g.ChooseStartingPlayer(g.Seats[chooser].ID, start); err != nil {
		t.Fatalf("ChooseStartingPlayer: %v", err)
	}
	if !g.MulligansOpen || g.StartingSeat != start {
		t.Fatalf("window open %v, starting seat %d, want open and %d", g.MulligansOpen, g.StartingSeat, start)
	}
	return g
}

func keep(t *testing.T, g *Game, seat int) {
	t.Helper()
	if err := g.KeepHand(g.Seats[seat].ID); err != nil {
		t.Fatalf("seat %d keep: %v", seat, err)
	}
}

func mulligan(t *testing.T, g *Game, seat int) {
	t.Helper()
	if err := g.Mulligan(g.Seats[seat].ID, OpeningHandSize); err != nil {
		t.Fatalf("seat %d mulligan: %v", seat, err)
	}
}

func wantDecider(t *testing.T, g *Game, want int) {
	t.Helper()
	if got := g.MulliganDecider(); got != want {
		t.Fatalf("decider = seat %d, want %d", got, want)
	}
}

// Every table size and every starting seat: decisions go starting seat
// first and then round the table, nobody out of turn.
func TestMulliganOrderFollowsTheStartingSeat(t *testing.T) {
	for _, n := range []int{2, 4} {
		for start := 0; start < n; start++ {
			g := startMulliganTable(t, n, start)
			for step := 0; step < n; step++ {
				want := (start + step) % n
				wantDecider(t, g, want)
				for seat := 0; seat < n; seat++ {
					if seat == want {
						continue
					}
					if err := g.KeepHand(g.Seats[seat].ID); !g.Seats[seat].HandKept && !errors.Is(err, ErrNotYourMulligan) {
						t.Fatalf("%d seats, start %d: seat %d kept out of turn: %v", n, start, seat, err)
					}
					if err := g.Mulligan(g.Seats[seat].ID, OpeningHandSize); !errors.Is(err, ErrNotYourMulligan) {
						t.Fatalf("%d seats, start %d: seat %d mulliganed out of turn: %v", n, start, seat, err)
					}
				}
				keep(t, g, want)
			}
			if g.MulligansOpen {
				t.Fatalf("%d seats, start %d: the window is still open after every seat kept", n, start)
			}
			wantDecider(t, g, -1)
			// The turn machinery took over: the first turn is the starting seat's.
			if g.Turn.ActiveSeat != start || g.Turn.Step == StepUntap && g.Turn.PriorityHolder == NoPriority {
				t.Fatalf("%d seats, start %d: turn %+v after the window closed", n, start, g.Turn)
			}
		}
	}
}

// A mulligan is a decision: the round goes on past that seat, and only
// the seats that mulliganed decide again, in turn order again.
func TestMulliganRepeatsInTurnOrderForThoseWhoMulliganed(t *testing.T) {
	g := startMulliganTable(t, 4, 2)

	// Round 1: seats 2, 3, 0, 1.
	wantDecider(t, g, 2)
	mulligan(t, g, 2)
	wantDecider(t, g, 3)
	keep(t, g, 3)
	wantDecider(t, g, 0)
	mulligan(t, g, 0)
	wantDecider(t, g, 1)
	keep(t, g, 1)

	// Round 2: only seats 2 and 0, seat 2 first. Seat 0 does not get
	// to go ahead of it, and the keepers are not asked again.
	wantDecider(t, g, 2)
	if err := g.KeepHand(g.Seats[0].ID); !errors.Is(err, ErrNotYourMulligan) {
		t.Fatalf("seat 0 kept ahead of seat 2 in round 2: %v", err)
	}
	mulligan(t, g, 2)
	wantDecider(t, g, 0)
	keep(t, g, 0)

	// Round 3: seat 2 alone.
	wantDecider(t, g, 2)
	if got := g.Seats[2].MulligansTaken; got != 2 {
		t.Fatalf("seat 2 mulligans taken = %d, want 2", got)
	}
	keep(t, g, 2)

	if g.MulligansOpen {
		t.Fatal("the window should be closed")
	}
	// A stale second keep from a tab that already kept is a no-op.
	if err := g.KeepHand(g.Seats[3].ID); err != nil {
		t.Fatalf("repeat keep: %v", err)
	}
}

// A seat that has left is skipped, never waited for: before its turn,
// when it is the one deciding, and when it was the last holdout.
func TestMulliganSkipsAConcededSeat(t *testing.T) {
	t.Run("conceded before its turn", func(t *testing.T) {
		g := startMulliganTable(t, 4, 1)
		if err := g.Concede(g.Seats[2].ID); err != nil {
			t.Fatal(err)
		}
		wantDecider(t, g, 1)
		keep(t, g, 1)
		wantDecider(t, g, 3) // seat 2 is skipped
		keep(t, g, 3)
		wantDecider(t, g, 0)
		keep(t, g, 0)
		if g.MulligansOpen {
			t.Fatal("the window should close without seat 2")
		}
	})
	t.Run("conceded while deciding", func(t *testing.T) {
		g := startMulliganTable(t, 4, 3)
		keep(t, g, 3)
		wantDecider(t, g, 0)
		if err := g.Concede(g.Seats[0].ID); err != nil {
			t.Fatal(err)
		}
		wantDecider(t, g, 1)
	})
	t.Run("conceded as the last holdout closes the window", func(t *testing.T) {
		g := startMulliganTable(t, 3, 0)
		keep(t, g, 0)
		keep(t, g, 1)
		wantDecider(t, g, 2)
		if err := g.Concede(g.Seats[2].ID); err != nil {
			t.Fatal(err)
		}
		if g.State != StateActive {
			t.Skipf("the game ended with two seats left: %v", g.State)
		}
		if g.MulligansOpen {
			t.Fatal("the window stayed open for a seat that left")
		}
	})
	t.Run("conceded after mulliganing, before the next round", func(t *testing.T) {
		g := startMulliganTable(t, 4, 0)
		mulligan(t, g, 0)
		keep(t, g, 1)
		if err := g.Concede(g.Seats[0].ID); err != nil {
			t.Fatal(err)
		}
		// Seat 0 is gone, so round 1 continues with 2 and 3, and the
		// next round has nobody in it.
		wantDecider(t, g, 2)
		keep(t, g, 2)
		keep(t, g, 3)
		if g.MulligansOpen {
			t.Fatal("the window should close once the remaining seats have kept")
		}
	})
}

// Undo and the snapshot keep who has answered this round.
func TestMulliganRoundSurvivesCloneAndSnapshot(t *testing.T) {
	g := startMulliganTable(t, 4, 1)
	keep(t, g, 1)
	mulligan(t, g, 2)
	wantDecider(t, g, 3)

	c := g.Clone()
	wantDecider(t, c, 3)
	// Move on, then undo into the clone: the round is where it was.
	keep(t, g, 3)
	wantDecider(t, g, 0)
	g.RestoreFrom(c)
	wantDecider(t, g, 3)

	snap := g.CaptureSnapshot()
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	wantDecider(t, restored, 3)
	if !restored.Seats[2].MulliganDecided || restored.Seats[0].MulliganDecided {
		t.Fatalf("round flags after restore: seat 2 %v, seat 0 %v", restored.Seats[2].MulliganDecided, restored.Seats[0].MulliganDecided)
	}
	// A file written before the field existed restores as a fresh round:
	// the first seat from the starting seat that has not kept decides.
	for _, p := range restored.Seats {
		p.MulliganDecided = false
	}
	wantDecider(t, restored, 2)
}
