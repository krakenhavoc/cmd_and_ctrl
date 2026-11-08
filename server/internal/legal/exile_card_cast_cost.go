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
//
// #5193: an announced optional "you may collect evidence N" names its
// cards on the same list, through the same walk and floor
// (exileCardsPayment's any-number arm). A set of optional costs the
// graveyard cannot pay drops the move (CR 701.59b); declining it is
// always a move of its own.
//
// #5240: for collect evidence X sized by the targets the floor is not known
// until a move's targets are, so this returns every candidate, cheapest fuel
// first, like the X form; exileCardsForTargets cuts it per target set.
func (e *enumerator) exileCardsCastPayment(castID uuid.UUID, mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) (ids []uuid.UUID, ok bool) {
	ec := exileCardsCostOf(mandatory, optional, chosen)
	if ec == nil {
		return nil, true
	}
	if ec.CountFromX || ec.IsCollectEvidenceFromTargets() {
		return e.cheapestFuelFirst(e.g.ExileCostOptionsForEffect(e.seat, castID, ec)), true
	}
	return e.exileCardsPayment(ec, castID, nil)
}

// exileCardsForTargets is the payment a move with `targets` names for a
// collect evidence X sized by the targets (#5240, Urgent Necropsy): the
// cheapest-first pool cut where its total mana value first reaches the X
// those targets fix (Game.CollectEvidenceTargetsFloorForEffect, the number
// CastSpell checks), none at X = 0. ok is false when the whole pool falls
// short, and the target set is not offered (CR 118.3). Any other cost's
// payment is `pool` unchanged.
func (e *enumerator) exileCardsForTargets(paying *game.AdditionalCost, pool []uuid.UUID, targets []game.TargetRef) ([]uuid.UUID, bool) {
	if paying == nil || !paying.ExileCards.IsCollectEvidenceFromTargets() {
		return pool, true
	}
	floor := e.g.CollectEvidenceTargetsFloorForEffect(targets)
	return e.g.ExileCostCheapestSufficientForEffect(e.seat, game.CollectEvidenceCost(floor, ""), pool)
}

// exileCardsCostOf is the graveyard exile clause a cast's plan pays: the
// mandatory cost's, or an announced optional cost's (collect evidence,
// #5193). effects.Register keeps the two apart, so there is at most one.
func exileCardsCostOf(mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) *game.ExileCost {
	if mandatory != nil && mandatory.ExileCards != nil {
		return mandatory.ExileCards
	}
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && optional[i].ExileCards != nil {
			return optional[i].ExileCards
		}
	}
	return nil
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
