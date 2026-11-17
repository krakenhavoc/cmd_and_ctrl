package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nimble Mongoose — Creature — Mongoose {G}, 1/1 (Odyssey, #5613):
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Threshold — This creature gets +2/+2 as long as there are seven or
//	 more cards in your graveyard."
//
// Shroud is a printed keyword (CR 702.18a), which the targeting gate
// reads for every player, its controller included. The threshold is the
// layer 7c self-modifier (CR 613.4c) Krosan Beast uses, at +2/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1059f089-2693-4c89-92c5-39fe49beebb3",
		Name:            "Nimble Mongoose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Static:          []game.StaticAbility{fraThresholdSelfPT(2, 2)},
	})
}
