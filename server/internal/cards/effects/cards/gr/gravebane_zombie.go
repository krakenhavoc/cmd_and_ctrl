package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravebane Zombie — Creature — Zombie {3}{B}, 3/2 (Mirage, S82 OLD batch 810,
// #5633):
//
//	"If this creature would die, put it on top of its owner's library instead."
//
// A self replacement (CR 614.6) of the move from the battlefield to a
// graveyard, redirected to the top of its owner's library. Only a death is
// replaced: a card that is milled or discarded still goes to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d04a836b-7e42-424d-9600-42a40d0dfccd",
		Name:         "Gravebane Zombie",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.S82OldB810DiesToTopOfLibraryInstead("Gravebane Zombie: put on top of its owner's library instead of dying"),
		},
	})
}
