package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traveling Philosopher — Creature — Human Advisor {1}{W}, 2/2 (Theros, #5561):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1adc4691-20cf-4f55-af85-3132fe1ff2a8",
		Name:         "Traveling Philosopher",
		Completeness: effects.CompletenessFull,
	})
}
