package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demon's Due — Instant {3}{B} (Streets of New Capenna, #5491):
//
//	"Scry 2, then draw two cards. You lose 2 life."
//
// Each sentence waits for the prompt the one before asked (CR 608.2c), so
// the draw takes the card the scry left on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf65a5f5-5536-4616-a09e-2416439c7bc6",
		Name:         "Demon's Due",
		Purpose:      game.Purpose{Scry: 2, Draws: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
			}),
		),
	})
}
