package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresDiscardYourHand is the "Discard your hand" cost probe
// (#1600): an activated or mana ability whose cost carries the hand
// form of the discard component (game.DiscardCost.Hand). Kept out of
// registry.go so the seam's entry there stays one hunk.
func declaresDiscardYourHand(s effects.Spec) bool {
	for _, a := range s.Activated {
		if a.Cost.DiscardCards.DiscardsHand() {
			return true
		}
	}
	for _, m := range s.ManaAbilities {
		if m.Cost.DiscardCards.DiscardsHand() {
			return true
		}
	}
	return false
}
