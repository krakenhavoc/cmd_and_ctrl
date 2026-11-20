package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crossway Vampire — Creature — Vampire {1}{R}{R}, 3/2 (Innistrad, #5572):
//
//	"When this creature enters, target creature can't block this turn."
//
// Summit Intimidator's entry: the creature is targeted as the trigger
// goes on the stack, and can't block for the rest of the turn.
//
// No simplification.
func init() {
	const label = "Crossway Vampire — target creature can't block this turn"
	Register(Spec{
		OracleID:     "543071da-fc89-4274-a2ae-d5d755d1f2a9",
		Name:         "Crossway Vampire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WhenThisEnters(label, targetCreatureCantBlockThisTurn(label)),
					TargetCreature("target creature")),
				ForTargets(CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
