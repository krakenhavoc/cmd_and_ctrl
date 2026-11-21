package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keen Glidemaster — Creature — Human Soldier {1}{U}, 2/1 (Jumpstart
// Azorius deck request, #4894):
//
//	"{2}{U}: Target creature gains flying until end of turn."
//
// An instant-speed activated grant on any creature, its own included.
// A target that left or stopped being a creature in response does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98deac9e-28ea-4d11-8d75-833bf6d66e6e",
		Name:         "Keen Glidemaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Keen Glidemaster — flying until end of turn"),
		}},
	})
}
