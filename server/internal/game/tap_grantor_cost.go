package game

import (
	"slices"

	"github.com/google/uuid"
)

// tap_grantor_cost.go — "Tap <the permanent that granted this ability>" as
// a cost component (AbilityCost.TapGrantor).
//
//	CR 113.10  an effect gives an object an ability
//	CR 118.3   a cost can't be paid by something that can't pay it: an
//	           already tapped permanent can't be tapped
//	CR 602.2b  costs are paid as the ability is activated
//
// Fishing Pole is the card: "Equipped creature has '{1}, {T}, Tap Fishing
// Pole: Put a bait counter on Fishing Pole.'" The ability is the
// creature's, so the {T} taps the creature and a plain "tap this" would
// tap it twice; this names the row's grantor instead, the same
// AbilityOrigin.GrantedBy The Dominion Bracelet's exile-the-grantor cost
// reads (exile_grantor_cost.go). The grantor is the specific object that
// granted the row, not another permanent with the same name.
//
// It is not the {T} symbol of the Equipment, so a summoning-sick grantor
// (an Equipment is never a creature, but an Aura-like grantor could be)
// is not refused for it (CR 302.6 hangs sickness on the source's own {T}).

// tapGrantorOfLocked is the permanent a TapGrantor cost taps: the row's
// grantor, which must still be on the battlefield and untapped. Nil and no
// error when the cost has no such component. Caller must hold g.mu.
func (g *Game) tapGrantorOfLocked(cost AbilityCost, origin AbilityOrigin) (uuid.UUID, error) {
	if !cost.TapGrantor {
		return uuid.Nil, nil
	}
	if !origin.Granted() || origin.GrantedBy == uuid.Nil || g.Battlefield == nil {
		return uuid.Nil, ErrInvalidParam
	}
	c := findBattlefieldCard(g, origin.GrantedBy)
	if c == nil || c.Tapped {
		return uuid.Nil, ErrInvalidParam
	}
	return origin.GrantedBy, nil
}

// validateTapGrantorCostLocked checks the grantor against the other
// components of the same payment: one permanent pays one component
// (CR 118.3), so the grantor may not also be sacrificed, returned,
// exiled or tapped for another component, or be the source whose own {T}
// or sacrifice the same cost pays. Caller must hold g.mu.
func validateTapGrantorCostLocked(grantor, sourceID uuid.UUID, cost AbilityCost, moved ...[]uuid.UUID) error {
	if grantor == uuid.Nil {
		return nil
	}
	if grantor == sourceID && (cost.Tap || cost.SacrificeSelf || cost.ExileSelf || cost.ReturnSelf || cost.BottomSelf) {
		return ErrInvalidParam
	}
	for _, ids := range moved {
		if slices.Contains(ids, grantor) {
			return ErrInvalidParam
		}
	}
	return nil
}

// payTapGrantorCostLocked taps the grantor as part of a cost and announces
// it, as every other tap cost does. A grantor that has gone since the
// validation is skipped. Caller must hold g.mu.
func (g *Game) payTapGrantorCostLocked(playerID, grantor uuid.UUID) {
	if grantor == uuid.Nil {
		return
	}
	c := findBattlefieldCard(g, grantor)
	if c == nil || c.Tapped {
		return
	}
	c.Tapped = true
	g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: grantor})
}
