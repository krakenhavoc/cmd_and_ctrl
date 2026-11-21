package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "144e3d90-8ed8-4c21-941d-fc59cec042a2",
		Name:          "Dig Up the Body",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(1)},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.MillCards{N: 2}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Dig Up the Body — you may return a creature card from your graveyard to your hand",
					Match:    game.Card.IsCreature,
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
		),
	})
}
