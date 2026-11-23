package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feast of the Unicorn — Enchantment — Aura {3}{B} (Homelands, #5635):
//
//	"Enchant creature
//	 Enchanted creature gets +4/+0."
//
// An Aura (CR 303.4) whose static is a layer 7c change (CR 613.4c) to
// the creature it enchants, for as long as it stays attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "274d89b8-1e59-4992-9299-dc793b7f6752",
		Name:         "Feast of the Unicorn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(4, 0)},
	})
}
