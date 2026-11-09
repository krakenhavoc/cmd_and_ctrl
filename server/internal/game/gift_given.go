package game

import "github.com/google/uuid"

// gift_given.go — "whenever you give a gift" (Jolly Gerbils, CR 702.174).
//
// A gift is given when its effect happens for the chosen opponent
// (CR 702.174d–i): a spell's, as its resolution begins, a permanent's,
// from its entry trigger. The promise alone is not the gift.

// EventGiftGiven records that Actor gave a gift (CR 702.174). Source is
// the card whose gift it was and Target the recipient. It is its own kind,
// as EventConnived is, so a "whenever you give a gift" payoff fires on nothing else.
const EventGiftGiven EventKind = "gift_given"

// GiftGivenForEffect announces that giver gave a gift to recipient on
// behalf of source. Caller must hold g.mu.
func (g *Game) GiftGivenForEffect(giver, recipient, source uuid.UUID) {
	g.EmitEvent(Event{Kind: EventGiftGiven, Actor: giver, Source: source, Target: recipient})
}
