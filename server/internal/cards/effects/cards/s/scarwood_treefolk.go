package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarwood Treefolk — Creature — Treefolk {3}{G}, 3/5 (Time Spiral,
// #5596):
//
//	"This creature enters tapped."
//
// The tapped entry is the shared self-replacement (CR 614.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "874b944a-8d08-4143-b477-685087396a88",
		Name:         "Scarwood Treefolk",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
