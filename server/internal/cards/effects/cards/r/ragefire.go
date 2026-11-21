package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ragefire — Sorcery {1}{R} (Ravnica Allegiance, #5521):
//
//	"Ragefire deals 3 damage to target creature."
//
// Repulsor Rays's sentence: 3 damage to the creature
// (damageToFirstTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91b8c0bc-2599-4423-a23f-5dc95b23cf25",
		Name:         "Ragefire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
