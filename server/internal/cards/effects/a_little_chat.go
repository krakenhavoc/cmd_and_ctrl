package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "1167fe30-32b8-4381-8067-a05cd345977b",
		Name:          "A Little Chat",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(1)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return LookAtTopThenTakeOneRestOnBottom(2,
				"A Little Chat — put one of them into your hand and the other on the bottom")(ctx.Game, item)
		},
	})
}
