package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Decree — World Enchantment {2}{U}{U}:
//
//	"All creatures lose flying and islandwalk."
//
// Gravity Sphere with islandwalk added: one layer-6 removal over every
// creature on the battlefield (LoseKeywords). Islandwalk is the
// engine's landwalk token, the one the block check reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06eb8b90-aad8-41c1-bbc5-85ded2fe76bc",
		Name:         "Mystic Decree",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{LoseKeywords(AllCreatures, "flying", "islandwalk")},
	})
}
