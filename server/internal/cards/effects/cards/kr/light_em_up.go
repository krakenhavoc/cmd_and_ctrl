package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "c6932bcf-d66b-42c1-8ca4-6a4573e0f6bf",
		Name:          "Light 'Em Up",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(2)},
		Targets:       effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:       effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 2}.Apply(ctx)
		},
	})
}
