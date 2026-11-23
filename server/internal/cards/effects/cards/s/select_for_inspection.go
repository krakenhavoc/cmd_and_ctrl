package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Select for Inspection — Instant {U} (Kaladesh, #5539):
//
//	"Return target tapped creature to its owner's hand. Scry 1. (Look
//	 at the top card of your library. You may put that card on the
//	 bottom.)"
//
// Harbinger of the Tides' tapped-creature target, then Voyage's End's
// scry after the bounce.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfee4ba9-3cf2-4e8b-a4a7-21974828cd54",
		Name:         "Select for Inspection",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetCreature("target tapped creature", effects.HobB3TappedCreature()),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.BounceToHand{Target: id}.Apply(ctx) }),
			effects.Scry{N: 1},
		),
	})
}
