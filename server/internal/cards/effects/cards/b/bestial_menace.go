package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bestial Menace — Sorcery {3}{G}{G}:
//
//	"Create a 1/1 green Snake creature token, a 2/2 green Wolf creature
//	 token, and a 3/3 green Elephant creature token."
//
// Three token creations in one resolution, one per printed token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef08d371-9d6c-4a2d-a444-d475fbf1b633",
		Name:         "Bestial Menace",
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: 1},
			effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1},
			effects.CreateToken{Template: effects.TokenCard("3/3 green Elephant"), N: 1},
		),
	})
}
