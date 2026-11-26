package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harsh Sustenance — Instant {1}{W}{B} (Fate Reforged):
//
//	"Harsh Sustenance deals X damage to any target and you gain X life, where X is
//	 the number of creatures you control."
//
// X is counted once, as the spell resolves, and the life is gained even when the
// damage is prevented (CR 608.2c) as long as the target is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5cf471f-032c-4ebb-823c-8841de209216",
		Name:         "Harsh Sustenance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: targets[0].ID, Amount: x}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: x}.Apply(ctx)
		},
	})
}
