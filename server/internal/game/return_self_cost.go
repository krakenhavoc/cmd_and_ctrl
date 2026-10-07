package game

import "github.com/google/uuid"

// return_self_cost.go — #2028: returning the ability's own permanent to
// its owner's hand as part of the cost (AbilityCost.ReturnSelf).
//
// Gossamer Chains' "Return this enchantment to its owner's hand:",
// Shigeki, Jukai Visionary's "{1}{G}, {T}, Return Shigeki to its
// owner's hand:". CR 602.2b runs CR 601.2h for an ability, so the
// return is part of announcing it: the permanent is in its owner's
// hand before anyone can respond, and an ability countered on the stack
// has still cost it.
//
// The validator lives here and the payment reuses ReturnToHandCost's
// payer (return_cost.go), because a return-self is that component with
// the pick fixed to the source: one battlefield exit to the OWNER's
// hand, with MustSettleNow, and the owner's CR 903.9b answer asked
// before the payment began (askCostCommanderLocked, ADR 0115).

// validateReturnSelfCostLocked checks a ReturnSelf component without
// moving anything — ADR 0020 §3's "validate everything, then pay
// everything".
//
//   - The ability must have been activated from the battlefield: only a
//     permanent can be returned. effects.Register already refuses the
//     component on an ability that functions from another zone
//     (AbilityNeedsPermanentSource), so this is the runtime half of the
//     same rule.
//   - The source pays one component (CR 118.3). A cost that also
//     sacrifices or exiles the source is refused, and so is naming the
//     source as one of the permanents another component sacrifices,
//     returns or exiles.
//
// Caller must hold g.mu.
func validateReturnSelfCostLocked(sourceID uuid.UUID, srcZone ZoneKind, cost AbilityCost, moved ...[]uuid.UUID) error {
	if !cost.ReturnSelf {
		return nil
	}
	if srcZone != ZoneBattlefield {
		return ErrActivationZoneNotAllowed
	}
	if cost.SacrificeSelf || cost.ExileSelf {
		return ErrInvalidParam
	}
	for _, ids := range moved {
		for _, id := range ids {
			if id == sourceID {
				return ErrInvalidParam
			}
		}
	}
	return nil
}

// payReturnSelfCostLocked pays a ReturnSelf component: the source goes
// to its owner's hand through payReturnToHandCostLocked, the payer the
// ReturnToHand component uses, so the move is the one battlefield exit
// with cause cost. The exit records the permanent's last-known
// information (CR 608.2h), which is what the ability's effect reads for
// "this permanent" once the card in the hand is a new object
// (CR 400.7).
//
// Call only after validateReturnSelfCostLocked has passed, and before
// the ability's stack item is built, so the leaves-the-battlefield
// triggers it queues sit above the ability, as the other
// components that move permanents do.
//
// Caller must hold g.mu.
func (g *Game) payReturnSelfCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, answers map[uuid.UUID]bool) error {
	if !cost.ReturnSelf {
		return nil
	}
	_, err := g.payReturnToHandCostLocked(playerID, sourceID, []uuid.UUID{sourceID}, answers)
	return err
}
