package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noble Steeds — Enchantment {2}{W} (Alliances, #5634):
//
//	"{1}{W}: Target creature gains first strike until end of turn."
//
// A layer 6 grant (CR 613.1f) to the target that ends at cleanup
// (CR 514.2). A target that has left by the time it resolves gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6005e47-d545-4c16-b54a-bfe1dde61a98",
		Name:         "Noble Steeds",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target creature gains first strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("first strike", "Noble Steeds — first strike until end of turn"),
		}},
	})
}
