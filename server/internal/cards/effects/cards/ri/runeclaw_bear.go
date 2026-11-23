package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runeclaw Bear — Creature — Bear {1}{G}, 2/2 (Magic 2010, #5585):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec49dfcf-d16d-4621-af4b-4a6f09043221",
		Name:         "Runeclaw Bear",
		Completeness: effects.CompletenessFull,
	})
}
