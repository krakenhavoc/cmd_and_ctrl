package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scarwood Treefolk — Creature — Treefolk {3}{G}, 3/5 (Time Spiral,
// #5596):
//
//	"This creature enters tapped."
//
// The tapped entry is the shared self-replacement (CR 614.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "874b944a-8d08-4143-b477-685087396a88",
		Name:         "Scarwood Treefolk",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
	})
}
