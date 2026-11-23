package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bronze Sable — Artifact Creature — Sable {2}, 2/1 (Theros, #5561):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f27e5e11-0ad0-448b-8760-75ed1b97e7d8",
		Name:         "Bronze Sable",
		Completeness: effects.CompletenessFull,
	})
}
