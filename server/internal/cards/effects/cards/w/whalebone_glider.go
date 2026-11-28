package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whalebone Glider — Artifact {2} (Ice Age, S82 OLD, #5636):
//
//	"{2}, {T}: Target creature with power 3 or less gains flying until end of
//	 turn."
//
// Elvish Herder's grant on a power-limited target, judged on announce and again
// as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a65fadeb-c9bc-409f-8e90-ff8e55485a4c",
		Name:         "Whalebone Glider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target creature with power 3 or less gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
			Effect:  effects.S82OldB726TargetGainsKeywordUntilEOT("Whalebone Glider", "flying"),
		}},
	})
}
