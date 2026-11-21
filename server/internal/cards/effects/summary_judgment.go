package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ae198ce9-097b-4204-b599-12bdb36f5195",
		Name:         "Summary Judgment",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if b14InYourMainPhase(ctx.Game, ctx.Controller()) {
				return DealDamageToTheTarget(5)(ctx.Game, item)
			}
			return DealDamageToTheTarget(3)(ctx.Game, item)
		},
	})
}
