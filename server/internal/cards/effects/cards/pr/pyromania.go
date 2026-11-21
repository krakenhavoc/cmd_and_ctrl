package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyromania — Enchantment {2}{R}:
//
//	"{1}{R}, Discard a card at random: This enchantment deals 1 damage to
//	 any target.
//	 {1}{R}, Sacrifice this enchantment: It deals 1 damage to any target."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// The sacrifice ability's damage comes from the enchantment as it last
// existed (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c89cf19-8dc6-422e-bbf8-4dd5a399c630",
		Name:         "Pyromania",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}, Discard a card at random: This enchantment deals 1 damage to any target.",
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardAtRandom(1, "a card at random")),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			},
			{
				Label:   "{1}{R}, Sacrifice this enchantment: It deals 1 damage to any target.",
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			},
		},
	})
}
