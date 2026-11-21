package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harmattan Efreet — Creature — Efreet {2}{U}{U}, 2/2 (Mirage, #5633):
//
//	"Flying
//	 {1}{U}{U}: Target creature gains flying until end of turn."
//
// Flying is the printed keyword. The ability is Keen Glidemaster's at
// {1}{U}{U}: any creature, the Efreet's own included, at instant speed,
// through the shared ebTargetCreatureGainsUntilEOT body. A target that
// left or stopped being a creature in response gains nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84b1d4e1-ca8a-4eed-b5c4-364876938aee",
		Name:            "Harmattan Efreet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}{U}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{1}{U}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Harmattan Efreet — flying until end of turn"),
		}},
	})
}
