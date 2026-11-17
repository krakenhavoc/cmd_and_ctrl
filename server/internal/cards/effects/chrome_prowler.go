package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chrome Prowler — Artifact Creature — Phyrexian Cat {2}{U}, 3/2:
//
//	"Flash
//	 When this creature enters, tap target creature an opponent
//	 controls."
//
// A plain tap: no stun counter and no hold, so the creature untaps in its
// controller's next untap step. The target is chosen as the trigger goes on
// the stack; one that is gone or no longer an opposing creature by
// resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d16c2fa-4118-4b81-92ca-6c8e9efe03b6",
		Name:            "Chrome Prowler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Chrome Prowler — tap target creature an opponent controls", tapEachLegalTargetBody),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
