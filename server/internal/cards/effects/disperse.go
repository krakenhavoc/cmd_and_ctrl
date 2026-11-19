package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disperse — Instant {1}{U} (Scars of Mirrodin, #5577):
//
//	"Return target nonland permanent to its owner's hand."
//
// Boomerang Basics' bounce without the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d8fe3f8-5027-4f9e-999e-e277caa96478",
		Name:         "Disperse",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return bounceTheTargetToHand(ctx.Game, item)
		},
	})
}
