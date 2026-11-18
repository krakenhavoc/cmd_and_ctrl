package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "8ce63f60-0211-46c7-8fb6-c5484ebfa029",
		Name:          "Cut of the Profits",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(3)},
		XMatters:      true,
		OnResolve:     ResolveSteps(drawXStep(), loseXStep()),
	})
}
