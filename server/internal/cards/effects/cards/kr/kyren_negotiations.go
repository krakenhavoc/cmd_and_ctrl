package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kyren Negotiations — Enchantment {2}{R}{R} (Mercadian Masques, S82 OLD,
// #5619):
//
//	"Tap an untapped creature you control: This enchantment deals 1 damage to
//	 target player or planeswalker."
//
// Tapping a creature is not the {T} symbol (CR 302.6), so a creature that
// arrived this turn may pay it (Purple Pentapus's cost). The damage is the
// enchantment's own, to a player or a planeswalker, and the ability can be
// activated as often as there are untapped creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50090188-651d-4402-968f-3f6bf9276029",
		Name:         "Kyren Negotiations",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: This enchantment deals 1 damage to target player or planeswalker.",
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
