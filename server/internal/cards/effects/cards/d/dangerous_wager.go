package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dangerous Wager — {1}{R} Instant: "Discard your hand, then draw two cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e0d60d7-cb27-4e45-80d4-c29aef5d56b6",
		Name:         "Dangerous Wager",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: 2}.Apply(ctx)
			}),
		),
	})
}
