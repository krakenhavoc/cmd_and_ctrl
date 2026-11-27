package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Remove — Instant {U} (Portal Second Age, S82 older sets batch 736, #5625):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	 Return target attacking creature to its owner's hand."
//
// Just Fate's and Champion's Victory's cast condition; the target must be an attacking creature as it is chosen and as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "a1ed0f7e-56c3-4866-89c4-77b81a952955",
		Name:               "Remove",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		Targets:            effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:          effects.BounceTheTarget,
	})
}
