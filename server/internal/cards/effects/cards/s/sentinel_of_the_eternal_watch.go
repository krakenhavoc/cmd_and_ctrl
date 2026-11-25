package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinel of the Eternal Watch — Creature — Giant Soldier {5}{W}, 4/6 (Magic
// Origins, S82 older sets):
//
//	"Vigilance
//	 At the beginning of combat on each opponent's turn, tap target creature
//	 that player controls."
//
// Web of Inertia's beginning of combat on an opponent's turn, with a target
// chosen as the trigger goes on the stack among the creatures the active
// player controls (the player whose turn it is, "that player") and re-checked as
// it resolves (CR 608.2b). With no such creature the trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "572b1aeb-d2a0-433d-b167-86e683c2127d",
		Name:            "Sentinel of the Eternal Watch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepBeginCombat, false), effects.ByAnOpponent),
					"Sentinel of the Eternal Watch — tap target creature that player controls",
					effects.ThousandMoonsCrackshotTap),
				effects.TargetCreature("target creature that player controls", effects.ControlledByTheActivePlayer())),
		},
	})
}
