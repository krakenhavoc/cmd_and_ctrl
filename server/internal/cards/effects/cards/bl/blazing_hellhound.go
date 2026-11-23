package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blazing Hellhound — Creature — Elemental Dog {2}{B}{R}, 4/3 (Magic Origins, #5549):
//
//	"{1}, Sacrifice another creature: This creature deals 1 damage to any target."
//
// Hurler Cyclops's ability: the sacrifice is another creature, paid as a
// cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "598a7410-35e8-4d62-9bf1-d44ee7717040",
		Name:         "Blazing Hellhound",
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
