package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "54530a84-8278-4c93-a2cb-f5c86619a053",
		Name:          "Cut Your Losses",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(2)},
		Targets:       TargetPlayer("target player"),
		OnResolve:     millHalfOfTargetPlayer,
	})
}
