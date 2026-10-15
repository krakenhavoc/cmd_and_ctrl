package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// handExilePayments is the payments a cast offers a mandatory hand-exile
// additional cost (#3175, the Marches: "you may exile any number of blue
// cards from your hand. This spell costs {2} less to cast for each card
// exiled this way"). [nil] for every card without one, so the caller's
// loop runs once and the enumeration is unchanged.
//
// The count is part of the announcement and changes the price, so each
// count is its own move, priced with its own discount
// (CostQuery.ExilingFromHand): none, then one, two, … cards off the front
// of the engine's own candidate walk (game.HandExileOptionsForEffect) in
// the policy's cheapest-fuel order, up to the cost-payment cap. Which
// cards of a given count is not worth a cross product: the policy has
// already said which it would rather lose (ADR 0100 §6's discipline for a
// variable sacrifice).
func (e *enumerator) handExilePayments(card game.Card) [][]uuid.UUID {
	ac := game.AdditionalCostFor(game.CatalogKey(card))
	if ac == nil || ac.ExileFromHand == nil {
		return [][]uuid.UUID{nil}
	}
	pool := e.cheapestFuelFirst(e.g.HandExileOptionsForEffect(e.seat, card.InstanceID, ac.ExileFromHand))
	n := min(len(pool), e.capOr(maxEnumeratedCostPayments))
	if len(pool) > n {
		e.noteCut(CapCostPayments, len(pool)-n, false)
	}
	out := make([][]uuid.UUID, 0, n+1)
	out = append(out, nil)
	for k := 1; k <= n; k++ {
		out = append(out, append([]uuid.UUID(nil), pool[:k]...))
	}
	return out
}
