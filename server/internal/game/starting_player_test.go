package game

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
)

// The tie rule, with no randomness in it: only the seats on the high
// result roll again, in seat order (ADR 0121 §1, decision 1).
func TestOpeningRoundLeadersAreOnlyTheTiedHighSeats(t *testing.T) {
	round := OpeningRollRound{
		Seats: []int{0, 1, 2, 3},
		Rolls: []OpeningRollDie{{Seat: 2, Result: 20}, {Seat: 0, Result: 20}, {Seat: 1, Result: 5}, {Seat: 3, Result: 1}},
	}
	leaders, high := openingRoundLeaders(round)
	if high != 20 || !reflect.DeepEqual(leaders, []int{0, 2}) {
		t.Fatalf("leaders = %v on %d, want [0 2] on 20", leaders, high)
	}
	round = OpeningRollRound{Seats: []int{0, 2}, Rolls: []OpeningRollDie{{Seat: 0, Result: 3}, {Seat: 2, Result: 17}}}
	if leaders, high = openingRoundLeaders(round); high != 17 || !reflect.DeepEqual(leaders, []int{2}) {
		t.Fatalf("leaders = %v on %d, want [2] on 17", leaders, high)
	}
}

func TestStartRollIsDeterministicAndPublic(t *testing.T) {
	type result struct {
		winner int
		rolls  [][2]int // seat, natural result
	}
	start := func() result {
		g := NewGame()
		for seat := 0; seat < 4; seat++ {
			if _, err := g.AddPlayer(fmt.Sprintf("P%d", seat+1), buildTestDeck("C")); err != nil {
				t.Fatalf("AddPlayer: %v", err)
			}
		}
		if err := g.StartWithFirstPlayerRoll(rand.New(rand.NewPCG(41, 97))); err != nil {
			t.Fatalf("StartWithFirstPlayerRoll: %v", err)
		}
		got := result{winner: g.StartingSeat}
		for _, ev := range g.Events {
			if ev.Kind != EventRollDie {
				continue
			}
			seat := -1
			for i, p := range g.Seats {
				if p.ID == ev.Actor {
					seat = i
					break
				}
			}
			if seat < 0 || ev.Sides != startingPlayerDieSides || ev.Source != [16]byte{} {
				t.Fatalf("opening roll event = %+v", ev)
			}
			got.rolls = append(got.rolls, [2]int{seat, ev.Amount})
		}
		if len(got.rolls) < len(g.Seats) {
			t.Fatalf("opening rolls = %v, want at least one per seat", got.rolls)
		}
		if g.Turn.ActiveSeat != got.winner || g.Turn.OrderSeat != got.winner {
			t.Fatalf("turn = %+v, winner seat = %d", g.Turn, got.winner)
		}
		return got
	}
	first, second := start(), start()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different opening rolls:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

func TestRoundReturnsToTheRolledStartingSeat(t *testing.T) {
	turn := newStartingTurn(2)
	want := []struct {
		seat  int
		round int
	}{{3, 1}, {0, 1}, {1, 1}, {2, 2}}
	for _, step := range want {
		turn.Step = StepCleanup
		turn = turn.advance(4, 2)
		if turn.ActiveSeat != step.seat || turn.Round != step.round {
			t.Fatalf("after rotation: seat=%d round=%d, want seat=%d round=%d", turn.ActiveSeat, turn.Round, step.seat, step.round)
		}
	}
}
