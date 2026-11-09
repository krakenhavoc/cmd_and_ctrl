package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exileCastPayment is the permanent(s) a mandatory exile-a-permanent
// additional cost names (#3402, Soul Exchange): the activation's own
// candidate walk and payment search (ExilePermanentsOptionsForEffect,
// exilePermanentPayments), so a payment this builds is one
// validateExilePermanentsCostLocked accepts. ONE payment, the policy's
// cheapest-fuel-first pick: the cast has nothing else to choose between
// two creatures, as teamworkBlightPayment's discipline has it. ok is
// false when the seat has nothing to exile, which drops the cast from
// the move list (the view ships the same picker present and short).
func (e *enumerator) exileCastPayment(castID uuid.UUID, mandatory *game.AdditionalCost) (ids []uuid.UUID, ok bool) {
	if mandatory == nil || mandatory.ExilePermanents.Empty() {
		return nil, true
	}
	ec := mandatory.ExilePermanents
	sets := e.exilePermanentPayments(e.g.ExilePermanentsOptionsForEffect(e.seat, castID, ec), ec, castID)
	if len(sets) == 0 {
		return nil, false
	}
	return sets[0], true
}

// returnCastPayment is the permanent(s) a return-a-permanent additional
// cost names (#5247, Fear of Isolation): exileCastPayment one destination
// over, through the activation's own walk and payment search
// (ReturnToHandOptionsForEffect, returnPayments), so a payment this builds
// is one validateReturnToHandCostLocked accepts. ONE payment, the policy's
// cheapest-fuel-first pick. ok is false when the seat has nothing to
// return, which drops the cast from the move list (CR 118.3).
//
// #5358: the clause is the mandatory slot's or an announced optional
// cost's (Chocobo Kick's land-returning kicker); effects.Register keeps
// at most one. With nothing to return only the kicked announcement is
// dropped — the unkicked set never names this clause.
func (e *enumerator) returnCastPayment(castID uuid.UUID, mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) (ids []uuid.UUID, ok bool) {
	rc := returnCostOf(mandatory, optional, chosen)
	if rc.Empty() {
		return nil, true
	}
	sets := e.returnPayments(e.g.ReturnToHandOptionsForEffect(e.seat, castID, rc), rc, castID)
	if len(sets) == 0 {
		return nil, false
	}
	return sets[0], true
}

// returnCostOf is the return-to-hand clause a cast's plan pays: the
// mandatory cost's, or an announced optional cost's (#5358). nil for none.
func returnCostOf(mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) *game.ReturnToHandCost {
	if mandatory != nil && !mandatory.ReturnToHand.Empty() {
		return mandatory.ReturnToHand
	}
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && !optional[i].ReturnToHand.Empty() {
			return optional[i].ReturnToHand
		}
	}
	return nil
}
