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

// BottomSelfZoneSupported reports whether a BottomSelf component can be
// paid from `z`: the battlefield (Timestream Navigator's "Put this
// creature on the bottom of its owner's library") or, since the Star
// Trek batch 24 cards, the graveyard (Daniels, Temporal Agent's "Put
// this card from your graveyard on the bottom of your library"). It is
// ExileSelfZoneSupported's sibling: effects.Register refuses the
// component on an ability that functions from anywhere else, at boot,
// and the validator below refuses an activation from anywhere else at
// runtime, from the one predicate.
func BottomSelfZoneSupported(z ZoneKind) bool {
	return z == ZoneBattlefield || z == ZoneGraveyard
}

// validateBottomSelfCostLocked checks a BottomSelf component without
// moving anything (ADR 0020 §3): the ability must be activated from a
// zone the component can be paid from (BottomSelfZoneSupported), the
// source pays one component (CR 118.3), and no other component may name
// the source as one of its picks.
//
// Caller must hold g.mu.
func validateBottomSelfCostLocked(sourceID uuid.UUID, srcZone ZoneKind, cost AbilityCost, moved ...[]uuid.UUID) error {
	if !cost.BottomSelf {
		return nil
	}
	if !BottomSelfZoneSupported(srcZone) {
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
// From the graveyard it is the same move out of the graveyard (the
// activation path has already checked the card is the activator's own,
// CR 108.4); the library card is a new object (CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) payBottomSelfCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, answers map[uuid.UUID]bool) error {
	if !cost.BottomSelf {
		return nil
	}
	if findBattlefieldCard(g, sourceID) == nil && !g.cardInAGraveyardLocked(sourceID) {
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

// cardInAGraveyardLocked reports whether any player's graveyard holds
// the card.
//
// Caller must hold g.mu.
func (g *Game) cardInAGraveyardLocked(id uuid.UUID) bool {
	for _, p := range g.Seats {
		if p.Graveyard.Contains(id) {
			return true
		}
	}
	return false
}
