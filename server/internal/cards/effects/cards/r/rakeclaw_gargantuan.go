package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakeclaw Gargantuan — Creature — Beast {2}{R}{G}{W}, 5/3 (Commander 2013, S82
// OLD b385):
//
//	"{1}: Target creature with power 5 or greater gains first strike until end
//	 of turn."
//
// Beacon Behemoth's row with first strike. The power is the target's current
// power, checked as the ability is activated and again as it resolves
// (CR 608.2b); it can target the Gargantuan itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8dbb4a8f-78e9-4ceb-824d-bb67bdf939db",
		Name:         "Rakeclaw Gargantuan",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Target creature with power 5 or greater gains first strike until end of turn.",
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature with power 5 or greater", effects.PowerGE(5)),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("first strike", "Rakeclaw Gargantuan — first strike"),
		}},
	})
}
