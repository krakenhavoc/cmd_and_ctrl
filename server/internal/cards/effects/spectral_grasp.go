package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectral Grasp — Enchantment — Aura {1}{W} (Commander Masters, #5948):
//
//	"Enchant creature
//	 Enchanted creature can't attack you or planeswalkers you control.
//	 Enchanted creature can't block creatures you control."
//
// "You" is the Aura's controller, read as the layers run, so a stolen Aura
// protects its new controller (CR 508.1c). The attack bar is the shared
// attached-creature restriction; the block bar is a pair rule on the
// enchanted creature as a blocker against attackers the Aura's controller
// controls (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d2377fb-ac3b-453b-b632-abe918d72529",
		Name:         "Spectral Grasp",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{cmdr2B19AttachedCantAttackYou()},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnAttached(), ControlledBySourceController(), "enchanted creature can't block creatures you control"),
		},
	})
}
