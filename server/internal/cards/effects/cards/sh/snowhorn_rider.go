package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snowhorn Rider — Creature — Human Warrior {3}{G}{U}{R}, 5/5 (Khans of
// Tarkir, #5554):
//
//	"Trample
//	 Morph {2}{G}{U}{R} (You may cast this card face down as a 2/2
//	 creature for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape: the printed keywords are data the face-down 2/2
// does not have (CR 708.2a). Morph is the keyword's declaration
// (morph.go, CR 702.37): cast face down for {3} it is a nameless 2/2
// with no text (CR 708.2a), and {2}{G}{U}{R} turns it face up as the
// same object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3d5a753b-ed2d-4646-ade3-5daeba2a7270",
		Name:             "Snowhorn Rider",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{G}{U}{R}")},
	})
}
