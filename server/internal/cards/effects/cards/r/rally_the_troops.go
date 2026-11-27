package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rally the Troops — Instant {W}
// (Portal Three Kingdoms, #5622):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step."
//	"Untap all creatures you control."
//
// The cast condition is Heavy Fog's: the declare attackers step, after a
// creature was declared attacking you (CR 508.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "555927b3-bbc8-48b7-9a04-e8ebaabccf3e",
		Name:               "Rally the Troops",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
		},
	})
}
