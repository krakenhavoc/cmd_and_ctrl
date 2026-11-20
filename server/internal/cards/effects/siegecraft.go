package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a49a9418-72d1-4ab1-89d9-a1f65ae23d35",
		Name:         "Siegecraft",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(2, 4)},
	})
}
