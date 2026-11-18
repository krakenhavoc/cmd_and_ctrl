package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "40b0b93d-4075-42d6-bcd8-b2f9910cc7d2",
		Name:          "Audacious Swap",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(2)},
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets:       TargetPermanent("target nonenchantment permanent", Permanent(), Not(Enchantment())),
		OnResolve:     audaciousSwapResolve,
	})
}
