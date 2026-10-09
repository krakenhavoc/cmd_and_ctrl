package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
)

// deficitPolicy is the policy DeficitOpen asks: the shipped tuning, so
// every contestant's offers are judged by one rule.
var deficitPolicy = New()

// DeficitOpen reports whether the seat's mana deficit is open for the
// card with this instance ID in the window in: whether rampPremium
// would pay that card a premium for one more mana a turn (ADR 0126 §2:
// the largest mana value among the seat's OTHER cards in hand and its
// commander with its tax, capped at RampWantCap, above the mana its
// sources make). It is the computation the heuristic prices a rock
// with, under DefaultConfig, exported for the arena's A2 count of
// offers made while the deficit was open (ADR 0136's amendment of
// 2026-10-09, #2435). False when the seat's view has no hand.
func DeficitOpen(in aiseat.Input, cardID string) bool {
	// newState and rampFor read only the tuning, so no lock: arena
	// games call this from many goroutines.
	st := deficitPolicy.newState(in)
	c := st.mine[cardID]
	if c == nil {
		c = st.castSource(cardID)
	}
	if c == nil {
		return false
	}
	return deficitPolicy.rampFor(st, c, 1) > 0
}
