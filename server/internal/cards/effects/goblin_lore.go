package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d6fe945c-b561-4e35-8c8e-7dcaa2ec45ab",
		Name:         "Goblin Lore",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 4, Discards: 3},
		OnResolve: ResolveSteps(
			DrawCards{N: 4},
			DiscardCards{N: 3},
		),
	})
}
