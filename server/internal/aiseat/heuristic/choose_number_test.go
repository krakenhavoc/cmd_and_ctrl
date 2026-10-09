package heuristic_test

import (
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// choose_number_test.go — ADR 0129's amendment of 2026-10-09 (#1941):
// the heuristic's answers to a number paid in life, or also dealt to the
// chooser.

// rankAmounts ranks each offered amount of a pay_amount prompt for seat 0
// at `life`.
func rankAmounts(t *testing.T, life int, pa protocol.PayAmountView, amounts []int) map[int]float64 {
	t.Helper()
	pol := heuristic.New()
	const id = "33333333-3333-3333-3333-333333333333"
	v := newView([]protocol.PlayerView{newSeat(0, withLife(life)), newSeat(1)}, withChoice(protocol.PendingChoiceView{
		ID: id, Kind: "pay_amount", Chooser: seatID(0).String(), PayAmount: &pa,
	}))
	in := input(0, v)
	for _, n := range amounts {
		in.Moves = append(in.Moves, choiceMove(t, 0, id, fmt.Sprintf("answer %d", n), map[string]any{"amount": n}))
	}
	out := map[int]float64{}
	for i, n := range amounts {
		out[n] = rankValue(t, pol, in, in.Moves[i].Label)
	}
	return out
}

// Volcano Hellion: the lethal amount when it leaves the seat at 10 or
// more, and nothing otherwise; never its own life total.
func TestChooseNumberSelfDamageTakesTheGoalOnlyAboveTheFloor(t *testing.T) {
	pa := protocol.PayAmountView{Min: 0, Max: 1_000_000, NoMax: true, Goal: 4, Marks: []int{4, 40},
		Unit: "damage", Resource: "none", SelfDamage: true}
	rich := rankAmounts(t, 40, pa, []int{0, 4, 40})
	if rich[4] <= rich[0] || rich[4] <= rich[40] {
		t.Errorf("at 40 life: %v, want the goal of 4 on top", rich)
	}
	if rich[40] >= rich[0] {
		t.Errorf("at 40 life: dealing itself 40 (%.3f) is not below nothing (%.3f)", rich[40], rich[0])
	}
	pa.Marks = []int{4, 12}
	poor := rankAmounts(t, 12, pa, []int{0, 4, 12})
	if poor[0] <= poor[4] {
		t.Errorf("at 12 life: %v, want nothing above the goal (it would leave 8)", poor)
	}
}

// A life payment: the card's goal when affordable, and nothing when
// there is no goal.
func TestPayLifeAmountPaysTheGoal(t *testing.T) {
	pa := protocol.PayAmountView{Min: 0, Max: 40, Goal: 20, Unit: "power", Resource: "life"}
	r := rankAmounts(t, 40, pa, []int{0, 20, 40})
	if r[20] <= r[0] || r[20] <= r[40] {
		t.Errorf("%v, want paying 20 on top", r)
	}
	pa.Goal = 0
	r = rankAmounts(t, 40, pa, []int{0, 40})
	if r[0] <= r[40] {
		t.Errorf("no goal: %v, want paying nothing on top", r)
	}
}
