package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Buzz Bots — {1}{U} Artifact Creature — Robot Insect 1/1 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Flying, vigilance
//	 When this creature dies, draw a card."
//
// Ithilien Kingfisher's dies trigger with one more printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "434e720f-2bfa-49b6-a5ac-fe0c0b24764d",
		Name:            "Buzz Bots",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Buzz Bots — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
