package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Vandal — Creature — Orc Warrior {1}{R}, 1/1 (Dominaria, #5530):
//
//	"{T}, Sacrifice an artifact: This creature deals 2 damage to any target."
//
// Orcish Mechanics' ability: tap it and sacrifice an artifact for 2
// damage to any target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "933a1811-7e1a-4173-b1c9-34b2c858ba68",
		Name:         "Orcish Vandal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: This creature deals 2 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
