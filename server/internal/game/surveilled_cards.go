package game

import "github.com/google/uuid"

// surveilled_cards.go — which cards a player surveilled into their
// graveyard this turn (#5429, Eye of Duskmantle).
//
//	CR 701.25a  "To 'surveil N' means to look at the top N cards of your
//	            library, then put any number of them into your graveyard
//	            and the rest on top of your library in any order."
//
// The engine already knew that a PLAYER surveilled this turn (EventSurveil
// in the turn's log) and where a graveyard card came FROM
// (Card.PutInGraveyardFrom, stamped by MoveCard). "Cards in your graveyard
// you've surveilled this turn" is a third fact, about the card: it is in a
// graveyard because this turn's surveil put it there. A card milled from
// the same library a moment later has the same origin and is not one.
//
// So it is one bool on the card, Card.Surveilled, written by the surveil's
// own continuation (surveilRouted) over the cards its graveyard leg really
// put into the graveyard: a leg Rest in Peace exiled instead, or a
// commander its owner sent to the command zone (CR 903.9), is not among
// them. The "this turn" half is the arrival record's
// (Card.PutInGraveyardThisTurn), which the turn boundary already clears,
// and MoveCard forgets the mark on every move, so a card that leaves the
// graveyard and comes back is a new object (CR 400.7) that was not
// surveilled. A card kept on top of the library was looked at but never
// moved, and is not marked.
//
// "You've surveilled" needs no player on the mark: a player surveils only
// their own library (CR 701.25a), and the route puts a card into its
// owner's graveyard, so the surveilling player is the owner of every card
// the mark is on. It is written whether or not anything reads it; the Eye
// that arrives later in the turn sees the turn's earlier surveils.

// SurveilledThisTurn reports whether the card is in a graveyard because its
// owner surveilled it there during the current turn.
func (c Card) SurveilledThisTurn() bool {
	return c.Surveilled && c.PutInGraveyardThisTurn
}

// markSurveilledLocked stamps Card.Surveilled on each of `ids` that is in
// playerID's graveyard, the cards a surveil's graveyard leg put there.
//
// Caller must hold g.mu.
func (g *Game) markSurveilledLocked(playerID uuid.UUID, ids []uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Graveyard == nil || len(ids) == 0 {
		return
	}
	for i := range p.Graveyard.Cards {
		c := &p.Graveyard.Cards[i]
		for _, id := range ids {
			if c.InstanceID == id {
				c.Surveilled = true
				break
			}
		}
	}
}
