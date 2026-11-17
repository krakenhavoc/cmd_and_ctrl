package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mossbeard Ancient — Creature — Treefolk {5}{G}{G}, 7/7 (Dominaria United, S82):
//
//	"Trample
//	 When this creature enters, you gain 5 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "26afc8c1-c23f-4212-a2aa-386c71c4a7f9",
		Name:            "Mossbeard Ancient",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mossbeard Ancient — you gain 5 life", Do(GainLife{Amount: 5})),
		},
	})
}
