package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aggressive Urge — Instant {1}{G} (Jumpstart, #5510):
//
//	"Target creature gets +1/+1 until end of turn.
//	 Draw a card."
//
// The draw is part of the same resolution: a spell whose only target has
// become illegal does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43f5a93b-0f8d-48d2-ab9d-275d44cf88b5",
		Name:         "Aggressive Urge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 1))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Aggressive Urge — +1/+1"}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
