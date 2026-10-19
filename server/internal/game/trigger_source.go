package game

import "github.com/google/uuid"

// TriggerSourceClaimants lists the players entitled to announce a
// trigger by hand for the card cardID (#4131); the bool is false when no
// zone holds the card, so the caller can leave "I don't know that card"
// to the mutation's own ErrCardNotFound.
//
// A permanent on the battlefield has one claimant, its controller. Once
// the card has left, a leaves-the-battlefield or dies trigger belongs to
// whoever controlled it as it last existed there (CR 603.3a, CR 603.10a)
// and not to whoever controls the card now, which is its owner's
// graveyard or hand: so the claimants are the current controller, the
// owner, and the last-known controller read off the card's most recent
// EventLTB. That last one is what lets a player whose stolen creature
// died announce its dies trigger.
func (g *Game) TriggerSourceClaimants(cardID uuid.UUID) ([]uuid.UUID, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	z, pos := g.locateCardLocked(cardID)
	if z == nil {
		return nil, false
	}
	c := z.Cards[pos]
	claimants := []uuid.UUID{c.Controller}
	if findBattlefieldCard(g, cardID) != nil {
		return claimants, true
	}
	claimants = append(claimants, c.Owner)
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == EventLTB && ev.CardID == cardID && ev.LastKnownController != uuid.Nil {
			claimants = append(claimants, ev.LastKnownController)
			break
		}
	}
	return claimants, true
}
