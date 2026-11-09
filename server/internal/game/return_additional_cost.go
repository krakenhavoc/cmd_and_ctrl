package game

import "github.com/google/uuid"

// return_additional_cost.go — #5247: "As an additional cost to cast this
// spell, return a permanent you control to its owner's hand" (Fear of
// Isolation).
//
// The clause is the activated-ability component, ReturnToHandCost, on
// the MANDATORY slot of a cast (AdditionalCost.ReturnToHand), so the
// candidate walk (ReturnToHandOptionsForEffect), the payability
// predicate (ReturnToHandPayable), the validator and the payer are the
// activation's own (return_cost.go), and the view, the enumerator and the
// engine cannot disagree about which permanent pays (#544). What a cast
// adds is only where the picks ride and when they are paid:
//
//   - The caster names them on CastSpellParams.ReturnIDs, the name an
//     activation already uses (`return_ids`), chosen as the spell is
//     announced (CR 601.2b).
//   - Validated with every other component before anything is paid
//     (ADR 0020 §3), and kept apart from the permanents another component
//     of the same cast sacrifices or exiles, so one permanent cannot pay
//     two of them (CR 118.3).
//   - A commander among them is asked CR 903.9b BEFORE anything is paid,
//     with the alternative cost's return (Daze) and a splice card's, and
//     the announcement parks on that question (parked_announcement.go),
//     so a restore taken while it is open keeps the pick.
//   - Returned with the spell already on the stack (CR 601.2a before
//     601.2h), beside the sacrifices and the exiles, so a
//     leaves-the-battlefield trigger lands above it, each permanent to
//     its OWNER's hand (CR 400.3): a permanent you control but don't own
//     goes to its owner. Nothing dies and nothing is sacrificed.
//
// #5358: the same component on an OPTIONAL slot is a kicker (Chocobo
// Kick's "Kicker—Return a land you control to its owner's hand"): it
// joins the plan only when announced (CR 601.2b), and from there it is
// validated, asked about and paid exactly as above. Unkicked, a
// return_ids is refused as on any spell without the cost.
//
// A tapped permanent pays (returning is not tapping), and so may one
// tapped for mana while the spell is being cast: CR 601.2g has mana
// abilities activated before CR 601.2h pays the costs, so the auto-tapper
// treats a permanent named here as it treats one named to a sacrifice
// (CastAutoTapExclusions). Paying a cost does not target (CR 601.2h), so
// a hexproof permanent pays too.

// planReturnToHand is the return-to-hand component the announced plan
// demands, or nil: the mandatory slot's, or (#5358) an announced optional
// cost's — Chocobo Kick's "Kicker—Return a land you control to its
// owner's hand". effects.Register allows at most one across the two
// slots (checkOneCastReturn), and a card-shaped optional cost is paid at
// most once, so there is at most one.
func planReturnToHand(plan []costPayment) *ReturnToHandCost {
	for _, pay := range plan {
		if !pay.cost.ReturnToHand.Empty() {
			return pay.cost.ReturnToHand
		}
	}
	return nil
}

// validateCastReturnToHandLocked checks the permanents named to pay the
// plan's return-to-hand cost without moving anything: none when the plan
// has no such component (a stray return_ids is refused rather than
// ignored), else the activation validator's rules (exactly Count, each
// once, on the battlefield, the caster's, matching the clause without
// the targeting gate). `alsoSpent` is every permanent another component
// of the same cast takes off the battlefield — one permanent pays one
// component (CR 118.3).
//
// Caller must hold g.mu.
func (g *Game) validateCastReturnToHandLocked(playerID, castID uuid.UUID, plan []costPayment, ids, alsoSpent []uuid.UUID) error {
	rc := planReturnToHand(plan)
	if rc == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if sharesAnID(ids, alsoSpent) {
		return ErrInvalidParam
	}
	g.RecomputeLayersIfStaleLocked()
	return g.validateReturnToHandCostLocked(playerID, castID, rc, ids)
}

// castReturnAlsoSpent is every permanent a cast's other components take
// off the battlefield: the additional cost's sacrifices and exiles, and
// the alternative cost's card payment (a sacrifice or a return of its
// own, Daze's Island).
func castReturnAlsoSpent(params CastSpellParams) []uuid.UUID {
	return unionIDList(params.SacrificeIDs, params.ExilePermanentIDs, params.AltCostIDs)
}
