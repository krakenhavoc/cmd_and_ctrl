package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Just Fate — Instant {2}{W} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	 Destroy target attacking creature."
//
// The cast condition is Heavy Fog's and Assassin's Blade's; the target must be an attacking creature as it is chosen and as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "2619692a-07a0-4f50-88b0-4cec6dbc8600",
		Name:               "Just Fate",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		Targets:            effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:          effects.DestroyTheTargetPermanent,
	})
}
