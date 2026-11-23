package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volt Charge — Instant {2}{R}:
//
//	"Volt Charge deals 3 damage to any target. Proliferate. (Choose any
//	 number of permanents and/or players, then give each another counter of
//	 each kind already there.)"
//
// Lightning Bolt's damage, then the controller's proliferate choice
// (Proliferate, #2525). With its only target gone the spell does not
// resolve at all (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c57907d1-c5a1-45c7-80a8-38fedf86dd19",
		Name:         "Volt Charge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(3)(item, ctx); err != nil {
				return err
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
