package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pensive Minotaur — Creature — Minotaur Warrior {2}{R}, 2/3 (Journey into Nyx, #5558).
//
// A vanilla creature: the catalog entry records that nothing is missing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cafad9c8-9974-4596-92d4-b7fb615a0353",
		Name:         "Pensive Minotaur",
		Completeness: effects.CompletenessFull,
	})
}
