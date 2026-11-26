package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slime Molding — Sorcery {X}{G} (Return to Ravnica, #5566):
//
//	"Create an X/X green Ooze creature token."
//
// Wurmcalling's X/X token on the table's green Ooze: the token is made at X/X
// (a 0/0 at X = 0, which dies as a state-based action, CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e01c8122-9159-4f28-ac6c-338bd889650e",
		Name:         "Slime Molding",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			tok := ctx.TextToken(effects.TokenCard("1/1 green Ooze"))
			tok.Power, tok.Toughness, tok.PrintedPTKnown = x, x, true
			return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
		},
	})
}
