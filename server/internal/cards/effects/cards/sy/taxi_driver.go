package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taxi Driver — Creature — Human Pilot {1}{R}, 3/1 (Marvel Super Heroes
// Commander, #4455):
//
//	"{1}, {T}: Target creature gains haste until end of turn."
//
// Hanweir Battlements' haste ability, paid with {1} instead of {R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a17a92f7-e95c-4d0e-a43e-7e6b41a59139",
		Name:         "Taxi Driver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Taxi Driver — haste until end of turn"),
		}},
	})
}
