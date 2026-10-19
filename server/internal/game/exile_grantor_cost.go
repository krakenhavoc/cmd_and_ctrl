package game

import (
	"slices"

	"github.com/google/uuid"
)

// exile_grantor_cost.go — AbilityCost.ExileGrantor (#4000, ADR 0093's
// 2026-10-19 amendment): a granted ability whose cost exiles the
// permanent that granted it. The Dominion Bracelet:
//
//	Equipped creature gets +1/+1 and has "{15}, Exile The Dominion
//	Bracelet: You control target opponent during their next turn. This
//	ability costs {X} less to activate, where X is this creature's
//	power. Activate only as a sorcery."
//
// The ability is the equipped creature's (CR 113.10), so its source is
// the creature, and "Exile The Dominion Bracelet" names another object:
// the row's grantor, which the layer pass already records on the row's
// origin (AbilityOrigin.GrantedBy). The cost is read from that origin
// when the activation names the row, validated with every other
// component and paid with the permanent-moving ones, through the one
// exit route.

// exileGrantorOfLocked is the permanent a row's ExileGrantor component
// exiles: the row's grantor, on the battlefield. ErrInvalidParam when
// the row has no grantor (it is not a granted row, or the grant is a
// copy's) or the grantor is no longer on the battlefield; uuid.Nil when
// the cost has no such component. Caller must hold g.mu.
func (g *Game) exileGrantorOfLocked(cost AbilityCost, origin AbilityOrigin) (uuid.UUID, error) {
	if !cost.ExileGrantor {
		return uuid.Nil, nil
	}
	if !origin.Granted() || origin.GrantedBy == uuid.Nil || g.Battlefield == nil || !g.Battlefield.Contains(origin.GrantedBy) {
		return uuid.Nil, ErrInvalidParam
	}
	return origin.GrantedBy, nil
}

// validateExileGrantorCostLocked checks the grantor against the other
// components of the same payment: one permanent pays one component
// (CR 118.3), so the grantor may not be the source when the cost also
// moves the source, and may not be among the permanents the payment
// sacrifices, returns or exiles. Caller must hold g.mu.
func validateExileGrantorCostLocked(grantor, sourceID uuid.UUID, cost AbilityCost, moved ...[]uuid.UUID) error {
	if grantor == uuid.Nil {
		return nil
	}
	if grantor == sourceID && (cost.SacrificeSelf || cost.ExileSelf || cost.ReturnSelf || cost.BottomSelf) {
		return ErrInvalidParam
	}
	for _, ids := range moved {
		if slices.Contains(ids, grantor) {
			return ErrInvalidParam
		}
	}
	return nil
}

// payExileGrantorCostLocked exiles the grantor as part of a cost: the
// one exit route, marked as a cost and settled at once (CR 601.2h,
// 602.2b), with the owner's CR 903.9 answer for a commander asked
// before the payment began. It leaves the battlefield, so its
// leaves-the-battlefield triggers see it, and the grant ends with it.
// Caller must hold g.mu.
func (g *Game) payExileGrantorCostLocked(playerID, sourceID, grantor uuid.UUID, answers map[uuid.UUID]bool) error {
	if grantor == uuid.Nil {
		return nil
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:          grantor,
		Dst:             ZoneExile,
		Actor:           playerID,
		Source:          sourceID,
		Cause:           MoveCause{Kind: MoveCauseCost, Controller: playerID},
		MustSettleNow:   true,
		commanderAnswer: commanderAnswerFor(answers, grantor),
	})
	return err
}
