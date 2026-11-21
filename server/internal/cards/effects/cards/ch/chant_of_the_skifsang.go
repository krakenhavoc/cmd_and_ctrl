package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5c902588-b16e-4679-b89e-7b11c928061f",
		Name:         "Chant of the Skifsang",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(-13, 0)},
	})
}
