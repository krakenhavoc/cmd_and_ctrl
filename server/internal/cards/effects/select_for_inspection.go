package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "cfee4ba9-3cf2-4e8b-a4a7-21974828cd54",
		Name:         "Select for Inspection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      TargetCreature("target tapped creature", hobB3TappedCreature()),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return BounceToHand{Target: id}.Apply(ctx) }),
			Scry{N: 1},
		),
	})
}
