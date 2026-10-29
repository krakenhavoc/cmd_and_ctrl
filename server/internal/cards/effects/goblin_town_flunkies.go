package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin-town Flunkies — Creature — Goblin Soldier {1}{R}, 1/1 (The Hobbit):
//
//	"Haste
//	 When this creature enters, amass Goblins 1."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "85ceb4ac-7a1c-4c7e-95a3-515a5b5d6116",
		Name:            "Goblin-town Flunkies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Goblin-town Flunkies — amass Goblins 1", Do(Amass{Subtype: "Goblin", N: 1})),
		},
	})
}
