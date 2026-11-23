package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magma Jet — Instant {1}{R} (Fifth Dawn, #5604):
//
//	"Magma Jet deals 2 damage to any target. Scry 2."
//
// Shock's 2 damage, then the scry, which happens whenever the spell
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f292253-64d3-4cf2-881f-a3eea4fda388",
		Name:         "Magma Jet",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Scry: 2, Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(2)(item, ctx); err != nil {
				return err
			}
			return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
