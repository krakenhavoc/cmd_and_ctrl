package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wolf Cove Villager — Creature — Human Peasant {W}, 2/2 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"This creature enters tapped."
//
// The enters-tapped replacement (CR 614.1d), as Dungeon Crawler's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "549a3f53-f90e-4af0-bbc5-2c4b9df30008",
		Name:         "Wolf Cove Villager",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
	})
}
