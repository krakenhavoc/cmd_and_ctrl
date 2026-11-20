package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witness of the Ages — Artifact Creature — Golem {6}, 4/4 (Khans of
// Tarkir, #5554):
//
//	"Morph {5} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape with nothing else printed. Morph is the
// keyword's declaration (morph.go, CR 702.37): cast face down for {3} it
// is a nameless 2/2 with no text (CR 708.2a), and {5} turns it face up
// as the same object (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "afeb9de8-7030-4797-a9c8-5a99de3c410f",
		Name:             "Witness of the Ages",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{5}")},
	})
}
