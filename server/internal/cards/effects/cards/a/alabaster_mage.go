package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alabaster Mage — Creature — Human Wizard {1}{W}, 2/1 (Jumpstart):
//
//	"{1}{W}: Target creature you control gains lifelink until end of turn."
//
// No {T} in the cost, so it can be activated with summoning sickness and
// more than once. The target is re-checked on resolution (CR 608.2b): a
// creature that changed control in response gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea7a8acb-f01f-4a69-9b3c-954d8f0d3155",
		Name:         "Alabaster Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target creature you control gains lifelink until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.JmpB7FirstLegalTargetGains("Alabaster Mage — lifelink until end of turn", "lifelink"),
		}},
	})
}
