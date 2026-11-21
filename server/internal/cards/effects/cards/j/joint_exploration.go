package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Joint Exploration — Instant {1}{U} (Dominaria United):
//
//	"Kicker {G} (You may pay an additional {G} as you cast this spell.)
//	 Scry 2, then draw a card. If this spell was kicked, you may put a
//	 land card from your hand onto the battlefield."
//
// Three instructions in printed order, each waiting for the one before
// it (CR 608.2c): the scry's prompt is answered before the draw, and the
// draw lands before the land is chosen, so the drawn card may be the
// land. The land is put, not played (CR 305.4), untapped.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9eba1789-6e58-4d78-8a0d-b54a8be76ec5",
		Name:          "Joint Exploration",
		Purpose:       game.Purpose{Scry: 2, Draws: 1, Lands: 1, LandsUntapped: 1},
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{G}")},
		OnResolve: effects.ResolveSteps(
			effects.Scry{N: 2},
			effects.DrawCards{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.MayPutALandFromHand("Joint Exploration").Apply(ctx)
			}),
		),
	})
}
