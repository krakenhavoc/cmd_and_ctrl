package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// pay_amount.go — ADR 0129 §7: answering "you may pay any amount of {E}",
// and, since the ADR's amendment of 2026-10-09 (#1941), "pay any amount
// of life" and "an amount of damage of your choice to you and target
// creature".
//
// The prompt carries the card's own threshold (`goal`): the smallest
// amount that reaches what the card is for — the target's toughness for
// Harnessed Lightning, every counter for Rampaging Aetherhood's +1/+1
// counters, the target's lethal damage for Volcano Hellion, the cards
// that refill Necrodominance's hand. The heuristic answers exactly that,
// and otherwise the floor (nothing paid, or the smallest number): a
// counter of energy is worth Weights.Energy kept, and an amount that
// falls short of the goal buys less than it costs.
//
// An answer paid in LIFE — a life payment, or a number that is also
// dealt to the chooser — is held to the floor Phyrexian life is held to
// (phyrexianLifeFloor): the goal is taken only when it leaves the seat at
// or above it, and an answer that would take the seat's last life is
// refused outright. The enumerator offers nothing, the smallest answer,
// the goal, the card's marks and the ceiling, and the floor is always
// one of them.

// payAmountValue ranks one offered amount.
func (st *state) payAmountValue(ch *protocol.PendingChoiceView, amount *int) (float64, string) {
	if ch == nil || ch.PayAmount == nil || amount == nil {
		return 0, "pay amount: unreadable"
	}
	pa := ch.PayAmount
	n := *amount
	floor := 0
	if pa.Resource == "none" {
		floor = pa.Min
	}
	spendsLife := pa.Resource == "life" || pa.SelfDamage
	if spendsLife && n > floor {
		left := st.myLife() - n
		switch {
		case left <= 0:
			return suicideValue, "pay amount: would pay its last life"
		case left < phyrexianLifeFloor:
			return phyrexianLifeDeclined, "pay amount: would take life below the floor"
		}
	}
	switch {
	case pa.Goal > 0 && n == pa.Goal:
		return 2, "pay amount: exactly the card's threshold"
	case n == floor:
		if spendsLife {
			return 1, "pay amount: keep the life"
		}
		return 1, "pay amount: keep the energy"
	default:
		// Short of the goal, or past it: spent for less than it is
		// worth. Less spent is less wasted.
		return 0.5 - float64(n)/1000, "pay amount: not the threshold"
	}
}
