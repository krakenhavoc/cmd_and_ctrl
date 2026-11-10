package game

import "github.com/google/uuid"

// life_x_cost.go — "Pay X life" on an activated ability (Krumar
// Initiate's "{X}{B}, {T}, Pay X life: This creature endures X"), the
// case life_cost_count.go left out and AbilityCost.LifeX now covers.
//
// The amount is the announced X (CR 107.3a), read once at announce with
// the rest of the cost (CR 601.2f–g via CR 602.2b), so the life total is
// a CEILING on X rather than a number the card prints: CR 119.4 lets a
// player pay life only if their life total is at least the payment, and
// CR 119.8 refuses it outright while their life can't change. Energy's
// "Pay X {E}" (energy_cost.go) is the sibling, bounded the same way.

// AbilityLifeCostAtXLocked is AbilityLifeCostLocked with X announced as
// `x`: the life an activation charges, plus X for "Pay X life". ok is
// false for a count whose key is not registered.
//
// Caller must hold g.mu (read or write).
func (g *Game) AbilityLifeCostAtXLocked(activator, source uuid.UUID, cost AbilityCost, x int) (life int, ok bool) {
	life, ok = g.AbilityLifeCostLocked(activator, source, cost)
	if ok && cost.LifeX && x > 0 {
		life += x
	}
	return life, ok
}

// LifeXCeilingLocked is the largest X a "Pay X life" cost lets `p`
// announce: the most life they can pay beyond the printed part `printed`,
// capped at maxX. -1 when the cost has no life X, or the seat cannot pay
// even the printed part.
//
// Caller must hold g.mu (read or write).
func (g *Game) LifeXCeilingLocked(p *Player, cost AbilityCost, printed, maxX int) int {
	if !cost.LifeX || p == nil {
		return -1
	}
	x := p.Life - printed
	if x > maxX {
		x = maxX
	}
	for ; x >= 0; x-- {
		if g.CanPayLifeLocked(p, printed+x) {
			return x
		}
	}
	return -1
}
