package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pin to the Earth — Enchantment — Aura {1}{U} (Journey into Nyx):
//
//	"Enchant creature
//	 Enchanted creature gets -6/-0."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "605624ee-b7ca-4728-993f-e13b18cc57f3",
		Name:         "Pin to the Earth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-6, 0)},
	})
}
