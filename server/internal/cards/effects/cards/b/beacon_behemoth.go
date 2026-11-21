package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beacon Behemoth — Creature — Beast {3}{G}{G}, 5/3 (Conflux, #5587):
//
//	"{1}: Target creature with power 5 or greater gains vigilance until
//	 end of turn."
//
// Spearbreaker Behemoth's row with vigilance. The power is the target's
// current power, checked as the ability is activated and again as it
// resolves (CR 608.2b); it can target the Behemoth itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25571850-d5c6-460f-b212-9fc9aa705277",
		Name:         "Beacon Behemoth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Target creature with power 5 or greater gains vigilance until end of turn.",
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature with power 5 or greater", effects.PowerGE(5)),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("vigilance", "Beacon Behemoth — vigilance"),
		}},
	})
}
