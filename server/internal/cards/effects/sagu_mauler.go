package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sagu Mauler — Creature — Beast {4}{G}{U}, 6/6 (Khans of Tarkir,
// #5554):
//
//	"Trample, hexproof
//	 Morph {3}{G}{U} (You may cast this card face down as a 2/2
//	 creature for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape: the printed keywords are data the face-down 2/2
// does not have (CR 708.2a). Morph is the keyword's declaration
// (morph.go, CR 702.37): cast face down for {3} it is a nameless 2/2
// with no text (CR 708.2a), and {3}{G}{U} turns it face up as the same
// object (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8d877ee9-1751-48a1-9f2f-02897dae153b",
		Name:             "Sagu Mauler",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample", "hexproof"},
		AlternativeCosts: []game.AlternativeCost{Morph("{3}{G}{U}")},
	})
}
