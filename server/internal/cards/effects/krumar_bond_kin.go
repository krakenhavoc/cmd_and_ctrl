package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krumar Bond-Kin — Creature — Orc Warrior {3}{B}{B}, 5/3 (Khans of
// Tarkir, #5554):
//
//	"Morph {4}{B} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape with nothing else printed. Morph is the
// keyword's declaration (morph.go, CR 702.37): cast face down for {3} it
// is a nameless 2/2 with no text (CR 708.2a), and {4}{B} turns it face
// up as the same object (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8e34f487-aa75-4afb-af75-3349de456c79",
		Name:             "Krumar Bond-Kin",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{4}{B}")},
	})
}
