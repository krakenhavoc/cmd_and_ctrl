package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrath of the Bloodmane — Instant {2}{R} (Reality Fracture, tracker #2795):
//
//	"This spell costs {1} less to cast if you control a legendary
//	 creature.
//	 Wrath of the Bloodmane deals 4 damage to target creature or
//	 planeswalker."
//
// The discount is a self cost modifier (ADR 0048). The spell is an
// instant, so it can never count itself as the legendary creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1994d6a-984d-4b75-9198-1f1822eab6eb",
		Name:         "Wrath of the Bloodmane",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a legendary creature.",
				wrathOfTheBloodmaneControlsLegend),
		},
		Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 4}.Apply(ctx)
		},
	})
}

// wrathOfTheBloodmaneControlsLegend is "if you control a legendary
// creature", read off the post-layer battlefield.
func wrathOfTheBloodmaneControlsLegend(q game.CostQuery) bool {
	if q.Game == nil || q.Game.Battlefield == nil {
		return false
	}
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller && c.IsCreature() && c.IsLegendary() {
			return true
		}
	}
	return false
}
