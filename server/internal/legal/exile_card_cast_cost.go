package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exileCardsCastPayment is the graveyard cards a mandatory
// exile-N-cards-from-your-graveyard additional cost names (#3536,
// Abhorrent Oculus): the activation's own candidate walk and payment
// (exileCardsPayment over ExileCostOptionsForEffect), so a payment this
// builds is one validateExileCardsCostLocked accepts. ONE payment, the
// policy's cheapest-fuel-first pick: the cast has nothing else to choose
// between two sets, as teamworkBlightPayment's discipline has it. ok is
// false when the graveyard cannot cover the count, which drops the cast
// from the move list (the view ships the same picker present and short,
// and the cast gate greys it).
//
// #1948: for the X form ("exile X cards") the count is the X the move
// announces, which is not known yet, so this returns EVERY candidate,
// cheapest fuel first; exileCardsForX cuts it to the announced X when the
// move is built, and exileXCeiling caps the X search at its length.
func (e *enumerator) exileCardsCastPayment(castID uuid.UUID, mandatory *game.AdditionalCost) (ids []uuid.UUID, ok bool) {
	if mandatory == nil || mandatory.ExileCards == nil {
		return nil, true
	}
	if mandatory.ExileCards.CountFromX {
		return e.cheapestFuelFirst(e.g.ExileCostOptionsForEffect(e.seat, castID, mandatory.ExileCards)), true
	}
	return e.exileCardsPayment(mandatory.ExileCards, castID, nil)
}

// exileXCeiling is the largest X an "exile X cards from your graveyard"
// cost lets the seat announce (#1948): the engine's own
// ExileCardsXCeilingForEffect, the number CastSpell refuses above, so the
// enumerator never offers an X the engine refuses (#544). noXCeiling for
// every other cost.
func (e *enumerator) exileXCeiling(castID uuid.UUID, addCost *game.AdditionalCost) int {
	if addCost == nil || addCost.ExileCards == nil || !addCost.ExileCards.CountFromX {
		return noXCeiling
	}
	return e.g.ExileCardsXCeilingForEffect(e.seat, castID, addCost.ExileCards)
}

// exileCardsForX is the payment a move at X names: the first X of the
// cheapest-first pool for the X form, the fixed payment otherwise.
func exileCardsForX(paying *game.AdditionalCost, pool []uuid.UUID, x int) []uuid.UUID {
	if paying == nil || paying.ExileCards == nil || !paying.ExileCards.CountFromX {
		return pool
	}
	if x <= 0 {
		return nil
	}
	if x > len(pool) {
		x = len(pool)
	}
	return pool[:x]
}

// withoutIDs is pool with every id in `taken` removed, order kept. The
// delve pool is cut by it so a card the exile cost names is never also
// offered to delve (CR 118.3).
func withoutIDs(pool, taken []uuid.UUID) []uuid.UUID {
	if len(taken) == 0 || len(pool) == 0 {
		return pool
	}
	skip := make(map[uuid.UUID]bool, len(taken))
	for _, id := range taken {
		skip[id] = true
	}
	out := make([]uuid.UUID, 0, len(pool))
	for _, id := range pool {
		if !skip[id] {
			out = append(out, id)
		}
	}
	return out
}
