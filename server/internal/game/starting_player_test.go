package game

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestChooseStartingSeatRerollsOnlyTiedLeaders(t *testing.T) {
	rolls := map[int][]int{
		0: {20, 3},
		1: {5},
		2: {20, 17},
		3: {1},
	}
	calls := make([]int, 0, 6)
	winner := chooseStartingSeat(4, func(seat int) int {
		calls = append(calls, seat)
		result := rolls[seat][0]
		rolls[seat] = rolls[seat][1:]
		return result
	})
	if winner != 2 {
		t.Fatalf("winner = seat %d, want seat 2", winner)
	}
	if want := []int{0, 1, 2, 3, 0, 2}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("roll calls = %v, want %v", calls, want)
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
