package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fiery Justice — Sorcery {R}{G}{W} (Double Masters 2022, #5948):
//
//	"Fiery Justice deals 5 damage divided as you choose among any number of
//	 targets. Target opponent gains 5 life."
//
// Two target clauses chosen as the spell is cast (CR 601.2c): up to five
// targets that divide the 5 damage, each given at least 1 (CR 601.2d), and
// one opponent who gains the life. The opponent may also be one of the damage
// targets. Each clause is read for its own picks at resolution, so a target
// that became illegal takes nothing (its share is not redistributed) and
// the life is gained only if that opponent is still a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "333809cb-e196-45f2-8a67-31374438e56e",
		Name:         "Fiery Justice",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetAny().WithCount(0, 5).Dividing(effects.Divide(5)),
			effects.TargetPlayer("target opponent", effects.Opponent()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DealDividedShares(ctx, 0) }),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.ClauseTargets(1) {
					if t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
						continue
					}
					return effects.GainLife{Player: t.ID, Amount: 5}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
