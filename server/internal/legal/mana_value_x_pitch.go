package legal

import (
	"github.com/google/uuid"
)

// manaValueXPitchPayments is the pitch payments of an alternative cost
// whose card is bound to the announced X (#3175, CR 107.3a / 601.2b):
// Shining Shoal's "exile a white card with mana value X from your hand".
// The card IS the announcement, so a payment and its X are one choice.
//
// One payment per DISTINCT mana value among the candidates, the first
// card of each value in `pool`'s order (the policy's cheapest fuel
// first), with the X it announces beside it. Two cards of one value buy
// the same spell; the policy has already said which it would rather
// lose. A value under `floor` is left out: a card whose effect is X
// (XMatters) is not offered at X = 0, #810's rule. At most the cost-
// payment cap of values, the smallest X last cut.
//
// The value is read off the card in hand, where an {X} in its own cost
// counts as 0 (CR 202.3e) — the reading validateAlternativeCostPaymentLocked
// holds the named card to.
func (e *enumerator) manaValueXPitchPayments(pool []uuid.UUID, floor int) (sets [][]uuid.UUID, xs []int) {
	limit := e.capOr(maxEnumeratedCostPayments)
	seen := make(map[int]bool, len(pool))
	cut := 0
	for _, id := range pool {
		mv, ok := e.handManaValue(id)
		if !ok || mv < floor || seen[mv] {
			continue
		}
		seen[mv] = true
		if len(sets) >= limit {
			cut++
			continue
		}
		sets = append(sets, []uuid.UUID{id})
		xs = append(xs, mv)
	}
	if cut > 0 {
		e.noteCut(CapCostPayments, cut, false)
	}
	return sets, xs
}

// handManaValue is the mana value of a card in the seat's hand, as the
// engine reads it there, and false for a card not in the hand or with a
// cost the engine cannot read (which meets no mana-value bound).
func (e *enumerator) handManaValue(id uuid.UUID) (int, bool) {
	for i := range e.p.Hand.Cards {
		if c := e.p.Hand.Cards[i]; c.InstanceID == id {
			return e.g.ManaValueForEffect(c)
		}
	}
	return 0, false
}
