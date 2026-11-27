package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eightfold Maze — Instant {2}{W} (Portal Three Kingdoms, S82 OLD b713, #5622):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Destroy target attacking creature."
//
// Heavy Fog's cast restriction (YouWereAttackedThisStep) over Chastise's
// attacking-creature target: judged against combat at announce and at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "9097457c-ceb3-4944-8f01-c28ddefbb270",
		Name:               "Eightfold Maze",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		Targets:            effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:          effects.DestroyTheTargetPermanent,
	})
}
