package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summary Judgment — Instant {1}{W} (Ravnica Remastered):
//
//	"Summary Judgment deals 3 damage to target tapped creature.
//	 Addendum — If you cast this spell during your main phase, it deals 5
//	 damage instead."
//
// Addendum is read as the spell resolves (it cannot change phase on the
// stack): it is the caster's own main phase. The tapped clause is judged at
// cast and again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae198ce9-097b-4204-b599-12bdb36f5195",
		Name:         "Summary Judgment",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
				return effects.DealDamageToTheTarget(5)(ctx.Game, item)
			}
			return effects.DealDamageToTheTarget(3)(ctx.Game, item)
		},
	})
}
