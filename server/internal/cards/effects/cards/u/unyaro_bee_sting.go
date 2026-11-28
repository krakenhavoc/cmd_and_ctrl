package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unyaro Bee Sting — Sorcery {3}{G} (Mirage, #5633):
//
//	"Unyaro Bee Sting deals 2 damage to any target."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a500313b-35e3-4ebc-9144-e9486784757b",
		Name:         "Unyaro Bee Sting",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
