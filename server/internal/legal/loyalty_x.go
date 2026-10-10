package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loyalty_x.go — #1944 (ADR 0032's amendment of 2026-10-10): the
// enumerator's half of a loyalty cost of −X.
//
// A mana {X} is offered once, at the largest affordable value, because
// a bigger X only ever buys more. A −X loyalty cost is the other way
// round: every point of X is a loyalty counter the planeswalker loses,
// and the last one costs the planeswalker. So the X is a real choice,
// and the enumerator offers each value from the floor to the loyalty
// there (CR 606.6), one move per value with its own MoveCost.Loyalty.
// The policy prices them: a damage row declared DamageIsX kills its
// target at the smallest X that is lethal, and every X past that is
// loyalty spent for nothing.

// loyaltyXCeiling is the largest X a −X loyalty cost lets the seat
// announce for a permanent holding `loyalty` counters: the loyalty less
// the printed part, capped at maxX. -1 when the cost has no loyalty X.
func loyaltyXCeiling(cost game.AbilityCost, loyalty, maxX int) int {
	if !cost.LoyaltyX || cost.Loyalty == nil {
		return -1
	}
	ceiling := loyalty + *cost.Loyalty
	if ceiling > maxX {
		ceiling = maxX
	}
	if ceiling < 0 {
		return -1
	}
	return ceiling
}

// loyaltyXRungs is the X values one announcement is offered at: every
// value from floor to ceiling for a −X loyalty cost whose X nothing in
// the announcement fixed, and the single -1 ("no rung, the X the
// announcement already carries") for every other ability.
func loyaltyXRungs(cost game.AbilityCost, annX, floor, ceiling int) []int {
	if !cost.LoyaltyX || annX >= 0 || ceiling < floor {
		return []int{-1}
	}
	out := make([]int, 0, ceiling-floor+1)
	for x := floor; x <= ceiling; x++ {
		out = append(out, x)
	}
	return out
}

// tapAtRung is one tap-cost payment at one rung of a −X loyalty
// cost's X (rung -1 for every other ability).
type tapAtRung struct {
	taps []uuid.UUID
	rung int
}

// tapsAtRungs is every tap payment at every rung, rung-major, so the
// payment loop walks the product without another level of nesting.
func tapsAtRungs(tapSets [][]uuid.UUID, rungs []int) []tapAtRung {
	out := make([]tapAtRung, 0, len(tapSets)*len(rungs))
	for _, r := range rungs {
		for _, taps := range tapSets {
			out = append(out, tapAtRung{taps: taps, rung: r})
		}
	}
	return out
}
