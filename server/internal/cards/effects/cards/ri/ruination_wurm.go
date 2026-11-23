package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruination Wurm — Creature — Wurm {4}{R}{G}, 7/6 (Gatecrash, #5565):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d084fe0-4896-4a1d-839c-81f55591c18a",
		Name:         "Ruination Wurm",
		Completeness: effects.CompletenessFull,
	})
}
