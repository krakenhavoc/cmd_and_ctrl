package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furious Strength — Enchantment — Aura {2}{R} (Marvel Super Heroes
// Commander, #4455): "Enchant creature. Enchanted creature gets +2/+2 and
// has menace."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efefb339-bb7c-4b26-abed-8434f955122b",
		Name:         "Furious Strength",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(2, 2), GrantToAttached("menace")},
	})
}
