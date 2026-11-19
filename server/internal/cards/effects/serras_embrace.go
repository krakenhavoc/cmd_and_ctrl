package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serra's Embrace — Enchantment — Aura {2}{W}{W} (Foundations Jumpstart):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying and vigilance."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d6ba936-4a15-4c40-aaa6-71605fb732d1",
		Name:         "Serra's Embrace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying"),
			GrantToAttached("vigilance"),
		},
	})
}
