package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "1167fe30-32b8-4381-8067-a05cd345977b",
		Name:          "A Little Chat",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(1)},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LookAtTopThenTakeOneRestOnBottom(2,
				"A Little Chat — put one of them into your hand and the other on the bottom")(ctx.Game, item)
		},
	})
}
