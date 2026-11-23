package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Lore — Sorcery {1}{R} (Portal Second Age, #5625):
//
//	"Draw four cards, then discard three cards at random."
//
// Control of the Court's text: four draws, then three random discards
// from the hand as it stands once the draws have settled (CR 608.2c).
// The discards are random (CR 701.9b), so the player chooses nothing; a
// hand of fewer than three discards what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6fe945c-b561-4e35-8c8e-7dcaa2ec45ab",
		Name:         "Goblin Lore",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4, Discards: 3},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 4},
			effects.DiscardCards{N: 3},
		),
	})
}
