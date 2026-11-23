package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meteor Storm — Enchantment {R}{G}:
//
//	"{2}{R}{G}, Discard two cards at random: This enchantment deals 4 damage
//	 to any target."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// Two cards, drawn together: a hand of one can't pay it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03f96c23-de0b-4b85-81f8-febc850aa621",
		Name:         "Meteor Storm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}{G}, Discard two cards at random: This enchantment deals 4 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}{G}"), effects.DiscardAtRandom(2, "two cards at random")),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(4),
		}},
	})
}
