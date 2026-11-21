package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mire's Grasp — Enchantment — Aura {1}{B} (Theros Beyond Death, #5514):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-3."
//
// Dead Weight's shape, -3/-3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1009d2a2-a145-46e1-ba14-7a42735062fc",
		Name:         "Mire's Grasp",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(-3, -3)},
	})
}
