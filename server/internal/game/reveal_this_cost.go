package game

import "github.com/google/uuid"

// reveal_this_cost.go — "Reveal this card from your hand" as the cost of an
// ACTIVATED ability that works from the hand (Forecast, CR 702.57a; Sky
// Hussar). The twin of reveal_hand_cost.go for the one card that carries the
// ability: AbilityCost.RevealThis.
//
// It is paid at activation (CR 601.2h): the card is visible to every player as
// the cost is paid, before anyone may respond, and an ability that is countered
// has still shown it. Nothing moves (CR 701.20b): the card stays in the hand,
// known to every seat for as long as it sits there (RevealForEffect marks them
// knowers).

// payRevealThisCostLocked reveals the source card to the table. A source that
// is not in a hand (it has left, or the ability was activated from elsewhere)
// shows nothing: a card in a public zone is already known.
//
// Caller must hold g.mu.
func (g *Game) payRevealThisCostLocked(playerID, sourceID uuid.UUID) {
	if sourceID == uuid.Nil {
		return
	}
	z := g.FindCardZoneForEffect(sourceID)
	if z == nil || z.Kind != ZoneHand {
		return
	}
	reason := "reveal this card from your hand"
	if c, ok := g.LookupCardForEffect(sourceID); ok {
		reason = c.Name + " — " + reason
	}
	g.RevealForEffect(RevealSpec{Player: playerID, Source: sourceID, Reason: reason, Cards: []uuid.UUID{sourceID}})
}
