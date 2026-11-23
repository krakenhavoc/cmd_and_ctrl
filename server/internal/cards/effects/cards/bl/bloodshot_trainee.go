package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodshot Trainee — Creature — Goblin Warrior {3}{R}, 2/3 (Future Sight, #5594):
//
//	"{T}: This creature deals 4 damage to target creature. Activate only
//	 if this creature's power is 4 or greater."
//
// The condition reads the Trainee's power through the layers as the
// ability is announced (CR 602.5b), so a pump that ends later does not
// take back an activation already on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87ab1c79-79f5-48a9-af1a-7c876684c8d9",
		Name:         "Bloodshot Trainee",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: This creature deals 4 damage to target creature. Activate only if this creature's power is 4 or greater.",
			Purpose:   effects.ForTargets(effects.DamageToTarget(0, 4)),
			Cost:      effects.TapCost(),
			Condition: effects.FutB1SourcePowerAtLeast(4),
			Targets:   effects.TargetCreature("target creature"),
			Effect:    effects.SourceDealsDamageToEachLegalTarget(4),
		}},
	})
}
