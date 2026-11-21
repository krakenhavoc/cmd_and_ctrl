package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Air Marshal — Creature — Human Soldier {1}{U}, 2/1 (The Brothers' War,
// #5482):
//
//	"{3}: Target Soldier gains flying until end of turn."
//
// Harmattan Efreet's grant over a Soldier target, the Marshal itself
// included. A target that left or stopped being a Soldier in response
// gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59b2edf0-00e4-4cd4-88ac-6dd74e9b1b14",
		Name:         "Air Marshal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Target Soldier gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCreature("target Soldier", effects.OfCreatureType("Soldier")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Air Marshal — flying until end of turn"),
		}},
	})
}
