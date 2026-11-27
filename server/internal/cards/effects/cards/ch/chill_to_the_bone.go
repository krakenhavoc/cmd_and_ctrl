package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chill to the Bone — Instant {3}{B} (Coldsnap, #5597):
//
//	"Destroy target nonsnow creature."
//
// A destroy with a "nonsnow" target clause, read through the layers so a
// permanent that an ice counter made snow is not a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcc53728-3f6e-48cf-a042-393a2d8098fc",
		Name:         "Chill to the Bone",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nonsnow creature", effects.Not(effects.Snow())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
