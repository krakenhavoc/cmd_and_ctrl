package game

import "github.com/google/uuid"

// planar_deck_effects.go — effects that look at and reorder the planar deck
// (ADR 0216 §6, #4870): Stairs to Infinity's "reveal the top card of your
// planar deck. You may put it on the bottom of your planar deck".
//
// The planar deck is outside the card index, so RevealForEffect cannot find
// its cards; these methods reveal them directly. A reveal is not a move
// (CR 701.20b), and a revealed card is hidden again once the effect is done
// with it: a face-down plane (CR 311.2) is not a public object.

// RevealTopOfPlanarDeckForEffect reveals the top card of the planar deck to
// every seated player and returns it, or false when the deck is empty or
// the table is not playing Planechase. The reveal is announced as an
// EventRevealCards carrying the card's name.
//
// Caller must hold g.mu (write).
func (g *Game) RevealTopOfPlanarDeckForEffect(player, source uuid.UUID, reason string) (Card, bool) {
	deck := g.planarDeckForLocked(player)
	if deck == nil {
		return Card{}, false
	}
	top, err := deck.Top()
	if err != nil {
		return Card{}, false
	}
	viewers := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p != nil {
			viewers = append(viewers, p.ID)
		}
	}
	deck.Cards[len(deck.Cards)-1].AddKnowersAll(viewers)
	g.EmitEvent(Event{
		Kind:        EventRevealCards,
		Actor:       player,
		Source:      source,
		CardID:      top.InstanceID,
		Label:       reason,
		OldZone:     ZoneCommand,
		PublicNames: map[uuid.UUID]string{top.InstanceID: top.Name},
	})
	return top, true
}

// PutTopOfPlanarDeckOnBottomForEffect moves the planar deck's top card to
// its bottom if it is still `cardID`, face down again. It reports whether
// it moved. A card that is not on top any more (the deck changed while a
// prompt was open) is left alone.
//
// Caller must hold g.mu (write).
func (g *Game) PutTopOfPlanarDeckOnBottomForEffect(player, cardID uuid.UUID) bool {
	deck := g.planarDeckForLocked(player)
	if deck == nil {
		return false
	}
	top, err := deck.Top()
	if err != nil || top.InstanceID != cardID {
		return false
	}
	c, _ := deck.PopTop()
	c.ClearKnown()
	deck.PushBottom(c)
	return true
}

// HideTopOfPlanarDeckForEffect turns a revealed top card face down again
// without moving it: the effect that revealed it is over.
//
// Caller must hold g.mu (write).
func (g *Game) HideTopOfPlanarDeckForEffect(player, cardID uuid.UUID) {
	deck := g.planarDeckForLocked(player)
	if deck == nil || len(deck.Cards) == 0 {
		return
	}
	if last := &deck.Cards[len(deck.Cards)-1]; last.InstanceID == cardID {
		last.ClearKnown()
	}
}
