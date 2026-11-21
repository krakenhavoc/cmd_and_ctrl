package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrage Ogre — Creature — Ogre Warrior {3}{R}{R}, 3/3 (Scars of Mirrodin, #5577):
//
//	"{T}, Sacrifice an artifact: This creature deals 2 damage to any target."
//
// Orcish Mechanics' ability: the artifact is sacrificed as a cost
// (CR 602.2b), so it is gone before the target is dealt damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d556c71e-ce16-4229-880b-744790f93797",
		Name:         "Barrage Ogre",
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
