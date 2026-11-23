package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Artillery — Creature — Goblin Warrior {1}{R}{R}, 1/3 (Magic 2010, #5585):
//
//	"{T}: This creature deals 2 damage to any target and 3 damage to you."
//
// A tap ability with Spicy Oatmeal Pizza's body: 2 damage to the target and
// 3 to its controller, dealt at once. If the target is gone, the ability
// does nothing, the damage to you included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06db7444-0225-48c8-ab04-d69ff06e97b6",
		Name:         "Goblin Artillery",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to any target and 3 damage to you.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.TmtB6DamageAnyTargetAndYou(2, 3),
		}},
	})
}
