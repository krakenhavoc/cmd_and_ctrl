package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geistflame — Instant {R} (Innistrad, #5572):
//
//	"Geistflame deals 1 damage to any target.
//	 Flashback {3}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Firebolt's body at 1 damage with the shared Flashback ({3}{R}).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2f01c3d9-e0fc-4cdf-b4db-daeb8bb24bc3",
		Name:             "Geistflame",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetAny(),
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 1)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		OnResolve:        effects.DamageToFirstTargetStep(1),
	})
}
