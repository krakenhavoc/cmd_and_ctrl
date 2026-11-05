package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malfunctioning Holodeck — Artifact {2}{W} (Star Trek, #4973):
//
//	"Ward {2}
//	 When this artifact enters, exile target nonland permanent an
//	 opponent controls until this artifact leaves the battlefield."
//
// Banishing Light's two rows (exileUntilThisLeavesTriggers, CR 610.3) on
// an artifact, plus ward {2} (Ward). If the Holodeck leaves before its
// trigger resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f04c38b-d6bd-405f-b69c-009df87801dd",
		Name:         "Malfunctioning Holodeck",
		Completeness: CompletenessFull,
		Triggered: append([]game.TriggeredAbility{
			Ward(WardMana("{2}"), "Malfunctioning Holodeck — ward {2}"),
		}, exileUntilThisLeavesTriggers("Malfunctioning Holodeck")...),
	})
}
