package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fen Stalker — Creature — Nightstalker {3}{B}, 3/2 (Prophecy, #5794):
//
//	"This creature has fear as long as you control no untapped lands.
//	 (It can't be blocked except by artifact creatures and/or black
//	 creatures.)"
//
// A layer 6 grant of fear (CR 613.1f, 702.36) to itself while its
// controller controls no untapped lands, read live (CR 611.3a,
// no_untapped_lands.go). Fear is the engine's block rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfd00c0e-d034-4b39-a3d3-f8b8842045dd",
		Name:         "Fen Stalker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SelfKeywordWhileNoUntappedLands("fear")},
	})
}
