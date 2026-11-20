package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whip-Spine Drake — Creature — Drake {3}{U}{U}, 3/3 (Future Sight, #5594):
//
//	"Flying
//	 Morph {2}{W} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// A flier with morph (CR 702.37a). The face-down cast for {3}, the 2/2
// body and the turn-face-up special action are the keyword's; the card
// file names only the printed morph cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "70e165a1-494a-4728-8d7e-1a9f306414db",
		Name:             "Whip-Spine Drake",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Morph("{2}{W}")},
	})
}
