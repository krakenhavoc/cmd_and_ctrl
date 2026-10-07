package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// energy_payment_test.go — ADR 0129 §3 and §7: the heuristic's answers to
// energy paid while an ability resolves.

// A pay_amount prompt pays the card's threshold, and otherwise nothing.
func TestPayAmountPaysTheThreshold(t *testing.T) {
	pol := heuristic.New()
	const id = "11111111-1111-1111-1111-111111111111"
	amounts := []int{0, 1, 3, 7}
	rank := func(goal int) map[int]float64 {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withChoice(protocol.PendingChoiceView{
			ID: id, Kind: "pay_amount", Chooser: seatID(0).String(),
			PayAmount: &protocol.PayAmountView{Min: 1, Max: 7, Goal: goal, Unit: "damage"},
		}))
		in := input(0, v)
		for _, n := range amounts {
			in.Moves = append(in.Moves, choiceMove(t, 0, id, "pay "+string(rune('0'+n)), map[string]any{"amount": n}))
		}
		out := map[int]float64{}
		for i, n := range amounts {
			out[n] = rankValue(t, pol, in, in.Moves[i].Label)
		}
		return out
	}
	withGoal := rank(3)
	for _, n := range []int{0, 1, 7} {
		if withGoal[3] <= withGoal[n] {
			t.Errorf("goal 3: paying 3 (%.3f) is not above paying %d (%.3f)", withGoal[3], n, withGoal[n])
		}
	}
	noGoal := rank(0)
	for _, n := range []int{1, 3, 7} {
		if noGoal[0] <= noGoal[n] {
			t.Errorf("no goal: paying nothing (%.3f) is not above paying %d (%.3f)", noGoal[0], n, noGoal[n])
		}
	}
}

// A fixed energy payment is paid, priced at Weights.Energy: still above
// the decline, and a dearer payment is worth a little less.
func TestEnergyPaymentIsPaidAndPriced(t *testing.T) {
	pol := heuristic.New()
	const id = "22222222-2222-2222-2222-222222222222"
	value := func(n int) (pay, decline float64) {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withChoice(protocol.PendingChoiceView{
			ID: id, Kind: "pay_unless", Chooser: seatID(0).String(), PayEnergy: &n,
		}))
		in := input(0, v,
			choiceMove(t, 0, id, "pay", map[string]any{"apply": true}),
			choiceMove(t, 0, id, "decline", map[string]any{"apply": false}))
		return rankValue(t, pol, in, "pay"), rankValue(t, pol, in, "decline")
	}
	pay2, decline := value(2)
	pay8, _ := value(8)
	if pay2 <= decline || pay8 <= decline {
		t.Errorf("pay 2 %.3f, pay 8 %.3f, decline %.3f: paying must beat declining", pay2, pay8, decline)
	}
	if pay8 >= pay2 {
		t.Errorf("eight energy (%.3f) priced no higher than two (%.3f)", pay8, pay2)
	}
}
