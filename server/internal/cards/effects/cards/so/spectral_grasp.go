package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4d2377fb-ac3b-453b-b632-abe918d72529",
		Name:         "Spectral Grasp",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.Cmdr2B19AttachedCantAttackYou()},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnAttached(), effects.ControlledBySourceController(), "enchanted creature can't block creatures you control"),
		},
	})
}
