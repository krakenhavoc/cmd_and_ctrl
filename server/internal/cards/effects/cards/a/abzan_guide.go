package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Guide — Creature — Human Warrior {3}{W}{B}{G}, 4/4 (Khans of
// Tarkir, #5554):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)
//	 Morph {2}{W}{B}{G} (You may cast this card face down as a 2/2
//	 creature for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape: the printed keywords are data the face-down 2/2
// does not have (CR 708.2a). Morph is the keyword's declaration
// (morph.go, CR 702.37): cast face down for {3} it is a nameless 2/2
// with no text (CR 708.2a), and {2}{W}{B}{G} turns it face up as the
// same object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "85b2fb89-5ebe-4c13-bc12-38d1417e5880",
		Name:             "Abzan Guide",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{W}{B}{G}")},
	})
}
