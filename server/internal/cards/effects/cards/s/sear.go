package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sear — Instant {1}{R} (Lorwyn Eclipsed, #4458):
//
//	"Sear deals 4 damage to target creature or planeswalker."
//
// Molten Exhale's body without the behold: 4 damage to one target
// creature or planeswalker (CR 120.3c, 120.3e). If the target is illegal
// as Sear resolves, Sear does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23b11fdc-61b5-4f5e-ad80-f81b729fc91d",
		Name:         "Sear",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
