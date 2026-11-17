package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frightcrawler — Creature — Horror {1}{B}, 1/1 (Odyssey, #5613):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +2/+2 and can't block."
//
// Fear is a printed keyword (CR 702.36). The threshold is two statics read
// live: the layer 7c self-modifier (CR 613.4c) and a layer 6 CantBlock on
// itself (thresholdSelfCantBlock), so below seven cards it blocks as a
// 1/1 and from the seventh card it attacks as a 3/3 and cannot block.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "96b611e9-38cc-4326-bb38-73084e99e6f5",
		Name:            "Frightcrawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Static:          []game.StaticAbility{fraThresholdSelfPT(2, 2), thresholdSelfCantBlock()},
	})
}
