package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0006faf6-7a61-426c-9034-579f2cfcfa83",
		Name:         "Sensory Deprivation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{pumpEnchantedCreature(-3, 0)},
	})
}
