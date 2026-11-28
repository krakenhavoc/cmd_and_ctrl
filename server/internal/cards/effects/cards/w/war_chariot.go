package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Chariot — Artifact {3} (Ice Age, S82 OLD #5636):
//
//	"{3}, {T}: Target creature gains trample until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "989c698a-600e-47d8-acaf-3ca140dcd150",
		Name:         "War Chariot",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature gains trample until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("trample", "War Chariot — trample"),
		}},
	})
}
