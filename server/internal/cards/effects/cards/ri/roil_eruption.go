package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roil Eruption — Sorcery {1}{R} (Foundations Jumpstart):
//
//	"Kicker {5} (You may pay an additional {5} as you cast this spell.)
//	 Roil Eruption deals 3 damage to any target. If this spell was kicked,
//	 it deals 5 damage instead."
//
// Burst Lightning's shape at 3 and 5.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "6798163a-864f-4844-96b1-77585a7e7ab4",
		Name:          "Roil Eruption",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:       effects.TargetAny(),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{5}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			amount := 3
			if ctx.WasKicked() {
				amount = 5
			}
			for _, t := range ctx.LegalTargets() {
				return effects.DealDamage{Target: t.ID, Amount: amount, Source: item.SourceCardID}.Apply(ctx)
			}
			return nil
		},
	})
}
