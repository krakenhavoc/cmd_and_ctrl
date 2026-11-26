package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguimancy — Sorcery {4}{B} (Born of the Gods, S82 OLD b378, #5559):
//
//	"You draw X cards and you lose X life, where X is your devotion to black."
//
// Skeletal Scrying's two steps: the life loss waits for a prompt the draw asked
// (a dredge offer; CR 608.2c). X is your devotion to black (CR 700.5), counted
// at each step: nothing the draw does changes the permanents that count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0a85525-b174-41d9-8871-870af29d3480",
		Name:         "Sanguimancy",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: effects.S82OldB378DevotionOf(ctx, "B")}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if x := effects.S82OldB378DevotionOf(ctx, "B"); x > 0 {
					return effects.GainLife{Player: ctx.Controller(), Amount: -x}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
