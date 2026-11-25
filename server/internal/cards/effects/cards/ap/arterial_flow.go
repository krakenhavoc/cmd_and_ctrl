package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arterial Flow — Sorcery {1}{B}{B}:
//
//	"Each opponent discards two cards. If you control a Vampire, each opponent
//	 loses 2 life and you gain 2 life."
//
// The discards come first, each opponent choosing their own (CR 608.2c); the
// Vampire is looked for after them, as the spell resolves, and the drain is a
// life loss from each opponent and one 2-life gain for the controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5e2fb92-4678-410c-8c83-42aead9d29d3",
		Name:         "Arterial Flow",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				effects.EachOpponentDiscards(ctx, 2)
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.ControlsA(ctx.TextSubtype("Vampire"))(ctx.Game, ctx.Controller()) {
					return nil
				}
				if err := (effects.EachOpponentLosesLife{N: 2}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Amount: 2}.Apply(ctx)
			}),
		),
	})
}
