package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metamorphic Wurm — Creature — Elephant Wurm {3}{G}{G}, 3/3 (Odyssey,
// #5613):
//
//	"Threshold — This creature gets +4/+4 as long as there are seven or
//	 more cards in your graveyard."
//
// Krosan Beast's threshold at +4/+4: the layer 7c self-modifier (CR
// 613.4c), read live on every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "512477f0-1083-423b-9675-a8bc852f7d88",
		Name:         "Metamorphic Wurm",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{fraThresholdSelfPT(4, 4)},
	})
}
