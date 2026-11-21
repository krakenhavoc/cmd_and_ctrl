package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Assault — Enchantment {2}{B}{R} (Invasion, #5616):
//
//	"{1}, Pay 2 life: This enchantment deals 1 damage to any target."
//
// Seal of Fire's damage body (sourceDealsDamageToEachLegalTarget) behind
// a mana-and-life cost. The life is part of the cost, paid as the
// ability is activated, so a player with less than 2 life can't
// activate it. There is no {T}, so it can be activated any number of
// times. A target gone by resolution takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "320820e4-3c7e-4e4d-b23c-63bf13a68e73",
		Name:         "Reckless Assault",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Pay 2 life: This enchantment deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.PayLife(2)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
