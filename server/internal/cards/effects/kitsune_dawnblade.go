package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitsune Dawnblade — Creature — Fox Samurai {4}{W}, 2/3 (Saviors of
// Kamigawa, #5601):
//
//	"Bushido 1
//	 When this creature enters, you may tap target creature."
//
// Bushido 1 is the shared keyword trigger (CR 702.45a). The tap is chosen
// as the trigger goes on the stack and the "may" asked as it resolves
// (CR 603.5); a target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd954a2a-5799-4c65-af8a-47adb1ce5e91",
		Name:         "Kitsune Dawnblade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Bushido(1, "Kitsune Dawnblade"),
			Targeting(
				Optional(WhenThisEnters("Kitsune Dawnblade — you may tap target creature", tapTheFirstLegalTarget),
					"Kitsune Dawnblade — tap target creature?"),
				TargetCreature("target creature"),
			),
		},
	})
}
