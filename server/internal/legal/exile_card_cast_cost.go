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
func (e *enumerator) exileCardsCastPayment(castID uuid.UUID, mandatory *game.AdditionalCost) (ids []uuid.UUID, ok bool) {
	if mandatory == nil || mandatory.ExileCards == nil {
		return nil, true
	}
	return e.exileCardsPayment(mandatory.ExileCards, castID, nil)
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
