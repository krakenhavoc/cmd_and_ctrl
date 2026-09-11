package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Scour — Instant {U}:
//
//	"Target player mills two cards."
//	"Draw a card."
//
// A cantrip that happens to fill a graveyard — usually your own, in a
// deck that wants cards there. The mill is TARGETED and the draw is
// not, so the draw still happens when the target becomes illegal,
// which is why the two are separate statements here rather than one
// early return.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "83101ba8-a569-4827-8c53-9ca0dfcd59a7",
		Name:     "Thought Scour",
		Targets:  TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// CR 608.2b: an illegal target is skipped, but the rest
			// of the spell still resolves — "draw a card" is a
			// separate sentence with no target in it.
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetPlayer &&
				ctx.IsTargetLegal(item.Targets[0]) {
				if err := (MillCards{Player: item.Targets[0].ID, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
