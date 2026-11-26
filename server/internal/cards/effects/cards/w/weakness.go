package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weakness — Enchantment — Aura {B} (Magic 2010, #5585):
//
//	"Enchant creature
//	 Enchanted creature gets -2/-1."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f07a24c0-bf3c-4733-9473-c6be3b16950e",
		Name:         "Weakness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-2, -1)},
	})
}
