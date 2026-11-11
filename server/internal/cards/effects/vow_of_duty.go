package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vow of Duty — Enchantment — Aura {2}{W} (Commander wave 2, #5368):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has vigilance, and can't attack you or
//	 planeswalkers you control."
//
// Three statics on the Aura, all read off its attachment so they end when
// it leaves: the layer 7c pump, the keyword grant, and the CR 508.1c
// attack restriction naming the Aura's controller (cmdr2B19AttachedCantAttackYou).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e7bddbd-e448-4aea-9d18-b44bd45837c5",
		Name:         "Vow of Duty",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("vigilance"),
			cmdr2B19AttachedCantAttackYou(),
		},
	})
}
