package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "54530a84-8278-4c93-a2cb-f5c86619a053",
		Name:          "Cut Your Losses",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(2)},
		Targets:       effects.TargetPlayer("target player"),
		OnResolve:     effects.MillHalfOfTargetPlayer,
	})
}
