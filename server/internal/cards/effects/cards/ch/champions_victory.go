package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Champion's Victory — Instant {U} (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Return target attacking creature to its owner's hand."
//
// Heavy Fog's cast restriction over an attacking-creature target, judged
// against combat at announce and at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "8df38d95-1707-4f83-9d64-41f0be2a3b3b",
		Name:               "Champion's Victory",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		Targets:            effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:          effects.BounceTheTarget,
	})
}
