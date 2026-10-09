package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Something Worth Saving — Instant {1}{G} (Reality Fracture, tracker #2795):
//
//	"Mill four cards. You may put a permanent card from among them into
//	 your hand. You gain 1 life."
//
// The pick is over the cards that actually reached the graveyard (CR
// 400.7), and the life is gained once it is answered, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46cf0e3e-3c3e-4eb8-a71a-4b4b87744ced",
		Name:         "Something Worth Saving",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return rfSpellBMillThenMayTakePermanent(ctx, 4,
				"Something Worth Saving — put a permanent card milled this way into your hand",
				func(c *Context) error {
					return GainLife{Player: c.Controller(), Amount: 1}.Apply(c)
				})
		},
	})
}
