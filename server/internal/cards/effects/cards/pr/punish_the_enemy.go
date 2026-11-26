package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Punish the Enemy — Instant {4}{R} (Dragon's Maze, S82 OLD b405, #5564):
//
//	"Punish the Enemy deals 3 damage to target player or planeswalker and 3
//	 damage to target creature."
//
// Hungry Flames' shape with the amounts swapped: two clauses, one damage
// instance each, and a clause whose target left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09f73f77-a9bc-4718-be7d-ba6e08576e3a",
		Name:         "Punish the Enemy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Clauses(effects.TargetPlayerOrPlaneswalker(), effects.TargetCreature("target creature")),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3), effects.DamageToTarget(1, 3)),
		OnResolve:    effects.DamageEachClauseTarget(3, 3),
	})
}
