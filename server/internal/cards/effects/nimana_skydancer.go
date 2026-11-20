package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nimana Skydancer — Creature — Human Rogue {2}{B}, 2/1 (Zendikar Rising,
// #5508):
//
//	"Flash
//	 Flying
//	 When this creature enters, target opponent mills two cards."
//
// Flash and flying are printed. The enters trigger is Extractor Demon's
// targetPlayerMillsTwo aimed at an opponent, chosen as it goes on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "27a66362-f2f5-4d46-9faf-a73752753304",
		Name:            "Nimana Skydancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Nimana Skydancer — target opponent mills two cards", targetPlayerMillsTwo),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}
