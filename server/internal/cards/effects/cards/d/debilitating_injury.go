package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Debilitating Injury — Enchantment — Aura {1}{B} (Khans of Tarkir,
// #5554):
//
//	"Enchant creature
//	 Enchanted creature gets -2/-2."
//
// Stab Wound's -2/-2 static on the enchanted creature (layer 7c),
// without the upkeep drain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52eab77d-9a07-4e14-8872-72681d3b3d0e",
		Name:         "Debilitating Injury",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(-2, -2)},
	})
}
