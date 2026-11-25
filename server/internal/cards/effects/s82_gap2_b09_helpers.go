package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b09_helpers.go — shared shapes of GAP2 batch 9 (#5948): the
// card-side names of the #6093 and #6095 seams. A file of its own, so
// concurrent batches never collide.

// DurationUntilEndOfCombatOnYourNextTurn is "until end of combat on your
// next turn" (#6093, Brazen Cannonade): it lasts through the rest of this
// turn and ends as `player`'s next turn's combat phase ends (CR 511.3).
// See game.UntilEndOfCombatOnYourNextTurn.
func DurationUntilEndOfCombatOnYourNextTurn(ctx *Context, player uuid.UUID) game.Duration {
	return ctx.Game.UntilEndOfCombatOnYourNextTurnDuration(player)
}

// RemoveAPlusOneCounterInsteadOfUntapping is the quoted ability "If this
// creature would untap during your untap step, remove a +1/+1 counter from
// it instead. If you do, untap it." as an untap replacement a grant bundle
// carries (#6095, game.UntapCounterRemoval).
func RemoveAPlusOneCounterInsteadOfUntapping(label string) game.UntapCounterRemoval {
	return game.UntapCounterRemoval{Counter: game.CounterPlusOne, Label: label}
}

// TargetPlayerGetsTheEmblem is "Target player gets an emblem with …" as an
// activated ability's whole effect (Rowan Kenrith's and Will Kenrith's −8):
// the card's EmblemSpec goes to the first still-legal target player (CR
// 114.2), and "you" on the emblem is that player.
func TargetPlayerGetsTheEmblem(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	p, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return CreateEmblem{Player: p}.Apply(ctx)
}
