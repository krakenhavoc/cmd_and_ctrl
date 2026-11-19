package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slimebind — Enchantment — Aura {1}{U} (Foundations Jumpstart):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -4/-0."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "87aa2cbe-841c-40ff-9b1c-b259b4cc631e",
		Name:            "Slimebind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static:          []game.StaticAbility{PumpAttached(-4, 0)},
	})
}
