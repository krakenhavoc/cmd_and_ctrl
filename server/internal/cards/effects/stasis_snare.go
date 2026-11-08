package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stasis Snare — Enchantment {1}{W}{W} (Foundations, #4465):
//
//	"Flash
//	 When this enchantment enters, exile target creature an opponent
//	 controls until this enchantment leaves the battlefield."
//
// Makeshift Binding's exile without the life gain, on a flash
// enchantment. The exile lasts while the Snare stays (CR 610.3): if it
// leaves before the trigger resolves, nothing is exiled, and once it
// leaves the card returns under its owner's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a83768e-672e-4fec-8931-853f5e96d43d",
		Name:            "Stasis Snare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Stasis Snare — exile target creature an opponent controls until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves(
					"Stasis Snare — the exiled creature returns when Stasis Snare leaves the battlefield")),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
