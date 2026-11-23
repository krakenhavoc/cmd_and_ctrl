package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wurmcalling — Sorcery {X}{G} (Time Spiral, #5596):
//
//	"Buyback {2}{G} (You may pay an additional {2}{G} as you cast this
//	 spell. If you do, put this card into your hand as it resolves.)
//	 Create an X/X green Wurm creature token."
//
// The table's green Wurm at X/X (a 0/0 at X = 0, which dies as a
// state-based action, CR 704.5f). Buyback is the shared constructor (CR
// 702.27a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "8dd0f86d-6f9e-47e3-97dd-3a3fc075ea93",
		Name:          "Wurmcalling",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		XMatters:      true,
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{2}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			tok := ctx.TextToken(effects.TokenCard("6/6 green Wurm"))
			tok.Power, tok.Toughness, tok.PrintedPTKnown = x, x, true
			return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
		},
	})
}
