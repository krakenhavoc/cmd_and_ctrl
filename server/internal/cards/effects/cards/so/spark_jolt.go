package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spark Jolt — Instant {R} (Theros):
//
//	"Spark Jolt deals 1 damage to any target.
//	 Scry 1."
//
// Magma Jet's shape: the damage, then the scry, which happens whenever the
// spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f24634f3-49aa-4362-8faa-93598b9324d9",
		Name:         "Spark Jolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Scry: 1, Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(1)(item, ctx); err != nil {
				return err
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
