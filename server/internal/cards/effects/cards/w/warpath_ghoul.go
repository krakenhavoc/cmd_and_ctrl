package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warpath Ghoul — Creature — Zombie {2}{B}, 3/2 (Magic 2010, #5585):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17ae652c-b1c7-4dde-ae9e-50b9e286776c",
		Name:         "Warpath Ghoul",
		Completeness: effects.CompletenessFull,
	})
}
