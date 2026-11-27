package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jagged Lightning — {3}{R}{R} Sorcery (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Jagged Lightning deals 3 damage to each of two target creatures."
//
// Exactly two targets; each one still legal as it resolves takes 3 damage
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5d5cfc0-17e0-42d1-b5bd-9f27b11191f9",
		Name:         "Jagged Lightning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DealDamageToEachLegalTarget(3),
	})
}
