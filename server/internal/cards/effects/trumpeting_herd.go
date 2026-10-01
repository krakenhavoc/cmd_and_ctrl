package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "69501650-ed48-4ebf-9287-b0338c5bc5d5",
		Name:            "Trumpeting Herd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("3/3 green Elephant"), N: 1}.Apply(ctx)
		},
	})
}
