package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "df45a344-5937-4a3f-a80f-581ea456ea05",
		Name:          "Rob the Archives",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(1)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
