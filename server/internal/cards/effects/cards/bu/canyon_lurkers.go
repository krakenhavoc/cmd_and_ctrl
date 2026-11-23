package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canyon Lurkers — Creature — Human Rogue {4}{R}, 5/2 (Khans of Tarkir,
// #5554):
//
//	"Morph {3}{R} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Ainok Tracker's shape with nothing else printed. Morph is the
// keyword's declaration (morph.go, CR 702.37): cast face down for {3} it
// is a nameless 2/2 with no text (CR 708.2a), and {3}{R} turns it face
// up as the same object (CR 708.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a744c5d7-c903-446a-bb30-30788058d9db",
		Name:             "Canyon Lurkers",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{3}{R}")},
	})
}
