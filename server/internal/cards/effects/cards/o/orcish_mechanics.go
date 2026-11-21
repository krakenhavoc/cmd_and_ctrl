package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Mechanics — Creature — Orc {2}{R}, 1/1 (Antiquities, #5640):
//
//	"{T}, Sacrifice an artifact: This creature deals 2 damage to any
//	 target."
//
// Pia and Kiran Nalaar's sacrifice ping with a {T} in place of the mana:
// the artifact is paid at announce, the target is any creature, player,
// planeswalker or battle, and a target that left in response leaves the
// ability to do nothing (DealDamageToTheTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ec58835-de2d-4064-89a9-f92db80bc276",
		Name:         "Orcish Mechanics",
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
