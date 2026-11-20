package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icatian Lieutenant — Creature — Human Soldier {W}{W}, 1/2 (Fallen
// Empires, #5637):
//
//	"{1}{W}: Target Soldier creature gets +1/+0 until end of turn."
//
// The target must be a Soldier as the ability is activated and again
// as it resolves (CR 608.2b); the Lieutenant is a Soldier, so it may
// pump itself. No {T} in the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51e9af95-745b-45b9-b870-7f768578fa4f",
		Name:         "Icatian Lieutenant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target Soldier creature gets +1/+0 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target Soldier creature", effects.HasSubtype("Soldier")),
			Effect:  effects.PumpAndGrantFirstTarget(1, 0, "Icatian Lieutenant — +1/+0 until end of turn"),
		}},
	})
}
