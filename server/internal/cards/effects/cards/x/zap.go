package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zap — Instant {2}{R} (Invasion, S82 #5616):
//
//	"Zap deals 1 damage to any target.
//	 Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56115482-3fd4-45fb-b800-be68c5509cf2",
		Name:         "Zap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Draws: 1, Targets: effects.ForTargets(effects.DamageToTarget(0, 1)).Targets, Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
