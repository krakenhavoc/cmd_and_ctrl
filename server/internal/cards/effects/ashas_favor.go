package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Asha's Favor — Enchantment — Aura {2}{W} (Conflux, #5587):
//
//	"Enchant creature
//	 Enchanted creature has flying, first strike, and vigilance."
//
// Battle Mastery's shape with three keywords: one layer-6 grant to the
// enchanted creature (CR 613.1f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97e7ada6-dee1-4bb3-a42d-38a61d946bd4",
		Name:         "Asha's Favor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			GrantToAttached("flying", "first strike", "vigilance"),
		},
	})
}
