package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sensory Deprivation — Enchantment — Aura {U} (Innistrad, #5572):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-0."
//
// Torment's -3/-0 on the enchanted creature (pumpEnchantedCreature, a
// layer-7c modify).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0006faf6-7a61-426c-9034-579f2cfcfa83",
		Name:         "Sensory Deprivation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(-3, 0)},
	})
}
