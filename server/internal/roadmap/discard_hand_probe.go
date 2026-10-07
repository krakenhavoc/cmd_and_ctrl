package roadmap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declaresDiscardX is the "Discard X cards" cost probe (#2527): an
// activated ability whose discard component counts from the announced
// X (game.DiscardCost.CountFromX).
func declaresDiscardX(s effects.Spec) bool {
	for _, a := range s.Activated {
		if game.DiscardCountFromX(a.Cost.DiscardCards) {
			return true
		}
	}
	return false
}

// declaresDiscardManaValueX is the "Discard a card with mana value X"
// cost probe (#2190): an activated ability whose discard component
// announces the discarded card's mana value as X
// (game.DiscardCost.ManaValueX).
func declaresDiscardManaValueX(s effects.Spec) bool {
	for _, a := range s.Activated {
		if game.DiscardManaValueX(a.Cost.DiscardCards) {
			return true
		}
	}
	return false
}

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
