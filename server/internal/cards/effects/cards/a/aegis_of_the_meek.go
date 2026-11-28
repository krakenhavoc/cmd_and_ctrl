package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aegis of the Meek — Artifact {3} (Ice Age, S82 OLD, #5636):
//
//	"{1}, {T}: Target 1/1 creature gets +1/+2 until end of turn."
//
// The target must be 1/1 as it is chosen and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0cf78877-bf08-413c-8ecb-8c3265ac1c4a",
		Name:         "Aegis of the Meek",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Target 1/1 creature gets +1/+2 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 2)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target 1/1 creature", effects.OldB834PowerToughness(1, 1)),
			Effect:  effects.OldB834TargetGets(1, 2, "Aegis of the Meek — +1/+2"),
		}},
	})
}
