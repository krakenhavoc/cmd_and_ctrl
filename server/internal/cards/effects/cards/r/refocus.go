package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "0101f891-e54c-43e5-86f3-263d95060fcd",
		Name:         "Refocus",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.UntapTarget{Target: id}.Apply(ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
