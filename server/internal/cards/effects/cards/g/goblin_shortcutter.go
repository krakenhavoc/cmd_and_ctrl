package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Shortcutter — Creature — Goblin Scout {1}{R}, 2/1 (Jumpstart,
// #5510):
//
//	"When this creature enters, target creature can't block this turn."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and the
// restriction is Ahn-Crop Crasher's: the creature cannot be declared as a
// blocker until the turn ends (CR 509.1b). A target gone by resolution makes
// the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Goblin Shortcutter — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "29863706-fa4e-49ed-9f40-69ae503abba3",
		Name:         "Goblin Shortcutter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
				effects.TargetCreature("target creature")),
				effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
