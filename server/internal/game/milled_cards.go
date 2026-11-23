package game

import "github.com/google/uuid"

// milled_cards.go — which cards a mill put into a graveyard this turn
// (Raul, Trouble Shooter).
//
//	CR 701.17a  "To mill a number of cards, a player puts that many cards
//	            from the top of their library into their graveyard."
//
// The engine already knew where a graveyard card came FROM
// (Card.PutInGraveyardFrom, stamped by MoveCard) and emitted EventMill for a
// card a mill put there. "Cards in your graveyard that were milled this turn"
// is a fact about the card: it is in a graveyard because a mill put it there
// this turn. A card put there by a search (Entomb) or a surveil has the same
// origin and is not one.
//
// So it is one bool on the card, Card.Milled, written where the mill's
// EventMill is emitted (the one place a milled card lands in a graveyard)
// and forgotten by MoveCard on every move, so a card that leaves the
// graveyard and comes back is a new object (CR 400.7) that was not milled.
// The "this turn" half is the arrival record's (Card.PutInGraveyardThisTurn),
// which the turn boundary already clears. It is written whether or not
// anything reads it: a Raul that arrives later in the turn sees the turn's
// earlier mills.

// MilledThisTurn reports whether the card is in a graveyard because a mill
// put it there during the current turn.
func (c Card) MilledThisTurn() bool {
	return c.Milled && c.PutInGraveyardThisTurn
}

// markMilledLocked stamps Card.Milled on the card `id` in `z`, the graveyard
// a mill's route has just landed it in. A card that is not there (a commander
// its owner sent to the command zone, CR 903.9) is left alone.
//
// Caller must hold g.mu.
func (g *Game) markMilledLocked(z *Zone, id uuid.UUID) {
	if z == nil {
		return
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			z.Cards[i].Milled = true
			return
		}
	}
}
