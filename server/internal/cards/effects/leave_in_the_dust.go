package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leave in the Dust — Instant {3}{U} (Jumpstart Azorius deck request,
// #4894):
//
//	"Return target nonland permanent to its owner's hand.
//	 Draw a card."
//
// The draw follows the bounce as its continuation, so a commander's
// owner answering the CR 903.9 prompt first does not reorder the two.
// A target that became illegal in response makes the whole spell do
// nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7bb843b6-c55b-4796-b1a7-2fff7b05c276",
		Name:         "Leave in the Dust",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BounceToHand{Target: id, ThenTail: ctx.TailN(thenDrawCards, 1)}.Apply(ctx)
		},
	})
}
