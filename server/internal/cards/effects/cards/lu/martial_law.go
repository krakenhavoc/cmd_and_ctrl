package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Martial Law — Enchantment {2}{W}{W} (Return to Ravnica, #5566):
//
//	"At the beginning of your upkeep, detain target creature an opponent controls."
//
// The target is chosen as the trigger goes on the stack; with no creature an
// opponent controls the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfcf9f38-eb26-4033-b7c0-45ed29417aa5",
		Name:         "Martial Law",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep("Martial Law — detain target creature an opponent controls", effects.OldB423DetainTargets("Martial Law — detain")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
