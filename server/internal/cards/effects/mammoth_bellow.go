package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mammoth Bellow — Sorcery {2}{G}{U}{R} (#3931):
//
//	"Create a 5/5 green Elephant creature token.
//	 Harmonize {5}{G}{U}{R}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "6f2e4607-bece-42ba-a2a3-5dc50f246793",
		Name:             "Mammoth Bellow",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{5}{G}{U}{R}")},
		Purpose:          game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("5/5 green Elephant"), N: 1}.Apply(ctx)
		},
	})
}
