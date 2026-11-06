package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glass Casket — Artifact {1}{W} (Wilds of Eldraine, #4471):
//
//	"When this artifact enters, exile target creature an opponent controls
//	 with mana value 3 or less until this artifact leaves the battlefield."
//
// The Oblivion Ring family's entry trigger (exileChosenTargetUntilThisLeaves)
// over a creature an opponent controls whose mana value is 3 or less
// (CR 202.3), read at targeting and again as the trigger resolves
// (CR 608.2b). The card comes back under its owner's control when the
// Casket leaves (CR 610.3c); if the Casket has already left when the
// trigger resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9167046-7908-4d16-8cc1-3eb02d1ba547",
		Name:         "Glass Casket",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Glass Casket — exile target creature an opponent controls with mana value 3 or less until this artifact leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Glass Casket — the exiled card returns when Glass Casket leaves the battlefield")),
				TargetCreature("target creature an opponent controls with mana value 3 or less", OpponentControls(), ManaValueLE(3))),
		},
	})
}
