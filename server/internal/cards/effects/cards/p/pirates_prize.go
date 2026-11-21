package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pirate's Prize — Sorcery {3}{U} (Ixalan, #5532):
//
//	"Draw two cards. Create a Treasure token."
//
// Glimmerburst's shape: two cards, then a Treasure; the Treasure waits
// for any prompt the draw asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6de8bdc0-9b2a-4105-bbac-d5a2f73f52dc",
		Name:         "Pirate's Prize",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Tokens: 1},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.CreateToken{Template: effects.TreasureToken(), N: 1},
		),
	})
}
