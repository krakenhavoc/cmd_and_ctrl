package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "df45a344-5937-4a3f-a80f-581ea456ea05",
		Name:          "Rob the Archives",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(1)},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
