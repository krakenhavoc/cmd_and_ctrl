package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hurler Cyclops — Creature — Cyclops {3}{R}{R}, 5/4 (Dominaria United):
//
//	"{1}, Sacrifice another creature: This creature deals 1 damage to any
//	 target."
//
// The sacrifice is a cost, so the Cyclops itself is not a legal payment
// (SacrificeAnotherN). The damage is dealt by the Cyclops: if it has left
// by resolution, its last-known information is used (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fac57730-8fd8-4702-8b35-4ef0ed5fa41a",
		Name:         "Hurler Cyclops",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
