package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "d4bbdd0c-e578-49ed-b2f7-b1decb04ee87",
		Name:         "Pressure Point",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.TapTarget{Target: id}.Apply(ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
