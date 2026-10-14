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
