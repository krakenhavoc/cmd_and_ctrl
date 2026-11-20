package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chant of the Skifsang — Enchantment — Aura {2}{U} (Dark Ascension, #5571):
//
//	"Enchant creature
//	 Enchanted creature gets -13/-0."
//
// Torment's layer 7c static with -13/-0. A power below zero deals no
// damage (CR 120.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c902588-b16e-4679-b89e-7b11c928061f",
		Name:         "Chant of the Skifsang",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{pumpEnchantedCreature(-13, 0)},
	})
}
