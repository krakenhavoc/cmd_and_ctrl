package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mage il-Vec — Creature — Human Wizard {2}{R}, 2/2:
//
//	"{T}, Discard a card at random: This creature deals 1 damage to any
//	 target."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adea7db6-634c-4c5d-be40-264b4acffc53",
		Name:         "Mage il-Vec",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card at random: This creature deals 1 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardAtRandom(1, "a card at random")),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
