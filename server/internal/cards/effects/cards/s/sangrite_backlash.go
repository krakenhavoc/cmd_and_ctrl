package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sangrite Backlash — Enchantment — Aura {B/G}{R} (Alara Reborn, #5586):
//
//	"Enchant creature
//	 Enchanted creature gets +3/-3."
//
// Immolation's shape: one layer 7c modification of the enchanted
// creature; a creature with toughness 3 or less dies to the state-based
// check.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03220765-8a9d-4d4a-88d5-98df975ce13c",
		Name:         "Sangrite Backlash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(3, -3)},
	})
}
