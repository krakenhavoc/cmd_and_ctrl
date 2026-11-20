package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin War Paint — Enchantment — Aura {1}{R} (Zendikar, #5583):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has haste."
//
// Enchant creature is the shared EnchantCreature clause; the pump and
// the keyword are the shared attachment statics.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67ca79d7-9064-4605-8625-b1cfe5cb1b45",
		Name:         "Goblin War Paint",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("haste"),
		},
	})
}
