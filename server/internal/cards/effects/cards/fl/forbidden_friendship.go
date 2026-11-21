package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forbidden Friendship — Sorcery {1}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Create a 1/1 red Dinosaur creature token with haste and a 1/1 white
//	 Human Soldier creature token."
//
// Two tokens from the table, the Dinosaur first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6192840-5e5d-48fa-a256-3b1e8c1c307a",
		Name:         "Forbidden Friendship",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 red Dinosaur with haste"), N: 1},
			effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1},
		),
	})
}
