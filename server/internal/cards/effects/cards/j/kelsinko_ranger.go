package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kelsinko Ranger — Creature — Human Ranger {W}, 1/1 (Ice Age, S82 OLD #5636):
//
//	"{1}{W}: Target green creature gains first strike until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d47e598-690e-4850-97e4-7136bc7b497e",
		Name:         "Kelsinko Ranger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target green creature gains first strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target green creature", effects.OfColor("G")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("first strike", "Kelsinko Ranger — first strike"),
		}},
	})
}
