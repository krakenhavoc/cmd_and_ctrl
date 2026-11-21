package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Segmented Krotiq — Creature — Insect {5}{G}, 6/5 (Dragons of Tarkir,
// #5551):
//
//	"Megamorph {6}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "05221ba6-53fd-4f45-b8b4-493724f75963",
		Name:             "Segmented Krotiq",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{6}{G}")},
	})
}
