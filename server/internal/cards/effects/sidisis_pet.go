package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sidisi's Pet — Creature — Zombie Ape {3}{B}, 1/4 (Khans of Tarkir,
// #5554):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)
//	 Morph {1}{B} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape: the printed keywords are data the face-down 2/2
// does not have (CR 708.2a). Morph is the keyword's declaration
// (morph.go, CR 702.37): cast face down for {3} it is a nameless 2/2
// with no text (CR 708.2a), and {1}{B} turns it face up as the same
// object (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9b183d76-1dee-4ecf-9226-e1aa600b1247",
		Name:             "Sidisi's Pet",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{Morph("{1}{B}")},
	})
}
