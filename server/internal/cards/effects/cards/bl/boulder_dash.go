package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boulder Dash — Sorcery {1}{R} (Lorwyn Eclipsed, #4458):
//
//	"Boulder Dash deals 2 damage to any target and 1 damage to any other
//	 target."
//
// Arc Trail's text and shape, and its body (damageEachClauseTarget): two
// clauses, slot 0 takes 2 and slot 1 takes 1, and slot 1 is Distinct
// ("any other target"). Each slot is re-checked on its own at resolution
// (CR 608.2b), so if one target has left the other is still dealt its
// damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30a9ad0e-1b23-41ce-bacc-3b27f78ef628",
		Name:         "Boulder Dash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Clauses(effects.TargetAny(), effects.Distinct(effects.TargetAny())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2), effects.DamageToTarget(1, 1)),
		OnResolve:    effects.DamageEachClauseTarget(2, 1),
	})
}
