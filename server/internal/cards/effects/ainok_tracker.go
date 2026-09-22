package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ainok Tracker — 3/3 Dog Scout for {5}{R}:
//
//	"First strike. Morph {4}{R}"
//
// The plainest morph in the game and the reason it is in the catalog:
// everything the card does beyond its keyword is the keyword, so this
// file is the proof that #1194's engine carries morph with nothing on
// the card's side. Cast face down for {3} it is a nameless 2/2 with
// no first strike and no text (CR 708.2a); {4}{R} at any time turns it
// face up into the 3/3 that strikes first, without the permanent
// becoming a new object (CR 708.8) — so a face-down Tracker that has
// been attacking all along keeps attacking, and now does so first.
func init() {
	Register(Spec{
		OracleID:        "bf84a598-12d3-406d-8eeb-40592e782b87",
		Name:            "Ainok Tracker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		AlternativeCosts: []game.AlternativeCost{
			Morph("{4}{R}"),
		},
	})
}
