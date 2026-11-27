package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gelatinous Genesis — Sorcery {X}{X}{G} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Create X X/X green Ooze creature tokens."
//
// X is the one number announced with the cast (CR 107.3a); both {X}
// symbols in the cost are paid at that value. The count of tokens and the
// size of each token are both X, and an X of 0 makes nothing. The token's
// size is set as it is created, so it is a 0/0 placeholder whose printed
// P/T is the card's */* (the Ooze row of the token table, shared with
// Mystic Genesis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "291e4b91-17c7-4b7b-8057-f0d361424016",
		Name:         "Gelatinous Genesis",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			tok := ctx.TextToken(effects.TokenCard("0/0 green Ooze"))
			tok.Power, tok.Toughness = x, x
			return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, x)
		},
	})
}
