package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Font of Ire — Enchantment {1}{R} (Journey into Nyx, #5558):
//
//	"{3}{R}, Sacrifice this enchantment: It deals 5 damage to target
//	 player or planeswalker."
//
// A sacrifice-as-cost ability, activated at instant speed. The
// Font is the damage's source, as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88d4c611-3211-41b2-8453-6210f8394a0c",
		Name:         "Font of Ire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, Sacrifice this enchantment: It deals 5 damage to target player or planeswalker.",
			Purpose: ForTargets(DamageToTarget(0, 5)),
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{3}{R}"},
			Targets: targetPlayerOrPlaneswalker(),
			Effect:  DealDamageToTheTarget(5),
		}},
	})
}
