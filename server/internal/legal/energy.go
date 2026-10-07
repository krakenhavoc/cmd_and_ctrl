package legal

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// energy.go — ADR 0129 §7: the enumerator's half of paying energy.
//
// A fixed "Pay N {E}" is a gate (game.EnergyShortfall, the predicate the
// engine refuses with) and a price (MoveCost.Energy). "Pay X {E}" also
// bounds the announced X: CR 118.3 forbids announcing more than the
// seat can pay, so the largest X offered is the energy left after the
// printed part, capped by Options.MaxX like every other X search.

// energyXCeiling is the largest X a "Pay X {E}" cost lets the seat
// announce: its energy less the printed part, capped at maxX. -1 when
// the cost has no energy X, or the seat cannot pay even the printed
// part.
func energyXCeiling(cost game.AbilityCost, energy, maxX int) int {
	if !cost.EnergyX {
		return -1
	}
	ceiling := energy - cost.Energy
	if ceiling > maxX {
		ceiling = maxX
	}
	if ceiling < 0 {
		return -1
	}
	return ceiling
}

// capEnergyX bounds a solved activation's X by the energy ceiling. With
// {X} in the mana too (no printed card has both), X is one number and
// both components charge it (CR 107.3a), so the smaller bound wins; with
// none, the energy alone sets it, and the largest affordable value is
// offered, as for a mana X. False when the floor cannot be met.
func capEnergyX(pay abilityManaPayment, cost game.AbilityCost, ceiling, floor int) (abilityManaPayment, bool) {
	if ceiling < 0 {
		return pay, false
	}
	if cost.XSlots() == 0 || pay.xValue > ceiling {
		pay.xValue = ceiling
	}
	if pay.xValue < floor {
		return pay, false
	}
	return pay, true
}
