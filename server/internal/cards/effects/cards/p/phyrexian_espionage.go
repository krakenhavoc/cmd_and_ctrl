package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Espionage — {2}{U} Sorcery (Dominaria United):
//
//	"Kicker {1}{B} (You may pay an additional {1}{B} as you cast this
//	 spell.)
//	 Draw two cards. If this spell was kicked, each opponent discards a
//	 card."
//
// Two steps (ResolveSteps): the discards are asked after the draw.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "32dce35e-182d-41be-9fb2-b0738d0781b8",
		Name:          "Phyrexian Espionage",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Draws: 2},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{B}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item)
			}),
		),
	})
}
