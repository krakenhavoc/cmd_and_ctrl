package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moment of Triumph — Instant {W} (Foundations, #4465):
//
//	"Target creature gets +2/+2 until end of turn. You gain 2 life."
//
// Moment of Craving's shape with a pump: the still-legal target gets
// +2/+2 (boostTargetUntilEOT), then the caster gains 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57579799-2582-417f-9864-4560cac39ac3",
		Name:         "Moment of Triumph",
		Purpose:      game.Purpose{LifeGain: 2, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 2))},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTargetUntilEOT("Moment of Triumph — +2/+2", 2, 2)(ctx.Game, item); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}
