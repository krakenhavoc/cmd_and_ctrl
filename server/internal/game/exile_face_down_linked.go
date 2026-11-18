package game

import "github.com/google/uuid"

// ExileFaceDownLinkedForEffect exiles `cardID` face down (CR 406.3a: nobody may
// look at it, its controller included) and linked to the permanent
// `sourceID` as the object it is now (CR 607.2a, 400.7), for "exile the top
// card of your library face down" in a replacement that a later ability of the
// same card reads back as "cards exiled with <this>" (Asmodeus the Archfiend's
// "If you would draw a card, exile the top card of your library face down
// instead" and "Return all cards exiled with Asmodeus to their owner's hand").
//
// It is the face-down twin of ExileLinkedForEffect and the blind twin of
// ExileHiddenLinkedForEffect: that one lands FaceDownHidden, whose viewer is the
// controller of the source, which is the wrong rule here. The card lands
// FaceDownExiled, so no one is a knower, and ExiledWithForEffect still finds it.
// A source that is no longer on the battlefield exiles the card with no link
// (it is no object any ability can be reading any more).
//
// Answers whether the move paused on a CR 903.9 commander prompt.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileFaceDownLinkedForEffect(sourceID, actor, cardID uuid.UUID) (bool, error) {
	route := zoneRoute{
		CardID:   cardID,
		Dst:      ZoneExile,
		Actor:    actor,
		Source:   sourceID,
		FaceDown: FaceDownExiled,
	}
	if src := findBattlefieldCard(g, sourceID); src != nil {
		route.ExiledWith = PermissionCardRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}
	}
	return g.routeCardToZoneLocked(route)
}
