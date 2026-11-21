package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ojutai Interceptor — Creature — Bird Soldier {3}{U}, 3/1 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 Megamorph {3}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The printed keywords ride PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ecc42076-5cce-4e00-a21b-34846ea41df2",
		Name:             "Ojutai Interceptor",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{3}{U}")},
	})
}
