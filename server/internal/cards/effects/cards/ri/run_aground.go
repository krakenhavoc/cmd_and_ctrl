package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Run Aground — Instant {3}{U} (Ixalan, S82 older sets):
//
//	"Put target artifact or creature on top of its owner's library."
//
// Time Ebb's body over an artifact or creature target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ef9bf18-6b09-4398-a502-6867d999ab0c",
		Name:         "Run Aground",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
