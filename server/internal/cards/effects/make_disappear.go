package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "5f12f665-f083-48ae-b61a-541051c996ec",
		Name:          "Make Disappear",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(1)},
		Targets:       TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return counterTheTargetSpellUnlessPaid("Make Disappear", "{2}")(ctx.Game, item)
		},
	})
}
