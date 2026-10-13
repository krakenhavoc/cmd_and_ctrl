package game

import "github.com/google/uuid"

// reveal_hand_cost.go — "Reveal your hand" as the cost of an ACTIVATED
// ability (#3250; CR 602.2b, CR 701.20a). Sasaya, Orochi Ascendant:
// "Reveal your hand: If you have seven or more land cards in your hand,
// flip Sasaya."
//
// AbilityCost.RevealCards reveals a count of cards the activator names;
// "your hand" has no count to name, so it is a flag
// (AbilityCost.RevealHand), the way "Discard your hand" is a form of the
// discard component. It is paid at activation (CR 601.2h): the hand is
// visible to every player as the cost is paid, before anyone may
// respond, and an ability that is countered has still shown it. Moving
// the reveal into the resolution would change when opponents learn the
// hand, which is the whole point of the cost.
//
// Nothing moves (CR 701.20b). The cards stay known to every seat for as
// long as they sit in the hand (RevealForEffect marks them knowers); a
// card drawn afterwards is not known. An empty hand pays the cost in
// full (CR 118.3) and has nothing to show, so no reveal frame fires; the
// activation itself is still logged.

// payRevealHandCostLocked reveals every card in playerID's hand to the
// table, in hand order. When the ability is activated from the hand
// (cycling-style), the source is in the hand like any other card and is
// shown with it: the cost says "your hand", not "your other cards".
//
// Caller must hold g.mu.
func (g *Game) payRevealHandCostLocked(playerID, sourceID uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil || len(p.Hand.Cards) == 0 {
		return
	}
	ids := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for i := range p.Hand.Cards {
		ids = append(ids, p.Hand.Cards[i].InstanceID)
	}
	reason := "reveal your hand"
	if c, ok := g.LookupCardForEffect(sourceID); ok {
		reason = c.Name + " — " + reason
	}
	g.RevealForEffect(RevealSpec{Player: playerID, Source: sourceID, Reason: reason, Cards: ids})
}
