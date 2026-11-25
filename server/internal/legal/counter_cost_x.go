package legal

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// counterChoicesAtX is the counter payment and the X a move offers for a
// "{X}, …, Remove X <kind> counters from this permanent" cost (#6085,
// game.CounterCountIsX, Talon of Pain): ONE announced X pays both (CR
// 107.3a), so the move announces the largest X that the pool (x, the
// largest X the mana solve could afford) and the counters the source
// holds can both pay, and removes exactly that many. Nothing is offered
// (nil) when that X is under the enumerated floor — a card that declares
// XMatters is not offered at X = 0 (#810).
//
// choices is counterPaymentChoices' answer for the variable self form:
// one payment per kind the source holds of the printed kind (so at most
// one), each taking every counter, or the one empty payment when it holds
// none.
func counterChoicesAtX(choices []counterChoice, x, floor int) ([]counterChoice, int) {
	held, at := 0, -1
	for i, cc := range choices {
		if cc.total > held {
			held, at = cc.total, i
		}
	}
	n := min(x, held)
	if n < floor || n < 0 {
		return nil, 0
	}
	if n == 0 || at < 0 {
		return []counterChoice{{}}, 0
	}
	cc := choices[at]
	part := cc.parts[0]
	part.n = n
	cc.parts = []counterPart{part}
	cc.total = n
	return []counterChoice{cc}, n
}

// counterCountIsXFloor is the X floor the enumerator holds such a cost to.
func counterCountIsXFloor(source *game.Card, cost game.AbilityCost) int {
	return enumeratedXFloor(game.CatalogAbilityKey(*source), cost.FloorX())
}
