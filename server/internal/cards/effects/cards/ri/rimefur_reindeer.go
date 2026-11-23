package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimefur Reindeer — Creature — Elk {3}{W}, 3/4 (Wilds of Eldraine,
// #4471):
//
//	"Whenever an enchantment you control enters, tap target creature an
//	 opponent controls."
//
// The trigger condition is eerie's enchantment half
// (EnchantmentEnteredUnderYourControl): any enchantment entering under
// your control, a Role token included. The target is chosen as the
// trigger goes on the stack (CR 603.3d); with no opposing creature the
// trigger is removed. A target gone by resolution is not tapped (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66dc3237-3543-4fc2-96f1-c8d1eff7c04c",
		Name:         "Rimefur Reindeer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.EnchantmentEnteredUnderYourControl,
				"Rimefur Reindeer — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
