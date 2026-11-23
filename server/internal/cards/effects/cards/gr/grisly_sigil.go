package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "b9d9d118-a073-4e73-90fe-026462d1d895",
		Name:          "Grisly Sigil",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(1)},
		Targets:       effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 1
			if ctx.Game.PermanentWasDealtNoncombatDamageThisTurn(id) {
				n = 3
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: n}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
		},
	})
}
