package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chaplain's Blessing — Sorcery {W} (Shadows over Innistrad, #5543):
//
//	"You gain 5 life."
//
// Sacred Nectar's life gain at five.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "195837b4-e5e4-479d-8dab-6ac77f9fcd83",
		Name:         "Chaplain's Blessing",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 5}),
	})
}
