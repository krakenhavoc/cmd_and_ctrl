package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "144e3d90-8ed8-4c21-941d-fc59cec042a2",
		Name:          "Dig Up the Body",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(1)},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return MillCards{N: 2}.Apply(ctx) }),
			Instruction(func(ctx *Context) error {
				return ReturnChosenFromGraveyard{
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
