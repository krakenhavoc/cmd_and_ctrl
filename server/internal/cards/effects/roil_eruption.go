package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "6798163a-864f-4844-96b1-77585a7e7ab4",
		Name:          "Roil Eruption",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(DamageToTarget(0, 3)),
		Targets:       TargetAny(),
		OptionalCosts: []game.AdditionalCost{Kicker("{5}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amount := 3
			if ctx.WasKicked() {
				amount = 5
			}
			for _, t := range ctx.LegalTargets() {
				return DealDamage{Target: t.ID, Amount: amount, Source: item.SourceCardID}.Apply(ctx)
			}
			return nil
		},
	})
}
