package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Refocus — Instant {1}{U} (Fate Reforged,
// #5552):
//
//	"Untap target creature.
//	 Draw a card."
//
// Pressure Point's shape with an untap: the untap is on the target if it is
// still legal as the spell resolves (CR 608.2b), and the draw happens either
// way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0101f891-e54c-43e5-86f3-263d95060fcd",
		Name:         "Refocus",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return UntapTarget{Target: id}.Apply(ctx) }),
			DrawCards{N: 1},
		),
	})
}
