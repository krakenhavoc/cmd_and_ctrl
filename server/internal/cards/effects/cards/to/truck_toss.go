package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Truck Toss — {2}{R}{R} Instant (Marvel Super Heroes, #4455):
//
//	"This spell costs {2} less to cast if you control a Vehicle.
//	 Truck Toss deals 4 damage to any target."
//
// Wizard's Retort's conditional self cost modifier over a Vehicle, then
// Lightning Strike's any-target damage at four.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d765055-da62-43f7-9580-98e9d164753b",
		Name:         "Truck Toss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you control a Vehicle.",
				effects.YouControlA(effects.HasSubtype("Vehicle"))),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 4}.Apply(ctx)
		},
	})
}
