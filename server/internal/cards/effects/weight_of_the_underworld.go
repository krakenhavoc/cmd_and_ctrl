package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weight of the Underworld — Enchantment — Aura {3}{B} (Born of the Gods,
// #5559):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-2."
//
// Dead Weight at -3/-2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b960d17b-3679-4dc5-9a5d-c8e7fd81fa9a",
		Name:         "Weight of the Underworld",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{pumpEnchantedCreature(-3, -2)},
	})
}
