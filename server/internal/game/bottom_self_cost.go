package game

import "github.com/google/uuid"

// bottom_self_cost.go — #2726: putting the ability's own permanent on
// the bottom of its owner's library as part of the cost
// (AbilityCost.BottomSelf).
//
// Timestream Navigator's "{2}{U}{U}, {T}, Put this creature on the
// bottom of its owner's library:". ReturnSelf's sibling one zone over:
// CR 602.2b runs CR 601.2h for an ability, so the move is part of
// announcing it, and an ability countered on the stack has still cost
// the permanent. See return_self_cost.go for the shared reasoning.

// validateBottomSelfCostLocked checks a BottomSelf component without
// moving anything (ADR 0020 §3): the ability must be activated from the
// battlefield, the source pays one component (CR 118.3), and no other
// component may name the source as one of its picks.
//
// Caller must hold g.mu.
func validateBottomSelfCostLocked(sourceID uuid.UUID, srcZone ZoneKind, cost AbilityCost, moved ...[]uuid.UUID) error {
	if !cost.BottomSelf {
		return nil
	}
	if srcZone != ZoneBattlefield {
		return ErrActivationZoneNotAllowed
	}
	if cost.SacrificeSelf || cost.ExileSelf || cost.ReturnSelf {
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

// payBottomSelfCostLocked pays a BottomSelf component: the source goes
// to the bottom of its OWNER's library through the one exit path, with
// cause cost and MustSettleNow, so the leaves-the-battlefield triggers
// it queues see it leave and sit above the ability (called before the
// stack item is built). The exit records the permanent's last-known
// information (CR 608.2h), which the effect reads for "this permanent".
// The owner's CR 903.9b answer was asked before the payment began.
//
// Caller must hold g.mu.
func (g *Game) payBottomSelfCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, answers map[uuid.UUID]bool) error {
	if !cost.BottomSelf {
		return nil
	}
	if findBattlefieldCard(g, sourceID) == nil {
		return nil
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:          sourceID,
		Dst:             ZoneLibrary,
		ToBottom:        true,
		Actor:           playerID,
		Source:          sourceID,
		Cause:           MoveCause{Kind: MoveCauseCost, Controller: playerID},
		MustSettleNow:   true,
		commanderAnswer: commanderAnswerFor(answers, sourceID),
	})
	return err
}
