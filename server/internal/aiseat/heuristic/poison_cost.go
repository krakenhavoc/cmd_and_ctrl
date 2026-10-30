package heuristic

// poison_cost.go — #4613: the price of getting poison counters as a cost.
//
// The Serpent Society's "Ward—Get five poison counters" is a payment the
// caster can always make (a player can always get counters), so the
// enumerator offers it at any poison total and legal.MoveCost.Poison
// carries the count. What the bot weighs is what the counters do to it.
// CR 704.5c loses the game at ten, so a payment that reaches ten is a
// refusal, exactly as a life payment to zero is (costValue), and a smaller
// one is priced on the life scale: Weights.PoisonCostValue.

// poisonLifeEquivalent is how much life one poison counter is worth to a
// seat. Ten counters lose the game and the life curve is calibrated on a
// twenty-life game, so a counter is two life: the whole poison track is the
// same twenty-point scale LifeCostValue is calibrated on, and the counters
// never leave on their own.
const poisonLifeEquivalent = 2

// PoisonCostValue is what getting `pay` more poison counters costs a seat
// that already has `poison`, in the units Strength is measured in: the
// same integral of the life curve LifeCostValue takes, over a track that
// runs from the seat's counters to the tenth. It is a price, positive.
// A payment that reaches ten counters is not priced here; that is a loss
// of the game, and costValue refuses it outright.
func (w Weights) PoisonCostValue(poison, pay int) float64 {
	if pay <= 0 {
		return 0
	}
	room := poisonLethal - poison
	if room <= pay {
		room = pay + 1
	}
	return w.LifeCostValue(room*poisonLifeEquivalent, pay*poisonLifeEquivalent)
}

// poisonCostLethal reports whether getting n more poison counters would
// take the bot to ten, which loses the game (CR 704.5c).
func poisonCostLethal(st *state, n int) bool {
	if n <= 0 || st.seat == nil {
		return false
	}
	return st.seat.Poison+n >= poisonLethal
}

// poisonCostValue is the price of getting n poison counters, as a number
// to subtract. Zero when the view does not carry the seat's counters.
func poisonCostValue(st *state, n int) float64 {
	if n <= 0 || st.seat == nil {
		return 0
	}
	return st.w.PoisonCostValue(st.seat.Poison, n)
}
