package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ef414347-bbe3-47cb-b900-88275561c43e",
		Name:         "Neighborhood Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AnotherCreatureWithPower2OrLessEnteredUnderYourControl,
					"Neighborhood Guardian — target creature you control gets +1/+1 until end of turn",
					boostTargetUntilEOT("Neighborhood Guardian — +1/+1", 1, 1)),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
