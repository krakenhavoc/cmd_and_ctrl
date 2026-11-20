package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Sunstriker — Creature — Bird Warrior {1}{W}{W}, 1/1 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 Double strike (This creature deals both first-strike and regular combat damage.)
//	 Megamorph {4}{W} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The printed keywords ride PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "60b0fd7c-a4cc-4a41-8455-debcfeccff81",
		Name:             "Aven Sunstriker",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "double strike"},
		AlternativeCosts: []game.AlternativeCost{Megamorph("{4}{W}")},
	})
}
