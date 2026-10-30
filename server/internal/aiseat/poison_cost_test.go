package aiseat_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// poison_cost_test.go — #4613: a confirm whose accept branch gives the seat
// poison counters ("Ward—Get five poison counters") reaches the bot priced.
// The prompt declares PoisonCost, the enumerator stamps it on the accept move
// (legal.MoveCost.Poison) and the policy weighs it against the ten that lose
// the game (CR 704.5c): a payment that would reach ten is never taken, and a
// small one is taken only while it is cheap.

// poisonChoice queues "get N poison counters" on a seat that already has
// `have`, lets the bot answer it, and reports whether it took the accept
// branch and the labels it used.
func poisonChoice(t *testing.T, have, n int) (accepted bool, took []string) {
	t.Helper()
	g := newSettledTable(t, 13)
	seat := g.Seats[0]
	if have > 0 {
		if err := g.SetPoison(seat.ID, have); err != nil {
			t.Fatalf("SetPoison: %v", err)
		}
	}
	g.WithWriteLock(func() {
		queueConfirmHooked(g, game.ConfirmPrompt{
			Chooser:      seat.ID,
			Question:     "Ward — get poison counters, or it is countered",
			AcceptLabel:  "Get poison counters",
			DeclineLabel: "Let it be countered",
			PoisonCost:   n,
		}, func(*game.Game) error {
			accepted = true
			return nil
		}, nil)
	})
	took = driveChoices(t, g, heuristic.New(), seat.ID, 4)
	return accepted, took
}

func TestHeuristicDeclinesAFivePoisonWard(t *testing.T) {
	for _, have := range []int{0, 3, 5, 9} {
		accepted, took := poisonChoice(t, have, 5)
		if accepted {
			t.Errorf("at %d poison the bot took five more; it took %v", have, took)
		}
		if len(took) != 1 || !strings.Contains(took[0], "Let it be countered") {
			t.Errorf("at %d poison took %v, want the decline branch", have, took)
		}
	}
}

// One counter is cheap on a clean seat and dear near ten. And a payment
// that reaches ten is refused at any size.
func TestHeuristicPricesASmallPoisonPaymentAgainstTheTrackToTen(t *testing.T) {
	for _, tc := range []struct {
		have, n int
		want    bool
	}{
		{0, 1, true},
		{8, 1, false},
		{9, 1, false}, // the tenth counter loses the game
		{6, 4, false}, // exactly ten
	} {
		accepted, took := poisonChoice(t, tc.have, tc.n)
		if accepted != tc.want {
			t.Errorf("at %d poison, +%d: accepted = %v, want %v (took %v)", tc.have, tc.n, accepted, tc.want, took)
		}
	}
}

func TestPoisonCostValueIsMonotoneAndZeroForNothing(t *testing.T) {
	w := heuristic.DefaultWeights()
	if got := w.PoisonCostValue(4, 0); got != 0 {
		t.Errorf("PoisonCostValue(4, 0) = %v, want 0", got)
	}
	prev := 0.0
	for pay := 1; pay <= 5; pay++ {
		v := w.PoisonCostValue(0, pay)
		if v <= prev {
			t.Errorf("PoisonCostValue(0, %d) = %v, not above %v", pay, v, prev)
		}
		prev = v
	}
	if near, far := w.PoisonCostValue(8, 1), w.PoisonCostValue(0, 1); near <= far {
		t.Errorf("a counter at 8 poison costs %v, no more than at 0 (%v)", near, far)
	}
}
