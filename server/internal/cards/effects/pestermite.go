package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pestermite — Creature — Faerie Rogue {2}{U}, 2/1 (Foundations Jumpstart):
//
//	"Flash
//	 Flying
//	 When this creature enters, you may tap or untap target permanent."
//
// Granite Witness's trigger: the target is chosen as the trigger goes on the
// stack, the "you may" is asked then, and tap or untap is picked as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83cf4235-eb4e-4c55-839a-f2740d41ed5e",
		Name:            "Pestermite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WhenThisEnters("Pestermite — tap or untap target permanent",
						b38TapOrUntapTarget("Pestermite")),
					TargetPermanent("target permanent", Permanent())),
				"Pestermite — tap or untap target permanent?"),
		},
	})
}
