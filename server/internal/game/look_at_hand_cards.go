package game

import "github.com/google/uuid"

// look_at_hand_cards.go — "look at those cards", of cards in another
// player's hand (#5457).
//
// "Target opponent chooses X cards from their hand. Look at those cards."
// (Extract Brain) is a look (CR 701.20), not a reveal: only the looker
// learns the cards, and no event names them to the table.

// LookAtHandCardsForEffect marks `looker` as having looked at each of
// `ids` that is in `owner`'s hand, and returns the ones it found there,
// in the order given. A card that has left the hand is skipped.
//
// The knowledge stays with the card for as long as it stays in that
// hand, as a Thoughtseize's does.
//
// Caller must hold g.mu (write).
func (g *Game) LookAtHandCardsForEffect(looker, owner uuid.UUID, ids []uuid.UUID) []uuid.UUID {
	p := g.playerByIDLocked(owner)
	if p == nil || p.Hand == nil || looker == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for _, id := range ids {
		for i := range p.Hand.Cards {
			if p.Hand.Cards[i].InstanceID == id {
				p.Hand.Cards[i].AddKnower(looker)
				out = append(out, id)
				break
			}
		}
	}
	return out
}
