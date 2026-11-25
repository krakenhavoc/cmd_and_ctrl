package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jaya's Greeting — Instant {1}{R} (War of the Spark, #5520):
//
//	"Jaya's Greeting deals 3 damage to target creature. Scry 1."
//
// The scry happens even when the creature is gone (CR 608.2b only skips the
// instructions that need the illegal target); with no legal target at all the
// spell does not resolve.
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.DamageToTarget(0, 3))
	purpose.Scry = 1
	effects.Register(effects.Spec{
		OracleID:     "1be474c6-5a9f-4c58-adc6-c1508348adea",
		Name:         "Jaya's Greeting",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
