package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Griptide — Instant {3}{U} (Dark Ascension, #5571):
//
//	"Put target creature on top of its owner's library."
//
// Time Ebb at instant speed: the tuck goes to the OWNER's library, whoever
// controls the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a6bbc45-6cb8-40ff-9f7c-7b2c5713dd78",
		Name:         "Griptide",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
