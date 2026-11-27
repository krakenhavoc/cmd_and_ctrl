package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swelter — Sorcery {3}{R} (Judgment, #5611):
//
//	"Swelter deals 2 damage to each of two target creatures."
//
// Furious Reprisal's shape: exactly two targets (a clause of count 2), each
// dealt 2 damage by the spell; a target that has become illegal is skipped (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1169108-1bc8-42ca-a056-a9358fba10cb",
		Name:         "Swelter",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve:    effects.DealDamageToEachLegalTarget(2),
	})
}
