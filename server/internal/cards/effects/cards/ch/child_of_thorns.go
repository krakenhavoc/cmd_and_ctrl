package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Child of Thorns — Creature — Spirit {G}, 1/1 (Betrayers of Kamigawa, S82
// OLD, #5602):
//
//	"Sacrifice this creature: Target creature gets +1/+1 until end of
//	 turn."
//
// The sacrifice is a cost, so the Child is gone whatever happens to the
// target and is not a legal target in practice; a target that has left
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Child of Thorns — target creature gets +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "f94002a1-582f-4b9d-a4f7-dae816f21c1d",
		Name:         "Child of Thorns",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT(label, 1, 1),
		}},
	})
}
