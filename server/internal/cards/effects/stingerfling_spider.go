package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stingerfling Spider — Creature — Spider {4}{G}, 2/5 (Magic 2012, #5573):
//
//	"Reach
//	 When this creature enters, you may destroy target creature with
//	 flying."
//
// Reach, and Hammerheim Deadeye's enters trigger made optional (Optional):
// the target is chosen as it goes on the stack and the destroy is asked
// for as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "056770e2-8ab7-424d-a2d8-7be6d6cc73f1",
		Name:            "Stingerfling Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(WhenThisEnters("Stingerfling Spider — destroy target creature with flying", destroyFirstLegalCardTarget),
				TargetCreature("target creature with flying", HasKeyword("flying"))),
				"Stingerfling Spider — destroy target creature with flying?"),
		},
	})
}
