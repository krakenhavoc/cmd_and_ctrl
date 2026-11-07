package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Makeshift Binding — Enchantment {2}{W} (Murders at Karlov Manor,
// #4469):
//
//	"When this enchantment enters, exile target creature an opponent
//	 controls until this enchantment leaves the battlefield. You gain 2
//	 life."
//
// Dimensional Exile's exile (exileChosenTargetUntilThisLeaves): the
// creature returns under its owner's control when the Binding leaves
// (CR 610.3c), and nothing is exiled if the Binding left before the
// trigger resolved (CR 610.3b). The 2 life follows the exile either way.
// A target that is gone by then leaves the trigger with no legal target,
// so it does nothing at all, the life included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "942f3f06-1df0-40b4-8537-71cbec111852",
		Name:         "Makeshift Binding",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Makeshift Binding — exile target creature an opponent controls until this enchantment leaves the battlefield; you gain 2 life",
				exileChosenTargetUntilThisLeavesThenGainLife(
					"Makeshift Binding — the exiled creature returns when Makeshift Binding leaves the battlefield", 2)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
