package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manic Vandal — Creature — Human Warrior {2}{R}, 2/2 (Magic 2011,
// #5579):
//
//	"When this creature enters, destroy target artifact."
//
// Uktabi Orangutan's enters trigger: the artifact is chosen as the
// trigger goes on the stack, and with no artifact on the battlefield the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c12bdbf-80cd-4edb-bebb-b4efcf532272",
		Name:         "Manic Vandal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Manic Vandal — destroy target artifact", destroyFirstLegalCardTarget),
				TargetPermanent("target artifact", Artifact())),
		},
	})
}
