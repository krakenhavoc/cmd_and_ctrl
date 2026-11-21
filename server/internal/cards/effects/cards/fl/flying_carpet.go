package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flying Carpet — Artifact {4} (Arabian Nights, #5641):
//
//	"{2}, {T}: Target creature gains flying until end of turn."
//
// Keen Glidemaster's grant from an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "453524fb-2ffa-421f-883c-b77ae2850c52",
		Name:         "Flying Carpet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Flying Carpet — flying until end of turn"),
		}},
	})
}
