package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dirgur Nemesis — Creature — Serpent {5}{U}, 6/5 (Dragons of Tarkir,
// #5551):
//
//	"Defender
//	 Megamorph {6}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The printed keywords ride PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c08b444b-6617-4199-b9da-5d4e81cd7597",
		Name:             "Dirgur Nemesis",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"defender"},
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{6}{U}")},
	})
}
