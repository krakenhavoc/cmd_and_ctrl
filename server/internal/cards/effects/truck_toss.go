package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7d765055-da62-43f7-9580-98e9d164753b",
		Name:         "Truck Toss",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you control a Vehicle.",
				YouControlA(HasSubtype("Vehicle"))),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 4}.Apply(ctx)
		},
	})
}
