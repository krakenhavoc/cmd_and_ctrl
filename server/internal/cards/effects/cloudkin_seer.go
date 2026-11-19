package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudkin Seer — Creature — Elemental Wizard {2}{U}, 2/1 (Ravnica: Clue
// Edition):
//
//	"Flying
//	 When this creature enters, draw a card."
//
// Flying is the engine keyword; the trigger is Elvish Visionary's cantrip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09bd4e1c-9861-481f-80dc-4de955c8d3af",
		Name:            "Cloudkin Seer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cloudkin Seer — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
