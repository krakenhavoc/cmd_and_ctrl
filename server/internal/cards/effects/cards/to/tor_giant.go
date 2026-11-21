package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tor Giant — Creature — Giant {3}{R}, 3/3 (Ice Age, #5636):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2457a9f7-8a46-4c7a-b35e-bcbf4a79c365",
		Name:         "Tor Giant",
		Completeness: effects.CompletenessFull,
	})
}
