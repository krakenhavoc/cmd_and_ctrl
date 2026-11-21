package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repulsor Rays — Sorcery {R} (Marvel Super Heroes Commander, #4455):
// "Repulsor Rays deals 3 damage to target creature." Concentrated
// Fire's damage body at three damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "053a90b8-61b1-4390-afc0-c5eaef0138a9",
		Name:         "Repulsor Rays",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
