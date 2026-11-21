package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trailblazing Historian — Creature — Human Shaman 1/3 {1}{R} (March of the Machine, S82):
//
//	"Haste
//	 {T}: Another target creature gains haste until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "399770b9-eab7-4512-aa1d-65c5bd81094f",
		Name:            "Trailblazing Historian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Another target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Trailblazing Historian — haste until end of turn"),
		}},
	})
}
