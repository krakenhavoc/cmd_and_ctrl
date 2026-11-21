package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage-Eye Harrier — Creature — Bird Warrior {4}{W}, 1/5 (Khans of
// Tarkir, #5554):
//
//	"Flying
//	 Morph {3}{W} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape: the printed keywords are data the face-down 2/2
// does not have (CR 708.2a). Morph is the keyword's declaration
// (morph.go, CR 702.37): cast face down for {3} it is a nameless 2/2
// with no text (CR 708.2a), and {3}{W} turns it face up as the same
// object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "589d1c6e-5808-424c-a6e4-dadef90d119e",
		Name:             "Sage-Eye Harrier",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{3}{W}")},
	})
}
