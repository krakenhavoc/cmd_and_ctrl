package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incurable Ogre — {3}{R} Creature — Ogre Mutant 5/1 (Shards of Alara, S82 OLD b513):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94190aaf-db91-4a3c-a2a7-e90868e856ff",
		Name:         "Incurable Ogre",
		Completeness: effects.CompletenessFull,
	})
}
