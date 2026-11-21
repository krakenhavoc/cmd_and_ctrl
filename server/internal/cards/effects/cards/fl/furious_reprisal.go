package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furious Reprisal — Sorcery {3}{R} (Foundations Jumpstart):
//
//	"Furious Reprisal deals 2 damage to each of two targets."
//
// Volcanic Salvo's damage to each target, over exactly two distinct
// "any target" picks. A target that became illegal takes nothing and the
// other still does (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38e1ab26-95f9-43dd-9bda-13c7d8ba47ea",
		Name:         "Furious Reprisal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(2, 2),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve:    effects.DealDamageToEachLegalTarget(2),
	})
}
