package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Torment — Enchantment — Aura {1}{B} (Stronghold, #5627):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-0."
//
// The enchant ability is the Aura spell's target (CR 303.4a), and the
// shared pumpEnchantedCreature static gives the host -3/-0 in layer 7c
// while it is a creature. Power can go below zero; a creature with
// negative power deals no combat damage. It is declared a harmful Aura,
// as Cryoshatter is, so it is aimed at an opponent's creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b53ec6e6-fcc9-4471-88c5-7ad0fbd7bbea",
		Name:         "Torment",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{pumpEnchantedCreature(-3, 0)},
	})
}
