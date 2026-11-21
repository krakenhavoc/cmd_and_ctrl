package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Concentrated Fire — Instant {3}{R} (Marvel Super Heroes Commander, #4455):
// "Concentrated Fire deals 5 damage to target creature." Lightning Axe's
// damage body without the additional cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fcb2342-202d-4baa-82ad-66d782142264",
		Name:         "Concentrated Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
