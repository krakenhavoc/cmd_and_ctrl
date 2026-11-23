package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Fireslinger — Creature — Goblin Warrior {R}, 1/1 (Foundations
// Jumpstart):
//
//	"{T}: This creature deals 1 damage to target player or
//	 planeswalker."
//
// Goblin Sharpshooter's tap ping (sourceDealsOneToFirstTarget) with Boros
// Charm's "target player or planeswalker" clause. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "396564be-3747-49dd-98dc-afb81a375f18",
		Name:         "Goblin Fireslinger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target player or planeswalker.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsOneToFirstTarget,
		}},
	})
}
