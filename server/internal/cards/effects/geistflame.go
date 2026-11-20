package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "2f01c3d9-e0fc-4cdf-b4db-daeb8bb24bc3",
		Name:             "Geistflame",
		Completeness:     CompletenessFull,
		Targets:          TargetAny(),
		Purpose:          ForTargets(DamageToTarget(0, 1)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}")},
		OnResolve:        damageToFirstTarget(1),
	})
}
