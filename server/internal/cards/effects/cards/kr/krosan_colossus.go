package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Colossus — Creature — Beast {6}{G}{G}{G}, 9/9:
//
//	"Morph {6}{G}{G}"
//
// Morph alone: the face-down cast and the turn-up are the keyword's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "703ab192-7e0a-4f02-8381-568ad2d999ca",
		Name:             "Krosan Colossus",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{6}{G}{G}")},
	})
}
