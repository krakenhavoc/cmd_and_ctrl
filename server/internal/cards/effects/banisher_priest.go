package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banisher Priest — Creature — Human Cleric {1}{W}{W}, 2/2:
//
//	"When this creature enters, exile target creature an opponent controls
//	 until this creature leaves the battlefield."
//
// The Oblivion Ring family's shared tail (exileChosenTargetUntilThisLeaves)
// over a creature an opponent controls: the card returns under its owner's
// control when the Priest leaves (CR 610.3c); if the Priest has already left
// when the trigger resolves nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f560b83-32d4-4bb4-a956-8f5db18599db",
		Name:         "Banisher Priest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Banisher Priest — exile target creature an opponent controls until this creature leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Banisher Priest — the exiled creature returns when Banisher Priest leaves the battlefield")),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
