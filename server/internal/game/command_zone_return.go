package game

import "github.com/google/uuid"

// command_zone_return.go — "return it to the command zone" for a
// commander on the battlefield (Hellkite Courser, #5368).
//
// CR 903.9b is a replacement for a commander headed for a hand or a
// library; going to the command zone itself is not one of the zones it
// replaces, so the move needs no question. The card keeps its identity
// (the commander tally is keyed by it) and leaves the battlefield like
// any other permanent: its leaves-the-battlefield triggers fire, and
// nothing is paid or taxed. The destination is the card's owner's
// command zone (ADR 0207's junkyard for an Astrotorium-backed card).

// ReturnToCommandZoneForEffect puts a permanent into its owner's command
// zone through the routed exit every move shares. A card that is not on
// the battlefield, or that is already in a command zone, is left alone
// (CR 400.7: a commander that went somewhere else is not the object the
// effect meant).
//
// Caller must hold g.mu.
func (g *Game) ReturnToCommandZoneForEffect(cardID uuid.UUID) error {
	if cardID == uuid.Nil {
		return nil
	}
	if z := g.findCardZoneLocked(cardID); z == nil || z.Kind != ZoneBattlefield {
		return nil
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{CardID: cardID, Dst: ZoneCommand})
	return err
}
