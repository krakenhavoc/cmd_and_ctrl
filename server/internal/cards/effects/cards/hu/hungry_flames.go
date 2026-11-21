package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hungry Flames — Instant {2}{R} (Aether Revolt,
// #5536):
//
//	"Hungry Flames deals 3 damage to target creature and 2 damage to target
//	 player or planeswalker."
//
// Two clauses, Arc Trail's shape (damageEachClauseTarget): one damage
// instance, and a clause whose target left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2c00788-14e7-4ab5-8c71-2a176621a177",
		Name:         "Hungry Flames",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Clauses(effects.TargetCreature("target creature"), effects.TargetPlayerOrPlaneswalker()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3), effects.DamageToTarget(1, 2)),
		OnResolve:    effects.DamageEachClauseTarget(3, 2),
	})
}
