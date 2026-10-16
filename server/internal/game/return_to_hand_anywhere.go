package game

import "github.com/google/uuid"

// return_to_hand_anywhere.go — "return it to your hand" for a card that
// may be sitting in any of the zones a discard can leave it in (#3859).
//
// Quagnoth's "When a spell or ability an opponent controls causes you to
// discard this card, return it to your hand" names no zone. The card was
// discarded wherever it went (CR 701.9a defines a discard by where the
// card came FROM): the graveyard in the ordinary case, exile when
// madness (CR 702.35a) or Rest in Peace sent it there, the top of the
// library under Library of Leng, the command zone when CR 903.9 offered
// it to a commander. "It" is the card as it now exists in that zone (the
// move was the discard itself), so it comes back from whichever one it
// is in.

// ReturnFromAnyZoneToHandForEffect puts the card into its owner's hand
// from the graveyard, exile, the library or the command zone, whichever
// holds it, and reports whether it moved. A card anywhere else (still in
// the hand, on the battlefield, on the stack, gone from the game) is left
// alone: nothing to return (CR 608.2b).
//
// actor is the player doing it and source the card whose ability it is;
// both ride the zone-change event.
//
// Caller must hold g.mu.
func (g *Game) ReturnFromAnyZoneToHandForEffect(actor, source, cardID uuid.UUID) (bool, error) {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return false, nil
	}
	switch z.Kind {
	case ZoneGraveyard, ZoneExile, ZoneLibrary, ZoneCommand:
	default:
		return false, nil
	}
	c, ok := g.cardInZoneLocked(z, cardID)
	if !ok || c.IsToken() {
		return false, nil
	}
	if _, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:   cardID,
		Dst:      ZoneHand,
		DstOwner: c.Owner,
		Actor:    actor,
		Source:   source,
	}); err != nil {
		return false, err
	}
	now := g.findCardZoneLocked(cardID)
	return now != nil && now.Kind == ZoneHand, nil
}
