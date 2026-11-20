package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vithian Renegades — Creature — Human Shaman {1}{R}{G}, 3/2 (Alara
// Reborn, #5586):
//
//	"When this creature enters, destroy target artifact."
//
// Viridian Corrupter's enters trigger: the target is chosen as it goes on
// the stack (CR 603.3d), and one that left in response is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78fa248d-7f45-4d43-b3e3-c726db938b0f",
		Name:         "Vithian Renegades",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Vithian Renegades — destroy target artifact", destroyChosenPermanent),
				TargetPermanent("target artifact", Artifact())),
		},
	})
}
