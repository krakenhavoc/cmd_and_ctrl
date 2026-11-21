package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siegecraft — Enchantment — Aura {3}{W} (Khans of Tarkir, #5554):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+4."
//
// The shared PumpAttached static at +2/+4 on the enchanted creature
// (layer 7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a49a9418-72d1-4ab1-89d9-a1f65ae23d35",
		Name:         "Siegecraft",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(2, 4)},
	})
}
