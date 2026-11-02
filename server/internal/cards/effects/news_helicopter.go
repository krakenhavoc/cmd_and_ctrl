package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// News Helicopter — Artifact Creature — Construct {3}, 1/1
// (Marvel's Spider-Man):
//
//	"Flying
//	 When this creature enters, create a 1/1 green and white Human
//	 Citizen creature token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "192dd1e2-b6c2-432a-a1b4-e52ff39b3db2",
		Name:            "News Helicopter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("News Helicopter — create a 1/1 green and white Human Citizen creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Human Citizen"), N: 1})),
		},
	})
}
