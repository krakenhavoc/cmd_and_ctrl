package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jade Bearer — Creature — Merfolk Shaman {G}, 1/1 (Rivals of Ixalan,
// #5531):
//
//	"When this creature enters, put a +1/+1 counter on another target
//	 Merfolk you control."
//
// "Another" is object identity (Another), so the Bearer can't pick
// itself; with no other Merfolk of yours the trigger is removed for want
// of a target (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72ddbc9b-f584-433f-9d2d-6e4c374129c4",
		Name:         "Jade Bearer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Jade Bearer — put a +1/+1 counter on another target Merfolk you control",
				rfCreatureFCountersOnTarget(1)),
				Another(TargetCreature("another target Merfolk you control", YouControl(), Subtype("Merfolk")))),
		},
	})
}
