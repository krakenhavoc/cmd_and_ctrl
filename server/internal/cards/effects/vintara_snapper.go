package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vintara Snapper — Creature — Turtle {G}{G}, 2/2 (Prophecy, #5794):
//
//	"This creature has shroud as long as you control no untapped lands.
//	 (It can't be the target of spells or abilities.)"
//
// A layer 6 grant of shroud (CR 613.1f, 702.18a) to itself while its
// controller controls no untapped lands, read live (CR 611.3a,
// no_untapped_lands.go). Shroud is the engine's targeting rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57c5cbd6-432d-4d48-9fc6-5b8a6eb223f1",
		Name:         "Vintara Snapper",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SelfKeywordWhileNoUntappedLands("shroud")},
	})
}
