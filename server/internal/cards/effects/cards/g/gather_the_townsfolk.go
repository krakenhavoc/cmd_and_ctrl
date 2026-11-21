package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gather the Townsfolk — Sorcery {1}{W} (Innistrad Remastered):
//
//	"Create two 1/1 white Human creature tokens.
//	 Fateful hour — If you have 5 or less life, create five of those tokens
//	 instead."
//
// Fateful hour is an ability word (CR 207.2c): your life total is read as the
// spell resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff45677d-977e-4ef5-859c-52393ed71a6c",
		Name:         "Gather the Townsfolk",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 2
			if p := ctx.Game.PlayerByIDForEffect(item.Controller); p != nil && p.Life <= 5 {
				n = 5
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: n}.Apply(ctx)
		},
	})
}
