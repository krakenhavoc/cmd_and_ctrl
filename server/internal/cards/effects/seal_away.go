package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seal Away — Enchantment {1}{W} (Dominaria, #5530):
//
//	"Flash
//	 When this enchantment enters, exile target tapped creature an opponent controls until this enchantment leaves the battlefield."
//
// Super Villain Lockup's shape: flash, and the tapped creature is exiled
// until this enchantment leaves the battlefield (CR 610.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8909015-ea49-47cb-8d37-653904f965bb",
		Name:            "Seal Away",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets: TargetCreature("target tapped creature an opponent controls",
					OpponentControls(), b751Tapped()),
				Key:    "Seal Away — exile target tapped creature an opponent controls",
				Effect: exileChosenTargetUntilThisLeaves("Seal Away — the exiled creature returns when the enchantment leaves the battlefield"),
			},
		},
	})
}
