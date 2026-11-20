package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assassin's Blade — Instant {1}{B} (Portal, #5630):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Destroy target nonblack attacking creature."
//
// Heavy Fog's cast restriction (youWereAttackedThisStep): the declare
// attackers step, in which a creature was declared attacking you.
// The target must be a nonblack creature that is attacking as the
// spell is cast and as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "76da2150-34b9-4483-99df-131e1c5468d5",
		Name:               "Assassin's Blade",
		Completeness:       CompletenessFull,
		CastCondition:      youWereAttackedThisStep,
		CastConditionLabel: attackedThisStepLabel,
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		Targets:            TargetCreature("target nonblack attacking creature", NonBlack(), AttackingCreature()),
		OnResolve:          destroyTheTargetPermanent,
	})
}
