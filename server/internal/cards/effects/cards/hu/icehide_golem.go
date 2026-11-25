package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icehide Golem — Snow Artifact Creature — Golem {S}, 2/2 (Modern Horizons):
//
//	"({S} can be paid with one mana from a snow source.)"
//
// A vanilla 2/2 whose whole cost is the snow symbol: the engine asks for one
// mana of a snow source (CR 107.4h) and refuses any other mana. The snow
// supertype and the artifact creature types come from the printed type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29b925e9-1e98-471c-9b23-e0372ac86485",
		Name:         "Icehide Golem",
		Completeness: effects.CompletenessFull,
	})
}
