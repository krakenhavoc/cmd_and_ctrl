package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acrobatic Maneuver — Instant {2}{W} (Kaladesh, #5539):
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control.
//	 Draw a card."
//
// Blur's flicker under the owner's control (CR 400.7: it returns as a new
// object), then the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1aa1cf50-ee79-4bef-826b-5f1a9e1aa1a5",
		Name:         "Acrobatic Maneuver",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return Flicker{Target: id}.Apply(ctx) }),
			DrawCards{N: 1},
		),
	})
}
