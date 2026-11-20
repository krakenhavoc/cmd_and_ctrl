package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icatian Priest — Creature — Human Cleric {W}, 1/1 (Fallen Empires,
// #5637):
//
//	"{1}{W}{W}: Target creature gets +1/+1 until end of turn."
//
// Any creature, the Priest itself included. No {T} in the cost, so it
// can be activated the turn the Priest arrives, and as often as it is
// paid for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a9d29ce-e8e2-402f-b0c2-873f70418234",
		Name:         "Icatian Priest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{W}: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.ManaCost("{1}{W}{W}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PumpAndGrantFirstTarget(1, 1, "Icatian Priest — +1/+1 until end of turn"),
		}},
	})
}
