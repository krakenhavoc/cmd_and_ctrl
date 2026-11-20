package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marsh Hulk — Creature — Zombie Ogre {4}{B}{B}, 4/6 (Dragons of Tarkir,
// #5551):
//
//	"Megamorph {6}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "979a581b-22a0-4461-8638-d219587323d7",
		Name:             "Marsh Hulk",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Megamorph("{6}{B}")},
	})
}
