package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turn into a Pumpkin — Instant {3}{U} (Throne of Eldraine, #5516):
//
//	"Return target nonland permanent to its owner's hand. Draw a card.
//	 Adamant — If at least three blue mana was spent to cast this spell, create a Food token."
//
// The bounce, the draw and the adamant Food, each a step that waits for any
// prompt the one before it asked (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "cfda6f65-3755-40e5-8061-a21941c4b1bd",
		Name:           "Turn into a Pumpkin",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"U"},
		Purpose:        game.Purpose{Draws: 1},
		Targets:        TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return bounceTheTarget(ctx.Item, ctx) }),
			DrawCards{N: 1},
			eldB1AdamantSpentFood("U"),
		),
	})
}
