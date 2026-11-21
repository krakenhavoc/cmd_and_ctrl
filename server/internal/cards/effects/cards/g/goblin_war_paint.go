package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "67ca79d7-9064-4605-8625-b1cfe5cb1b45",
		Name:         "Goblin War Paint",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("haste"),
		},
	})
}
