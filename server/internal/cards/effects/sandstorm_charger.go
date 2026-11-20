package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandstorm Charger — Creature — Beast {4}{W}, 3/4 (Dragons of Tarkir,
// #5551):
//
//	"Megamorph {4}{W} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "aa7a9689-e79a-4562-a621-eadc45966ce8",
		Name:             "Sandstorm Charger",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Megamorph("{4}{W}")},
	})
}
