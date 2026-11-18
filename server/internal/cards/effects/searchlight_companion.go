package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Searchlight Companion — Artifact Creature — Drone {3}, 1/1 (Kamigawa:
// Neon Dynasty, #5493):
//
//	"Flying
//	 When this creature enters, create a 1/1 colorless Spirit creature
//	 token."
//
// An enters trigger that makes the printed 1/1 colorless Spirit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "426569c7-d868-4af1-bc83-5b5200532d17",
		Name:            "Searchlight Companion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Searchlight Companion — create a 1/1 colorless Spirit creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 1})),
		},
	})
}
