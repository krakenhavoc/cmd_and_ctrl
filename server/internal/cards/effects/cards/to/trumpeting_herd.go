package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trumpeting Herd — Sorcery {2}{G}{G}:
//
//	"Create a 3/3 green Elephant creature token.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69501650-ed48-4ebf-9287-b0338c5bc5d5",
		Name:            "Trumpeting Herd",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("3/3 green Elephant"), N: 1}.Apply(ctx)
		},
	})
}
