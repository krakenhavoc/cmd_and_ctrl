package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "40b0b93d-4075-42d6-bcd8-b2f9910cc7d2",
		Name:          "Audacious Swap",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(2)},
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets:       effects.TargetPermanent("target nonenchantment permanent", effects.Permanent(), effects.Not(effects.Enchantment())),
		OnResolve:     effects.AudaciousSwapResolve,
	})
}
