package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadly Recluse — Creature — Spider {1}{G}, 1/2 (Magic 2010, #5585):
//
//	"Reach (This creature can block creatures with flying.)
//	 Deathtouch (Any amount of damage this deals to a creature is enough
//	 to destroy it.)"
//
// Thornweald Archer's keywords: reach (CR 702.17b) and deathtouch (CR 702.2b), both carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b941291-1802-4ced-9869-39fa43825550",
		Name:            "Deadly Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
	})
}
