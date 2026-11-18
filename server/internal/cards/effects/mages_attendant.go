package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mage's Attendant — Creature — Cat Rogue {2}{W}, 3/2 (Streets of New
// Capenna, #5491):
//
//	"When this creature enters, create a 1/1 blue Wizard creature token with
//	 "{1}, Sacrifice this token: Counter target noncreature spell unless its
//	 controller pays {1}.""
//
// An enters trigger making the batch's Wizard template
// (MagesAttendantWizardToken), whose ability is the token's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94b1e7ed-87f3-44e6-bf34-186dd9a1a0d1",
		Name:         "Mage's Attendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mage's Attendant — create a 1/1 blue Wizard creature token with a counter ability",
				Do(CreateToken{Template: MagesAttendantWizardToken(), N: 1})),
		},
	})
}
