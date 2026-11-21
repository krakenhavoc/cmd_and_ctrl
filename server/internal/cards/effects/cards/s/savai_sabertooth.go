package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savai Sabertooth — Creature — Cat {1}{W}, 3/1 (Ikoria: Lair of Behemoths, #5512):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24f2cf61-22db-4e34-898f-6baa2957d3ae",
		Name:         "Savai Sabertooth",
		Completeness: effects.CompletenessFull,
	})
}
