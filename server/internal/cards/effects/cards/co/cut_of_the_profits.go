package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "8ce63f60-0211-46c7-8fb6-c5484ebfa029",
		Name:          "Cut of the Profits",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(3)},
		XMatters:      true,
		OnResolve:     effects.ResolveSteps(effects.DrawXStep(), effects.LoseXStep()),
	})
}
