package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marang River Skeleton — Creature — Skeleton {1}{B}, 1/1 (Dragons of Tarkir,
// #5551):
//
//	"{B}: Regenerate this creature.
//	 Megamorph {3}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The regeneration is Drowned's ability: a shield on this creature
// (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0cfbbb26-9e1c-4b79-8b32-dcc05739f5ea",
		Name:             "Marang River Skeleton",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{3}{B}")},
		Activated:        []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}
