package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neighborhood Guardian — Creature — Unicorn {1}{W}, 2/2 (Murders at
// Karlov Manor, #4469):
//
//	"Whenever another creature you control with power 2 or less enters,
//	 target creature you control gets +1/+1 until end of turn."
//
// Mentor of the Meek's trigger condition, read as the creature enters,
// with Daybreak Charger's pump on a creature you control, chosen as the
// trigger goes on the stack: the newcomer, the Guardian or any other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef414347-bbe3-47cb-b900-88275561c43e",
		Name:         "Neighborhood Guardian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AnotherCreatureWithPower2OrLessEnteredUnderYourControl,
					"Neighborhood Guardian — target creature you control gets +1/+1 until end of turn",
					effects.BoostTargetUntilEOT("Neighborhood Guardian — +1/+1", 1, 1)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
