package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Spy — Creature — Goblin Rogue {R}, 1/1 (Invasion, S82 older sets batch 664, #5616):
//
//	"Play with the top card of your library revealed."
//
// Courser of Kruphix's revealed top card with no permission attached: the
// engine shows the top card of its controller's library to every player while
// the Spy is on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "5f1e8aa2-6b85-4e2e-92e4-876470447d6f",
		Name:              "Goblin Spy",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
	})
}
