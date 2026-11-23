package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ritual of Rejuvenation — Instant {2}{W} (Ixalan, #5532):
//
//	"You gain 4 life.
//	 Draw a card."
//
// Revitalize's body with 4 life: the draw waits for any prompt the
// life gain asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c9d8065-9ab0-4468-a11c-383bd036029f",
		Name:         "Ritual of Rejuvenation",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 4}, effects.DrawCards{N: 1}),
	})
}
