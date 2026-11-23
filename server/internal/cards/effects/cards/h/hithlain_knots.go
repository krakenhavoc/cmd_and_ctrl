package cardsh

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hithlain Knots — Instant {1}{U} (The Hobbit Eternal, HOC):
//
//	"Tap target creature. Scry 1.
//	 Draw a card."
//
// The two sentences are one spell resolving in order (CR 608.2c): the tap,
// the scry, then the draw, which waits for the scry's answer (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c80fc55-86a0-403e-82ce-09ee03095d8a",
		Name:         "Hithlain Knots",
		Purpose:      game.Purpose{Scry: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.TapTarget{Target: id}.Apply(ctx) }),
			effects.Scry{N: 1},
			effects.DrawCards{N: 1},
		),
	})
}
