package game

import "github.com/google/uuid"

// PutFromOutsideOnLibraryTopForEffect puts one card `playerID` owns outside
// the game on top of their library: The Raven's Warning's "you may put a
// card you own from outside the game on top of your library" (#5505). It
// is PutFromOutsideIntoLibraryForEffect's other end: like it, only an
// effect that brings a card in reaches it (CR 400.11c), the card moves as a
// zone change from outside the game, and nobody shuffles. The owner chose
// the card, so they know what is on top of the library.
//
// A card that is not in the player's container moves nothing (a second
// submission of the same ID is harmless) and reports false.
//
// Caller must hold g.mu.
func (g *Game) PutFromOutsideOnLibraryTopForEffect(playerID, cardID uuid.UUID) (bool, error) {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Sideboard == nil || p.Library == nil {
		return false, ErrCardNotFound
	}
	c, err := p.Sideboard.Remove(cardID)
	if err != nil {
		return false, nil
	}
	c.Owner = p.ID
	c.Controller = p.ID
	c.AddKnower(p.ID)
	p.Library.PushTop(c)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   p.ID,
		CardID:  c.InstanceID,
		OldZone: ZoneOutside,
		NewZone: ZoneLibrary,
	})
	return true, nil
}
