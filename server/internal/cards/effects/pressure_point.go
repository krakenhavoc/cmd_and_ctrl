package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pressure Point — Instant {1}{W} (Fate Reforged,
// #5552):
//
//	"Tap target creature.
//	 Draw a card."
//
// Hithlain Knots without the scry: the tap is on the target if it is still
// legal as the spell resolves (CR 608.2b), and the draw happens either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4bbdd0c-e578-49ed-b2f7-b1decb04ee87",
		Name:         "Pressure Point",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return TapTarget{Target: id}.Apply(ctx) }),
			DrawCards{N: 1},
		),
	})
}
