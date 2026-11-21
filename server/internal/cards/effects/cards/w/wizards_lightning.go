package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wizard's Lightning — Instant {2}{R} (Dominaria, #5530):
//
//	"This spell costs {2} less to cast if you control a Wizard.
//	 Wizard's Lightning deals 3 damage to any target."
//
// Truck Toss's discount, keyed on a Wizard: {2} less while you control
// one as you cast it (CR 601.2f), then 3 damage to any target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bd81743-84c8-43ec-86f5-68230036827b",
		Name:         "Wizard's Lightning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you control a Wizard.",
				effects.YouControlA(effects.HasSubtype("Wizard"))),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealDamageToTheTarget(3)(ctx.Game, item)
		},
	})
}
