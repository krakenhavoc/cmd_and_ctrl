package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b19_helpers.go — shared vocabulary for the S82 BRO batch 19 cards.

// s82Bro19BounceTarget is "return target <permanent> to its owner's hand"
// for an activated ability with one battlefield target (Drafna, Founder of
// Lat-Nam): the target goes back if it is still a permanent on the
// battlefield as the ability resolves.
func S82Bro19BounceTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	return BounceToHand{Target: target}.Apply(ctx)
}
