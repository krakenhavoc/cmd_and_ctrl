package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monastery Flock — Creature — Bird {2}{U}, 0/5 (Khans of Tarkir,
// #5554):
//
//	"Defender, flying
//	 Morph {U} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape: the printed keywords are data the face-down 2/2
// does not have (CR 708.2a). Morph is the keyword's declaration
// (morph.go, CR 702.37): cast face down for {3} it is a nameless 2/2
// with no text (CR 708.2a), and {U} turns it face up as the same object
// (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "fd05f23e-8974-445d-9059-d706a3740d89",
		Name:             "Monastery Flock",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"defender", "flying"},
		AlternativeCosts: []game.AlternativeCost{Morph("{U}")},
	})
}
