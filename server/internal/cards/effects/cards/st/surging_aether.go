package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surging Aether — Instant {3}{U} (Coldsnap, #5597):
//
//	"Ripple 4 (...)
//	 Return target permanent to its owner's hand."
//
// Temporal Fissure's bounce with ripple (CR 702.60a) as a trigger of the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "060b1a76-22a4-48a5-a353-4d0501367956",
		Name:         "Surging Aether",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target permanent"),
		Triggered:    []game.TriggeredAbility{effects.S82OldB566Ripple("Surging Aether", 4)},
		OnResolve:    effects.BounceTheTarget,
	})
}
