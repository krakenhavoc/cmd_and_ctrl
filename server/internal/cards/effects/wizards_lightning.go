package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2bd81743-84c8-43ec-86f5-68230036827b",
		Name:         "Wizard's Lightning",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you control a Wizard.",
				YouControlA(HasSubtype("Wizard"))),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DealDamageToTheTarget(3)(ctx.Game, item)
		},
	})
}
