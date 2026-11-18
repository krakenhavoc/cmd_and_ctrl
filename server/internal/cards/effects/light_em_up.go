package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "c6932bcf-d66b-42c1-8ca4-6a4573e0f6bf",
		Name:          "Light 'Em Up",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(2)},
		Targets:       TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:       ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: id, Amount: 2}.Apply(ctx)
		},
	})
}
