package legal

import (
	"fmt"
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

// payAmountOffers is what a pay_amount prompt (ADR 0129 §3) is offered
// as: nothing (a payment may always be declined), the smallest answer,
// the card's own threshold when it names one, the card's other marks
// (ADR 0129's amendment of 2026-10-09: the chooser's life total, for
// Volcano Hellion) and the ceiling, each once, ascending. A number with
// no ceiling does not offer its overflow guard. Every one is an answer
// the engine validates against (PayAmountPrompt.AnswerInBounds).
func payAmountOffers(pa *game.PayAmountPrompt) []int {
	candidates := []int{0, pa.Min, pa.Goal}
	candidates = append(candidates, pa.Marks...)
	if !pa.NoMax {
		candidates = append(candidates, pa.Max)
	}
	var out []int
	for _, n := range candidates {
		if pa.AnswerInBounds(n) && !slices.Contains(out, n) {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

// payAmountMove is one offered answer's label and price: "pay 3 {E}",
// "pay 5 life", "choose 4". The price is what the answer pays; a number
// that is not paid costs nothing, and its self-damage is the policy's
// to read off the prompt.
func payAmountMove(reason string, pa *game.PayAmountPrompt, amount int) (string, *MoveCost) {
	switch pa.ResourceOrEnergy() {
	case game.PayResourceLife:
		if amount == 0 {
			return reason + ": pay no life", nil
		}
		return fmt.Sprintf("%s: pay %d life", reason, amount), &MoveCost{Life: amount}
	case game.PayResourceNone:
		return fmt.Sprintf("%s: choose %d", reason, amount), nil
	}
	if amount == 0 {
		return reason + ": pay nothing", nil
	}
	return fmt.Sprintf("%s: pay %d {E}", reason, amount), withEnergy(nil, amount)
}
