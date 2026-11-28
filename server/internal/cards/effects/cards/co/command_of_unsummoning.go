package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Command of Unsummoning — Instant {2}{U} (Portal, S82 OLD batch 781, #5630):
//
//	"Cast this spell only during the declare attackers step and only if you've
//	 been attacked this step.
//	 Return one or two target attacking creatures to their owner's hand."
//
// Champion's Victory's cast restriction over one or two attacking-creature
// targets; each target that is still legal as it resolves returns to its
// owner's hand (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "fd57dfa5-f29d-4b79-8748-c34a73efb7f0",
		Name:               "Command of Unsummoning",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.YouWereAttackedThisStep,
		CastConditionLabel: effects.AttackedThisStepLabel,
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		Targets:            effects.TargetCreature("one or two target attacking creatures", effects.AttackingCreature()).WithCount(1, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.NeoB20BounceTargetedPermanents(ctx)
		},
	})
}
