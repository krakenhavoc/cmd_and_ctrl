package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hordeling Outburst — Sorcery {1}{R}{R}:
//
//	"Create three 1/1 red Goblin creature tokens."
//
// Dragon Fodder's big brother: one more mana, one more body, and the
// three-for-one that makes a sacrifice outlet worth the card. Same
// token, same lack of rider.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6450b8e-eb18-431c-9eb7-7daf107978b2",
		Name:         "Hordeling Outburst",
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.RedGoblinToken(),
				N:          3,
			}.Apply(ctx)
		},
	})
}
