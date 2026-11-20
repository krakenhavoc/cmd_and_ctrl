package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mire's Grasp — Enchantment — Aura {1}{B} (Theros Beyond Death, #5514):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-3."
//
// Dead Weight's shape, -3/-3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1009d2a2-a145-46e1-ba14-7a42735062fc",
		Name:         "Mire's Grasp",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{pumpEnchantedCreature(-3, -3)},
	})
}
