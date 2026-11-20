package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forbidden Friendship — Sorcery {1}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Create a 1/1 red Dinosaur creature token with haste and a 1/1 white
//	 Human Soldier creature token."
//
// Two tokens from the table, the Dinosaur first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6192840-5e5d-48fa-a256-3b1e8c1c307a",
		Name:         "Forbidden Friendship",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("1/1 red Dinosaur with haste"), N: 1},
			CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1},
		),
	})
}
