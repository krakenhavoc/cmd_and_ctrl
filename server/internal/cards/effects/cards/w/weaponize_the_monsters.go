package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weaponize the Monsters — Enchantment {R} (Ikoria: Lair of Behemoths, #5512):
//
//	"{2}, Sacrifice a creature: This enchantment deals 2 damage to any target."
//
// The creature is sacrificed as a cost, so it is gone before the ability
// resolves and may be the very creature the ability then hits only if it was
// chosen as a target (a sacrificed target is illegal as the ability resolves,
// CR 608.2b). The damage comes from the enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa2154fb-6756-4cf8-85a6-68264fb7a0b2",
		Name:         "Weaponize the Monsters",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a creature: This enchantment deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
