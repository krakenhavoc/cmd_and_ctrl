package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// pay_amount.go — ADR 0129 §7: answering "you may pay any amount of {E}".
//
// The prompt carries the card's own threshold (`goal`): the smallest
// amount that reaches what the card is for — the target's toughness for
// Harnessed Lightning, every counter for Rampaging Aetherhood's +1/+1
// counters. The heuristic pays exactly that, and otherwise pays nothing:
// a counter of energy is worth Weights.Energy kept, and an amount that
// falls short of the goal buys less than it costs. The enumerator offers
// nothing, the smallest payment, the goal and the ceiling, and paying
// nothing is always one of them.

// payAmountValue ranks one offered amount.
func payAmountValue(ch *protocol.PendingChoiceView, amount *int) (float64, string) {
	if ch == nil || ch.PayAmount == nil || amount == nil {
		return 0, "pay amount: unreadable"
	}
	n := *amount
	goal := ch.PayAmount.Goal
	switch {
	case goal > 0 && n == goal:
		return 2, "pay amount: exactly the card's threshold"
	case n == 0:
		return 1, "pay amount: keep the energy"
	default:
		// Short of the goal, or past it: energy spent for less than it
		// is worth. Less spent is less wasted.
		return 0.5 - float64(n)/1000, "pay amount: not the threshold"
	}
}
