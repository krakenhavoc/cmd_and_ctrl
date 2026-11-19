package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subjugator Angel — Creature — Angel {4}{W}{W}, 4/3 (Innistrad Remastered):
//
//	"Flying
//	 When this creature enters, tap all creatures your opponents control."
func init() {
	Register(Spec{
		OracleID:        "88b1e06c-5899-4e83-8204-e2c32c0c6aff",
		Name:            "Subjugator Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Subjugator Angel — tap all creatures your opponents control", tapAllCreaturesYourOpponentsControl),
		},
	})
}
