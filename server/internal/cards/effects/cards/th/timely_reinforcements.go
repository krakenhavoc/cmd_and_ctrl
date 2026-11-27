package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Timely Reinforcements — Sorcery {2}{W} (Magic 2012, S82 OLD batch 452,
// #5573):
//
//	"If you have less life than an opponent, you gain 6 life. If you
//	 control fewer creatures than an opponent, create three 1/1 white
//	 Soldier creature tokens."
//
// Each sentence is judged on its own as the spell resolves, against any
// opponent (Birth of the Imperium's per-opponent reading); the life gain
// is applied before the creature count is read, in the printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64f4f0ed-e924-459e-86a4-8202fae6ef3e",
		Name:         "Timely Reinforcements",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			me := ctx.Game.PlayerByIDForEffect(item.Controller)
			if me == nil {
				return nil
			}
			lessLife := effects.EachOpponent(ctx.Game, item.Controller, func(opp uuid.UUID) bool {
				o := ctx.Game.PlayerByIDForEffect(opp)
				return o != nil && o.Life > me.Life
			})
			if lessLife {
				if err := (effects.GainLife{Amount: 6}).Apply(ctx); err != nil {
					return err
				}
			}
			isCreature := func(c game.Card) bool { return c.IsCreature() }
			mine := effects.CountControlled(ctx.Game, item.Controller, isCreature)
			fewer := effects.EachOpponent(ctx.Game, item.Controller, func(opp uuid.UUID) bool {
				return effects.CountControlled(ctx.Game, opp, isCreature) > mine
			})
			if !fewer {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 3}.Apply(ctx)
		},
	})
}
