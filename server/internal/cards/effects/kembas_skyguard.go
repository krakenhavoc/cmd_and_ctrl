package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kemba's Skyguard — Creature — Cat Knight {1}{W}{W}, 2/2 (Duel Decks:
// Elspeth vs. Tezzeret, #5578):
//
//	"Flying
//	 When this creature enters, you gain 2 life."
//
// Mandroid Squadron's enters lifegain on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c8167af-2eb0-4f0a-a1b9-b61b65bac261",
		Name:            "Kemba's Skyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kemba's Skyguard — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
