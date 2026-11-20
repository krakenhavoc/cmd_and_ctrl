package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Fire — Instant {R} (Dominaria, #5530):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 Shivan Fire deals 2 damage to target creature. If this spell was kicked, it deals 4 damage instead."
//
// Kicker read off the cast: 2 damage, or 4 when kicked, to the target
// creature if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "854fd120-9a51-4d37-9922-7e4b0464e0f5",
		Name:          "Shivan Fire",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(DamageToTarget(0, 2)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Kicker("{4}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.WasKicked() {
				return DealDamageToTheTarget(4)(ctx.Game, item)
			}
			return DealDamageToTheTarget(2)(ctx.Game, item)
		},
	})
}
