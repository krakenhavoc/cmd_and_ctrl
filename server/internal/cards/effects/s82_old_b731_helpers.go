package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b731_helpers.go — Urza's Saga, batch 731 (#5624, S82).

// OldB731ReturnThisToOwnersHand is "return this creature to its owner's hand"
// as a triggered ability's whole effect (Imaginary Pet): the object that
// triggered, if it is still that object (CR 400.7); one that left and came back
// is a new object and stays.
func OldB731ReturnThisToOwnersHand(g *game.Game, item *game.StackItem) error {
	if SourceIsNewObject(g, item) {
		return nil
	}
	return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
