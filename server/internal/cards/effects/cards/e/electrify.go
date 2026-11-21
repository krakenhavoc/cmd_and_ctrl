package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Electrify — Instant {3}{R}
// (Amonkhet, #5535):
//
//	"Electrify deals 4 damage to target creature."
//
// Flame Slash's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da952213-9a3c-4e13-bc5a-55d1e5dbaa5b",
		Name:         "Electrify",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
