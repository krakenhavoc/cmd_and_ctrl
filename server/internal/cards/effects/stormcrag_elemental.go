package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormcrag Elemental — Creature — Elemental {5}{R}, 5/5 (Dragons of Tarkir,
// #5551):
//
//	"Trample
//	 Megamorph {4}{R}{R} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The printed keywords ride PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f28c18a1-13f8-4360-adc7-9d828ab1e933",
		Name:             "Stormcrag Elemental",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{Megamorph("{4}{R}{R}")},
	})
}
