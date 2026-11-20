package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surrakar Marauder — Creature — Surrakar {1}{B}, 2/1 (Zendikar, #5583):
//
//	"Landfall — Whenever a land you control enters, this creature gains
//	 intimidate until end of turn. (It can't be blocked except by
//	 artifact creatures and/or creatures that share a color with it.)"
//
// Landfall grants intimidate (CR 702.13) to the creature itself until
// end of turn, a layer-6 effect pinned to the source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23c52252-1e65-4e04-98d3-0ac06a565016",
		Name:         "Surrakar Marauder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Surrakar Marauder — gains intimidate until end of turn (landfall)",
				thisGainsKeywordUntilEndOfTurn("intimidate", "Surrakar Marauder — gains intimidate until end of turn")),
		},
	})
}
