package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woolly Loxodon — Creature — Elephant Warrior {5}{G}{G}, 6/7 (Khans of
// Tarkir, #5554):
//
//	"Morph {5}{G} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape with nothing else printed. Morph is the
// keyword's declaration (morph.go, CR 702.37): cast face down for {3} it
// is a nameless 2/2 with no text (CR 708.2a), and {5}{G} turns it face
// up as the same object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b588df7a-64d8-449e-b573-13badc8cdf56",
		Name:             "Woolly Loxodon",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{5}{G}")},
	})
}
