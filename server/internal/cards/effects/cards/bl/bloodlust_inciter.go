package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodlust Inciter — Creature — Human Warrior {R}, 1/1 (Foundations
// Jumpstart):
//
//	"{T}: Target creature gains haste until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "826b83ba-0fe5-4f1d-91c8-799a1100d1e6",
		Name:         "Bloodlust Inciter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Bloodlust Inciter — haste until end of turn"),
		}},
	})
}
