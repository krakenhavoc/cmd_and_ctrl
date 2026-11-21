package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gutter Skulk — Creature — Zombie Rat {1}{B}, 2/2 (Gatecrash, #5565):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e775053-ab50-4cf7-8eef-e19a6eb79e1b",
		Name:         "Gutter Skulk",
		Completeness: effects.CompletenessFull,
	})
}
