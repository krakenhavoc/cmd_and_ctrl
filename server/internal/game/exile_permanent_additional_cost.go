package game

import "github.com/google/uuid"

// exile_permanent_additional_cost.go — #3402: "As an additional cost to
// cast this spell, exile a creature you control" (Soul Exchange).
//
// The clause is the activated-ability component, ExilePermanentsCost, on
// the MANDATORY slot of a cast (AdditionalCost.ExilePermanents), so the
// candidate walk, the clause check and the payer are the activation's
// own (exile_permanent_cost.go) and the two surfaces cannot disagree
// about who may pay (#544). What a cast adds is only where the picks
// ride and when they are paid:
//
//   - The caster names them on CastSpellParams.ExilePermanentIDs, the
//     name an activation already uses (`exile_permanent_ids`).
//   - They are paid with the spell already on the stack (CR 601.2a before
//     601.2h), beside the sacrifices, so a leaves-the-battlefield trigger
//     lands above the spell. Nothing dies and nothing is sacrificed.
//   - They are recorded on the spell's PaidCost: Exiled (the instance
//     IDs, as an ability's exile cost records them) and ExiledObjects
//     (each as the battlefield object it was, CR 400.7), which is what
//     "if the exiled creature was a Thrull" reads as last-known
//     information (CR 608.2h) through Context.ExiledPermanents.

// planExilePermanents is the exile-a-permanent component the announced
// plan demands, or nil. effects.Register allows it only on the mandatory
// slot, so there is at most one.
func planExilePermanents(plan []costPayment) *ExilePermanentsCost {
	for _, pay := range plan {
		if !pay.cost.ExilePermanents.Empty() {
			return pay.cost.ExilePermanents
		}
	}
	return nil
}

// validateCastExilePermanentsLocked checks the permanents named to pay
// the plan's exile-a-permanent cost without moving anything: none when
// the plan has no such component, else the activation validator's rules
// (exactly Count, each once, on the battlefield, the caster's, of the
// clause's type). `alsoSpent` is every permanent another component of
// the same cast pays with — one permanent pays one component (CR 118.3).
//
// Caller must hold g.mu.
func (g *Game) validateCastExilePermanentsLocked(playerID, castID uuid.UUID, plan []costPayment, ids, alsoSpent []uuid.UUID) error {
	ec := planExilePermanents(plan)
	if ec == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	return g.validateExilePermanentsCostLocked(playerID, castID, ec, ids, alsoSpent)
}

// exiledRefsLocked names the permanents an exile payment is about to take
// as the battlefield objects they are, for PaidCost.ExiledObjects. Read
// BEFORE the payment moves them, for sacrificeRefsLocked's reason.
//
// Caller must hold g.mu.
func (g *Game) exiledRefsLocked(ids []uuid.UUID) []ObjectRef {
	return g.sacrificeRefsLocked(ids)
}

// paidWithExiledPermanents folds the permanents the additional cost
// exiles into the record that lands on the stack item.
func paidWithExiledPermanents(paid PaidCost, ids []uuid.UUID, refs []ObjectRef) PaidCost {
	if len(ids) == 0 {
		return paid
	}
	paid.Exiled = append(paid.Exiled, ids...)
	paid.ExiledObjects = refs
	return paid
}
