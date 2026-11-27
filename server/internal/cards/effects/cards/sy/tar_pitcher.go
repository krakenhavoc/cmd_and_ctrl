package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tar Pitcher — Creature — Goblin Shaman {3}{R}, 2/2 (Lorwyn):
//
//	"{T}, Sacrifice a Goblin: This creature deals 2 damage to any target."
//
// Barrage Ogre's ability over Goblins. The Goblin is a cost (CR 602.2b): any
// Goblin you control, the Pitcher itself included, since it is one; the damage
// is dealt as the Pitcher last existed if it was the sacrifice (CR 608.2h).
// Tapping it needs it free of summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6991ab8-36f7-40c4-ad37-06e847d92664",
		Name:         "Tar Pitcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a Goblin: This creature deals 2 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a Goblin", effects.Subtype("Goblin"))),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
