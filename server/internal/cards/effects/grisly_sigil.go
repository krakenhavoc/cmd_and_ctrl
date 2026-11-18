package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "b9d9d118-a073-4e73-90fe-026462d1d895",
		Name:          "Grisly Sigil",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(1)},
		Targets:       TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 1
			if ctx.Game.PermanentWasDealtNoncombatDamageThisTurn(id) {
				n = 3
			}
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: n}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
		},
	})
}
