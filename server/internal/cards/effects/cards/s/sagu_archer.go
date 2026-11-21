package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sagu Archer — Creature — Snake Archer {4}{G}, 2/5 (Khans of Tarkir,
// #5554):
//
//	"Reach (This creature can block creatures with flying.)
//	 Morph {4}{G} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape: the printed keywords are data the face-down 2/2
// does not have (CR 708.2a). Morph is the keyword's declaration
// (morph.go, CR 702.37): cast face down for {3} it is a nameless 2/2
// with no text (CR 708.2a), and {4}{G} turns it face up as the same
// object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "cf8a8482-e243-4155-bc21-a1e474c6f1e7",
		Name:             "Sagu Archer",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{4}{G}")},
	})
}
