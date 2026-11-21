package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prizefighter Construct — Artifact Creature — Construct {5}, 6/2 (Aether Revolt, #5536):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see bear_cub.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f4cfd69-c739-4398-93a5-e6238b7cd83b",
		Name:         "Prizefighter Construct",
		Completeness: effects.CompletenessFull,
	})
}
