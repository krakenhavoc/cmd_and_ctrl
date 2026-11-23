package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starstorm — Instant {X}{R}{R}:
//
//	"Starstorm deals X damage to each creature.
//	 Cycling {3}"
//
// Mass damage, one damage instance (damageEachMatching); a creature whose
// toughness is above X survives. Cycling is the shared CR 702.29a ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34d112c3-cc2d-4428-bd7f-568bda61bd30",
		Name:         "Starstorm",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{3}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), ctx.X())
		},
	})
}
