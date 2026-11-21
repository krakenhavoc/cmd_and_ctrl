package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Behemoth — Creature — Beast {5}{W}, 3/6 (Khans of Tarkir, #5554):
//
//	"Morph {4}{W} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape with nothing else printed. Morph is the
// keyword's declaration (morph.go, CR 702.37): cast face down for {3} it
// is a nameless 2/2 with no text (CR 708.2a), and {4}{W} turns it face
// up as the same object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4792a05e-b557-4132-8303-235feefe0d99",
		Name:             "War Behemoth",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{4}{W}")},
	})
}
