package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "854fd120-9a51-4d37-9922-7e4b0464e0f5",
		Name:          "Shivan Fire",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.DamageToTarget(0, 2)),
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.WasKicked() {
				return effects.DealDamageToTheTarget(4)(ctx.Game, item)
			}
			return effects.DealDamageToTheTarget(2)(ctx.Game, item)
		},
	})
}
