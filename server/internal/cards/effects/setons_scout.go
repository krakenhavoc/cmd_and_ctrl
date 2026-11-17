package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seton's Scout — Creature — Centaur Druid Scout Archer {1}{G}, 2/1
// (Torment, #5612):
//
//	"Reach (This creature can block creatures with flying.)
//	 Threshold — This creature gets +2/+2 as long as there are seven or
//	 more cards in your graveyard."
//
// Reach is a printed keyword (CR 702.17a). Threshold is an ability word
// with no rules of its own (CR 207.2c): the layer 7c self-modifier (CR
// 613.4c) Krosan Beast uses, at +2/+2, read live on every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b2a32ca-ef5b-40a5-acf7-aefac1d09e5b",
		Name:            "Seton's Scout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static:          []game.StaticAbility{fraThresholdSelfPT(2, 2)},
	})
}
