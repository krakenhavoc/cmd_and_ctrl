package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pull from Tomorrow — Instant for {X}{U}{U}:
//
//	"Draw X cards, then discard a card."
//
// The discard is a cost of doing business on every other deck and a
// bonus on this one. Printed order matters and the engine keeps it:
// the discard choice is queued against the post-draw hand, so a card
// just drawn is a legal pitch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1a23235-3076-475c-a68a-db29cf2a9dba",
		Name:         "Pull from Tomorrow",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The discard is asked after the draw's own prompts (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player: ctx.Controller(),
					Source: ctx.Source(),
					N:      1,
				})
				return nil
			}),
		),
	})
}
