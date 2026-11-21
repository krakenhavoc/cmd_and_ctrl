package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephyr Charge — Enchantment {1}{U} (Magic 2014, #5563):
//
//	"{1}{U}: Target creature gains flying until end of turn."
//
// Harmattan Efreet's grant at {1}{U}, from an enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "576128ba-28aa-4ef9-9e72-8ca13f42c9b4",
		Name:         "Zephyr Charge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Zephyr Charge — flying until end of turn"),
		}},
	})
}
