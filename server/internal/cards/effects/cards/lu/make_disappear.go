package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "5f12f665-f083-48ae-b61a-541051c996ec",
		Name:          "Make Disappear",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(1)},
		Targets:       effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CounterTheTargetSpellUnlessPaid("Make Disappear", "{2}")(ctx.Game, item)
		},
	})
}
