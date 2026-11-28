package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Robe of Mirrors — Enchantment — Aura {U} (Exodus, S82 OLD b739, #5626):
//
//	"Enchant creature
//	 Enchanted creature has shroud."
//
// Alexi's Cloak without flash: the shared attachment keyword grant (layer 6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "093a20e5-ff14-41c7-b16c-1f745ddf6942",
		Name:         "Robe of Mirrors",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("shroud")},
	})
}
