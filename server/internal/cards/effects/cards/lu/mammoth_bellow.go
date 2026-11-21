package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mammoth Bellow — Sorcery {2}{G}{U}{R} (#3931):
//
//	"Create a 5/5 green Elephant creature token.
//	 Harmonize {5}{G}{U}{R}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6f2e4607-bece-42ba-a2a3-5dc50f246793",
		Name:             "Mammoth Bellow",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Harmonize("{5}{G}{U}{R}")},
		Purpose:          game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("5/5 green Elephant"), N: 1}.Apply(ctx)
		},
	})
}
