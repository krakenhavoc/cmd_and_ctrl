package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fervent Cathar — {2}{R} Creature — Human Knight 2/1 (Avacyn Restored, #5570):
//
//	"Haste
//	 When this creature enters, target creature can't block this turn."
//
// Voldaren Duelist's haste and enters trigger: the target can't
// block this turn.
//
// No simplification.
func init() {
	const label = "Fervent Cathar — target creature can't block this turn"
	Register(Spec{
		OracleID:        "94bcae2f-86a6-4d38-b8e1-5ebf69dc7cbc",
		Name:            "Fervent Cathar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WhenThisEnters(label, targetCreatureCantBlockThisTurn(label)),
					TargetCreature("target creature")),
				ForTargets(CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
