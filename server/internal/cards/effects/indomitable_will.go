package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indomitable Will — Enchantment — Aura {1}{W} (Foundations Jumpstart):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+2."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cca60afe-b044-4401-8322-170aa015873c",
		Name:            "Indomitable Will",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{PumpAttached(1, 2)},
	})
}
