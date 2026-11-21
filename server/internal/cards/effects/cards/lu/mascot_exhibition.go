package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mascot Exhibition — Sorcery — Lesson {7} (Strixhaven: School of Mages,
// #5502):
//
//	"Create a 2/1 white and black Inkling creature token with flying, a
//	 3/2 red and white Spirit creature token, and a 4/4 blue and red
//	 Elemental creature token."
//
// The three mascots of Inkling, Spirit and Elemental Summoning, created in
// printed order as one resolution. Being a Lesson is its type line, which
// learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1ea2cba-f184-4a27-9971-3a56b0a0e390",
		Name:         "Mascot Exhibition",
		Purpose:      game.Purpose{Tokens: 3},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("2/1 white and black Inkling with flying"), N: 1},
			effects.CreateToken{Template: effects.TokenCard("3/2 red and white Spirit"), N: 1},
			effects.CreateToken{Template: effects.TokenCard("4/4 blue and red Elemental"), N: 1},
		),
	})
}
