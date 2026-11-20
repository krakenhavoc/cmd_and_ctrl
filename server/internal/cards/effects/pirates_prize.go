package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pirate's Prize — Sorcery {3}{U} (Ixalan, #5532):
//
//	"Draw two cards. Create a Treasure token."
//
// Glimmerburst's shape: two cards, then a Treasure; the Treasure waits
// for any prompt the draw asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6de8bdc0-9b2a-4105-bbac-d5a2f73f52dc",
		Name:         "Pirate's Prize",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Tokens: 1},
		OnResolve: ResolveSteps(
			DrawCards{N: 2},
			CreateToken{Template: TreasureToken(), N: 1},
		),
	})
}
