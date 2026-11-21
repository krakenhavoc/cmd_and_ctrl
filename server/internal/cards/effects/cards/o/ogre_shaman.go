package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Shaman — Creature — Ogre Shaman {3}{R}{R}, 3/3:
//
//	"{2}, Discard a card at random: This creature deals 2 damage to any
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
		OracleID:     "04c7bb20-6e40-4e79-a0ec-ced920d3491e",
		Name:         "Ogre Shaman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Discard a card at random: This creature deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.DiscardAtRandom(1, "a card at random")),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
